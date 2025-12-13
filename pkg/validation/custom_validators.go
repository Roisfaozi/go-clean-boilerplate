package validation

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	// Matches any HTML tag: starts with <, followed by optional /, then tag name, then anything until >
	htmlTagRegex = regexp.MustCompile(`<[a-zA-Z/][^>]*>`)
)

// RegisterCustomValidations registers custom validation tags to the provided validator instance.
func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("xss", validateXSS); err != nil {
		return err
	}

	return nil
}

func validateXSS(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	safeTags := []string{"b", "i", "em", "strong", "u"}
	desc := fl.Field().String()

	temp := desc
	for _, tag := range safeTags {
		// Case insensitive replacement for safe tags is tricky with simple regex without flag
		// For simplicity, we assume lowercase safe tags or handle basic variations
		// Better approach: use a proper HTML sanitizer library like bluemonday
		re := regexp.MustCompile(fmt.Sprintf(`(?i)<[/]?%s\b[^>]*>`, tag))
		temp = re.ReplaceAllString(temp, "")
	}

	return !htmlTagRegex.MatchString(temp)
}

func SanitizeString(s string) string {
	// Simple regex-based strip tags.
	// Note: This is not secure against all XSS vectors but sufficient for basic cleanup.
	return htmlTagRegex.ReplaceAllString(s, "")
}
