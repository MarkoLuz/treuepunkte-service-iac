package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"treuepunkte/internal/domain"
)

type Store interface {
    RedeemPoints(ctx context.Context, customerID, reference string, points int, idemKey string) (bool, error)
}

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func isMySQLDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (r *Repository) AccruePoints(ctx context.Context, customerID, orderID string, points int, idemKey string) (bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT IGNORE INTO customers (customer_id)
		VALUES (?)
	`, customerID)
	if err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT IGNORE INTO balances (customer_id, active_points, pending_points)
		VALUES (?, 0, 0)
	`, customerID)
	if err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO points_ledger (
			customer_id,
			order_id,
			kind,
			status,
			points,
			idempotency_key
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		customerID,
		orderID,
		"accrue",
		"pending",
		points,
		nullIfEmpty(idemKey),
	)
	if err != nil {
		if isMySQLDuplicate(err) {
			return false, domain.ErrConflict
		}
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE balances
		SET pending_points = pending_points + ?
		WHERE customer_id = ?
	`, points, customerID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) ConfirmAccrue(ctx context.Context, customerID, orderID, idemKey string) (bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	var points int
	var status string

	err = tx.QueryRowContext(ctx, `
		SELECT points, status
		FROM points_ledger
		WHERE customer_id = ?
		  AND order_id = ?
		  AND kind = 'accrue'
	`, customerID, orderID).Scan(&points, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.ErrAccrueNotFound
		}
		return false, err
	}

	if status != "pending" {
		return false, domain.ErrTransactionNotPending
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO points_ledger (
			customer_id,
			order_id,
			kind,
			status,
			points,
			activated_at,
			idempotency_key
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`,
		customerID,
		orderID,
		"confirm",
		"active",
		points,
		nullIfEmpty(idemKey),
	)
	if err != nil {
		if isMySQLDuplicate(err) {
			return false, domain.ErrConflict
		}
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE balances
		SET pending_points = pending_points - ?,
		    active_points = active_points + ?
		WHERE customer_id = ?
	`, points, points, customerID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) RevokePoints(ctx context.Context, customerID, orderID, returnID, idemKey string) (bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	var points int

	err = tx.QueryRowContext(ctx, `
		SELECT points
		FROM points_ledger
		WHERE customer_id = ?
		  AND order_id = ?
		  AND kind = 'accrue'
	`, customerID, orderID).Scan(&points)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.ErrAccrueNotFound
		}
		return false, err
	}

	var confirmExists int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM points_ledger
		WHERE customer_id = ?
		  AND order_id = ?
		  AND kind = 'confirm'
	`, customerID, orderID).Scan(&confirmExists)
	if err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO points_ledger (
			customer_id,
			order_id,
			return_id,
			kind,
			status,
			points,
			idempotency_key
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		customerID,
		orderID,
		returnID,
		"revoke",
		"active",
		-points,
		nullIfEmpty(idemKey),
	)
	if err != nil {
		if isMySQLDuplicate(err) {
			return false, domain.ErrConflict
		}
		return false, err
	}

	if confirmExists > 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE balances
			SET active_points = active_points - ?
			WHERE customer_id = ?
		`, points, customerID)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE balances
			SET pending_points = pending_points - ?
			WHERE customer_id = ?
		`, points, customerID)
	}
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) RedeemPoints(ctx context.Context, customerID, reference string, points int, idemKey string) (bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	var existingCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM points_ledger
		WHERE customer_id = ?
		  AND reference = ?
		  AND kind = 'redeem'
	`, customerID, reference).Scan(&existingCount)
	if err != nil {
		return false, err
	}

	if existingCount > 0 {
		return false, domain.ErrConflict
	}

	var activePoints int
	err = tx.QueryRowContext(ctx, `
		SELECT active_points
		FROM balances
		WHERE customer_id = ?
	`, customerID).Scan(&activePoints)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.ErrNotFound
		}
		return false, err
	}

	if activePoints < points {
		return false, domain.ErrInsufficientActivePoints
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO points_ledger (
			customer_id,
			reference,
			kind,
			status,
			points,
			idempotency_key
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		customerID,
		reference,
		"redeem",
		"active",
		-points,
		nullIfEmpty(idemKey),
	)
	if err != nil {
		if isMySQLDuplicate(err) {
			return false, domain.ErrConflict
		}
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE balances
		SET active_points = active_points - ?
		WHERE customer_id = ?
	`, points, customerID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) RestorePoints(ctx context.Context, customerID, reference, idemKey string) (bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	var existingCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM points_ledger
		WHERE customer_id = ?
		  AND reference = ?
		  AND kind = 'restore'
	`, customerID, reference).Scan(&existingCount)
	if err != nil {
		return false, err
	}

	if existingCount > 0 {
		return false, domain.ErrConflict
	}

	var redeemPoints int
	err = tx.QueryRowContext(ctx, `
		SELECT points
		FROM points_ledger
		WHERE customer_id = ?
		  AND reference = ?
		  AND kind = 'redeem'
	`, customerID, reference).Scan(&redeemPoints)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.ErrRedeemNotFound
		}
		return false, err
	}

	if redeemPoints >= 0 {
		return false, domain.ErrConflict
	}

	restorePoints := -redeemPoints

	_, err = tx.ExecContext(ctx, `
		INSERT INTO points_ledger (
			customer_id,
			reference,
			kind,
			status,
			points,
			idempotency_key
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		customerID,
		reference,
		"restore",
		"active",
		restorePoints,
		nullIfEmpty(idemKey),
	)
	if err != nil {
		if isMySQLDuplicate(err) {
			return false, domain.ErrConflict
		}
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE balances
		SET active_points = active_points + ?
		WHERE customer_id = ?
	`, restorePoints, customerID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) GetBalance(ctx context.Context, customerID string) (domain.Balance, error) {
	var balance domain.Balance

	err := r.DB.QueryRowContext(ctx, `
		SELECT customer_id, active_points, pending_points
		FROM balances
		WHERE customer_id = ?
	`, customerID).Scan(
		&balance.CustomerID,
		&balance.ActivePoints,
		&balance.PendingPoints,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Balance{}, domain.ErrNotFound
		}
		return domain.Balance{}, err
	}

	return balance, nil
}

func (r *Repository) GetTransactions(ctx context.Context, customerID string) ([]domain.Transaction, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, customer_id, order_id, reference, return_id, kind, status, points, occurred_at
		FROM points_ledger
		WHERE customer_id = ?
		ORDER BY id ASC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction

	for rows.Next() {
		var tx domain.Transaction

		err := rows.Scan(
			&tx.ID,
			&tx.CustomerID,
			&tx.OrderID,
			&tx.Reference,
			&tx.ReturnID,
			&tx.Kind,
			&tx.Status,
			&tx.Points,
			&tx.OccurredAt,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
