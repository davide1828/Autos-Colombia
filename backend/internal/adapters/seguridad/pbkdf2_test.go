package seguridad

import (
	"encoding/hex"
	"strings"
	"testing"
)

// CP-S01 · RNF-08 · PBKDF2-HMAC-SHA256 reproduce los vectores de prueba conocidos (contraseña "password", sal "salt").
func TestDerivar_VectoresConocidos(t *testing.T) {
	casos := []struct {
		iter int
		esp  string
	}{
		{1, "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"},
		{2, "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"},
		{4096, "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a"},
	}
	for _, c := range casos {
		if got := hex.EncodeToString(derivar([]byte("password"), []byte("salt"), c.iter, 32)); got != c.esp {
			t.Errorf("iter=%d: obtuve %s", c.iter, got)
		}
	}
}

// CP-S02 · RNF-08 · Cada hash lleva una sal distinta y la clave solo verifica con el texto exacto.
func TestHashYVerificar(t *testing.T) {
	h := NuevoPBKDF2(1000)
	a, err := h.Hash("Clave-Segura-123")
	if err != nil || !strings.HasPrefix(a, "pbkdf2-sha256$1000$") {
		t.Fatalf("hash: %q %v", a, err)
	}
	b, _ := h.Hash("Clave-Segura-123")
	if a == b {
		t.Fatal("la sal debe ser distinta en cada hash")
	}
	if !h.Verificar(a, "Clave-Segura-123") || h.Verificar(a, "clave-segura-123") || h.Verificar(a, "") {
		t.Fatal("verificación incorrecta")
	}
	// Un hash creado con otras iteraciones sigue verificándose (el costo viaja dentro del hash).
	if !NuevoPBKDF2(5).Verificar(a, "Clave-Segura-123") {
		t.Fatal("debe verificar con las iteraciones guardadas en el hash")
	}
}

// CP-S03 · RNF-08 · Un hash corrupto o con costo absurdo no verifica ni bloquea la CPU.
func TestVerificar_HashMalFormado(t *testing.T) {
	h := NuevoPBKDF2(10)
	for _, malo := range []string{"", "texto", "pbkdf2-sha256$x$a$b", "pbkdf2-sha256$0$YQ$YQ", "pbkdf2-sha256$99999999999$YQ$YQ", "md5$1$YQ$YQ", "pbkdf2-sha256$1$***$YQ", "pbkdf2-sha256$1$YQ$"} {
		if h.Verificar(malo, "x") {
			t.Errorf("%q no debía verificar", malo)
		}
	}
}

// CP-S04 · RNF-08 · El costo por defecto es de al menos 600 000 iteraciones.
func TestIteracionesPorDefecto(t *testing.T) {
	if NuevoPBKDF2(0).Iteraciones != IteracionesPorDefecto || IteracionesPorDefecto < 600_000 {
		t.Fatal("el costo por defecto debe ser al menos 600000")
	}
}
