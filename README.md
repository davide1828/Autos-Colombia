# Parqueadero "Autos Colombia" — Iteración 1: entrada y salida de vehículos

Aplicación web responsive (PWA) para celulares, tablets y computadores.

- **Backend:** Go (arquitectura hexagonal: `domain`, `application`, `adapters/httpapi`, `adapters/postgres`)
- **Frontend:** React + Vite
- **Base de datos:** PostgreSQL 16 (`backend/db/schema.sql`, `backend/db/seed.sql`)
- **Documentación y diagramas:** `docs/` (fuentes PlantUML importables en StarUML + PNG)

## Ejecutar

```bash
# 1. Base de datos (Docker)
docker compose up -d db

# 2. Compilar el frontend dentro del backend
make build            # o: cd frontend && npm ci && npm run build && cp -r dist ../backend/static

# 3. API + frontend en http://localhost:8080  (API_TOKEN: mínimo 16 caracteres)
cd backend && API_TOKEN=una-clave-larga-123 go run ./cmd/api
```

Todo en contenedores: `API_TOKEN=una-clave-larga-123 docker compose --profile app up --build`.

Variables: `API_TOKEN` (obligatoria), `DATABASE_URL`, `PORT` (8080), `STATIC_DIR` (./static).
La API **no arranca sin `API_TOKEN`** (solo con `ALLOW_INSECURE_NO_AUTH=true`, para desarrollo local).
Al abrir la app, escriba esa clave en la pantalla de acceso.

## Calidad

```bash
make lint      # gofmt, go vet, eslint
make test      # backend con -race + frontend (vitest)
make test-integration   # incluye PostgreSQL real (use una base DESECHABLE)
```

- Backend: arquitectura hexagonal; pruebas de dominio, casos de uso, API HTTP, configuración y PostgreSQL (incluye carrera de concurrencia y restricciones de la base).
- Seguridad: clave obligatoria y comparada en tiempo constante, límite de intentos fallidos, cabeceras de seguridad (CSP, X-Frame-Options…), tiempos de espera del servidor, cierre ordenado, consultas parametrizadas.
- Zona horaria: la vigencia de la mensualidad se evalúa con el día calendario de Colombia (UTC-5).
- Frontend: React sin librerías de UI; accesible (etiquetas, foco visible, `aria-live`, no depende solo del color), responsive de 320 px a escritorio.

## API

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/ingresos` | Registra el ingreso `{placa, id_celda, novedad?}` |
| POST | `/api/salidas` | Registra la salida `{placa, novedad?}` |
| POST | `/api/novedades` | Registra una novedad `{placa, tipo, descripcion}` |
| GET | `/api/celdas?estado=DISPONIBLE` | Consulta celdas |
| GET | `/api/registros/activos` | Vehículos dentro |
| GET | `/api/sesion` | Valida la clave de acceso |

Todas exigen `Authorization: Bearer <API_TOKEN>`.

## Mockups

Archivo de Figma: https://www.figma.com/design/DhCeukFFGQef4egiIl3UbY/

Está actualizado al diseño final de la interfaz (25 de octubre de 2026): componentes reutilizables (placa colombiana, selector de celda, bahía del patio, navegación inferior, mensajes, chips y botón), paleta como variables de color (asfalto, papel, placa, libre, ocupada), tipografía Barlow / Barlow Condensed, ocho pantallas para celular (acceso, ingreso diligenciado, ingreso exitoso, rechazo por mensualidad, salida, salida exitosa, novedad y patio), la vista de computador y un prototipo navegable (inicia en la pantalla de acceso).
