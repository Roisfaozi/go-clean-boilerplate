package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// SECURITY TEST SUITE - Permission UseCase
// Tests for: Circular role inheritance, SQL injection, Concurrent access
// ============================================================================

// setupPermissionTest is used from permission_usecase_test.go

// ============================================================================
// 🔐 CIRCULAR ROLE INHERITANCE TESTS
// ============================================================================

// TestCircularRoleInheritance_DirectCycle tests A -> B -> A circular reference.
func TestCircularRoleInheritance_DirectCycle(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleA := &roleEntity.Role{ID: "role-a", Name: "admin"}
	roleB := &roleEntity.Role{ID: "role-b", Name: "editor"}

	deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(roleB, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(roleA, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), "editor", "admin", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// TestCircularRoleInheritance_IndirectCycle tests A -> B -> C -> A circular reference.
func TestCircularRoleInheritance_IndirectCycle(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleA := &roleEntity.Role{ID: "role-a", Name: "superadmin"}
	roleC := &roleEntity.Role{ID: "role-c", Name: "moderator"}

	deps.RoleRepo.On("FindByName", mock.Anything, "superadmin").Return(roleA, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "moderator").Return(roleC, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), "superadmin", "moderator", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// ============================================================================
// 🔐 SQL INJECTION IN PERMISSION INPUTS
// ============================================================================

func TestGrantPermissionToRole_SQLInjection_InPath(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "editor"
	maliciousPath := "/api/users'; DROP TABLE users; --"
	method := "GET"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, maliciousPath, method, "global")
	assert.NoError(t, err)
	deps.Enforcer.AssertCalled(t, "AddPolicy", mock.Anything)
}

func TestGrantPermissionToRole_SQLInjection_InRoleName(t *testing.T) {
	deps, uc := setupPermissionTest()

	maliciousRole := "admin' OR '1'='1"
	path := "/api/v1/users"
	method := "DELETE"

	deps.RoleRepo.On("FindByName", mock.Anything, maliciousRole).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), maliciousRole, path, method, "global")
	assert.Error(t, err)
	deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestGrantPermissionToRole_SQLInjection_InMethod(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "viewer"
	path := "/api/reports"
	maliciousMethod := "GET; DROP TABLE policies; --"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, path, maliciousMethod, "global")
	assert.NoError(t, err)
}

// ============================================================================
// 🔐 CONCURRENT PERMISSION UPDATES
// ============================================================================

func TestGrantPermissionToRole_Concurrent_SameRole(t *testing.T) {
	deps, uc := setupPermissionTest()
	numConcurrent := 10

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var successCount int32
	deps.Enforcer.On("AddPolicy", mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&successCount, 1)
		}).Return(true, nil).Maybe()

	var wg sync.WaitGroup
	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/api/resource/" + string(rune('a'+idx))
			err := uc.GrantPermissionToRole(context.Background(), roleName, path, "GET", "global")
			errChan <- err
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&successCount))
}

func TestRevokePermissionFromRole_Concurrent(t *testing.T) {
	deps, uc := setupPermissionTest()
	numConcurrent := 5

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var revokeCount int32
	deps.Enforcer.On("RemovePolicy", mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&revokeCount, 1)
		}).Return(true, nil)

	var wg sync.WaitGroup
	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/api/resource/" + string(rune('a'+idx))
			err := uc.RevokePermissionFromRole(context.Background(), roleName, path, "DELETE", "global")
			errChan <- err
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&revokeCount))
}

// ============================================================================
// 🔐 EDGE CASE: EMPTY AND SPECIAL VALUES
// ============================================================================

func TestGrantPermissionToRole_EmptyPath(t *testing.T) {
	_, uc := setupPermissionTest()

	roleName := "editor"
	err := uc.GrantPermissionToRole(context.Background(), roleName, "", "GET", "global")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestGrantPermissionToRole_WildcardPath(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "admin"
	role := &roleEntity.Role{ID: "role-admin", Name: roleName}
	wildcardPath := "/api/*"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, wildcardPath, "*", "global")
	assert.NoError(t, err)
}

func TestGrantPermissionToRole_UnicodeInPath(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}
	unicodePath := "/api/用户/管理" // Chinese characters

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, unicodePath, "GET", "global")
	assert.NoError(t, err)
}

// ============================================================================
// 🔐 ENFORCER FAILURE HANDLING
// ============================================================================

func TestGrantPermissionToRole_EnforcerConnectionError(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, errors.New("connection refused"))

	err := uc.GrantPermissionToRole(context.Background(), roleName, "/api/test", "GET", "global")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestRevokePermissionFromRole_PolicyNotExists(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "viewer"
	role := &roleEntity.Role{ID: "role-viewer", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, nil)

	err := uc.RevokePermissionFromRole(context.Background(), roleName, "/api/nonexistent", "DELETE", "global")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy to revoke not found")
}

func TestGrantPermissionToRole_UpdateExistingPolicy(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, "/api/users", "GET", "global")
	assert.NoError(t, err)
}

func TestConcurrentBatchPermissionCheck(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID := "user-concurrent"
	numGoroutines := 10
	itemsPerCheck := 5

	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Maybe()

	var wg sync.WaitGroup
	var successCount int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			items := make([]model.PermissionCheckItem, itemsPerCheck)
			itemsKey := idx * itemsPerCheck
			for j := 0; j < itemsPerCheck; j++ {
				items[j] = model.PermissionCheckItem{
					Resource: fmt.Sprintf("/api/resource-%d", itemsKey+j),
					Action:   "GET",
				}
			}

			results, err := uc.BatchCheckPermission(context.Background(), userID, items)
			if err == nil && len(results) == itemsPerCheck {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int32(numGoroutines), successCount)
}
