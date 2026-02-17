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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	database.DB = db
	// AutoMigrate all models used in handler tests
	db.AutoMigrate(
		&models.Connection{},
		&models.VisualQuery{},
		&models.Collection{},
		&models.User{},
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
