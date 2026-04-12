package core

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

func UseCORS(allowOrigins []string, additionalHeaders ...string) gin.HandlerFunc {
	allowedHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-AccessToken",
		"Authorization",
		"Accept",
		"Origin",
		"Cache-Control",
		"X-Requested-With",
		"X-Recaptcha-Response",
		"X-Device-ID",
		"X-Platform",
		"X-Project-ID",
	}

	allowedHeaders = append(allowedHeaders, additionalHeaders...)

	return cors.New(cors.Options{
		MaxAge:           86400,
		AllowCredentials: true,
		AllowedOrigins:   allowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   allowedHeaders,
	})
}
