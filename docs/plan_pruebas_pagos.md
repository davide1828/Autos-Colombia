# Plan resumido de pruebas — Gestión de pagos

**Proyecto:** Parqueadero Autos Colombia · **Iteración:** 3 · **Repositorio:** <https://github.com/davide1828/Autos-Colombia>

| Código | Escenario / pasos | Resultado esperado | Tipo / estado |
|---|---|---|---|
| CP-PG-01 | Registrar un pago con cliente existente, monto positivo, mes actual y medio EFECTIVO. | HTTP 201; quedan pago `PAGADO` y mensualidad completa del mismo mes. | Aplicación y HTTP · automatizado |
| CP-PG-02 | Enviar monto cero/negativo, cliente inválido o medio no admitido. | HTTP 400; no se crean filas de pago ni mensualidad. | Aplicación y HTTP · automatizado |
| CP-PG-03 | Enviar periodo mal formado o anterior al mes actual. | HTTP 400; no se altera la vigencia. | Aplicación y HTTP · automatizado |
| CP-PG-04 | Repetir el pago activo para cliente y mes ya pagados. | HTTP 409; no se crean duplicados. | Aplicación y PostgreSQL · automatizado |
| CP-PG-05 | Consultar historial de cliente con y sin pagos. | HTTP 200; pagos ordenados por periodo descendente; cliente sin pagos recibe `[]`. | HTTP / PostgreSQL · automatizado |
| CP-PG-06 | Descargar recibo PDF de un pago. | HTTP 200, `application/pdf`; número `AC-AAAA-######`, campos mínimos y huella SHA-256 reproducible. El documento indica que no es factura DIAN. | Aplicación y HTTP · automatizado |
| CP-PG-07 | Intentar ingreso sin periodo mensual vigente; luego registrar un pago vigente e intentar de nuevo. | Primer ingreso rechazado sin ocupar celda; tras pagar, ingreso permitido. | Aplicación y PostgreSQL · automatizado |
| CP-PG-08 | Cancelar pago vigente y revisar la mensualidad y el historial. | Pago queda `CANCELADO`; periodo asociado se revoca en la misma transacción; historial lo conserva; no permite nuevos ingresos sin otro periodo vigente. | PostgreSQL · automatizado |
| CP-PG-09 | Encargado cancela; administrador cancela; repetir cancelación. | Encargado recibe HTTP 403; administrador recibe HTTP 200; segunda cancelación recibe HTTP 409. | HTTP / roles · automatizado |
| CP-PG-10 | Registrar cliente nuevo con vehículo; repetir documento con placa nueva. | Se crea cliente y vehículo en el primer intento; el segundo conserva el cliente y agrega vehículo. Placa repetida se rechaza. | API de gestión · integración automatizada existente / revisión manual de UI |
| CP-PG-11 | Medir respuesta de registro, consulta, recibo y cancelación bajo carga representativa. | Cada operación completa en menos de 2 s; error 504 al superar el límite de 1,8 s del handler. | Rendimiento · pendiente de ejecutar en ambiente de despliegue |
| CP-PG-12 | Forzar error antes del `COMMIT` en el registro/cancelación y volver a consultar. | No queda pago parcial, mensualidad parcial ni cambio de estado parcial. | PostgreSQL · transacción implementada; inyección de fallo dedicada pendiente |

## Ejecución reproducible

- Unitarias y HTTP: `cd backend; go test ./...`.
- Frontend: `cd frontend; npm test; npm run lint; npm run build`.
- PostgreSQL real: definir `TEST_DATABASE_URL` con una **base desechable** y ejecutar `cd backend; go test ./internal/adapters/postgres -count=1`. Las pruebas recrean el esquema público; no usar una base con datos de trabajo.
- El límite temporal de 1,8 s es una protección de la API, no una certificación de rendimiento. CP-PG-11 requiere medición de carga con infraestructura y concurrencia representativas.

## Criterios de salida

La iteración se considera funcionalmente verificada cuando CP-PG-01 a CP-PG-10 pasan, el esquema de PostgreSQL incluye las restricciones del pago y `go test`, `npm test`, `npm run lint` y `npm run build` finalizan sin error. RNF-16 sigue pendiente de integración DIAN; CP-PG-11 requiere evidencia del ambiente objetivo.
