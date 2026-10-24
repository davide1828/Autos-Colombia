// Paquete config: lectura y validación de la configuración desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

const longitudMinimaToken = 16

type Config struct {
	DatabaseURL string
	APIToken    string
	Port        string
	StaticDir   string
	// PermitirSinAutenticacion solo debe usarse en desarrollo local.
	PermitirSinAutenticacion bool
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
	}
	return c, c.Validar()
}

func (c Config) Validar() error {
	if p, err := strconv.Atoi(c.Port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("PORT inválido: %q", c.Port)
	}
	if c.APIToken == "" && !c.PermitirSinAutenticacion {
		return errors.New("API_TOKEN es obligatorio (o ALLOW_INSECURE_NO_AUTH=true solo para desarrollo local)")
	}
	if c.APIToken != "" && len(c.APIToken) < longitudMinimaToken {
		return fmt.Errorf("API_TOKEN debe tener al menos %d caracteres", longitudMinimaToken)
	}
	return nil
}
