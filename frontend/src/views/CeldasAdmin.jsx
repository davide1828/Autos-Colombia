import { useCallback, useEffect, useState } from "react";
import { api } from "../api.js";

export default function CeldasAdmin({ alActualizar }) {
  const [celdas, setCeldas] = useState([]);
  const [activos, setActivos] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState("");
  const [exito, setExito] = useState("");

  // Modal Nueva Celda
  const [modalCrear, setModalCrear] = useState(false);
  const [codigo, setCodigo] = useState("");
  const [tipo, setTipo] = useState("CARRO");
  const [guardando, setGuardando] = useState(false);
  const [errorModal, setErrorModal] = useState("");

  const cargar = useCallback(async () => {
    setCargando(true);
    try {
      const [c, a] = await Promise.all([api.celdas(), api.activos()]);
      setCeldas(c);
      setActivos(a);
      setError("");
    } catch (err) {
      setError(err.message || "Error al cargar las celdas.");
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  // Mapa de vehículos que ocupan celda
  const mapaOcupadas = {};
  activos.forEach((reg) => {
    mapaOcupadas[reg.id_celda] = `${reg.placa} (${reg.cliente})`;
  });

  // Métricas
  const total = celdas.length;
  const disponibles = celdas.filter((c) => c.estado === "DISPONIBLE").length;
  const ocupadas = celdas.filter((c) => c.estado === "OCUPADA").length;
  const inactivas = celdas.filter((c) => c.estado === "INACTIVA").length;

  async function handleToggleEstado(c) {
    if (c.estado === "OCUPADA") {
      setError("Regla RN-09: La celda está ocupada. Registre primero la salida del vehículo.");
      return;
    }
    const nuevoEstado = c.estado === "DISPONIBLE" ? "INACTIVA" : "DISPONIBLE";
    setError("");
    setExito("");
    try {
      await api.cambiarEstadoCelda(c.id_celda, nuevoEstado);
      setExito(`Celda ${c.codigo} cambiada a estado ${nuevoEstado}.`);
      await cargar();
      if (alActualizar) alActualizar();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCrearCelda(e) {
    e.preventDefault();
    setErrorModal("");
    const cod = codigo.trim().toUpperCase();
    if (!cod) {
      setErrorModal("Ingrese el código de la celda.");
      return;
    }

    setGuardando(true);
    try {
      await api.crearCelda({ codigo: cod, tipo });
      setModalCrear(false);
      setCodigo("");
      setTipo("CARRO");
      setExito(`Celda ${cod} registrada correctamente.`);
      await cargar();
      if (alActualizar) alActualizar();
    } catch (err) {
      setErrorModal(err.message);
    } finally {
      setGuardando(false);
    }
  }

  return (
    <div className="panel-admin">
      <header className="panel-admin__cabecera">
        <div>
          <h2>Gestión de Celdas y Bahías</h2>
          <p className="ayuda">Administre el inventario de celdas, asigne tipos e inhabilite bahías en obra.</p>
        </div>
        <button type="button" className="boton" onClick={() => setModalCrear(true)}>
          + Registrar Celda
        </button>
      </header>

      {error && <p className="mensaje-alerta mensaje-alerta--error" role="alert">{error}</p>}
      {exito && <p className="mensaje-alerta mensaje-alerta--exito" role="status">{exito}</p>}

      {/* Tarjetas KPI */}
      <div className="tarjetas-grid">
        <div className="tarjeta-kpi">
          <p className="tarjeta-kpi__titulo">Total Celdas</p>
          <p className="tarjeta-kpi__valor">{total}</p>
        </div>
        <div className="tarjeta-kpi tarjeta-kpi--exito">
          <p className="tarjeta-kpi__titulo">Disponibles</p>
          <p className="tarjeta-kpi__valor">{disponibles}</p>
        </div>
        <div className="tarjeta-kpi tarjeta-kpi--alerta">
          <p className="tarjeta-kpi__titulo">Ocupadas</p>
          <p className="tarjeta-kpi__valor">{ocupadas}</p>
        </div>
        <div className="tarjeta-kpi tarjeta-kpi--inactivo">
          <p className="tarjeta-kpi__titulo">Inactivas (Obras)</p>
          <p className="tarjeta-kpi__valor">{inactivas}</p>
        </div>
      </div>

      {/* Tabla de Celdas */}
      <div className="tabla-contenedor">
        <table className="tabla-admin">
          <thead>
            <tr>
              <th>CÓDIGO</th>
              <th>TIPO PERMITIDO</th>
              <th>ESTADO ACTUAL</th>
              <th>VEHÍCULO ALOJADO</th>
              <th>GESTIÓN DE ESTADO</th>
            </tr>
          </thead>
          <tbody>
            {cargando ? (
              <tr>
                <td colSpan="5" className="celda-centro">Cargando celdas…</td>
              </tr>
            ) : celdas.length === 0 ? (
              <tr>
                <td colSpan="5" className="celda-centro">No hay celdas registradas.</td>
              </tr>
            ) : (
              celdas.map((c) => {
                const ocupante = mapaOcupadas[c.id_celda];
                return (
                  <tr key={c.id_celda}>
                    <td><strong>{c.codigo}</strong></td>
                    <td>
                      <span className="chip-tipo">
                        {c.tipo === "CARRO" ? "🚗 CARRO" : "🏍️ MOTO"}
                      </span>
                    </td>
                    <td>
                      <span className={`badge badge--${c.estado.toLowerCase()}`}>
                        {c.estado}
                      </span>
                    </td>
                    <td>{ocupante ? <strong>{ocupante}</strong> : <span className="texto-atenuado">— (Libre)</span>}</td>
                    <td>
                      {c.estado === "OCUPADA" ? (
                        <span className="texto-protegido" title="Regla RN-09: No es posible inhabilitar celdas con vehículos dentro">
                          🔒 Ocupada
                        </span>
                      ) : (
                        <button
                          type="button"
                          className={`boton-link ${c.estado === "DISPONIBLE" ? "boton-link--peligro" : "boton-link--exito"}`}
                          onClick={() => handleToggleEstado(c)}
                        >
                          {c.estado === "DISPONIBLE" ? "Inhabilitar" : "Habilitar"}
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <div className="alerta-regla">
        <p><strong>Regla RN-09:</strong> No es posible inhabilitar ni modificar celdas que tengan un vehículo dentro. Registre primero la salida.</p>
      </div>

      {/* Modal Registrar Celda */}
      {modalCrear && (
        <div className="modal-fondo" role="dialog" aria-modal="true" aria-labelledby="modal-celda-titulo">
          <div className="modal-caja">
            <header className="modal-cabecera">
              <h3 id="modal-celda-titulo">Registrar Nueva Celda</h3>
              <button type="button" className="icono-boton" onClick={() => setModalCrear(false)}>✕</button>
            </header>
            <form onSubmit={handleCrearCelda} className="modal-cuerpo" noValidate>
              {errorModal && <p className="ayuda ayuda--error">{errorModal}</p>}
              <div className="campo">
                <label className="etiqueta">Código de Celda *</label>
                <input
                  type="text"
                  value={codigo}
                  onChange={(e) => setCodigo(e.target.value.toUpperCase())}
                  placeholder="ej. A-05 o M-03 (máx 10 car)"
                  maxLength={10}
                  autoFocus
                  required
                />
              </div>
              <div className="campo">
                <label className="etiqueta">Tipo de Vehículo Permitido *</label>
                <div className="opciones-radio">
                  <label>
                    <input
                      type="radio"
                      name="tipo"
                      value="CARRO"
                      checked={tipo === "CARRO"}
                      onChange={() => setTipo("CARRO")}
                    />
                    🚗 CARRO
                  </label>
                  <label>
                    <input
                      type="radio"
                      name="tipo"
                      value="MOTO"
                      checked={tipo === "MOTO"}
                      onChange={() => setTipo("MOTO")}
                    />
                    🏍️ MOTO
                  </label>
                </div>
              </div>
              <footer className="modal-acciones">
                <button type="button" className="boton boton--secundario" onClick={() => setModalCrear(false)} disabled={guardando}>
                  Cancelar
                </button>
                <button type="submit" className="boton" disabled={guardando}>
                  {guardando ? "Guardando…" : "Crear Celda"}
                </button>
              </footer>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

