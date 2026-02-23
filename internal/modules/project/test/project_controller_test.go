package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	projectHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupControllerTest() (*projectHttp.ProjectController, *mocks.MockProjectUseCase) {
	mockUseCase := new(mocks.MockProjectUseCase)
	validate := config.NewValidator()
	controller := projectHttp.NewProjectController(mockUseCase, validate)
	return controller, mockUseCase
}

func TestProjectController_Create_ValidationFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUseCase := setupControllerTest()

	// Setup router
	r := gin.New()
	// Add user_id and org_id for context, simulating middleware
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")

		// Set organization_id in request context (for database.GetOrganizationID)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, database.OrganizationIDKey, "org-1")
		c.Request = c.Request.WithContext(ctx)
	})
	r.POST("/projects", controller.Create)

	// Create request with invalid data (XSS payload)
	reqBody := model.CreateProjectRequest{
		Name:   "<script>alert(1)</script>", // Invalid: XSS
		Domain: "valid-domain.com",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// If validation is missing (bug), CreateProject WILL be called with "org-1"
	// But since we fixed it, it should NOT be called. We use Maybe() to allow both behaviors during development/verification.
	mockUseCase.On("CreateProject", mock.Anything, "user-1", "org-1", mock.AnythingOfType("model.CreateProjectRequest")).Return(&model.ProjectResponse{}, nil).Maybe()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// If validation works, status should be 400.
	// If validation is missing (current state), it will likely be 201.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectController_Update_ValidationFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUseCase := setupControllerTest()

	r := gin.New()
	r.PUT("/projects/:id", controller.Update)

	// Request with invalid data (XSS payload)
	reqBody := model.UpdateProjectRequest{
		Name: "<script>alert(1)</script>", // Invalid: XSS
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/projects/p1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// If bug exists, UpdateProject will be called.
	mockUseCase.On("UpdateProject", mock.Anything, "p1", mock.AnythingOfType("model.UpdateProjectRequest")).Return(&model.ProjectResponse{}, nil).Maybe()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
