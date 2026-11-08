package postgres

// Iteración 2: persistencia de usuarios, clientes, celdas y auditoría.

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"

	"autoscolombia/parqueadero/internal/domain"
)

const codigoUnico = "23505" // unique_violation

func esUnico(err error) bool {
	var pe *pq.Error
	return errors.As(err, &pe) && pe.Code == codigoUnico
}

// ---------- usuarios ----------

const colsUsuario = `id_usuario, usuario, nombre, rol, clave_hash, activo`

func escanearUsuario(f func(...any) error) (*domain.Usuario, error) {
	var u domain.Usuario
	if err := f(&u.ID, &u.Usuario, &u.Nombre, &u.Rol, &u.ClaveHash, &u.Activo); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) UsuarioPorLogin(ctx context.Context, login string) (*domain.Usuario, error) {
	u, err := escanearUsuario(r.db.QueryRowContext(ctx, `SELECT `+colsUsuario+` FROM usuario WHERE usuario=$1`, login).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUsuarioNoExiste
	}
	return u, err
}

func (r *Repo) UsuarioPorID(ctx context.Context, id int) (*domain.Usuario, error) {
	u, err := escanearUsuario(r.db.QueryRowContext(ctx, `SELECT `+colsUsuario+` FROM usuario WHERE id_usuario=$1`, id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUsuarioNoExiste
	}
	return u, err
}

func (r *Repo) Usuarios(ctx context.Context) ([]domain.Usuario, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+colsUsuario+` FROM usuario ORDER BY usuario`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Usuario{}
	for rows.Next() {
		u, err := escanearUsuario(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *Repo) ContarUsuarios(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM usuario`).Scan(&n)
	return n, err
}

func (r *Repo) CrearUsuario(ctx context.Context, u *domain.Usuario) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO usuario (usuario, nombre, rol, clave_hash, activo) VALUES ($1,$2,$3,$4,$5) RETURNING id_usuario`,
		u.Usuario, u.Nombre, string(u.Rol), u.ClaveHash, u.Activo).Scan(&u.ID)
	if esUnico(err) {
		return domain.ErrUsuarioDuplicado
	}
	return err
}

// ActualizarUsuario serializa los cambios sobre los administradores activos con bloqueo de filas, de modo que
// dos administradores que se desactiven a la vez no puedan dejar el sistema sin ninguno.
func (r *Repo) ActualizarUsuario(ctx context.Context, u *domain.Usuario) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id_usuario FROM usuario WHERE rol='ADMINISTRADOR' AND activo FOR UPDATE`)
	if err != nil {
		return err
	}
	rows.Close()
	res, err := tx.ExecContext(ctx,
		`UPDATE usuario SET nombre=$2, rol=$3, activo=$4, clave_hash=$5 WHERE id_usuario=$1`,
		u.ID, u.Nombre, string(u.Rol), u.Activo, u.ClaveHash)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrUsuarioNoExiste
	}
	var admins int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM usuario WHERE rol='ADMINISTRADOR' AND activo`).Scan(&admins); err != nil {
		return err
	}
	if admins == 0 {
		return domain.ErrUltimoAdministrador
	}
	return tx.Commit()
}

// ---------- clientes y vehículos ----------

var escapeLike = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// cargarDetalle completa vehículos y mensualidades de los clientes dados.
func (r *Repo) cargarDetalle(ctx context.Context, base []domain.ClienteDetalle) error {
	if len(base) == 0 {
		return nil
	}
	ids := make([]int64, len(base))
	pos := map[int]int{}
	for i := range base {
		ids[i] = int64(base[i].ID)
		pos[base[i].ID] = i
		base[i].Vehiculos, base[i].Mensualidades = []domain.Vehiculo{}, []domain.PeriodoMensualidad{}
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id_vehiculo, placa, tipo, id_cliente FROM vehiculo WHERE id_cliente = ANY($1) ORDER BY placa`, pq.Array(ids))
	if err != nil {
		return err
	}
	for rows.Next() {
		var v domain.Vehiculo
		if err := rows.Scan(&v.ID, &v.Placa, &v.Tipo, &v.ClienteID); err != nil {
			rows.Close()
			return err
		}
		base[pos[v.ClienteID]].Vehiculos = append(base[pos[v.ClienteID]].Vehiculos, v)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx,
		`SELECT id_mensualidad, id_cliente, to_char(fecha_inicio,'YYYY-MM-DD'), to_char(fecha_fin,'YYYY-MM-DD')
		   FROM mensualidad WHERE id_cliente = ANY($1) ORDER BY fecha_inicio DESC`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p domain.PeriodoMensualidad
		var cli int
		if err := rows.Scan(&p.ID, &cli, &p.FechaInicio, &p.FechaFin); err != nil {
			return err
		}
		base[pos[cli]].Mensualidades = append(base[pos[cli]].Mensualidades, p)
	}
	return rows.Err()
}

func (r *Repo) BuscarClientes(ctx context.Context, q string, limite int) ([]domain.ClienteDetalle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_cliente, documento, nombre, COALESCE(telefono,'') FROM cliente
		  WHERE documento ILIKE $1 ESCAPE '\' OR nombre ILIKE $1 ESCAPE '\' ORDER BY nombre, id_cliente LIMIT $2`,
		"%"+escapeLike.Replace(q)+"%", limite)
	if err != nil {
		return nil, err
	}
	out := []domain.ClienteDetalle{}
	for rows.Next() {
		var c domain.ClienteDetalle
		if err := rows.Scan(&c.ID, &c.Documento, &c.Nombre, &c.Telefono); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, c)
	}
	rows.Close()
	return out, r.cargarDetalle(ctx, out)
}

func (r *Repo) ClientePorID(ctx context.Context, id int) (*domain.ClienteDetalle, error) {
	var c domain.ClienteDetalle
	err := r.db.QueryRowContext(ctx, `SELECT id_cliente, documento, nombre, COALESCE(telefono,'') FROM cliente WHERE id_cliente=$1`, id).
		Scan(&c.ID, &c.Documento, &c.Nombre, &c.Telefono)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrClienteNoExiste
	}
	if err != nil {
		return nil, err
	}
	lista := []domain.ClienteDetalle{c}
	if err := r.cargarDetalle(ctx, lista); err != nil {
		return nil, err
	}
	return &lista[0], nil
}

func (r *Repo) ActualizarCliente(ctx context.Context, c *domain.Cliente) error {
	res, err := r.db.ExecContext(ctx, `UPDATE cliente SET nombre=$2, telefono=NULLIF($3,'') WHERE id_cliente=$1`, c.ID, c.Nombre, c.Telefono)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrClienteNoExiste
	}
	return nil
}

// AgregarMensualidad bloquea la fila del cliente para que dos registros simultáneos no puedan solaparse.
func (r *Repo) AgregarMensualidad(ctx context.Context, m *domain.Mensualidad) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var uno int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM cliente WHERE id_cliente=$1 FOR UPDATE`, m.ClienteID).Scan(&uno); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrClienteNoExiste
	} else if err != nil {
		return err
	}
	ini, fin := m.Inicio.Format("2006-01-02"), m.Fin.Format("2006-01-02")
	var solapa bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM mensualidad WHERE id_cliente=$1 AND fecha_inicio <= $3::date AND fecha_fin >= $2::date)`,
		m.ClienteID, ini, fin).Scan(&solapa); err != nil {
		return err
	}
	if solapa {
		return domain.ErrMensualidadSolapada
	}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES ($1,$2,$3) RETURNING id_mensualidad`,
		m.ClienteID, ini, fin).Scan(&m.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repo) VehiculoPorID(ctx context.Context, id int) (*domain.Vehiculo, error) {
	var v domain.Vehiculo
	err := r.db.QueryRowContext(ctx, `SELECT id_vehiculo, placa, tipo, id_cliente FROM vehiculo WHERE id_vehiculo=$1`, id).
		Scan(&v.ID, &v.Placa, &v.Tipo, &v.ClienteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrVehiculoNoRegistrado
	}
	return &v, err
}

func (r *Repo) ActualizarTipoVehiculo(ctx context.Context, id int, tipo domain.TipoVehiculo) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var uno int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM vehiculo WHERE id_vehiculo=$1 FOR UPDATE`, id).Scan(&uno); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrVehiculoNoRegistrado
	} else if err != nil {
		return err
	}
	var dentro bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM registro WHERE id_vehiculo=$1 AND fecha_salida IS NULL)`, id).Scan(&dentro); err != nil {
		return err
	}
	if dentro {
		return domain.ErrVehiculoDentro
	}
	if _, err := tx.ExecContext(ctx, `UPDATE vehiculo SET tipo=$2 WHERE id_vehiculo=$1`, id, string(tipo)); err != nil {
		return err
	}
	return tx.Commit()
}

// ---------- celdas ----------

func (r *Repo) CrearCelda(ctx context.Context, c *domain.Celda) error {
	err := r.db.QueryRowContext(ctx, `INSERT INTO celda (codigo, tipo, estado) VALUES ($1,$2,'DISPONIBLE') RETURNING id_celda`,
		c.Codigo, string(c.Tipo)).Scan(&c.ID)
	if esUnico(err) {
		return domain.ErrCeldaDuplicada
	}
	return err
}

// ActualizarCelda bloquea la fila: si en ese instante se registra un ingreso, uno de los dos espera al otro.
func (r *Repo) ActualizarCelda(ctx context.Context, c *domain.Celda) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var estado string
	if err := tx.QueryRowContext(ctx, `SELECT estado FROM celda WHERE id_celda=$1 FOR UPDATE`, c.ID).Scan(&estado); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrCeldaNoExiste
	} else if err != nil {
		return err
	}
	if domain.EstadoCelda(estado) == domain.CeldaOcupada {
		return domain.ErrCeldaOcupada
	}
	if _, err := tx.ExecContext(ctx, `UPDATE celda SET codigo=$2, tipo=$3 WHERE id_celda=$1`, c.ID, c.Codigo, string(c.Tipo)); err != nil {
		if esUnico(err) {
			return domain.ErrCeldaDuplicada
		}
		return err
	}
	return tx.Commit()
}

func (r *Repo) CambiarEstadoCelda(ctx context.Context, id int, estado domain.EstadoCelda) error {
	res, err := r.db.ExecContext(ctx, `UPDATE celda SET estado=$2 WHERE id_celda=$1 AND estado<>'OCUPADA'`, id, string(estado))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 1 {
		return nil
	}
	var existe bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM celda WHERE id_celda=$1)`, id).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return domain.ErrCeldaNoExiste
	}
	return domain.ErrCeldaOcupada
}

// ---------- auditoría ----------

func (r *Repo) RegistrarAuditoria(ctx context.Context, a *domain.Auditoria) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO auditoria (momento, id_usuario, usuario, accion, entidad, id_entidad, detalle)
		 VALUES ($1, NULLIF($2,0), $3, $4, $5, NULLIF($6,0), NULLIF($7,'')) RETURNING id_auditoria`,
		a.Momento, a.UsuarioID, a.Usuario, a.Accion, a.Entidad, a.EntidadID, a.Detalle).Scan(&a.ID)
}

func (r *Repo) Auditorias(ctx context.Context, limite int) ([]domain.Auditoria, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_auditoria, momento, COALESCE(id_usuario,0), usuario, accion, entidad, COALESCE(id_entidad,0), COALESCE(detalle,'')
		   FROM auditoria ORDER BY id_auditoria DESC LIMIT $1`, limite)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Auditoria{}
	for rows.Next() {
		var a domain.Auditoria
		if err := rows.Scan(&a.ID, &a.Momento, &a.UsuarioID, &a.Usuario, &a.Accion, &a.Entidad, &a.EntidadID, &a.Detalle); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// EsquemaActualizado verifica que la base tenga las tablas y reglas de la Iteración 2. Las bases creadas con la
// Iteración 1 deben actualizarse con db/schema.sql (es repetible: solo agrega lo que falta).
func (r *Repo) EsquemaActualizado(ctx context.Context) error {
	var tablas, triggers int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM unnest(ARRAY['usuario','auditoria']) t WHERE to_regclass('public.'||t) IS NOT NULL`).Scan(&tablas); err != nil {
		return err
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM pg_trigger WHERE tgname IN ('trg_auditoria_inmutable','trg_registro_tipos') AND NOT tgisinternal`).Scan(&triggers); err != nil {
		return err
	}
	if tablas != 2 || triggers != 2 {
		return errors.New("el esquema de la base de datos está desactualizado: aplique backend/db/schema.sql (es repetible)")
	}
	return nil
}
