//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLifecycle_FullFlow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	jwtManager := jwt.NewJWTManager("lifecycle-secret", "lifecycle-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil)

	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	authz := authRepository.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	// Start worker to process audit logs
	cleanupHandler := handlers.NewCleanupTaskHandler(tokenRepo, userRepo, auditRepo, env.Logger)
	workerCfg := worker.WorkerConfig{} // Minimal config
	processor := worker.NewRedisTaskProcessor(redisOpt, env.Logger, cleanupHandler, auditUC, auditRepo, workerCfg)
	env.StartWorker(processor)

	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, oRepo, tm, env.Logger, nil, authz, taskDistributor, nil)
	userUC := userUseCase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, nil)

	ctx := context.Background()

	regReq := &userModel.RegisterUserRequest{
		Username: "lifecycle", Email: "lifecycle@example.com", Password: "password123", Name: "Life Cycle",
	}
	userResp, err := userUC.Create(ctx, regReq)
	require.NoError(t, err)
	userID := userResp.ID

	loginReq := authModel.LoginRequest{Username: regReq.Username, Password: regReq.Password}
	loginResp, _, err := authUC.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	updateReq := &userModel.UpdateUserRequest{
		ID: userID, Name: "Updated Life",
	}
	updateResp, err := userUC.Update(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Life", updateResp.Name)

	deleteReq := &userModel.DeleteUserRequest{ID: userID}
	err = userUC.DeleteUser(ctx, userID, deleteReq)
	require.NoError(t, err)

	// Manually trigger outbox sync since the worker might be slow
	err = handlers.NewOutboxTaskHandler(auditRepo, env.Logger).ProcessAuditOutbox(ctx, nil)
	require.NoError(t, err)

	// Wait for any final async processing
	time.Sleep(2 * time.Second)

	logs, _, err := auditUC.GetLogsDynamic(ctx, &querybuilder.DynamicFilter{
		Sort: &[]querybuilder.SortModel{{ColId: "CreatedAt", Sort: "asc"}},
	})
	require.NoError(t, err)

	var userLogs []auditModel.AuditLogResponse
	for _, l := range logs {
		if l.UserID == userID || l.EntityID == userID {
			userLogs = append(userLogs, l)
		}
	}

	require.GreaterOrEqual(t, len(userLogs), 4, "Should have at least 4 audit entries for this lifecycle")

	actions := make(map[string]bool)
	for _, l := range userLogs {
		actions[l.Action] = true
	}

	assert.True(t, actions["CREATE"], "CREATE log missing")
	assert.True(t, actions["LOGIN"], "LOGIN log missing")
	assert.True(t, actions["UPDATE"], "UPDATE log missing")
	assert.True(t, actions["DELETE"], "DELETE log missing")
}
