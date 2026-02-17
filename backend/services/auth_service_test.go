package services

import (
	"testing"

	"insight-engine-backend/database"
	"insight-engine-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupAuthServiceTestDB initializes in-memory SQLite DB for AuthService tests
func setupAuthServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err, "Failed to migrate test database")

	database.DB = db
	return db
}

// createMockEmailService creates a mock email service with sending disabled
func createMockEmailService() *EmailService {
	svc := NewEmailService()
	svc.Mock = true
	return svc
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: Register
// ─────────────────────────────────────────────────────────────────────────────

func TestAuthService_Register_Success(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	result, err := authSvc.Register("newuser@example.com", "newuser", "SecurePass123!", "New User")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.User)
	assert.Equal(t, "newuser@example.com", result.User.Email)
	assert.Equal(t, "newuser", result.User.Username)
	assert.Equal(t, "New User", result.User.Name)
	assert.NotEmpty(t, result.User.ID)

	// Verify password hash in DB
	var dbUser models.User
	err = db.First(&dbUser, "email = ?", "newuser@example.com").Error
	assert.NoError(t, err)
	assert.NotEqual(t, "SecurePass123!", dbUser.Password)
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte("SecurePass123!"))
	assert.NoError(t, err, "Password should be hashed with bcrypt")
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	// First registration
	_, err := authSvc.Register("dup@example.com", "user1", "SecurePass123!", "User One")
	require.NoError(t, err)

	// Second registration with same email
	_, err = authSvc.Register("dup@example.com", "user2", "SecurePass123!", "User Two")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already registered")
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	// First registration
	_, err := authSvc.Register("a@example.com", "sameuser", "SecurePass123!", "User A")
	require.NoError(t, err)

	// Second registration with same username
	_, err = authSvc.Register("b@example.com", "sameuser", "SecurePass123!", "User B")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username already taken")
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	testCases := []struct {
		email  string
		reason string
	}{
		{"", "empty email"},
		{"invalid", "no @ symbol"},
		{"@example.com", "no local part"},
		{"test@", "no domain"},
	}

	for _, tc := range testCases {
		t.Run(tc.reason, func(t *testing.T) {
			_, err := authSvc.Register(tc.email, "validuser", "SecurePass123!", "User")
			assert.Error(t, err, "Should fail for: %s", tc.reason)
		})
	}
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.Register("valid@example.com", "validuser", "short", "User")
	assert.Error(t, err)
}

func TestAuthService_Register_EmptyUsername(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.Register("valid@example.com", "", "SecurePass123!", "User")
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: GetUserByEmail / GetUserByUsername
// ─────────────────────────────────────────────────────────────────────────────

func TestAuthService_GetUserByEmail_Found(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())
	_, err := authSvc.Register("findme@example.com", "findme", "SecurePass123!", "Find Me")
	require.NoError(t, err)

	user, err := authSvc.GetUserByEmail("findme@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "findme@example.com", user.Email)
}

func TestAuthService_GetUserByEmail_NotFound(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.GetUserByEmail("nonexistent@example.com")
	assert.Error(t, err)
}

func TestAuthService_GetUserByUsername_Found(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())
	_, err := authSvc.Register("user@example.com", "findusername", "SecurePass123!", "User")
	require.NoError(t, err)

	user, err := authSvc.GetUserByUsername("findusername")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "findusername", user.Username)
}

func TestAuthService_GetUserByUsername_NotFound(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.GetUserByUsername("ghostuser")
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: VerifyPassword
// ─────────────────────────────────────────────────────────────────────────────

func TestAuthService_VerifyPassword_Correct(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	// Generate a hash
	hash, err := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)
	require.NoError(t, err)

	err = authSvc.VerifyPassword(string(hash), "TestPassword123!")
	assert.NoError(t, err)
}

func TestAuthService_VerifyPassword_Incorrect(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	hash, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	err = authSvc.VerifyPassword(string(hash), "WrongPassword")
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: ChangePassword
// ─────────────────────────────────────────────────────────────────────────────

func TestAuthService_ChangePassword_Success(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	// Register user first
	result, err := authSvc.Register("change@example.com", "changeuser", "OldPass123!", "Change User")
	require.NoError(t, err)

	// Change password
	err = authSvc.ChangePassword(result.User.ID, "OldPass123!", "NewPass456!")
	assert.NoError(t, err)

	// Verify old password no longer works
	var updated models.User
	db.First(&updated, "id = ?", result.User.ID)
	err = bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("OldPass123!"))
	assert.Error(t, err, "Old password should no longer work")

	// Verify new password works
	err = bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("NewPass456!"))
	assert.NoError(t, err, "New password should work")
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	result, err := authSvc.Register("wrong@example.com", "wronguser", "CurrentPass123!", "Wrong User")
	require.NoError(t, err)

	err = authSvc.ChangePassword(result.User.ID, "WrongCurrentPass!", "NewPass456!")
	assert.Error(t, err)
}

func TestAuthService_ChangePassword_ShortNewPassword(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	result, err := authSvc.Register("short@example.com", "shortuser", "CurrentPass123!", "Short User")
	require.NoError(t, err)

	err = authSvc.ChangePassword(result.User.ID, "CurrentPass123!", "short")
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: RequestPasswordReset / ResetPassword / ValidateResetToken
// ─────────────────────────────────────────────────────────────────────────────

func TestAuthService_RequestPasswordReset_ExistingUser(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.Register("reset@example.com", "resetuser", "Pass123!", "Reset User")
	require.NoError(t, err)

	// Should NOT return error (prevents email enumeration)
	err = authSvc.RequestPasswordReset("reset@example.com")
	assert.NoError(t, err)

	// Verify token was generated
	var user models.User
	db.First(&user, "email = ?", "reset@example.com")
	assert.NotNil(t, user.PasswordResetToken)
	assert.NotEmpty(t, user.PasswordResetToken)
}

func TestAuthService_RequestPasswordReset_NonExistentUser(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	// Should NOT return error (prevents email enumeration)
	err := authSvc.RequestPasswordReset("ghost@example.com")
	assert.NoError(t, err, "Should not reveal whether email exists")
}

func TestAuthService_ResetPassword_ValidToken(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.Register("valid-reset@example.com", "validreset", "OldPass123!", "Valid Reset")
	require.NoError(t, err)

	err = authSvc.RequestPasswordReset("valid-reset@example.com")
	require.NoError(t, err)

	// Get the token
	var user models.User
	db.First(&user, "email = ?", "valid-reset@example.com")
	require.NotEmpty(t, user.PasswordResetToken)

	// Reset password
	err = authSvc.ResetPassword(user.PasswordResetToken, "BrandNew123!")
	assert.NoError(t, err)

	// Verify new password works
	var updated models.User
	db.First(&updated, "email = ?", "valid-reset@example.com")
	err = bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("BrandNew123!"))
	assert.NoError(t, err)
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	err := authSvc.ResetPassword("invalid-token-12345", "NewPass123!")
	assert.Error(t, err)
}

func TestAuthService_ValidateResetToken_Valid(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	defer db.Exec("DELETE FROM users")

	authSvc := NewAuthService(db, createMockEmailService())

	_, err := authSvc.Register("validate@example.com", "validateuser", "Pass123!", "Validate User")
	require.NoError(t, err)

	err = authSvc.RequestPasswordReset("validate@example.com")
	require.NoError(t, err)

	var user models.User
	db.First(&user, "email = ?", "validate@example.com")

	valid, err := authSvc.ValidateResetToken(user.PasswordResetToken)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestAuthService_ValidateResetToken_Invalid(t *testing.T) {
	db := setupAuthServiceTestDB(t)
	authSvc := NewAuthService(db, createMockEmailService())

	valid, err := authSvc.ValidateResetToken("bogus-token")
	if err == nil {
		assert.False(t, valid)
	}
}
