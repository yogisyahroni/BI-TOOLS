package services

import (
	"fmt"
	"insight-engine-backend/models"

	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService handles notification operations
type NotificationService struct {
	db           *gorm.DB
	wsHub        *WebSocketHub
	slackService *SlackService
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, wsHub *WebSocketHub, slackService *SlackService) *NotificationService {
	return &NotificationService{
		db:           db,
		wsHub:        wsHub,
		slackService: slackService,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Push notification via WebSocket if user is connected
	if s.wsHub.IsUserConnected(notification.UserID) {
		s.wsHub.BroadcastToUser(notification.UserID, "notification", notification)
	}

	return nil
}

// SendNotification creates and sends a notification to a user
func (s *NotificationService) SendNotification(userID string, title, message, notifType, link string, metadata map[string]interface{}) error {
	notification := &models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		Link:    link,
	}

	// Convert metadata to JSONB if provided
	if metadata != nil {
		// GORM will handle the conversion to datatypes.JSON
		// notification.Metadata = metadata
	}

	// Persist and socket push
	if err := s.CreateNotification(notification); err != nil {
		return err
	}

	// Check if we should send to Slack based on type or metadata
	// For now, we check if metadata has a "slack_channel" key or if it's a critical alert
	if s.slackService != nil {
		var slackChannel string
		if metadata != nil {
			if sc, ok := metadata["slack_channel"].(string); ok {
				slackChannel = sc
			}
		}

		// Also auto-send critical alerts to default channel if configured
		if notifType == "alert_critical" || slackChannel != "" {
			// Construct attachment
			attachment := SlackAttachment{
				Title:     title,
				Text:      message,
				TitleLink: link,
				Ts:        time.Now().Unix(),
				Footer:    "InsightEngine AI Notification",
			}

			// Set color based on type
			if notifType == "alert_critical" {
				attachment.Color = "#ff0000"
			} else if notifType == "alert_warning" {
				attachment.Color = "#ffcc00"
			} else {
				attachment.Color = "#36a64f"
			}

			// Run in background to not block main thread
			go func() {
				if err := s.slackService.SendNotification(slackChannel, "", []SlackAttachment{attachment}); err != nil {
					fmt.Printf("Failed to send slack notification: %v\n", err)
				}
			}()
		}
	}

	return nil
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(userID string, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	// Get total count
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated notifications
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(userID string, limit int) ([]models.Notification, error) {
	var notifications []models.Notification

	if err := s.db.Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID string) (int64, error) {
	var count int64

	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID uuid.UUID, userID string) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID string) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID uuid.UUID, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}

	return nil
}

// DeleteReadNotifications deletes all read notifications for a user
func (s *NotificationService) DeleteReadNotifications(userID string) error {
	if err := s.db.Where("user_id = ? AND is_read = ?", userID, true).
		Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("failed to delete read notifications: %w", err)
	}

	return nil
}

// GetNotificationsByType retrieves notifications by type for a user
func (s *NotificationService) GetNotificationsByType(userID string, notifType string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification

	if err := s.db.Where("user_id = ? AND type = ?", userID, notifType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}

	return notifications, nil
}

// BroadcastSystemNotification sends a notification to all connected users
func (s *NotificationService) BroadcastSystemNotification(title, message, notifType string) error {
	// Get all connected users
	connectedUsers := s.wsHub.GetConnectedUsers()

	for _, userIDStr := range connectedUsers {
		// userID is already string
		userID := userIDStr

		notification := &models.Notification{
			UserID:  userID,
			Title:   title,
			Message: message,
			Type:    notifType,
		}

		// Create notification in DB
		if err := s.db.Create(notification).Error; err != nil {
			continue
		}

		// Push via WebSocket
		s.wsHub.BroadcastToUser(userIDStr, "notification", notification)
	}

	return nil
}
