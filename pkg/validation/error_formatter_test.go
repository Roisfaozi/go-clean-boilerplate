package validation_test

import (
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestValidationStruct struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	Age      int    `json:"age" validate:"min=18"`
}

func TestFormatValidationErrors_SingleError(t *testing.T) {
	v := validator.New()
	s := TestValidationStruct{
		Name:     "",
		Email:    "test@example.com",
		Password: "password123",
		Age:      20,
	}

	err := v.Struct(s)
	assert.Error(t, err)

	formattedError := validation.FormatValidationErrors(err)
	assert.Equal(t, "Name is required", formattedError)
}

func TestFormatValidationErrors_MultipleErrors(t *testing.T) {
	v := validator.New()
	s := TestValidationStruct{
		Name:     "ab",
		Email:    "invalid-email",
		Password: "password123",
		Age:      20,
	}

	err := v.Struct(s)
	assert.Error(t, err)

	formattedError := validation.FormatValidationErrors(err)
	expectedErrors := []string{
		"Name must be at least 3 characters long",
		"Email must be a valid email address",
	}
	for _, expected := range expectedErrors {
		assert.Contains(t, formattedError, expected)
	}
	assert.Contains(t, formattedError, "; ")
}

func TestFormatValidationErrors_NonValidationError(t *testing.T) {
	someError := errors.New("this is a generic error")
	formattedError := validation.FormatValidationErrors(someError)
	assert.Equal(t, "this is a generic error", formattedError)
}

func TestFormatValidationErrors_EmailError(t *testing.T) {
	v := validator.New()
	s := TestValidationStruct{
		Name:     "Valid Name",
		Email:    "not-an-email",
		Password: "password123",
		Age:      20,
	}

	err := v.Struct(s)
	assert.Error(t, err)

	formattedError := validation.FormatValidationErrors(err)
	assert.Equal(t, "Email must be a valid email address", formattedError)
}

func TestFormatValidationErrors_MinMaxErrors(t *testing.T) {
	v := validator.New()
	s := TestValidationStruct{
		Name:     "Valid Name",
		Email:    "test@example.com",
		Password: "short",
		Age:      20,
	}
	err := v.Struct(s)
	assert.Error(t, err)
	formattedError := validation.FormatValidationErrors(err)
	assert.Equal(t, "Password must be at least 6 characters long", formattedError)

	s = TestValidationStruct{
		Name:     "Valid Name",
		Email:    "test@example.com",
		Password: "averylongpasswordthatisovertwentycharacters",
		Age:      20,
	}
	err = v.Struct(s)
	assert.Error(t, err)
	formattedError = validation.FormatValidationErrors(err)
	assert.Equal(t, "Password must be at most 20 characters long", formattedError)
}
