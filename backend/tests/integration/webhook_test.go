package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Webhook struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description string   `json:"description"`
	IsActive    bool     `json:"isActive"`
}

func TestWebhookFlow(t *testing.T) {
	// 1. Setup Callback Server
	received := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Webhook received!")
		// Verify signature header exists
		if r.Header.Get("X-Insight-Signature") == "" {
			t.Error("Missing X-Insight-Signature header")
		}
		w.WriteHeader(http.StatusOK)
		close(received)
	}))
	defer server.Close()

	// 2. Login
	token := login(t)

	// 3. Create Webhook
	webhook := Webhook{
		Name:        "Test Webhook",
		URL:         server.URL,
		Events:      []string{"test.event"},
		Description: "Integration Test Webhook",
		IsActive:    true,
	}

	webhookID := createWebhook(t, token, webhook)
	t.Logf("Created webhook with ID: %s", webhookID)

	// 4. Trigger Test Event
	triggerTestEvent(t, token, webhookID)

	// 5. Wait for Webhook Delivery
	select {
	case <-received:
		t.Log("Webhook delivery confirmed")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for webhook delivery")
	}

	// 6. Delete Webhook
	deleteWebhook(t, token, webhookID)
}

// --- Helpers ---

func login(t *testing.T) string {
	// Use existing demo user credentials or separate test user
	// For integration tests on running env, we assume the user exists or we create one.
	// We'll use the one from test_webhook.py: demo@insight.com / demo123
	// OR create a new one to be safe. Let's create one.

	// Register a new user for testing
	uniqueID := uuid.New().String()
	email := fmt.Sprintf("webhook_test_%s@example.com", uniqueID)
	password := "Test@1234"
	username := fmt.Sprintf("User_%s", uniqueID)

	// Register
	registerBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"username": username,
	})
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(registerBody))
	if err != nil {
		t.Fatalf("Registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Registration failed: %s", string(body))
	}

	// Manual Verification via DB
	dbDSN := "postgresql://postgres:1234@localhost:5432/Inside_engineer1?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database for verification: %v", err)
	}

	// Update user to be verified
	if err := db.Exec("UPDATE users SET email_verified = true, email_verified_at = NOW(), status = 'active' WHERE email = ?", email).Error; err != nil {
		t.Fatalf("Failed to verify user in DB: %v", err)
	}

	// Give a small delay for DB consistency (optional but safe)
	time.Sleep(100 * time.Millisecond)

	// Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(loginBody))

	if err != nil || resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login failed: %s", string(body))
	}
	defer resp.Body.Close()

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)
	token := res["token"]
	t.Logf("Got token: %s", token)
	if token == "" {
		t.Fatal("Login succeeded but returned empty token")
	}
	return token
}

func createWebhook(t *testing.T, token string, w Webhook) string {
	body, _ := json.Marshal(w)
	req, _ := http.NewRequest("POST", baseURL+"/webhooks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Create Webhook failed (Status %d): %s", resp.StatusCode, string(body))
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if res["id"] == nil {
		t.Fatalf("Create Webhook returned no ID: %v", res)
	}
	return res["id"].(string)
}

func triggerTestEvent(t *testing.T, token, id string) {
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/webhooks/%s/test", baseURL, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func deleteWebhook(t *testing.T, token, id string) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/webhooks/%s", baseURL, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
