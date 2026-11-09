import { useState } from "react";
import { api, sesion } from "../api.js";

export default function Login({ onEntrar, aviso }) {
  const [usuario, setUsuario] = useState("");
  const [clave, setClave] = useState("");
  const [error, setError] = useState(aviso || "");
  const [cargando, setCargando] = useState(false);

  async function enviar(e) {
    e.preventDefault();
    const u = usuario.trim();
    if (!u) {
      setError("Escriba su nombre de usuario.");
      return;
    }
    if (!clave) {
      setError("Escriba su contraseña.");
      return;
    }

    setCargando(true);
    setError("");

    try {
      const res = await api.login(u, clave);
      onEntrar(res.usuario);
    } catch (err) {
      sesion.borrar();
      if (err.estado === 401) {
        setError("Usuario o contraseña incorrectos. Revise sus datos e intente de nuevo.");
      } else if (err.estado === 429) {
        setError("Demasiados intentos fallidos consecutivos. Por seguridad, espere 1 minuto.");
      } else {
        setError(err.message || "No se pudo iniciar sesión.");
      }
    } finally {
      setCargando(false);
    }
  }

  return (
    <main className="login">
      <form className="login__tarjeta" onSubmit={enviar} noValidate>
        <p className="marca marca--grande">Autos Colombia</p>
        <h1>Iniciar sesión</h1>
        <p className="login__subtitulo">Sistema de Información del Parqueadero</p>

        <div className="campo">
          <label htmlFor="usuario" className="etiqueta">
            Usuario
          </label>
          <input
            id="usuario"
            type="text"
            value={usuario}
            onChange={(e) => setUsuario(e.target.value)}
            autoComplete="username"
            placeholder="ej. admin o encargado"
            autoFocus
            disabled={cargando}
          />
        </div>

        <div className="campo">
          <label htmlFor="clave" className="etiqueta">
            Contraseña
          </label>
          <input
            id="clave"
            type="password"
            value={clave}
            onChange={(e) => setClave(e.target.value)}
            autoComplete="current-password"
            placeholder="••••••••••"
            disabled={cargando}
            aria-invalid={error ? "true" : "false"}
            aria-describedby="login-error"
          />
          {error && (
            <p id="login-error" className="ayuda ayuda--error" role="alert">
              {error}
            </p>
          )}
        </div>

        <button type="submit" className="boton boton--completo" disabled={cargando}>
          {cargando ? "Verificando credenciales…" : "Ingresar"}
        </button>
      </form>
    </main>
  );
}
