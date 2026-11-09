# Parqueadero "Autos Colombia" — Iteración 2
## 2. Historias de Usuario (Gestión de Usuarios y Gestión de Celdas)

Cada historia sigue la estructura ágil estándar **Como [rol], quiero [capacidad], para [beneficio]**, con sus criterios de aceptación expresados en formato estructurado **Dado / Cuando / Entonces**, cubriendo tanto los caminos felices como los flujos de rechazo y excepciones implementados en el sistema.

---

### HU-06 · Gestión de usuarios del sistema
**Relacionada con:** RF-09, RF-10, RF-11, RF-14, RN-07, RN-08  
**Rol:** Administrador  

> **Como** Administrador del parqueadero,  
> **quiero** registrar, actualizar y administrar los usuarios del sistema asignando sus roles (`ADMINISTRADOR` o `ENCARGADO`) y estados,  
> **para** controlar quién puede operar el sistema y asegurar que cada persona tenga únicamente los permisos correspondientes a sus funciones.

* **CA-06.1 · Registro exitoso de usuario:**
  * **Dado** que el Administrador está autenticado y diligencia un login único de 3 a 30 caracteres (minúsculas, números, punto o guion), un nombre completo válido, una contraseña que cumple la política de robustez y un rol válido (`ADMINISTRADOR` o `ENCARGADO`).
  * **Cuando** el Administrador solicita la creación del usuario.
  * **Entonces** el sistema guarda el usuario activo, almacena la contraseña hasheada con PBKDF2-SHA256, genera un registro de auditoría (`accion: CREAR_USUARIO`) y retorna código HTTP 201 Created.

* **CA-06.2 · Rechazo por usuario duplicado:**
  * **Dado** que ya existe un usuario registrado con el mismo nombre de inicio de sesión (`usuario`).
  * **Cuando** el Administrador intenta registrar un nuevo usuario con ese mismo login.
  * **Entonces** el sistema rechaza la solicitud con HTTP 409 Conflict y el mensaje *"ya existe un usuario con ese nombre de usuario"*, sin alterar la base de datos.

* **CA-06.3 · Rechazo por contraseña débil:**
  * **Dado** que el Administrador ingresa una contraseña menor a 10 caracteres, o sin números, o sin letras, o igual al nombre de usuario.
  * **Cuando** se envía la solicitud de creación.
  * **Entonces** el sistema responde con HTTP 400 Bad Request indicando que la contraseña no cumple las políticas de longitud y complejidad exigidas.

* **CA-06.4 · Inactivación de usuario común:**
  * **Dado** un usuario activo con rol `ENCARGADO`.
  * **Cuando** el Administrador actualiza su estado a inactivo (`activo = false`).
  * **Entonces** el sistema persiste el cambio, registra el evento en auditoría y rechaza inmediatamente cualquier intento posterior de usar el token de dicho usuario con HTTP 401 Unauthorized.

* **CA-06.5 · Regla de protección del último administrador activo:**
  * **Dado** que en el sistema existe un único usuario activo con rol `ADMINISTRADOR`.
  * **Cuando** se intenta inactivar dicho usuario o cambiarle el rol a `ENCARGADO`.
  * **Entonces** el sistema bloquea la operación con HTTP 409 Conflict y el error *"debe quedar al menos un administrador activo"*, garantizando la gobernabilidad del sistema.

---

### HU-07 · Autenticación y cambio de credenciales
**Relacionada con:** RF-08, RF-12, RF-13, RN-07  
**Rol:** Usuario del sistema (Administrador o Encargado)  

> **Como** usuario del sistema,  
> **quiero** iniciar sesión con mis credenciales individuales y actualizar mi contraseña periódicamente,  
> **para** acceder a las funciones del parqueadero de manera protegida y mantener mi cuenta segura.

* **CA-07.1 · Inicio de sesión exitoso:**
  * **Dado** un usuario registrado y activo con su clave correcta.
  * **Cuando** envía sus credenciales al endpoint de autenticación.
  * **Entonces** el sistema emite un token de sesión seguro, informa la fecha de expiración, retorna los datos del perfil y reinicia el contador de intentos fallidos.

* **CA-07.2 · Credenciales erróneas:**
  * **Dado** un nombre de usuario inexistente o una contraseña incorrecta.
  * **Cuando** se realiza la solicitud de inicio de sesión.
  * **Entonces** el sistema responde con HTTP 401 Unauthorized y el mensaje genérico *"usuario o contraseña incorrectos"* (sin revelar cuál de los dos datos falló), incrementando el contador de intentos fallidos de la IP.

* **CA-07.3 · Bloqueo preventivo por fuerza bruta:**
  * **Dado** que una misma dirección IP acumula 5 intentos fallidos consecutivos en menos de 60 segundos.
  * **Cuando** realiza un nuevo intento de autenticación.
  * **Entonces** el sistema responde con HTTP 429 Too Many Requests, cabecera `Retry-After: 60` y el mensaje *"demasiados intentos fallidos; espere un minuto"*.

* **CA-07.4 · Cambio de contraseña personal exitoso:**
  * **Dado** un usuario autenticado que proporciona su contraseña actual correcta y una nueva contraseña que cumple la política de robustez.
  * **Cuando** solicita el cambio de contraseña.
  * **Entonces** el sistema actualiza el hash de la clave, registra la auditoría correspondiente y confirma el cambio con HTTP 200 OK.

* **CA-07.5 · Rechazo por confirmación errónea de clave actual:**
  * **Dado** un usuario autenticado que ingresa erróneamente su contraseña actual.
  * **Cuando** solicita el cambio de clave.
  * **Entonces** el sistema rechaza la operación con HTTP 400 Bad Request (*"la contraseña actual no es correcta"*), manteniendo la clave anterior intacta.

---

### HU-08 · Gestión y mantenimiento de celdas
**Relacionada con:** RF-15, RF-16, RF-17, RN-03, RN-09  
**Rol:** Administrador  

> **Como** Administrador del parqueadero,  
> **quiero** crear celdas y habilitar o inhabilitar bahías de estacionamiento por obras o mantenimiento,  
> **para** gestionar la capacidad real del patio y evitar que se asignen celdas no aptas para el parqueo.

* **CA-08.1 · Registro exitoso de celda:**
  * **Dado** que el Administrador suministra un código alfanumérico no existente de hasta 10 caracteres (ej. `A-01`, `M-02`) y un tipo de vehículo válido (`CARRO` o `MOTO`).
  * **Cuando** confirma la creación.
  * **Entonces** el sistema registra la nueva celda en estado `DISPONIBLE`, genera auditoría y responde con HTTP 201 Created.

* **CA-08.2 · Rechazo por celda duplicada:**
  * **Dado** que ya existe una celda registrada con el código `A-01`.
  * **Cuando** se intenta registrar otra celda con el código `A-01` (incluso variando minúsculas).
  * **Entonces** el sistema normaliza a mayúsculas y rechaza con HTTP 409 Conflict (*"ya existe una celda con ese código"*).

* **CA-08.3 · Inhabilitación de celda disponible por mantenimiento:**
  * **Dado** una celda en estado `DISPONIBLE`.
  * **Cuando** el Administrador cambia su estado a `INACTIVA`.
  * **Entonces** el sistema actualiza el estado a `INACTIVA`, excluye la celda del listado de celdas libres para ingresos operativos y registra el evento en la auditoría.

* **CA-08.4 · Bloqueo de inhabilitación o cambio en celda ocupada:**
  * **Dado** una celda en estado `OCUPADA` que aloja un vehículo en estancia.
  * **Cuando** el Administrador intenta cambiar su estado a `INACTIVA` o modificar su código/tipo.
  * **Entonces** el sistema bloquea la modificación con HTTP 409 Conflict y el mensaje *"la celda está ocupada; registre primero la salida del vehículo"*.

* **CA-08.5 · Reactivación de celda:**
  * **Dado** una celda en estado `INACTIVA` cuyos trabajos de mantenimiento finalizaron.
  * **Cuando** el Administrador actualiza su estado a `DISPONIBLE`.
  * **Entonces** el sistema habilita la celda nuevamente para que los encargados puedan asignarla en nuevos ingresos.

---

### HU-09 · Gestión de clientes, vehículos y mensualidades
**Relacionada con:** RF-18, RF-19, RN-01, RN-03, RN-10  
**Rol:** Administrador / Encargado  

> **Como** operador del parqueadero,  
> **quiero** registrar clientes con sus datos de contacto, asociarles sus vehículos y renovar sus mensualidades,  
> **para** asegurar que solo los vehículos autorizados con contrato vigente puedan ingresar al patio.

* **CA-09.1 · Registro de cliente y vehículos:**
  * **Dado** un cliente nuevo con documento, nombre completo, teléfono y correo electrónico válidos, junto con las placas de sus vehículos y el tipo de cada uno (`CARRO` o `MOTO`).
  * **Cuando** se guarda el cliente.
  * **Entonces** el sistema almacena la información, normaliza las placas al formato oficial y habilita al cliente en el sistema.

* **CA-09.2 · Adición de periodo de mensualidad válido:**
  * **Dado** un cliente existente y un nuevo rango de fechas (`fecha_inicio` y `fecha_fin`) que no se cruza con ninguna mensualidad previa del mismo cliente.
  * **Cuando** se registra el periodo de mensualidad.
  * **Entonces** el sistema crea la mensualidad y actualiza el indicador de vigencia del cliente.

* **CA-09.3 · Rechazo por mensualidad solapada:**
  * **Dado** un cliente que ya cuenta con una mensualidad del `2026-11-01` al `2026-11-30`.
  * **Cuando** se intenta registrar un periodo del `2026-11-15` al `2026-12-15`.
  * **Entonces** el sistema rechaza la operación con HTTP 409 Conflict y el mensaje *"la mensualidad se solapa con otra ya registrada para el cliente"*.

* **CA-09.4 · Modificación de tipo de vehículo fuera del patio:**
  * **Dado** un vehículo registrado que NO se encuentra dentro del parqueadero en ese momento.
  * **Cuando** el Administrador actualiza su tipo (ej. de `MOTO` a `CARRO`).
  * **Entonces** el sistema actualiza el tipo exitosamente y registra la acción en la bitácora de auditoría.

* **CA-09.5 · Rechazo de modificación de tipo con vehículo dentro:**
  * **Dado** un vehículo que tiene un ingreso activo en el patio (ocupando una celda de su tipo).
  * **When** se intenta cambiar su tipo.
  * **Then** el sistema rechaza el cambio con HTTP 409 Conflict (*"el vehículo está dentro del parqueadero; registre primero su salida"*), protegiendo la regla de compatibilidad de la celda.

---

### HU-10 · Auditoría inmutable de operaciones
**Relacionada con:** RF-20, RF-21, RN-11  
**Rol:** Administrador  

> **Como** Administrador del parqueadero,  
> **quiero** consultar la bitácora cronológica de acciones administrativas realizadas en el sistema,  
> **para** tener trazabilidad total, transparencia operativa y soporte ante auditorías o reclamaciones.

* **CA-10.1 · Trazabilidad automática y no repudio:**
  * **Dado** un usuario autenticado que realiza cualquier operación administrativa (crear usuario, cambiar rol, inhabilitar celda, editar cliente).
  * **Cuando** la transacción se confirma en la base de datos.
  * **Entonces** se inserta de manera automática una fila en la tabla `auditoria` con marca de tiempo UTC, ID y login del usuario, acción ejecutada, entidad afectada, ID de la entidad y detalles de los cambios.

* **CA-10.2 · Consulta de bitácora:**
  * **Dado** el Administrador autenticado consultando el endpoint `/api/auditoria`.
  * **Cuando** solicita el listado.
  * **Entonces** el sistema retorna los registros ordenados descendentemente por fecha/hora, mostrando quién realizó cada modificación.

* **CA-10.3 · Inmutabilidad estricta contra manipulación:**
  * **Dado** cualquier intento de ejecutar una sentencia `UPDATE` o `DELETE` sobre la tabla `auditoria` (incluso de forma directa en el motor de base de datos PostgreSQL).
  * **Cuando** se ejecuta la instrucción.
  * **Entonces** el trigger `trg_auditoria_inmutable` intercepta la orden y aborta la transacción con una excepción irrevocable: *"la auditoría no se modifica ni se elimina"*.

