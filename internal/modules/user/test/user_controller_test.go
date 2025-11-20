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
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestUserHandler_RegisterUser_Success(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase) // Corrected mock struct name
	handler := userHandler.NewUserHandler(mockUseCase, logrus.New())
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	reqBody := &model.RegisterUserRequest{
		Name:     "testuser",
		Password: "password123",
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
	mockUseCase := new(mocks.MockUserUseCase) // Corrected mock struct name
	handler := userHandler.NewUserHandler(mockUseCase, logrus.New())
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	reqBody := &model.RegisterUserRequest{
		Name:     "existing_user",
		Password: "password123",
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

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase) // Corrected mock struct name
	handler := userHandler.NewUserHandler(mockUseCase, logrus.New())
	router := setupUserTestRouter()
	router.GET("/users/me", handler.GetCurrentUser)

	userID := "user-123"
	resBody := &model.UserResponse{
		ID:   userID,
		Name: "currentuser",
	}

	mockUseCase.On("Current", mock.Anything, &model.GetUserRequest{ID: userID}).Return(resBody, nil)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)

	// Manually set user_id in context for test
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
	mockUseCase := new(mocks.MockUserUseCase) // Corrected mock struct name
	handler := userHandler.NewUserHandler(mockUseCase, logrus.New())
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
