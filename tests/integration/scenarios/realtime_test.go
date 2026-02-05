//go:build integration
// +build integration

package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_RealTime_LoginBroadcast(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	wsConfig := &ws.WebSocketConfig{
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         54 * time.Second,
		MaxMessageSize:     512,
		DistributedEnabled: true,
		RedisPrefix:        "ws_broadcast:",
	}
	wsManager := ws.NewWebSocketManager(wsConfig, env.Logger, env.Redis)
	go wsManager.Run()

	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, wsManager, nil, env.Enforcer, nil, nil)

	// Create User
	user := setup.CreateTestUser(t, env.DB, "ws_user", "ws@test.com", "pass")

	// Create Organization and link User (Required for Login broadcast)
	orgID := uuid.New().String()
	org := &orgEntity.Organization{
		ID:      orgID,
		Name:    "WS Test Org",
		Slug:    "ws-test-org",
		OwnerID: user.ID,
		Status:  "active",
	}
	err := oRepo.Create(context.Background(), org, "owner")
	require.NoError(t, err)

	// Subscribe to the organization channel
	channelName := fmt.Sprintf("ws_broadcast:org_%s_notifications", orgID)
	pubsub := env.Redis.Subscribe(context.Background(), channelName)
	defer pubsub.Close()

	// Wait for subscription confirmation
	_, err = pubsub.Receive(context.Background())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Perform Login
	_, _, err = authService.Login(context.Background(), model.LoginRequest{Username: "ws_user", Password: "pass"})
	require.NoError(t, err)

	select {
	case msg := <-pubsub.Channel():
		assert.Equal(t, channelName, msg.Channel)

		var notification map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &notification)
		assert.NoError(t, err)

		assert.Equal(t, "user_login", notification["type"])
		assert.Equal(t, user.ID, notification["user_id"])

	case <-time.After(5 * time.Second):
		assert.Fail(t, "Timeout waiting for WebSocket broadcast via Redis")
	}
}
