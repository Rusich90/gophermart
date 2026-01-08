package validator

import (
	"github.com/Rusich90/gophermart.git/internal/http/dto"
)

func ValidateWithdrawRequest(req *dto.WithdrawRequest) map[string]string {
	errors := make(map[string]string)

	if req.Order == "" {
		errors["order"] = "order number is required"
	} else if len(req.Order) < 1 || len(req.Order) > 50 {
		errors["order"] = "order number must be between 1 and 50 characters"
	}

	if req.Sum <= 0 {
		errors["sum"] = "amount must be greater than zero"
	} else if req.Sum < 0.01 {
		errors["sum"] = "amount must be at least 0.01"
	}

	return errors
}
