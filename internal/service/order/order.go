package order

import (
	"context"
	"strings"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces"
	"go.uber.org/zap"
)

type Repository interface {
	FindByOrderID(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError)
	CreateOrder(ctx context.Context, orderID string, userID int32, status entities.OrderStatus) *entities.DomainError
	FindByUserID(ctx context.Context, userID int32, limit int64, offset int64) ([]entities.Order, *entities.DomainError)
	CountOrdersByUserID(ctx context.Context, userID int32) (int32, *entities.DomainError)
}

type OrderQueue interface {
	AddOrderIDToQueue(orderID string)
}

type Order struct {
	repository Repository
	logger     interfaces.Logger
	orderQueue OrderQueue
}

func (o *Order) OrdersListUser(
	ctx context.Context,
	userID int32,
	pagination *pagination.Pagination,
) ([]entities.Order, *entities.DomainError) {
	countOrders, err := o.repository.CountOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if countOrders == 0 {
		return nil, nil
	}

	pagination.SetTotalCount(int64(countOrders))

	orders, err := o.repository.FindByUserID(ctx, userID, pagination.GetPerPage(), pagination.GetOffset())
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *Order) CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError {
	orderID = strings.Trim(orderID, " ")

	errDomain := o.validateOrderID(ctx, orderID, userID)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.repository.CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew)
	if errDomain != nil {
		return errDomain
	}

	o.orderQueue.AddOrderIDToQueue(orderID)

	o.logger.Info("Заказ отправлен в очередь на обработку", zap.String("order_id", orderID))

	return nil
}

func (o *Order) validateOrderID(ctx context.Context, id string, userID int32) *entities.DomainError {
	if !helpers.ValidateLuhn(id) {
		return entities.NewUnprocessableEntity(nil, "неверный формат номера заказа")
	}

	ord, err := o.repository.FindByOrderID(ctx, id)
	if err != nil {
		return err
	}

	if ord != nil {
		if ord.UserID == userID {
			return entities.NewOkError(nil, "номер заказа уже был загружен этим пользователем")
		}

		return entities.NewConflictError(nil, "номер заказа уже был загружен другим пользователем")
	}

	return nil
}

func NewOrder(
	repository Repository,
	orderQueue OrderQueue,
	logger interfaces.Logger,
) *Order {
	return &Order{
		repository: repository,
		logger:     logger,
		orderQueue: orderQueue,
	}
}
