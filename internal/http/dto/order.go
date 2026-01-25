package dto

import (
	"time"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
)

type OrderResponse struct {
	Number     string                  `json:"number"`
	Status     domainorder.OrderStatus `json:"status"`
	Accrual    float64                 `json:"accrual,omitempty"`
	UploadedAt time.Time               `json:"uploaded_at"`
}
