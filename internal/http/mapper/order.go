package mapper

import (
	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
)

func OrderToDTO(order domainorder.Order) dto.OrderResponse {
	return dto.OrderResponse{
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: order.CreatedAt,
	}
}

func OrdersToDTO(orders []domainorder.Order) []dto.OrderResponse {
	dtoList := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		dtoList[i] = OrderToDTO(o)
	}
	return dtoList
}
