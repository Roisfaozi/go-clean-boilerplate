//go:build integration
// +build integration

package modules

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookIntegration_FullLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	// Setup components
	redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
	distributor := worker.NewRedisTaskDistributor(redisOpt)
	webhookRepo := repository.NewWebhookRepository(env.DB, env.Logger)
	webhookHandler := handlers.NewWebhookHandler(webhookRepo, env.Logger)
	cleanupHandler := handlers.NewCleanupTaskHandler(nil, nil, nil, env.Logger)
	processor := worker.NewRedisTaskProcessor(redisOpt, env.Logger, cleanupHandler, webhookHandler, nil, nil, worker.WorkerConfig{})
	env.StartWorker(processor)

	validate := validator.New()
	webhookUC := usecase.NewWebhookUseCase(webhookRepo, distributor, env.Logger, validate)
	ctx := context.Background()

	t.Run("Positive: Successful Trigger and Delivery", func(t *testing.T) {
		received := make(chan bool, 1)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received <- true
			w.WriteHeader(http.StatusOK)
		}))
		defer mockServer.Close()

		wh, err := webhookUC.Create(ctx, model.CreateWebhookRequest{
			Name: "Pos Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.created"}, Secret: "super-secret-key",
		})
		require.NoError(t, err)
		require.NotNil(t, wh)

		err = webhookUC.Trigger(ctx, model.TriggerWebhookRequest{
			OrganizationID: "org-1", EventType: "user.created", Payload: map[string]string{"data": "ok"},
		})
		require.NoError(t, err)

		select {
		case <-received:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Webhook not delivered")
		}

		// Verify Log
		var logs []any
		require.Eventually(t, func() bool {
			logs, _ = webhookUC.FindLogs(ctx, wh.ID, "org-1", 1, 0)
			return len(logs) > 0
		}, 5*time.Second, 200*time.Millisecond)
		require.NotEmpty(t, logs)
	})

	t.Run("Negative: Target Server Error 500", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		wh, err := webhookUC.Create(ctx, model.CreateWebhookRequest{
			Name: "Neg Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.error"}, Secret: "super-secret-key",
		})
		require.NoError(t, err)
		require.NotNil(t, wh)

		_ = webhookUC.Trigger(ctx, model.TriggerWebhookRequest{
			OrganizationID: "org-1", EventType: "user.error", Payload: map[string]string{"err": "fail"},
		})

		// Wait for log
		require.Eventually(t, func() bool {
			logs, _ := webhookUC.FindLogs(ctx, wh.ID, "org-1", 1, 0)
			return len(logs) > 0
		}, 5*time.Second, 200*time.Millisecond)
	})

	t.Run("Edge: Inactive Webhook should not trigger", func(t *testing.T) {
		callCount := int32(0)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
		}))
		defer mockServer.Close()

		wh, err := webhookUC.Create(ctx, model.CreateWebhookRequest{
			Name: "Inactive Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.inactive"}, Secret: "super-secret-key",
		})
		require.NoError(t, err)

		// Deactivate
		active := false
		_, err = webhookUC.Update(ctx, wh.ID, "org-1", model.UpdateWebhookRequest{IsActive: &active})
		require.NoError(t, err)

		_ = webhookUC.Trigger(ctx, model.TriggerWebhookRequest{
			OrganizationID: "org-1", EventType: "user.inactive", Payload: map[string]string{"id": "1"},
		})

		time.Sleep(1 * time.Second)
		assert.Equal(t, int32(0), atomic.LoadInt32(&callCount))
	})

	t.Run("Security: HMAC Signature Verification", func(t *testing.T) {
		secret := "secure-key-123"
		payload := map[string]string{"secure": "data"}
		payloadJSON, _ := json.Marshal(payload)

		sigChan := make(chan string, 1)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sigChan <- r.Header.Get("X-Webhook-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer mockServer.Close()

		_, err := webhookUC.Create(ctx, model.CreateWebhookRequest{
			Name: "Sec Test", OrganizationID: "org-sec", URL: mockServer.URL, Events: []string{"secure.event"}, Secret: secret,
		})
		require.NoError(t, err)

		_ = webhookUC.Trigger(ctx, model.TriggerWebhookRequest{
			OrganizationID: "org-sec", EventType: "secure.event", Payload: payload,
		})

		select {
		case receivedSignature := <-sigChan:
			h := hmac.New(sha256.New, []byte(secret))
			h.Write(payloadJSON)
			expectedSignature := hex.EncodeToString(h.Sum(nil))
			assert.Equal(t, expectedSignature, receivedSignature)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}
	})
}
