-- Datos adicionales de prueba para vehiculos y clientes
INSERT INTO cliente (documento, nombre, telefono) VALUES
 ('1004004004', 'Sofia Castro', '3004445566'),
 ('1005005005', 'Andres Mejia', '3005556677'),
 ('1006006006', 'Valentina Rios', '3006667788')
ON CONFLICT (documento) DO NOTHING;

-- Mensualidades (vigentes y vencidas)
INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin)
SELECT id_cliente, CURRENT_DATE - 2, CURRENT_DATE + 28 FROM cliente WHERE documento = '1004004004' AND NOT EXISTS (SELECT 1 FROM mensualidad WHERE id_cliente = cliente.id_cliente);

INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin)
SELECT id_cliente, CURRENT_DATE - 15, CURRENT_DATE + 15 FROM cliente WHERE documento = '1005005005' AND NOT EXISTS (SELECT 1 FROM mensualidad WHERE id_cliente = cliente.id_cliente);

INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin)
SELECT id_cliente, CURRENT_DATE - 45, CURRENT_DATE - 15 FROM cliente WHERE documento = '1006006006' AND NOT EXISTS (SELECT 1 FROM mensualidad WHERE id_cliente = cliente.id_cliente);

-- Vehiculos de prueba (Motos y Carros)
-- 1. Moto ABC12D (la del ejemplo en el placeholder, con mensualidad vigente)
INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'ABC12D', 'MOTO', id_cliente FROM cliente WHERE documento = '1001001001'
ON CONFLICT (placa) DO NOTHING;

-- 2. Otras motos con mensualidad vigente
INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'KTM45E', 'MOTO', id_cliente FROM cliente WHERE documento = '1004004004'
ON CONFLICT (placa) DO NOTHING;

INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'MOT01A', 'MOTO', id_cliente FROM cliente WHERE documento = '1005005005'
ON CONFLICT (placa) DO NOTHING;

-- 3. Moto con mensualidad vencida (para probar rechazo por mensualidad)
INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'YAM88B', 'MOTO', id_cliente FROM cliente WHERE documento = '1006006006'
ON CONFLICT (placa) DO NOTHING;

-- 4. Carros con mensualidad vigente
INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'BMW321', 'CARRO', id_cliente FROM cliente WHERE documento = '1005005005'
ON CONFLICT (placa) DO NOTHING;

INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'TOY789', 'CARRO', id_cliente FROM cliente WHERE documento = '1004004004'
ON CONFLICT (placa) DO NOTHING;

-- 5. Carro con mensualidad vencida (para probar rechazo por mensualidad)
INSERT INTO vehiculo (placa, tipo, id_cliente)
SELECT 'REN101', 'CARRO', id_cliente FROM cliente WHERE documento = '1006006006'
ON CONFLICT (placa) DO NOTHING;

-- Celdas adicionales
INSERT INTO celda (codigo, tipo) VALUES
 ('M-03', 'MOTO'),
 ('A-04', 'CARRO')
ON CONFLICT (codigo) DO NOTHING;
