package config

import (
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
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}
