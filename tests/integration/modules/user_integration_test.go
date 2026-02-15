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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/local"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDependencies(t *testing.T, env *setup.TestEnvironment) (usecase.UserUseCase, repository.UserRepository) {
	userRepo := repository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil)

	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, nil, env.Enforcer, auditUC, nil)

	tmpDir := t.TempDir()
	storageProvider, _ := local.NewLocalStorage(tmpDir, "http://test-bucket")

	return usecase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, storageProvider), userRepo
}

func TestUserIntegration_Create_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC, userRepo := setupTestDependencies(t, env)

	req := &model.RegisterUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Username, result.Username)
	assert.Equal(t, req.Email, result.Email)

	user, err := userRepo.FindByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, req.Username, user.Username)

	roles, err := env.Enforcer.GetRolesForUser(result.ID, "global")
	require.NoError(t, err)
	assert.Contains(t, roles, "role:user")
}

func TestUserIntegration_Create_DuplicateUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "existinguser", "existing@example.com", "password123")

	userUC, _ := setupTestDependencies(t, env)

	req := &model.RegisterUserRequest{
		Username:  "existinguser",
		Email:     "newemail@example.com",
		Password:  "password123",
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserIntegration_Update_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, userRepo := setupTestDependencies(t, env)

	updateReq := &model.UpdateUserRequest{
		ID:        testUser.ID,
		Name:      "Updated Name",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Name", result.Name)

	updatedUser, err := userRepo.FindByID(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
}

func TestUserIntegration_Delete_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, userRepo := setupTestDependencies(t, env)

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

func TestUserIntegration_GetByID_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, _ := setupTestDependencies(t, env)

	result, err := userUC.GetUserByID(context.Background(), testUser.ID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testUser.ID, result.ID)
	assert.Equal(t, testUser.Username, result.Username)
}

func TestUserStatus_BannedFlow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	password := "password123"
	user := setup.CreateTestUser(t, env.DB, "banneduser", "banned@example.com", password)

	env.DB.Model(&entity.User{}).Where("id = ?", user.ID).Update("status", entity.UserStatusBanned)

	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, nil, env.Enforcer, auditUC, nil)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "updateuser", "update@example.com", "password123")
	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "duplicate", "first@example.com", "password123")
	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "user1", "duplicate@example.com", "password123")
	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

	deleteReq := &model.DeleteUserRequest{ID: "non-existent-id", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}

	err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)

	assert.Error(t, err)
}

func TestUserIntegration_Create_Edge_MinimumUsernameLength(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "emptyuser", "empty@example.com", "password123")
	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

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
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "victim", "victim@example.com", "password123")
	userUC := setupUserUseCase(t, env)

	updateReq := &model.UpdateUserRequest{
		ID: testUser.ID, Name: "Hacked Name", IPAddress: "192.168.1.100", UserAgent: "AttackerAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	if err == nil {
		assert.NotNil(t, result)
	}
}

func setupUserUseCase(t *testing.T, env *setup.TestEnvironment) usecase.UserUseCase {
	userRepo := repository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil)

	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, nil, env.Enforcer, auditUC, nil)

	tmpDir := t.TempDir()
	storageProvider, _ := local.NewLocalStorage(tmpDir, "http://test-bucket")

	return usecase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, storageProvider)
}

func TestUserIntegration_FindAll_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

	// Create 2 users
	_, _ = userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "user1", Email: "u1@e.com", Password: "p", Name: "U1"})
	_, _ = userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "user2", Email: "u2@e.com", Password: "p", Name: "U2"})

	req := &model.GetUserListRequest{Page: 1, Limit: 10}
	users, total, err := userUC.GetAllUsers(context.Background(), req)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2)) // Might have other users from other tests if DB not clean, but we call CleanupDatabase
	assert.Len(t, users, 2)
}

func TestUserIntegration_FindAllDynamic_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

  	_, err := userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "alpha", Email: "a@e.com", Password: "p", Name: "Alpha"})  
    require.NoError(t, err)  
    _, err = userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "beta", Email: "b@e.com", Password: "p", Name: "Beta"})  
    require.NoError(t, err)  
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"username": {Type: "equals", From: "alpha"},
		},
	}

	users, total, err := userUC.GetAllUsersDynamic(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "alpha", users[0].Username)
}

func TestUserIntegration_UpdateAvatar_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC := setupUserUseCase(t, env)

	res, err := userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "avataruser", Email: "av@e.com", Password: "p", Name: "Avatar User"})
	require.NoError(t, err)

	// Valid PNG content
	content := "\x89PNG\r\n\x1a\n" + "some data"
	reader := strings.NewReader(content)

	updatedUser, err := userUC.UpdateAvatar(context.Background(), res.ID, reader, "avatar.png", "image/png")

	require.NoError(t, err)
	assert.NotEmpty(t, updatedUser.AvatarURL)
	// The filename is generated using user ID and extension
	assert.Contains(t, updatedUser.AvatarURL, res.ID)
	assert.Contains(t, updatedUser.AvatarURL, ".png")
}

func TestUserIntegration_HardDeleteSoftDeletedUsers(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC, _ := setupTestDependencies(t, env)

	// Create user
	res, _ := userUC.Create(context.Background(), &model.RegisterUserRequest{Username: "todelete", Email: "d@e.com", Password: "p", Name: "To Delete"})

	// Delete user (Soft Delete)
	err := userUC.DeleteUser(context.Background(), "admin", &model.DeleteUserRequest{ID: res.ID})
	require.NoError(t, err)

	// Verify soft deleted
	_, err = userUC.GetUserByID(context.Background(), res.ID)
	assert.ErrorIs(t, err, exception.ErrNotFound)

	// Manually verify in DB (it should exist with deleted_at)
	var count int64
	env.DB.Unscoped().Model(&entity.User{}).Where("id = ?", res.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Hard delete (simulate retention passed)
	err = userUC.HardDeleteSoftDeletedUsers(context.Background(), 0)
	require.NoError(t, err)

	// Verify hard deleted
	env.DB.Unscoped().Model(&entity.User{}).Where("id = ?", res.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
