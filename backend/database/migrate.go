package database

import (
	"insight-engine-backend/models"
	"log"
)

// Migrate runs auto-migration for database models
func Migrate() {
	if DB == nil {
		log.Fatal("Database connection not initialized")
	}

	log.Println("🔄 Running Database Migrations...")

	// AutoMigrate models
	// AutoMigrate models
	// Use separate migration for generic models
	if err := DB.AutoMigrate(
		&models.Dashboard{},
		&models.DashboardCard{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.DataClassification{},
		&models.ColumnMetadata{},
		&models.ColumnPermission{},
		&models.Alert{},
		&models.AlertHistory{},
		&models.AlertAcknowledgment{},
		&models.AlertNotificationChannelConfig{},
		// Core Entities
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.Connection{},
		// Semantic Layer
		&models.SemanticModel{},
		&models.SemanticDimension{},
		&models.SemanticMetric{},
		&models.SemanticRelationship{},
		// Business Glossary
		&models.BusinessTerm{},
		&models.TermColumnMapping{},
	); err != nil {
		log.Printf("⚠️ Core models migration warning: %v", err)
	}

	// Migrate Webhooks separately to ensure they are created even if core migration has warnings
	if err := DB.AutoMigrate(
		&models.Webhook{},
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

	log.Println("✅ Database Migrations completed (with potential warnings)")
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
