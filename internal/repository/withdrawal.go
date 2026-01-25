package repository

import (
	"context"
	"errors"
	"fmt"

	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawalRepo struct {
	db *pgxpool.Pool
}

func NewWithdrawalRepo(db *pgxpool.Pool) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

func (r *WithdrawalRepo) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainwithdrawal.Withdrawal, error) {
	var withdrawals []domainwithdrawal.Withdrawal

	query := `
		SELECT order_num, user_id, sum, created_at
		FROM withdrawals 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals by user id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal domainwithdrawal.Withdrawal
		err := rows.Scan(
			&withdrawal.OrderNum,
			&withdrawal.UserID,
			&withdrawal.Sum,
			&withdrawal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (r *WithdrawalRepo) Create(ctx context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (order_num, user_id, sum, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		withdrawal.OrderNum,
		withdrawal.UserID,
		withdrawal.Sum,
		withdrawal.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domainwithdrawal.ErrOrderNumConflict
		}
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

func (r *WithdrawalRepo) GetSumByUserID(ctx context.Context, userID *uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(sum), 0)
		FROM withdrawals 
		WHERE user_id = $1
	`

	var totalSum float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&totalSum)
	if err != nil {
		return 0, fmt.Errorf("failed to get sum of withdrawals by user id: %w", err)
	}

	return totalSum, nil
}
