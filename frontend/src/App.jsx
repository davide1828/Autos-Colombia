import { useCallback, useEffect, useState } from "react";
import { api, configurar, sesion } from "./api.js";
import Icono from "./components/Icono.jsx";
import Login from "./components/Login.jsx";
import Patio from "./components/Patio.jsx";
import { resumenOcupacion } from "./lib/format.js";
import Ingreso from "./views/Ingreso.jsx";
import Novedad from "./views/Novedad.jsx";
import Salida from "./views/Salida.jsx";

const VISTAS = [
  { id: "ingreso", texto: "Ingreso", icono: "entrar" },
  { id: "salida", texto: "Salida", icono: "salir" },
  { id: "novedad", texto: "Novedad", icono: "novedad" },
  { id: "patio", texto: "Patio", icono: "patio", soloMovil: true },
];
const REFRESCO_MS = 30_000;

function useConexion() {
  const [enLinea, setEnLinea] = useState(navigator.onLine);
  useEffect(() => {
    const a = () => setEnLinea(true), b = () => setEnLinea(false);
    window.addEventListener("online", a); window.addEventListener("offline", b);
    return () => { window.removeEventListener("online", a); window.removeEventListener("offline", b); };
  }, []);
  return enLinea;
}

function Tablero({ sinSesion }) {
  const [vista, setVista] = useState("ingreso");
  const [celdas, setCeldas] = useState([]);
  const [activos, setActivos] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState("");
  const [ahora, setAhora] = useState(() => Date.now());
  const enLinea = useConexion();

  const refrescar = useCallback(async () => {
    try {
      const [c, a] = await Promise.all([api.celdas(), api.activos()]);
      setCeldas(c); setActivos(a); setError("");
    } catch (e) { setError(e.message); }
    finally { setCargando(false); setAhora(Date.now()); }
  }, []);

  useEffect(() => {
    refrescar();
    const t = setInterval(() => { if (!document.hidden) refrescar(); }, REFRESCO_MS);
    const v = () => { if (!document.hidden) refrescar(); };
    document.addEventListener("visibilitychange", v);
    return () => { clearInterval(t); document.removeEventListener("visibilitychange", v); };
  }, [refrescar]);

  // Con pantalla ancha el patio siempre está visible junto al formulario: no existe la pestaña "Patio".
  useEffect(() => {
    const m = window.matchMedia("(min-width: 720px)");
    const ajustar = () => { if (m.matches) setVista((v) => (v === "patio" ? "ingreso" : v)); };
    ajustar();
    m.addEventListener("change", ajustar);
    return () => m.removeEventListener("change", ajustar);
  }, []);

  const r = resumenOcupacion(celdas);
  const patio = <Patio celdas={celdas} activos={activos} ahora={ahora} cargando={cargando} error={error} />;

  return (
    <div className="app">
      <a className="saltar" href="#principal">Saltar al contenido</a>
      <header className="barra">
        <p className="marca">Autos Colombia</p>
        <p className="barra__estado" aria-label={`${r.ocupadas} de ${r.enServicio} celdas ocupadas`}>
          <span className="barra__cifra">{r.ocupadas}/{r.enServicio}</span> ocupadas
        </p>
        <button type="button" className="icono-boton icono-boton--claro" onClick={sinSesion} aria-label="Cerrar sesión">
          <Icono nombre="salirSesion" />
        </button>
      </header>

      <nav className="nav" aria-label="Secciones">
        <ul>
          {VISTAS.map((v) => (
            <li key={v.id} className={v.soloMovil ? "solo-movil" : undefined}>
              <button type="button" className="nav__item" aria-current={vista === v.id ? "page" : undefined} onClick={() => setVista(v.id)}>
                <Icono nombre={v.icono} tamano={22} /><span>{v.texto}</span>
              </button>
            </li>
          ))}
        </ul>
        <button type="button" className="nav__salir" onClick={sinSesion}><Icono nombre="salirSesion" tamano={20} /><span>Cerrar sesión</span></button>
      </nav>

      <main id="principal" className="contenido">
        {!enLinea && <p className="aviso" role="status"><Icono nombre="sinConexion" tamano={18} /> Sin conexión. Los registros se guardan cuando vuelva la red.</p>}
        <div className="contenido__rejilla">
          <div className={`columna columna--accion${vista === "patio" ? " oculta-movil" : ""}`}>
            {vista === "ingreso" && <Ingreso celdas={celdas} alRegistrar={refrescar} />}
            {vista === "salida" && <Salida activos={activos} ahora={ahora} alRegistrar={refrescar} />}
            {vista === "novedad" && <Novedad activos={activos} ahora={ahora} />}
          </div>
          <div className={`columna columna--patio${vista === "patio" ? "" : " oculta-movil"}`}>{patio}</div>
        </div>
      </main>
    </div>
  );
}

export default function App() {
  const [autenticado, setAutenticado] = useState(Boolean(sesion.leer()));
  const [aviso, setAviso] = useState("");

  const salir = useCallback((mensaje = "") => { sesion.borrar(); setAviso(mensaje); setAutenticado(false); }, []);
  useEffect(() => { configurar({ onNoAutorizado: () => salir("La sesión terminó. Escriba la clave de nuevo.") }); }, [salir]);

  return autenticado
    ? <Tablero sinSesion={() => salir()} />
    : <Login aviso={aviso} onEntrar={() => { setAviso(""); setAutenticado(true); }} />;
}
