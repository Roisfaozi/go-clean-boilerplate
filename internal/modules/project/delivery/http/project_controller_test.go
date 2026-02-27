package http_test

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
	t.Run("Success", func(t *testing.T) {
		mockUseCase := new(mocks.MockProjectUseCase)
		handler := newTestProjectHandler(mockUseCase)
		router := setupProjectTestRouter()
		router.POST("/projects", handler.Create)

		reqBody := model.CreateProjectRequest{
			Name:   "Test Project",
			Domain: "test-domain",
		}
		resBody := &model.ProjectResponse{
			ID:     "p1",
			Name:   "Test Project",
			Domain: "test-domain",
		}

		mockUseCase.On("CreateProject", mock.Anything, mock.Anything, mock.Anything, reqBody).Return(resBody, nil).Once()

		jsonBody := `{"name":"Test Project", "domain":"test-domain"}`
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Organization-ID", "org-1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error - Empty Name", func(t *testing.T) {
		mockUseCase := new(mocks.MockProjectUseCase)
		handler := newTestProjectHandler(mockUseCase)
		router := setupProjectTestRouter()
		router.POST("/projects", handler.Create)

		jsonBody := `{"name":"", "domain":"test-domain"}`
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Validation Error - Empty Domain", func(t *testing.T) {
		mockUseCase := new(mocks.MockProjectUseCase)
		handler := newTestProjectHandler(mockUseCase)
		router := setupProjectTestRouter()
		router.POST("/projects", handler.Create)

		jsonBody := `{"name":"Test Project", "domain":""}`
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestProjectController_Update_Validation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUseCase := new(mocks.MockProjectUseCase)
		handler := newTestProjectHandler(mockUseCase)
		router := setupProjectTestRouter()
		router.PUT("/projects/:id", handler.Update)

		projectID := "p1"
		reqBody := model.UpdateProjectRequest{
			Name: "Updated Name",
		}
		resBody := &model.ProjectResponse{
			ID:   projectID,
			Name: "Updated Name",
		}

		mockUseCase.On("UpdateProject", mock.Anything, projectID, reqBody).Return(resBody, nil).Once()

		jsonBody := `{"name":"Updated Name"}`
		req, _ := http.NewRequest(http.MethodPut, "/projects/"+projectID, bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}
