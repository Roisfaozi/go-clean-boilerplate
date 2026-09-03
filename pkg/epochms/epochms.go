package epochms

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Time represents epoch milliseconds as int64.
type Time int64

func Now() Time {
	return Time(time.Now().UnixMilli())
}

func From(t time.Time) Time {
	return Time(t.UnixMilli())
}

func (t Time) Std() time.Time {
	return time.UnixMilli(int64(t)).UTC()
}

func (t Time) Add(d time.Duration) Time {
	return Time(int64(t) + d.Milliseconds())
}

func (t Time) Sub(u Time) time.Duration {
	return time.Duration(int64(t)-int64(u)) * time.Millisecond
}

func (t Time) Before(u Time) bool {
	return t < u
}

func (t Time) After(u Time) bool {
	return t > u
}

func (t Time) In(tz *time.Location) time.Time {
	if tz == nil {
		panic("epochms: nil timezone location")
	}
	return time.UnixMilli(int64(t)).In(tz)
}

func (t Time) RFC3339(tz *time.Location) string {
	return t.In(tz).Format(time.RFC3339)
}

func (t Time) BusinessDate(tz *time.Location) string {
	return t.In(tz).Format("2006-01-02")
}

func StartOfDay(dayStr string, tz *time.Location) (Time, error) {
	if tz == nil {
		panic("epochms: nil timezone location")
	}
	parsed, err := time.ParseInLocation("2006-01-02", dayStr, tz)
	if err != nil {
		return 0, fmt.Errorf("invalid business date format: %w", err)
	}
	start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, tz)
	return From(start), nil
}

func EndOfDay(dayStr string, tz *time.Location) (Time, error) {
	if tz == nil {
		panic("epochms: nil timezone location")
	}
	parsed, err := time.ParseInLocation("2006-01-02", dayStr, tz)
	if err != nil {
		return 0, fmt.Errorf("invalid business date format: %w", err)
	}
	end := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999000000, tz)
	return From(end), nil
}

// Database driver interfaces
func (t Time) Value() (driver.Value, error) {
	return int64(t), nil
}

func (t *Time) Scan(value interface{}) error {
	if value == nil {
		*t = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*t = Time(v)
	case int:
		*t = Time(v)
	case float64:
		*t = Time(int64(v))
	case []byte:
		var n int64
		if _, err := fmt.Sscanf(string(v), "%d", &n); err != nil {
			return fmt.Errorf("epochms: cannot scan %v into Time", value)
		}
		*t = Time(n)
	default:
		return fmt.Errorf("epochms: unsupported scan type %T", value)
	}
	return nil
}

// JSON interfaces
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(t))
}

func (t *Time) Unmarshaler(data []byte) error {
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*t = Time(n)
	return nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	return t.Unmarshaler(data)
}
