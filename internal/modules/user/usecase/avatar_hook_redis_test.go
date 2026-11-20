package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tus"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredisClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func TestRedisProcessedUploadTracker_CrossInstanceReplaySuppression(t *testing.T) {
	mr, rdb := setupMiniredisClient(t)

	trackerA := NewRedisProcessedUploadTracker(rdb)
	trackerB := NewRedisProcessedUploadTracker(rdb)

	ucA := &stubAvatarUserUseCase{}
	ucB := &stubAvatarUserUseCase{}

	hookA := &AvatarHook{
		UserUseCase:  ucA,
		ProcessedTTL: 10 * time.Minute,
		Tracker:      trackerA,
	}
	hookB := &AvatarHook{
		UserUseCase:  ucB,
		ProcessedTTL: 10 * time.Minute,
		Tracker:      trackerB,
	}

	event := tus.UploadEvent{
		UploadID: "upload-cross-instance-1",
		FileURL:  "https://cdn.example.com/avatars/upload-cross-instance-1.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	require.NoError(t, hookA.HandleUpload(context.Background(), event))
	require.NoError(t, hookB.HandleUpload(context.Background(), event))

	assert.Equal(t, int32(1), ucA.calls.Load())
	assert.Equal(t, int32(0), ucB.calls.Load(), "second instance must skip already processed uploadID in Redis")
	assert.True(t, mr.Exists(redisProcessedUploadKeyPrefix+event.UploadID))
	assert.Greater(t, mr.TTL(redisProcessedUploadKeyPrefix+event.UploadID), time.Duration(0))
}

func TestRedisProcessedUploadTracker_UnmarkProcessedRecoveryAfterFailure(t *testing.T) {
	mr, rdb := setupMiniredisClient(t)

	ucA := &stubAvatarUserUseCase{
		setAvatarFn: func(_ context.Context, _, _ string) error {
			return errors.New("temporary db outage")
		},
	}
	ucB := &stubAvatarUserUseCase{}

	hookA := &AvatarHook{
		UserUseCase:  ucA,
		ProcessedTTL: 10 * time.Minute,
		Tracker:      NewRedisProcessedUploadTracker(rdb),
	}
	hookB := &AvatarHook{
		UserUseCase:  ucB,
		ProcessedTTL: 10 * time.Minute,
		Tracker:      NewRedisProcessedUploadTracker(rdb),
	}

	event := tus.UploadEvent{
		UploadID: "upload-recover-1",
		FileURL:  "https://cdn.example.com/avatars/upload-recover-1.png",
		Metadata: map[string]string{"authenticated_user_id": "user-1"},
	}

	// First instance fails -> key must be unmarked in Redis.
	err := hookA.HandleUpload(context.Background(), event)
	require.Error(t, err)
	assert.Equal(t, int32(1), ucA.calls.Load())
	assert.False(t, mr.Exists(redisProcessedUploadKeyPrefix+event.UploadID), "failed upload must release Redis claim")

	// Second instance retries -> succeeds and persists key.
	require.NoError(t, hookB.HandleUpload(context.Background(), event))
	assert.Equal(t, int32(1), ucB.calls.Load())
	assert.True(t, mr.Exists(redisProcessedUploadKeyPrefix+event.UploadID))

	// Replay on first instance is now deduplicated.
	require.NoError(t, hookA.HandleUpload(context.Background(), event))
	assert.Equal(t, int32(1), ucA.calls.Load())
}

func TestRedisProcessedUploadTracker_TTLExpiryAllowsReprocessing(t *testing.T) {
	mr, rdb := setupMiniredisClient(t)

	tracker := NewRedisProcessedUploadTracker(rdb)
	claimed, err := tracker.MarkProcessedIfAbsent(context.Background(), "upload-ttl-1", 5*time.Minute)
	require.NoError(t, err)
	assert.True(t, claimed)

	claimed, err = tracker.MarkProcessedIfAbsent(context.Background(), "upload-ttl-1", 5*time.Minute)
	require.NoError(t, err)
	assert.False(t, claimed)

	mr.FastForward(6 * time.Minute)

	claimed, err = tracker.MarkProcessedIfAbsent(context.Background(), "upload-ttl-1", 5*time.Minute)
	require.NoError(t, err)
	assert.True(t, claimed)
}

func TestRedisProcessedUploadTracker_NilClientSafety(t *testing.T) {
	tracker := NewRedisProcessedUploadTracker(nil)

	claimed, err := tracker.MarkProcessedIfAbsent(context.Background(), "upload-nil-1", time.Minute)
	require.NoError(t, err)
	assert.True(t, claimed)

	require.NoError(t, tracker.UnmarkProcessed(context.Background(), "upload-nil-1"))

	var nilTracker *RedisProcessedUploadTracker
	claimed, err = nilTracker.MarkProcessedIfAbsent(context.Background(), "upload-nil-2", time.Minute)
	require.NoError(t, err)
	assert.True(t, claimed)
	require.NoError(t, nilTracker.UnmarkProcessed(context.Background(), "upload-nil-2"))
}
