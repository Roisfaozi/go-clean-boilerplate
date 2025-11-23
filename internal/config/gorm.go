package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(config *AppConfig, log *logrus.Logger) *gorm.DB {
	username := config.Mysql.User
	password := config.Mysql.Password
	host := config.Mysql.Host
	port := config.Mysql.Port
	database := config.Mysql.DBName
	idleConnection := config.Mysql.IdleConnection
	maxConnection := config.Mysql.MaxConnection
	maxLifeTimeConnection := config.Mysql.MaxLifeTimeConnection

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, database)
	log.Errorf("DSN: %s", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Info,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	connection, err := db.DB()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	connection.SetMaxIdleConns(idleConnection)
	connection.SetMaxOpenConns(maxConnection)
	connection.SetConnMaxLifetime(time.Second * time.Duration(maxLifeTimeConnection))

	return db
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	// Skip if no logger is set
	if l.Logger == nil {
		return
	}

	// Format the message with arguments
	msg := fmt.Sprintf(message, args...)

	// Log with appropriate level
	if len(args) > 0 {
		// This is an error or slow query
		l.Logger.Debugf("GORM: %s", msg)
	} else {
		// This is a regular SQL query
		l.Logger.Debugf("GORM: %s", msg)
	}

}
