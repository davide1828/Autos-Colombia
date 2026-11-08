// Paquete config: lectura y validación de la configuración desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"autoscolombia/parqueadero/internal/domain"
)

const longitudMinimaToken = 16

type Config struct {
	DatabaseURL string
	APIToken    string
	Port        string
	StaticDir   string
	// PermitirSinAutenticacion solo debe usarse en desarrollo local.
	PermitirSinAutenticacion bool

	// Primer administrador: se crea al arrancar solo si todavía no existe ningún usuario.
	AdminUsuario string
	AdminNombre  string
	AdminClave   string
	// HashIteraciones es el costo de PBKDF2 (0 = valor por defecto, 600000).
	HashIteraciones int
}

func env(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}

// Cargar lee el entorno y falla rápido si la configuración es insegura o inválida.
func Cargar() (Config, error) {
	sinAuth, _ := strconv.ParseBool(os.Getenv("ALLOW_INSECURE_NO_AUTH"))
	c := Config{
		DatabaseURL:              env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/parqueadero?sslmode=disable"),
		APIToken:                 os.Getenv("API_TOKEN"),
		Port:                     env("PORT", "8080"),
		StaticDir:                env("STATIC_DIR", "./static"),
		PermitirSinAutenticacion: sinAuth,
		AdminUsuario:             env("ADMIN_USUARIO", "admin"),
		AdminNombre:              env("ADMIN_NOMBRE", "Administrador"),
		AdminClave:               os.Getenv("ADMIN_CLAVE"),
	}
	if v := os.Getenv("HASH_ITERACIONES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1000 {
			return c, fmt.Errorf("HASH_ITERACIONES inválido: %q (mínimo 1000)", v)
		}
		c.HashIteraciones = n
	}
	return c, c.Validar()
}

func (c Config) Validar() error {
	if p, err := strconv.Atoi(c.Port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("PORT inválido: %q", c.Port)
	}
	if c.APIToken == "" && !c.PermitirSinAutenticacion {
		return errors.New("API_TOKEN es obligatorio (es el secreto que firma las sesiones; o ALLOW_INSECURE_NO_AUTH=true solo para desarrollo local)")
	}
	if c.APIToken != "" && len(c.APIToken) < longitudMinimaToken {
		return fmt.Errorf("API_TOKEN debe tener al menos %d caracteres", longitudMinimaToken)
	}
	if c.AdminClave != "" {
		if _, err := domain.NormalizarLogin(c.AdminUsuario); err != nil {
			return fmt.Errorf("ADMIN_USUARIO inválido: %w", err)
		}
		if err := domain.ValidarClave(c.AdminClave, c.AdminUsuario); err != nil {
			return fmt.Errorf("ADMIN_CLAVE inválida: %w", err)
		}
	}
	return nil
}
