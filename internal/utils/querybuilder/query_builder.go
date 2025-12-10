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

		field, found := tType.FieldByName(key)
		if !found {

			var fieldFound bool
			for i := 0; i < tType.NumField(); i++ {
				if strings.EqualFold(tType.Field(i).Name, key) {
					field = tType.Field(i)
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				warnings = append(warnings, fmt.Sprintf("Field '%s' not found in model", key))
				continue
			}
		}

		dbCol := GetDBFieldName(field)
		op := f.Type

		switch op {
		case "contains":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, "%"+val+"%")
			}
		case "notContains":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s NOT LIKE ?", dbCol))
				args = append(args, "%"+val+"%")
			}
		case "startsWith":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, val+"%")
			}
		case "endsWith":
			if val, ok := f.From.(string); ok {
				queryParts = append(queryParts, fmt.Sprintf("%s LIKE ?", dbCol))
				args = append(args, "%"+val)
			}
		case "equals":
			queryParts = append(queryParts, fmt.Sprintf("%s = ?", dbCol))
			args = append(args, f.From)
		case "notEqual":
			queryParts = append(queryParts, fmt.Sprintf("%s <> ?", dbCol))
			args = append(args, f.From)
		case "lessThan":
			queryParts = append(queryParts, fmt.Sprintf("%s < ?", dbCol))
			args = append(args, f.From)
		case "lessThanOrEqual":
			queryParts = append(queryParts, fmt.Sprintf("%s <= ?", dbCol))
			args = append(args, f.From)
		case "greaterThan":
			queryParts = append(queryParts, fmt.Sprintf("%s > ?", dbCol))
			args = append(args, f.From)
		case "greaterThanOrEqual":
			queryParts = append(queryParts, fmt.Sprintf("%s >= ?", dbCol))
			args = append(args, f.From)
		case "inRange":
			queryParts = append(queryParts, fmt.Sprintf("%s >= ? AND %s <= ?", dbCol, dbCol))
			args = append(args, f.From, f.To)
		case "in":
			queryParts = append(queryParts, fmt.Sprintf("%s IN (?)", dbCol))
			args = append(args, f.From)
		case "notIn":
			queryParts = append(queryParts, fmt.Sprintf("%s NOT IN (?)", dbCol))
			args = append(args, f.From)
		case "isNull":
			queryParts = append(queryParts, fmt.Sprintf("%s IS NULL", dbCol))
		case "notNull":
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
		field, found := tType.FieldByName(s.ColId)
		if !found {
			for i := 0; i < tType.NumField(); i++ {
				if strings.EqualFold(tType.Field(i).Name, s.ColId) {
					field = tType.Field(i)
					found = true
					break
				}
			}
		}

		if !found {
			// Skip or error? Prefer safe skip or strict error.
			// Let's error for sort to avoid unexpected ordering.
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
