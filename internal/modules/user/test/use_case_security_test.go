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

	inputUsername := "<script>alert('XSS')</script>User"
	// Expected sanitized output: tags removed or escaped.
	// pkg.SanitizeString uses html.EscapeString.
	expectedUsername := "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;User"

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
		Name:     "Old Name",
	}

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: inputUsername,
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Expect FindByUsername call for uniqueness check
	// The usecase checks uniqueness with the sanitized username?
	// Currently it checks with the raw input because it sanitizes AFTER checking? No.
	// In Create: request.Username = pkg.SanitizeString(request.Username); then check uniqueness.
	// In Update: currently NO sanitization.

	// If sanitization is added, it should check uniqueness using the sanitized username.
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Capture the user passed to Update and verify Username is sanitized
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Update(context.Background(), request)

	// Since we expect the test to FAIL initially (because Update doesn't sanitize),
	// the mock expectation for "Update" with expectedUsername might not be met,
	// or "FindByUsername" might be called with the raw input.
	// However, for reproduction, we write the test asserting correct behavior.

	assert.NoError(t, err)
	if result != nil {
		assert.Equal(t, expectedUsername, result.Username)
	}

	// Verify that Update was called with sanitized username
	deps.Repo.AssertExpectations(t)
}
