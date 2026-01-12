package http

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Rusich90/gophermart.git/internal/client/accrual"
	"github.com/Rusich90/gophermart.git/internal/http/middleware"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
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

	if err := runMigrations(cfg.DatabaseURI, cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	accrualClient := accrual.NewAccrualClient(
		cfg.AccrualSystemAddress,
		10,
		5,
		logger,
	)

	userRepo := repository.NewUserRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	withdrawalRepo := repository.NewWithdrawalRepo(db)

	authService := service.NewAuthService(userRepo, []byte(cfg.AuthSecret))
	orderService := service.NewOrderService(orderRepo, accrualClient, logger)
	withdrawalService := service.NewWithdrawalService(withdrawalRepo)
	balanceService := service.NewBalanceService(orderRepo, withdrawalRepo)

	authHandler := handler.NewAuthHandler(authService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService, logger)
	balanceHandler := handler.NewBalanceHandler(balanceService, logger)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware())

	public := r.Group("/api/user")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	protected := r.Group("/api/user")
	protected.Use(middleware.AuthMiddleware([]byte(cfg.AuthSecret), logger))
	{
		protected.GET("/orders", orderHandler.GetAllByUserID)
		protected.POST("/orders", orderHandler.AddOrder)

		protected.GET("/balance", balanceHandler.GetByUserID)
		protected.POST("/balance/withdraw", balanceHandler.Withdraw)

		protected.GET("/withdrawals", withdrawalHandler.GetAllByUserID)
	}

	return r, db, nil
}

func runMigrations(databaseURL string, cfg *config.Config) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.MigrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
