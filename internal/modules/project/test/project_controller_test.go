package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	httpDelivery "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectUseCase ...
type MockProjectUseCase struct {
	mock.Mock
}

func (m *MockProjectUseCase) CreateProject(ctx context.Context, userID string, orgID string, req model.CreateProjectRequest) (*model.ProjectResponse, error) {
	args := m.Called(ctx, userID, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) GetProjects(ctx context.Context, orgID string) ([]*model.ProjectResponse, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) GetProjectByID(ctx context.Context, id string) (*model.ProjectResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) UpdateProject(ctx context.Context, id string, req model.UpdateProjectRequest) (*model.ProjectResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) DeleteProject(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestProjectController_Create_Validation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUC := new(MockProjectUseCase)
	validate := config.NewValidator()
	controller := httpDelivery.NewProjectController(mockUC, validate)

	r := gin.New()
	r.POST("/projects", controller.Create)

	// If validation is missing, the controller will proceed to call CreateProject on usecase.
	// We mock it to succeed to simulate the vulnerability (invalid data getting processed).
	mockUC.On("CreateProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&model.ProjectResponse{ID: "123"}, nil)

	// Test Case: Empty Name (Should Fail Validation)
	reqBody := model.CreateProjectRequest{
		Name:   "", // Invalid: required, min=1
		Domain: "valid.com",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	// Assertions
	// If vulnerability exists: w.Code will be 201 (Created)
	// If fixed: w.Code will be 422 (Unprocessable Entity) or 400 (Bad Request)

	if w.Code == http.StatusCreated {
		t.Log("VULNERABILITY CONFIRMED: Created project with empty name")
		t.Fail() // Fail the test to indicate vulnerability exists
	} else {
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code, "Expected validation error")
	}
}

func TestProjectController_Update_Validation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUC := new(MockProjectUseCase)
	validate := config.NewValidator()
	controller := httpDelivery.NewProjectController(mockUC, validate)

	r := gin.New()
	r.PUT("/projects/:id", controller.Update)

	// Mock successful update if validation is bypassed
	mockUC.On("UpdateProject", mock.Anything, "123", mock.Anything).Return(&model.ProjectResponse{ID: "123"}, nil)

// Test Case: XSS Payload (Should Fail Validation)

	// So let's test XSS vulnerability.
	reqBody := model.UpdateProjectRequest{
		Name: "<script>alert(1)</script>",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/projects/123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	// If vulnerability exists (no XSS check): w.Code will be 200
	// If fixed (XSS check added & validation enabled): w.Code will be 422/400

	if w.Code == http.StatusOK {
t.Errorf("Vulnerability confirmed: project updated with XSS payload")
	} else {
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code, "Expected validation error")
	}
}
