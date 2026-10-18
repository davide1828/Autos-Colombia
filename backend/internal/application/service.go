// Paquete application: casos de uso. Orquesta el dominio a través de puertos.
package application

import (
	"context"
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
	Celdas(ctx context.Context, estado string) ([]domain.Celda, error)
	RegistrosActivos(ctx context.Context) ([]domain.RegistroDetalle, error)
}

type NovedadEntrada struct {
	Tipo        domain.TipoNovedad `json:"tipo"`
	Descripcion string             `json:"descripcion"`
}

type Servicio struct {
	repo  Repositorio
	ahora func() time.Time
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
func (s *Servicio) RegistrarIngreso(ctx context.Context, placa string, celdaID int, nov *NovedadEntrada) (*domain.Registro, error) {
	placa, err := domain.NormalizarPlaca(placa)
	if err != nil {
		return nil, err
	}
	v, err := s.repo.VehiculoPorPlaca(ctx, placa)
	if err != nil {
		return nil, err
	}
	ahora := s.ahora()
	ok, err := s.repo.MensualidadVigente(ctx, v.ClienteID, ahora)
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
	return r, nil
}

// RegistrarSalida — RF-03.
func (s *Servicio) RegistrarSalida(ctx context.Context, placa string, nov *NovedadEntrada) (*domain.Registro, error) {
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
	ahora := s.ahora()
	r.Salida = &ahora
	if err := s.repo.GuardarSalida(ctx, r, n); err != nil {
		return nil, err
	}
	return r, nil
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
	return n, nil
}

// ConsultarCeldas — RF-05. estado vacío = todas.
func (s *Servicio) ConsultarCeldas(ctx context.Context, estado string) ([]domain.Celda, error) {
	return s.repo.Celdas(ctx, estado)
}

func (s *Servicio) VehiculosDentro(ctx context.Context) ([]domain.RegistroDetalle, error) {
	return s.repo.RegistrosActivos(ctx)
}
