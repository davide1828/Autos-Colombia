package httpapi

// Sesiones: token firmado con HMAC-SHA256 (sin estado en el servidor) con vencimiento.

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

// DuracionSesion es la vigencia de una sesión (una jornada de trabajo).
const DuracionSesion = 8 * time.Hour

var errTokenInvalido = errors.New("sesión inválida o vencida")

type firmador struct {
	clave []byte
	ttl   time.Duration
	ahora func() time.Time
}

type reclamos struct {
	ID      int        `json:"i"`
	Usuario string     `json:"u"`
	Rol     domain.Rol `json:"r"`
	Exp     int64      `json:"e"`
}

func (f *firmador) firmar(payload string) string {
	m := hmac.New(sha256.New, f.clave)
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// emitir crea un token para el actor y devuelve también su vencimiento.
func (f *firmador) emitir(a application.Actor) (string, time.Time) {
	exp := f.ahora().Add(f.ttl)
	datos, _ := json.Marshal(reclamos{ID: a.ID, Usuario: a.Usuario, Rol: a.Rol, Exp: exp.Unix()})
	payload := base64.RawURLEncoding.EncodeToString(datos)
	return payload + "." + f.firmar(payload), exp
}

// verificar comprueba la firma (en tiempo constante) y el vencimiento.
func (f *firmador) verificar(token string) (application.Actor, error) {
	partes := strings.Split(token, ".")
	if len(partes) != 2 || subtle.ConstantTimeCompare([]byte(f.firmar(partes[0])), []byte(partes[1])) != 1 {
		return application.Actor{}, errTokenInvalido
	}
	datos, err := base64.RawURLEncoding.DecodeString(partes[0])
	if err != nil {
		return application.Actor{}, errTokenInvalido
	}
	var c reclamos
	if json.Unmarshal(datos, &c) != nil || f.ahora().Unix() >= c.Exp {
		return application.Actor{}, errTokenInvalido
	}
	return application.Actor{ID: c.ID, Usuario: c.Usuario, Rol: c.Rol}, nil
}
