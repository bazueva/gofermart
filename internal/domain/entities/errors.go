package entities

import (
	"github.com/samber/lo"
)

type DomainError struct {
	ErrorType ErrorType
	SourceErr error
	Text      string
}

type ErrorType int

const (
	InternalServerErrorType ErrorType = iota
	ConflictErrorType
	BadRequestErrorType
	UnauthorizedErrorType
	UnprocessableEntityErrorType
	OkEntityErrorType
	NoContentErrorType
	RetriableErrorType
	PaymentRequiredErrorType
	ToManyRequestErrorType
)

func (e *DomainError) Error() string {
	return e.Text
}

func NewInternalServerError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: InternalServerErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "Internal Server Error"),
	}
}

func NewUnauthorizedError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: UnauthorizedErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "пользователь не аутентифицирован"),
	}
}

func NewOkError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: OkEntityErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewUnprocessableEntity(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: UnprocessableEntityErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewConflictError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: ConflictErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewBadRequestError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: BadRequestErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewNoContentError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: NoContentErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewRetriableError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: RetriableErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewPaymentRequiredError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: PaymentRequiredErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewTooManyRequestError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: ToManyRequestErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "Слишком много запросов"),
	}
}
