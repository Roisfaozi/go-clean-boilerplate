package telemetry

import (
	"database/sql"
	"time"

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

	WSSendDropsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_ws_send_drops_total",
			Help: "Total number of outbound WebSocket messages dropped due to full client send buffer",
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

	// Worker Task metrics
	WorkerTasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_worker_tasks_total",
			Help: "Total number of background worker tasks processed",
		},
		[]string{"task_type", "status"},
	)

	WorkerTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_worker_task_duration_seconds",
			Help:    "Duration of background worker task processing in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"task_type", "status"},
	)

	// Webhook delivery metrics
	WebhookDeliveriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_webhook_deliveries_total",
			Help: "Total number of webhook delivery attempts by bounded outcome",
		},
		[]string{"outcome"},
	)

	WebhookDeliveryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_webhook_delivery_duration_seconds",
			Help:    "Duration of webhook delivery attempts in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"outcome"},
	)

	// Email delivery metrics
	EmailDeliveriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_email_deliveries_total",
			Help: "Total number of email delivery attempts by bounded outcome",
		},
		[]string{"outcome"},
	)

	EmailDeliveryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_email_delivery_duration_seconds",
			Help:    "Duration of email delivery attempts in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"outcome"},
	)

	// Audit outbox metrics
	AuditOutboxPending = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_audit_outbox_pending",
			Help: "Number of pending audit outbox entries observed in the latest worker batch",
		},
	)
)

// RecordWebhookDelivery records a webhook delivery attempt with a bounded outcome label.
func RecordWebhookDelivery(outcome string, duration time.Duration) {
	normalized := normalizeWebhookOutcome(outcome)
	WebhookDeliveriesTotal.WithLabelValues(normalized).Inc()
	WebhookDeliveryDuration.WithLabelValues(normalized).Observe(duration.Seconds())
}

func normalizeWebhookOutcome(outcome string) string {
	switch outcome {
	case "success", "retryable_error", "permanent_error", "error":
		return outcome
	default:
		return "error"
	}
}

// RecordEmailDelivery records an email delivery attempt with a bounded outcome label.
func RecordEmailDelivery(outcome string, duration time.Duration) {
	normalized := normalizeEmailOutcome(outcome)
	EmailDeliveriesTotal.WithLabelValues(normalized).Inc()
	EmailDeliveryDuration.WithLabelValues(normalized).Observe(duration.Seconds())
}

func normalizeEmailOutcome(outcome string) string {
	if outcome == "success" {
		return "success"
	}
	return "error"
}

// IncWSSendDrop increments the WebSocket send drop counter.
func IncWSSendDrop() {
	WSSendDropsTotal.Inc()
}

// SetAuditOutboxPending sets the current number of pending audit outbox entries.
func SetAuditOutboxPending(count int64) {
	AuditOutboxPending.Set(float64(count))
}

type dbStatsCollector struct {
	db     *sql.DB
	dbName string

	openConnections    *prometheus.Desc
	inUseConnections   *prometheus.Desc
	idleConnections    *prometheus.Desc
	maxOpenConnections *prometheus.Desc
	waitCountTotal     *prometheus.Desc
	waitDurationTotal  *prometheus.Desc
}

// NewDBStatsCollector returns a Prometheus collector that reads sql.DB.Stats() on scrape.
func NewDBStatsCollector(db *sql.DB, dbName string) prometheus.Collector {
	if dbName == "" {
		dbName = "default"
	}
	labels := []string{"db"}
	return &dbStatsCollector{
		db:     db,
		dbName: dbName,
		openConnections: prometheus.NewDesc(
			"app_db_pool_open_connections",
			"The number of established connections both in use and idle",
			labels, nil,
		),
		inUseConnections: prometheus.NewDesc(
			"app_db_pool_in_use_connections",
			"The number of connections currently in use",
			labels, nil,
		),
		idleConnections: prometheus.NewDesc(
			"app_db_pool_idle_connections",
			"The number of idle connections",
			labels, nil,
		),
		maxOpenConnections: prometheus.NewDesc(
			"app_db_pool_max_open_connections",
			"Maximum number of open connections to the database",
			labels, nil,
		),
		waitCountTotal: prometheus.NewDesc(
			"app_db_pool_wait_count_total",
			"The total number of connections waited for",
			labels, nil,
		),
		waitDurationTotal: prometheus.NewDesc(
			"app_db_pool_wait_duration_seconds_total",
			"The total time blocked waiting for a new connection in seconds",
			labels, nil,
		),
	}
}

func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.openConnections
	ch <- c.inUseConnections
	ch <- c.idleConnections
	ch <- c.maxOpenConnections
	ch <- c.waitCountTotal
	ch <- c.waitDurationTotal
}

func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	if c.db == nil {
		return
	}
	stats := c.db.Stats()
	ch <- prometheus.MustNewConstMetric(c.openConnections, prometheus.GaugeValue, float64(stats.OpenConnections), c.dbName)
	ch <- prometheus.MustNewConstMetric(c.inUseConnections, prometheus.GaugeValue, float64(stats.InUse), c.dbName)
	ch <- prometheus.MustNewConstMetric(c.idleConnections, prometheus.GaugeValue, float64(stats.Idle), c.dbName)
	ch <- prometheus.MustNewConstMetric(c.maxOpenConnections, prometheus.GaugeValue, float64(stats.MaxOpenConnections), c.dbName)
	ch <- prometheus.MustNewConstMetric(c.waitCountTotal, prometheus.CounterValue, float64(stats.WaitCount), c.dbName)
	ch <- prometheus.MustNewConstMetric(c.waitDurationTotal, prometheus.CounterValue, stats.WaitDuration.Seconds(), c.dbName)
}

type realtimeStatsCollector struct {
	sseCountFn func() int
	wsCountFn  func() int

	sseActive *prometheus.Desc
	wsActive  *prometheus.Desc
}

// NewRealtimeStatsCollector returns a Prometheus collector that reads active SSE and WS connection counts on scrape.
func NewRealtimeStatsCollector(sseCountFn func() int, wsCountFn func() int) prometheus.Collector {
	return &realtimeStatsCollector{
		sseCountFn: sseCountFn,
		wsCountFn:  wsCountFn,
		sseActive: prometheus.NewDesc(
			"app_sse_active_connections",
			"Number of currently active SSE connections",
			nil, nil,
		),
		wsActive: prometheus.NewDesc(
			"app_ws_active_connections",
			"Number of currently active WebSocket connections",
			nil, nil,
		),
	}
}

func (c *realtimeStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sseActive
	ch <- c.wsActive
}

func (c *realtimeStatsCollector) Collect(ch chan<- prometheus.Metric) {
	sseCount := 0
	if c.sseCountFn != nil {
		sseCount = c.sseCountFn()
	}
	wsCount := 0
	if c.wsCountFn != nil {
		wsCount = c.wsCountFn()
	}
	ch <- prometheus.MustNewConstMetric(c.sseActive, prometheus.GaugeValue, float64(sseCount))
	ch <- prometheus.MustNewConstMetric(c.wsActive, prometheus.GaugeValue, float64(wsCount))
}
