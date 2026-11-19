package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

type pagoReq struct {
	ClienteID int              `json:"id_cliente"`
	Monto     int64            `json:"monto"`
	Periodo   string           `json:"periodo"`
	Medio     domain.MedioPago `json:"medio_pago"`
}

const tiempoMaximoPago = 1800 * time.Millisecond

func conTiempoMaximoPago(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelar := context.WithTimeout(r.Context(), tiempoMaximoPago)
		defer cancelar()
		next(w, r.WithContext(ctx))
	}
}

func (a *API) crearPago(w http.ResponseWriter, r *http.Request) {
	var in pagoReq
	if !leer(w, r, &in) {
		return
	}
	pago, err := a.svc.RegistrarPago(r.Context(), in.ClienteID, in.Monto, in.Periodo, in.Medio)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, http.StatusCreated, pago)
}

func (a *API) historialPagos(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	pagos, err := a.svc.HistorialPagos(r.Context(), id)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, http.StatusOK, pagos)
}

func (a *API) reciboPago(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	pago, contenido, err := a.svc.ReciboPago(r.Context(), id)
	if err != nil {
		a.responderError(w, err)
		return
	}
	numero := fmt.Sprintf("AC-%d-%06d", pago.Fecha.In(application.ZonaColombia).Year(), pago.ID)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+numero+`.pdf"`)
	w.Header().Set("X-Receipt-Number", numero)
	w.Header().Set("X-Receipt-Hash", pago.SelloDigital)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(contenido)
}

func (a *API) cancelarPago(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	if err := a.svc.CancelarPago(r.Context(), id); err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, http.StatusOK, map[string]string{"estado": string(domain.PagoCancelado)})
}

func (a *API) registrarRutasPago(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/pagos", conTiempoMaximoPago(a.requiere(ambos, a.crearPago)))
	mux.HandleFunc("GET /api/clientes/{id}/pagos", conTiempoMaximoPago(a.requiere(ambos, a.historialPagos)))
	mux.HandleFunc("GET /api/pagos/{id}/recibo", conTiempoMaximoPago(a.requiere(ambos, a.reciboPago)))
	mux.HandleFunc("POST /api/pagos/{id}/cancelar", conTiempoMaximoPago(a.requiere(soloAdmin, a.cancelarPago)))
}
