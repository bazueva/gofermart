package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewCreateOrderWithdraw(
	orderID string,
	userID int32,
	bonusSum float64,
	status postgres.Expression,
) postgres.InsertStatement {
	return table.Orders.
		INSERT(
			table.Orders.OrderID,
			table.Orders.UserID,
			table.Orders.Status,
			table.Orders.BonusSum,
		).
		VALUES(
			orderID,
			userID,
			status,
			bonusSum,
		).
		RETURNING(table.Orders.ID.AS("id"))
}
