package response

type WebResponse[T any] struct {
	Data   T             `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
	Error  string        `json:"error,omitempty"`
}

type PageResponse[T any] struct {
	Data         []T          `json:"data,omitempty"`
	PageMetadata PageMetadata `json:"paging,omitempty"`
}

type PageMetadata struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

// WebResponseAny is a non-generic wrapper for WebResponse to be used in Swagger docs
type WebResponseAny struct {
	Data   interface{}   `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
	Error  string        `json:"error,omitempty"`
}
