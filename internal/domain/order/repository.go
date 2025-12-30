package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepo interface {
	GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]Order, error)
}
