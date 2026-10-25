# Prompt para continuar en otro chat

> Copia todo lo que está debajo de la línea y pégalo en el chat nuevo, junto con el ZIP `parqueadero_EA1_codigo.zip` y el PDF `España_Cristian_EA1.pdf`.

---

Eres mi asistente para la **Evidencia de Aprendizaje 1 (EA1)** de *Ingeniería de Software* (docente: Esteban Montoya, IU Digital de Antioquia). Soy **Cristian David España Figueroa**, desarrollador full-stack (Go/Fiber, React, PostgreSQL), fundador de EXARIX S.A.S., en Medellín. Háblame directo y en español, con pasos concretos. No afirmes que algo funciona si no lo verificaste; si no puedes hacer algo, dilo.

## La tarea
Sistema de información del parqueadero **"Autos Colombia"** (parqueo por mensualidad). Requerimientos globales: entrada/salida de vehículos, usuarios y celdas, novedades sobre vehículos y pagos. **Esta entrega es la Iteración 1: gestión de entrada y salida.** Entregable: un PDF (portada, introducción, entregables, conclusiones) nombrado `España_Cristian_EA1`, con: (1) requerimientos funcionales y no funcionales, (2) historias de usuario, (3) arquitectura en StarUML, (4) mockups en Figma, (5) MER, modelo relacional y físico, (6) casos de uso en StarUML, (7) diagrama de clases en StarUML, (8) diagrama de paquetes en StarUML, (9) código corriendo y en GitHub. La estructura sigue el ejemplo de clase de "préstamo de equipos" (necesidad → regla → función; historias Dado/Cuando/Entonces; capas; coherencia entre representaciones).

## Decisiones ya tomadas
- Aplicación **web responsive (PWA)**: celular, tablet y PC. Stack: React + Vite, Go (hexagonal), PostgreSQL 16.
- La capa HTTP usa `net/http` (en mi sandbox no se pudo descargar Fiber); está aislada en un adaptador, así que cambiarla a Fiber no toca dominio ni aplicación.
- Reglas: RN-01 solo ingresa vehículo registrado con mensualidad vigente; RN-02 un registro activo por vehículo; RN-03 un registro activo por celda y celda del mismo tipo (CARRO/MOTO); RN-04 la salida exige ingreso activo y libera la celda; RN-05 novedades (DANIO/FALTANTE/OTRA) sobre vehículos dentro, al ingresar, al salir o durante la estancia; RN-06 el historial no se borra.
- Alcance: clientes, vehículos, mensualidades y celdas ya existen (datos semilla); un solo actor, el Encargado; sin reservas.

## Estado actual (todo verificado, salvo lo marcado)
- **Backend** (`backend/`): dominio, casos de uso, adaptadores `httpapi`, `postgres` y `memory`, `config`. **55 pruebas pasan con PostgreSQL real y `-race`** (51 sin base de datos, 4 se omiten). Cobertura: dominio 100 %, config 100 %, API 90 %, aplicación 89 %, persistencia 74 %. Incluye una carrera real de concurrencia y las restricciones de la base. Corregido: bug de zona horaria (la mensualidad se evalúa con el día de Colombia, UTC-5).
- **Seguridad**: `API_TOKEN` obligatorio (≥16 caracteres) o la API no arranca (solo `ALLOW_INSECURE_NO_AUTH=true` en desarrollo); comparación en tiempo constante; límite de 10 intentos fallidos por minuto por IP; CSP y cabeceras de seguridad; timeouts del servidor; cierre ordenado; consultas parametrizadas. Limitación conocida: es una clave compartida, **no hay usuarios ni roles**.
- **Frontend** (`frontend/`): rediseño profesional. Identidad: campo de **placa con aspecto de placa colombiana**, tablero del patio con bahías, tipografía Barlow/Barlow Condensed empaquetada, pantalla de acceso, validación, atajos de "vehículos dentro", accesibilidad (foco, aria-live, no solo color), responsive de 320 px a escritorio (en ≥720 px formulario y patio van juntos). ESLint sin errores, 17 pruebas de lógica, y un recorrido end-to-end con navegador real (13 comprobaciones, todas OK).
- **Calidad/DevOps**: `Makefile`, `Dockerfile`, `docker-compose.yml`, `.env.example`, `.github/workflows/ci.yml`. **El Dockerfile y el workflow de CI NO se han ejecutado** (mi sandbox no tiene Docker ni GitHub Actions): probarlos y corregir si hace falta.
- **Documentación**: PDF de 16 páginas (portada completa con mi nombre, materia, código del curso PREICA2602B020081, periodo 2026_2_B2 y docente Esteban Montoya Molina; enlace de Figma incluido; fecha 25 de octubre de 2026). Fuentes PlantUML de los diagramas en `docs/diagramas/` (+PNG), capturas en `docs/pantallas/`, y el generador del PDF en `docs/make_pdf.py` (config en `docs/pdf_config.json`).
- **Figma**: archivo `https://www.figma.com/design/DhCeukFFGQef4egiIl3UbY/` (fileKey `DhCeukFFGQef4egiIl3UbY`, plan team `team::1693855459363039951`). **Actualizado al diseño final (25-oct-2026)**: secciones *Componentes*, *Pantallas · celular* (8 pantallas: acceso, ingreso diligenciado, ingreso exitoso, rechazo, salida, salida exitosa, novedad, patio) y *Pantalla · computador*; variables de color en la colección "Autos Colombia · Final"; prototipo navegable (27 enlaces; flujo inicia en Acceso). Se eliminó el diseño azul anterior (queda en el historial de versiones). Limitaciones: íconos dibujados a mano (no los del código); la pantalla de rechazo no está enlazada en el prototipo.

## Pendiente (en este orden)
1. **GitHub**: crear el repositorio, `git remote add origin <url>` y `git push -u origin main` (el ZIP trae `.git` con el historial). Pasarte la URL para ponerla en el PDF (sección 9).
2. ~~Actualizar Figma~~ **Hecho** (ver arriba).
3. **StarUML**: redibujar arquitectura, casos de uso, clases y paquetes (los `.puml` son la guía; StarUML no importa PlantUML). Reemplazar las imágenes del PDF.
4. Regenerar el PDF con `python3 docs/make_pdf.py docs/pdf_config.json` (las rutas ya son relativas; agregar `"github": "<url>"` en `docs/pdf_config.json` y las exportaciones de StarUML) y revisarlo visualmente.
5. Opcional: pruebas de componentes del frontend, OpenAPI, usuarios y roles; medir RNF-01 (2 s con 200 celdas y 5 usuarios) y RNF-02 (usabilidad), que hoy solo están definidos, no medidos.
6. Próximas iteraciones: administración de usuarios, clientes, vehículos y celdas; gestión de pagos (la entidad `mensualidad` ya está modelada).

## Cómo ejecutar
```bash
docker compose up -d db
cd frontend && npm ci && npm run build && cp -r dist ../backend/static && cd ..
cd backend && API_TOKEN=una-clave-larga-123 go run ./cmd/api     # http://localhost:8080
make test            # backend con -race + frontend
make test-integration  # con PostgreSQL real (base desechable)
```
API: `POST /api/ingresos`, `POST /api/salidas`, `POST /api/novedades`, `GET /api/celdas?estado=`, `GET /api/registros/activos`, `GET /api/sesion` (todas con `Authorization: Bearer <API_TOKEN>`).

## Notas sobre el entorno de trabajo (por si vuelves a usar herramientas)
- Un sandbox puede no tener acceso al proxy de módulos de Go (`proxy.golang.org`); funcionó `GOPROXY=direct GOSUMDB=off` para dependencias de github.com. No se pueden descargar archivos de figma.com desde el sandbox.
- Con un conector de Figma, leer primero las guías `figma-use` y `figma-generate-design` antes de escribir; los scripts de `use_figma` son atómicos (si fallan, no queda nada a medias).

Empieza resumiéndome en 5 líneas lo que entendiste y pregúntame cuál de los pendientes quiero atacar primero.
