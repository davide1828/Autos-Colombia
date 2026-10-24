import { describe, expect, it } from "vitest";
import { formatearDuracion, horaCorta, minutosDesde, normalizarPlaca, resumenOcupacion, validarPlaca } from "./format.js";

describe("normalizarPlaca", () => {
  it("pasa a mayúsculas y quita símbolos y espacios", () => {
    expect(normalizarPlaca(" abc-123 ")).toBe("ABC123");
    expect(normalizarPlaca("xyz 98a")).toBe("XYZ98A");
  });
  it("limita a 6 caracteres y tolera valores vacíos", () => {
    expect(normalizarPlaca("ABC1234567")).toBe("ABC123");
    expect(normalizarPlaca(undefined)).toBe("");
  });
});

describe("validarPlaca", () => {
  it("acepta carro y moto", () => {
    expect(validarPlaca("ABC123")).toBe("");
    expect(validarPlaca("XYZ98A")).toBe("");
  });
  it("explica el error", () => {
    expect(validarPlaca("")).toMatch(/Escriba/);
    expect(validarPlaca("AB12")).toMatch(/3 letras/);
    expect(validarPlaca("123ABC")).toMatch(/3 letras/);
  });
});

describe("formatearDuracion", () => {
  it.each([
    [0, "menos de 1 min"], [-5, "menos de 1 min"], [1, "1 min"], [35, "35 min"], [60, "1 h"],
    [95, "1 h 35 min"], [1439, "23 h 59 min"], [1440, "1 d"], [1500, "1 d 1 h"], [2880 + 120, "2 d 2 h"],
  ])("%i min → %s", (min, esperado) => expect(formatearDuracion(min)).toBe(esperado));
});

describe("minutosDesde", () => {
  it("cuenta minutos completos y nunca es negativo", () => {
    const base = Date.parse("2026-10-20T15:00:00Z");
    expect(minutosDesde("2026-10-20T14:25:30Z", base)).toBe(34);
    expect(minutosDesde("2026-10-20T16:00:00Z", base)).toBe(0);
  });
});

describe("horaCorta", () => {
  it("usa la hora de Colombia (UTC-5)", () => {
    expect(horaCorta("2026-10-20T15:05:00Z")).toMatch(/^10:05/);
  });
});

describe("resumenOcupacion", () => {
  it("cuenta por estado", () => {
    const r = resumenOcupacion([{ estado: "OCUPADA" }, { estado: "DISPONIBLE" }, { estado: "DISPONIBLE" }, { estado: "INACTIVA" }]);
    expect(r).toEqual({ total: 4, ocupadas: 1, libres: 2, inactivas: 1, enServicio: 3 });
  });
});
