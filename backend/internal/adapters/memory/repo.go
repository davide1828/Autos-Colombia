// Paquete memory: implementación en memoria de application.Repositorio.
// Aplica las mismas reglas de unicidad que la base de datos; se usa en pruebas.
package memory

import (
	"context"
	"sync"
	"time"

	"autoscolombia/parqueadero/internal/domain"
)

type Repo struct {
	mu        sync.Mutex
	vehiculos map[string]*domain.Vehiculo
	clientes  map[int]string // id -> nombre
	hasta     map[int]string // id cliente -> último día de la mensualidad (yyyy-mm-dd)
	celdas    map[int]*domain.Celda
	registros []*domain.Registro
	novedades []domain.Novedad
	sigReg    int
	sigNov    int
}

func Nuevo() *Repo {
	return &Repo{
		vehiculos: map[string]*domain.Vehiculo{}, clientes: map[int]string{},
		hasta: map[int]string{}, celdas: map[int]*domain.Celda{},
	}
}

func (r *Repo) AgregarCliente(id int, nombre, mensualidadHasta string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clientes[id] = nombre
	if mensualidadHasta != "" {
		r.hasta[id] = mensualidadHasta
	}
}

func (r *Repo) AgregarVehiculo(v domain.Vehiculo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := v
	r.vehiculos[v.Placa] = &cp
}

func (r *Repo) AgregarCelda(c domain.Celda) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := c
	r.celdas[c.ID] = &cp
}

func (r *Repo) Novedades() []domain.Novedad {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]domain.Novedad(nil), r.novedades...)
}

func (r *Repo) EstadoCelda(id int) domain.EstadoCelda {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.celdas[id].Estado
}

func (r *Repo) VehiculoPorPlaca(_ context.Context, placa string) (*domain.Vehiculo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.vehiculos[placa]; ok {
		cp := *v
		return &cp, nil
	}
	return nil, domain.ErrVehiculoNoRegistrado
}

func (r *Repo) MensualidadVigente(_ context.Context, clienteID int, dia time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fin, ok := r.hasta[clienteID]
	return ok && dia.Format("2006-01-02") <= fin, nil
}

func (r *Repo) Celda(_ context.Context, id int) (*domain.Celda, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.celdas[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, domain.ErrCeldaNoExiste
}

func (r *Repo) RegistroActivoDeVehiculo(_ context.Context, vehiculoID int) (*domain.Registro, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.registros {
		if g.VehiculoID == vehiculoID && g.EstaActivo() {
			cp := *g
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *Repo) GuardarIngreso(_ context.Context, g *domain.Registro, n *domain.Novedad) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := r.celdas[g.CeldaID]
	if c == nil || c.Estado != domain.CeldaDisponible {
		return domain.ErrCeldaNoDisponible
	}
	for _, o := range r.registros {
		if o.EstaActivo() && o.VehiculoID == g.VehiculoID {
			return domain.ErrVehiculoYaDentro
		}
	}
	c.Estado = domain.CeldaOcupada
	r.sigReg++
	g.ID = r.sigReg
	cp := *g
	r.registros = append(r.registros, &cp)
	if n != nil {
		r.sigNov++
		n.ID, n.RegistroID = r.sigNov, g.ID
		r.novedades = append(r.novedades, *n)
	}
	return nil
}

func (r *Repo) GuardarSalida(_ context.Context, g *domain.Registro, n *domain.Novedad) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.registros {
		if o.ID == g.ID && o.EstaActivo() {
			o.Salida = g.Salida
			r.celdas[o.CeldaID].Estado = domain.CeldaDisponible
			if n != nil {
				r.sigNov++
				n.ID, n.RegistroID = r.sigNov, g.ID
				r.novedades = append(r.novedades, *n)
			}
			return nil
		}
	}
	return domain.ErrSinIngresoActivo
}

func (r *Repo) GuardarNovedad(_ context.Context, n *domain.Novedad) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sigNov++
	n.ID = r.sigNov
	r.novedades = append(r.novedades, *n)
	return nil
}

func (r *Repo) Celdas(_ context.Context, estado string) ([]domain.Celda, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.Celda{}
	for id := 1; id <= len(r.celdas)+100; id++ { // orden estable por id
		if c, ok := r.celdas[id]; ok && (estado == "" || string(c.Estado) == estado) {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (r *Repo) RegistrosActivos(_ context.Context) ([]domain.RegistroDetalle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.RegistroDetalle{}
	for _, g := range r.registros {
		if !g.EstaActivo() {
			continue
		}
		var placa string
		var cliente int
		for p, v := range r.vehiculos {
			if v.ID == g.VehiculoID {
				placa, cliente = p, v.ClienteID
			}
		}
		out = append(out, domain.RegistroDetalle{RegistroID: g.ID, Placa: placa, Cliente: r.clientes[cliente], Celda: r.celdas[g.CeldaID].Codigo, Entrada: g.Entrada})
	}
	return out, nil
}
