package setup

import (
	"context"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	integrationSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TestServer struct {
	Server   *httptest.Server
	DB       *gorm.DB
	Redis    *redis.Client
	Enforcer usecase.IEnforcer
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

func SetupTusTestServer(t *testing.T) *TestServer {
	ctx := context.Background()
	env := integrationSetup.SetupIntegrationEnvironment(t)
	s3URL, s3Bucket := integrationSetup.SetupRustFS(t, ctx)

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
			Watcher: config.WatcherConfig{Enabled: false},
		},
		RateLimit: config.RateLimitConfig{Enabled: false},
		Storage: config.StorageConfig{
			Driver: "s3",
			S3: struct {
				Endpoint       string `mapstructure:"endpoint"`
				Region         string `mapstructure:"region"`
				Bucket         string `mapstructure:"bucket"`
				AccessKey      string `mapstructure:"access_key"`
				SecretKey      string `mapstructure:"secret_key"`
				UseSSL         bool   `mapstructure:"use_ssl"`
				ForcePathStyle bool   `mapstructure:"force_path_style"`
			}{
				Endpoint:       s3URL,
				Region:         "us-east-1",
				Bucket:         s3Bucket,
				AccessKey:      "rustfsadmin",
				SecretKey:      "rustfsadmin",
				UseSSL:         false,
				ForcePathStyle: true,
			},
		},
		Tus: config.TusConfig{
			BasePath: "/api/v1/upload/files/",
		},
	}

	// Create bucket using a temporary client
	awsCfg, _ := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("rustfsadmin", "rustfsadmin", "")),
	)
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3URL)
		o.UsePathStyle = true
	})

	// Wait a bit for RustFS
	time.Sleep(2 * time.Second)

	_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s3Bucket),
	})
	require.NoError(t, err)

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

func CreateAdminAndLogin(t *testing.T, server *TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AdminPass123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "user_admin"
		u.Email = "user_admin@test.com"
		u.Password = string(hash)
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AdminPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	return loginRes.Data.AccessToken
}
