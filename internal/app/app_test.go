package app

import (
	"errors"
	"testing"

	appMocks "github.com/bazueva/gofermart/internal/app/mocks"
	"github.com/bazueva/gofermart/internal/context"
	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/bazueva/gofermart/internal/models/forms"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestApp_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("ctx without Auth ctx", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)
		logger.EXPECT().Error("ctx is not contextAuth", mock2.Anything)

		appTest := NewApp(nil, nil, logger)
		err := appTest.CreateOrder(t.Context(), "test")

		assert.Equal(t, "Internal Server Error", err.Error())
	})

	t.Run("userID = 0", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.NewAuth(t.Context())
		appTest := NewApp(nil, nil, logger)
		err := appTest.CreateOrder(ctx, "test")

		assert.Equal(t, "пользователь не аутентифицирован", err.Error())
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
	})

	t.Run("error create order", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.NewAuth(t.Context())
		ctx.WithUserID(20)

		orderService := appMocks.NewMockOrderService(t)
		orderService.EXPECT().CreateOrder(ctx, "12323", int32(20)).
			Return(entities.NewInternalServerError(nil, ""))

		appTest := NewApp(nil, orderService, logger)
		err := appTest.CreateOrder(ctx, "12323")

		assert.Equal(t, "Internal Server Error", err.Error())
		assert.Equal(t, entities.InternalServerErrorType, err.ErrorType)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		logger := mocks.NewMockLogger(t)

		ctx := context.NewAuth(t.Context())
		ctx.WithUserID(20)

		orderService := appMocks.NewMockOrderService(t)
		orderService.EXPECT().CreateOrder(ctx, "12323", int32(20)).
			Return(nil)

		appTest := NewApp(nil, orderService, logger)
		err := appTest.CreateOrder(ctx, "12323")

		assert.Nil(t, err)
	})
}

func TestApp_CheckJWTToken_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("error userService", func(t *testing.T) {
		mockUserService := appMocks.NewMockUserService(t)

		domainErr := entities.NewBadRequestError(
			errors.New("empty token"),
			"token cannot be empty",
		)
		mockUserService.EXPECT().CheckJWTToken("").
			Return(int32(0), domainErr).
			Once()

		appMocks := &app{
			userService: mockUserService,
		}

		userID, err := appMocks.CheckJWTToken("")

		assert.Error(t, err)
		assert.Equal(t, int32(0), userID)
		assert.Equal(t, entities.BadRequestErrorType, err.ErrorType)

		mockUserService.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockUserService := appMocks.NewMockUserService(t)

		mockUserService.EXPECT().CheckJWTToken("token").
			Return(int32(1), nil).
			Once()

		appTest := &app{
			userService: mockUserService,
		}

		userID, err := appTest.CheckJWTToken("token")

		assert.Nil(t, err)
		assert.Equal(t, int32(1), userID)
	})
}

func TestApp_Login(t *testing.T) {
	t.Parallel()

	t.Run("success - valid credentials", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		request := models.LoginRequest{
			Login:    "testuser",
			Password: "correct_password",
		}
		expectedToken := "valid.jwt.token"

		mockUserService.EXPECT().
			Login(t.Context(), forms.LoginForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return(expectedToken, nil)

		appTest := &app{
			userService: mockUserService,
		}

		token, err := appTest.Login(t.Context(), request)

		assert.Nil(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("error - invalid credentials", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.LoginRequest{
			Login:    "testuser",
			Password: "wrong_password",
		}
		domainErr := entities.NewUnauthorizedError(
			errors.New("invalid credentials"),
			"invalid login or password",
		)

		mockUserService.EXPECT().
			Login(ctx, forms.LoginForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return("", domainErr)

		appTest := &app{
			userService: mockUserService,
		}

		token, err := appTest.Login(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, "", token)
		assert.Equal(t, entities.UnauthorizedErrorType, err.ErrorType)
		mockUserService.AssertExpectations(t)
	})
}

func TestApp_Register(t *testing.T) {
	t.Parallel()

	t.Run("success - valid registration", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.RegisterRequest{
			Login:    "newuser",
			Password: "password123",
		}
		expectedToken := "valid.jwt.token"

		mockUserService.EXPECT().
			Register(ctx, forms.UserForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return(expectedToken, nil)

		appTest := &app{
			userService: mockUserService,
		}

		token, err := appTest.Register(ctx, request)

		assert.Nil(t, err)
		assert.Equal(t, expectedToken, token)
		mockUserService.AssertExpectations(t)
	})

	t.Run("error - user already exists", func(t *testing.T) {
		t.Parallel()

		mockUserService := appMocks.NewMockUserService(t)

		ctx := t.Context()
		request := models.RegisterRequest{
			Login:    "existinguser",
			Password: "password123",
		}
		domainErr := entities.NewConflictError(
			errors.New("user already exists"),
			"user with this login already exists",
		)

		mockUserService.EXPECT().
			Register(ctx, forms.UserForm{
				Login:    request.Login,
				Password: request.Password,
			}).
			Return("", domainErr).
			Once()

		appTest := &app{
			userService: mockUserService,
		}

		token, err := appTest.Register(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, "", token)
		assert.Equal(t, entities.ConflictErrorType, err.ErrorType)
		mockUserService.AssertExpectations(t)
	})
}
