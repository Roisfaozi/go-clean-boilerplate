package config

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultDSNCharset   = "utf8mb4"
	defaultDSNParseTime = "True"
	defaultDSNLocation  = "UTC"
)

func NewDatabase(cfg *AppConfig, log *logrus.Logger) (*gorm.DB, *sqlx.DB) {
	sqlxDB := NewSQLXDatabase(cfg, log)
	gormDB := NewDatabaseFromSQLX(cfg, sqlxDB, log)
	return gormDB, sqlxDB
}

func NewDatabaseFromSQLX(cfg *AppConfig, sqlxDB *sqlx.DB, log *logrus.Logger) *gorm.DB {
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

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlxDB.DB,
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to initialize GORM with existing connection: %v", err)
	}

	if cfg.Telemetry.Enabled {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			log.Errorf("Failed to instrument GORM with OTEL: %v", err)
		}
	}

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
