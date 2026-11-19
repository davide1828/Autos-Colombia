import { useState } from "react";
import { api } from "../api.js";

export default function ClientesVehiculos() {
  const [formulario, setFormulario] = useState({
    documento: "",
    nombre: "",
    telefono: "",
    placa: "",
    tipo: "CARRO",
  });
  const [guardando, setGuardando] = useState(false);
  const [error, setError] = useState("");
  const [exito, setExito] = useState("");

  function cambiar(campo, valor) {
    setFormulario((actual) => ({ ...actual, [campo]: valor }));
  }

  async function enviar(e) {
    e.preventDefault();
    setError("");
    setExito("");
    setGuardando(true);
    try {
      const resultado = await api.crearVehiculo({
        placa: formulario.placa.trim().toUpperCase(),
        tipo: formulario.tipo,
        cliente: {
          documento: formulario.documento.trim(),
          nombre: formulario.nombre.trim(),
          telefono: formulario.telefono.trim(),
        },
      });
      setExito(
        `${resultado.cliente_nuevo ? "Cliente registrado" : "Cliente existente"} y vehículo ${resultado.placa} asociado correctamente.`,
      );
      setFormulario({ documento: "", nombre: "", telefono: "", placa: "", tipo: "CARRO" });
    } catch (e) {
      setError(e.message);
    } finally {
      setGuardando(false);
    }
  }

  return (
    <section className="panel-admin" aria-labelledby="clientes-vehiculos-titulo">
      <header className="panel-admin__cabecera">
        <div>
          <h2 id="clientes-vehiculos-titulo">Clientes y vehículos</h2>
          <p className="ayuda">Registre un cliente junto con su primer vehículo o asocie otro vehículo a un cliente existente.</p>
        </div>
      </header>
      {error && <p className="mensaje-alerta mensaje-alerta--error" role="alert">{error}</p>}
      {exito && <p className="mensaje-alerta mensaje-alerta--exito" role="status">{exito}</p>}
      <form className="formulario-cliente" onSubmit={enviar}>
        <label className="campo">
          <span className="etiqueta">Documento del cliente</span>
          <input value={formulario.documento} onChange={(e) => cambiar("documento", e.target.value)} required maxLength={20} />
        </label>
        <label className="campo">
          <span className="etiqueta">Nombre completo</span>
          <input value={formulario.nombre} onChange={(e) => cambiar("nombre", e.target.value)} required maxLength={100} />
        </label>
        <label className="campo">
          <span className="etiqueta">Teléfono (opcional)</span>
          <input value={formulario.telefono} onChange={(e) => cambiar("telefono", e.target.value)} maxLength={20} />
        </label>
        <label className="campo">
          <span className="etiqueta">Placa</span>
          <input value={formulario.placa} onChange={(e) => cambiar("placa", e.target.value)} required maxLength={6} autoCapitalize="characters" />
        </label>
        <label className="campo">
          <span className="etiqueta">Tipo de vehículo</span>
          <select value={formulario.tipo} onChange={(e) => cambiar("tipo", e.target.value)}>
            <option value="CARRO">Carro</option>
            <option value="MOTO">Moto</option>
          </select>
        </label>
        <p className="ayuda formulario-cliente__nota">
          Si el documento ya existe, el sistema conservará los datos del cliente y asociará el nuevo vehículo. Las cuentas de encargados se crean en «Usuarios».
        </p>
        <button type="submit" className="boton formulario-cliente__accion" disabled={guardando}>
          {guardando ? "Guardando…" : "Registrar cliente y vehículo"}
        </button>
      </form>
    </section>
  );
}
