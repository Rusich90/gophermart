package handler

import (
	"io"
	"net/http"
	"strings"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/Rusich90/gophermart.git/internal/http/authcontext"
	"github.com/Rusich90/gophermart.git/internal/http/mapper"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderService *service.OrderService
	logger       *zap.Logger
}

func NewOrderHandler(orderService *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, logger: logger}
}

func (h *OrderHandler) GetAllByUserID(c *gin.Context) {
	userID := authcontext.RequireUserID(c)
	if userID == nil {
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

func (h *OrderHandler) AddOrder(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	defer c.Request.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "number is required"})
		return
	}

	userID := authcontext.RequireUserID(c)
	if userID == nil {
		return
	}

	err = h.orderService.AddOrder(c.Request.Context(), userID, orderNumber)

	if err != nil {
		switch {
		case domainorder.IsErrOrderOwnedByOtherUser(err):
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "not owner"})
		case domainorder.IsErrOrderAlreadyUploaded(err):
			c.Status(http.StatusOK)
		case domainorder.IsErrInvalidOrderNumber(err):
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "number is invalid"})
		default:
			h.logger.Error("Failed addOrder", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.Status(http.StatusAccepted)
}
