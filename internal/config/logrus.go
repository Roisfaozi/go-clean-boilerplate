package config

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"

	"github.com/sirupsen/logrus"
)

func NewLogrus(config *AppConfig) *logrus.Logger {

	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Log.Level)

	if err != nil {

		logger.SetLevel(logrus.InfoLevel)

		logger.Warnf("Invalid log level '%s'. Defaulting to 'info'.", config.Log.Level)

	} else {

		logger.SetLevel(level)

	}

	logger.SetReportCaller(true)

	if config.Server.AppEnv == "development" {

		logger.SetFormatter(&logrus.TextFormatter{

			ForceColors: true,

			FullTimestamp: true,

			TimestampFormat: "2006-01-02 15:04:05.000",

			CallerPrettyfier: func(f *runtime.Frame) (string, string) {

				filename := path.Base(f.File)

				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)

			},
		})

	} else {

		logger.SetFormatter(&logrus.JSONFormatter{

			TimestampFormat: "2006-01-02 15:04:05.000",

			CallerPrettyfier: func(f *runtime.Frame) (string, string) {

				filename := path.Base(f.File)

				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)

			},
		})

	}

	return logger

}

func LogWithContext(ctx context.Context, logger *logrus.Logger) *logrus.Entry {

	entry := logrus.NewEntry(logger)

	if reqID, ok := ctx.Value(constants.RequestIDKey).(string); ok {

		entry = entry.WithField("request_id", reqID)

	}

	if userID, ok := ctx.Value(constants.UserIDKey).(string); ok {

		entry = entry.WithField("user_id", userID)

	}

	return entry

}

func LogError(ctx context.Context, logger *logrus.Logger, err error, message string) {
	entry := LogWithContext(ctx, logger)

	entry.WithField("error_detail", fmt.Sprintf("%+v", err)).
		WithError(err).
		Error(message)
}
