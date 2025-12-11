package querybuilder

// Filter represents a single filtering operation
type Filter struct {
	Type string      `json:"type"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// SortModel represents sorting instructions
type SortModel struct {
	ColId string `json:"colId"`
	Sort  string `json:"sort"`
}

// DynamicFilter is the main payload for dynamic queries
type DynamicFilter struct {
	Filter map[string]Filter `json:"filter,omitempty"`
	Sort   *[]SortModel      `json:"sort,omitempty"`
}

// PreloadEntity represents a relationship to preload
type PreloadEntity struct {
	Entity string
	Args   []interface{}
}
