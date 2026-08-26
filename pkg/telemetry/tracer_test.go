package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitTracer_Success(t *testing.T) {
	cfg := Config{
		ServiceName:  "test-service",
		CollectorURL: "localhost:4317",
		Insecure:     true,
		SampleRatio:  0.5,
	}

	shutdown, err := InitTracer(cfg)
	require.NoError(t, err)
	require.NotNil(t, shutdown)

	err = shutdown(context.Background())
	assert.NoError(t, err)
}
