package pkg

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "Hello World", "hello-world"},
		{"With special characters", "Hello! @World#", "hello-world"},
		{"Multiple spaces and dashes", "Hello   World---Test", "hello-world-test"},
		{"Only special characters", "!@#$%^&*", "uuid"},
		{"Empty string", "", "uuid"},
		{"Already a slug", "hello-world", "hello-world"},
		{"Trailing spaces", " Hello World ", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
