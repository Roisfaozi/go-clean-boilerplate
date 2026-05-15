package sso

import (
	"net/http"
)

type MockTransport struct {
	ServerURL string
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"

	mockServerURL := req.URL
	var err error
	if mockServerURL, err = mockServerURL.Parse(t.ServerURL + req.URL.Path); err != nil {
		return nil, err
	}
	req.URL = mockServerURL
	req.Host = req.URL.Host

	return http.DefaultTransport.RoundTrip(req)
}
