package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
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
	handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Webhook-Signature"))
		assert.Equal(t, "user.created", r.Header.Get("X-Webhook-Event"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, nil
	})}

	payload := tasks.WebhookTriggerPayload{
		WebhookID: "wh-1",
		URL:       "https://example.test/webhook",
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

func TestWebhookHandler_ProcessTaskWebhookTrigger_FailureLogPersistence(t *testing.T) {
	t.Run("network error returns log persistence failure", func(t *testing.T) {
		repo := new(mocks.MockWebhookRepository)
		log := logrus.New()
		handler := NewWebhookHandler(repo, log)
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		})}

		payload := tasks.WebhookTriggerPayload{
			WebhookID: "wh-network-fail",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.created",
			Payload:   `{"id":"user-1"}`,
		}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeWebhookTrigger, payloadBytes)

		repo.On("CreateLog", mock.Anything, mock.MatchedBy(func(logEntry interface{}) bool {
			entry, ok := logEntry.(*entity.WebhookLog)
			return ok && entry.WebhookID == payload.WebhookID && entry.ErrorMessage != ""
		})).Return(errors.New("log store down"))

		err := handler.ProcessTaskWebhookTrigger(context.Background(), task)

		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "webhook request failed"))
		assert.True(t, strings.Contains(err.Error(), "failed to save webhook log"))
		repo.AssertExpectations(t)
	})

	t.Run("server error includes log persistence failure", func(t *testing.T) {
		repo := new(mocks.MockWebhookRepository)
		log := logrus.New()
		handler := NewWebhookHandler(repo, log)
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"status":"error"}`)),
			}, nil
		})}

		payload := tasks.WebhookTriggerPayload{
			WebhookID: "wh-server-fail",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.created",
			Payload:   `{"id":"user-1"}`,
		}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeWebhookTrigger, payloadBytes)

		repo.On("CreateLog", mock.Anything, mock.MatchedBy(func(logEntry interface{}) bool {
			entry, ok := logEntry.(*entity.WebhookLog)
			return ok && entry.WebhookID == payload.WebhookID && entry.ResponseStatusCode == http.StatusInternalServerError
		})).Return(errors.New("log store down"))

		err := handler.ProcessTaskWebhookTrigger(context.Background(), task)

		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "upstream server error: 500"))
		assert.True(t, strings.Contains(err.Error(), "failed to save webhook log"))
		repo.AssertExpectations(t)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
