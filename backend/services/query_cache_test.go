package services_test

import (
	"context"
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/services"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func HelperSetupRedis(t *testing.T) (*miniredis.Miniredis, *services.RedisCache) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	config := services.RedisCacheConfig{
		Host:           mr.Addr(),
		Password:       "",
		DB:             0,
		MaxRetries:     1,
		PoolSize:       1,
		EnableFallback: true,
		FallbackTTL:    1 * time.Minute,
	}

	rc, err := services.NewRedisCache(config)
	require.NoError(t, err)

	return mr, rc
}

func TestRedisCache_SetGet(t *testing.T) {
	mr, rc := HelperSetupRedis(t)
	defer mr.Close()
	defer rc.Close()

	ctx := context.Background()
	key := "test:key"
	value := []byte("hello world")

	// Test Set
	err := rc.Set(ctx, key, value, 1*time.Minute)
	require.NoError(t, err)

	// Test Get
	got, err := rc.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, got)

	// Test Exists
	exists, err := rc.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Fallback(t *testing.T) {
	mr, rc := HelperSetupRedis(t)
	defer rc.Close()

	ctx := context.Background()
	key := "fallback:key"
	value := []byte("fallback data")

	// Set data normally
	err := rc.Set(ctx, key, value, 1*time.Minute)
	require.NoError(t, err)

	// Kill Redis to force fallback
	mr.Close()

	// Give it a moment? No, Miniredis close is immediate usually.
	// rc.Get should fail to hit Redis, but succeed via local fallback if implemented purely in memory on write
	// Note: RedisCache implementation writes to local fallback on Set() if enabled.

	got, err := rc.Get(ctx, key)
	require.NoError(t, err, "Should not error if fallback is successful")
	assert.Equal(t, value, got, "Should get data from fallback")
}

func TestQueryCache_GenerateCacheKey(t *testing.T) {
	// QueryCache doesn't need real redis for key generation
	qc := services.NewQueryCache(nil, time.Minute)

	conn := &models.Connection{ID: "conn-123"}
	limit10 := 10
	config := &models.VisualQueryConfig{
		Tables: []models.TableSelection{{Name: "users"}},
		Limit:  &limit10,
	}
	userId := "user-456"

	key := qc.GenerateCacheKey(config, conn, userId)
	assert.Contains(t, key, "cache:vq:")

	// Same inputs should match
	key2 := qc.GenerateCacheKey(config, conn, userId)
	assert.Equal(t, key, key2)

	// Different inputs should differ
	limit20 := 20
	config.Limit = &limit20
	key3 := qc.GenerateCacheKey(config, conn, userId)
	assert.NotEqual(t, key, key3)
}

func TestQueryCache_SetGetResult(t *testing.T) {
	mr, rc := HelperSetupRedis(t)
	defer mr.Close()
	defer rc.Close()

	qc := services.NewQueryCache(rc, 10*time.Minute)
	ctx := context.Background()

	key := "query:result:1"
	result := &models.QueryResult{
		Columns:  []string{"id", "name"},
		Rows:     [][]interface{}{{"1", "Alice"}, {"2", "Bob"}},
		RowCount: 2,
	}
	tags := []string{"table:users"}

	// Set
	err := qc.SetCachedResult(ctx, key, result, tags)
	require.NoError(t, err)

	// Get
	cached, err := qc.GetCachedResult(ctx, key)
	require.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, result.Columns, cached.Columns)
	assert.Equal(t, result.RowCount, cached.RowCount)
	// JSON marshaling numbers typically become float64, check basic equality only
	assert.Equal(t, len(result.Rows), len(cached.Rows))
}

func TestQueryCache_Invalidation(t *testing.T) {
	mr, rc := HelperSetupRedis(t)
	defer mr.Close()
	defer rc.Close()

	qc := services.NewQueryCache(rc, 10*time.Minute)
	ctx := context.Background()

	// Store result with tags
	key := "query:to:invalidate"
	result := &models.QueryResult{RowCount: 1}
	err := qc.SetCachedResult(ctx, key, result, []string{"conn:123"})
	require.NoError(t, err)

	// Ensure it exists
	exists, _ := rc.Exists(ctx, key)
	assert.True(t, exists)

	// Invalidate by connection tag
	err = qc.InvalidateConnection(ctx, "123")
	require.NoError(t, err)

	// Ensure it's gone
	exists, _ = rc.Exists(ctx, key)
	assert.False(t, exists)
}

func TestRedisCache_WarmKeys(t *testing.T) {
	mr, rc := HelperSetupRedis(t)
	defer mr.Close()
	defer rc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register warm key
	callCount := 0
	rc.RegisterWarmKey(services.WarmKeySpec{
		Key: "warm:key",
		TTL: 1 * time.Minute,
		Loader: func(ctx context.Context) ([]byte, error) {
			callCount++
			return []byte(fmt.Sprintf("data-%d", callCount)), nil
		},
		Interval: 100 * time.Millisecond,
	})

	// Start warming
	rc.StartWarming(ctx)

	// Wait for first warm
	time.Sleep(200 * time.Millisecond)

	// Verify data is in cache
	val, err := rc.Get(ctx, "warm:key")
	require.NoError(t, err)
	assert.Contains(t, string(val), "data-")
}
