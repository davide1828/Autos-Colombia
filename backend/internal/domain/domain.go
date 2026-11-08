// Paquete domain: entidades y reglas de negocio del parqueadero "Autos Colombia".
// No depende de HTTP ni de la base de datos.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
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

	// Registro de vehículos (RF-08).
	ErrPlacaYaRegistrada    = errors.New("la placa ya está registrada en el sistema")
	ErrTipoVehiculoInvalido = errors.New("el tipo de vehículo debe ser CARRO o MOTO")
	ErrClienteInvalido      = errors.New("el cliente requiere documento (máx. 20 caracteres) y nombre (máx. 100); el teléfono admite máx. 20")
	ErrMensualidadInvalida  = errors.New("la mensualidad requiere fechas válidas (AAAA-MM-DD) y la fecha final no puede ser anterior a la inicial")
)

type Cliente struct {
	ID        int    `json:"id_cliente"`
	Documento string `json:"documento"`
	Nombre    string `json:"nombre"`
	Telefono  string `json:"telefono,omitempty"`
}

// NormalizarCliente quita espacios sobrantes y valida los datos obligatorios del cliente.
func NormalizarCliente(c Cliente) (Cliente, error) {
	c.Documento = strings.TrimSpace(c.Documento)
	c.Nombre = strings.Join(strings.Fields(c.Nombre), " ")
	c.Telefono = strings.TrimSpace(c.Telefono)
	if c.Documento == "" || c.Nombre == "" ||
		utf8.RuneCountInString(c.Documento) > 20 || utf8.RuneCountInString(c.Nombre) > 100 || utf8.RuneCountInString(c.Telefono) > 20 {
		return Cliente{}, ErrClienteInvalido
	}
	return c, nil
}

// Valido indica si el tipo es CARRO o MOTO.
func (t TipoVehiculo) Valido() bool { return t == Carro || t == Moto }

// Mensualidad es el periodo (días calendario, ambos extremos incluidos) en que el cliente puede usar el parqueadero.
type Mensualidad struct {
	ID        int       `json:"id_mensualidad"`
	ClienteID int       `json:"id_cliente"`
	Inicio    time.Time `json:"fecha_inicio"`
	Fin       time.Time `json:"fecha_fin"`
}

const formatoDia = "2006-01-02"

// NuevaMensualidad construye una mensualidad a partir de fechas AAAA-MM-DD y valida el rango.
func NuevaMensualidad(inicio, fin string) (*Mensualidad, error) {
	i, err1 := time.Parse(formatoDia, strings.TrimSpace(inicio))
	f, err2 := time.Parse(formatoDia, strings.TrimSpace(fin))
	if err1 != nil || err2 != nil || f.Before(i) {
		return nil, ErrMensualidadInvalida
	}
	return &Mensualidad{Inicio: i, Fin: f}, nil
}

// EstaVigente indica si el día calendario de "dia" está dentro del periodo.
func (m Mensualidad) EstaVigente(dia time.Time) bool {
	d := dia.Format(formatoDia)
	return d >= m.Inicio.Format(formatoDia) && d <= m.Fin.Format(formatoDia)
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
