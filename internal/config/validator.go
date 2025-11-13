package config

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NewValidator creates and configures a new instance of the validator.
// It registers a custom tag name function to use the 'json' tag in error messages,
// which makes validation errors more API-friendly.
func NewValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}
