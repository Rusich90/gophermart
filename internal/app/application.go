package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Rusich90/gophermart.git/internal/client/accrual"
	"github.com/Rusich90/gophermart.git/internal/database"
	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/logger"
	"github.com/Rusich90/gophermart.git/internal/repository"
	"github.com/Rusich90/gophermart.git/internal/server"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Rusich90/gophermart.git/config"
)

type Application struct {
	DB            *pgxpool.Pool
	Logger        *zap.Logger
	Config        *config.Config
	AccrualClient accrual.AccrualInterface
	Repositories  *Repositories
	Services      *Services
	Handlers      *Handlers
	Server        *server.Server
}

// Repository constructors
func NewUserRepo(db *pgxpool.Pool) domainuser.UserRepo {
	return repository.NewUserRepo(db)
}

func NewOrderRepo(db *pgxpool.Pool) domainorder.OrderRepo {
	return repository.NewOrderRepo(db)
}

func NewWithdrawalRepo(db *pgxpool.Pool) domainwithdrawal.WithdrawalRepo {
	return repository.NewWithdrawalRepo(db)
}

// Service constructors
func (app *Application) NewAuthService(userRepo domainuser.UserRepo) *service.AuthService {
	return service.NewAuthService(userRepo, []byte(app.Config.AuthSecret))
}

func (app *Application) NewBalanceService(orderRepo domainorder.OrderRepo, withdrawalRepo domainwithdrawal.WithdrawalRepo) *service.BalanceService {
	return service.NewBalanceService(orderRepo, withdrawalRepo)
}

func (app *Application) NewOrderService(orderRepo domainorder.OrderRepo, accrualClient accrual.AccrualInterface, logger *zap.Logger) *service.OrderService {
	return service.NewOrderService(orderRepo, accrualClient, logger, time.Duration(app.Config.PollInterval)*time.Second)
}

func (app *Application) NewWithdrawalService(withdrawalRepo domainwithdrawal.WithdrawalRepo) *service.WithdrawalService {
	return service.NewWithdrawalService(withdrawalRepo)
}

// Handler constructors
func (app *Application) NewAuthHandler(authService *service.AuthService) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, app.Logger)
}

func (app *Application) NewBalanceHandler(balanceService *service.BalanceService) *handler.BalanceHandler {
	return handler.NewBalanceHandler(balanceService, app.Logger)
}

func (app *Application) NewOrderHandler(orderService *service.OrderService) *handler.OrderHandler {
	return handler.NewOrderHandler(orderService, app.Logger)
}

func (app *Application) NewWithdrawalHandler(withdrawalService *service.WithdrawalService) *handler.WithdrawalHandler {
	return handler.NewWithdrawalHandler(withdrawalService, app.Logger)
}

func (app *Application) initRepositories() error {
	userRepo := NewUserRepo(app.DB)
	orderRepo := NewOrderRepo(app.DB)
	withdrawalRepo := NewWithdrawalRepo(app.DB)

	app.Repositories = &Repositories{
		UserRepo:       userRepo,
		OrderRepo:      orderRepo,
		WithdrawalRepo: withdrawalRepo,
	}

	return nil
}

func (app *Application) initServices() error {
	authService := app.NewAuthService(app.Repositories.UserRepo)
	balanceService := app.NewBalanceService(app.Repositories.OrderRepo, app.Repositories.WithdrawalRepo)
	orderService := app.NewOrderService(app.Repositories.OrderRepo, app.AccrualClient, app.Logger)
	withdrawalService := app.NewWithdrawalService(app.Repositories.WithdrawalRepo)

	app.Services = &Services{
		AuthService:       authService,
		BalanceService:    balanceService,
		OrderService:      orderService,
		WithdrawalService: withdrawalService,
	}

	return nil
}

func (app *Application) initHandlers() error {
	authHandler := app.NewAuthHandler(app.Services.AuthService)
	balanceHandler := app.NewBalanceHandler(app.Services.BalanceService)
	orderHandler := app.NewOrderHandler(app.Services.OrderService)
	withdrawalHandler := app.NewWithdrawalHandler(app.Services.WithdrawalService)

	app.Handlers = &Handlers{
		AuthHandler:       authHandler,
		BalanceHandler:    balanceHandler,
		OrderHandler:      orderHandler,
		WithdrawalHandler: withdrawalHandler,
	}

	return nil
}

type Repositories struct {
	UserRepo       domainuser.UserRepo
	OrderRepo      domainorder.OrderRepo
	WithdrawalRepo domainwithdrawal.WithdrawalRepo
}

type Services struct {
	AuthService       *service.AuthService
	OrderService      *service.OrderService
	WithdrawalService *service.WithdrawalService
	BalanceService    *service.BalanceService
}

type Handlers struct {
	AuthHandler       *handler.AuthHandler
	OrderHandler      *handler.OrderHandler
	WithdrawalHandler *handler.WithdrawalHandler
	BalanceHandler    *handler.BalanceHandler
}

func Initialize(cfg *config.Config) (*Application, error) {
	db, err := database.ConnectToDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx := context.Background()
	if err := database.PingDB(ctx, db); err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := database.RunMigrations(cfg.DatabaseURI, cfg); err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger, err := logger.NewLogger()
	if err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	accrualClient := accrual.NewAccrualClient(
		cfg.AccrualSystemAddress,
		cfg.MaxConcurrentRequests,
		5,
		logger,
	)

	app := &Application{
		DB:            db,
		Logger:        logger,
		Config:        cfg,
		AccrualClient: accrualClient,
	}

	if err := app.initRepositories(); err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := app.initServices(); err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := app.initHandlers(); err != nil {
		database.CloseDB(db)
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	return app, nil
}

func (app *Application) Run() error {
	app.Server = server.NewServer(
		app.Logger,
		*app.Handlers.AuthHandler,
		*app.Handlers.OrderHandler,
		*app.Handlers.BalanceHandler,
		*app.Handlers.WithdrawalHandler,
		[]byte(app.Config.AuthSecret),
		app.Config.RunAddress,
	)

	return app.Server.Run()
}
