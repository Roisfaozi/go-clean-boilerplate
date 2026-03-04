package test

import (
	"context"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetUserByID_ContextCancellation tests that context cancellation is propagated to the repository.
func TestGetUserByID_ContextCancellation(t *testing.T) {
	// Setup dependencies
	mockRepo, _, _, _, _, _, _, uc := setupTestUserUseCase()

	// Create a context that is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Setup expectations
	// The repository should be called with a context.
	// We can't strictly assert "ctx.Err() != nil" inside the mock matching easily without a custom matcher,
	// but we can assert that the context passed IS the same context.
	expectedErr := context.Canceled

	mockRepo.On("FindByID", mock.MatchedBy(func(c context.Context) bool {
		return c == ctx
	}), "user-123").Return(nil, expectedErr)

	// Execute
	result, err := uc.GetUserByID(ctx, "user-123")

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Helper setup function (reused from user_usecase_test.go logic if available, but simplified here for isolation)
// Assuming standard mocks are available in internal/modules/user/test/mocks based on previous file listings
func setupTestUserUseCase() (
	*mocks.MockUserRepository,
	interface{}, // DB generic
	interface{}, // Enforcer generic
	interface{}, // Audit generic
	interface{}, // Auth generic
	interface{}, // Storage generic
	*logrus.Logger,
	usecase.UserUseCase,
) {
	mockRepo := new(mocks.MockUserRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.FatalLevel)

	// We need nil/mock placeholders for other deps to construct the usecase
	// Since GetUserByID only needs Repo and Logger, others can be nil or simple mocks if NewUserUseCase enforces non-nil.
	// Looking at user_usecase.go, NewUserUseCase takes interfaces.
	// checking if it panics on nil. The implementation struct just assigns them.
	// u.Repo.FindByID is called.

	// However, we need to respect the constructor signature.
	// NewUserUseCase(db, log, repo, enforcer, audit, auth, storage)

	// We might need to mock these if NewUserUseCase checks them, or if we want to be safe.
	// Based on previous files, I'll use simple nil or new() for interfaces if possible,
	// or valid mocks if I need to import them.
	// For this specific test, we only access Repo.

	// Re-using the MockTransactionManager from before would be good if available, or just nil if not used in GetUserByID.
	// GetUserByID does NOT use transaction manager (lines 144-160 of user_usecase.go).

	// So we pass nil for others.

	uc := usecase.NewUserUseCase(nil, logger, mockRepo, nil, nil, nil, nil)

	return mockRepo, nil, nil, nil, nil, nil, logger, uc
}
