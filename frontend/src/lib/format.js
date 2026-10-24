// Funciones puras de formato y validación (probadas en format.test.js).

export const PLACA_RE = /^[A-Z]{3}\d{2}[A-Z0-9]$/;
const ZONA = "America/Bogota";

/** Mayúsculas, solo letras y números, máximo 6 caracteres. */
export function normalizarPlaca(valor) {
  return String(valor ?? "").toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 6);
}

/** Devuelve el mensaje de error o "" si la placa es válida. */
export function validarPlaca(placa) {
  if (!placa) return "Escriba la placa del vehículo.";
  if (!PLACA_RE.test(placa)) return "La placa debe tener 3 letras y 3 caracteres, por ejemplo ABC123 o ABC12D.";
  return "";
}

/** 0 → "menos de 1 min", 35 → "35 min", 95 → "1 h 35 min", 1500 → "1 d 1 h". */
export function formatearDuracion(minutos) {
  const m = Math.max(0, Math.floor(minutos));
  if (m < 1) return "menos de 1 min";
  if (m < 60) return `${m} min`;
  if (m < 1440) {
    const h = Math.floor(m / 60), r = m % 60;
    return r ? `${h} h ${r} min` : `${h} h`;
  }
  const d = Math.floor(m / 1440), h = Math.floor((m % 1440) / 60);
  return h ? `${d} d ${h} h` : `${d} d`;
}

export function minutosDesde(iso, ahora = Date.now()) {
  return Math.max(0, Math.floor((ahora - new Date(iso).getTime()) / 60000));
}

/** "10:33 p. m." en hora de Colombia, sin importar la zona del dispositivo. */
export function horaCorta(iso) {
  return new Intl.DateTimeFormat("es-CO", { hour: "numeric", minute: "2-digit", timeZone: ZONA }).format(new Date(iso));
}

export function resumenOcupacion(celdas) {
  const ocupadas = celdas.filter((c) => c.estado === "OCUPADA").length;
  const libres = celdas.filter((c) => c.estado === "DISPONIBLE").length;
  const inactivas = celdas.length - ocupadas - libres;
  return { total: celdas.length, ocupadas, libres, inactivas, enServicio: celdas.length - inactivas };
}
