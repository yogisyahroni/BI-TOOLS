package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"insight-engine-backend/models"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentService handles comment operations
type CommentService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
}

// NewCommentService creates a new comment service
func NewCommentService(db *gorm.DB, notificationSvc *NotificationService) *CommentService {
	return &CommentService{
		db:              db,
		notificationSvc: notificationSvc,
	}
}

// mentionRegex matches @username patterns in text
var mentionRegex = regexp.MustCompile(`@(\w+)`)

// ExtractMentions extracts @username patterns from content and returns usernames
func ExtractMentions(content string) []string {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	usernames := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			username := strings.ToLower(match[1])
			if !seen[username] {
				usernames = append(usernames, username)
				seen[username] = true
			}
		}
	}

	return usernames
}

// CreateComment creates a new comment with mention support
func (s *CommentService) CreateComment(userID string, req *models.CommentCreateRequest) (*models.Comment, error) {
	// Validate entity type
	entityType, valid := models.ValidateEntityType(req.EntityType)
	if !valid {
		return nil, errors.New("invalid entity type")
	}

	// Check permissions
	if err := s.checkEntityAccess(userID, entityType, req.EntityID); err != nil {
		return nil, err
	}

	// If it's a reply, validate parent comment
	if req.ParentID != nil && *req.ParentID != "" {
		var parentComment models.Comment
		if err := s.db.Where("id = ?", *req.ParentID).First(&parentComment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent comment not found")
			}
			return nil, fmt.Errorf("failed to fetch parent comment: %w", err)
		}

		// Ensure parent is on the same entity
		if parentComment.EntityType != entityType || parentComment.EntityID != req.EntityID {
			return nil, errors.New("parent comment belongs to different entity")
		}

		// Prevent nested replies beyond 1 level (parent must be top-level)
		if parentComment.IsReply() {
			return nil, errors.New("cannot reply to a reply - only one level of nesting allowed")
		}
	}

	// Extract mentions from content
	mentionedUsernames := ExtractMentions(req.Content)
	mentionedUserIDs := make([]string, 0)

	// Lookup mentioned users
	if len(mentionedUsernames) > 0 {
		var mentionedUsers []models.User
		if err := s.db.Where("LOWER(username) IN ?", mentionedUsernames).Find(&mentionedUsers).Error; err != nil {
			LogError("mention_lookup_error", "Failed to lookup mentioned users", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, user := range mentionedUsers {
				mentionedUserIDs = append(mentionedUserIDs, user.ID.String())
			}
		}
	}

	// Create comment
	comment := models.Comment{
		ID:         uuid.New().String(),
		EntityType: entityType,
		EntityID:   req.EntityID,
		UserID:     userID,
		Content:    req.Content,
		ParentID:   req.ParentID,
		IsResolved: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Set mentions
	if err := comment.SetMentionedUserIDs(mentionedUserIDs); err != nil {
		LogError("mention_set_error", "Failed to set mentions", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := s.db.Create(&comment).Error; err != nil {
		LogError("comment_create_error", "Failed to create comment", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Load user relationship
	s.db.Preload("User").First(&comment, "id = ?", comment.ID)

	// Create notifications for mentioned users
	s.createMentionNotifications(&comment, mentionedUserIDs, userID)

	LogInfo("comment_created", "Comment created successfully", map[string]interface{}{
		"comment_id":    comment.ID,
		"user_id":       userID,
		"entity_type":   entityType,
		"entity_id":     req.EntityID,
		"mention_count": len(mentionedUserIDs),
	})

	return &comment, nil
}

// GetCommentByID retrieves a comment by ID with all relationships
func (s *CommentService) GetCommentByID(commentID string) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Preload("User").
		Preload("Parent.User").
		Preload("Replies.User").
		Preload("Annotation").
		First(&comment, "id = ?", commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, fmt.Errorf("failed to fetch comment: %w", err)
	}

	return &comment, nil
}

// GetCommentsByEntity retrieves comments for an entity with pagination and filtering
func (s *CommentService) GetCommentsByEntity(userID string, filter *models.CommentFilter) ([]models.Comment, int64, error) {
	// Check access to entity
	if filter.EntityType != nil && filter.EntityID != nil {
		if err := s.checkEntityAccess(userID, *filter.EntityType, *filter.EntityID); err != nil {
			return nil, 0, err
		}
	}

	query := s.db.Model(&models.Comment{})

	// Apply filters
	if filter.EntityType != nil {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}
	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", *filter.EntityID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.IsResolved != nil {
		query = query.Where("is_resolved = ?", *filter.IsResolved)
	}

	// Handle parent filter
	if filter.ParentID != nil {
		if *filter.ParentID == "" {
			// Top-level comments only
			query = query.Where("parent_id IS NULL")
		} else if *filter.ParentID != "*" {
			// Specific parent
			query = query.Where("parent_id = ?", *filter.ParentID)
		}
		// "*" means all comments including replies
	} else {
		// Default: top-level only
		query = query.Where("parent_id IS NULL")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Apply sorting
	sortField := "created_at"
	if filter.SortBy == "popular" {
		// For popularity, we'd need a more complex query with subquery
		// For now, use created_at as fallback
		sortField = "created_at"
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Fetch comments with relationships
	var comments []models.Comment
	if err := query.Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Preload("User")
		}).
		Preload("Annotation").
		Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch comments: %w", err)
	}

	return comments, total, nil
}

// UpdateComment updates a comment (only by owner)
func (s *CommentService) UpdateComment(commentID, userID string, content string) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch comment: %w", err)
	}

	// Extract new mentions
	mentionedUsernames := ExtractMentions(content)
	mentionedUserIDs := make([]string, 0)

	if len(mentionedUsernames) > 0 {
		var mentionedUsers []models.User
		if err := s.db.Where("LOWER(username) IN ?", mentionedUsernames).Find(&mentionedUsers).Error; err != nil {
			LogError("mention_lookup_error", "Failed to lookup mentioned users", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, user := range mentionedUsers {
				mentionedUserIDs = append(mentionedUserIDs, user.ID.String())
			}
		}
	}

	// Get old mentions for notification comparison
	oldMentions := comment.GetMentionedUserIDs()
	oldMentionMap := make(map[string]bool)
	for _, id := range oldMentions {
		oldMentionMap[id] = true
	}

	// Update comment
	comment.Content = content
	comment.UpdatedAt = time.Now()
	if err := comment.SetMentionedUserIDs(mentionedUserIDs); err != nil {
		LogError("mention_set_error", "Failed to set mentions", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := s.db.Save(&comment).Error; err != nil {
		LogError("comment_update_error", "Failed to update comment", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID,
		})
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	// Notify only new mentions
	newMentions := make([]string, 0)
	for _, id := range mentionedUserIDs {
		if !oldMentionMap[id] {
			newMentions = append(newMentions, id)
		}
	}
	s.createMentionNotifications(&comment, newMentions, userID)

	// Reload with relationships
	s.db.Preload("User").
		Preload("Replies.User").
		Preload("Annotation").
		First(&comment, "id = ?", comment.ID)

	LogInfo("comment_updated", "Comment updated successfully", map[string]interface{}{
		"comment_id": commentID,
		"user_id":    userID,
	})

	return &comment, nil
}

// DeleteComment deletes a comment (only by owner or admin)
func (s *CommentService) DeleteComment(commentID, userID string) error {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("comment not found")
		}
		return fmt.Errorf("failed to fetch comment: %w", err)
	}

	// Check if user can delete (owner or admin)
	canDelete := comment.UserID == userID
	if !canDelete {
		isAdmin, err := s.isUserAdmin(userID)
		if err != nil || !isAdmin {
			return errors.New("you do not have permission to delete this comment")
		}
	}

	// Delete associated annotation if exists
	if err := s.db.Where("comment_id = ?", commentID).Delete(&models.Annotation{}).Error; err != nil {
		LogError("annotation_delete_error", "Failed to delete annotation", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID,
		})
	}

	// Delete the comment (replies will be cascaded if FK constraint has ON DELETE CASCADE)
	if err := s.db.Delete(&comment).Error; err != nil {
		LogError("comment_delete_error", "Failed to delete comment", map[string]interface{}{
			"error":      err.Error(),
			"comment_id": commentID,
		})
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	LogInfo("comment_deleted", "Comment deleted successfully", map[string]interface{}{
		"comment_id": commentID,
		"user_id":    userID,
	})

	return nil
}

// ResolveComment marks a comment as resolved or unresolved
func (s *CommentService) ResolveComment(commentID, userID string, isResolved bool) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, fmt.Errorf("failed to fetch comment: %w", err)
	}

	// Check permissions - owner, entity owner, or admin
	canResolve, err := s.canResolveComment(userID, &comment)
	if err != nil {
		return nil, err
	}
	if !canResolve {
		return nil, errors.New("you do not have permission to resolve this comment")
	}

	// Update resolved status
	comment.IsResolved = isResolved
	comment.UpdatedAt = time.Now()

	if err := s.db.Save(&comment).Error; err != nil {
		LogError("comment_resolve_error", "Failed to update comment resolve status", map[string]interface{}{
			"error":       err.Error(),
			"comment_id":  commentID,
			"is_resolved": isResolved,
		})
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	// Reload with relationships
	s.db.Preload("User").
		Preload("Replies.User").
		Preload("Annotation").
		First(&comment, "id = ?", commentID)

	action := "unresolved"
	if isResolved {
		action = "resolved"
	}

	LogInfo("comment_resolved", fmt.Sprintf("Comment %s", action), map[string]interface{}{
		"comment_id": commentID,
		"user_id":    userID,
	})

	return &comment, nil
}

// GetReplies gets all replies for a parent comment
func (s *CommentService) GetReplies(parentID string, userID string) ([]models.Comment, error) {
	// First verify parent exists
	var parent models.Comment
	if err := s.db.Where("id = ?", parentID).First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parent comment not found")
		}
		return nil, fmt.Errorf("failed to fetch parent comment: %w", err)
	}

	// Check access
	if err := s.checkEntityAccess(userID, parent.EntityType, parent.EntityID); err != nil {
		return nil, err
	}

	var replies []models.Comment
	if err := s.db.Where("parent_id = ?", parentID).
		Preload("User").
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch replies: %w", err)
	}

	return replies, nil
}

// SearchMentions searches for users to mention
func (s *CommentService) SearchMentions(query string, excludeUserID string, limit int) ([]models.User, error) {
	if query == "" {
		return []models.User{}, nil
	}

	var users []models.User
	dbQuery := s.db.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ?)",
		"%"+strings.ToLower(query)+"%",
		"%"+strings.ToLower(query)+"%",
		"%"+strings.ToLower(query)+"%").
		Where("id != ?", excludeUserID).
		Limit(limit).
		Find(&users)

	if err := dbQuery.Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetRecentMentions gets recently mentioned users by a specific user
func (s *CommentService) GetRecentMentions(userID string, limit int) ([]models.User, error) {
	// Get distinct mentioned users from recent comments by this user
	var mentionedUserIDs []string
	if err := s.db.Model(&models.Comment{}).
		Where("user_id = ?", userID).
		Where("mentions IS NOT NULL").
		Order("created_at DESC").
		Limit(50).
		Pluck("mentions", &mentionedUserIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch recent mentions: %w", err)
	}

	// Parse JSON arrays and collect unique user IDs
	seen := make(map[string]bool)
	uniqueIDs := make([]string, 0)
	for _, mentionsJSON := range mentionedUserIDs {
		var ids []string
		if err := json.Unmarshal([]byte(mentionsJSON), &ids); err != nil {
			continue
		}
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				uniqueIDs = append(uniqueIDs, id)
			}
		}
	}

	if len(uniqueIDs) == 0 {
		return []models.User{}, nil
	}

	// Fetch user details
	var users []models.User
	if err := s.db.Where("id IN ?", uniqueIDs).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch mentioned users: %w", err)
	}

	return users, nil
}

// CreateAnnotation creates a new annotation with an associated comment
func (s *CommentService) CreateAnnotation(userID string, req *models.AnnotationCreateRequest) (*models.Comment, error) {
	// Validate entity type for chart
	entityType := models.EntityTypeChart

	// Check permissions
	if err := s.checkEntityAccess(userID, entityType, req.ChartID); err != nil {
		return nil, err
	}

	// Extract mentions from content
	mentionedUsernames := ExtractMentions(req.Content)
	mentionedUserIDs := make([]string, 0)

	if len(mentionedUsernames) > 0 {
		var mentionedUsers []models.User
		if err := s.db.Where("LOWER(username) IN ?", mentionedUsernames).Find(&mentionedUsers).Error; err != nil {
			LogError("mention_lookup_error", "Failed to lookup mentioned users", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, user := range mentionedUsers {
				mentionedUserIDs = append(mentionedUserIDs, user.ID.String())
			}
		}
	}

	// Create comment first
	comment := models.Comment{
		ID:         uuid.New().String(),
		EntityType: entityType,
		EntityID:   req.ChartID,
		UserID:     userID,
		Content:    req.Content,
		ParentID:   nil,
		IsResolved: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := comment.SetMentionedUserIDs(mentionedUserIDs); err != nil {
		LogError("mention_set_error", "Failed to set mentions", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create annotation
	positionJSON, _ := json.Marshal(req.Position)
	annotation := models.Annotation{
		ID:        uuid.New().String(),
		CommentID: comment.ID,
		ChartID:   req.ChartID,
		XValue:    req.XValue,
		YValue:    req.YValue,
		XCategory: req.XCategory,
		YCategory: req.YCategory,
		Position:  positionJSON,
		Type:      req.Type,
		Color:     req.Color,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if annotation.Color == "" {
		annotation.Color = "#F59E0B"
	}

	// Start transaction
	tx := s.db.Begin()

	// Create comment
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		LogError("comment_create_error", "Failed to create annotation comment", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Create annotation
	if err := tx.Create(&annotation).Error; err != nil {
		tx.Rollback()
		LogError("annotation_create_error", "Failed to create annotation", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to create annotation: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load relationships
	s.db.Preload("User").
		Preload("Annotation").
		First(&comment, "id = ?", comment.ID)

	// Create notifications
	s.createMentionNotifications(&comment, mentionedUserIDs, userID)

	LogInfo("annotation_created", "Annotation created successfully", map[string]interface{}{
		"comment_id":    comment.ID,
		"annotation_id": annotation.ID,
		"user_id":       userID,
		"chart_id":      req.ChartID,
	})

	return &comment, nil
}

// GetAnnotationsByChart retrieves all annotations for a chart
func (s *CommentService) GetAnnotationsByChart(chartID string, userID string) ([]models.Comment, error) {
	// Check access
	if err := s.checkEntityAccess(userID, models.EntityTypeChart, chartID); err != nil {
		return nil, err
	}

	var comments []models.Comment
	if err := s.db.Where("entity_type = ? AND entity_id = ?", models.EntityTypeChart, chartID).
		Where("EXISTS (SELECT 1 FROM annotations WHERE annotations.comment_id = comments.id)").
		Preload("User").
		Preload("Annotation").
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch annotations: %w", err)
	}

	return comments, nil
}

// UpdateAnnotation updates an annotation
func (s *CommentService) UpdateAnnotation(annotationID, userID string, req *models.AnnotationCreateRequest) (*models.Comment, error) {
	// Find annotation
	var annotation models.Annotation
	if err := s.db.Where("id = ?", annotationID).First(&annotation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("annotation not found")
		}
		return nil, fmt.Errorf("failed to fetch annotation: %w", err)
	}

	// Find associated comment
	var comment models.Comment
	if err := s.db.Where("id = ?", annotation.CommentID).First(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch associated comment: %w", err)
	}

	// Find the card to get its dashboard ID
	var card models.DashboardCard
	if err := s.db.Where("id = ?", annotation.ChartID).First(&card).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch associated card: %w", err)
	}

	// Find the dashboard to check its owner
	var dashboard models.Dashboard
	if err := s.db.Where("id = ?", card.DashboardID).First(&dashboard).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch associated dashboard: %w", err)
	}

	// Check ownership
	canUpdate := comment.UserID == userID
	// Check ownership via dashboard
	if dashboard.UserID.String() == userID {
		canUpdate = true
	}

	if !canUpdate {
		isAdmin, err := s.isUserAdmin(userID)
		if err != nil || !isAdmin {
			return nil, errors.New("you do not have permission to update this annotation")
		}
	}

	// Update annotation
	annotation.XValue = req.XValue
	annotation.YValue = req.YValue
	annotation.XCategory = req.XCategory
	annotation.YCategory = req.YCategory
	positionJSON, _ := json.Marshal(req.Position)
	annotation.Position = positionJSON
	annotation.Type = req.Type
	if req.Color != "" {
		annotation.Color = req.Color
	}
	annotation.UpdatedAt = time.Now()

	// Update comment content
	comment.Content = req.Content
	comment.UpdatedAt = time.Now()

	tx := s.db.Begin()

	if err := tx.Save(&annotation).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update annotation: %w", err)
	}

	if err := tx.Save(&comment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload with relationships
	s.db.Preload("User").
		Preload("Annotation").
		First(&comment, "id = ?", comment.ID)

	return &comment, nil
}

// DeleteAnnotation deletes an annotation and its comment
func (s *CommentService) DeleteAnnotation(annotationID, userID string) error {
	// Find annotation
	var annotation models.Annotation
	if err := s.db.Where("id = ?", annotationID).First(&annotation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("annotation not found")
		}
		return fmt.Errorf("failed to fetch annotation: %w", err)
	}

	// Find associated comment
	var comment models.Comment
	if err := s.db.Where("id = ?", annotation.CommentID).First(&comment).Error; err != nil {
		return fmt.Errorf("failed to fetch associated comment: %w", err)
	}

	// Check ownership
	canDelete := comment.UserID == userID
	if !canDelete {
		isAdmin, err := s.isUserAdmin(userID)
		if err != nil || !isAdmin {
			return errors.New("you do not have permission to delete this annotation")
		}
	}

	tx := s.db.Begin()

	// Delete annotation
	if err := tx.Delete(&annotation).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete annotation: %w", err)
	}

	// Delete comment
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	LogInfo("annotation_deleted", "Annotation deleted successfully", map[string]interface{}{
		"annotation_id": annotationID,
		"comment_id":    comment.ID,
		"user_id":       userID,
	})

	return nil
}

// Helper methods

func (s *CommentService) createMentionNotifications(comment *models.Comment, mentionedUserIDs []string, authorID string) {
	if s.notificationSvc == nil || len(mentionedUserIDs) == 0 {
		return
	}

	// Get author info
	var author models.User
	if err := s.db.Where("id = ?", authorID).First(&author).Error; err != nil {
		author.Name = "Someone"
	}

	// Get entity info for link
	entityLink := s.buildEntityLink(comment.EntityType, comment.EntityID)

	for _, mentionedUserID := range mentionedUserIDs {
		// Skip if user mentions themselves
		if mentionedUserID == authorID {
			continue
		}

		notifType := "mention"
		if comment.IsReply() {
			notifType = "mention_reply"
		}

		err := s.notificationSvc.SendNotification(
			mentionedUserID,
			fmt.Sprintf("%s mentioned you", author.Name),
			fmt.Sprintf("%s mentioned you in a comment on %s", author.Name, comment.EntityType),
			notifType,
			entityLink,
			map[string]interface{}{
				"comment_id":  comment.ID,
				"entity_type": comment.EntityType,
				"entity_id":   comment.EntityID,
				"author_id":   authorID,
			},
		)

		if err != nil {
			LogError("mention_notification_error", "Failed to send mention notification", map[string]interface{}{
				"error":     err.Error(),
				"user_id":   mentionedUserID,
				"author_id": authorID,
			})
		}
	}
}

func (s *CommentService) buildEntityLink(entityType models.CommentEntityType, entityID string) string {
	switch entityType {
	case models.EntityTypeDashboard:
		return fmt.Sprintf("/dashboards/%s", entityID)
	case models.EntityTypeQuery:
		return fmt.Sprintf("/queries/%s", entityID)
	case models.EntityTypeChart:
		return fmt.Sprintf("/charts/%s", entityID)
	case models.EntityTypePipeline:
		return fmt.Sprintf("/pipelines/%s", entityID)
	case models.EntityTypeDataflow:
		return fmt.Sprintf("/dataflows/%s", entityID)
	case models.EntityTypeCollection:
		return fmt.Sprintf("/collections/%s", entityID)
	default:
		return "/"
	}
}

func (s *CommentService) checkEntityAccess(userID string, entityType models.CommentEntityType, entityID string) error {
	// This is a simplified check - in production, you'd check actual permissions
	// For now, we assume if the user can view the entity, they can comment
	switch entityType {
	case models.EntityTypeDashboard:
		var count int64
		if err := s.db.Model(&models.Dashboard{}).Where("id = ?", entityID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("dashboard not found")
		}
	case models.EntityTypeQuery:
		var count int64
		if err := s.db.Model(&models.SavedQuery{}).Where("id = ?", entityID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("query not found")
		}
	case models.EntityTypeChart:
		// Charts are typically part of dashboards, for now just check if chart ID exists
		// In a real implementation, you'd check the chart's existence
		return nil
	case models.EntityTypePipeline:
		var count int64
		if err := s.db.Model(&models.Pipeline{}).Where("id = ?", entityID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("pipeline not found")
		}
	case models.EntityTypeDataflow:
		var count int64
		if err := s.db.Model(&models.Dataflow{}).Where("id = ?", entityID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("dataflow not found")
		}
	case models.EntityTypeCollection:
		var count int64
		if err := s.db.Model(&models.Collection{}).Where("id = ?", entityID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("collection not found")
		}
	default:
		return errors.New("unsupported entity type")
	}

	return nil
}

func (s *CommentService) canResolveComment(userID string, comment *models.Comment) (bool, error) {
	// Comment owner can resolve
	if comment.UserID == userID {
		return true, nil
	}

	// Check if user is admin
	isAdmin, err := s.isUserAdmin(userID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	// Check if user owns the entity
	switch comment.EntityType {
	case models.EntityTypeDashboard:
		var dashboard models.Dashboard
		if err := s.db.Where("id = ?", comment.EntityID).First(&dashboard).Error; err != nil {
			return false, err
		}
		if dashboard.UserID.String() == userID {
			return true, nil
		}
	case models.EntityTypeQuery:
		var query models.SavedQuery
		if err := s.db.Where("id = ?", comment.EntityID).First(&query).Error; err != nil {
			return false, err
		}
		if query.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (s *CommentService) isUserAdmin(userID string) (bool, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.IsSuperAdmin(), nil
}
