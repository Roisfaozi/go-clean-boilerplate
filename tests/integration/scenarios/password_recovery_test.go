//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scenario: Password Recovery Lifecycle
// User Forgets Password -> Request Token -> Reset Password -> Login with New Password
func TestScenario_PasswordRecovery_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Modules
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB) // Hybrid Repo
	aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)

	auditService := auditUC.NewAuditUseCase(aucRepo, env.Logger)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)

	// Note: We pass nil for TaskDistributor because we don't want to test the worker here,
	// we want to test the logic flow. The logic should still save to DB even if distributor is nil
	// (or we can verify it doesn't panic). Based on reading, it handles nil distributor gracefully.
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, env.Enforcer, auditService, nil)

	// 2. Create User
	oldPassword := "OldPass123!"
	newPassword := "NewPass456!"
	user := setup.CreateTestUser(t, env.DB, "forgot_user", "forgot@test.com", oldPassword)

	// 3. Request Forgot Password
	err := authService.ForgotPassword(context.Background(), user.Email)
	require.NoError(t, err)

	// 4. Intercept Token from DB
	// Since the usecase generates it internally and sends it via email (which we mocked/skipped),
	// we must retrieve it from the database table directly to simulate the user receiving it.
	var resetToken authEntity.PasswordResetToken
	err = env.DB.Where("email = ?", user.Email).First(&resetToken).Error
	require.NoError(t, err, "Reset token should be saved in DB")
	assert.NotEmpty(t, resetToken.Token)

	// 5. Reset Password
	err = authService.ResetPassword(context.Background(), resetToken.Token, newPassword)
	require.NoError(t, err)

	// 6. Verify Token is consumed/deleted
	var checkToken authEntity.PasswordResetToken
	err = env.DB.Where("token = ?", resetToken.Token).First(&checkToken).Error
	assert.Error(t, err, "Token should be deleted after use")

	// 7. Attempt Login with OLD Password (Should Fail)
	_, _, err = authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: oldPassword,
	})
	assert.Error(t, err, "Login with old password should fail")

	// 8. Attempt Login with NEW Password (Should Success)
	resp, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: newPassword,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}
