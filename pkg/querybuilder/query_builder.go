package querybuilder

import (
	"reflect"
	"regexp"
	"strings"
)

// GetDBFieldName extracts the database column name from the db tag or converts the field name to snake_case.
func GetDBFieldName(tType reflect.Type, fieldName string) (string, bool) {
	if isSensitiveField(fieldName) {
		return "", false
	}

	// 1. Try direct match
	field, found := tType.FieldByName(fieldName)

	// 2. If not found, try case-insensitive or tag-based match
	if !found {
		for i := 0; i < tType.NumField(); i++ {
			f := tType.Field(i)

			// Check exact name (case-insensitive)
			if strings.EqualFold(f.Name, fieldName) {
				field = f
				found = true
				break
			}

			// Check DB tag
			dbTag := f.Tag.Get("db")
			if dbTag != "" {
				col := extractColumnNameFromDBTag(dbTag)
				if col != "" && strings.EqualFold(col, fieldName) {
					field = f
					found = true
					break
				}
			}

			// Check JSON tag
			jsonTag := f.Tag.Get("json")
			if jsonTag != "" {
				col := extractColumnNameFromJsonTag(jsonTag)
				if strings.EqualFold(col, fieldName) {
					field = f
					found = true
					break
				}
			}

			// Check Snake Case of field name
			if strings.EqualFold(ToSnakeCase(f.Name), fieldName) {
				field = f
				found = true
				break
			}
		}
	}

	if !found || isSensitiveField(field.Name) {
		return "", false
	}

	// Now that we have the field, get its DB column name
	dbTag := field.Tag.Get("db")
	if dbTag != "" {
		if colName := extractColumnNameFromDBTag(dbTag); colName != "" {
			return colName, true
		}
	}

	return ToSnakeCase(field.Name), true
}

func extractColumnNameFromDBTag(tag string) string {
	parts := strings.Split(tag, ",")
	if len(parts) > 0 && parts[0] != "-" {
		return parts[0]
	}
	return ""
}

func extractColumnNameFromJsonTag(tag string) string {
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func isSensitiveField(fieldName string) bool {
	sensitiveFields := map[string]bool{
		"Password": true,
		"Token":    true,
		"Secret":   true,
		"Key":      true,
		"Salt":     true,
	}
	return sensitiveFields[fieldName]
}
