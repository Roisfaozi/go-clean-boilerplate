package setup

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	Enforcer  usecase.IEnforcer
	Logger    *logrus.Logger
	Ctx       context.Context
	MySQLAddr string
	RedisAddr string
	S3URL     string
	S3Bucket  string
	Closers   []func()
}

func (env *TestEnvironment) StartWorker(processor interface {
	Start() error
	Shutdown()
}) {
	go func() {
		if err := processor.Start(); err != nil {
			env.Logger.Errorf("Failed to start worker in test: %v", err)
		}
	}()
	env.AddCloser(processor.Shutdown)
}

func (env *TestEnvironment) AddCloser(closer func()) {
	env.Closers = append(env.Closers, closer)
}

func (env *TestEnvironment) Cleanup() {
	for i := len(env.Closers) - 1; i >= 0; i-- {
		env.Closers[i]()
	}
}

func SetupRustFS(t *testing.T, ctx context.Context) (string, string) {
	bucket := "test-bucket"
	// Ensure we wait for both the port and a successful response from health check or similar
	req := testcontainers.ContainerRequest{
		Image:        "rustfs/rustfs:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"RUSTFS_ACCESS_KEY":     "rustfsadmin",
			"RUSTFS_SECRET_KEY":     "rustfsadmin",
			"RUSTFS_CONSOLE_ENABLE": "true",
			"RUSTFS_VOLUMES":        "/data",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9000/tcp"),
			wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
		).WithDeadline(60 * time.Second),
	}

	rustfsC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Skipping: Failed to start RustFS: %v", err)
		return "", ""
	}

	t.Cleanup(func() {
		_ = rustfsC.Terminate(ctx)
	})

	host, err := rustfsC.Host(ctx)
	require.NoError(t, err)

	p, err := rustfsC.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	s3URL := fmt.Sprintf("http://%s:%s", host, p.Port())

	return s3URL, bucket
}

func SetupIntegrationEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Create a cleaner test output
	logger.SetLevel(logrus.FatalLevel)

	initOnce.Do(func() {
		var err error
		// logger.Info("🐳 Starting Shared Integration Containers...") // Suppressed

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
			"redis:7.2-alpine",
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
		globalRDB = redis.NewClient(&redis.Options{
			Addr:            redisAddr,
			Protocol:        3,
			DisableIdentity: true, // Suppress maint_notifications handshake error
		})

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

func connectWithRetry(connStr string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysqlDriver.Open(connStr), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   logger.Default.LogMode(logger.Silent),
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

func SetupCasbin(t *testing.T, db *gorm.DB, logger *logrus.Logger) usecase.IEnforcer {
	// Ensure config path is correct relative to integration tests
	cfg := &config.AppConfig{
		Casbin: config.CasbinConfig{
			Enabled: true,
			Model:   "../../../internal/config/casbin_model.conf",
			Watcher: config.WatcherConfig{Enabled: false},
		},
	}
	enforcer, err := config.NewCasbinEnforcer(cfg, db, logger)
	require.NoError(t, err, "Failed to setup Casbin enforcer")
	return usecase.NewTransactionalEnforcer(enforcer, cfg.Casbin.Model)
}

func SetupRedisContainer(ctx context.Context) (*redisContainer.RedisContainer, string, error) {
	redisC, err := redisContainer.Run(ctx,
		"redis:7.2-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", err
	}

	p, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		return nil, "", err
	}

	return redisC, p.Port(), nil
}

func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
