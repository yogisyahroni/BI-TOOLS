package services

import (
	"context"
	"testing"
	"time"

	"insight-engine-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupUserAdminTestDB initializes in-memory SQLite DB for testing
func setupUserAdminTestDB(t *testing.T) (*gorm.DB, *UserAdminService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect to test database")
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	// Manually create users table to avoid SQLite error with uuid_generate_v4()
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		username TEXT UNIQUE,
		name TEXT,
		password TEXT,
		role TEXT DEFAULT 'user',
		email_verified NUMERIC DEFAULT 0,
		email_verified_at TIMESTAMP,
		email_verification_token TEXT,
		email_verification_expires TIMESTAMP,
		password_reset_token TEXT,
		password_reset_expires TIMESTAMP,
		provider TEXT,
		provider_id TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		status TEXT DEFAULT 'active',
		deactivated_at TIMESTAMP,
		deactivated_by TEXT,
		deactivation_reason TEXT,
		impersonation_token TEXT,
		impersonation_expires TIMESTAMP,
		impersonated_by TEXT
	)`)

	// Migrate schemas (skip User as it is manually created)
	db.AutoMigrate(&models.AuditLog{})

	// Initialize dependencies
	auditService := NewAuditService(db)
	return db, NewUserAdminService(db, auditService)
}

func cleanupUserAdminTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM audit_logs")
}

func TestGetUsers(t *testing.T) {
	db, service := setupUserAdminTestDB(t)
	defer cleanupUserAdminTestDB(db)

	// Create test users
	users := []models.User{
		{
			ID:       uuid.New(),
			Name:     "Admin User",
			Email:    "admin@example.com",
			Username: "admin",
			Role:     "admin",
			Status:   models.UserStatusActive,
		},
		{
			ID:       uuid.New(),
			Name:     "Regular User",
			Email:    "user@example.com",
			Username: "user",
			Role:     "user",
			Status:   models.UserStatusActive,
		},
		{
			ID:       uuid.New(),
			Name:     "Inactive User",
			Email:    "inactive@example.com",
			Username: "inactive",
			Role:     "user",
			Status:   models.UserStatusInactive,
		},
	}
	err := db.Create(&users).Error
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("GetAllUsers", func(t *testing.T) {
		filter := &UserFilter{Limit: 10}
		resp, err := service.GetUsers(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(3), resp.Total)
		assert.Len(t, resp.Users, 3)
	})

	t.Run("FilterByRole", func(t *testing.T) {
		filter := &UserFilter{Role: "admin", Limit: 10}
		resp, err := service.GetUsers(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, "admin", resp.Users[0].Username)
	})

	t.Run("FilterByStatus", func(t *testing.T) {
		filter := &UserFilter{Status: string(models.UserStatusInactive), Limit: 10}
		resp, err := service.GetUsers(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, "inactive", resp.Users[0].Username)
	})

	t.Run("Search", func(t *testing.T) {
		filter := &UserFilter{Search: "Regular", Limit: 10}
		resp, err := service.GetUsers(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, "user", resp.Users[0].Username)
	})
}

func TestUpdateUser(t *testing.T) {
	db, service := setupUserAdminTestDB(t)
	defer cleanupUserAdminTestDB(db)

	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Name:     "Original Name",
		Email:    "original@example.com",
		Username: "original",
		Role:     "user",
	}
	db.Create(&user)

	ctx := context.Background()

	t.Run("UpdateNameAndRole", func(t *testing.T) {
		req := &UpdateUserRequest{
			Name: "Updated Name",
			Role: "admin",
		}
		updatedUser, err := service.UpdateUser(ctx, userID.String(), req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, "admin", updatedUser.Role)

		// Verify in DB
		var dbUser models.User
		db.First(&dbUser, "id = ?", userID)
		assert.Equal(t, "Updated Name", dbUser.Name)
		assert.Equal(t, "admin", dbUser.Role)
	})

	t.Run("UpdateEmailDuplicate", func(t *testing.T) {
		// Create another user
		otherUser := models.User{
			ID:       uuid.New(),
			Name:     "Other",
			Email:    "other@example.com",
			Username: "other",
		}
		db.Create(&otherUser)

		req := &UpdateUserRequest{
			Email: "other@example.com", // Duplicate
		}
		_, err := service.UpdateUser(ctx, userID.String(), req)
		assert.Error(t, err)
		// This line is not part of the original code, but was in the provided snippet.
		// Assuming it's a placeholder or an example of a query that might use this pattern.
		// query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?)", search, search, search)
		assert.Equal(t, "email already in use", err.Error())
	})
}

func TestActivateDeactivateUser(t *testing.T) {
	db, service := setupUserAdminTestDB(t)
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Name:     "Test User",
		Email:    "test@example.com",
		Username: "test",
		Status:   models.UserStatusActive,
	}
	db.Create(&user)

	ctx := context.Background()

	t.Run("DeactivateUser", func(t *testing.T) {
		deactivatedUser, err := service.DeactivateUser(ctx, userID.String(), adminID, "Violation of terms")
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusInactive, deactivatedUser.Status)
		assert.Equal(t, adminID, *deactivatedUser.DeactivatedBy)
		assert.Equal(t, "Violation of terms", *deactivatedUser.DeactivationReason)
		assert.NotNil(t, deactivatedUser.DeactivatedAt)
	})

	t.Run("ActivateUser", func(t *testing.T) {
		activatedUser, err := service.ActivateUser(ctx, userID.String(), adminID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusActive, activatedUser.Status)
		assert.Nil(t, activatedUser.DeactivatedBy)
		assert.Nil(t, activatedUser.DeactivationReason)
		assert.Nil(t, activatedUser.DeactivatedAt)
	})
}

func TestDeleteUser(t *testing.T) {
	db, service := setupUserAdminTestDB(t)
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Name:     "Delete Me",
		Email:    "delete@example.com",
		Username: "deleteme",
	}
	db.Create(&user)

	ctx := context.Background()

	err := service.DeleteUser(ctx, userID.String(), adminID)
	require.NoError(t, err)

	// Verify deletion
	var count int64
	db.Model(&models.User{}).Where("id = ?", userID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestImpersonation(t *testing.T) {
	db, service := setupUserAdminTestDB(t)
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Name:     "Target User",
		Email:    "target@example.com",
		Username: "target",
		Status:   models.UserStatusActive,
	}
	db.Create(&user)

	ctx := context.Background()

	var token string

	t.Run("ImpersonateUser", func(t *testing.T) {
		resp, err := service.ImpersonateUser(ctx, userID.String(), adminID)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, userID, resp.UserID)
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), resp.ExpiresAt, 1*time.Minute)
		token = resp.Token
	})

	t.Run("ValidateImpersonationToken", func(t *testing.T) {
		impersonatedUser, err := service.ValidateImpersonationToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, userID, impersonatedUser.ID)

		// Token should be cleared after use
		var dbUser models.User
		db.First(&dbUser, "id = ?", userID)
		assert.Empty(t, dbUser.ImpersonationToken)
	})
}
