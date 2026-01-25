package user

import "errors"

var (
	ErrLoginAlreadyExists = errors.New("user with this login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func IsErrLoginAlreadyExists(err error) bool {
	return errors.Is(err, ErrLoginAlreadyExists)
}

func IsErrInvalidCredentials(err error) bool {
	return errors.Is(err, ErrInvalidCredentials)
}
