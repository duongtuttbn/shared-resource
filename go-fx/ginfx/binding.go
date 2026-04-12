package ginfx

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var ErrMissingRequestBody = errors.New("missing request body")

// ShouldBindUri is a convenience wrapper for Gin's ShouldBindUri.
func ShouldBindUri(c *gin.Context, obj any) error { // nolint:revive
	return c.ShouldBindUri(obj)
}

// ShouldBindQuery is a convenience wrapper for Gin's ShouldBindQuery.
func ShouldBindQuery(c *gin.Context, obj any) error {
	return c.ShouldBindQuery(obj)
}

// ShouldBind is a wrapper for gin's ShouldBind that provide additional feature.
func ShouldBind(c *gin.Context, obj any) error {
	return checkBindingError(c.ShouldBind(obj))
}

// ShouldBindJSON is a wrapper for gin's ShouldBindJSON that provide additional feature.
func ShouldBindJSON(c *gin.Context, obj any) error {
	return checkBindingError(c.ShouldBindJSON(obj))
}

func checkBindingError(err error) error {
	if errors.Is(err, io.EOF) {
		return ErrMissingRequestBody
	}

	return err
}
