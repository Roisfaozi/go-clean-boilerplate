package epochms_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/epochms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpochMs(t *testing.T) {
	wib, err := time.LoadLocation("Asia/Jakarta")
	require.NoError(t, err)
	wit, err := time.LoadLocation("Asia/Jayapura")
	require.NoError(t, err)

	// Instant: 2026-09-01 17:30:00 UTC -> 2026-09-02 00:30:00 WIB, 2026-09-02 02:30:00 WIT
	stdTime := time.Date(2026, 9, 1, 17, 30, 0, 0, time.UTC)
	ep := epochms.From(stdTime)

	t.Run("WIB vs WIT business date difference", func(t *testing.T) {
		// 16:30 UTC -> 2026-09-01 23:30 WIB (Sept 1), 2026-09-02 01:30 WIT (Sept 2)
		tDiff := time.Date(2026, 9, 1, 16, 30, 0, 0, time.UTC)
		epDiff := epochms.From(tDiff)
		assert.Equal(t, "2026-09-01", epDiff.BusinessDate(wib))
		assert.Equal(t, "2026-09-02", epDiff.BusinessDate(wit))
	})

	t.Run("Nil timezone panic", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = ep.In(nil)
		})
		assert.Panics(t, func() {
			_, _ = epochms.StartOfDay("2026-09-01", nil)
		})
		assert.Panics(t, func() {
			_, _ = epochms.EndOfDay("2026-09-01", nil)
		})
	})

	t.Run("DB Round Trip", func(t *testing.T) {
		val, err := ep.Value()
		require.NoError(t, err)
		assert.Equal(t, stdTime.UnixMilli(), val)

		var scanned epochms.Time
		err = scanned.Scan(val)
		require.NoError(t, err)
		assert.Equal(t, ep, scanned)
	})

	t.Run("JSON Marshaling emits number", func(t *testing.T) {
		bytes, err := json.Marshal(ep)
		require.NoError(t, err)
		assert.Equal(t, []byte(jsonNumberString(stdTime.UnixMilli())), bytes)

		var unmarshaled epochms.Time
		err = json.Unmarshal(bytes, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, ep, unmarshaled)
	})

	t.Run("Milliseconds guard check", func(t *testing.T) {
		now := epochms.Now()
		// Epoch milli for year 2026 is > 1.7e12, epoch sec is ~1.7e9
		assert.True(t, int64(now) > 1_000_000_000_000, "value must be in milliseconds, got: %d", now)
	})
}

func jsonNumberString(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
