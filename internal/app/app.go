package app

import (
	"context"
	"errors"

	contextPkg "github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
	"go.uber.org/zap"
)

/**
Оркестратор для управлением взаимодействием сервисов.
*/

type UserService interface {
	Register(ctx context.Context, user forms.UserForm) (string, *entities.DomainError)
	Login(ctx context.Context, user forms.LoginForm) (string, *entities.DomainError)
	CheckJWTToken(token string) (int32, *entities.DomainError)
}

type OrderService interface {
	CreateOrder(ctx context.Context, orderID string, userID int32) *entities.DomainError
}

type app struct {
	userService  UserService
	orderService OrderService
	logger       interfaces.Logger
}

func (a *app) CreateOrder(ctx context.Context, orderID string) *entities.DomainError {
	contextAuth, ok := ctx.(*contextPkg.Auth)
	if !ok {
		err := errors.New("ctx without ctx.Auth")
		a.logger.Error("ctx is not contextAuth", zap.Error(err))

		return entities.NewInternalServerError(err, "")
	}

	userID := contextAuth.UserID()
	if userID == 0 {
		return entities.NewUnauthorizedError(nil, "")
	}

	err := a.orderService.CreateOrder(ctx, orderID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) CheckJWTToken(token string) (int32, *entities.DomainError) {
	userID, err := a.userService.CheckJWTToken(token)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (a *app) Login(ctx context.Context, request models.LoginRequest) (string, *entities.DomainError) {
	tokenJWT, err := a.userService.Login(ctx, forms.LoginForm{
		Login:    request.Login,
		Password: request.Password,
	})

	if err != nil {
		return "", err
	}

	return tokenJWT, nil
}

func (a *app) Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError) {
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
) *app {
	return &app{
		userService:  userService,
		orderService: orderService,
		logger:       logger,
	}
}
