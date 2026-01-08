package handler

import (
	"net/http"
	"time"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/errs"
	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
	"github.com/Rusich90/gophermart.git/internal/http/mapper"
	"github.com/Rusich90/gophermart.git/internal/http/validator"
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
	userID := authcontext.RequireUserID(c)
	if userID == nil {
		return
	}

	balance, err := h.balanceService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
		return
	}

	balanceDTO := mapper.BalanceToDTO(*balance)

	c.JSON(http.StatusOK, balanceDTO)
}

func (h *BalanceHandler) Withdraw(c *gin.Context) {
	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErrors := validator.ValidateWithdrawRequest(&req)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	userID := authcontext.RequireUserID(c)
	if userID == nil {
		return
	}

	withdraw := &domainwithdrawal.Withdrawal{
		OrderNum:  req.Order,
		UserID:    *userID,
		Sum:       req.Sum,
		CreatedAt: time.Now(),
	}

	err := h.balanceService.Withdraw(c.Request.Context(), withdraw)
	if err != nil {
		switch {
		case domainorder.IsErrInvalidOrderNumber(err):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "number is invalid"})
			return
		case errs.IsErrInsufficientFunds(err):
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{"error": "insufficient funds"})
			return
		case domainwithdrawal.IsErrOrderNumConflict(err):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "number is invalid"})
			return
		default:
			h.logger.Error("Failed withdraw", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
			return
		}
	}

	c.Status(http.StatusOK)
}
