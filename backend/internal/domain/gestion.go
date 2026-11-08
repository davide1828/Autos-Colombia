package domain

// Iteración 2 · Gestión de usuarios (del sistema y clientes) y gestión de celdas.

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	// Usuarios del sistema (RF-13 a RF-18).
	ErrUsuarioInvalido  = errors.New("el usuario debe tener de 3 a 30 caracteres (minúsculas, números, punto, guion o guion bajo) y un nombre de máximo 100 caracteres")
	ErrClaveDebil       = errors.New("la contraseña debe tener de 10 a 128 caracteres, con al menos una letra y un número, y no puede ser igual al usuario")
	ErrRolInvalido      = errors.New("el rol debe ser ADMINISTRADOR o ENCARGADO")
	ErrUsuarioDuplicado = errors.New("ya existe un usuario con ese nombre de usuario")
	ErrUsuarioNoExiste  = errors.New("el usuario no existe")
	ErrCredenciales     = errors.New("usuario o contraseña incorrectos")
	// ErrClaveActualIncorrecta se usa al cambiar la propia contraseña: la sesión es válida, la confirmación no.
	ErrClaveActualIncorrecta = errors.New("la contraseña actual no es correcta")
	ErrSinPermiso            = errors.New("su rol no permite realizar esta operación")
	ErrUltimoAdministrador   = errors.New("debe quedar al menos un administrador activo")

	// Clientes y vehículos (RF-09 a RF-12).
	ErrClienteNoExiste     = errors.New("el cliente no existe")
	ErrMensualidadSolapada = errors.New("la mensualidad se solapa con otra ya registrada para el cliente")
	ErrVehiculoDentro      = errors.New("el vehículo está dentro del parqueadero; registre primero su salida")

	// Celdas (RF-19 a RF-21).
	ErrCeldaInvalida       = errors.New("la celda requiere un código de máximo 10 caracteres (letras, números o guion) y un tipo CARRO o MOTO")
	ErrCeldaDuplicada      = errors.New("ya existe una celda con ese código")
	ErrCeldaOcupada        = errors.New("la celda está ocupada; registre primero la salida del vehículo")
	ErrEstadoCeldaInvalido = errors.New("el estado debe ser DISPONIBLE o INACTIVA (la celda se ocupa al registrar un ingreso)")
)

// Rol define qué puede hacer un usuario del sistema.
type Rol string

const (
	RolAdministrador Rol = "ADMINISTRADOR"
	RolEncargado     Rol = "ENCARGADO"
)

func (r Rol) Valido() bool { return r == RolAdministrador || r == RolEncargado }

// Usuario es una persona que opera el sistema (no confundir con Cliente, que es quien usa el parqueadero).
type Usuario struct {
	ID        int    `json:"id_usuario"`
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	Rol       Rol    `json:"rol"`
	Activo    bool   `json:"activo"`
	ClaveHash string `json:"-"` // nunca sale por la API
}

var reLogin = regexp.MustCompile(`^[a-z0-9._-]{3,30}$`)

// NormalizarLogin pasa a minúsculas y valida el nombre de usuario.
func NormalizarLogin(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if !reLogin.MatchString(s) {
		return "", ErrUsuarioInvalido
	}
	return s, nil
}

// NormalizarNombre compacta espacios y valida el nombre visible (1 a 100 caracteres).
func NormalizarNombre(s string) (string, error) {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" || utf8.RuneCountInString(s) > 100 {
		return "", ErrUsuarioInvalido
	}
	return s, nil
}

// ValidarClave aplica la política de contraseñas: 10 a 128 caracteres, con letra y número, distinta del usuario.
func ValidarClave(clave, login string) error {
	n := utf8.RuneCountInString(clave)
	if n < 10 || n > 128 || strings.EqualFold(clave, login) {
		return ErrClaveDebil
	}
	var letra, numero bool
	for _, r := range clave {
		letra = letra || unicode.IsLetter(r)
		numero = numero || unicode.IsDigit(r)
	}
	if !letra || !numero {
		return ErrClaveDebil
	}
	return nil
}

var reCodigoCelda = regexp.MustCompile(`^[A-Z0-9-]{1,10}$`)

// NormalizarCodigoCelda pasa a mayúsculas y valida el código (ej. A-01).
func NormalizarCodigoCelda(s string) (string, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if !reCodigoCelda.MatchString(s) {
		return "", ErrCeldaInvalida
	}
	return s, nil
}

// EstadoGestionable indica si el estado puede fijarse a mano (la ocupación solo la produce un ingreso).
func (e EstadoCelda) EstadoGestionable() bool { return e == CeldaDisponible || e == CeldaInactiva }

// PeriodoMensualidad es una mensualidad expresada en días (AAAA-MM-DD).
type PeriodoMensualidad struct {
	ID          int    `json:"id_mensualidad"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
}

// ClienteDetalle reúne al cliente con sus vehículos y mensualidades para consultarlo.
type ClienteDetalle struct {
	Cliente
	Vehiculos     []Vehiculo           `json:"vehiculos"`
	Mensualidades []PeriodoMensualidad `json:"mensualidades"`
	Vigente       bool                 `json:"mensualidad_vigente"`
}

// Auditoria deja constancia de quién hizo qué y cuándo (RN-06: no se modifica ni se elimina).
type Auditoria struct {
	ID        int64     `json:"id_auditoria"`
	Momento   time.Time `json:"momento"`
	UsuarioID int       `json:"id_usuario"`
	Usuario   string    `json:"usuario"`
	Accion    string    `json:"accion"`
	Entidad   string    `json:"entidad"`
	EntidadID int       `json:"id_entidad"`
	Detalle   string    `json:"detalle,omitempty"`
}
