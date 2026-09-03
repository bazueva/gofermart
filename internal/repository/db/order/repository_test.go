package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		t.Parallel()

		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusNew

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status)
        VALUES ($1, $2, 'NEW')
        RETURNING orders.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err = repo.CreateOrder(ctx, orderID, userID, status)

		assert.Nil(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateOrder", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusProcessing

		expectedSQL := `INSERT INTO public.orders (order_id, user_id, status)
        VALUES ($1, $2, 'PROCESSING')
        RETURNING orders.id AS "id";`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID, hydrateDomainToOrdersStatusEnum(status)).
			WillReturnError(errorDB)

		err = repo.CreateOrder(ctx, orderID, userID, status)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository CreateOrder", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		orderID := "order-123"
		userID := int32(123)
		status := entities.OrdersStatusInvalid

		expectedSQL := `INSERT INTO orders (order_id, user_id, status) VALUES ($1, $2, $3) RETURNING id`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID, userID, hydrateDomainToOrdersStatusEnum(status)).
			WillReturnError(context.DeadlineExceeded)

		err = repo.CreateOrder(ctx, orderID, userID, status)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}

func TestRepository_FindByOrderID(t *testing.T) {
	t.Parallel()

	t.Run("success - find order", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		rows := sqlmock.NewRows([]string{"orders.id", "orders.order_id", "orders.status", "orders.user_id", "orders.created_at"}).
			AddRow(1, "order-123", "PROCESSED", 123, time.Now())

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnRows(rows)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(1), result.ID)
		assert.Equal(t, "order-123", result.OrderID)
		assert.Equal(t, int32(123), result.UserID)
		assert.Equal(t, entities.OrdersStatusProcessed, result.Status)
	})

	t.Run("error - order not found", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "order-not-found"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Nil(t, err)
		assert.Nil(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(errorDB)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		orderID := "order-123"
		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.order_id = $1::text;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(orderID).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindByOrderID(ctx, orderID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}

func TestRepository_CountOrdersByUserID(t *testing.T) {
	t.Parallel()

	t.Run("success - find orders", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		userID := int32(123)
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		rows := sqlmock.NewRows([]string{"count"}).
			AddRow(1)

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID).
			WillReturnRows(rows)

		result, err := repo.CountOrdersByUserID(ctx, userID)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(1), result)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		userID := int32(123)
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID).
			WillReturnError(errorDB)

		result, err := repo.CountOrdersByUserID(ctx, userID)

		assert.Equal(t, int32(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		userID := int32(123)
		expectedSQL := `SELECT COUNT(orders.id) AS "count"
        FROM public.orders
        WHERE orders.user_id = $1::integer;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.CountOrdersByUserID(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, int32(0), result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}

func TestRepository_FindByUserID(t *testing.T) {
	t.Parallel()

	t.Run("success - find orders by user ID", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		userID := int32(123)
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		createdAt := time.Now()
		rows := sqlmock.NewRows([]string{"orders.id", "orders.order_id", "orders.status", "orders.user_id", "orders.created_at"}).
			AddRow(1, "12345678903", "NEW", 123, createdAt).
			AddRow(2, "123456789015", "PROCESSED", 123, createdAt)

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID, limit, offset).
			WillReturnRows(rows)

		result, err := repo.FindByUserID(ctx, userID, limit, offset)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, entities.Order{
			ID:        1,
			OrderID:   "12345678903",
			Status:    entities.OrdersStatusNew,
			UserID:    123,
			CreatedAt: &createdAt,
		}, result[0])

		assert.Equal(t, entities.Order{
			ID:        2,
			OrderID:   "123456789015",
			Status:    entities.OrdersStatusProcessed,
			UserID:    123,
			CreatedAt: &createdAt,
		}, result[1])
	})

	t.Run("success - no orders found", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		userID := int32(123)
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID, limit, offset).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindByUserID(ctx, userID, limit, offset)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		userID := int32(123)
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID, limit, offset).
			WillReturnError(errorDB)

		result, err := repo.FindByUserID(ctx, userID, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindByOrderID", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		userID := int32(123)
		limit := int64(20)
		offset := int64(0)

		expectedSQL := `SELECT orders.id AS "orders.id",
             orders.order_id AS "orders.order_id",
             orders.status AS "orders.status",
             orders.user_id AS "orders.user_id",
             orders.created_at AS "orders.created_at"
        FROM public.orders
        WHERE orders.user_id = $1::integer
        ORDER BY orders.created_at DESC
        LIMIT $2
        OFFSET $3;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(userID, limit, offset).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindByUserID(ctx, userID, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}

func TestRepository_FindStaleOrders(t *testing.T) {
	t.Parallel()

	t.Run("success - find stale orders", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE (orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)
        ORDER BY orders.created_at ASC
        LIMIT $2;`

		rows := sqlmock.NewRows([]string{"orders.order_id"}).
			AddRow("12345678903").
			AddRow("123456789015")

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnRows(rows)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "12345678903", result[0])
		assert.Equal(t, "123456789015", result[1])
	})

	t.Run("success - no stale orders found", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE (orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)
        ORDER BY orders.created_at ASC
        LIMIT $2;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnError(qrm.ErrNoRows)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("error repository FindStaleOrders", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE (orders.status IN ('NEW', 'PROCESSING')) AND (orders.created_at < $1::timestamp with time zone)
        ORDER BY orders.created_at ASC
        LIMIT $2;`

		mock.ExpectQuery(expectedSQL).
			WithArgs(sqlmock.AnyArg(), limit).
			WillReturnError(errorDB)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository FindStaleOrders", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		statuses := []entities.OrderStatus{
			entities.OrdersStatusNew,
			entities.OrdersStatusProcessing,
		}
		limit := int64(10)

		expectedSQL := `SELECT orders.order_id AS "orders.order_id"
        FROM public.orders
        WHERE (orders.status IN ($1::text, $2::text)) AND (orders.created_at < $3::timestamp with time zone)
        LIMIT $4;`

		mock.ExpectQuery(expectedSQL).
			WithArgs("NEW", "PROCESSING", sqlmock.AnyArg(), limit).
			WillReturnError(context.DeadlineExceeded)

		result, err := repo.FindStaleOrders(ctx, statuses, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}

func TestRepository_UpdateStatusAndBonus(t *testing.T) {
	t.Parallel()

	t.Run("success - update status and bonus", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "12345678903"
		status := entities.OrdersStatusProcessed
		sum := int32(100)

		expectedSQL := `UPDATE public.orders
        SET (status, bonus_sum, updated_at, processed_at) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs(model.OrdersStatus_Processed, sum, sqlmock.AnyArg(), sqlmock.AnyArg(), orderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.UpdateStatusAndBonus(ctx, orderID, status, sum)

		assert.Nil(t, err)
	})

	t.Run("success - update status invalid and bonus", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "12345678903"
		status := entities.OrdersStatusInvalid
		sum := int32(100)

		expectedSQL := `UPDATE public.orders
        SET (status, bonus_sum, updated_at) = ($1, $2, $3)
        WHERE orders.order_id = $4::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs(model.OrdersStatus_Invalid, sum, sqlmock.AnyArg(), orderID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.UpdateStatusAndBonus(ctx, orderID, status, sum)

		assert.Nil(t, err)
	})

	t.Run("error - database error", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		errorDB := errors.New("database connection failed")

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().
			Error("error repository UpdateStatusAndBonus", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx := t.Context()

		orderID := "12345678903"
		status := entities.OrdersStatusProcessed
		sum := int32(100)

		expectedSQL := `UPDATE public.orders
        SET (status, bonus_sum, updated_at, processed_at) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs("PROCESSED", sum, sqlmock.AnyArg(), orderID).
			WillReturnError(errorDB)

		err = repo.UpdateStatusAndBonus(ctx, orderID, status, sum)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})

	t.Run("context timeout", func(t *testing.T) {
		db, mock, err := helpers.SqlMockTest(t)
		require.NoError(t, err)
		defer func() {
			_ = db.Close()
		}()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("error repository UpdateStatusAndBonus", mock2.Anything)

		repo := NewRepository(db, logger)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		orderID := "12345678903"
		status := entities.OrdersStatusProcessed
		sum := int32(100)

		expectedSQL := `UPDATE public.orders
        SET (status, bonus_sum, updated_at, processed_at) = ($1, $2, $3, $4)
        WHERE orders.order_id = $5::text;`

		mock.ExpectExec(expectedSQL).
			WithArgs("PROCESSED", sum, orderID).
			WillReturnError(context.DeadlineExceeded)

		err = repo.UpdateStatusAndBonus(ctx, orderID, status, sum)

		assert.Error(t, err)
		assert.IsType(t, &entities.DomainError{}, err)
		assert.Equal(t, entities.InternalServerErrorType, err.(*entities.DomainError).ErrorType)
	})
}
