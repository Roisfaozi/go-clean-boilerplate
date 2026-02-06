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

func TestRoleHandler_Create_XSS_Attribute_Sanitization(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	// Payload with "safe" tag <b> but with malicious attribute.
	// Current xss validator might allow <b>, and if Sanitize is not called, it passes raw.
	// SanitizeString strips ALL tags, so we expect "Bold".
	createRequest := model.CreateRoleRequest{
		Name:        "<b onmouseover=alert(1)>Bold</b>",
		Description: "XSS Attempt",
	}
	requestBody, _ := json.Marshal(createRequest)

	// Expect the UseCase to receive the SANITIZED name "Bold"
	expectedReq := &model.CreateRoleRequest{
		Name:        "Bold",
		Description: "XSS Attempt",
	}

	// We match the struct. Note: CreateRoleRequest might have other fields or methods,
	// but testify checks field equality.
	mockUseCase.On("Create", mock.Anything, expectedReq).Return(&model.RoleResponse{ID: "uuid", Name: "Bold"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// If sanitization happens, it should return 201 Created.
	// If it fails (raw string passed), the mock won't match, and likely return 500 or panic depending on mock setup,
	// or simply fail the assertion later.
	// Actually, if mock call doesn't match, mock.Called returns panic if not configured otherwise, or returns default values.
	// But testify mocks usually panic on unexpected call or return nothing.

	// However, if the controller passes raw string, the mock expectation for "Bold" won't match "<b...>Bold</b>".
	// So the mock will say "Unexpected call to Create with arguments ...".

	assert.Equal(t, http.StatusCreated, w.Code)

	mockUseCase.AssertExpectations(t)
}
