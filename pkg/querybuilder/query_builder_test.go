package querybuilder

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Age       int    `db:"age" json:"age"`
	DeletedAt int64  `db:"deleted_at" json:"deleted_at"`
}

func TestBuildRawQuery_Operators(t *testing.T) {
	tests := []struct {
		name          string
		filter        *DynamicFilter
		expectedQuery string
		expectError   bool
	}{
		{
			name: "Contains Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Name": {Type: "contains", From: "Test"},
				},
			},
			expectedQuery: "name LIKE ?",
			expectError:   false,
		},
		{
			name: "Between Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "between", From: 10, To: 20},
				},
			},
			expectedQuery: "age BETWEEN ? AND ?",
			expectError:   false,
		},
		{
			name: "In Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "in", From: []int{1, 2, 3}},
				},
			},
			expectedQuery: "age IN (?, ?, ?)",
			expectError:   false,
		},
		{
			name: "Unknown Field",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Unknown": {Type: "equals", From: 1},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := BuildRawQuery(&TestModel{}, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, res.WhereSQL)
			}
		})
	}
}

func TestBuildRawQuery_Sort(t *testing.T) {
	sorts := []SortModel{
		{ColId: "Name", Sort: "asc"},
		{ColId: "Age", Sort: "desc"},
	}
	f := &DynamicFilter{
		Sort: &sorts,
	}

	res, err := BuildRawQuery(&TestModel{}, f)
	require.NoError(t, err)

	assert.Equal(t, "name ASC, age DESC", res.OrderBy)
}

func TestGetDBFieldName(t *testing.T) {
	tType := reflect.TypeOf(TestModel{})

	name, ok := GetDBFieldName(tType, "Name")
	assert.True(t, ok)
	assert.Equal(t, "name", name)

	name, ok = GetDBFieldName(tType, "name")
	assert.True(t, ok)
	assert.Equal(t, "name", name)

	name, ok = GetDBFieldName(tType, "Age")
	assert.True(t, ok)
	assert.Equal(t, "age", name)
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "user_id", ToSnakeCase("UserID"))
	assert.Equal(t, "my_field_name", ToSnakeCase("MyFieldName"))
}
