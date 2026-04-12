package jwtfx

import "time"

type JwtConfig struct {
	JwtDefaultExpired    time.Duration `json:"jwt_default_expired" mapstructure:"jwt_default_expired"`
	JwtIssuer            string        `json:"jwt_issuer" mapstructure:"jwt_issuer"`
	JwtKey               string        `json:"jwt_key" mapstructure:"jwt_key"`
	JwtJwkPrivateKeyJSON string        `json:"jwt_jwk_private_key_json" mapstructure:"jwt_jwk_private_key_json"`
	JwtJwksJSON          string        `json:"jwt_jwks_json" mapstructure:"jwt_jwks_json"`
	JwtJwksURL           string        `json:"jwt_jwks_url" mapstructure:"jwt_jwks_url"`
}
