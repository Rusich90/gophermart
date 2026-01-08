package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepo interface {
	GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]Order, error)
	GetSumByUserID(ctx context.Context, userID *uuid.UUID) (float64, error)
	GetByNumber(ctx context.Context, number string) (Order, error)
	Create(ctx context.Context, number string, userID *uuid.UUID) error
}
