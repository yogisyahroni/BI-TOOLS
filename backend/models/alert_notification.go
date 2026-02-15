package models

// This file contains additional DTO types for alert notifications
// The main model definitions are in alert.go

// CreateAlertNotificationChannelRequest represents a request to create a notification channel
type CreateAlertNotificationChannelRequest struct {
	ChannelType AlertNotificationChannel `json:"channelType" validate:"required"`
	IsEnabled   bool                     `json:"isEnabled"`
	Config      map[string]interface{}   `json:"config,omitempty"`
}

// UpdateAlertNotificationChannelRequest represents a request to update a notification channel
type UpdateAlertNotificationChannelRequest struct {
	IsEnabled *bool                  `json:"isEnabled,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// TestNotificationChannelRequest represents a request to test a notification channel
type TestNotificationChannelRequest struct {
	ChannelType AlertNotificationChannel `json:"channelType" validate:"required"`
	Config      map[string]interface{}   `json:"config,omitempty"`
}

// TestNotificationChannelResponse represents the response from testing a notification channel
type TestNotificationChannelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
