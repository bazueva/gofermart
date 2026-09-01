package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"go.uber.org/zap"
)

type App interface {
	Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError)
	Login(ctx context.Context, request models.LoginRequest) (string, *entities.DomainError)
	CreateOrder(ctx context.Context, id string) *entities.DomainError
}

type handler struct {
	logger interfaces.Logger
	app    App
}

func NewHandler(logger interfaces.Logger, application App) *handler {
	return &handler{
		logger: logger,
		app:    application,
	}
}

func (h *handler) LoginUser(w http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			h.logger.Error("body close error", zap.Error(err))
		}
	}()

	var loginRequest models.LoginRequest

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&loginRequest)
	if err != nil {
		h.errorHandler(w, err, http.StatusBadRequest)

		return
	}

	tokenJWT, errDomain := h.app.Login(request.Context(), loginRequest)
	if errDomain != nil {
		h.errorHandler(w, errDomain, 0)

		return
	}

	// для тестов
	w.Header().Set("Authorization", "Bearer "+tokenJWT)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenJWT,
	})
}

func (h *handler) RegisterUser(w http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			h.logger.Error("body close error", zap.Error(err))
		}
	}()

	var regRequest models.RegisterRequest

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&regRequest)
	if err != nil {
		h.errorHandler(w, err, http.StatusBadRequest)

		return
	}

	tokenJWT, errDomain := h.app.Register(request.Context(), regRequest)
	if errDomain != nil {
		h.errorHandler(w, errDomain, 0)

		return
	}

	// для тестов
	w.Header().Set("Authorization", "Bearer "+tokenJWT)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenJWT,
	})
}

func (h *handler) errorHandler(writer http.ResponseWriter, err error, statusCode int) {
	if domainError, ok := errors.AsType[*entities.DomainError](err); ok {
		switch domainError.ErrorType {
		case entities.ConflictErrorType:
			statusCode = http.StatusConflict
		case entities.BadRequestErrorType:
			statusCode = http.StatusBadRequest
		case entities.UnauthorizedErrorType:
			statusCode = http.StatusUnauthorized
		case entities.UnprocessableEntityErrorType:
			statusCode = http.StatusUnprocessableEntity
		case entities.OkEntityErrorType:
			statusCode = http.StatusOK

		default:
			statusCode = http.StatusInternalServerError
		}
	}

	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(map[string]string{
		"error": err.Error(),
	})
}

func (h *handler) CreateOrder(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.errorHandler(writer, err, http.StatusBadRequest)

		return
	}

	defer func() {
		_ = request.Body.Close()
	}()

	orderID := string(body)
	if orderID == "" {
		h.errorHandler(writer, errors.New("Не указан номер заказа"), http.StatusBadRequest)

		return
	}

	errDomain := h.app.CreateOrder(request.Context(), orderID)
	if errDomain != nil {
		h.errorHandler(writer, errDomain, 0)

		return
	}

	writer.WriteHeader(http.StatusAccepted)
}
