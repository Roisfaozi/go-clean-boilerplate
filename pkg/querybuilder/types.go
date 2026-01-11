package querybuilder

type Filter struct {
	Type string      `json:"type" validate:"required,oneof=equals contains in between gt gte lt lte ne"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

type SortModel struct {
	ColId string `json:"colId" validate:"required,max=100"`
	Sort  string `json:"sort" validate:"required,oneof=asc desc ASC DESC"`
}

type DynamicFilter struct {
	Filter   map[string]Filter `json:"filter,omitempty" validate:"omitempty,dive,keys,max=100,endkeys"`
	Sort     *[]SortModel      `json:"sort,omitempty" validate:"omitempty,dive"`
	Page     int               `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize int               `json:"page_size,omitempty" validate:"omitempty,min=1"`
}

type PreloadEntity struct {
	Entity string
	Args   []interface{}
}
