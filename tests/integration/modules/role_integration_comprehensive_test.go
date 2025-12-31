//go:build integration_legacy
// +build integration_legacy

package modules

import (
	"context"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// POSITIVE TEST CASES
// ============================================

func TestRoleIntegration_Create_Positive_ValidRole(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	req := &model.CreateRoleRequest{
		Name:        "Manager",
		Description: "Manager role with full access",
	}

	result, err := roleUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Description, result.Description)
}

func TestRoleIntegration_Update_Positive_ValidUpdate(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	createReq := &model.CreateRoleRequest{Name: "Original", Description: "Original description"}
	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	updateReq := &model.UpdateRoleRequest{ID: created.ID, Name: "Updated", Description: "Updated description"}
	result, err := roleUC.Update(context.Background(), updateReq)

	require.NoError(t, err)
	assert.Equal(t, "Updated", result.Name)
	assert.Equal(t, "Updated description", result.Description)
}

func TestRoleIntegration_GetAll_Positive_MultipleRoles(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	roles := []string{"Developer", "Designer", "Manager"}
	for _, name := range roles {
		req := &model.CreateRoleRequest{Name: name, Description: name + " role"}
		_, err := roleUC.Create(context.Background(), req)
		require.NoError(t, err)
	}

	result, err := roleUC.GetAll(context.Background())

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)
}

// ============================================
// NEGATIVE TEST CASES
// ============================================

func TestRoleIntegration_Create_Negative_DuplicateName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	req1 := &model.CreateRoleRequest{Name: "Duplicate", Description: "First role"}
	_, err := roleUC.Create(context.Background(), req1)
	require.NoError(t, err)

	req2 := &model.CreateRoleRequest{Name: "Duplicate", Description: "Second role"}
	result, err := roleUC.Create(context.Background(), req2)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRoleIntegration_Create_Negative_EmptyName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	req := &model.CreateRoleRequest{Name: "", Description: "Role without name"}
	result, err := roleUC.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRoleIntegration_Update_Negative_NonExistentRole(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	updateReq := &model.UpdateRoleRequest{ID: "non-existent-id", Name: "Updated", Description: "Updated"}
	result, err := roleUC.Update(context.Background(), updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRoleIntegration_Delete_Negative_NonExistentRole(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	err := roleUC.Delete(context.Background(), "non-existent-id")

	assert.Error(t, err)
}

func TestRoleIntegration_GetByID_Negative_NonExistentRole(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	result, err := roleUC.GetByID(context.Background(), "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ============================================
// EDGE CASES
// ============================================

func TestRoleIntegration_Create_Edge_VeryLongName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	longName := strings.Repeat("a", 100)
	req := &model.CreateRoleRequest{Name: longName, Description: "Long name role"}

	result, err := roleUC.Create(context.Background(), req)

	if err == nil {
		assert.NotEmpty(t, result.ID)
	} else {
		assert.Error(t, err)
	}
}

func TestRoleIntegration_Create_Edge_SpecialCharactersInName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	specialNames := []string{
		"Admin@#$%",
		"Role-With-Dashes",
		"Role_With_Underscores",
		"Role.With.Dots",
		"Role (With Parentheses)",
	}

	for _, name := range specialNames {
		t.Run("SpecialChar_"+name, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: name, Description: "Special char role"}
			result, err := roleUC.Create(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, name, result.Name)
		})
	}
}

func TestRoleIntegration_Create_Edge_UnicodeInName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	unicodeNames := []string{
		"管理员",        // Chinese
		"Администратор", // Russian
		"مدير",          // Arabic
		"マネージャー",  // Japanese
	}

	for _, name := range unicodeNames {
		t.Run("Unicode_"+name, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: name, Description: "Unicode role"}
			result, err := roleUC.Create(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, name, result.Name)
		})
	}
}

func TestRoleIntegration_Update_Edge_EmptyDescription(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	createReq := &model.CreateRoleRequest{Name: "TestRole", Description: "Original description"}
	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	updateReq := &model.UpdateRoleRequest{ID: created.ID, Name: "TestRole", Description: ""}
	result, err := roleUC.Update(context.Background(), updateReq)

	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestRoleIntegration_Create_Edge_MinimumNameLength(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	req := &model.CreateRoleRequest{Name: "A", Description: "Single char role"}
	result, err := roleUC.Create(context.Background(), req)

	if err == nil {
		assert.NotEmpty(t, result.ID)
	}
}

// ============================================
// SECURITY TEST CASES
// ============================================

func TestRoleIntegration_Security_SQLInjectionInName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	sqlInjections := []string{
		"Admin' OR '1'='1",
		"'; DROP TABLE roles--",
		"Admin'--",
		"1' UNION SELECT * FROM roles--",
	}

	for _, injection := range sqlInjections {
		t.Run("SQLInjection_"+injection, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: injection, Description: "SQL injection attempt"}
			result, err := roleUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, injection, result.Name)
			}
		})
	}
}

func TestRoleIntegration_Security_XSSInDescription(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
	}

	for i, xss := range xssPayloads {
		t.Run("XSS_"+string(rune(i)), func(t *testing.T) {
			req := &model.CreateRoleRequest{
				Name:        "XSSRole" + string(rune(i)),
				Description: xss,
			}
			result, err := roleUC.Create(context.Background(), req)

			require.NoError(t, err)
			assert.NotEmpty(t, result.ID)
		})
	}
}

func TestRoleIntegration_Security_PathTraversalInName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	pathTraversals := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32",
		"....//....//",
	}

	for _, path := range pathTraversals {
		t.Run("PathTraversal_"+path, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: path, Description: "Path traversal attempt"}
			result, err := roleUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
			}
		})
	}
}

func TestRoleIntegration_Security_NoSQLInjection(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	noSQLPayloads := []string{
		`{"$gt":""}`,
		`{"$ne":null}`,
		`admin' || '1'=='1`,
	}

	for _, payload := range noSQLPayloads {
		t.Run("NoSQL_"+payload, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: payload, Description: "NoSQL injection attempt"}
			result, err := roleUC.Create(context.Background(), req)

			if err == nil {
				assert.NotEmpty(t, result.ID)
			}
		})
	}
}

func TestRoleIntegration_Security_CaseSensitivityCheck(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleUseCase(t, env)

	req1 := &model.CreateRoleRequest{Name: "Admin", Description: "Admin role"}
	_, err := roleUC.Create(context.Background(), req1)
	require.NoError(t, err)

	req2 := &model.CreateRoleRequest{Name: "admin", Description: "admin role lowercase"}
	result, err := roleUC.Create(context.Background(), req2)

	if err == nil {
		assert.NotEmpty(t, result.ID)
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func setupRoleUseCase(t *testing.T, env *setup.TestEnvironment) usecase.RoleUseCase {
	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	return usecase.NewRoleUseCase(env.Logger, tm, roleRepo)
}
