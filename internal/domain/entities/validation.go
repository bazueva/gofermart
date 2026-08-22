package entities

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidateError struct {
	Field string
	Error string
}

func ConvertValidatorErrors(err error) []ValidateError {
	if err == nil {
		return nil
	}

	resultErrors := make([]ValidateError, 0, 10)

	if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, e := range validateErrs {
			errorText := ""
			switch e.Tag() {
			case "required":
				errorText = "Поле обязательно для заполнения"
			case "min":
				errorText = fmt.Sprintf("Поле должно содержать не менее %s символов", e.Param())
			case "max":
				errorText = fmt.Sprintf("Поле должно содержать не более %s символов", e.Param())
			}

			resultErrors = append(resultErrors, ValidateError{
				Field: strings.ToLower(e.StructField()),
				Error: errorText,
			})
		}
	}

	return resultErrors
}
