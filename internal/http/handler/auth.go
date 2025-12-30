package handler

import (
	"net/http"

	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
	"github.com/Rusich90/gophermart.git/internal/http/validator"
	"github.com/Rusich90/gophermart.git/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErrors := validator.ValidateRegisterRequest(&req)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if domainuser.IsErrLoginAlreadyExists(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	token, err := h.authService.GenerateToken(userID)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.SetCookie("token", token, 24*60*60, "/", "", false, true)

}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErrors := validator.ValidateLoginRequest(&req)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if domainuser.IsErrInvalidCredentials(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("Failed to login user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.SetCookie("token", token, 24*60*60, "/", "", false, true)

	c.Status(http.StatusOK)
}
