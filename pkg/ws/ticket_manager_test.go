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
		Addr:            mr.Addr(),
		DisableIdentity: true,
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
		Addr:            mr.Addr(),
		DisableIdentity: true,
		MaxRetries:      -1, // Don't retry/log when we close the server
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

func TestRedisTicketManager_RecentUnusedReplay(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:            mr.Addr(),
		DisableIdentity: true,
	})
	tm := ws.NewRedisTicketManager(rdb, 30*time.Second)

	ctx := context.Background()
	t1, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	require.NotEmpty(t, t1)

	t2, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.Equal(t, t1, t2, "repeat CreateTicket within replay window should return same unused ticket")

	userCtx, err := tm.ValidateTicket(ctx, t2)
	require.NoError(t, err)
	assert.Equal(t, "user-1", userCtx.UserID)
	assert.Equal(t, "org-1", userCtx.OrganizationID)
	assert.Equal(t, "session-1", userCtx.SessionID)
	assert.Equal(t, "admin", userCtx.Role)
	assert.Equal(t, "alice", userCtx.Username)

	_, err = tm.ValidateTicket(ctx, t1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired ticket")
}

func TestRedisTicketManager_ConsumedTicketIssuesFreshTicket(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:            mr.Addr(),
		DisableIdentity: true,
	})
	tm := ws.NewRedisTicketManager(rdb, 30*time.Second)

	ctx := context.Background()
	t1, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)

	_, err = tm.ValidateTicket(ctx, t1)
	require.NoError(t, err)

	t2, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, t1, t2, "consumed ticket must never be replayed")

	// Old ticket remains invalid
	_, err = tm.ValidateTicket(ctx, t1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired ticket")

	// New ticket is valid
	userCtx, err := tm.ValidateTicket(ctx, t2)
	require.NoError(t, err)
	assert.Equal(t, "user-1", userCtx.UserID)
}

func TestRedisTicketManager_DistinctContextIssuesDistinctTickets(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:            mr.Addr(),
		DisableIdentity: true,
	})
	tm := ws.NewRedisTicketManager(rdb, 30*time.Second)

	ctx := context.Background()
	base, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)

	diffOrg, err := tm.CreateTicket(ctx, "user-1", "org-2", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, base, diffOrg)

	diffSession, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-2", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, base, diffSession)

	diffUser, err := tm.CreateTicket(ctx, "user-2", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, base, diffUser)

	diffRole, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "member", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, base, diffRole)

	diffUsername, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "bob")
	require.NoError(t, err)
	assert.NotEqual(t, base, diffUsername)
}

func TestRedisTicketManager_RecentWindowExpiration(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:            mr.Addr(),
		DisableIdentity: true,
	})
	tm := ws.NewRedisTicketManager(rdb, 30*time.Second)

	ctx := context.Background()
	t1, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)

	// Still within default 5s replay window
	mr.FastForward(3 * time.Second)
	t1Replay, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.Equal(t, t1, t1Replay)

	// Past default 5s replay window (total 6s elapsed)
	mr.FastForward(3 * time.Second)
	t2, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, t1, t2, "CreateTicket after recentWindow expires must issue a fresh ticket")

	// Both tickets remain valid until consumed or TTL expires
	_, err = tm.ValidateTicket(ctx, t1)
	require.NoError(t, err)
	_, err = tm.ValidateTicket(ctx, t2)
	require.NoError(t, err)
}

func TestRedisTicketManager_SetRecentReplayWindow(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:            mr.Addr(),
		DisableIdentity: true,
	})
	tm := ws.NewRedisTicketManager(rdb, 30*time.Second)

	ctx := context.Background()

	// Disable replay window
	tm.SetRecentReplayWindow(0)
	t1, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	t2, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, t1, t2, "disabling replay window should always mint fresh tickets")

	// Configure custom 2s replay window
	tm.SetRecentReplayWindow(2 * time.Second)
	t3, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	t3Replay, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.Equal(t, t3, t3Replay)

	mr.FastForward(3 * time.Second)
	t4, err := tm.CreateTicket(ctx, "user-1", "org-1", "session-1", "admin", "alice")
	require.NoError(t, err)
	assert.NotEqual(t, t3, t4)
}
