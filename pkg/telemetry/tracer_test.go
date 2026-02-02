package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTracer(t *testing.T) {
	// This might try to connect to localhost:4317 but should not fail hard
	shutdown, err := InitTracer("test-service", "localhost:4317")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Verify global tracer provider is set
	tp := otel.GetTracerProvider()
	assert.NotNil(t, tp)

	// Clean up
	if shutdown != nil {
		_ = shutdown(context.Background())
	}
}

func TestStartSpan(t *testing.T) {
	// Setup
	ctx := context.Background()
	spanName := "test-span"

	// Execute
	newCtx, endSpan := StartSpan(ctx, spanName)

	// Assert
	assert.NotNil(t, newCtx)
	assert.NotNil(t, endSpan)
	assert.NotEqual(t, ctx, newCtx) // Context should be modified (contain span)

	// End span (should not panic)
	endSpan()
}
