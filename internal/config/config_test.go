package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_USER", "testuser")
	t.Setenv("MYSQL_PASSWORD", "testpass")
	t.Setenv("MYSQL_DBNAME", "testdb")
	t.Setenv("JWT_ACCESS_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_REFRESH_SECRET", "01234567890123456789012345678901")
}

func TestNewConfig_DefaultValues(t *testing.T) {
	setupTestEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "localhost", cfg.Mysql.Host)
	assert.Equal(t, 3306, cfg.Mysql.Port)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, false, cfg.Casbin.Enabled)
	assert.Equal(t, true, cfg.RateLimit.Enabled)
	assert.Equal(t, "memory", cfg.RateLimit.Store)
	assert.Equal(t, "local", cfg.Storage.Driver)
}

func TestNewConfig_JWTDurations(t *testing.T) {
	setupTestEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenDuration)
	assert.Equal(t, 720*time.Hour, cfg.JWT.RefreshTokenDuration)
}

func TestNewConfig_SecurityDefaults(t *testing.T) {
	setupTestEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.Security.MaxLoginAttempts)
	assert.Equal(t, 30*time.Minute, cfg.Security.LockoutDuration)
}

func TestNewConfig_CircuitBreakerDefaults(t *testing.T) {
	setupTestEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.True(t, cfg.CircuitBreaker.Enabled)
	assert.Equal(t, uint32(5), cfg.CircuitBreaker.MaxRequests)
	assert.Equal(t, 60*time.Second, cfg.CircuitBreaker.Interval)
	assert.Equal(t, 30*time.Second, cfg.CircuitBreaker.Timeout)
}

func TestNewConfig_MetricsAuthValidation(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("METRICS_AUTH_ENABLED", "true")
	t.Setenv("METRICS_USERNAME", "")
	t.Setenv("METRICS_PASSWORD", "")

	_, err := NewConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metrics auth is enabled but username or password is missing")
}

func TestNewConfig_TrustedProxiesParsing(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("SERVER_TRUSTED_PROXIES", "10.0.0.1, 10.0.0.2, 192.168.1.1")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2", "192.168.1.1"}, cfg.Server.TrustedProxies)
}

func TestNewConfig_CORSOriginsParsing(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com")

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, cfg.CORS.AllowedOrigins)
}

func TestNewConfig_StorageDrivers(t *testing.T) {
	setupTestEnv(t)

	cfg, err := NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "local", cfg.Storage.Driver)
	assert.Equal(t, "./uploads", cfg.Storage.Local.RootPath)
}

func TestNewConfig_Validation_RequiredFields(t *testing.T) {
	setupTestEnv(t)
	// MYSQL_PORT has validate:"required" (non-zero)
	t.Setenv("MYSQL_PORT", "0")

	_, err := NewConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config validation failed")
}
