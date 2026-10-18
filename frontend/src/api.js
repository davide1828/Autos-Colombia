// Cliente de la API REST. El token se guarda solo durante la sesión del navegador.
const getToken = () => sessionStorage.getItem("token") || "";
export const setToken = (t) => sessionStorage.setItem("token", t);

async function llamar(ruta, opciones = {}) {
  const res = await fetch(ruta, {
    ...opciones,
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
  });
  const datos = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(datos.mensaje || "No se pudo completar la operación");
  return datos;
}

const post = (ruta, cuerpo) => llamar(ruta, { method: "POST", body: JSON.stringify(cuerpo) });

export const api = {
  celdas: (estado = "") => llamar(`/api/celdas${estado ? `?estado=${estado}` : ""}`),
  activos: () => llamar("/api/registros/activos"),
  ingreso: (d) => post("/api/ingresos", d),
  salida: (d) => post("/api/salidas", d),
  novedad: (d) => post("/api/novedades", d),
};
