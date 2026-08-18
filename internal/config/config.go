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

const (
	defaultServerPort                    = 8080
	defaultServerReadTimeout             = "30s"
	defaultServerWriteTimeout            = "30s"
	defaultLogLevel                      = "info"
	defaultMySQLHost                     = "localhost"
	defaultMySQLPort                     = 3306
	defaultRedisAddr                     = "localhost:6379"
	defaultJWTAccessDuration             = "15m"
	defaultJWTRefreshDuration            = "24h"
	defaultSecurityMaxLoginAttempts      = 5
	defaultSecurityLockoutDuration       = "30m"
	defaultSecurityMaxConcurrentSessions = 3
	defaultCookieSameSite                = "lax"
	defaultCookieDomain                  = ""
	defaultCasbinEnabled                 = true
	defaultCasbinModel                   = "internal/config/casbin_model.conf"
	defaultCasbinDefaultRole             = "role:user"
	defaultCasbinDefaultDomain           = "global"
	defaultCasbinWatcherEnabled          = false
	defaultCasbinWatcherChannel          = "/casbin"
	defaultRateLimitEnabled              = true
	defaultRateLimitRPS                  = 10.0
	defaultRateLimitBurst                = 20
	defaultRateLimitStore                = "memory"
	defaultSMTPHost                      = "localhost"
	defaultSMTPPort                      = 1025
	defaultSMTPUsername                  = ""
	defaultSMTPPassword                  = ""
	defaultSMTPFromSender                = "NexusOS Admin"
	defaultSMTPFromEmail                 = "no-reply@nexusos.dev"
	defaultCircuitBreakerEnabled         = true
	defaultCircuitBreakerMaxRequests     = 5
	defaultCircuitBreakerInterval        = "60s"
	defaultCircuitBreakerTimeout         = "30s"
	defaultWebSocketRedisPrefix          = "ws_broadcast:"
	defaultMetricsEnabled                = true
	defaultMetricsAuthEnabled            = false
	defaultStorageDriver                 = "local"
	defaultStorageLocalRootPath          = "./uploads"
	defaultStorageLocalBaseURL           = "http://localhost:8080/uploads"
	defaultStorageS3UseSSL               = true
	defaultTusBasePath                   = "/api/v1/upload/files/"
	defaultPprofEnabled                  = false
	defaultPprofPort                     = 6060
	defaultTelemetryServiceName          = "go-clean-api"
	defaultTelemetryCollectorURL         = "localhost:4317"
	defaultAppEnvDevelopment             = "development"
	defaultAppEnvLocal                   = "local"
	defaultAppEnvDev                     = "dev"
	defaultAppEnvTest                    = "test"
	defaultAppEnvTesting                 = "testing"
	storageDriverLocal                   = "local"
	storageDriverS3                      = "s3"
)

type AppConfig struct {
	Server         ServerConfig         `mapstructure:"server"`
	Mysql          MySqlConfig          `mapstructure:"mysql"`
	Redis          RedisConfig          `mapstructure:"redis"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	Security       SecurityConfig       `mapstructure:"security"`
	Cookie         CookieConfig         `mapstructure:"cookie"`
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
	Tus   TusConfig   `mapstructure:"tus"`
	Pprof PprofConfig `mapstructure:"pprof"`
	SSO   SSOConfig   `mapstructure:"sso"`
}

type SSOConfig struct {
	Google    OAuthProviderConfig `mapstructure:"google"`
	Microsoft OAuthProviderConfig `mapstructure:"microsoft"`
	GitHub    OAuthProviderConfig `mapstructure:"github"`
}

type OAuthProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
}

type TusConfig struct {
	BasePath string `mapstructure:"base_path"`
}

type PprofConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
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
	Port            int           `mapstructure:"port" validate:"required"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	AppName         string        `mapstructure:"app_name"`
	AppEnv          string        `mapstructure:"app_env"`
	TrustedProxies  []string      `mapstructure:"trusted_proxies"`
	FrontendBaseURL string        `mapstructure:"frontend_base_url"`
}

type SecurityConfig struct {
	MaxLoginAttempts      int           `mapstructure:"max_login_attempts"`
	LockoutDuration       time.Duration `mapstructure:"lockout_duration"`
	MaxConcurrentSessions int           `mapstructure:"max_concurrent_sessions"`
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

type CookieConfig struct {
	Domain   string `mapstructure:"domain"`
	SameSite string `mapstructure:"same_site" validate:"omitempty,oneof=lax strict none"`
	Secure   *bool  `mapstructure:"secure"`
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
	Enabled       bool          `mapstructure:"enabled"`
	Model         string        `mapstructure:"model"`
	DefaultRole   string        `mapstructure:"default_role"`
	DefaultDomain string        `mapstructure:"default_domain"`
	Watcher       WatcherConfig `mapstructure:"watcher"`
}

type WatcherConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Channel string `mapstructure:"channel"`
}

// envOnlyKeys lists configuration keys that have no default and therefore need
// an explicit environment binding. Without this, viper leaves them empty even
// when the matching variable is set.
var envOnlyKeys = []string{
	"server.frontend_base_url",
	"cookie.secure",
	"redis.dial_timeout",
	"redis.read_timeout",
	"redis.write_timeout",
}

func NewConfig() (*AppConfig, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading configuration from environment variables")
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// viper's AutomaticEnv only resolves keys it already knows about (defaults,
	// config file, or explicit binds). Any field that is populated purely by
	// Unmarshal would be silently dropped, so every default-less key that must
	// come from the environment is bound here explicitly.
	for _, key := range envOnlyKeys {
		_ = v.BindEnv(key)
	}

	v.SetDefault("server.port", defaultServerPort)
	v.SetDefault("server.read_timeout", defaultServerReadTimeout)
	v.SetDefault("server.write_timeout", defaultServerWriteTimeout)
	v.SetDefault("log.level", defaultLogLevel)
	v.SetDefault("mysql.host", defaultMySQLHost)
	v.SetDefault("mysql.port", defaultMySQLPort)
	v.SetDefault("redis.addr", defaultRedisAddr)
	v.SetDefault("jwt.access_duration", defaultJWTAccessDuration)
	v.SetDefault("jwt.refresh_duration", defaultJWTRefreshDuration)
	v.SetDefault("security.max_login_attempts", defaultSecurityMaxLoginAttempts)
	v.SetDefault("security.lockout_duration", defaultSecurityLockoutDuration)
	v.SetDefault("security.max_concurrent_sessions", defaultSecurityMaxConcurrentSessions)
	v.SetDefault("cookie.same_site", defaultCookieSameSite)
	v.SetDefault("cookie.domain", defaultCookieDomain)
	v.SetDefault("casbin.enabled", defaultCasbinEnabled)
	v.SetDefault("casbin.model", defaultCasbinModel)
	v.SetDefault("casbin.default_role", defaultCasbinDefaultRole)
	v.SetDefault("casbin.default_domain", defaultCasbinDefaultDomain)
	v.SetDefault("casbin.watcher.enabled", defaultCasbinWatcherEnabled)
	v.SetDefault("casbin.watcher.channel", defaultCasbinWatcherChannel)
	// v.SetDefault("cors.allowed_origins", "*") // Removed unsafe default
	v.SetDefault("rate_limit.enabled", defaultRateLimitEnabled)
	v.SetDefault("rate_limit.rps", defaultRateLimitRPS)
	v.SetDefault("rate_limit.burst", defaultRateLimitBurst)
	v.SetDefault("rate_limit.store", defaultRateLimitStore)
	v.SetDefault("smtp.host", defaultSMTPHost)
	v.SetDefault("smtp.port", defaultSMTPPort)
	v.SetDefault("smtp.username", defaultSMTPUsername)
	v.SetDefault("smtp.password", defaultSMTPPassword)
	v.SetDefault("smtp.from_sender", defaultSMTPFromSender)
	v.SetDefault("smtp.from_email", defaultSMTPFromEmail)
	v.SetDefault("circuit_breaker.enabled", defaultCircuitBreakerEnabled)
	v.SetDefault("circuit_breaker.max_requests", defaultCircuitBreakerMaxRequests)
	v.SetDefault("circuit_breaker.interval", defaultCircuitBreakerInterval)
	v.SetDefault("circuit_breaker.timeout", defaultCircuitBreakerTimeout)
	v.SetDefault("websocket.distributed_enabled", false)
	v.SetDefault("websocket.redis_prefix", defaultWebSocketRedisPrefix)
	v.SetDefault("metrics.enabled", defaultMetricsEnabled)
	v.SetDefault("metrics.auth_enabled", defaultMetricsAuthEnabled)
	// v.SetDefault("metrics.username", "admin")      // Removed hardcoded default
	// v.SetDefault("metrics.password", "metrics123") // Removed hardcoded default
	v.SetDefault("storage.driver", defaultStorageDriver)
	v.SetDefault("storage.local.root_path", defaultStorageLocalRootPath)
	v.SetDefault("storage.local.base_url", defaultStorageLocalBaseURL)
	v.SetDefault("storage.s3.use_ssl", defaultStorageS3UseSSL)
	v.SetDefault("tus.base_path", defaultTusBasePath)
	v.SetDefault("pprof.enabled", defaultPprofEnabled)
	v.SetDefault("pprof.port", defaultPprofPort)

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

	cfg.Tus.BasePath = v.GetString("tus.base_path")

	cfg.Pprof.Enabled = v.GetBool("pprof.enabled")
	cfg.Pprof.Port = v.GetInt("pprof.port")

	cfg.JWT.AccessTokenSecret = v.GetString("jwt.access_secret")
	cfg.JWT.RefreshTokenSecret = v.GetString("jwt.refresh_secret")

	cfg.SSO.Google.ClientID = v.GetString("sso.google.client_id")
	cfg.SSO.Google.ClientSecret = v.GetString("sso.google.client_secret")
	cfg.SSO.Google.RedirectURL = v.GetString("sso.google.redirect_url")

	cfg.SSO.Microsoft.ClientID = v.GetString("sso.microsoft.client_id")
	cfg.SSO.Microsoft.ClientSecret = v.GetString("sso.microsoft.client_secret")
	cfg.SSO.Microsoft.RedirectURL = v.GetString("sso.microsoft.redirect_url")

	cfg.SSO.GitHub.ClientID = v.GetString("sso.github.client_id")
	cfg.SSO.GitHub.ClientSecret = v.GetString("sso.github.client_secret")
	cfg.SSO.GitHub.RedirectURL = v.GetString("sso.github.redirect_url")

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
	cfg.Casbin.DefaultRole = v.GetString("casbin.default_role")
	cfg.Casbin.DefaultDomain = v.GetString("casbin.default_domain")
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
