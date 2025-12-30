package handler

import (
	"net/http"

	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/mapper"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderService *service.OderService
	logger       *zap.Logger
}

func NewOrderHandler(orderService *service.OderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, logger: logger}
}

func (h *OrderHandler) GetAllByUserID(c *gin.Context) {
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

	orders, err := h.orderService.GetAllByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unknown error"})
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	ordersDTO := mapper.OrdersToDTO(orders)

	c.JSON(http.StatusOK, ordersDTO)
}
