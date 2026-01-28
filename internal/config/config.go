package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Server         ServerConfig         `mapstructure:"server"`
	Mysql          MySqlConfig          `mapstructure:"mysql"`
	Redis          RedisConfig          `mapstructure:"redis"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	Security       SecurityConfig       `mapstructure:"security"`
	Log            LoggerConfig         `mapstructure:"log"`
	WebSocket      WebSocketConfig      `mapstructure:"websocket"`
	Casbin         CasbinConfig         `mapstructure:"casbin"`
	CORS           CORSConfig           `mapstructure:"cors"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
	RateLimit      RateLimitConfig      `mapstructure:"rate_limit"`
	SMTP           SMTPConfig           `mapstructure:"smtp"`
	Storage        StorageConfig        `mapstructure:"storage"`
	Metrics        struct {
		Enabled     bool   `env:"METRICS_ENABLED" envDefault:"false"`
		AuthEnabled bool   `env:"METRICS_AUTH_ENABLED" envDefault:"false"`
		Username    string `env:"METRICS_USER"`
		Password    string `env:"METRICS_PASS"`
	}

	Telemetry struct {
		Enabled      bool   `env:"OTEL_ENABLED" envDefault:"false"`
		ServiceName  string `env:"OTEL_SERVICE_NAME" envDefault:"go-clean-api"`
		CollectorURL string `env:"OTEL_COLLECTOR_URL" envDefault:"localhost:4317"`
	}
}

type StorageConfig struct {
	Driver string `mapstructure:"driver" validate:"required,oneof=local s3"`
	Local  struct {
		RootPath string `mapstructure:"root_path"`
		BaseURL  string `mapstructure:"base_url"`
	} `mapstructure:"local"`
	S3 struct {
		Endpoint       string `mapstructure:"endpoint"`
		Region         string `mapstructure:"region"`
		Bucket         string `mapstructure:"bucket"`
		AccessKey      string `mapstructure:"access_key"`
		SecretKey      string `mapstructure:"secret_key"`
		UseSSL         bool   `mapstructure:"use_ssl"`
		ForcePathStyle bool   `mapstructure:"force_path_style"`
	} `mapstructure:"s3"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port" validate:"required"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	AppName        string        `mapstructure:"app_name"`
	AppEnv         string        `mapstructure:"app_env"`
	TrustedProxies []string      `mapstructure:"trusted_proxies"`
}

type SecurityConfig struct {
	MaxLoginAttempts int           `mapstructure:"max_login_attempts"`
	LockoutDuration  time.Duration `mapstructure:"lockout_duration"`
}

type MetricsConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	AuthEnabled bool   `mapstructure:"auth_enabled"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
}

type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	RPS     float64 `mapstructure:"rps"`
	Burst   int     `mapstructure:"burst"`
	Store   string  `mapstructure:"store"` // "memory" or "redis"
}

type SMTPConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	FromSender string `mapstructure:"from_sender"`
	FromEmail  string `mapstructure:"from_email"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CircuitBreakerConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	MaxRequests uint32        `mapstructure:"max_requests"`
	Interval    time.Duration `mapstructure:"interval"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type MySqlConfig struct {
	Host                  string `mapstructure:"host" validate:"required"`
	Port                  int    `mapstructure:"port" validate:"required"`
	User                  string `mapstructure:"user" validate:"required"`
	Password              string `mapstructure:"password" validate:"required"`
	DBName                string `mapstructure:"dbname" validate:"required"`
	IdleConnection        int    `mapstructure:"idle_connection"`
	MaxConnection         int    `mapstructure:"max_connection"`
	MaxLifeTimeConnection int    `mapstructure:"max_life_time_connection"`
}

type RedisConfig struct {
	Addr         string        `mapstructure:"addr" validate:"required"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type JWTConfig struct {
	AccessTokenSecret    string        `mapstructure:"access_secret" validate:"required,min=32"`
	RefreshTokenSecret   string        `mapstructure:"refresh_secret" validate:"required,min=32"`
	AccessTokenDuration  time.Duration `mapstructure:"access_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_duration"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type CasbinConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Model   string        `mapstructure:"model"`
	Watcher WatcherConfig `mapstructure:"watcher"`
}

type WatcherConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Channel string `mapstructure:"channel"`
}

func NewConfig() (*AppConfig, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading configuration from environment variables")
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("log.level", "info")
	v.SetDefault("mysql.host", "localhost")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("jwt.access_duration", "15m")
	v.SetDefault("jwt.refresh_duration", "720h")
	v.SetDefault("security.max_login_attempts", 5)
	v.SetDefault("security.lockout_duration", "30m")
	v.SetDefault("casbin.enabled", false)
	v.SetDefault("casbin.model", "internal/config/casbin_model.conf")
	v.SetDefault("casbin.watcher.enabled", false)
	v.SetDefault("casbin.watcher.channel", "/casbin")
	// v.SetDefault("cors.allowed_origins", "*") // Removed unsafe default
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.rps", 10.0)
	v.SetDefault("rate_limit.burst", 20)
	v.SetDefault("rate_limit.store", "memory")
	v.SetDefault("smtp.host", "localhost")
	v.SetDefault("smtp.port", 1025)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from_sender", "NexusOS Admin")
	v.SetDefault("smtp.from_email", "no-reply@nexusos.dev")
	v.SetDefault("circuit_breaker.enabled", true)
	v.SetDefault("circuit_breaker.max_requests", 5)
	v.SetDefault("circuit_breaker.interval", "60s")
	v.SetDefault("circuit_breaker.timeout", "30s")
	v.SetDefault("websocket.distributed_enabled", false)
	v.SetDefault("websocket.redis_prefix", "ws_broadcast:")
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.auth_enabled", false)
	// v.SetDefault("metrics.username", "admin")      // Removed hardcoded default
	// v.SetDefault("metrics.password", "metrics123") // Removed hardcoded default
	v.SetDefault("storage.driver", "local")
	v.SetDefault("storage.local.root_path", "./uploads")
	v.SetDefault("storage.local.base_url", "http://localhost:8080/uploads")
	v.SetDefault("storage.s3.use_ssl", true)

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.CircuitBreaker.Enabled = v.GetBool("circuit_breaker.enabled")
	cfg.CircuitBreaker.MaxRequests = v.GetUint32("circuit_breaker.max_requests")
	cfg.CircuitBreaker.Interval = v.GetDuration("circuit_breaker.interval")
	cfg.CircuitBreaker.Timeout = v.GetDuration("circuit_breaker.timeout")

	cfg.Storage.Driver = v.GetString("storage.driver")
	cfg.Storage.Local.RootPath = v.GetString("storage.local.root_path")
	cfg.Storage.Local.BaseURL = v.GetString("storage.local.base_url")
	cfg.Storage.S3.Endpoint = v.GetString("storage.s3.endpoint")
	cfg.Storage.S3.Region = v.GetString("storage.s3.region")
	cfg.Storage.S3.Bucket = v.GetString("storage.s3.bucket")
	cfg.Storage.S3.AccessKey = v.GetString("storage.s3.access_key")
	cfg.Storage.S3.SecretKey = v.GetString("storage.s3.secret_key")
	cfg.Storage.S3.UseSSL = v.GetBool("storage.s3.use_ssl")
	cfg.Storage.S3.ForcePathStyle = v.GetBool("storage.s3.force_path_style")

	cfg.JWT.AccessTokenSecret = v.GetString("jwt.access_secret")
	cfg.JWT.RefreshTokenSecret = v.GetString("jwt.refresh_secret")

	cfg.Security.MaxLoginAttempts = v.GetInt("security.max_login_attempts")
	cfg.Security.LockoutDuration = v.GetDuration("security.lockout_duration")

	cfg.Redis.Addr = v.GetString("redis.addr")
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.Redis.DB = v.GetInt("redis.db")
	cfg.Redis.PoolSize = v.GetInt("redis.pool_size")

	cfg.WebSocket.DistributedEnabled = v.GetBool("websocket.distributed_enabled")
	cfg.WebSocket.RedisPrefix = v.GetString("websocket.redis_prefix")

	cfg.Server.Port = v.GetInt("server.port")
	cfg.Server.AppEnv = v.GetString("server.app_env")
	cfg.Server.AppName = v.GetString("server.app_name")
	cfg.Server.ReadTimeout = v.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = v.GetDuration("server.write_timeout")
	if trustedProxiesStr := v.GetString("server.trusted_proxies"); trustedProxiesStr != "" && len(cfg.Server.TrustedProxies) == 0 {
		proxies := strings.Split(trustedProxiesStr, ",")
		for i := range proxies {
			proxies[i] = strings.TrimSpace(proxies[i])
		}
		cfg.Server.TrustedProxies = proxies
	}

	cfg.Log.Level = v.GetString("log.level")

	cfg.Mysql.Host = v.GetString("mysql.host")
	cfg.Mysql.Port = v.GetInt("mysql.port")
	cfg.Mysql.User = v.GetString("mysql.user")
	cfg.Mysql.Password = v.GetString("mysql.password")
	cfg.Mysql.DBName = v.GetString("mysql.dbname")
	cfg.Mysql.IdleConnection = v.GetInt("mysql.idle_connection")
	cfg.Mysql.MaxConnection = v.GetInt("mysql.max_connection")
	cfg.Mysql.MaxLifeTimeConnection = v.GetInt("mysql.max_life_time_connection")

	cfg.Casbin.Enabled = v.GetBool("casbin.enabled")
	cfg.Casbin.Model = v.GetString("casbin.model")
	cfg.Casbin.Watcher.Enabled = v.GetBool("casbin.watcher.enabled")
	cfg.Casbin.Watcher.Channel = v.GetString("casbin.watcher.channel")

	cfg.Metrics.Enabled = v.GetBool("metrics.enabled")
	cfg.Metrics.AuthEnabled = v.GetBool("metrics.auth_enabled")
	cfg.Metrics.Username = v.GetString("metrics.username")
	cfg.Metrics.Password = v.GetString("metrics.password")

	if corsStr := v.GetString("cors.allowed_origins"); corsStr != "" && len(cfg.CORS.AllowedOrigins) == 0 {
		origins := strings.Split(corsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		cfg.CORS.AllowedOrigins = origins
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if cfg.Metrics.AuthEnabled {
		if cfg.Metrics.Username == "" || cfg.Metrics.Password == "" {
			return nil, fmt.Errorf("metrics auth is enabled but username or password is missing")
		}
	}

	return &cfg, nil
}
