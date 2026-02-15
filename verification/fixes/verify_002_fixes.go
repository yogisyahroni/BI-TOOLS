// verification/fixes/verify_002_fixes.go
// VERIFY-002: Authentication Flow Fixes
// Fixes common issues found during verification

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestConfiguration holds test configuration
type TestConfiguration struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	BackendURL string
}

func loadConfig() *TestConfiguration {
	return &TestConfiguration{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "insightengine"),
		BackendURL: getEnv("BACKEND_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type FixResult struct {
	Issue   string `json:"issue"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

var results []FixResult

func addResult(issue, status, details string) {
	results = append(results, FixResult{
		Issue:   issue,
		Status:  status,
		Details: details,
	})
}

// FIX 1: Check and fix email verification token expiration
func fixEmailVerificationTokens(db *gorm.DB) {
	fmt.Println("🔧 FIX 1: Checking email verification tokens...")

	// Check if users have expired tokens
	var expiredCount int64
	twoDaysAgo := time.Now().Add(-48 * time.Hour)

	db.Raw(`
		SELECT COUNT(*) FROM users 
		WHERE email_verification_token != '' 
		AND email_verified_at IS NULL 
		AND created_at < ?
	`, twoDaysAgo).Scan(&expiredCount)

	if expiredCount > 0 {
		fmt.Printf("  Found %d users with expired verification tokens\n", expiredCount)

		// Clear expired tokens
		result := db.Exec(`
			UPDATE users 
			SET email_verification_token = ''
			WHERE email_verification_token != '' 
			AND email_verified_at IS NULL 
			AND created_at < ?
		`, twoDaysAgo)

		if result.Error != nil {
			addResult("Email Verification Tokens", "FAIL", result.Error.Error())
			fmt.Printf("  ❌ Failed to clear expired tokens: %v\n", result.Error)
		} else {
			addResult("Email Verification Tokens", "FIXED",
				fmt.Sprintf("Cleared %d expired tokens", expiredCount))
			fmt.Printf("  ✅ Cleared %d expired tokens\n", expiredCount)
		}
	} else {
		addResult("Email Verification Tokens", "OK", "No expired tokens found")
		fmt.Println("  ✅ No expired tokens found")
	}
}

// FIX 2: Verify encryption is working
func verifyEncryption(db *gorm.DB) {
	fmt.Println("\n🔧 FIX 2: Verifying credential encryption...")

	// Check if any credentials are stored in plaintext
	var plaintextCount int64
	db.Raw(`
		SELECT COUNT(*) FROM connections 
		WHERE connection_string NOT LIKE '%=%' 
		AND connection_string NOT LIKE '{%}'
		AND connection_string != ''
	`).Scan(&plaintextCount)

	if plaintextCount > 0 {
		addResult("Credential Encryption", "WARNING",
			fmt.Sprintf("Found %d potentially unencrypted credentials", plaintextCount))
		fmt.Printf("  ⚠️  Found %d potentially plaintext credentials\n", plaintextCount)
		fmt.Println("     These should be encrypted - run encryption migration")
	} else {
		addResult("Credential Encryption", "OK", "Credentials appear encrypted")
		fmt.Println("  ✅ Credentials appear to be encrypted")
	}

	// Check for ENCRYPTION_KEY
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		addResult("Encryption Key", "MISSING", "ENCRYPTION_KEY not set")
		fmt.Println("  ❌ ENCRYPTION_KEY environment variable not set!")
		fmt.Println("     Generate with: openssl rand -base64 32")
	} else {
		keyLen := len(encKey)
		if keyLen == 32 || keyLen == 44 { // 32 bytes or base64 encoded
			addResult("Encryption Key", "OK", "Key is properly configured")
			fmt.Println("  ✅ Encryption key is configured")
		} else {
			addResult("Encryption Key", "INVALID", fmt.Sprintf("Key length is %d, expected 32", keyLen))
			fmt.Printf("  ❌ Encryption key length is %d, expected 32 bytes\n", keyLen)
		}
	}
}

// FIX 3: Check audit logging configuration
func verifyAuditLogging(db *gorm.DB) {
	fmt.Println("\n🔧 FIX 3: Verifying audit logging...")

	// Check if audit_logs table exists and has entries
	var tableExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'audit_logs'
		)
	`).Scan(&tableExists)

	if !tableExists {
		addResult("Audit Logging Table", "MISSING", "audit_logs table does not exist")
		fmt.Println("  ❌ audit_logs table does not exist!")
		fmt.Println("     Run migrations to create audit_logs table")
		return
	}

	// Check recent audit logs
	var recentCount int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	db.Raw(`
		SELECT COUNT(*) FROM audit_logs 
		WHERE created_at > ?
	`, oneHourAgo).Scan(&recentCount)

	if recentCount == 0 {
		addResult("Audit Logging", "WARNING", "No recent audit logs found")
		fmt.Println("  ⚠️  No audit logs in the last hour")
		fmt.Println("     Verify audit logging is properly integrated")
	} else {
		addResult("Audit Logging", "OK", fmt.Sprintf("Found %d recent audit logs", recentCount))
		fmt.Printf("  ✅ Found %d audit logs in the last hour\n", recentCount)
	}
}

// FIX 4: Check OAuth configuration
func verifyOAuthConfig() {
	fmt.Println("\n🔧 FIX 4: Verifying OAuth configuration...")

	providers := []string{
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
		"OKTA_CLIENT_ID",
		"OKTA_CLIENT_SECRET",
	}

	configured := 0
	missing := []string{}

	for _, provider := range providers {
		if os.Getenv(provider) != "" {
			configured++
		} else {
			missing = append(missing, provider)
		}
	}

	if configured == 0 {
		addResult("OAuth Configuration", "WARNING", "No OAuth providers configured")
		fmt.Println("  ⚠️  No OAuth providers configured")
		fmt.Println("     Set at least GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET for Google OAuth")
	} else {
		addResult("OAuth Configuration", "OK", fmt.Sprintf("%d providers configured", configured/2))
		fmt.Printf("  ✅ %d OAuth provider(s) configured\n", configured/2)

		if len(missing) > 0 {
			fmt.Printf("     Missing (optional): %s\n", strings.Join(missing, ", "))
		}
	}
}

// FIX 5: Check rate limiting configuration
func verifyRateLimiting() {
	fmt.Println("\n🔧 FIX 5: Verifying rate limiting...")

	// Check Redis connection
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		addResult("Rate Limiting (Redis)", "WARNING", "REDIS_HOST not set, using in-memory")
		fmt.Println("  ⚠️  REDIS_HOST not set, rate limiting will use in-memory store")
		fmt.Println("     For production, set REDIS_HOST for distributed rate limiting")
	} else {
		addResult("Rate Limiting (Redis)", "OK", "Redis configured for rate limiting")
		fmt.Println("  ✅ Redis configured for distributed rate limiting")
	}

	// Check rate limit values
	rateLimit := os.Getenv("RATE_LIMIT_REQUESTS")
	if rateLimit == "" {
		addResult("Rate Limit Value", "DEFAULT", "Using default 120 requests/minute")
		fmt.Println("  ℹ️  Using default rate limit: 120 requests/minute")
	} else {
		addResult("Rate Limit Value", "OK", fmt.Sprintf("Custom rate limit: %s", rateLimit))
		fmt.Printf("  ✅ Custom rate limit configured: %s\n", rateLimit)
	}
}

// FIX 6: Check CORS configuration
func verifyCORS() {
	fmt.Println("\n🔧 FIX 6: Verifying CORS configuration...")

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		addResult("CORS Configuration", "WARNING", "ALLOWED_ORIGINS not set, using default")
		fmt.Println("  ⚠️  ALLOWED_ORIGINS not set")
		fmt.Println("     CORS will allow localhost:3000 (development mode)")
		fmt.Println("     For production, set ALLOWED_ORIGINS=https://yourdomain.com")
	} else {
		addResult("CORS Configuration", "OK", fmt.Sprintf("Allowed origins: %s", allowedOrigins))
		fmt.Printf("  ✅ CORS configured for: %s\n", allowedOrigins)
	}
}

// FIX 7: Check JWT configuration
func verifyJWTConfig() {
	fmt.Println("\n🔧 FIX 7: Verifying JWT configuration...")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		addResult("JWT Secret", "MISSING", "JWT_SECRET not set")
		fmt.Println("  ❌ JWT_SECRET environment variable not set!")
		fmt.Println("     Generate with: openssl rand -base64 64")
	} else {
		if len(jwtSecret) < 32 {
			addResult("JWT Secret", "WEAK", "JWT secret is too short")
			fmt.Printf("  ⚠️  JWT secret is only %d characters, should be at least 32\n", len(jwtSecret))
		} else {
			addResult("JWT Secret", "OK", "JWT secret is properly configured")
			fmt.Println("  ✅ JWT secret is configured")
		}
	}

	jwtExpiry := os.Getenv("JWT_EXPIRY")
	if jwtExpiry == "" {
		addResult("JWT Expiry", "DEFAULT", "Using default 24 hours")
		fmt.Println("  ℹ️  Using default JWT expiry: 24 hours")
	} else {
		addResult("JWT Expiry", "OK", fmt.Sprintf("Custom expiry: %s", jwtExpiry))
		fmt.Printf("  ✅ JWT expiry configured: %s\n", jwtExpiry)
	}
}

// FIX 8: Test database connectivity
func testDatabaseConnection(config *TestConfiguration) (*gorm.DB, error) {
	fmt.Println("\n🔧 FIX 8: Testing database connectivity...")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser,
		config.DBPassword, config.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		addResult("Database Connection", "FAIL", err.Error())
		fmt.Printf("  ❌ Cannot connect to database: %v\n", err)
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		addResult("Database Connection", "FAIL", err.Error())
		fmt.Printf("  ❌ Database error: %v\n", err)
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		addResult("Database Connection", "FAIL", err.Error())
		fmt.Printf("  ❌ Cannot ping database: %v\n", err)
		return nil, err
	}

	addResult("Database Connection", "OK", "Connected successfully")
	fmt.Println("  ✅ Database connection successful")

	return db, nil
}

// Generate fixes report
func generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 VERIFY-002 FIXES REPORT")
	fmt.Println(strings.Repeat("=", 70))

	fixed := 0
	warnings := 0
	failed := 0
	ok := 0

	for _, result := range results {
		switch result.Status {
		case "FIXED":
			fixed++
			fmt.Printf("✅ FIXED: %s\n   %s\n\n", result.Issue, result.Details)
		case "OK":
			ok++
			fmt.Printf("✅ OK: %s\n   %s\n\n", result.Issue, result.Details)
		case "WARNING":
			warnings++
			fmt.Printf("⚠️  WARNING: %s\n   %s\n\n", result.Issue, result.Details)
		case "FAIL", "MISSING", "INVALID", "WEAK":
			failed++
			fmt.Printf("❌ %s: %s\n   %s\n\n", result.Status, result.Issue, result.Details)
		}
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Summary: %d Fixed, %d OK, %d Warnings, %d Failed\n", fixed, ok, warnings, failed)
	fmt.Println(strings.Repeat("=", 70))

	if failed > 0 {
		fmt.Println("\n❌ VERIFY-002 completed with failures")
		fmt.Println("Please address the FAILED items above before proceeding")
	} else if warnings > 0 {
		fmt.Println("\n⚠️  VERIFY-002 completed with warnings")
		fmt.Println("Review WARNINGS but can proceed with caution")
	} else {
		fmt.Println("\n✅ VERIFY-002 completed successfully!")
		fmt.Println("All authentication fixes applied")
	}
}

func main() {
	fmt.Println("🚀 VERIFY-002: Authentication Flow Fixes")
	fmt.Println("🔧 Applying fixes for common authentication issues")
	fmt.Println(strings.Repeat("=", 70))

	// Load configuration
	config := loadConfig()

	// Test database connection
	db, err := testDatabaseConnection(config)
	if err != nil {
		fmt.Println("\n❌ Cannot proceed without database connection")
		os.Exit(1)
	}

	// Apply fixes
	fixEmailVerificationTokens(db)
	verifyEncryption(db)
	verifyAuditLogging(db)
	verifyOAuthConfig()
	verifyRateLimiting()
	verifyCORS()
	verifyJWTConfig()

	// Generate report
	generateReport()

	// Exit with error code if there are failures
	for _, result := range results {
		if result.Status == "FAIL" || result.Status == "MISSING" || result.Status == "INVALID" {
			os.Exit(1)
		}
	}
}
