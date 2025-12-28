package validator

import (
	"github.com/Rusich90/gophermart.git/internal/http/dto"
)

func ValidateRegisterRequest(req *dto.AuthRequest) map[string]string {
	errors := make(map[string]string)

	if req.Login == "" {
		errors["login"] = "login is required"
	} else if len(req.Login) < 3 || len(req.Login) > 50 {
		errors["login"] = "login must be between 3 and 50 characters"
	}

	if req.Password == "" {
		errors["password"] = "password is required"
	} else if len(req.Password) < 6 || len(req.Password) > 100 {
		errors["password"] = "password must be between 6 and 100 characters"
	}

	return errors
}

func ValidateLoginRequest(req *dto.AuthRequest) map[string]string {
	errors := make(map[string]string)

	if req.Login == "" {
		errors["login"] = "login is required"
	}

	if req.Password == "" {
		errors["password"] = "password is required"
	}

	return errors
}
