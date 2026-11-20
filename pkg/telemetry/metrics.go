package telemetry

import (
	"database/sql"
	"strings"
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

	AuditOutboxOldestAgeSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_audit_outbox_oldest_age_seconds",
			Help: "Age in seconds of the oldest pending audit outbox entry observed in the latest worker batch",
		},
	)

	// Redis operation metrics
	RedisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_redis_operations_total",
			Help: "Total number of Redis operations by bounded command and status",
		},
		[]string{"cmd", "status"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_redis_operation_duration_seconds",
			Help:    "Duration of Redis operations in seconds",
			Buckets: []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"cmd", "status"},
	)

	// Rate-limit rejection metrics
	RateLimitRejectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_rate_limit_rejections_total",
			Help: "Total number of HTTP requests rejected by rate limiters",
		},
		[]string{"limiter"},
	)

	// TUS upload metrics
	TUSUploadSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_tus_upload_size_bytes",
			Help:    "Size of completed TUS uploads in bytes",
			Buckets: []float64{1024, 10240, 102400, 1048576, 5242880, 10485760, 52428800, 104857600},
		},
		[]string{"driver"},
	)

	TUSUploadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_tus_upload_duration_seconds",
			Help:    "Duration of TUS upload completion hook handling in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"driver", "status"},
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

// SetAuditOutboxOldestAgeSeconds sets the age in seconds of the oldest pending audit outbox entry.
func SetAuditOutboxOldestAgeSeconds(seconds float64) {
	if seconds < 0 {
		seconds = 0
	}
	AuditOutboxOldestAgeSeconds.Set(seconds)
}

// RecordRedisOperation records a Redis command execution with bounded command and status labels.
func RecordRedisOperation(cmd, status string, duration time.Duration) {
	normalizedCmd := normalizeRedisCmd(cmd)
	normalizedStatus := normalizeRedisStatus(status)
	if duration < 0 {
		duration = 0
	}
	RedisOperationsTotal.WithLabelValues(normalizedCmd, normalizedStatus).Inc()
	RedisOperationDuration.WithLabelValues(normalizedCmd, normalizedStatus).Observe(duration.Seconds())
}

func normalizeRedisCmd(cmd string) string {
	switch strings.ToLower(strings.TrimSpace(cmd)) {
	case "get", "set", "del", "getdel", "setnx", "expire", "ttl", "incr",
		"eval", "evalsha", "sadd", "srem", "smembers", "scard",
		"zadd", "zremrangebyscore", "zrange", "exists", "ping", "pipeline":
		return strings.ToLower(strings.TrimSpace(cmd))
	default:
		return "other"
	}
}

func normalizeRedisStatus(status string) string {
	switch status {
	case "ok", "miss", "error":
		return status
	default:
		return "error"
	}
}

// RecordRateLimitRejection increments the rate-limit rejection counter for a bounded limiter type.
func RecordRateLimitRejection(limiter string) {
	RateLimitRejectionsTotal.WithLabelValues(normalizeRateLimiter(limiter)).Inc()
}

func normalizeRateLimiter(limiter string) string {
	switch limiter {
	case "ip", "user", "memory":
		return limiter
	default:
		return "other"
	}
}

// RecordTUSUpload records completion metrics for a TUS upload using bounded driver and status labels.
func RecordTUSUpload(driver, status string, sizeBytes int64, duration time.Duration) {
	normalizedDriver := normalizeTUSDriver(driver)
	normalizedStatus := normalizeTUSStatus(status)
	if sizeBytes < 0 {
		sizeBytes = 0
	}
	if duration < 0 {
		duration = 0
	}
	StorageUploadsTotal.WithLabelValues(normalizedDriver, normalizedStatus).Inc()
	TUSUploadSizeBytes.WithLabelValues(normalizedDriver).Observe(float64(sizeBytes))
	TUSUploadDuration.WithLabelValues(normalizedDriver, normalizedStatus).Observe(duration.Seconds())
}

func normalizeTUSDriver(driver string) string {
	if driver == "s3" {
		return "s3"
	}
	return "local"
}

func normalizeTUSStatus(status string) string {
	if status == "success" {
		return "success"
	}
	return "error"
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
