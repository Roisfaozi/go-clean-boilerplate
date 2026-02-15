package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisPresenceManager(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	logger := logrus.New()

	// SUT (System Under Test)
	manager := NewPresenceManager(redisClient, logger, 5*time.Minute)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	userData := &PresenceUser{
		UserID:    userID,
		Name:      "John Doe",
		AvatarURL: "https://avatar.com/john",
		Role:      "admin",
	}

	t.Run("SetUserOnline - Data is stored in Redis", func(t *testing.T) {
		err := manager.SetUserOnline(ctx, orgID, userID, userData)
		assert.NoError(t, err)

		// Check Org ZSET
		score, err := mr.ZScore(fmt.Sprintf("presence:org:%s", orgID), userID)
		assert.NoError(t, err)
		assert.True(t, score > 0)

		// Check User Metadata
		val, err := mr.Get(fmt.Sprintf("presence:user:%s", userID))
		assert.NoError(t, err)
		assert.NotEmpty(t, val)

		var storedUser PresenceUser
		err = json.Unmarshal([]byte(val), &storedUser)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", storedUser.Name)
		assert.Equal(t, "online", storedUser.Status)
	})

	t.Run("GetOnlineUsers - Returns list of active users", func(t *testing.T) {
		users, err := manager.GetOnlineUsers(ctx, orgID)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, userID, users[0].UserID)
	})

	t.Run("Data Isolation - Different Org gets empty list", func(t *testing.T) {
		users, err := manager.GetOnlineUsers(ctx, "other-org")
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})

	t.Run("SetUserOffline - Data is removed", func(t *testing.T) {
		err := manager.SetUserOffline(ctx, orgID, userID)
		assert.NoError(t, err)

		// Verify removal from ZSET
		_, err = mr.ZScore(fmt.Sprintf("presence:org:%s", orgID), userID)
		assert.Error(t, err) // Should not exist

		// Verify metadata removal
		val, _ := mr.Get(fmt.Sprintf("presence:user:%s", userID))
		assert.Empty(t, val)
	})

	t.Run("PruneStaleUsers - Removes inactive users", func(t *testing.T) {
		// 1. Add an active user
		_ = manager.SetUserOnline(ctx, "org-A", "user-active", &PresenceUser{Name: "Active"})

		// 2. Add a stale user manually to miniredis ZSET with old score
		_, err := mr.ZAdd("presence:org:org-A", 1000, "user-stale") // Very old timestamp
		require.NoError(t, err)
		err = mr.Set("presence:user:user-stale", `{"name":"Stale"}`)
		require.NoError(t, err)

		// 3. Prune with 1 minute threshold
		removed, err := manager.PruneStaleUsers(ctx, 1*time.Minute)
		assert.NoError(t, err)

		// Verify "user-stale" is in removed list for "org-A"
		assert.Contains(t, removed["org-A"], "user-stale")
		assert.NotContains(t, removed["org-A"], "user-active")

		// Verify deletion from Redis
		members, _ := mr.ZMembers("presence:org:org-A")
		assert.NotContains(t, members, "user-stale")
		assert.Contains(t, members, "user-active")

		val, _ := mr.Get("presence:user:user-stale")
		assert.Empty(t, val)
	})
}
