package memory

// Iteración 2: gestión de usuarios, clientes, celdas y auditoría en memoria (misma semántica que PostgreSQL).

import (
	"context"
	"sort"
	"strings"

	"autoscolombia/parqueadero/internal/domain"
)

func copiaUsuario(u *domain.Usuario) *domain.Usuario { c := *u; return &c }

func (r *Repo) UsuarioPorLogin(_ context.Context, login string) (*domain.Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.usuarios {
		if u.Usuario == login {
			return copiaUsuario(u), nil
		}
	}
	return nil, domain.ErrUsuarioNoExiste
}

func (r *Repo) UsuarioPorID(_ context.Context, id int) (*domain.Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.usuarios[id]; ok {
		return copiaUsuario(u), nil
	}
	return nil, domain.ErrUsuarioNoExiste
}

func (r *Repo) Usuarios(_ context.Context) ([]domain.Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Usuario, 0, len(r.usuarios))
	for _, u := range r.usuarios {
		out = append(out, *u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Usuario < out[j].Usuario })
	return out, nil
}

func (r *Repo) ContarUsuarios(_ context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.usuarios), nil
}

func (r *Repo) CrearUsuario(_ context.Context, u *domain.Usuario) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.usuarios {
		if o.Usuario == u.Usuario {
			return domain.ErrUsuarioDuplicado
		}
	}
	r.sigUsu++
	u.ID = r.sigUsu
	r.usuarios[u.ID] = copiaUsuario(u)
	return nil
}

func (r *Repo) ActualizarUsuario(_ context.Context, u *domain.Usuario) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.usuarios[u.ID]; !ok {
		return domain.ErrUsuarioNoExiste
	}
	previo := r.usuarios[u.ID]
	r.usuarios[u.ID] = copiaUsuario(u)
	for _, o := range r.usuarios {
		if o.Rol == domain.RolAdministrador && o.Activo {
			return nil
		}
	}
	r.usuarios[u.ID] = previo // dejaría al sistema sin administradores activos
	return domain.ErrUltimoAdministrador
}

// ---- clientes y vehículos ----

func (r *Repo) detalle(id int) domain.ClienteDetalle {
	d := domain.ClienteDetalle{Cliente: domain.Cliente{ID: id, Documento: r.documento[id], Nombre: r.clientes[id], Telefono: r.telefono[id]},
		Vehiculos: []domain.Vehiculo{}, Mensualidades: []domain.PeriodoMensualidad{}}
	for _, v := range r.vehiculos {
		if v.ClienteID == id {
			d.Vehiculos = append(d.Vehiculos, *v)
		}
	}
	sort.Slice(d.Vehiculos, func(i, j int) bool { return d.Vehiculos[i].Placa < d.Vehiculos[j].Placa })
	for _, m := range r.mens[id] {
		d.Mensualidades = append(d.Mensualidades, domain.PeriodoMensualidad{ID: m.ID, FechaInicio: m.Inicio.Format("2006-01-02"), FechaFin: m.Fin.Format("2006-01-02")})
	}
	sort.Slice(d.Mensualidades, func(i, j int) bool { return d.Mensualidades[i].FechaInicio > d.Mensualidades[j].FechaInicio })
	return d
}

func (r *Repo) BuscarClientes(_ context.Context, q string, limite int) ([]domain.ClienteDetalle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q = strings.ToLower(q)
	out := []domain.ClienteDetalle{}
	for id := range r.clientes {
		if q == "" || strings.Contains(strings.ToLower(r.clientes[id]), q) || strings.Contains(strings.ToLower(r.documento[id]), q) {
			out = append(out, r.detalle(id))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nombre < out[j].Nombre })
	if len(out) > limite {
		out = out[:limite]
	}
	return out, nil
}

func (r *Repo) ClientePorID(_ context.Context, id int) (*domain.ClienteDetalle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clientes[id]; !ok {
		return nil, domain.ErrClienteNoExiste
	}
	d := r.detalle(id)
	return &d, nil
}

func (r *Repo) ActualizarCliente(_ context.Context, c *domain.Cliente) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clientes[c.ID]; !ok {
		return domain.ErrClienteNoExiste
	}
	r.clientes[c.ID], r.telefono[c.ID] = c.Nombre, c.Telefono
	return nil
}

func (r *Repo) AgregarMensualidad(_ context.Context, m *domain.Mensualidad) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clientes[m.ClienteID]; !ok {
		return domain.ErrClienteNoExiste
	}
	for _, o := range r.mens[m.ClienteID] {
		if !o.Inicio.After(m.Fin) && !o.Fin.Before(m.Inicio) {
			return domain.ErrMensualidadSolapada
		}
	}
	r.sigMens++
	m.ID = r.sigMens
	r.mens[m.ClienteID] = append(r.mens[m.ClienteID], *m)
	return nil
}

func (r *Repo) buscarVehiculo(id int) *domain.Vehiculo {
	for _, v := range r.vehiculos {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func (r *Repo) VehiculoPorID(_ context.Context, id int) (*domain.Vehiculo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v := r.buscarVehiculo(id); v != nil {
		c := *v
		return &c, nil
	}
	return nil, domain.ErrVehiculoNoRegistrado
}

func (r *Repo) ActualizarTipoVehiculo(_ context.Context, id int, tipo domain.TipoVehiculo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.buscarVehiculo(id)
	if v == nil {
		return domain.ErrVehiculoNoRegistrado
	}
	for _, g := range r.registros {
		if g.VehiculoID == id && g.EstaActivo() {
			return domain.ErrVehiculoDentro
		}
	}
	v.Tipo = tipo
	return nil
}

// ---- celdas ----

func (r *Repo) CrearCelda(_ context.Context, c *domain.Celda) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	max := 0
	for id, o := range r.celdas {
		if o.Codigo == c.Codigo {
			return domain.ErrCeldaDuplicada
		}
		if id > max {
			max = id
		}
	}
	c.ID = max + 1
	cp := *c
	r.celdas[c.ID] = &cp
	return nil
}

func (r *Repo) ActualizarCelda(_ context.Context, c *domain.Celda) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	actual, ok := r.celdas[c.ID]
	if !ok {
		return domain.ErrCeldaNoExiste
	}
	if actual.Estado == domain.CeldaOcupada {
		return domain.ErrCeldaOcupada
	}
	for id, o := range r.celdas {
		if id != c.ID && o.Codigo == c.Codigo {
			return domain.ErrCeldaDuplicada
		}
	}
	actual.Codigo, actual.Tipo = c.Codigo, c.Tipo
	return nil
}

func (r *Repo) CambiarEstadoCelda(_ context.Context, id int, estado domain.EstadoCelda) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.celdas[id]
	if !ok {
		return domain.ErrCeldaNoExiste
	}
	if c.Estado == domain.CeldaOcupada {
		return domain.ErrCeldaOcupada
	}
	c.Estado = estado
	return nil
}

// ---- auditoría ----

func (r *Repo) RegistrarAuditoria(_ context.Context, a *domain.Auditoria) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a.ID = int64(len(r.auditoria) + 1)
	r.auditoria = append(r.auditoria, *a)
	return nil
}

func (r *Repo) Auditorias(_ context.Context, limite int) ([]domain.Auditoria, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Auditoria, 0, limite)
	for i := len(r.auditoria) - 1; i >= 0 && len(out) < limite; i-- {
		out = append(out, r.auditoria[i])
	}
	return out, nil
}
