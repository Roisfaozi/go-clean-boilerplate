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
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type WebhookHandler struct {
	repo repository.WebhookRepository
	log  *logrus.Logger
}

func NewWebhookHandler(repo repository.WebhookRepository, log *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		repo: repo,
		log:  log,
	}
}

func (h *WebhookHandler) ProcessTaskWebhookTrigger(ctx context.Context, t *asynq.Task) error {
	var payload tasks.WebhookTriggerPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	start := time.Now()

	// Create signature
	h_mac := hmac.New(sha256.New, []byte(payload.Secret))
	h_mac.Write([]byte(payload.Payload))
	signature := hex.EncodeToString(h_mac.Sum(nil))

	timestamp := fmt.Sprintf("%d", start.UnixMilli())

	req, err := http.NewRequestWithContext(ctx, "POST", payload.URL, bytes.NewBuffer([]byte(payload.Payload)))
	if err != nil {
		h.log.WithError(err).Error("Failed to create webhook request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", payload.EventType)
	req.Header.Set("X-Webhook-ID", payload.WebhookID)
	req.Header.Set("X-Webhook-Timestamp", timestamp)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)

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
		logEntry.ErrorMessage = err.Error()
		_ = h.repo.CreateLog(ctx, logEntry)
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			h.log.WithError(closeErr).Warn("Failed to close webhook response body")
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	logEntry.ResponseStatusCode = resp.StatusCode
	logEntry.ResponseBody = string(body)

	if resp.StatusCode >= 400 {
		logEntry.ErrorMessage = fmt.Sprintf("received status code %d", resp.StatusCode)
	}

	if err := h.repo.CreateLog(ctx, logEntry); err != nil {
		h.log.WithError(err).Error("Failed to save webhook log")
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("upstream server error: %d", resp.StatusCode)
	}

	return nil
}
