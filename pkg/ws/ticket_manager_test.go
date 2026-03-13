package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTicketManager(t *testing.T) (*ws.RedisTicketManager, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	tm := ws.NewRedisTicketManager(rdb, 1*time.Second) // Short TTL for testing
	return tm, mr
}

func TestRedisTicketManager_CreateAndValidate(t *testing.T) {
	tm, mr := setupTicketManager(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "user-1"
	orgID := "org-1"
	sessionID := "session-1"
	role := "admin"
	username := "admin_user"

	// 1. Create Ticket
	ticket, err := tm.CreateTicket(ctx, userID, orgID, sessionID, role, username)
	require.NoError(t, err)
	require.NotEmpty(t, ticket)

	// 2. Validate Ticket (Success)
	userCtx, err := tm.ValidateTicket(ctx, ticket)
	require.NoError(t, err)
	assert.Equal(t, userID, userCtx.UserID)
	assert.Equal(t, orgID, userCtx.OrganizationID)
	assert.Equal(t, sessionID, userCtx.SessionID)
	assert.Equal(t, role, userCtx.Role)
	assert.Equal(t, username, userCtx.Username)

	// 3. Validate Ticket Again (Should Fail - One Time Use)
	_, err = tm.ValidateTicket(ctx, ticket)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired ticket")
}

func TestRedisTicketManager_BadData(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	tm := ws.NewRedisTicketManager(rdb, 0) // Should default to 30s

	rdb.Set(context.Background(), "ws:ticket:bad-json", "{bad-json}", 0)
	_, err = tm.ValidateTicket(context.Background(), "bad-json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")

	mr.Close()
	// Error on get
	_, err = tm.ValidateTicket(context.Background(), "anything")
	require.Error(t, err)
}

func TestRedisTicketManager_Expiration(t *testing.T) {
	tm, mr := setupTicketManager(t)
	defer mr.Close()

	ctx := context.Background()
	ticket, err := tm.CreateTicket(ctx, "u1", "o1", "s1", "r1", "run1")
	require.NoError(t, err)

	// Fast forward time greater than TTL (1s)
	mr.FastForward(2 * time.Second)

	// Validate Ticket (Should Fail - Expired)
	_, err = tm.ValidateTicket(ctx, ticket)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired ticket")
}

func TestRedisTicketManager_InvalidTicket(t *testing.T) {
	tm, mr := setupTicketManager(t)
	defer mr.Close()

	ctx := context.Background()

	// Validate non-existent ticket
	_, err := tm.ValidateTicket(ctx, "non-existent-ticket")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired ticket")
}
