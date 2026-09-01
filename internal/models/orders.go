package models

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int32  `json:"accrual,omitempty"`
	UploadedAt string `json:"uploadedAt"`
}
