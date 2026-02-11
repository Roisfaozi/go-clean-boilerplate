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

func setupAccessIntegration(env *setup.TestEnvironment) usecase.IAccessUseCase {
	repo := repository.NewAccessRepository(env.DB, env.Logger)
	return usecase.NewAccessUseCase(repo, env.Logger)
}

func TestAccessIntegration_CreateAccessRight_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	req := model.CreateAccessRightRequest{
		Name:        "User Management",
		Description: "Manage users",
	}

	ar, err := uc.CreateAccessRight(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, ar.ID)
	assert.Equal(t, req.Name, ar.Name)
}

func TestAccessIntegration_CreateAccessRight_Fail_Duplicate(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	req := model.CreateAccessRightRequest{Name: "Dup", Description: "d"}
	_, err := uc.CreateAccessRight(context.Background(), req)
	require.NoError(t, err)

	_, err = uc.CreateAccessRight(context.Background(), req)
	assert.Error(t, err)
}

func TestAccessIntegration_DeleteAccessRight_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Del", Description: "d"})
	require.NoError(t, err)

	err = uc.DeleteAccessRight(context.Background(), ar.ID)
	assert.NoError(t, err)
}

func TestAccessIntegration_DeleteAccessRight_Fail_NotFound(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)
	err := uc.DeleteAccessRight(context.Background(), "ghost-id")
	assert.Error(t, err)
}

func TestAccessIntegration_CreateEndpoint_LinkToAccessRight(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	ar, _ := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Roles", Description: "d"})
	ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/api/v1/roles", Method: "GET"})
	require.NoError(t, err)

	err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
	require.NoError(t, err)

	list, _ := uc.GetAllAccessRights(context.Background())
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

func TestAccessIntegration_DeleteEndpoint_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	ep, _ := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/temp", Method: "POST"})
	err := uc.DeleteEndpoint(context.Background(), ep.ID)
	assert.NoError(t, err)
}

func TestAccessIntegration_DynamicSearch_AccessRights(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	_, _ = uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "User Management"})
	_, _ = uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Audit Logs"})

	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "contains", From: "User"},
		},
	}
	list, _, err := uc.GetAccessRightsDynamic(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, list.Data, 1)
	assert.Equal(t, "User Management", list.Data[0].Name)
}

func TestAccessIntegration_Security_SQLInjectionPrevention(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	payload := "name' OR '1'='1"
	ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: payload})
	if err == nil {
		assert.Equal(t, payload, ar.Name)
	}
}
