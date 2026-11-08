package httpapi_test

// Pruebas HTTP de la Iteración 2 (sesiones, permisos por rol, usuarios, clientes, celdas y auditoría).

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

// servidorConServicio devuelve también el servicio para preparar situaciones (usuarios desactivados, etc.).
func servidorConServicio(t *testing.T, reloj func() time.Time) (http.Handler, *application.Servicio) {
	t.Helper()
	r := memory.Nuevo()
	r.AgregarClienteCompleto(1, "1001", "Carlos Gómez", "3001234567", "2099-12-31")
	r.AgregarClienteCompleto(2, "1002", "Marta Rojas", "", "2000-01-01")
	r.AgregarVehiculo(domain.Vehiculo{ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarVehiculo(domain.Vehiculo{ID: 2, Placa: "DEF456", Tipo: domain.Carro, ClienteID: 2})
	r.AgregarCelda(domain.Celda{ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	r.AgregarCelda(domain.Celda{ID: 2, Codigo: "A-02", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	svc := application.NuevoServicio(r, reloj).ConHasheador(hashRapido)
	svc.CrearUsuario(context.Background(), "admin1", "Ana Administradora", claveAdmin, domain.RolAdministrador)
	svc.CrearUsuario(context.Background(), "encargado1", "Luis Encargado", claveEncargado, domain.RolEncargado)
	return httpapi.Nueva(svc, secreto, nil).ConReloj(reloj).Rutas(), svc
}

func inicioSesion(t *testing.T, h http.Handler, usuario, clave string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	return llamar(t, h, "POST", "/api/login", fmt.Sprintf(`{"usuario":%q,"clave":%q}`, usuario, clave), "")
}

// CP-H01 · RF-13 · Inicio de sesión: devuelve token, vencimiento y datos del usuario sin su hash.
func TestLogin(t *testing.T) {
	h := nuevoServidor(t)
	rec, m := inicioSesion(t, h, " ADMIN1 ", claveAdmin)
	u, _ := m["usuario"].(map[string]any)
	if rec.Code != 200 || m["token"] == "" || m["expira"] == nil || u["usuario"] != "admin1" || u["rol"] != "ADMINISTRADOR" {
		t.Fatalf("login: %d %v", rec.Code, m)
	}
	if strings.Contains(rec.Body.String(), "pbkdf2") || strings.Contains(rec.Body.String(), "clave_hash") || strings.Contains(rec.Body.String(), claveAdmin) {
		t.Fatalf("la respuesta no debe exponer hash ni contraseña: %s", rec.Body.String())
	}
	if rec, m = llamar(t, h, "GET", "/api/sesion", "", m["token"].(string)); rec.Code != 200 || m["estado"] != "ok" {
		t.Fatalf("la sesión debía ser válida: %d %v", rec.Code, m)
	}
}

// CP-H02 · RF-13 · Fallos de login: mismo mensaje para cualquier causa; formato inválido.
func TestLogin_Rechazos(t *testing.T) {
	h := nuevoServidor(t)
	mensajes := map[string]bool{}
	for _, c := range [][2]string{{"admin1", "Clave-Mala-2026"}, {"fantasma", claveAdmin}, {"", ""}, {"a b", claveAdmin}} {
		rec, m := inicioSesion(t, h, c[0], c[1])
		if rec.Code != 401 || m["error"] != "no_autorizado" {
			t.Errorf("%v: esperaba 401, obtuve %d %v", c, rec.Code, m)
		}
		mensajes[fmt.Sprint(m["mensaje"])] = true
	}
	if len(mensajes) != 1 {
		t.Errorf("todos los fallos debían dar el mismo mensaje: %v", mensajes)
	}
	if rec, _ := llamar(t, h, "POST", "/api/login", `{"usuario":"admin1","clave":"x","extra":1}`, ""); rec.Code != 400 {
		t.Errorf("campos desconocidos debían dar 400, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/login", `no es json`, ""); rec.Code != 400 {
		t.Errorf("JSON inválido debía dar 400, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/login", "", ""); rec.Code != 405 {
		t.Errorf("GET /api/login debía dar 405, obtuve %d", rec.Code)
	}
}

// CP-H14 · RF-13 · Diez intentos fallidos en un minuto bloquean el acceso (el undécimo recibe 429, incluso con la clave correcta).
func TestLogin_BloqueaTrasDiezFallos(t *testing.T) {
	h := nuevoServidor(t)
	for i := 0; i < 10; i++ {
		if rec, _ := inicioSesion(t, h, "admin1", "Clave-Mala-2026"); rec.Code != 401 {
			t.Fatalf("intento %d: %d", i, rec.Code)
		}
	}
	rec, m := inicioSesion(t, h, "admin1", claveAdmin) // incluso la clave correcta queda bloqueada
	if rec.Code != 429 || m["error"] != "demasiados_intentos" {
		t.Fatalf("esperaba 429, obtuve %d %v", rec.Code, m)
	}
}

// CP-H03 · RNF-09 · La sesión vence a las 8 horas (la última hora válida y la primera inválida).
func TestSesion_Vencimiento(t *testing.T) {
	ahora := time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)
	h, _ := servidorConServicio(t, func() time.Time { return ahora })
	_, m := inicioSesion(t, h, "encargado1", claveEncargado)
	tok := m["token"].(string)
	ahora = ahora.Add(httpapi.DuracionSesion - time.Second)
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", tok); rec.Code != 200 {
		t.Fatalf("un segundo antes de vencer debía servir: %d", rec.Code)
	}
	ahora = ahora.Add(time.Second)
	if rec, m := llamar(t, h, "GET", "/api/sesion", "", tok); rec.Code != 401 || m["error"] != "no_autorizado" {
		t.Fatalf("a las 8 horas debía vencer: %d %v", rec.Code, m)
	}
}

// CP-H04 · RNF-04 · Tokens manipulados, vacíos o sin firma se rechazan.
func TestSesion_TokensInvalidos(t *testing.T) {
	h := nuevoServidor(t)
	_, m := inicioSesion(t, h, "admin1", claveAdmin)
	tok := m["token"].(string)
	partes := strings.Split(tok, ".")
	cambiar := func(s string) string { // altera un carácter
		if s[0] == 'A' {
			return "B" + s[1:]
		}
		return "A" + s[1:]
	}
	for nombre, malo := range map[string]string{
		"firma alterada": partes[0] + "." + cambiar(partes[1]), "contenido alterado": cambiar(partes[0]) + "." + partes[1],
		"sin firma": partes[0], "solo punto": ".", "vacío": "", "basura": "xxx.yyy", "tres partes": tok + ".z",
	} {
		if rec, _ := llamar(t, nuevoServidor(t), "GET", "/api/sesion", "", malo); rec.Code != 401 {
			t.Errorf("%s: esperaba 401, obtuve %d", nombre, rec.Code)
		}
	}
}

// CP-H05 · RF-15, RF-17 · Desactivar o degradar a un usuario surte efecto de inmediato, aunque su token siga vigente.
func TestSesion_CambiosSurtenEfectoDeInmediato(t *testing.T) {
	h, svc := servidorConServicio(t, time.Now)
	tokAdmin := tokenDe(t, h, tokenAdmin)
	_, m := inicioSesion(t, h, "encargado1", claveEncargado)
	tokEnc := m["token"].(string)
	us, _ := svc.Usuarios(context.Background())
	var idEnc int
	for _, u := range us {
		if u.Usuario == "encargado1" {
			idEnc = u.ID
		}
	}
	if rec, _ := llamar(t, h, "GET", "/api/usuarios", "", tokEnc); rec.Code != 403 {
		t.Fatalf("el encargado no debía listar usuarios: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "PATCH", fmt.Sprintf("/api/usuarios/%d", idEnc), `{"rol":"ADMINISTRADOR"}`, tokAdmin); rec.Code != 200 {
		t.Fatalf("ascenso: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/usuarios", "", tokEnc); rec.Code != 200 {
		t.Fatalf("tras el ascenso debía poder listar usuarios: %d", rec.Code)
	}
	// Desactivado: el token vigente deja de servir de inmediato.
	if rec, _ := llamar(t, h, "PATCH", fmt.Sprintf("/api/usuarios/%d", idEnc), `{"activo":false}`, tokAdmin); rec.Code != 200 {
		t.Fatalf("desactivación: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", tokEnc); rec.Code != 401 {
		t.Fatalf("el usuario desactivado debía recibir 401, obtuve %d", rec.Code)
	}
	if rec, _ := inicioSesion(t, h, "encargado1", claveEncargado); rec.Code != 401 {
		t.Fatalf("el usuario desactivado no puede iniciar sesión: %d", rec.Code)
	}
}

// CP-H06 · RF-17 · Matriz de permisos por rol: ENCARGADO recibe 403 en las rutas de administración.
func TestPermisosPorRol(t *testing.T) {
	h := nuevoServidor(t)
	administracion := []struct{ metodo, ruta, cuerpo string }{
		{"POST", "/api/celdas", `{"codigo":"Z-01","tipo":"CARRO"}`},
		{"PATCH", "/api/celdas/1", `{"codigo":"Z-02"}`},
		{"PUT", "/api/celdas/1/estado", `{"estado":"INACTIVA"}`},
		{"GET", "/api/usuarios", ``},
		{"POST", "/api/usuarios", `{"usuario":"nuevo1","nombre":"Nuevo","clave":"Clave-Nueva-2026","rol":"ENCARGADO"}`},
		{"PATCH", "/api/usuarios/1", `{"nombre":"X"}`},
		{"POST", "/api/usuarios/1/clave", `{"nueva":"Clave-Nueva-2026"}`},
		{"GET", "/api/auditoria", ``},
	}
	for _, c := range administracion {
		if rec, m := llamar(t, h, c.metodo, c.ruta, c.cuerpo, token); rec.Code != 403 || m["error"] != "sin_permiso" {
			t.Errorf("encargado %s %s: esperaba 403, obtuve %d %v", c.metodo, c.ruta, rec.Code, m)
		}
		if rec, _ := llamar(t, h, c.metodo, c.ruta, c.cuerpo, ""); rec.Code != 401 {
			t.Errorf("sin sesión %s %s: esperaba 401, obtuve %d", c.metodo, c.ruta, rec.Code)
		}
		if rec, _ := llamar(t, h, c.metodo, c.ruta, c.cuerpo, tokenAdmin); rec.Code >= 400 {
			t.Errorf("administrador %s %s: esperaba éxito, obtuve %d", c.metodo, c.ruta, rec.Code)
		}
	}
	// Las rutas de operación diaria y de clientes sí son de ambos roles.
	for _, ruta := range []string{"/api/celdas", "/api/registros/activos", "/api/clientes", "/api/clientes/1", "/api/sesion"} {
		for _, tk := range []string{token, tokenAdmin} {
			if rec, _ := llamar(t, h, "GET", ruta, "", tk); rec.Code != 200 {
				t.Errorf("%s con %s: esperaba 200, obtuve %d", ruta, tk, rec.Code)
			}
		}
	}
}

// CP-H07 · RF-16 · Cambiar la propia contraseña por la API.
func TestCambiarMiClave_API(t *testing.T) {
	h := nuevoServidor(t)
	if rec, _ := llamar(t, h, "POST", "/api/sesion/clave", `{"actual":"incorrecta","nueva":"Nueva-Clave-2027"}`, token); rec.Code != 400 {
		t.Errorf("actual incorrecta debía dar 400 (no 401: la sesión sigue siendo válida), obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", token); rec.Code != 200 {
		t.Errorf("tras equivocarse de contraseña actual la sesión debía seguir válida, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/sesion/clave", `{"actual":"`+claveEncargado+`","nueva":"corta"}`, token); rec.Code != 400 {
		t.Errorf("nueva débil debía dar 400, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/sesion/clave", `{"actual":"`+claveEncargado+`","nueva":"Nueva-Clave-2027"}`, token); rec.Code != 200 {
		t.Fatalf("cambio válido debía dar 200, obtuve %d", rec.Code)
	}
	if rec, _ := inicioSesion(t, h, "encargado1", claveEncargado); rec.Code != 401 {
		t.Error("la clave anterior debía dejar de servir")
	}
	if rec, _ := inicioSesion(t, h, "encargado1", "Nueva-Clave-2027"); rec.Code != 200 {
		t.Error("la clave nueva debía servir")
	}
}

// CP-H08 · RF-14, RF-15, RF-16 · Gestión de usuarios por la API.
func TestUsuarios_API(t *testing.T) {
	h := nuevoServidor(t)
	rec, m := llamar(t, h, "POST", "/api/usuarios", `{"usuario":"Maria.R","nombre":" María  Ríos ","clave":"Clave-Maria-2026","rol":"ENCARGADO"}`, tokenAdmin)
	if rec.Code != 201 || m["usuario"] != "maria.r" || m["nombre"] != "María Ríos" || m["activo"] != true || m["id_usuario"] == nil {
		t.Fatalf("creación: %d %v", rec.Code, m)
	}
	if strings.Contains(rec.Body.String(), "pbkdf2") {
		t.Fatal("no debe exponer el hash")
	}
	id := fmt.Sprintf("%.0f", m["id_usuario"])
	if rec, m = llamar(t, h, "POST", "/api/usuarios", `{"usuario":"maria.r","nombre":"Otra","clave":"Clave-Maria-2026","rol":"ENCARGADO"}`, tokenAdmin); rec.Code != 409 {
		t.Errorf("duplicado debía dar 409, obtuve %d %v", rec.Code, m)
	}
	for nombre, cuerpo := range map[string]string{
		"clave débil":      `{"usuario":"u.debil","nombre":"N","clave":"123","rol":"ENCARGADO"}`,
		"rol inválido":     `{"usuario":"u.rol","nombre":"N","clave":"Clave-Maria-2026","rol":"ROOT"}`,
		"usuario inválido": `{"usuario":"a b","nombre":"N","clave":"Clave-Maria-2026","rol":"ENCARGADO"}`,
		"campo extra":      `{"usuario":"u.extra","nombre":"N","clave":"Clave-Maria-2026","rol":"ENCARGADO","activo":false}`,
	} {
		if rec, _ := llamar(t, h, "POST", "/api/usuarios", cuerpo, tokenAdmin); rec.Code != 400 {
			t.Errorf("%s: esperaba 400, obtuve %d", nombre, rec.Code)
		}
	}
	if rec, m = llamar(t, h, "PATCH", "/api/usuarios/"+id, `{"nombre":"María Ríos Gómez","activo":false}`, tokenAdmin); rec.Code != 200 || m["nombre"] != "María Ríos Gómez" || m["activo"] != false {
		t.Errorf("edición: %d %v", rec.Code, m)
	}
	if rec, _ := llamar(t, h, "POST", "/api/usuarios/"+id+"/clave", `{"nueva":"Clave-Restablecida-1"}`, tokenAdmin); rec.Code != 200 {
		t.Errorf("restablecer: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/usuarios/"+id+"/clave", `{"nueva":"x"}`, tokenAdmin); rec.Code != 400 {
		t.Errorf("restablecer con clave débil debía dar 400, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "PATCH", "/api/usuarios/9999", `{"nombre":"X"}`, tokenAdmin); rec.Code != 404 {
		t.Errorf("usuario inexistente debía dar 404, obtuve %d", rec.Code)
	}
	// El único administrador no puede desactivarse a sí mismo.
	if rec, m = llamar(t, h, "PATCH", "/api/usuarios/1", `{"activo":false}`, tokenAdmin); rec.Code != 409 {
		t.Errorf("último administrador debía dar 409, obtuve %d %v", rec.Code, m)
	}
	req := httptest.NewRequest("GET", "/api/usuarios", nil)
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, h, tokenAdmin))
	out := httptest.NewRecorder()
	h.ServeHTTP(out, req)
	var lista []map[string]any
	if json.Unmarshal(out.Body.Bytes(), &lista) != nil || len(lista) != 3 || strings.Contains(out.Body.String(), "pbkdf2") {
		t.Errorf("listado: %s", out.Body.String())
	}
}

// CP-H09 · RF-09, RF-10, RF-11, RF-12 · Clientes, mensualidades y vehículos por la API.
func TestClientes_API(t *testing.T) {
	h, _ := servidorConServicio(t, time.Now)
	req := func(ruta string) []map[string]any {
		r := httptest.NewRequest("GET", ruta, nil)
		r.Header.Set("Authorization", "Bearer "+tokenDe(t, h, token))
		out := httptest.NewRecorder()
		h.ServeHTTP(out, r)
		var l []map[string]any
		if err := json.Unmarshal(out.Body.Bytes(), &l); err != nil {
			t.Fatalf("%s: %v %s", ruta, err, out.Body.String())
		}
		return l
	}
	if l := req("/api/clientes?q=carlos"); len(l) != 1 || l[0]["documento"] != "1001" || l[0]["mensualidad_vigente"] != true || len(l[0]["vehiculos"].([]any)) != 1 {
		t.Errorf("búsqueda: %+v", l)
	}
	if l := req("/api/clientes?q=zzz"); l == nil || len(l) != 0 {
		t.Errorf("sin resultados debía devolver []: %+v", l)
	}
	rec, m := llamar(t, h, "GET", "/api/clientes/2", "", token)
	if rec.Code != 200 || m["mensualidad_vigente"] != false || m["nombre"] != "Marta Rojas" {
		t.Errorf("detalle: %d %v", rec.Code, m)
	}
	if rec, _ := llamar(t, h, "GET", "/api/clientes/999", "", token); rec.Code != 404 {
		t.Errorf("cliente inexistente: %d", rec.Code)
	}
	for _, id := range []string{"abc", "0", "-1", "1.5", "99999999999999999999"} {
		if rec, _ := llamar(t, h, "GET", "/api/clientes/"+id, "", token); rec.Code != 400 && rec.Code != 404 && rec.Code != 405 {
			t.Errorf("id %q: esperaba 4xx, obtuve %d", id, rec.Code)
		}
	}
	if rec, m = llamar(t, h, "PATCH", "/api/clientes/2", `{"nombre":" Marta  Lucía Rojas ","telefono":"3159998877"}`, token); rec.Code != 200 || m["nombre"] != "Marta Lucía Rojas" || m["documento"] != "1002" {
		t.Errorf("edición: %d %v", rec.Code, m)
	}
	if rec, _ := llamar(t, h, "PATCH", "/api/clientes/2", `{"nombre":"  ","telefono":""}`, token); rec.Code != 400 {
		t.Errorf("nombre vacío: %d", rec.Code)
	}
	// Mensualidad: registrar, solapar, renovar y poder ingresar.
	if rec, _ := llamar(t, h, "POST", "/api/ingresos", `{"placa":"DEF456","id_celda":1}`, token); rec.Code != 422 {
		t.Fatalf("antes de renovar debía dar 422: %d", rec.Code)
	}
	rec, m = llamar(t, h, "POST", "/api/clientes/2/mensualidades", `{"fecha_inicio":"2026-01-01","fecha_fin":"2099-12-31"}`, token)
	if rec.Code != 201 || m["fecha_inicio"] != "2026-01-01" || m["id_mensualidad"] == nil {
		t.Fatalf("mensualidad: %d %v", rec.Code, m)
	}
	if rec, _ = llamar(t, h, "POST", "/api/clientes/2/mensualidades", `{"fecha_inicio":"2030-01-01","fecha_fin":"2030-02-01"}`, token); rec.Code != 409 {
		t.Errorf("solapada: esperaba 409, obtuve %d", rec.Code)
	}
	if rec, _ = llamar(t, h, "POST", "/api/clientes/2/mensualidades", `{"fecha_inicio":"2030-02-01","fecha_fin":"2030-01-01"}`, token); rec.Code != 409 && rec.Code != 400 {
		t.Errorf("invertida: %d", rec.Code)
	}
	if rec, _ = llamar(t, h, "POST", "/api/clientes/999/mensualidades", `{"fecha_inicio":"2030-01-01","fecha_fin":"2030-02-01"}`, token); rec.Code != 404 {
		t.Errorf("cliente inexistente: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/ingresos", `{"placa":"DEF456","id_celda":1}`, token); rec.Code != 201 {
		t.Errorf("tras renovar debía ingresar: %d", rec.Code)
	}
	// Tipo de vehículo.
	if rec, _ := llamar(t, h, "PATCH", "/api/vehiculos/2", `{"tipo":"MOTO"}`, token); rec.Code != 409 {
		t.Errorf("vehículo dentro: esperaba 409, obtuve %d", rec.Code)
	}
	if rec, m = llamar(t, h, "PATCH", "/api/vehiculos/1", `{"tipo":"MOTO"}`, token); rec.Code != 200 || m["tipo"] != "MOTO" {
		t.Errorf("cambio de tipo: %d %v", rec.Code, m)
	}
	if rec, _ := llamar(t, h, "PATCH", "/api/vehiculos/1", `{"tipo":"BUS"}`, token); rec.Code != 400 {
		t.Errorf("tipo inválido: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "PATCH", "/api/vehiculos/999", `{"tipo":"CARRO"}`, token); rec.Code != 404 {
		t.Errorf("vehículo inexistente: %d", rec.Code)
	}
}

// CP-H10 · RF-19, RF-20, RF-21 · Celdas por la API (solo administrador).
func TestCeldas_API(t *testing.T) {
	h := nuevoServidor(t)
	rec, m := llamar(t, h, "POST", "/api/celdas", `{"codigo":" b-01 ","tipo":"MOTO"}`, tokenAdmin)
	if rec.Code != 201 || m["codigo"] != "B-01" || m["estado"] != "DISPONIBLE" || m["id_celda"] == nil {
		t.Fatalf("creación: %d %v", rec.Code, m)
	}
	id := fmt.Sprintf("%.0f", m["id_celda"])
	for nombre, c := range map[string]struct {
		cuerpo string
		code   int
	}{
		"código repetido": {`{"codigo":"A-01","tipo":"CARRO"}`, 409}, "código inválido": {`{"codigo":"A 01","tipo":"CARRO"}`, 400},
		"tipo inválido": {`{"codigo":"C-01","tipo":"BUS"}`, 400}, "vacío": {`{}`, 400}, "campo extra": {`{"codigo":"C-02","tipo":"CARRO","estado":"OCUPADA"}`, 400},
	} {
		if rec, _ := llamar(t, h, "POST", "/api/celdas", c.cuerpo, tokenAdmin); rec.Code != c.code {
			t.Errorf("%s: esperaba %d, obtuve %d", nombre, c.code, rec.Code)
		}
	}
	if rec, m = llamar(t, h, "PATCH", "/api/celdas/"+id, `{"codigo":"b-02","tipo":"CARRO"}`, tokenAdmin); rec.Code != 200 || m["codigo"] != "B-02" || m["tipo"] != "CARRO" {
		t.Errorf("edición: %d %v", rec.Code, m)
	}
	if rec, _ = llamar(t, h, "PATCH", "/api/celdas/"+id, `{"codigo":"A-01"}`, tokenAdmin); rec.Code != 409 {
		t.Errorf("código de otra celda: %d", rec.Code)
	}
	if rec, m = llamar(t, h, "PUT", "/api/celdas/"+id+"/estado", `{"estado":"INACTIVA"}`, tokenAdmin); rec.Code != 200 || m["estado"] != "INACTIVA" {
		t.Errorf("inactivar: %d %v", rec.Code, m)
	}
	if rec, _ = llamar(t, h, "PUT", "/api/celdas/"+id+"/estado", `{"estado":"OCUPADA"}`, tokenAdmin); rec.Code != 400 {
		t.Errorf("fijar OCUPADA: %d", rec.Code)
	}
	if rec, _ = llamar(t, h, "PUT", "/api/celdas/999/estado", `{"estado":"INACTIVA"}`, tokenAdmin); rec.Code != 404 {
		t.Errorf("celda inexistente: %d", rec.Code)
	}
	// Celda ocupada: ni editar ni inactivar.
	if rec, _ = llamar(t, h, "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":1}`, token); rec.Code != 201 {
		t.Fatalf("ingreso: %d", rec.Code)
	}
	if rec, _ = llamar(t, h, "PATCH", "/api/celdas/1", `{"codigo":"Z-09"}`, tokenAdmin); rec.Code != 409 {
		t.Errorf("editar ocupada: %d", rec.Code)
	}
	if rec, _ = llamar(t, h, "PUT", "/api/celdas/1/estado", `{"estado":"INACTIVA"}`, tokenAdmin); rec.Code != 409 {
		t.Errorf("inactivar ocupada: %d", rec.Code)
	}
}

// CP-H11 · RF-18 · La auditoría muestra quién hizo cada operación.
func TestAuditoria_API(t *testing.T) {
	h := nuevoServidor(t)
	llamar(t, h, "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":1}`, token)
	llamar(t, h, "POST", "/api/celdas", `{"codigo":"Q-01","tipo":"CARRO"}`, tokenAdmin)
	r := httptest.NewRequest("GET", "/api/auditoria?limite=10", nil)
	r.Header.Set("Authorization", "Bearer "+tokenDe(t, h, tokenAdmin))
	out := httptest.NewRecorder()
	h.ServeHTTP(out, r)
	var l []map[string]any
	if err := json.Unmarshal(out.Body.Bytes(), &l); err != nil || len(l) != 4 {
		t.Fatalf("auditoría (2 inicios de sesión + 2 operaciones): %v %s", err, out.Body.String())
	}
	por := map[string]map[string]any{}
	for _, x := range l {
		por[x["accion"].(string)] = x
	}
	if por["CELDA_CREADA"]["usuario"] != "admin1" || por["INGRESO"]["usuario"] != "encargado1" || por["INGRESO"]["detalle"] != "ABC123 en A-01" || por["INICIO_SESION"] == nil {
		t.Errorf("registros inesperados: %+v", l)
	}
	if l[0]["accion"] != "CELDA_CREADA" {
		t.Errorf("el más reciente debía ir primero: %v", l[0]["accion"])
	}
}

// CP-H12 · RNF-04 · Rutas desconocidas y métodos no permitidos responden en JSON.
func TestRutasNuevas_404y405(t *testing.T) {
	h := nuevoServidor(t)
	for _, c := range []struct {
		metodo, ruta string
		code         int
	}{
		{"DELETE", "/api/celdas/1", 405}, {"DELETE", "/api/usuarios/1", 405}, {"DELETE", "/api/clientes/1", 405}, {"PUT", "/api/usuarios", 405},
		{"GET", "/api/celdas/1/estado", 405}, {"GET", "/api/clientes/1/mensualidades", 405}, {"POST", "/api/auditoria", 405},
		{"GET", "/api/clientes/1/otra", 404}, {"GET", "/api/inexistente", 404}, {"GET", "/api/usuarios/abc/clave", 404},
	} {
		rec, m := llamar(t, h, c.metodo, c.ruta, "", tokenAdmin)
		if rec.Code != c.code || m["error"] == nil {
			t.Errorf("%s %s: esperaba %d JSON, obtuve %d %v", c.metodo, c.ruta, c.code, rec.Code, m)
		}
	}
}

// CP-H13 · RNF-04 · Modo de desarrollo sin autenticación: todo se permite como administrador y no hay login.
func TestModoDesarrolloSinAutenticacion(t *testing.T) {
	svc := application.NuevoServicio(memory.Nuevo(), nil).ConHasheador(hashRapido)
	h := httpapi.Nueva(svc, "", nil).Rutas()
	if rec, m := llamar(t, h, "GET", "/api/sesion", "", ""); rec.Code != 200 {
		t.Errorf("sesión: %d %v", rec.Code, m)
	} else if u := m["usuario"].(map[string]any); u["rol"] != "ADMINISTRADOR" || u["usuario"] != "desarrollo" {
		t.Errorf("actor de desarrollo: %v", u)
	}
	if rec, _ := llamar(t, h, "POST", "/api/celdas", `{"codigo":"D-01","tipo":"CARRO"}`, ""); rec.Code != 201 {
		t.Errorf("crear celda sin sesión en desarrollo: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", "/api/login", `{"usuario":"a","clave":"b"}`, ""); rec.Code != 400 {
		t.Errorf("login en desarrollo debía dar 400, obtuve %d", rec.Code)
	}
}
