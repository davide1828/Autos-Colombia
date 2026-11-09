# Parqueadero "Autos Colombia" — Iteración 2
## 3. Diagrama y Especificación de Casos de Uso (StarUML)
*(Gestión de Usuarios y Gestión de Celdas)*

### 3.1 Descripción del Modelo y Actores

En esta segunda iteración se incorporan los roles de usuario y las funciones administrativas del sistema. Se modelan dos actores humanos principales:

1. **Administrador:** Responsable de la gestión de identidades y accesos (alta de usuarios, asignación de roles `ADMINISTRADOR` o `ENCARGADO`, activación/inactivación y reseteo de claves), parametrización de celdas (alta, cambio de tipo y bloqueo preventivo por mantenimiento), administración de contratos de mensualidad y consulta inmutable de la bitácora de auditoría.
2. **Encargado:** Operador del patio responsable de los flujos diarios de la Iteración 1 (ingreso, salida, novedades y patio), quien en esta iteración cuenta con credenciales individuales de acceso, inicio de sesión seguro y capacidad de actualizar su contraseña personal.

---

### 3.2 Diagrama de Casos de Uso (Iteración 2)

* **Archivo de proyecto StarUML:** [`docs/diagramas/parqueadero_iteracion_2.mdj`](file:///c:/universidad/ingenieria%20de%20software/parqueadero_EA1_codigo/parqueadero/docs/diagramas/parqueadero_iteracion_2.mdj)
* **Fuente PlantUML:** [`docs/diagramas/08_casos_uso_iteracion_2.puml`](file:///c:/universidad/ingenieria%20de%20software/parqueadero_EA1_codigo/parqueadero/docs/diagramas/08_casos_uso_iteracion_2.puml)
* **Imagen de alta resolución exportada:** [`docs/diagramas/casos_uso_iteracion_2.png`](file:///c:/universidad/ingenieria%20de%20software/parqueadero_EA1_codigo/parqueadero/docs/diagramas/casos_uso_iteracion_2.png)

---

### 3.3 Especificación Detallada de los Casos de Uso

| Caso de Uso | Actores | Descripción | Precondición | Postcondición | Relaciones `<<include>>` |
|---|---|---|---|---|---|
| **CU-06 Iniciar sesión** | Administrador, Encargado | Permite al usuario autenticarse en el sistema con su nombre de usuario y contraseña para obtener un token de sesión seguro. | Usuario registrado y en estado `activo`. | Token emitido y contador de intentos fallidos reiniciado. | `<<include>>` Validar credenciales y tasa de intentos. |
| **CU-07 Cambiar propia contraseña** | Administrador, Encargado | Permite a cualquier usuario autenticado actualizar su contraseña personal confirmando su clave actual. | Sesión activa y válida. | Hash de la contraseña actualizado; sesiones previas invalidadas. | `<<include>>` Validar política de robustez de contraseña. |
| **CU-08 Administrar usuarios del sistema** | Administrador | Permite registrar nuevos operadores, actualizar nombres, modificar roles, inactivar cuentas y resetear claves. | Administrador autenticado. | Usuario persistido/actualizado; evento registrado en auditoría. | `<<include>>` Validar política de contraseña.<br>`<<include>>` Garantizar último admin.<br>`<<include>>` Registrar auditoría inmutable. |
| **CU-09 Administrar celdas** | Administrador | Permite dar de alta celdas con código y tipo (`CARRO`/`MOTO`), editar características y alternar estado (`DISPONIBLE`/`INACTIVA`). | Administrador autenticado. | Celda creada o actualizada; auditoría registrada. | `<<include>>` Validar celda no ocupada.<br>`<<include>>` Registrar auditoría inmutable. |
| **CU-10 Administrar clientes y mensualidades** | Administrador, Encargado | Permite asociar vehículos a clientes y registrar nuevos periodos de mensualidad sin cruce de fechas. | Usuario autenticado (gestión de clientes) / Administrador (edición de tipos). | Contrato registrado; vigencia actualizada. | `<<include>>` Validar no solapamiento de fechas.<br>`<<include>>` Validar vehículo fuera de patio.<br>`<<include>>` Registrar auditoría inmutable. |
| **CU-11 Consultar bitácora de auditoría** | Administrador | Permite consultar el historial cronológico descendente de acciones administrativas (quién, cuándo, entidad, detalle). | Administrador autenticado. | Registros devueltos sin alteración. | Ninguna (consulta pura de sólo lectura). |

---

### 3.4 Justificación de los Casos de Uso Incluidos (`<<include>>`)

1. **`Validar credenciales y tasa de intentos`:** Compartido en el login para verificar hash PBKDF2 y bloquear IPs que acumulen más de 5 fallos por minuto (RN-07).
2. **`Validar política de robustez de contraseña`:** Reutilizado tanto en el cambio de contraseña personal (CU-07) como en la creación/reseteo administrativo (CU-08). Exige 10-128 caracteres, letras, números y disimilitud con el login.
3. **`Garantizar al menos un administrador activo`:** Sub-flujo obligatorio en CU-08. Bloquea cualquier actualización que reduzca a cero el número de administradores activos del sistema (RN-08).
4. **`Validar celda no ocupada`:** Sub-flujo en CU-09. Impide que una celda en uso pase a `INACTIVA` o cambie de tipo/código mientras un vehículo esté dentro (RN-09).
5. **`Validar no solapamiento de fechas de mensualidad`:** Verifica que los rangos de vigencia no se crucen para el mismo cliente (RN-10).
6. **`Validar vehículo fuera del parqueadero`:** Impide alterar el tipo de vehículo (`CARRO`/`MOTO`) mientras mantenga una estancia activa en el patio (RN-03).
7. **`Registrar evento de auditoría inmutable`:** Transversal a toda operación de escritura administrativa (CU-08, CU-09, CU-10), asegurando trazabilidad no repudiable (RN-11).

