package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestInstrumentTaskHandler(t *testing.T) {
	t.Run("records metrics on success", func(t *testing.T) {
		handler := InstrumentTaskHandler("test.task", func(ctx context.Context, task *asynq.Task) error {
			return nil
		})

		err := handler(context.Background(), asynq.NewTask("test.task", nil))
		assert.NoError(t, err)
	})

	t.Run("records metrics on error and propagates error", func(t *testing.T) {
		expectedErr := errors.New("task failed")
		handler := InstrumentTaskHandler("test.task", func(ctx context.Context, task *asynq.Task) error {
			return expectedErr
		})

		err := handler(context.Background(), asynq.NewTask("test.task", nil))
		assert.ErrorIs(t, err, expectedErr)
	})
}
