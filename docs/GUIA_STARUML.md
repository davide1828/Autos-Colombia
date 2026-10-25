# Guía: redibujar los 4 diagramas en StarUML

Objetivo: reemplazar las imágenes de PlantUML del PDF por diagramas hechos en StarUML (lo pide el enunciado de la EA1). Los `.puml` de `docs/diagramas/` son tu plano: aquí está todo lo que debes crear, en el orden recomendado.

> Los nombres de menús pueden variar un poco según la versión de StarUML (5 o 6). Si algo no está donde digo, busca la opción por su nombre en la barra de menús o en la caja de herramientas de la izquierda.

## 0. Preparación (una sola vez)

1. Abre StarUML y crea un proyecto nuevo. Guárdalo de inmediato como `docs/diagramas/autos_colombia.mdj` (así queda en el repositorio como evidencia).
2. En el **Model Explorer** (panel derecho) verás el modelo raíz. Los diagramas se crean desde el menú **Model → Add Diagram** o con clic derecho sobre un elemento del modelo.
3. Trabaja en este orden: **casos de uso → clases → paquetes → arquitectura**. Así reutilizas los nombres y no te contradices.
4. Reglas de estilo para todos: títulos en español, mismos nombres que en el PDF, sin diagramas desbordados (cabe en una página vertical, ancho máximo ~16 cm).

---

## 1. Casos de uso (Figura 6)

**Tipo de diagrama:** *Use Case Diagram*. Nómbralo `Casos de uso - Iteración 1: entrada y salida de vehículos`.

**Elementos a crear**

| Elemento | Nombre exacto |
|---|---|
| Subject (límite del sistema) | `Sistema de parqueadero Autos Colombia` |
| Actor | `Encargado` |
| Caso de uso | `CU-01 Consultar celdas disponibles` |
| Caso de uso | `CU-02 Registrar ingreso` |
| Caso de uso | `CU-03 Registrar salida` |
| Caso de uso | `CU-04 Registrar novedad` |
| Caso de uso | `CU-05 Consultar vehículos dentro` |
| Caso de uso | `Validar mensualidad vigente` |
| Caso de uso | `Validar celda en ocupación` |

**Relaciones**

- Asociación (línea simple) de `Encargado` a CU-01, CU-02, CU-03, CU-04 y CU-05.
- `CU-02 Registrar ingreso` **«include»** → `Validar mensualidad vigente`.
- `CU-02 Registrar ingreso` **«include»** → `Validar celda en ocupación`.
- `CU-04 Registrar novedad` **«extend»** → `CU-02 Registrar ingreso` (nota: "si hay novedad").
- `CU-04 Registrar novedad` **«extend»** → `CU-03 Registrar salida` (nota: "si hay novedad").

**Cuidado con la dirección de las flechas**

- *include*: la flecha sale del caso base (CU-02) y llega al incluido.
- *extend*: la flecha sale del caso que extiende (CU-04) y llega al caso base (CU-02 / CU-03).

**Verificación**: son 7 óvalos, 1 actor, 5 asociaciones, 2 include, 2 extend. Los códigos CU-01…CU-05 coinciden con la matriz de trazabilidad (sección 10 del PDF).

**Exportar como:** `casos_uso.png`.

---

## 2. Diagrama de clases (Figura 7)

**Tipo de diagrama:** *Class Diagram*. Nómbralo `Diagrama de clases - Dominio y servicio de aplicación`.

**Enumeraciones** (herramienta *Enumeration*; los valores se agregan en *Literals*)

| Enumeración | Literales |
|---|---|
| `TipoVehiculo` | CARRO, MOTO |
| `EstadoCelda` | DISPONIBLE, OCUPADA, INACTIVA |
| `TipoNovedad` | DANIO, FALTANTE, OTRA |
| `MomentoNovedad` | ENTRADA, SALIDA, DURANTE_ESTANCIA |

**Clases** (atributos privados «-», operaciones públicas «+»)

| Clase | Atributos | Operaciones |
|---|---|---|
| `Cliente` | id: int, documento: string, nombre: string | — |
| `Mensualidad` | id: int, fechaInicio: date, fechaFin: date | estaVigente(dia): bool |
| `Vehiculo` | id: int, placa: string, tipo: TipoVehiculo | — |
| `Celda` | id: int, codigo: string, tipo: TipoVehiculo, estado: EstadoCelda | puedeOcuparse(): bool |
| `Registro` | id: int, fechaEntrada: datetime, fechaSalida: datetime | estaActivo(): bool, duracion(ahora): duration |
| `Novedad` | id: int, tipo: TipoNovedad, descripcion: string, momento: MomentoNovedad, fecha: datetime | validar() |
| `Servicio` (estereotipo `aplicación`) | — | registrarIngreso(placa, idCelda, novedad), registrarSalida(placa, novedad), registrarNovedad(placa, novedad), consultarCeldas(estado), vehiculosDentro() |

**Interfaz y adaptador**

- `Repositorio` con la herramienta *Interface*, estereotipo `puerto`.
- `RepoPostgres` como clase, estereotipo `adaptador`.

**Asociaciones** (en *Properties* de cada extremo escribe la multiplicidad)

| Origen (mult.) | Destino (mult.) | Nombre |
|---|---|---|
| Cliente (1) | Vehiculo (0..*) | posee |
| Cliente (1) | Mensualidad (0..*) | contrata |
| Vehiculo (1) | Registro (0..*) | genera |
| Celda (1) | Registro (0..*) | aloja |
| Registro (1) | Novedad (0..*) | tiene |

**Dependencias** (flecha punteada, herramienta *Dependency*)

- Vehiculo → TipoVehiculo; Celda → EstadoCelda; Novedad → TipoNovedad; Novedad → MomentoNovedad.
- Servicio → Registro; Servicio → Novedad.
- `Servicio` → `Repositorio` con asociación dirigida y nombre `usa`.

**Realización**

- `RepoPostgres` → `Repositorio` con *Interface Realization* (flecha punteada con triángulo hueco).

**Verificación**: 7 clases + 1 interfaz + 1 adaptador + 4 enumeraciones; 5 asociaciones con multiplicidades; la regla "un único registro activo por vehículo y celda" va en el texto del PDF, no en el diagrama.

**Exportar como:** `clases.png`. Si queda muy ancho, organiza el modelo en tres filas: enumeraciones abajo, dominio al centro, `Servicio / Repositorio / RepoPostgres` arriba o a la izquierda.

---

## 3. Diagrama de paquetes (Figura 8)

StarUML no tiene un tipo "Package Diagram" propio: usa un **Class Diagram** (o *General Diagram*) y dentro la herramienta **Package**. Nómbralo `Diagrama de paquetes`.

**Paquetes de agrupación** (contenedores grandes)

- `Presentación`
- `Aplicación y dominio`
- `Persistencia`

**Paquetes internos** (dibújalos dentro del contenedor correspondiente)

| Dentro de | Paquete |
|---|---|
| Presentación | `frontend (React)` |
| Presentación | `adapters/httpapi` |
| Aplicación y dominio | `application` |
| Aplicación y dominio | `domain` |
| Persistencia | `adapters/postgres` |
| Persistencia | `db (schema.sql, seed.sql)` |
| (suelto, fuera de los grupos) | `cmd/api (arranque)` |

**Dependencias** (flecha punteada, herramienta *Dependency*)

| Desde | Hacia | Etiqueta |
|---|---|---|
| frontend (React) | adapters/httpapi | REST/JSON |
| adapters/httpapi | application | — |
| application | domain | — |
| adapters/postgres | application | implementa Repositorio |
| adapters/postgres | domain | — |
| adapters/postgres | db (schema.sql, seed.sql) | — |
| cmd/api (arranque) | adapters/httpapi | — |
| cmd/api (arranque) | adapters/postgres | — |
| cmd/api (arranque) | application | — |

**Verificación**: ninguna flecha sale de `domain` (el dominio no depende de nada); `application` solo depende de `domain`. Es el mensaje central del diagrama y coincide con el texto de la sección 8.

**Exportar como:** `paquetes.png`.

---

## 4. Arquitectura en capas (Figura 1)

StarUML no trae un diagrama de "capas"; usa un **Component Diagram** o un **Deployment Diagram** (más sencillo para mostrar el dispositivo del usuario). Recomendado: *Deployment Diagram*. Nómbralo `Arquitectura en capas (hexagonal) - Autos Colombia`.

**Elementos**

| Elemento (herramienta) | Nombre exacto |
|---|---|
| Actor | `Encargado` |
| Node | `Celular / Tablet / PC (navegador, PWA)` |
| Component (dentro del node) | `Frontend React (formularios, mensajes)` |
| Node o Package grande | `Servidor (Go)` |
| Component dentro del servidor | `Capa de presentación — Adaptador HTTP / API REST JSON (validación de entrada, token, códigos HTTP)` |
| Component dentro del servidor | `Capa de aplicación — Casos de uso: RegistrarIngreso, RegistrarSalida, RegistrarNovedad, Consultas` |
| Component dentro del servidor | `Capa de dominio — Entidades y reglas de negocio (Vehiculo, Celda, Registro, Novedad)` |
| Interface o Component dentro del servidor | `Puerto: Repositorio (interfaz)` |
| Component dentro del servidor | `Capa de persistencia — Adaptador PostgreSQL (database/sql)` |
| Node (o Component) fuera del servidor | `PostgreSQL` |

**Relaciones**

| Desde | Hacia | Etiqueta |
|---|---|---|
| Encargado | Frontend React | — |
| Frontend React | Capa de presentación | HTTPS / JSON |
| Capa de presentación | Capa de aplicación | — |
| Capa de aplicación | Capa de dominio | — |
| Capa de aplicación | Puerto: Repositorio | — |
| Capa de persistencia | Puerto: Repositorio | implementa (punteada, triángulo hueco) |
| Capa de persistencia | PostgreSQL | SQL |

**Verificación**: el dominio no apunta a la persistencia ni a HTTP; la persistencia *implementa* el puerto que define la aplicación (inversión de dependencias, la idea hexagonal del PDF).

**Exportar como:** `arquitectura.png`.

---

## 5. Exportar y reemplazar en el PDF

1. Para cada diagrama: clic en el fondo del diagrama → **File → Export Diagram As → Image (PNG)**. Opciones recomendadas: fondo blanco, escala 2x o más, sin márgenes enormes.
2. Guarda los cuatro archivos con **estos nombres exactos** dentro de `docs/diagramas/` (sobrescribiendo los de PlantUML):
   - `arquitectura.png`
   - `casos_uso.png`
   - `clases.png`
   - `paquetes.png`
3. Guarda también el proyecto `autos_colombia.mdj` en esa misma carpeta.
4. Regenera el PDF desde la raíz del proyecto:
   ```
   python3 docs/make_pdf.py docs/pdf_config.json
   ```
5. En el PDF, los pies de figura dicen "fuente PlantUML: docs/diagramas/…puml". Cámbialos por "fuente StarUML: docs/diagramas/autos_colombia.mdj" buscando `PlantUML` en `docs/make_pdf.py`.
6. Revisa las páginas 7, 12, 13 y 14: que ningún diagrama quede cortado, que el texto sea legible y que las flechas tengan la dirección correcta.

## 6. Lista final de control

- [ ] Los códigos CU-01 a CU-05 son los mismos que en historias y trazabilidad.
- [ ] Los nombres de clases, atributos y enumeraciones son idénticos a los del código (`domain/`) y al modelo físico.
- [ ] Multiplicidades del diagrama de clases = cardinalidades del modelo entidad-relación.
- [ ] El dominio no depende de nada (paquetes y arquitectura).
- [ ] Las 4 imágenes están en `docs/diagramas/` y el `.mdj` también.
- [ ] El PDF se regeneró y se revisó página por página.
