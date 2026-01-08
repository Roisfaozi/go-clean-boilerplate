package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUseCase is a mock implementation of usecase.UserUseCase
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) GetUserByID(ctx context.Context, id string) (*model.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) Current(ctx context.Context, request *model.GetUserRequest) (*model.UserResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) Update(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) UpdateStatus(ctx context.Context, userID, status string) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserUseCase) DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error {
	args := m.Called(ctx, actorUserID, request)
	return args.Error(0)
}

func TestUserController_GetAllUsers_Validation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUsecase := new(MockUserUseCase)
	logger := logrus.New()
	validate := validator.New()
	controller := NewUserController(mockUsecase, logger, validate)

	router := gin.New()
	router.GET("/users", controller.GetAllUsers)

	// Case 1: Excessive Limit (should fail validation)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users?limit=1001", nil)
	router.ServeHTTP(w, req)

	// Assertions for FIXED state
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code) // Validation error
	mockUsecase.AssertNotCalled(t, "GetAllUsers")
}
