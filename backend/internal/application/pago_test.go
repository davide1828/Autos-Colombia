package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/adapters/memory"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

func TestPagoCreaMensualidadYCancelacionLaRevoca(t *testing.T) {
	ctx := context.Background()
	ahora := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	repo := memory.Nuevo()
	repo.AgregarClienteCompleto(25, "10025", "Nora Ruiz", "", "")
	svc := application.NuevoServicio(repo, func() time.Time { return ahora })

	pago, err := svc.RegistrarPago(ctx, 25, 180000, "2026-11", domain.MedioTransferencia)
	if err != nil {
		t.Fatal(err)
	}
	if pago.ID == 0 || !pago.PeriodoInicio.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) ||
		!pago.PeriodoFin.Equal(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("periodo mensual inesperado: %+v", pago)
	}
	if vigente, err := repo.MensualidadVigente(ctx, 25, ahora); err != nil || !vigente {
		t.Fatalf("el pago debe activar la mensualidad: vigente=%v error=%v", vigente, err)
	}
	if _, err := svc.RegistrarPago(ctx, 25, 180000, "2026-11", domain.MedioTransferencia); !errors.Is(err, domain.ErrPagoDuplicado) {
		t.Fatalf("el pago duplicado debe rechazarse, obtenido: %v", err)
	}
	if err := svc.CancelarPago(ctx, pago.ID); err != nil {
		t.Fatal(err)
	}
	if vigente, err := repo.MensualidadVigente(ctx, 25, ahora); err != nil || vigente {
		t.Fatalf("la cancelacion debe revocar la mensualidad: vigente=%v error=%v", vigente, err)
	}
	if err := svc.CancelarPago(ctx, pago.ID); !errors.Is(err, domain.ErrPagoYaCancelado) {
		t.Fatalf("cancelar dos veces debe rechazarse, obtenido: %v", err)
	}
	if _, err := svc.RegistrarPago(ctx, 25, 180000, "2026-11", domain.MedioTransferencia); err != nil {
		t.Fatalf("un mes cuyo pago se cancelo debe poder pagarse de nuevo: %v", err)
	}
}

func TestPagoRechazaPeriodoPasadoMontoYMedioInvalidos(t *testing.T) {
	ahora := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	repo := memory.Nuevo()
	repo.AgregarClienteCompleto(25, "10025", "Nora Ruiz", "", "")
	svc := application.NuevoServicio(repo, func() time.Time { return ahora })

	for _, prueba := range []struct {
		nombre string
		monto  int64
		mes    string
		medio  domain.MedioPago
	}{
		{"monto cero", 0, "2026-11", domain.MedioEfectivo},
		{"mes mal formado", 180000, "2026-13", domain.MedioEfectivo},
		{"mes pasado", 180000, "2026-10", domain.MedioEfectivo},
		{"medio inválido", 180000, "2026-11", "CHEQUE"},
	} {
		t.Run(prueba.nombre, func(t *testing.T) {
			_, err := svc.RegistrarPago(context.Background(), 25, prueba.monto, prueba.mes, prueba.medio)
			if err == nil {
				t.Fatal("se esperaba rechazar la solicitud")
			}
		})
	}
}

func TestReciboPagoGeneraPDFConSello(t *testing.T) {
	ctx := context.Background()
	ahora := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	repo := memory.Nuevo()
	repo.AgregarClienteCompleto(25, "10025", "Nora Ruiz", "", "")
	svc := application.NuevoServicio(repo, func() time.Time { return ahora })
	pago, err := svc.RegistrarPago(ctx, 25, 180000, "2026-11", domain.MedioEfectivo)
	if err != nil {
		t.Fatal(err)
	}

	recibo, pdf, err := svc.ReciboPago(ctx, pago.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-1.4") || !strings.Contains(string(pdf), "AC-2026-000001") ||
		!strings.Contains(string(pdf), "Sello digital") ||
		len(recibo.SelloDigital) != 64 || recibo.ClienteNombre != "Nora Ruiz" {
		t.Fatalf("el recibo no tiene la estructura o datos esperados: pago=%+v", recibo)
	}
}
