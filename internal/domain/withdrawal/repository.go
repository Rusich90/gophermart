package withdrawal

import (
	"context"

	"github.com/google/uuid"
)

type WithdrawalRepo interface {
	GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]Withdrawal, error)
	GetSumByUserID(ctx context.Context, userID *uuid.UUID) (float64, error)
	Create(ctx context.Context, withdrawal *Withdrawal) error
}
