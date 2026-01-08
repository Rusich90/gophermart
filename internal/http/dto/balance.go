package dto

type BalanceResponse struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
