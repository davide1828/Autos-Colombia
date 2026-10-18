package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

// repoMemoria es un doble de prueba del puerto Repositorio.
type repoMemoria struct {
	vehiculos map[string]*domain.Vehiculo
	vigentes  map[int]bool
	celdas    map[int]*domain.Celda
	activos   map[int]*domain.Registro // por vehículo
	novedades []domain.Novedad
	sig       int
}

func nuevoRepo() *repoMemoria {
	return &repoMemoria{
		vehiculos: map[string]*domain.Vehiculo{
			"ABC123": {ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1},
			"DEF456": {ID: 2, Placa: "DEF456", Tipo: domain.Carro, ClienteID: 2},
			"XYZ98A": {ID: 3, Placa: "XYZ98A", Tipo: domain.Moto, ClienteID: 1},
		},
		vigentes: map[int]bool{1: true, 2: false},
		celdas: map[int]*domain.Celda{
			1: {ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible},
			2: {ID: 2, Codigo: "A-02", Tipo: domain.Carro, Estado: domain.CeldaOcupada},
			3: {ID: 3, Codigo: "M-01", Tipo: domain.Moto, Estado: domain.CeldaDisponible},
		},
		activos: map[int]*domain.Registro{},
	}
}

func (r *repoMemoria) VehiculoPorPlaca(_ context.Context, p string) (*domain.Vehiculo, error) {
	if v, ok := r.vehiculos[p]; ok {
		return v, nil
	}
	return nil, domain.ErrVehiculoNoRegistrado
}
func (r *repoMemoria) MensualidadVigente(_ context.Context, c int, _ time.Time) (bool, error) {
	return r.vigentes[c], nil
}
func (r *repoMemoria) Celda(_ context.Context, id int) (*domain.Celda, error) {
	if c, ok := r.celdas[id]; ok {
		return c, nil
	}
	return nil, domain.ErrCeldaNoExiste
}
func (r *repoMemoria) RegistroActivoDeVehiculo(_ context.Context, v int) (*domain.Registro, error) {
	return r.activos[v], nil
}
func (r *repoMemoria) GuardarIngreso(_ context.Context, g *domain.Registro, n *domain.Novedad) error {
	r.sig++
	g.ID = r.sig
	r.activos[g.VehiculoID] = g
	r.celdas[g.CeldaID].Estado = domain.CeldaOcupada
	if n != nil {
		n.RegistroID = g.ID
		r.novedades = append(r.novedades, *n)
	}
	return nil
}
func (r *repoMemoria) GuardarSalida(_ context.Context, g *domain.Registro, n *domain.Novedad) error {
	delete(r.activos, g.VehiculoID)
	r.celdas[g.CeldaID].Estado = domain.CeldaDisponible
	if n != nil {
		n.RegistroID = g.ID
		r.novedades = append(r.novedades, *n)
	}
	return nil
}
func (r *repoMemoria) GuardarNovedad(_ context.Context, n *domain.Novedad) error {
	r.novedades = append(r.novedades, *n)
	return nil
}
func (r *repoMemoria) Celdas(context.Context, string) ([]domain.Celda, error) { return nil, nil }
func (r *repoMemoria) RegistrosActivos(context.Context) ([]domain.RegistroDetalle, error) {
	return nil, nil
}

func servicio(r *repoMemoria) *application.Servicio {
	fijo := time.Date(2026, 10, 18, 8, 0, 0, 0, time.UTC)
	return application.NuevoServicio(r, func() time.Time { return fijo })
}

func TestIngresoValido_OcupaCelda(t *testing.T) {
	r := nuevoRepo()
	reg, err := servicio(r).RegistrarIngreso(context.Background(), " abc-123 ", 1, nil)
	if err != nil || reg.ID == 0 {
		t.Fatalf("esperaba ingreso exitoso, err=%v", err)
	}
	if r.celdas[1].Estado != domain.CeldaOcupada {
		t.Fatal("la celda debía quedar ocupada")
	}
}

func TestIngresoRechazos(t *testing.T) {
	casos := []struct {
		nombre  string
		placa   string
		celda   int
		prepara func(*repoMemoria)
		quiere  error
	}{
		{"placa inválida", "12", 1, nil, domain.ErrPlacaInvalida},
		{"vehículo no registrado", "ZZZ999", 1, nil, domain.ErrVehiculoNoRegistrado},
		{"mensualidad vencida", "DEF456", 1, nil, domain.ErrMensualidadVencida},
		{"vehículo ya dentro", "ABC123", 1, func(r *repoMemoria) { r.activos[1] = &domain.Registro{ID: 9, VehiculoID: 1, CeldaID: 2} }, domain.ErrVehiculoYaDentro},
		{"celda inexistente", "ABC123", 99, nil, domain.ErrCeldaNoExiste},
		{"celda ocupada", "ABC123", 2, nil, domain.ErrCeldaNoDisponible},
		{"celda de otro tipo", "ABC123", 3, nil, domain.ErrCeldaIncompatible},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := nuevoRepo()
			if c.prepara != nil {
				c.prepara(r)
			}
			_, err := servicio(r).RegistrarIngreso(context.Background(), c.placa, c.celda, nil)
			if !errors.Is(err, c.quiere) {
				t.Fatalf("esperaba %v, obtuve %v", c.quiere, err)
			}
			if len(r.activos) > 0 && c.nombre != "vehículo ya dentro" {
				t.Fatal("un rechazo no debe crear registros")
			}
		})
	}
}

func TestIngresoConNovedad(t *testing.T) {
	r := nuevoRepo()
	_, err := servicio(r).RegistrarIngreso(context.Background(), "ABC123", 1,
		&application.NovedadEntrada{Tipo: domain.NovedadDanio, Descripcion: "Rayón en la puerta izquierda"})
	if err != nil || len(r.novedades) != 1 || r.novedades[0].Momento != domain.MomentoEntrada {
		t.Fatalf("la novedad de entrada no se guardó bien: %v %+v", err, r.novedades)
	}
}

func TestIngresoNovedadInvalida_NoRegistra(t *testing.T) {
	r := nuevoRepo()
	_, err := servicio(r).RegistrarIngreso(context.Background(), "ABC123", 1,
		&application.NovedadEntrada{Tipo: domain.NovedadDanio, Descripcion: "  "})
	if !errors.Is(err, domain.ErrNovedadInvalida) || len(r.activos) != 0 {
		t.Fatalf("esperaba rechazo sin registro, err=%v", err)
	}
}

func TestSalida_LiberaCelda(t *testing.T) {
	r := nuevoRepo()
	s := servicio(r)
	if _, err := s.RegistrarIngreso(context.Background(), "ABC123", 1, nil); err != nil {
		t.Fatal(err)
	}
	reg, err := s.RegistrarSalida(context.Background(), "ABC123", nil)
	if err != nil || reg.Salida == nil {
		t.Fatalf("salida falló: %v", err)
	}
	if r.celdas[1].Estado != domain.CeldaDisponible {
		t.Fatal("la celda debía quedar disponible")
	}
}

func TestSalidaSinIngreso(t *testing.T) {
	_, err := servicio(nuevoRepo()).RegistrarSalida(context.Background(), "ABC123", nil)
	if !errors.Is(err, domain.ErrSinIngresoActivo) {
		t.Fatalf("esperaba ErrSinIngresoActivo, obtuve %v", err)
	}
}

func TestNovedadDuranteEstancia(t *testing.T) {
	r := nuevoRepo()
	s := servicio(r)
	_, err := s.RegistrarNovedad(context.Background(), "ABC123", application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: "Alarma sonando"})
	if !errors.Is(err, domain.ErrSinIngresoActivo) {
		t.Fatalf("sin ingreso activo debía rechazar, obtuve %v", err)
	}
	_, _ = s.RegistrarIngreso(context.Background(), "ABC123", 1, nil)
	n, err := s.RegistrarNovedad(context.Background(), "ABC123", application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: "Alarma sonando"})
	if err != nil || n.Momento != domain.MomentoEstancia {
		t.Fatalf("novedad no registrada: %v", err)
	}
}

func TestDuracionRegistro(t *testing.T) {
	ent := time.Date(2026, 10, 18, 8, 0, 0, 0, time.UTC)
	sal := ent.Add(90 * time.Minute)
	if d := (domain.Registro{Entrada: ent, Salida: &sal}).Duracion(sal); d != 90*time.Minute {
		t.Fatalf("duración incorrecta: %v", d)
	}
}
