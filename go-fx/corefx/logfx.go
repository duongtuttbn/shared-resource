package corefx

import (
	"context"
	"github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"strings"
	"time"
	"tla-backend/pkg/go-kit/log"
)

// newLogger create a logger instance.
func newLogger(p LoggerParams) (*logrus.Logger, error) {
	level := parseLogLevel(p.Config.Log.Level)
	if p.Config.Profile == ProfileDebug {
		level = logrus.DebugLevel
	}
	logger := logrus.New()

	logger.SetLevel(level)

	switch p.Config.Log.Format {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: false})
	}
	if p.Config.SentryDsn == "" {
		return logger, nil
	}
	// Setup sentry.
	environment := ProfileDevelopment
	if p.Config.Profile == ProfileProduction {
		environment = ProfileProduction
	}
	release := "unknown" // TODO: get release from git
	// Send only ERROR and higher level logs to Sentry
	sentryLevels := []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
	// Initialize Sentry
	sentryHook, err := sentrylogrus.NewEventHook(sentryLevels, sentry.ClientOptions{
		Dsn:           p.Config.SentryDsn,
		EnableTracing: false,
		Environment:   environment,
		Release:       release,
	})
	if err != nil {
		return nil, err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			sentryHook.Flush(5 * time.Second)
			return nil
		},
	})

	logger.AddHook(sentryHook)
	return logger, nil
}

func parseLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

type LoggerParams struct {
	fx.In
	Config    CoreConfig
	Lifecycle fx.Lifecycle
}

// NewGlobalLogger create a logger instance and register it globally.
func NewGlobalLogger(p LoggerParams) (log.Logger, error) {
	logger, err := newLogger(p)
	if err != nil {
		return nil, err
	}
	log.SetDefaultLogger(logger)
	return log.Root(), nil
}
