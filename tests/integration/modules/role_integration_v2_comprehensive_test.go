//go:build integration
// +build integration

package modules

import (
	"context"
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

func TestRoleIntegration_Positive_Create(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	resp, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:test", Description: "desc"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "role:test", resp.Name)
}

func TestRoleIntegration_Negative_Create_DuplicateName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	_, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:dup", Description: "desc"})
	require.NoError(t, err)

	resp, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:dup", Description: "desc2"})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRoleIntegration_Edge_Create_LongName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	longName := "role:" + makeString('a', 60)
	resp, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: longName, Description: "desc"})
	if err == nil {
		assert.NotNil(t, resp)
		assert.Equal(t, longName, resp.Name)
	} else {
		assert.Error(t, err)
		assert.Nil(t, resp)
	}
}

func TestRoleIntegration_Security_Delete_SuperadminForbidden(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	err := uc.Delete(context.Background(), "role:superadmin")
	assert.Error(t, err)
}

func TestRoleIntegration_Positive_GetAll(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	_, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:a", Description: "desc"})
	require.NoError(t, err)
	_, err = uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:b", Description: "desc"})
	require.NoError(t, err)

	roles, err := uc.GetAll(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 2)
}

func TestRoleIntegration_Positive_DynamicSearch(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupRoleUseCaseV2(env)
	_, err := uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:developer", Description: "desc"})
	require.NoError(t, err)
	_, err = uc.Create(context.Background(), &model.CreateRoleRequest{Name: "role:designer", Description: "desc"})
	require.NoError(t, err)

	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "contains", From: "dev"},
		},
	}

	roles, err := uc.GetAllRolesDynamic(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "role:developer", roles[0].Name)
}

func setupRoleUseCaseV2(env *setup.TestEnvironment) usecase.RoleUseCase {
	repo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	return usecase.NewRoleUseCase(env.Logger, tm, repo)
}

func makeString(ch byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
