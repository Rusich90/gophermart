package validator_test

import (
	"testing"

	"github.com/Rusich90/gophermart.git/internal/http/dto"
	"github.com/Rusich90/gophermart.git/internal/http/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *dto.AuthRequest
		expected map[string]string
	}{
		{
			name:    "valid request",
			request: &dto.AuthRequest{Login: "user123", Password: "password123"},
			expected: map[string]string{},
		},
		{
			name:    "empty login",
			request: &dto.AuthRequest{Login: "", Password: "password123"},
			expected: map[string]string{
				"login": "login is required",
			},
		},
		{
			name:    "short login",
			request: &dto.AuthRequest{Login: "ab", Password: "password123"},
			expected: map[string]string{
				"login": "login must be between 3 and 50 characters",
			},
		},
		{
			name:    "long login",
			request: &dto.AuthRequest{Login: "aVeryLongLoginThatExceedsTheMaximumAllowedLengthOfFiftyCharactersWhichIsNotPermitted", Password: "password123"},
			expected: map[string]string{
				"login": "login must be between 3 and 50 characters",
			},
		},
		{
			name:    "empty password",
			request: &dto.AuthRequest{Login: "user123", Password: ""},
			expected: map[string]string{
				"password": "password is required",
			},
		},
		{
			name:    "short password",
			request: &dto.AuthRequest{Login: "user123", Password: "12345"},
			expected: map[string]string{
				"password": "password must be between 6 and 100 characters",
			},
		},
		{
			name:    "long password",
			request: &dto.AuthRequest{Login: "user123", Password: generateLongString(101)},
			expected: map[string]string{
				"password": "password must be between 6 and 100 characters",
			},
		},
		{
			name:    "multiple errors",
			request: &dto.AuthRequest{Login: "", Password: ""},
			expected: map[string]string{
				"login":    "login is required",
				"password": "password is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRegisterRequest(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *dto.AuthRequest
		expected map[string]string
	}{
		{
			name:    "valid request",
			request: &dto.AuthRequest{Login: "user123", Password: "password123"},
			expected: map[string]string{},
		},
		{
			name:    "empty login",
			request: &dto.AuthRequest{Login: "", Password: "password123"},
			expected: map[string]string{
				"login": "login is required",
			},
		},
		{
			name:    "empty password",
			request: &dto.AuthRequest{Login: "user123", Password: ""},
			expected: map[string]string{
				"password": "password is required",
			},
		},
		{
			name:    "multiple errors",
			request: &dto.AuthRequest{Login: "", Password: ""},
			expected: map[string]string{
				"login":    "login is required",
				"password": "password is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateLoginRequest(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to generate a string of specified length
func generateLongString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}