package querybuilder

import (
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Mock model for testing
type TestModel struct {
	ID        int            `gorm:"column:id"`
	Name      string         `gorm:"column:name"`
	Age       int            `gorm:"column:age"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func setupDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true, // DryRun prevents actual DB execution, useful for SQL generation check
	})
	return db
}

func TestGenerateDynamicQuery(t *testing.T) {
	db := setupDB()

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
			expectedQuery: "name LIKE",
			expectError:   false,
		},
		{
			name: "Between Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "between", From: 10, To: 20},
				},
			},
			expectedQuery: "age BETWEEN",
			expectError:   false,
		},
		{
			name: "In Operator",
			filter: &DynamicFilter{
				Filter: map[string]Filter{
					"Age": {Type: "in", From: []int{1, 2, 3}},
				},
			},
			expectedQuery: "age IN",
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
			query := db.Model(&TestModel{})
			resQuery, err := GenerateDynamicQuery(query, &TestModel{}, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// ToSQL generates the SQL string
				sql := resQuery.Find(&[]TestModel{}).Statement.SQL.String()
				assert.Contains(t, sql, tt.expectedQuery)
			}
		})
	}
}

func TestGenerateDynamicSort(t *testing.T) {
	db := setupDB()

	sorts := []SortModel{
		{ColId: "Name", Sort: "asc"},
		{ColId: "Age", Sort: "desc"},
	}
	f := &DynamicFilter{
		Sort: &sorts,
	}

	query := db.Model(&TestModel{})
	resQuery, err := GenerateDynamicSort(query, &TestModel{}, f)
	require.NoError(t, err)

	sql := resQuery.Find(&[]TestModel{}).Statement.SQL.String()

	// GORM SQL generation order for ORDER BY might vary slightly or be combined
	assert.Contains(t, sql, "ORDER BY")
	assert.Contains(t, sql, "name asc")
	assert.Contains(t, sql, "age desc")
}

func TestGetDBFieldName(t *testing.T) {
	tType := reflect.TypeOf(TestModel{})

	name, ok := GetDBFieldName(tType, "Name")
	assert.True(t, ok)
	assert.Equal(t, "name", name)

	name, ok = GetDBFieldName(tType, "name") // Lowercase lookup
	assert.True(t, ok)
	assert.Equal(t, "name", name)

	name, ok = GetDBFieldName(tType, "Age")
	assert.True(t, ok)
	assert.Equal(t, "age", name)
}

func TestToSnakeCase(t *testing.T) {
	// Updated expectations based on regex implementation behavior
	assert.Equal(t, "user_id", ToSnakeCase("UserID"))
	assert.Equal(t, "my_field_name", ToSnakeCase("MyFieldName"))
}
