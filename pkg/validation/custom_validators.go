package validation

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
)

// RegisterCustomValidations registers custom validation tags to the provided validator instance.
func RegisterCustomValidations(v *validator.Validate) error {
	// Register 'xss' validation
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
		temp = regexp.MustCompile(fmt.Sprintf("<[/]?%s[^>]*>", tag)).ReplaceAllString(temp, "")
	}

	return !htmlTagRegex.MatchString(temp)
}

// SanitizeString removes HTML tags from a string.
func SanitizeString(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}
