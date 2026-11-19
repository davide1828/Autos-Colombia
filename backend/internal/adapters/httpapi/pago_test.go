package httpapi_test

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

func TestPagosHTTPRegistroHistorialReciboYCancelacion(t *testing.T) {
	repo := memory.Nuevo()
	repo.AgregarClienteCompleto(25, "10025", "Nora Ruiz", "", "")
	svc := application.NuevoServicio(repo, time.Now).ConHasheador(hashRapido)
	if _, err := svc.CrearUsuario(context.Background(), "admin.pagos", "Admin Pagos", claveAdmin, domain.RolAdministrador); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CrearUsuario(context.Background(), "operador.pagos", "Operador Pagos", claveEncargado, domain.RolEncargado); err != nil {
		t.Fatal(err)
	}
	h := httpapi.Nueva(svc, secreto, nil).Rutas()
	_, admin := inicioSesion(t, h, "admin.pagos", claveAdmin)
	_, operador := inicioSesion(t, h, "operador.pagos", claveEncargado)

	mes := time.Now().In(application.ZonaColombia).Format("2006-01")
	rec, pago := llamar(t, h, "POST", "/api/pagos",
		fmt.Sprintf(`{"id_cliente":25,"monto":180000,"periodo":%q,"medio_pago":"EFECTIVO"}`, mes),
		admin["token"].(string))
	if rec.Code != http.StatusCreated || pago["id_pago"] == nil {
		t.Fatalf("registrar pago: %d %v", rec.Code, pago)
	}
	id := int(pago["id_pago"].(float64))

	reqHistorial := httptest.NewRequest("GET", "/api/clientes/25/pagos", nil)
	reqHistorial.Header.Set("Authorization", "Bearer "+operador["token"].(string))
	recHistorial := httptest.NewRecorder()
	h.ServeHTTP(recHistorial, reqHistorial)
	var historial []map[string]any
	if err := json.Unmarshal(recHistorial.Body.Bytes(), &historial); err != nil ||
		recHistorial.Code != http.StatusOK || len(historial) != 1 {
		t.Fatalf("historial: %d %s", recHistorial.Code, recHistorial.Body.String())
	}
	recibo := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/pagos/%d/recibo", id), nil)
	req.Header.Set("Authorization", "Bearer "+operador["token"].(string))
	h.ServeHTTP(recibo, req)
	if recibo.Code != http.StatusOK || recibo.Header().Get("Content-Type") != "application/pdf" ||
		recibo.Header().Get("X-Receipt-Number") == "" || len(recibo.Header().Get("X-Receipt-Hash")) != 64 ||
		!strings.HasPrefix(recibo.Body.String(), "%PDF-1.4") {
		t.Fatalf("recibo PDF: status=%d headers=%v", recibo.Code, recibo.Header())
	}

	if rec, _ := llamar(t, h, "POST", fmt.Sprintf("/api/pagos/%d/cancelar", id), `{}`, operador["token"].(string)); rec.Code != http.StatusForbidden {
		t.Fatalf("el encargado no debe cancelar pagos: %d", rec.Code)
	}
	if rec, _ := llamar(t, h, "POST", fmt.Sprintf("/api/pagos/%d/cancelar", id), `{}`, admin["token"].(string)); rec.Code != http.StatusOK {
		t.Fatalf("cancelar pago: %d %s", rec.Code, rec.Body.String())
	}
	if vigente, err := repo.MensualidadVigente(context.Background(), 25, time.Now()); err != nil || vigente {
		t.Fatalf("la cancelación debe revocar la mensualidad: vigente=%v error=%v", vigente, err)
	}
}

func TestPagosHTTPRechazaDatosInvalidosYMetodosNoPermitidos(t *testing.T) {
	h := nuevoServidor(t)
	token := tokenDe(t, h, tokenAdmin)
	for _, cuerpo := range []string{
		`{"id_cliente":1,"monto":0,"periodo":"2099-12","medio_pago":"EFECTIVO"}`,
		`{"id_cliente":1,"monto":1000,"periodo":"2020-01","medio_pago":"EFECTIVO"}`,
		`{"id_cliente":1,"monto":1000,"periodo":"2099-12","medio_pago":"OTRO"}`,
	} {
		if rec, _ := llamar(t, h, "POST", "/api/pagos", cuerpo, token); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuvo %d", cuerpo, rec.Code)
		}
	}
	if rec, _ := llamar(t, h, "DELETE", "/api/pagos", "", token); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/pagos debía dar 405, obtuvo %d", rec.Code)
	}
}
