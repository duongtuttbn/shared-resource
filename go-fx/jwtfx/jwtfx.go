package jwtfx

import (
	"go.uber.org/fx"
)

// NewModule create a Module that register both JwtEncoder and JwtDecoder.
func NewModule() fx.Option {
	return fx.Module("jwt",
		fx.Provide(NewJwtEncoder, NewJwtDecoder, NewJwtService),
	)
}

// NewDecodeModule create a Module that register JwtDecoder.
// The JwtService created by this method does not support encoding.
func NewDecodeModule() fx.Option {
	return fx.Module("jwt.decoder",
		fx.Provide(NewJwtDecoder, NewJwtService),
	)
}
