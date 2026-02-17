package services

import (
	"context"
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/resilience"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQueryCache is a mock implementation of QueryCacheInterface
type MockQueryCache struct {
	mock.Mock
}

func (m *MockQueryCache) GetCachedResult(ctx context.Context, key string) (*models.QueryResult, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func (m *MockQueryCache) SetCachedResult(ctx context.Context, key string, result *models.QueryResult, tags []string) error {
	args := m.Called(ctx, key, result, tags)
	return args.Error(0)
}

func (m *MockQueryCache) GenerateCacheKey(config *models.VisualQueryConfig, conn *models.Connection, userId string) string {
	args := m.Called(config, conn, userId)
	return args.String(0)
}

func (m *MockQueryCache) GenerateRawQueryCacheKey(connectionId string, sql string, params []interface{}, limit *int, offset *int) string {
	// Simple deterministic key for test
	return fmt.Sprintf("cache:raw:%s:%s", connectionId, sql)
}

func (m *MockQueryCache) InvalidateQuery(ctx context.Context, visualQueryId string) error {
	args := m.Called(ctx, visualQueryId)
	return args.Error(0)
}

func (m *MockQueryCache) InvalidateConnection(ctx context.Context, connectionId string) error {
	args := m.Called(ctx, connectionId)
	return args.Error(0)
}

func (m *MockQueryCache) InvalidateUser(ctx context.Context, userId string) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func TestExecute_CacheHit(t *testing.T) {
	// Setup
	mockCB := &resilience.MockCircuitBreaker{NameVal: "test-cb"}
	mockQC := new(MockQueryCache)
	// NewQueryExecutor(cb resilience.CircuitBreaker, qo *QueryOptimizer, qc QueryCacheInterface)
	executor := NewQueryExecutor(mockCB, nil, mockQC)

	ctx := context.Background()
	conn := &models.Connection{ID: "conn-1", Type: "postgres"}
	sqlQuery := "SELECT 1"

	// Expectations
	cachedResult := &models.QueryResult{
		Rows:   [][]interface{}{{"1"}},
		Cached: true,
	}

	cacheKey := "cache:raw:conn-1:SELECT 1"
	mockQC.On("GetCachedResult", ctx, cacheKey).Return(cachedResult, nil)

	// Execute
	result, err := executor.Execute(ctx, conn, sqlQuery, nil, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cachedResult, result)
	assert.True(t, result.Cached)

	mockQC.AssertExpectations(t)
}
