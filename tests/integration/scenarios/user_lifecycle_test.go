//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteUserLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:moderator" })

	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	uUC := userUseCase.NewUserUseCase(env.Logger, tm, userRepo, env.Enforcer, auditUC)
	aUC := authUseCase.NewAuthUsecase(jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC, nil)
	pUC := permissionUseCase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	// 1. Register User
	registerReq := &userModel.RegisterUserRequest{
		Username: "lifecycleuser", Email: "lifecycle@example.com", Password: "password123",
		Name: "Lifecycle User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}
	user, err := uUC.Create(context.Background(), registerReq)
	require.NoError(t, err)

	// 2. Login
	loginReq := authModel.LoginRequest{Username: "lifecycleuser", Password: "password123", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	loginResp, refreshToken, err := aUC.Login(context.Background(), loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	// 3. Update Profile
	updateReq := &userModel.UpdateUserRequest{ID: user.ID, Name: "Updated User", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	updatedUser, err := uUC.Update(context.Background(), updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated User", updatedUser.Name)

	// 4. Assign Role
	err = pUC.AssignRoleToUser(context.Background(), user.ID, "role:moderator")
	require.NoError(t, err)

	// 5. Refresh Token
	newToken, _, err := aUC.RefreshToken(context.Background(), refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken.AccessToken)

	// 6. Logout
	claims, _ := jwtManager.ValidateAccessToken(newToken.AccessToken)
	err = aUC.RevokeToken(context.Background(), user.ID, claims.SessionID)
	require.NoError(t, err)

	// 7. Delete User
	deleteReq := &userModel.DeleteUserRequest{ID: user.ID, IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	err = uUC.DeleteUser(context.Background(), "admin-id", deleteReq)
	require.NoError(t, err)
}

func TestRBACWorkflow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	pUC := permissionUseCase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)
	testUser := setup.CreateTestUser(t, env.DB, "rbacuser", "rbac@example.com", "password123")

	// 1. Grant Permission to Role
	err := pUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/admin/dashboard", "GET")
	require.NoError(t, err)

	// 2. Assign Role to User
	err = pUC.AssignRoleToUser(context.Background(), testUser.ID, "role:admin")
	require.NoError(t, err)

	// 3. Verify Access
	allowed, err := env.Enforcer.Enforce(testUser.ID, "/api/v1/admin/dashboard", "GET")
	require.NoError(t, err)
	assert.True(t, allowed)

	// 4. Verify Role via Casbin
	roles, _ := env.Enforcer.GetRolesForUser(testUser.ID)
	assert.Contains(t, roles, "role:admin")
}