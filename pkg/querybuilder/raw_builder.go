package querybuilder

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// SQLQueryResult contains the generated SQL fragments and arguments.
type SQLQueryResult struct {
	WhereSQL string
	OrderBy  string
	Limit    int
	Offset   int
	Args     []any
}

// BuildRawQuery translates DynamicFilter into a parameterized WHERE and ORDER BY SQL clause.
func BuildRawQuery(model any, filter *DynamicFilter) (*SQLQueryResult, error) {
	result := &SQLQueryResult{
		Limit:  10,
		Offset: 0,
	}

	if filter == nil {
		return result, nil
	}

	tType := reflect.TypeOf(model)
	if tType.Kind() == reflect.Pointer {
		tType = tType.Elem()
	}

	whereSQL, args, err := buildFilterClauses(tType, filter.Filter)
	if err != nil {
		return nil, err
	}
	result.WhereSQL = whereSQL
	result.Args = args

	orderBy, err := buildOrderClause(tType, filter.Sort)
	if err != nil {
		return nil, err
	}
	result.OrderBy = orderBy

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	result.Limit = pageSize
	result.Offset = (page - 1) * pageSize

	return result, nil
}

func buildFilterClauses(tType reflect.Type, filters map[string]Filter) (string, []any, error) {
	if len(filters) == 0 {
		return "", nil, nil
	}

	filterKeys := make([]string, 0, len(filters))
	for k := range filters {
		filterKeys = append(filterKeys, k)
	}
	sort.Strings(filterKeys)

	var whereClauses []string
	var args []any

	for _, fieldName := range filterKeys {
		cond := filters[fieldName]
		dbFieldName, ok := GetDBFieldName(tType, fieldName)
		if !ok {
			return "", nil, fmt.Errorf("invalid field for filtering: %s", fieldName)
		}

		clause, clauseArgs, err := buildSingleCondition(dbFieldName, cond)
		if err != nil {
			return "", nil, err
		}
		whereClauses = append(whereClauses, clause)
		args = append(args, clauseArgs...)
	}

	return strings.Join(whereClauses, " AND "), args, nil
}

func buildSingleCondition(dbFieldName string, cond Filter) (string, []any, error) {
	switch cond.Type {
	case "equals":
		return fmt.Sprintf("%s = ?", dbFieldName), []any{cond.From}, nil
	case "contains":
		return fmt.Sprintf("%s LIKE ?", dbFieldName), []any{fmt.Sprintf("%%%v%%", cond.From)}, nil
	case "in":
		return buildInCondition(dbFieldName, cond.From)
	case "between":
		return fmt.Sprintf("%s BETWEEN ? AND ?", dbFieldName), []any{cond.From, cond.To}, nil
	case "gt":
		return fmt.Sprintf("%s > ?", dbFieldName), []any{cond.From}, nil
	case "gte":
		return fmt.Sprintf("%s >= ?", dbFieldName), []any{cond.From}, nil
	case "lt":
		return fmt.Sprintf("%s < ?", dbFieldName), []any{cond.From}, nil
	case "lte":
		return fmt.Sprintf("%s <= ?", dbFieldName), []any{cond.From}, nil
	case "ne":
		return fmt.Sprintf("%s != ?", dbFieldName), []any{cond.From}, nil
	default:
		return "", nil, fmt.Errorf("unsupported filter type: %s", cond.Type)
	}
}

func buildInCondition(dbFieldName string, from any) (string, []any, error) {
	val := reflect.ValueOf(from)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return "", nil, fmt.Errorf("invalid value for 'in' filter, must be a slice or array")
	}
	n := val.Len()
	if n == 0 {
		return "1=0", nil, nil
	}
	placeholders := make([]string, n)
	args := make([]any, n)
	for i := range n {
		placeholders[i] = "?"
		args[i] = val.Index(i).Interface()
	}
	return fmt.Sprintf("%s IN (%s)", dbFieldName, strings.Join(placeholders, ", ")), args, nil
}

func buildOrderClause(tType reflect.Type, sortModel *[]SortModel) (string, error) {
	if sortModel == nil || len(*sortModel) == 0 {
		return "", nil
	}
	var orderClauses []string
	for _, s := range *sortModel {
		dbFieldName, ok := GetDBFieldName(tType, s.ColId)
		if !ok {
			return "", fmt.Errorf("invalid field for sorting: %s", s.ColId)
		}
		dir := "ASC"
		if strings.EqualFold(s.Sort, "desc") {
			dir = "DESC"
		}
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", dbFieldName, dir))
	}
	return strings.Join(orderClauses, ", "), nil
}
