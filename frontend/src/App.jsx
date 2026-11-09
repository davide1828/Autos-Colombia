import { useCallback, useEffect, useState } from "react";
import { api, configurar, sesion } from "./api.js";
import Icono from "./components/Icono.jsx";
import Login from "./components/Login.jsx";
import Patio from "./components/Patio.jsx";
import CambioClaveModal from "./components/CambioClaveModal.jsx";
import { resumenOcupacion } from "./lib/format.js";
import Ingreso from "./views/Ingreso.jsx";
import Novedad from "./views/Novedad.jsx";
import Salida from "./views/Salida.jsx";
import UsuariosAdmin from "./views/UsuariosAdmin.jsx";
import CeldasAdmin from "./views/CeldasAdmin.jsx";
import AuditoriaAdmin from "./views/AuditoriaAdmin.jsx";

const REFRESCO_MS = 30_000;

function useConexion() {
  const [enLinea, setEnLinea] = useState(navigator.onLine);
  useEffect(() => {
    const a = () => setEnLinea(true), b = () => setEnLinea(false);
    window.addEventListener("online", a);
    window.addEventListener("offline", b);
    return () => {
      window.removeEventListener("online", a);
      window.removeEventListener("offline", b);
    };
  }, []);
  return enLinea;
}

function Tablero({ usuario, sinSesion }) {
  const esAdmin = usuario?.rol === "ADMINISTRADOR";
  const [vista, setVista] = useState(esAdmin ? "usuarios" : "ingreso");
  const [celdas, setCeldas] = useState([]);
  const [activos, setActivos] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState("");
  const [ahora, setAhora] = useState(() => Date.now());
  const [modalClave, setModalClave] = useState(false);
  const enLinea = useConexion();

  const refrescar = useCallback(async () => {
    try {
      const [c, a] = await Promise.all([api.celdas(), api.activos()]);
      setCeldas(c);
      setActivos(a);
      setError("");
    } catch (e) {
      setError(e.message);
    } finally {
      setCargando(false);
      setAhora(Date.now());
    }
  }, []);

  useEffect(() => {
    refrescar();
    const t = setInterval(() => {
      if (!document.hidden) refrescar();
    }, REFRESCO_MS);
    const v = () => {
      if (!document.hidden) refrescar();
    };
    document.addEventListener("visibilitychange", v);
    return () => {
      clearInterval(t);
      document.removeEventListener("visibilitychange", v);
    };
  }, [refrescar]);

  // Lista de vistas según el rol (RBAC)
  const vistasDisponibles = esAdmin
    ? [
        { id: "usuarios", texto: "Usuarios", icono: "entrar" },
        { id: "celdasAdmin", texto: "Celdas", icono: "patio" },
        { id: "auditoria", texto: "Auditoría", icono: "novedad" },
        { id: "ingreso", texto: "Ingreso", icono: "entrar" },
        { id: "salida", texto: "Salida", icono: "salir" },
        { id: "novedad", texto: "Novedad", icono: "novedad" },
        { id: "patio", texto: "Patio", icono: "patio" },
      ]
    : [
        { id: "ingreso", texto: "Ingreso", icono: "entrar" },
        { id: "salida", texto: "Salida", icono: "salir" },
        { id: "novedad", texto: "Novedad", icono: "novedad" },
        { id: "patio", texto: "Patio", icono: "patio", soloMovil: true },
      ];

  // Con pantalla ancha en modo encargado, el patio está siempre al lado
  useEffect(() => {
    if (!esAdmin) {
      const m = window.matchMedia("(min-width: 720px)");
      const ajustar = () => {
        if (m.matches) setVista((v) => (v === "patio" ? "ingreso" : v));
      };
      ajustar();
      m.addEventListener("change", ajustar);
      return () => m.removeEventListener("change", ajustar);
    }
  }, [esAdmin]);

  const r = resumenOcupacion(celdas);
  const patio = <Patio celdas={celdas} activos={activos} ahora={ahora} cargando={cargando} error={error} />;

  const esVistaAdmin = ["usuarios", "celdasAdmin", "auditoria"].includes(vista);

  return (
    <div className="app">
      <a className="saltar" href="#principal">
        Saltar al contenido
      </a>
      <header className="barra">
        <div className="barra__inicio">
          <p className="marca">Autos Colombia</p>
          <span className="barra__rol-tag">{usuario?.rol || "OPERADOR"}</span>
        </div>

        <p className="barra__estado" aria-label={`${r.ocupadas} de ${r.enServicio} celdas ocupadas`}>
          <span className="barra__cifra">{r.ocupadas}/{r.enServicio}</span> ocupadas
        </p>

        <div className="barra__usuario">
          <span className="barra__usuario-nombre" title={usuario?.nombre}>
            {usuario?.nombre || usuario?.usuario}
          </span>
          <button
            type="button"
            className="icono-boton icono-boton--claro"
            onClick={() => setModalClave(true)}
            title="Cambiar mi contraseña"
            aria-label="Cambiar mi contraseña"
          >
            🔑
          </button>
          <button
            type="button"
            className="icono-boton icono-boton--claro"
            onClick={sinSesion}
            title="Cerrar sesión"
            aria-label="Cerrar sesión"
          >
            <Icono nombre="salirSesion" />
          </button>
        </div>
      </header>

      <nav className="nav" aria-label="Secciones">
        <ul>
          {vistasDisponibles.map((v) => (
            <li key={v.id} className={v.soloMovil ? "solo-movil" : undefined}>
              <button
                type="button"
                className="nav__item"
                aria-current={vista === v.id ? "page" : undefined}
                onClick={() => setVista(v.id)}
              >
                <Icono nombre={v.icono} tamano={22} />
                <span>{v.texto}</span>
              </button>
            </li>
          ))}
        </ul>
        <button type="button" className="nav__salir" onClick={sinSesion}>
          <Icono nombre="salirSesion" tamano={20} />
          <span>Cerrar sesión</span>
        </button>
      </nav>

      <main id="principal" className="contenido">
        {!enLinea && (
          <p className="aviso" role="status">
            <Icono nombre="sinConexion" tamano={18} /> Sin conexión. Los registros se guardan cuando vuelva la red.
          </p>
        )}

        {esVistaAdmin ? (
          <div className="contenido__admin">
            {vista === "usuarios" && <UsuariosAdmin usuarioActual={usuario?.usuario} />}
            {vista === "celdasAdmin" && <CeldasAdmin alActualizar={refrescar} />}
            {vista === "auditoria" && <AuditoriaAdmin />}
          </div>
        ) : (
          <div className="contenido__rejilla">
            <div className={`columna columna--accion${vista === "patio" ? " oculta-movil" : ""}`}>
              {vista === "ingreso" && <Ingreso celdas={celdas} alRegistrar={refrescar} />}
              {vista === "salida" && <Salida activos={activos} ahora={ahora} alRegistrar={refrescar} />}
              {vista === "novedad" && <Novedad activos={activos} ahora={ahora} />}
              {vista === "patio" && esAdmin && patio}
            </div>
            <div className={`columna columna--patio${vista === "patio" ? "" : " oculta-movil"}`}>
              {patio}
            </div>
          </div>
        )}
      </main>

      <CambioClaveModal
        abierto={modalClave}
        onCerrar={() => setModalClave(false)}
        usuarioActual={usuario?.usuario}
      />
    </div>
  );
}

export default function App() {
  const [autenticado, setAutenticado] = useState(Boolean(sesion.leer()));
  const [usuario, setUsuario] = useState(() => sesion.obtenerUsuario());
  const [aviso, setAviso] = useState("");

  const salir = useCallback((mensaje = "") => {
    sesion.borrar();
    setUsuario(null);
    setAviso(mensaje);
    setAutenticado(false);
  }, []);

  useEffect(() => {
    configurar({
      onNoAutorizado: () => salir("La sesión expiró o fue revocada. Inicie sesión de nuevo."),
    });
  }, [salir]);

  return autenticado ? (
    <Tablero usuario={usuario} sinSesion={() => salir()} />
  ) : (
    <Login
      aviso={aviso}
      onEntrar={(u) => {
        setAviso("");
        setUsuario(u);
        setAutenticado(true);
      }}
    />
  );
}
