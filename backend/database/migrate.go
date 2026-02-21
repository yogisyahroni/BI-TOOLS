package database

import (
	"insight-engine-backend/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Migrate runs auto-migration for database models
func Migrate() {
	if DB == nil {
		log.Fatal("Database connection not initialized")
	}

	log.Println("🔄 Running Database Migrations...")

	// Enable uuid-ossp extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Printf("ΓÜá∩╕Å Warning: Failed to enable uuid-ossp extension: %v", err)
	}

	// AutoMigrate models
	// Use separate migration for generic models
	// Force Drop tables for clean schema (Dev/Verify only - resolving type mismatches)
	log.Println("Dropping tables to ensure clean schema...")
	if err := DB.Migrator().DropTable(&models.CollectionItem{}, &models.Collection{}, &models.Pulse{}, &models.Dashboard{}, &models.DashboardCard{}, &models.User{}); err != nil {
		log.Printf("Failed to drop tables: %v", err)
	}

	log.Println("Running AutoMigrate...")
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Collection{},
		&models.CollectionItem{},
		&models.Dashboard{},
		&models.DashboardCard{},
		&models.Pulse{},
		&models.WebhookConfig{},
		&models.WebhookLog{},
		&models.AIUsageRequest{},
		&models.AIBudget{},
		&models.BudgetAlert{},
		&models.RateLimitConfig{},
		&models.RateLimitViolation{},
		&models.ColumnPermission{},
		// Dependencies for Alert
		&models.SavedQuery{},
		&models.QueryVersion{},
		&models.Alert{},
		&models.AlertHistory{},
		&models.AlertAcknowledgment{},
		&models.AlertNotificationChannelConfig{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.AuditLog{},
		&models.DataClassification{},
		&models.ColumnMetadata{},
		// Core Entities
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.Connection{},
		// Semantic Layer
		&models.SemanticModel{},
		&models.SemanticDimension{},
		&models.SemanticMetric{},
		&models.SemanticRelationship{},
		// Queries
		// SavedQuery and QueryVersion moved up to satisfy Alert dependency
		// Business Glossary
		&models.BusinessTerm{},
		&models.TermColumnMapping{},
		// Export Jobs
		&models.ExportJob{},
		// Stories & Presentations
		&models.Story{}, // TASK-161
		// Pulses
		&models.Pulse{}, // TASK-156
	); err != nil {
		log.Printf("⚠️ Core models migration warning: %v", err)
	}

	// ISOLATED MIGRATION FOR ALERTS & AUDIT LOGS (DEBUGGING MISSING TABLES)
	log.Println("🔧 Running ISOLATED migration for Alerts and AuditLogs...")
	if err := DB.AutoMigrate(
		&models.Alert{},
		&models.AlertHistory{},
		&models.AlertAcknowledgment{},
		&models.AlertNotificationChannelConfig{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.AuditLog{},
	); err != nil {
		log.Printf("❌ Alert/AuditLog migration FAILED: %v", err)
	} else {
		log.Println("✅ Alert/AuditLog migration successful")
	}

	// FORCE MIGRATE PULSE (TASK-156 DEBUG) - Redundant but harmless to keep
	if err := DB.AutoMigrate(&models.Pulse{}); err != nil {
		log.Printf("❌ Pulse migration FAILED: %v", err)
	} else {
		log.Println("✅ Pulse migration successful")
	}

	// FORCE MIGRATE DASHBOARDS (Fix for 500 error)
	log.Println("Force running Dashboard migration...")
	if err := DB.AutoMigrate(&models.Dashboard{}, &models.DashboardCard{}); err != nil {
		log.Printf("❌ Dashboard migration FAILED: %v", err)
	} else {
		log.Println("✅ Dashboard migration successful")
	}

	// FORCE MIGRATE STORY (Fix for 500 error)
	log.Println("Force running Story migration...")
	if err := DB.AutoMigrate(&models.Story{}); err != nil {
		log.Printf("❌ Story migration FAILED: %v", err)
	} else {
		log.Println("✅ Story migration successful")
	}

	// Migrate Webhooks separately to ensure they are created even if core migration has warnings
	if err := DB.AutoMigrate(
		&models.Webhook{},
		&models.WebhookConfig{},
		&models.WebhookLog{},
	); err != nil {
		log.Printf("⚠️ Webhook migration warning: %v", err)
	}

	// Migrate AI Usage and Budget models
	if err := DB.AutoMigrate(
		&models.AIUsageRequest{},
		&models.AIBudget{},
		&models.BudgetAlert{},
		&models.RateLimitConfig{},
		&models.RateLimitViolation{},
	); err != nil {
		log.Printf("⚠️ AI Usage migration warning: %v", err)
	}

	// Migrate Embed Tokens
	if err := DB.AutoMigrate(&models.EmbedToken{}); err != nil {
		log.Printf("⚠️ Embed Token migration warning: %v", err)
	}

	// Migrate Shares
	if err := DB.AutoMigrate(&models.Share{}); err != nil {
		log.Printf("⚠️ Share migration warning: %v", err)
	}

	// Migrate Collections
	if err := DB.AutoMigrate(&models.Collection{}); err != nil {
		log.Printf("⚠️ Collection migration warning: %v", err)
	}

	// Handle User migration with resilience
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Printf("⚠️ User migration warning (indexes might need manual fix): %v", err)
	}

	// Explicitly ensure critical columns exist (AutoMigrate might fail on indexes)
	columns := []string{
		"EmailVerified",
		"EmailVerificationToken",
		"EmailVerificationExpires",
		"PasswordResetToken",
		"PasswordResetExpires",
		"Provider",
		"ProviderID",
	}

	for _, col := range columns {
		if !DB.Migrator().HasColumn(&models.User{}, col) {
			log.Printf("🔧 Manually adding '%s' column to users table...", col)
			if err := DB.Migrator().AddColumn(&models.User{}, col); err != nil {
				log.Printf("❌ Failed to add '%s' column: %v", col, err)
			} else {
				log.Printf("✅ '%s' column added successfully", col)
			}
		}
	}

	SeedDefaultUser()

	log.Println("✅ Database Migrations completed (with potential warnings)")
}

// SeedDefaultUser creates a demo user if one doesn't exist
func SeedDefaultUser() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		log.Println("🌱 Seeding default user yogisyahroni766.ysr@gmai.com...")

		password := "Namakamu766!!"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		user := models.User{
			Email:         "yogisyahroni766.ysr@gmai.com",
			Username:      "yogi_admin",
			Name:          "Yogi Syahroni",
			Role:          "admin",
			EmailVerified: true,
			Status:        "active",
			Password:      string(hashedPassword),
		}

		if err := DB.Create(&user).Error; err != nil {
			log.Printf("❌ Failed to seed default user: %v", err)
		} else {
			log.Println("✅ Default user seeded successfully")
		}
	}
}

// SeedDataClassifications ensures default data classifications exist
func SeedDataClassifications() {
	classifications := []models.DataClassification{
		{Name: "Public", Description: "Data available to everyone", Color: "#22c55e"},
		{Name: "Internal", Description: "Data for internal use only", Color: "#3b82f6"},
		{Name: "Confidential", Description: "Sensitive business data", Color: "#f59e0b"},
		{Name: "PII", Description: "Personally Identifiable Information", Color: "#ef4444"},
	}

	for _, c := range classifications {
		// Use FirstOrCreate to avoid duplicates
		if err := DB.Where("name = ?", c.Name).FirstOrCreate(&c).Error; err != nil {
			log.Printf("⚠️ Failed to seed classification %s: %v", c.Name, err)
		}
	}

	log.Println("✅ Data Classifications seeded")
}
