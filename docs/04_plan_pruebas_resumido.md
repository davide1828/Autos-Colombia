# Parqueadero "Autos Colombia" — Iteración 2
## 4. Formato de Plan de Pruebas Resumido
*(Gestión de Usuarios y Gestión de Celdas)*

### 4.1 Información General del Plan de Pruebas

* **Proyecto:** Sistema de Información del Parqueadero “Autos Colombia”
* **Fase:** Iteración 2 — Gestión de Usuarios, Roles (RBAC), Sesiones, Celdas y Auditoría
* **Responsable:** Cristian David España Figueroa
* **Herramientas de Ejecución:** `go test -race ./...` (Suite de pruebas unitarias y de integración), Vitest (Frontend)
* **Entorno de Pruebas:** Go 1.22+, PostgreSQL 16 (Docker Engine), Mock Repository en memoria
* **Estado General:** **100% Ejecutadas y Aprobadas (0 Fallos)**

---

### 4.2 Matriz del Plan de Pruebas Resumido

| ID | Módulo | Caso de Prueba | Precondiciones | Pasos de Ejecución | Datos de Entrada | Resultado Esperado | Resultado Obtenido | Estado |
|---|---|---|---|---|---|---|---|:---:|
| **CP-01** | **Seguridad** | Autenticación válida de usuario y emisión de sesión | Usuario registrado y activo en base de datos. | 1. Enviar solicitud POST a `/api/login`<br>2. Procesar hash PBKDF2<br>3. Generar token | `usuario: "admin"`, `clave: "ClaveSegura123"` | HTTP 200 OK con token de sesión firmado, tiempo de expiración y perfil del usuario. | Token emitido correctamente; contador de intentos reiniciado. | **Aprobado** |
| **CP-02** | **Seguridad** | Rechazo por credenciales erróneas | Usuario registrado o inexistente. | 1. Enviar credenciales incorrectas a `/api/login`<br>2. Comparar hash en tiempo constante | `usuario: "admin"`, `clave: "ClaveInvalida"` | HTTP 401 Unauthorized (*"usuario o contraseña incorrectos"*). Se incrementa el contador de fallos de la IP. | HTTP 401 retornado; mensaje genérico sin filtrar si falló el login o la clave. | **Aprobado** |
| **CP-03** | **Seguridad** | Bloqueo por fuerza bruta tras 5 intentos fallidos | IP sin bloqueo previo. | 1. Enviar 5 solicitudes consecutivas con clave errónea en < 60s<br>2. Enviar 6to intento | 6 solicitudes desde la misma IP | HTTP 429 Too Many Requests con cabecera `Retry-After: 60` y mensaje preventivo. | HTTP 429 retornado en el 6to intento; acceso bloqueado por 1 minuto. | **Aprobado** |
| **CP-04** | **Usuarios** | Registro exitoso de nuevo usuario por Administrador | Administrador autenticado con token Bearer. | 1. Enviar POST a `/api/usuarios`<br>2. Validar formato login y contraseña<br>3. Hashear clave e insertar | `usuario: "operador1"`, `nombre: "Carlos Ruiz"`, `rol: "ENCARGADO"`, `clave: "ClaveOperador2026"` | HTTP 201 Created. Usuario guardado con `activo=true` y evento registrado en auditoría. | Usuario persistido; hash PBKDF2 almacenado con sal única; auditoría generada. | **Aprobado** |
| **CP-05** | **Usuarios** | Rechazo por nombre de usuario duplicado | Ya existe un usuario con login `"admin"`. | 1. Intentar registrar nuevo usuario con login `"admin"` | `usuario: "admin"`, `nombre: "Otro Admin"`, `rol: "ADMINISTRADOR"`, `clave: "ClaveAdmin999"` | HTTP 409 Conflict (*"ya existe un usuario con ese nombre de usuario"*). | HTTP 409 retornado; se mantiene la unicidad en la tabla `usuario`. | **Aprobado** |
| **CP-06** | **Usuarios** | Rechazo por contraseña débil o inválida | Administrador autenticado. | 1. Enviar clave menor a 10 caracteres o sin dígitos | `usuario: "cajero"`, `clave: "corta1"` | HTTP 400 Bad Request indicando que la contraseña no cumple la política de robustez. | HTTP 400 retornado con mensaje descriptivo de la política de seguridad. | **Aprobado** |
| **CP-07** | **Usuarios** | Regla de protección del último administrador activo | Solo existe 1 administrador activo en el sistema. | 1. Administrador intenta inactivar su propia cuenta o cambiar su rol a ENCARGADO | `id: 1`, `activo: false` o `rol: "ENCARGADO"` | HTTP 409 Conflict (*"debe quedar al menos un administrador activo"*). Operación abortada. | Error atómico `ErrUltimoAdministrador`; la base de datos mantiene al administrador activo. | **Aprobado** |
| **CP-08** | **Usuarios** | Cambio personal de contraseña con validación de clave actual | Usuario autenticado en el sistema. | 1. Enviar POST a `/api/sesion/clave`<br>2. Validar clave anterior<br>3. Hashear nueva clave | `clave_actual: "ClaveVieja123"`, `clave_nueva: "ClaveNuevaSegura2026"` | HTTP 200 OK. Contraseña actualizada; auditoría registrada. | Hash actualizado exitosamente; la sesión previa exige nueva validación. | **Aprobado** |
| **CP-09** | **Celdas** | Registro exitoso de nueva celda | Administrador autenticado. | 1. Enviar POST a `/api/celdas`<br>2. Validar código y tipo | `codigo: "A-05"`, `tipo: "CARRO"` | HTTP 201 Created. Celda registrada en estado `DISPONIBLE`; auditoría registrada. | Celda creada correctamente con código normalizado en mayúsculas. | **Aprobado** |
| **CP-10** | **Celdas** | Rechazo por código de celda duplicado | Existe la celda con código `"A-01"`. | 1. Intentar crear celda con código `"a-01"` | `codigo: "a-01"`, `tipo: "CARRO"` | HTTP 409 Conflict (*"ya existe una celda con ese código"*). | Código normalizado a mayúsculas y rechazado por colisión única. | **Aprobado** |
| **CP-11** | **Celdas** | Inhabilitación de celda disponible por mantenimiento | Celda en estado `DISPONIBLE` sin vehículos dentro. | 1. Enviar PUT a `/api/celdas/1/estado`<br>2. Validar que no tenga registro activo | `estado: "INACTIVA"` | HTTP 200 OK. Celda pasa a `INACTIVA` y se excluye del listado operativo para ingresos. | Celda inhabilitada; evento reflejado en bitácora de auditoría. | **Aprobado** |
| **CP-12** | **Celdas** | Bloqueo de inhabilitación sobre celda ocupada | Celda con vehículo alojado en estancia activa. | 1. Intentar cambiar estado de la celda a `INACTIVA` | `id_celda: 2 (ocupada)`, `estado: "INACTIVA"` | HTTP 409 Conflict (*"la celda está ocupada; registre primero la salida del vehículo"*). | Rechazo inmediato; la celda permanece en estado `OCUPADA` garantizando la integridad. | **Aprobado** |
| **CP-13** | **Clientes** | Rechazo de mensualidad con fechas solapadas | Cliente con mensualidad del 01/11 al 30/11. | 1. Intentar agregar periodo del 15/11 al 15/12 para el mismo cliente | `id_cliente: 1`, `fecha_inicio: "2026-11-15"`, `fecha_fin: "2026-12-15"` | HTTP 409 Conflict (*"la mensualidad se solapa con otra ya registrada para el cliente"*). | Validación de solapamiento exitosa; no se registran rangos duplicados. | **Aprobado** |
| **CP-14** | **Auditoría** | Inmutabilidad de la tabla de auditoría ante manipulación | Existen registros en la tabla `auditoria`. | 1. Intentar ejecutar `UPDATE auditoria SET accion='OTRA'` o `DELETE FROM auditoria` | Sentencia SQL directa sobre PostgreSQL | El trigger `trg_auditoria_inmutable` aborta la transacción arrojando excepción. | Excepción lanzada: *"la auditoría no se modifica ni se elimina"*. Registros intactos. | **Aprobado** |

---

### 4.3 Evidencia de Ejecución en Consola

```bash
$ cd backend && go test -v -race ./internal/application ./internal/adapters/httpapi
=== RUN   TestAutenticar_Valido
--- PASS: TestAutenticar_Valido (0.01s)
=== RUN   TestAutenticar_Rechazos
--- PASS: TestAutenticar_Rechazos (0.01s)
=== RUN   TestCrearUsuario_Valido
--- PASS: TestCrearUsuario_Valido (0.01s)
=== RUN   TestUltimoAdministradorProtegido
--- PASS: TestUltimoAdministradorProtegido (0.00s)
=== RUN   TestUltimoAdministrador_Simultaneo
--- PASS: TestUltimoAdministrador_Simultaneo (0.02s)
=== RUN   TestCrearCelda
--- PASS: TestCrearCelda (0.00s)
=== RUN   TestActualizarCelda
--- PASS: TestActualizarCelda (0.00s)
=== RUN   TestAgregarMensualidad_Solapes
--- PASS: TestAgregarMensualidad_Solapes (0.00s)
PASS
ok      autoscolombia/parqueadero/internal/application      0.403s
ok      autoscolombia/parqueadero/internal/adapters/httpapi 1.078s
```

