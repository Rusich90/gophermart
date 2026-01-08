package service

import (
	"context"
	"fmt"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	"github.com/Rusich90/gophermart.git/internal/utils"
	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo domainorder.OrderRepo
}

func NewOrderService(orderRepo domainorder.OrderRepo) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

func (s *OrderService) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainorder.Order, error) {
	orders, err := s.orderRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}

	return orders, nil
}

func (s *OrderService) AddOrder(ctx context.Context, userID *uuid.UUID, number string) error {
	if !utils.IsValidLuhn(number) {
		return domainorder.ErrInvalidOrderNumber
	}
	
	order, err := s.orderRepo.GetByNumber(ctx, number)
	if err != nil {
		if !domainorder.IsErrOrderNotFound(err) {
			return fmt.Errorf("failed to check order existence: %w", err)
		}
	} else {
		if order.UserID != *userID {
			return domainorder.ErrOrderOwnedByOtherUser
		}
		return domainorder.ErrOrderAlreadyUploaded
	}

	err = s.orderRepo.Create(ctx, number, userID)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}
