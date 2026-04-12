package ginfx_test

import (
	"testing"
	"github.com/duongtuttbn/shared-resource/go-fx/ginfx"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const (
	global = "global"
	auth   = "auth"
)

func newGlobalV1() ginfx.RouteGroup {
	return ginfx.NewRouteGroup(global, "/v1")
}

func newRequireAuth() ginfx.RouteGroup {
	return ginfx.NewRouteGroup(auth, "", func(_ *gin.Context) {
		// require auth logic
	}).WithParent(global)
}

type todoHandler struct{}

func newTodoHandler() todoHandler {
	return todoHandler{}
}

func (t todoHandler) Group() ginfx.RouteGroupName {
	return auth
}

func (t todoHandler) Router(r gin.IRouter) {
	r.GET("todos", func(c *gin.Context) {
		ginfx.WriteResponse(c, "ok")
	})
}

func TestRouteGroup(t *testing.T) {
	app := fxtest.New(
		t,
		fx.Provide(
			func() ginfx.ServerConfig {
				return ginfx.ServerConfig{
					Port: 8080,
				}
			},
			ginfx.NewGinEngine,
			ginfx.AsGroup(newGlobalV1),
			ginfx.AsGroup(newRequireAuth),

			ginfx.AsModule(newTodoHandler),
		),
		fx.Invoke(func(engine *gin.Engine) {
			routes := engine.Routes()

			assert.Len(t, routes, 1)
			println(routes[0].Handler)
		}),
	)
	defer app.RequireStop()
	app.RequireStart()
}
