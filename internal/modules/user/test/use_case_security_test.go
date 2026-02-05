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

func TestUserUseCase_Update_XSS_Sanitization(t *testing.T) {
	deps, uc := setupUserTest()

	// <b> is allowed by validation (pkg/validation/custom_validators.go)
	// but should be sanitized by UseCase to match Create behavior
	inputUsername := "<b>bold</b>"
	expectedUsername := "&lt;b&gt;bold&lt;/b&gt;"

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: inputUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "original",
		Name:     "Original Name",
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Ensure username uniqueness check passes
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// This is the key assertion: verify that Repo.Update receives the sanitized username
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Update(context.Background(), request)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
