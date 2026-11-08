package domain_test

import (
	"errors"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/domain"
)

func TestNormalizarPlaca(t *testing.T) {
	validas := map[string]string{"abc123": "ABC123", " ABC-123 ": "ABC123", "xyz 98a": "XYZ98A", "DEF45G": "DEF45G"}
	for in, quiere := range validas {
		got, err := domain.NormalizarPlaca(in)
		if err != nil || got != quiere {
			t.Errorf("NormalizarPlaca(%q) = %q, %v; quería %q", in, got, err, quiere)
		}
	}
	for _, in := range []string{"", "AB123", "ABCD123", "123ABC", "AB-12", "ABC1234", "ABC12$"} {
		if _, err := domain.NormalizarPlaca(in); !errors.Is(err, domain.ErrPlacaInvalida) {
			t.Errorf("NormalizarPlaca(%q) debía rechazarse", in)
		}
	}
}

func TestNovedadValidar(t *testing.T) {
	ok := domain.Novedad{Tipo: domain.NovedadDanio, Descripcion: "Rayón"}
	if err := ok.Validar(); err != nil {
		t.Fatalf("novedad válida rechazada: %v", err)
	}
	for _, n := range []domain.Novedad{
		{Tipo: "OTRO", Descripcion: "x"}, {Tipo: domain.NovedadOtra, Descripcion: "   "}, {},
	} {
		if err := n.Validar(); !errors.Is(err, domain.ErrNovedadInvalida) {
			t.Errorf("novedad %+v debía rechazarse", n)
		}
	}
}

func TestRegistro(t *testing.T) {
	ent := time.Date(2026, 10, 18, 8, 0, 0, 0, time.UTC)
	r := domain.Registro{Entrada: ent}
	if !r.EstaActivo() || r.Duracion(ent.Add(time.Hour)) != time.Hour {
		t.Fatal("un registro sin salida está activo y su duración llega hasta ahora")
	}
	sal := ent.Add(90 * time.Minute)
	r.Salida = &sal
	if r.EstaActivo() || r.Duracion(ent.Add(10*time.Hour)) != 90*time.Minute {
		t.Fatal("un registro cerrado no está activo y su duración llega hasta la salida")
	}
}

func TestCeldaPuedeOcuparse(t *testing.T) {
	for estado, quiere := range map[domain.EstadoCelda]bool{domain.CeldaDisponible: true, domain.CeldaOcupada: false, domain.CeldaInactiva: false} {
		if got := (domain.Celda{Estado: estado}).PuedeOcuparse(); got != quiere {
			t.Errorf("estado %s: got %v", estado, got)
		}
	}
}

// CP-V07 · RF-08 · Cliente: espacios compactados y límites de documento (20), nombre (100) y teléfono (20).
func TestNormalizarCliente(t *testing.T) {
	c, err := domain.NormalizarCliente(domain.Cliente{Documento: " 123 ", Nombre: "  Ana   María  Pérez ", Telefono: " 300 "})
	if err != nil || c.Documento != "123" || c.Nombre != "Ana María Pérez" || c.Telefono != "300" {
		t.Fatalf("normalización: %+v %v", c, err)
	}
	largo := func(n int) string { return stringN("x", n) }
	invalidos := []domain.Cliente{
		{Documento: "", Nombre: "Ana"}, {Documento: "1", Nombre: "   "},
		{Documento: largo(21), Nombre: "Ana"}, {Documento: "1", Nombre: largo(101)}, {Documento: "1", Nombre: "Ana", Telefono: largo(21)},
	}
	for i, c := range invalidos {
		if _, err := domain.NormalizarCliente(c); !errors.Is(err, domain.ErrClienteInvalido) {
			t.Errorf("caso %d: esperaba ErrClienteInvalido, obtuve %v", i, err)
		}
	}
}

func stringN(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// CP-V08 · RF-08 · Solo CARRO y MOTO son tipos de vehículo válidos.
func TestTipoVehiculoValido(t *testing.T) {
	if !domain.Carro.Valido() || !domain.Moto.Valido() || domain.TipoVehiculo("BUS").Valido() || domain.TipoVehiculo("").Valido() {
		t.Fatal("solo CARRO y MOTO son tipos válidos")
	}
}

// CP-V09 · RF-08, RF-11 · Mensualidad: periodo inclusivo, un solo día válido y fechas inválidas rechazadas.
func TestMensualidad(t *testing.T) {
	m, err := domain.NuevaMensualidad("2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatal(err)
	}
	for dia, quiere := range map[string]bool{"2026-09-30": false, "2026-10-01": true, "2026-10-31": true, "2026-11-01": false} {
		d, _ := time.Parse("2006-01-02", dia)
		if m.EstaVigente(d) != quiere {
			t.Errorf("EstaVigente(%s) debía ser %v", dia, quiere)
		}
	}
	if _, err := domain.NuevaMensualidad("2026-10-31", "2026-10-01"); !errors.Is(err, domain.ErrMensualidadInvalida) {
		t.Error("fin anterior al inicio debía fallar")
	}
	for _, par := range [][2]string{{"", "2026-10-01"}, {"2026-13-01", "2026-12-01"}, {"01/10/2026", "31/10/2026"}} {
		if _, err := domain.NuevaMensualidad(par[0], par[1]); !errors.Is(err, domain.ErrMensualidadInvalida) {
			t.Errorf("fechas %v debían fallar", par)
		}
	}
	if m1, err := domain.NuevaMensualidad("2026-10-05", "2026-10-05"); err != nil || !m1.EstaVigente(m1.Inicio) {
		t.Error("una mensualidad de un solo día es válida")
	}
}
