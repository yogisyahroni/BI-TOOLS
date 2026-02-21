package database

import (
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// ConfigureReadReplicas configures read/write splitting using DB_READ_REPLICAS env var
// Format: comma-separated DSNs
// Example: postgres://u:p@host1:5432/db,postgres://u:p@host2:5432/db
func ConfigureReadReplicas(db *gorm.DB) {
	replicaDSNsEnv := os.Getenv("DB_READ_REPLICAS")
	if replicaDSNsEnv == "" {
		// Log at info level so we know it's not configured
		log.Println("ℹ️ No read replicas configured (DB_READ_REPLICAS empty). Using primary for all queries.")
		return
	}

	rawDSNs := strings.Split(replicaDSNsEnv, ",")
	var replicas []gorm.Dialector

	for _, dsn := range rawDSNs {
		trimmed := strings.TrimSpace(dsn)
		if trimmed != "" {
			// Validate DSN strictly if possible, or just append
			replicas = append(replicas, postgres.Open(trimmed))
		}
	}

	if len(replicas) == 0 {
		log.Println("⚠️ DB_READ_REPLICAS parsed but no valid DSNs found.")
		return
	}

	// Register dbresolver
	// We use RandomPolicy by default for simple load balancing
	err := db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}))

	if err != nil {
		log.Printf("❌ Failed to configure read replicas: %v", err)
		// We do not fatal here, as the primary is still available.
		// However, in strict production this might be worth alerting on.
		return
	}

	log.Printf("✅ Configured %d read replica(s) with RandomPolicy", len(replicas))
}
