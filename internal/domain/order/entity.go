package order

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	Number    string      `db:"number"`
	UserID    uuid.UUID   `db:"user_id"`
	Status    OrderStatus `db:"status"`
	Accrual   float64     `db:"accrual"`
	CreatedAt time.Time   `db:"created_at"`
}
