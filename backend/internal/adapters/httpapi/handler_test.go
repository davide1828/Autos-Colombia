package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

const token = "token-de-prueba-123456"

func nuevoServidor(t *testing.T) http.Handler {
	t.Helper()
	r := memory.Nuevo()
	r.AgregarCliente(1, "Carlos Gómez", "2099-12-31")
	r.AgregarCliente(2, "Marta Rojas", "2000-01-01")
	r.AgregarVehiculo(domain.Vehiculo{ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarVehiculo(domain.Vehiculo{ID: 2, Placa: "DEF456", Tipo: domain.Carro, ClienteID: 2})
	r.AgregarCelda(domain.Celda{ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	r.AgregarCelda(domain.Celda{ID: 2, Codigo: "A-02", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	return httpapi.Nueva(application.NuevoServicio(r, nil), token, nil).Rutas()
}

func llamar(t *testing.T, h http.Handler, metodo, ruta, cuerpo, tok string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
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
	if rec, _ := llamar(t, h, "GET", "/api/sesion", "", token); rec.Code != 429 {
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
