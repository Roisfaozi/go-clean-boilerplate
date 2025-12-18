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
}

type ServerConfig struct {
	Port         int           `mapstructure:"port" validate:"required"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	AppName      string        `mapstructure:"app_name"`
	AppEnv       string        `mapstructure:"app_env"`
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

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

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
