package config

import (
	"fmt"
	"path"
	"runtime"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/sirupsen/logrus"
)

// TraceContextHook attaches RequestID from context to the log entry
type TraceContextHook struct{}

func (h *TraceContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TraceContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if reqID, ok := entry.Context.Value(constants.RequestIDKey).(string); ok {
			entry.Data["request_id"] = reqID
		}
		// Also trace UserID if available (e.g. from authenticated context)
		if userID, ok := entry.Context.Value(constants.UserIDKey).(string); ok {
			entry.Data["user_id"] = userID
		}
	}
	return nil
}

func NewLogrus(config *AppConfig) *logrus.Logger {
	logger := logrus.New()

	// Add Trace Hook
	logger.AddHook(&TraceContextHook{})

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
			ForceColors:     true,
			FullTimestamp:   true,
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