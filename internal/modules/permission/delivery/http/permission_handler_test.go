package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	permHandler "github.com/Roisfaozi/casbin-db/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestGrantPermission_Success(t *testing.T) {
	// Setup
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionHandler(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	// In the original handler, GrantPermission returns a 200 OK with a message.
	// For a resource creation/permission grant, 201 Created might be more appropriate.
	// We will assume the handler is changed to return 201, or we adjust the test.
	// Let's test for 200 as per the original `Success` function.
	// UPDATE: The original `GrantPermission` in the handler uses `response.Success` which is 200 OK.
	// Let's change the handler to use `response.Created` which is 201, which is more semantically correct.
	// And then test for 201. Oh wait, I can't change the handler code, I must adapt the test.
	// The handler uses `response.Success`, which calls `SuccessResponse` with `http.StatusOK`.
	// The test for `TestGrantPermission_Success` in `permission_handler_test.go` has a bug where it asserts for `responseBody["message"]` directly.
	// The `SuccessResponse` function wraps the data in a `WebResponse` struct, so the JSON is `{"data": {"message": "..."}}`.
	// The test needs to be corrected to look inside the "data" object.
	router.POST("/permissions/grant", handler.GrantPermission)

	// Mocking
	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
	}
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method).Return(nil)

	// Create request
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	// The handler's GrantPermission uses response.Success which returns http.StatusOK (200)
	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	// Correctly access the nested data object
	data, ok := responseBody["data"].(map[string]interface{})
	assert.True(t, ok, "Response should have a 'data' object")
	assert.Equal(t, "Permission granted successfully", data["message"])

	mockUseCase.AssertExpectations(t)
}

func TestGrantPermission_InvalidBody(t *testing.T) {
	// Setup
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionHandler(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	// Create request with invalid body
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(`{"role": "editor",`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "GrantPermissionToRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGrantPermission_UseCaseError(t *testing.T) {
	// Setup
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionHandler(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	// Mocking
	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
	}
	mockError := errors.New("use case failed")
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method).Return(mockError)

	// Create request
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
