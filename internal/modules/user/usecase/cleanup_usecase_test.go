package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_Success(t *testing.T) {
	deps, uc := setupUserTest()
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
	deps, uc := setupUserTest()
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
	deps, uc := setupUserTest()
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
	deps, uc := setupUserTest()
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
	deps, uc := setupUserTest()
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
