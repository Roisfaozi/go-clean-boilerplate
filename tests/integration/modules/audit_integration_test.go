//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuditIntegration(env *setup.TestEnvironment) auditUseCase.AuditUseCase {
	repo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	return auditUseCase.NewAuditUseCase(repo, env.Logger, nil, nil)
}

func TestAuditIntegration_LogActivity_And_Query(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAuditIntegration(env)

	req := model.CreateAuditLogRequest{
		UserID: "user-1", Action: "CREATE", Entity: "User", EntityID: "entity-1",
		OldValues: map[string]any{"a": 1}, NewValues: map[string]any{"a": 2},
		IPAddress: "127.0.0.1", UserAgent: "TestAgent",
	}
	err := uc.LogActivity(context.Background(), req)
	require.NoError(t, err)

	filter := &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{"entity": {Type: "equals", From: "User"}}}
	logs, _, err := uc.GetLogsDynamic(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "CREATE", logs[0].Action)
	assert.Equal(t, "User", logs[0].Entity)
}

func TestAuditIntegration_LogActivity_MissingRequiredFields(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAuditIntegration(env)

	err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{})
	assert.Error(t, err)
}

func TestAuditIntegration_LogActivity_NilValues(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAuditIntegration(env)

	err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{
		UserID: "user-1", Action: "UPDATE", Entity: "User", EntityID: "entity-1",
		OldValues: nil, NewValues: nil,
	})
	require.NoError(t, err)
}

func TestAuditIntegration_Security_SQLInjectionInEntity(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAuditIntegration(env)

	payload := "User'; DROP TABLE audit_logs;--"
	err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{
		UserID: "user-1", Action: "CREATE", Entity: payload, EntityID: "entity-1",
	})

	if err == nil {
		filter := &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{"entity": {Type: "equals", From: payload}}}
		logs, _, err := uc.GetLogsDynamic(context.Background(), filter)
		require.NoError(t, err)
		assert.NotEmpty(t, logs)
		assert.Equal(t, payload, logs[0].Entity)
	} else {
		assert.Error(t, err)
	}
}
