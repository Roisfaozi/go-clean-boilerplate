package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type userTestDeps struct {
	UC *mocks.MockUserUseCase
}

func setupTest() (*userTestDeps, *userHttp.UserController, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	deps := &userTestDeps{UC: new(mocks.MockUserUseCase)}

	log := logrus.New()
	log.SetOutput(io.Discard)

	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)

	controller := userHttp.NewUserController(deps.UC, log, validate)

	r := gin.New()

	return deps, controller, r
}

func TestSetup(t *testing.T) {
	deps, c, r := setupTest()
	uc := deps.UC
	assert.NotNil(t, c)
	assert.NotNil(t, uc)
	assert.NotNil(t, r)
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.POST("/users/register", c.RegisterUser)

		reqBody := model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockRes := &model.UserResponse{
			ID:       "1",
			Username: "testuser",
			Email:    "test@example.com",
			Name:     "Test User",
		}

		uc.EXPECT().Create(mock.Anything, mock.MatchedBy(func(req *model.RegisterUserRequest) bool {
			return req.Username == "testuser"
		})).Return(mockRes, nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		_, c, r := setupTest()
		r.POST("/users/register", c.RegisterUser)

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.POST("/users/register", c.RegisterUser)

		reqBody := model.RegisterUserRequest{
			Username: "", // Invalid username
			Email:    "invalid-email",
			Password: "short",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.POST("/users/register", c.RegisterUser)

		reqBody := model.RegisterUserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}
		jsonBody, _ := json.Marshal(reqBody)

		uc.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, exception.ErrConflict).Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestGetCurrentUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.GetCurrentUser(ctx)
		})

		mockRes := &model.UserResponse{
			ID:       "123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		uc.EXPECT().Current(mock.Anything, &model.GetUserRequest{ID: "123"}).Return(mockRes, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Unauthorized - Missing Context", func(t *testing.T) {
		_, c, r := setupTest()
		r.GET("/users/me", c.GetCurrentUser)

		req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.GetCurrentUser(ctx)
		})

		uc.EXPECT().Current(mock.Anything, &model.GetUserRequest{ID: "123"}).Return(nil, exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PUT("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateUser(ctx)
		})

		reqBody := model.UpdateUserRequest{
			Username: "updateduser",
			Name:     "Updated Name",
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockRes := &model.UserResponse{
			ID:       "123",
			Username: "updateduser",
			Name:     "Updated Name",
		}

		uc.EXPECT().Update(mock.Anything, mock.MatchedBy(func(req *model.UpdateUserRequest) bool {
			return req.ID == "123" && req.Username == "updateduser"
		})).Return(mockRes, nil).Once()

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Unauthorized - Missing Context", func(t *testing.T) {
		_, c, r := setupTest()
		r.PUT("/users/me", c.UpdateUser)

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		_, c, r := setupTest()
		r.PUT("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateUser(ctx)
		})

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.PUT("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateUser(ctx)
		})

		reqBody := model.UpdateUserRequest{
			Username: "sh", // Invalid username length
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PUT("/users/me", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateUser(ctx)
		})

		reqBody := model.UpdateUserRequest{
			Username: "updateduser",
			Name:     "Updated Name",
		}
		jsonBody, _ := json.Marshal(reqBody)

		uc.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestUpdateAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PATCH("/users/me/avatar", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateAvatar(ctx)
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.jpg")
		_, _ = part.Write([]byte("fake image data"))
		_ = writer.Close()

		mockRes := &model.UserResponse{
			ID:        "123",
			AvatarURL: "https://s3.com/test.jpg",
		}

		uc.EXPECT().UpdateAvatar(mock.Anything, "123", mock.Anything, "test.jpg", mock.Anything).Return(mockRes, nil).Once()

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Unauthorized - Missing Context", func(t *testing.T) {
		_, c, r := setupTest()
		r.PATCH("/users/me/avatar", c.UpdateAvatar)

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("No File", func(t *testing.T) {
		_, c, r := setupTest()
		r.PATCH("/users/me/avatar", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateAvatar(ctx)
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		_ = writer.Close() // Empty form data

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("File Too Large", func(t *testing.T) {
		_, c, r := setupTest()
		r.PATCH("/users/me/avatar", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateAvatar(ctx)
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.jpg")
		// Write exactly 2MB + 1 byte
		largeData := strings.Repeat("a", (2*1024*1024)+1)
		_, _ = part.Write([]byte(largeData))
		_ = writer.Close()

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PATCH("/users/me/avatar", func(ctx *gin.Context) {
			ctx.Set("user_id", "123")
			c.UpdateAvatar(ctx)
		})

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.jpg")
		_, _ = part.Write([]byte("fake image data"))
		_ = writer.Close()

		uc.EXPECT().UpdateAvatar(mock.Anything, "123", mock.Anything, "test.jpg", mock.Anything).Return(nil, exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestUpdateUserStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PATCH("/users/:id/status", c.UpdateUserStatus)

		reqBody := model.UpdateUserStatusRequest{
			Status: "suspended",
		}
		jsonBody, _ := json.Marshal(reqBody)

		uc.EXPECT().UpdateStatus(mock.Anything, "123", "suspended").Return(nil).Once()

		req, _ := http.NewRequest(http.MethodPatch, "/users/123/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		_, c, r := setupTest()
		r.PATCH("/users/:id/status", c.UpdateUserStatus)

		req, _ := http.NewRequest(http.MethodPatch, "/users/123/status", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.PATCH("/users/:id/status", c.UpdateUserStatus)

		reqBody := model.UpdateUserStatusRequest{
			Status: "invalid_status", // Not in oneof
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPatch, "/users/123/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.PATCH("/users/:id/status", c.UpdateUserStatus)

		reqBody := model.UpdateUserStatusRequest{
			Status: "banned",
		}
		jsonBody, _ := json.Marshal(reqBody)

		uc.EXPECT().UpdateStatus(mock.Anything, "123", "banned").Return(exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodPatch, "/users/123/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestGetAllUsers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users", c.GetAllUsers)

		mockRes := []*model.UserResponse{
			{ID: "1", Username: "user1"},
			{ID: "2", Username: "user2"},
		}

		uc.EXPECT().GetAllUsers(mock.Anything, mock.MatchedBy(func(req *model.GetUserListRequest) bool {
			return req.Page == 1 && req.Limit == 10
		})).Return(mockRes, int64(2), nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Query Parameters", func(t *testing.T) {
		_, c, r := setupTest()
		r.GET("/users", c.GetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users?page=invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.GET("/users", c.GetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users?page=-1", nil) // Invalid page
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users", c.GetAllUsers)

		uc.EXPECT().GetAllUsers(mock.Anything, mock.Anything).Return(nil, int64(0), exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users/:id", c.GetUserByID)

		mockRes := &model.UserResponse{
			ID:       "123",
			Username: "testuser",
		}

		uc.EXPECT().GetUserByID(mock.Anything, "123").Return(mockRes, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.GET("/users/:id", c.GetUserByID)

		uc.EXPECT().GetUserByID(mock.Anything, "123").Return(nil, exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.DELETE("/users/:id", func(ctx *gin.Context) {
			ctx.Set("user_id", "admin_id")
			c.DeleteUser(ctx)
		})

		uc.EXPECT().DeleteUser(mock.Anything, "admin_id", mock.MatchedBy(func(req *model.DeleteUserRequest) bool {
			return req.ID == "123"
		})).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Unauthorized - Missing Context", func(t *testing.T) {
		_, c, r := setupTest()
		r.DELETE("/users/:id", c.DeleteUser)

		req, _ := http.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.DELETE("/users/:id", func(ctx *gin.Context) {
			ctx.Set("user_id", "admin_id")
			c.DeleteUser(ctx)
		})

		uc.EXPECT().DeleteUser(mock.Anything, "admin_id", mock.Anything).Return(exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestGetUsersDynamic(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.POST("/users/search", c.GetUsersDynamic)

		reqBody := querybuilder.DynamicFilter{
			Page:     1,
			PageSize: 10,
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockRes := []*model.UserResponse{
			{ID: "1", Username: "user1"},
			{ID: "2", Username: "user2"},
		}

		uc.EXPECT().GetAllUsersDynamic(mock.Anything, mock.MatchedBy(func(req *querybuilder.DynamicFilter) bool {
			return req.Page == 1 && req.PageSize == 10
		})).Return(mockRes, int64(2), nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		uc.AssertExpectations(t)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		_, c, r := setupTest()
		r.POST("/users/search", c.GetUsersDynamic)

		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.POST("/users/search", c.GetUsersDynamic)

		reqBody := querybuilder.DynamicFilter{
			Page: -1, // Invalid page
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.POST("/users/search", c.GetUsersDynamic)

		reqBody := querybuilder.DynamicFilter{
			Page:     1,
			PageSize: 10,
		}
		jsonBody, _ := json.Marshal(reqBody)

		uc.EXPECT().GetAllUsersDynamic(mock.Anything, mock.Anything).Return(nil, int64(0), exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		uc.AssertExpectations(t)
	})
}

func TestRoutes(t *testing.T) {
	_, c, r := setupTest()
	group := r.Group("/api/v1")

	userHttp.RegisterPublicRoutes(group, c)
	userHttp.RegisterAuthenticatedRoutes(group, c)
	userHttp.RegisterAuthorizedRoutes(group, c)

	routes := r.Routes()
	var paths []string
	for _, route := range routes {
		paths = append(paths, route.Method+" "+route.Path)
	}

	assert.Contains(t, paths, "POST /api/v1/users/register")
	assert.Contains(t, paths, "GET /api/v1/users/me")
	assert.Contains(t, paths, "PUT /api/v1/users/me")
	assert.Contains(t, paths, "PATCH /api/v1/users/me/avatar")
	assert.Contains(t, paths, "GET /api/v1/users")
	assert.Contains(t, paths, "POST /api/v1/users/search")
	assert.Contains(t, paths, "GET /api/v1/users/:id")
	assert.Contains(t, paths, "PATCH /api/v1/users/:id/status")
	assert.Contains(t, paths, "DELETE /api/v1/users/:id")
}

func TestVulnerability_RegisterUser_XSS(t *testing.T) {
	_, c, r := setupTest()
	r.POST("/users/register", c.RegisterUser)

	reqBody := model.RegisterUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Name:     "<script>alert(1)</script>", // XSS Payload
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Since Name fails XSS validation, it should return UnprocessableEntity (422)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestEdgeCase_RegisterUser_LongUsername(t *testing.T) {
	_, c, r := setupTest()
	r.POST("/users/register", c.RegisterUser)

	reqBody := model.RegisterUserRequest{
		Username: strings.Repeat("a", 101), // Exceeds max length of 100
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestVulnerability_UpdateUser_XSS(t *testing.T) {
	_, c, r := setupTest()
	r.PUT("/users/me", func(ctx *gin.Context) {
		ctx.Set("user_id", "123")
		c.UpdateUser(ctx)
	})

	reqBody := model.UpdateUserRequest{
		Username: "<img src=x onerror=alert(1)>", // XSS Payload
		Name:     "Updated Name",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
