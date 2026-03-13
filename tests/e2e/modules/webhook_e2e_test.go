//go:build e2e
// +build e2e

package modules

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	integrationSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookE2E_UserRegistrationTrigger(t *testing.T) {
	// 1. Setup Environment
	server := setup.SetupTestServer(t)
	defer server.Server.Close()

	// 2. Create Superadmin for Auth
	admin := integrationSetup.CreateTestUser(t, server.DB, "superadmin_e2e", "superadmin_e2e@test.com", "Password123!")
	// Assign superadmin role in global domain via Enforcer to ensure it's loaded
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	err = server.Enforcer.LoadPolicy()
	require.NoError(t, err)

	// 3. Login to get token
	loginResp := server.Client.POST("/api/v1/auth/login", map[string]string{
		"username": "superadmin_e2e",
		"password": "Password123!",
	})
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed: %d %s", loginResp.StatusCode, loginResp.String())
	}

	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = loginResp.JSON(&loginData)
	require.NoError(t, err)
	server.Client.Token = loginData.Data.AccessToken
	t.Log("Login successful")

	// 4. Setup Mock Webhook Receiver
	receivedPayload := make(chan []byte, 1)
	mockReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Mock receiver received request")
		body, _ := io.ReadAll(r.Body)
		receivedPayload <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer mockReceiver.Close()
	t.Logf("Mock receiver started at %s", mockReceiver.URL)

	// 5. Create Webhook via API
	webhookReq := model.CreateWebhookRequest{
		Name:           "E2E Registration Webhook",
		OrganizationID: "global",
		URL:            mockReceiver.URL,
		Events:         []string{"user.created"},
		Secret:         "e2e-secret-key-12345",
	}

	t.Log("Creating webhook via API...")
	createResp := server.Client.POST("/api/v1/webhooks?organization_id=global", webhookReq)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Create webhook failed: %d %s", createResp.StatusCode, createResp.String())
	}
	t.Log("Webhook created")

	// 6. Trigger Event: Register a new user
	regReq := map[string]string{
		"username": "new_e2e_user",
		"email":    "new_e2e@example.com",
		"password": "Password123!",
		"fullname": "E2E User",
	}

	t.Log("Registering new user to trigger webhook...")
	regResp := server.Client.POST("/api/v1/users/register", regReq)
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("User registration failed: %d %s", regResp.StatusCode, regResp.String())
	}
	t.Log("User registered")

	// 7. Verify Webhook Delivery
	select {
	case payload := <-receivedPayload:
		var data map[string]interface{}
		err = json.Unmarshal(payload, &data)
		require.NoError(t, err)
		// Payload from user registration usually contains the user info
		assert.Equal(t, "new_e2e_user", data["username"])
		assert.Equal(t, "new_e2e@example.com", data["email"])
	case <-time.After(10 * time.Second):
		t.Fatal("E2E Webhook delivery timeout")
	}
}
