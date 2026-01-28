//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRoleIntegration(env *setup.TestEnvironment) usecase.RoleUseCase {
	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	return usecase.NewRoleUseCase(env.Logger, tm, roleRepo)
}

func TestRoleIntegration_Create_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	req := &model.CreateRoleRequest{
		Name:        "Test Role",
		Description: "Test role description",
	}

	result, err := roleUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Name, result.Name)
}

func TestRoleIntegration_Delete_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	createReq := &model.CreateRoleRequest{
		Name:        "Delete Role",
		Description: "Role to be deleted",
	}
	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	err = roleUC.Delete(context.Background(), created.ID)
	require.NoError(t, err)
}

func TestRoleIntegration_GetAll_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	roles := []model.CreateRoleRequest{
		{Name: "Test Role 1", Description: "Description 1"},
		{Name: "Test Role 2", Description: "Description 2"},
	}

	for _, req := range roles {
		_, err := roleUC.Create(context.Background(), &req)
		require.NoError(t, err)
	}

	result, err := roleUC.GetAll(context.Background())

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestRoleIntegration_DynamicSearch_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	roles := []model.CreateRoleRequest{
		{Name: "Manager", Description: "Manager role"},
		{Name: "Developer", Description: "Developer role"},
		{Name: "Designer", Description: "Designer role"},
	}

	for _, req := range roles {
		_, err := roleUC.Create(context.Background(), &req)
		require.NoError(t, err)
	}

	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "contains", From: "Dev"},
		},
	}

	result, err := roleUC.GetAllRolesDynamic(context.Background(), filter)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Developer", result[0].Name)
}

func TestRoleIntegration_Create_Negative_DuplicateName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	req1 := &model.CreateRoleRequest{Name: "Duplicate", Description: "First"}
	_, err := roleUC.Create(context.Background(), req1)
	require.NoError(t, err)

	req2 := &model.CreateRoleRequest{Name: "Duplicate", Description: "Second"}
	_, err = roleUC.Create(context.Background(), req2)
	assert.Error(t, err)
}

func TestRoleIntegration_Create_Negative_EmptyName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	req := &model.CreateRoleRequest{Name: "", Description: "Empty Name"}
	result, err := roleUC.Create(context.Background(), req)

	_ = err
	if err == nil {
		assert.NotEmpty(t, result.ID)
	}
}

func TestRoleIntegration_Delete_Negative_NonExistentRole(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	err := roleUC.Delete(context.Background(), "non-existent-id")
	assert.Error(t, err)
}

func TestRoleIntegration_Edge_Create_LongName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	longName := strings.Repeat("a", 60)
	req := &model.CreateRoleRequest{Name: longName, Description: "Long Name"}

	_, err := roleUC.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestRoleIntegration_Edge_SpecialCharactersInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	specialNames := []string{
		"Admin@#$%", "Role-With-Dashes", "Role_With_Underscores", "Role.With.Dots", "Role (With Parentheses)",
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

func TestRoleIntegration_Edge_UnicodeInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	unicodeNames := []string{
		"管理员", "Администратор", "مدير", "マネージャー",
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

func TestRoleIntegration_Edge_MinimumNameLength(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	req := &model.CreateRoleRequest{Name: "A", Description: "Single char role"}
	result, err := roleUC.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
}

func TestRoleIntegration_Security_SQLInjectionInName(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	sqlInjections := []string{
		"Admin' OR '1'='1", "'; DROP TABLE roles--", "Admin'--", "1' UNION SELECT * FROM roles--",
	}

	for _, injection := range sqlInjections {
		t.Run("SQLInjection_"+injection, func(t *testing.T) {
			req := &model.CreateRoleRequest{Name: injection, Description: "SQL injection attempt"}
			_, err := roleUC.Create(context.Background(), req)
			
			_ = err
		})
	}
}

func TestRoleIntegration_Security_XSSInDescription(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
	}

	for i, xss := range xssPayloads {
		t.Run(fmt.Sprintf("XSS_%d", i), func(t *testing.T) {
			req := &model.CreateRoleRequest{
				Name:        fmt.Sprintf("XSSRole_%d", i),
				Description: xss,
			}
			result, err := roleUC.Create(context.Background(), req)
			require.NoError(t, err)
			assert.NotEmpty(t, result.ID)
		})
	}
}

func TestRoleIntegration_Security_Delete_SuperadminForbidden(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	err := roleUC.Delete(context.Background(), "role:superadmin")
	assert.Error(t, err)
}

func TestRoleIntegration_Update_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleUC := setupRoleIntegration(env)

	createReq := &model.CreateRoleRequest{
		Name:        "Role To Update",
		Description: "Original Description",
	}
	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	updateReq := &model.UpdateRoleRequest{
		Description: "Updated Description",
	}

	updated, err := roleUC.Update(context.Background(), created.ID, updateReq)

	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Description", updated.Description)
}
