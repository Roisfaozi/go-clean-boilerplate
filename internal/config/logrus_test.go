package config

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceContextHook_EnrichesLogEntry(t *testing.T) {
	hook := &TraceContextHook{}

	t.Run("Enriches RequestID and UserID from AuthContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constants.RequestIDKey, "req-123")
		ctx = authcontext.WithUserID(ctx, "user-456")

		entry := &logrus.Entry{
			Context: ctx,
			Data:    make(logrus.Fields),
		}

		err := hook.Fire(entry)
		assert.NoError(t, err)
		assert.Equal(t, "req-123", entry.Data["request_id"])
		assert.Equal(t, "user-456", entry.Data["user_id"])
		assert.Nil(t, entry.Data["trace_id"])
		assert.Nil(t, entry.Data["span_id"])
	})

	t.Run("Enriches TraceID and SpanID when valid OTEL span present", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		tracer := tp.Tracer("test")
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		entry := &logrus.Entry{
			Context: ctx,
			Data:    make(logrus.Fields),
		}

		err := hook.Fire(entry)
		assert.NoError(t, err)
		assert.NotEmpty(t, entry.Data["trace_id"])
		assert.NotEmpty(t, entry.Data["span_id"])
	})
}
