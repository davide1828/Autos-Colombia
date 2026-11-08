// Paquete application: casos de uso. Orquesta el dominio a través de puertos.
package application

import (
	"context"
	"errors"
	"time"

	"autoscolombia/parqueadero/internal/domain"
)

// Repositorio es el puerto de salida hacia la persistencia.
type Repositorio interface {
	VehiculoPorPlaca(ctx context.Context, placa string) (*domain.Vehiculo, error)
	MensualidadVigente(ctx context.Context, clienteID int, dia time.Time) (bool, error)
	Celda(ctx context.Context, id int) (*domain.Celda, error)
	RegistroActivoDeVehiculo(ctx context.Context, vehiculoID int) (*domain.Registro, error)
	// GuardarIngreso inserta el registro, ocupa la celda y guarda la novedad opcional en una sola transacción.
	GuardarIngreso(ctx context.Context, r *domain.Registro, n *domain.Novedad) error
	// GuardarSalida cierra el registro, libera la celda y guarda la novedad opcional en una sola transacción.
	GuardarSalida(ctx context.Context, r *domain.Registro, n *domain.Novedad) error
	GuardarNovedad(ctx context.Context, n *domain.Novedad) error
	// GuardarVehiculo registra el vehículo en una sola transacción: reutiliza el cliente si ya existe su documento
	// (conserva el nombre guardado y lo deja en c) o lo crea, inserta el vehículo y, si viene, la mensualidad.
	// Devuelve ErrPlacaYaRegistrada si la placa ya existe (aun con operaciones simultáneas).
	GuardarVehiculo(ctx context.Context, c *domain.Cliente, v *domain.Vehiculo, m *domain.Mensualidad) (clienteNuevo bool, err error)
	Celdas(ctx context.Context, estado string) ([]domain.Celda, error)
	RegistrosActivos(ctx context.Context) ([]domain.RegistroDetalle, error)

	RepositorioGestion // Iteración 2: usuarios, clientes, celdas y auditoría
}

type NovedadEntrada struct {
	Tipo        domain.TipoNovedad `json:"tipo"`
	Descripcion string             `json:"descripcion"`
}

// ResultadoIngreso es lo que el encargado necesita ver tras registrar un ingreso.
type ResultadoIngreso struct {
	Registro *domain.Registro
	Placa    string
	Celda    string // código de la celda (ej. A-01)
}

// ClienteEntrada son los datos del cliente al registrar un vehículo.
type ClienteEntrada struct {
	Documento string `json:"documento"`
	Nombre    string `json:"nombre"`
	Telefono  string `json:"telefono,omitempty"`
}

// MensualidadEntrada es el periodo de la mensualidad (fechas AAAA-MM-DD).
type MensualidadEntrada struct {
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
}

// ResultadoVehiculo es lo que el encargado necesita ver tras registrar un vehículo.
type ResultadoVehiculo struct {
	Vehiculo              *domain.Vehiculo
	Cliente               string // nombre del cliente asociado
	ClienteNuevo          bool   // false si se reutilizó un cliente existente (mismo documento)
	MensualidadRegistrada bool
}

// ResultadoSalida incluye el tiempo de permanencia.
type ResultadoSalida struct {
	Registro *domain.Registro
	Placa    string
	Celda    string
	Minutos  int
}

// ZonaColombia: Colombia no tiene horario de verano, UTC-5 todo el año.
var ZonaColombia = time.FixedZone("COT", -5*60*60)

type Servicio struct {
	repo      Repositorio
	ahora     func() time.Time
	hash      Hasheador
	hashFalso string // hash de un usuario inexistente: iguala el tiempo de respuesta del login
}

func NuevoServicio(repo Repositorio, ahora func() time.Time) *Servicio {
	if ahora == nil {
		ahora = time.Now
	}
	return &Servicio{repo: repo, ahora: ahora}
}

func (s *Servicio) novedad(in *NovedadEntrada, m domain.MomentoNovedad, vehiculoID int) (*domain.Novedad, error) {
	if in == nil {
		return nil, nil
	}
	n := &domain.Novedad{VehiculoID: vehiculoID, Tipo: in.Tipo, Descripcion: in.Descripcion, Momento: m, Fecha: s.ahora()}
	if err := n.Validar(); err != nil {
		return nil, err
	}
	return n, nil
}

// RegistrarIngreso — RF-01, RF-02, RF-05.
func (s *Servicio) RegistrarIngreso(ctx context.Context, placa string, celdaID int, nov *NovedadEntrada) (*ResultadoIngreso, error) {
	placa, err := domain.NormalizarPlaca(placa)
	if err != nil {
		return nil, err
	}
	v, err := s.repo.VehiculoPorPlaca(ctx, placa)
	if err != nil {
		return nil, err
	}
	ahora := s.ahora()
	// La vigencia de la mensualidad se evalúa con el día calendario de Colombia, no con el de UTC.
	ok, err := s.repo.MensualidadVigente(ctx, v.ClienteID, ahora.In(ZonaColombia))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrMensualidadVencida
	}
	activo, err := s.repo.RegistroActivoDeVehiculo(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if activo != nil {
		return nil, domain.ErrVehiculoYaDentro
	}
	c, err := s.repo.Celda(ctx, celdaID)
	if err != nil {
		return nil, err
	}
	if !c.PuedeOcuparse() {
		return nil, domain.ErrCeldaNoDisponible
	}
	if c.Tipo != v.Tipo {
		return nil, domain.ErrCeldaIncompatible
	}
	n, err := s.novedad(nov, domain.MomentoEntrada, v.ID)
	if err != nil {
		return nil, err
	}
	r := &domain.Registro{VehiculoID: v.ID, CeldaID: c.ID, Entrada: ahora}
	if err := s.repo.GuardarIngreso(ctx, r, n); err != nil {
		return nil, err
	}
	s.auditar(ctx, "INGRESO", "registro", r.ID, v.Placa+" en "+c.Codigo)
	return &ResultadoIngreso{Registro: r, Placa: v.Placa, Celda: c.Codigo}, nil
}

// RegistrarSalida — RF-03.
func (s *Servicio) RegistrarSalida(ctx context.Context, placa string, nov *NovedadEntrada) (*ResultadoSalida, error) {
	placa, err := domain.NormalizarPlaca(placa)
	if err != nil {
		return nil, err
	}
	v, err := s.repo.VehiculoPorPlaca(ctx, placa)
	if err != nil {
		return nil, err
	}
	r, err := s.repo.RegistroActivoDeVehiculo(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, domain.ErrSinIngresoActivo
	}
	n, err := s.novedad(nov, domain.MomentoSalida, v.ID)
	if err != nil {
		return nil, err
	}
	celda, err := s.repo.Celda(ctx, r.CeldaID)
	if err != nil {
		return nil, err
	}
	ahora := s.ahora()
	r.Salida = &ahora
	if err := s.repo.GuardarSalida(ctx, r, n); err != nil {
		return nil, err
	}
	s.auditar(ctx, "SALIDA", "registro", r.ID, v.Placa+" de "+celda.Codigo)
	return &ResultadoSalida{Registro: r, Placa: v.Placa, Celda: celda.Codigo, Minutos: int(r.Duracion(ahora).Minutes())}, nil
}

// RegistrarNovedad — RF-04: novedad sobre un vehículo que está dentro.
func (s *Servicio) RegistrarNovedad(ctx context.Context, placa string, in NovedadEntrada) (*domain.Novedad, error) {
	placa, err := domain.NormalizarPlaca(placa)
	if err != nil {
		return nil, err
	}
	v, err := s.repo.VehiculoPorPlaca(ctx, placa)
	if err != nil {
		return nil, err
	}
	r, err := s.repo.RegistroActivoDeVehiculo(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, domain.ErrSinIngresoActivo
	}
	n, err := s.novedad(&in, domain.MomentoEstancia, v.ID)
	if err != nil {
		return nil, err
	}
	n.RegistroID = r.ID
	if err := s.repo.GuardarNovedad(ctx, n); err != nil {
		return nil, err
	}
	s.auditar(ctx, "NOVEDAD", "registro", r.ID, v.Placa+" · "+string(n.Tipo))
	return n, nil
}

// ConsultarCeldas — RF-05. estado vacío = todas.
func (s *Servicio) ConsultarCeldas(ctx context.Context, estado string) ([]domain.Celda, error) {
	return s.repo.Celdas(ctx, estado)
}

func (s *Servicio) VehiculosDentro(ctx context.Context) ([]domain.RegistroDetalle, error) {
	return s.repo.RegistrosActivos(ctx)
}

// RegistrarVehiculo — RF-08. Registra un vehículo con su cliente y, opcionalmente, una mensualidad.
func (s *Servicio) RegistrarVehiculo(ctx context.Context, placa string, tipo domain.TipoVehiculo, cli ClienteEntrada, men *MensualidadEntrada) (*ResultadoVehiculo, error) {
	placa, err := domain.NormalizarPlaca(placa)
	if err != nil {
		return nil, err
	}
	if !tipo.Valido() {
		return nil, domain.ErrTipoVehiculoInvalido
	}
	c, err := domain.NormalizarCliente(domain.Cliente{Documento: cli.Documento, Nombre: cli.Nombre, Telefono: cli.Telefono})
	if err != nil {
		return nil, err
	}
	var m *domain.Mensualidad
	if men != nil {
		if m, err = domain.NuevaMensualidad(men.FechaInicio, men.FechaFin); err != nil {
			return nil, err
		}
	}
	switch _, err := s.repo.VehiculoPorPlaca(ctx, placa); {
	case err == nil:
		return nil, domain.ErrPlacaYaRegistrada
	case !errors.Is(err, domain.ErrVehiculoNoRegistrado):
		return nil, err
	}
	v := &domain.Vehiculo{Placa: placa, Tipo: tipo}
	nuevo, err := s.repo.GuardarVehiculo(ctx, &c, v, m)
	if err != nil {
		return nil, err
	}
	s.auditar(ctx, "VEHICULO_REGISTRADO", "vehiculo", v.ID, placa+" · "+c.Nombre)
	return &ResultadoVehiculo{Vehiculo: v, Cliente: c.Nombre, ClienteNuevo: nuevo, MensualidadRegistrada: m != nil}, nil
}
