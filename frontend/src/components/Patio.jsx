import { formatearDuracion, minutosDesde, resumenOcupacion } from "../lib/format.js";
import Icono from "./Icono.jsx";
import { MiniPlaca } from "./PlacaInput.jsx";

const ESTADOS = {
  DISPONIBLE: { clase: "libre", texto: "Libre", icono: "libre" },
  OCUPADA: { clase: "ocupada", texto: "Ocupada", icono: "ocupada" },
  INACTIVA: { clase: "fuera", texto: "Fuera de servicio", icono: "fuera" },
};

/** Tablero del patio: cada celda es una bahía; si está ocupada muestra la placa del vehículo. */
export default function Patio({ celdas, activos, ahora, cargando, error }) {
  const r = resumenOcupacion(celdas);
  const porCelda = new Map(activos.map((a) => [a.celda, a]));

  return (
    <section className="patio" aria-labelledby="patio-titulo">
      <div className="patio__cabecera">
        <h2 id="patio-titulo">Patio</h2>
        <p className="patio__resumen" aria-live="polite">
          {cargando && celdas.length === 0 ? "Cargando…" : <><strong>{r.ocupadas}</strong> de {r.enServicio} celdas ocupadas</>}
        </p>
      </div>
      {error && <p className="ayuda ayuda--error" role="alert">{error}</p>}
      {!cargando && celdas.length === 0 && !error && (
        <p className="vacio">Todavía no hay celdas registradas en el sistema.</p>
      )}
      <ul className="bahias">
        {celdas.map((c) => {
          const e = ESTADOS[c.estado] || ESTADOS.INACTIVA;
          const dentro = porCelda.get(c.codigo);
          return (
            <li key={c.id_celda} className={`bahia bahia--${e.clase}`}>
              <div className="bahia__fila">
                <span className="bahia__codigo">{c.codigo}</span>
                <Icono nombre={c.tipo === "MOTO" ? "moto" : "carro"} tamano={18} />
              </div>
              {dentro ? (
                <>
                  <MiniPlaca placa={dentro.placa} />
                  <span className="bahia__detalle">{formatearDuracion(minutosDesde(dentro.fecha_entrada, ahora))} · {dentro.cliente}</span>
                </>
              ) : (
                <span className="bahia__estado"><Icono nombre={e.icono} tamano={14} />{e.texto}</span>
              )}
            </li>
          );
        })}
      </ul>
      {celdas.length > 0 && activos.length === 0 && (
        <p className="vacio">No hay vehículos dentro. Cuando registre un ingreso aparecerá aquí con su placa.</p>
      )}
    </section>
  );
}
