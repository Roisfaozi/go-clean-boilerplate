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

func TestGitHubProvider_GetUserInfo_SuccessWithPublicEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 12345,
			"login": "octocat",
			"name": "Octocat",
			"email": "octocat@github.com",
			"avatar_url": "https://github.com/images/error/octocat_happy.gif"
		}`))
	}))
	defer server.Close()

	provider := NewGitHubProvider(ProviderConfig{})

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
	assert.Equal(t, "octocat@github.com", userInfo.Email)
	assert.Equal(t, "12345", userInfo.ProviderID)
	assert.Equal(t, "Octocat", userInfo.Name)
}

func TestGitHubProvider_GetUserInfo_FallbackToEmailEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"id": 67890,
				"login": "private_user"
			}`))
			return
		}

		if r.URL.Path == "/user/emails" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"email": "secondary@example.com",
					"primary": false,
					"verified": true
				},
				{
					"email": "primary@example.com",
					"primary": true,
					"verified": true
				}
			]`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := NewGitHubProvider(ProviderConfig{})

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
	assert.Equal(t, "primary@example.com", userInfo.Email)
	assert.Equal(t, "67890", userInfo.ProviderID)
	assert.Equal(t, "private_user", userInfo.Name)
}

func TestGitHubProvider_GetUserInfo_ErrorNoEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id": 111, "login": "noemail"}`))
			return
		}

		if r.URL.Path == "/user/emails" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"email": "unverified@example.com", "primary": true, "verified": false}]`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	provider := NewGitHubProvider(ProviderConfig{})

	hijackClient := &http.Client{
		Transport: &MockTransport{
			ServerURL: server.URL,
		},
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hijackClient)

	token := &oauth2.Token{AccessToken: "mock-token"}
	_, err := provider.GetUserInfo(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no verified email found")
}

func TestGitHubProvider_GetUserInfo_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewGitHubProvider(ProviderConfig{})

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
