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
	Server    ServerConfig    `mapstructure:"server"`
	Mysql     MySqlConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LoggerConfig    `mapstructure:"log"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Casbin    CasbinConfig    `mapstructure:"casbin"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port" validate:"required"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	AppName        string        `mapstructure:"app_name"`
	AppEnv         string        `mapstructure:"app_env"`
	TrustedProxies []string      `mapstructure:"trusted_proxies"`
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

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
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
	v.SetDefault("casbin.enabled", false)
	v.SetDefault("casbin.model", "internal/config/casbin_model.conf")
	v.SetDefault("casbin.watcher.enabled", false)
	v.SetDefault("casbin.watcher.channel", "/casbin")
	v.SetDefault("cors.allowed_origins", "*")
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.rps", 10.0)
	v.SetDefault("rate_limit.burst", 20)
	v.SetDefault("rate_limit.store", "memory")
	v.SetDefault("websocket.distributed_enabled", false)
	v.SetDefault("websocket.redis_prefix", "ws_broadcast:")
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.auth_enabled", false)
	v.SetDefault("metrics.username", "admin")
	v.SetDefault("metrics.password", "metrics123")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.JWT.AccessTokenSecret = v.GetString("jwt.access_secret")
	cfg.JWT.RefreshTokenSecret = v.GetString("jwt.refresh_secret")

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

	return &cfg, nil
}
