package application

// Iteración 2: casos de uso de gestión de usuarios (del sistema y clientes), gestión de celdas y auditoría.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"autoscolombia/parqueadero/internal/domain"
)

// ErrSinHasheador indica que el servicio se creó sin configurar el hasheador de contraseñas.
var ErrSinHasheador = errors.New("el servicio no tiene configurado el hasheador de contraseñas")

// Hasheador es el puerto para proteger contraseñas (la implementación vive en un adaptador).
type Hasheador interface {
	Hash(clave string) (string, error)
	Verificar(hash, clave string) bool
}

// RepositorioGestion es el puerto de persistencia de la Iteración 2.
type RepositorioGestion interface {
	// Usuarios del sistema.
	UsuarioPorLogin(ctx context.Context, login string) (*domain.Usuario, error)
	UsuarioPorID(ctx context.Context, id int) (*domain.Usuario, error)
	Usuarios(ctx context.Context) ([]domain.Usuario, error)
	ContarUsuarios(ctx context.Context) (int, error)
	CrearUsuario(ctx context.Context, u *domain.Usuario) error // ErrUsuarioDuplicado
	// ActualizarUsuario guarda nombre, rol, estado y clave de forma atómica y devuelve ErrUltimoAdministrador
	// si el cambio dejaría al sistema sin administradores activos (aun con operaciones simultáneas).
	ActualizarUsuario(ctx context.Context, u *domain.Usuario) error

	// Clientes y vehículos.
	BuscarClientes(ctx context.Context, q string, limite int) ([]domain.ClienteDetalle, error)
	ClientePorID(ctx context.Context, id int) (*domain.ClienteDetalle, error)
	ActualizarCliente(ctx context.Context, c *domain.Cliente) error // ErrClienteNoExiste
	// AgregarMensualidad rechaza con ErrMensualidadSolapada si el periodo cruza otro del mismo cliente.
	AgregarMensualidad(ctx context.Context, m *domain.Mensualidad) error
	VehiculoPorID(ctx context.Context, id int) (*domain.Vehiculo, error)
	// ActualizarTipoVehiculo falla con ErrVehiculoDentro si el vehículo tiene un ingreso activo.
	ActualizarTipoVehiculo(ctx context.Context, id int, tipo domain.TipoVehiculo) error

	// Celdas.
	CrearCelda(ctx context.Context, c *domain.Celda) error // ErrCeldaDuplicada
	// ActualizarCelda cambia código y tipo; ErrCeldaOcupada si tiene un vehículo dentro.
	ActualizarCelda(ctx context.Context, c *domain.Celda) error
	// CambiarEstadoCelda alterna DISPONIBLE/INACTIVA; ErrCeldaOcupada si tiene un vehículo dentro.
	CambiarEstadoCelda(ctx context.Context, id int, estado domain.EstadoCelda) error

	// Auditoría (solo se agrega y se consulta; nunca se modifica ni se elimina).
	RegistrarAuditoria(ctx context.Context, a *domain.Auditoria) error
	Auditorias(ctx context.Context, limite int) ([]domain.Auditoria, error)
}

// Actor es quien ejecuta una operación (se toma de la sesión).
type Actor struct {
	ID      int
	Usuario string
	Rol     domain.Rol
}

type claveActor struct{}

func ConActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, claveActor{}, a)
}

func ActorDe(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(claveActor{}).(Actor)
	return a, ok
}

// ConHasheador configura la protección de contraseñas (obligatoria para usar usuarios con clave).
func (s *Servicio) ConHasheador(h Hasheador) *Servicio {
	s.hash = h
	if h != nil {
		s.hashFalso, _ = h.Hash("contraseña-inexistente-0")
	}
	return s
}

func (s *Servicio) hoy() time.Time { return s.ahora().In(ZonaColombia) }

// auditar deja constancia de la operación si hay un actor en el contexto (RF-18). Es un registro
// adicional a la operación ya confirmada: si falla no deshace la operación del encargado.
func (s *Servicio) auditar(ctx context.Context, accion, entidad string, id int, detalle string) {
	if a, ok := ActorDe(ctx); ok {
		s.auditarComo(ctx, a, accion, entidad, id, detalle)
	}
}

func (s *Servicio) auditarComo(ctx context.Context, a Actor, accion, entidad string, id int, detalle string) {
	if len(detalle) > 300 {
		detalle = detalle[:300]
	}
	_ = s.repo.RegistrarAuditoria(ctx, &domain.Auditoria{
		Momento: s.ahora(), UsuarioID: a.ID, Usuario: a.Usuario, Accion: accion, Entidad: entidad, EntidadID: id, Detalle: detalle,
	})
}

// ---------- Usuarios del sistema (RF-13 a RF-16) ----------

// Autenticar valida usuario y contraseña. Siempre responde ErrCredenciales ante cualquier fallo
// (usuario inexistente, inactivo o clave errónea) para no revelar qué usuarios existen.
func (s *Servicio) Autenticar(ctx context.Context, login, clave string) (*domain.Usuario, error) {
	if s.hash == nil {
		return nil, ErrSinHasheador
	}
	l, err := domain.NormalizarLogin(login)
	if err != nil {
		s.hash.Verificar(s.hashFalso, clave)
		return nil, domain.ErrCredenciales
	}
	u, err := s.repo.UsuarioPorLogin(ctx, l)
	switch {
	case errors.Is(err, domain.ErrUsuarioNoExiste):
		s.hash.Verificar(s.hashFalso, clave) // mismo costo que un usuario real
		return nil, domain.ErrCredenciales
	case err != nil:
		return nil, err
	}
	if !s.hash.Verificar(u.ClaveHash, clave) || !u.Activo {
		return nil, domain.ErrCredenciales
	}
	s.auditarComo(ctx, Actor{ID: u.ID, Usuario: u.Usuario, Rol: u.Rol}, "INICIO_SESION", "usuario", u.ID, "")
	return u, nil
}

// CrearUsuario — RF-14.
func (s *Servicio) CrearUsuario(ctx context.Context, login, nombre, clave string, rol domain.Rol) (*domain.Usuario, error) {
	if s.hash == nil {
		return nil, ErrSinHasheador
	}
	l, err := domain.NormalizarLogin(login)
	if err != nil {
		return nil, err
	}
	n, err := domain.NormalizarNombre(nombre)
	if err != nil {
		return nil, err
	}
	if !rol.Valido() {
		return nil, domain.ErrRolInvalido
	}
	if err := domain.ValidarClave(clave, l); err != nil {
		return nil, err
	}
	h, err := s.hash.Hash(clave)
	if err != nil {
		return nil, err
	}
	u := &domain.Usuario{Usuario: l, Nombre: n, Rol: rol, Activo: true, ClaveHash: h}
	if err := s.repo.CrearUsuario(ctx, u); err != nil {
		return nil, err
	}
	s.auditar(ctx, "USUARIO_CREADO", "usuario", u.ID, fmt.Sprintf("%s (%s)", u.Usuario, u.Rol))
	return u, nil
}

// Usuarios — RF-15.
func (s *Servicio) Usuarios(ctx context.Context) ([]domain.Usuario, error) {
	return s.repo.Usuarios(ctx)
}

// Usuario devuelve al usuario del sistema por su identificador (la API lo usa para validar cada sesión).
func (s *Servicio) Usuario(ctx context.Context, id int) (*domain.Usuario, error) {
	return s.repo.UsuarioPorID(ctx, id)
}

// CambiosUsuario son los campos editables; los que vienen nulos no se tocan.
type CambiosUsuario struct {
	Nombre *string     `json:"nombre,omitempty"`
	Rol    *domain.Rol `json:"rol,omitempty"`
	Activo *bool       `json:"activo,omitempty"`
}

// ActualizarUsuario — RF-15. Nunca deja al sistema sin un administrador activo.
func (s *Servicio) ActualizarUsuario(ctx context.Context, id int, c CambiosUsuario) (*domain.Usuario, error) {
	u, err := s.repo.UsuarioPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	var cambios []string
	if c.Nombre != nil {
		n, err := domain.NormalizarNombre(*c.Nombre)
		if err != nil {
			return nil, err
		}
		u.Nombre = n
		cambios = append(cambios, "nombre")
	}
	if c.Rol != nil {
		if !c.Rol.Valido() {
			return nil, domain.ErrRolInvalido
		}
		if u.Rol != *c.Rol {
			cambios = append(cambios, fmt.Sprintf("rol %s→%s", u.Rol, *c.Rol))
		}
		u.Rol = *c.Rol
	}
	if c.Activo != nil {
		if u.Activo != *c.Activo {
			cambios = append(cambios, map[bool]string{true: "activado", false: "desactivado"}[*c.Activo])
		}
		u.Activo = *c.Activo
	}
	if err := s.repo.ActualizarUsuario(ctx, u); err != nil {
		return nil, err
	}
	s.auditar(ctx, "USUARIO_EDITADO", "usuario", u.ID, u.Usuario+": "+strings.Join(cambios, ", "))
	return u, nil
}

// RestablecerClave — RF-16 (administrador).
func (s *Servicio) RestablecerClave(ctx context.Context, id int, nueva string) error {
	u, err := s.repo.UsuarioPorID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.fijarClave(ctx, u, nueva); err != nil {
		return err
	}
	s.auditar(ctx, "CLAVE_RESTABLECIDA", "usuario", u.ID, u.Usuario)
	return nil
}

// CambiarMiClave — RF-16 (cualquier usuario sobre su propia cuenta).
func (s *Servicio) CambiarMiClave(ctx context.Context, id int, actual, nueva string) error {
	if s.hash == nil {
		return ErrSinHasheador
	}
	u, err := s.repo.UsuarioPorID(ctx, id)
	if err != nil {
		return err
	}
	if !s.hash.Verificar(u.ClaveHash, actual) {
		return domain.ErrClaveActualIncorrecta
	}
	if err := s.fijarClave(ctx, u, nueva); err != nil {
		return err
	}
	s.auditar(ctx, "CLAVE_CAMBIADA", "usuario", u.ID, u.Usuario)
	return nil
}

func (s *Servicio) fijarClave(ctx context.Context, u *domain.Usuario, nueva string) error {
	if s.hash == nil {
		return ErrSinHasheador
	}
	if err := domain.ValidarClave(nueva, u.Usuario); err != nil {
		return err
	}
	h, err := s.hash.Hash(nueva)
	if err != nil {
		return err
	}
	u.ClaveHash = h
	return s.repo.ActualizarUsuario(ctx, u)
}

// AsegurarAdministrador crea el primer administrador si todavía no existe ningún usuario (arranque).
func (s *Servicio) AsegurarAdministrador(ctx context.Context, login, nombre, clave string) (bool, error) {
	n, err := s.repo.ContarUsuarios(ctx)
	if err != nil || n > 0 {
		return false, err
	}
	if _, err := s.CrearUsuario(ctx, login, nombre, clave, domain.RolAdministrador); err != nil {
		return false, err
	}
	return true, nil
}

// ---------- Clientes y vehículos (RF-09 a RF-12) ----------

const limiteClientes = 50

func (s *Servicio) conVigencia(c *domain.ClienteDetalle) {
	hoy := s.hoy().Format("2006-01-02")
	c.Vigente = false
	for _, m := range c.Mensualidades {
		if hoy >= m.FechaInicio && hoy <= m.FechaFin {
			c.Vigente = true
		}
	}
}

// BuscarClientes — RF-09. Busca por documento o por parte del nombre; sin texto lista los primeros por nombre.
func (s *Servicio) BuscarClientes(ctx context.Context, q string) ([]domain.ClienteDetalle, error) {
	lista, err := s.repo.BuscarClientes(ctx, strings.TrimSpace(q), limiteClientes)
	if lista == nil {
		lista = []domain.ClienteDetalle{} // la API responde [] y no null
	}
	for i := range lista {
		s.conVigencia(&lista[i])
	}
	return lista, err
}

func (s *Servicio) Cliente(ctx context.Context, id int) (*domain.ClienteDetalle, error) {
	c, err := s.repo.ClientePorID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.conVigencia(c)
	return c, nil
}

// ActualizarCliente — RF-10. El documento identifica al cliente y no se modifica.
func (s *Servicio) ActualizarCliente(ctx context.Context, id int, nombre, telefono string) (*domain.ClienteDetalle, error) {
	actual, err := s.repo.ClientePorID(ctx, id)
	if err != nil {
		return nil, err
	}
	c, err := domain.NormalizarCliente(domain.Cliente{ID: id, Documento: actual.Documento, Nombre: nombre, Telefono: telefono})
	if err != nil {
		return nil, err
	}
	if err := s.repo.ActualizarCliente(ctx, &c); err != nil {
		return nil, err
	}
	s.auditar(ctx, "CLIENTE_EDITADO", "cliente", id, c.Documento)
	return s.Cliente(ctx, id)
}

// AgregarMensualidad — RF-11.
func (s *Servicio) AgregarMensualidad(ctx context.Context, clienteID int, in MensualidadEntrada) (*domain.Mensualidad, error) {
	m, err := domain.NuevaMensualidad(in.FechaInicio, in.FechaFin)
	if err != nil {
		return nil, err
	}
	m.ClienteID = clienteID
	if err := s.repo.AgregarMensualidad(ctx, m); err != nil {
		return nil, err
	}
	s.auditar(ctx, "MENSUALIDAD_REGISTRADA", "cliente", clienteID, in.FechaInicio+" a "+in.FechaFin)
	return m, nil
}

// ActualizarTipoVehiculo — RF-12.
func (s *Servicio) ActualizarTipoVehiculo(ctx context.Context, id int, tipo domain.TipoVehiculo) (*domain.Vehiculo, error) {
	if !tipo.Valido() {
		return nil, domain.ErrTipoVehiculoInvalido
	}
	if err := s.repo.ActualizarTipoVehiculo(ctx, id, tipo); err != nil {
		return nil, err
	}
	v, err := s.repo.VehiculoPorID(ctx, id)
	if err == nil {
		s.auditar(ctx, "VEHICULO_EDITADO", "vehiculo", id, v.Placa+" → "+string(tipo))
	}
	return v, err
}

// ---------- Celdas (RF-19 a RF-21) ----------

// CrearCelda — RF-19. Toda celda nueva nace DISPONIBLE.
func (s *Servicio) CrearCelda(ctx context.Context, codigo string, tipo domain.TipoVehiculo) (*domain.Celda, error) {
	cod, err := domain.NormalizarCodigoCelda(codigo)
	if err != nil {
		return nil, err
	}
	if !tipo.Valido() {
		return nil, domain.ErrCeldaInvalida
	}
	c := &domain.Celda{Codigo: cod, Tipo: tipo, Estado: domain.CeldaDisponible}
	if err := s.repo.CrearCelda(ctx, c); err != nil {
		return nil, err
	}
	s.auditar(ctx, "CELDA_CREADA", "celda", c.ID, fmt.Sprintf("%s (%s)", c.Codigo, c.Tipo))
	return c, nil
}

// ActualizarCelda — RF-20. Solo si la celda no tiene un vehículo dentro.
func (s *Servicio) ActualizarCelda(ctx context.Context, id int, codigo *string, tipo *domain.TipoVehiculo) (*domain.Celda, error) {
	c, err := s.repo.Celda(ctx, id)
	if err != nil {
		return nil, err
	}
	if codigo != nil {
		if c.Codigo, err = domain.NormalizarCodigoCelda(*codigo); err != nil {
			return nil, err
		}
	}
	if tipo != nil {
		if !tipo.Valido() {
			return nil, domain.ErrCeldaInvalida
		}
		c.Tipo = *tipo
	}
	if err := s.repo.ActualizarCelda(ctx, c); err != nil {
		return nil, err
	}
	s.auditar(ctx, "CELDA_EDITADA", "celda", c.ID, fmt.Sprintf("%s (%s)", c.Codigo, c.Tipo))
	return s.repo.Celda(ctx, id)
}

// CambiarEstadoCelda — RF-21. Inactivar o reactivar (nunca se elimina: RN-06).
func (s *Servicio) CambiarEstadoCelda(ctx context.Context, id int, estado domain.EstadoCelda) (*domain.Celda, error) {
	if !estado.EstadoGestionable() {
		return nil, domain.ErrEstadoCeldaInvalido
	}
	if err := s.repo.CambiarEstadoCelda(ctx, id, estado); err != nil {
		return nil, err
	}
	c, err := s.repo.Celda(ctx, id)
	if err == nil {
		s.auditar(ctx, "CELDA_ESTADO", "celda", id, c.Codigo+" → "+string(estado))
	}
	return c, err
}

// Auditoria — RF-18. Devuelve los últimos movimientos (más recientes primero).
func (s *Servicio) Auditoria(ctx context.Context, limite int) ([]domain.Auditoria, error) {
	if limite < 1 || limite > 200 {
		limite = 50
	}
	return s.repo.Auditorias(ctx, limite)
}
