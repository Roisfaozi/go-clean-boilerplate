package sso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogleProvider_GetUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth2/v2/userinfo", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "12345",
			"email": "test@example.com",
			"verified_email": true,
			"name": "Test User",
			"picture": "https://example.com/avatar.jpg"
		}`))
	}))
	defer server.Close()

	provider := NewGoogleProvider(ProviderConfig{})

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{AccessToken: "mock-token"}
	userInfo, err := provider.GetUserInfo(ctx, token)

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "12345", userInfo.ProviderID)
	assert.Equal(t, "Test User", userInfo.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", userInfo.AvatarURL)
}

func TestGoogleProvider_GetUserInfo_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewGoogleProvider(ProviderConfig{})

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{AccessToken: "mock-token"}
	_, err := provider.GetUserInfo(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code 500")
}

func TestGoogleProvider_GetUserInfo_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer server.Close()

	provider := NewGoogleProvider(ProviderConfig{})

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{AccessToken: "mock-token"}
	_, err := provider.GetUserInfo(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed decoding user info")
}

func TestGoogleProvider_GetLoginURL(t *testing.T) {
	provider := NewGoogleProvider(ProviderConfig{
		ClientID:    "test-client",
		RedirectURL: "http://localhost/callback",
	})

	url := provider.GetLoginURL("test-state")
	assert.Contains(t, url, "client_id=test-client")
	assert.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback")
	assert.Contains(t, url, "state=test-state")
}

func TestGoogleProvider_ExchangeCode(t *testing.T) {
	// Simple test to ensure it delegates to oauth2.Config correctly.
	// Since we can't easily mock oauth2.Config.Exchange without intercepting the HTTP call.
	// We'll intercept the HTTP call.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`access_token=mock-token&token_type=bearer`))
	}))
	defer server.Close()

	provider := NewGoogleProvider(ProviderConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	})
	provider.config.Endpoint = oauth2.Endpoint{
		TokenURL: server.URL,
	}

	hijackClient := &http.Client{
		// No need for MockTransport because TokenURL is an exact URL, not relative.
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token, err := provider.ExchangeCode(ctx, "mock-code")

	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "mock-token", token.AccessToken)
}


func TestGoogleProvider_GetUserInfo_GetError(t *testing.T) {
	provider := NewGoogleProvider(ProviderConfig{})

	// Context with timeout to intentionally fail the HTTP request if it tries to dial.
	// Alternatively, we can use a mock transport that always returns an error.
	hijackClient := &http.Client{
		Transport: &ErrorTransport{},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{AccessToken: "mock-token"}
	_, err := provider.GetUserInfo(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed getting user info")
}

type ErrorTransport struct{}

func (t *ErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, assert.AnError
}

func TestGoogleProvider_EdgeAndVulnerability(t *testing.T) {
	provider := NewGoogleProvider(ProviderConfig{})

	// Edge Case: Empty token gets unauthorized
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server1.Close()

	hijackClient1 := &http.Client{
		Transport: &MockTransport{
			ServerURL: server1.URL,
		},
	}
	ctx1 := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient1)
	_, err := provider.GetUserInfo(ctx1, &oauth2.Token{AccessToken: ""})
	assert.Error(t, err)

	// Edge Case: Malformed JSON Response (Vulnerability simulation)
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{ "id": `)) // Incomplete JSON
	}))
	defer server2.Close()

	hijackClient2 := &http.Client{
		Transport: &MockTransport{
			ServerURL: server2.URL,
		},
	}
	ctx2 := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient2)

	token := &oauth2.Token{AccessToken: "mock-token"}
	_, err = provider.GetUserInfo(ctx2, token)
	assert.Error(t, err)
}
