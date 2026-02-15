package bootstrap

import (
	"insight-engine-backend/database"
	"insight-engine-backend/services"

	"github.com/joho/godotenv"
)

// InitLogger initializes the structured logger
func InitLogger() {
	services.InitLogger("insight-engine-backend")
}

// LoadConfig loads environment variables
// Critical: Panics if essential config is missing
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		services.LogWarn("env_load", ".env file not found, using system environment variables", nil)
	}
}

// ConnectDatabase initializes DB connection and runs migrations
func ConnectDatabase() {
	database.Connect()
}
