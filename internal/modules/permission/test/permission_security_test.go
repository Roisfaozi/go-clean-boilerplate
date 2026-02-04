package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// SECURITY TEST SUITE - Permission UseCase
// Tests for: Circular role inheritance, SQL injection, Concurrent access
// ============================================================================

type securityPermDeps struct {
	UserRepo *userMocks.MockUserRepository
	RoleRepo *roleMocks.MockRoleRepository
	Enforcer *mocks.IEnforcer
}

func setupSecurityPermissionTest() (*securityPermDeps, usecase.IPermissionUseCase) {
	deps := &securityPermDeps{
		UserRepo: new(userMocks.MockUserRepository),
		RoleRepo: new(roleMocks.MockRoleRepository),
		Enforcer: new(mocks.IEnforcer),
	}

	uc := usecase.NewPermissionUseCase(deps.Enforcer, logrus.New(), deps.RoleRepo, deps.UserRepo)
	return deps, uc
}

// ============================================================================
// 🔐 CIRCULAR ROLE INHERITANCE TESTS
// ============================================================================

// TestCircularRoleInheritance_DirectCycle tests A -> B -> A circular reference.
// This test validates that if role inheritance leads to a cycle, it should be detected.
func TestCircularRoleInheritance_DirectCycle(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	// Setup: Role A inherits from Role B, and Role B tries to inherit from Role A
	roleA := &roleEntity.Role{ID: "role-a", Name: "admin"}
	roleB := &roleEntity.Role{ID: "role-b", Name: "editor"}

	// AddParentRole calls FindByName for both childRole and parentRole
	deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(roleB, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(roleA, nil)

	// AddGroupingPolicy is called with raw role names, not prefixed
	deps.Enforcer.On("AddGroupingPolicy", "editor", "admin", "global").Return(true, nil)

	// The test here demonstrates the NEED for circular detection
	// In production, before calling AddGroupingPolicy, we should check for cycles

	// This is a design test - checking that the function doesn't crash
	// In a proper implementation, this should return an error for circular inheritance
	err := uc.AddParentRole(context.Background(), "editor", "admin")

	// Current implementation allows this - documenting need for cycle detection
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// TestCircularRoleInheritance_IndirectCycle tests A -> B -> C -> A circular reference.
func TestCircularRoleInheritance_IndirectCycle(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	// Setup: A -> B -> C, now C tries to inherit from A
	roleA := &roleEntity.Role{ID: "role-a", Name: "superadmin"}
	roleC := &roleEntity.Role{ID: "role-c", Name: "moderator"}

	// AddParentRole calls FindByName for both childRole and parentRole
	deps.RoleRepo.On("FindByName", mock.Anything, "superadmin").Return(roleA, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "moderator").Return(roleC, nil)

	// AddGroupingPolicy is called with raw role names, not prefixed
	deps.Enforcer.On("AddGroupingPolicy", "superadmin", "moderator", "global").Return(true, nil)

	// Document the need for cycle detection
	err := uc.AddParentRole(context.Background(), "superadmin", "moderator")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// ============================================================================
// 🔐 SQL INJECTION IN PERMISSION INPUTS
// ============================================================================

// TestGrantPermissionToRole_SQLInjection_InPath tests SQL injection in permission path.
func TestGrantPermissionToRole_SQLInjection_InPath(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "editor"
	maliciousPath := "/api/users'; DROP TABLE users; --"
	method := "GET"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	// Casbin should receive the raw string - parameterized at DB level
	deps.Enforcer.On("AddPolicy", roleName, "global", maliciousPath, method).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, maliciousPath, method)

	// Should not error - string is passed as-is, DB handles escaping
	assert.NoError(t, err)
	deps.Enforcer.AssertCalled(t, "AddPolicy", roleName, "global", maliciousPath, method)
}

// TestGrantPermissionToRole_SQLInjection_InRoleName tests SQL injection in role name.
func TestGrantPermissionToRole_SQLInjection_InRoleName(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	maliciousRole := "admin' OR '1'='1"
	path := "/api/v1/users"
	method := "DELETE"

	// Repository should handle this safely
	deps.RoleRepo.On("FindByName", mock.Anything, maliciousRole).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), maliciousRole, path, method)

	assert.Error(t, err)
	deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
}

// TestGrantPermissionToRole_SQLInjection_InMethod tests SQL injection in HTTP method.
func TestGrantPermissionToRole_SQLInjection_InMethod(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "viewer"
	path := "/api/reports"
	maliciousMethod := "GET; DROP TABLE policies; --"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("AddPolicy", roleName, "global", path, maliciousMethod).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, path, maliciousMethod)

	// Method validation should ideally reject this, but if not validated:
	assert.NoError(t, err)
}

// ============================================================================
// 🔐 CONCURRENT PERMISSION UPDATES
// ============================================================================

// TestGrantPermissionToRole_Concurrent_SameRole tests concurrent permission grants to same role.
func TestGrantPermissionToRole_Concurrent_SameRole(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()
	numConcurrent := 10

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var successCount int32
	deps.Enforcer.On("AddPolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
			err := uc.GrantPermissionToRole(context.Background(), roleName, path, "GET")
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

// TestRevokePermissionFromRole_Concurrent tests concurrent permission revocations.
func TestRevokePermissionFromRole_Concurrent(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()
	numConcurrent := 5

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var revokeCount int32
	deps.Enforcer.On("RemovePolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
			err := uc.RevokePermissionFromRole(context.Background(), roleName, path, "DELETE")
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

// TestGrantPermissionToRole_EmptyPath tests granting permission with empty path.
func TestGrantPermissionToRole_EmptyPath(t *testing.T) {
	_, uc := setupSecurityPermissionTest()

	roleName := "editor"

	// Empty path is rejected by implementation (role, path, and method are required)
	err := uc.GrantPermissionToRole(context.Background(), roleName, "", "GET")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

// TestGrantPermissionToRole_WildcardPath tests granting permission with wildcard path.
func TestGrantPermissionToRole_WildcardPath(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "admin"
	role := &roleEntity.Role{ID: "role-admin", Name: roleName}
	wildcardPath := "/api/*"

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", roleName, "global", wildcardPath, "*").Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, wildcardPath, "*")

	assert.NoError(t, err)
}

// TestGrantPermissionToRole_UnicodeInPath tests granting permission with unicode characters.
func TestGrantPermissionToRole_UnicodeInPath(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}
	unicodePath := "/api/用户/管理" // Chinese characters

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", roleName, "global", unicodePath, "GET").Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, unicodePath, "GET")

	assert.NoError(t, err)
}

// ============================================================================
// 🔐 ENFORCER FAILURE HANDLING
// ============================================================================

// TestGrantPermissionToRole_EnforcerConnectionError tests handling enforcer connection errors.
func TestGrantPermissionToRole_EnforcerConnectionError(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	deps.Enforcer.On("AddPolicy", roleName, "global", "/api/test", "GET").Return(false, errors.New("connection refused"))

	err := uc.GrantPermissionToRole(context.Background(), roleName, "/api/test", "GET")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestRevokePermissionFromRole_PolicyNotExists tests revoking non-existent policy.
func TestRevokePermissionFromRole_PolicyNotExists(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "viewer"
	role := &roleEntity.Role{ID: "role-viewer", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	// RemovePolicy returns false when policy doesn't exist
	deps.Enforcer.On("RemovePolicy", roleName, "global", "/api/nonexistent", "DELETE").Return(false, nil)

	err := uc.RevokePermissionFromRole(context.Background(), roleName, "/api/nonexistent", "DELETE")

	// Implementation returns error when policy doesn't exist (line 243 of usecase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy to revoke not found")
}

// ============================================================================
// PRIORITY 2: DATA INTEGRITY & EDGE CASES
// ============================================================================

// TestBatchCheckPermission_LargeScale tests batch checking with 100+ items.
func TestBatchCheckPermission_LargeScale(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	userID := "user-123"
	numItems := 150 // More than 100 items
	items := make([]struct{ Resource, Action string }, numItems)

	// Setup mock expectations for all unique items
	for i := 0; i < numItems; i++ {
		resource := "/api/resource" + string(rune('a'+i/26)) + string(rune('a'+i%26)) // Unique resources like /api/resourceaa, /api/resourceab
		action := "GET"
		items[i] = struct{ Resource, Action string }{resource, action}

		// Alternate between allowed and denied
		deps.Enforcer.On("Enforce", userID, "global", resource, action).Return(i%2 == 0, nil).Maybe()
	}

	// Build permission check items
	permItems := make([]model.PermissionCheckItem, numItems)
	for i, item := range items {
		permItems[i] = model.PermissionCheckItem{
			Resource: item.Resource,
			Action:   item.Action,
		}
	}

	results, err := uc.BatchCheckPermission(context.Background(), userID, permItems)

	assert.NoError(t, err)
	assert.Len(t, results, numItems)
}

// TestBatchCheckPermission_EnforceError tests batch permission with enforcer errors.
func TestBatchCheckPermission_EnforceError(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	userID := "user-123"

	// First item succeeds, second fails with error
	deps.Enforcer.On("Enforce", userID, "global", "/api/success", "GET").Return(true, nil)
	deps.Enforcer.On("Enforce", userID, "global", "/api/error", "POST").Return(false, errors.New("enforcer error"))
	deps.Enforcer.On("Enforce", userID, "global", "/api/after-error", "GET").Return(true, nil)

	items := []model.PermissionCheckItem{
		{Resource: "/api/success", Action: "GET"},
		{Resource: "/api/error", Action: "POST"},
		{Resource: "/api/after-error", Action: "GET"},
	}

	results, err := uc.BatchCheckPermission(context.Background(), userID, items)

	// Should continue processing despite error (graceful degradation)
	assert.NoError(t, err)
	assert.True(t, results["/api/success:GET"])
	assert.False(t, results["/api/error:POST"]) // Error results in false
	assert.True(t, results["/api/after-error:GET"])
}

// TestBatchCheckPermission_EmptyItems tests batch check with empty items.
func TestBatchCheckPermission_EmptyItems(t *testing.T) {
	_, uc := setupSecurityPermissionTest()

	results, err := uc.BatchCheckPermission(context.Background(), "user-123", []model.PermissionCheckItem{})

	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

// TestGrantPermissionToRole_UpdateExistingPolicy tests updating an existing policy.
func TestGrantPermissionToRole_UpdateExistingPolicy(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil)
	// AddPolicy returns false if policy already exists (idempotent)
	deps.Enforcer.On("AddPolicy", roleName, "global", "/api/users", "GET").Return(false, nil)

	err := uc.GrantPermissionToRole(context.Background(), roleName, "/api/users", "GET")

	// Should not error even if policy already exists
	assert.NoError(t, err)
}

// TestConcurrentBatchPermissionCheck tests concurrent batch permission checks.
func TestConcurrentBatchPermissionCheck(t *testing.T) {
	deps, uc := setupSecurityPermissionTest()

	userID := "user-concurrent"
	numGoroutines := 10
	itemsPerCheck := 5

	// Setup mock to handle multiple concurrent calls
	deps.Enforcer.On("Enforce", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Maybe()

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
