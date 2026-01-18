package test

import (
	"context"
	"errors"
	"io"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// setupCleanupTest creates test dependencies for cleanup tests
func setupCleanupTest() (*userTestDeps, usecase.UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: new(permMocks.IEnforcer),
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_Success(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 30

	// Mock HardDeleteSoftDeletedUsers - Success
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_HardDeleteSoftDeletedUsers_NoRecordsToDelete(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 90

	// Mock HardDeleteSoftDeletedUsers - No records found, but no error
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

// ============================================================================
// ❌ NEGATIVE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_DatabaseError(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 30

	// Mock HardDeleteSoftDeletedUsers - Database error
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).
		Return(errors.New("database connection lost"))

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_HardDeleteSoftDeletedUsers_InvalidRetentionDays(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	// Negative retention days
	retentionDays := -10

	// Mock HardDeleteSoftDeletedUsers - Should handle invalid input
	// Note: Current implementation doesn't validate, but repository might
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).
		Return(errors.New("invalid retention days"))

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
}

// ============================================================================
// 🔄 EDGE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_ZeroRetentionDays(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	// Zero retention days - delete all soft-deleted users immediately
	retentionDays := 0

	// Mock HardDeleteSoftDeletedUsers - Should work with 0 days
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
