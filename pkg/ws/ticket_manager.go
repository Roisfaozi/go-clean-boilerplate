package ws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultTicketTTL          = 30 * time.Second
	defaultRecentReplayWindow = 5 * time.Second
)

// TicketManager defines the interface for managing one-time WebSocket authentication tickets.
type TicketManager interface {
	// CreateTicket generates a new one-time ticket for the given user context.
	CreateTicket(ctx context.Context, userID, orgID, sessionID, role, username string) (string, error)
	// ValidateTicket validates a ticket and returns the associated user context.
	// The ticket is invalidated immediately after use.
	ValidateTicket(ctx context.Context, ticket string) (*UserContext, error)
}

// UserContext holds the user information associated with a ticket.
type UserContext struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	SessionID      string `json:"session_id"`
	Role           string `json:"role"`
	Username       string `json:"username"`
}

// RedisTicketManager implements TicketManager using Redis.
type RedisTicketManager struct {
	redisClient  *redis.Client
	ttl          time.Duration
	recentWindow time.Duration
}

// NewRedisTicketManager creates a new instance of RedisTicketManager.
func NewRedisTicketManager(redisClient *redis.Client, ttl time.Duration) *RedisTicketManager {
	if ttl <= 0 {
		ttl = defaultTicketTTL
	}
	recentWindow := defaultRecentReplayWindow
	if ttl < recentWindow {
		recentWindow = ttl
	}
	return &RedisTicketManager{
		redisClient:  redisClient,
		ttl:          ttl,
		recentWindow: recentWindow,
	}
}

// SetRecentReplayWindow configures the deduplication window for replaying an
// unused ticket for identical user context. A non-positive duration disables replay.
func (m *RedisTicketManager) SetRecentReplayWindow(d time.Duration) {
	if d <= 0 {
		m.recentWindow = 0
		return
	}
	if d > m.ttl {
		d = m.ttl
	}
	m.recentWindow = d
}

// CreateTicket generates a new ticket and stores it in Redis.
// When a recent replay window is active and an unused ticket for the same user
// context was issued within that window, the existing unused ticket is returned.
func (m *RedisTicketManager) CreateTicket(ctx context.Context, userID, orgID, sessionID, role, username string) (string, error) {
	userCtx := UserContext{
		UserID:         userID,
		OrganizationID: orgID,
		SessionID:      sessionID,
		Role:           role,
		Username:       username,
	}

	data, err := json.Marshal(userCtx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user context: %w", err)
	}

	var recentKey string
	if m.recentWindow > 0 {
		recentKey = m.recentTicketKey(data)
		recentTicket, err := m.redisClient.Get(ctx, recentKey).Result()
		if err == nil && recentTicket != "" {
			if m.redisClient.Exists(ctx, m.ticketKey(recentTicket)).Val() > 0 {
				return recentTicket, nil
			}
		}
	}

	ticket := uuid.New().String()
	key := m.ticketKey(ticket)

	err = m.redisClient.Set(ctx, key, data, m.ttl).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store ticket in redis: %w", err)
	}

	if m.recentWindow > 0 {
		_ = m.redisClient.Set(ctx, recentKey, ticket, m.recentWindow).Err()
	}

	return ticket, nil
}

// ValidateTicket retrieves and deletes the ticket from Redis (One-Time Use).
func (m *RedisTicketManager) ValidateTicket(ctx context.Context, ticket string) (*UserContext, error) {
	key := m.ticketKey(ticket)

	// Use GetDel to ensure atomicity (requires Redis 6.2+)
	// If GetDel is not available/supported by the client version or redis version,
	// we fall back to Get + Del pipeline, but standard go-redis supports GetDel.
	val, err := m.redisClient.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid or expired ticket")
		}
		return nil, fmt.Errorf("failed to validate ticket: %w", err)
	}

	var userCtx UserContext
	if err := json.Unmarshal([]byte(val), &userCtx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user context: %w", err)
	}

	return &userCtx, nil
}

func (m *RedisTicketManager) ticketKey(ticket string) string {
	return fmt.Sprintf("ws:ticket:%s", ticket)
}

func (m *RedisTicketManager) recentTicketKey(userCtxPayload []byte) string {
	sum := sha256.Sum256(userCtxPayload)
	return "ws:ticket:recent:" + hex.EncodeToString(sum[:])
}
