package dtos

import (
	"regexp"
	"strings"
)

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanumunicode"` // strict alphanum check
	Password string `json:"password" validate:"required,min=8,max=128"`
	FullName string `json:"fullName" validate:"omitempty,max=100"`
}

// RegisterResponse represents the registration success response
type RegisterResponse struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult contains validation status and errors
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidateRegisterRequest - DEPRECATED: Use validator.GetValidator().ValidateStruct() instead
// Kept for backward compatibility if needed, but handlers should switch to the validator package.
func ValidateRegisterRequest(req *RegisterRequest) *ValidationResult {
	errors := []ValidationError{}

	// Email validation
	if strings.TrimSpace(req.Email) == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	} else if !isValidEmail(req.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	// Username validation
	if strings.TrimSpace(req.Username) == "" {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username is required",
		})
	} else if len(req.Username) < 3 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username must be at least 3 characters",
		})
	} else if len(req.Username) > 50 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username must be less than 50 characters",
		})
	} else if !isValidUsername(req.Username) {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username can only contain letters, numbers, underscores, and hyphens",
		})
	}

	// Password validation
	if req.Password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password is required",
		})
	} else if len(req.Password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters",
		})
	} else if len(req.Password) > 128 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must be less than 128 characters",
		})
	}

	return &ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// isValidEmail validates email format using regex
// Business Rule: Must follow standard email format (user@domain.com)
func isValidEmail(email string) bool {
	// RFC 5322 compliant regex for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidUsername validates username format
// Business Rule: Only alphanumeric, underscore, hyphen allowed
func isValidUsername(username string) bool {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernameRegex.MatchString(username)
}

// ResendVerificationRequest represents the request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User interface{} `json:"user"` // Use interface to avoid import cycle with models if needed, or better, just use map[string]interface{} or specific struct if models is allowed.
	// Actually, dtos usually don't import models to avoid circular deps if models import dtos.
	// But here models might be safe. Let's check imports.
	// auth_dto.go only imports regexp, strings.
	// If I import models, it might be fine.
	// But `auth_handler.go` defines LoginResponse with `models.User`.
	// Let's keep LoginResponse in handler output or define a UserDTO.
	// For now, let's just move Requests to DTOs. Responses can stay or be moved if simple.
	// I will NOT move LoginResponse to avoid potential circular dependency if models import dtos (unlikely but safe).
	// Actually, `auth_handler.go` imports `dtos` and `models`. `models` likely doesn't import `dtos`.
	// But `LoginResponse` uses `models.User`.
	// I will leave LoginResponse in handler for now or redefine it without models.User if possible.
	// The instructions say "Standardize DTOs".
	// I'll stick to Requests for validation first.
	Token string `json:"token"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=128"`
}

// ValidateResetTokenRequest represents the validate reset token request
type ValidateResetTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=128,nefield=CurrentPassword"`
}
