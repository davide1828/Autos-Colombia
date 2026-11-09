// Cliente de la API REST para el Sistema Parqueadero Autos Colombia (Iteraciones 1 y 2).
const TOKEN_KEY = "ac_token";
const USUARIO_KEY = "ac_usuario";
let alNoAutorizado = () => {};

export const sesion = {
  leer: () => sessionStorage.getItem(TOKEN_KEY) || "",
  guardar: (token, usuario) => {
    sessionStorage.setItem(TOKEN_KEY, token);
    if (usuario) {
      sessionStorage.setItem(USUARIO_KEY, JSON.stringify(usuario));
    }
  },
  obtenerUsuario: () => {
    try {
      return JSON.parse(sessionStorage.getItem(USUARIO_KEY) || "null");
    } catch {
      return null;
    }
  },
  borrar: () => {
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(USUARIO_KEY);
  },
};

export const configurar = ({ onNoAutorizado }) => {
  alNoAutorizado = onNoAutorizado;
};

export class ApiError extends Error {
  constructor(mensaje, estado) {
    super(mensaje);
    this.estado = estado;
  }
}

async function llamar(ruta, opciones = {}, { avisarSiNoAutorizado = true } = {}) {
  let res;
  const token = sesion.leer();
  const headers = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  try {
    res = await fetch(ruta, {
      ...opciones,
      headers: { ...headers, ...opciones.headers },
    });
  } catch {
    throw new ApiError("No hay conexión con el servidor. Revise la red e inténtelo de nuevo.", 0);
  }

  const datos = await res.json().catch(() => ({}));

  if (res.status === 401 && avisarSiNoAutorizado) {
    alNoAutorizado();
  }

  if (!res.ok) {
    throw new ApiError(datos.mensaje || "No se pudo completar la operación.", res.status);
  }

  return datos;
}

const post = (ruta, cuerpo) => llamar(ruta, { method: "POST", body: JSON.stringify(cuerpo) });
const patch = (ruta, cuerpo) => llamar(ruta, { method: "PATCH", body: JSON.stringify(cuerpo) });
const put = (ruta, cuerpo) => llamar(ruta, { method: "PUT", body: JSON.stringify(cuerpo) });

export const api = {
  // ---------- Sesión y Seguridad (Iteración 2) ----------
  login: async (usuario, clave) => {
    const res = await post("/api/login", { usuario, clave });
    sesion.guardar(res.token, res.usuario);
    return res;
  },
  sesion: () => llamar("/api/sesion", {}, { avisarSiNoAutorizado: false }),
  cambiarMiClave: (actual, nueva) => post("/api/sesion/clave", { actual, nueva }),

  // ---------- Patio Operativo (Iteración 1) ----------
  celdas: () => llamar("/api/celdas"),
  activos: () => llamar("/api/registros/activos"),
  ingreso: (d) => post("/api/ingresos", d),
  salida: (d) => post("/api/salidas", d),
  novedad: (d) => post("/api/novedades", d),

  // ---------- Gestión de Usuarios (Iteración 2 - Admin) ----------
  usuarios: () => llamar("/api/usuarios"),
  crearUsuario: (d) => post("/api/usuarios", d),
  actualizarUsuario: (id, d) => patch(`/api/usuarios/${id}`, d),
  restablecerClave: (id, nueva) => post(`/api/usuarios/${id}/clave`, { nueva }),

  // ---------- Gestión de Celdas (Iteración 2 - Admin) ----------
  crearCelda: (d) => post("/api/celdas", d),
  actualizarCelda: (id, d) => patch(`/api/celdas/${id}`, d),
  cambiarEstadoCelda: (id, estado) => put(`/api/celdas/${id}/estado`, { estado }),

  // ---------- Clientes y Mensualidades (Iteración 2) ----------
  clientes: (q = "") => llamar(`/api/clientes?q=${encodeURIComponent(q)}`),
  cliente: (id) => llamar(`/api/clientes/${id}`),
  agregarMensualidad: (idCliente, d) => post(`/api/clientes/${idCliente}/mensualidades`, d),

  // ---------- Auditoría Inmutable (Iteración 2 - Admin) ----------
  auditoria: (limite = 50) => llamar(`/api/auditoria?limite=${limite}`),
};
