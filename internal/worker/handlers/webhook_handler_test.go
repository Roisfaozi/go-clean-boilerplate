package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWebhookHandler_ProcessTaskWebhookTrigger(t *testing.T) {
	repo := new(mocks.MockWebhookRepository)
	log := logrus.New()
	handler := NewWebhookHandler(repo, log)

	// Setup mock server to receive webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Webhook-Signature"))
		assert.Equal(t, "user.created", r.Header.Get("X-Webhook-Event"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	payload := tasks.WebhookTriggerPayload{
		WebhookID: "wh-1",
		URL:       server.URL,
		Secret:    "secret",
		EventType: "user.created",
		Payload:   `{"id":"user-1"}`,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask(tasks.TypeWebhookTrigger, payloadBytes)

	repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil)

	err := handler.ProcessTaskWebhookTrigger(context.Background(), task)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
