package config

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// NewCasbinEnforcer initializes and returns a Casbin enforcer based on the application configuration.
// It returns nil if Casbin is disabled in the configuration.
func NewCasbinEnforcer(cfg *AppConfig, db *gorm.DB, log *logrus.Logger) (*casbin.Enforcer, error) {
	if !cfg.Casbin.Enabled {
		log.Info("Casbin is disabled.")
		return nil, nil
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin gorm adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(cfg.Casbin.Model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Initialize Watcher if enabled
	if cfg.Casbin.Watcher.Enabled {
		watcher, err := rediswatcher.NewWatcher(cfg.Redis.Addr, rediswatcher.WatcherOptions{
			Channel: cfg.Casbin.Watcher.Channel,
			Options: redis.Options{
				Password: cfg.Redis.Password,
				DB:       cfg.Redis.DB,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create casbin redis watcher: %w", err)
		}
		if err := enforcer.SetWatcher(watcher); err != nil {
			return nil, fmt.Errorf("failed to set casbin watcher: %w", err)
		}
		log.Info("Casbin Redis watcher initialized.")
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policy: %w", err)
	}
	log.Info("Casbin enforcer initialized and policies loaded.")

	return enforcer, nil
}
