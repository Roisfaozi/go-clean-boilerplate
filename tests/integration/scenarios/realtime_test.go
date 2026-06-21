//go:build integration
// +build integration

package scenarios

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_RealTime_LoginBroadcast(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	wsConfig := &ws.WebSocketConfig{
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         54 * time.Second,
		MaxMessageSize:     512,
		DistributedEnabled: true,
		RedisPrefix:        "ws_broadcast:",
	}
	presenceManager := ws.NewPresenceManager(env.Redis, env.Logger, 5*time.Minute)
	wsManager := ws.NewWebSocketManager(wsConfig, env.Logger, env.Redis, presenceManager)
	go wsManager.Run()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	publisher := delivery.NewEventPublisher(wsManager, nil, env.Logger)
	authz := authRepo.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, publisher, authz, nil, nil, make(map[string]sso.Provider))

	// 1. Create Organization
	orgID := "test-org-123"
	user := setup.CreateTestUser(t, env.DB, "ws_user", "ws@test.com", "pass")

	org := &orgEntity.Organization{
		ID:      orgID,
		Name:    "Test Org",
		Slug:    "test-org",
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := oRepo.Create(context.Background(), org, "role:admin")
	require.NoError(t, err)

	// 2. Subscribe to Organization Channel (Redis Channel Name: ws_broadcast:org_<org_id>_notifications)
	expectedChannel := "ws_broadcast:org_" + orgID + "_notifications"
	pubsub := env.Redis.Subscribe(context.Background(), expectedChannel)
	defer pubsub.Close()

	_, err = pubsub.Receive(context.Background())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, _, err = authService.Login(context.Background(), model.LoginRequest{Username: "ws_user", Password: "pass"})
	require.NoError(t, err)

	select {
	case msg := <-pubsub.Channel():

		assert.Equal(t, expectedChannel, msg.Channel)

		var notification map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &notification)
		assert.NoError(t, err)

		assert.Equal(t, "user_login", notification["type"])
		assert.Equal(t, user.ID, notification["user_id"])

	case <-time.After(5 * time.Second):
		assert.Fail(t, "Timeout waiting for WebSocket broadcast via Redis on channel "+expectedChannel)
	}
}
