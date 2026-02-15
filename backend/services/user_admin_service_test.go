package services

import (
	"context"
	"testing"
	"time"

	"insight-engine-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupUserAdminTestDB initializes in-memory SQLite DB for testing
func setupUserAdminTestDB() (*gorm.DB, *UserAdminService) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	// Migrate schemas
	db.AutoMigrate(&models.User{}, &models.AuditLog{})

	// Initialize dependencies
	auditService := NewAuditService(db)
	userAdminService := NewUserAdminService(db, auditService)

	return db, userAdminService
}

func cleanupUserAdminTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM audit_logs")
}

func TestGetUsers(t *testing.T) {
	db, service := setupUserAdminTestDB()
	defer cleanupUserAdminTestDB(db)

	// Create test users
	users := []models.User{
		{
			ID:       "user-1",
			Name:     "Admin User",
			Email:    "admin@example.com",
			Username: "admin",
			Role:     "admin",
			Status:   models.UserStatusActive,
		},
		{
			ID:       "user-2",
			Name:     "Regular User",
			Email:    "user@example.com",
			Username: "user",
			Role:     "user",
			Status:   models.UserStatusActive,
		},
		{
			ID:       "user-3",
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
	db, service := setupUserAdminTestDB()
	defer cleanupUserAdminTestDB(db)

	user := models.User{
		ID:       "user-update",
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
		updatedUser, err := service.UpdateUser(ctx, user.ID, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, "admin", updatedUser.Role)

		// Verify in DB
		var dbUser models.User
		db.First(&dbUser, "id = ?", user.ID)
		assert.Equal(t, "Updated Name", dbUser.Name)
		assert.Equal(t, "admin", dbUser.Role)
	})

	t.Run("UpdateEmailDuplicate", func(t *testing.T) {
		// Create another user
		otherUser := models.User{
			ID:       "user-other",
			Name:     "Other",
			Email:    "other@example.com",
			Username: "other",
		}
		db.Create(&otherUser)

		req := &UpdateUserRequest{
			Email: "other@example.com", // Duplicate
		}
		_, err := service.UpdateUser(ctx, user.ID, req)
		assert.Error(t, err)
		// This line is not part of the original code, but was in the provided snippet.
		// Assuming it's a placeholder or an example of a query that might use this pattern.
		// query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?)", search, search, search)
		assert.Equal(t, "email already in use", err.Error())
	})
}

func TestActivateDeactivateUser(t *testing.T) {
	db, service := setupUserAdminTestDB()
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	user := models.User{
		ID:       "user-active",
		Name:     "Test User",
		Email:    "test@example.com",
		Username: "test",
		Status:   models.UserStatusActive,
	}
	db.Create(&user)

	ctx := context.Background()

	t.Run("DeactivateUser", func(t *testing.T) {
		deactivatedUser, err := service.DeactivateUser(ctx, user.ID, adminID, "Violation of terms")
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusInactive, deactivatedUser.Status)
		assert.Equal(t, adminID, *deactivatedUser.DeactivatedBy)
		assert.Equal(t, "Violation of terms", *deactivatedUser.DeactivationReason)
		assert.NotNil(t, deactivatedUser.DeactivatedAt)
	})

	t.Run("ActivateUser", func(t *testing.T) {
		activatedUser, err := service.ActivateUser(ctx, user.ID, adminID)
		require.NoError(t, err)
		assert.Equal(t, models.UserStatusActive, activatedUser.Status)
		assert.Nil(t, activatedUser.DeactivatedBy)
		assert.Nil(t, activatedUser.DeactivationReason)
		assert.Nil(t, activatedUser.DeactivatedAt)
	})
}

func TestDeleteUser(t *testing.T) {
	db, service := setupUserAdminTestDB()
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	user := models.User{
		ID:       "user-delete",
		Name:     "Delete Me",
		Email:    "delete@example.com",
		Username: "deleteme",
	}
	db.Create(&user)

	ctx := context.Background()

	err := service.DeleteUser(ctx, user.ID, adminID)
	require.NoError(t, err)

	// Verify deletion
	var count int64
	db.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestImpersonation(t *testing.T) {
	db, service := setupUserAdminTestDB()
	defer cleanupUserAdminTestDB(db)

	adminID := "admin-123"
	user := models.User{
		ID:       "user-impersonate",
		Name:     "Target User",
		Email:    "target@example.com",
		Username: "target",
		Status:   models.UserStatusActive,
	}
	db.Create(&user)

	ctx := context.Background()

	var token string

	t.Run("ImpersonateUser", func(t *testing.T) {
		resp, err := service.ImpersonateUser(ctx, user.ID, adminID)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, user.ID, resp.UserID)
		assert.WithinDuration(t, time.Now().Add(15*time.Minute), resp.ExpiresAt, 1*time.Minute)
		token = resp.Token
	})

	t.Run("ValidateImpersonationToken", func(t *testing.T) {
		impersonatedUser, err := service.ValidateImpersonationToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, user.ID, impersonatedUser.ID)

		// Token should be cleared after use
		var dbUser models.User
		db.First(&dbUser, "id = ?", user.ID)
		assert.Empty(t, dbUser.ImpersonationToken)
	})
}
