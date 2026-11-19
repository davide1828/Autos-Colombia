import { useEffect, useState } from "react";
import { api } from "../api.js";

function mesActual() {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "America/Bogota", year: "numeric", month: "2-digit" }).format(new Date());
}

function fechaCorta(valor) {
  return String(valor).slice(0, 10);
}

function pesos(valor) {
  return new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 }).format(valor);
}

export default function PagosAdmin({ esAdmin }) {
  const [filtro, setFiltro] = useState("");
  const [clientes, setClientes] = useState([]);
  const [cliente, setCliente] = useState(null);
  const [pagos, setPagos] = useState([]);
  const [periodo, setPeriodo] = useState(mesActual);
  const [monto, setMonto] = useState("");
  const [medio, setMedio] = useState("EFECTIVO");
  const [cargando, setCargando] = useState(false);
  const [guardando, setGuardando] = useState(false);
  const [error, setError] = useState("");
  const [exito, setExito] = useState("");

  useEffect(() => {
    let vigente = true;
    const timer = setTimeout(async () => {
      try {
        const resultado = await api.clientes(filtro.trim());
        if (vigente) setClientes(resultado);
      } catch (e) {
        if (vigente) setError(e.message);
      }
    }, 250);
    return () => {
      vigente = false;
      clearTimeout(timer);
    };
  }, [filtro]);

  async function seleccionarCliente(c) {
    setCliente(c);
    setError("");
    setExito("");
    setCargando(true);
    try {
      const [detalle, historial] = await Promise.all([api.cliente(c.id_cliente), api.historialPagos(c.id_cliente)]);
      setCliente(detalle);
      setPagos(historial);
    } catch (e) {
      setError(e.message);
      setPagos([]);
    } finally {
      setCargando(false);
    }
  }

  async function registrar(e) {
    e.preventDefault();
    setError("");
    setExito("");
    if (!cliente || !periodo || !Number.isSafeInteger(Number(monto)) || Number(monto) < 1) {
      setError("Seleccione un cliente, un mes válido y un valor entero mayor que cero.");
      return;
    }
    setGuardando(true);
    try {
      const pago = await api.registrarPago({
        id_cliente: cliente.id_cliente,
        monto: Number(monto),
        periodo,
        medio_pago: medio,
      });
      setExito(`Pago ${pago.id_pago} registrado. El periodo mensual quedó vigente.`);
      setMonto("");
      await seleccionarCliente(cliente);
    } catch (e) {
      setError(e.message);
    } finally {
      setGuardando(false);
    }
  }

  async function descargarRecibo(pago) {
    setError("");
    try {
      const blob = await api.reciboPago(pago.id_pago);
      const url = URL.createObjectURL(blob);
      const enlace = document.createElement("a");
      enlace.href = url;
      enlace.download = `Autos-Colombia-recibo-${pago.id_pago}.pdf`;
      enlace.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (e) {
      setError(e.message);
    }
  }

  async function cancelar(pago) {
    if (!window.confirm(`¿Cancelar el pago ${pago.id_pago}? También se revocará la mensualidad asociada.`)) return;
    setError("");
    setExito("");
    try {
      await api.cancelarPago(pago.id_pago);
      setExito("Pago cancelado y mensualidad revocada.");
      await seleccionarCliente(cliente);
    } catch (e) {
      setError(e.message);
    }
  }

  return (
    <section className="panel-admin" aria-labelledby="pagos-titulo">
      <header className="panel-admin__cabecera">
        <div>
          <h2 id="pagos-titulo">Gestión de pagos</h2>
          <p className="ayuda">Registre mensualidades, consulte el historial y descargue recibos internos.</p>
        </div>
      </header>
      <p className="alerta-regla">
        El recibo PDF es un comprobante interno con sello SHA-256; no es una factura electrónica validada por la DIAN.
        Un pago registra la mensualidad del mes seleccionado. Un mes vencido bloquea nuevos ingresos hasta registrar un pago vigente.
      </p>
      {error && <p className="mensaje-alerta mensaje-alerta--error" role="alert">{error}</p>}
      {exito && <p className="mensaje-alerta mensaje-alerta--exito" role="status">{exito}</p>}

      <div className="pago-layout">
        <section className="pago-seccion" aria-labelledby="cliente-pago-titulo">
          <h3 id="cliente-pago-titulo">1. Seleccionar cliente</h3>
          <label className="campo">
            <span className="etiqueta">Buscar por documento o nombre</span>
            <input className="campo-busqueda" value={filtro} onChange={(e) => setFiltro(e.target.value)} />
          </label>
          <ul className="lista-clientes">
            {clientes.map((c) => (
              <li key={c.id_cliente}>
                <button
                  type="button"
                  className="lista-clientes__item"
                  aria-pressed={cliente?.id_cliente === c.id_cliente}
                  onClick={() => seleccionarCliente(c)}
                >
                  <strong>{c.nombre}</strong>
                  <span>{c.documento}</span>
                </button>
              </li>
            ))}
            {!clientes.length && <li className="vacio">No se encontraron clientes.</li>}
          </ul>
          {cliente && (
            <p className="ayuda">
              Seleccionado: <strong>{cliente.nombre}</strong> · {cliente.documento}
            </p>
          )}
        </section>

        <form className="pago-seccion" onSubmit={registrar}>
          <h3>2. Registrar mensualidad</h3>
          <label className="campo">
            <span className="etiqueta">Mes cubierto</span>
            <input
              type="month"
              value={periodo}
              min={mesActual()}
              onChange={(e) => setPeriodo(e.target.value)}
              required
            />
          </label>
          <label className="campo">
            <span className="etiqueta">Valor en pesos colombianos</span>
            <input
              type="number"
              min="1"
              step="1"
              inputMode="numeric"
              value={monto}
              onChange={(e) => setMonto(e.target.value)}
              required
            />
          </label>
          <label className="campo">
            <span className="etiqueta">Medio de pago</span>
            <select value={medio} onChange={(e) => setMedio(e.target.value)}>
              <option value="EFECTIVO">Efectivo</option>
              <option value="TRANSFERENCIA">Transferencia</option>
              <option value="TARJETA">Tarjeta</option>
            </select>
          </label>
          <button type="submit" className="boton" disabled={guardando || !cliente}>
            {guardando ? "Registrando…" : "Registrar pago"}
          </button>
        </form>
      </div>

      <section className="pago-seccion" aria-labelledby="historial-titulo">
        <h3 id="historial-titulo">Historial de pagos</h3>
        {!cliente ? (
          <p className="vacio">Seleccione un cliente para consultar sus pagos.</p>
        ) : (
          <div className="tabla-contenedor">
            <table className="tabla-admin">
              <thead>
                <tr>
                  <th>MES</th><th>VALOR</th><th>MEDIO</th><th>ESTADO</th><th>RECIBO / ACCIONES</th>
                </tr>
              </thead>
              <tbody>
                {cargando ? (
                  <tr><td colSpan="5" className="celda-centro">Cargando historial…</td></tr>
                ) : pagos.length === 0 ? (
                  <tr><td colSpan="5" className="celda-centro">Este cliente todavía no tiene pagos.</td></tr>
                ) : pagos.map((p) => (
                  <tr key={p.id_pago}>
                    <td>{fechaCorta(p.periodo_inicio).slice(0, 7)}</td>
                    <td>{pesos(p.monto)}</td>
                    <td>{p.medio_pago}</td>
                    <td><span className={`badge ${p.estado === "PAGADO" ? "badge--activo" : "badge--inactivo"}`}>{p.estado}</span></td>
                    <td className="acciones-fila">
                      <button type="button" className="boton-link" onClick={() => descargarRecibo(p)}>PDF</button>
                      {esAdmin && p.estado === "PAGADO" && (
                        <button type="button" className="boton-link boton-link--peligro" onClick={() => cancelar(p)}>Cancelar</button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </section>
  );
}
