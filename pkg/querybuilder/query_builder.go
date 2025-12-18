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

	if hasSoftDeleteField(tType) {
		db = db.Where("deleted_at IS NULL")
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

func hasSoftDeleteField(tType reflect.Type) bool {
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	_, found := tType.FieldByName("DeletedAt")
	return found
}

func GetDBFieldName(tType reflect.Type, fieldName string) (string, bool) {
	field, found := tType.FieldByName(fieldName)
	if !found {
		for i := 0; i < tType.NumField(); i++ {
			f := tType.Field(i)
			if strings.EqualFold(f.Name, fieldName) {
				field = f
				found = true
				break
			}
		}
		if !found {
			return "", false
		}
	}

	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		if colName := extractColumnNameFromGormTag(gormTag); colName != "" {
			return colName, true
		}
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag != "" {
		if colName := extractColumnNameFromJsonTag(jsonTag); colName != "" {
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
