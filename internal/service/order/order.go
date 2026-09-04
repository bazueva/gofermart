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
	UserBalance(ctx context.Context, db interfaces.Tx, id int32) (float64, *entities.DomainError)
	CreateOrderWithWithdraw(ctx context.Context, db interfaces.Tx, userID int32, orderID string, bonusSum float64) *entities.DomainError
	BeginTransaction(ctx context.Context) (interfaces.Tx, error)
}

type OrderQueue interface {
	AddOrderIDToQueue(orderID string)
}

type Order struct {
	repository Repository
	logger     interfaces.Logger
	orderQueue OrderQueue
}

const (
	lockTypeCreateOrderWithdraw = 100
)

func (o *Order) BalanceWithdraw(ctx context.Context, userID int32, withdraw entities.BalanceWithdraw) *entities.DomainError {
	withdraw.Order = strings.Trim(withdraw.Order, " ")

	errDomain := o.validateOrderID(ctx, withdraw.Order, userID)
	if errDomain != nil {
		if errDomain.ErrorType != entities.InternalServerErrorType {
			return entities.NewBadRequestError(nil, "неверный номер заказа")
		}

		return errDomain
	}

	tx, _ := o.repository.BeginTransaction(ctx)
	defer tx.Rollback()

	var ok bool
	err := tx.QueryRowContext(ctx, "SELECT pg_try_advisory_xact_lock($1, $2);", lockTypeCreateOrderWithdraw, int64(userID)).
		Scan(&ok)
	if err != nil {
		o.logger.Error("ошибка выполнения запроса блокировки", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	if !ok {
		o.logger.Warn("запрос отклонен: операция уже выполняется параллельно", zap.Int32("userID", userID))

		return entities.NewTooManyRequestError(nil, "операция уже выполняется, попробуйте позже")
	}

	userBalance, errDomain := o.repository.UserBalance(ctx, tx, userID)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.validateWithdraw(userID, userBalance, withdraw)
	if errDomain != nil {
		return errDomain
	}

	errDomain = o.repository.CreateOrderWithWithdraw(ctx, tx, userID, withdraw.Order, withdraw.Sum)
	if errDomain != nil {
		return errDomain
	}

	err = tx.Commit()
	if err != nil {
		o.logger.Error("ошибка Commit", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	return nil
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

func (o *Order) validateWithdraw(userID int32, userBalance float64, withdraw entities.BalanceWithdraw) *entities.DomainError {
	if userBalance <= 0 {
		if userBalance < 0 {
			o.logger.Error("У пользователя отрицательный баланс",
				zap.Int32("user_id", userID),
				zap.Float64("balance", userBalance),
			)
		}

		return entities.NewPaymentRequiredError(nil, "на счету недостаточно средств")
	}

	if withdraw.Sum <= float64(0) {
		return entities.NewBadRequestError(nil, "сумма для списания должна быть больше 0")
	}

	if withdraw.Sum > userBalance {
		return entities.NewPaymentRequiredError(nil, "на счету недостаточно средств")
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
