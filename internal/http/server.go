package http

import (
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/Rusich90/gophermart.git/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func setupMiddleware(r *gin.Engine, logger *zap.Logger) {
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware())
}

func registerPublicRoutes(public *gin.RouterGroup, authHandler *handler.AuthHandler) {
	public.POST("/register", authHandler.Register)
	public.POST("/login", authHandler.Login)
}

func registerProtectedRoutes(
	protected *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	withdrawalHandler *handler.WithdrawalHandler,
	secret []byte,
	logger *zap.Logger,
) {
	protected.Use(middleware.AuthMiddleware(secret, logger))

	protected.GET("/orders", orderHandler.GetAllByUserID)
	protected.POST("/orders", orderHandler.AddOrder)

	protected.GET("/balance", balanceHandler.GetByUserID)
	protected.POST("/balance/withdraw", balanceHandler.Withdraw)

	protected.GET("/withdrawals", withdrawalHandler.GetAllByUserID)
}

func SetupServer(
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	withdrawalHandler *handler.WithdrawalHandler,
	authSecret []byte,
) (*gin.Engine, error) {
	r := gin.New()
	setupMiddleware(r, logger)

	public := r.Group("/api/user")
	registerPublicRoutes(public, authHandler)

	protected := r.Group("/api/user")
	registerProtectedRoutes(protected, authHandler, orderHandler, balanceHandler, withdrawalHandler, authSecret, logger)

	return r, nil
}
