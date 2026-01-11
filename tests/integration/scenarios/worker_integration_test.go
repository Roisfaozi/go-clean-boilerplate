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

func TestScenario_WorkerIntegration_SendEmail(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	redisOpt := asynq.RedisClientOpt{
		Addr: env.RedisAddr,
	}
	distributor := worker.NewRedisTaskDistributor(redisOpt)

	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	payload := &tasks.SendEmailPayload{
		To:      "test@example.com",
		Subject: "Integration Test Email",
		Body:    "This is a test body.",
	}

	ctx := context.Background()
	err := distributor.DistributeTaskSendEmail(ctx, payload, asynq.MaxRetry(2))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	pendingTasks, err := inspector.ListPendingTasks("default", asynq.Page(1), asynq.PageSize(10))
	require.NoError(t, err)

	var foundTask *asynq.TaskInfo
	for _, task := range pendingTasks {
		if task.Type == tasks.TypeSendEmail {
			foundTask = task
			break
		}
	}

	require.NotNil(t, foundTask, "SendEmail task not found in pending queue")

	var actualPayload tasks.SendEmailPayload
	err = json.Unmarshal(foundTask.Payload, &actualPayload)
	require.NoError(t, err)

	assert.Equal(t, payload.To, actualPayload.To)
	assert.Equal(t, payload.Subject, actualPayload.Subject)
	assert.Equal(t, payload.Body, actualPayload.Body)
	assert.Equal(t, 2, foundTask.MaxRetry)
}
