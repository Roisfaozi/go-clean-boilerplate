package usecase

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisProcessedUploadKeyPrefix = "tus:processed:"

// RedisProcessedUploadTracker persists processed TUS upload IDs in Redis so
// upload completion callbacks are deduplicated across application instances.
type RedisProcessedUploadTracker struct {
	client *redis.Client
}

var _ ProcessedUploadTracker = (*RedisProcessedUploadTracker)(nil)

func NewRedisProcessedUploadTracker(client *redis.Client) *RedisProcessedUploadTracker {
	return &RedisProcessedUploadTracker{client: client}
}

func (t *RedisProcessedUploadTracker) MarkProcessedIfAbsent(ctx context.Context, uploadID string, ttl time.Duration) (bool, error) {
	if t == nil || t.client == nil || uploadID == "" {
		return true, nil
	}
	if ttl <= 0 {
		ttl = defaultProcessedUploadTTL
	}
	return t.client.SetNX(ctx, redisProcessedUploadKeyPrefix+uploadID, "1", ttl).Result()
}

func (t *RedisProcessedUploadTracker) UnmarkProcessed(ctx context.Context, uploadID string) error {
	if t == nil || t.client == nil || uploadID == "" {
		return nil
	}
	return t.client.Del(ctx, redisProcessedUploadKeyPrefix+uploadID).Err()
}
