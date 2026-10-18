import { useCallback, useEffect, useState } from "react";
import { api, setToken } from "./api.js";

const TIPOS = [
  { v: "DANIO", t: "Daño" },
  { v: "FALTANTE", t: "Faltante" },
  { v: "OTRA", t: "Otra" },
];

function Mensaje({ m }) {
  if (!m) return null;
  return <p role="status" className={`msg ${m.ok ? "ok" : "error"}`}>{m.texto}</p>;
}

function CampoNovedad({ valor, onChange }) {
  return (
    <fieldset>
      <legend>Novedad (opcional)</legend>
      <select value={valor.tipo} onChange={(e) => onChange({ ...valor, tipo: e.target.value })}>
        {TIPOS.map((t) => <option key={t.v} value={t.v}>{t.t}</option>)}
      </select>
      <textarea
        rows={2}
        maxLength={500}
        placeholder="Ej.: rayón en la puerta izquierda"
        value={valor.descripcion}
        onChange={(e) => onChange({ ...valor, descripcion: e.target.value })}
      />
    </fieldset>
  );
}

const novedadVacia = { tipo: "DANIO", descripcion: "" };
const aEnvio = (n) => (n.descripcion.trim() ? n : undefined);

function Ingreso({ celdas, onCambio }) {
  const [placa, setPlaca] = useState("");
  const [celda, setCelda] = useState("");
  const [nov, setNov] = useState(novedadVacia);
  const [msg, setMsg] = useState(null);

  async function enviar(e) {
    e.preventDefault();
    try {
      await api.ingreso({ placa, id_celda: Number(celda), novedad: aEnvio(nov) });
      setMsg({ ok: true, texto: `Ingreso registrado para ${placa.toUpperCase()}.` });
      setPlaca(""); setCelda(""); setNov(novedadVacia); onCambio();
    } catch (err) { setMsg({ ok: false, texto: err.message }); }
  }

  return (
    <form onSubmit={enviar}>
      <h2>Registrar ingreso</h2>
      <label>Placa
        <input value={placa} onChange={(e) => setPlaca(e.target.value)} placeholder="ABC123" maxLength={8} required autoCapitalize="characters" />
      </label>
      <label>Celda disponible
        <select value={celda} onChange={(e) => setCelda(e.target.value)} required>
          <option value="">Seleccione…</option>
          {celdas.map((c) => <option key={c.id_celda} value={c.id_celda}>{c.codigo} · {c.tipo === "CARRO" ? "Carro" : "Moto"}</option>)}
        </select>
      </label>
      <CampoNovedad valor={nov} onChange={setNov} />
      <button type="submit">Confirmar ingreso</button>
      <Mensaje m={msg} />
    </form>
  );
}

function Salida({ onCambio }) {
  const [placa, setPlaca] = useState("");
  const [nov, setNov] = useState(novedadVacia);
  const [msg, setMsg] = useState(null);

  async function enviar(e) {
    e.preventDefault();
    try {
      const r = await api.salida({ placa, novedad: aEnvio(nov) });
      setMsg({ ok: true, texto: `Salida registrada. Permanencia: ${r.minutos} min.` });
      setPlaca(""); setNov(novedadVacia); onCambio();
    } catch (err) { setMsg({ ok: false, texto: err.message }); }
  }

  return (
    <form onSubmit={enviar}>
      <h2>Registrar salida</h2>
      <label>Placa
        <input value={placa} onChange={(e) => setPlaca(e.target.value)} placeholder="ABC123" maxLength={8} required autoCapitalize="characters" />
      </label>
      <CampoNovedad valor={nov} onChange={setNov} />
      <button type="submit">Confirmar salida</button>
      <Mensaje m={msg} />
    </form>
  );
}

function Novedad() {
  const [placa, setPlaca] = useState("");
  const [nov, setNov] = useState(novedadVacia);
  const [msg, setMsg] = useState(null);

  async function enviar(e) {
    e.preventDefault();
    try {
      await api.novedad({ placa, ...nov });
      setMsg({ ok: true, texto: "Novedad registrada." });
      setPlaca(""); setNov(novedadVacia);
    } catch (err) { setMsg({ ok: false, texto: err.message }); }
  }

  return (
    <form onSubmit={enviar}>
      <h2>Registrar novedad</h2>
      <label>Placa del vehículo dentro
        <input value={placa} onChange={(e) => setPlaca(e.target.value)} placeholder="ABC123" maxLength={8} required autoCapitalize="characters" />
      </label>
      <CampoNovedad valor={nov} onChange={setNov} />
      <button type="submit">Guardar novedad</button>
      <Mensaje m={msg} />
    </form>
  );
}

function Estado({ activos, todas }) {
  const hora = (s) => new Date(s).toLocaleString("es-CO", { dateStyle: "short", timeStyle: "short" });
  return (
    <section>
      <h2>Vehículos dentro ({activos.length})</h2>
      {activos.length === 0 ? <p className="vacio">No hay vehículos dentro.</p> : (
        <ul className="tarjetas">
          {activos.map((a) => (
            <li key={a.id_registro}>
              <span><strong>{a.placa}</strong> · celda {a.celda}</span>
              <small>{a.cliente} · desde {hora(a.fecha_entrada)}</small>
            </li>
          ))}
        </ul>
      )}
      <h2>Celdas</h2>
      <ul className="celdas">
        {todas.map((c) => <li key={c.id_celda} className={c.estado.toLowerCase()}>{c.codigo}<small>{c.estado.toLowerCase()}</small></li>)}
      </ul>
    </section>
  );
}

export default function App() {
  const [vista, setVista] = useState("ingreso");
  const [todas, setTodas] = useState([]);
  const [activos, setActivos] = useState([]);
  const [error, setError] = useState("");
  const [token, setTok] = useState(sessionStorage.getItem("token") || "");

  const cargar = useCallback(async () => {
    try {
      const [c, a] = await Promise.all([api.celdas(), api.activos()]);
      setTodas(c); setActivos(a); setError("");
    } catch (e) { setError(e.message); }
  }, []);

  useEffect(() => { cargar(); }, [cargar, token]);

  const disponibles = todas.filter((c) => c.estado === "DISPONIBLE");
  const vistas = [["ingreso", "Ingreso"], ["salida", "Salida"], ["novedad", "Novedad"], ["estado", "Estado"]];

  return (
    <div className="app">
      <header>
        <h1>Autos Colombia</h1>
        <input
          className="token" type="password" placeholder="Clave de acceso" value={token}
          onChange={(e) => { setTok(e.target.value); setToken(e.target.value); }}
          aria-label="Clave de acceso"
        />
      </header>
      {error && <p className="msg error" role="alert">{error}</p>}
      <main>
        {vista === "ingreso" && <Ingreso celdas={disponibles} onCambio={cargar} />}
        {vista === "salida" && <Salida onCambio={cargar} />}
        {vista === "novedad" && <Novedad />}
        {vista === "estado" && <Estado activos={activos} todas={todas} />}
      </main>
      <nav>
        {vistas.map(([k, t]) => (
          <button key={k} className={vista === k ? "activa" : ""} onClick={() => { setVista(k); cargar(); }}>{t}</button>
        ))}
      </nav>
    </div>
  );
}
