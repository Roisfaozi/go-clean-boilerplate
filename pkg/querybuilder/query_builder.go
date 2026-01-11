package querybuilder

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

func GenerateDynamicQuery(db *gorm.DB, model interface{}, filter *DynamicFilter) (*gorm.DB, error) {
	if filter == nil {
		return db, nil
	}

	tType := reflect.TypeOf(model)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	for fieldName, condition := range filter.Filter {
		dbFieldName, ok := GetDBFieldName(tType, fieldName)
		if !ok {
			return nil, fmt.Errorf("invalid field for filtering: %s", fieldName)
		}

		switch condition.Type {
		case "equals":
			db = db.Where(fmt.Sprintf("%s = ?", dbFieldName), condition.From)
		case "contains":
			db = db.Where(fmt.Sprintf("%s LIKE ?", dbFieldName), fmt.Sprintf("%%%v%%", condition.From))
		case "in":
			val := reflect.ValueOf(condition.From)
			if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
				db = db.Where(fmt.Sprintf("%s IN (?)", dbFieldName), condition.From)
			} else {
				return nil, fmt.Errorf("invalid value for 'in' filter, must be a slice or array")
			}
		case "between":
			db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", dbFieldName), condition.From, condition.To)
		case "gt":
			db = db.Where(fmt.Sprintf("%s > ?", dbFieldName), condition.From)
		case "gte":
			db = db.Where(fmt.Sprintf("%s >= ?", dbFieldName), condition.From)
		case "lt":
			db = db.Where(fmt.Sprintf("%s < ?", dbFieldName), condition.From)
		case "lte":
			db = db.Where(fmt.Sprintf("%s <= ?", dbFieldName), condition.From)
		case "ne":
			db = db.Where(fmt.Sprintf("%s != ?", dbFieldName), condition.From)
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", condition.Type)
		}
	}

	return db, nil
}

func GenerateDynamicSort(db *gorm.DB, model interface{}, filter *DynamicFilter) (*gorm.DB, error) {
	if filter == nil || filter.Sort == nil || len(*filter.Sort) == 0 {
		return db, nil
	}

	tType := reflect.TypeOf(model)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	for _, sort := range *filter.Sort {
		dbFieldName, ok := GetDBFieldName(tType, sort.ColId)
		if !ok {
			return nil, fmt.Errorf("invalid field for sorting: %s", sort.ColId)
		}

		order := "asc"
		if strings.ToLower(sort.Sort) == "desc" {
			order = "desc"
		}
		db = db.Order(fmt.Sprintf("%s %s", dbFieldName, order))
	}

	return db, nil
}

// GetDBFieldName extracts the database column name from the gorm tag or converts the field name to snake_case.
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

			// Check GORM tag
			gormTag := f.Tag.Get("gorm")
			if gormTag != "" {
				col := extractColumnNameFromGormTag(gormTag)
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
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		if colName := extractColumnNameFromGormTag(gormTag); colName != "" {
			return colName, true
		}
	}

	return ToSnakeCase(field.Name), true
}

func extractColumnNameFromGormTag(tag string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
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
