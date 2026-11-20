package config

import (
	"context"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsStrictCasbinEnv(t *testing.T) {
	tests := []struct {
		name     string
		appEnv   string
		expected bool
	}{
		{name: "empty is non strict", appEnv: "", expected: false},
		{name: "local is non strict", appEnv: "local", expected: false},
		{name: "dev is non strict", appEnv: "dev", expected: false},
		{name: "development is non strict", appEnv: "development", expected: false},
		{name: "test is non strict", appEnv: "test", expected: false},
		{name: "testing is non strict", appEnv: "testing", expected: false},
		{name: "production is strict", appEnv: "production", expected: true},
		{name: "staging is strict", appEnv: "staging", expected: true},
		{name: "case and spaces normalize", appEnv: " Production ", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStrictCasbinEnv(tt.appEnv); got != tt.expected {
				t.Fatalf("isStrictCasbinEnv(%q) = %v, want %v", tt.appEnv, got, tt.expected)
			}
		})
	}
}

func TestRegisterPrometheusCollectorsIdempotent(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	sqlxDB := sqlx.NewDb(sqlDB, "mysql")
	logger := logrus.New()
	sseManager := sse.NewManager()
	defer sseManager.Stop()
	wsManager := ws2.NewWebSocketManager(NewDefaultWebSocketConfig().ToPkgConfig(), logger, nil, nil)
	defer wsManager.Stop()

	cfg := &AppConfig{
		Mysql: MySqlConfig{DBName: "casbin_test"},
	}

	reg := prometheus.NewRegistry()

	// Register twice to verify AlreadyRegisteredError is ignored without panic.
	require.NotPanics(t, func() {
		registerPrometheusCollectorsWith(reg, cfg, sqlxDB, sseManager, wsManager, logger)
		registerPrometheusCollectorsWith(reg, cfg, sqlxDB, sseManager, wsManager, logger)
	})

	families, err := reg.Gather()
	require.NoError(t, err)

	metricNames := make(map[string]bool, len(families))
	for _, mf := range families {
		metricNames[mf.GetName()] = true
	}

	assert.True(t, metricNames["app_db_pool_open_connections"])
	assert.True(t, metricNames["app_sse_active_connections"])
	assert.True(t, metricNames["app_ws_active_connections"])
}

func TestApplicationShutdownOrderAndIdempotency(t *testing.T) {
	logger := logrus.New()
	var events []string

	sseManager := sse.NewManager()
	wsManager := ws2.NewWebSocketManager(NewDefaultWebSocketConfig().ToPkgConfig(), logger, nil, nil)
	go wsManager.Run()

	app := &Application{
		Server: &http.Server{},
		Log:    logger,
		TaskProcessor: &stubTaskProcessor{
			onStop: func() {
				events = append(events, "processor:stop")
			},
		},
		BroadcasterCancel: func() {
			events = append(events, "broadcaster:stop")
		},
		TracerShutdown: func(context.Context) error {
			events = append(events, "tracer:shutdown")
			return nil
		},
		wsManager:  wsManager,
		sseManager: sseManager,
	}

	require.NoError(t, app.Shutdown(context.Background()))
	require.Equal(t, []string{
		"processor:stop",
		"broadcaster:stop",
		"tracer:shutdown",
	}, events)

	// Calling Shutdown a second time must not panic (e.g. sseManager.Stop() guarded by sync.Once).
	require.NotPanics(t, func() {
		require.NoError(t, app.Shutdown(context.Background()))
	})
}
