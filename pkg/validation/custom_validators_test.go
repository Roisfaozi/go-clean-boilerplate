package validation_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Comment string `validate:"xss"`
}

type TestSlugStruct struct {
	Slug string `validate:"slug"`
}

func TestValidateXSS(t *testing.T) {
	v := validator.New()
	err := validation.RegisterCustomValidations(v)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Clean string",
			input:    "This is a clean comment.",
			expected: true,
		},
		{
			name:     "String with allowed tags",
			input:    "This is a <b>bold</b> and <i>italic</i> comment.",
			expected: true,
		},
		{
			name:     "Simple XSS script tag",
			input:    "This contains <script>alert('xss')</script> malicious code.",
			expected: false,
		},
		{
			name:     "XSS with img onerror",
			input:    "Comment with <img src=x onerror=alert('xss')> payload.",
			expected: false,
		},
		{
			name:     "XSS with svg onload",
			input:    "<svg onload=alert(1)>",
			expected: false,
		},
		{
			name:     "Mixed case script tag",
			input:    "Comment with <sCrIpT>alert(1)</sCrIpT> tag.",
			expected: false,
		},
		{
			name:     "XSS with no closing tag",
			input:    "<script>alert(1)",
			expected: false,
		},
		{
			name:     "XSS with single quote",
			input:    "<img src='x' onerror='alert(1)'>",
			expected: false,
		},
		{
			name:     "XSS with double quote",
			input:    "<img src=\"x\" onerror=\"alert(1)\">",
			expected: false,
		},
		{
			name:     "XSS with encoded characters",
			input:    "<%2Fscript><script>alert(1)</script>",
			expected: false,
		},
		{
			name:     "String with only allowed tags",
			input:    "<b><i>Hello</i></b> World",
			expected: true,
		},
		{
			name:     "String with disallowed but incomplete tag",
			input:    "This is <script",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Comment: tt.input}
			err := v.Struct(s)
			if tt.expected {
				assert.NoError(t, err, "Expected no validation error for input: %s", tt.input)
			} else {
				if assert.Error(t, err, "Expected validation error for input: %s", tt.input) {
					assert.Contains(t, err.Error(), "xss", "Expected XSS validation error tag")
				}
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	v := validator.New()
	err := validation.RegisterCustomValidations(v)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid Slug",
			input:    "my-slug-123",
			expected: true,
		},
		{
			name:     "Valid Slug - Single Word",
			input:    "slug",
			expected: true,
		},
		{
			name:     "Invalid Slug - Uppercase",
			input:    "My-Slug",
			expected: false,
		},
		{
			name:     "Invalid Slug - Starts with Dash",
			input:    "-slug",
			expected: false,
		},
		{
			name:     "Invalid Slug - Ends with Dash",
			input:    "slug-",
			expected: false,
		},
		{
			name:     "Invalid Slug - Special Characters",
			input:    "slug@!",
			expected: false,
		},
		{
			name:     "Invalid Slug - Spaces",
			input:    "my slug",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestSlugStruct{Slug: tt.input}
			err := v.Struct(s)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "slug")
			}
		})
	}
}
