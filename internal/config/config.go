package config

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// AppConfig holds all configuration for the application, loaded from environment variables.
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
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	AppName      string        `mapstructure:"app_name"`
	AppEnv       string        `mapstructure:"app_env"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	RPS     float64 `mapstructure:"rps"`
	Burst   int     `mapstructure:"burst"`
	Store   string  `mapstructure:"store"` // "memory" or "redis"
}

// CORSConfig holds CORS-related configuration.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// MySqlConfig holds PostgreSQL database connection details.
type MySqlConfig struct {
	Host                  string `mapstructure:"host"`
	Port                  int    `mapstructure:"port"`
	User                  string `mapstructure:"user"`
	Password              string `mapstructure:"password"`
	DBName                string `mapstructure:"dbname"`
	IdleConnection        int    `mapstructure:"idle_connection"`
	MaxConnection         int    `mapstructure:"max_connection"`
	MaxLifeTimeConnection int    `mapstructure:"max_life_time_connection"`
}

// RedisConfig holds Redis connection details.
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// JWTConfig holds JWT-related configuration.
type JWTConfig struct {
	AccessTokenSecret    string        `mapstructure:"access_secret"`
	RefreshTokenSecret   string        `mapstructure:"refresh_secret"`
	AccessTokenDuration  time.Duration `mapstructure:"access_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_duration"`
}

// LoggerConfig holds logging level configuration.
type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// CasbinConfig holds Casbin-related configuration.
type CasbinConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Model   string        `mapstructure:"model"`
	Watcher WatcherConfig `mapstructure:"watcher"`
}

// WatcherConfig holds Casbin Redis watcher configuration.
type WatcherConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Channel string `mapstructure:"channel"`
}

// NewConfig initializes and returns the application's configuration by reading from
// a .env file and environment variables.
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
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("jwt.access_duration", "15m")
	v.SetDefault("jwt.refresh_duration", "720h")
	v.SetDefault("casbin.enabled", false)
	v.SetDefault("casbin.model", "internal/config/casbin_model.conf")
	v.SetDefault("casbin.watcher.enabled", false)
	v.SetDefault("casbin.watcher.channel", "/casbin")

	// CORS Defaults
	v.SetDefault("cors.allowed_origins", "*")

	// Rate Limit Defaults
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.rps", 10.0)
	v.SetDefault("rate_limit.burst", 20)
	v.SetDefault("rate_limit.store", "memory")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Parse CORS allowed origins from comma-separated string if needed
	// Note: viper unmarshal might handle slice if env var is list, but comma-string is safer for .env
	if corsStr := v.GetString("cors.allowed_origins"); corsStr != "" && len(cfg.CORS.AllowedOrigins) == 0 {
		origins := strings.Split(corsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		cfg.CORS.AllowedOrigins = origins
	}

	// Manual overrides if needed (viper unmarshal should cover most)
	cfg.JWT.AccessTokenSecret = v.GetString("jwt.access_secret")
	cfg.JWT.RefreshTokenSecret = v.GetString("jwt.refresh_secret")

	cfg.Redis.Addr = v.GetString("redis.addr")
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.Redis.DB = v.GetInt("redis.db")
	cfg.Redis.PoolSize = v.GetInt("redis.pool_size")

	cfg.Server.Port = v.GetInt("server.port")
	cfg.Server.AppEnv = v.GetString("server.app_env")
	cfg.Server.AppName = v.GetString("server.app_name")
	cfg.Server.ReadTimeout = v.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = v.GetDuration("server.write_timeout")

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

	//Rate Limit Defaults
	cfg.RateLimit.Enabled = v.GetBool("rate_limit.enabled")
	cfg.RateLimit.RPS = v.GetFloat64("rate_limit.rps")
	cfg.RateLimit.Burst = v.GetInt("rate_limit.burst")
	cfg.RateLimit.Store = v.GetString("rate_limit.store")

	return &cfg, nil
}
