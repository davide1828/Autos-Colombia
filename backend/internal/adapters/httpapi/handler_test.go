package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/adapters/seguridad"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

const secreto = "token-de-prueba-123456" // firma las sesiones

// Alias de sesión: llamar() inicia sesión con el usuario correspondiente y usa el token real.
const (
	token      = "@encargado"
	tokenAdmin = "@admin"
)

const (
	claveEncargado = "Clave-Encargado-1"
	claveAdmin     = "Clave-Admin-2026"
)

var (
	sesiones   sync.Map // http.Handler -> map de alias -> token
	hashRapido = seguridad.NuevoPBKDF2(1000)
)

func nuevoServidor(t *testing.T) http.Handler {
	t.Helper()
	r := memory.Nuevo()
	r.AgregarCliente(1, "Carlos Gómez", "2099-12-31")
	r.AgregarCliente(2, "Marta Rojas", "2000-01-01")
	r.AgregarVehiculo(domain.Vehiculo{ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarVehiculo(domain.Vehiculo{ID: 2, Placa: "DEF456", Tipo: domain.Carro, ClienteID: 2})
	r.AgregarCelda(domain.Celda{ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	r.AgregarCelda(domain.Celda{ID: 2, Codigo: "A-02", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	svc := application.NuevoServicio(r, nil).ConHasheador(hashRapido)
	svc.CrearUsuario(context.Background(), "admin1", "Ana Administradora", claveAdmin, domain.RolAdministrador)
	svc.CrearUsuario(context.Background(), "encargado1", "Luis Encargado", claveEncargado, domain.RolEncargado)
	return httpapi.Nueva(svc, secreto, nil).Rutas()
}

// tokenDe inicia sesión (una vez por servidor y alias) y devuelve el token firmado.
func tokenDe(t *testing.T, h http.Handler, alias string) string {
	t.Helper()
	m, _ := sesiones.LoadOrStore(h, &sync.Map{})
	cache := m.(*sync.Map)
	if tok, ok := cache.Load(alias); ok {
		return tok.(string)
	}
	usuario, clave := "encargado1", claveEncargado
	if alias == tokenAdmin {
		usuario, clave = "admin1", claveAdmin
	}
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"usuario":"`+usuario+`","clave":"`+clave+`"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out struct{ Token string }
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.Token == "" {
		t.Fatalf("no se pudo iniciar sesión como %s: %d %s", usuario, rec.Code, rec.Body.String())
	}
	cache.Store(alias, out.Token)
	return out.Token
}

func llamar(t *testing.T, h http.Handler, metodo, ruta, cuerpo, tok string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	if strings.HasPrefix(tok, "@") {
		tok = tokenDe(t, h, tok)
	}
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	return rec, m
}

func TestAutenticacion(t *testing.T) {
	h := nuevoServidor(t)
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", ""); rec.Code != 401 {
		t.Fatalf("sin token esperaba 401, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", "otro"); rec.Code != 401 {
		t.Fatalf("token incorrecto esperaba 401, obtuve %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", token); rec.Code != 200 {
		t.Fatalf("token correcto esperaba 200, obtuve %d", rec.Code)
	}
}

func TestLimiteDeIntentosFallidos(t *testing.T) {
	h := nuevoServidor(t)
	valido := tokenDe(t, h, token) // se obtiene antes de que la IP quede bloqueada
	for i := 0; i < 10; i++ {
		if rec, _ := llamar(t, h, "GET", "/api/sesion", "", "mala"); rec.Code != 401 {
			t.Fatalf("intento %d: esperaba 401, obtuve %d", i, rec.Code)
		}
	}
	rec, m := llamar(t, h, "GET", "/api/sesion", "", "mala")
	if rec.Code != 429 || m["error"] != "demasiados_intentos" || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("esperaba 429 con Retry-After, obtuve %d %v", rec.Code, m)
	}
	// Incluso la clave correcta queda bloqueada mientras dure la ventana desde la misma IP.
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", valido); rec.Code != 429 {
		t.Fatalf("la IP bloqueada debía recibir 429, obtuve %d", rec.Code)
	}
}

func TestFlujoCompleto(t *testing.T) {
	h := nuevoServidor(t)

	rec, m := llamar(t, h, "POST", "/api/ingresos", `{"placa":"abc-123","id_celda":1,"novedad":{"tipo":"DANIO","descripcion":"Rayón"}}`, token)
	if rec.Code != 201 || m["placa"] != "ABC123" || m["celda"] != "A-01" || m["id_registro"] == nil {
		t.Fatalf("ingreso: %d %v", rec.Code, m)
	}

	rec, m = llamar(t, h, "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":2}`, token)
	if rec.Code != 409 || m["error"] != "conflicto" {
		t.Fatalf("ingreso repetido debía dar 409: %d %v", rec.Code, m)
	}

	rec, _ = llamar(t, h, "GET", "/api/celdas?estado=DISPONIBLE", "", token)
	var celdas []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &celdas)
	if rec.Code != 200 || len(celdas) != 1 || celdas[0]["codigo"] != "A-02" {
		t.Fatalf("celdas disponibles: %d %s", rec.Code, rec.Body.String())
	}

	rec, _ = llamar(t, h, "GET", "/api/registros/activos", "", token)
	var activos []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &activos)
	if len(activos) != 1 || activos[0]["placa"] != "ABC123" || activos[0]["cliente"] != "Carlos Gómez" {
		t.Fatalf("activos: %s", rec.Body.String())
	}

	rec, m = llamar(t, h, "POST", "/api/novedades", `{"placa":"ABC123","tipo":"OTRA","descripcion":"Alarma activada"}`, token)
	if rec.Code != 201 || m["momento"] != "DURANTE_ESTANCIA" {
		t.Fatalf("novedad: %d %v", rec.Code, m)
	}

	rec, m = llamar(t, h, "POST", "/api/salidas", `{"placa":"ABC123"}`, token)
	if rec.Code != 200 || m["celda"] != "A-01" || m["minutos"] == nil || m["fecha_salida"] == nil {
		t.Fatalf("salida: %d %v", rec.Code, m)
	}

	rec, m = llamar(t, h, "POST", "/api/salidas", `{"placa":"ABC123"}`, token)
	if rec.Code != 409 {
		t.Fatalf("segunda salida debía dar 409: %d %v", rec.Code, m)
	}
}

func TestCodigosDeError(t *testing.T) {
	h := nuevoServidor(t)
	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		codigo                       int
		clave                        string
	}{
		{"placa inválida", "POST", "/api/ingresos", `{"placa":"12","id_celda":1}`, 400, "datos_invalidos"},
		{"vehículo no registrado", "POST", "/api/ingresos", `{"placa":"ZZZ999","id_celda":1}`, 404, "no_encontrado"},
		{"celda inexistente", "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":99}`, 404, "no_encontrado"},
		{"mensualidad vencida", "POST", "/api/ingresos", `{"placa":"DEF456","id_celda":1}`, 422, "regla_de_negocio"},
		{"novedad inválida", "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":1,"novedad":{"tipo":"DANIO","descripcion":""}}`, 400, "datos_invalidos"},
		{"JSON roto", "POST", "/api/ingresos", `{"placa":`, 400, "solicitud_invalida"},
		{"campo desconocido", "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":1,"x":1}`, 400, "solicitud_invalida"},
		{"datos extra tras el JSON", "POST", "/api/ingresos", `{"placa":"ABC123","id_celda":1}{"a":1}`, 400, "solicitud_invalida"},
		{"estado de celda inválido", "GET", "/api/celdas?estado=LIBRE", "", 400, "solicitud_invalida"},
		{"ruta inexistente", "GET", "/api/nada", "", 404, "no_encontrado"},
		{"salida sin ingreso", "POST", "/api/salidas", `{"placa":"ABC123"}`, 409, "conflicto"},
		{"novedad sin ingreso", "POST", "/api/novedades", `{"placa":"ABC123","tipo":"OTRA","descripcion":"x"}`, 409, "conflicto"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			rec, m := llamar(t, h, c.metodo, c.ruta, c.cuerpo, token)
			if rec.Code != c.codigo || m["error"] != c.clave {
				t.Fatalf("esperaba %d/%s, obtuve %d %v", c.codigo, c.clave, rec.Code, m)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Fatalf("Content-Type inesperado: %q", ct)
			}
		})
	}
}

func TestMetodoNoPermitido(t *testing.T) {
	rec, m := llamar(t, nuevoServidor(t), "DELETE", "/api/ingresos", "", token)
	if rec.Code != http.StatusMethodNotAllowed || m["error"] != "metodo_no_permitido" || rec.Header().Get("Allow") != "POST" {
		t.Fatalf("esperaba 405 JSON con Allow: %d %v", rec.Code, m)
	}
}

func TestSaludSinAutenticacion(t *testing.T) {
	if rec, _ := llamar(t, nuevoServidor(t), "GET", "/healthz", "", ""); rec.Code != 200 {
		t.Fatalf("healthz debía responder 200, obtuvo %d", rec.Code)
	}
}

// CP-V06 · RF-08 · POST /api/vehiculos: 201, 400, 401 y 409; el vehículo registrado ingresa de inmediato.
func TestRegistrarVehiculo_API(t *testing.T) {
	h := nuevoServidor(t)
	cuerpo := `{"placa":"jkl-456","tipo":"CARRO","cliente":{"documento":"5005005005","nombre":"Ana Pérez","telefono":"3004445566"},"mensualidad":{"fecha_inicio":"2026-01-01","fecha_fin":"2099-12-31"}}`

	if rec, _ := llamar(t, h, "POST", "/api/vehiculos", cuerpo, ""); rec.Code != 401 {
		t.Fatalf("sin token esperaba 401, obtuve %d", rec.Code)
	}
	rec, m := llamar(t, h, "POST", "/api/vehiculos", cuerpo, token)
	if rec.Code != 201 || m["placa"] != "JKL456" || m["tipo"] != "CARRO" || m["cliente"] != "Ana Pérez" || m["cliente_nuevo"] != true || m["mensualidad_registrada"] != true || m["id_vehiculo"] == nil {
		t.Fatalf("registro: %d %v", rec.Code, m)
	}
	// Ya registrado: puede ingresar de inmediato.
	if rec, m := llamar(t, h, "POST", "/api/ingresos", `{"placa":"JKL456","id_celda":1}`, token); rec.Code != 201 {
		t.Fatalf("ingreso del vehículo recién registrado: %d %v", rec.Code, m)
	}
	// Placa repetida → 409.
	if rec, m := llamar(t, h, "POST", "/api/vehiculos", cuerpo, token); rec.Code != 409 || m["error"] != "conflicto" {
		t.Fatalf("placa repetida debía dar 409: %d %v", rec.Code, m)
	}
	invalidos := map[string]string{
		"placa inválida":    `{"placa":"AB12","tipo":"CARRO","cliente":{"documento":"1","nombre":"Ana"}}`,
		"tipo inválido":     `{"placa":"MNO12P","tipo":"BUS","cliente":{"documento":"1","nombre":"Ana"}}`,
		"sin cliente":       `{"placa":"MNO12P","tipo":"MOTO"}`,
		"fechas invertidas": `{"placa":"MNO12P","tipo":"MOTO","cliente":{"documento":"1","nombre":"Ana"},"mensualidad":{"fecha_inicio":"2026-10-31","fecha_fin":"2026-10-01"}}`,
	}
	for nombre, c := range invalidos {
		if rec, m := llamar(t, h, "POST", "/api/vehiculos", c, token); rec.Code != 400 || m["error"] != "datos_invalidos" {
			t.Errorf("%s: esperaba 400, obtuve %d %v", nombre, rec.Code, m)
		}
	}
	if rec, _ := llamar(t, h, "POST", "/api/vehiculos", `{"placa":"MNO12P","tipo":"MOTO","extra":1}`, token); rec.Code != 400 {
		t.Errorf("campos desconocidos debían dar 400, obtuve %d", rec.Code)
	}
}
