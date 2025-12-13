package config

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

// NewLogrus creates a new Logrus logger instance based on the application configuration.
func NewLogrus(config *AppConfig) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
		logger.Warnf("Invalid log level '%s'. Defaulting to 'info'.", config.Log.Level)
	} else {
		logger.SetLevel(level)
	}

	// Add report caller to see file and line number where log was called
	logger.SetReportCaller(true)

	// Use TextFormatter for development environment, JSONFormatter for others
	if config.Server.AppEnv != "production" {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				// Shorten file path to just filename:line
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				// Shorten file path
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	}

	return logger
}

// LogWithContext extracts the Request ID from context (if available) and returns a logger entry.
func LogWithContext(ctx context.Context, logger *logrus.Logger) *logrus.Entry {
	entry := logrus.NewEntry(logger)

	if reqID, ok := ctx.Value("request_id").(string); ok {
		entry = entry.WithField("request_id", reqID)
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		entry = entry.WithField("user_id", userID)
	}

	return entry
}

// LogError logs an error with context, stack trace, and additional message.
func LogError(ctx context.Context, logger *logrus.Logger, err error, message string) {
	entry := LogWithContext(ctx, logger)

	// Use %+v to print stack trace if the error was wrapped with pkg/errors.
	// If not wrapped, it prints the error string.
	entry.WithField("error_detail", fmt.Sprintf("%+v", err)).
		WithError(err).
		Error(message)
}
