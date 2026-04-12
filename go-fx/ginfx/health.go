package ginfx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func NewHealthModule() fx.Option {
	return fx.Module("health",
		fx.Provide(
			AsModule(NewHealthHandler),
		),
	)
}

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Router(r gin.IRouter) {
	r.GET("/health", func(c *gin.Context) {
		WriteResponse(c, map[string]string{
			"status": "ok",
		})
	})
}
