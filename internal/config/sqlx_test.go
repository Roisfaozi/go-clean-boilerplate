package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMySQLDSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  MySqlConfig
		want string
	}{
		{
			name: "standard config",
			cfg: MySqlConfig{
				User:     "appuser",
				Password: "secret",
				Host:     "127.0.0.1",
				Port:     3306,
				DBName:   "casbin_db",
			},
			want: "appuser:secret@tcp(127.0.0.1:3306)/casbin_db?charset=utf8mb4&parseTime=True&loc=UTC",
		},
		{
			name: "empty password and custom port",
			cfg: MySqlConfig{
				User:     "root",
				Password: "",
				Host:     "mysql",
				Port:     3307,
				DBName:   "test_db",
			},
			want: "root:@tcp(mysql:3307)/test_db?charset=utf8mb4&parseTime=True&loc=UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildMySQLDSN(tt.cfg))
		})
	}
}

func TestOpenSQLXDB(t *testing.T) {
	baseCfg := func(telemetryEnabled bool) *AppConfig {
		return &AppConfig{
			Mysql: MySqlConfig{
				User:     "root",
				Password: "password",
				Host:     "localhost",
				Port:     3306,
				DBName:   "app",
			},
			Telemetry: TelemetryConfig{
				Enabled: telemetryEnabled,
			},
		}
	}

	t.Run("telemetry disabled uses sqlx.Open", func(t *testing.T) {
		db, err := openSQLXDB(baseCfg(false))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func() { _ = db.Close() }()

		assert.Equal(t, "mysql", db.DriverName())
	})

	t.Run("telemetry enabled wraps otelsql with mysql driver name", func(t *testing.T) {
		db, err := openSQLXDB(baseCfg(true))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func() { _ = db.Close() }()

		assert.Equal(t, "mysql", db.DriverName())
	})
}
