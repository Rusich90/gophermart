package domain

import (
	"time"

	"github.com/google/uuid"
)

type Withdrawal struct {
	OrderNum  string    `db:"order_num"`
	UserID    uuid.UUID `db:"user_id"`
	Sum       float64   `db:"sum"`
	CreatedAt time.Time `db:"created_at"`
}
