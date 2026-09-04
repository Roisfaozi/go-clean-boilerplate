package response_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	response.WriteSuccess(w, http.StatusOK, map[string]string{"foo": "bar"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":{"foo":"bar"}`)
}

func TestWriteHTTPError(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
	}{
		{exception.ErrBadRequest, http.StatusBadRequest},
		{exception.ErrUnauthorized, http.StatusUnauthorized},
		{exception.ErrForbidden, http.StatusForbidden},
		{exception.ErrNotFound, http.StatusNotFound},
		{exception.ErrConflict, http.StatusConflict},
		{exception.ErrValidationError, http.StatusUnprocessableEntity},
		{exception.ErrTooManyRequests, http.StatusTooManyRequests},
		{errors.New("other"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		response.WriteHTTPError(w, tt.err, "error msg")
		assert.Equal(t, tt.wantStatus, w.Code)
		assert.Contains(t, w.Body.String(), `"message":"error msg"`)
	}
}

func TestDecodeJSON(t *testing.T) {
	type Body struct {
		Name string `json:"name"`
	}

	// Valid
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
	var b Body
	err := response.DecodeJSON(req, &b, 1024)
	require.NoError(t, err)
	assert.Equal(t, "test", b.Name)

	// Extra tokens
	reqExtra := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}{"name":"extra"}`))
	var bExtra Body
	err = response.DecodeJSON(reqExtra, &bExtra, 1024)
	assert.Error(t, err)
}
