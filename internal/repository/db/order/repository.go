package order

import (
	"context"
	"errors"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/repository/db/order/queries"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/go-jet/jet/v2/qrm"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 1 * time.Second
)

type repository struct {
	db     interfaces.DB
	logger interfaces.Logger
}

func (r *repository) CreateOrder(ctx context.Context, orderID string, userID int32, status entities.OrderStatus) *entities.DomainError {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result struct {
		ID int32
	}
	err := queries.NewCreateOrder(orderID, userID, hydrateDomainToOrdersStatus(status)).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil {
		r.logger.Error("error repository CreateOrder", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	return nil

}

func (r *repository) FindByOrderID(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result model.Orders

	err := queries.NewFindByOrderID(orderID).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		r.logger.Error("error repository FindByOrderID", zap.Error(err))

		return nil, entities.NewInternalServerError(err, "")
	}

	if result.ID == 0 {
		return nil, nil
	}

	return &entities.Order{
		ID:        result.ID,
		OrderID:   result.OrderID,
		Status:    hydrateOrdersStatusToDomain(result.Status),
		UserID:    result.UserID,
		CreatedAt: result.CreatedAt,
	}, nil
}

func NewRepository(db interfaces.DB, logger interfaces.Logger) *repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}
