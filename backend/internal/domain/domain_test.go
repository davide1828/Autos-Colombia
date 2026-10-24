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
