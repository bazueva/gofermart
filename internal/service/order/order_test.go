package order

import (
	"errors"
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/service/order/mocks"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestOrder_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		mockQueue := mocks.NewMockOrderQueue(t)
		mockLogger := interfacesMocks.NewMockLogger(t)

		ctx := t.Context()
		orderID := "12345678903"
		userID := int32(123)

		mockRepo.EXPECT().
			FindByOrderID(ctx, orderID).Return(nil, nil)

		mockRepo.EXPECT().
			CreateOrder(ctx, orderID, userID, entities.OrdersStatusNew).
			Return(nil).
			Once()

		mockQueue.EXPECT().AddOrderIDToQueue(orderID)

		mockLogger.EXPECT().Info("Заказ отправлен в очередь на обработку", mock2.Anything)

		o := &Order{
			repository: mockRepo,
			orderQueue: mockQueue,
			logger:     mockLogger,
		}

		err := o.CreateOrder(ctx, orderID, userID)

		assert.Nil(t, err)
	})

	t.Run("error - validation luhn failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		orderID := "order-123"
		userID := int32(123)

		o := &Order{
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

		o := &Order{
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

		o := &Order{
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

		o := &Order{
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

		o := &Order{
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

		o := &Order{
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

		o := &Order{
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

		o := &Order{
			repository: mockRepo,
		}

		err := o.validateOrderID(ctx, orderID, userID)

		assert.Equal(t, "номер заказа уже был загружен другим пользователем", err.Error())
		assert.Equal(t, entities.ConflictErrorType, err.ErrorType)
	})
}

func TestOrder_OrdersListUser(t *testing.T) {
	t.Parallel()

	t.Run("success - get orders list", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		expectedOrders := []entities.Order{
			{
				ID:      1,
				OrderID: "12345678903",
				UserID:  filter.UserID,
				Status:  entities.OrdersStatusNew,
			},
			{
				ID:      2,
				OrderID: "123456789015",
				UserID:  filter.UserID,
				Status:  entities.OrdersStatusProcessed,
			},
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, filter, int64(20), int64(0)).
			Return(expectedOrders, nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Nil(t, err)
		assert.Len(t, orders, 2)
	})

	t.Run("success - no orders found", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(0), nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Nil(t, err)
		assert.Nil(t, orders)
	})

	t.Run("error - count orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to count orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(0), domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - find orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		filter := entities.OrderFilter{
			UserID:    int32(123),
			OrderType: new(entities.OrderFilterAddBalanceType),
		}
		pag := pagination.NewPagination(1, 20)

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, filter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, filter, pag.GetPerPage(), pag.GetOffset()).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersListUser(ctx, filter.UserID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}

func TestOrder_OrdersWithdrawalsListUser(t *testing.T) {
	t.Parallel()

	t.Run("success - get withdrawals list", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		expectedOrders := []entities.Order{
			{
				ID:       1,
				OrderID:  "12345678903",
				UserID:   userID,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -100.50,
			},
			{
				ID:       2,
				OrderID:  "123456789015",
				UserID:   userID,
				Status:   entities.OrdersStatusProcessed,
				BonusSum: -200.00,
			},
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, ordersFilter, pag.GetPerPage(), pag.GetOffset()).
			Return(expectedOrders, nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Nil(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, int64(2), pag.TotalCount())
	})

	t.Run("success - no withdrawals found", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(0), nil)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Nil(t, err)
		assert.Nil(t, orders)
	})

	t.Run("error - count orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to count orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(0), domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})

	t.Run("error - find orders failed", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)

		ctx := t.Context()
		userID := int32(123)
		pag := pagination.NewPagination(1, 20)

		ordersFilter := entities.OrderFilter{
			UserID:    userID,
			OrderType: new(entities.OrderFilterWriteOffBalanceType),
		}

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to find orders",
		)

		mockRepo.EXPECT().
			CountOrdersByUserID(ctx, ordersFilter).
			Return(int32(2), nil)

		mockRepo.EXPECT().
			FindByUserID(ctx, ordersFilter, pag.GetPerPage(), pag.GetOffset()).
			Return(nil, domainErr)

		o := &Order{
			repository: mockRepo,
		}

		orders, err := o.OrdersWithdrawalsListUser(ctx, userID, pag)

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, domainErr, err)
	})
}
