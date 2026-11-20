package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	headerWebhookDeliveryID        = "X-Webhook-Delivery-ID"
	webhookDedupeKeyPrefix         = "worker:webhook:dedupe:"
	defaultWebhookDedupeTTL        = 24 * time.Hour
	defaultWebhookDedupeMaxEntries = 10000
)

type WebhookHandler struct {
	repo      repository.WebhookRepository
	log       *logrus.Logger
	client    *http.Client
	rdb       *redis.Client
	dedupeTTL time.Duration

	mu        sync.Mutex
	dedupeMap map[string]time.Time
}

func NewWebhookHandler(repo repository.WebhookRepository, log *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		repo: repo,
		log:  log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		dedupeTTL: defaultWebhookDedupeTTL,
	}
}

func (h *WebhookHandler) SetRedisClient(rdb *redis.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rdb = rdb
}

func webhookDedupeKey(webhookID, eventID string) string {
	sum := sha256.Sum256([]byte(webhookID + ":" + eventID))
	return webhookDedupeKeyPrefix + hex.EncodeToString(sum[:])
}

func (h *WebhookHandler) claimDedupe(ctx context.Context, key string) (bool, error) {
	h.mu.Lock()
	rdb := h.rdb
	ttl := h.dedupeTTL
	if ttl <= 0 {
		ttl = defaultWebhookDedupeTTL
	}
	if rdb != nil {
		h.mu.Unlock()
		return rdb.SetNX(ctx, key, "1", ttl).Result()
	}
	defer h.mu.Unlock()

	now := time.Now()
	if h.dedupeMap == nil {
		h.dedupeMap = make(map[string]time.Time)
	}
	for k, exp := range h.dedupeMap {
		if !exp.After(now) {
			delete(h.dedupeMap, k)
		}
	}
	if exp, ok := h.dedupeMap[key]; ok && exp.After(now) {
		return false, nil
	}
	if len(h.dedupeMap) >= defaultWebhookDedupeMaxEntries {
		var oldestKey string
		var oldestExp time.Time
		first := true
		for k, exp := range h.dedupeMap {
			if first || exp.Before(oldestExp) {
				oldestKey = k
				oldestExp = exp
				first = false
			}
		}
		if !first {
			delete(h.dedupeMap, oldestKey)
		}
	}
	h.dedupeMap[key] = now.Add(ttl)
	return true, nil
}

func (h *WebhookHandler) releaseDedupe(ctx context.Context, key string) {
	if key == "" {
		return
	}
	h.mu.Lock()
	rdb := h.rdb
	if rdb != nil {
		h.mu.Unlock()
		unclaimCtx := ctx
		if unclaimCtx == nil || unclaimCtx.Err() != nil {
			unclaimCtx = context.Background()
		}
		if err := rdb.Del(unclaimCtx, key).Err(); err != nil && h.log != nil {
			h.log.WithContext(unclaimCtx).WithError(err).Warnf("Failed to release webhook dedupe key %s", key)
		}
		return
	}
	delete(h.dedupeMap, key)
	h.mu.Unlock()
}

func (h *WebhookHandler) ProcessTaskWebhookTrigger(ctx context.Context, t *asynq.Task) error {
	var payload tasks.WebhookTriggerPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	eventID := strings.TrimSpace(payload.EventID)
	var dedupeKey string
	if eventID != "" {
		dedupeKey = webhookDedupeKey(payload.WebhookID, eventID)
		claimed, err := h.claimDedupe(ctx, dedupeKey)
		if err != nil {
			return fmt.Errorf("failed to claim webhook dedupe key: %w", err)
		}
		if !claimed {
			if h.log != nil {
				h.log.WithContext(ctx).Infof("Skipping duplicate webhook delivery webhook_id=%s event_id=%s", payload.WebhookID, eventID)
			}
			return nil
		}
	}

	start := time.Now()
	outcome := "error"
	defer func() {
		telemetry.RecordWebhookDelivery(outcome, time.Since(start))
	}()

	// Create signature
	h_mac := hmac.New(sha256.New, []byte(payload.Secret))
	h_mac.Write([]byte(payload.Payload))
	signature := hex.EncodeToString(h_mac.Sum(nil))

	timestamp := fmt.Sprintf("%d", start.UnixMilli())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, payload.URL, bytes.NewBuffer([]byte(payload.Payload)))
	if err != nil {
		outcome = "permanent_error"
		h.log.WithError(err).Error("Failed to create webhook request")
		return err
	}

	req.Header.Set(headerContentType, headerValueApplicationJSON)
	req.Header.Set(headerWebhookSignature, signature)
	req.Header.Set(headerWebhookEvent, payload.EventType)
	req.Header.Set(headerWebhookID, payload.WebhookID)
	req.Header.Set(headerWebhookTimestamp, timestamp)
	if eventID != "" {
		req.Header.Set(headerWebhookDeliveryID, eventID)
	}

	resp, err := h.client.Do(req)

	executionTime := time.Since(start).Milliseconds()

	logEntry := &entity.WebhookLog{
		ID:            uuid.New().String(),
		WebhookID:     payload.WebhookID,
		EventType:     payload.EventType,
		Payload:       payload.Payload,
		ExecutionTime: executionTime,
		CreatedAt:     time.Now().UnixMilli(),
	}

	if err != nil {
		h.releaseDedupe(ctx, dedupeKey)
		outcome = "retryable_error"
		logEntry.ErrorMessage = err.Error()
		if logErr := h.saveWebhookLog(ctx, logEntry); logErr != nil {
			return fmt.Errorf("webhook request failed: %w; failed to save webhook log: %v", err, logErr)
		}
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.log.WithError(err).Warn("Failed to close webhook response body")
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	logEntry.ResponseStatusCode = resp.StatusCode
	logEntry.ResponseBody = string(body)

	if resp.StatusCode >= 400 {
		logEntry.ErrorMessage = fmt.Sprintf("received status code %d", resp.StatusCode)
	}

	if err := h.saveWebhookLog(ctx, logEntry); err != nil {
		if resp.StatusCode >= 500 {
			h.releaseDedupe(ctx, dedupeKey)
			outcome = "retryable_error"
			return fmt.Errorf("upstream server error: %d; failed to save webhook log: %w", resp.StatusCode, err)
		}
	}

	if resp.StatusCode >= 500 {
		h.releaseDedupe(ctx, dedupeKey)
		outcome = "retryable_error"
		return fmt.Errorf("upstream server error: %d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		outcome = "permanent_error"
		return nil
	}

	outcome = "success"
	return nil
}

func (h *WebhookHandler) saveWebhookLog(ctx context.Context, logEntry *entity.WebhookLog) error {
	if err := h.repo.CreateLog(ctx, logEntry); err != nil {
		h.log.WithError(err).Error("Failed to save webhook log")
		return err
	}

	return nil
}
