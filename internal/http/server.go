package http

import (
	"context"
	"fmt"
	"time"

	"github.com/Rusich90/gophermart.git/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/Rusich90/gophermart.git/internal/service"
	"go.uber.org/zap"
)

func SetupServer(cfg *config.Config) (*gin.Engine, *pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := db.Ping(pingCtx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	userRepo := repository.NewUserRepo(db)
	authService := service.NewAuthService(userRepo, []byte(cfg.AuthSecret))
	authHandler := handler.NewAuthHandler(authService, logger)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware())

	r.POST("/api/user/register", authHandler.Register)
	r.POST("/api/user/login", authHandler.Login)

	return r, db, nil
}
