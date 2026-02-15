package dtos

type EmbedTokenRequest struct {
	DashboardID    string                 `json:"dashboard_id" validate:"required"`
	Expiration     int                    `json:"expiration_minutes" validate:"min=1,max=1440"` // Max 24 hours
	AllowedFilters map[string]interface{} `json:"allowed_filters"`
	HiddenWidgets  []string               `json:"hidden_widgets"`
	Theme          string                 `json:"theme"` // "light" or "dark"
}

type EmbedTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	EmbedURL  string `json:"embed_url"`
}
