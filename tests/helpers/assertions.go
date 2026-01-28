package helpers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func AssertStatusCode(t *testing.T, expected int, actual int) {
	assert.Equal(t, expected, actual, "Status code mismatch")
}

func AssertJSONResponse(t *testing.T, body []byte, path string, expected interface{}) {
	result := gjson.GetBytes(body, path)

	switch v := expected.(type) {
	case string:
		assert.Equal(t, v, result.String())
	case int:
		assert.Equal(t, int64(v), result.Int())
	case bool:
		assert.Equal(t, v, result.Bool())
	case float64:
		assert.Equal(t, v, result.Float())
	}
}

func AssertJSONContains(t *testing.T, body []byte, path string) {
	result := gjson.GetBytes(body, path)
	assert.True(t, result.Exists(), "Path %s does not exist in JSON", path)
}

func AssertJSONNotEmpty(t *testing.T, body []byte, path string) {
	result := gjson.GetBytes(body, path)
	assert.True(t, result.Exists(), "Path %s does not exist", path)

	switch result.Type {
	case gjson.String:
		assert.NotEmpty(t, result.String(), "String at %s is empty", path)
	case gjson.Number:
		assert.NotZero(t, result.Int(), "Number at %s is zero", path)
	}
}

func AssertValidationError(t *testing.T, resp *http.Response, body []byte) {
	assert.Equal(t, 422, resp.StatusCode, "Expected validation error (422)")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "message")
}

func AssertUnauthorized(t *testing.T, resp *http.Response) {
	assert.Equal(t, 401, resp.StatusCode, "Expected unauthorized (401)")
}

func AssertForbidden(t *testing.T, resp *http.Response) {
	assert.Equal(t, 403, resp.StatusCode, "Expected forbidden (403)")
}

func AssertNotFound(t *testing.T, resp *http.Response) {
	assert.Equal(t, 404, resp.StatusCode, "Expected not found (404)")
}

func AssertSuccess(t *testing.T, resp *http.Response) {
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"Expected success status code (2xx), got %d", resp.StatusCode)
}
