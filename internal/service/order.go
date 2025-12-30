package service

import (
	"context"
	"fmt"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/google/uuid"
)

type OderService struct {
	orderRepo domainorder.OrderRepo
}

func NewOrderService(orderRepo domainorder.OrderRepo) *OderService {
	return &OderService{orderRepo: orderRepo}
}

func (s *OderService) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainorder.Order, error) {
	urls, err := s.orderRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}

	return urls, nil
}
