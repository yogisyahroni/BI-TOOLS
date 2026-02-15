package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"insight-engine-backend/database"
	"insight-engine-backend/dtos"
	"insight-engine-backend/models"
	"insight-engine-backend/services"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupAuthTestDB initializes in-memory SQLite DB for testing
func setupAuthTestDB() {
	var err error
	// Use shared cache for in-memory DB to share state across connections
	database.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	sqlDB, err := database.DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	// Migrate schemas
	database.DB.AutoMigrate(&models.User{})
}

// setupAuthTestApp creates a Fiber app with auth routes for testing
func setupAuthTestApp() (*fiber.App, *services.AuthService) {
	setupAuthTestDB() // Initialize DB

	app := fiber.New()

	emailService := services.NewEmailService()
	emailService.Mock = true // Disable real email sending
	authService := services.NewAuthService(database.DB, emailService)
	authHandler := NewAuthHandler(authService)

	app.Post("/api/auth/register", authHandler.Register)
	app.Post("/api/auth/login", authHandler.Login)
	app.Get("/api/auth/verify-email", authHandler.VerifyEmail)
	app.Post("/api/auth/resend-verification", authHandler.ResendVerification)

	return app, authService
}

// cleanupTestUser removes test user from database
func cleanupTestUser(email string) {
	database.DB.Where("email = ?", email).Delete(&models.User{})
}

// TestRegister_Success tests successful user registration
func TestRegister_Success(t *testing.T) {
	app, _ := setupAuthTestApp()

	// Test data
	reqBody := dtos.RegisterRequest{
		Email:    "test.success@example.com",
		Username: "testusersuccess",
		Password: "SecurePass123!",
		FullName: "Test User",
	}

	// Cleanup after test
	defer cleanupTestUser(reqBody.Email)

	// Create request
	// jsonBody, err := json.Marshal(reqBody)
	// require.NoError(t, err)
	// t.Logf("JSON Body: %s", string(jsonBody))

	jsonBody := []byte(`{"email":"test.success@example.com","username":"testusersuccess","password":"SecurePass123!","fullName":"Test User"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(jsonBody))

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)

	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Logf("Response Body: %s", buf.String())
	}
	require.Equal(t, 201, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Assertions
	assert.Equal(t, "success", result["status"])
	data := result["data"].(map[string]interface{})
	assert.NotEmpty(t, data["userId"])
	assert.Equal(t, reqBody.Email, data["email"])
	assert.Equal(t, reqBody.Username, data["username"])
	assert.Equal(t, "Registration successful. Please check your email to verify your account.", data["message"])

	// Verify user created in database
	var user models.User
	err = database.DB.Where("email = ?", reqBody.Email).First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, reqBody.Email, user.Email)
	assert.Equal(t, reqBody.Username, user.Username)
	assert.Equal(t, reqBody.FullName, user.Name)

	// Verify password is hashed (not plaintext)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password))
	assert.NoError(t, err, "Password should be properly hashed")
}

// TestRegister_DuplicateEmail tests registration with existing email
func TestRegister_DuplicateEmail(t *testing.T) {
	app, _ := setupAuthTestApp()

	// Create first user
	reqBody := dtos.RegisterRequest{
		Email:    "test.duplicate@example.com",
		Username: "testuserdup", // Pure alpha to avoid strict validation issues
		Password: "SecurePass123!",
		FullName: "Test User",
	}

	// Cleanup after test
	defer cleanupTestUser(reqBody.Email)

	// Create first user
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	respFirst, err := app.Test(req)
	assert.NoError(t, err)
	if respFirst.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(respFirst.Body)
		t.Fatalf("First registration failed. Status: %d, Body: %s", respFirst.StatusCode, buf.String())
	}

	// Try to create second user with same email
	reqBody2 := dtos.RegisterRequest{
		Email:    "test.duplicate@example.com", // Same email
		Username: "testuserdup2",               // Different username
		Password: "SecurePass123!",
		FullName: "Test User 2",
	}

	jsonBody2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "error", result["status"])
	assert.Contains(t, result["message"], "email already registered")
}

// TestRegister_DuplicateUsername tests registration with existing username
func TestRegister_DuplicateUsername(t *testing.T) {
	app, _ := setupAuthTestApp()

	// Create first user
	reqBody := dtos.RegisterRequest{
		Email:    "test.user1@example.com",
		Username: "testuserunique",
		Password: "SecurePass123!",
		FullName: "Test User",
	}

	// Cleanup after test
	defer cleanupTestUser(reqBody.Email)

	defer cleanupTestUser("test.user2@example.com")

	// Create first user
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	respFirst, errFirst := app.Test(req)
	require.NoError(t, errFirst)
	if respFirst.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(respFirst.Body)
		t.Fatalf("Registration failed. Status: %d, Body: %s", respFirst.StatusCode, buf.String())
	}

	// Try to create second user with same username
	reqBody2 := dtos.RegisterRequest{
		Email:    "test.user2@example.com", // Different email
		Username: "testuserunique",         // Same username
		Password: "SecurePass123!",
		FullName: "Test User 2",
	}

	jsonBody2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "error", result["status"])
	assert.Contains(t, result["message"], "username already taken")
}

// TestRegister_InvalidEmail tests registration with invalid email formats
func TestRegister_InvalidEmail(t *testing.T) {
	app, _ := setupAuthTestApp()

	invalidEmails := []string{
		"",
		"invalid-email",
		"@example.com",
		"test@",
		"test@.com",
	}

	for _, email := range invalidEmails {
		reqBody := dtos.RegisterRequest{
			Email:    email,
			Username: "testuser_invalid",
			Password: "SecurePass123!",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode, "Should fail for email: %s", email)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "error", result["status"])
	}
}

// TestRegister_InvalidUsername tests registration with invalid usernames
func TestRegister_InvalidUsername(t *testing.T) {
	app, _ := setupAuthTestApp()

	invalidUsernames := []struct {
		username string
		reason   string
	}{
		{"", "empty username"},
		{"ab", "too short (min 3)"},
		{"user@name", "special characters"},
		{"user name", "spaces"},
	}

	for _, tc := range invalidUsernames {
		reqBody := dtos.RegisterRequest{
			Email:    "test.invalid@example.com",
			Username: tc.username,
			Password: "SecurePass123!",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err, "Test case: %s", tc.reason)
		assert.Equal(t, 400, resp.StatusCode, "Should fail for username: %s (%s)", tc.username, tc.reason)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "error", result["status"])
	}
}

// TestRegister_InvalidPassword tests registration with weak passwords
func TestRegister_InvalidPassword(t *testing.T) {
	app, _ := setupAuthTestApp()

	invalidPasswords := []struct {
		password string
		reason   string
	}{
		{"", "empty password"},
		{"1234567", "too short (7 chars, min 8)"},
		{"short", "too short (5 chars)"},
	}

	for _, tc := range invalidPasswords {
		reqBody := dtos.RegisterRequest{
			Email:    "test.password@example.com",
			Username: "testuser_password",
			Password: tc.password,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err, "Test case: %s", tc.reason)
		assert.Equal(t, 400, resp.StatusCode, "Should fail for password: %s (%s)", tc.password, tc.reason)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "error", result["status"])
	}
}

// TestRegister_MissingRequiredFields tests registration with missing fields
func TestRegister_MissingRequiredFields(t *testing.T) {
	app, _ := setupAuthTestApp()

	testCases := []struct {
		name     string
		reqBody  dtos.RegisterRequest
		expected string
	}{
		{
			name:     "missing email",
			reqBody:  dtos.RegisterRequest{Username: "testuser", Password: "SecurePass123!"},
			expected: "email",
		},
		{
			name:     "missing username",
			reqBody:  dtos.RegisterRequest{Email: "test@example.com", Password: "SecurePass123!"},
			expected: "username",
		},
		{
			name:     "missing password",
			reqBody:  dtos.RegisterRequest{Email: "test@example.com", Username: "testuser"},
			expected: "password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, 400, resp.StatusCode)

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			assert.Equal(t, "error", result["status"])
		})
	}
}

// TestRegister_CanLoginAfterRegistration tests that user can login immediately after registration
func TestRegister_CanLoginAfterRegistration(t *testing.T) {
	app, _ := setupAuthTestApp()

	// Register user
	registerBody := dtos.RegisterRequest{
		Email:    "verify.login@example.com",
		Username: "testuserlogin",
		Password: "SecurePass123!",
		FullName: "Test User",
	}

	// Cleanup
	defer cleanupTestUser(registerBody.Email)

	// Register
	jsonBody, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Fatalf("Registration failed. Status: %d, Body: %s", resp.StatusCode, buf.String())
	}
	assert.Equal(t, 201, resp.StatusCode)

	// Login with same credentials
	loginBody := map[string]string{
		"email":    registerBody.Email,
		"password": registerBody.Password,
	}
	loginJson, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginJson))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq)
	assert.NoError(t, err)
	assert.Equal(t, 200, loginResp.StatusCode)

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	assert.NotNil(t, loginResult["user"])
	assert.NotEmpty(t, loginResult["token"])
}

// TestRegister_InvalidJSON tests registration with malformed JSON
func TestRegister_InvalidJSON(t *testing.T) {
	app, _ := setupAuthTestApp()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "error", result["status"])
	assert.Contains(t, result["message"], "Invalid request body")
}
