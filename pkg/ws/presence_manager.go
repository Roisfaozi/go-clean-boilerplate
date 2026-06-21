package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// PresenceManager defines operations for tracking online users
type PresenceManager interface {
	SetUserOnline(ctx context.Context, orgID, userID string, userData *PresenceUser) error
	SetUserOffline(ctx context.Context, orgID, userID string) error
	GetOnlineUsers(ctx context.Context, orgID string) ([]PresenceUser, error)
	RefreshUserHeartbeat(ctx context.Context, orgID, userID string) error
	PruneStaleUsers(ctx context.Context, timeout time.Duration) (map[string][]string, error)
}

type PresenceUser struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	Status    string `json:"status"` // online, away
	LastSeen  int64  `json:"last_seen"`
}

type RedisPresenceManager struct {
	redisClient *redis.Client
	log         *logrus.Logger
	ttl         time.Duration
}

func NewPresenceManager(redisClient *redis.Client, log *logrus.Logger, ttl time.Duration) *RedisPresenceManager {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &RedisPresenceManager{
		redisClient: redisClient,
		log:         log,
		ttl:         ttl,
	}
}

func (m *RedisPresenceManager) SetUserOnline(ctx context.Context, orgID, userID string, userData *PresenceUser) error {
	pipe := m.redisClient.Pipeline()
	now := time.Now().Unix()

	keyOrg := fmt.Sprintf("presence:org:%s", orgID)
	pipe.ZAdd(ctx, keyOrg, redis.Z{
		Score:  float64(now),
		Member: userID,
	})

	keyUser := fmt.Sprintf("presence:user:%s", userID)
	userData.Status = "online"
	userData.LastSeen = time.Now().UnixMilli()
	data, _ := json.Marshal(userData)

	pipe.Set(ctx, keyUser, data, m.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (m *RedisPresenceManager) SetUserOffline(ctx context.Context, orgID, userID string) error {
	pipe := m.redisClient.Pipeline()

	keyOrg := fmt.Sprintf("presence:org:%s", orgID)
	pipe.ZRem(ctx, keyOrg, userID)

	keyUser := fmt.Sprintf("presence:user:%s", userID)
	pipe.Del(ctx, keyUser)

	_, err := pipe.Exec(ctx)
	return err
}

func (m *RedisPresenceManager) GetOnlineUsers(ctx context.Context, orgID string) ([]PresenceUser, error) {
	keyOrg := fmt.Sprintf("presence:org:%s", orgID)

	userIDs, err := m.redisClient.ZRange(ctx, keyOrg, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return []PresenceUser{}, nil
	}

	pipe := m.redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(userIDs))

	for i, uid := range userIDs {
		keyUser := fmt.Sprintf("presence:user:%s", uid)
		cmds[i] = pipe.Get(ctx, keyUser)
	}

	_, _ = pipe.Exec(ctx)

	users := make([]PresenceUser, 0, len(userIDs))
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			var u PresenceUser
			if err := json.Unmarshal([]byte(val), &u); err == nil {
				users = append(users, u)
			}
		}
	}

	return users, nil
}

func (m *RedisPresenceManager) RefreshUserHeartbeat(ctx context.Context, orgID, userID string) error {
	pipe := m.redisClient.Pipeline()
	now := time.Now().Unix()

	keyOrg := fmt.Sprintf("presence:org:%s", orgID)
	pipe.ZAdd(ctx, keyOrg, redis.Z{
		Score:  float64(now),
		Member: userID,
	})

	keyUser := fmt.Sprintf("presence:user:%s", userID)
	pipe.Expire(ctx, keyUser, m.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (m *RedisPresenceManager) PruneStaleUsers(ctx context.Context, timeout time.Duration) (map[string][]string, error) {
	iter := m.redisClient.Scan(ctx, 0, "presence:org:*", 0).Iterator()
	staleThreshold := time.Now().Add(-timeout).Unix()
	removedUsers := make(map[string][]string)

	for iter.Next(ctx) {
		keyOrg := iter.Val()
		var orgID string
		_, _ = fmt.Sscanf(keyOrg, "presence:org:%s", &orgID)

		staleIDs, err := m.redisClient.ZRangeByScore(ctx, keyOrg, &redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", staleThreshold),
		}).Result()

		if err != nil || len(staleIDs) == 0 {
			continue
		}

		pipe := m.redisClient.Pipeline()
		pipe.ZRemRangeByScore(ctx, keyOrg, "-inf", fmt.Sprintf("%d", staleThreshold))
		for _, uid := range staleIDs {
			pipe.Del(ctx, fmt.Sprintf("presence:user:%s", uid))
		}

		if _, err := pipe.Exec(ctx); err == nil {
			removedUsers[orgID] = staleIDs
		}
	}

	return removedUsers, iter.Err()
}
