package entities

import "time"

type OrderStatus string

const (
	OrdersStatusNew        OrderStatus = "NEW"
	OrdersStatusProcessing OrderStatus = "PROCESSING"
	OrdersStatusInvalid    OrderStatus = "INVALID"
	OrdersStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID          int32
	OrderID     string
	Status      OrderStatus
	UserID      int32
	BonusSum    float64
	CreatedAt   *time.Time
	ProcessedAt *time.Time
}
