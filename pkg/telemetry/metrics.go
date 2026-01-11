package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// User metrics
	UserLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_user_logins_total",
			Help: "Total number of user login attempts",
		},
		[]string{"status"}, // success, failed
	)

	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_user_registrations_total",
			Help: "Total number of new user registrations",
		},
	)

	// WebSocket metrics
	ActiveWSConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_websocket_active_connections",
			Help: "Number of currently active WebSocket connections",
		},
	)

	// Storage metrics
	StorageUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_storage_uploads_total",
			Help: "Total number of file uploads",
		},
		[]string{"driver", "status"},
	)

	// Cleanup Job metrics
	CleanupTasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_cleanup_tasks_total",
			Help: "Total number of maintenance cleanup tasks executed",
		},
		[]string{"task_type", "status"},
	)
)
