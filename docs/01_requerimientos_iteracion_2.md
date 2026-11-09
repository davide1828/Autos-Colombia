# Parqueadero "Autos Colombia" — Iteración 2
## 1. Requerimientos Funcionales y No Funcionales (Gestión de Usuarios y Gestión de Celdas)

### 1.1 Contexto y Reglas de Negocio de la Iteración 2

En la **Iteración 1** se abordó la operación diaria del patio (registro de ingreso, salida y novedades con un encargado y clave de acceso compartida). En esta **Iteración 2**, el sistema evoluciona hacia una plataforma multiusuario con control de acceso basado en roles (**RBAC**), administración de usuarios del sistema, gestión completa del inventario de celdas, trazabilidad con auditoría inmutable y soporte a clientes y vehículos asociados.

#### Reglas de Negocio (Continuación de EA1)
* **RN-01 (Mensualidad vigente):** La fecha de fin debe ser mayor o igual a la fecha actual evaluada en la zona horaria legal de Colombia (UTC-5).
* **RN-02 (Unicidad de estancia):** Un vehículo no puede tener más de un registro de ingreso activo en el parqueadero.
* **RN-03 (Compatibilidad y exclusividad de celda):** Una celda solo puede estar ocupada por un vehículo a la vez. El tipo de la celda (`CARRO` o `MOTO`) debe coincidir exactamente con el tipo del vehículo. En la base de datos se garantiza atómicamente mediante el trigger `trg_registro_tipos`.
* **RN-04 (Cierre y liberación):** La salida libera inmediatamente la celda a estado `DISPONIBLE` y calcula el tiempo de estancia.
* **RN-05 (Registro de novedades):** Las novedades pueden registrarse al ingresar, al salir o durante la estancia.
* **RN-06 (Inmutabilidad del histórico):** Los registros operativos y las novedades nunca se eliminan del sistema.
* **RN-07 (Autenticación y sesión segura):** El acceso al sistema requiere credenciales individuales. Las contraseñas deben cumplir políticas de robustez (mínimo 10 caracteres, letra y número, distinta al login), hasheadas con salt criptográfico individual (PBKDF2-SHA256). Se aplica limitación de intentos fallidos por IP (máx. 5 intentos/minuto).
* **RN-08 (Garantía del último administrador activo):** El sistema debe impedir cualquier operación (inactivación, cambio de rol o eliminación) que deje al sistema sin al menos un usuario activo con rol `ADMINISTRADOR`.
* **RN-09 (Integridad de celdas ocupadas):** Una celda en estado `OCUPADA` no puede pasar a estado `INACTIVA` ni sufrir modificaciones en su código o tipo mientras tenga un vehículo alojado. Debe registrarse primero la salida.
* **RN-10 (No solapamiento de periodos de mensualidad):** Ningún cliente puede tener periodos de mensualidad superpuestos entre sí.
* **RN-11 (Auditoría administrativa inmutable):** Toda acción administrativa (creación, edición o cambio de estado de usuarios, celdas y vehículos) debe generar un registro de auditoría (quién, cuándo, acción, entidad y detalle). La tabla de auditoría rechaza modificaciones y eliminaciones mediante trigger (`trg_auditoria_inmutable`).

---

### 1.2 Requerimientos Funcionales (RF)

| ID | Requerimiento Funcional | Regla(s) Asociada(s) | Prioridad |
|---|---|---|:---:|
| **RF-08** | **Autenticación de usuarios:** El sistema debe permitir el inicio de sesión a usuarios registrados (Administrador o Encargado) mediante usuario y contraseña, emitiendo un token de sesión seguro (Bearer/JWT) con tiempo de expiración. | RN-07 | Alta |
| **RF-09** | **Control de acceso basado en roles (RBAC):** El sistema debe verificar el rol del usuario en cada petición HTTP. Las operaciones de administración (gestión de usuarios, celdas, vehículos, clientes y consulta de auditoría) son exclusivas de `ADMINISTRADOR`. El rol `ENCARGADO` solo opera entradas, salidas, novedades, patio y el cambio de su propia clave. | RN-07 | Alta |
| **RF-10** | **Registro de usuarios del sistema:** El sistema debe permitir al Administrador registrar nuevos usuarios indicando nombre completo (1 a 100 caracteres), nombre de usuario único en minúsculas (3 a 30 caracteres alfanuméricos y separadores), rol (`ADMINISTRADOR` o `ENCARGADO`) y contraseña inicial segura. | RN-07 | Alta |
| **RF-11** | **Modificación y estado de usuarios:** El sistema debe permitir al Administrador modificar el nombre, rol y estado (`activo`/`inactivo`) de un usuario, garantizando que siempre permanezca al menos un administrador activo en el sistema. | RN-08 | Alta |
| **RF-12** | **Cambio de contraseña por el propio usuario:** El sistema debe permitir a cualquier usuario autenticado actualizar su contraseña personal validando su contraseña actual y la política de fortaleza de la nueva clave. | RN-07 | Media |
| **RF-13** | **Restablecimiento administrativo de contraseñas:** El sistema debe permitir al Administrador restablecer la contraseña de cualquier usuario sin necesidad de conocer la anterior, aplicando las políticas de seguridad. | RN-07 | Media |
| **RF-14** | **Revocación inmediata de sesión:** El sistema debe invalidar inmediatamente la sesión activa de un usuario si este es inactivado o si se le modifica su rol, consultando la vigencia en la base de datos en cada petición protegida. | RN-07, RN-08 | Alta |
| **RF-15** | **Registro de celdas:** El sistema debe permitir al Administrador registrar celdas asignando un código único alfanumérico en mayúsculas de hasta 10 caracteres (ej. `A-01`, `M-05`) y el tipo de vehículo permitido (`CARRO` o `MOTO`). | RN-03 | Alta |
| **RF-16** | **Gestión de estado de celdas (Habilitar / Inhabilitar):** El sistema debe permitir al Administrador alternar el estado operativo de una celda entre `DISPONIBLE` e `INACTIVA` (por mantenimiento o daño), impidiendo inhabilitarla si la celda está ocupada. | RN-09 | Alta |
| **RF-17** | **Modificación de características de celda:** El sistema debe permitir al Administrador editar el código o tipo de una celda, validando que la celda no se encuentre ocupada al momento del cambio. | RN-03, RN-09 | Media |
| **RF-18** | **Gestión de clientes y mensualidades:** El sistema debe permitir registrar clientes con su documento, nombre, teléfono y correo, asociarles vehículos, y registrar mensualidades validando que no se solapen periodos de cobertura para el mismo cliente. | RN-01, RN-10 | Alta |
| **RF-19** | **Actualización de tipo de vehículo con guardia de estancia:** El sistema debe permitir al Administrador corregir el tipo de un vehículo (`CARRO`/`MOTO`), validando que el vehículo no tenga un ingreso activo en el patio para no corromper la compatibilidad con su celda asignada. | RN-03 | Media |
| **RF-20** | **Bitácora de auditoría inmutable:** El sistema debe registrar de forma automática y transparente cada acción de alta, modificación o cambio de estado sobre usuarios, celdas y vehículos, almacenando momento UTC, usuario responsable, entidad, identificador y detalle; impidiendo la edición o eliminación de los registros. | RN-11 | Alta |
| **RF-21** | **Consulta de auditoría:** El sistema debe permitir al Administrador consultar el histórico cronológico de acciones de auditoría registradas en el sistema. | RN-11 | Media |

---

### 1.3 Requerimientos No Funcionales (RNF)

| ID | Categoría | Requerimiento No Funcional Verificable | Métrica / Criterio de Aceptación |
|---|---|---|---|
| **RNF-08** | **Seguridad Criptográfica** | Las contraseñas de los usuarios nunca se guardan en texto claro. Deben protegerse mediante algoritmo PBKDF2 con función pseudoaleatoria HMAC-SHA256, sal criptográfica aleatoria de 16 bytes y un factor de costo de al menos 100.000 iteraciones. | Cobertura al 100% en pruebas del paquete `adapters/seguridad`. Verificación en tiempo constante (`subtle.ConstantTimeCompare`). |
| **RNF-09** | **Seguridad Perimetral y Fuerza Bruta** | El sistema debe limitar las solicitudes fallidas de inicio de sesión por dirección IP origen para mitigar ataques de fuerza bruta. | Máximo 5 intentos erróneos por minuto. Si se excede, el servidor retorna HTTP 429 (`Too Many Requests`) con cabecera `Retry-After: 60`. |
| **RNF-10** | **Integridad y Concurrencia Transaccional** | Las operaciones de actualización crítica (modificación de celda, asignación de vehículo, validación del último administrador) deben ejecutarse en transacciones ACID con bloqueos pesimistas (`SELECT ... FOR SHARE` / `FOR UPDATE`). | Comprobado con pruebas de estrés concurrentes simulando solicitudes paralelas contra PostgreSQL. |
| **RNF-11** | **Trazabilidad y No Repudio** | El registro de auditoría debe ser inalterable a nivel de motor de base de datos. | Trigger `trg_auditoria_inmutable` en PostgreSQL que arroja excepción ante cualquier intento de `UPDATE` o `DELETE` sobre la tabla `auditoria`. |
| **RNF-12** | **Desempeño en Autenticación y Autorización** | La validación de credenciales y la verificación de token Bearer en cada solicitud no deben incrementar la latencia promedio del backend en más de 5 ms. | Medición mediante pruebas de API con servidor en memoria y en contenedor. |
| **RNF-13** | **Mantenibilidad y Arquitectura Limpia** | La separación en capas (Arquitectura Hexagonal: `domain`, `application`, `adapters`) debe mantenerse sin acoplamientos inversos. Los casos de uso de gestión no deben depender de la tecnología de transporte (HTTP) ni del motor de base de datos específico. | Verificación mediante inyección de dependencias en `application/gestion.go` y doble implementación de persistencia (`postgres` y `memory`). |

