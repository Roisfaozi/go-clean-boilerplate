//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	e2eSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionalEnforcer_PersistsCompleteGroupingPolicy(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewSQLXTransactionManager(env.SQLXDB, env.Logger)
	userID := "user-regression"

	err := tm.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		_, err := env.Enforcer.WithContext(txCtx).AddGroupingPolicy(userID, "role:user", "global")
		return err
	})
	require.NoError(t, err)

	var row struct {
		Ptype string `db:"ptype"`
		V0    string `db:"v0"`
		V1    string `db:"v1"`
		V2    string `db:"v2"`
	}

	err = env.SQLXDB.Get(&row, "SELECT ptype, v0, v1, v2 FROM casbin_rule WHERE ptype = ? AND v0 = ? LIMIT 1", "g", userID)
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

	tm := tx.NewSQLXTransactionManager(env.SQLXDB, env.Logger)
	organizations := orgRepo.NewOrganizationRepository(env.DB)
	members := orgRepo.NewOrganizationMemberRepository(env.DB)
	roles := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	uc := orgUsecase.NewOrganizationUseCase(env.Logger, tm, organizations, members, nil, env.Enforcer, roles)

	resp, err := uc.CreateOrganization(context.Background(), user.ID, &orgModel.CreateOrganizationRequest{
		Name: "Regression Org",
		Slug: "regression-org",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var rules []struct {
		Ptype string `db:"ptype"`
		V0    string `db:"v0"`
		V1    string `db:"v1"`
		V2    string `db:"v2"`
		V3    string `db:"v3"`
	}

	err = env.SQLXDB.Select(&rules, "SELECT ptype, v0, v1, v2, COALESCE(v3, '') AS v3 FROM casbin_rule WHERE v2 = ? OR v1 = ? ORDER BY id ASC", resp.ID, resp.ID)
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
		Ptype string `db:"ptype"`
		V0    string `db:"v0"`
		V1    string `db:"v1"`
		V2    string `db:"v2"`
	}

	err := server.DB.Select(&rows, "SELECT ptype, v0, v1, v2 FROM casbin_rule WHERE ptype = ?", "g")
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	for _, row := range rows {
		t.Logf("grouping row: ptype=%s v0=%q v1=%q v2=%q", row.Ptype, row.V0, row.V1, row.V2)
		assert.NotEmpty(t, row.V0, "malformed grouping policy row: %+v", row)
		assert.NotEmpty(t, row.V1, "malformed grouping policy row: %+v", row)
		assert.NotEmpty(t, row.V2, "malformed grouping policy row: %+v", row)
	}
}
