package ginauth

import (
	"github.com/duongtuttbn/shared-resource/go-fx/ginfx"
	"github.com/duongtuttbn/shared-resource/go-fx/jwtfx"
	"github.com/duongtuttbn/shared-resource/go-kit/core"
)

func NewJwtMiddleware(jwtService *jwtfx.JwtService) ginfx.Middleware {
	return ginfx.MiddlewareHandler{
		HandlerFunc: core.UseAuth(jwtService),
	}
}
