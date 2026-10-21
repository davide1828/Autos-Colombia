# Parqueadero "Autos Colombia" — Iteración 1: entrada y salida de vehículos

Aplicación web responsive (PWA) para celulares, tablets y computadores.

- **Backend:** Go (arquitectura hexagonal: `domain`, `application`, `adapters/httpapi`, `adapters/postgres`)
- **Frontend:** React + Vite
- **Base de datos:** PostgreSQL 16 (`backend/db/schema.sql`, `backend/db/seed.sql`)
- **Documentación y diagramas:** `docs/` (fuentes PlantUML importables en StarUML + PNG)

## Ejecutar

```bash
# 1. Base de datos (opción Docker)
docker compose up -d db

# 2. Frontend (genera backend/static)
cd frontend && npm install && npm run build && cp -r dist ../backend/static && cd ..

# 3. API + frontend en http://localhost:8080
cd backend
API_TOKEN=cambiar-esta-clave go run ./cmd/api
```

Variables de entorno: `DATABASE_URL`, `API_TOKEN`, `PORT` (8080), `STATIC_DIR` (./static).
Usuario de prueba: ingrese la clave `API_TOKEN` en el campo "Clave de acceso" de la app.

## Pruebas

```bash
cd backend && go test ./...
```

## API

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/ingresos` | Registra el ingreso `{placa, id_celda, novedad?}` |
| POST | `/api/salidas` | Registra la salida `{placa, novedad?}` |
| POST | `/api/novedades` | Registra una novedad `{placa, tipo, descripcion}` |
| GET | `/api/celdas?estado=DISPONIBLE` | Consulta celdas |
| GET | `/api/registros/activos` | Vehículos dentro |

Todas exigen `Authorization: Bearer <API_TOKEN>`.

## Mockups

Archivo de Figma: https://www.figma.com/design/DhCeukFFGQef4egiIl3UbY/
