// This file is intentionally left blank.
// All JWT configuration logic has been consolidated into internal/config/config.go
// to create a single, centralized configuration management system.
package config

import (
	"time"
)

func (c *AppConfig) GetAccessTokenSecret() string {
	return c.JWT.AccessTokenSecret
}

func (c *AppConfig) GetRefreshTokenSecret() string {
	return c.JWT.RefreshTokenSecret
}

func (c *AppConfig) GetAccessTokenDuration() time.Duration {
	return c.JWT.AccessTokenDuration
}

func (c *AppConfig) GetRefreshTokenDuration() time.Duration {
	return c.JWT.RefreshTokenDuration
}
