package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewCreateOrder(orderID string, userID int32, status postgres.Expression) postgres.InsertStatement {
	return table.Orders.
		INSERT(
			table.Orders.OrderID,
			table.Orders.UserID,
			table.Orders.Status,
		).
		VALUES(
			orderID,
			userID,
			status,
		).
		RETURNING(table.Orders.ID.AS("id"))
}
