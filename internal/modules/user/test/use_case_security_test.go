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

func TestUserUseCase_Update_Security_Sanitization(t *testing.T) {
	deps, uc := setupUserTest()

	// Input with allowed HTML tag by xss validator (<b>) but should be sanitized for storage
	rawUsername := "<b>BoldUser</b>"
	expectedUsername := "&lt;b&gt;BoldUser&lt;/b&gt;"

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: rawUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Expect check against SANITIZED username
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Expect Update with sanitized username
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == "user123" && u.Username == expectedUsername
	})).Return(nil)

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Update(context.Background(), request)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}
