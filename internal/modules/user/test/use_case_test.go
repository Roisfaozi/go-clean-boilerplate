package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/mocking"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserUseCase_Create_Success(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	// Create test data
	testID := "test123"
	testReq := &model.RegisterUserRequest{
		ID:       testID,
		Name:     "Test User",
		Password: "password123",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return((*entity.User)(nil), gorm.ErrRecordNotFound)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(testReq.Password))
		return u.ID == testID && u.Name == testReq.Name && err == nil
	})).Return(nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Create(context.Background(), testReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testID, result.ID)
	assert.Equal(t, "Test User", result.Name)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Create_UserAlreadyExists(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "test123"
	testReq := &model.RegisterUserRequest{
		ID:       testID,
		Name:     "Test User",
		Password: "password123",
	}

	existingUser := &entity.User{
		ID:   testID,
		Name: "Existing User",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return(existingUser, nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Create(context.Background(), testReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrConflict, err)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "test123"
	expectedUser := &entity.User{
		ID:   testID,
		Name: "Test User",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return(expectedUser, nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.GetUserByID(context.Background(), testID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testID, result.ID)
	assert.Equal(t, "Test User", result.Name)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_GetUserByID_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "non-existent"

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return((*entity.User)(nil), gorm.ErrRecordNotFound)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.GetUserByID(context.Background(), testID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, exception.ErrNotFound)) // FIXED

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Current_Success(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "test123"
	testReq := &model.GetUserRequest{
		ID: testID,
	}

	expectedUser := &entity.User{
		ID:   testID,
		Name: "Test User",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return(expectedUser, nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Current(context.Background(), testReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testID, result.ID)
	assert.Equal(t, "Test User", result.Name)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Update_Success(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "test123"
	testReq := &model.UpdateUserRequest{
		ID:   testID,
		Name: "Updated Name",
	}

	existingUser := &entity.User{
		ID:   testID,
		Name: "Original Name",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return(existingUser, nil)

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == testID && u.Name == testReq.Name
	})).Return(nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Update(context.Background(), testReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testID, result.ID)
	assert.Equal(t, "Updated Name", result.Name)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Create_InvalidRequest(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	// Create invalid test data (missing required fields)
	testReq := &model.RegisterUserRequest{
		ID:       "", // Empty ID is invalid
		Name:     "", // Empty Name is invalid
		Password: "", // Empty Password is invalid
	}

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Create(context.Background(), testReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrBadRequest, err)

	// Verify no repository methods were called for invalid input
	mockRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserUseCase_Update_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "non-existent"
	testReq := &model.UpdateUserRequest{
		ID:   testID,
		Name: "Updated Name",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return((*entity.User)(nil), gorm.ErrRecordNotFound)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Update(context.Background(), testReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, exception.ErrNotFound)) // FIXED

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Current_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "non-existent"
	testReq := &model.GetUserRequest{
		ID: testID,
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return((*entity.User)(nil), gorm.ErrRecordNotFound)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Current(context.Background(), testReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, exception.ErrNotFound)) // FIXED

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Logout_Success(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "test123"
	testReq := &model.LogoutUserRequest{
		ID: testID,
	}

	existingUser := &entity.User{
		ID:    testID,
		Name:  "Test User",
		Token: "existing-token",
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return(existingUser, nil)

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == testID && u.Token == ""
	})).Return(nil)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Logout(context.Background(), testReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testID, result.ID)

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_Logout_UserNotFound(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)

	testID := "non-existent"
	testReq := &model.LogoutUserRequest{
		ID: testID,
	}

	// Mock expectations
	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args[1].(func(context.Context) error)
			_ = fn(context.Background())
		})

	mockRepo.On("FindByID", mock.Anything, testID).
		Return((*entity.User)(nil), gorm.ErrRecordNotFound)

	// Execute
	uc := usecase.NewUserUseCase(
		logrus.New(),
		validator.New(),
		mockTM,
		mockRepo,
	)

	result, err := uc.Logout(context.Background(), testReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, exception.ErrNotFound)) // FIXED

	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}
