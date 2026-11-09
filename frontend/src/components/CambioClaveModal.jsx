import { useState } from "react";
import { api } from "../api.js";

export default function CambioClaveModal({ abierto, onCerrar, usuarioActual }) {
  const [actual, setActual] = useState("");
  const [nueva, setNueva] = useState("");
  const [confirmar, setConfirmar] = useState("");
  const [error, setError] = useState("");
  const [exito, setExito] = useState(false);
  const [guardando, setGuardando] = useState(false);

  if (!abierto) return null;

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setExito(false);

    if (!actual) {
      setError("Ingrese su contraseña actual.");
      return;
    }
    if (nueva.length < 10 || nueva.length > 128) {
      setError("La nueva contraseña debe tener entre 10 y 128 caracteres.");
      return;
    }
    if (!/[A-Za-z]/.test(nueva) || !/[0-9]/.test(nueva)) {
      setError("La contraseña debe incluir al menos una letra y un número.");
      return;
    }
    if (usuarioActual && nueva.toLowerCase() === usuarioActual.toLowerCase()) {
      setError("La contraseña no puede ser igual a su nombre de usuario.");
      return;
    }
    if (nueva !== confirmar) {
      setError("La nueva contraseña y su confirmación no coinciden.");
      return;
    }

    setGuardando(true);
    try {
      await api.cambiarMiClave(actual, nueva);
      setExito(true);
      setTimeout(() => {
        onCerrar();
        setActual("");
        setNueva("");
        setConfirmar("");
        setExito(false);
      }, 1500);
    } catch (err) {
      setError(err.message || "Error al actualizar la contraseña.");
    } finally {
      setGuardando(false);
    }
  }

  return (
    <div className="modal-fondo" role="dialog" aria-modal="true" aria-labelledby="modal-clave-titulo">
      <div className="modal-caja">
        <header className="modal-cabecera">
          <h2 id="modal-clave-titulo">Cambiar Contraseña</h2>
          <button type="button" className="icono-boton" onClick={onCerrar} aria-label="Cerrar modal">
            ✕
          </button>
        </header>

        {exito ? (
          <div className="modal-cuerpo">
            <p className="mensaje-exito">✓ Contraseña actualizada correctamente.</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="modal-cuerpo" noValidate>
            {error && (
              <p className="ayuda ayuda--error" role="alert">
                {error}
              </p>
            )}

            <div className="campo">
              <label htmlFor="clave-actual" className="etiqueta">
                Contraseña actual *
              </label>
              <input
                id="clave-actual"
                type="password"
                value={actual}
                onChange={(e) => setActual(e.target.value)}
                autoComplete="current-password"
                required
              />
            </div>

            <div className="campo">
              <label htmlFor="clave-nueva" className="etiqueta">
                Nueva contraseña *
              </label>
              <input
                id="clave-nueva"
                type="password"
                value={nueva}
                onChange={(e) => setNueva(e.target.value)}
                autoComplete="new-password"
                placeholder="Mínimo 10 caracteres, letra y número"
                required
              />
              <p className="ayuda">Mínimo 10 caracteres con al menos una letra y un dígito.</p>
            </div>

            <div className="campo">
              <label htmlFor="clave-confirmar" className="etiqueta">
                Confirmar nueva contraseña *
              </label>
              <input
                id="clave-confirmar"
                type="password"
                value={confirmar}
                onChange={(e) => setConfirmar(e.target.value)}
                autoComplete="new-password"
                required
              />
            </div>

            <footer className="modal-acciones">
              <button type="button" className="boton boton--secundario" onClick={onCerrar} disabled={guardando}>
                Cancelar
              </button>
              <button type="submit" className="boton" disabled={guardando}>
                {guardando ? "Guardando…" : "Actualizar Contraseña"}
              </button>
            </footer>
          </form>
        )}
      </div>
    </div>
  );
}

