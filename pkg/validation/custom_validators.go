package validation

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[a-zA-Z/][^>]*>`)
)

func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("xss", validateXSS); err != nil {
		return err
	}
	if err := v.RegisterValidation("slug", validateSlug); err != nil {
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

		re := regexp.MustCompile(fmt.Sprintf(`(?i)<[/]?%s\b[^>]*>`, tag))
		temp = re.ReplaceAllString(temp, "")
	}

	return !htmlTagRegex.MatchString(temp)
}

func validateSlug(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	// Slug must be lowercase alphanumeric with optional dashes, not starting/ending with dash
	match, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, fl.Field().String())
	return match
}

func SanitizeString(s string) string {
	p := bluemonday.StrictPolicy()
	return p.Sanitize(s)
}
