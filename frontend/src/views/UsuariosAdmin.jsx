import { useCallback, useEffect, useState } from "react";
import { api } from "../api.js";

export default function UsuariosAdmin({ usuarioActual }) {
  const [usuarios, setUsuarios] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState("");
  const [exito, setExito] = useState("");
  const [filtroTexto, setFiltroTexto] = useState("");
  const [filtroRol, setFiltroRol] = useState("TODOS");
  const [filtroEstado, setFiltroEstado] = useState("TODOS");

  // Modal Crear Usuario
  const [modalCrear, setModalCrear] = useState(false);
  const [nuevoUsuario, setNuevoUsuario] = useState("");
  const [nuevoNombre, setNuevoNombre] = useState("");
  const [nuevaClave, setNuevaClave] = useState("");
  const [nuevoRol, setNuevoRol] = useState("ENCARGADO");
  const [guardando, setGuardando] = useState(false);
  const [errorModal, setErrorModal] = useState("");

  // Modal Reset Clave
  const [modalReset, setModalReset] = useState(null);
  const [claveReset, setClaveReset] = useState("");

  const cargar = useCallback(async () => {
    setCargando(true);
    try {
      const data = await api.usuarios();
      setUsuarios(data);
      setError("");
    } catch (err) {
      setError(err.message || "Error al cargar la lista de usuarios.");
    } finally {
      setCargando(false);
    }
  }, []);

  useEffect(() => {
    cargar();
  }, [cargar]);

  // Contar cuántos administradores activos existen
  const adminsActivos = usuarios.filter((u) => u.rol === "ADMINISTRADOR" && u.activo).length;

  async function handleToggleActivo(u) {
    if (u.rol === "ADMINISTRADOR" && u.activo && adminsActivos <= 1) {
      setError("Regla RN-08: Debe quedar al menos un administrador activo en el sistema.");
      return;
    }
    setError("");
    setExito("");
    try {
      await api.actualizarUsuario(u.id_usuario, { activo: !u.activo });
      setExito(`Usuario '${u.usuario}' ${!u.activo ? "activado" : "inactivado"} exitosamente.`);
      await cargar();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleCrearUsuario(e) {
    e.preventDefault();
    setErrorModal("");
    if (!nuevoUsuario.trim()) {
      setErrorModal("Ingrese el nombre de usuario.");
      return;
    }
    if (!nuevoNombre.trim()) {
      setErrorModal("Ingrese el nombre completo.");
      return;
    }
    if (nuevaClave.length < 10) {
      setErrorModal("La contraseña debe tener mínimo 10 caracteres.");
      return;
    }

    setGuardando(true);
    try {
      await api.crearUsuario({
        usuario: nuevoUsuario.trim().toLowerCase(),
        nombre: nuevoNombre.trim(),
        clave: nuevaClave,
        rol: nuevoRol,
      });
      setModalCrear(false);
      setNuevoUsuario("");
      setNuevoNombre("");
      setNuevaClave("");
      setNuevoRol("ENCARGADO");
      setExito("Nuevo usuario registrado correctamente.");
      await cargar();
    } catch (err) {
      setErrorModal(err.message);
    } finally {
      setGuardando(false);
    }
  }

  async function handleResetClave(e) {
    e.preventDefault();
    if (claveReset.length < 10) {
      setError("La nueva clave debe tener al menos 10 caracteres.");
      return;
    }
    try {
      await api.restablecerClave(modalReset.id_usuario, claveReset);
      setExito(`Contraseña restablecida para el usuario '${modalReset.usuario}'.`);
      setModalReset(null);
      setClaveReset("");
    } catch (err) {
      setError(err.message);
    }
  }

  const usuariosFiltrados = usuarios.filter((u) => {
    const texto = filtroTexto.toLowerCase();
    const coincideTexto = u.usuario.toLowerCase().includes(texto) || u.nombre.toLowerCase().includes(texto);
    const coincideRol = filtroRol === "TODOS" || u.rol === filtroRol;
    const coincideEstado =
      filtroEstado === "TODOS" || (filtroEstado === "ACTIVO" && u.activo) || (filtroEstado === "INACTIVO" && !u.activo);
    return coincideTexto && coincideRol && coincideEstado;
  });

  return (
    <div className="panel-admin">
      <header className="panel-admin__cabecera">
        <div>
          <h2>Gestión de Usuarios del Sistema (RBAC)</h2>
          <p className="ayuda">Administre operadores, roles asignados y estados de cuenta.</p>
        </div>
        <button type="button" className="boton" onClick={() => setModalCrear(true)}>
          + Nuevo Usuario
        </button>
      </header>

      {error && <p className="mensaje-alerta mensaje-alerta--error" role="alert">{error}</p>}
      {exito && <p className="mensaje-alerta mensaje-alerta--exito" role="status">{exito}</p>}

      {/* Barra de Filtros */}
      <div className="panel-admin__filtros">
        <input
          type="search"
          placeholder="Buscar por usuario o nombre..."
          value={filtroTexto}
          onChange={(e) => setFiltroTexto(e.target.value)}
          className="campo-busqueda"
        />
        <select value={filtroRol} onChange={(e) => setFiltroRol(e.target.value)} className="selector-filtro">
          <option value="TODOS">Rol: Todos</option>
          <option value="ADMINISTRADOR">Administrador</option>
          <option value="ENCARGADO">Encargado</option>
        </select>
        <select value={filtroEstado} onChange={(e) => setFiltroEstado(e.target.value)} className="selector-filtro">
          <option value="TODOS">Estado: Todos</option>
          <option value="ACTIVO">Solo Activos</option>
          <option value="INACTIVO">Solo Inactivos</option>
        </select>
      </div>

      {/* Tabla de Usuarios */}
      <div className="tabla-contenedor">
        <table className="tabla-admin">
          <thead>
            <tr>
              <th>USUARIO</th>
              <th>NOMBRE COMPLETO</th>
              <th>ROL ASIGNADO</th>
              <th>ESTADO</th>
              <th>ACCIONES</th>
            </tr>
          </thead>
          <tbody>
            {cargando ? (
              <tr>
                <td colSpan="5" className="celda-centro">Cargando usuarios…</td>
              </tr>
            ) : usuariosFiltrados.length === 0 ? (
              <tr>
                <td colSpan="5" className="celda-centro">No se encontraron usuarios.</td>
              </tr>
            ) : (
              usuariosFiltrados.map((u) => {
                const esUltimoAdmin = u.rol === "ADMINISTRADOR" && u.activo && adminsActivos <= 1;
                return (
                  <tr key={u.id_usuario}>
                    <td><strong>{u.usuario}</strong></td>
                    <td>{u.nombre}</td>
                    <td>
                      <span className={`badge badge--${u.rol.toLowerCase()}`}>{u.rol}</span>
                    </td>
                    <td>
                      <span className={`badge badge--${u.activo ? "activo" : "inactivo"}`}>
                        {u.activo ? "ACTIVO" : "INACTIVO"}
                      </span>
                    </td>
                    <td>
                      <div className="acciones-fila">
                        <button
                          type="button"
                          className="boton-link"
                          onClick={() => setModalReset(u)}
                          title="Restablecer contraseña"
                        >
                          Clave
                        </button>
                        {esUltimoAdmin ? (
                          <span className="texto-protegido" title="Regla RN-08: Último administrador activo">
                            🛡️ Protegido
                          </span>
                        ) : (
                          <button
                            type="button"
                            className={`boton-link ${u.activo ? "boton-link--peligro" : "boton-link--exito"}`}
                            onClick={() => handleToggleActivo(u)}
                          >
                            {u.activo ? "Inactivar" : "Activar"}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <div className="alerta-regla">
        <p><strong>Regla RN-08:</strong> El sistema impide desactivar o degradar al último Administrador activo para asegurar la gobernabilidad del sistema.</p>
      </div>

      {/* Modal Crear Usuario */}
      {modalCrear && (
        <div className="modal-fondo" role="dialog" aria-modal="true" aria-labelledby="modal-crear-titulo">
          <div className="modal-caja">
            <header className="modal-cabecera">
              <h3 id="modal-crear-titulo">Nuevo Usuario del Sistema</h3>
              <button type="button" className="icono-boton" onClick={() => setModalCrear(false)}>✕</button>
            </header>
            <form onSubmit={handleCrearUsuario} className="modal-cuerpo" noValidate>
              {errorModal && <p className="ayuda ayuda--error">{errorModal}</p>}
              <div className="campo">
                <label className="etiqueta">Nombre de usuario (Login) *</label>
                <input
                  type="text"
                  value={nuevoUsuario}
                  onChange={(e) => setNuevoUsuario(e.target.value)}
                  placeholder="ej. smorales (3-30 car, minúsculas)"
                  required
                />
              </div>
              <div className="campo">
                <label className="etiqueta">Nombre completo *</label>
                <input
                  type="text"
                  value={nuevoNombre}
                  onChange={(e) => setNuevoNombre(e.target.value)}
                  placeholder="ej. Santiago Morales Restrepo"
                  required
                />
              </div>
              <div className="campo">
                <label className="etiqueta">Rol en el sistema *</label>
                <div className="opciones-radio">
                  <label>
                    <input
                      type="radio"
                      name="rol"
                      value="ENCARGADO"
                      checked={nuevoRol === "ENCARGADO"}
                      onChange={() => setNuevoRol("ENCARGADO")}
                    />
                    <strong>ENCARGADO</strong> (Operación de patio, entradas y salidas)
                  </label>
                  <label>
                    <input
                      type="radio"
                      name="rol"
                      value="ADMINISTRADOR"
                      checked={nuevoRol === "ADMINISTRADOR"}
                      onChange={() => setNuevoRol("ADMINISTRADOR")}
                    />
                    <strong>ADMINISTRADOR</strong> (Control total, usuarios, celdas y auditoría)
                  </label>
                </div>
              </div>
              <div className="campo">
                <label className="etiqueta">Contraseña inicial *</label>
                <input
                  type="password"
                  value={nuevaClave}
                  onChange={(e) => setNuevaClave(e.target.value)}
                  placeholder="Mínimo 10 car, al menos una letra y un número"
                  required
                />
                <p className="ayuda">Cumple la política de robustez RN-07.</p>
              </div>
              <footer className="modal-acciones">
                <button type="button" className="boton boton--secundario" onClick={() => setModalCrear(false)} disabled={guardando}>
                  Cancelar
                </button>
                <button type="submit" className="boton" disabled={guardando}>
                  {guardando ? "Registrando…" : "Guardar Usuario"}
                </button>
              </footer>
            </form>
          </div>
        </div>
      )}

      {/* Modal Restablecer Clave */}
      {modalReset && (
        <div className="modal-fondo" role="dialog" aria-modal="true" aria-labelledby="modal-reset-titulo">
          <div className="modal-caja">
            <header className="modal-cabecera">
              <h3 id="modal-reset-titulo">Restablecer Contraseña</h3>
              <button type="button" className="icono-boton" onClick={() => setModalReset(null)}>✕</button>
            </header>
            <form onSubmit={handleResetClave} className="modal-cuerpo">
              <p>Asigne una nueva contraseña para el operador <strong>{modalReset.nombre}</strong> (<code>{modalReset.usuario}</code>):</p>
              <div className="campo">
                <label className="etiqueta">Nueva Contraseña *</label>
                <input
                  type="password"
                  value={claveReset}
                  onChange={(e) => setClaveReset(e.target.value)}
                  placeholder="Mínimo 10 caracteres"
                  autoFocus
                  required
                />
              </div>
              <footer className="modal-acciones">
                <button type="button" className="boton boton--secundario" onClick={() => setModalReset(null)}>
                  Cancelar
                </button>
                <button type="submit" className="boton">
                  Guardar Nueva Contraseña
                </button>
              </footer>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

