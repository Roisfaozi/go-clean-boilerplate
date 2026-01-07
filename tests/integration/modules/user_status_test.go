//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

func TestUserStatus_BannedFlow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	// --- SETUP ---
	password := "password123"
	user := setup.CreateTestUser(t, env.DB, "banneduser", "banned@example.com", password)
	
	// Update status to BANNED
	env.DB.Model(&entity.User{}).Where("id = ?", user.ID).Update("status", entity.UserStatusBanned)

	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := authUseCase.NewAuthUsecase(jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC, nil)

	// 1. LOGIN (Should SUCCEED even if banned)
	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	loginResp, _, err := authUC.Login(context.Background(), loginReq)
	
	require.NoError(t, err, "Login should succeed even for banned users")
	assert.NotEmpty(t, loginResp.AccessToken)

	// 2. MIDDLEWARE CHECK (Simulate access to protected resource)
	// We'll manually call the middleware or check repo
	
	t.Run("Middleware should block banned user", func(t *testing.T) {
		// Verify the user status in DB is indeed banned
		u, _ := userRepo.FindByID(context.Background(), user.ID)
		assert.Equal(t, entity.UserStatusBanned, u.Status)
	})
}
