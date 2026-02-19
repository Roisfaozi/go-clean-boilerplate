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

type ExtendedValidationStruct struct {
	Username string `json:"username" validate:"alphanum"`
	ID       string `json:"id" validate:"uuid"`
	IsActive string `json:"is_active" validate:"boolean"`
	Slug     string `json:"slug" validate:"slug"`
}

type InvalidTypeStruct struct {
	Number int `validate:"xss"`
	Flag   bool `validate:"slug"`
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

func TestFormatValidationErrors_OtherTypes(t *testing.T) {
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	tests := []struct {
		name     string
		input    ExtendedValidationStruct
		expected string
	}{
		{
			name: "Alphanum Error",
			input: ExtendedValidationStruct{
				Username: "invalid space",
				ID:       "00000000-0000-0000-0000-000000000000", // Valid UUID
				IsActive: "true",
				Slug:     "valid-slug",
			},
			expected: "Username must contain only alphanumeric characters",
		},
		{
			name: "UUID Error",
			input: ExtendedValidationStruct{
				Username: "validUser",
				ID:       "not-a-uuid",
				IsActive: "true",
				Slug:     "valid-slug",
			},
			expected: "ID must be a valid UUID",
		},
		{
			name: "Boolean Error",
			input: ExtendedValidationStruct{
				Username: "validUser",
				ID:       "00000000-0000-0000-0000-000000000000",
				IsActive: "not-bool",
				Slug:     "valid-slug",
			},
			expected: "IsActive must be a boolean value",
		},
		{
			name: "Default Error (Slug)",
			input: ExtendedValidationStruct{
				Username: "validUser",
				ID:       "00000000-0000-0000-0000-000000000000",
				IsActive: "true",
				Slug:     "Invalid Slug",
			},
			expected: "Slug failed on 'slug' validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)
			assert.Error(t, err)
			formattedError := validation.FormatValidationErrors(err)
			assert.Equal(t, tt.expected, formattedError)
		})
	}
}

func TestValidation_InvalidType(t *testing.T) {
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	s := InvalidTypeStruct{
		Number: 123,
		Flag:   true,
	}

	// Should fail validation because type is not string, if validator returns false for non-string
	err := v.Struct(s)
	assert.Error(t, err)
}
