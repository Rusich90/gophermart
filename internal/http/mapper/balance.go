package mapper

import (
	servicedto "github.com/Rusich90/gophermart.git/internal/dto"
	"github.com/Rusich90/gophermart.git/internal/http/dto"
)

func BalanceToDTO(balance servicedto.BalanceDTO) dto.BalanceResponse {
	return dto.BalanceResponse{
		Current:   balance.CurrentAmount,
		Withdrawn: balance.TotalWithdrawal,
	}
}
