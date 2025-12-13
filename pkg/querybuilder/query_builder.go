package querybuilder

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// GenerateDynamicQuery constructs a GORM query based on dynamic filters.
func GenerateDynamicQuery(db *gorm.DB, model interface{}, filter *DynamicFilter) (*gorm.DB, error) {
	if filter == nil {
		return db, nil
	}

	db = db.Where("deleted_at IS NULL") // Default soft delete

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
			// Ensure condition.From is a slice/array
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

// GenerateDynamicSort applies sorting conditions to a GORM query.
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

// GetDBFieldName extracts the database column name from a struct field, prioritizing 'gorm' tag, then 'json' tag, then snake_case of the field name.
func GetDBFieldName(tType reflect.Type, fieldName string) (string, bool) {
	field, found := tType.FieldByName(fieldName)
	if !found {
		return "", false
	}

	// Check 'gorm' tag
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		if colName := extractColumnNameFromGormTag(gormTag); colName != "" {
			return colName, true
		}
	}

	// Check 'json' tag
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		if colName := extractColumnNameFromJsonTag(jsonTag); colName != "" {
			return colName, true
		}
	}

	// Default to snake_case
	return ToSnakeCase(fieldName), true
}

// extractColumnNameFromGormTag parses the 'gorm' tag to find the column name.
// e.g., "column:user_name;type:varchar(100);uniqueIndex" -> "user_name"
func extractColumnNameFromGormTag(tag string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// extractColumnNameFromJsonTag parses the 'json' tag to find the column name.
// e.g., "user_name,omitempty" -> "user_name"
func extractColumnNameFromJsonTag(tag string) string {
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(input string) string {
	var result []rune
	for i, r := range input {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
