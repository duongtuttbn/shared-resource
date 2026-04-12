package corefx

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"tla-backend/pkg/go-kit/log"
)

const (
	ProfileProduction  = "production"
	ProfileDevelopment = "development"
	ProfileDebug       = "debug"
)

type CoreConfig struct {
	Profile   string     `json:"profile" mapstructure:"profile"`
	Log       log.Config `json:"log" mapstructure:"log"`
	SentryDsn string     `json:"sentry_dsn" mapstructure:"sentry_dsn"`
}

// NewModule register core module and enable health check support.
func NewModule() fx.Option {
	return fx.Options(
		fx.WithLogger(func() fxevent.Logger {
			return &LogrusLogger{Logger: logrus.New()}
		}),
		fx.Module("corefx",
			fx.Provide(NewGlobalLogger),
			fx.Invoke(func(_ log.Logger) {
				// force initialization of logger, which also initialize config.
			}),
		),
	)
}

// As register already registered type T under multiple interfaces.
// Useful if you need a single required object to provide multiple required types.
// This method allows you to inject the original object, and all type it registered by this function.
func As[T any](types ...any) any {
	annotations := make([]fx.Annotation, 0, len(types))
	for i := range types {
		annotations = append(annotations, fx.As(types[i]))
	}

	return fx.Annotate(
		func(t T) T { return t },
		annotations...,
	)
}

// From create a function that accepts and return self.
// This method can be used with other As... methods of multiple fx packages when you want to keep both the original type and annotated type
// after annotated.
// For example, fx.Provide(newMyService, AsInterface(From[*myService])).
func From[T any]() any {
	return func(t T) T { return t }
}
