import { useCallback, useEffect, useState } from "react";
import { api } from "../api.js";

export default function AuditoriaAdmin() {
  const [registros, setRegistros] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState("");
  const [filtro, setFiltro] = useState("");

  const cargar = useCallback(async () => {
    setCargando(true);
    try {
      const data = await api.auditoria(100);
      setRegistros(data);
      setError("");
    } catch (err) {
      setError(err.message || "Error al cargar la bitácora de auditoría.");
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  const filtrados = registros.filter((r) => {
    const q = filtro.toLowerCase();
    return (
      r.usuario.toLowerCase().includes(q) ||
      r.accion.toLowerCase().includes(q) ||
      r.entidad.toLowerCase().includes(q) ||
      (r.detalle && r.detalle.toLowerCase().includes(q))
    );
  });

  return (
    <div className="panel-admin">
      <header className="panel-admin__cabecera">
        <div>
          <h2>Bitácora de Auditoría del Sistema</h2>
          <p className="ayuda">Registro inalterable de operaciones administrativas y modificaciones de datos (RN-11).</p>
        </div>
        <button type="button" className="boton boton--secundario" onClick={cargar} disabled={cargando}>
          {cargando ? "Actualizando…" : "🔄 Refrescar"}
        </button>
      </header>

      {error && <p className="mensaje-alerta mensaje-alerta--error" role="alert">{error}</p>}

      <div className="panel-admin__filtros">
        <input
          type="search"
          placeholder="Filtrar por usuario, acción o detalle..."
          value={filtro}
          onChange={(e) => setFiltro(e.target.value)}
          className="campo-busqueda"
        />
      </div>

      <div className="tabla-contenedor">
        <table className="tabla-admin">
          <thead>
            <tr>
              <th>FECHA / HORA UTC</th>
              <th>USUARIO</th>
              <th>ACCIÓN</th>
              <th>ENTIDAD</th>
              <th>ID</th>
              <th>DETALLE DE LA OPERACIÓN</th>
            </tr>
          </thead>
          <tbody>
            {cargando ? (
              <tr>
                <td colSpan="6" className="celda-centro">Cargando bitácora de auditoría…</td>
              </tr>
            ) : filtrados.length === 0 ? (
              <tr>
                <td colSpan="6" className="celda-centro">No hay registros de auditoría que coincidan.</td>
              </tr>
            ) : (
              filtrados.map((item) => (
                <tr key={item.id_auditoria}>
                  <td className="texto-atenuado">
                    {new Date(item.momento).toLocaleString("es-CO", { timeZone: "America/Bogota" })}
                  </td>
                  <td><strong>{item.usuario}</strong></td>
                  <td>
                    <span className="badge badge--accion">{item.accion}</span>
                  </td>
                  <td>{item.entidad}</td>
                  <td className="texto-atenuado">{item.id_entidad || "—"}</td>
                  <td>{item.detalle || "—"}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="alerta-regla">
        <p><strong>Garantía de No Repudio (RN-11):</strong> La base de datos protege esta tabla mediante el disparador <code>trg_auditoria_inmutable</code>; cualquier intento de modificación o borrado es abortado por el motor.</p>
      </div>
    </div>
  );
}

