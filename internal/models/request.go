package models

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BalanceWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
