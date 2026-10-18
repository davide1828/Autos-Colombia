// Paquete domain: entidades y reglas de negocio del parqueadero "Autos Colombia".
// No depende de HTTP ni de la base de datos.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type TipoVehiculo string

const (
	Carro TipoVehiculo = "CARRO"
	Moto  TipoVehiculo = "MOTO"
)

type EstadoCelda string

const (
	CeldaDisponible EstadoCelda = "DISPONIBLE"
	CeldaOcupada    EstadoCelda = "OCUPADA"
	CeldaInactiva   EstadoCelda = "INACTIVA"
)

type TipoNovedad string

const (
	NovedadDanio    TipoNovedad = "DANIO"
	NovedadFaltante TipoNovedad = "FALTANTE"
	NovedadOtra     TipoNovedad = "OTRA"
)

type MomentoNovedad string

const (
	MomentoEntrada  MomentoNovedad = "ENTRADA"
	MomentoSalida   MomentoNovedad = "SALIDA"
	MomentoEstancia MomentoNovedad = "DURANTE_ESTANCIA"
)

// Errores de negocio (cada uno se traduce a un código HTTP en el adaptador web).
var (
	ErrPlacaInvalida        = errors.New("la placa no tiene un formato válido (ej. ABC123 o ABC12D)")
	ErrVehiculoNoRegistrado = errors.New("el vehículo no está registrado en el sistema")
	ErrMensualidadVencida   = errors.New("el cliente no tiene mensualidad vigente")
	ErrVehiculoYaDentro     = errors.New("el vehículo ya tiene un ingreso activo")
	ErrCeldaNoExiste        = errors.New("la celda no existe")
	ErrCeldaNoDisponible    = errors.New("la celda no está disponible")
	ErrCeldaIncompatible    = errors.New("la celda no corresponde al tipo de vehículo")
	ErrSinIngresoActivo     = errors.New("el vehículo no tiene un ingreso activo")
	ErrNovedadInvalida      = errors.New("la novedad requiere un tipo válido y una descripción")
)

type Cliente struct {
	ID        int    `json:"id_cliente"`
	Documento string `json:"documento"`
	Nombre    string `json:"nombre"`
}

type Vehiculo struct {
	ID        int          `json:"id_vehiculo"`
	Placa     string       `json:"placa"`
	Tipo      TipoVehiculo `json:"tipo"`
	ClienteID int          `json:"id_cliente"`
}

type Celda struct {
	ID     int          `json:"id_celda"`
	Codigo string       `json:"codigo"`
	Tipo   TipoVehiculo `json:"tipo"`
	Estado EstadoCelda  `json:"estado"`
}

// PuedeOcuparse indica si la celda está disponible para recibir un vehículo.
func (c Celda) PuedeOcuparse() bool { return c.Estado == CeldaDisponible }

type Registro struct {
	ID         int        `json:"id_registro"`
	VehiculoID int        `json:"id_vehiculo"`
	CeldaID    int        `json:"id_celda"`
	Entrada    time.Time  `json:"fecha_entrada"`
	Salida     *time.Time `json:"fecha_salida,omitempty"`
}

func (r Registro) EstaActivo() bool { return r.Salida == nil }

// Duracion de la estadía hasta la salida (o hasta "ahora" si sigue activo).
func (r Registro) Duracion(ahora time.Time) time.Duration {
	fin := ahora
	if r.Salida != nil {
		fin = *r.Salida
	}
	return fin.Sub(r.Entrada)
}

type Novedad struct {
	ID          int            `json:"id_novedad"`
	RegistroID  int            `json:"id_registro"`
	VehiculoID  int            `json:"id_vehiculo"`
	Tipo        TipoNovedad    `json:"tipo"`
	Descripcion string         `json:"descripcion"`
	Momento     MomentoNovedad `json:"momento"`
	Fecha       time.Time      `json:"fecha"`
}

// RegistroDetalle es la vista que ve el encargado de un vehículo dentro del parqueadero.
type RegistroDetalle struct {
	RegistroID int       `json:"id_registro"`
	Placa      string    `json:"placa"`
	Cliente    string    `json:"cliente"`
	Celda      string    `json:"celda"`
	Entrada    time.Time `json:"fecha_entrada"`
}

var reglaPlaca = regexp.MustCompile(`^[A-Z]{3}[0-9]{2}[0-9A-Z]$`)

// NormalizarPlaca pasa a mayúsculas, quita espacios/guiones y valida el formato colombiano.
func NormalizarPlaca(p string) (string, error) {
	p = strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(p)))
	if !reglaPlaca.MatchString(p) {
		return "", ErrPlacaInvalida
	}
	return p, nil
}

// Validar comprueba que la novedad tenga tipo conocido y descripción.
func (n Novedad) Validar() error {
	switch n.Tipo {
	case NovedadDanio, NovedadFaltante, NovedadOtra:
	default:
		return ErrNovedadInvalida
	}
	if strings.TrimSpace(n.Descripcion) == "" {
		return ErrNovedadInvalida
	}
	return nil
}
