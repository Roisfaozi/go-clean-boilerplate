package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *AppConfig, log *logrus.Logger) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.User, cfg.Mysql.Password, cfg.Mysql.Host, cfg.Mysql.Port, cfg.Mysql.DBName)

	newLogger := logger.New(
		&logrusWriter{Logger: log},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Instrument GORM with OpenTelemetry
	if cfg.Telemetry.Enabled {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			log.Errorf("Failed to instrument GORM with OTEL: %v", err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Mysql.IdleConnection)
	sqlDB.SetMaxOpenConns(cfg.Mysql.MaxConnection)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Mysql.MaxLifeTimeConnection) * time.Second)

	return db
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	if l.Logger == nil {
		return
	}

	msg := fmt.Sprintf(message, args...)

	if len(args) > 0 {
		l.Logger.Debugf("GORM: %s", msg)
	} else {
		l.Logger.Debugf("GORM: %s", msg)
	}

}
