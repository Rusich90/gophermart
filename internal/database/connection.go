package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToDB(config *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, config.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func PingDB(ctx context.Context, db *pgxpool.Pool) error {
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := db.Ping(pingCtx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func CloseDB(db *pgxpool.Pool) {
	if db != nil {
		db.Close()
	}
}
