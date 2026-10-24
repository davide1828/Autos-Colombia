package httpapi

import (
	"net/http"
	"strings"
)

// Estaticos sirve el frontend compilado. Los archivos con hash en /assets/ se cachean un año;
// el resto (index.html, manifest) se revalida siempre para que las actualizaciones lleguen.
func Estaticos(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fs.ServeHTTP(w, r)
	})
}
