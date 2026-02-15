package test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_UpdateAvatar_CleanupOnDBFailure(t *testing.T) {
	deps, uc := setupUserTest()
	userID := "user123"
	user := &entity.User{ID: userID}
	expectedURL := "https://storage.com/avatars/user123.png"

	// Mock User Found
	deps.Repo.On("FindByID", mock.Anything, userID).Return(user, nil)

	// Mock Upload Success
	deps.Storage.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, "image/png").Return(expectedURL, nil)

	// Mock Transaction
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Mock DB Update Failure
	deps.Repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	// EXPECT Cleanup: DeleteFile should be called
	deps.Storage.On("DeleteFile", mock.Anything, expectedURL).Return(nil)

	// Create valid image content (needs to be large enough to detect type)
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	pngContent := "\x89PNG\r\n\x1a\n" + strings.Repeat("a", 512)
	file := strings.NewReader(pngContent)

	_, err := uc.UpdateAvatar(context.Background(), userID, file, "avatar.png", "image/png")

	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}
