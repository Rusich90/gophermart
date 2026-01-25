package withdrawal

import "errors"

var (
	ErrOrderNumConflict = errors.New("withdrawal with this order number already exists")
)

func IsErrOrderNumConflict(err error) bool {
	return errors.Is(err, ErrOrderNumConflict)
}
