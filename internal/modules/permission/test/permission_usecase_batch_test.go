package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupBatchTest creates test dependencies for batch permission tests
func setupBatchTest() (*mocks.IEnforcer, usecase.IPermissionUseCase) {
	enforcer := new(mocks.IEnforcer)
	roleRepo := new(roleMocks.MockRoleRepository)
	userRepo := new(userMocks.MockUserRepository)
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Default behavior for enforcer with context to return itself
	enforcer.On("WithContext", mock.Anything).Return(enforcer)

	uc := usecase.NewPermissionUseCase(enforcer, log, roleRepo, userRepo)

	return enforcer, uc
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestPermissionUseCase_BatchCheckPermission_Success_AllAllowed(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-123"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/users", Action: "POST"},
		{Resource: "/api/roles", Action: "GET"},
	}

	// Mock Enforce - All allowed
	enforcer.On("Enforce", userID, "global", "/api/users", "GET").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/users", "POST").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/roles", "GET").Return(true, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.True(t, results["/api/users:POST"])
	assert.True(t, results["/api/roles:GET"])
	enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_BatchCheckPermission_Success_Mixed(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-456"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},    // Allowed
		{Resource: "/api/users", Action: "DELETE"}, // Denied
		{Resource: "/api/roles", Action: "POST"},   // Allowed
		{Resource: "/api/admin", Action: "GET"},    // Denied
	}

	// Mock Enforce - Mixed results
	enforcer.On("Enforce", userID, "global", "/api/users", "GET").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/users", "DELETE").Return(false, nil)
	enforcer.On("Enforce", userID, "global", "/api/roles", "POST").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/admin", "GET").Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 4, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/users:DELETE"])
	assert.True(t, results["/api/roles:POST"])
	assert.False(t, results["/api/admin:GET"])
	enforcer.AssertExpectations(t)
}

// ❌ NEGATIVE CASES
func TestPermissionUseCase_BatchCheckPermission_EmptyUserID(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := ""
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
	}

	// Mock Enforce - Should still be called with empty userID
	enforcer.On("Enforce", userID, "global", "/api/users", "GET").Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Current implementation doesn't validate empty userID
	// It will just return false for all permissions
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.False(t, results["/api/users:GET"])
	enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_BatchCheckPermission_EmptyItems(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-789"
	items := []model.PermissionCheckItem{} // Empty list

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should return empty map
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
	enforcer.AssertNotCalled(t, "Enforce")
}

// 🔄 EDGE CASES
func TestPermissionUseCase_BatchCheckPermission_LargeItemList(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-101"

	// Create 100 items
	items := make([]model.PermissionCheckItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = model.PermissionCheckItem{
			Resource: "/api/resource",
			Action:   "GET",
		}
		// Mock each enforce call
		enforcer.On("Enforce", userID, "global", "/api/resource", "GET").Return(true, nil).Once()
	}

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Note: All items have same resource:action, so only 1 key in map
	assert.Equal(t, 1, len(results))
	assert.True(t, results["/api/resource:GET"])
	enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_BatchCheckPermission_EnforcerError(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-202"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/roles", Action: "POST"}, // This will error
		{Resource: "/api/admin", Action: "GET"},
	}

	// Mock Enforce - One with error
	enforcer.On("Enforce", userID, "global", "/api/users", "GET").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/roles", "POST").
		Return(false, errors.New("casbin database error"))
	enforcer.On("Enforce", userID, "global", "/api/admin", "GET").Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should not fail, but log error and continue
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/roles:POST"]) // Error treated as false
	assert.False(t, results["/api/admin:GET"])
	enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_BatchCheckPermission_DuplicateItems(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-303"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/users", Action: "GET"}, // Duplicate
		{Resource: "/api/users", Action: "GET"}, // Duplicate
	}

	// Mock Enforce - Will be called 3 times for duplicates
	enforcer.On("Enforce", userID, "global", "/api/users", "GET").Return(true, nil).Times(3)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Map will have only 1 entry (last one wins)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))
	assert.True(t, results["/api/users:GET"])
	enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_BatchCheckPermission_SpecialCharactersInResource(t *testing.T) {
	enforcer, uc := setupBatchTest()
	ctx := context.Background()

	userID := "user-404"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users/:id/profile", Action: "GET"},
		{Resource: "/api/files/*.pdf", Action: "READ"},
		{Resource: "/api/search?query=*", Action: "POST"},
	}

	// Mock Enforce
	enforcer.On("Enforce", userID, "global", "/api/users/:id/profile", "GET").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/files/*.pdf", "READ").Return(true, nil)
	enforcer.On("Enforce", userID, "global", "/api/search?query=*", "POST").Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users/:id/profile:GET"])
	assert.True(t, results["/api/files/*.pdf:READ"])
	assert.False(t, results["/api/search?query=*:POST"])
	enforcer.AssertExpectations(t)
}
