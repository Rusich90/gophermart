package domain

import "errors"

var (
	ErrOrderOwnedByOtherUser = errors.New("order is owned by another user")
	ErrOrderAlreadyUploaded  = errors.New("order already uploaded by user")
	ErrOrderNotFound         = errors.New("order not found")
)

func IsErrOrderOwnedByOtherUser(err error) bool {
	return errors.Is(err, ErrOrderOwnedByOtherUser)
}

func IsErrOrderAlreadyUploaded(err error) bool {
	return errors.Is(err, ErrOrderAlreadyUploaded)
}

func IsErrOrderNotFound(err error) bool {
	return errors.Is(err, ErrOrderNotFound)
}
