package ginfx

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"tla-backend/pkg/go-kit/core"
	"tla-backend/pkg/go-kit/lerror"
	"tla-backend/pkg/go-kit/log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

type Response struct {
	RequestID string              `json:"request_id,omitempty"`
	Code      int                 `json:"code,omitempty"`
	Message   string              `json:"message,omitempty"`
	Data      interface{}         `json:"data,omitempty"`
	Errors    map[string][]string `json:"errors,omitempty"`
}

// ResultHandler convert a function that return result/error into a route function.
func ResultHandler[T any](fn func(context.Context) (T, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		result, err := fn(c)
		if err != nil {
			WriteError(c, err)
			return
		}
		WriteResponse(c, result)
	}
}

// ResultHandlerGin convert a function that return result/error into a route function.
// This variant accepts *gin.Context for backward compatibility.
func ResultHandlerGin[T any](fn func(c *gin.Context) (T, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		result, err := fn(c)
		if err != nil {
			WriteError(c, err)
			return
		}
		WriteResponse(c, result)
	}
}

var errorHandlers = make([]WriteErrorHandler, 0)

// WriteErrorHandler write matched error to gin context.
// return true if matched, otherwise return false.
type WriteErrorHandler func(c *gin.Context, err error) bool

// ErrorMatcher match error, return matched error or nil if not matched.
type (
	ErrorMatcher        func(err error) error
	ErrorResponseWriter func(c *gin.Context, err error)
)

// RegisterWriteErrorHandler register a WriteErrorHandler to global error handler.
func RegisterWriteErrorHandler(handlers ...WriteErrorHandler) {
	errorHandlers = append(errorHandlers, handlers...)
}

// RegisterWriteErrorCode associate error with status code.
// the WriteError will return the associated error with this status code and default message.
func RegisterWriteErrorCode(matcher ErrorMatcher, status int, errorMessage ...string) {
	msg := ""
	if len(errorMessage) > 0 {
		msg = errorMessage[0]
	} else {
		msg = http.StatusText(status)
	}
	RegisterWriteErrorHandler(handleError(matcher, func(c *gin.Context, _ error) {
		WriteMessage(c, msg, status)
	}))
}

// RegisterWriteErrorResponse associate error with specified ErrorResponseWriter.
func RegisterWriteErrorResponse(matcher ErrorMatcher, writer ErrorResponseWriter) {
	RegisterWriteErrorHandler(handleError(matcher, writer))
}

func handleError(match ErrorMatcher, write ErrorResponseWriter) WriteErrorHandler {
	return func(c *gin.Context, err error) bool {
		matchedErr := match(err)
		if matchedErr == nil {
			return false
		}
		write(c, matchedErr)
		return true
	}
}

// ValidateStruct validate a struct.
func ValidateStruct(v any) error {
	return binding.Validator.ValidateStruct(v)
}

// RegisterValidation a simple wrapper for validator.RegisterValidation.
func RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation(tag, fn, callValidationEvenIfNull...)
		if err != nil {
			panic(err)
		}
	}
}

// RegisterValidationCtx a simple wrapper for validator.RegisterValidationCtx.
func RegisterValidationCtx(tag string, fn validator.FuncCtx, callValidationEvenIfNull ...bool) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidationCtx(tag, fn, callValidationEvenIfNull...)
		if err != nil {
			panic(err)
		}
	}
}

// MatchErrorIs create a matcher that match error using errors.Is.
func MatchErrorIs(target error) ErrorMatcher {
	return func(err error) error {
		if errors.Is(err, target) {
			return err
		}
		return nil
	}
}

// MatchErrorAs create a matcher that match error using errors.As.
func MatchErrorAs(target any) ErrorMatcher {
	return func(err error) error {
		t := target
		if errors.As(err, &t) {
			return t.(error)
		}
		return nil
	}
}

// WriteError write error as JSON.
func WriteError(c *gin.Context, err error, status ...int) {
	for i := range errorHandlers {
		handled := errorHandlers[i]
		if handled(c, err) {
			return
		}
	}

	statusCode := http.StatusInternalServerError
	message := ""
	var serr *lerror.XError
	if errors.As(err, &serr) {
		statusCode = serr.Status
		message = serr.Message
	}

	// WriteErrorBinding aren't registered as a WriteErrorHandler to allow user to override it behavior.
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		ve := lerror.ValidationError{
			Violations: validationErrors,
			Message:    message,
		}
		errors.As(err, &ve)
		// For validation error, always return 422.
		// Message can still be customized.
		writeErrorBinding(c, ve, http.StatusUnprocessableEntity)
		return
	}

	// Set default status only if passed in error is not serror.Error
	// which makes serror.Error status and message higher priority.
	if serr == nil {
		if len(status) > 0 {
			statusCode = status[0]
		}
		message = http.StatusText(statusCode)
	}

	// Logging for case 500.
	LogInternalServerError(c, err, statusCode)
	WriteMessage(c, message, statusCode)
}

// LogInternalServerError log the error if status is 500, otherwise log as debug.
func LogInternalServerError(c *gin.Context, err error, status int) {
	requestID := lo.CoalesceOrEmpty(core.GetRequestID(c), "[DISABLED]")
	if status == http.StatusInternalServerError {
		log.Errorf("Request ID: %s, error: %+v", requestID, err)
	} else {
		log.Debugf("Request ID: %s, error: %+v", requestID, err)
	}
}

// WriteErrorBinding write validator.ValidationErrors error with standardized format.
func WriteErrorBinding[T lerror.ValidationError | validator.ValidationErrors](c *gin.Context, err T, status ...int) {
	statusCode := http.StatusUnprocessableEntity
	if len(status) > 0 {
		statusCode = status[0]
	}
	if e, ok := any(err).(validator.ValidationErrors); ok {
		writeErrorBinding(c, lerror.ValidationError{Violations: e}, statusCode)
		return
	}
	writeErrorBinding(c, any(err).(lerror.ValidationError), statusCode)
}

// writeErrorBinding write validator.ValidationErrors error with standardized format.
// Allow specifying a message.
func writeErrorBinding(c *gin.Context, e lerror.ValidationError, statusCode int) {
	fieldErrors := []validator.FieldError(e.Violations)
	body := e.BaseErrorMessages
	if body == nil {
		body = make(map[string][]string, len(fieldErrors))
	}
	if e.Prefix != "" {
		e.Prefix += "."
	}
	for _, err := range fieldErrors {
		f := err.Namespace()
		// Remove the first item of Namespace (as it is the struct name)
		if i := strings.IndexRune(f, '.'); i >= 0 && i < len(f)-1 {
			f = f[i+1:]
		}
		f = e.Prefix + f
		if _, ok := body[f]; !ok {
			body[f] = make([]string, 0, 1)
		}
		body[f] = append(body[f], err.Tag())
	}

	c.JSON(statusCode, Response{
		RequestID: core.GetRequestID(c),
		Message:   e.Message,
		Code:      statusCode,
		Errors:    body,
	})
}

func WriteResponse(c *gin.Context, data interface{}, status ...int) {
	statusCode := http.StatusOK
	if len(status) > 0 {
		statusCode = status[0]
	}
	c.JSON(statusCode, Response{
		RequestID: core.GetRequestID(c),
		Code:      statusCode,
		Data:      data,
	})
}

// WriteResult write the result if there is no error.
func WriteResult[T any](c *gin.Context, result T, err error, status ...int) {
	if err != nil {
		WriteError(c, err)
		return
	}
	WriteResponse(c, result, status...)
}

// WriteResult0 write the status if there is no error.
func WriteResult0(c *gin.Context, err error, status ...int) {
	if err != nil {
		WriteError(c, err)
		return
	}
	statusCode := http.StatusOK
	if len(status) > 0 {
		statusCode = status[0]
	}
	WriteStatus(c, statusCode)
}

// WriteStatus write status as JSON with the message is the status text.
func WriteStatus(c *gin.Context, status int) {
	WriteMessage(c, http.StatusText(status), status)
}

// WriteMessage write a message as JSON.
func WriteMessage(c *gin.Context, message string, status ...int) {
	statusCode := http.StatusOK
	if len(status) > 0 {
		statusCode = status[0]
	}
	c.JSON(statusCode, Response{
		RequestID: core.GetRequestID(c),
		Code:      statusCode,
		Message:   message,
	})
}
