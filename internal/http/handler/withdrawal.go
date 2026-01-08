package handler

import (
	"net/http"

	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/mapper"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WithdrawalHandler struct {
	withdrawalService *service.WithdrawalService
	logger            *zap.Logger
}

func NewWithdrawalHandler(withdrawalService *service.WithdrawalService, logger *zap.Logger) *WithdrawalHandler {
	return &WithdrawalHandler{withdrawalService: withdrawalService, logger: logger}
}

func (h *WithdrawalHandler) GetAllByUserID(c *gin.Context) {
	userID := authcontext.RequireUserID(c)
	if userID == nil {
		return
	}

	withdrawals, err := h.withdrawalService.GetAllByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	ordersDTO := mapper.WithdrawalsToDTO(withdrawals)

	c.JSON(http.StatusOK, ordersDTO)
}
