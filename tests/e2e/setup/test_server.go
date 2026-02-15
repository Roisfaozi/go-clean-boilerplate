package setup

import (
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	integrationSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type TestServer struct {
	Server   *httptest.Server
	DB       *gorm.DB
	Redis    *redis.Client
	Enforcer *casbin.Enforcer
	BaseURL  string
	Client   *TestClient
}

func SetupTestServer(t *testing.T) *TestServer {
	env := integrationSetup.SetupIntegrationEnvironment(t)

	dsn := env.MySQLAddr
	parts := strings.Split(dsn, "@tcp(")
	hostPortAndDB := strings.Split(parts[1], ")/")
	hostPort := strings.Split(hostPortAndDB[0], ":")
	host := hostPort[0]
	port, _ := strconv.Atoi(hostPort[1])

	cfg := &config.AppConfig{
		Server: config.ServerConfig{
			Port:         0,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			AppName:      "test-app",
			AppEnv:       "test",
		},
		Mysql: config.MySqlConfig{
			Host:                  host,
			Port:                  port,
			User:                  "test",
			Password:              "test",
			DBName:                "test_db",
			IdleConnection:        10,
			MaxConnection:         100,
			MaxLifeTimeConnection: 3600,
		},
		Redis: config.RedisConfig{
			Addr:     env.RedisAddr,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		JWT: config.JWTConfig{
			AccessTokenSecret:    "test-access-secret-32-chars-long-min-length",
			RefreshTokenSecret:   "test-refresh-secret-32-chars-long-min-length",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 24 * time.Hour,
		},
		Security: config.SecurityConfig{
			MaxLoginAttempts: 5,
			LockoutDuration:  30 * time.Minute,
		},
		Casbin: config.CasbinConfig{
			Enabled: true,
			Model:   "../../../internal/config/casbin_model.conf",
			Watcher: config.WatcherConfig{
				Enabled: false,
				Channel: "/casbin",
			},
		},
		RateLimit: config.RateLimitConfig{
			Enabled: false,
		},
		Storage: config.StorageConfig{
			Driver: "local",
			Local: struct {
				RootPath string `mapstructure:"root_path"`
				BaseURL  string `mapstructure:"base_url"`
			}{
				RootPath: "./test_uploads",
				BaseURL:  "http://localhost/uploads",
			},
		},
	}

	app, err := config.NewApplication(cfg)
	require.NoError(t, err)

	server := httptest.NewServer(app.Server.Handler)
	client := NewTestClient(server.URL)

	return &TestServer{
		Server:   server,
		DB:       env.DB,
		Redis:    env.Redis,
		Enforcer: app.Enforcer,
		BaseURL:  server.URL,
		Client:   client,
	}
}

func (s *TestServer) Cleanup() {
	if s.Server != nil {
		s.Server.Close()
	}
}
