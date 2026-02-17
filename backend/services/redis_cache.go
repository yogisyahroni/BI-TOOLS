package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache manages Redis connections and cache operations with
// graceful fallback, hit-rate monitoring, and cache warming.
type RedisCache struct {
	client *redis.Client

	// ---- Graceful fallback ----
	// localFallback is a thread-safe in-memory LRU-ish map used when
	// Redis is unreachable. Entries are evicted after fallbackTTL.
	localFallback   map[string]fallbackEntry
	fallbackMu      sync.RWMutex
	fallbackEnabled bool
	fallbackTTL     time.Duration

	// ---- Hit-rate counters (in-process) ----
	localHits   int64
	localMisses int64

	// ---- Warm keys ----
	warmKeys []WarmKeySpec
}

// fallbackEntry is an in-memory cache entry used when Redis is down
type fallbackEntry struct {
	Value     []byte
	ExpiresAt time.Time
}

// WarmKeySpec describes a key that should be proactively cached
type WarmKeySpec struct {
	Key      string
	Loader   func(ctx context.Context) ([]byte, error) // function that produces the value
	TTL      time.Duration
	Interval time.Duration // how often to refresh (0 = TTL/2)
}

// RedisCacheConfig holds Redis configuration
type RedisCacheConfig struct {
	Host           string
	Password       string
	DB             int
	MaxRetries     int
	PoolSize       int
	EnableFallback bool          // enable in-memory fallback when Redis is unreachable
	FallbackTTL    time.Duration // how long fallback entries survive (default 5m)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits             int64   `json:"hits"`
	Misses           int64   `json:"misses"`
	HitRate          float64 `json:"hitRate"`
	TotalKeys        int64   `json:"totalKeys"`
	MemoryUsage      int64   `json:"memoryUsage"` // bytes
	Uptime           int64   `json:"uptime"`      // seconds
	ConnectedClients int64   `json:"connectedClients"`
	EvictedKeys      int64   `json:"evictedKeys"`
	ExpiredKeys      int64   `json:"expiredKeys"`
	FallbackActive   bool    `json:"fallbackActive"` // true = currently using local fallback
	LocalHits        int64   `json:"localHits"`      // in-process counter
	LocalMisses      int64   `json:"localMisses"`    // in-process counter
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config RedisCacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Host,
		Password:     config.Password,
		DB:           config.DB,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fallbackTTL := config.FallbackTTL
	if fallbackTTL == 0 {
		fallbackTTL = 5 * time.Minute
	}

	return &RedisCache{
		client:          client,
		localFallback:   make(map[string]fallbackEntry),
		fallbackEnabled: config.EnableFallback,
		fallbackTTL:     fallbackTTL,
	}, nil
}

// ---- Core Operations ----

// Get retrieves a value from cache with graceful fallback
func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		atomic.AddInt64(&rc.localMisses, 1)
		return nil, nil // Cache miss
	}
	if err != nil {
		// Redis unreachable — try local fallback
		if rc.fallbackEnabled {
			if fbVal := rc.getFromFallback(key); fbVal != nil {
				atomic.AddInt64(&rc.localHits, 1)
				return fbVal, nil
			}
		}
		atomic.AddInt64(&rc.localMisses, 1)
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	atomic.AddInt64(&rc.localHits, 1)

	// Write-through to local fallback
	if rc.fallbackEnabled {
		rc.setFallback(key, val, rc.fallbackTTL)
	}

	return val, nil
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := rc.client.Set(ctx, key, value, ttl).Err()

	// Always update local fallback regardless of Redis result
	if rc.fallbackEnabled {
		fbTTL := ttl
		if fbTTL == 0 || fbTTL > rc.fallbackTTL {
			fbTTL = rc.fallbackTTL
		}
		rc.setFallback(key, value, fbTTL)
	}

	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()

	rc.deleteFallback(key)

	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// DeleteByPattern removes all keys matching a pattern
func (rc *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64

	for {
		var keys []string
		var err error

		keys, cursor, err = rc.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
		}

		if len(keys) > 0 {
			if _, delErr := rc.client.Del(ctx, keys...).Result(); delErr != nil {
				return fmt.Errorf("failed to delete keys: %w", delErr)
			}
			// Also clear from local fallback
			for _, k := range keys {
				rc.deleteFallback(k)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

// ---- Statistics (full INFO parsing) ----

// GetStats retrieves comprehensive cache statistics from Redis INFO
func (rc *RedisCache) GetStats(ctx context.Context) (*CacheStats, error) {
	dbSize, err := rc.client.DBSize(ctx).Result()
	if err != nil {
		// Redis may be down — return local counters only
		lHits := atomic.LoadInt64(&rc.localHits)
		lMisses := atomic.LoadInt64(&rc.localMisses)
		total := lHits + lMisses
		hitRate := 0.0
		if total > 0 {
			hitRate = float64(lHits) / float64(total) * 100
		}
		return &CacheStats{
			FallbackActive: true,
			LocalHits:      lHits,
			LocalMisses:    lMisses,
			HitRate:        hitRate,
		}, nil
	}

	stats := &CacheStats{
		TotalKeys:   dbSize,
		LocalHits:   atomic.LoadInt64(&rc.localHits),
		LocalMisses: atomic.LoadInt64(&rc.localMisses),
	}

	// Parse INFO stats section for hit/miss counters
	infoStats, err := rc.client.Info(ctx, "stats").Result()
	if err == nil {
		stats.Hits = parseInfoInt(infoStats, "keyspace_hits")
		stats.Misses = parseInfoInt(infoStats, "keyspace_misses")
		stats.EvictedKeys = parseInfoInt(infoStats, "evicted_keys")
		stats.ExpiredKeys = parseInfoInt(infoStats, "expired_keys")
		totalOps := stats.Hits + stats.Misses
		if totalOps > 0 {
			stats.HitRate = float64(stats.Hits) / float64(totalOps) * 100
		}
	}

	// Parse INFO memory
	infoMem, err := rc.client.Info(ctx, "memory").Result()
	if err == nil {
		stats.MemoryUsage = parseInfoInt(infoMem, "used_memory")
	}

	// Parse INFO server
	infoSrv, err := rc.client.Info(ctx, "server").Result()
	if err == nil {
		stats.Uptime = parseInfoInt(infoSrv, "uptime_in_seconds")
	}

	// Parse INFO clients
	infoCli, err := rc.client.Info(ctx, "clients").Result()
	if err == nil {
		stats.ConnectedClients = parseInfoInt(infoCli, "connected_clients")
	}

	return stats, nil
}

// ---- Tag-based invalidation ----

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence %s: %w", key, err)
	}
	return count > 0, nil
}

// SetWithTags stores a value with associated tags for invalidation
func (rc *RedisCache) SetWithTags(ctx context.Context, key string, value []byte, ttl time.Duration, tags []string) error {
	pipe := rc.client.Pipeline()

	pipe.Set(ctx, key, value, ttl)

	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		pipe.SAdd(ctx, tagKey, key)
		pipe.Expire(ctx, tagKey, ttl+1*time.Minute)
	}

	_, err := pipe.Exec(ctx)

	// Update local fallback
	if rc.fallbackEnabled {
		fbTTL := ttl
		if fbTTL == 0 || fbTTL > rc.fallbackTTL {
			fbTTL = rc.fallbackTTL
		}
		rc.setFallback(key, value, fbTTL)
	}

	if err != nil {
		return fmt.Errorf("failed to set key with tags: %w", err)
	}

	return nil
}

// InvalidateByTag removes all keys associated with a tag
func (rc *RedisCache) InvalidateByTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)

	keys, err := rc.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get tag members: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	pipe := rc.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
		rc.deleteFallback(key)
	}
	pipe.Del(ctx, tagKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate by tag: %w", err)
	}

	return nil
}

// ---- Cache Warming ----

// RegisterWarmKey adds a key to be proactively cached
func (rc *RedisCache) RegisterWarmKey(spec WarmKeySpec) {
	rc.warmKeys = append(rc.warmKeys, spec)
}

// StartWarming starts background goroutines to keep warm keys fresh.
// Call this once after registering all warm keys. The goroutines
// are stopped when ctx is cancelled.
func (rc *RedisCache) StartWarming(ctx context.Context) {
	for _, spec := range rc.warmKeys {
		go rc.warmLoop(ctx, spec)
	}
}

func (rc *RedisCache) warmLoop(ctx context.Context, spec WarmKeySpec) {
	interval := spec.Interval
	if interval == 0 {
		interval = spec.TTL / 2
	}
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Warm immediately on startup
	rc.warmOnce(ctx, spec)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rc.warmOnce(ctx, spec)
		}
	}
}

func (rc *RedisCache) warmOnce(ctx context.Context, spec WarmKeySpec) {
	warmCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	data, err := spec.Loader(warmCtx)
	if err != nil {
		LogInfo("cache_warm_failed", "Failed to warm cache key", map[string]interface{}{
			"key":   spec.Key,
			"error": err.Error(),
		})
		return
	}

	if setErr := rc.Set(warmCtx, spec.Key, data, spec.TTL); setErr != nil {
		LogInfo("cache_warm_set_failed", "Failed to set warmed cache key", map[string]interface{}{
			"key":   spec.Key,
			"error": setErr.Error(),
		})
	}
}

// ---- Graceful Fallback (in-memory) ----

func (rc *RedisCache) getFromFallback(key string) []byte {
	rc.fallbackMu.RLock()
	defer rc.fallbackMu.RUnlock()

	entry, ok := rc.localFallback[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return entry.Value
}

func (rc *RedisCache) setFallback(key string, value []byte, ttl time.Duration) {
	rc.fallbackMu.Lock()
	defer rc.fallbackMu.Unlock()

	// Cap fallback size at 10,000 entries to prevent memory explosion
	if len(rc.localFallback) >= 10000 {
		rc.evictOldestFallback()
	}

	rc.localFallback[key] = fallbackEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (rc *RedisCache) deleteFallback(key string) {
	if !rc.fallbackEnabled {
		return
	}
	rc.fallbackMu.Lock()
	defer rc.fallbackMu.Unlock()
	delete(rc.localFallback, key)
}

// evictOldestFallback removes the oldest 20% of entries
func (rc *RedisCache) evictOldestFallback() {
	// Simple approach: delete expired entries first
	now := time.Now()
	for k, v := range rc.localFallback {
		if now.After(v.ExpiresAt) {
			delete(rc.localFallback, k)
		}
	}
	// If still over 80% capacity, delete random entries until under limit
	limit := 8000
	for k := range rc.localFallback {
		if len(rc.localFallback) <= limit {
			break
		}
		delete(rc.localFallback, k)
	}
}

// ---- Lifecycle ----

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Ping checks if Redis is reachable
func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// FlushDB clears all keys in the current database (use with caution!)
func (rc *RedisCache) FlushDB(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// ResetLocalCounters resets the in-process hit/miss counters
func (rc *RedisCache) ResetLocalCounters() {
	atomic.StoreInt64(&rc.localHits, 0)
	atomic.StoreInt64(&rc.localMisses, 0)
}

// ---- INFO Parsing Helper ----

// parseInfoInt extracts an integer value from a Redis INFO response by field name
func parseInfoInt(info string, field string) int64 {
	lines := strings.Split(info, "\n")
	prefix := field + ":"
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			valStr := strings.TrimPrefix(line, prefix)
			valStr = strings.TrimSpace(valStr)
			val, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				return 0
			}
			return val
		}
	}
	return 0
}
