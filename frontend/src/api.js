// Cliente de la API REST. La clave se guarda solo durante la sesión del navegador.
const CLAVE = "ac_token";
let alNoAutorizado = () => {};

export const sesion = {
  leer: () => sessionStorage.getItem(CLAVE) || "",
  guardar: (t) => sessionStorage.setItem(CLAVE, t),
  borrar: () => sessionStorage.removeItem(CLAVE),
};
export const configurar = ({ onNoAutorizado }) => { alNoAutorizado = onNoAutorizado; };

export class ApiError extends Error {
  constructor(mensaje, estado) { super(mensaje); this.estado = estado; }
}

async function llamar(ruta, opciones = {}, { avisarSiNoAutorizado = true } = {}) {
  let res;
  try {
    res = await fetch(ruta, {
      ...opciones,
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${sesion.leer()}` },
    });
  } catch {
    throw new ApiError("No hay conexión con el servidor. Revise la red e inténtelo de nuevo.", 0);
  }
  const datos = await res.json().catch(() => ({}));
  if (res.status === 401 && avisarSiNoAutorizado) alNoAutorizado();
  if (!res.ok) throw new ApiError(datos.mensaje || "No se pudo completar la operación.", res.status);
  return datos;
}

const post = (ruta, cuerpo) => llamar(ruta, { method: "POST", body: JSON.stringify(cuerpo) });

export const api = {
  verificarClave: (clave) => {
    sesion.guardar(clave);
    return llamar("/api/sesion", {}, { avisarSiNoAutorizado: false });
  },
  celdas: () => llamar("/api/celdas"),
  activos: () => llamar("/api/registros/activos"),
  ingreso: (d) => post("/api/ingresos", d),
  salida: (d) => post("/api/salidas", d),
  novedad: (d) => post("/api/novedades", d),
};
