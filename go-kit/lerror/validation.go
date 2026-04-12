package lerror

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError a wrapper for validator.ValidationErrors that provide additional context.
type ValidationError struct {
	// BaseErrorMessages is the base errors map that will be used to build the response. You can use this field if you want to return
	// a validation error without having to create a list of Violations first.
	// Field in the errors will be overwritten by the validation errors if they have the same key.
	BaseErrorMessages map[string][]string
	// Violations is the list of validation errors.
	Violations validator.ValidationErrors
	// Message additional message describing the error.
	Message string
	// Prefix adds prefix to all error fields in the response.
	Prefix string
}

func (ve ValidationError) Error() string {
	return ve.Violations.Error()
}

func (ve ValidationError) Unwrap() error {
	return ve.Violations
}

// NewFieldViolation return a ValidationError with base error messages described by field and errors.
func NewFieldViolation(field string, errors ...string) error {
	return ValidationError{
		BaseErrorMessages: map[string][]string{
			field: errors,
		},
	}
}

// WrapPViolation return a ValidationError from validator.ValidationErrors, with optional prefixes.
func WrapPViolation(err validator.ValidationErrors, prefixes ...string) error {
	prefix := ""
	if len(prefixes) > 0 {
		if len(prefixes) > 1 {
			prefix = strings.Join(prefixes, ".")
		} else {
			prefix = prefixes[0]
		}
	}
	return ValidationError{
		Violations: err,
		Prefix:     prefix,
	}
}
