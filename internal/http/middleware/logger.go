package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		logger.Info("HTTP request",
			zap.String("uri", c.Request.URL.RequestURI()),
			zap.String("method", c.Request.Method),
			zap.Duration("duration", duration),
			zap.Int("status_code", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}
