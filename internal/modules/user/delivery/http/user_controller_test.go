package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockUserUseCase) GetAllUsers(ctx context.Context, request *model.GetUserListRequest) ([]*model.UserResponse, int64, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.UserResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserUseCase) GetAllUsersDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.UserResponse, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.UserResponse), args.Get(1).(int64), args.Error(2)
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

func (m *MockUserUseCase) UpdateAvatar(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (*model.UserResponse, error) {
	args := m.Called(ctx, userID, file, filename, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error {
	args := m.Called(ctx, retentionDays)
	return args.Error(0)
}

func (m *MockUserUseCase) DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error {
	args := m.Called(ctx, actorUserID, request)
	return args.Error(0)
}

func newTestController() (*UserController, *MockUserUseCase) {
	mockUsecase := new(MockUserUseCase)
	logger := logrus.New()
	validate := validator.New()
	// Register custom validations
	_ = validation.RegisterCustomValidations(validate)
	controller := NewUserController(mockUsecase, logger, validate)
	return controller, mockUsecase
}

func TestUserController_GetAllUsers_Validation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := newTestController()

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

func TestUserController_GetUsersDynamic_Validation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := newTestController()

	router := gin.New()
	router.POST("/users/search", controller.GetUsersDynamic)

	// Case 1: Invalid Sort Direction (should fail if validation exists, but initially will pass or error differently)
	// We want to verify that validation IS enforced.

	invalidSort := []querybuilder.SortModel{
		{ColId: "username", Sort: "INVALID_DIRECTION"},
	}
	filter := querybuilder.DynamicFilter{
		Sort: &invalidSort,
	}
	body, _ := json.Marshal(filter)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users/search", bytes.NewReader(body))

	// We expect NO call to usecase because validation should fail first
	// If validation fails, it returns 422.

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected validation error for invalid sort direction")
	mockUsecase.AssertNotCalled(t, "GetAllUsersDynamic")
}

func TestUserController_RegisterUser_XSS(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := newTestController()

	router := gin.New()
	router.POST("/users/register", controller.RegisterUser)

	// Case 1: XSS in Username
	reqBody := model.RegisterUserRequest{
		Username: "<script>alert('xss')</script>",
		Password: "password123",
		Name:     "Normal Name",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users/register", bytes.NewReader(body))

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected validation error for XSS in Username")
	mockUsecase.AssertNotCalled(t, "Create")

	// Case 2: XSS in Name
	reqBody = model.RegisterUserRequest{
		Username: "validuser",
		Password: "password123",
		Name:     "<img src=x onerror=alert(1)>",
		Email:    "test2@example.com",
	}
	body, _ = json.Marshal(reqBody)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/users/register", bytes.NewReader(body))

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected validation error for XSS in Name")
	mockUsecase.AssertNotCalled(t, "Create")
}
