package response_test

import (
	"encoding/json"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestWebResponseSuccess_JSONMarshalling(t *testing.T) {
	resp := response.WebResponseSuccess[string]{
		Data: "test data",
		Paging: &response.PageMetadata{
			Page:      1,
			Size:      10,
			Limit:     10,
			Total:     100,
			TotalItem: 100,
			TotalPage: 10,
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"data":"test data"`)
	assert.Contains(t, string(data), `"paging"`)
}

func TestWebResponseSuccess_OmitEmpty(t *testing.T) {
	resp := response.WebResponseSuccess[string]{
		Data: "test",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"paging"`)
}

func TestWebResponseError_JSONMarshalling(t *testing.T) {
	resp := response.WebResponseError[any]{
		Message: "Something went wrong",
		Error:   "internal error",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"message":"Something went wrong"`)
	assert.Contains(t, string(data), `"error":"internal error"`)
}

func TestPageResponse_JSONMarshalling(t *testing.T) {
	resp := response.PageResponse[string]{
		Data: []string{"item1", "item2", "item3"},
		PageMetadata: response.PageMetadata{
			Page:      1,
			Size:      10,
			Limit:     10,
			Total:     3,
			TotalItem: 3,
			TotalPage: 1,
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"item1"`)
	assert.Contains(t, string(data), `"paging"`)
}

func TestPageMetadata_AllFields(t *testing.T) {
	meta := response.PageMetadata{
		Page:      2,
		Size:      25,
		Limit:     50,
		Total:     100,
		TotalItem: 75,
		TotalPage: 4,
	}

	data, err := json.Marshal(meta)
	assert.NoError(t, err)

	var parsed response.PageMetadata
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	assert.Equal(t, 2, parsed.Page)
	assert.Equal(t, 25, parsed.Size)
	assert.Equal(t, 50, parsed.Limit)
	assert.Equal(t, int64(100), parsed.Total)
	assert.Equal(t, int64(75), parsed.TotalItem)
	assert.Equal(t, int64(4), parsed.TotalPage)
}

func TestWebResponseAny_JSONMarshalling(t *testing.T) {
	resp := response.WebResponseAny{
		Data:   map[string]interface{}{"key": "value"},
		Errors: "validation failed",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"data"`)
	assert.Contains(t, string(data), `"errors":"validation failed"`)
}

func TestWebResponseAny_NilPaging(t *testing.T) {
	resp := response.WebResponseAny{
		Data:   "test",
		Paging: nil,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), `"paging"`)
}
