package config

import "testing"

func TestIsStrictCasbinEnv(t *testing.T) {
	tests := []struct {
		name     string
		appEnv   string
		expected bool
	}{
		{name: "empty is non strict", appEnv: "", expected: false},
		{name: "local is non strict", appEnv: "local", expected: false},
		{name: "dev is non strict", appEnv: "dev", expected: false},
		{name: "development is non strict", appEnv: "development", expected: false},
		{name: "test is non strict", appEnv: "test", expected: false},
		{name: "testing is non strict", appEnv: "testing", expected: false},
		{name: "production is strict", appEnv: "production", expected: true},
		{name: "staging is strict", appEnv: "staging", expected: true},
		{name: "case and spaces normalize", appEnv: " Production ", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStrictCasbinEnv(tt.appEnv); got != tt.expected {
				t.Fatalf("isStrictCasbinEnv(%q) = %v, want %v", tt.appEnv, got, tt.expected)
			}
		})
	}
}
