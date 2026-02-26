package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	projectHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
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

func TestProjectController_Create(t *testing.T) {
	mockUseCase := new(mocks.MockProjectUseCase)
	handler := newTestProjectHandler(mockUseCase)
	router := setupProjectTestRouter()
	router.POST("/projects", handler.Create)

	t.Run("Success", func(t *testing.T) {
		reqBody := model.CreateProjectRequest{
			Name:   "My Project",
			Domain: "myproject.com",
		}
		resBody := &model.ProjectResponse{
			ID:     "proj-1",
			Name:   "My Project",
			Domain: "myproject.com",
		}

		mockUseCase.On("CreateProject", mock.Anything, "user-1", "org-1", reqBody).Return(resBody, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Inject Organization ID via Context
		ctx := database.SetOrganizationContext(req.Context(), "org-1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user-1") // Set user_id in Gin context

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error - Missing Fields", func(t *testing.T) {
		reqBody := model.CreateProjectRequest{
			Name: "", // Missing Name
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "CreateProject")
	})

	t.Run("Validation Error - XSS in Name", func(t *testing.T) {
		reqBody := model.CreateProjectRequest{
			Name:   "<script>alert(1)</script>",
			Domain: "valid.com",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "CreateProject")
	})

	t.Run("Validation Error - Name Too Long", func(t *testing.T) {
		longName := ""
		for i := 0; i < 101; i++ {
			longName += "a"
		}
		reqBody := model.CreateProjectRequest{
			Name:   longName,
			Domain: "valid.com",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "CreateProject")
	})
}

func TestProjectController_Update(t *testing.T) {
	mockUseCase := new(mocks.MockProjectUseCase)
	handler := newTestProjectHandler(mockUseCase)
	router := setupProjectTestRouter()
	router.PUT("/projects/:id", handler.Update)

	t.Run("Success", func(t *testing.T) {
		reqBody := model.UpdateProjectRequest{
			Name: "Updated Name",
		}
		resBody := &model.ProjectResponse{
			ID:   "proj-1",
			Name: "Updated Name",
		}

		mockUseCase.On("UpdateProject", mock.Anything, "proj-1", reqBody).Return(resBody, nil).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/projects/proj-1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "proj-1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error - XSS in Name", func(t *testing.T) {
		reqBody := model.UpdateProjectRequest{
			Name: "<img src=x onerror=alert(1)>",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/projects/proj-1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "proj-1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "UpdateProject")
	})

	t.Run("Validation Error - Invalid Status", func(t *testing.T) {
		reqBody := model.UpdateProjectRequest{
			Status: "invalid_status",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/projects/proj-1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "proj-1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockUseCase.AssertNotCalled(t, "UpdateProject")
	})

	t.Run("Not Found", func(t *testing.T) {
		reqBody := model.UpdateProjectRequest{
			Name: "Updated",
		}

		mockUseCase.On("UpdateProject", mock.Anything, "proj-1", reqBody).Return(nil, exception.ErrNotFound).Once()

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPut, "/projects/proj-1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "proj-1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}
