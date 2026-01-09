//go:build integration
// +build integration

package scenarios

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScenario_WorkerIntegration_SendEmail verifies that the TaskDistributor
// correctly enqueues a "SendEmail" task into the Redis Asynq queue.
func TestScenario_WorkerIntegration_SendEmail(t *testing.T) {
	// 1. Setup Environment
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	// We don't need to clean DB, but we should probably clean Redis or use a unique queue?
	// For now, let's just proceed.

	// 2. Setup Task Distributor
	// We need to parse Redis Addr for Asynq Opts
	redisOpt := asynq.RedisClientOpt{
		Addr: env.RedisAddr,
	}
	distributor := worker.NewRedisTaskDistributor(redisOpt)

	// Ensure we close the client after test
	// Note: The interface doesn't expose Close(), but the concrete type does.
	// In real app, this is managed by FX or lifecycle.
	// distributor.(*worker.RedisTaskDistributor).Close() // If needed

	// 3. Setup Asynq Inspector to verify the queue state
	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	// 4. Define Payload and Execute
	payload := &tasks.SendEmailPayload{
		To:      "test@example.com",
		Subject: "Integration Test Email",
		Body:    "This is a test body.",
	}

	ctx := context.Background()
	err := distributor.DistributeTaskSendEmail(ctx, payload, asynq.MaxRetry(2))
	require.NoError(t, err)

	// 5. Verify Task is in Queue (Status: Pending)
	// Default queue is usually "default" unless specified
	// We give it a tiny moment to persist to Redis
	time.Sleep(100 * time.Millisecond)

	pendingTasks, err := inspector.ListPendingTasks("default", asynq.Page(1), asynq.PageSize(10))
	require.NoError(t, err)

	// Find our task
	var foundTask *asynq.TaskInfo
	for _, task := range pendingTasks {
		if task.Type == tasks.TypeSendEmail {
			foundTask = task
			break
		}
	}

	require.NotNil(t, foundTask, "SendEmail task not found in pending queue")

	// 6. Verify Payload
	var actualPayload tasks.SendEmailPayload
	err = json.Unmarshal(foundTask.Payload, &actualPayload)
	require.NoError(t, err)

	assert.Equal(t, payload.To, actualPayload.To)
	assert.Equal(t, payload.Subject, actualPayload.Subject)
	assert.Equal(t, payload.Body, actualPayload.Body)
	assert.Equal(t, 2, foundTask.MaxRetry) // Verify option was passed
}
