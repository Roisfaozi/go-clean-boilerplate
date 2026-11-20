package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhookHandler_ProcessTaskWebhookTrigger(t *testing.T) {
	repo := mocks.NewMockWebhookRepository(t)
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
		repo := mocks.NewMockWebhookRepository(t)
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
		repo := mocks.NewMockWebhookRepository(t)
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

func TestWebhookHandler_DedupeAndDeliveryID(t *testing.T) {
	t.Run("in-memory dedupe suppresses duplicate EventID and sends X-Webhook-Delivery-ID", func(t *testing.T) {
		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)
		handler := NewWebhookHandler(repo, log)

		var httpCalls atomic.Int32
		var seenDeliveryID string
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			httpCalls.Add(1)
			seenDeliveryID = r.Header.Get(headerWebhookDeliveryID)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})}

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-dedupe-1",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.created",
			Payload:   `{"id":"user-1"}`,
			EventID:   "evt-100",
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Once()

		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))

		assert.Equal(t, int32(1), httpCalls.Load(), "duplicate EventID must not trigger 2nd HTTP call")
		assert.Equal(t, "evt-100", seenDeliveryID)
		repo.AssertExpectations(t)
	})

	t.Run("redis dedupe suppresses duplicate EventID across instances", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		defer func() { _ = rdb.Close() }()

		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)

		var httpCalls atomic.Int32
		client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			httpCalls.Add(1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})}

		handlerA := NewWebhookHandler(repo, log)
		handlerA.SetRedisClient(rdb)
		handlerA.client = client

		handlerB := NewWebhookHandler(repo, log)
		handlerB.SetRedisClient(rdb)
		handlerB.client = client

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-redis-1",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.created",
			Payload:   `{"id":"user-1"}`,
			EventID:   "evt-redis-1",
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Once()

		require.NoError(t, handlerA.ProcessTaskWebhookTrigger(context.Background(), task))
		require.NoError(t, handlerB.ProcessTaskWebhookTrigger(context.Background(), task))

		assert.Equal(t, int32(1), httpCalls.Load())
		assert.True(t, mr.Exists(webhookDedupeKey("wh-redis-1", "evt-redis-1")))
		repo.AssertExpectations(t)
	})

	t.Run("5xx unmarks claim so retry succeeds", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		defer func() { _ = rdb.Close() }()

		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)
		handler := NewWebhookHandler(repo, log)
		handler.SetRedisClient(rdb)

		var httpCalls atomic.Int32
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if httpCalls.Add(1) == 1 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader(`{"err":"boom"}`)),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})}

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-5xx",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.updated",
			Payload:   `{"id":"user-1"}`,
			EventID:   "evt-5xx-1",
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Twice()

		// 1st attempt returns 500 -> error returned and claim released.
		err = handler.ProcessTaskWebhookTrigger(context.Background(), task)
		require.Error(t, err)
		assert.False(t, mr.Exists(webhookDedupeKey("wh-5xx", "evt-5xx-1")))

		// 2nd attempt retries -> succeeds and persists claim.
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		assert.True(t, mr.Exists(webhookDedupeKey("wh-5xx", "evt-5xx-1")))

		// 3rd attempt (duplicate) -> suppressed without HTTP call or log.
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		assert.Equal(t, int32(2), httpCalls.Load())
		repo.AssertExpectations(t)
	})

	t.Run("network error unmarks claim so retry succeeds", func(t *testing.T) {
		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)
		handler := NewWebhookHandler(repo, log)

		var httpCalls atomic.Int32
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if httpCalls.Add(1) == 1 {
				return nil, errors.New("dial tcp timeout")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})}

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-net",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.updated",
			Payload:   `{"id":"user-1"}`,
			EventID:   "evt-net-1",
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Twice()

		require.Error(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		assert.Equal(t, int32(2), httpCalls.Load())
		repo.AssertExpectations(t)
	})

	t.Run("4xx permanent error keeps claim", func(t *testing.T) {
		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)
		handler := NewWebhookHandler(repo, log)

		var httpCalls atomic.Int32
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			httpCalls.Add(1)
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"err":"bad request"}`)),
			}, nil
		})}

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-4xx",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.updated",
			Payload:   `{"id":"user-1"}`,
			EventID:   "evt-4xx-1",
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Once()

		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		assert.Equal(t, int32(1), httpCalls.Load(), "4xx permanent error must stay claimed")
		repo.AssertExpectations(t)
	})

	t.Run("empty EventID preserves backwards-compatible behavior without false blocking", func(t *testing.T) {
		repo := mocks.NewMockWebhookRepository(t)
		log := logrus.New()
		log.SetOutput(io.Discard)
		handler := NewWebhookHandler(repo, log)

		var httpCalls atomic.Int32
		handler.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			httpCalls.Add(1)
			assert.Empty(t, r.Header.Get(headerWebhookDeliveryID))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})}

		task, err := tasks.NewWebhookTriggerTask(tasks.WebhookTriggerPayload{
			WebhookID: "wh-legacy",
			URL:       "https://example.test/webhook",
			Secret:    "secret",
			EventType: "user.updated",
			Payload:   `{"id":"user-1"}`,
		})
		require.NoError(t, err)

		repo.On("CreateLog", mock.Anything, mock.Anything).Return(nil).Twice()

		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		require.NoError(t, handler.ProcessTaskWebhookTrigger(context.Background(), task))
		assert.Equal(t, int32(2), httpCalls.Load())
		repo.AssertExpectations(t)
	})
}
