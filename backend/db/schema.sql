-- Modelo físico (PostgreSQL) - Parqueadero Autos Colombia - Iteración 1
CREATE TABLE IF NOT EXISTS cliente (
    id_cliente  SERIAL PRIMARY KEY,
    documento   VARCHAR(20)  NOT NULL UNIQUE,
    nombre      VARCHAR(100) NOT NULL,
    telefono    VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS mensualidad (
    id_mensualidad SERIAL PRIMARY KEY,
    id_cliente     INT  NOT NULL REFERENCES cliente(id_cliente),
    fecha_inicio   DATE NOT NULL,
    fecha_fin      DATE NOT NULL,
    CONSTRAINT ck_mensualidad_fechas CHECK (fecha_fin >= fecha_inicio)
);

CREATE TABLE IF NOT EXISTS vehiculo (
    id_vehiculo SERIAL PRIMARY KEY,
    placa       VARCHAR(6)  NOT NULL UNIQUE,
    tipo        VARCHAR(5)  NOT NULL CHECK (tipo IN ('CARRO','MOTO')),
    id_cliente  INT         NOT NULL REFERENCES cliente(id_cliente)
);

CREATE TABLE IF NOT EXISTS celda (
    id_celda SERIAL PRIMARY KEY,
    codigo   VARCHAR(10) NOT NULL UNIQUE,
    tipo     VARCHAR(5)  NOT NULL CHECK (tipo IN ('CARRO','MOTO')),
    estado   VARCHAR(10) NOT NULL DEFAULT 'DISPONIBLE'
             CHECK (estado IN ('DISPONIBLE','OCUPADA','INACTIVA'))
);

CREATE TABLE IF NOT EXISTS registro (
    id_registro   SERIAL PRIMARY KEY,
    id_vehiculo   INT         NOT NULL REFERENCES vehiculo(id_vehiculo),
    id_celda      INT         NOT NULL REFERENCES celda(id_celda),
    fecha_entrada TIMESTAMPTZ NOT NULL,
    fecha_salida  TIMESTAMPTZ,
    CONSTRAINT ck_registro_fechas CHECK (fecha_salida IS NULL OR fecha_salida >= fecha_entrada)
);

-- Reglas 2 y 3: un solo registro activo por vehículo y por celda.
CREATE UNIQUE INDEX IF NOT EXISTS uq_registro_vehiculo_activo ON registro(id_vehiculo) WHERE fecha_salida IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_registro_celda_activa    ON registro(id_celda)    WHERE fecha_salida IS NULL;

CREATE TABLE IF NOT EXISTS novedad (
    id_novedad  SERIAL PRIMARY KEY,
    id_registro INT         NOT NULL REFERENCES registro(id_registro),
    id_vehiculo INT         NOT NULL REFERENCES vehiculo(id_vehiculo),
    tipo        VARCHAR(10) NOT NULL CHECK (tipo IN ('DANIO','FALTANTE','OTRA')),
    descripcion VARCHAR(500) NOT NULL,
    momento     VARCHAR(20) NOT NULL CHECK (momento IN ('ENTRADA','SALIDA','DURANTE_ESTANCIA')),
    fecha       TIMESTAMPTZ NOT NULL
);

-- Índices según las consultas del sistema
CREATE INDEX IF NOT EXISTS ix_mensualidad_cliente_fechas ON mensualidad(id_cliente, fecha_inicio, fecha_fin);
CREATE INDEX IF NOT EXISTS ix_celda_estado               ON celda(estado);
CREATE INDEX IF NOT EXISTS ix_novedad_registro           ON novedad(id_registro);
