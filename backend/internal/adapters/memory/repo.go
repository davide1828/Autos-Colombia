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
	docs      map[string]int // documento -> id cliente
	sigCli    int
	sigVeh    int
	mens      map[int][]domain.Mensualidad // id cliente -> periodos
	documento map[int]string               // id cliente -> documento
	telefono  map[int]string               // id cliente -> teléfono
	usuarios  map[int]*domain.Usuario
	sigUsu    int
	sigMens   int
	auditoria []domain.Auditoria
	celdas    map[int]*domain.Celda
	registros []*domain.Registro
	novedades []domain.Novedad
	sigReg    int
	sigNov    int
}

func Nuevo() *Repo {
	return &Repo{
		vehiculos: map[string]*domain.Vehiculo{}, clientes: map[int]string{}, docs: map[string]int{}, sigCli: 1000, sigVeh: 1000,
		mens: map[int][]domain.Mensualidad{}, documento: map[int]string{}, telefono: map[int]string{},
		usuarios: map[int]*domain.Usuario{}, celdas: map[int]*domain.Celda{},
	}
}

// AgregarCliente registra un cliente de prueba; mensualidadHasta (AAAA-MM-DD) crea una mensualidad vigente hasta esa fecha.
func (r *Repo) AgregarCliente(id int, nombre, mensualidadHasta string) {
	r.AgregarClienteCompleto(id, "DOC"+nombre, nombre, "", mensualidadHasta)
}

// AgregarClienteCompleto permite fijar documento y teléfono en las pruebas.
func (r *Repo) AgregarClienteCompleto(id int, documento, nombre, telefono, mensualidadHasta string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clientes[id], r.documento[id], r.telefono[id], r.docs[documento] = nombre, documento, telefono, id
	if mensualidadHasta != "" {
		fin, _ := time.Parse("2006-01-02", mensualidadHasta)
		r.mens[id] = append(r.mens[id], domain.Mensualidad{ClienteID: id, Inicio: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC), Fin: fin})
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
	for _, m := range r.mens[clienteID] {
		if m.EstaVigente(dia) {
			return true, nil
		}
	}
	return false, nil
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

// GuardarVehiculo aplica las mismas reglas que la base de datos: placa única y cliente único por documento.
func (r *Repo) GuardarVehiculo(_ context.Context, c *domain.Cliente, v *domain.Vehiculo, m *domain.Mensualidad) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.vehiculos[v.Placa]; ok {
		return false, domain.ErrPlacaYaRegistrada
	}
	nuevo := false
	id, ok := r.docs[c.Documento]
	if !ok {
		r.sigCli++
		id, nuevo = r.sigCli, true
		r.docs[c.Documento] = id
		r.clientes[id], r.documento[id], r.telefono[id] = c.Nombre, c.Documento, c.Telefono
	}
	c.ID, c.Nombre = id, r.clientes[id]
	r.sigVeh++
	v.ID, v.ClienteID = r.sigVeh, id
	cp := *v
	r.vehiculos[v.Placa] = &cp
	if m != nil {
		m.ClienteID = id
		r.sigMens++
		m.ID = r.sigMens
		r.mens[id] = append(r.mens[id], *m)
	}
	return nuevo, nil
}
