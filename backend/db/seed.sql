-- Datos iniciales de ejemplo (clientes, vehículos y celdas ya registrados)
INSERT INTO cliente (documento, nombre, telefono) VALUES
 ('1001001001','Carlos Gómez','3001112233'),
 ('1002002002','Marta Rojas','3002223344'),
 ('1003003003','Luis Pérez','3003334455'),
 ('1004004004','Sofía Castro','3004445566'),
 ('1005005005','Andrés Mejía','3005556677'),
 ('1006006006','Valentina Ríos','3006667788');

INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES
 (1, CURRENT_DATE - 10, CURRENT_DATE + 20),
 (2, CURRENT_DATE - 60, CURRENT_DATE - 30),   -- vencida
 (3, CURRENT_DATE - 5,  CURRENT_DATE + 25),
 (4, CURRENT_DATE - 2,  CURRENT_DATE + 28),
 (5, CURRENT_DATE - 15, CURRENT_DATE + 15),
 (6, CURRENT_DATE - 45, CURRENT_DATE - 15);   -- vencida

INSERT INTO vehiculo (placa, tipo, id_cliente) VALUES
 ('ABC123','CARRO',1),
 ('ABC12D','MOTO',1),
 ('XYZ98A','MOTO',1),
 ('DEF456','CARRO',2),
 ('GHI789','CARRO',3),
 ('TOY789','CARRO',4),
 ('KTM45E','MOTO',4),
 ('BMW321','CARRO',5),
 ('MOT01A','MOTO',5),
 ('REN101','CARRO',6),
 ('YAM88B','MOTO',6);

INSERT INTO celda (codigo, tipo) VALUES
 ('A-01','CARRO'),
 ('A-02','CARRO'),
 ('A-03','CARRO'),
 ('A-04','CARRO'),
 ('M-01','MOTO'),
 ('M-02','MOTO'),
 ('M-03','MOTO');

UPDATE celda SET estado='INACTIVA' WHERE codigo='A-03';
