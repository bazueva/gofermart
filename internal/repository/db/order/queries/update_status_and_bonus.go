package queries

import (
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func NewUpdateStatusAndBonus(orderID string, status entities.OrderStatus, bonusSum float64) postgres.UpdateStatement {
	columns := postgres.ColumnList{
		table.Orders.Status,
		table.Orders.BonusSum,
		table.Orders.UpdatedAt,
	}

	if status == entities.OrdersStatusProcessed {
		columns = append(columns, table.Orders.ProcessedAt)
	}

	model := dbModel.Orders{
		Status:      hydrateDomainToOrdersStatus(status),
		BonusSum:    new(bonusSum),
		ProcessedAt: new(time.Now()),
		UpdatedAt:   new(time.Now()),
	}

	return table.Orders.
		UPDATE(columns).
		MODEL(model).
		WHERE(table.Orders.OrderID.EQ(postgres.String(orderID)))
}
