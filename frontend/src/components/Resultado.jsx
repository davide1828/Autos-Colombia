import Icono from "./Icono.jsx";

/** Confirmación (o error) de la última operación, anunciada a lectores de pantalla. */
export default function Resultado({ tipo, titulo, children, onCerrar }) {
  const esError = tipo === "error";
  return (
    <section className={`resultado resultado--${tipo}`} role={esError ? "alert" : "status"}>
      <Icono nombre={esError ? "novedad" : "check"} tamano={24} />
      <div className="resultado__texto">
        <h3>{titulo}</h3>
        <p>{children}</p>
      </div>
      {onCerrar && (
        <button type="button" className="icono-boton" onClick={onCerrar} aria-label="Cerrar mensaje"><Icono nombre="cerrar" tamano={18} /></button>
      )}
    </section>
  );
}
