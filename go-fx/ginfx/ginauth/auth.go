package ginauth

import (
	"go.uber.org/fx"
	"tla-backend/pkg/go-fx/ginfx"
	"tla-backend/pkg/go-fx/jwtfx"
)

// NewJwtModule register JwtEncoder and JwtDecoder.
// This method also register a middleware and a UserTokenService for resolving user, typically required for access control on server.
func NewJwtModule() fx.Option {
	return fx.Options(
		jwtfx.NewModule(),
		fx.Module("ginauth",
			fx.Provide(
				ginfx.AsMiddleware(NewJwtMiddleware),
			),
		),
	)
}

// NewJwtDecodeModule register JwtDecoder only.
// This method also register a middleware and a UserTokenService for resolving user, typically required for access control on server.
func NewJwtDecodeModule() fx.Option {
	return fx.Options(
		jwtfx.NewDecodeModule(),
		fx.Module("auth",
			fx.Provide(
				ginfx.AsMiddleware(NewJwtMiddleware),
			),
		),
	)
}
