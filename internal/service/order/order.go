package order

import (
	"context"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
)

type Repository interface {
	FindByOrderID(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError)
	CreateOrder(ctx context.Context, orderID string, userID int32, status entities.OrderStatus) *entities.DomainError
}

type order struct {
	repository Repository
}

func (o *order) CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError {
	err := o.validateOrderID(ctx, orderID, userID)
	if err != nil {
		return err
	}

	err = o.repository.CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew)
	if err != nil {
		return err
	}

	// куда-то отправить на обработку orderID

	return nil
}

func (o *order) validateOrderID(ctx context.Context, id string, userID int32) *entities.DomainError {
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

func NewOrder(repository Repository) *order {
	return &order{
		repository: repository,
	}
}
