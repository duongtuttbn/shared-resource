package ginfx

import (
	"time"
)

type ServerConfig struct {
	Port                 int           `json:"port" mapstructure:"port"`
	CorsAllowedOrigins   []string      `json:"cors_allowed_origins" mapstructure:"cors_allowed_origins"`
	CorsAllowedHeaders   []string      `json:"cors_allowed_headers" mapstructure:"cors_allowed_headers"`
	CorsAllowCredentials bool          `json:"cors_allow_credentials" mapstructure:"cors_allow_credentials"`
	ServerTimeout        time.Duration `json:"server_timeout" mapstructure:"server_timeout"`
	TrustedPlatform      string        `json:"trusted_platform" mapstructure:"trusted_platform"`
	DisableRequestID     bool          `json:"disable_request_id" mapstructure:"disable_request_id"`
}
