package service

import (
	"context"
	"fmt"

	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/google/uuid"
)

type WithdrawalService struct {
	withdrawalRepo domainwithdrawal.WithdrawalRepo
}

func NewWithdrawalService(withdrawalRepo domainwithdrawal.WithdrawalRepo) *WithdrawalService {
	return &WithdrawalService{withdrawalRepo: withdrawalRepo}
}

func (s *WithdrawalService) GetAllByUserID(ctx context.Context, userID *uuid.UUID) ([]domainwithdrawal.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.GetAllByUserID: %w", err)
	}

	return withdrawals, nil
}
