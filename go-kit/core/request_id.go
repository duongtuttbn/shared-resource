package core

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "request_id"

func UseRequestID(c *gin.Context) {
	requestID := time.Now().UnixMilli()
	c.Set(requestIDContextKey, strconv.FormatInt(requestID, 10))
}

func UseRequestIDWithGenerator(gen func() string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(requestIDContextKey, gen())
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(requestIDContextKey)
}
