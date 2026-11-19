# Matriz de trazabilidad — Iteración 3

| Requisito | Historia | Diseño / prueba | Implementación |
|---|---|---|---|
| RF-22 Registrar mensualidad pagada, validar datos y evitar repetidos | HU-11 | `diagramas/14_casos_uso_pagos.puml`; `diagramas/16_secuencia_gestion_pagos.puml`; CP-PG-01 a CP-PG-04 | `backend/internal/domain/pago.go`; `backend/internal/application/pago.go`; `backend/internal/adapters/httpapi/pago.go`; `backend/internal/adapters/postgres/pago.go`; `backend/db/schema.sql`; `frontend/src/views/PagosAdmin.jsx` |
| RF-23 Consultar historial por cliente | HU-12 | `diagramas/14_casos_uso_pagos.puml`; CP-PG-05 | `backend/internal/adapters/postgres/pago.go`; `backend/internal/adapters/httpapi/pago.go`; `frontend/src/views/PagosAdmin.jsx` |
| RF-24 Descargar recibo PDF interno, numerado y con huella | HU-12 | `diagramas/16_secuencia_gestion_pagos.puml`; CP-PG-06 | `backend/internal/application/pago.go`; `backend/internal/adapters/httpapi/pago.go`; `frontend/src/api.js`; `frontend/src/views/PagosAdmin.jsx` |
| RF-25 Rechazar ingreso si no hay mensualidad vigente | HU-11 | `diagramas/16_secuencia_gestion_pagos.puml`; CP-PG-07 | `backend/internal/application/service.go`; `backend/internal/adapters/postgres/repo.go`; transacción de pago en `backend/internal/adapters/postgres/pago.go` |
| RF-26 Cancelar pago, revocar periodo y conservar historial | HU-12 | `diagramas/14_casos_uso_pagos.puml`; `diagramas/16_secuencia_gestion_pagos.puml`; CP-PG-08 y CP-PG-09 | `backend/internal/adapters/postgres/pago.go`; `backend/internal/application/pago.go`; autorización en `backend/internal/adapters/httpapi/pago.go` |
| RF-27 Registrar cliente y vehículo | HU-13 | `pantallas/17_registro_cliente_vehiculo.svg`; CP-PG-10 | `backend/internal/application/service.go`; `backend/internal/adapters/httpapi/gestion.go`; `frontend/src/views/ClientesVehiculos.jsx` |
| RF-28 Crear cuentas de operadores | HU-14 | Casos de uso previos en `diagramas/parqueadero_iteracion_3.mdj` | `backend/internal/application/gestion.go`; `backend/internal/adapters/httpapi/gestion.go`; `frontend/src/views/UsuariosAdmin.jsx` |
| RNF-14 Límite de 1,8 s en rutas de pagos | HU-11 / HU-12 | CP-PG-11; métrica pendiente de carga en producción | `backend/internal/adapters/httpapi/pago.go`; respuesta `504` en `backend/internal/adapters/httpapi/handler.go` |
| RNF-15 Transacciones e integridad | HU-11 / HU-12 | CP-PG-12; prueba de integración PostgreSQL | `backend/internal/adapters/postgres/pago.go`; restricciones e índices en `backend/db/schema.sql` |
| RNF-16 Facturación electrónica DIAN | Fuera del alcance aceptado | Se declara pendiente; el PDF generado es recibo interno y no equivale a factura electrónica | Requiere proveedor DIAN, certificado y credenciales; no se marca como cumplido |
| RNF-17 Sesión y permisos | HU-11 / HU-12 | CP-PG-09 | `backend/internal/adapters/httpapi/handler.go`; `backend/internal/adapters/httpapi/pago.go` |

## Evidencia ejecutada

- Backend: `cd backend; go test ./...`.
- Frontend: `cd frontend; npm test`, `npm run lint` y `npm run build`.
- PostgreSQL real: `cd backend; go test ./internal/adapters/postgres -count=1` con `TEST_DATABASE_URL` apuntando a una base desechable.
- Los casos CP-PG-01 a CP-PG-10 se ejecutan como pruebas unitarias, HTTP o integración según se indica en `plan_pruebas_pagos.md`. RNF-14 requiere medir carga en el entorno de despliegue; los tests funcionales no constituyen una certificación de rendimiento.
