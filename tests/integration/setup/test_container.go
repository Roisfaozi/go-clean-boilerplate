package setup

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
	mysqlDriver "gorm.io/driver/mysql"
)

// Singleton instances
var (
	mysqlC    *mysql.MySQLContainer
	redisC    *redisContainer.RedisContainer
	globalDB  *gorm.DB
	globalRDB *redis.Client
	mysqlAddr string
	redisAddr string
	initOnce  sync.Once
)

type TestEnvironment struct {
	DB        *gorm.DB
	Redis     *redis.Client
	Enforcer  *casbin.Enforcer
	Logger    *logrus.Logger
	Ctx       context.Context
	MySQLAddr string
	RedisAddr string
}

// SetupIntegrationEnvironment initializes the shared containers once using singleton pattern.
func SetupIntegrationEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	initOnce.Do(func() {
		var err error
		logger.Info("🐳 Starting Shared Integration Containers...")

		mysqlC, err = mysql.RunContainer(ctx,
			testcontainers.WithImage("mysql:8.0"),
			mysql.WithDatabase("test_db"),
			mysql.WithUsername("test"),
			mysql.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("port: 3306  MySQL Community Server").
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to start MySQL: %v", err))
		}

		redisC, err = redisContainer.RunContainer(ctx,
			testcontainers.WithImage("redis:7-alpine"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("Ready to accept connections").
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to start Redis: %v", err))
		}

		mysqlAddr, err = mysqlC.ConnectionString(ctx)
		if err != nil {
			panic(err)
		}
		globalDB, err = connectWithRetry(mysqlAddr, 5)
		if err != nil {
			panic(err)
		}

		redisAddr, err = redisC.Endpoint(ctx, "")
		if err != nil {
			panic(err)
		}
		globalRDB = redis.NewClient(&redis.Options{Addr: redisAddr})

		RunMigrations(nil, globalDB)
	})

	require.NotNil(t, globalDB, "Database should be initialized")
	require.NotNil(t, globalRDB, "Redis should be initialized")

	CleanupDatabase(t, globalDB)
	_ = globalRDB.FlushDB(ctx).Err()
	SeedTestData(t, globalDB)
	enforcer := SetupCasbin(t, globalDB, logger)

	return &TestEnvironment{
		DB:        globalDB,
		Redis:     globalRDB,
		Enforcer:  enforcer,
		Logger:    logger,
		Ctx:       ctx,
		MySQLAddr: mysqlAddr,
		RedisAddr: redisAddr,
	}
}

func (env *TestEnvironment) Cleanup() {
	// No-op for singleton containers
}

func connectWithRetry(connStr string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysqlDriver.Open(connStr), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   nil,
		})
		if err == nil {
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					return db, nil
				}
			}
		}
		time.Sleep(time.Second * 1)
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
}

func SetupCasbin(t *testing.T, db *gorm.DB, logger *logrus.Logger) *casbin.Enforcer {
	cfg := &config.AppConfig{
		Casbin: config.CasbinConfig{
			Enabled: true,
			Model:   "../../../internal/config/casbin_model.conf",
			Watcher: config.WatcherConfig{Enabled: false},
		},
	}
	enforcer, err := config.NewCasbinEnforcer(cfg, db, logger)
	require.NoError(t, err, "Failed to setup Casbin enforcer")
	return enforcer
}
