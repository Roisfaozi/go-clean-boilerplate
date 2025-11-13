package config

import (
	"time"

	"github.com/spf13/viper"
)

type JWTConfig struct {
	accessTokenSecret    string
	refreshTokenSecret   string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJWTConfig(config *viper.Viper) *JWTConfig {
	return &JWTConfig{
		accessTokenSecret:    config.GetString("jwt.access.secret"),
		refreshTokenSecret:   config.GetString("jwt.refresh.secret"),
		accessTokenDuration:  config.GetDuration("jwt.access.duration"),
		refreshTokenDuration: config.GetDuration("jwt.refresh.duration"),
	}
}

func (c *JWTConfig) GetAccessTokenSecret() string {
	return c.accessTokenSecret
}

func (c *JWTConfig) GetRefreshTokenSecret() string {
	return c.refreshTokenSecret
}

func (c *JWTConfig) GetAccessTokenDuration() time.Duration {
	return c.accessTokenDuration
}

func (c *JWTConfig) GetRefreshTokenDuration() time.Duration {
	return c.refreshTokenDuration
}
