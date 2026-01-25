package http

import (
	"context"
	"fmt"
	"time"

	"github.com/Rusich90/gophermart.git/internal/client/accrual"
	"github.com/Rusich90/gophermart.git/internal/database"
	"github.com/Rusich90/gophermart.git/internal/http/middleware"
	"github.com/Rusich90/gophermart.git/internal/logger"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rusich90/gophermart.git/config"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/service"
)

func SetupServer(cfg *config.Config) (*gin.Engine, *pgxpool.Pool, error) {
	db, err := database.ConnectToDB(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx := context.Background()
	if err := database.PingDB(ctx, db); err != nil {
		database.CloseDB(db)
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := database.RunMigrations(cfg.DatabaseURI, cfg); err != nil {
		database.CloseDB(db)
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger, err := logger.NewLogger()
	if err != nil {
		database.CloseDB(db)
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
	orderService := service.NewOrderService(orderRepo, accrualClient, logger, time.Duration(cfg.PollInterval)*time.Second)
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
