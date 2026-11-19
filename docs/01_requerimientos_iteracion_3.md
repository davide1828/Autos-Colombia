# Iteración 3 — Gestión de pagos

## Alcance y reglas de negocio

El encargado registra manualmente el valor recibido, el mes que cubre y el medio de pago. El sistema no procesa tarjetas ni confirma transferencias con una pasarela: el operador registra el pago después de verificarlo. El periodo cubierto es un mes calendario completo, el mes actual o uno futuro. El valor se captura en pesos colombianos enteros y debe ser mayor que cero.

Al registrar el pago, el sistema crea en una sola transacción el pago y su mensualidad. No admite dos pagos activos para el mismo cliente y mes, ni periodos que se solapen con una mensualidad ya registrada. La vigencia de la mensualidad controla los nuevos ingresos: un vehículo no puede entrar sin una mensualidad vigente. Una cancelación conserva el pago en el historial, revoca la mensualidad vinculada y no expulsa un vehículo que ya esté dentro; sí impide nuevos ingresos mientras no exista otra mensualidad vigente.

El sistema genera un recibo PDF interno, numerado `AC-AAAA-######`, con un resumen de la operación y una huella SHA-256 reproducible. **Esta huella no es una firma digital ni acredita facturación electrónica DIAN.** Para cumplir esa obligación se requiere integrar un proveedor autorizado y configurar certificado y credenciales tributarias, elementos que no forman parte del alcance acordado para esta entrega.

## Requerimientos funcionales

| ID | Requerimiento | Criterio de aceptación |
|---|---|---|
| RF-22 | Registrar el pago de una mensualidad para un cliente existente. | Un pago válido guarda cliente, monto, mes, fecha y medio de pago, y activa la mensualidad del mismo periodo en una transacción. Montos no positivos, periodo pasado, cliente inexistente y medios no admitidos se rechazan. |
| RF-23 | Consultar el historial de pagos de un cliente. | El historial presenta periodo, valor, medio y estado, ordenados desde el periodo más reciente. Un cliente sin pagos recibe una lista vacía. |
| RF-24 | Generar y descargar el recibo del pago en PDF. | El PDF contiene identificador, nombre y documento del cliente, valor, fecha, periodo, medio, estado y huella SHA-256. Se identifica claramente como recibo interno, no factura DIAN. |
| RF-25 | Impedir el ingreso cuando el cliente no tiene mensualidad vigente. | La validación de ingreso rechaza la operación sin ocupar la celda. Registrar una mensualidad pagada y vigente permite el ingreso. |
| RF-26 | Cancelar un pago y revocar su mensualidad asociada. | Solo un administrador puede cancelar. El pago queda en el historial como cancelado, se elimina la mensualidad activa asociada y la operación se confirma atómicamente. No se puede cancelar dos veces. |
| RF-27 | Registrar clientes y asociarles vehículos. | Encargado o administrador pueden registrar un cliente con su primer vehículo o asociar un vehículo adicional por documento. El documento existente conserva los datos del cliente ya guardados. |
| RF-28 | Crear y administrar cuentas de operadores. | El administrador puede registrar encargados y otros administradores desde la sección Usuarios existente; se mantienen validación de clave y permisos por rol. |

## Requerimientos no funcionales

| ID | Requerimiento | Implementación / evidencia |
|---|---|---|
| RNF-14 | Las rutas de pagos deben finalizar en menos de 2 segundos. | Las solicitudes tienen un límite de ejecución de 1,8 s. Si vence, la API responde `504`. El límite acota el tiempo de proceso; falta ejecutar una prueba de carga en el ambiente de despliegue para certificar el objetivo bajo concurrencia. |
| RNF-15 | Los cambios financieros y de vigencia deben ser atómicos y persistentes. | PostgreSQL registra pago y mensualidad dentro de una transacción; la cancelación elimina la mensualidad y cambia el estado del pago dentro de otra. Índices y restricciones evitan duplicados activos. |
| RNF-16 | Facturación electrónica de acuerdo con la normativa vigente. | **Pendiente / no satisfecho por esta versión.** El sistema produce un recibo interno con huella SHA-256, no una factura firmada ni validada por DIAN. Requiere proveedor autorizado, certificado y credenciales tributarias. |
| RNF-17 | Proteger las rutas de pagos y las acciones administrativas. | Las rutas requieren sesión; consulta y registro permiten los roles de administrador y encargado; cancelación queda restringida al administrador. |

## Historias de usuario

### HU-11 — Registrar pago y habilitar mensualidad

**Como encargado**, quiero registrar el pago mensual de un cliente para dejar constancia del recaudo y habilitar sus ingresos durante el periodo pagado.

- **Dado** un cliente existente sin mensualidad superpuesta, **cuando** registro un monto entero positivo, un medio permitido y el mes actual o futuro, **entonces** se guardan el pago y la mensualidad en una sola operación.
- **Dado** un pago ya registrado para ese cliente y mes, **cuando** se intenta registrarlo otra vez, **entonces** se rechaza como conflicto y no se genera una segunda mensualidad.
- **Dado** un cliente sin mensualidad vigente, **cuando** intenta ingresar, **entonces** el ingreso se rechaza y la celda no se ocupa.

### HU-12 — Consultar pagos, recibos y cancelaciones

**Como administrador**, quiero revisar el historial, descargar recibos internos y cancelar un pago incorrecto para conservar trazabilidad y corregir su vigencia.

- **Dado** un pago, **cuando** solicito el recibo, **entonces** recibo un PDF descargable con su número y huella SHA-256.
- **Dado** un pago vigente, **cuando** lo cancelo como administrador, **entonces** permanece en el historial con estado `CANCELADO` y su mensualidad se revoca.
- **Dado** un pago cancelado, **cuando** se intenta cancelarlo de nuevo, **entonces** la API informa conflicto y no altera los datos.
- **Dado** un encargado, **cuando** intenta cancelar, **entonces** recibe `403`.

### HU-13 — Registrar clientes y vehículos

**Como encargado**, quiero crear un cliente y asociarle un vehículo, o asociar otro vehículo a un cliente existente, para mantener al día el padrón que usa la operación y los pagos.

- **Dado** un documento nuevo y datos válidos del cliente y vehículo, **cuando** guardo el formulario, **entonces** se registra el cliente con su vehículo.
- **Dado** un documento ya registrado y una placa nueva, **cuando** guardo el formulario, **entonces** se conserva el cliente existente y se asocia el nuevo vehículo.
- **Dado** una placa duplicada o datos inválidos, **cuando** guardo, **entonces** se informa el error y no se duplica el registro.

### HU-14 — Crear cuentas de operadores

**Como administrador**, quiero crear cuentas de encargado o administrador para dar acceso al personal autorizado. La funcionalidad ya disponible en la sección **Usuarios** se mantiene en esta iteración.

## Diagramas y artefactos relacionados

- Casos de uso: `diagramas/14_casos_uso_pagos.puml`.
- Secuencia del pago y sus alternativas: `diagramas/16_secuencia_gestion_pagos.puml`.
- Componentes actualizados para Iteraciones 1, 2 y 3: `diagramas/13_diagrama_componentes.puml`.
- Proyecto StarUML: `diagramas/parqueadero_iteracion_3.mdj`.
- Mockups: `pantallas/16_gestion_pagos.svg` y `pantallas/17_registro_cliente_vehiculo.svg`.
- Casos de prueba: `plan_pruebas_pagos.md`.
- Trazabilidad: `trazabilidad_iteracion_3.md`.

## Referencias

- Código fuente y evidencias de versión: <https://github.com/davide1828/Autos-Colombia>.
- Las cuentas de usuario del sistema y la administración existente de usuarios se documentan en `frontend/src/views/UsuariosAdmin.jsx`; el registro de clientes y vehículos está en `frontend/src/views/ClientesVehiculos.jsx`.
