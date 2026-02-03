package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRoleHandler_Create_Untrimmed_Reproduction demonstrates that currently
// the Role Controller passes untrimmed whitespace to the UseCase.
func TestRoleHandler_Create_Sanitized(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	// Input with leading/trailing whitespace
	createRequest := model.CreateRoleRequest{Name: "   admin   ", Description: "  desc  "}
	requestBody, _ := json.Marshal(createRequest)

	// Expectation: The UseCase receives the input TRIMMED and SANITIZED
	expectedRequest := model.CreateRoleRequest{Name: "admin", Description: "desc"}
	mockUseCase.On("Create", mock.Anything, &expectedRequest).Return(&model.RoleResponse{ID: "uuid", Name: "admin"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// If the test passes, it means the controller DID trim the input.
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}
