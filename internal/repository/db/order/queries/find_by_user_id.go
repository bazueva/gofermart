package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewFindByUserID(userID int32, limit, offset int64) postgres.SelectStatement {
	return postgres.SELECT(
		table.Orders.ID,
		table.Orders.OrderID,
		table.Orders.Status,
		table.Orders.UserID,
		table.Orders.CreatedAt,
		table.Orders.ProcessedAt,
		table.Orders.BonusSum,
	).FROM(table.Orders).
		WHERE(table.Orders.UserID.EQ(postgres.Int32(userID))).
		ORDER_BY(table.Orders.CreatedAt.DESC()).
		LIMIT(limit).
		OFFSET(offset)
}
