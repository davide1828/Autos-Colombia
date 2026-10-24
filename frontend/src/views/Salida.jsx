import { useState } from "react";
import { api } from "../api.js";
import NovedadCampo, { novedadVacia } from "../components/NovedadCampo.jsx";
import PlacaInput, { MiniPlaca } from "../components/PlacaInput.jsx";
import Rapidos from "../components/Rapidos.jsx";
import Resultado from "../components/Resultado.jsx";
import { formatearDuracion, validarPlaca } from "../lib/format.js";

export default function Salida({ activos, ahora, alRegistrar }) {
  const [placa, setPlaca] = useState("");
  const [abierta, setAbierta] = useState(false);
  const [nov, setNov] = useState(novedadVacia);
  const [errores, setErrores] = useState({});
  const [resultado, setResultado] = useState(null);
  const [enviando, setEnviando] = useState(false);

  async function enviar(e) {
    e.preventDefault();
    const err = { placa: validarPlaca(placa), nov: abierta && !nov.descripcion.trim() ? "Describa la novedad o quítela." : "" };
    setErrores(err);
    if (err.placa || err.nov) return;
    setEnviando(true); setResultado(null);
    try {
      const r = await api.salida({ placa, novedad: abierta ? nov : undefined });
      setResultado({ tipo: "ok", titulo: "Salida registrada", cuerpo: <><MiniPlaca placa={r.placa} /> estuvo {formatearDuracion(r.minutos)}. La celda <strong>{r.celda}</strong> ya está libre.</> });
      setPlaca(""); setAbierta(false); setNov(novedadVacia);
      alRegistrar();
    } catch (ex) {
      setResultado({ tipo: "error", titulo: "No se registró la salida", cuerpo: ex.message });
    } finally { setEnviando(false); }
  }

  return (
    <form className="panel" onSubmit={enviar} noValidate aria-labelledby="t-salida">
      <h2 id="t-salida">Registrar salida</h2>
      {resultado && <Resultado tipo={resultado.tipo} titulo={resultado.titulo} onCerrar={() => setResultado(null)}>{resultado.cuerpo}</Resultado>}
      <PlacaInput id="salida-placa" valor={placa} onCambio={setPlaca} error={errores.placa} />
      <Rapidos activos={activos} ahora={ahora} seleccionada={placa} onElegir={setPlaca} />
      <NovedadCampo id="salida-nov" abierta={abierta} onAbrir={setAbierta} valor={nov} onCambio={setNov} error={errores.nov} />
      <button type="submit" className="boton" disabled={enviando}>{enviando ? "Registrando…" : "Registrar salida"}</button>
    </form>
  );
}
