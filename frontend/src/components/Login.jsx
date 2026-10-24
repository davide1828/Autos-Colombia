import { useState } from "react";
import { api, sesion } from "../api.js";

export default function Login({ onEntrar, aviso }) {
  const [clave, setClave] = useState("");
  const [error, setError] = useState(aviso || "");
  const [cargando, setCargando] = useState(false);

  async function enviar(e) {
    e.preventDefault();
    if (!clave.trim()) { setError("Escriba la clave de acceso."); return; }
    setCargando(true); setError("");
    try {
      await api.verificarClave(clave.trim());
      onEntrar();
    } catch (err) {
      sesion.borrar();
      setError(err.estado === 401 ? "La clave no es correcta. Revísela e inténtelo de nuevo." : err.message);
    } finally { setCargando(false); }
  }

  return (
    <main className="login">
      <form className="login__tarjeta" onSubmit={enviar} noValidate>
        <p className="marca marca--grande">Autos Colombia</p>
        <h1>Acceso del encargado</h1>
        <div className="campo">
          <label htmlFor="clave" className="etiqueta">Clave de acceso</label>
          <input id="clave" type="password" value={clave} onChange={(e) => setClave(e.target.value)} autoComplete="current-password"
            aria-invalid={error ? "true" : "false"} aria-describedby="clave-error" autoFocus />
          <p id="clave-error" className="ayuda ayuda--error" role={error ? "alert" : undefined}>{error}</p>
        </div>
        <button type="submit" className="boton" disabled={cargando}>{cargando ? "Verificando…" : "Entrar"}</button>
      </form>
    </main>
  );
}
