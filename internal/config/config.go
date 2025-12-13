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
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	AppName      string        `mapstructure:"app_name"`
	AppEnv       string        `mapstructure:"app_env"`
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

	// Set default values
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

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	//JWT
	cfg.JWT.AccessTokenSecret = v.GetString("jwt.access_secret")
	cfg.JWT.RefreshTokenSecret = v.GetString("jwt.refresh_secret")

	//redis
	cfg.Redis.Addr = v.GetString("redis.addr")
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.Redis.DB = v.GetInt("redis.db")
	cfg.Redis.PoolSize = v.GetInt("redis.pool_size")

	//server
	cfg.Server.Port = v.GetInt("server.port")
	cfg.Server.AppEnv = v.GetString("server.app_env")
	cfg.Server.AppName = v.GetString("server.app_name")
	cfg.Server.ReadTimeout = v.GetDuration("server.read_timeout")
	cfg.Server.WriteTimeout = v.GetDuration("server.write_timeout")

	//log
	cfg.Log.Level = v.GetString("log.level")

	//mysql
	cfg.Mysql.Host = v.GetString("mysql.host")
	cfg.Mysql.Port = v.GetInt("mysql.port")
	cfg.Mysql.User = v.GetString("mysql.user")
	cfg.Mysql.Password = v.GetString("mysql.password")
	cfg.Mysql.DBName = v.GetString("mysql.dbname")
	cfg.Mysql.IdleConnection = v.GetInt("mysql.idle_connection")
	cfg.Mysql.MaxConnection = v.GetInt("mysql.max_connection")
	cfg.Mysql.MaxLifeTimeConnection = v.GetInt("mysql.max_life_time_connection")

	//casbin
	cfg.Casbin.Enabled = v.GetBool("casbin.enabled")
	cfg.Casbin.Model = v.GetString("casbin.model")
	cfg.Casbin.Watcher.Enabled = v.GetBool("casbin.watcher.enabled")
	cfg.Casbin.Watcher.Channel = v.GetString("casbin.watcher.channel")
	return &cfg, nil
}
