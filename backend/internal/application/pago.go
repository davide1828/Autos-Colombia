package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"autoscolombia/parqueadero/internal/domain"
)

type ServicioPago struct {
	repo domain.RepositorioPago
	now  func() time.Time
}

func NuevoServicioPago(repo domain.RepositorioPago, now func() time.Time) *ServicioPago {
	if now == nil {
		now = time.Now
	}
	return &ServicioPago{repo: repo, now: now}
}

func (s *ServicioPago) CrearPago(ctx context.Context, clienteID int, monto int64, periodo string, medio domain.MedioPago) (*domain.Pago, error) {
	mes, err := time.Parse("2006-01", periodo)
	if err != nil || mes.Format("2006-01") != periodo {
		return nil, domain.ErrPeriodoPagoInvalido
	}
	if clienteID < 1 || monto < 1 || !medio.Valido() {
		return nil, domain.ErrPagoInvalido
	}
	hoy := s.now().In(ZonaColombia)
	inicioMes := time.Date(hoy.Year(), hoy.Month(), 1, 0, 0, 0, 0, time.UTC)
	inicio := time.Date(mes.Year(), mes.Month(), 1, 0, 0, 0, 0, time.UTC)
	if inicio.Before(inicioMes) {
		return nil, domain.ErrPeriodoPagoInvalido
	}
	p := &domain.Pago{
		ClienteID:     clienteID,
		Monto:         monto,
		Fecha:         s.now(),
		PeriodoInicio: inicio,
		PeriodoFin:    inicio.AddDate(0, 1, -1),
		Metodo:        medio,
		Estado:        domain.PagoPagado,
	}
	if err := s.repo.CrearPago(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ServicioPago) Historial(ctx context.Context, clienteID int) ([]domain.Pago, error) {
	if clienteID < 1 {
		return nil, domain.ErrClienteNoExiste
	}
	return s.repo.HistorialPagos(ctx, clienteID)
}

func (s *ServicioPago) GenerarRecibo(ctx context.Context, id int) (*domain.Pago, []byte, error) {
	if id < 1 {
		return nil, nil, domain.ErrPagoNoExiste
	}
	p, err := s.repo.ObtenerPago(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	p.SelloDigital = selloRecibo(p)
	if err := s.repo.GuardarSelloRecibo(ctx, id, p.Estado, p.SelloDigital); err != nil {
		return nil, nil, err
	}
	pdf, err := generarPDFRecibo(p)
	if err != nil {
		return nil, nil, err
	}
	return p, pdf, nil
}

func (s *ServicioPago) CancelarPago(ctx context.Context, id int) error {
	if id < 1 {
		return domain.ErrPagoNoExiste
	}
	return s.repo.CancelarPago(ctx, id)
}

func selloRecibo(p *domain.Pago) string {
	contenido := fmt.Sprintf("%d|%d|%s|%s|%d|%s|%s|%s|%s",
		p.ID, p.ClienteID, p.ClienteNombre, p.Documento, p.Monto, p.Fecha.UTC().Format(time.RFC3339Nano),
		p.PeriodoInicio.Format("2006-01-02"), p.PeriodoFin.Format("2006-01-02"), string(p.Metodo)+"|"+string(p.Estado))
	suma := sha256.Sum256([]byte(contenido))
	return hex.EncodeToString(suma[:])
}

func generarPDFRecibo(p *domain.Pago) ([]byte, error) {
	numero := fmt.Sprintf("AC-%d-%06d", p.Fecha.In(ZonaColombia).Year(), p.ID)
	lineas := []string{
		"AUTOS COLOMBIA - RECIBO INTERNO DE PAGO",
		"Recibo: " + numero,
		"Cliente: " + p.ClienteNombre,
		"Documento: " + p.Documento,
		fmt.Sprintf("Valor: COP %d", p.Monto),
		"Medio de pago: " + string(p.Metodo),
		"Fecha: " + p.Fecha.In(ZonaColombia).Format("2006-01-02 15:04"),
		"Mensualidad: " + p.PeriodoInicio.Format("2006-01-02") + " a " + p.PeriodoFin.Format("2006-01-02"),
		"Estado: " + string(p.Estado),
		"Sello digital (SHA-256): " + p.SelloDigital,
		"Documento interno; no es factura electronica validada por la DIAN.",
	}
	stream := strings.Builder{}
	stream.WriteString("BT\n/F1 12 Tf\n50 790 Td\n")
	for i, linea := range lineas {
		if i > 0 {
			stream.WriteString("0 -22 Td\n")
		}
		fmt.Fprintf(&stream, "(%s) Tj\n", escaparPDF(transliterarASCII(linea)))
	}
	stream.WriteString("ET\n")

	contenido := []byte(stream.String())
	objetos := [][]byte{
		[]byte("<< /Type /Catalog /Pages 2 0 R >>"),
		[]byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"),
		[]byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>"),
		[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contenido), contenido)),
	}
	var pdf strings.Builder
	pdf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objetos)+1)
	for i, objeto := range objetos {
		offsets[i+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, objeto)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return []byte(pdf.String()), nil
}

func escaparPDF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	return strings.ReplaceAll(s, ")", "\\)")
}

func transliterarASCII(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch {
		case r <= 0x7e && r >= 0x20:
			out.WriteRune(r)
		case r == 'á' || r == 'à' || r == 'ä' || r == 'Á' || r == 'À' || r == 'Ä':
			out.WriteByte('a')
		case r == 'é' || r == 'è' || r == 'ë' || r == 'É' || r == 'È' || r == 'Ë':
			out.WriteByte('e')
		case r == 'í' || r == 'ì' || r == 'ï' || r == 'Í' || r == 'Ì' || r == 'Ï':
			out.WriteByte('i')
		case r == 'ó' || r == 'ò' || r == 'ö' || r == 'Ó' || r == 'Ò' || r == 'Ö':
			out.WriteByte('o')
		case r == 'ú' || r == 'ù' || r == 'ü' || r == 'Ú' || r == 'Ù' || r == 'Ü':
			out.WriteByte('u')
		case r == 'ñ' || r == 'Ñ':
			out.WriteByte('n')
		default:
			out.WriteByte('?')
		}
	}
	return out.String()
}

func (s *Servicio) RegistrarPago(ctx context.Context, clienteID int, monto int64, periodo string, medio domain.MedioPago) (*domain.Pago, error) {
	p, err := s.pagos.CrearPago(ctx, clienteID, monto, periodo, medio)
	if err == nil {
		s.auditar(ctx, "PAGO_REGISTRADO", "pago", p.ID, fmt.Sprintf("cliente %d · %s", clienteID, periodo))
	}
	return p, err
}

func (s *Servicio) HistorialPagos(ctx context.Context, clienteID int) ([]domain.Pago, error) {
	return s.pagos.Historial(ctx, clienteID)
}

func (s *Servicio) ReciboPago(ctx context.Context, pagoID int) (*domain.Pago, []byte, error) {
	p, pdf, err := s.pagos.GenerarRecibo(ctx, pagoID)
	if err == nil {
		s.auditar(ctx, "RECIBO_GENERADO", "pago", pagoID, "")
	}
	return p, pdf, err
}

func (s *Servicio) CancelarPago(ctx context.Context, pagoID int) error {
	err := s.pagos.CancelarPago(ctx, pagoID)
	if err == nil {
		s.auditar(ctx, "PAGO_CANCELADO", "pago", pagoID, "")
	}
	return err
}
