package dto

type BalanceResponse struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}
