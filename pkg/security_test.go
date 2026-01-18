package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Safe string", "hello world", false},
		{"SQL injection - OR 1=1", "admin' OR 1=1 --", true},
		{"SQL injection - DROP TABLE", "DROP TABLE users", true},
		{"SQL injection - UNION SELECT", "UNION SELECT * FROM users", true},
		{"SQL injection - comment", "admin --", true},
		{"SQL injection - semicolon", "admin;", true},
		{"Safe email", "test@example.com", false},
		{"Safe username", "user123", false},

		// False positive prevention (Substrings)
		{"Safe word 'selection'", "Natural selection", false}, // Contains 'select' but not as whole word
		{"Safe name 'Selecta'", "Selecta", false}, // Contains 'select' but not as whole word
		{"Safe name 'Benedict'", "Benedict", false},

		// Limitations (Whole words are still flagged)
		// "Grant" is a name, but also a keyword. This simple regex cannot distinguish.
		{"Ambiguous name 'Grant'", "Grant", true},
		{"Sentence with keyword", "Please select an option", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ContainsSQLInjection(tt.input), "Input: %s", tt.input)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "hello world", "hello world"},
		{"String with spaces", "  hello world  ", "hello world"},
		{"HTML tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Special chars", "foo & bar", "foo &amp; bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SanitizeString(tt.input))
		})
	}
}
