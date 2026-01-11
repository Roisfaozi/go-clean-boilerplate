package pkg_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/stretchr/testify/assert"
)

func TestContainsSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Safe string", "hello world", false},
		{"SQL comment --", "admin' --", true},
		{"SQL comment /*", "admin /* comment */", true},
		{"SQL Union", "UNION SELECT * FROM users", true},
		{"SQL Select", "SELECT * FROM users", true},
		{"SQL Drop", "DROP TABLE users", true},
		{"Mixed case", "SeLeCt * FrOm", true},
		{"Quote", "O'Reilly", true},
		{"Semicolon", "param; DROP TABLE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pkg.ContainsSQLInjection(tt.input))
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "hello", "hello"},
		{"Trim space", "  hello  ", "hello"},
		{"HTML tag", "<script>", "&lt;script&gt;"},
		{"Quotes", "\"", "&#34;"},
		{"Ampersand", "&", "&amp;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pkg.SanitizeString(tt.input))
		})
	}
}
