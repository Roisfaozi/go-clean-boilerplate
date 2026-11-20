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

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type userTestDeps struct {
	UC *mocks.MockUserUseCase
}

func setupTest() (*userTestDeps, *userHttp.UserController, *http.ServeMux) {
	deps := &userTestDeps{UC: new(mocks.MockUserUseCase)}

	log := logrus.New()
	log.SetOutput(io.Discard)

	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)

	controller := userHttp.NewUserController(deps.UC, log, validate)

	r := http.NewServeMux()

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
		r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

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
		r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

		req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

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
		r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

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
		r.HandleFunc("GET /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPGetCurrentUser(w, req.WithContext(ctx))
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
		r.HandleFunc("GET /users/me", c.HTTPGetCurrentUser)

		req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.HandleFunc("GET /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPGetCurrentUser(w, req.WithContext(ctx))
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
		r.HandleFunc("PUT /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateUser(w, req.WithContext(ctx))
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
		r.HandleFunc("PUT /users/me", c.HTTPUpdateUser)

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("PUT /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateUser(w, req.WithContext(ctx))
		})

		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("PUT /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateUser(w, req.WithContext(ctx))
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
		r.HandleFunc("PUT /users/me", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateUser(w, req.WithContext(ctx))
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
		r.HandleFunc("PATCH /users/me/avatar", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateAvatar(w, req.WithContext(ctx))
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
		r.HandleFunc("PATCH /users/me/avatar", c.HTTPUpdateAvatar)

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("No File", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("PATCH /users/me/avatar", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateAvatar(w, req.WithContext(ctx))
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
		r.HandleFunc("PATCH /users/me/avatar", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateAvatar(w, req.WithContext(ctx))
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
		r.HandleFunc("PATCH /users/me/avatar", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
			c.HTTPUpdateAvatar(w, req.WithContext(ctx))
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
		r.HandleFunc("PATCH /users/{id}/status", c.HTTPUpdateUserStatus)

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
		r.HandleFunc("PATCH /users/{id}/status", c.HTTPUpdateUserStatus)

		req, _ := http.NewRequest(http.MethodPatch, "/users/123/status", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("PATCH /users/{id}/status", c.HTTPUpdateUserStatus)

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
		r.HandleFunc("PATCH /users/{id}/status", c.HTTPUpdateUserStatus)

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
		r.HandleFunc("GET /users", c.HTTPGetAllUsers)

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
		r.HandleFunc("GET /users", c.HTTPGetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users?page=invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("GET /users", c.HTTPGetAllUsers)

		req, _ := http.NewRequest(http.MethodGet, "/users?page=-1", nil) // Invalid page
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.HandleFunc("GET /users", c.HTTPGetAllUsers)

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
		r.HandleFunc("GET /users/{id}", c.HTTPGetUserByID)

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
		r.HandleFunc("GET /users/{id}", c.HTTPGetUserByID)

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
		r.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "admin_id")
			c.HTTPDeleteUser(w, req.WithContext(ctx))
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
		r.HandleFunc("DELETE /users/{id}", c.HTTPDeleteUser)

		req, _ := http.NewRequest(http.MethodDelete, "/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, c, r := setupTest()
		uc := deps.UC
		r.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, req *http.Request) {
			ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "admin_id")
			c.HTTPDeleteUser(w, req.WithContext(ctx))
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
		r.HandleFunc("POST /users/search", c.HTTPGetUsersDynamic)

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
		r.HandleFunc("POST /users/search", c.HTTPGetUsersDynamic)

		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, c, r := setupTest()
		r.HandleFunc("POST /users/search", c.HTTPGetUsersDynamic)

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
		r.HandleFunc("POST /users/search", c.HTTPGetUsersDynamic)

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

func TestVulnerability_RegisterUser_XSS(t *testing.T) {
	_, c, r := setupTest()
	r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

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
	r.HandleFunc("POST /users/register", c.HTTPRegisterUser)

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
	r.HandleFunc("PUT /users/me", func(w http.ResponseWriter, req *http.Request) {
		ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "123")
		c.HTTPUpdateUser(w, req.WithContext(ctx))
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
