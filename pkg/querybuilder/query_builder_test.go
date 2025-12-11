package querybuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock model for testing
type TestModel struct {
	ID        int    `gorm:"column:id"`
	Name      string `gorm:"column:name"`
	Age       int
	DeletedBy *int `gorm:"column:deleted_by"`
}

func TestGenerateDynamicQuery(t *testing.T) {
	tests := []struct {
		name          string
		filter        *DynamicFilter
		expectedQuery string
		expectedArgs  []interface{}
		expectedWarns int
	}{
		{
			name: "Contains Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Name": {Type: "contains", From: "Test"},
				},
			},
			expectedQuery: "deleted_by IS NULL AND name LIKE ?",
			expectedArgs:  []interface{}{"%Test%"},
			expectedWarns: 0,
		},
		{
			name: "InRange Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "inRange", From: 10, To: 20},
				},
			},
			expectedQuery: "deleted_by IS NULL AND age >= ? AND age <= ?",
			expectedArgs:  []interface{}{10, 20},
			expectedWarns: 0,
		},
		{
			name: "In Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "in", From: []int{1, 2, 3}},
				},
			},
			expectedQuery: "deleted_by IS NULL AND age IN (?)",
			expectedArgs:  []interface{}{[]int{1, 2, 3}},
			expectedWarns: 0,
		},
		{
			name: "Unknown Field",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Unknown": {Type: "equals", From: 1},
				},
			},
			expectedQuery: "deleted_by IS NULL",
			expectedArgs:  []interface{}{},
			expectedWarns: 1,
		},
		{
			name:          "Empty Filter (Default Soft Delete)",
			filter:        &DynamicFilter{},
			expectedQuery: "deleted_by IS NULL",
			expectedArgs:  []interface{}{},
			expectedWarns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, args, warns, err := GenerateDynamicQuery[TestModel](tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedQuery, q)

			assert.Equal(t, len(tt.expectedArgs), len(args))
			if len(args) > 0 {
				assert.Equal(t, tt.expectedArgs[0], args[0])
			}

			assert.Len(t, warns, tt.expectedWarns)
		})
	}
}

func TestGenerateDynamicSort(t *testing.T) {
	sorts := []SortModel{
		{ColId: "Name", Sort: "asc"},
		{ColId: "Age", Sort: "desc"},
	}
	f := &DynamicFilter{Sort: &sorts}

	s, err := GenerateDynamicSort[TestModel](f)
	require.NoError(t, err)
	assert.Contains(t, s, "name ASC")
	assert.Contains(t, s, "age DESC")
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "user_id", ToSnakeCase("UserID"))
	assert.Equal(t, "my_field_name", ToSnakeCase("MyFieldName"))
}
