package domain_test

import (
	"errors"
	"strings"
	"testing"

	"autoscolombia/parqueadero/internal/domain"
)

// CP-D01 · RF-14 · Nombre de usuario: minúsculas, números, punto, guion y guion bajo; 3 a 30 caracteres.
func TestNormalizarLogin(t *testing.T) {
	for entrada, esperado := range map[string]string{" Ana.Perez ": "ana.perez", "ABC": "abc", "a-b_c.1": "a-b_c.1", strings.Repeat("x", 30): strings.Repeat("x", 30)} {
		if got, err := domain.NormalizarLogin(entrada); err != nil || got != esperado {
			t.Errorf("%q → %q %v", entrada, got, err)
		}
	}
	for _, malo := range []string{"", "ab", strings.Repeat("x", 31), "a b", "a@b", "ñandú", "a/b", "a;b", "' or 1=1 --"} {
		if _, err := domain.NormalizarLogin(malo); !errors.Is(err, domain.ErrUsuarioInvalido) {
			t.Errorf("%q debía rechazarse", malo)
		}
	}
}

// CP-D02 · RF-14, RNF-10 · Política de contraseñas.
func TestValidarClave(t *testing.T) {
	buenas := []string{"abcdefgh12", "Clave-Segura-2026", "contraseña2026ñ", strings.Repeat("a", 127) + "1"}
	for _, c := range buenas {
		if err := domain.ValidarClave(c, "usuario"); err != nil {
			t.Errorf("%q debía aceptarse: %v", c, err)
		}
	}
	malas := map[string]string{"9 caracteres": "abcdefg12", "129 caracteres": strings.Repeat("a", 128) + "1", "sin número": "SoloLetrasAqui", "sin letra": "1234567890", "vacía": "",
		"igual al usuario": "usuario1234", "igual al usuario en mayúsculas": "USUARIO1234"}
	for nombre, c := range malas {
		if err := domain.ValidarClave(c, "usuario1234"); !errors.Is(err, domain.ErrClaveDebil) {
			t.Errorf("%s debía rechazarse", nombre)
		}
	}
}

// CP-D03 · RF-19 · Código de celda.
func TestNormalizarCodigoCelda(t *testing.T) {
	for entrada, esperado := range map[string]string{" a-01 ": "A-01", "M1": "M1", "0": "0", "ABCDE-1234": "ABCDE-1234"} {
		if got, err := domain.NormalizarCodigoCelda(entrada); err != nil || got != esperado {
			t.Errorf("%q → %q %v", entrada, got, err)
		}
	}
	for _, malo := range []string{"", "   ", "ABCDE-12345", "A 01", "A_01", "A.01", "Ñ-01", "A-01;"} {
		if _, err := domain.NormalizarCodigoCelda(malo); !errors.Is(err, domain.ErrCeldaInvalida) {
			t.Errorf("%q debía rechazarse", malo)
		}
	}
}

func TestRolYEstadoGestionable(t *testing.T) {
	if !domain.RolAdministrador.Valido() || !domain.RolEncargado.Valido() || domain.Rol("ROOT").Valido() || domain.Rol("").Valido() {
		t.Error("solo ADMINISTRADOR y ENCARGADO son roles válidos")
	}
	if !domain.CeldaDisponible.EstadoGestionable() || !domain.CeldaInactiva.EstadoGestionable() || domain.CeldaOcupada.EstadoGestionable() || domain.EstadoCelda("X").EstadoGestionable() {
		t.Error("solo DISPONIBLE e INACTIVA se fijan a mano")
	}
}
