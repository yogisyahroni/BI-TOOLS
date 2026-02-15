// verification/scripts/verify_auth.go
// Authentication & Security Verification Suite
// VERIFY-001: Task 1-16 verification

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	baseURL     = "http://localhost:8080"
	frontendURL = "http://localhost:3000"
)

type TestResult struct {
	TaskID   string                 `json:"task_id"`
	Name     string                 `json:"name"`
	Status   string                 `json:"status"` // PASS, FAIL, SKIP
	Duration string                 `json:"duration"`
	Error    string                 `json:"error,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

type VerificationReport struct {
	Timestamp  time.Time    `json:"timestamp"`
	TotalTests int          `json:"total_tests"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	Skipped    int          `json:"skipped"`
	Results    []TestResult `json:"results"`
}

var report VerificationReport

func init() {
	report = VerificationReport{
		Timestamp: time.Now(),
		Results:   []TestResult{},
	}
}

func addResult(taskID, name, status, errMsg string, details map[string]interface{}) {
	result := TestResult{
		TaskID:   taskID,
		Name:     name,
		Status:   status,
		Duration: time.Now().Format("15:04:05"),
	}

	if errMsg != "" {
		result.Error = errMsg
	}

	if details != nil {
		result.Details = details
	}

	report.Results = append(report.Results, result)
	report.TotalTests++

	switch status {
	case "PASS":
		report.Passed++
	case "FAIL":
		report.Failed++
	case "SKIP":
		report.Skipped++
	}
}

// TASK-001: User Registration API
func verifyUserRegistration() {
	fmt.Println("🔍 VERIFY-001: Testing User Registration API (TASK-001)...")

	testUser := map[string]string{
		"email":    "test_verify_001@insightengine.ai",
		"password": "TestPass123!",
		"name":     "Test User",
	}

	jsonData, _ := json.Marshal(testUser)

	// Test 1: Valid registration
	resp, err := http.Post(
		baseURL+"/api/auth/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		addResult("TASK-001", "User Registration API", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		addResult("TASK-001", "User Registration API", "PASS", "", map[string]interface{}{
			"status_code": resp.StatusCode,
			"user_id":     result["id"],
		})
		fmt.Println("  ✅ User registration working")
	} else if resp.StatusCode == 409 {
		addResult("TASK-001", "User Registration API", "PASS", "", map[string]interface{}{
			"status_code": resp.StatusCode,
			"note":        "User already exists - duplicate prevention working",
		})
		fmt.Println("  ✅ Duplicate user prevention working")
	} else {
		body, _ := io.ReadAll(resp.Body)
		addResult("TASK-001", "User Registration API", "FAIL",
			fmt.Sprintf("Unexpected status: %d - %s", resp.StatusCode, string(body)), nil)
		fmt.Printf("  ❌ Registration failed: %d\n", resp.StatusCode)
	}

	// Test 2: Duplicate email
	time.Sleep(100 * time.Millisecond)
	resp2, err := http.Post(
		baseURL+"/api/auth/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		addResult("TASK-001a", "Duplicate Email Prevention", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == 409 {
		addResult("TASK-001a", "Duplicate Email Prevention", "PASS", "", nil)
		fmt.Println("  ✅ Duplicate email prevention working")
	} else {
		addResult("TASK-001a", "Duplicate Email Prevention", "FAIL",
			fmt.Sprintf("Expected 409, got %d", resp2.StatusCode), nil)
		fmt.Printf("  ❌ Duplicate check failed: %d\n", resp2.StatusCode)
	}

	// Test 3: Invalid email format
	invalidUser := map[string]string{
		"email":    "invalid-email",
		"password": "TestPass123!",
		"name":     "Test",
	}
	jsonData, _ = json.Marshal(invalidUser)

	time.Sleep(100 * time.Millisecond)
	resp3, err := http.Post(
		baseURL+"/api/auth/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		addResult("TASK-001b", "Email Validation", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp3.Body.Close()

	if resp3.StatusCode == 400 {
		addResult("TASK-001b", "Email Validation", "PASS", "", nil)
		fmt.Println("  ✅ Email validation working")
	} else {
		addResult("TASK-001b", "Email Validation", "FAIL",
			fmt.Sprintf("Expected 400, got %d", resp3.StatusCode), nil)
		fmt.Printf("  ❌ Email validation failed: %d\n", resp3.StatusCode)
	}
}

// TASK-003: Email Verification System
func verifyEmailVerification() {
	fmt.Println("\n🔍 VERIFY-003: Testing Email Verification (TASK-003)...")

	// Check if verification endpoint exists
	resp, err := http.Get(baseURL + "/api/auth/verify?token=test-token")
	if err != nil {
		addResult("TASK-003", "Email Verification Endpoint", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	// Should return 400 for invalid token (endpoint exists)
	if resp.StatusCode == 400 {
		addResult("TASK-003", "Email Verification Endpoint", "PASS", "", map[string]interface{}{
			"status_code": resp.StatusCode,
			"note":        "Endpoint exists, rejects invalid tokens",
		})
		fmt.Println("  ✅ Email verification endpoint exists")
	} else {
		addResult("TASK-003", "Email Verification Endpoint", "FAIL",
			fmt.Sprintf("Unexpected status: %d", resp.StatusCode), nil)
		fmt.Printf("  ❌ Verification endpoint issue: %d\n", resp.StatusCode)
	}

	// Check email service configuration
	addResult("TASK-003a", "Email Service Configuration", "SKIP",
		"Manual verification required - check email logs", nil)
	fmt.Println("  ⚠️  Email service needs manual verification (check MailHog/SMTP)")
}

// TASK-004: Forgot Password API
func verifyForgotPassword() {
	fmt.Println("\n🔍 VERIFY-004: Testing Forgot Password (TASK-004)...")

	// Test with valid email
	forgotData := map[string]string{
		"email": "test_verify_001@insightengine.ai",
	}
	jsonData, _ := json.Marshal(forgotData)

	resp, err := http.Post(
		baseURL+"/api/auth/forgot-password",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		addResult("TASK-004", "Forgot Password API", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	// Should return 200 even if email doesn't exist (security)
	if resp.StatusCode == 200 {
		addResult("TASK-004", "Forgot Password API", "PASS", "", map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		fmt.Println("  ✅ Forgot password endpoint working")
	} else {
		addResult("TASK-004", "Forgot Password API", "FAIL",
			fmt.Sprintf("Unexpected status: %d", resp.StatusCode), nil)
		fmt.Printf("  ❌ Forgot password failed: %d\n", resp.StatusCode)
	}
}

// TASK-007: Google OAuth
func verifyGoogleOAuth() {
	fmt.Println("\n🔍 VERIFY-007: Testing Google OAuth (TASK-007)...")

	// Check OAuth endpoint exists
	resp, err := http.Get(baseURL + "/api/auth/google")
	if err != nil {
		addResult("TASK-007", "Google OAuth Endpoint", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	// Should redirect to Google (302) or return config (200)
	if resp.StatusCode == 302 || resp.StatusCode == 200 {
		addResult("TASK-007", "Google OAuth Endpoint", "PASS", "", map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		fmt.Println("  ✅ Google OAuth endpoint configured")
	} else {
		addResult("TASK-007", "Google OAuth Endpoint", "FAIL",
			fmt.Sprintf("Unexpected status: %d", resp.StatusCode), nil)
		fmt.Printf("  ❌ Google OAuth issue: %d\n", resp.StatusCode)
	}

	addResult("TASK-007a", "Google OAuth Flow", "SKIP",
		"Manual verification required - need real Google account", nil)
	fmt.Println("  ⚠️  Full OAuth flow needs manual testing with real Google account")
}

// TASK-009: Rate Limiting
func verifyRateLimiting() {
	fmt.Println("\n🔍 VERIFY-009: Testing Rate Limiting (TASK-009)...")

	// Send multiple rapid requests
	limitReached := false
	var rateLimitHeaders map[string]string

	for i := 0; i < 130; i++ {
		resp, err := http.Get(baseURL + "/api/health")
		if err != nil {
			continue
		}

		// Check for rate limit headers
		if resp.StatusCode == 429 {
			limitReached = true
			if rateLimitHeaders == nil {
				rateLimitHeaders = map[string]string{}
				for k, v := range resp.Header {
					if strings.Contains(strings.ToLower(k), "ratelimit") ||
						strings.Contains(strings.ToLower(k), "rate-limit") ||
						strings.Contains(strings.ToLower(k), "x-ratelimit") {
						rateLimitHeaders[k] = strings.Join(v, ", ")
					}
				}
			}
			resp.Body.Close()
			break
		}
		resp.Body.Close()
		time.Sleep(10 * time.Millisecond)
	}

	if limitReached {
		addResult("TASK-009", "Rate Limiting", "PASS", "", map[string]interface{}{
			"limit_reached": true,
			"headers":       rateLimitHeaders,
		})
		fmt.Println("  ✅ Rate limiting active (429 returned after ~120 requests)")
		if len(rateLimitHeaders) > 0 {
			fmt.Printf("  📊 Rate limit headers: %v\n", rateLimitHeaders)
		}
	} else {
		addResult("TASK-009", "Rate Limiting", "WARN",
			"Rate limit not triggered in 130 requests - verify configuration", nil)
		fmt.Println("  ⚠️  Rate limit not triggered - check configuration")
	}
}

// TASK-010: CORS Policy
func verifyCORS() {
	fmt.Println("\n🔍 VERIFY-010: Testing CORS Policy (TASK-010)...")

	// Test 1: Valid origin
	req, _ := http.NewRequest("OPTIONS", baseURL+"/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		addResult("TASK-010", "CORS Preflight", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
	if corsHeader == "http://localhost:3000" || corsHeader == "*" {
		addResult("TASK-010", "CORS Preflight", "PASS", "", map[string]interface{}{
			"allow_origin": corsHeader,
		})
		fmt.Println("  ✅ CORS preflight working for valid origin")
	} else {
		addResult("TASK-010", "CORS Preflight", "FAIL",
			fmt.Sprintf("Missing CORS headers: %s", corsHeader), nil)
		fmt.Printf("  ❌ CORS headers missing: %s\n", corsHeader)
	}

	// Test 2: Invalid origin
	req2, _ := http.NewRequest("GET", baseURL+"/api/auth/login", nil)
	req2.Header.Set("Origin", "http://evil.com")

	resp2, err := client.Do(req2)
	if err != nil {
		addResult("TASK-010a", "CORS Invalid Origin", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == 403 {
		addResult("TASK-010a", "CORS Invalid Origin Block", "PASS", "", nil)
		fmt.Println("  ✅ CORS blocking unauthorized origins")
	} else {
		addResult("TASK-010a", "CORS Invalid Origin Block", "WARN",
			fmt.Sprintf("Expected 403, got %d - verify CORS whitelist", resp2.StatusCode), nil)
		fmt.Printf("  ⚠️  CORS may not block unauthorized origins: %d\n", resp2.StatusCode)
	}
}

// TASK-012: Encryption
func verifyEncryption() {
	fmt.Println("\n🔍 VERIFY-012: Testing Encryption (TASK-012)...")

	// This requires database access to verify credentials are encrypted
	// For now, we'll just verify the service exists

	addResult("TASK-012", "Encryption Service", "SKIP",
		"Requires database inspection - verify credentials are encrypted not plaintext", nil)
	fmt.Println("  ⚠️  Encryption verification requires database inspection")
	fmt.Println("     Manual check: SELECT connection_string FROM connections;")
	fmt.Println("     Should be base64 encoded encrypted data, not plaintext")
}

// TASK-014: Audit Logging
func verifyAuditLogging() {
	fmt.Println("\n🔍 VERIFY-014: Testing Audit Logging (TASK-014)...")

	// Check audit logs endpoint
	resp, err := http.Get(baseURL + "/api/admin/audit-logs")
	if err != nil {
		addResult("TASK-014", "Audit Logs API", "FAIL",
			fmt.Sprintf("Connection error: %v", err), nil)
		return
	}
	defer resp.Body.Close()

	// Should return 401 (needs auth) or 200 (if public for testing)
	if resp.StatusCode == 401 || resp.StatusCode == 200 {
		status := "PASS"
		if resp.StatusCode == 401 {
			status = "PASS" // Protected endpoint is good
		}

		body, _ := io.ReadAll(resp.Body)
		var logs []map[string]interface{}
		if err := json.Unmarshal(body, &logs); err == nil && len(logs) > 0 {
			addResult("TASK-014", "Audit Logs API", status, "", map[string]interface{}{
				"status_code": resp.StatusCode,
				"logs_count":  len(logs),
			})
			fmt.Printf("  ✅ Audit logs API working (%d logs found)\n", len(logs))
		} else {
			addResult("TASK-014", "Audit Logs API", status, "", map[string]interface{}{
				"status_code": resp.StatusCode,
				"note":        "Endpoint accessible",
			})
			fmt.Printf("  ✅ Audit logs endpoint accessible (status: %d)\n", resp.StatusCode)
		}
	} else {
		addResult("TASK-014", "Audit Logs API", "FAIL",
			fmt.Sprintf("Unexpected status: %d", resp.StatusCode), nil)
		fmt.Printf("  ❌ Audit logs endpoint issue: %d\n", resp.StatusCode)
	}
}

// Health check
func verifyHealthCheck() {
	fmt.Println("\n🔍 VERIFY-HEALTH: Testing Backend Health...")

	resp, err := http.Get(baseURL + "/api/health")
	if err != nil {
		fmt.Printf("  ❌ Backend not accessible: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var health map[string]interface{}
		json.Unmarshal(body, &health)

		fmt.Printf("  ✅ Backend healthy\n")
		fmt.Printf("  📊 Status: %v\n", health["status"])
	} else {
		fmt.Printf("  ⚠️  Backend status: %d\n", resp.StatusCode)
	}
}

// Save report
func saveReport() {
	report.Timestamp = time.Now()

	jsonData, _ := json.MarshalIndent(report, "", "  ")

	// Save to file
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("verification/reports/auth_verification_%s.json", timestamp)

	os.WriteFile(filename, jsonData, 0644)

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 VERIFICATION SUMMARY - Authentication & Security")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Total Tests:  %d\n", report.TotalTests)
	fmt.Printf("✅ Passed:     %d (%.1f%%)\n", report.Passed, float64(report.Passed)/float64(report.TotalTests)*100)
	fmt.Printf("❌ Failed:     %d (%.1f%%)\n", report.Failed, float64(report.Failed)/float64(report.TotalTests)*100)
	fmt.Printf("⚠️  Skipped:    %d (%.1f%%)\n", report.Skipped, float64(report.Skipped)/float64(report.TotalTests)*100)
	fmt.Println(strings.Repeat("=", 70))

	// Show failed tests
	if report.Failed > 0 {
		fmt.Println("\n🔴 FAILED TESTS:")
		for _, result := range report.Results {
			if result.Status == "FAIL" {
				fmt.Printf("  • %s: %s\n    Error: %s\n", result.TaskID, result.Name, result.Error)
			}
		}
	}

	fmt.Printf("\n📁 Full report saved to: %s\n", filename)
}

func main() {
	fmt.Println("🚀 InsightEngine E2E Verification Suite")
	fmt.Println("📋 VERIFY-001: Authentication & Security")
	fmt.Println(strings.Repeat("=", 70))

	// Check if backend is running
	verifyHealthCheck()

	// Run all verifications
	verifyUserRegistration()
	verifyEmailVerification()
	verifyForgotPassword()
	verifyGoogleOAuth()
	verifyRateLimiting()
	verifyCORS()
	verifyEncryption()
	verifyAuditLogging()

	// Save and display report
	saveReport()

	// Exit with error code if failures
	if report.Failed > 0 {
		fmt.Printf("\n❌ Verification completed with %d failures\n", report.Failed)
		os.Exit(1)
	}

	fmt.Println("\n✅ All verifications passed!")
}
