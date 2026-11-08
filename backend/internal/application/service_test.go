package application_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

var ctx = context.Background()

// Datos de prueba: cliente 1 con mensualidad vigente, cliente 2 vencida.
func nuevoEscenario(t *testing.T, ahora time.Time) (*application.Servicio, *memory.Repo) {
	t.Helper()
	r := memory.Nuevo()
	r.AgregarCliente(1, "Carlos Gómez", "2026-10-31")
	r.AgregarCliente(2, "Marta Rojas", "2026-09-01")
	r.AgregarVehiculo(domain.Vehiculo{ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarVehiculo(domain.Vehiculo{ID: 2, Placa: "DEF456", Tipo: domain.Carro, ClienteID: 2})
	r.AgregarVehiculo(domain.Vehiculo{ID: 3, Placa: "XYZ98A", Tipo: domain.Moto, ClienteID: 1})
	r.AgregarVehiculo(domain.Vehiculo{ID: 4, Placa: "GHI789", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarCelda(domain.Celda{ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	r.AgregarCelda(domain.Celda{ID: 2, Codigo: "A-02", Tipo: domain.Carro, Estado: domain.CeldaOcupada})
	r.AgregarCelda(domain.Celda{ID: 3, Codigo: "M-01", Tipo: domain.Moto, Estado: domain.CeldaDisponible})
	r.AgregarCelda(domain.Celda{ID: 4, Codigo: "A-03", Tipo: domain.Carro, Estado: domain.CeldaInactiva})
	r.AgregarCelda(domain.Celda{ID: 5, Codigo: "A-04", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	return application.NuevoServicio(r, func() time.Time { return ahora }), r
}

var mediodia = time.Date(2026, 10, 18, 17, 0, 0, 0, time.UTC) // 12:00 en Colombia

func TestIngresoValido_OcupaCeldaYDevuelveDatos(t *testing.T) {
	s, r := nuevoEscenario(t, mediodia)
	res, err := s.RegistrarIngreso(ctx, " abc-123 ", 1, nil)
	if err != nil {
		t.Fatalf("esperaba ingreso exitoso: %v", err)
	}
	if res.Registro.ID == 0 || res.Placa != "ABC123" || res.Celda != "A-01" {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	if r.EstadoCelda(1) != domain.CeldaOcupada {
		t.Fatal("la celda debía quedar ocupada")
	}
}

func TestIngresoRechazos(t *testing.T) {
	casos := []struct {
		nombre string
		placa  string
		celda  int
		antes  func(*application.Servicio)
		quiere error
	}{
		{"placa inválida", "12", 1, nil, domain.ErrPlacaInvalida},
		{"vehículo no registrado", "ZZZ999", 1, nil, domain.ErrVehiculoNoRegistrado},
		{"mensualidad vencida", "DEF456", 1, nil, domain.ErrMensualidadVencida},
		{"vehículo ya dentro", "ABC123", 5, func(s *application.Servicio) { _, _ = s.RegistrarIngreso(ctx, "ABC123", 1, nil) }, domain.ErrVehiculoYaDentro},
		{"celda inexistente", "ABC123", 99, nil, domain.ErrCeldaNoExiste},
		{"celda ocupada", "ABC123", 2, nil, domain.ErrCeldaNoDisponible},
		{"celda inactiva", "ABC123", 4, nil, domain.ErrCeldaNoDisponible},
		{"celda de otro tipo", "ABC123", 3, nil, domain.ErrCeldaIncompatible},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s, r := nuevoEscenario(t, mediodia)
			if c.antes != nil {
				c.antes(s)
			}
			activosAntes, _ := r.RegistrosActivos(ctx)
			_, err := s.RegistrarIngreso(ctx, c.placa, c.celda, nil)
			if !errors.Is(err, c.quiere) {
				t.Fatalf("esperaba %v, obtuve %v", c.quiere, err)
			}
			if despues, _ := r.RegistrosActivos(ctx); len(despues) != len(activosAntes) {
				t.Fatal("un rechazo no debe crear registros")
			}
		})
	}
}

// Regresión: a las 8 p. m. en Colombia ya es "mañana" en UTC, pero la mensualidad
// que vence hoy sigue vigente.
func TestMensualidadSeEvaluaConElDiaDeColombia(t *testing.T) {
	ultimoDia := time.Date(2026, 11, 1, 1, 30, 0, 0, time.UTC) // 31-oct 20:30 en Colombia = 1-nov en UTC
	s, _ := nuevoEscenario(t, ultimoDia)
	if _, err := s.RegistrarIngreso(ctx, "ABC123", 1, nil); err != nil {
		t.Fatalf("el último día de la mensualidad debía permitir el ingreso: %v", err)
	}
	siguiente := time.Date(2026, 11, 1, 6, 0, 0, 0, time.UTC) // 1-nov 01:00 en Colombia
	s2, _ := nuevoEscenario(t, siguiente)
	if _, err := s2.RegistrarIngreso(ctx, "ABC123", 1, nil); !errors.Is(err, domain.ErrMensualidadVencida) {
		t.Fatalf("al día siguiente debía estar vencida, obtuve %v", err)
	}
}

func TestIngresoConNovedad(t *testing.T) {
	s, r := nuevoEscenario(t, mediodia)
	_, err := s.RegistrarIngreso(ctx, "ABC123", 1, &application.NovedadEntrada{Tipo: domain.NovedadDanio, Descripcion: "Rayón en la puerta izquierda"})
	nov := r.Novedades()
	if err != nil || len(nov) != 1 || nov[0].Momento != domain.MomentoEntrada || nov[0].RegistroID == 0 {
		t.Fatalf("la novedad de entrada no se guardó bien: %v %+v", err, nov)
	}
}

func TestIngresoNovedadInvalida_NoRegistraNiOcupaCelda(t *testing.T) {
	s, r := nuevoEscenario(t, mediodia)
	_, err := s.RegistrarIngreso(ctx, "ABC123", 1, &application.NovedadEntrada{Tipo: domain.NovedadDanio, Descripcion: "  "})
	if !errors.Is(err, domain.ErrNovedadInvalida) || r.EstadoCelda(1) != domain.CeldaDisponible {
		t.Fatalf("esperaba rechazo sin efectos, err=%v", err)
	}
}

func TestSalida_LiberaCeldaYCalculaPermanencia(t *testing.T) {
	reloj := mediodia
	r := memory.Nuevo()
	r.AgregarCliente(1, "Carlos", "2026-10-31")
	r.AgregarVehiculo(domain.Vehiculo{ID: 1, Placa: "ABC123", Tipo: domain.Carro, ClienteID: 1})
	r.AgregarCelda(domain.Celda{ID: 1, Codigo: "A-01", Tipo: domain.Carro, Estado: domain.CeldaDisponible})
	s := application.NuevoServicio(r, func() time.Time { return reloj })

	if _, err := s.RegistrarIngreso(ctx, "ABC123", 1, nil); err != nil {
		t.Fatal(err)
	}
	reloj = reloj.Add(95 * time.Minute)
	res, err := s.RegistrarSalida(ctx, "ABC123", &application.NovedadEntrada{Tipo: domain.NovedadFaltante, Descripcion: "Falta el espejo"})
	if err != nil {
		t.Fatalf("salida falló: %v", err)
	}
	if res.Minutos != 95 || res.Celda != "A-01" || res.Registro.Salida == nil {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	if r.EstadoCelda(1) != domain.CeldaDisponible {
		t.Fatal("la celda debía quedar disponible")
	}
	if n := r.Novedades(); len(n) != 1 || n[0].Momento != domain.MomentoSalida {
		t.Fatalf("novedad de salida no guardada: %+v", n)
	}
}

func TestSalidaSinIngreso(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	if _, err := s.RegistrarSalida(ctx, "ABC123", nil); !errors.Is(err, domain.ErrSinIngresoActivo) {
		t.Fatalf("esperaba ErrSinIngresoActivo, obtuve %v", err)
	}
}

func TestNovedadDuranteEstancia(t *testing.T) {
	s, r := nuevoEscenario(t, mediodia)
	nov := application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: "Alarma sonando"}
	if _, err := s.RegistrarNovedad(ctx, "ABC123", nov); !errors.Is(err, domain.ErrSinIngresoActivo) {
		t.Fatalf("sin ingreso activo debía rechazar, obtuve %v", err)
	}
	_, _ = s.RegistrarIngreso(ctx, "ABC123", 1, nil)
	n, err := s.RegistrarNovedad(ctx, "ABC123", nov)
	if err != nil || n.Momento != domain.MomentoEstancia || n.RegistroID == 0 {
		t.Fatalf("novedad no registrada: %v %+v", err, n)
	}
	if len(r.Novedades()) != 1 {
		t.Fatal("la novedad debía quedar guardada")
	}
}

func TestConsultas(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	disp, _ := s.ConsultarCeldas(ctx, "DISPONIBLE")
	if len(disp) != 3 {
		t.Fatalf("esperaba 3 celdas disponibles, hubo %d", len(disp))
	}
	_, _ = s.RegistrarIngreso(ctx, "ABC123", 1, nil)
	dentro, _ := s.VehiculosDentro(ctx)
	if len(dentro) != 1 || dentro[0].Placa != "ABC123" || dentro[0].Cliente != "Carlos Gómez" || dentro[0].Celda != "A-01" {
		t.Fatalf("vehículos dentro inesperados: %+v", dentro)
	}
}

// Concurrencia: dos encargados intentan ocupar la misma celda a la vez; solo uno lo logra.
func TestIngresoConcurrente_UnaSolaCelda(t *testing.T) {
	s, r := nuevoEscenario(t, mediodia)
	r.AgregarVehiculo(domain.Vehiculo{ID: 5, Placa: "JKL012", Tipo: domain.Carro, ClienteID: 1})
	placas := []string{"ABC123", "GHI789", "JKL012"}
	var wg sync.WaitGroup
	var exitos atomic.Int32
	for _, p := range placas {
		wg.Add(1)
		go func(placa string) {
			defer wg.Done()
			if _, err := s.RegistrarIngreso(ctx, placa, 1, nil); err == nil {
				exitos.Add(1)
			} else if !errors.Is(err, domain.ErrCeldaNoDisponible) {
				t.Errorf("error inesperado: %v", err)
			}
		}(p)
	}
	wg.Wait()
	if exitos.Load() != 1 {
		t.Fatalf("solo un ingreso debía lograrse, hubo %d", exitos.Load())
	}
}

// ---- RF-08: registrar vehículo ----

var mensOct = &application.MensualidadEntrada{FechaInicio: "2026-10-01", FechaFin: "2026-10-31"}
var anaPerez = application.ClienteEntrada{Documento: "5005005005", Nombre: " Ana  Pérez ", Telefono: "3004445566"}

// CP-V01 · RF-08 · Registrar un vehículo con cliente nuevo; luego puede ingresar.
func TestRegistrarVehiculo_ClienteNuevoYLuegoPuedeIngresar(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	res, err := s.RegistrarVehiculo(ctx, " jkl-456 ", domain.Carro, anaPerez, mensOct)
	if err != nil {
		t.Fatalf("esperaba registro exitoso: %v", err)
	}
	if res.Vehiculo.Placa != "JKL456" || res.Vehiculo.ID == 0 || res.Vehiculo.ClienteID == 0 || res.Cliente != "Ana Pérez" || !res.ClienteNuevo || !res.MensualidadRegistrada {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	// El vehículo recién registrado ya puede ingresar (RN-01) en una celda libre de su tipo.
	if _, err := s.RegistrarIngreso(ctx, "JKL456", 1, nil); err != nil {
		t.Fatalf("el vehículo registrado debía poder ingresar: %v", err)
	}
}

// CP-V02 · RF-08 · Sin mensualidad registrada el vehículo no puede ingresar (RN-01).
func TestRegistrarVehiculo_SinMensualidadNoPuedeIngresar(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	res, err := s.RegistrarVehiculo(ctx, "JKL456", domain.Moto, anaPerez, nil)
	if err != nil || res.MensualidadRegistrada {
		t.Fatalf("registro sin mensualidad: %v %+v", err, res)
	}
	if _, err := s.RegistrarIngreso(ctx, "JKL456", 3, nil); !errors.Is(err, domain.ErrMensualidadVencida) {
		t.Fatalf("sin mensualidad el ingreso debía rechazarse: %v", err)
	}
}

// CP-V03 · RF-08 · El mismo documento reutiliza al cliente (no se duplica) y comparte su mensualidad.
func TestRegistrarVehiculo_ReutilizaClientePorDocumento(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	if _, err := s.RegistrarVehiculo(ctx, "JKL456", domain.Carro, anaPerez, mensOct); err != nil {
		t.Fatal(err)
	}
	// Segundo vehículo del mismo cliente (mismo documento): no se duplica y conserva el nombre guardado.
	otro := application.ClienteEntrada{Documento: "5005005005", Nombre: "Otro nombre"}
	res, err := s.RegistrarVehiculo(ctx, "MNO12P", domain.Moto, otro, nil)
	if err != nil || res.ClienteNuevo || res.Cliente != "Ana Pérez" {
		t.Fatalf("debía reutilizar al cliente existente: %v %+v", err, res)
	}
	// Y como la mensualidad es del cliente, la moto también puede ingresar.
	if _, err := s.RegistrarIngreso(ctx, "MNO12P", 3, nil); err != nil {
		t.Fatalf("la moto del mismo cliente debía poder ingresar: %v", err)
	}
}

// CP-V04 · RF-08 · Placa repetida o inválida, tipo inválido, cliente incompleto y fechas inválidas se rechazan sin dejar datos a medias.
func TestRegistrarVehiculo_Rechazos(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	casos := []struct {
		nombre string
		placa  string
		tipo   domain.TipoVehiculo
		cli    application.ClienteEntrada
		men    *application.MensualidadEntrada
		quiere error
	}{
		{"placa duplicada", "ABC123", domain.Carro, anaPerez, mensOct, domain.ErrPlacaYaRegistrada},
		{"placa duplicada con otro formato", " abc-123 ", domain.Carro, anaPerez, mensOct, domain.ErrPlacaYaRegistrada},
		{"placa inválida", "AB12", domain.Carro, anaPerez, mensOct, domain.ErrPlacaInvalida},
		{"tipo inválido", "JKL456", "BUS", anaPerez, mensOct, domain.ErrTipoVehiculoInvalido},
		{"tipo vacío", "JKL456", "", anaPerez, mensOct, domain.ErrTipoVehiculoInvalido},
		{"sin documento", "JKL456", domain.Carro, application.ClienteEntrada{Nombre: "Ana"}, mensOct, domain.ErrClienteInvalido},
		{"sin nombre", "JKL456", domain.Carro, application.ClienteEntrada{Documento: "1"}, mensOct, domain.ErrClienteInvalido},
		{"fechas invertidas", "JKL456", domain.Carro, anaPerez, &application.MensualidadEntrada{FechaInicio: "2026-10-31", FechaFin: "2026-10-01"}, domain.ErrMensualidadInvalida},
		{"fecha mal escrita", "JKL456", domain.Carro, anaPerez, &application.MensualidadEntrada{FechaInicio: "ayer", FechaFin: "2026-10-01"}, domain.ErrMensualidadInvalida},
	}
	for _, c := range casos {
		if _, err := s.RegistrarVehiculo(ctx, c.placa, c.tipo, c.cli, c.men); !errors.Is(err, c.quiere) {
			t.Errorf("%s: esperaba %v, obtuve %v", c.nombre, c.quiere, err)
		}
	}
	// Ningún rechazo debe haber dejado el vehículo a medias.
	if _, err := s.RegistrarIngreso(ctx, "JKL456", 1, nil); !errors.Is(err, domain.ErrVehiculoNoRegistrado) {
		t.Fatalf("tras los rechazos la placa no debía existir: %v", err)
	}
}

// CP-V05 · RF-08, RNF-11 · Doce registros simultáneos de la misma placa: solo uno gana.
func TestRegistrarVehiculo_MismaPlacaSimultanea_SoloUnoGana(t *testing.T) {
	s, _ := nuevoEscenario(t, mediodia)
	var ok, dup int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.RegistrarVehiculo(ctx, "JKL456", domain.Carro, anaPerez, mensOct)
			switch {
			case err == nil:
				atomic.AddInt32(&ok, 1)
			case errors.Is(err, domain.ErrPlacaYaRegistrada):
				atomic.AddInt32(&dup, 1)
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()
	if ok != 1 || dup != 11 {
		t.Fatalf("esperaba 1 registro y 11 rechazos, hubo %d y %d", ok, dup)
	}
}
