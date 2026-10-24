// Íconos de trazo (24×24). Decorativos: siempre van acompañados de texto.
const RUTAS = {
  carro: <><path d="M5 16V11l2-5h10l2 5v5" /><path d="M3 16h18v3H3z" /><circle cx="7.5" cy="13.5" r=".6" fill="currentColor" /><circle cx="16.5" cy="13.5" r=".6" fill="currentColor" /></>,
  moto: <><circle cx="6" cy="16" r="3" /><circle cx="18" cy="16" r="3" /><path d="M6 16l4-7h4l3 7M9 9H7M14 9l2-3h2" /></>,
  entrar: <><path d="M14 4h5v16h-5" /><path d="M4 12h10M10 8l4 4-4 4" /></>,
  salir: <><path d="M10 4H5v16h5" /><path d="M10 12h10M16 8l4 4-4 4" /></>,
  novedad: <><path d="M12 3l9 16H3z" /><path d="M12 10v4M12 17v.5" /></>,
  patio: <><rect x="3" y="3" width="8" height="8" rx="1" /><rect x="13" y="3" width="8" height="8" rx="1" /><rect x="3" y="13" width="8" height="8" rx="1" /><rect x="13" y="13" width="8" height="8" rx="1" /></>,
  check: <path d="M5 12.5l4.5 4.5L19 7.5" />,
  cerrar: <path d="M6 6l12 12M18 6L6 18" />,
  salirSesion: <><path d="M9 4H5v16h4" /><path d="M16 8l4 4-4 4M20 12H9" /></>,
  mas: <path d="M12 5v14M5 12h14" />,
  menos: <path d="M5 12h14" />,
  libre: <circle cx="12" cy="12" r="5" />,
  ocupada: <circle cx="12" cy="12" r="5" fill="currentColor" />,
  fuera: <><circle cx="12" cy="12" r="6" /><path d="M8 16l8-8" /></>,
  sinConexion: <><path d="M3 9a14 14 0 0118 0M6 12.5a9 9 0 0112 0M9.5 16a4 4 0 015 0" /><path d="M4 4l16 16" /></>,
};

export default function Icono({ nombre, tamano = 20 }) {
  return (
    <svg className="icono" width={tamano} height={tamano} viewBox="0 0 24 24" fill="none" stroke="currentColor"
      strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" focusable="false">
      {RUTAS[nombre]}
    </svg>
  );
}
