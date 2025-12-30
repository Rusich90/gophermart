package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	Number    string    `db:"number"`
	UserID    uuid.UUID `db:"user_id"`
	Status    string    `db:"status"`
	Accrual   float64   `db:"accrual"`
	CreatedAt time.Time `db:"created_at"`
}
