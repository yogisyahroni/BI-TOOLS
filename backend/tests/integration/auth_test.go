package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAuthFlow(t *testing.T) {
	// 1. Setup Test User Data
	uniqueID := uuid.New().String()
	email := fmt.Sprintf("auth_test_%s@example.com", uniqueID)
	password := "Test@1234"
	username := fmt.Sprintf("User_%s", uniqueID)

	// 2. Test Registration
	t.Run("Register New User", func(t *testing.T) {
		status := registerUser(t, email, password, username)
		assert.Equal(t, http.StatusCreated, status)
	})

	// 3. Test Login Before Verification (Should fail if verification enforced, but default might allow?)
	// Actually, system requires verification. Let's verify via DB.

	// 4. Verify User via DB
	verifyUserInDB(t, email)

	// 5. Test Login Success
	var token string
	t.Run("Login Success", func(t *testing.T) {
		token = loginUser(t, email, password)
		assert.NotEmpty(t, token)
	})

	// 6. Test duplicate registration
	t.Run("Register Duplicate Email", func(t *testing.T) {
		status := registerUser(t, email, password, username)
		assert.NotEqual(t, http.StatusCreated, status)
		// Expect 409 Conflict or 400 Bad Request depending on implementation
	})

	// 7. Test Login Wrong Password
	t.Run("Login Wrong Password", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": "WrongPassword123!",
		})
		resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// 8. Test Protected Endpoint
	t.Run("Access Protected Endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/workspaces", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 9. Test Protected Endpoint No Token
	t.Run("Access Protected Endpoint No Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/workspaces", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func registerUser(t *testing.T, email, password, username string) int {
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
	return resp.StatusCode
}

func verifyUserInDB(t *testing.T, email string) {
	dbDSN := getTestDSN()
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database for verification: %v", err)
	}

	if err := db.Exec("UPDATE users SET email_verified = true, email_verified_at = NOW(), status = 'active' WHERE email = ?", email).Error; err != nil {
		t.Fatalf("Failed to verify user in DB: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func loginUser(t *testing.T, email, password string) string {
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login failed (Status %d): %s", resp.StatusCode, string(body))
	}

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)
	return res["token"]
}
