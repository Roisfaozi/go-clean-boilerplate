package querybuilder

// Filter represents a single filtering operation
type Filter struct {
	Type string      `json:"type" validate:"required,oneof=equals contains in between gt gte lt lte ne"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// SortModel represents sorting instructions
type SortModel struct {
	ColId string `json:"colId" validate:"required,max=100"`
	Sort  string `json:"sort" validate:"required,oneof=asc desc ASC DESC"`
}

// DynamicFilter is the main payload for dynamic queries
type DynamicFilter struct {
	Filter map[string]Filter `json:"filter,omitempty" validate:"omitempty,dive,keys,max=100,endkeys"`
	Sort   *[]SortModel      `json:"sort,omitempty" validate:"omitempty,dive"`
}

// PreloadEntity represents a relationship to preload
type PreloadEntity struct {
	Entity string
	Args   []interface{}
}
