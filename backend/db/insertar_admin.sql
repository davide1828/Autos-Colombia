-- Insertar administrador inicial (solo si no existe)
INSERT INTO usuario (usuario, nombre, rol, clave_hash, activo)
VALUES ('admin', 'Administrador Principal', 'ADMINISTRADOR', 'pbkdf2-sha256$600000$mCjgvKXOjqefKAyW01HVRA$Ceo4NSzlTaqIpqO5dnD5p/GpX2TeDAco1FGlqociWM8', true)
ON CONFLICT (usuario) DO NOTHING;
