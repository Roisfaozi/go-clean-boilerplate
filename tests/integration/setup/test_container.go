package setup

import (
	"context"
	"fmt"
	"os/exec"
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
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

func SetupIntegrationEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	initOnce.Do(func() {
		var err error
		logger.Info("🐳 Starting Shared Integration Containers...")

		if !IsDockerAvailable() {
			_ = fmt.Errorf("docker not available")
			return
		}

		mysqlC, err = mysql.Run(ctx,
			"mysql:lts",
			mysql.WithDatabase("test_db"),
			mysql.WithUsername("test"),
			mysql.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("port: 3306  MySQL Community Server").
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			t.Skipf("Skipping integration tests: Docker environment not available or failed to start: %v", err)
			return
		}

		redisC, err = redisContainer.Run(ctx,
			"redis:8.4-alpine",
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

		mysqlAddr = mysqlAddr + "?parseTime=true"
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

	if globalDB == nil {
		t.Skip("Skipping integration tests: Database not initialized (likely due to missing Docker)")
		return nil
	}

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

func SetupRedisContainer(ctx context.Context) (*redisContainer.RedisContainer, string, error) {
	redisC, err := redisContainer.Run(ctx,
		"redis:8.4-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", err
	}

	// Remove the protocol if present (e.g., "redis://") as client usually expects host:port
	// Actually Endpoint returns host:port, so no need to strip protocol usually unless mapped port retrieval is weird.
	// But redis.NewClient Options.Addr expects "host:port". redisC.Endpoint returns exactly that.

	// Extract port is tricky from Endpoint directly if we want just port, but we need host:port for client.
	// The function signature in test expects (container, port), but actually it uses it as Addr.
	// Let's return the full address as "port" string for simplicity in the test usage which does fmt.Sprintf("localhost:%s", port) - WAIT.
	// If test does `fmt.Sprintf("localhost:%s", redisPort)`, it expects ONLY port.
	
	// Let's get the mapped port.
	p, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		return nil, "", err
	}
	
	return redisC, p.Port(), nil
}

func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
