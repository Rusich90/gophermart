package service

import (
	"context"
	"fmt"
	"math"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/dto"
	"github.com/google/uuid"
)

type BalanceService struct {
	orderRepo      domainorder.OrderRepo
	withdrawalRepo domainwithdrawal.WithdrawalRepo
}

func NewBalanceService(orderRepo domainorder.OrderRepo, withdrawalRepo domainwithdrawal.WithdrawalRepo) *BalanceService {
	return &BalanceService{orderRepo: orderRepo, withdrawalRepo: withdrawalRepo}
}

func (s *BalanceService) GetByUserID(ctx context.Context, userID *uuid.UUID) (*dto.BalanceDTO, error) {
	totalAccrual, err := s.orderRepo.GetSumByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}

	totalWithdrawal, err := s.withdrawalRepo.GetSumByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}
	balanceDTO := dto.BalanceDTO{
		TotalAccrual:    totalAccrual,
		TotalWithdrawal: totalWithdrawal,
		CurrentAmount:   math.Round((totalAccrual-totalWithdrawal)*10) / 10,
	}

	return &balanceDTO, nil
}
