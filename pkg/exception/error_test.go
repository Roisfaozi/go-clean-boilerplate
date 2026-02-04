package exception

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsAreDefined(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrBadRequest", ErrBadRequest, "bad request"},
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", ErrForbidden, "forbidden"},
		{"ErrInternalServer", ErrInternalServer, "internal server error"},
		{"ErrConflict", ErrConflict, "data already exists"},
		{"ErrUnprocessableEntity", ErrUnprocessableEntity, "unprocessable entity"},
		{"ErrValidationError", ErrValidationError, "validation error"},
		{"ErrTooManyRequests", ErrTooManyRequests, "to many requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrBadRequest,
		ErrNotFound,
		ErrUnauthorized,
		ErrForbidden,
		ErrInternalServer,
		ErrConflict,
		ErrUnprocessableEntity,
		ErrValidationError,
		ErrTooManyRequests,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j {
				assert.False(t, errors.Is(err1, err2), "errors should be distinct: %v and %v", err1, err2)
			}
		}
	}
}

func TestErrorsIsComparison(t *testing.T) {
	// Test that errors.Is works correctly
	wrapped := errors.New("wrapped: " + ErrNotFound.Error())
	assert.False(t, errors.Is(wrapped, ErrNotFound), "wrapped string should not match with Is")

	// Test actual wrapping
	wrappedErr := errors.New("context: not found")
	assert.False(t, errors.Is(wrappedErr, ErrNotFound))
}

func TestErrorWrapping(t *testing.T) {
	// Test fmt.Errorf wrapping with %w
	wrapped := errors.Join(errors.New("context"), ErrNotFound)
	assert.True(t, errors.Is(wrapped, ErrNotFound))
}
