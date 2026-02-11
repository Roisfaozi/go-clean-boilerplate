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

func TestUserUseCase_Update_Sanitization(t *testing.T) {
	deps, uc := setupUserTest()

	// Input with HTML characters that should be escaped if Sanitized
	inputUsername := "user<name>"
	expectedUsername := "user&lt;name&gt;" // pkg.SanitizeString uses html.EscapeString

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: inputUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
	// Mock that new username (sanitized) does not exist
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Capture the user passed to Update and verify Username is sanitized
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Update(context.Background(), request)

	// Since we expect FAILURE (it's not implemented yet), we assert NoError for the call,
	// but the Mock assertion will fail because it receives "user<name>" instead of "user&lt;name&gt;"
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
