package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"insight-engine-backend/database"
	"insight-engine-backend/dtos"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/validator"
	"insight-engine-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler with dependencies
// Following Dependency Injection pattern for testability
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user account with validated input and sends a verification email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.RegisterRequest true "Registration Request"
// @Success 201 {object} dtos.RegisterResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Parse request body - Anti-Mass Assignment: Explicit field mapping
	var req dtos.RegisterRequest

	// Use manual JSON unmarshaling to avoid BodyParser issues in tests
	body := c.Body()
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid request body",
			})
		}
	} else {
		// If body is empty, return error as fields are required
		// But let validator handle it (though validator will fail if fields are empty)
	}

	// Validate input - Strict Mode
	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Validation failed",
			"errors":  err.Error(), // Validator returns formatted errors
		})
	}

	// Call service to create user and send verification email
	result, err := h.authService.Register(req.Email, req.Username, req.Password, req.FullName)
	if err != nil {
		// Handle specific business errors
		errMsg := err.Error()
		if errMsg == "email already registered" || errMsg == "username already taken" {
			return c.Status(409).JSON(fiber.Map{
				"status":  "error",
				"message": errMsg,
			})
		}

		// Log unexpected errors
		services.LogError("registration_failed", "Registration failed", map[string]interface{}{"error": err.Error()})
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Internal server error",
			"debug":   err.Error(), // Temporary for debugging
		})
	}

	// Prepare response message
	message := "Registration successful"
	if !result.VerificationSent {
		message = "Registration successful. Please contact support to verify your email."
	} else {
		message = "Registration successful. Please check your email to verify your account."
	}

	// Success response - HTTP 201 Created
	return c.Status(201).JSON(fiber.Map{
		"status": "success",
		"data": dtos.RegisterResponse{
			UserID:   result.User.ID.String(),
			Email:    result.User.Email,
			Username: result.User.Username,
			Message:  message,
		},
	})
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verifies user email using token from email link.
// @Tags Auth
// @Param token query string true "Verification Token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Verification token is required",
		})
	}

	user, err := h.authService.VerifyEmail(token)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "invalid or expired verification token" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired verification link. Please request a new one.",
			})
		}
		if errMsg == "email already verified" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Email already verified. Please sign in.",
			})
		}
		if errMsg == "verification token has expired" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Verification link has expired. Please request a new one.",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to verify email. Please try again.",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"message":  "Email verified successfully",
			"email":    user.Email,
			"verified": true,
		},
	})
}

// ResendVerification handles resending verification email
// @Summary Resend verification email
// @Description Resends verification email to user.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.ResendVerificationRequest true "Resend Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	var req dtos.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Email is already validated by Struct Validator

	err := h.authService.ResendVerificationEmail(req.Email)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "email already verified" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Email already verified. Please sign in.",
			})
		}

		// For security, don't reveal if email exists or not
		// Return generic success message
		return c.Status(200).JSON(fiber.Map{
			"status":  "success",
			"message": "If the email exists, a verification email has been sent.",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Verification email sent. Please check your inbox.",
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticates user and returns JWT token. Checks if email is verified.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.LoginRequest true "Login Request"
// @Success 200 {object} dtos.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check if email is verified
	fmt.Printf("DEBUG: Login checking email verification for %s. Verified: %v\n", user.Email, user.EmailVerified)
	if !user.EmailVerified {
		return c.Status(403).JSON(fiber.Map{
			"error":         "Email not verified",
			"message":       "Please verify your email before signing in.",
			"needsVerified": true,
		})
	}
	// Verify Password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		fmt.Printf("DEBUG: Login failed for %s. Error: %v. StoredHash: %s\n", req.Email, err, user.Password)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	fmt.Printf("DEBUG: Login user.ID from DB: '%s'\n", user.ID)

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["id"] = user.ID
	claims["email"] = user.Email
	claims["name"] = user.Name
	claims["role"] = user.Role
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // 3 days

	fmt.Printf("DEBUG: Login claims sub: '%v'\n", claims["sub"])

	// Sign token
	secret := os.Getenv("NEXTAUTH_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not login"})
	}

	return c.JSON(fiber.Map{
		"user":  user,
		"token": t,
	})
}

// ForgotPassword handles password reset request
// @Summary Request password reset
// @Description Generates reset token and sends reset email. Always returns success.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dtos.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Email validation handled by struct validator

	// Request password reset - always returns success for security
	err := h.authService.RequestPasswordReset(req.Email)
	if err != nil {
		// Log error but don't reveal to user (will use structured JSON logging in Phase 5)
	}

	// Always return success to prevent email enumeration
	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "If an account exists with this email, a password reset link has been sent.",
	})
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Resets user password using valid token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dtos.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Reset password
	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "invalid or expired reset token" || errMsg == "reset token has expired" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired reset token. Please request a new password reset.",
			})
		}
		if errMsg == "password must be at least 8 characters" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": errMsg,
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to reset password. Please try again.",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Password reset successfully. Please sign in with your new password.",
	})
}

// ValidateResetTokenRequest represents the validate reset token request
// Deprecated: Use dtos.ValidateResetTokenRequest if simple struct is needed, usually query param.
// But handler uses query param `token`. The struct was logically unused or for Swagger.
// Since handler logic uses `c.Query("token")`, we don't need BodyParser.
// I'll keep the handler logic as is (using Query), but remove the struct definition.

// ValidateResetToken checks if reset token is valid
// POST /api/auth/validate-reset-token
// Business Rule: Validates reset token without resetting password
func (h *AuthHandler) ValidateResetToken(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Token is required",
		})
	}

	isValid, err := h.authService.ValidateResetToken(token)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to validate token",
		})
	}

	if !isValid {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid or expired reset token",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"valid": true,
		},
	})
}

// ChangePassword handles password change for authenticated users
// @Summary Change password
// @Description Changes password after verifying current password. Requires authentication.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtos.ChangePasswordRequest true "Change Password Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	var req dtos.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validator.GetValidator().ValidateStruct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Change password
	err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "current password is incorrect" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Current password is incorrect",
			})
		}
		if errMsg == "new password must be different from current password" {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "New password must be different from current password",
			})
		}
		if errMsg == "user not found" {
			return c.Status(404).JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to change password. Please try again.",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Password changed successfully",
	})
}
