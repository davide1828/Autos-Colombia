package postgres_test

// Iteración 2 contra PostgreSQL real. Aquí se comprueban las reglas que solo la base de datos puede garantizar
// con operaciones simultáneas (último administrador, solapamiento de mensualidades, códigos únicos) y la
// inmutabilidad de la auditoría. Códigos CP-Pxx enlazados con el plan de pruebas.

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/adapters/postgres"
	"autoscolombia/parqueadero/internal/adapters/seguridad"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

var relojOct = func() time.Time { return time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC) }

func servicioReal(db *sql.DB) *application.Servicio {
	db.SetMaxOpenConns(20)
	return application.NuevoServicio(postgres.Nuevo(db), relojOct).ConHasheador(seguridad.NuevoPBKDF2(1000))
}

func mustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

func contar(t *testing.T, db *sql.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

const claveP = "Clave-Prueba-2026"

// CP-P01 · RNF-05 · El esquema es repetible (idempotente) y la base rechaza roles y usuarios inválidos.
func TestEsquemaIdempotenteYRestricciones(t *testing.T) {
	db := baseDeDatos(t)
	esquema, _ := os.ReadFile("../../../db/schema.sql")
	for i := 0; i < 2; i++ {
		if _, err := db.Exec(string(esquema)); err != nil {
			t.Fatalf("reaplicar schema.sql (vuelta %d): %v", i+1, err)
		}
	}
	for nombre, q := range map[string]string{
		"rol inválido":                `INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('a.b','N','ROOT','h')`,
		"usuario en mayúsculas":       `INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('ABC','N','ENCARGADO','h')`,
		"usuario de 2 caracteres":     `INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('ab','N','ENCARGADO','h')`,
		"usuario con espacio":         `INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('a b','N','ENCARGADO','h')`,
		"sin hash":                    `INSERT INTO usuario (usuario, nombre, rol) VALUES ('abc','N','ENCARGADO')`,
		"auditoría sin usuario":       `INSERT INTO auditoria (accion, entidad) VALUES ('X','y')`,
		"auditoría con usuario ajeno": `INSERT INTO auditoria (id_usuario, usuario, accion, entidad) VALUES (999,'x','X','y')`,
	} {
		if _, err := db.Exec(q); err == nil {
			t.Errorf("%s debía rechazarse por la base de datos", nombre)
		}
	}
	mustExec(t, db, `INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('abc','N','ENCARGADO','h')`)
	if _, err := db.Exec(`INSERT INTO usuario (usuario, nombre, rol, clave_hash) VALUES ('abc','Otro','ENCARGADO','h')`); err == nil {
		t.Error("el usuario repetido debía rechazarse")
	}
}

// CP-P02 · RF-13, RF-14, RNF-08 · Crear y autenticar usuarios de punta a punta; la clave jamás se guarda en claro.
func TestUsuariosContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	s := servicioReal(db)
	u, err := s.CrearUsuario(ctx, "Ana.R", " Ana  Ríos ", claveP, domain.RolAdministrador)
	if err != nil || u.ID == 0 || u.Usuario != "ana.r" {
		t.Fatalf("creación: %v %+v", err, u)
	}
	_, err = s.CrearUsuario(ctx, "ANA.R", "Otra", claveP, domain.RolEncargado)
	if !errors.Is(err, domain.ErrUsuarioDuplicado) {
		t.Errorf("duplicado: %v", err)
	}
	var hash, nombre string
	db.QueryRow(`SELECT clave_hash, nombre FROM usuario WHERE usuario='ana.r'`).Scan(&hash, &nombre)
	if strings.Contains(hash, claveP) || !strings.HasPrefix(hash, "pbkdf2-sha256$1000$") || nombre != "Ana Ríos" {
		t.Fatalf("guardado: hash=%q nombre=%q", hash, nombre)
	}
	if got, err := s.Autenticar(ctx, "ana.r", claveP); err != nil || got.ID != u.ID {
		t.Fatalf("autenticar: %v %+v", err, got)
	}
	if _, err := s.Autenticar(ctx, "ana.r", claveP+"x"); !errors.Is(err, domain.ErrCredenciales) {
		t.Errorf("clave incorrecta: %v", err)
	}
	if err := s.CambiarMiClave(ctx, u.ID, claveP, "Otra-Clave-2027"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Autenticar(ctx, "ana.r", "Otra-Clave-2027"); err != nil {
		t.Errorf("clave nueva: %v", err)
	}
	if creado, _ := s.AsegurarAdministrador(ctx, "x.x", "X", claveP); creado {
		t.Error("con usuarios existentes no debe crear otro administrador")
	}
}

// CP-P03 · RF-15, RNF-11 · Dos administradores que se desactivan a la vez (repetido 25 veces): solo uno lo logra.
func TestUltimoAdministradorSimultaneoContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	s := servicioReal(db)
	a, _ := s.CrearUsuario(ctx, "admin.a", "Admin A", claveP, domain.RolAdministrador)
	b, _ := s.CrearUsuario(ctx, "admin.b", "Admin B", claveP, domain.RolAdministrador)
	off := false
	for ronda := 0; ronda < 25; ronda++ {
		mustExec(t, db, `UPDATE usuario SET activo=TRUE`)
		var ok, rechazados int32
		var wg sync.WaitGroup
		for _, id := range []int{a.ID, b.ID} {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := s.ActualizarUsuario(ctx, id, application.CambiosUsuario{Activo: &off})
				switch {
				case err == nil:
					atomic.AddInt32(&ok, 1)
				case errors.Is(err, domain.ErrUltimoAdministrador):
					atomic.AddInt32(&rechazados, 1)
				default:
					t.Errorf("error inesperado: %v", err)
				}
			}(id)
		}
		wg.Wait()
		if ok != 1 || rechazados != 1 || contar(t, db, `SELECT count(*) FROM usuario WHERE rol='ADMINISTRADOR' AND activo`) != 1 {
			t.Fatalf("ronda %d: éxitos=%d rechazos=%d: debía quedar exactamente 1 administrador activo", ronda, ok, rechazados)
		}
	}
}

// CP-P04 · RF-18, RNF-14 · La auditoría se puede leer pero no modificar ni borrar (RN-06), ni siquiera con SQL directo.
func TestAuditoriaInmutable(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1)
	s := servicioReal(db)
	admin, _ := s.CrearUsuario(ctx, "admin", "Admin", claveP, domain.RolAdministrador)
	conActor := application.ConActor(ctx, application.Actor{ID: admin.ID, Usuario: "admin", Rol: domain.RolAdministrador})
	if _, err := s.RegistrarIngreso(conActor, "AAA000", 1, nil); err != nil {
		t.Fatal(err)
	}
	dev := application.ConActor(ctx, application.Actor{Usuario: "desarrollo", Rol: domain.RolAdministrador}) // ID 0: sin usuario
	if _, err := s.RegistrarSalida(dev, "AAA000", nil); err != nil {
		t.Fatal(err)
	}
	reg, err := s.Auditoria(ctx, 10)
	if err != nil || len(reg) != 2 || reg[0].Accion != "SALIDA" || reg[0].UsuarioID != 0 || reg[0].Usuario != "desarrollo" ||
		reg[1].Accion != "INGRESO" || reg[1].UsuarioID != admin.ID || reg[1].Detalle != "AAA000 en A-01" {
		t.Fatalf("auditoría: %v %+v", err, reg)
	}
	for _, q := range []string{`UPDATE auditoria SET accion='X'`, `DELETE FROM auditoria`, `DELETE FROM auditoria WHERE id_auditoria=1`} {
		_, err := db.Exec(q)
		if err == nil || !strings.Contains(err.Error(), "no se modifica ni se elimina") {
			t.Errorf("%s debía rechazarse por el trigger: %v", q, err)
		}
	}
	if contar(t, db, `SELECT count(*) FROM auditoria`) != 2 {
		t.Error("los registros de auditoría deben seguir intactos")
	}
}

// CP-P05 · RF-09 · Búsqueda: los comodines de LIKE se tratan como texto, no como patrón; carga vehículos y mensualidades.
func TestBuscarClientesContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1)
	mustExec(t, db, `INSERT INTO cliente (documento, nombre, telefono) VALUES ('50%', 'Descuento 100% Real', '300'), ('60_1', 'Con_Guion', NULL), ('70\1', 'Barra\Invertida', NULL)`)
	mustExec(t, db, `INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES (1,'2026-11-01','2026-11-30'), (1,'2026-12-01','2026-12-31')`)
	s := servicioReal(db)
	// "1" coincide por documento con el cliente 1, "60_1" y "70\1", y por nombre con "Descuento 100% Real".
	for q, esperados := range map[string]int{"%": 1, "_": 1, `\`: 1, "100% real": 1, "con_guion": 1, "carlos": 1, "GÓMEZ": 1, "1": 4, "": 5, "no existe": 0} {
		lista, err := s.BuscarClientes(ctx, q)
		if err != nil || len(lista) != esperados {
			t.Errorf("búsqueda %q: esperaba %d, obtuve %d (%v)", q, esperados, len(lista), err)
		}
	}
	lista, _ := s.BuscarClientes(ctx, "carlos")
	c := lista[0]
	if len(c.Vehiculos) != 2 || c.Vehiculos[0].Placa != "AAA000" || c.Vehiculos[1].Placa != "XYZ98A" || len(c.Mensualidades) != 3 || c.Mensualidades[0].FechaInicio != "2026-12-01" || !c.Vigente {
		t.Errorf("detalle de Carlos: vehículos=%+v mensualidades=%+v vigente=%v", c.Vehiculos, c.Mensualidades, c.Vigente)
	}
	sin, _ := s.BuscarClientes(ctx, "con_guion")
	if sin[0].Telefono != "" || len(sin[0].Vehiculos) != 0 || sin[0].Vigente || sin[0].Vehiculos == nil || sin[0].Mensualidades == nil {
		t.Errorf("cliente sin teléfono ni vehículos: %+v", sin[0])
	}
	if _, err := s.Cliente(ctx, 9999); !errors.Is(err, domain.ErrClienteNoExiste) {
		t.Errorf("cliente inexistente: %v", err)
	}
	for i := 0; i < 60; i++ { // el límite de 50 resultados
		mustExec(t, db, `INSERT INTO cliente (documento, nombre) VALUES ($1, $2)`, fmt.Sprintf("L%03d", i), fmt.Sprintf("Masivo %03d", i))
	}
	if lista, _ = s.BuscarClientes(ctx, "masivo"); len(lista) != 50 {
		t.Errorf("el límite es 50, devolvió %d", len(lista))
	}
}

// CP-P06 · RF-11, RNF-11 · Mensualidades simultáneas que se solapan: solo una entra (25 rondas); las contiguas sí caben.
func TestMensualidadSolapadaSimultanea(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	s := servicioReal(db)
	for ronda := 0; ronda < 25; ronda++ {
		mustExec(t, db, `DELETE FROM mensualidad WHERE id_cliente=2`)
		var ok, solapadas int32
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				// Todos los periodos comparten el 15 de noviembre.
				_, err := s.AgregarMensualidad(ctx, 2, application.MensualidadEntrada{FechaInicio: fmt.Sprintf("2026-11-%02d", 1+i), FechaFin: fmt.Sprintf("2026-11-%02d", 15+i)})
				switch {
				case err == nil:
					atomic.AddInt32(&ok, 1)
				case errors.Is(err, domain.ErrMensualidadSolapada):
					atomic.AddInt32(&solapadas, 1)
				default:
					t.Errorf("error inesperado: %v", err)
				}
			}(i)
		}
		wg.Wait()
		if ok != 1 || solapadas != 7 || contar(t, db, `SELECT count(*) FROM mensualidad WHERE id_cliente=2`) != 1 {
			t.Fatalf("ronda %d: éxitos=%d solapadas=%d: debía quedar exactamente 1 mensualidad", ronda, ok, solapadas)
		}
	}
	mustExec(t, db, `DELETE FROM mensualidad WHERE id_cliente=2`)
	for _, p := range [][2]string{{"2026-11-01", "2026-11-30"}, {"2026-12-01", "2026-12-31"}, {"2026-10-01", "2026-10-31"}} {
		if _, err := s.AgregarMensualidad(ctx, 2, application.MensualidadEntrada{FechaInicio: p[0], FechaFin: p[1]}); err != nil {
			t.Errorf("periodo contiguo %v debía aceptarse: %v", p, err)
		}
	}
	if _, err := s.AgregarMensualidad(ctx, 2, application.MensualidadEntrada{FechaInicio: "2026-12-31", FechaFin: "2027-01-05"}); !errors.Is(err, domain.ErrMensualidadSolapada) {
		t.Errorf("compartir un solo día debía solapar: %v", err)
	}
	if _, err := s.AgregarMensualidad(ctx, 999, application.MensualidadEntrada{FechaInicio: "2030-01-01", FechaFin: "2030-01-31"}); !errors.Is(err, domain.ErrClienteNoExiste) {
		t.Errorf("cliente inexistente: %v", err)
	}
}

// CP-P07 · RF-12, RNF-11 · Cambiar el tipo mientras el vehículo ingresa: nunca ocurren ambas cosas (25 rondas).
func TestCambioDeTipoVsIngresoSimultaneo(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1) // AAA000 es CARRO y A-01 es CARRO
	s := servicioReal(db)
	for ronda := 0; ronda < 25; ronda++ {
		mustExec(t, db, `DELETE FROM novedad; DELETE FROM registro; UPDATE celda SET estado='DISPONIBLE'; UPDATE vehiculo SET tipo='CARRO' WHERE placa='AAA000'`)
		var ingreso, cambio error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _, ingreso = s.RegistrarIngreso(ctx, "AAA000", 1, nil) }()
		go func() { defer wg.Done(); _, cambio = s.ActualizarTipoVehiculo(ctx, 1, domain.Moto) }()
		wg.Wait()
		if ingreso == nil && cambio == nil {
			t.Fatalf("ronda %d: el vehículo ingresó en una celda de carros y además cambió a moto", ronda)
		}
		if ingreso != nil && !errors.Is(ingreso, domain.ErrCeldaIncompatible) {
			t.Fatalf("ronda %d: ingreso rechazado por un motivo inesperado: %v", ronda, ingreso)
		}
		if cambio != nil && !errors.Is(cambio, domain.ErrVehiculoDentro) {
			t.Fatalf("ronda %d: cambio rechazado por un motivo inesperado: %v", ronda, cambio)
		}
	}
	if _, err := s.ActualizarTipoVehiculo(ctx, 9999, domain.Carro); !errors.Is(err, domain.ErrVehiculoNoRegistrado) {
		t.Errorf("vehículo inexistente: %v", err)
	}
}

// CP-P08 · RF-19, RNF-11 · Códigos de celda únicos aun con creaciones simultáneas.
func TestCrearCeldaSimultanea(t *testing.T) {
	db := baseDeDatos(t)
	s := servicioReal(db)
	var ok, dup int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.CrearCelda(ctx, "N-01", domain.Carro)
			if err == nil {
				atomic.AddInt32(&ok, 1)
			} else if errors.Is(err, domain.ErrCeldaDuplicada) {
				atomic.AddInt32(&dup, 1)
			} else {
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()
	if ok != 1 || dup != 9 || contar(t, db, `SELECT count(*) FROM celda WHERE codigo='N-01'`) != 1 {
		t.Fatalf("esperaba 1 creación y 9 duplicados, hubo %d y %d", ok, dup)
	}
}

// CP-P09 · RF-20, RF-21 · Editar o inactivar celdas ocupadas se rechaza; tras la salida se permite.
func TestCeldasContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1)
	s := servicioReal(db)
	if _, err := s.RegistrarIngreso(ctx, "AAA000", 1, nil); err != nil {
		t.Fatal(err)
	}
	nuevo := "Z-01"
	if _, err := s.ActualizarCelda(ctx, 1, &nuevo, nil); !errors.Is(err, domain.ErrCeldaOcupada) {
		t.Errorf("editar ocupada: %v", err)
	}
	if _, err := s.CambiarEstadoCelda(ctx, 1, domain.CeldaInactiva); !errors.Is(err, domain.ErrCeldaOcupada) {
		t.Errorf("inactivar ocupada: %v", err)
	}
	if contar(t, db, `SELECT count(*) FROM celda WHERE id_celda=1 AND codigo='A-01' AND estado='OCUPADA'`) != 1 {
		t.Error("los rechazos no debían cambiar la celda")
	}
	if _, err := s.RegistrarSalida(ctx, "AAA000", nil); err != nil {
		t.Fatal(err)
	}
	dup := "A-02"
	if _, err := s.ActualizarCelda(ctx, 1, &dup, nil); !errors.Is(err, domain.ErrCeldaDuplicada) {
		t.Errorf("código de otra celda: %v", err)
	}
	tipo := domain.Moto
	if c, err := s.ActualizarCelda(ctx, 1, &nuevo, &tipo); err != nil || c.Codigo != "Z-01" || c.Tipo != domain.Moto {
		t.Errorf("edición tras la salida: %v %+v", err, c)
	}
	if c, err := s.CambiarEstadoCelda(ctx, 1, domain.CeldaInactiva); err != nil || c.Estado != domain.CeldaInactiva {
		t.Errorf("inactivar: %v %+v", err, c)
	}
	if _, err := s.RegistrarIngreso(ctx, "XYZ98A", 1, nil); !errors.Is(err, domain.ErrCeldaNoDisponible) {
		t.Errorf("ingreso en celda inactiva: %v", err)
	}
	if _, err := s.CambiarEstadoCelda(ctx, 999, domain.CeldaInactiva); !errors.Is(err, domain.ErrCeldaNoExiste) {
		t.Errorf("celda inexistente: %v", err)
	}
	if _, err := s.ActualizarCelda(ctx, 999, &nuevo, nil); !errors.Is(err, domain.ErrCeldaNoExiste) {
		t.Errorf("editar celda inexistente: %v", err)
	}
}

// CP-P10 · RF-10 · Editar clientes: el documento no cambia y el teléfono vacío se guarda como nulo.
func TestEditarClienteContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	s := servicioReal(db)
	c, err := s.ActualizarCliente(ctx, 1, " Carlos  Alberto Gómez ", " 3101112233 ")
	if err != nil || c.Nombre != "Carlos Alberto Gómez" || c.Telefono != "3101112233" || c.Documento != "1" {
		t.Fatalf("edición: %v %+v", err, c)
	}
	if c, err = s.ActualizarCliente(ctx, 1, "Carlos Alberto Gómez", ""); err != nil || c.Telefono != "" {
		t.Fatalf("borrar teléfono: %v %+v", err, c)
	}
	if contar(t, db, `SELECT count(*) FROM cliente WHERE id_cliente=1 AND telefono IS NULL`) != 1 {
		t.Error("el teléfono vacío debía guardarse como NULL")
	}
	if _, err := s.ActualizarCliente(ctx, 999, "X", ""); !errors.Is(err, domain.ErrClienteNoExiste) {
		t.Errorf("cliente inexistente: %v", err)
	}
}

// CP-P11 · RNF-12 · Los datos iniciales entregados (seed.sql) son compatibles con el esquema y repetibles.
func TestSemillaCompatible(t *testing.T) {
	db := baseDeDatos(t)
	semilla, err := os.ReadFile("../../../db/seed.sql")
	if err != nil {
		t.Skip("no hay seed.sql")
	}
	if _, err := db.Exec(string(semilla)); err != nil {
		t.Fatalf("aplicar seed.sql: %v", err)
	}
	if contar(t, db, `SELECT count(*) FROM celda`) == 0 || contar(t, db, `SELECT count(*) FROM vehiculo`) == 0 || contar(t, db, `SELECT count(*) FROM cliente`) == 0 {
		t.Fatal("la semilla debía dejar celdas, clientes y vehículos")
	}
	esquema, _ := os.ReadFile("../../../db/schema.sql")
	if _, err := db.Exec(string(esquema)); err != nil {
		t.Fatalf("reaplicar el esquema sobre datos existentes: %v", err)
	}
	s := servicioReal(db)
	if lista, err := s.BuscarClientes(ctx, ""); err != nil || len(lista) == 0 {
		t.Errorf("búsqueda sobre la semilla: %v %d", err, len(lista))
	}
}

// CP-P12 · RN-03, RNF-11 · La base de datos rechaza un registro entre un vehículo y una celda de distinto tipo (SQL directo).
func TestTriggerTiposCompatibles(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1) // AAA000 CARRO; XYZ98A MOTO; celdas A-01 CARRO, A-02 CARRO, M-01 MOTO
	if _, err := db.Exec(`INSERT INTO registro (id_vehiculo, id_celda, fecha_entrada) VALUES (3, 1, now())`); err == nil || !strings.Contains(err.Error(), "no coinciden") {
		t.Errorf("moto en celda de carros debía rechazarse: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO registro (id_vehiculo, id_celda, fecha_entrada) VALUES (1, 3, now())`); err == nil {
		t.Error("carro en celda de motos debía rechazarse")
	}
	if _, err := db.Exec(`INSERT INTO registro (id_vehiculo, id_celda, fecha_entrada) VALUES (1, 1, now())`); err != nil {
		t.Errorf("carro en celda de carros debía aceptarse: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO registro (id_vehiculo, id_celda, fecha_entrada) VALUES (3, 3, now())`); err != nil {
		t.Errorf("moto en celda de motos debía aceptarse: %v", err)
	}
}

// CP-P13 · RNF-12 · El arranque detecta una base creada con la Iteración 1 (sin las tablas nuevas).
func TestEsquemaActualizado(t *testing.T) {
	db := baseDeDatos(t)
	repo := postgres.Nuevo(db)
	if err := repo.EsquemaActualizado(ctx); err != nil {
		t.Fatalf("con el esquema completo no debía fallar: %v", err)
	}
	mustExec(t, db, `DROP TRIGGER trg_registro_tipos ON registro`)
	if err := repo.EsquemaActualizado(ctx); err == nil || !strings.Contains(err.Error(), "desactualizado") {
		t.Fatalf("sin el trigger debía avisar que el esquema está desactualizado: %v", err)
	}
	esquema, _ := os.ReadFile("../../../db/schema.sql")
	if _, err := db.Exec(string(esquema)); err != nil {
		t.Fatal(err)
	}
	if err := repo.EsquemaActualizado(ctx); err != nil {
		t.Fatalf("tras reaplicar schema.sql debía quedar al día: %v", err)
	}
	mustExec(t, db, `DROP TABLE auditoria`)
	if err := repo.EsquemaActualizado(ctx); err == nil {
		t.Fatal("sin la tabla auditoria debía fallar")
	}
}
