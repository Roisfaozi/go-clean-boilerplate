package sso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestMicrosoftProvider_GetUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1.0/me", r.URL.Path)
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ms-user-123",
			"displayName": "Microsoft User",
			"userPrincipalName": "msuser@tenant.onmicrosoft.com",
			"mail": "msuser@example.com"
		}`))
	}))
	defer server.Close()

	provider := NewMicrosoftProvider(ProviderConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/callback",
	})
	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{
		AccessToken: "mock-token",
		TokenType:   "Bearer",
	}

	userInfo, err := provider.GetUserInfo(ctx, token)

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "msuser@example.com", userInfo.Email)
	assert.Equal(t, "ms-user-123", userInfo.ProviderID)
	assert.Equal(t, "Microsoft User", userInfo.Name)
}

func TestMicrosoftProvider_GetUserInfo_FallbackEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ms-user-456",
			"displayName": "Microsoft User 2",
			"userPrincipalName": "upn@example.com",
			"mail": ""
		}`))
	}))
	defer server.Close()

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	provider := NewMicrosoftProvider(ProviderConfig{})
	token := &oauth2.Token{AccessToken: "mock-token"}

	userInfo, err := provider.GetUserInfo(ctx, token)

	require.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "upn@example.com", userInfo.Email)
}

func TestMicrosoftProvider_GetUserInfo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	provider := NewMicrosoftProvider(ProviderConfig{})
	token := &oauth2.Token{AccessToken: "mock-token"}

	_, err := provider.GetUserInfo(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code 500")
}

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
