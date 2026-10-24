import { formatearDuracion, minutosDesde } from "../lib/format.js";
import { MiniPlaca } from "./PlacaInput.jsx";

/** Atajo: vehículos que están dentro, para llenar la placa con un toque. */
export default function Rapidos({ activos, ahora, seleccionada, onElegir }) {
  if (activos.length === 0) return null;
  return (
    <div className="rapidos">
      <p className="etiqueta" id="rapidos-t">Vehículos dentro</p>
      <ul aria-labelledby="rapidos-t">
        {activos.map((a) => (
          <li key={a.id_registro}>
            <button type="button" className="rapido" aria-pressed={seleccionada === a.placa} onClick={() => onElegir(a.placa)}>
              <MiniPlaca placa={a.placa} />
              <small>{a.celda} · {formatearDuracion(minutosDesde(a.fecha_entrada, ahora))}</small>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
