package core

import (
	"net/http"
	"tla-backend/pkg/go-kit/lerror"
	"tla-backend/pkg/go-kit/log"

	"github.com/gin-gonic/gin"
)

var ErrorHandler = func(c *gin.Context, err error, statusCode ...int) {
	requestID := GetRequestID(c)
	if requestID == "" {
		log.Warn("RequestID not found in context. Use middleware.UseRequestID to add request RequestID automatically")
	}

	log.Errorf("Request Id: %s %+v", requestID, err)

	response := Response{
		RequestID: requestID,
	}

	if !lerror.IsLError(err) {
		response.Code = lerror.InternalServer.ToInt()
		responseCode := http.StatusInternalServerError
		if len(statusCode) > 0 {
			responseCode = statusCode[0]
		}
		c.JSON(responseCode, response)
		return
	}

	resError := lerror.Unwrap(err)
	response.Code = resError.Code
	response.Message = resError.Message

	c.JSON(resError.Status, response)
}

var ResponseHandler = func(c *gin.Context, data interface{}, status ...int) {
	statusCode := http.StatusOK
	if len(status) > 0 {
		statusCode = status[0]
	}
	c.JSON(statusCode, Response{
		Code:    2000,
		Message: "Success",
		Data:    data,
	})
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type Response struct {
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func (h *Handler) Error(c *gin.Context, err error, statusCode ...int) {
	ErrorHandler(c, err, statusCode...)
}

func (h *Handler) Response(c *gin.Context, data interface{}, status ...int) {
	ResponseHandler(c, data, status...)
}

func (h *Handler) GetClientIP(c *gin.Context) string {
	return GetClientIP(c.Request)
}
