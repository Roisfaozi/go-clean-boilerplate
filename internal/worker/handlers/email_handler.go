package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

const (
	emailDedupeKeyPrefix         = "worker:email:dedupe:"
	defaultEmailDedupeTTL        = 24 * time.Hour
	defaultEmailDedupeMaxEntries = 10000
)

// SMTPConfig is defined locally/interface to avoid import cycle
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromSender string
	FromEmail  string
}

type EmailTaskHandler struct {
	logger    *logrus.Logger
	cfg       SMTPConfig
	rdb       *redis.Client
	dedupeTTL time.Duration
	sendFn    func(m *gomail.Message) error

	mu        sync.Mutex
	dedupeMap map[string]time.Time
}

func NewEmailTaskHandler(logger *logrus.Logger, cfg SMTPConfig) *EmailTaskHandler {
	return &EmailTaskHandler{
		logger:    logger,
		cfg:       cfg,
		dedupeTTL: defaultEmailDedupeTTL,
	}
}

func NewEmailTaskHandlerWithDedupe(logger *logrus.Logger, cfg SMTPConfig, rdb *redis.Client) *EmailTaskHandler {
	h := NewEmailTaskHandler(logger, cfg)
	h.rdb = rdb
	return h
}

func (h *EmailTaskHandler) SetRedisClient(rdb *redis.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rdb = rdb
}

func (h *EmailTaskHandler) SetSendFunc(fn func(m *gomail.Message) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sendFn = fn
}

func emailDedupeKey(payload *tasks.SendEmailPayload) string {
	if key := strings.TrimSpace(payload.IdempotencyKey); key != "" {
		return emailDedupeKeyPrefix + key
	}
	sum := sha256.Sum256([]byte(
		strings.ToLower(strings.TrimSpace(payload.To)) + "\x00" +
			payload.Subject + "\x00" +
			payload.Body,
	))
	return emailDedupeKeyPrefix + hex.EncodeToString(sum[:])
}

func (h *EmailTaskHandler) claimDedupe(ctx context.Context, key string) (bool, error) {
	h.mu.Lock()
	rdb := h.rdb
	ttl := h.dedupeTTL
	if ttl <= 0 {
		ttl = defaultEmailDedupeTTL
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
	if len(h.dedupeMap) >= defaultEmailDedupeMaxEntries {
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

func (h *EmailTaskHandler) releaseDedupe(ctx context.Context, key string) {
	h.mu.Lock()
	rdb := h.rdb
	if rdb != nil {
		h.mu.Unlock()
		unclaimCtx := ctx
		if unclaimCtx == nil || unclaimCtx.Err() != nil {
			unclaimCtx = context.Background()
		}
		if err := rdb.Del(unclaimCtx, key).Err(); err != nil && h.logger != nil {
			h.logger.WithContext(unclaimCtx).WithError(err).Warnf("Failed to release email dedupe key %s", key)
		}
		return
	}
	delete(h.dedupeMap, key)
	h.mu.Unlock()
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	if len(name) > 2 {
		name = name[:2] + "***"
	} else {
		name = "***"
	}
	return name + "@" + parts[1]
}

func (h *EmailTaskHandler) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	maskedTo := MaskEmail(payload.To)
	dedupeKey := emailDedupeKey(&payload)
	claimed, err := h.claimDedupe(ctx, dedupeKey)
	if err != nil {
		return fmt.Errorf("failed to claim email dedupe key: %w", err)
	}
	if !claimed {
		if h.logger != nil {
			h.logger.WithContext(ctx).Infof("Skipping duplicate email to %s (key=%s)", maskedTo, dedupeKey)
		}
		return nil
	}

	start := time.Now()
	if h.logger != nil {
		h.logger.WithContext(ctx).Infof("Sending real email to %s via %s:%d", maskedTo, h.cfg.Host, h.cfg.Port)
	}

	m := gomail.NewMessage()
	m.SetHeader(headerEmailFrom, fmt.Sprintf("%s <%s>", h.cfg.FromSender, h.cfg.FromEmail))
	m.SetHeader(headerEmailTo, payload.To)
	m.SetHeader(headerEmailSubject, payload.Subject)
	m.SetBody(headerValueTextHTML, payload.Body)

	h.mu.Lock()
	sendFn := h.sendFn
	h.mu.Unlock()

	var sendErr error
	if sendFn != nil {
		sendErr = sendFn(m)
	} else {
		d := gomail.NewDialer(h.cfg.Host, h.cfg.Port, h.cfg.Username, h.cfg.Password)
		sendErr = d.DialAndSend(m)
	}

	if sendErr != nil {
		h.releaseDedupe(ctx, dedupeKey)
		telemetry.RecordEmailDelivery("error", time.Since(start))
		if h.logger != nil {
			h.logger.WithContext(ctx).Errorf("Failed to send email to %s: %v", maskedTo, sendErr)
		}
		return fmt.Errorf("failed to send email: %w", sendErr)
	}

	telemetry.RecordEmailDelivery("success", time.Since(start))
	if h.logger != nil {
		h.logger.WithContext(ctx).Infof("SUCCESS: Email sent to %s", maskedTo)
	}
	return nil
}
