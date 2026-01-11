package response

type WebResponseSuccess[T any] struct {
	Data   T             `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
}

type WebResponseError[T any] struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PageResponse[T any] struct {
	Data         []T          `json:"data,omitempty"`
	PageMetadata PageMetadata `json:"paging,omitempty"`
}

type PageMetadata struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

type WebResponseAny struct {
	Data   interface{}   `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
	Error  string        `json:"error,omitempty"`
}
