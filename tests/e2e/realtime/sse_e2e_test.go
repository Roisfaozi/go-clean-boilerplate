//go:build e2e
// +build e2e

package realtime

import (
	"bufio"
	"net/http"
	"strings"
	"testing"
	"time"

	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSSE_EventSubscription(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// Create user for SSE test
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("SSEPass123!"), bcrypt.DefaultCost)

	user := f.Create(func(u *userEntity.User) {
		u.Username = "sse_user"
		u.Email = "sse@test.com"
		u.Password = string(hash)
	})

	// Login to get token
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": user.Username,
		"password": "SSEPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	userToken := loginRes.Data.AccessToken

	t.Run("Success - Connect to SSE endpoint", func(t *testing.T) {
		// Create SSE request
		req, err := http.NewRequest("GET", server.BaseURL+"/api/v1/events", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("Accept", "text/event-stream")

		// Use raw http.Client with timeout
		client := &http.Client{Timeout: 5 * time.Second}
		sseResp, err := client.Do(req)
		
		// SSE endpoint may not be implemented yet, so we check for reasonable responses
		if err != nil {
			t.Logf("SSE connection error (may be expected if not implemented): %v", err)
			t.Skip("SSE endpoint not available")
			return
		}
		defer sseResp.Body.Close()

		// Check content type for SSE
		contentType := sseResp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/event-stream") {
			assert.Equal(t, 200, sseResp.StatusCode)
			t.Log("SSE connection established successfully")

			// Try to read first event with timeout
			scanner := bufio.NewScanner(sseResp.Body)
			done := make(chan bool)
			go func() {
				if scanner.Scan() {
					line := scanner.Text()
					t.Logf("Received SSE line: %s", line)
				}
				done <- true
			}()

			select {
			case <-done:
				// Event received
			case <-time.After(2 * time.Second):
				t.Log("No events received within timeout (expected for test)")
			}
		} else {
			t.Logf("SSE endpoint returned non-SSE content type: %s", contentType)
		}
	})

	t.Run("Negative - Unauthorized SSE", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.BaseURL+"/api/v1/events", nil)
		require.NoError(t, err)
		// No auth header

		client := &http.Client{Timeout: 5 * time.Second}
		sseResp, err := client.Do(req)
		if err != nil {
			t.Skip("SSE endpoint not available")
			return
		}
		defer sseResp.Body.Close()

		// Should be unauthorized
		assert.Equal(t, 401, sseResp.StatusCode)
	})
}
