package telemetry

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBStatsCollector(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(15)
	expectedStats := db.Stats()

	reg := prometheus.NewRegistry()
	collector := NewDBStatsCollector(db, "primary")
	require.NoError(t, reg.Register(collector))

	families, err := reg.Gather()
	require.NoError(t, err)

	byName := make(map[string]float64, len(families))
	for _, mf := range families {
		require.Len(t, mf.GetMetric(), 1)
		m := mf.GetMetric()[0]
		labels := m.GetLabel()
		require.Len(t, labels, 1)
		assert.Equal(t, "db", labels[0].GetName())
		assert.Equal(t, "primary", labels[0].GetValue())

		switch {
		case m.GetGauge() != nil:
			byName[mf.GetName()] = m.GetGauge().GetValue()
		case m.GetCounter() != nil:
			byName[mf.GetName()] = m.GetCounter().GetValue()
		}
	}

	expectedNames := []string{
		"app_db_pool_open_connections",
		"app_db_pool_in_use_connections",
		"app_db_pool_idle_connections",
		"app_db_pool_max_open_connections",
		"app_db_pool_wait_count_total",
		"app_db_pool_wait_duration_seconds_total",
	}
	for _, name := range expectedNames {
		_, ok := byName[name]
		assert.True(t, ok, "missing metric family %s", name)
	}

	assert.Equal(t, float64(expectedStats.OpenConnections), byName["app_db_pool_open_connections"])
	assert.Equal(t, float64(expectedStats.InUse), byName["app_db_pool_in_use_connections"])
	assert.Equal(t, float64(expectedStats.Idle), byName["app_db_pool_idle_connections"])
	assert.Equal(t, float64(15), byName["app_db_pool_max_open_connections"])
	assert.Equal(t, float64(expectedStats.WaitCount), byName["app_db_pool_wait_count_total"])
	assert.Equal(t, expectedStats.WaitDuration.Seconds(), byName["app_db_pool_wait_duration_seconds_total"])
}

func TestDBStatsCollector_NilDBAndDefaultName(t *testing.T) {
	reg := prometheus.NewRegistry()
	collector := NewDBStatsCollector(nil, "")
	require.NoError(t, reg.Register(collector))

	families, err := reg.Gather()
	require.NoError(t, err)
	assert.Empty(t, families)

	ch := make(chan *prometheus.Desc, 10)
	collector.Describe(ch)
	close(ch)
	assert.Len(t, ch, 6)
}

func TestRealtimeStatsCollector(t *testing.T) {
	sseCount := 7
	wsCount := 12

	reg := prometheus.NewRegistry()
	collector := NewRealtimeStatsCollector(
		func() int { return sseCount },
		func() int { return wsCount },
	)
	require.NoError(t, reg.Register(collector))

	assert.Equal(t, 2, testutil.CollectAndCount(collector))

	families, err := reg.Gather()
	require.NoError(t, err)

	byName := make(map[string]float64, len(families))
	for _, mf := range families {
		require.Len(t, mf.GetMetric(), 1)
		byName[mf.GetName()] = mf.GetMetric()[0].GetGauge().GetValue()
	}

	assert.Equal(t, float64(7), byName["app_sse_active_connections"])
	assert.Equal(t, float64(12), byName["app_ws_active_connections"])

	// Subsequent scrape reflects updated counts.
	sseCount = 3
	wsCount = 19
	families, err = reg.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		byName[mf.GetName()] = mf.GetMetric()[0].GetGauge().GetValue()
	}
	assert.Equal(t, float64(3), byName["app_sse_active_connections"])
	assert.Equal(t, float64(19), byName["app_ws_active_connections"])
}

func TestRealtimeStatsCollector_NilCallbacks(t *testing.T) {
	reg := prometheus.NewRegistry()
	collector := NewRealtimeStatsCollector(nil, nil)
	require.NoError(t, reg.Register(collector))

	families, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, families, 2)
	for _, mf := range families {
		require.Len(t, mf.GetMetric(), 1)
		assert.Equal(t, float64(0), mf.GetMetric()[0].GetGauge().GetValue())
	}
}

func TestRecordWebhookDelivery_BoundedOutcomes(t *testing.T) {
	beforeSuccess := testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("success"))
	beforeRetryable := testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("retryable_error"))
	beforePermanent := testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("permanent_error"))
	beforeError := testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("error"))

	RecordWebhookDelivery("success", 50*time.Millisecond)
	RecordWebhookDelivery("retryable_error", 100*time.Millisecond)
	RecordWebhookDelivery("permanent_error", 20*time.Millisecond)
	RecordWebhookDelivery("error", 10*time.Millisecond)
	RecordWebhookDelivery("unbounded_raw_error_123", 15*time.Millisecond)

	assert.Equal(t, beforeSuccess+1, testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("success")))
	assert.Equal(t, beforeRetryable+1, testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("retryable_error")))
	assert.Equal(t, beforePermanent+1, testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("permanent_error")))
	assert.Equal(t, beforeError+2, testutil.ToFloat64(WebhookDeliveriesTotal.WithLabelValues("error")))

	// Ensure unbounded outcome was never registered as a label value.
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() == "app_webhook_deliveries_total" || mf.GetName() == "app_webhook_delivery_duration_seconds" {
			for _, m := range mf.GetMetric() {
				for _, l := range m.GetLabel() {
					assert.NotEqual(t, "unbounded_raw_error_123", l.GetValue())
				}
			}
		}
	}
}

func TestRecordEmailDelivery_BoundedOutcomes(t *testing.T) {
	beforeSuccess := testutil.ToFloat64(EmailDeliveriesTotal.WithLabelValues("success"))
	beforeError := testutil.ToFloat64(EmailDeliveriesTotal.WithLabelValues("error"))

	RecordEmailDelivery("success", 120*time.Millisecond)
	RecordEmailDelivery("error", 40*time.Millisecond)
	RecordEmailDelivery("smtp_550_unknown", 30*time.Millisecond)

	assert.Equal(t, beforeSuccess+1, testutil.ToFloat64(EmailDeliveriesTotal.WithLabelValues("success")))
	assert.Equal(t, beforeError+2, testutil.ToFloat64(EmailDeliveriesTotal.WithLabelValues("error")))

	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	for _, mf := range families {
		if mf.GetName() == "app_email_deliveries_total" || mf.GetName() == "app_email_delivery_duration_seconds" {
			for _, m := range mf.GetMetric() {
				for _, l := range m.GetLabel() {
					assert.NotEqual(t, "smtp_550_unknown", l.GetValue())
				}
			}
		}
	}
}

func TestWSSendDropsAndAuditOutboxPending(t *testing.T) {
	beforeDrops := testutil.ToFloat64(WSSendDropsTotal)
	IncWSSendDrop()
	IncWSSendDrop()
	assert.Equal(t, beforeDrops+2, testutil.ToFloat64(WSSendDropsTotal))

	SetAuditOutboxPending(42)
	assert.Equal(t, float64(42), testutil.ToFloat64(AuditOutboxPending))
	SetAuditOutboxPending(0)
	assert.Equal(t, float64(0), testutil.ToFloat64(AuditOutboxPending))
}
