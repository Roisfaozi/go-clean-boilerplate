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

func setupProjectControllerTest() (*mocks.MockProjectUseCase, *projectHttp.ProjectController, *gin.Engine) {
	mockUseCase := new(mocks.MockProjectUseCase)
	validate := validator.New()
	controller := projectHttp.NewProjectController(mockUseCase, validate)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")

		// Set Org ID in request context as expected by database.GetOrganizationID
		ctx := database.SetOrganizationContext(c.Request.Context(), "org-1")
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	})

	return mockUseCase, controller, r
}

func TestProjectController_Create_Success(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.POST("/projects", controller.Create)

	reqBody := model.CreateProjectRequest{
		Name:   "New Project",
		Domain: "example.com",
	}
	body, _ := json.Marshal(reqBody)

	expectedResponse := &model.ProjectResponse{
		ID:             "p1",
		OrganizationID: "org-1",
		UserID:         "user-1",
		Name:           "New Project",
		Domain:         "example.com",
		Status:         "active",
	}

	mockUseCase.On("CreateProject", mock.Anything, "user-1", "org-1", reqBody).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "New Project", data["name"])
}

func TestProjectController_Create_ValidationFailure(t *testing.T) {
	_, controller, r := setupProjectControllerTest()
	r.POST("/projects", controller.Create)

	// Empty Name and Domain should fail validation
	reqBody := model.CreateProjectRequest{
		Name:   "",
		Domain: "",
	}
	body, _ := json.Marshal(reqBody)

	// Validation should fail BEFORE calling UseCase
	// So we don't expect any calls to UseCase

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// This is where we expect the BUG to manifest.
	// If validation is missing, it will proceed to call UseCase (which isn't mocked for empty input) or fail differently.
	// If validation works, it should return 422 (Unprocessable Entity).
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestProjectController_GetAll_Success(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.GET("/projects", controller.GetAll)

	expectedProjects := []*model.ProjectResponse{
		{ID: "p1", Name: "Project 1"},
		{ID: "p2", Name: "Project 2"},
	}

	mockUseCase.On("GetProjects", mock.Anything, "org-1").Return(expectedProjects, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectController_GetByID_Success(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.GET("/projects/:id", controller.GetByID)

	expectedProject := &model.ProjectResponse{ID: "p1", Name: "Project 1"}

	mockUseCase.On("GetProjectByID", mock.Anything, "p1").Return(expectedProject, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/projects/p1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectController_Update_Success(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.PUT("/projects/:id", controller.Update)

	reqBody := model.UpdateProjectRequest{
		Name: "Updated Project",
	}
	body, _ := json.Marshal(reqBody)

	expectedResponse := &model.ProjectResponse{ID: "p1", Name: "Updated Project"}

	mockUseCase.On("UpdateProject", mock.Anything, "p1", reqBody).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/projects/p1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectController_Delete_Success(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.DELETE("/projects/:id", controller.Delete)

	mockUseCase.On("DeleteProject", mock.Anything, "p1").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/p1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectController_Delete_Error(t *testing.T) {
	mockUseCase, controller, r := setupProjectControllerTest()
	r.DELETE("/projects/:id", controller.Delete)

	mockUseCase.On("DeleteProject", mock.Anything, "p1").Return(errors.New("failed"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/projects/p1", nil)
	r.ServeHTTP(w, req)

	// Since we use response.HandleError, it probably returns 500 or 400 depending on error type.
	// Assuming 500 for generic error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
