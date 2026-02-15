package database

import (
	"database/sql"
	"insight-engine-backend/models"
	"log"
	"os"
	"strconv"
	"time"
)

// ConnectionPoolService manages database connection pool settings
type ConnectionPoolService struct {
	db *sql.DB
}

// NewConnectionPoolService creates a new instance
func NewConnectionPoolService() *ConnectionPoolService {
	return &ConnectionPoolService{}
}

// Configure applies pool settings to a database connection
func (s *ConnectionPoolService) Configure(db *sql.DB, config models.ConnectionPoolConfig) {
	s.db = db

	// Apply settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	log.Printf("✅ Connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%s, MaxIdleTime=%s",
		config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime, config.ConnMaxIdleTime)
}

// LoadConfigFromEnv loads pool configuration from environment variables
func (s *ConnectionPoolService) LoadConfigFromEnv() models.ConnectionPoolConfig {
	config := models.DefaultPoolConfig()

	if val := os.Getenv("DB_MAX_OPEN_CONNS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			config.MaxOpenConns = v
		}
	}

	if val := os.Getenv("DB_MAX_IDLE_CONNS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			config.MaxIdleConns = v
		}
	}

	if val := os.Getenv("DB_CONN_MAX_LIFETIME"); val != "" {
		if v, err := time.ParseDuration(val); err == nil {
			config.ConnMaxLifetime = v
		}
	}

	if val := os.Getenv("DB_CONN_MAX_IDLE_TIME"); val != "" {
		if v, err := time.ParseDuration(val); err == nil {
			config.ConnMaxIdleTime = v
		}
	}

	return config
}

// GetStats returns current pool statistics
func (s *ConnectionPoolService) GetStats() map[string]interface{} {
	if s.db == nil {
		return map[string]interface{}{"status": "not_configured"}
	}

	stats := s.db.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
	}
}
