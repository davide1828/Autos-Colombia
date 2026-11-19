package postgres_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"autoscolombia/parqueadero/internal/adapters/postgres"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

func TestPagoPostgresActivaMensualidadYCancelacionEsAtomica(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1)
	reloj := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	s := application.NuevoServicio(postgres.Nuevo(db), func() time.Time { return reloj })

	pago, err := s.RegistrarPago(ctx, 1, 180000, "2026-11", domain.MedioEfectivo)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegistrarPago(ctx, 1, 180000, "2026-11", domain.MedioEfectivo); !errors.Is(err, domain.ErrPagoDuplicado) {
		t.Fatalf("el pago repetido debe rechazarse como duplicado, obtenido: %v", err)
	}
	historial, err := s.HistorialPagos(ctx, 1)
	if err != nil || len(historial) != 1 || historial[0].Monto != 180000 || historial[0].ClienteNombre != "Carlos Gómez" {
		t.Fatalf("historial: error=%v pagos=%+v", err, historial)
	}
	recibo, pdf, err := s.ReciboPago(ctx, pago.ID)
	if err != nil || len(recibo.SelloDigital) != 64 || !strings.HasPrefix(string(pdf), "%PDF-1.4") {
		t.Fatalf("recibo PostgreSQL: error=%v pago=%+v", err, recibo)
	}
	if _, err := s.RegistrarIngreso(ctx, "AAA000", 1, nil); err != nil {
		t.Fatalf("el pago debe habilitar el ingreso del mes: %v", err)
	}
	if err := s.CancelarPago(ctx, pago.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegistrarSalida(ctx, "AAA000", nil); err != nil {
		t.Fatalf("registrar salida tras cancelacion: %v", err)
	}
	if _, err := s.RegistrarIngreso(ctx, "AAA000", 1, nil); !errors.Is(err, domain.ErrMensualidadVencida) {
		t.Fatalf("la mensualidad cancelada no debe permitir nuevos ingresos: %v", err)
	}
	if count := contar(t, db, `SELECT count(*) FROM pago WHERE id_pago=$1 AND estado='CANCELADO' AND id_mensualidad IS NULL`, pago.ID); count != 1 {
		t.Fatalf("el pago cancelado debe conservarse sin la mensualidad activa; filas: %d", count)
	}
	if count := contar(t, db, `SELECT count(*) FROM mensualidad WHERE id_cliente=1 AND fecha_inicio='2026-11-01'`); count != 0 {
		t.Fatalf("la cancelacion debe revocar la mensualidad en la misma transaccion; filas: %d", count)
	}
}
