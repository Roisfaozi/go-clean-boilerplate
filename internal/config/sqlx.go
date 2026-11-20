package config

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	defaultDSNCharset   = "utf8mb4"
	defaultDSNParseTime = "True"
	defaultDSNLocation  = "UTC"
)

func buildMySQLDSN(cfg MySqlConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		defaultDSNCharset, defaultDSNParseTime, defaultDSNLocation)
}

func openSQLXDB(cfg *AppConfig) (*sqlx.DB, error) {
	dsn := buildMySQLDSN(cfg.Mysql)
	if cfg.Telemetry.Enabled {
		sqlDB, err := otelsql.Open("mysql", dsn, otelsql.WithAttributes(semconv.DBSystemMySQL))
		if err != nil {
			return nil, err
		}
		return sqlx.NewDb(sqlDB, "mysql"), nil
	}
	return sqlx.Open("mysql", dsn)
}

// NewSQLXDatabase creates and connects a new *sqlx.DB connection pool.
func NewSQLXDatabase(cfg *AppConfig, log *logrus.Logger) *sqlx.DB {
	db, err := openSQLXDB(cfg)
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
