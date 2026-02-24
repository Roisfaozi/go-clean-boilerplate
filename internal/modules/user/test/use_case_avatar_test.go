package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper to create a reader with valid PNG header
func createValidImageReader(content string) io.Reader {
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return io.MultiReader(strings.NewReader(string(header)), strings.NewReader(content))
}

// setupAvatarTest creates test dependencies for avatar tests
func setupAvatarTest() (*userTestDeps, userUseCase.UserUseCase) {
	mockEnforcer := new(permMocks.IEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := userUseCase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_Success(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-123.png"

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
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-123.png", "image/png").
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
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-456"
	filename := "new-avatar.png"
	contentType := "image/png"
	fileContent := createValidImageReader("new-fake-image-data")
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
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-456.png", "image/png").
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
	deps, uc := setupAvatarTest()
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
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-789"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser3",
		Email:    "test3@example.com",
		Name:     "Test User 3",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Error
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-789.png", "image/png").
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
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-101"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-101.png"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser4",
		Email:    "test4@example.com",
		Name:     "Test User 4",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Success
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-101.png", "image/png").
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
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-202"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-202.png"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser5",
		Email:    "test5@example.com",
		Name:     "Test User 5",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-202.png", "image/png").
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
	deps, uc := setupAvatarTest()
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

	// Mock Storage Upload - Should NOT be called
	// deps.Storage.On("UploadFile", ...).Return(...)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrValidationError, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertNotCalled(t, "UploadFile")
}

func TestUserUseCase_UpdateAvatar_FileTooLarge(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-404"
	filename := "huge-image.jpg"
	contentType := "image/png"
	// Simulate large file with valid header
	largeContent := createValidImageReader(strings.Repeat("x", 10*1024*1024)) // 10MB

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser7",
		Email:    "test7@example.com",
		Name:     "Test User 7",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Should reject file too large
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-404.png", "image/png").
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
