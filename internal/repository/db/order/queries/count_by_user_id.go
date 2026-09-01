package queries

import (
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewCountByUserID(userID int32) postgres.SelectStatement {
	return postgres.SELECT(
		postgres.COUNT(table.Orders.ID),
	).FROM(table.Orders).
		WHERE(table.Orders.UserID.EQ(postgres.Int32(userID)))
}
