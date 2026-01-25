package mapper

import (
	domainwithdrawal "github.com/Rusich90/gophermart.git/internal/domain/withdrawal"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
)

func WithdrawalToDTO(withdrawal domainwithdrawal.Withdrawal) dto.WithdrawalResponse {
	return dto.WithdrawalResponse{
		Order:       withdrawal.OrderNum,
		Sum:         withdrawal.Sum,
		ProcessedAt: withdrawal.CreatedAt,
	}
}

func WithdrawalsToDTO(withdrawals []domainwithdrawal.Withdrawal) []dto.WithdrawalResponse {
	dtoList := make([]dto.WithdrawalResponse, len(withdrawals))
	for i, w := range withdrawals {
		dtoList[i] = WithdrawalToDTO(w)
	}
	return dtoList
}
