package config

import (
	"context"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

// CoreFxModule provides application foundation singletons: config, logger, DB, redis, validator, transaction manager.
var CoreFxModule = fx.Module("core",
	fx.Provide(
		NewLogrus,
		NewValidator,
		func(cfg *AppConfig, log *logrus.Logger) *sqlx.DB {
			return NewSQLXDatabase(cfg, log)
		},
		func(cfg *AppConfig, log *logrus.Logger) *redis.Client {
			return NewRedisConfig(cfg, log)
		},
		func(db *sqlx.DB, log *logrus.Logger) tx.WithTransactionManager {
			return tx.NewSQLXTransactionManager(db, log)
		},
		func(cfg *AppConfig) (*Application, error) {
			return NewApplication(cfg)
		},
	),
	fx.Invoke(RegisterAppLifecycle),
)

// RegisterAppLifecycle registers lifecycle hooks for graceful startup and shutdown of Application.
func RegisterAppLifecycle(lc fx.Lifecycle, app *Application, log *logrus.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Infof("Server starting on %s", app.Server.Addr)
				if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Errorf("Server error: %v", err)
				}
			}()

			go func() {
				log.Info("Starting Scheduler...")
				if err := app.Scheduler.Start(); err != nil {
					log.Errorf("Failed to start scheduler: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down application...")
			return app.Shutdown(ctx)
		},
	})
}
