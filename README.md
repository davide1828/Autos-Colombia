# Autos Colombia - Parqueadero

Aplicación web para gestionar el ingreso, salida y control de vehículos en un parqueadero, con soporte para usuarios administrativos y flujo de operación desde dispositivos móviles y escritorio.

## ¿Qué hace el sistema?

La plataforma permite:

- registrar ingresos de vehículos con selección de celda
- registrar salidas y cálculo del tiempo transcurrido
- controlar celdas disponibles y ocupadas
- manejar novedades asociadas a un vehículo
- validar usuarios mensuales y condiciones de acceso
- mantener una auditoría de movimientos y administración del patio
- ofrecer una interfaz responsive y de fácil uso en celular y escritorio

## Stack tecnológico

- Frontend: React + Vite
- Backend: Go
- Base de datos: PostgreSQL 16
- Arquitectura: enfoque hexagonal / capas (`domain`, `application`, `adapters`, `config`)
- Seguridad: token de API, hash PBKDF2, validación de credenciales, límites de intentos y headers de seguridad

## Estructura del proyecto

```text
.
├── backend/
│   ├── cmd/api/                 # punto de entrada de la API
│   ├── db/                      # esquema y datos semilla de PostgreSQL
│   ├── internal/
│   │   ├── adapters/            # HTTP, Postgres, seguridad, memoria
│   │   ├── application/        # casos de uso y servicios
│   │   ├── config/             # carga de variables de entorno
│   │   └── domain/             # entidades y reglas de negocio
│   └── static/                 # assets del frontend compilados
├── frontend/
│   ├── src/                    # código del cliente React
│   ├── package.json            # scripts y dependencias del frontend
│   └── vite.config.js         # configuración de Vite
├── docs/
│   ├── diagramas/              # diagramas PlantUML y modelos
│   └── pantallas/             # mockups y recursos visuales
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── README.md
└── iniciar.bat / iniciar.ps1
```

## Requisitos

- Docker y Docker Compose
- Go 1.22+
- Node.js 18+ y npm
- Puerto 8080 disponible para la app

## Inicio rápido

### 1) Levantar la base de datos

```bash
docker compose up -d db
```

Esto crea la base PostgreSQL con el esquema inicial y los datos semilla.

### 2) Compilar el frontend y dejarlo disponible para la API

```bash
make build
```

Equivalente a:

```bash
cd frontend
npm ci
npm run build
cp -r dist ../backend/static
```

### 3) Ejecutar la API

```bash
cd backend
API_TOKEN=una-clave-larga-123 go run ./cmd/api
```

La aplicación queda disponible en:

- http://localhost:8080

> La variable `API_TOKEN` es obligatoria. Debe tener al menos 16 caracteres. En desarrollo local también puede usarse `ALLOW_INSECURE_NO_AUTH=true`, pero solo como excepción y no como configuración de producción.

## Variables de entorno

La API utiliza estas variables:

- `API_TOKEN`: secreto requerido para autenticar solicitudes HTTP
- `DATABASE_URL`: URL de conexión a PostgreSQL
- `PORT`: puerto HTTP principal (por defecto `8080`)
- `STATIC_DIR`: carpeta donde están los assets estáticos (por defecto `./static`)
- `ALLOW_INSECURE_NO_AUTH`: habilita modo local inseguro para pruebas rápidas
- `ADMIN_USUARIO`, `ADMIN_NOMBRE`, `ADMIN_CLAVE`: crean un administrador inicial si existe un usuario vacío
- `HASH_ITERACIONES`: costo de PBKDF2 para hash de contraseñas

## Ejecutar con Docker Compose

```bash
API_TOKEN=una-clave-larga-123 docker compose --profile app up --build
```

Esto levanta la base de datos y la aplicación en un único flujo con la configuración del proyecto.

## Comandos útiles

```bash
make db
make build
make lint
make test
make test-integration
```

Descripción:

- `make lint`: valida formato Go y lint del frontend
- `make test`: ejecuta pruebas del backend con `-race` y pruebas del frontend
- `make test-integration`: ejecuta validaciones contra PostgreSQL real

## API principal

Todas las rutas protegidas requieren el encabezado:

```http
Authorization: Bearer <API_TOKEN>
```

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/ingresos` | Registra un ingreso con placa, celda y opcional novedad |
| POST | `/api/salidas` | Registra la salida de un vehículo |
| POST | `/api/novedades` | Guarda una novedad asociada a una placa |
| GET | `/api/celdas?estado=DISPONIBLE` | Consulta celdas por estado |
| GET | `/api/registros/activos` | Muestra los vehículos presentes |
| GET | `/api/sesion` | Valida la sesión o clave de acceso |

## Calidad y seguridad

El proyecto incluye criterios relevantes de calidad:

- arquitectura en capas con separación de responsabilidades
- pruebas de dominio, aplicación, HTTP y PostgreSQL
- validación de concurrencia y restricciones de base de datos
- seguridad con comparación constante de secretos
- headers de seguridad para prevención de abuso de la interfaz
- manejo ordenado de cierre del servidor
- zona horaria de Colombia UTC-5 para cálculos de vigencia de mensualidad
- frontend accesible y responsive

## Documentación y diagramas

La carpeta `docs/` contiene los artefactos del proyecto:

- diagramas UML de arquitectura y flujo
- modelado relacional y físico
- casos de uso y secuencias
- mockups de pantallas y prototipos de UI

## Estado del proyecto

Este repositorio corresponde a una iteración funcional del sistema de parqueadero con foco en operación de ingreso/salida, administración, validación de negocio y experiencia de usuario móvil.

Si quieres, puedo dejarte una segunda versión del README orientada a:

1. presentación más comercial y profesional,
2. documentación técnica más detallada para desarrolladores,
3. versión enfocada en entrega académica con objetivos, arquitectura y evidencia de pruebas.
