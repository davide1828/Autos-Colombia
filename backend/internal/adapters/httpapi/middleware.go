package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type respuestaConEstado struct {
	http.ResponseWriter
	estado int
}

func (r *respuestaConEstado) WriteHeader(c int) { r.estado = c; r.ResponseWriter.WriteHeader(c) }

func idPeticion() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Envolver aplica, de afuera hacia adentro: identificador, recuperación de pánicos,
// cabeceras de seguridad y registro estructurado de cada petición.
func Envolver(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := idPeticion()
		w.Header().Set("X-Request-Id", id)
		cabecerasDeSeguridad(w)
		rw := &respuestaConEstado{ResponseWriter: w, estado: http.StatusOK}
		inicio := time.Now()
		defer func() {
			if p := recover(); p != nil {
				log.Error("pánico recuperado", "id", id, "panico", p, "pila", string(debug.Stack()))
				if rw.estado == http.StatusOK {
					escribirError(rw, http.StatusInternalServerError, "error_interno", "ocurrió un error inesperado")
				}
			}
			log.Info("petición", "id", id, "metodo", r.Method, "ruta", r.URL.Path, "estado", rw.estado, "ms", time.Since(inicio).Milliseconds())
		}()
		next.ServeHTTP(rw, r)
	})
}

func cabecerasDeSeguridad(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
}

// limitador cuenta intentos fallidos de autenticación por IP en una ventana de tiempo.
type limitador struct {
	mu       sync.Mutex
	max      int
	ventana  time.Duration
	intentos map[string][]time.Time
	ahora    func() time.Time
}

func nuevoLimitador(max int, ventana time.Duration) *limitador {
	return &limitador{max: max, ventana: ventana, intentos: map[string][]time.Time{}, ahora: time.Now}
}

func (l *limitador) recientes(ip string) []time.Time {
	corte := l.ahora().Add(-l.ventana)
	var vivos []time.Time
	for _, t := range l.intentos[ip] {
		if t.After(corte) {
			vivos = append(vivos, t)
		}
	}
	return vivos
}

func (l *limitador) bloqueado(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.intentos[ip] = l.recientes(ip)
	return len(l.intentos[ip]) >= l.max
}

func (l *limitador) falla(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.intentos[ip] = append(l.recientes(ip), l.ahora())
}

func ipDe(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return h
}
