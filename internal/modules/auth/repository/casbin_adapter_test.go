package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type casbinAdapterTestDeps struct {
	Enforcer *mocks.MockIEnforcer
}

func setupCasbinAdapterTest() (*casbinAdapterTestDeps, repository.AuthzManager) {
	deps := &casbinAdapterTestDeps{
		Enforcer: new(mocks.MockIEnforcer),
	}
	adapter := repository.NewCasbinAdapter(deps.Enforcer, "", "")
	return deps, adapter
}

func TestCasbinAdapter_AssignDefaultRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deps, adapter := setupCasbinAdapterTest()
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", []interface{}{"user-1", "role:user", "global"}).Return(true, nil)

		err := adapter.AssignDefaultRole(context.Background(), "user-1")

		assert.NoError(t, err)
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		deps, adapter := setupCasbinAdapterTest()
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", []interface{}{"user-1", "role:user", "global"}).Return(false, errors.New("db error"))

		err := adapter.AssignDefaultRole(context.Background(), "user-1")

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("nil enforcer", func(t *testing.T) {
		adapter := repository.NewCasbinAdapter(nil, "", "")
		err := adapter.AssignDefaultRole(context.Background(), "user-1")

		assert.NoError(t, err)
	})
}

func TestCasbinAdapter_GetRolesForUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deps, adapter := setupCasbinAdapterTest()
		deps.Enforcer.On("GetRolesForUser", "user-1", []string{"global"}).Return([]string{"role:admin", "role:user"}, nil)

		roles, err := adapter.GetRolesForUser(context.Background(), "user-1", "")

		assert.NoError(t, err)
		assert.Equal(t, []string{"role:admin", "role:user"}, roles)
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("with domain", func(t *testing.T) {
		deps, adapter := setupCasbinAdapterTest()
		deps.Enforcer.On("GetRolesForUser", "user-1", []string{"custom_domain"}).Return([]string{"role:custom"}, nil)

		roles, err := adapter.GetRolesForUser(context.Background(), "user-1", "custom_domain")

		assert.NoError(t, err)
		assert.Equal(t, []string{"role:custom"}, roles)
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		deps, adapter := setupCasbinAdapterTest()
		deps.Enforcer.On("GetRolesForUser", "user-1", []string{"global"}).Return(nil, errors.New("enforcer error"))

		roles, err := adapter.GetRolesForUser(context.Background(), "user-1", "")

		assert.Error(t, err)
		assert.Nil(t, roles)
		assert.Equal(t, "enforcer error", err.Error())
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("nil enforcer", func(t *testing.T) {
		adapter := repository.NewCasbinAdapter(nil, "", "")
		roles, err := adapter.GetRolesForUser(context.Background(), "user-1", "")

		assert.NoError(t, err)
		assert.Nil(t, roles)
	})
}
