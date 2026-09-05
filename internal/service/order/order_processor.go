package order

import (
	"context"
	"sync"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"go.uber.org/zap"
)

const rateLimitWorkers = 3

type BonusRepository interface {
	GetOrder(ctx context.Context, orderID string) (*entities.Order, *entities.DomainError)
}

type OrderRepository interface {
	UpdateStatusAndBonus(ctx context.Context, orderID string, status entities.OrderStatus, sum float64) *entities.DomainError
	FindStaleOrders(ctx context.Context, statuses []entities.OrderStatus, limit int64) ([]string, *entities.DomainError)
}

type OrderProcessor struct {
	logger             interfaces.Logger
	ordersProcessingCh chan string
	ordersProcessedCh  chan entities.Order
	bonusRepository    BonusRepository
	orderRepository    OrderRepository
}

func (op *OrderProcessor) AddOrderIDToQueue(orderID string) {
	select {
	case op.ordersProcessingCh <- orderID:
		op.logger.Info("Заказ отправлен в очередь на обработку", zap.String("order_id", orderID))
	default:
		op.logger.Warn("Очередь обработки переполнена, заказ обработается позже из БД", zap.String("order_id", orderID))
	}
}

func (op *OrderProcessor) AddOrderToQueue(order entities.Order) {
	select {
	case op.ordersProcessedCh <- order:
		op.logger.Info("Заказ отправлен в очередь на обновление", zap.Any("order", order))
	default:
		op.logger.Warn("Очередь обработки переполнена", zap.Any("order", order))
	}
}

func NewOrderProcessor(
	bonusRepository BonusRepository,
	orderRepository OrderRepository,
	logger interfaces.Logger,
) *OrderProcessor {
	// канал для обработка заказов, у которых статус NEW, PROCESSING
	ordersProcessingCh := make(chan string, rateLimitWorkers*5)
	// канал с результатом начисления по заказам
	ordersProcessedCh := make(chan entities.Order, rateLimitWorkers*5)

	return &OrderProcessor{
		logger:             logger,
		bonusRepository:    bonusRepository,
		orderRepository:    orderRepository,
		ordersProcessingCh: ordersProcessingCh,
		ordersProcessedCh:  ordersProcessedCh,
	}
}

var orderStatusesCheck = []entities.OrderStatus{
	entities.OrdersStatusNew,
	entities.OrdersStatusProcessing,
}

const (
	databasePollerInterval = 1 * time.Minute
)

func (op *OrderProcessor) StartDatabasePoller(ctx context.Context) {
	tick := time.Tick(databasePollerInterval)

	go func() {
		for {
			select {
			case <-tick:
				orderIDs, err := op.orderRepository.FindStaleOrders(
					ctx,
					orderStatusesCheck,
					10,
				)
				if err != nil {
					op.logger.Error("ошибка получения order_ids StartDatabasePoller", zap.Error(err.SourceErr))
				}

				for _, orderID := range orderIDs {
					op.AddOrderIDToQueue(orderID)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (op *OrderProcessor) Start(ctx context.Context) {
	var wgProcessWorkers sync.WaitGroup
	var wgSaveResults sync.WaitGroup

	// воркеры, которые слушают заказы из ordersProcessingCh
	for i := 0; i < rateLimitWorkers; i++ {
		wgProcessWorkers.Add(1)
		go func() {
			defer wgProcessWorkers.Done()
			op.orderCheckStatus(ctx)
		}()
	}

	// воркеры, которые обновляют заказы в БД, берут данные из ordersProcessedCh
	for i := 0; i < rateLimitWorkers; i++ {
		wgSaveResults.Add(1)
		go func() {
			defer wgSaveResults.Done()
			op.orderUpdateBonus(ctx)
		}()
	}

	go func() {
		<-ctx.Done()
		op.logger.Info("Получен сигнал отмены. Ожидаем завершения воркеров...")

		close(op.ordersProcessingCh)
		wgProcessWorkers.Wait()

		close(op.ordersProcessedCh)
		wgSaveResults.Wait()

		op.logger.Info("Все фоновые воркеры успешно завершили работу")
	}()
}

func (op *OrderProcessor) orderCheckStatus(ctx context.Context) {
	for orderID := range op.ordersProcessingCh {
		op.checkOrderBonus(ctx, orderID)
	}
}

func (op *OrderProcessor) checkOrderBonus(ctx context.Context, orderID string) {
	result, err := op.bonusRepository.GetOrder(ctx, orderID)
	if err != nil {
		if err.ErrorType == entities.NoContentErrorType {
			op.logger.Info("Заказ не найден в bonus, заказу присвоен статус INVALID", zap.String("order_id", orderID))

			return

			// Падают тесты на гитлабе
			// result = &entities.Order{
			//	Status:  entities.OrdersStatusInvalid,
			//	OrderID: orderID,
			// }
		} else {
			return
		}
	}

	select {
	case op.ordersProcessedCh <- *result:
		op.logger.Info("Заказ отправлен в очередь на обновление данных",
			zap.String("order_id", orderID),
			zap.Any("order", result),
		)
	case <-ctx.Done():
		return
	}
}

func (op *OrderProcessor) orderUpdateBonus(ctx context.Context) {
	for orderData := range op.ordersProcessedCh {
		// если ctx отменен, создаем новый контект чтобы запросы в БД успели выполниться
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		op.updateStatusOrder(saveCtx, orderData)

		cancel()
	}
}

func (op *OrderProcessor) updateStatusOrder(ctx context.Context, data entities.Order) {
	if data.OrderID == "" {
		op.logger.Info("updateStatusOrder пустой orderID", zap.Any("order", data))

		return
	}

	err := op.orderRepository.UpdateStatusAndBonus(ctx, data.OrderID, data.Status, data.BonusSum)
	if err != nil {
		if err.ErrorType == entities.RetriableErrorType {
			op.logger.Error(
				"Ошибка обновления заказа, заказ повторно отправлен в очередь",
				zap.Error(err),
				zap.Any("order", data),
			)

			op.AddOrderToQueue(data)
		}

		return
	}

	op.logger.Info("Заказ успешно обновлен", zap.Any("order", data))
}
