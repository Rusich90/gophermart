package domain

import (
	"context"
)

type UserRepo interface {
	Create(ctx context.Context, user User) error
	GetByLogin(ctx context.Context, login string) (User, error)
	LoginExists(ctx context.Context, login string) (bool, error)
}
