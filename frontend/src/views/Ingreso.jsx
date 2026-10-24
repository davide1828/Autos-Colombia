import { useState } from "react";
import { api } from "../api.js";
import CeldaSelector from "../components/CeldaSelector.jsx";
import NovedadCampo, { novedadVacia } from "../components/NovedadCampo.jsx";
import PlacaInput, { MiniPlaca } from "../components/PlacaInput.jsx";
import Resultado from "../components/Resultado.jsx";
import { horaCorta, validarPlaca } from "../lib/format.js";

export default function Ingreso({ celdas, alRegistrar }) {
  const [placa, setPlaca] = useState("");
  const [celda, setCelda] = useState("");
  const [abierta, setAbierta] = useState(false);
  const [nov, setNov] = useState(novedadVacia);
  const [errores, setErrores] = useState({});
  const [resultado, setResultado] = useState(null);
  const [enviando, setEnviando] = useState(false);

  async function enviar(e) {
    e.preventDefault();
    const err = {
      placa: validarPlaca(placa),
      celda: celda ? "" : "Elija una celda libre.",
      nov: abierta && !nov.descripcion.trim() ? "Describa la novedad o quítela." : "",
    };
    setErrores(err);
    if (err.placa || err.celda || err.nov) return;
    setEnviando(true); setResultado(null);
    try {
      const r = await api.ingreso({ placa, id_celda: Number(celda), novedad: abierta ? nov : undefined });
      setResultado({ tipo: "ok", titulo: "Ingreso registrado", cuerpo: <><MiniPlaca placa={r.placa} /> quedó en la celda <strong>{r.celda}</strong> a las {horaCorta(r.fecha_entrada)}</> });
      setPlaca(""); setCelda(""); setAbierta(false); setNov(novedadVacia);
      alRegistrar();
    } catch (ex) {
      setResultado({ tipo: "error", titulo: "No se registró el ingreso", cuerpo: ex.message });
      if (ex.estado === 409 || ex.estado === 422) alRegistrar(); // el estado del patio pudo cambiar
    } finally { setEnviando(false); }
  }

  return (
    <form className="panel" onSubmit={enviar} noValidate aria-labelledby="t-ingreso">
      <h2 id="t-ingreso">Registrar ingreso</h2>
      {resultado && <Resultado tipo={resultado.tipo} titulo={resultado.titulo} onCerrar={() => setResultado(null)}>{resultado.cuerpo}</Resultado>}
      <PlacaInput id="ingreso-placa" valor={placa} onCambio={setPlaca} error={errores.placa} />
      <CeldaSelector celdas={celdas} valor={celda} onCambio={setCelda} error={errores.celda} />
      <NovedadCampo id="ingreso-nov" abierta={abierta} onAbrir={setAbierta} valor={nov} onCambio={setNov} error={errores.nov} />
      <button type="submit" className="boton" disabled={enviando}>{enviando ? "Registrando…" : "Registrar ingreso"}</button>
    </form>
  );
}
