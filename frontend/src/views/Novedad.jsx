import { useState } from "react";
import { api } from "../api.js";
import NovedadCampo, { novedadVacia } from "../components/NovedadCampo.jsx";
import PlacaInput, { MiniPlaca } from "../components/PlacaInput.jsx";
import Rapidos from "../components/Rapidos.jsx";
import Resultado from "../components/Resultado.jsx";
import { validarPlaca } from "../lib/format.js";

export default function Novedad({ activos, ahora }) {
  const [placa, setPlaca] = useState("");
  const [nov, setNov] = useState(novedadVacia);
  const [errores, setErrores] = useState({});
  const [resultado, setResultado] = useState(null);
  const [enviando, setEnviando] = useState(false);

  async function enviar(e) {
    e.preventDefault();
    const err = { placa: validarPlaca(placa), nov: nov.descripcion.trim() ? "" : "Describa lo que ocurrió." };
    setErrores(err);
    if (err.placa || err.nov) return;
    setEnviando(true); setResultado(null);
    try {
      await api.novedad({ placa, tipo: nov.tipo, descripcion: nov.descripcion.trim() });
      setResultado({ tipo: "ok", titulo: "Novedad guardada", cuerpo: <>Quedó registrada para <MiniPlaca placa={placa} />.</> });
      setPlaca(""); setNov(novedadVacia);
    } catch (ex) {
      setResultado({ tipo: "error", titulo: "No se guardó la novedad", cuerpo: ex.message });
    } finally { setEnviando(false); }
  }

  return (
    <form className="panel" onSubmit={enviar} noValidate aria-labelledby="t-novedad">
      <h2 id="t-novedad">Registrar novedad</h2>
      {resultado && <Resultado tipo={resultado.tipo} titulo={resultado.titulo} onCerrar={() => setResultado(null)}>{resultado.cuerpo}</Resultado>}
      <PlacaInput id="novedad-placa" etiqueta="Placa del vehículo dentro" valor={placa} onCambio={setPlaca} error={errores.placa} />
      <Rapidos activos={activos} ahora={ahora} seleccionada={placa} onElegir={setPlaca} />
      <NovedadCampo id="novedad" obligatoria valor={nov} onCambio={setNov} error={errores.nov} />
      <button type="submit" className="boton" disabled={enviando}>{enviando ? "Guardando…" : "Guardar novedad"}</button>
    </form>
  );
}
