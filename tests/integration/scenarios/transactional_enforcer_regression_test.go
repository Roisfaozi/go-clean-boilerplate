//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	e2eSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionalEnforcer_PersistsCompleteGroupingPolicy(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	userID := "user-regression"

	err := tm.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		_, err := env.Enforcer.WithContext(txCtx).AddGroupingPolicy(userID, "role:user", "global")
		return err
	})
	require.NoError(t, err)

	var row struct {
		Ptype string
		V0    string
		V1    string
		V2    string
	}

	err = env.DB.Table("casbin_rule").
		Select("ptype, v0, v1, v2").
		Where("ptype = ? AND v0 = ?", "g", userID).
		Take(&row).Error
	require.NoError(t, err)

	assert.Equal(t, "g", row.Ptype)
	assert.Equal(t, userID, row.V0)
	assert.Equal(t, "role:user", row.V1)
	assert.Equal(t, "global", row.V2)
}

func TestOrganizationUseCase_CreateOrganization_PersistsCompletePolicies(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	user := setup.CreateTestUser(t, env.DB, "orgcase", "orgcase@example.com", "Password123!")
	_, err := env.Enforcer.AddGroupingPolicy(user.ID, "role:user", "global")
	require.NoError(t, err)

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	organizations := orgRepo.NewOrganizationRepository(env.DB)
	members := orgRepo.NewOrganizationMemberRepository(env.DB)
	uc := orgUsecase.NewOrganizationUseCase(env.Logger, tm, organizations, members, nil, env.Enforcer)

	resp, err := uc.CreateOrganization(context.Background(), user.ID, &orgModel.CreateOrganizationRequest{
		Name: "Regression Org",
		Slug: "regression-org",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var rules []struct {
		Ptype string
		V0    string
		V1    string
		V2    string
		V3    string
	}

	err = env.DB.Table("casbin_rule").
		Select("ptype, v0, v1, v2, v3").
		Where("v2 = ? OR v1 = ?", resp.ID, resp.ID).
		Order("id ASC").
		Find(&rules).Error
	require.NoError(t, err)
	require.NotEmpty(t, rules)

	for _, rule := range rules {
		if rule.Ptype == "g" {
			assert.NotEmpty(t, rule.V0)
			assert.NotEmpty(t, rule.V1)
			assert.NotEmpty(t, rule.V2)
		}
	}
}

func TestAuthRegister_DoesNotPersistMalformedGroupingPolicies(t *testing.T) {
	server := e2eSetup.SetupTestServer(t)
	defer server.Cleanup()

	resp := server.Client.POST("/api/v1/auth/register", map[string]any{
		"name":     "Regression User",
		"username": "reg_user_http",
		"email":    "reg_user_http@example.com",
		"password": "Password123!",
	})
	require.Equal(t, 201, resp.StatusCode)

	var rows []struct {
		Ptype string
		V0    string
		V1    string
		V2    string
	}

	err := server.DB.Table("casbin_rule").
		Select("ptype, v0, v1, v2").
		Where("ptype = ?", "g").
		Find(&rows).Error
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	for _, row := range rows {
		t.Logf("grouping row: ptype=%s v0=%q v1=%q v2=%q", row.Ptype, row.V0, row.V1, row.V2)
		assert.NotEmpty(t, row.V0, "malformed grouping policy row: %+v", row)
		assert.NotEmpty(t, row.V1, "malformed grouping policy row: %+v", row)
		assert.NotEmpty(t, row.V2, "malformed grouping policy row: %+v", row)
	}
}
