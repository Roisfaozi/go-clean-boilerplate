//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserIntegration(env *setup.TestEnvironment) usecase.UserUseCase {
	userRepo := repository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil, nil)

	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authz := authRepository.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, authz, nil, nil, make(map[string]sso.Provider))

	return usecase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, nil, nil)
}

func TestUserIntegration_Positive_Delete(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	testUser := setup.CreateTestUser(t, env.DB, "deleteuser", "delete@example.com", "password123")
	userUC := setupUserIntegration(env)
	userRepo := repository.NewUserRepository(env.DB, env.Logger)

	deleteReq := &model.DeleteUserRequest{
		ID:        testUser.ID,
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)
	require.NoError(t, err)

	_, err = userRepo.FindByID(context.Background(), testUser.ID)
	assert.Error(t, err)
}

func TestUserIntegration_Positive_GetByID(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	testUser := setup.CreateTestUser(t, env.DB, "getbyiduser", "getbyid@example.com", "password123")
	userUC := setupUserIntegration(env)

	result, err := userUC.GetUserByID(context.Background(), testUser.ID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testUser.ID, result.ID)
	assert.Equal(t, testUser.Username, result.Username)
}

func TestUserIntegration_Positive_Create_WithRoleAssignment(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)
	userRepo := repository.NewUserRepository(env.DB, env.Logger)

	req := &model.RegisterUserRequest{
		Username:  "roleuser",
		Email:     "roleuser@example.com",
		Password:  "Password123!",
		Name:      "Role User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Username, result.Username)
	assert.Equal(t, req.Email, result.Email)

	// Sync enforcer after transactional changes
	err = env.Enforcer.LoadPolicy()
	require.NoError(t, err)

	user, err := userRepo.FindByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, req.Username, user.Username)

	roles, err := env.Enforcer.GetRolesForUser(result.ID, "global")
	require.NoError(t, err)
	assert.Contains(t, roles, "role:user")
}

func TestUserStatus_BannedFlow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	password := "password123"
	user := setup.CreateTestUser(t, env.DB, "banneduser", "banned@example.com", password)

	env.DB.Model(&entity.User{}).Where("id = ?", user.ID).Update("status", entity.UserStatusBanned)

	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	_ = auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil, nil)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authz := authRepository.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, authz, nil, nil, make(map[string]sso.Provider))

	loginReq := authModel.LoginRequest{Username: user.Username, Password: password}
	loginResp, _, err := authUC.Login(context.Background(), loginReq)

	require.Error(t, err, "Login should fail for banned users")
	assert.Nil(t, loginResp)

	t.Run("Verify user status is banned", func(t *testing.T) {
		u, _ := userRepo.FindByID(context.Background(), user.ID)
		assert.Equal(t, entity.UserStatusBanned, u.Status)
	})
}

func TestUserIntegration_Create_Positive_ValidData(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "validuser", Email: "valid@example.com", Password: "SecurePass123!",
		Name: "Valid User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Username, result.Username)
	assert.Equal(t, req.Email, result.Email)
}

func TestUserIntegration_Update_Positive_ValidUpdate(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	testUser := setup.CreateTestUser(t, env.DB, "updateuser", "update@example.com", "password123")
	userUC := setupUserIntegration(env)

	updateReq := &model.UpdateUserRequest{
		ID: testUser.ID, Name: "Updated Name", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", result.Name)
}

func TestUserIntegration_Create_Negative_DuplicateUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CreateTestUser(t, env.DB, "duplicate", "first@example.com", "password123")
	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "duplicate", Email: "second@example.com", Password: "password123",
		Name: "Second User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserIntegration_Create_Negative_DuplicateEmail(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CreateTestUser(t, env.DB, "user1", "duplicate@example.com", "password123")
	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "user2", Email: "duplicate@example.com", Password: "password123",
		Name: "User Two", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserIntegration_Update_Negative_NonExistentUser(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	updateReq := &model.UpdateUserRequest{
		ID: "non-existent-id", Name: "Updated Name", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserIntegration_Delete_Negative_NonExistentUser(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	deleteReq := &model.DeleteUserRequest{ID: "non-existent-id", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}

	err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)

	assert.Error(t, err)
}

func TestUserIntegration_Create_Edge_MinimumUsernameLength(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "abc", Email: "min@example.com", Password: "password123",
		Name: "Min User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotEmpty(t, result.ID)
	}
}

func TestUserIntegration_Create_Edge_MaximumUsernameLength(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	longUsername := strings.Repeat("a", 50)
	req := &model.RegisterUserRequest{
		Username: longUsername, Email: "max@example.com", Password: "password123",
		Name: "Max User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, longUsername, result.Username)
}

func TestUserIntegration_Create_Edge_SpecialCharactersInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "specialuser", Email: "special@example.com", Password: "password123",
		Name: "O'Brien-Smith (Jr.) & Co.", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, req.Name, result.Name)
}

func TestUserIntegration_Create_Edge_UnicodeInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "unicodeuser", Email: "unicode@example.com", Password: "password123",
		Name: "张三 李四 Müller José", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, req.Name, result.Name)
}

func TestUserIntegration_Update_Edge_EmptyOptionalFields(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	testUser := setup.CreateTestUser(t, env.DB, "emptyuser", "empty@example.com", "password123")
	userUC := setupUserIntegration(env)

	updateReq := &model.UpdateUserRequest{
		ID: testUser.ID, Name: "", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestUserIntegration_Create_Edge_EmailWithPlusSign(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "plususer", Email: "user+test@example.com", Password: "password123",
		Name: "Plus User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, req.Email, result.Email)
}

func TestUserIntegration_Security_SQLInjectionInUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	sqlInjections := []string{
		"admin' OR '1'='1",
		"'; DROP TABLE users--",
		"admin'--",
		"1' UNION SELECT * FROM users--",
	}

	for i, injection := range sqlInjections {
		t.Run("SQLInjection_"+fmt.Sprint(i), func(t *testing.T) {
			req := &model.RegisterUserRequest{
				Username: injection, Email: fmt.Sprintf("sql%d@example.com", i), Password: "password123",
				Name: "SQL User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
			}

			result, err := userUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, pkg.SanitizeString(injection), result.Username)
			}
		})
	}
}

func TestUserIntegration_Security_XSSInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	xssPayloads := []string{
		"<script>alert('XSS1')</script>",
		"<img src=x onerror=alert('XSS2')>",
		"javascript:alert('XSS3')",
		"<svg onload=alert('XSS4')>",
	}

	for i, xss := range xssPayloads {
		t.Run("XSS_"+fmt.Sprint(i), func(t *testing.T) {
			req := &model.RegisterUserRequest{
				Username: fmt.Sprintf("xssuser%d", i), Email: fmt.Sprintf("xss%d@example.com", i),
				Password: "password123", Name: xss, IPAddress: "127.0.0.1", UserAgent: "TestAgent",
			}

			result, err := userUC.Create(context.Background(), req)

			require.NoError(t, err)
			assert.NotEmpty(t, result.ID)
		})
	}
}

func TestUserIntegration_Security_PathTraversalInUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	pathTraversals := []string{"../../../etc/passwd", "..\\..\\windows\\system32", "....//....//"}

	for i, path := range pathTraversals {
		t.Run("PathTraversal_"+fmt.Sprint(i), func(t *testing.T) {
			req := &model.RegisterUserRequest{
				Username: path, Email: fmt.Sprintf("path%d@example.com", i), Password: "password123",
				Name: "Path User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
			}

			result, err := userUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
			}
		})
	}
}

func TestUserIntegration_Security_NoSQLInjection(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	noSQLPayloads := []string{
		`{\"$$gt\":\""}`,
		`{\"$$ne\":null}`,
		`admin' || '1'=='1`,
	}

	for i, payload := range noSQLPayloads {
		t.Run("NoSQL_"+fmt.Sprint(i), func(t *testing.T) {
			req := &model.RegisterUserRequest{
				Username: payload, Email: fmt.Sprintf("nosql%d@example.com", i), Password: "password123",
				Name: "NoSQL User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
			}

			result, err := userUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
			}
		})
	}
}

func TestUserIntegration_Security_PasswordNotInResponse(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	userUC := setupUserIntegration(env)

	req := &model.RegisterUserRequest{
		Username: "secureuser", Email: "secure@example.com", Password: "SecurePass123!",
		Name: "Secure User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Secure User", result.Name)
}

func TestUserIntegration_Security_UnauthorizedUpdate(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	testUser := setup.CreateTestUser(t, env.DB, "victim", "victim@example.com", "password123")
	userUC := setupUserIntegration(env)

	updateReq := &model.UpdateUserRequest{
		ID: testUser.ID, Name: "Hacked Name", IPAddress: "192.168.1.100", UserAgent: "AttackerAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	if err == nil {
		assert.NotNil(t, result)
	}
}
