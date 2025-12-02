package test_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	userHandler "github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func newTestUserHandler(mockUseCase *mocks.MockUserUseCase) *userHandler.UserHandler {
	return userHandler.NewUserHandler(mockUseCase, logrus.New(), validator.New())
}

func TestUserHandler_RegisterUser_Success(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	reqBody := &model.RegisterUserRequest{
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Email:    "test@example.com",
	}
	resBody := &model.UserResponse{
		ID:   "user-123",
		Name: "testuser",
	}

	mockUseCase.On("Create", mock.Anything, reqBody).Return(resBody, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestUserHandler_RegisterUser_Conflict(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	reqBody := &model.RegisterUserRequest{
		Username: "existing_user",
		Password: "password123",
		Name:     "Existing User",
		Email:    "existing@example.com",
	}
	mockUseCase.On("Create", mock.Anything, reqBody).Return(nil, exception.ErrConflict)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestUserHandler_RegisterUser_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	// Invalid payload: empty username, short password
	reqBody := &model.RegisterUserRequest{
		Username: "",
		Password: "123", 
		Name:     "Test User",
		Email:    "test@example.com",
	}

	// UseCase.Create should NOT be called because validation fails first
	// No mock setup needed for UseCase

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.GET("/users/me", handler.GetCurrentUser)

	userID := "user-123"
	resBody := &model.UserResponse{
		ID:   userID,
		Name: "currentuser",
	}

	mockUseCase.On("Current", mock.Anything, &model.GetUserRequest{ID: userID}).Return(resBody, nil)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	data, _ := responseBody["data"].(map[string]interface{})
	assert.Equal(t, userID, data["id"])

	mockUseCase.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_NotFound(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.GET("/users/me", handler.GetCurrentUser)

	userID := "not-found-user"
	mockUseCase.On("Current", mock.Anything, &model.GetUserRequest{ID: userID}).Return(nil, exception.ErrNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.GET("/users", handler.GetAllUsers)

	t.Run("Success", func(t *testing.T) {
		expectedUsers := []*model.UserResponse{
			{ID: "user-1", Name: "User One"},
			{ID: "user-2", Name: "User Two"},
		}
		expectedReq := &model.GetUserListRequest{Page: 0, Limit: 0, Username: "", Email: ""}

		mockUseCase.On("GetAllUsers", mock.Anything, expectedReq).Return(expectedUsers, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response response.WebResponseSuccess[[]*model.UserResponse]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, response.Data)

		mockUseCase.AssertExpectations(t)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		expectedReq := &model.GetUserListRequest{Page: 0, Limit: 0, Username: "", Email: ""}
		mockUseCase.On("GetAllUsers", mock.Anything, expectedReq).Return(nil, exception.ErrInternalServer).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserByID(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.GET("/users/:id", handler.GetUserByID)

	t.Run("Success", func(t *testing.T) {
		userID := "user-123"
		expectedUser := &model.UserResponse{ID: userID, Name: "Test User"}

		mockUseCase.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response response.WebResponseSuccess[*model.UserResponse]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, response.Data)

		mockUseCase.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := "not-found-id"
		mockUseCase.On("GetUserByID", mock.Anything, userID).Return(nil, exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodGet, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.DELETE("/users/:id", handler.DeleteUser)

	t.Run("Success", func(t *testing.T) {
		userID := "user-to-delete"
		mockUseCase.On("DeleteUser", mock.Anything, userID).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		userID := "not-found-id"
		mockUseCase.On("DeleteUser", mock.Anything, userID).Return(exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}
