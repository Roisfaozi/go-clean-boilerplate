//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	roleUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_AdminSecurity_AccountSuspension(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)

	auditService := auditUC.NewAuditUseCase(aucRepo, env.Logger)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)

	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, tm, env.Logger, nil, nil, env.Enforcer, auditService, nil)

	userService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, auditService, authService, nil)

	password := "Pass123!"
	user := setup.CreateTestUser(t, env.DB, "suspend_target", "suspend@test.com", password)

	loginResp, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	sessions, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, sessions)

	err = userService.UpdateStatus(context.Background(), user.ID, userEntity.UserStatusBanned)
	require.NoError(t, err)

	sessionsAfterBan, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Empty(t, sessionsAfterBan, "All sessions should be revoked after ban")

	_, err = authService.ValidateAccessToken(loginResp.AccessToken)
	assert.Error(t, err, "Token should be invalid after revocation")
}

func TestScenario_AdminSecurity_RBAC_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepoData := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepoData)

	roleName := "content_editor"
	_, err := roleService.Create(context.Background(), &roleModel.CreateRoleRequest{Name: roleName})
	require.NoError(t, err)

	path, method := "/api/v1/articles", "POST"
	err = permService.GrantPermissionToRole(context.Background(), roleName, path, method)
	require.NoError(t, err)

	user := setup.CreateTestUser(t, env.DB, "editor_user", "editor@test.com", "pass")
	err = permService.AssignRoleToUser(context.Background(), user.ID, roleName)
	require.NoError(t, err)

	ok, err := env.Enforcer.Enforce(roleName, path, method)
	assert.NoError(t, err)
	assert.True(t, ok, "Role should have permission")

	userRoles, _ := env.Enforcer.GetRolesForUser(user.ID)
	assert.Contains(t, userRoles, roleName)
}

func TestScenario_AdminSecurity_TokenRotation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, tm, env.Logger, nil, nil, env.Enforcer, nil, nil)

	user := setup.CreateTestUser(t, env.DB, "rotator", "rot@test.com", "pass")
	_, rt1, err := authService.Login(context.Background(), authModel.LoginRequest{Username: user.Username, Password: "pass"})
	require.NoError(t, err)

	_, rt2, err := authService.RefreshToken(context.Background(), rt1)
	require.NoError(t, err)
	assert.NotEqual(t, rt1, rt2)

	_, _, err = authService.RefreshToken(context.Background(), rt1)
	assert.Error(t, err, "Reuse of old refresh token should fail")

	_, rt3, err := authService.RefreshToken(context.Background(), rt2)
	assert.NoError(t, err)
	assert.NotEmpty(t, rt3)
}
