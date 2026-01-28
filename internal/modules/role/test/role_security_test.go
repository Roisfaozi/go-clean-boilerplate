package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleHandler_Create_Sanitization(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	handler := roleHttp.NewRoleController(mockUseCase, logrus.New(), v)
	router.POST("/roles", handler.Create)

	// "<b>admin</b>" is allowed by xss validator (safe tag),
	// but should be stripped by Sanitize() if it was called.
	createRequest := model.CreateRoleRequest{Name: "<b>admin</b>", Description: "<b>Admin</b> role"}
	requestBody, _ := json.Marshal(createRequest)

	// Expectation: The controller should call Create with SANITIZED inputs (tags stripped)
	// If Sanitize() is missing in controller, this expectation will fail because it will receive "<b>admin</b>"
	expectedRequest := &model.CreateRoleRequest{Name: "admin", Description: "Admin role"}

	mockUseCase.On("Create", mock.Anything, expectedRequest).Return(&model.RoleResponse{ID: "uuid", Name: "admin"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	// This will fail if the controller passes unsanitized data
	mockUseCase.AssertExpectations(t)
}
