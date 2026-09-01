package order

import (
	"github.com/bazueva/gofermart/internal/domain/entities"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
)

func hydrateOrdersStatusToDomain(status dbModel.OrdersStatus) entities.OrderStatus {
	switch status {
	case dbModel.OrdersStatus_New:
		return entities.OrdersStatusNew
	case dbModel.OrdersStatus_Processing:
		return entities.OrdersStatusProcessing
	case dbModel.OrdersStatus_Invalid:
		return entities.OrdersStatusInvalid
	case dbModel.OrdersStatus_Processed:
		return entities.OrdersStatusProcessed
	default:
		return entities.OrdersStatusNew
	}
}

func hydrateDomainToOrdersStatus(status entities.OrderStatus) dbModel.OrdersStatus {
	switch status {
	case entities.OrdersStatusNew:
		return dbModel.OrdersStatus_New
	case entities.OrdersStatusProcessing:
		return dbModel.OrdersStatus_Processing
	case entities.OrdersStatusInvalid:
		return dbModel.OrdersStatus_Invalid
	case entities.OrdersStatusProcessed:
		return dbModel.OrdersStatus_Processed
	default:
		return dbModel.OrdersStatus(status)
	}
}
