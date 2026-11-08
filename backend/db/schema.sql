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

-- =====================================================================
-- Iteración 2: gestión de usuarios del sistema, roles y auditoría
-- =====================================================================
CREATE TABLE IF NOT EXISTS usuario (
    id_usuario  SERIAL PRIMARY KEY,
    usuario     VARCHAR(30)  NOT NULL UNIQUE,
    nombre      VARCHAR(100) NOT NULL,
    rol         VARCHAR(15)  NOT NULL,
    clave_hash  TEXT         NOT NULL,
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    creado      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT ck_usuario_login CHECK (usuario ~ '^[a-z0-9._-]{3,30}$'),
    CONSTRAINT ck_usuario_rol   CHECK (rol IN ('ADMINISTRADOR', 'ENCARGADO'))
);

-- Quién hizo qué y cuándo. id_usuario es NULL solo en el modo de desarrollo sin autenticación.
CREATE TABLE IF NOT EXISTS auditoria (
    id_auditoria BIGSERIAL PRIMARY KEY,
    momento      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    id_usuario   INT REFERENCES usuario(id_usuario),
    usuario      VARCHAR(30)  NOT NULL,
    accion       VARCHAR(40)  NOT NULL,
    entidad      VARCHAR(20)  NOT NULL,
    id_entidad   INT,
    detalle      VARCHAR(300)
);
CREATE INDEX IF NOT EXISTS ix_auditoria_momento ON auditoria (momento DESC);

-- RN-06: la auditoría solo se agrega; nunca se modifica ni se elimina.
CREATE OR REPLACE FUNCTION auditoria_inmutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'la auditoría no se modifica ni se elimina';
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_auditoria_inmutable ON auditoria;
CREATE TRIGGER trg_auditoria_inmutable BEFORE UPDATE OR DELETE ON auditoria
    FOR EACH ROW EXECUTE FUNCTION auditoria_inmutable();

-- RN-03 en la base de datos: un registro solo puede unir un vehículo con una celda de su mismo tipo.
-- Bloquea la fila del vehículo (FOR SHARE) para que un cambio de tipo simultáneo espere a este ingreso.
CREATE OR REPLACE FUNCTION registro_tipos_compatibles() RETURNS trigger AS $$
DECLARE
    tipo_vehiculo VARCHAR(5);
    tipo_celda    VARCHAR(5);
BEGIN
    SELECT tipo INTO tipo_vehiculo FROM vehiculo WHERE id_vehiculo = NEW.id_vehiculo FOR SHARE;
    SELECT tipo INTO tipo_celda    FROM celda    WHERE id_celda    = NEW.id_celda;
    IF tipo_vehiculo IS DISTINCT FROM tipo_celda THEN
        RAISE EXCEPTION 'el tipo del vehículo y el de la celda no coinciden'
            USING ERRCODE = 'check_violation', CONSTRAINT = 'ck_registro_tipos_compatibles';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_registro_tipos ON registro;
CREATE TRIGGER trg_registro_tipos BEFORE INSERT ON registro
    FOR EACH ROW EXECUTE FUNCTION registro_tipos_compatibles();
