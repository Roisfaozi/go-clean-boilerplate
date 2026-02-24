package test_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	projectHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupProjectTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func newTestProjectHandler(mockUseCase *mocks.MockProjectUseCase) *projectHttp.ProjectController {
	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)
	return projectHttp.NewProjectController(mockUseCase, validate)
}

func TestProjectController_Create_Validation(t *testing.T) {
	mockUseCase := new(mocks.MockProjectUseCase)
	handler := newTestProjectHandler(mockUseCase)
	router := setupProjectTestRouter()
	router.POST("/projects", handler.Create)

	mockUseCase.On("CreateProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&model.ProjectResponse{ID: "123"}, nil).Maybe()

	// Case 1: Empty Name
	jsonBody := `{"name": "", "domain": "example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected validation failure for empty name")

	// Case 2: XSS in Name
	jsonBodyXSS := `{"name": "<script>alert(1)</script>", "domain": "example.com"}`
	reqXSS, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(jsonBodyXSS))
	reqXSS.Header.Set("Content-Type", "application/json")

	wXSS := httptest.NewRecorder()
	router.ServeHTTP(wXSS, reqXSS)
	assert.Equal(t, http.StatusUnprocessableEntity, wXSS.Code, "Expected XSS validation failure")
}

func TestProjectController_Update_Validation(t *testing.T) {
	mockUseCase := new(mocks.MockProjectUseCase)
	handler := newTestProjectHandler(mockUseCase)
	router := setupProjectTestRouter()
	router.PUT("/projects/:id", handler.Update)

	mockUseCase.On("UpdateProject", mock.Anything, mock.Anything, mock.Anything).Return(&model.ProjectResponse{ID: "123"}, nil).Maybe()

	// Case 1: Empty Name
	// Currently passes with 200 because omitempty on string value allows "" and UseCase ignores "".
	jsonBody := `{"name": ""}`
	req, _ := http.NewRequest(http.MethodPut, "/projects/123", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected success (ignored) for empty name in update due to omitempty")

	// Case 2: XSS in Name
	// This should FAIL with 422 because XSS content is not empty, so omitempty doesn't skip, and xss validator runs.
	jsonBodyXSS := `{"name": "<img src=x onerror=alert(1)>"}`
	reqXSS, _ := http.NewRequest(http.MethodPut, "/projects/123", bytes.NewBufferString(jsonBodyXSS))
	reqXSS.Header.Set("Content-Type", "application/json")

	wXSS := httptest.NewRecorder()
	router.ServeHTTP(wXSS, reqXSS)
	assert.Equal(t, http.StatusUnprocessableEntity, wXSS.Code, "Expected XSS validation failure in update")
}
