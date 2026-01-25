package server

import (
	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func registerPublicRoutes(public *gin.RouterGroup, authHandler handler.AuthHandler) {
	public.POST("/register", authHandler.Register)
	public.POST("/login", authHandler.Login)
}

func registerProtectedRoutes(
	protected *gin.RouterGroup,
	orderHandler handler.OrderHandler,
	balanceHandler handler.BalanceHandler,
	withdrawalHandler handler.WithdrawalHandler,
	secret []byte,
	logger *zap.Logger,
) {
	protected.Use(AuthMiddleware(secret, logger))

	protected.GET("/orders", orderHandler.GetAllByUserID)
	protected.POST("/orders", orderHandler.AddOrder)

	protected.GET("/balance", balanceHandler.GetByUserID)
	protected.POST("/balance/withdraw", balanceHandler.Withdraw)

	protected.GET("/withdrawals", withdrawalHandler.GetAllByUserID)
}
