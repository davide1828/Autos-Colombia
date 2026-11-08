// Paquete seguridad: adaptador que implementa application.Hasheador con PBKDF2-HMAC-SHA256
// usando solo la biblioteca estándar (sin dependencias externas).
package seguridad

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// IteracionesPorDefecto sigue la recomendación de OWASP para PBKDF2-HMAC-SHA256.
const IteracionesPorDefecto = 600_000

const (
	prefijo        = "pbkdf2-sha256"
	maxIteraciones = 10_000_000 // evita que un hash corrupto deje la CPU ocupada
	tamSal         = 16
	tamClave       = 32
)

type PBKDF2 struct{ Iteraciones int }

// NuevoPBKDF2 crea el hasheador; con iter <= 0 usa IteracionesPorDefecto.
func NuevoPBKDF2(iter int) *PBKDF2 {
	if iter <= 0 {
		iter = IteracionesPorDefecto
	}
	return &PBKDF2{Iteraciones: iter}
}

// Hash devuelve "pbkdf2-sha256$iteraciones$sal$hash" (sal aleatoria por contraseña).
func (p *PBKDF2) Hash(clave string) (string, error) {
	sal := make([]byte, tamSal)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	dk := derivar([]byte(clave), sal, p.Iteraciones, tamClave)
	return fmt.Sprintf("%s$%d$%s$%s", prefijo, p.Iteraciones, base64.RawStdEncoding.EncodeToString(sal), base64.RawStdEncoding.EncodeToString(dk)), nil
}

// Verificar compara en tiempo constante; un hash mal formado simplemente no coincide.
func (p *PBKDF2) Verificar(hash, clave string) bool {
	partes := strings.Split(hash, "$")
	if len(partes) != 4 || partes[0] != prefijo {
		return false
	}
	iter, err := strconv.Atoi(partes[1])
	if err != nil || iter < 1 || iter > maxIteraciones {
		return false
	}
	sal, err1 := base64.RawStdEncoding.DecodeString(partes[2])
	esperado, err2 := base64.RawStdEncoding.DecodeString(partes[3])
	if err1 != nil || err2 != nil || len(esperado) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(derivar([]byte(clave), sal, iter, len(esperado)), esperado) == 1
}

// derivar implementa PBKDF2 (RFC 8018) con HMAC-SHA256.
func derivar(clave, sal []byte, iter, largo int) []byte {
	prf := hmac.New(sha256.New, clave)
	tam := prf.Size()
	bloques := (largo + tam - 1) / tam
	dk := make([]byte, 0, bloques*tam)
	u := make([]byte, tam)
	var idx [4]byte
	for b := 1; b <= bloques; b++ {
		prf.Reset()
		prf.Write(sal)
		binary.BigEndian.PutUint32(idx[:], uint32(b))
		prf.Write(idx[:])
		dk = prf.Sum(dk)
		t := dk[len(dk)-tam:]
		copy(u, t)
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for i := range t {
				t[i] ^= u[i]
			}
		}
	}
	return dk[:largo]
}
