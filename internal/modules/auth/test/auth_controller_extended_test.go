package test_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Login_Extended(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	t.Run("Validation_Error_Empty_Fields", func(t *testing.T) {
		reqBody := model.LoginRequest{Username: "", Password: ""}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Binding_Error_Malformed_JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{invalid-json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_RefreshToken_Extended(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	t.Run("Missing_Cookie", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
		// No cookie set

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUseCase.AssertNotCalled(t, "RefreshToken", mock.Anything, mock.Anything)
	})
}

func TestAuthHandler_Logout_Extended(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	t.Run("Missing_User_In_Context", func(t *testing.T) {
		// Manually create context without setting user_id/session_id
		w := httptest.NewRecorder()
		// We can't easily use router here because middleware sets context, but we are testing controller logic
		// that assumes middleware *might* fail or not be present (defensive coding).
		// However, AuthMiddleware usually ensures these are set.
		// Testing the controller directly:
		ctx, _ := newGinContextWithRecorder(w)
		ctx.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)

		handler.Logout(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUseCase.AssertNotCalled(t, "RevokeToken", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("RevokeToken_UseCase_Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := newGinContextWithRecorder(w)
		ctx.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)
		ctx.Set("user_id", "user1")
		ctx.Set("session_id", "sess1")

		mockUseCase.On("RevokeToken", mock.Anything, "user1", "sess1").Return(usecase.ErrTokenRevoked)

		handler.Logout(ctx)

		// Controller maps generic errors to 500 in Logout usually, or specific if mapped
		// In implementation: response.HandleError(c, err, "Logout failed")
		// ErrTokenRevoked is usually 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
