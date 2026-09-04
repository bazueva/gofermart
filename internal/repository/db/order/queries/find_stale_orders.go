package queries

import (
	"time"

	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewFindStaleOrders(statuses []postgres.Expression, limit int64) postgres.SelectStatement {
	staleThreshold := time.Now().Add(-2 * time.Minute)

	return postgres.
		SELECT(table.Orders.OrderID).
		FROM(table.Orders).
		WHERE(
			table.Orders.Status.IN(statuses...).
				AND(
					table.Orders.CreatedAt.LT(postgres.TimestampzT(staleThreshold)),
				),
		).
		ORDER_BY(table.Orders.CreatedAt.ASC()).
		LIMIT(limit)
}
