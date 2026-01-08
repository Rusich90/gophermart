package service

import (
	"context"
	"fmt"
	"math"

	domainorder "github.com/Rusich90/gophermart.git/internal/domain/order"
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/dto"
	"github.com/Rusich90/gophermart.git/internal/errs"
	"github.com/Rusich90/gophermart.git/internal/utils"
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

func (s *BalanceService) Withdraw(ctx context.Context, withdrawal *domainwithdrawal.Withdrawal) error {
	if !utils.IsValidLuhn(withdrawal.OrderNum) {
		return domainorder.ErrInvalidOrderNumber
	}

	balance, err := s.GetByUserID(ctx, &withdrawal.UserID)
	if err != nil {
		return fmt.Errorf("get balance error: %w", err)
	}

	if balance.CurrentAmount < withdrawal.Sum {
		return errs.ErrInsufficientFunds
	}

	err = s.withdrawalRepo.Create(ctx, withdrawal)
	if err != nil {
		if !domainwithdrawal.IsErrOrderNumConflict(err) {
			return nil
		}
		return fmt.Errorf("withdrawalRepo.Create: %w", err)
	}

	return nil
}
