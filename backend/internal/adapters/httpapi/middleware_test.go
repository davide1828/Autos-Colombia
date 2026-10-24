package httpapi_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
)

var silencio = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestEnvolver_CabecerasYRecuperacionDePanico(t *testing.T) {
	h := httpapi.Envolver(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") }), silencio)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
	if rec.Code != 500 {
		t.Fatalf("un pánico debía convertirse en 500, obtuve %d", rec.Code)
	}
	for _, c := range []string{"X-Request-Id", "X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Referrer-Policy"} {
		if rec.Header().Get(c) == "" {
			t.Errorf("falta la cabecera %s", c)
		}
	}
}

func TestEstaticos_Cache(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "assets"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "assets", "app-abc123.js"), []byte("1"), 0o644)
	h := httpapi.Estaticos(dir)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/assets/app-abc123.js", nil))
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("los assets con hash deben cachearse: %q", got)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" || rec.Code != 200 {
		t.Fatalf("index.html debe revalidarse: %q (%d)", got, rec.Code)
	}
}
