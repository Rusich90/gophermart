package authcontext

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func SetUserID(c *gin.Context, id *uuid.UUID) {
	c.Set(string(userIDKey), id)
}

func GetUserID(c *gin.Context) (*uuid.UUID, error) {
	val, exists := c.Get(string(userIDKey))
	if !exists || val == nil {
		return nil, errors.New("userID not found in context")
	}

	id, ok := val.(*uuid.UUID)
	if !ok {
		return nil, errors.New("userID has invalid type in context")
	}

	return id, nil
}

func RequireUserID(c *gin.Context) *uuid.UUID {
	userID, err := GetUserID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return nil
	}
	return userID
}
