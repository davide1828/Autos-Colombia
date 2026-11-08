// Paquete postgres: adaptador de persistencia (implementa application.Repositorio).
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"autoscolombia/parqueadero/internal/domain"
)

type Repo struct{ db *sql.DB }

func Nuevo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) VehiculoPorPlaca(ctx context.Context, placa string) (*domain.Vehiculo, error) {
	var v domain.Vehiculo
	err := r.db.QueryRowContext(ctx,
		`SELECT id_vehiculo, placa, tipo, id_cliente FROM vehiculo WHERE placa=$1`, placa).
		Scan(&v.ID, &v.Placa, &v.Tipo, &v.ClienteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrVehiculoNoRegistrado
	}
	return &v, err
}

func (r *Repo) MensualidadVigente(ctx context.Context, clienteID int, dia time.Time) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM mensualidad
		   WHERE id_cliente=$1 AND $2::date BETWEEN fecha_inicio AND fecha_fin)`,
		clienteID, dia.Format("2006-01-02")).Scan(&ok) // día calendario en la zona de "dia"
	return ok, err
}

func (r *Repo) Celda(ctx context.Context, id int) (*domain.Celda, error) {
	var c domain.Celda
	err := r.db.QueryRowContext(ctx,
		`SELECT id_celda, codigo, tipo, estado FROM celda WHERE id_celda=$1`, id).
		Scan(&c.ID, &c.Codigo, &c.Tipo, &c.Estado)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCeldaNoExiste
	}
	return &c, err
}

func (r *Repo) RegistroActivoDeVehiculo(ctx context.Context, vehiculoID int) (*domain.Registro, error) {
	var g domain.Registro
	err := r.db.QueryRowContext(ctx,
		`SELECT id_registro, id_vehiculo, id_celda, fecha_entrada
		   FROM registro WHERE id_vehiculo=$1 AND fecha_salida IS NULL`, vehiculoID).
		Scan(&g.ID, &g.VehiculoID, &g.CeldaID, &g.Entrada)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &g, err
}

func insertarNovedad(ctx context.Context, tx *sql.Tx, n *domain.Novedad) error {
	return tx.QueryRowContext(ctx,
		`INSERT INTO novedad (id_registro,id_vehiculo,tipo,descripcion,momento,fecha)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id_novedad`,
		n.RegistroID, n.VehiculoID, n.Tipo, n.Descripcion, n.Momento, n.Fecha).Scan(&n.ID)
}

func (r *Repo) GuardarIngreso(ctx context.Context, g *domain.Registro, n *domain.Novedad) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ocupa la celda solo si sigue disponible (evita carreras entre dos encargados).
	res, err := tx.ExecContext(ctx,
		`UPDATE celda SET estado='OCUPADA' WHERE id_celda=$1 AND estado='DISPONIBLE'`, g.CeldaID)
	if err != nil {
		return err
	}
	if filas, _ := res.RowsAffected(); filas == 0 {
		return domain.ErrCeldaNoDisponible
	}
	err = tx.QueryRowContext(ctx,
		`INSERT INTO registro (id_vehiculo,id_celda,fecha_entrada) VALUES ($1,$2,$3) RETURNING id_registro`,
		g.VehiculoID, g.CeldaID, g.Entrada).Scan(&g.ID)
	if err != nil {
		var pe *pq.Error
		if errors.As(err, &pe) && pe.Code == "23505" {
			if pe.Constraint == "uq_registro_vehiculo_activo" {
				return domain.ErrVehiculoYaDentro
			}
			return domain.ErrCeldaNoDisponible
		}
		if errors.As(err, &pe) && pe.Code == "23514" && pe.Constraint == "ck_registro_tipos_compatibles" {
			return domain.ErrCeldaIncompatible // el tipo del vehículo cambió mientras se registraba el ingreso
		}
		return err
	}
	if n != nil {
		n.RegistroID = g.ID
		if err := insertarNovedad(ctx, tx, n); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repo) GuardarSalida(ctx context.Context, g *domain.Registro, n *domain.Novedad) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE registro SET fecha_salida=$1 WHERE id_registro=$2 AND fecha_salida IS NULL`, g.Salida, g.ID)
	if err != nil {
		return err
	}
	if filas, _ := res.RowsAffected(); filas == 0 {
		return domain.ErrSinIngresoActivo
	}
	if _, err := tx.ExecContext(ctx, `UPDATE celda SET estado='DISPONIBLE' WHERE id_celda=$1`, g.CeldaID); err != nil {
		return err
	}
	if n != nil {
		n.RegistroID = g.ID
		if err := insertarNovedad(ctx, tx, n); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repo) GuardarNovedad(ctx context.Context, n *domain.Novedad) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := insertarNovedad(ctx, tx, n); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repo) Celdas(ctx context.Context, estado string) ([]domain.Celda, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_celda, codigo, tipo, estado FROM celda WHERE ($1='' OR estado=$1) ORDER BY codigo`, estado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Celda{}
	for rows.Next() {
		var c domain.Celda
		if err := rows.Scan(&c.ID, &c.Codigo, &c.Tipo, &c.Estado); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) RegistrosActivos(ctx context.Context) ([]domain.RegistroDetalle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT g.id_registro, v.placa, c.nombre, ce.codigo, g.fecha_entrada
		   FROM registro g
		   JOIN vehiculo v ON v.id_vehiculo=g.id_vehiculo
		   JOIN cliente  c ON c.id_cliente=v.id_cliente
		   JOIN celda   ce ON ce.id_celda=g.id_celda
		  WHERE g.fecha_salida IS NULL ORDER BY g.fecha_entrada`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RegistroDetalle{}
	for rows.Next() {
		var d domain.RegistroDetalle
		if err := rows.Scan(&d.RegistroID, &d.Placa, &d.Cliente, &d.Celda, &d.Entrada); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GuardarVehiculo registra cliente (si es nuevo), vehículo y mensualidad en una sola transacción.
func (r *Repo) GuardarVehiculo(ctx context.Context, c *domain.Cliente, v *domain.Vehiculo, m *domain.Mensualidad) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	nuevo := true
	err = tx.QueryRowContext(ctx,
		`INSERT INTO cliente (documento, nombre, telefono) VALUES ($1,$2,NULLIF($3,''))
		 ON CONFLICT (documento) DO NOTHING RETURNING id_cliente`, c.Documento, c.Nombre, c.Telefono).Scan(&c.ID)
	if errors.Is(err, sql.ErrNoRows) { // el documento ya existe: se reutiliza el cliente
		nuevo = false
		err = tx.QueryRowContext(ctx, `SELECT id_cliente, nombre FROM cliente WHERE documento=$1`, c.Documento).Scan(&c.ID, &c.Nombre)
	}
	if err != nil {
		return false, err
	}
	v.ClienteID = c.ID
	err = tx.QueryRowContext(ctx,
		`INSERT INTO vehiculo (placa, tipo, id_cliente) VALUES ($1,$2,$3) RETURNING id_vehiculo`,
		v.Placa, string(v.Tipo), c.ID).Scan(&v.ID)
	var pe *pq.Error
	if errors.As(err, &pe) && pe.Code == "23505" {
		return false, domain.ErrPlacaYaRegistrada
	}
	if err != nil {
		return false, err
	}
	if m != nil {
		m.ClienteID = c.ID
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES ($1,$2,$3) RETURNING id_mensualidad`,
			c.ID, m.Inicio.Format("2006-01-02"), m.Fin.Format("2006-01-02")).Scan(&m.ID); err != nil {
			return false, err
		}
	}
	return nuevo, tx.Commit()
}
