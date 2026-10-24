// Paquete httpapi: adaptador de presentación (API REST JSON).
package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

type API struct {
	svc   *application.Servicio
	token string // si está vacío, la API no exige autenticación (solo desarrollo)
	log   *slog.Logger
	lim   *limitador
}

func Nueva(svc *application.Servicio, token string, log *slog.Logger) *API {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &API{svc: svc, token: token, log: log, lim: nuevoLimitador(10, time.Minute)}
}

func (a *API) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { escribir(w, 200, map[string]string{"estado": "ok"}) })
	mux.HandleFunc("GET /api/sesion", a.auth(func(w http.ResponseWriter, r *http.Request) { escribir(w, 200, map[string]string{"estado": "ok"}) }))
	mux.HandleFunc("POST /api/ingresos", a.auth(a.ingreso))
	mux.HandleFunc("POST /api/salidas", a.auth(a.salida))
	mux.HandleFunc("POST /api/novedades", a.auth(a.novedad))
	mux.HandleFunc("GET /api/celdas", a.auth(a.celdas))
	mux.HandleFunc("GET /api/registros/activos", a.auth(a.activos))
	// Comodín: distingue "la ruta no existe" (404) de "método no permitido" (405) con respuestas JSON.
	permitidos := map[string]string{
		"/api/ingresos": "POST", "/api/salidas": "POST", "/api/novedades": "POST",
		"/api/celdas": "GET", "/api/registros/activos": "GET", "/api/sesion": "GET",
	}
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if metodo, ok := permitidos[r.URL.Path]; ok {
			w.Header().Set("Allow", metodo)
			escribirError(w, http.StatusMethodNotAllowed, "metodo_no_permitido", "use el método "+metodo)
			return
		}
		escribirError(w, http.StatusNotFound, "no_encontrado", "la ruta no existe")
	})
	return mux
}

func (a *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.token != "" {
			ip := ipDe(r)
			if a.lim.bloqueado(ip) {
				w.Header().Set("Retry-After", "60")
				escribirError(w, http.StatusTooManyRequests, "demasiados_intentos", "demasiados intentos fallidos; espere un minuto")
				return
			}
			recibido := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(recibido), []byte(a.token)) != 1 {
				a.lim.falla(ip)
				a.log.Warn("autenticación fallida", "ip", ip)
				escribirError(w, 401, "no_autorizado", "clave de acceso inválida")
				return
			}
		}
		next(w, r)
	}
}

type ingresoReq struct {
	Placa   string                      `json:"placa"`
	CeldaID int                         `json:"id_celda"`
	Novedad *application.NovedadEntrada `json:"novedad,omitempty"`
}

func (a *API) ingreso(w http.ResponseWriter, r *http.Request) {
	var in ingresoReq
	if !leer(w, r, &in) {
		return
	}
	res, err := a.svc.RegistrarIngreso(r.Context(), in.Placa, in.CeldaID, in.Novedad)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, map[string]any{
		"id_registro": res.Registro.ID, "placa": res.Placa, "celda": res.Celda, "fecha_entrada": res.Registro.Entrada,
	})
}

type salidaReq struct {
	Placa   string                      `json:"placa"`
	Novedad *application.NovedadEntrada `json:"novedad,omitempty"`
}

func (a *API) salida(w http.ResponseWriter, r *http.Request) {
	var in salidaReq
	if !leer(w, r, &in) {
		return
	}
	res, err := a.svc.RegistrarSalida(r.Context(), in.Placa, in.Novedad)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, map[string]any{
		"id_registro": res.Registro.ID, "placa": res.Placa, "celda": res.Celda,
		"fecha_entrada": res.Registro.Entrada, "fecha_salida": res.Registro.Salida, "minutos": res.Minutos,
	})
}

type novedadReq struct {
	Placa       string             `json:"placa"`
	Tipo        domain.TipoNovedad `json:"tipo"`
	Descripcion string             `json:"descripcion"`
}

func (a *API) novedad(w http.ResponseWriter, r *http.Request) {
	var in novedadReq
	if !leer(w, r, &in) {
		return
	}
	n, err := a.svc.RegistrarNovedad(r.Context(), in.Placa, application.NovedadEntrada{Tipo: in.Tipo, Descripcion: in.Descripcion})
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, n)
}

func (a *API) celdas(w http.ResponseWriter, r *http.Request) {
	estado := strings.ToUpper(r.URL.Query().Get("estado"))
	if estado != "" && estado != "DISPONIBLE" && estado != "OCUPADA" && estado != "INACTIVA" {
		escribirError(w, 400, "solicitud_invalida", "estado debe ser DISPONIBLE, OCUPADA o INACTIVA")
		return
	}
	cs, err := a.svc.ConsultarCeldas(r.Context(), estado)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, cs)
}

func (a *API) activos(w http.ResponseWriter, r *http.Request) {
	ds, err := a.svc.VehiculosDentro(r.Context())
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, ds)
}

// --- utilidades ---

func leer(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil || dec.More() {
		escribirError(w, 400, "solicitud_invalida", "el cuerpo no es un JSON válido")
		return false
	}
	return true
}

func escribir(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}

func escribirError(w http.ResponseWriter, codigo int, clave, mensaje string) {
	escribir(w, codigo, map[string]string{"error": clave, "mensaje": mensaje})
}

// responderError traduce errores de negocio a códigos HTTP.
func (a *API) responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrPlacaInvalida), errors.Is(err, domain.ErrNovedadInvalida):
		escribirError(w, 400, "datos_invalidos", err.Error())
	case errors.Is(err, domain.ErrVehiculoNoRegistrado), errors.Is(err, domain.ErrCeldaNoExiste):
		escribirError(w, 404, "no_encontrado", err.Error())
	case errors.Is(err, domain.ErrVehiculoYaDentro), errors.Is(err, domain.ErrCeldaNoDisponible), errors.Is(err, domain.ErrSinIngresoActivo):
		escribirError(w, 409, "conflicto", err.Error())
	case errors.Is(err, domain.ErrMensualidadVencida), errors.Is(err, domain.ErrCeldaIncompatible):
		escribirError(w, 422, "regla_de_negocio", err.Error())
	default:
		a.log.Error("error interno", "error", err)
		escribirError(w, 500, "error_interno", "ocurrió un error inesperado")
	}
}
