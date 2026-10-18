-- Datos iniciales de ejemplo (clientes, vehículos y celdas ya registrados)
INSERT INTO cliente (documento, nombre, telefono) VALUES
 ('1001001001','Carlos Gómez','3001112233'),
 ('1002002002','Marta Rojas','3002223344'),
 ('1003003003','Luis Pérez','3003334455');
INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES
 (1, CURRENT_DATE - 10, CURRENT_DATE + 20),
 (2, CURRENT_DATE - 60, CURRENT_DATE - 30),   -- vencida
 (3, CURRENT_DATE - 5,  CURRENT_DATE + 25);
INSERT INTO vehiculo (placa, tipo, id_cliente) VALUES
 ('ABC123','CARRO',1), ('XYZ98A','MOTO',1), ('DEF456','CARRO',2), ('GHI789','CARRO',3);
INSERT INTO celda (codigo, tipo) VALUES
 ('A-01','CARRO'),('A-02','CARRO'),('A-03','CARRO'),('M-01','MOTO'),('M-02','MOTO');
UPDATE celda SET estado='INACTIVA' WHERE codigo='A-03';
