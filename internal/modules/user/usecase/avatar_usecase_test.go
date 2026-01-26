package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_Success(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-123"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-123.jpg"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "", // No existing avatar
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-123.jpg", contentType).
		Return(uploadedURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.AvatarURL == uploadedURL
	})).Return(nil)

	// Mock Audit Log
	deps.AuditUC.On("LogActivity", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "UPDATE_AVATAR" &&
			req.Entity == "User" &&
			req.EntityID == userID
	})).Return(nil)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uploadedURL, result.AvatarURL)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

func TestUserUseCase_UpdateAvatar_Success_ReplaceExisting(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-456"
	filename := "new-avatar.png"
	contentType := "image/png"
	fileContent := strings.NewReader("new-fake-image-data")
	oldAvatarURL := "https://storage.example.com/avatars/user-456-old.jpg"
	newAvatarURL := "https://storage.example.com/avatars/user-456.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser2",
		Email:     "test2@example.com",
		Name:      "Test User 2",
		AvatarURL: oldAvatarURL, // Has existing avatar
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload (replaces old one)
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-456.png", contentType).
		Return(newAvatarURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.AvatarURL == newAvatarURL
	})).Return(nil)

	// Mock Audit Log
	deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(nil)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newAvatarURL, result.AvatarURL)
	assert.NotEqual(t, oldAvatarURL, result.AvatarURL)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

// ============================================================================
// ❌ NEGATIVE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_UserNotFound(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "nonexistent-user"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")

	// Mock FindByID - User not found
	deps.Repo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrNotFound, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertNotCalled(t, "UploadFile")
}

func TestUserUseCase_UpdateAvatar_StorageUploadError(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-789"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser3",
		Email:    "test3@example.com",
		Name:     "Test User 3",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Error
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-789.jpg", contentType).
		Return("", errors.New("storage service unavailable"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.Repo.AssertNotCalled(t, "Update")
}

func TestUserUseCase_UpdateAvatar_DatabaseUpdateError(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-101"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-101.jpg"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser4",
		Email:    "test4@example.com",
		Name:     "Test User 4",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Success
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-101.jpg", contentType).
		Return(uploadedURL, nil)

	// Mock Update - Error
	deps.Repo.On("Update", ctx, mock.Anything).Return(errors.New("database connection lost"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

func TestUserUseCase_UpdateAvatar_AuditLogError(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-202"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-202.jpg"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser5",
		Email:    "test5@example.com",
		Name:     "Test User 5",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-202.jpg", contentType).
		Return(uploadedURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.Anything).Return(nil)

	// Mock Audit Log - Error (should not fail the operation)
	deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(errors.New("audit service down"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert - Should still succeed even if audit fails
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uploadedURL, result.AvatarURL)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

// ============================================================================
// 🔄 EDGE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_InvalidFileType(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-303"
	filename := "malicious.exe"
	contentType := "application/x-msdownload"
	fileContent := strings.NewReader("fake-exe-data")

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser6",
		Email:    "test6@example.com",
		Name:     "Test User 6",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Should reject invalid file type
	deps.Storage.On("UploadFile", ctx, fileContent, "avatars/user-303.exe", contentType).
		Return("", errors.New("invalid file type"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

func TestUserUseCase_UpdateAvatar_FileTooLarge(t *testing.T) {
	deps, uc := setupUserTest()
	ctx := context.Background()

	userID := "user-404"
	filename := "huge-image.jpg"
	contentType := "image/jpeg"
	// Simulate large file
	largeContent := strings.NewReader(strings.Repeat("x", 10*1024*1024)) // 10MB

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser7",
		Email:    "test7@example.com",
		Name:     "Test User 7",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Should reject file too large
	deps.Storage.On("UploadFile", ctx, largeContent, "avatars/user-404.jpg", contentType).
		Return("", errors.New("file size exceeds limit"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, largeContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}
