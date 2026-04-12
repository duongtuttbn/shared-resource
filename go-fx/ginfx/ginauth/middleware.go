package ginauth

import (
	"tla-backend/pkg/go-fx/ginfx"
	"tla-backend/pkg/go-fx/jwtfx"
	"tla-backend/pkg/go-kit/core"
)

func NewJwtMiddleware(jwtService *jwtfx.JwtService) ginfx.Middleware {
	return ginfx.MiddlewareHandler{
		HandlerFunc: core.UseAuth(jwtService),
	}
}
