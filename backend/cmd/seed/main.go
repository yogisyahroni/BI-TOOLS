package main

import (
	"log"
	"time"

	"insight-engine-backend/bootstrap"
	"insight-engine-backend/database"
	"insight-engine-backend/models"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize Env and DB
	bootstrap.InitLogger() // Logger is often required by other services or database
	bootstrap.LoadConfig()
	bootstrap.ConnectDatabase()

	// Upsert demo user
	if err := upsertUser("demo@example.com", "demo", "password123", "Demo User"); err != nil {
		log.Fatal(err)
	}

	// Upsert specific user from request
	if err := upsertUser("yogisyahroni766.ysr@gmail.com", "yogisyahroni766", "Namakamu766!!", "Yogi Syahroni"); err != nil {
		log.Fatal(err)
	}
}

func upsertUser(email, username, password, name string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user := models.User{
		Email:           email,
		Username:        username,
		Password:        string(hashedPassword),
		Name:            name,
		EmailVerified:   true,
		EmailVerifiedAt: &now,
	}

	var existing models.User
	if err := database.DB.Where("email = ?", user.Email).First(&existing).Error; err == nil {
		updates := map[string]interface{}{
			"password":          string(hashedPassword),
			"email_verified":    true,
			"email_verified_at": now,
		}
		if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		log.Printf("Γ£à User %s UPDATED and VERIFIED.\n", email)
	} else {
		if err := database.DB.Create(&user).Error; err != nil {
			return err
		}
		log.Printf("Γ£à User %s CREATED and VERIFIED.\n", email)
	}
	return nil
}
