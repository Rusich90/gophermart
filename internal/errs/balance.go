package errs

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

func IsErrInsufficientFunds(err error) bool {
	return errors.Is(err, ErrInsufficientFunds)
}
