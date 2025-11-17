package usecase

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// setupTest creates an in-memory Casbin enforcer for isolated testing.
func setupTest(t *testing.T) IPermissionUseCase {
	log := logrus.New()
	log.SetOutput(&nullWriter{})

	m, err := model.NewModelFromString(`
	[request_definition]
	r = sub, obj, act
	
	[policy_definition]
	p = sub, obj, act
	
	[role_definition]
	g = _, _
	
	[policy_effect]
	e = some(where (p.eft == allow))
	
	[matchers]
	m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	`)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	return NewPermissionUseCase(enforcer, log)
}

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestGetAllPermissions(t *testing.T) {
	t.Run("Success - Policies Exist", func(t *testing.T) {
		uc := setupTest(t)
		_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy("admin", "/api/v1/users", "GET")
		_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy("user", "/api/v1/users/me", "GET")

		permissions, err := uc.GetAllPermissions()

		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.Len(t, permissions, 2)
		assert.Contains(t, permissions, []string{"admin", "/api/v1/users", "GET"})
		assert.Contains(t, permissions, []string{"user", "/api/v1/users/me", "GET"})
	})

	t.Run("Success - No Policies", func(t *testing.T) {
		uc := setupTest(t)

		permissions, err := uc.GetAllPermissions()

		assert.NoError(t, err)
		assert.Len(t, permissions, 0)
	})
}

func TestGetPermissionsForRole(t *testing.T) {
	uc := setupTest(t)
	_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy("admin", "/api/v1/admin/dashboard", "GET")
	_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy("admin", "/api/v1/admin/users", "POST")
	_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy("user", "/api/v1/users/me", "GET")

	t.Run("Success - Role with Policies", func(t *testing.T) {
		permissions, err := uc.GetPermissionsForRole("admin")

		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.Len(t, permissions, 2)
		assert.Contains(t, permissions, []string{"admin", "/api/v1/admin/dashboard", "GET"})
		assert.Contains(t, permissions, []string{"admin", "/api/v1/admin/users", "POST"})
	})

	t.Run("Success - Role with No Policies", func(t *testing.T) {
		_, _ = uc.(*PermissionUseCase).enforcer.AddRoleForUser("some_user", "guest")

		permissions, err := uc.GetPermissionsForRole("guest")

		assert.NoError(t, err)
		assert.Len(t, permissions, 0)
	})

	t.Run("Success - Non-existent Role", func(t *testing.T) {
		permissions, err := uc.GetPermissionsForRole("non_existent_role")

		assert.NoError(t, err)
		assert.Len(t, permissions, 0)
	})
}

func TestUpdatePermission(t *testing.T) {
	t.Run("Success - Update Existing Permission", func(t *testing.T) {
		uc := setupTest(t)
		oldPolicy := []string{"admin", "/api/v1/old", "GET"}
		newPolicy := []string{"admin", "/api/v1/new", "POST"}
		oldPolicyInterfaces := make([]interface{}, len(oldPolicy))
		for i, v := range oldPolicy {
			oldPolicyInterfaces[i] = v
		}
		_, _ = uc.(*PermissionUseCase).enforcer.AddPolicy(oldPolicyInterfaces...)

		updated, err := uc.UpdatePermission(oldPolicy, newPolicy)

		assert.NoError(t, err)
		assert.True(t, updated)

		hasOld, _ := uc.(*PermissionUseCase).enforcer.HasPolicy(oldPolicyInterfaces...)
		assert.False(t, hasOld)

		newPolicyInterfaces := make([]interface{}, len(newPolicy))
		for i, v := range newPolicy {
			newPolicyInterfaces[i] = v
		}
		hasNew, _ := uc.(*PermissionUseCase).enforcer.HasPolicy(newPolicyInterfaces...)
		assert.True(t, hasNew)
	})

	t.Run("Failure - Old Permission Does Not Exist", func(t *testing.T) {
		uc := setupTest(t)
		oldPolicy := []string{"admin", "/api/v1/non-existent", "GET"}
		newPolicy := []string{"admin", "/api/v1/new", "POST"}

		updated, err := uc.UpdatePermission(oldPolicy, newPolicy)

		assert.Error(t, err)
		assert.False(t, updated)
		assert.Contains(t, err.Error(), "policy to update not found")
	})

	t.Run("Failure - Invalid Input", func(t *testing.T) {
		uc := setupTest(t)
		_, err := uc.UpdatePermission([]string{}, []string{"admin", "/api/v1/new", "POST"})
		assert.Error(t, err)

		_, err = uc.UpdatePermission([]string{"admin", "/api/v1/old", "GET"}, []string{})
		assert.Error(t, err)
	})
}

func TestAssignRoleToUser(t *testing.T) {
	t.Run("Success - Assign new role", func(t *testing.T) {
		uc := setupTest(t)
		userID := "user123"
		role := "editor"

		err := uc.AssignRoleToUser(userID, role)
		assert.NoError(t, err)

		hasRole, err := uc.(*PermissionUseCase).enforcer.HasRoleForUser(userID, role)
		assert.NoError(t, err)
		assert.True(t, hasRole)
	})
}

func TestGrantPermissionToRole(t *testing.T) {
	t.Run("Success - Grant new permission", func(t *testing.T) {
		uc := setupTest(t)
		role := "reporter"
		path := "/api/v1/reports"
		method := "GET"

		err := uc.GrantPermissionToRole(role, path, method)
		assert.NoError(t, err)

		hasPolicy, _ := uc.(*PermissionUseCase).enforcer.HasPolicy(role, path, method)
		assert.True(t, hasPolicy)
	})
}

func TestRevokePermissionFromRole(t *testing.T) {
	t.Run("Success - Revoke existing permission", func(t *testing.T) {
		uc := setupTest(t)
		role := "auditor"
		path := "/api/v1/logs"
		method := "GET"

		_, err := uc.(*PermissionUseCase).enforcer.AddPolicy(role, path, method)
		assert.NoError(t, err)
		hasPolicy, _ := uc.(*PermissionUseCase).enforcer.HasPolicy(role, path, method)
		assert.True(t, hasPolicy)

		err = uc.RevokePermissionFromRole(role, path, method)
		assert.NoError(t, err)

		hasPolicy, _ = uc.(*PermissionUseCase).enforcer.HasPolicy(role, path, method)
		assert.False(t, hasPolicy)
	})

	t.Run("Failure - Revoke non-existent permission", func(t *testing.T) {
		uc := setupTest(t)
		err := uc.RevokePermissionFromRole("ghost", "/dev/null", "READ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy to revoke not found")
	})
}
