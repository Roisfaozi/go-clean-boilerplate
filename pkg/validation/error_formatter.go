package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) string {
	var sb strings.Builder

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for i, e := range validationErrors {
			if i > 0 {
				sb.WriteString("; ")
			}

			field := e.Field()

			switch e.Tag() {
			case "required":
				_, _ = fmt.Fprintf(&sb, "%s is required", field)
			case "email":
				_, _ = fmt.Fprintf(&sb, "%s must be a valid email address", field)
			case "min":
				_, _ = fmt.Fprintf(&sb, "%s must be at least %s characters long", field, e.Param())
			case "max":
				_, _ = fmt.Fprintf(&sb, "%s must be at most %s characters long", field, e.Param())
			case "alphanum":
				_, _ = fmt.Fprintf(&sb, "%s must contain only alphanumeric characters", field)
			case "uuid":
				_, _ = fmt.Fprintf(&sb, "%s must be a valid UUID", field)
			case "boolean":
				_, _ = fmt.Fprintf(&sb, "%s must be a boolean value", field)
			default:
				_, _ = fmt.Fprintf(&sb, "%s failed on '%s' validation", field, e.Tag())
			}
		}
		return sb.String()
	}

	return err.Error()
}
