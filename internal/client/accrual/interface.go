package accrual

import "context"

type AccrualInterface interface {
	GetAccrualInfo(ctx context.Context, orderNumber string) (*AccrualResponse, error)
}
