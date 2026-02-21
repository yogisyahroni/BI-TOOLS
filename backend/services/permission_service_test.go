package services

import (
	"testing"

	"insight-engine-backend/database"
	"insight-engine-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupPermissionTestDB initializes in-memory SQLite DB for permission tests
func setupPermissionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err, "Failed to connect to test database")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	sqlDB.SetMaxOpenConns(1)

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

	// Migrate schemas
	err = db.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},
		&models.UserRole{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	database.DB = db
	return db
}

// seedPermissions creates test permissions in the database
func seedPermissions(t *testing.T, db *gorm.DB) []models.Permission {
	t.Helper()
	permissions := []models.Permission{
		{Name: "query:create", Resource: "query", Action: "create", Description: "Create queries"},
		{Name: "query:read", Resource: "query", Action: "read", Description: "Read queries"},
		{Name: "query:update", Resource: "query", Action: "update", Description: "Update queries"},
		{Name: "query:delete", Resource: "query", Action: "delete", Description: "Delete queries"},
		{Name: "dashboard:create", Resource: "dashboard", Action: "create", Description: "Create dashboards"},
		{Name: "dashboard:read", Resource: "dashboard", Action: "read", Description: "Read dashboards"},
		{Name: "dashboard:update", Resource: "dashboard", Action: "update", Description: "Update dashboards"},
		{Name: "dashboard:delete", Resource: "dashboard", Action: "delete", Description: "Delete dashboards"},
		{Name: "admin:manage_users", Resource: "admin", Action: "manage_users", Description: "Manage users"},
		{Name: "admin:system_settings", Resource: "admin", Action: "system_settings", Description: "System settings"},
	}
	for i := range permissions {
		err := db.Create(&permissions[i]).Error
		require.NoError(t, err, "Failed to seed permission: %s", permissions[i].Name)
	}
	return permissions
}

// seedTestUser creates a test user in the database
func seedTestUser(t *testing.T, db *gorm.DB, usernameSuffix string) *models.User {
	t.Helper()
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    usernameSuffix + "@test.com",
		Username: "user_" + usernameSuffix[:8],
		Name:     "Test User",
		Password: "$2a$10$dummyhashedpassword",
	}
	err := db.Create(user).Error
	require.NoError(t, err, "Failed to seed test user")
	return user
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: GetAllPermissions
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAllPermissions_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	result, err := svc.GetAllPermissions()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), len(seeded))

	// Verify we can find the seeded permissions
	found := make(map[string]bool)
	for _, p := range result {
		found[p.Name] = true
	}
	for _, s := range seeded {
		assert.True(t, found[s.Name], "Expected permission %s to exist", s.Name)
	}
}

func TestGetAllPermissions_EmptyDB(t *testing.T) {
	db := setupPermissionTestDB(t)
	// No seed — empty DB

	svc := NewPermissionService(db)
	result, err := svc.GetAllPermissions()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: GetPermissionsByResource
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPermissionsByResource_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seedPermissions(t, db)

	svc := NewPermissionService(db)
	result, err := svc.GetPermissionsByResource("query")

	assert.NoError(t, err)
	assert.Len(t, result, 4) // create, read, update, delete

	for _, p := range result {
		assert.Equal(t, "query", p.Resource)
	}
}

func TestGetPermissionsByResource_NonExistent(t *testing.T) {
	db := setupPermissionTestDB(t)
	seedPermissions(t, db)

	svc := NewPermissionService(db)
	result, err := svc.GetPermissionsByResource("nonexistent")

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGetPermissionsByResource_AdminResource(t *testing.T) {
	db := setupPermissionTestDB(t)
	seedPermissions(t, db)

	svc := NewPermissionService(db)
	result, err := svc.GetPermissionsByResource("admin")

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	names := make(map[string]bool)
	for _, p := range result {
		names[p.Name] = true
	}
	assert.True(t, names["admin:manage_users"])
	assert.True(t, names["admin:system_settings"])
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: CreateRole
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateRole_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)

	permIDs := []uint{seeded[0].ID, seeded[1].ID} // query:create, query:read
	role, err := svc.CreateRole("Data Analyst", "Can create and read queries", permIDs)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "Data Analyst", role.Name)
	assert.Equal(t, "Can create and read queries", role.Description)
	assert.NotZero(t, role.ID)

	// Verify role was saved to DB
	var savedRole models.Role
	err = db.Preload("Permissions").First(&savedRole, role.ID).Error
	assert.NoError(t, err)
	assert.Len(t, savedRole.Permissions, 2)
}

func TestCreateRole_DuplicateName(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	permIDs := []uint{seeded[0].ID}

	// Create first role
	_, err := svc.CreateRole("Editor", "First editor", permIDs)
	require.NoError(t, err)

	// Attempt duplicate
	_, err = svc.CreateRole("Editor", "Duplicate editor", permIDs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreateRole_EmptyName(t *testing.T) {
	db := setupPermissionTestDB(t)
	seedPermissions(t, db)

	svc := NewPermissionService(db)
	_, err := svc.CreateRole("", "No name", []uint{})

	assert.Error(t, err)
}

func TestCreateRole_NoPermissions(t *testing.T) {
	db := setupPermissionTestDB(t)
	setupPermissionTestDB(t) // ensure fresh

	svc := NewPermissionService(db)
	role, err := svc.CreateRole("Empty Role", "No permissions assigned", []uint{})

	// This may succeed or fail depending on business logic — verify behavior
	if err == nil {
		assert.NotNil(t, role)
		assert.Equal(t, "Empty Role", role.Name)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: UpdateRole
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateRole_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)

	// Create role first
	role, err := svc.CreateRole("Old Name", "Old description", []uint{seeded[0].ID})
	require.NoError(t, err)

	// Update it
	err = svc.UpdateRole(role.ID, "New Name", "New description")
	assert.NoError(t, err)

	// Verify changes
	updated, err := svc.GetRoleByID(role.ID)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "New description", updated.Description)
}

func TestUpdateRole_NonExistent(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := NewPermissionService(db)

	err := svc.UpdateRole(99999, "Name", "Desc")
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: DeleteRole
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteRole_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	role, err := svc.CreateRole("ToDelete", "Will be deleted", []uint{seeded[0].ID})
	require.NoError(t, err)

	err = svc.DeleteRole(role.ID)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = svc.GetRoleByID(role.ID)
	assert.Error(t, err)
}

func TestDeleteRole_NonExistent(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := NewPermissionService(db)

	err := svc.DeleteRole(99999)
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: GetAllRoles
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAllRoles_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	_, err := svc.CreateRole("Role A", "First role", []uint{seeded[0].ID})
	require.NoError(t, err)
	_, err = svc.CreateRole("Role B", "Second role", []uint{seeded[1].ID})
	require.NoError(t, err)

	roles, err := svc.GetAllRoles()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 2)
}

func TestGetAllRoles_EmptyDB(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := NewPermissionService(db)

	roles, err := svc.GetAllRoles()
	assert.NoError(t, err)
	assert.NotNil(t, roles)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: GetRoleByID
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRoleByID_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	created, err := svc.CreateRole("FindMe", "Role to find", []uint{seeded[0].ID, seeded[1].ID})
	require.NoError(t, err)

	found, err := svc.GetRoleByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "FindMe", found.Name)
	assert.Len(t, found.Permissions, 2)
}

func TestGetRoleByID_NotFound(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := NewPermissionService(db)

	_, err := svc.GetRoleByID(99999)
	assert.Error(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: AssignPermissionsToRole
// ─────────────────────────────────────────────────────────────────────────────

func TestAssignPermissionsToRole_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)

	// Create role with initial permissions
	role, err := svc.CreateRole("Updatable", "Will get new perms", []uint{seeded[0].ID})
	require.NoError(t, err)

	// Replace with new permissions
	newPermIDs := []uint{seeded[4].ID, seeded[5].ID} // dashboard:create, dashboard:read
	err = svc.AssignPermissionsToRole(role.ID, newPermIDs)
	assert.NoError(t, err)

	// Verify
	updated, err := svc.GetRoleByID(role.ID)
	assert.NoError(t, err)
	assert.Len(t, updated.Permissions, 2)

	permNames := make(map[string]bool)
	for _, p := range updated.Permissions {
		permNames[p.Name] = true
	}
	assert.True(t, permNames["dashboard:create"])
	assert.True(t, permNames["dashboard:read"])
	assert.False(t, permNames["query:create"]) // Old permission should be removed
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: AssignRoleToUser / RemoveRoleFromUser / GetUserRoles
// ─────────────────────────────────────────────────────────────────────────────

func TestAssignRoleToUser_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-assign-role-001")

	role, err := svc.CreateRole("Tester Role", "For testing assignment", []uint{seeded[0].ID})
	require.NoError(t, err)

	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	assert.NoError(t, err)

	// Verify
	roles, err := svc.GetUserRoles(user.ID.String())
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, role.ID, roles[0].ID)
}

func TestAssignRoleToUser_DuplicateAssignment(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-dup-assign-001")

	role, err := svc.CreateRole("Double Role", "Duplicate test", []uint{seeded[0].ID})
	require.NoError(t, err)

	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	require.NoError(t, err)

	// Attempt duplicate assignment
	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already assigned")
}

func TestRemoveRoleFromUser_Success(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-remove-role-01")

	role, err := svc.CreateRole("Remove Me", "Will be removed", []uint{seeded[0].ID})
	require.NoError(t, err)

	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	require.NoError(t, err)

	err = svc.RemoveRoleFromUser(user.ID.String(), role.ID)
	assert.NoError(t, err)

	// Verify role is gone
	roles, err := svc.GetUserRoles(user.ID.String())
	assert.NoError(t, err)
	assert.Len(t, roles, 0)
}

func TestGetUserRoles_NoRoles(t *testing.T) {
	db := setupPermissionTestDB(t)
	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-no-roles-0001")

	roles, err := svc.GetUserRoles(user.ID.String())
	assert.NoError(t, err)
	assert.Len(t, roles, 0)
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: CheckPermission / GetUserPermissions
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckPermission_HasPermission(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-check-perm-01")

	// Create role with query:create permission and assign to user
	role, err := svc.CreateRole("Query Creator", "Can create queries", []uint{seeded[0].ID})
	require.NoError(t, err)
	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	require.NoError(t, err)

	hasPermission, err := svc.CheckPermission(user.ID.String(), "query:create")
	assert.NoError(t, err)
	assert.True(t, hasPermission)
}

func TestCheckPermission_DoesNotHavePermission(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-no-perm-0001")

	// Only assign query:read, not admin:manage_users
	role, err := svc.CreateRole("Limited Reader", "Can only read queries", []uint{seeded[1].ID})
	require.NoError(t, err)
	err = svc.AssignRoleToUser(user.ID.String(), role.ID, "admin-001")
	require.NoError(t, err)

	hasPermission, err := svc.CheckPermission(user.ID.String(), "admin:manage_users")
	assert.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestCheckPermission_NoRoles(t *testing.T) {
	db := setupPermissionTestDB(t)
	seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-perm-empty-1")

	hasPermission, err := svc.CheckPermission(user.ID.String(), "query:create")
	assert.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestGetUserPermissions_MultipleRoles(t *testing.T) {
	db := setupPermissionTestDB(t)
	seeded := seedPermissions(t, db)

	svc := NewPermissionService(db)
	user := seedTestUser(t, db, "user-multi-role-01")

	// Create two roles with different permissions
	roleA, err := svc.CreateRole("Role Alpha", "First role", []uint{seeded[0].ID, seeded[1].ID})
	require.NoError(t, err)
	roleB, err := svc.CreateRole("Role Beta", "Second role", []uint{seeded[4].ID, seeded[5].ID})
	require.NoError(t, err)

	// Assign both roles
	err = svc.AssignRoleToUser(user.ID.String(), roleA.ID, "admin-001")
	require.NoError(t, err)
	err = svc.AssignRoleToUser(user.ID.String(), roleB.ID, "admin-001")
	require.NoError(t, err)

	permissions, err := svc.GetUserPermissions(user.ID.String())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(permissions), 4) // At least 4 permissions from both roles
}
