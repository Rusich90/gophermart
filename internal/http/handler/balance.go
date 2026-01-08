package handler

import (
	"net/http"

	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/mapper"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	balanceService *service.BalanceService
	logger         *zap.Logger
}

func NewBalanceHandler(balanceService *service.BalanceService, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService, logger: logger}
}

func (h *BalanceHandler) GetByUserID(c *gin.Context) {
	userID, err := authcontext.GetUserID(c)
	if err != nil {
		h.logger.Info("Failed to get user ID: ", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if userID == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	balance, err := h.balanceService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
		return
	}

	balanceDTO := mapper.BalanceToDTO(*balance)

	c.JSON(http.StatusOK, balanceDTO)
}
