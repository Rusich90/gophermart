package accrual

import "errors"

var (
	ErrOrderNotRegistered = errors.New("order not registered")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrInternalServer     = errors.New("server error")
)

func IsErrOrderNotRegistered(err error) bool {
	return errors.Is(err, ErrOrderNotRegistered)
}

func IsErrTooManyRequests(err error) bool {
	return errors.Is(err, ErrTooManyRequests)
}

func IsErrInternalServer(err error) bool {
	return errors.Is(err, ErrInternalServer)
}
