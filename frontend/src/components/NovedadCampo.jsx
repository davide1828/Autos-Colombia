import Icono from "./Icono.jsx";

const TIPOS = [["DANIO", "Daño"], ["FALTANTE", "Faltante"], ["OTRA", "Otra"]];
const MAX = 500;

/** Novedad opcional (con botón para desplegar) u obligatoria (siempre visible). */
export default function NovedadCampo({ id, obligatoria = false, abierta, onAbrir, valor, onCambio, error }) {
  const visible = obligatoria || abierta;
  return (
    <div className="novedad">
      {!obligatoria && (
        <button type="button" className="enlace" aria-expanded={abierta} aria-controls={`${id}-panel`} onClick={() => onAbrir(!abierta)}>
          <Icono nombre={abierta ? "menos" : "mas"} tamano={18} />
          {abierta ? "Quitar novedad" : "Agregar novedad"}
        </button>
      )}
      {visible && (
        <div id={`${id}-panel`} className="novedad__panel">
          <fieldset className="grupo">
            <legend className="etiqueta">Tipo de novedad</legend>
            <div className="segmentos" role="radiogroup">
              {TIPOS.map(([v, t]) => (
                <label key={v} className="segmento">
                  <input type="radio" name={`${id}-tipo`} value={v} checked={valor.tipo === v} onChange={() => onCambio({ ...valor, tipo: v })} />
                  <span>{t}</span>
                </label>
              ))}
            </div>
          </fieldset>
          <div className="campo">
            <label htmlFor={`${id}-desc`} className="etiqueta">Descripción</label>
            <textarea id={`${id}-desc`} rows={3} maxLength={MAX} value={valor.descripcion} placeholder="Ej.: rayón en la puerta izquierda"
              onChange={(e) => onCambio({ ...valor, descripcion: e.target.value })}
              aria-invalid={error ? "true" : "false"} aria-describedby={`${id}-desc-ayuda`} />
            <p id={`${id}-desc-ayuda`} className={error ? "ayuda ayuda--error" : "ayuda"} role={error ? "alert" : undefined}>
              {error || `${valor.descripcion.length} de ${MAX} caracteres.`}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

export const novedadVacia = { tipo: "DANIO", descripcion: "" };
