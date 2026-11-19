package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"autoscolombia/parqueadero/internal/domain"
)

const columnasPago = `p.id_pago, p.id_cliente, c.nombre, c.documento, p.monto, p.fecha,
	p.periodo_inicio, p.periodo_fin, p.medio_pago, p.estado, COALESCE(p.sello_digital, '')`

func (r *Repo) CrearPago(ctx context.Context, p *domain.Pago) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var clienteID int
	if err := tx.QueryRowContext(ctx, `SELECT id_cliente FROM cliente WHERE id_cliente=$1 FOR UPDATE`, p.ClienteID).Scan(&clienteID); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrClienteNoExiste
	} else if err != nil {
		return err
	}
	var duplicado bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM pago WHERE id_cliente=$1 AND periodo_inicio=$2 AND estado='PAGADO'
	)`, p.ClienteID, p.PeriodoInicio).Scan(&duplicado); err != nil {
		return err
	}
	if duplicado {
		return domain.ErrPagoDuplicado
	}
	var solapada bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM mensualidad WHERE id_cliente=$1 AND fecha_inicio <= $3 AND fecha_fin >= $2
	)`, p.ClienteID, p.PeriodoInicio, p.PeriodoFin).Scan(&solapada); err != nil {
		return err
	}
	if solapada {
		return domain.ErrMensualidadSolapada
	}
	var mensualidadID int
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES ($1,$2,$3) RETURNING id_mensualidad`,
		p.ClienteID, p.PeriodoInicio, p.PeriodoFin).Scan(&mensualidadID); err != nil {
		return err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO pago
		(id_cliente, id_mensualidad, monto, fecha, periodo_inicio, periodo_fin, medio_pago, estado)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id_pago`,
		p.ClienteID, mensualidadID, p.Monto, p.Fecha, p.PeriodoInicio, p.PeriodoFin,
		string(p.Metodo), string(p.Estado)).Scan(&p.ID)
	if esUnico(err) {
		return domain.ErrPagoDuplicado
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repo) HistorialPagos(ctx context.Context, clienteID int) ([]domain.Pago, error) {
	var existe bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM cliente WHERE id_cliente=$1)`, clienteID).Scan(&existe); err != nil {
		return nil, err
	}
	if !existe {
		return nil, domain.ErrClienteNoExiste
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+columnasPago+` FROM pago p
		JOIN cliente c ON c.id_cliente=p.id_cliente
		WHERE p.id_cliente=$1 ORDER BY p.periodo_inicio DESC, p.id_pago DESC`, clienteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pagos := []domain.Pago{}
	for rows.Next() {
		var p domain.Pago
		if err := rows.Scan(&p.ID, &p.ClienteID, &p.ClienteNombre, &p.Documento, &p.Monto, &p.Fecha,
			&p.PeriodoInicio, &p.PeriodoFin, &p.Metodo, &p.Estado, &p.SelloDigital); err != nil {
			return nil, err
		}
		pagos = append(pagos, p)
	}
	return pagos, rows.Err()
}

func (r *Repo) ObtenerPago(ctx context.Context, id int) (*domain.Pago, error) {
	var p domain.Pago
	err := r.db.QueryRowContext(ctx, `SELECT `+columnasPago+` FROM pago p
		JOIN cliente c ON c.id_cliente=p.id_cliente WHERE p.id_pago=$1`, id).Scan(
		&p.ID, &p.ClienteID, &p.ClienteNombre, &p.Documento, &p.Monto, &p.Fecha,
		&p.PeriodoInicio, &p.PeriodoFin, &p.Metodo, &p.Estado, &p.SelloDigital)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPagoNoExiste
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) GuardarSelloRecibo(ctx context.Context, id int, estado domain.EstadoPago, sello string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE pago SET sello_digital=$1 WHERE id_pago=$2 AND estado=$3`, sello, id, string(estado))
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var existe bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pago WHERE id_pago=$1)`, id).Scan(&existe); err != nil {
			return err
		}
		if existe {
			return domain.ErrPagoCambio
		}
		return domain.ErrPagoNoExiste
	}
	return nil
}

func (r *Repo) CancelarPago(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var estado domain.EstadoPago
	var mensualidadID sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT estado, id_mensualidad FROM pago WHERE id_pago=$1 FOR UPDATE`, id).
		Scan(&estado, &mensualidadID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPagoNoExiste
	}
	if err != nil {
		return err
	}
	if estado == domain.PagoCancelado {
		return domain.ErrPagoYaCancelado
	}
	if mensualidadID.Valid {
		if _, err := tx.ExecContext(ctx, `DELETE FROM mensualidad WHERE id_mensualidad=$1`, mensualidadID.Int64); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pago SET estado=$1, id_mensualidad=NULL WHERE id_pago=$2`, string(domain.PagoCancelado), id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar cancelación del pago: %w", err)
	}
	return nil
}
