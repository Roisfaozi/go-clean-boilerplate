package config

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// NewSQLXDatabase creates and connects a new *sqlx.DB connection pool.
func NewSQLXDatabase(cfg *AppConfig, log *logrus.Logger) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=%s",
		cfg.Mysql.User, cfg.Mysql.Password, cfg.Mysql.Host, cfg.Mysql.Port, cfg.Mysql.DBName,
		defaultDSNCharset, defaultDSNParseTime, defaultDSNLocation)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	db.SetMaxIdleConns(cfg.Mysql.IdleConnection)
	db.SetMaxOpenConns(cfg.Mysql.MaxConnection)
	db.SetConnMaxLifetime(time.Duration(cfg.Mysql.MaxLifeTimeConnection) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}
