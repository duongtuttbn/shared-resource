package ginfx

import (
	"net/http"
	"slices"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"go.uber.org/fx"
)

type Middleware interface {
	Handle(c *gin.Context)
	Priority() int
}

// AsMiddleware register function into global middlewares.
func AsMiddleware(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Middleware)),
		fx.ResultTags(`group:"middlewares"`),
	)
}

type MiddlewareHandler struct {
	gin.HandlerFunc
	MiddlewarePriority int
}

func (m MiddlewareHandler) Handle(ctx *gin.Context) {
	m.HandlerFunc(ctx)
}

func (m MiddlewareHandler) Priority() int {
	return m.MiddlewarePriority
}

func newCORSMiddleware(env ServerConfig) gin.HandlerFunc {
	options := cors.Options{
		MaxAge:           86400 * 30,
		AllowCredentials: env.CorsAllowCredentials,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{
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
		},
	}

	if len(env.CorsAllowedHeaders) > 0 {
		options.AllowedHeaders = append(options.AllowedHeaders, env.CorsAllowedHeaders...)
	}

	if options.AllowCredentials && slices.Contains(env.CorsAllowedOrigins, "*") {
		// Allow all origins for * with credentials enabled.
		options.AllowOriginFunc = func(_ string) bool {
			return true
		}
	} else {
		options.AllowedOrigins = env.CorsAllowedOrigins
	}
	return cors.New(options)
}

func newTimeoutMiddleware(env ServerConfig) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(env.ServerTimeout),
		timeout.WithResponse(func(context *gin.Context) {
			WriteStatus(context, http.StatusRequestTimeout)
		}),
	)
}
