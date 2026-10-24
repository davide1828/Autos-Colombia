import { normalizarPlaca } from "../lib/format.js";

/** Campo de placa con el aspecto de una placa colombiana: es el elemento central del formulario. */
export default function PlacaInput({ id, etiqueta = "Placa del vehículo", valor, onCambio, error, autoFoco = false }) {
  const idAyuda = `${id}-ayuda`;
  return (
    <div className="campo">
      <label htmlFor={id} className="etiqueta">{etiqueta}</label>
      <div className={`placa${error ? " placa--error" : ""}`}>
        <span className="placa__pais" aria-hidden="true">COLOMBIA</span>
        <input
          id={id} className="placa__input" value={valor} placeholder="ABC123" maxLength={7}
          onChange={(e) => onCambio(normalizarPlaca(e.target.value))}
          autoFocus={autoFoco} autoCapitalize="characters" autoComplete="off" autoCorrect="off" spellCheck={false}
          aria-invalid={error ? "true" : "false"} aria-describedby={idAyuda}
        />
      </div>
      <p id={idAyuda} className={error ? "ayuda ayuda--error" : "ayuda"} role={error ? "alert" : undefined}>
        {error || "Ejemplo: ABC123 (carro) o ABC12D (moto)."}
      </p>
    </div>
  );
}

/** Placa en miniatura para listas y celdas. */
export function MiniPlaca({ placa }) {
  return <span className="miniplaca">{placa}</span>;
}
