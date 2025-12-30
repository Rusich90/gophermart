package repository

import (
	"context"
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
