package app

import (
	"context"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
)

/**
Оркестратор для управлением взаимодействием сервисов.
*/

type UserService interface {
	Register(ctx context.Context, user forms.UserForm) (string, *entities.DomainError)
	Login(ctx context.Context, user forms.LoginForm) (string, *entities.DomainError)
	CheckJWTToken(token string) (int32, *entities.DomainError)
}

type app struct {
	userService UserService
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

// Регистрация производится по паре логин/пароль. Каждый логин должен быть уникальным.
//	После успешной регистрации должна происходить автоматическая аутентификация пользователя.
//Для передачи аутентификационных данных используйте механизм cookies или HTTP-заголовок Authorization

/*
Возможные коды ответа:
200 — пользователь успешно зарегистрирован и аутентифицирован;
400 — неверный формат запроса;
409 — логин уже занят;
500 — внутренняя ошибка сервера.
*/
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

func NewApp(userService UserService) *app {
	return &app{
		userService: userService,
	}
}
