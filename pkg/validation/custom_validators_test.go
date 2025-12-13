package validation_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Mock structure to hold a string field with the 'xss' tag
type TestStruct struct {
	Comment string `validate:"xss"`
}

func TestValidateXSS(t *testing.T) {
	v := validator.New()
	err := validation.RegisterCustomValidations(v) // Register the custom 'xss' validation
	assert.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool // true if validation passes (no XSS), false if it fails (XSS detected)
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
			expected: true, // Incomplete tag is not matched by regex <...>, so it passes
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

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String with script tags",
			input:    "Hello <script>alert(1)</script> World",
			expected: "Hello alert(1) World", // Content remains, tags removed
		},
		{
			name:     "String with image tags",
			input:    "Hello <img src=\"x\"> World",
			expected: "Hello  World",
		},
		{
			name:     "String with mixed HTML tags",
			input:    "<b>Bold</b> and <i>Italic</i> <a href=\"#\">Link</a> <script>alert(1)</script>.",
			expected: "Bold and Italic Link alert(1).", // Content remains
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Complex XSS payload",
			input:    "<svg/onload=alert(document.cookie)>Hello",
			expected: "Hello",
		},
		{
			name:     "Allowed tags are handled by SanitizeString - current implementation removes all",
			input:    "This is <b>bold</b>",
			expected: "This is bold", // Current SanitizeString implementation removes all tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
