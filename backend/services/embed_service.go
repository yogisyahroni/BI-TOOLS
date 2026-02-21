package services

import (
	"errors"
	"fmt"
	"insight-engine-backend/dtos"
	"insight-engine-backend/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type EmbedService struct {
	DB *gorm.DB
}

func NewEmbedService(db *gorm.DB) *EmbedService {
	return &EmbedService{
		DB: db,
	}
}

// GenerateEmbedToken creates a signed JWT for embedding a dashboard
func (s *EmbedService) GenerateEmbedToken(req dtos.EmbedTokenRequest) (*dtos.EmbedTokenResponse, error) {
	// 1. Get Secret
	secret := os.Getenv("EMBED_SECRET")
	if secret == "" {
		secret = os.Getenv("NEXTAUTH_SECRET")
		if secret == "" {
			return nil, errors.New("server configuration error: missing embed secret")
		}
	}

	// 2. Set Expiration
	if req.Expiration <= 0 {
		req.Expiration = 60 // Default 1 hour
	}
	expiresAt := time.Now().Add(time.Duration(req.Expiration) * time.Minute)

	// 3. Create Claims
	claims := jwt.MapClaims{
		"dashboard_id":    req.DashboardID,
		"allowed_filters": req.AllowedFilters,
		"hidden_widgets":  req.HiddenWidgets,
		"theme":           req.Theme,
		"exp":             expiresAt.Unix(),
		"iat":             time.Now().Unix(),
		"iss":             "insight-engine-backend",
		"sub":             "embed-token",
	}

	// 4. Sign Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	// 5. Construct Response
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	embedURL := fmt.Sprintf("%s/embed/dashboard/%s?token=%s", frontendURL, req.DashboardID, signedToken)

	return &dtos.EmbedTokenResponse{
		Token:     signedToken,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		EmbedURL:  embedURL,
	}, nil
}

// ValidateEmbedToken verifies the token and returns the dashboard configuration
func (s *EmbedService) ValidateEmbedToken(tokenString string) (*models.Dashboard, error) {
	secret := os.Getenv("EMBED_SECRET")
	if secret == "" {
		secret = os.Getenv("NEXTAUTH_SECRET")
	}

	// 1. Parse and Validate Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 2. Extract Dashboard ID
		dashboardID, ok := claims["dashboard_id"].(string)
		if !ok {
			return nil, errors.New("invalid token payload: missing dashboard_id")
		}

		// 3. Fetch Dashboard from DB
		var dashboard models.Dashboard
		result := s.DB.Preload("Cards").First(&dashboard, "id = ?", dashboardID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil, errors.New("dashboard not found")
			}
			return nil, result.Error
		}

		// 4. Note: We could apply "HiddenWidgets" filtering here if we want to be strict,
		// or let the frontend handle it. For security, applying it here is better.
		// But for now, let's just return the full dashboard and trust the client SDK to hide elements,
		// or filter the cards list.
		// Let's filter cards if hidden_widgets is present.
		if hiddenWidgetsRaw, ok := claims["hidden_widgets"].([]interface{}); ok {
			hiddenMap := make(map[string]bool)
			for _, hw := range hiddenWidgetsRaw {
				if s, ok := hw.(string); ok {
					hiddenMap[s] = true
				}
			}
			if len(hiddenMap) > 0 {
				filteredCards := []models.DashboardCard{} // Assuming models.DashboardCard is the type
				for _, card := range dashboard.Cards {
					if !hiddenMap[card.ID.String()] { // Use card.ID.String() for map key
						filteredCards = append(filteredCards, card)
					}
				}
				dashboard.Cards = filteredCards
			}
		}

		return &dashboard, nil
	}

	return nil, errors.New("invalid or expired token")
}
