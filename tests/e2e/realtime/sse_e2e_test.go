//go:build e2e
// +build e2e

package realtime

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEE2E_NotificationFlow(t *testing.T) {
	// 1. Setup Server
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// 2. Establish SSE Connection with Timeout
	sseURL := server.BaseURL + "/events"
	t.Logf("Connecting to SSE URL: %s", sseURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", sseURL, nil)
	req.Header.Set("Accept", "text/event-stream")

	httpClient := &http.Client{}
	respSSE, err := httpClient.Do(req)
	require.NoError(t, err)
	defer respSSE.Body.Close()

	require.Equal(t, 200, respSSE.StatusCode)
	t.Log("SSE Connection established")

	eventChan := make(chan string, 10)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(respSSE.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				eventChan <- data
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	// 3. Perform Action: Login via REST API
	f := fixtures.NewUserFactory(server.DB)
	user := f.Create()

	loginPayload := map[string]any{
		"username": user.Username,
		"password": "password123",
	}

	t.Log("Performing login to trigger SSE...")
	respLogin := server.Client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, respLogin.StatusCode)

	// 4. Verify Notification Received via SSE
	select {
	case data := <-eventChan:
		t.Logf("Received SSE data: %s", data)
		var notification struct {
			Type   string `json:"type"`
			UserID string `json:"user_id"`
		}
		err := json.Unmarshal([]byte(data), &notification)
		require.NoError(t, err)

		assert.Equal(t, "user_login", notification.Type)
		assert.Equal(t, user.ID, notification.UserID)

	case err := <-errChan:
		t.Fatalf("SSE stream error: %v", err)

	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for SSE notification")
	}
}
