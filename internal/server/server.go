package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Rusich90/gophermart.git/internal/http/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func NewServer(
	logger *zap.Logger,
	authHandler handler.AuthHandler,
	orderHandler handler.OrderHandler,
	balanceHandler handler.BalanceHandler,
	withdrawalHandler handler.WithdrawalHandler,
	authSecret []byte,
	runAddress string,
) *Server {
	router := gin.New()

	setupMiddleware(router, logger)

	public := router.Group("/api/user")
	registerPublicRoutes(public, authHandler)

	protected := router.Group("/api/user")
	registerProtectedRoutes(protected, orderHandler, balanceHandler, withdrawalHandler, authSecret, logger)

	httpServer := &http.Server{
		Addr:    runAddress,
		Handler: router,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

func (s *Server) Run() error {
	s.logger.Info(fmt.Sprintf("Starting server on %s", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("Server exited properly")
	return nil
}
