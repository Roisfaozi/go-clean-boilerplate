package usecase_test

import (
	"context"
	"testing"

	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionManager is a local mock for testing
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	// Simple implementation that executes the function directly
	return fn(ctx)
}

func setupSecurityRoleTest() (*mocks.MockRoleRepository, *permissionMocks.MockIPermissionUseCase, usecase.RoleUseCase) {
	mockRepo := new(mocks.MockRoleRepository)
	mockPerm := new(permissionMocks.MockIPermissionUseCase)
	logger := logrus.New()
	mockTM := new(MockTransactionManager)

	uc := usecase.NewRoleUseCase(logger, mockTM, mockRepo, mockPerm)
	return mockRepo, mockPerm, uc
}

// TestDeleteRole_SuperadminProtection tests that superadmin role cannot be deleted.
func TestDeleteRole_SuperadminProtection(t *testing.T) {
	repo, _, uc := setupSecurityRoleTest()

	roleID := "role-superadmin-id"
	role := &entity.Role{ID: roleID, Name: "role:superadmin"}

	repo.On("FindByID", mock.Anything, roleID).Return(role, nil)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrForbidden, err)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// TestDeleteRole_SuperadminProtection_CaseSensitive tests bypass attempts with case variations.
// Current implementation uses exact string match, so ensure we know behavior for variations.
func TestDeleteRole_SuperadminProtection_CaseSensitive(t *testing.T) {
	repo, perm, uc := setupSecurityRoleTest()

	// If a role is named "Role:SuperAdmin", it technically isn't "role:superadmin".
	// This test verifies if the system allows deleting it (which is correct behavior if strict).
	// But if the system intends to protect all case variations, this test will fail (or indicate need for fix).
	roleID := "role-fake-superadmin"
	roleName := "Role:SuperAdmin"
	role := &entity.Role{ID: roleID, Name: roleName}

	repo.On("FindByID", mock.Anything, roleID).Return(role, nil)
	repo.On("Delete", mock.Anything, roleID).Return(nil)
	perm.On("DeleteRole", mock.Anything, roleName).Return(nil)

	err := uc.Delete(context.Background(), roleID)

	// Currently expecting success because protection is exact match
	assert.NoError(t, err)
}

// TestUpdateRole_SuperadminProtection tests if updating superadmin is allowed.
func TestUpdateRole_SuperadminProtection(t *testing.T) {
	repo, _, uc := setupSecurityRoleTest()

	roleID := "role-superadmin-id"
	role := &entity.Role{ID: roleID, Name: "role:superadmin"}

	repo.On("FindByID", mock.Anything, roleID).Return(role, nil)

	// Expect Update to be called because currently there is NO protection on Update
	// If protection is added later, this expectation should verify it (expecting error instead).
	// For now, we assert that it succeeds (documenting current behavior).
	repo.On("Update", mock.Anything, role).Return(nil)

	req := &model.UpdateRoleRequest{
		Description: "Updated description",
	}

	_, err := uc.Update(context.Background(), roleID, req)

	assert.NoError(t, err)
}
