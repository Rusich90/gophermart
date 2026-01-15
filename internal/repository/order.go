package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainorder.Order, error) {
	var orders []domainorder.Order

	query := `
		SELECT number, user_id, status, accrual, created_at
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order domainorder.Order
		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) GetByNumber(ctx context.Context, number string) (domainorder.Order, error) {
	var order domainorder.Order

	query := `
		SELECT number, user_id, status, accrual, created_at
		FROM orders 
		WHERE number = $1
	`

	row := r.db.QueryRow(ctx, query, number)

	err := row.Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return order, domainorder.ErrOrderNotFound
		}
		return order, fmt.Errorf("failed to scan order: %w", err)
	}

	return order, nil
}

func (r *OrderRepo) Create(ctx context.Context, order *domainorder.Order) error {
	query := `
		INSERT INTO orders (number, user_id, status, accrual, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.Number,
		order.UserID,
		order.Status,
		order.Accrual,
		order.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to exec insert order: %w", err)
	}

	return nil
}

func (r *OrderRepo) Update(ctx context.Context, order *domainorder.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3
	`

	_, err := r.db.Exec(ctx, query, order.Status, order.Accrual, order.Number)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

func (r *OrderRepo) GetSumByUserID(ctx context.Context, userID *uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(accrual), 0)
		FROM orders 
		WHERE user_id = $1 AND status = $2
	`

	var totalSum float64
	err := r.db.QueryRow(ctx, query, userID, domainorder.PROCESSED).Scan(&totalSum)
	if err != nil {
		return 0, fmt.Errorf("failed to get sum of processed orders by user id: %w", err)
	}

	return totalSum, nil
}
