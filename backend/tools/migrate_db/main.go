package main

import (
	"fmt"
	"log"
	"os"

	"insight-engine-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	fmt.Printf("Connecting to DB: %s\n", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Running manual migration for Workspace...")

	err = db.AutoMigrate(&models.Workspace{})
	if err != nil {
		log.Fatal("Failed to migrate Workspace:", err)
	}
	fmt.Println("Workspace migration success!")

	err = db.AutoMigrate(&models.WorkspaceMember{})
	if err != nil {
		log.Fatal("Failed to migrate WorkspaceMember:", err)
	}
	fmt.Println("WorkspaceMember migration success!")

	// Semantic Layer
	err = db.AutoMigrate(&models.SemanticModel{})
	if err != nil {
		log.Fatal("Failed to migrate SemanticModel:", err)
	}
	fmt.Println("SemanticModel migration success!")

	err = db.AutoMigrate(&models.SemanticDimension{})
	if err != nil {
		log.Fatal("Failed to migrate SemanticDimension:", err)
	}
	fmt.Println("SemanticDimension migration success!")

	err = db.AutoMigrate(&models.SemanticMetric{})
	if err != nil {
		log.Fatal("Failed to migrate SemanticMetric:", err)
	}
	fmt.Println("SemanticMetric migration success!")

	err = db.AutoMigrate(&models.SemanticRelationship{})
	if err != nil {
		log.Fatal("Failed to migrate SemanticRelationship:", err)
	}
	fmt.Println("SemanticRelationship migration success!")

	// Core Features
	err = db.AutoMigrate(&models.Collection{})
	if err != nil {
		log.Fatal("Failed to migrate Collection:", err)
	}
	fmt.Println("Collection migration success!")

	err = db.AutoMigrate(&models.CollectionItem{})
	if err != nil {
		log.Fatal("Failed to migrate CollectionItem:", err)
	}
	fmt.Println("CollectionItem migration success!")

	// err = db.AutoMigrate(&models.Dashboard{})
	// if err != nil {
	// 	log.Fatal("Failed to migrate Dashboard:", err)
	// }
	// fmt.Println("Dashboard migration success!")

	// err = db.AutoMigrate(&models.DashboardCard{})
	// if err != nil {
	// 	log.Fatal("Failed to migrate DashboardCard:", err)
	// }
	// fmt.Println("DashboardCard migration success!")

	err = db.AutoMigrate(&models.Connection{})
	if err != nil {
		log.Fatal("Failed to migrate Connection:", err)
	}
	fmt.Println("Connection migration success!")

	// Pipelines
	err = db.AutoMigrate(&models.Pipeline{})
	if err != nil {
		log.Fatal("Failed to migrate Pipeline:", err)
	}
	fmt.Println("Pipeline migration success!")

	err = db.AutoMigrate(&models.JobExecution{})
	if err != nil {
		log.Fatal("Failed to migrate JobExecution:", err)
	}
	fmt.Println("JobExecution migration success!")

	err = db.AutoMigrate(&models.QualityRule{})
	if err != nil {
		log.Fatal("Failed to migrate QualityRule:", err)
	}
	fmt.Println("QualityRule migration success!")

	// Webhooks
	err = db.AutoMigrate(&models.WebhookConfig{})
	if err != nil {
		log.Fatal("Failed to migrate WebhookConfig:", err)
	}
	fmt.Println("WebhookConfig migration success!")

	// Check if table exists
	if db.Migrator().HasTable("workspaces") {
		fmt.Println("Table 'workspaces' exists.")
	} else {
		fmt.Println("Table 'workspaces' DOES NOT exist.")
	}
}
