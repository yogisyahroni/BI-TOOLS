package handlers

import (
	"context"
	"insight-engine-backend/database"
	"insight-engine-backend/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	database.DB = db

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

	// AutoMigrate all models used in handler tests
	db.AutoMigrate(
		&models.Connection{},
		&models.VisualQuery{},
		&models.Collection{},
		&models.QueryExecutionLog{},
	)
	return db
}

// MockQueryExecutor is a mock implementation of services.QueryExecutorInterface
type MockQueryExecutor struct {
	mock.Mock
}

func (m *MockQueryExecutor) Execute(ctx context.Context, conn *models.Connection, sqlQuery string, params []interface{}, limit *int, offset *int) (*models.QueryResult, error) {
	args := m.Called(ctx, conn, sqlQuery, params, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func (m *MockQueryExecutor) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}
