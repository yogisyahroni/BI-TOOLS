package main

import (
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🛠️ Starting Database Repair Tool...")

	// Load .env from root (need to adjust path potentially, or rely on system env)
	// Assuming running from backend root or passing env vars.
	// Try loading from parent directories
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load(".env"); err != nil {
				log.Println("⚠️  No .env file found, relying on environment variables")
			}
		}
	}

	database.Connect()

	// Log DB Name
	var dbName string
	database.DB.Raw("SELECT current_database()").Scan(&dbName)
	log.Printf("🔌 Connected to Database: %s", dbName)

	// List actual columns
	var existingColumns []string
	database.DB.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'users'").Scan(&existingColumns)
	log.Printf("📋 Current columns in 'users' table (all schemas cached): %v", existingColumns)

	// Check for multiple users tables
	var tables []struct {
		Schema string
		Table  string
	}
	database.DB.Raw("SELECT table_schema as schema, table_name as table FROM information_schema.tables WHERE table_name = 'users'").Scan(&tables)
	log.Printf("🔍 Found 'users' tables in schemas: %+v", tables)

	// 1. User Table Fixes via Raw SQL (Foolproof)
	log.Println("Checking 'users' table columns with Raw SQL...")

	// Force recreation of critical columns to ensure they are visible
	dropQueries := []string{
		"ALTER TABLE users DROP COLUMN IF EXISTS email_verification_token",
		"ALTER TABLE users DROP COLUMN IF EXISTS email_verification_expires",
	}
	for _, q := range dropQueries {
		database.DB.Exec(q)
	}

	queries := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verification_token TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verification_expires TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_token TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS password_reset_expires TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS provider TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'user'",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active'",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_by TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivation_reason TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonation_token TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonation_expires TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonated_by TEXT",
		// Add other potential missing columns
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_by TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivation_reason TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonation_token TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonation_expires TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS impersonated_by TEXT",
	}

	for _, query := range queries {
		log.Printf("Running: %s", query)
		if err := database.DB.Exec(query).Error; err != nil {
			log.Printf("❌ Failed: %v", err)
		} else {
			log.Printf("✅ Success")
		}
	}

	// Verify by selecting
	log.Println("Validating schema by selecting from users...")
	var dump []map[string]interface{}
	if err := database.DB.Raw("SELECT email_verification_token FROM users LIMIT 1").Scan(&dump).Error; err != nil {
		log.Printf("❌ Validation Failed: %v", err)
	} else {
		log.Println("✅ Validation Success: Column 'email_verification_token' is queryable.")
	}

	// 2. Ensure Dashboard Cards table exists (just in case)
	if !database.DB.Migrator().HasTable(&models.DashboardCard{}) {
		log.Println("➕ Creating 'dashboard_cards' table...")
		database.DB.AutoMigrate(&models.DashboardCard{})
	}

	log.Println("🎉 Database repair completed.")
}
