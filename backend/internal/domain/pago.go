package domain

import (
	"context"
	"errors"
	"time"
)

type MedioPago string

const (
	MedioEfectivo      MedioPago = "EFECTIVO"
	MedioTransferencia MedioPago = "TRANSFERENCIA"
	MedioTarjeta       MedioPago = "TARJETA"
)

func (m MedioPago) Valido() bool {
	return m == MedioEfectivo || m == MedioTransferencia || m == MedioTarjeta
}

type EstadoPago string

const (
	PagoPagado    EstadoPago = "PAGADO"
	PagoCancelado EstadoPago = "CANCELADO"
)

var (
	ErrPagoNoExiste        = errors.New("el pago no existe")
	ErrPagoDuplicado       = errors.New("ya existe un pago activo para ese mes")
	ErrPagoInvalido        = errors.New("el monto debe ser mayor que cero y el medio de pago debe ser válido")
	ErrPeriodoPagoInvalido = errors.New("el periodo debe ser el mes actual o un mes futuro en formato AAAA-MM")
	ErrPagoYaCancelado     = errors.New("el pago ya está cancelado")
	ErrPagoCambio          = errors.New("el estado del pago cambió; vuelva a generar el recibo")
)

type Pago struct {
	ID            int        `json:"id_pago"`
	ClienteID     int        `json:"id_cliente"`
	ClienteNombre string     `json:"cliente,omitempty"`
	Documento     string     `json:"documento,omitempty"`
	Monto         int64      `json:"monto"`
	Fecha         time.Time  `json:"fecha"`
	PeriodoInicio time.Time  `json:"periodo_inicio"`
	PeriodoFin    time.Time  `json:"periodo_fin"`
	Metodo        MedioPago  `json:"medio_pago"`
	Estado        EstadoPago `json:"estado"`
	SelloDigital  string     `json:"sello_digital,omitempty"`
}

type RepositorioPago interface {
	CrearPago(ctx context.Context, p *Pago) error
	HistorialPagos(ctx context.Context, clienteID int) ([]Pago, error)
	ObtenerPago(ctx context.Context, id int) (*Pago, error)
	GuardarSelloRecibo(ctx context.Context, id int, estado EstadoPago, sello string) error
	CancelarPago(ctx context.Context, id int) error
}
