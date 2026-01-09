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

// Scenario 1: Account Suspension & Real-time Force Logout
func TestScenario_AdminSecurity_AccountSuspension(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// Init Modules
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)

	auditService := auditUC.NewAuditUseCase(aucRepo, env.Logger)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)

	// Auth UseCase (for Login/Revoke)
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, env.Enforcer, auditService, nil)

	// User UseCase (for UpdateStatus)
	userService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, auditService, authService)

	// 1. Setup User & Login
	password := "Pass123!"
	user := setup.CreateTestUser(t, env.DB, "suspend_target", "suspend@test.com", password)

	loginResp, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	// Verify Session exists in Redis
	sessions, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, sessions)

	// 2. Admin Bans User
	err = userService.UpdateStatus(context.Background(), user.ID, userEntity.UserStatusBanned)
	require.NoError(t, err)

	// 3. Verify Session Revoked (Real-time force logout)
	sessionsAfterBan, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Empty(t, sessionsAfterBan, "All sessions should be revoked after ban")

	// 4. Verify Token Validation Fails
	_, err = authService.ValidateAccessToken(loginResp.AccessToken)
	assert.Error(t, err, "Token should be invalid after revocation")
}

// Scenario 2: RBAC Lifecycle (Create Role -> Assign -> Access)
func TestScenario_AdminSecurity_RBAC_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// Init Repos & Services
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepoData := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepoData)

	// 1. Create New Role
	roleName := "content_editor"
	_, err := roleService.Create(context.Background(), &roleModel.CreateRoleRequest{Name: roleName})
	require.NoError(t, err)

	// 2. Grant Permission to Role
	path, method := "/api/v1/articles", "POST"
	err = permService.GrantPermissionToRole(context.Background(), roleName, path, method)
	require.NoError(t, err)

	// 3. Create User & Assign Role
	user := setup.CreateTestUser(t, env.DB, "editor_user", "editor@test.com", "pass")
	err = permService.AssignRoleToUser(context.Background(), user.ID, roleName)
	require.NoError(t, err)

	// 4. Verify Access via Enforcer
	ok, err := env.Enforcer.Enforce(roleName, path, method)
	assert.NoError(t, err)
	assert.True(t, ok, "Role should have permission")

	// 5. Verify User Inheritance
	userRoles, _ := env.Enforcer.GetRolesForUser(user.ID)
	assert.Contains(t, userRoles, roleName)
}

// Scenario 3: Token Rotation Security
func TestScenario_AdminSecurity_TokenRotation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// Init Auth
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, env.Enforcer, nil, nil)

	// 1. Login
	user := setup.CreateTestUser(t, env.DB, "rotator", "rot@test.com", "pass")
	_, rt1, err := authService.Login(context.Background(), authModel.LoginRequest{Username: user.Username, Password: "pass"})
	require.NoError(t, err)

	// 2. Rotate Token (RT1 -> RT2)
	_, rt2, err := authService.RefreshToken(context.Background(), rt1)
	require.NoError(t, err)
	assert.NotEqual(t, rt1, rt2)

	// 3. Attempt Reuse RT1 (Security Check)
	_, _, err = authService.RefreshToken(context.Background(), rt1)
	assert.Error(t, err, "Reuse of old refresh token should fail")

	// 4. Use RT2 (Should Work)
	_, rt3, err := authService.RefreshToken(context.Background(), rt2)
	assert.NoError(t, err)
	assert.NotEmpty(t, rt3)
}
