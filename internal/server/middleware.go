package server

import (
	"github.com/Rusich90/gophermart.git/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func setupMiddleware(r *gin.Engine, logger *zap.Logger) {
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.GzipMiddleware())
}

func AuthMiddleware(secret []byte, logger *zap.Logger) gin.HandlerFunc {
	return middleware.AuthMiddleware(secret, logger)
}