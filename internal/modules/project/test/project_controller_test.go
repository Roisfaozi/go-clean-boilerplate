package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	projectHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupControllerTest() (*gin.Engine, *mocks.MockProjectUseCase, *projectHttp.ProjectController) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockUseCase := new(mocks.MockProjectUseCase)
	validate := validator.New()
	controller := projectHttp.NewProjectController(mockUseCase, validate)
	return r, mockUseCase, controller
}

func TestProjectController_Create_Success(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	// Setup middleware context simulation
	r.POST("/projects", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		// Simulate TenantMiddleware setting context
		ctx := database.SetOrganizationContext(c.Request.Context(), "org-123")
		c.Request = c.Request.WithContext(ctx)
		controller.Create(c)
	})

	reqBody := model.CreateProjectRequest{
		Name:   "Test Project",
		Domain: "test.com",
	}
	body, _ := json.Marshal(reqBody)

	expectedRes := &model.ProjectResponse{
		ID:             "proj-123",
		OrganizationID: "org-123",
		UserID:         "user-123",
		Name:           "Test Project",
		Domain:         "test.com",
		Status:         "active",
	}

	mockUseCase.On("CreateProject", mock.Anything, "user-123", "org-123", reqBody).Return(expectedRes, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestProjectController_Create_ValidationFail(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.POST("/projects", func(c *gin.Context) {
		// Even if context is set, validation should fail before UseCase is called
		c.Set("user_id", "user-123")
		ctx := database.SetOrganizationContext(c.Request.Context(), "org-123")
		c.Request = c.Request.WithContext(ctx)
		controller.Create(c)
	})

	// Invalid request: Empty name
	reqBody := model.CreateProjectRequest{
		Name:   "",
		Domain: "test.com",
	}
	body, _ := json.Marshal(reqBody)

	// UseCase should NOT be called
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateProject")
}

func TestProjectController_GetProjects_Success(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.GET("/projects", func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "org-123")
		c.Request = c.Request.WithContext(ctx)
		controller.GetAll(c)
	})

	expectedRes := []*model.ProjectResponse{
		{ID: "proj-1", Name: "P1"},
		{ID: "proj-2", Name: "P2"},
	}

	mockUseCase.On("GetProjects", mock.Anything, "org-123").Return(expectedRes, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestProjectController_GetProjectByID_Success(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.GET("/projects/:id", func(c *gin.Context) {
		controller.GetByID(c)
	})

	expectedRes := &model.ProjectResponse{ID: "proj-123", Name: "Test"}

	mockUseCase.On("GetProjectByID", mock.Anything, "proj-123").Return(expectedRes, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects/proj-123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestProjectController_Update_Success(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.PUT("/projects/:id", func(c *gin.Context) {
		controller.Update(c)
	})

	reqBody := model.UpdateProjectRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(reqBody)

	expectedRes := &model.ProjectResponse{ID: "proj-123", Name: "Updated Name"}

	mockUseCase.On("UpdateProject", mock.Anything, "proj-123", reqBody).Return(expectedRes, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/projects/proj-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}


func TestProjectController_Delete_Success(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.DELETE("/projects/:id", func(c *gin.Context) {
		controller.Delete(c)
	})

	mockUseCase.On("DeleteProject", mock.Anything, "proj-123").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/proj-123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestProjectController_Create_UseCaseError(t *testing.T) {
	r, mockUseCase, controller := setupControllerTest()

	r.POST("/projects", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		ctx := database.SetOrganizationContext(c.Request.Context(), "org-123")
		c.Request = c.Request.WithContext(ctx)
		controller.Create(c)
	})

	reqBody := model.CreateProjectRequest{
		Name:   "Test Project",
		Domain: "test.com",
	}
	body, _ := json.Marshal(reqBody)

	mockUseCase.On("CreateProject", mock.Anything, "user-123", "org-123", reqBody).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
