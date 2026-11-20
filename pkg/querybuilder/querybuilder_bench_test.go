package querybuilder_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type benchItem struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Age    int    `db:"age"`
	Status string `db:"status"`
}

func BenchmarkBuildRawQuery_SimpleFilter(b *testing.B) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "equals", From: "Alice"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := querybuilder.BuildRawQuery(&benchItem{}, filter); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildRawQuery_ComplexMultiFilterSort(b *testing.B) {
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name":   {Type: "contains", From: "Ali"},
			"age":    {Type: "between", From: 18, To: 65},
			"status": {Type: "in", From: []string{"active", "pending", "invited"}},
			"id":     {Type: "ne", From: "x"},
		},
		Sort: &[]querybuilder.SortModel{
			{ColId: "name", Sort: "asc"},
			{ColId: "age", Sort: "desc"},
		},
		Page:     3,
		PageSize: 50,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := querybuilder.BuildRawQuery(&benchItem{}, filter); err != nil {
			b.Fatal(err)
		}
	}
}
