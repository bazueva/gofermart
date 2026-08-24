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
