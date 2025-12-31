//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessIntegration_Positive_CreateAccessRight_And_List(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)

	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "User Management", Description: "Manage users"})
	require.NoError(t, err)
	assert.NotEmpty(t, ar.ID)

	list, err := uc.GetAllAccessRights(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list.Data), 1)
}

func TestAccessIntegration_Positive_CreateEndpoint_LinkToAccessRight(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)

	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Role Management", Description: "Manage roles"})
	require.NoError(t, err)

	ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/api/v1/roles", Method: "GET"})
	require.NoError(t, err)
	assert.NotEmpty(t, ep.ID)

	err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
	require.NoError(t, err)

	list, err := uc.GetAllAccessRights(context.Background())
	require.NoError(t, err)
	foundLinked := false
	for _, item := range list.Data {
		if item.ID == ar.ID {
			for _, e := range item.Endpoints {
				if e.ID == ep.ID {
					foundLinked = true
				}
			}
		}
	}
	assert.True(t, foundLinked)
}

func TestAccessIntegration_Negative_CreateAccessRight_DuplicateName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)

	_, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Dup", Description: "d"})
	require.NoError(t, err)

	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Dup", Description: "d"})
	assert.Error(t, err)
	assert.Nil(t, ar)
}

func TestAccessIntegration_Negative_DeleteAccessRight_NotFound(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)
	err := uc.DeleteAccessRight(context.Background(), "does-not-exist")
	assert.Error(t, err)
}

func TestAccessIntegration_Edge_DeleteEndpoint_NonExistent_IsNoop(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)
	err := uc.DeleteEndpoint(context.Background(), "does-not-exist")
	assert.NoError(t, err)
}

func TestAccessIntegration_Positive_DynamicSearch_AccessRights(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)

	_, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "User Management", Description: "Manage users"})
	require.NoError(t, err)
	_, err = uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Role Management", Description: "Manage roles"})
	require.NoError(t, err)

	filter := &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{"name": {Type: "contains", From: "User"}}}
	list, err := uc.GetAccessRightsDynamic(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, list.Data, 1)
	assert.Equal(t, "User Management", list.Data[0].Name)
}

func TestAccessIntegration_Security_SQLInjectionInName(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	uc := setupAccessUseCase(env)

	payload := "name' OR '1'='1"
	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: payload, Description: "d"})
	if err == nil {
		assert.NotNil(t, ar)
		assert.Equal(t, payload, ar.Name)
	} else {
		assert.Error(t, err)
	}
}

func setupAccessUseCase(env *setup.TestEnvironment) usecase.IAccessUseCase {
	repo := repository.NewAccessRepository(env.DB, env.Logger)
	return usecase.NewAccessUseCase(repo, env.Logger)
}
