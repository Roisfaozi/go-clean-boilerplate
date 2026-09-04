package querybuilder_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestItem struct {
	ID        string `json:"id" gorm:"column:id"`
	Name      string `json:"name" gorm:"column:name"`
	Age       int    `json:"age" gorm:"column:age"`
	Password  string `json:"password" gorm:"column:password"`
	SecretKey string `json:"secret_key" gorm:"column:secret_key"`
}

func TestBuildRawQuery_Basic(t *testing.T) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "equals", From: "Alice"},
			"age":  {Type: "gt", From: 18},
		},
		Sort: &[]querybuilder.SortModel{
			{ColId: "name", Sort: "asc"},
			{ColId: "age", Sort: "desc"},
		},
		Page:     2,
		PageSize: 20,
	}

	res, err := querybuilder.BuildRawQuery(&TestItem{}, filter)
	require.NoError(t, err)

	assert.Equal(t, "age > ? AND name = ?", res.WhereSQL)
	assert.Equal(t, []any{18, "Alice"}, res.Args)
	assert.Equal(t, "name ASC, age DESC", res.OrderBy)
	assert.Equal(t, 20, res.Limit)
	assert.Equal(t, 20, res.Offset)
}

func TestBuildRawQuery_EmptyIn(t *testing.T) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "in", From: []string{}},
		},
	}

	res, err := querybuilder.BuildRawQuery(&TestItem{}, filter)
	require.NoError(t, err)
	assert.Equal(t, "1=0", res.WhereSQL)
	assert.Empty(t, res.Args)
}

func TestBuildRawQuery_InWithElements(t *testing.T) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "in", From: []string{"A", "B"}},
		},
	}

	res, err := querybuilder.BuildRawQuery(&TestItem{}, filter)
	require.NoError(t, err)
	assert.Equal(t, "name IN (?, ?)", res.WhereSQL)
	assert.Equal(t, []any{"A", "B"}, res.Args)
}

func TestBuildRawQuery_SensitiveField(t *testing.T) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"password": {Type: "equals", From: "hidden"},
		},
	}

	_, err := querybuilder.BuildRawQuery(&TestItem{}, filter)
	assert.Error(t, err)
}
