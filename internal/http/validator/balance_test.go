package validator_test

import (
	"testing"

	"github.com/Rusich90/gophermart.git/internal/http/dto"
	"github.com/Rusich90/gophermart.git/internal/http/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidateWithdrawRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *dto.WithdrawRequest
		expected map[string]string
	}{
		{
			name:    "valid request",
			request: &dto.WithdrawRequest{Order: "12345", Sum: 100.50},
			expected: map[string]string{},
		},
		{
			name:    "empty order",
			request: &dto.WithdrawRequest{Order: "", Sum: 100.50},
			expected: map[string]string{
				"order": "order number is required",
			},
		},
		{
			name:    "short order",
			request: &dto.WithdrawRequest{Order: "", Sum: 100.50},
			expected: map[string]string{
				"order": "order number is required",
			},
		},
		{
			name:    "long order",
			request: &dto.WithdrawRequest{Order: generateLongString(51), Sum: 100.50},
			expected: map[string]string{
				"order": "order number must be between 1 and 50 characters",
			},
		},
		{
			name:    "zero sum",
			request: &dto.WithdrawRequest{Order: "12345", Sum: 0},
			expected: map[string]string{
				"sum": "amount must be greater than zero",
			},
		},
		{
			name:    "negative sum",
			request: &dto.WithdrawRequest{Order: "12345", Sum: -50.25},
			expected: map[string]string{
				"sum": "amount must be greater than zero",
			},
		},
		{
			name:    "too small sum",
			request: &dto.WithdrawRequest{Order: "12345", Sum: 0.001},
			expected: map[string]string{
				"sum": "amount must be at least 0.01",
			},
		},
		{
			name:    "multiple errors",
			request: &dto.WithdrawRequest{Order: "", Sum: 0},
			expected: map[string]string{
				"order": "order number is required",
				"sum":   "amount must be greater than zero",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateWithdrawRequest(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}