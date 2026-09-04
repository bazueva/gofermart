package app

import (
	"context"
	"errors"

	contextPkg "github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/domain/pagination"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, user forms.UserForm) (string, *entities.DomainError)
	Login(ctx context.Context, user forms.LoginForm) (string, *entities.DomainError)
	CheckJWTToken(token string) (int32, *entities.DomainError)
}

type OrderService interface {
	CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError
	OrdersListUser(ctx context.Context, userID int32, pagination *pagination.Pagination) ([]entities.Order, *entities.DomainError)
	BalanceWithdraw(ctx context.Context, userID int32, withdraw entities.BalanceWithdraw) *entities.DomainError
}

type App struct {
	userService  UserService
	orderService OrderService
	logger       interfaces.Logger
}

func (a *App) BalanceWithDraw(ctx context.Context, request models.BalanceWithdrawRequest) *entities.DomainError {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return err
	}

	err = a.orderService.BalanceWithdraw(ctx, userID, entities.BalanceWithdraw{
		Order: request.Order,
		Sum:   request.Sum,
	})

	if err != nil {
		return err
	}

	return nil
}

func (a *App) UserOrdersList(ctx context.Context, page int32, perPage int32) ([]entities.Order, *entities.DomainError) {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return nil, err
	}

	orders, err := a.orderService.OrdersListUser(ctx, userID, pagination.NewPagination(int64(page), int64(perPage)))
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (a *App) userIDFromContext(ctx context.Context, errorIfEmpty bool) (int32, *entities.DomainError) {
	contextAuth, ok := ctx.(*contextPkg.Auth)
	if !ok {
		err := errors.New("ctx without ctx.Auth")
		a.logger.Error("ctx is not contextAuth", zap.Error(err))

		return 0, entities.NewInternalServerError(err, "")
	}

	userID := contextAuth.UserID()
	if errorIfEmpty && userID == 0 {
		return 0, entities.NewUnauthorizedError(nil, "")
	}

	return userID, nil
}

func (a *App) CreateOrder(ctx context.Context, orderID string) *entities.DomainError {
	userID, err := a.userIDFromContext(ctx, true)
	if err != nil {
		return err
	}

	err = a.orderService.CreateOrder(ctx, orderID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) CheckJWTToken(token string) (int32, *entities.DomainError) {
	userID, err := a.userService.CheckJWTToken(token)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (a *App) Login(ctx context.Context, request models.LoginRequest) (string, *entities.DomainError) {
	tokenJWT, err := a.userService.Login(ctx, forms.LoginForm{
		Login:    request.Login,
		Password: request.Password,
	})

	if err != nil {
		return "", err
	}

	return tokenJWT, nil
}

func (a *App) Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError) {
	tokenJWT, err := a.userService.Register(ctx, forms.UserForm{
		Login:    request.Login,
		Password: request.Password,
	})
	if err != nil {
		return "", err
	}

	return tokenJWT, nil
}

func NewApp(
	userService UserService,
	orderService OrderService,
	logger interfaces.Logger,
) *App {
	return &App{
		userService:  userService,
		orderService: orderService,
		logger:       logger,
	}
}
