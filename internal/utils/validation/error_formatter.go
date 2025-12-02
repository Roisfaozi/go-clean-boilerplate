package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors converts validator errors into a friendly string message.
// It combines multiple errors into a single string separated by semicolons.
func FormatValidationErrors(err error) string {
	var sb strings.Builder

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for i, e := range validationErrors {
			if i > 0 {
				sb.WriteString("; ")
			}
			
			// Use Field() which returns the struct field name. 
			// Note: To use JSON tag names instead, we'd need to register a custom TagNameFunc with the validator instance,
			// but using the Struct Field Name is usually sufficient and standard.
			field := e.Field()

			switch e.Tag() {
			case "required":
				sb.WriteString(fmt.Sprintf("%s is required", field))
			case "email":
				sb.WriteString(fmt.Sprintf("%s must be a valid email address", field))
			case "min":
				sb.WriteString(fmt.Sprintf("%s must be at least %s characters long", field, e.Param()))
			case "max":
				sb.WriteString(fmt.Sprintf("%s must be at most %s characters long", field, e.Param()))
			case "alphanum":
				sb.WriteString(fmt.Sprintf("%s must contain only alphanumeric characters", field))
			case "uuid":
				sb.WriteString(fmt.Sprintf("%s must be a valid UUID", field))
			case "boolean":
				sb.WriteString(fmt.Sprintf("%s must be a boolean value", field))
			default:
				sb.WriteString(fmt.Sprintf("%s failed on '%s' validation", field, e.Tag()))
			}
		}
		return sb.String()
	}

	// If it's not a validation error, return the default error string
	return err.Error()
}
