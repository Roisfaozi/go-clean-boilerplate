package querybuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// GenerateDynamicQuery builds a secure WHERE clause and args from the filter
func GenerateDynamicQuery[T any](filter *DynamicFilter) (string, []interface{}, []string, error) {
	if filter == nil || len(filter.Filter) == 0 {

		baseQuery := getSoftDeleteClause[T]()
		if baseQuery != "" {
			return baseQuery, []interface{}{}, nil, nil
		}
		return "", []interface{}{}, nil, nil
	}

	var queryParts []string
	var args []interface{}
	var warnings []string

	var zero T
	tType := reflect.TypeOf(zero)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	sdClause := getSoftDeleteClause[T]()
	if sdClause != "" {
		queryParts = append(queryParts, sdClause)
	}

	for key, f := range filter.Filter {
		field, found := findField[T](key)
		if !found {
			warnings = append(warnings, fmt.Sprintf("Field '%s' not found in model", key))
			continue
		}

		dbCol := GetDBFieldName(field)
		op := f.Type

		switch op {
		case "contains":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, "%"+val+"%")
			}
		case "notContains", "not_contains":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s NOT LIKE ?", dbCol))
				args = append(args, "%"+val+"%")
			}
		case "startsWith", "starts_with":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, val+"%")
			}
		case "endsWith", "ends_with":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, "%"+val)
			}
		case "equals":
			queryParts = append(queryParts, fmt.Sprintf("%s = ?", dbCol))
			args = append(args, f.From)
		case "notEqual", "not_equal":
			queryParts = append(queryParts, fmt.Sprintf("%s <> ?", dbCol))
			args = append(args, f.From)
		case "lessThan", "less_than":
			queryParts = append(queryParts, fmt.Sprintf("%s < ?", dbCol))
			args = append(args, f.From)
		case "lessThanOrEqual", "less_than_or_equal":
			queryParts = append(queryParts, fmt.Sprintf("%s <= ?", dbCol))
			args = append(args, f.From)
		case "greaterThan", "greater_than":
			queryParts = append(queryParts, fmt.Sprintf("%s > ?", dbCol))
			args = append(args, f.From)
		case "greaterThanOrEqual", "greater_than_or_equal":
			queryParts = append(queryParts, fmt.Sprintf("%s >= ?", dbCol))
			args = append(args, f.From)
		case "inRange", "in_range":
			queryParts = append(queryParts, fmt.Sprintf("%s >= ? AND %s <= ?", dbCol, dbCol))
			args = append(args, f.From, f.To)
		case "in":
			queryParts = append(queryParts, fmt.Sprintf("%s IN (?)", dbCol))
			args = append(args, f.From)
		case "notIn", "not_in":
			queryParts = append(queryParts, fmt.Sprintf("%s NOT IN (?)", dbCol))
			args = append(args, f.From)
		case "isNull", "is_null":
			queryParts = append(queryParts, fmt.Sprintf("%s IS NULL", dbCol))
		case "notNull", "not_null":
			queryParts = append(queryParts, fmt.Sprintf("%s IS NOT NULL", dbCol))
		default:
			warnings = append(warnings, fmt.Sprintf("Operator '%s' not supported for field '%s'", op, key))
		}
	}

	return strings.Join(queryParts, " AND "), args, warnings, nil
}

// GenerateDynamicSort generates the ORDER BY clause
func GenerateDynamicSort[T any](filter *DynamicFilter) (string, error) {
	if filter == nil || filter.Sort == nil || len(*filter.Sort) == 0 {
		return "", nil
	}

	var sortParts []string

	var zero T
	tType := reflect.TypeOf(zero)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	for _, s := range *filter.Sort {
		field, found := findField[T](s.ColId)
		if !found {
			return "", fmt.Errorf("sort field '%s' not found", s.ColId)
		}

		dbCol := GetDBFieldName(field)
		direction := "ASC"
		if strings.ToLower(s.Sort) == "desc" {
			direction = "DESC"
		}
		sortParts = append(sortParts, fmt.Sprintf("%s %s", dbCol, direction))
	}

	return strings.Join(sortParts, ", "), nil
}

func Preload(db *gorm.DB, preloads []PreloadEntity) *gorm.DB {
	for _, p := range preloads {
		db = db.Preload(p.Entity, p.Args...)
	}
	return db
}

// GetDBFieldName resolves the database column name
func GetDBFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("gorm")
	if tag != "" {
		parts := strings.Split(tag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
	}
	return ToSnakeCase(field.Name)
}

// ToSnakeCase converts CamelCase to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (i+1 < len(runes) && unicode.IsLower(runes[i+1]) || unicode.IsLower(runes[i-1])) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Internal helper to check for DeletedBy or DeletedAt and return default clause
func getSoftDeleteClause[T any]() string {
	var zero T
	tType := reflect.TypeOf(zero)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	_, hasDeletedBy := tType.FieldByName("DeletedBy")
	if hasDeletedBy {
		f, _ := tType.FieldByName("DeletedBy")
		col := GetDBFieldName(f)
		return fmt.Sprintf("%s IS NULL", col)
	}

	_, hasDeletedAt := tType.FieldByName("DeletedAt")
	if hasDeletedAt {
		f, _ := tType.FieldByName("DeletedAt")
		col := GetDBFieldName(f)

		if f.Type.Kind() == reflect.Int64 || f.Type.Kind() == reflect.Int {
			// For int timestamps, usually 0 is active
			return fmt.Sprintf("%s = 0", col)
		}
		return fmt.Sprintf("%s IS NULL", col)
	}

	return ""
}

// parseToNumberOrNull attempts to parse a value to a number, helpful for numeric filters on string inputs
func parseToNumberOrNull(v interface{}) interface{} {
	s, ok := v.(string)
	if !ok {
		return v
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return v
}

// findField attempts to find a struct field by JSON tag, then snake_case name, then case-insensitive name
func findField[T any](key string) (reflect.StructField, bool) {
	var zero T
	tType := reflect.TypeOf(zero)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	// 1. Check strict match with JSON tag
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		jsonTag := field.Tag.Get("json")
		// JSON tag can be "name,omitempty", so we split
		tagName := strings.Split(jsonTag, ",")[0]
		if tagName == key {
			return field, true
		}
	}

	// 2. Check if key matches snake_case version of field name
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		snakeName := ToSnakeCase(field.Name)
		if snakeName == key {
			return field, true
		}
	}

	// 3. Fallback: Case-insensitive match on field Name or JSON tag
	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		if strings.EqualFold(field.Name, key) {
			return field, true
		}
		
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if strings.EqualFold(jsonTag, key) {
			return field, true
		}
	}

	return reflect.StructField{}, false
}