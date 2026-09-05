package order

import (
	"context"
	"errors"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	dbPkg "github.com/bazueva/gofermart/internal/repository/db"
	"github.com/bazueva/gofermart/internal/repository/db/order/queries"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 1 * time.Second
)

type repository struct {
	db              interfaces.DB
	logger          interfaces.Logger
	errorClassifier *dbPkg.PostgresErrorClassifier
}

func (r *repository) CreateOrderWithWithdraw(ctx context.Context, tx interfaces.Tx, userID int32, orderID string, bonusSum float64) *entities.DomainError {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := queries.NewCreateOrderWithdraw(
		orderID,
		userID,
		bonusSum*-1,
		entities.OrdersStatusProcessed,
	).
		ExecContext(ctxWithTimeout, tx)
	if err != nil {
		r.logger.Error("error repository CreateOrderWithWithdraw", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	return nil
}

func (r *repository) UserBalance(ctx context.Context, tx interfaces.Tx, userID int32) (float64, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result struct {
		Sum float64
	}

	err := queries.NewUserBalanceSum(userID).
		QueryContext(ctxWithTimeout, tx, &result)

	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		r.logger.Error("error repository FindStaleOrders", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	return result.Sum, nil
}

func (r *repository) FindStaleOrders(ctx context.Context, statuses []entities.OrderStatus, limit int64) ([]string, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result []model.Orders

	err := queries.NewFindStaleOrders(statuses, limit).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		r.logger.Error("error repository FindStaleOrders", zap.Error(err))

		return nil, entities.NewInternalServerError(err, "")
	}

	return lo.Map(result, func(item model.Orders, index int) string {
		return item.OrderID
	}), nil
}

func (r *repository) UpdateStatusAndBonus(
	ctx context.Context,
	orderID string,
	status entities.OrderStatus,
	sum float64,
) *entities.DomainError {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := queries.
		NewUpdateStatusAndBonus(orderID, status, sum).
		ExecContext(ctxWithTimeout, r.db)
	if err != nil {
		r.logger.Error("error repository UpdateStatusAndBonus", zap.Error(err))
		if r.errorClassifier.ClassifyRetry(err) == dbPkg.Retriable {
			return entities.NewRetriableError(err, "")
		}

		return entities.NewInternalServerError(err, "")
	}

	return nil
}

func (r *repository) CountOrdersByUserID(ctx context.Context, filter entities.OrderFilter) (int32, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result struct {
		Count int32
	}

	err := queries.NewCountByUserID(filter).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil {
		r.logger.Error("error repository FindByOrderID", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	return result.Count, nil
}

func (r *repository) FindByUserID(
	ctx context.Context,
	filter entities.OrderFilter,
	limit int64,
	offset int64,
) ([]entities.Order, *entities.DomainError) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result []model.Orders

	err := queries.NewFindByUserID(filter, limit, offset).
		QueryContext(ctxWithTimeout, r.db, &result)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		r.logger.Error("error repository FindByOrderID", zap.Error(err))

		return nil, entities.NewInternalServerError(err, "")
	}

	return lo.Map(result, func(item model.Orders, index int) entities.Order {
		return entities.Order{
			ID:          item.ID,
			OrderID:     item.OrderID,
			Status:      hydrateOrdersStatusToDomain(item.Status),
			UserID:      item.UserID,
			CreatedAt:   item.CreatedAt,
			ProcessedAt: item.ProcessedAt,
			BonusSum: func() float64 {
				if item.BonusSum != nil {
					return *item.BonusSum
				}

				return 0
			}(),
		}
	}), nil
}

func (r *repository) CreateOrder(ctx context.Context, orderID string, userID int32, status entities.OrderStatus) *entities.DomainError {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var result struct {
		ID int32
	}
	err := queries.NewCreateOrder(orderID, userID, status).
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

func (r *repository) BeginTransaction(ctx context.Context) (interfaces.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func NewRepository(db interfaces.DB, logger interfaces.Logger) *repository {
	return &repository{
		db:              db,
		logger:          logger,
		errorClassifier: dbPkg.NewPostgresErrorClassifier(),
	}
}
