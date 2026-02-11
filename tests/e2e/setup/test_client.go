package setup

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type TestClient struct {
	BaseURL string
	Token   string
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{BaseURL: baseURL}
}

func (c *TestClient) POST(path string, body interface{}, opts ...RequestOption) *Response {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", c.BaseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	resp, _ := http.DefaultClient.Do(req)
	return NewResponse(resp)
}

func (c *TestClient) GET(path string, opts ...RequestOption) *Response {
	req, _ := http.NewRequest("GET", c.BaseURL+path, nil)

	for _, opt := range opts {
		opt(req)
	}

	resp, _ := http.DefaultClient.Do(req)
	return NewResponse(resp)
}

func (c *TestClient) PUT(path string, body interface{}, opts ...RequestOption) *Response {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", c.BaseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	resp, _ := http.DefaultClient.Do(req)
	return NewResponse(resp)
}

func (c *TestClient) PATCH(path string, body interface{}, opts ...RequestOption) *Response {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PATCH", c.BaseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	resp, _ := http.DefaultClient.Do(req)
	return NewResponse(resp)
}

func (c *TestClient) DELETE(path string, bodyOrOpts ...interface{}) *Response {
	var body interface{}
	var opts []RequestOption

	for _, arg := range bodyOrOpts {
		switch v := arg.(type) {
		case RequestOption:
			opts = append(opts, v)
		default:
			body = v
		}
	}

	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest("DELETE", c.BaseURL+path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest("DELETE", c.BaseURL+path, nil)
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, _ := http.DefaultClient.Do(req)
	return NewResponse(resp)
}

type RequestOption func(*http.Request)

func WithAuth(token string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func WithCookie(name, value string) RequestOption {
	return func(req *http.Request) {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
}

func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func WithOrg(orgID string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set("X-Organization-ID", orgID)
	}
}

type Response struct {
	*http.Response
	BodyBytes []byte
}

func NewResponse(resp *http.Response) *Response {
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return &Response{
		Response:  resp,
		BodyBytes: bodyBytes,
	}
}

func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.BodyBytes, v)
}

func (r *Response) String() string {
	return string(r.BodyBytes)
}

func (r *Response) GetCookie(name string) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
