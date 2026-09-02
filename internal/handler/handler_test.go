package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/handler/mocks"
	interfacesMocks "github.com/bazueva/gofermart/internal/interfaces/mocks"
	"github.com/bazueva/gofermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandler_LoginUser(t *testing.T) {
	t.Parallel()

	t.Run("success - login user", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		loginRequest := models.LoginRequest{
			Login:    "testuser",
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		expectedToken := "valid.jwt.token"

		mockApp.EXPECT().
			Login(req.Context(), loginRequest).
			Return(expectedToken, nil)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.LoginUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, response["token"])
		assert.Equal(t, "Bearer "+expectedToken, w.Header().Get("Authorization"))
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte("{invalid json}")))
		w := httptest.NewRecorder()

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.LoginUser(w, req)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid character 'i' looking for beginning of object key string", response["error"])
	})

	t.Run("error - app login failed", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		loginRequest := models.LoginRequest{
			Login:    "testuser",
			Password: "wrong_password",
		}
		body, _ := json.Marshal(loginRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		domainErr := entities.NewUnauthorizedError(
			errors.New("invalid credentials"),
			"invalid login or password",
		)

		mockApp.EXPECT().
			Login(req.Context(), loginRequest).
			Return("", domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.LoginUser(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("error - empty login", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		loginRequest := models.LoginRequest{
			Login:    "",
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		domainErr := entities.NewBadRequestError(
			errors.New("login is required"),
			"login cannot be empty",
		)

		mockApp.EXPECT().
			Login(req.Context(), loginRequest).
			Return("", domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.LoginUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("success - register user", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		registerRequest := models.RegisterRequest{
			Login:    "newuser",
			Password: "password123",
		}
		body, _ := json.Marshal(registerRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		expectedToken := "valid.jwt.token"

		mockApp.EXPECT().
			Register(req.Context(), registerRequest).
			Return(expectedToken, nil)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.RegisterUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, response["token"])
		assert.Equal(t, "Bearer "+expectedToken, w.Header().Get("Authorization"))
	})

	t.Run("error - invalid request body", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("{invalid json}")))
		w := httptest.NewRecorder()

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.RegisterUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error - user already exists", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		registerRequest := models.RegisterRequest{
			Login:    "existinguser",
			Password: "password123",
		}
		body, _ := json.Marshal(registerRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		domainErr := entities.NewConflictError(
			errors.New("user already exists"),
			"user with this login already exists",
		)

		mockApp.EXPECT().
			Register(req.Context(), registerRequest).
			Return("", domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.RegisterUser(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("error - empty login", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		registerRequest := models.RegisterRequest{
			Login:    "",
			Password: "password123",
		}
		body, _ := json.Marshal(registerRequest)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		domainErr := entities.NewBadRequestError(
			errors.New("login is required"),
			"login cannot be empty",
		)

		mockApp.EXPECT().
			Register(req.Context(), registerRequest).
			Return("", domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.RegisterUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_CreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("success - create order", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		orderID := "12345678903"
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderID)))
		w := httptest.NewRecorder()

		mockApp.EXPECT().
			CreateOrder(req.Context(), orderID).
			Return(nil)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.CreateOrder(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("error - empty order ID", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("")))
		w := httptest.NewRecorder()

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.CreateOrder(w, req)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Не указан номер заказа", response["error"])
	})

	t.Run("error - create order", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		orderID := "12345678903"
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte(orderID)))
		w := httptest.NewRecorder()

		domainErr := entities.NewOkError(
			errors.New("order already uploaded"),
			"номер заказа уже был загружен этим пользователем",
		)

		mockApp.EXPECT().
			CreateOrder(req.Context(), orderID).
			Return(domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.CreateOrder(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_UserOrdersList(t *testing.T) {
	t.Parallel()

	t.Run("success - get user orders", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders?page=1&perPage=20", nil)
		w := httptest.NewRecorder()

		createdAt := time.Now()
		expectedOrders := []entities.Order{
			{
				ID:        1,
				OrderID:   "12345678903",
				UserID:    123,
				Status:    entities.OrdersStatusNew,
				CreatedAt: new(createdAt),
			},
		}

		mockApp.EXPECT().
			UserOrdersList(req.Context(), int32(1), int32(20)).
			Return(expectedOrders, nil)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.UserOrdersList(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []models.Order
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "12345678903", response[0].Number)
		assert.Equal(t, "NEW", response[0].Status)
	})

	t.Run("success - empty orders returns NoContent", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders?page=1&perPage=20", nil)
		w := httptest.NewRecorder()

		mockApp.EXPECT().
			UserOrdersList(req.Context(), int32(1), int32(20)).
			Return([]entities.Order{}, nil)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.UserOrdersList(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("error - internal server error", func(t *testing.T) {
		mockApp := mocks.NewMockApp(t)
		logger := interfacesMocks.NewMockLogger(t)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders?page=1&perPage=20", nil)
		w := httptest.NewRecorder()

		domainErr := entities.NewInternalServerError(
			errors.New("database error"),
			"failed to get orders",
		)

		mockApp.EXPECT().
			UserOrdersList(req.Context(), int32(1), int32(20)).
			Return(nil, domainErr)

		h := &handler{
			app:    mockApp,
			logger: logger,
		}

		h.UserOrdersList(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
