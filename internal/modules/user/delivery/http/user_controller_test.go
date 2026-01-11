package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *MockUserUseCase) DeleteUser(ctx context.Context, actorUserID string, request *model.DeleteUserRequest) error {
	args := m.Called(ctx, actorUserID, request)
	return args.Error(0)
}

type userControllerTestDeps struct {
	UseCase  *MockUserUseCase
	Logger   *logrus.Logger
	Validate *validator.Validate
}

func setupUserControllerTest() (*userControllerTestDeps, *UserController) {
	v := validator.New()
	_ = validation.RegisterCustomValidations(v) // Register custom validations (xss)

	deps := &userControllerTestDeps{
		UseCase:  new(MockUserUseCase),
		Logger:   logrus.New(),
		Validate: v,
	}
	deps.Logger.SetOutput(io.Discard)
	controller := NewUserController(deps.UseCase, deps.Logger, deps.Validate)
	return deps, controller
}

func TestUserController_RegisterUser(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/users", controller.RegisterUser)

		reqBody := model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "Password123!",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		deps.UseCase.On("Create", mock.Anything, mock.MatchedBy(func(r *model.RegisterUserRequest) bool {
			return r.Username == reqBody.Username
		})).Return(&model.UserResponse{ID: "1", Username: reqBody.Username}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Negative - Validation Error", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/users", controller.RegisterUser)

		// Invalid email
		reqBody := model.RegisterUserRequest{
			Username: "testuser",
			Email:    "invalid-email",
			Password: "Password123!",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "Create")
	})

	t.Run("Edge - Max Length", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/users", controller.RegisterUser)

		// Username > 100 chars (assuming max=100 tag)
		reqBody := model.RegisterUserRequest{
			Username: strings.Repeat("a", 101),
			Email:    "test@example.com",
			Password: "Password123!",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "Create")
	})

	t.Run("Vulnerability - XSS in Input", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/users", controller.RegisterUser)

		reqBody := model.RegisterUserRequest{
			Username: "<script>alert(1)</script>",
			Email:    "test@example.com",
			Password: "Password123!",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Now we expect 422 because of xss tag
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "Create")
	})
}

func TestUserController_UpdateUser(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Set Auth Middleware FIRST
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "1")
			c.Next()
		})

		router.PUT("/users/:id", controller.UpdateUser)

		reqBody := model.UpdateUserRequest{
			Username: "newname",
		}
		jsonBody, _ := json.Marshal(reqBody)

		deps.UseCase.On("Update", mock.Anything, mock.MatchedBy(func(r *model.UpdateUserRequest) bool {
			return r.Username == "newname"
		})).Return(&model.UserResponse{ID: "1", Username: "newname"}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserController_UpdateUserStatus(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PATCH("/users/:id/status", controller.UpdateUserStatus)

		reqBody := model.UpdateUserStatusRequest{
			Status: "active",
		}
		jsonBody, _ := json.Marshal(reqBody)

		deps.UseCase.On("UpdateStatus", mock.Anything, "1", "active").Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/1/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative - Invalid Status", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.PATCH("/users/:id/status", controller.UpdateUserStatus)

		reqBody := model.UpdateUserStatusRequest{
			Status: "INVALID_STATUS",
		}
		jsonBody, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/users/1/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "UpdateStatus")
	})
}

func TestUserController_GetUserByID(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/users/:id", controller.GetUserByID)

		deps.UseCase.On("GetUserByID", mock.Anything, "1").Return(&model.UserResponse{ID: "1", Username: "user"}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/1", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative - Not Found", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/users/:id", controller.GetUserByID)

		deps.UseCase.On("GetUserByID", mock.Anything, "999").Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/999", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserController_DeleteUser(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Mock context setup (user_id in context)
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "admin_id")
			c.Next()
		})

		router.DELETE("/users/:id", controller.DeleteUser)

		deps.UseCase.On("DeleteUser", mock.Anything, "admin_id", mock.MatchedBy(func(r *model.DeleteUserRequest) bool {
			return r.ID == "target_id"
		})).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/users/target_id", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserController_GetCurrentUser(t *testing.T) {
	t.Run("Positive - Success", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Mock context setup
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "my_id")
			c.Next()
		})

		router.GET("/me", controller.GetCurrentUser)

		deps.UseCase.On("Current", mock.Anything, mock.MatchedBy(func(r *model.GetUserRequest) bool {
			return r.ID == "my_id"
		})).Return(&model.UserResponse{ID: "my_id", Username: "me"}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/me", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative - No User ID in Context", func(t *testing.T) {
		_, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/me", controller.GetCurrentUser)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/me", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserController_GetAllUsers_Validation(t *testing.T) {
	t.Run("Edge - Excessive Limit", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/users", controller.GetAllUsers)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users?limit=1001", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "GetAllUsers")
	})
}

func TestUserController_GetUsersDynamic_Validation(t *testing.T) {
	t.Run("Edge - Invalid Sort Direction", func(t *testing.T) {
		deps, controller := setupUserControllerTest()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.POST("/users/search", controller.GetUsersDynamic)

		invalidSort := []querybuilder.SortModel{
			{ColId: "username", Sort: "INVALID_DIRECTION"},
		}
		filter := querybuilder.DynamicFilter{
			Sort: &invalidSort,
		}
		body, _ := json.Marshal(filter)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/search", bytes.NewReader(body))

		router.ServeHTTP(w, req)

		// Validation should fail before hitting usecase
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertNotCalled(t, "GetAllUsersDynamic")
	})
}
