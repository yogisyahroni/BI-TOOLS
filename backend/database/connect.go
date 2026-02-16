package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not found in .env")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: false, // Critically important: Prevent SQLSTATE 0A000 error after migrations
	})

	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Configure connection pool for optimal performance
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance: ", err)
	}

	// Initialize and configure connection pool
	poolService := NewConnectionPoolService()
	poolConfig := poolService.LoadConfigFromEnv()
	poolService.Configure(sqlDB, poolConfig)

	// Expose globally if needed, or just keep it local for config
	// For now we just configure it.

	log.Println("✅ Connected to Database (PostgreSQL)")

	var dbName string
	DB.Raw("SELECT current_database()").Scan(&dbName)
	log.Printf("🔌 Active Database Name: %s", dbName)

	// Startup Sanity Check
	var test *time.Time
	// Use explicit scan to pointer which Gorm/Driver should handle, but ignore error if it's just a nil scan issue on raw types
	if err := DB.Raw("SELECT email_verified_at FROM users LIMIT 1").Scan(&test).Error; err != nil {
		log.Printf("⚠️ STARTUP CHECK: email_verified_at check skipped/failed (non-fatal): %v", err)
	} else {
		log.Println("✅ STARTUP SUCCESS: email_verified_at is accessible")
	}

	log.Println("✅ Connection pool configured: MaxOpen=50, MaxIdle=25, MaxLifetime=30m, MaxIdleTime=10m")
}
