package order

import (
	"errors"
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/service/order/mocks"
	"github.com/stretchr/testify/assert"
)

func TestOrder_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).Return(nil, nil)

		mockRepo.EXPECT().
			CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew).
			Return(nil).
			Once()

		o := &order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Nil(t, err)
	})

	t.Run("error - validation luhn failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "order-123"
		userID := int32(123)

		o := &order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Equal(t, "неверный формат номера заказа", err.Error())
		assert.Equal(t, entities.UnprocessableEntityErrorType, err.ErrorType)
	})

	t.Run("error - FindByOrderID failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to create order",
		)

		mockRepo.EXPECT().FindByOrderID(ctx, orderID).
			Return(nil, domainErr)

		o := &order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - create order failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to create order",
		)

		mockRepo.EXPECT().FindByOrderID(ctx, orderID).
			Return(nil, nil)

		mockRepo.EXPECT().
			CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew).
			Return(domainErr)

		o := &order{
			repository: mockRepo,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})
}

func TestOrder_ValidateOrderID(t *testing.T) {
	t.Parallel()

	t.Run("success - valid order ID", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, nil)

		o := &order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Nil(t, err)
	})

	t.Run("error - validation luhn failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "order-123"
		userID := int32(123)

		o := &order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "неверный формат номера заказа", err.Error())
		assert.Equal(t, entities.UnprocessableEntityErrorType, err.ErrorType)
	})

	t.Run("error - FindByOrderID failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find order",
		)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(nil, domainErr)

		o := &order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Error(t, err)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - order already uploaded by this user", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		existingOrder := &entities.Order{
			ID:      1,
			OrderID: orderID,
			UserID:  userID,
			Status:  entities.OrdersStatusNew,
		}

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(existingOrder, nil)

		o := &order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "номер заказа уже был загружен этим пользователем", err.Error())
		assert.Equal(t, entities.OkEntityErrorType, err.ErrorType)
	})

	t.Run("error - order already uploaded by another user", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		existingOrder := &entities.Order{
			ID:      1,
			OrderID: orderID,
			UserID:  int32(456),
			Status:  entities.OrdersStatusNew,
		}

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).
			Return(existingOrder, nil)

		o := &order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "номер заказа уже был загружен другим пользователем", err.Error())
		assert.Equal(t, entities.ConflictErrorType, err.ErrorType)
	})
}
