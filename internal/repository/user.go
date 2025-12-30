package repository

import (
	"context"
	"fmt"

	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user domainuser.User) error {
	query := `INSERT INTO users (id, login, password, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Login, user.Password, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login string) (domainuser.User, error) {
	var user domainuser.User
	query := `SELECT id, login, password, created_at FROM users WHERE login = $1`
	err := r.db.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		return domainuser.User{}, fmt.Errorf("failed to get user by login: %w", err)
	}
	return user, nil
}

func (r *UserRepo) LoginExists(ctx context.Context, login string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)`
	err := r.db.QueryRow(ctx, query, login).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if login exists: %w", err)
	}
	return exists, nil
}
