package config_test

import (
	"strings"
	"testing"

	"autoscolombia/parqueadero/internal/config"
)

func TestValidar(t *testing.T) {
	casos := []struct {
		nombre string
		c      config.Config
		error  string // fragmento esperado; vacío = válido
	}{
		{"válida", config.Config{APIToken: strings.Repeat("a", 16), Port: "8080"}, ""},
		{"sin token", config.Config{Port: "8080"}, "API_TOKEN es obligatorio"},
		{"sin token pero permitido en desarrollo", config.Config{Port: "8080", PermitirSinAutenticacion: true}, ""},
		{"token corto", config.Config{APIToken: "corto", Port: "8080"}, "al menos"},
		{"puerto inválido", config.Config{APIToken: strings.Repeat("a", 16), Port: "abc"}, "PORT inválido"},
		{"puerto fuera de rango", config.Config{APIToken: strings.Repeat("a", 16), Port: "70000"}, "PORT inválido"},
		{"administrador válido", config.Config{APIToken: strings.Repeat("a", 16), Port: "8080", AdminUsuario: "admin", AdminClave: "Clave-Inicial-2026"}, ""},
		{"clave de administrador débil", config.Config{APIToken: strings.Repeat("a", 16), Port: "8080", AdminUsuario: "admin", AdminClave: "corta1"}, "ADMIN_CLAVE inválida"},
		{"clave sin números", config.Config{APIToken: strings.Repeat("a", 16), Port: "8080", AdminUsuario: "admin", AdminClave: "SoloLetrasLargas"}, "ADMIN_CLAVE inválida"},
		{"usuario de administrador inválido", config.Config{APIToken: strings.Repeat("a", 16), Port: "8080", AdminUsuario: "Admin Uno", AdminClave: "Clave-Inicial-2026"}, "ADMIN_USUARIO inválido"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := c.c.Validar()
			if c.error == "" && err != nil {
				t.Fatalf("no esperaba error: %v", err)
			}
			if c.error != "" && (err == nil || !strings.Contains(err.Error(), c.error)) {
				t.Fatalf("esperaba error con %q, obtuve %v", c.error, err)
			}
		})
	}
}

func TestCargar_LeeElEntorno(t *testing.T) {
	t.Setenv("API_TOKEN", strings.Repeat("x", 20))
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://u:p@h/db")
	c, err := config.Cargar()
	if err != nil || c.Port != "9090" || c.DatabaseURL != "postgres://u:p@h/db" || c.StaticDir != "./static" {
		t.Fatalf("configuración inesperada: %+v %v", c, err)
	}
}

func TestCargar_FallaSinToken(t *testing.T) {
	t.Setenv("API_TOKEN", "")
	t.Setenv("ALLOW_INSECURE_NO_AUTH", "")
	if _, err := config.Cargar(); err == nil {
		t.Fatal("sin API_TOKEN debía fallar al arrancar")
	}
	t.Setenv("ALLOW_INSECURE_NO_AUTH", "true")
	if _, err := config.Cargar(); err != nil {
		t.Fatalf("con ALLOW_INSECURE_NO_AUTH=true debía permitir arrancar: %v", err)
	}
}

// CP-G01 · RF-14 · La configuración exige un administrador inicial con clave que cumpla la política.
func TestCargar_Administrador(t *testing.T) {
	t.Setenv("API_TOKEN", strings.Repeat("x", 20))
	t.Setenv("ADMIN_CLAVE", "Clave-Inicial-2026")
	t.Setenv("HASH_ITERACIONES", "2000")
	c, err := config.Cargar()
	if err != nil || c.AdminUsuario != "admin" || c.AdminNombre != "Administrador" || c.AdminClave != "Clave-Inicial-2026" || c.HashIteraciones != 2000 {
		t.Fatalf("configuración inesperada: %+v %v", c, err)
	}
	t.Setenv("HASH_ITERACIONES", "10")
	if _, err := config.Cargar(); err == nil || !strings.Contains(err.Error(), "HASH_ITERACIONES") {
		t.Fatalf("un costo menor a 1000 debía rechazarse: %v", err)
	}
}
