package memory

import (
	"context"
	"sort"

	"autoscolombia/parqueadero/internal/domain"
)

func copiaPago(p *domain.Pago) *domain.Pago {
	copia := *p
	return &copia
}

func (r *Repo) CrearPago(_ context.Context, p *domain.Pago) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clientes[p.ClienteID]; !ok {
		return domain.ErrClienteNoExiste
	}
	for _, pago := range r.pagos {
		if pago.ClienteID == p.ClienteID && pago.PeriodoInicio.Equal(p.PeriodoInicio) && pago.Estado == domain.PagoPagado {
			return domain.ErrPagoDuplicado
		}
	}
	for _, mensualidad := range r.mens[p.ClienteID] {
		if !mensualidad.Inicio.After(p.PeriodoFin) && !mensualidad.Fin.Before(p.PeriodoInicio) {
			return domain.ErrMensualidadSolapada
		}
	}
	r.sigMens++
	m := domain.Mensualidad{ID: r.sigMens, ClienteID: p.ClienteID, Inicio: p.PeriodoInicio, Fin: p.PeriodoFin}
	r.mens[p.ClienteID] = append(r.mens[p.ClienteID], m)
	r.sigPago++
	p.ID = r.sigPago
	r.pagoMens[p.ID] = m.ID
	r.pagos[p.ID] = copiaPago(p)
	return nil
}

func (r *Repo) HistorialPagos(_ context.Context, clienteID int) ([]domain.Pago, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clientes[clienteID]; !ok {
		return nil, domain.ErrClienteNoExiste
	}
	pagos := []domain.Pago{}
	for _, pago := range r.pagos {
		if pago.ClienteID == clienteID {
			p := *pago
			p.ClienteNombre = r.clientes[clienteID]
			p.Documento = r.documento[clienteID]
			pagos = append(pagos, p)
		}
	}
	sort.Slice(pagos, func(i, j int) bool {
		if pagos[i].PeriodoInicio.Equal(pagos[j].PeriodoInicio) {
			return pagos[i].ID > pagos[j].ID
		}
		return pagos[i].PeriodoInicio.After(pagos[j].PeriodoInicio)
	})
	return pagos, nil
}

func (r *Repo) ObtenerPago(_ context.Context, id int) (*domain.Pago, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pago, ok := r.pagos[id]
	if !ok {
		return nil, domain.ErrPagoNoExiste
	}
	copia := copiaPago(pago)
	copia.ClienteNombre = r.clientes[pago.ClienteID]
	copia.Documento = r.documento[pago.ClienteID]
	return copia, nil
}

func (r *Repo) GuardarSelloRecibo(_ context.Context, id int, estado domain.EstadoPago, sello string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	pago, ok := r.pagos[id]
	if !ok {
		return domain.ErrPagoNoExiste
	}
	if pago.Estado != estado {
		return domain.ErrPagoCambio
	}
	pago.SelloDigital = sello
	return nil
}

func (r *Repo) CancelarPago(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	pago, ok := r.pagos[id]
	if !ok {
		return domain.ErrPagoNoExiste
	}
	if pago.Estado == domain.PagoCancelado {
		return domain.ErrPagoYaCancelado
	}
	mensualidadID := r.pagoMens[id]
	mensualidades := r.mens[pago.ClienteID][:0]
	for _, mensualidad := range r.mens[pago.ClienteID] {
		if mensualidad.ID != mensualidadID {
			mensualidades = append(mensualidades, mensualidad)
		}
	}
	r.mens[pago.ClienteID] = mensualidades
	pago.Estado = domain.PagoCancelado
	delete(r.pagoMens, id)
	return nil
}
