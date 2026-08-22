package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/interfaces"
	"github.com/bazueva/gofermart/internal/models"
	"go.uber.org/zap"
)

type App interface {
	Register(ctx context.Context, request models.RegisterRequest) (string, *entities.DomainError)
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

func (h *handler) RegisterUser(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
		default:
			statusCode = http.StatusInternalServerError
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(map[string]string{
		"error": err.Error(),
	})
}
