package test_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	userHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func newTestUserHandler(mockUseCase *mocks.MockUserUseCase) *userHandler.UserController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)
	return userHandler.NewUserController(mockUseCase, log, validate)
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

	jsonBody := `{"username":"testuser","password":"password123","fullname":"Test User","email":"test@example.com"}`

	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString(jsonBody))
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

	jsonBody := `{"username":"existing_user","password":"password123","fullname":"Existing User","email":"existing@example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString(jsonBody))
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

	testCases := []struct {
		name     string
		body     string
		expected int
	}{
		{
			name:     "Empty Username",
			body:     `{"password":"password123","fullname":"Test User","email":"test@example.com"}`,
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "Short Username",
			body:     `{"username":"abc","password":"password123","fullname":"Test User","email":"test@example.com"}`,
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "Short Password",
			body:     `{"username":"testuser","password":"123","fullname":"Test User","email":"test@example.com"}`,
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "Invalid Email",
			body:     `{"username":"testuser","password":"password123","fullname":"Test User","email":"invalid-email"}`,
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "Missing Name",
			body:     `{"username":"testuser","password":"password123","email":"test@example.com"}`,
			expected: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)

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
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	data, _ := responseBody["data"].(map[string]interface{})
	assert.Equal(t, userID, data["id"])

	mockUseCase.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_NotFound(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)

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

		mockUseCase.On("GetAllUsers", mock.Anything, expectedReq).Return(expectedUsers, int64(2), nil).Once()

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
		mockUseCase.On("GetAllUsers", mock.Anything, expectedReq).Return(nil, int64(0), exception.ErrInternalServer).Once()

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

	actorUserID := "admin-id"
	userID := "user-to-delete"

	t.Run("Success", func(t *testing.T) {
		mockUseCase.On("DeleteUser", mock.Anything, actorUserID, mock.MatchedBy(func(req *model.DeleteUserRequest) bool {
			return req.ID == userID
		})).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: userID}}
		c.Set("user_id", actorUserID)

		handler.DeleteUser(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockUseCase.On("DeleteUser", mock.Anything, actorUserID, mock.MatchedBy(func(req *model.DeleteUserRequest) bool {
			return req.ID == userID
		})).Return(exception.ErrNotFound).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: userID}}
		c.Set("user_id", actorUserID)

		handler.DeleteUser(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateStatus(t *testing.T) {
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)
		router := setupUserTestRouter()
		router.PATCH("/users/:id/status", handler.UpdateUserStatus)

		status := "banned"
		mockUseCase.On("UpdateStatus", mock.Anything, userID, status).Return(nil).Once()

		body := `{"status":"banned"}`
		req, _ := http.NewRequest(http.MethodPatch, "/users/"+userID+"/status", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error - Invalid Status", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)
		router := setupUserTestRouter()
		router.PATCH("/users/:id/status", handler.UpdateUserStatus)

		body := `{"status":"invalid-status"}`
		req, _ := http.NewRequest(http.MethodPatch, "/users/"+userID+"/status", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)
		router := setupUserTestRouter()
		router.PATCH("/users/:id/status", handler.UpdateUserStatus)

		status := "active"
		mockUseCase.On("UpdateStatus", mock.Anything, userID, status).Return(exception.ErrNotFound).Once()

		body := `{"status":"active"}`
		req, _ := http.NewRequest(http.MethodPatch, "/users/"+userID+"/status", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		reqBody := &model.UpdateUserRequest{
			ID: userID,
			Username: "newusername",
			Name: "New Name",
		}

		resBody := &model.UserResponse{
			ID: userID,
			Username: "newusername",
			Name: "New Name",
		}

		mockUseCase.On("Update", mock.Anything, reqBody).Return(resBody, nil).Once()

		jsonBody := `{"username":"newusername","name":"New Name"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateUser(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Conflict", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		mockUseCase.On("Update", mock.Anything, mock.Anything).Return(nil, exception.ErrConflict).Once()

		jsonBody := `{"username":"exists","name":"New Name"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateUser(c)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		jsonBody := `{"username":"ab"}` // Too short
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateUser(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Validation Error - Long Name", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		// Name > 100 chars
		longName := ""
		for i := 0; i < 101; i++ {
			longName += "a"
		}
		jsonBody := `{"username":"validuser", "name":"` + longName + `"}`
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateUser(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestUserHandler_GetUsersDynamic(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.POST("/users/search", handler.GetUsersDynamic)

	t.Run("Success", func(t *testing.T) {
		expectedUsers := []*model.UserResponse{{ID: "1", Name: "Test"}}
		mockUseCase.On("GetAllUsersDynamic", mock.Anything, mock.MatchedBy(func (f *querybuilder.DynamicFilter) bool {
            return f != nil
        })).Return(expectedUsers, int64(1), nil).Once()

		jsonBody := `{"filters":{"name":{"type":"contains","from":"Test"}}}`
		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error - Invalid Filter", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)
		router := setupUserTestRouter()
		router.POST("/users/search", handler.GetUsersDynamic)

		jsonBody := `{"page_size": 200}` // Exceeds max 100
		req, _ := http.NewRequest(http.MethodPost, "/users/search", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestUserHandler_UpdateAvatar(t *testing.T) {
	userID := "user-123"

	t.Run("Success", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		expectedRes := &model.UserResponse{ID: userID, AvatarURL: "http://s3/avatar.png"}
		mockUseCase.On("UpdateAvatar", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything).Return(expectedRes, nil).Once()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "avatar.png")

		_, err := part.Write([]byte("image data"))
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateAvatar(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Bad Request - Missing File", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", nil) // No multipart body

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateAvatar(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockUseCase := new(mocks.MockUserUseCase)
		handler := newTestUserHandler(mockUseCase)

		mockUseCase.On("UpdateAvatar", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything).Return(nil, exception.ErrInternalServer).Once()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "avatar.png")
		_, err := part.Write([]byte("image data"))

		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPatch, "/users/me/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		handler.UpdateAvatar(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_RegisterUser_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)
	router := setupUserTestRouter()
	router.POST("/users/register", handler.RegisterUser)

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "XSS in Username",
			body: `{"username":"<script>alert(1)</script>","password":"password123","fullname":"Test User","email":"test@example.com"}`,
		},
		{
			name: "XSS in Fullname",
			body: `{"username":"testuser","password":"password123","fullname":"<img src=x onerror=alert(1)>","email":"test@example.com"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			assert.Contains(t, w.Body.String(), "xss") // Ensure the error is related to XSS validation
		})
	}
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserHandler_UpdateUser_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockUserUseCase)
	handler := newTestUserHandler(mockUseCase)

	userID := "user-123"

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "XSS in Username",
			body: `{"username":"<script>alert(1)</script>"}`,
		},
		{
			name: "XSS in Name",
			body: `{"name":"<img src=x onerror=alert(1)>"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", userID)

			handler.UpdateUser(c)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			assert.Contains(t, w.Body.String(), "xss")
		})
	}
	mockUseCase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
