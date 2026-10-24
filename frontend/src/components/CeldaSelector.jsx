import Icono from "./Icono.jsx";

export default function CeldaSelector({ celdas, valor, onCambio, error }) {
  const libres = celdas.filter((c) => c.estado === "DISPONIBLE");
  return (
    <fieldset className="campo grupo" aria-describedby="celda-ayuda">
      <legend className="etiqueta">Celda libre</legend>
      {libres.length === 0 ? (
        <p className="vacio" id="celda-ayuda">No hay celdas libres. Registre una salida para liberar una.</p>
      ) : (
        <div className="celdas-elegir" role="radiogroup">
          {libres.map((c) => (
            <label key={c.id_celda} className="opcion">
              <input type="radio" name="celda" value={c.id_celda} checked={String(valor) === String(c.id_celda)}
                onChange={() => onCambio(c.id_celda)} />
              <span className="opcion__cuerpo">
                <Icono nombre={c.tipo === "MOTO" ? "moto" : "carro"} />
                <strong>{c.codigo}</strong>
                <small>{c.tipo === "MOTO" ? "Moto" : "Carro"}</small>
              </span>
            </label>
          ))}
        </div>
      )}
      <p id="celda-ayuda" className={error ? "ayuda ayuda--error" : "ayuda"} role={error ? "alert" : undefined}>
        {error || (libres.length ? "La celda debe ser del mismo tipo que el vehículo." : "")}
      </p>
    </fieldset>
  );
}
