package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserUseCase_Update_Security_UsernameSanitization(t *testing.T) {
	deps, uc := setupUserTest()

	// Input with HTML tags
	inputUsername := "<b>bold</b>"
	// Expected stored username (sanitized)
	expectedUsername := "&lt;b&gt;bold&lt;/b&gt;"

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: inputUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
	}

	// Mock: Find user by ID
	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Mock: Check uniqueness
	// The usecase should sanitize BEFORE checking uniqueness to ensure we check the actual value to be stored.
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	// Mock: Transaction
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Mock: Update
	// Expect the USER passed to update to have the SANITIZED username
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	// Mock: Audit
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	// Execute
	_, err := uc.Update(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
