//go:build integration
// +build integration

package scenarios

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scenario: Real-time Notification on Login
// Verifies that a Login action triggers a WebSocket broadcast.
func TestScenario_RealTime_LoginBroadcast(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Dependencies
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	// 2. Setup Real WebSocket Manager (using Redis)
	wsConfig := &ws.WebSocketConfig{
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         54 * time.Second,
		MaxMessageSize:     512,
		DistributedEnabled: true, // Enable distributed mode for Redis Pub/Sub
		RedisPrefix:        "ws_broadcast:",
	}
	wsManager := ws.NewWebSocketManager(wsConfig, env.Logger, env.Redis)
	go wsManager.Run() // Start the manager loop

	// 3. Setup Auth UseCase with WS Manager
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, wsManager, env.Enforcer, nil, nil)

	// 4. Client Subscribe: Listen to Redis Channel directly
	// We need to wait a bit for WS Manager to be ready, but since we subscribe directly to Redis,
	// we just need to ensure our subscription is active BEFORE login triggers publish.

	pubsub := env.Redis.Subscribe(context.Background(), "ws_broadcast:global_notifications")
	defer pubsub.Close()

	// Wait for subscription confirmation
	_, err := pubsub.Receive(context.Background())
	require.NoError(t, err)

	// Give a small buffer to ensure Redis has processed the subscription
	time.Sleep(100 * time.Millisecond)

	// 5. Perform Login
	user := setup.CreateTestUser(t, env.DB, "ws_user", "ws@test.com", "pass")
	_, _, err = authService.Login(context.Background(), model.LoginRequest{Username: "ws_user", Password: "pass"})
	require.NoError(t, err)

	// 6. Assert Message Received
	select {
	case msg := <-pubsub.Channel():
		// Verify Channel
		assert.Equal(t, "ws_broadcast:global_notifications", msg.Channel)

		// Verify Payload
		var notification map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &notification)
		assert.NoError(t, err)

		assert.Equal(t, "user_login", notification["type"])
		assert.Equal(t, user.ID, notification["user_id"])

	case <-time.After(5 * time.Second): // Increased timeout
		assert.Fail(t, "Timeout waiting for WebSocket broadcast via Redis")
	}
}
