package application_test

// Pruebas de la Iteración 2. Cada prueba lleva un código CP-xx que se enlaza con el plan de pruebas y la
// matriz de trazabilidad (requisito → prueba). Usan el repositorio en memoria, que aplica las mismas reglas
// que la base de datos; las reglas de concurrencia sobre PostgreSQL real están en adapters/postgres.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/adapters/seguridad"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

var hashRapido = seguridad.NuevoPBKDF2(1000)

const (
	claveBuena = "Clave-Buena-2026"
	otraClave  = "Otra-Clave-2027"
)

// nuevoGestion crea el servicio con hasheador rápido, el escenario de la Iteración 1 y un administrador.
func nuevoGestion(t *testing.T) (*application.Servicio, *memory.Repo) {
	t.Helper()
	s, r := nuevoEscenario(t, mediodia)
	s.ConHasheador(hashRapido)
	if _, err := s.CrearUsuario(ctx, "admin", "Ana Administradora", claveBuena, domain.RolAdministrador); err != nil {
		t.Fatal(err)
	}
	return s, r
}

func comoAdmin() context.Context {
	return application.ConActor(ctx, application.Actor{ID: 1, Usuario: "admin", Rol: domain.RolAdministrador})
}

func quiere(t *testing.T, nombre string, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("%s: esperaba %v, obtuve %v", nombre, want, got)
	}
}

// ================= Usuarios del sistema =================

// CP-U01 · RF-13 · Inicio de sesión válido (el usuario no distingue mayúsculas ni espacios al borde).
func TestAutenticar_Valido(t *testing.T) {
	s, _ := nuevoGestion(t)
	for _, login := range []string{"admin", " ADMIN ", "Admin"} {
		u, err := s.Autenticar(ctx, login, claveBuena)
		if err != nil || u.Usuario != "admin" || u.Rol != domain.RolAdministrador {
			t.Errorf("login %q: %v %+v", login, err, u)
		}
	}
	reg, _ := s.Auditoria(ctx, 10)
	if len(reg) != 3 || reg[0].Accion != "INICIO_SESION" || reg[0].Usuario != "admin" {
		t.Errorf("cada inicio de sesión debía quedar auditado: %+v", reg)
	}
}

// CP-U02 · RF-13 · Todo fallo de autenticación responde igual (no revela qué usuarios existen).
func TestAutenticar_Rechazos(t *testing.T) {
	s, _ := nuevoGestion(t)
	if _, err := s.CrearUsuario(ctx, "bloqueado", "Usuario Bloqueado", claveBuena, domain.RolEncargado); err != nil {
		t.Fatal(err)
	}
	u, _ := s.Usuarios(ctx)
	var id int
	for _, x := range u {
		if x.Usuario == "bloqueado" {
			id = x.ID
		}
	}
	off := false
	if _, err := s.ActualizarUsuario(ctx, id, application.CambiosUsuario{Activo: &off}); err != nil {
		t.Fatal(err)
	}
	casos := map[string][2]string{
		"clave incorrecta":           {"admin", "Clave-Mala-2026"},
		"clave vacía":                {"admin", ""},
		"usuario inexistente":        {"fantasma", claveBuena},
		"usuario inactivo con clave": {"bloqueado", claveBuena},
		"usuario con formato malo":   {"a b", claveBuena},
		"usuario vacío":              {"", claveBuena},
	}
	for nombre, c := range casos {
		_, err := s.Autenticar(ctx, c[0], c[1])
		quiere(t, nombre, err, domain.ErrCredenciales)
	}
}

// CP-U03 · RF-14 · Crear usuario: la contraseña se guarda protegida, nunca en claro.
func TestCrearUsuario_Valido(t *testing.T) {
	s, r := nuevoGestion(t)
	u, err := s.CrearUsuario(ctx, " Luis.Perez ", "  Luis   Pérez  ", claveBuena, domain.RolEncargado)
	if err != nil || u.ID == 0 || u.Usuario != "luis.perez" || u.Nombre != "Luis Pérez" || !u.Activo || u.Rol != domain.RolEncargado {
		t.Fatalf("creación: %v %+v", err, u)
	}
	guardado, _ := r.UsuarioPorID(ctx, u.ID)
	if guardado.ClaveHash == "" || strings.Contains(guardado.ClaveHash, claveBuena) || !hashRapido.Verificar(guardado.ClaveHash, claveBuena) {
		t.Fatalf("la clave debía guardarse protegida: %q", guardado.ClaveHash)
	}
	if _, err := s.Autenticar(ctx, "luis.perez", claveBuena); err != nil {
		t.Fatalf("el usuario nuevo debía poder iniciar sesión: %v", err)
	}
}

// CP-U04 · RF-14, RNF-10 · Límites y rechazos al crear usuarios.
func TestCrearUsuario_LimitesYRechazos(t *testing.T) {
	s, _ := nuevoGestion(t)
	letras := func(n int) string { return strings.Repeat("a", n) }
	ok := []struct{ nombre, login, nom, clave string }{
		{"login de 3 caracteres", "abc", "Nombre", claveBuena},
		{"login de 30 caracteres", letras(30), "Nombre", claveBuena},
		{"nombre de 100 caracteres", "nombre.largo", letras(100), claveBuena},
		{"clave de 10 caracteres", "clave.min", "Nombre", "abcdefgh12"},
		{"clave de 128 caracteres", "clave.max", "Nombre", letras(127) + "1"},
	}
	for _, c := range ok {
		if _, err := s.CrearUsuario(ctx, c.login, c.nom, c.clave, domain.RolEncargado); err != nil {
			t.Errorf("%s debía aceptarse: %v", c.nombre, err)
		}
	}
	malos := []struct {
		nombre, login, nom, clave string
		rol                       domain.Rol
		quiere                    error
	}{
		{"login de 2 caracteres", "ab", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"login de 31 caracteres", letras(31), "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"login con espacio", "a b", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"login con arroba", "a@b.co", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"login con tilde", "perú", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"nombre vacío", "sin.nombre", "   ", claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"nombre de 101 caracteres", "nombre.101", letras(101), claveBuena, domain.RolEncargado, domain.ErrUsuarioInvalido},
		{"rol inválido", "rol.malo", "N", claveBuena, "SUPERUSUARIO", domain.ErrRolInvalido},
		{"rol vacío", "rol.vacio", "N", claveBuena, "", domain.ErrRolInvalido},
		{"clave de 9 caracteres", "clave.9", "N", "abcdefg12", domain.RolEncargado, domain.ErrClaveDebil},
		{"clave de 129 caracteres", "clave.129", "N", letras(128) + "1", domain.RolEncargado, domain.ErrClaveDebil},
		{"clave sin números", "sin.num", "N", "SoloLetrasLargas", domain.RolEncargado, domain.ErrClaveDebil},
		{"clave sin letras", "sin.let", "N", "1234567890123", domain.RolEncargado, domain.ErrClaveDebil},
		{"clave igual al usuario", "usuario1234", "N", "usuario1234", domain.RolEncargado, domain.ErrClaveDebil},
		{"usuario repetido", "admin", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioDuplicado},
		{"usuario repetido en mayúsculas", "ADMIN", "N", claveBuena, domain.RolEncargado, domain.ErrUsuarioDuplicado},
	}
	for _, c := range malos {
		_, err := s.CrearUsuario(ctx, c.login, c.nom, c.clave, c.rol)
		quiere(t, c.nombre, err, c.quiere)
	}
}

// CP-U05 · RF-15 · Editar nombre, rol y estado; los campos no enviados no cambian.
func TestActualizarUsuario(t *testing.T) {
	s, _ := nuevoGestion(t)
	e, _ := s.CrearUsuario(ctx, "encargado", "Luis Pérez", claveBuena, domain.RolEncargado)
	nombre, rol, off := "Luis Alberto Pérez", domain.RolAdministrador, false
	u, err := s.ActualizarUsuario(comoAdmin(), e.ID, application.CambiosUsuario{Nombre: &nombre})
	if err != nil || u.Nombre != nombre || u.Rol != domain.RolEncargado || !u.Activo {
		t.Fatalf("solo nombre: %v %+v", err, u)
	}
	if u, err = s.ActualizarUsuario(comoAdmin(), e.ID, application.CambiosUsuario{Rol: &rol, Activo: &off}); err != nil || u.Rol != rol || u.Activo {
		t.Fatalf("rol y estado: %v %+v", err, u)
	}
	_, err = s.ActualizarUsuario(ctx, e.ID, application.CambiosUsuario{Rol: func() *domain.Rol { r := domain.Rol("X"); return &r }()})
	quiere(t, "rol inválido", err, domain.ErrRolInvalido)
	vacio := " "
	_, err = s.ActualizarUsuario(ctx, e.ID, application.CambiosUsuario{Nombre: &vacio})
	quiere(t, "nombre vacío", err, domain.ErrUsuarioInvalido)
	_, err = s.ActualizarUsuario(ctx, 999, application.CambiosUsuario{Nombre: &nombre})
	quiere(t, "usuario inexistente", err, domain.ErrUsuarioNoExiste)
	reg, _ := s.Auditoria(ctx, 10)
	if len(reg) < 2 || reg[0].Accion != "USUARIO_EDITADO" || !strings.Contains(reg[0].Detalle, "desactivado") || reg[0].Usuario != "admin" {
		t.Errorf("la edición debía quedar auditada con el actor: %+v", reg)
	}
}

// CP-U06 · RF-15 · Nunca se queda el sistema sin un administrador activo.
func TestUltimoAdministradorProtegido(t *testing.T) {
	s, _ := nuevoGestion(t)
	off, enc := false, domain.RolEncargado
	_, err := s.ActualizarUsuario(ctx, 1, application.CambiosUsuario{Activo: &off})
	quiere(t, "desactivar al único administrador", err, domain.ErrUltimoAdministrador)
	_, err = s.ActualizarUsuario(ctx, 1, application.CambiosUsuario{Rol: &enc})
	quiere(t, "degradar al único administrador", err, domain.ErrUltimoAdministrador)
	if u, _ := s.Usuario(ctx, 1); !u.Activo || u.Rol != domain.RolAdministrador {
		t.Fatalf("el rechazo no debía cambiar nada: %+v", u)
	}
	seg, _ := s.CrearUsuario(ctx, "admin2", "Segundo Admin", claveBuena, domain.RolAdministrador)
	if _, err := s.ActualizarUsuario(ctx, 1, application.CambiosUsuario{Activo: &off}); err != nil {
		t.Fatalf("con otro administrador activo sí se puede desactivar: %v", err)
	}
	_, err = s.ActualizarUsuario(ctx, seg.ID, application.CambiosUsuario{Activo: &off})
	quiere(t, "ahora el segundo es el último", err, domain.ErrUltimoAdministrador)
}

// CP-U07 · RF-15, RNF-11 · Dos administradores que se desactivan a la vez: solo uno puede hacerlo.
func TestUltimoAdministrador_Simultaneo(t *testing.T) {
	s, _ := nuevoGestion(t)
	seg, _ := s.CrearUsuario(ctx, "admin2", "Segundo Admin", claveBuena, domain.RolAdministrador)
	off := false
	var ok, rechazados int32
	var wg sync.WaitGroup
	for _, id := range []int{1, seg.ID} {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := s.ActualizarUsuario(ctx, id, application.CambiosUsuario{Activo: &off})
			if err == nil {
				atomic.AddInt32(&ok, 1)
			} else if errors.Is(err, domain.ErrUltimoAdministrador) {
				atomic.AddInt32(&rechazados, 1)
			}
		}(id)
	}
	wg.Wait()
	if ok != 1 || rechazados != 1 {
		t.Fatalf("esperaba 1 éxito y 1 rechazo, hubo %d y %d", ok, rechazados)
	}
}

// CP-U08 · RF-16 · Cambiar la propia contraseña exige la actual y cumple la política.
func TestCambiarMiClave(t *testing.T) {
	s, _ := nuevoGestion(t)
	quiere(t, "actual incorrecta", s.CambiarMiClave(ctx, 1, "Clave-Mala-2026", otraClave), domain.ErrClaveActualIncorrecta)
	quiere(t, "nueva débil", s.CambiarMiClave(ctx, 1, claveBuena, "corta1"), domain.ErrClaveDebil)
	quiere(t, "usuario inexistente", s.CambiarMiClave(ctx, 99, claveBuena, otraClave), domain.ErrUsuarioNoExiste)
	if err := s.CambiarMiClave(comoAdmin(), 1, claveBuena, otraClave); err != nil {
		t.Fatal(err)
	}
	_, err := s.Autenticar(ctx, "admin", claveBuena)
	quiere(t, "la clave anterior deja de servir", err, domain.ErrCredenciales)
	if _, err := s.Autenticar(ctx, "admin", otraClave); err != nil {
		t.Fatalf("la clave nueva debía servir: %v", err)
	}
}

// CP-U09 · RF-16 · El administrador restablece la contraseña de otro usuario.
func TestRestablecerClave(t *testing.T) {
	s, _ := nuevoGestion(t)
	e, _ := s.CrearUsuario(ctx, "encargado", "Luis Pérez", claveBuena, domain.RolEncargado)
	quiere(t, "clave débil", s.RestablecerClave(ctx, e.ID, "123"), domain.ErrClaveDebil)
	quiere(t, "usuario inexistente", s.RestablecerClave(ctx, 99, otraClave), domain.ErrUsuarioNoExiste)
	if err := s.RestablecerClave(comoAdmin(), e.ID, otraClave); err != nil {
		t.Fatal(err)
	}
	reg, _ := s.Auditoria(ctx, 5)
	if reg[0].Accion != "CLAVE_RESTABLECIDA" || strings.Contains(reg[0].Detalle, otraClave) {
		t.Errorf("la auditoría no debe contener contraseñas: %+v", reg[0])
	}
	if _, err := s.Autenticar(ctx, "encargado", otraClave); err != nil {
		t.Fatalf("la clave restablecida debía servir: %v", err)
	}
}

// CP-U10 · RF-14 · El primer administrador se crea solo cuando no hay usuarios.
func TestAsegurarAdministrador(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	s.ConHasheador(hashRapido)
	creado, err := s.AsegurarAdministrador(ctx, "admin", "Administrador", claveBuena)
	if err != nil || !creado {
		t.Fatalf("primera vez: %v %v", creado, err)
	}
	creado, err = s.AsegurarAdministrador(ctx, "otro", "Otro", claveBuena)
	if err != nil || creado {
		t.Fatalf("con usuarios existentes no debía crear: %v %v", creado, err)
	}
	if lista, _ := s.Usuarios(ctx); len(lista) != 1 {
		t.Fatalf("debía haber un solo usuario: %+v", lista)
	}
}

// CP-U11 · RNF-08 · Sin hasheador configurado el servicio se niega a manejar contraseñas.
func TestSinHasheador(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	_, err := s.CrearUsuario(ctx, "admin", "Admin", claveBuena, domain.RolAdministrador)
	quiere(t, "crear", err, application.ErrSinHasheador)
	_, err = s.Autenticar(ctx, "admin", claveBuena)
	quiere(t, "autenticar", err, application.ErrSinHasheador)
}

// ================= Clientes y vehículos =================

func clientesDePrueba(t *testing.T) (*application.Servicio, *memory.Repo) {
	t.Helper()
	s, r := nuevoGestion(t)
	r.AgregarClienteCompleto(10, "1020304050", "María Fernanda Rojas", "3001112233", "2026-10-31")
	r.AgregarClienteCompleto(11, "9988776655", "Pedro Pablo Gómez", "", "")
	r.AgregarVehiculo(domain.Vehiculo{ID: 50, Placa: "MFR123", Tipo: domain.Carro, ClienteID: 10})
	r.AgregarVehiculo(domain.Vehiculo{ID: 51, Placa: "MFR12A", Tipo: domain.Moto, ClienteID: 10})
	return s, r
}

// CP-C01 · RF-09 · Buscar clientes por documento o nombre, con sus vehículos.
func TestBuscarClientes(t *testing.T) {
	s, _ := clientesDePrueba(t)
	casos := map[string]int{"1020304050": 1, "mar": 2, "MARÍA": 1, "gómez": 2, "rojas": 2, "zzz": 0, "": 4, "  998877  ": 1}
	for q, esperados := range casos {
		lista, err := s.BuscarClientes(ctx, q)
		if err != nil || len(lista) != esperados {
			t.Errorf("búsqueda %q: esperaba %d, obtuve %d (%v)", q, esperados, len(lista), err)
		}
	}
	lista, _ := s.BuscarClientes(ctx, "zzz")
	if lista == nil {
		t.Error("sin resultados debía devolver lista vacía, no nula")
	}
	maria, _ := s.BuscarClientes(ctx, "1020304050")
	if len(maria[0].Vehiculos) != 2 || maria[0].Vehiculos[0].Placa != "MFR123" {
		t.Errorf("debía traer sus 2 vehículos ordenados por placa: %+v", maria[0].Vehiculos)
	}
}

// CP-C02 · RF-09 · La vigencia se calcula con el día de Colombia: vale el último día, no el siguiente.
func TestVigenciaDeMensualidad(t *testing.T) {
	s, r := clientesDePrueba(t) // hoy = 2026-10-18; María vigente hasta 2026-10-31
	c, _ := s.Cliente(ctx, 10)
	if !c.Vigente {
		t.Error("María debía estar vigente")
	}
	if c, _ = s.Cliente(ctx, 11); c.Vigente || len(c.Mensualidades) != 0 {
		t.Errorf("Pedro no tiene mensualidad: %+v", c)
	}
	r.AgregarClienteCompleto(12, "1", "Vencida Ayer", "", "2026-10-17")
	r.AgregarClienteCompleto(13, "2", "Vence Hoy", "", "2026-10-18")
	if c, _ = s.Cliente(ctx, 12); c.Vigente {
		t.Error("venció ayer: no debía estar vigente")
	}
	if c, _ = s.Cliente(ctx, 13); !c.Vigente {
		t.Error("vence hoy: debía estar vigente (último día incluido)")
	}
	_, err := s.Cliente(ctx, 999)
	quiere(t, "cliente inexistente", err, domain.ErrClienteNoExiste)
}

// CP-C03 · RF-10 · Editar cliente: el documento no cambia y se valida el nombre.
func TestActualizarCliente(t *testing.T) {
	s, _ := clientesDePrueba(t)
	c, err := s.ActualizarCliente(comoAdmin(), 10, "  María   Rojas  ", " 3150000000 ")
	if err != nil || c.Nombre != "María Rojas" || c.Telefono != "3150000000" || c.Documento != "1020304050" {
		t.Fatalf("edición: %v %+v", err, c)
	}
	if c, err = s.ActualizarCliente(ctx, 10, "María Rojas", ""); err != nil || c.Telefono != "" {
		t.Fatalf("el teléfono es opcional y puede borrarse: %v %+v", err, c)
	}
	_, err = s.ActualizarCliente(ctx, 10, "   ", "")
	quiere(t, "nombre vacío", err, domain.ErrClienteInvalido)
	_, err = s.ActualizarCliente(ctx, 10, strings.Repeat("n", 101), "")
	quiere(t, "nombre de 101 caracteres", err, domain.ErrClienteInvalido)
	_, err = s.ActualizarCliente(ctx, 10, "Nombre", strings.Repeat("1", 21))
	quiere(t, "teléfono de 21 caracteres", err, domain.ErrClienteInvalido)
	_, err = s.ActualizarCliente(ctx, 999, "Nombre", "")
	quiere(t, "cliente inexistente", err, domain.ErrClienteNoExiste)
}

// CP-C04 · RF-11 · Registrar mensualidades sin solapamientos.
func TestAgregarMensualidad_Solapes(t *testing.T) {
	s, r2 := clientesDePrueba(t)
	r2.AgregarClienteCompleto(20, "777", "Sin Mensualidad", "", "")
	m := func(i, f string) application.MensualidadEntrada {
		return application.MensualidadEntrada{FechaInicio: i, FechaFin: f}
	}
	if _, err := s.AgregarMensualidad(comoAdmin(), 20, m("2026-11-01", "2026-11-30")); err != nil {
		t.Fatalf("primera mensualidad: %v", err)
	}
	solapadas := map[string]application.MensualidadEntrada{
		"idéntica":                m("2026-11-01", "2026-11-30"),
		"comparte el último día":  m("2026-11-30", "2026-12-15"),
		"comparte el primer día":  m("2026-10-15", "2026-11-01"),
		"contenida":               m("2026-11-10", "2026-11-12"),
		"contiene a la existente": m("2026-10-01", "2026-12-31"),
		"un solo día dentro":      m("2026-11-15", "2026-11-15"),
		"cruza por el inicio":     m("2026-10-20", "2026-11-05"),
		"cruza por el final":      m("2026-11-25", "2026-12-05"),
	}
	for nombre, in := range solapadas {
		_, err := s.AgregarMensualidad(ctx, 20, in)
		quiere(t, nombre, err, domain.ErrMensualidadSolapada)
	}
	for nombre, in := range map[string]application.MensualidadEntrada{
		"empieza el día siguiente": m("2026-12-01", "2026-12-31"),
		"termina el día anterior":  m("2026-10-01", "2026-10-31"),
		"un solo día aparte":       m("2027-01-15", "2027-01-15"),
	} {
		if _, err := s.AgregarMensualidad(ctx, 20, in); err != nil {
			t.Errorf("%s debía aceptarse: %v", nombre, err)
		}
	}
	for nombre, in := range map[string]application.MensualidadEntrada{
		"fechas invertidas": m("2026-12-31", "2026-12-01"), "fecha mal escrita": m("31/12/2026", "2027-01-01"),
		"mes 13": m("2027-13-01", "2027-14-01"), "30 de febrero": m("2027-02-30", "2027-03-05"), "vacías": m("", ""),
	} {
		_, err := s.AgregarMensualidad(ctx, 20, in)
		quiere(t, nombre, err, domain.ErrMensualidadInvalida)
	}
	_, err := s.AgregarMensualidad(ctx, 999, m("2030-01-01", "2030-01-31"))
	quiere(t, "cliente inexistente", err, domain.ErrClienteNoExiste)
}

// CP-C05 · RF-11, RF-02 · Al renovar la mensualidad, el vehículo vuelve a poder ingresar.
func TestMensualidadRenovadaPermiteIngresar(t *testing.T) {
	s, _ := nuevoGestion(t)
	// DEF456 pertenece al cliente 2, con la mensualidad vencida.
	_, err := s.RegistrarIngreso(ctx, "DEF456", 1, nil)
	quiere(t, "antes de renovar", err, domain.ErrMensualidadVencida)
	if _, err := s.AgregarMensualidad(ctx, 2, application.MensualidadEntrada{FechaInicio: "2026-10-01", FechaFin: "2026-10-31"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegistrarIngreso(ctx, "DEF456", 1, nil); err != nil {
		t.Fatalf("tras renovar debía poder ingresar: %v", err)
	}
}

// CP-C06 · RF-12 · Cambiar el tipo de un vehículo solo si no está dentro.
func TestActualizarTipoVehiculo(t *testing.T) {
	s, _ := nuevoGestion(t)
	v, err := s.ActualizarTipoVehiculo(comoAdmin(), 2, domain.Moto) // DEF456 está fuera
	if err != nil || v.Tipo != domain.Moto {
		t.Fatalf("cambio de tipo: %v %+v", err, v)
	}
	_, err = s.ActualizarTipoVehiculo(ctx, 2, "BUS")
	quiere(t, "tipo inválido", err, domain.ErrTipoVehiculoInvalido)
	_, err = s.ActualizarTipoVehiculo(ctx, 999, domain.Carro)
	quiere(t, "vehículo inexistente", err, domain.ErrVehiculoNoRegistrado)
	if _, err := s.RegistrarIngreso(ctx, "ABC123", 1, nil); err != nil {
		t.Fatal(err)
	}
	_, err = s.ActualizarTipoVehiculo(ctx, 1, domain.Moto)
	quiere(t, "vehículo dentro", err, domain.ErrVehiculoDentro)
	if _, err := s.RegistrarSalida(ctx, "ABC123", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ActualizarTipoVehiculo(ctx, 1, domain.Moto); err != nil {
		t.Fatalf("tras la salida sí se puede: %v", err)
	}
}

// ================= Celdas =================

func contiene(cs []domain.Celda, codigo string) bool {
	for _, c := range cs {
		if c.Codigo == codigo {
			return true
		}
	}
	return false
}

// CP-E01 · RF-19 · Crear celda: nace disponible, el código se normaliza y ya se puede usar.
func TestCrearCelda(t *testing.T) {
	s, _ := nuevoGestion(t)
	c, err := s.CrearCelda(comoAdmin(), " b-07 ", domain.Carro)
	if err != nil || c.ID == 0 || c.Codigo != "B-07" || c.Estado != domain.CeldaDisponible {
		t.Fatalf("creación: %v %+v", err, c)
	}
	disp, _ := s.ConsultarCeldas(ctx, "DISPONIBLE")
	if !contiene(disp, "B-07") {
		t.Errorf("la celda nueva debía aparecer como disponible: %+v", disp)
	}
	if _, err := s.RegistrarIngreso(ctx, "ABC123", c.ID, nil); err != nil {
		t.Fatalf("la celda nueva debía poder ocuparse: %v", err)
	}
}

// CP-E02 · RF-19 · Límites y rechazos al crear celdas.
func TestCrearCelda_LimitesYRechazos(t *testing.T) {
	s, _ := nuevoGestion(t)
	for _, cod := range []string{"A", "ABCDE-1234", "0", "Z-99"} {
		if _, err := s.CrearCelda(ctx, cod, domain.Moto); err != nil {
			t.Errorf("código %q debía aceptarse: %v", cod, err)
		}
	}
	for nombre, c := range map[string]struct {
		cod  string
		tipo domain.TipoVehiculo
		want error
	}{
		"código vacío":                 {"", domain.Carro, domain.ErrCeldaInvalida},
		"código solo espacios":         {"   ", domain.Carro, domain.ErrCeldaInvalida},
		"código de 11 caracteres":      {"ABCDE-12345", domain.Carro, domain.ErrCeldaInvalida},
		"código con espacio":           {"A 01", domain.Carro, domain.ErrCeldaInvalida},
		"código con guion bajo":        {"A_01", domain.Carro, domain.ErrCeldaInvalida},
		"tipo inválido":                {"C-01", "BUS", domain.ErrCeldaInvalida},
		"tipo vacío":                   {"C-02", "", domain.ErrCeldaInvalida},
		"código repetido":              {"A-01", domain.Carro, domain.ErrCeldaDuplicada},
		"código repetido en minúscula": {"a-01", domain.Moto, domain.ErrCeldaDuplicada},
	} {
		_, err := s.CrearCelda(ctx, c.cod, c.tipo)
		quiere(t, nombre, err, c.want)
	}
}

// CP-E03 · RF-20 · Editar código y tipo solo si la celda no está ocupada.
func TestActualizarCelda(t *testing.T) {
	s, _ := nuevoGestion(t)
	cod, tipo := "a-09", domain.Moto
	c, err := s.ActualizarCelda(comoAdmin(), 5, &cod, &tipo)
	if err != nil || c.Codigo != "A-09" || c.Tipo != domain.Moto {
		t.Fatalf("edición: %v %+v", err, c)
	}
	solo := "A-10"
	if c, err = s.ActualizarCelda(ctx, 5, &solo, nil); err != nil || c.Codigo != "A-10" || c.Tipo != domain.Moto {
		t.Fatalf("solo el código: %v %+v", err, c)
	}
	dup := "A-01"
	_, err = s.ActualizarCelda(ctx, 5, &dup, nil)
	quiere(t, "código de otra celda", err, domain.ErrCeldaDuplicada)
	if _, err = s.ActualizarCelda(ctx, 5, &solo, nil); err != nil {
		t.Errorf("conservar su propio código no es duplicado: %v", err)
	}
	malo := "A 1"
	_, err = s.ActualizarCelda(ctx, 5, &malo, nil)
	quiere(t, "código inválido", err, domain.ErrCeldaInvalida)
	bus := domain.TipoVehiculo("BUS")
	_, err = s.ActualizarCelda(ctx, 5, nil, &bus)
	quiere(t, "tipo inválido", err, domain.ErrCeldaInvalida)
	_, err = s.ActualizarCelda(ctx, 999, &solo, nil)
	quiere(t, "celda inexistente", err, domain.ErrCeldaNoExiste)
	if _, err := s.RegistrarIngreso(ctx, "ABC123", 1, nil); err != nil {
		t.Fatal(err)
	}
	nuevo := "Z-01"
	_, err = s.ActualizarCelda(ctx, 1, &nuevo, nil)
	quiere(t, "celda ocupada", err, domain.ErrCeldaOcupada)
	if oc, _ := s.ConsultarCeldas(ctx, "OCUPADA"); !contiene(oc, "A-01") || contiene(oc, "Z-01") {
		t.Errorf("el rechazo no debía cambiar la celda: %+v", oc)
	}
}

// CP-E04 · RF-21 · Inactivar y reactivar celdas; la ocupación no se fija a mano.
func TestCambiarEstadoCelda(t *testing.T) {
	s, _ := nuevoGestion(t)
	c, err := s.CambiarEstadoCelda(comoAdmin(), 5, domain.CeldaInactiva)
	if err != nil || c.Estado != domain.CeldaInactiva {
		t.Fatalf("inactivar: %v %+v", err, c)
	}
	_, err = s.RegistrarIngreso(ctx, "ABC123", 5, nil)
	quiere(t, "ingreso en celda inactiva", err, domain.ErrCeldaNoDisponible)
	if c, err = s.CambiarEstadoCelda(ctx, 5, domain.CeldaDisponible); err != nil || c.Estado != domain.CeldaDisponible {
		t.Fatalf("reactivar: %v %+v", err, c)
	}
	if _, err = s.RegistrarIngreso(ctx, "ABC123", 5, nil); err != nil {
		t.Fatalf("reactivada debía aceptar ingresos: %v", err)
	}
	for _, e := range []domain.EstadoCelda{domain.CeldaOcupada, "ROTA", ""} {
		_, err = s.CambiarEstadoCelda(ctx, 1, e)
		quiere(t, "estado "+string(e), err, domain.ErrEstadoCeldaInvalido)
	}
	_, err = s.CambiarEstadoCelda(ctx, 5, domain.CeldaInactiva)
	quiere(t, "inactivar celda ocupada", err, domain.ErrCeldaOcupada)
	_, err = s.CambiarEstadoCelda(ctx, 999, domain.CeldaInactiva)
	quiere(t, "celda inexistente", err, domain.ErrCeldaNoExiste)
	if _, err := s.RegistrarSalida(ctx, "ABC123", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CambiarEstadoCelda(ctx, 5, domain.CeldaInactiva); err != nil {
		t.Fatalf("tras la salida ya se puede inactivar: %v", err)
	}
}

// ================= Auditoría =================

// CP-A01 · RF-18 · Las operaciones quedan con su autor; sin actor no se audita; más reciente primero.
func TestAuditoria(t *testing.T) {
	s, _ := nuevoGestion(t)
	enc := application.ConActor(ctx, application.Actor{ID: 7, Usuario: "encargado", Rol: domain.RolEncargado})
	if _, err := s.RegistrarIngreso(enc, "ABC123", 1, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegistrarSalida(enc, "ABC123", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegistrarNovedad(enc, "ABC123", application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: "x"}); err == nil {
		t.Fatal("la novedad sobre un vehículo fuera debía fallar")
	}
	if _, err := s.RegistrarIngreso(ctx, "DEF456", 2, nil); err == nil { // sin actor y rechazado: no audita
		t.Fatal("DEF456 tiene la mensualidad vencida")
	}
	reg, _ := s.Auditoria(ctx, 50)
	if len(reg) != 2 || reg[0].Accion != "SALIDA" || reg[1].Accion != "INGRESO" || reg[0].Usuario != "encargado" || reg[0].UsuarioID != 7 {
		t.Fatalf("auditoría inesperada: %+v", reg)
	}
	if reg[1].Detalle != "ABC123 en A-01" || reg[0].Momento.IsZero() {
		t.Errorf("detalle/momento: %+v", reg)
	}
	for i := 0; i < 60; i++ {
		s.CrearCelda(comoAdmin(), fmt.Sprintf("T-%d", i), domain.Carro)
	}
	if reg, _ = s.Auditoria(ctx, 0); len(reg) != 50 {
		t.Errorf("sin límite válido debía devolver 50, devolvió %d", len(reg))
	}
	if reg, _ = s.Auditoria(ctx, 5); len(reg) != 5 {
		t.Errorf("con límite 5 debía devolver 5, devolvió %d", len(reg))
	}
	if reg, _ = s.Auditoria(ctx, 201); len(reg) != 50 {
		t.Errorf("un límite fuera de rango vuelve al valor por defecto (50), devolvió %d", len(reg))
	}
}
