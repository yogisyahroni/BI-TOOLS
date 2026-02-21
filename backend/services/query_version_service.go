package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"insight-engine-backend/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryVersionService handles query versioning operations
type QueryVersionService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
}

// NewQueryVersionService creates a new query version service
func NewQueryVersionService(db *gorm.DB, notificationSvc *NotificationService) *QueryVersionService {
	return &QueryVersionService{
		db:              db,
		notificationSvc: notificationSvc,
	}
}

// CreateVersion creates a new version for a query
func (s *QueryVersionService) CreateVersion(queryID string, userID string, req *models.QueryVersionCreateRequest) (*models.QueryVersion, error) {
	// Get query
	var query models.SavedQuery
	if err := s.db.Where("id = ?", queryID).First(&query).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("query not found")
		}
		return nil, fmt.Errorf("failed to fetch query: %w", err)
	}

	// Check permissions - only owner or admin can create versions
	if query.UserID != userID {
		canCreate, err := s.canCreateVersion(userID, queryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canCreate {
			return nil, errors.New("you do not have permission to create versions for this query")
		}
	}

	// Get next version number for this query
	var nextVersion int
	s.db.Model(&models.QueryVersion{}).Where("query_id = ?", queryID).
		Select("COALESCE(MAX(version), 0) + 1").Scan(&nextVersion)

	// Create change summary if not provided and not auto-save
	changeSummary := req.ChangeSummary
	if changeSummary == "" && !req.IsAutoSave {
		changeSummary = s.generateChangeSummary(&query, userID)
	}

	// Serialize query data
	version, err := models.NewQueryVersion(&query, userID, changeSummary, req.IsAutoSave)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}
	version.Version = nextVersion

	// Set metadata if provided
	if req.Metadata != nil {
		metadata := &models.QueryVersionMetadata{
			SQLChanged:      false, // Will be calculated from diff with previous version
			MetadataChanged: false,
			ConfigChanged:   false,
			TagsChanged:     false,
		}
		version.SetMetadata(metadata)
	}

	// Calculate diff with previous version if not the first version
	if nextVersion > 1 {
		if err := s.calculateAndStoreDiff(version, queryID); err != nil {
			LogError("query_version_diff_error", "Failed to calculate diff", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Save version
	if err := s.db.Create(version).Error; err != nil {
		LogError("query_version_create_error", "Failed to create version", map[string]interface{}{
			"query_id": queryID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	// Load relationships
	s.db.Preload("CreatedByUser").First(version, "id = ?", version.ID)

	// Clean up old auto-save versions if needed
	if req.IsAutoSave {
		if err := s.cleanupAutoSaveVersions(queryID); err != nil {
			LogError("query_auto_save_cleanup_error", "Failed to cleanup old auto-save versions", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	LogInfo("query_version_created", "Query version created successfully", map[string]interface{}{
		"version_id":   version.ID,
		"query_id":     queryID,
		"version":      nextVersion,
		"is_auto_save": req.IsAutoSave,
		"user_id":      userID,
	})

	return version, nil
}

// AutoSaveVersion automatically saves a version if there are changes
func (s *QueryVersionService) AutoSaveVersion(queryID string, userID string) (*models.QueryVersion, error) {
	// Check if there are recent auto-saves (within 5 minutes)
	var recentAutoSave models.QueryVersion
	err := s.db.Where("query_id = ? AND created_by = ? AND is_auto_save = ? AND created_at > ?",
		queryID, userID, true, time.Now().Add(-5*time.Minute)).
		Order("created_at DESC").First(&recentAutoSave).Error

	if err == nil {
		// Recent auto-save exists, check if content changed
		var query models.SavedQuery
		s.db.Where("id = ?", queryID).First(&query)

		if query.SQL == recentAutoSave.SQL && query.Name == recentAutoSave.Name {
			// No significant changes, don't create a new version
			return &recentAutoSave, nil
		}
	}

	// Create new auto-save version
	req := &models.QueryVersionCreateRequest{
		ChangeSummary: "Auto-save",
		IsAutoSave:    true,
	}
	return s.CreateVersion(queryID, userID, req)
}

// GetVersions retrieves all versions for a query with pagination
func (s *QueryVersionService) GetVersions(queryID string, userID string, filter *models.QueryVersionFilter) ([]models.QueryVersion, int64, error) {
	// Check permissions
	var query models.SavedQuery
	if err := s.db.Where("id = ?", queryID).First(&query).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("query not found")
		}
		return nil, 0, fmt.Errorf("failed to fetch query: %w", err)
	}

	// Check if user can view versions
	if query.UserID != userID {
		canView, err := s.canViewVersions(userID, queryID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canView {
			return nil, 0, errors.New("you do not have permission to view versions for this query")
		}
	}

	// Build query
	queryBuilder := s.db.Model(&models.QueryVersion{}).Where("query_id = ?", queryID)

	// Apply filters
	if filter != nil {
		if filter.IsAutoSave != nil {
			queryBuilder = queryBuilder.Where("is_auto_save = ?", *filter.IsAutoSave)
		}
		if filter.CreatedBy != nil {
			queryBuilder = queryBuilder.Where("created_by = ?", *filter.CreatedBy)
		}
	}

	// Get total count
	var total int64
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count versions: %w", err)
	}

	// Apply sorting
	orderBy := "created_at DESC"
	if filter != nil && filter.OrderBy != "" {
		switch filter.OrderBy {
		case "date_asc":
			orderBy = "created_at ASC"
		case "version_desc":
			orderBy = "version DESC"
		case "version_asc":
			orderBy = "version ASC"
		}
	}
	queryBuilder = queryBuilder.Order(orderBy)

	// Apply pagination
	if filter != nil {
		if filter.Limit > 0 {
			queryBuilder = queryBuilder.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			queryBuilder = queryBuilder.Offset(filter.Offset)
		}
	}

	// Fetch versions with user info
	var versions []models.QueryVersion
	if err := queryBuilder.Preload("CreatedByUser").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch versions: %w", err)
	}

	return versions, total, nil
}

// GetVersion retrieves a specific version
func (s *QueryVersionService) GetVersion(versionID string, userID string) (*models.QueryVersion, error) {
	var version models.QueryVersion
	if err := s.db.Preload("CreatedByUser").Preload("Query").First(&version, "id = ?", versionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("version not found")
		}
		return nil, fmt.Errorf("failed to fetch version: %w", err)
	}

	// Check permissions
	if version.Query != nil && version.Query.UserID != userID {
		canView, err := s.canViewVersions(userID, version.QueryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canView {
			return nil, errors.New("you do not have permission to view this version")
		}
	}

	return &version, nil
}

// RestoreVersion restores a query to a specific version
func (s *QueryVersionService) RestoreVersion(versionID string, userID string) (*models.QueryVersionRestoreResponse, error) {
	// Get version
	version, err := s.GetVersion(versionID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can restore
	canRestore, err := s.canRestoreVersion(userID, version.QueryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canRestore {
		return nil, errors.New("you do not have permission to restore this version")
	}

	// Get query
	var query models.SavedQuery
	if err := s.db.Where("id = ?", version.QueryID).First(&query).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch query: %w", err)
	}

	// Update query using transaction
	tx := s.db.Begin()

	// Update query fields
	updates := map[string]interface{}{
		"name":        version.Name,
		"description": version.Description,
		"sql":         version.SQL,
		"ai_prompt":   version.AIPrompt,
	}

	if err := tx.Model(&query).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update query: %w", err)
	}

	// Update visualization config
	if version.VisualizationConfig != nil {
		if err := tx.Model(&query).Update("visualization_config", version.VisualizationConfig).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update visualization config: %w", err)
		}
	}

	// Update tags
	if version.Tags != nil {
		tags, _ := version.GetTags()
		if err := tx.Model(&query).Update("tags", tags).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit restore: %w", err)
	}

	// Create a new version to record the restore action
	restoreReq := &models.QueryVersionCreateRequest{
		ChangeSummary: fmt.Sprintf("Restored from version %d", version.Version),
		IsAutoSave:    false,
		Metadata: map[string]interface{}{
			"restored_from_version":    version.Version,
			"restored_from_version_id": version.ID,
		},
	}
	if _, err := s.CreateVersion(query.ID, userID, restoreReq); err != nil {
		LogError("query_restore_version_create_error", "Failed to create restore version", map[string]interface{}{
			"error": err.Error(),
		})
	}

	LogInfo("query_version_restored", "Query version restored successfully", map[string]interface{}{
		"version_id":  version.ID,
		"query_id":    query.ID,
		"restored_to": version.Version,
		"user_id":     userID,
	})

	// Send notification
	if s.notificationSvc != nil {
		s.notificationSvc.SendNotification(
			userID,
			"Query Restored",
			fmt.Sprintf("Query '%s' was restored to version %d", query.Name, version.Version),
			"success",
			fmt.Sprintf("/queries/%s", query.ID),
			map[string]interface{}{
				"query_id":       query.ID,
				"version_id":     version.ID,
				"version_number": version.Version,
			},
		)
	}

	return &models.QueryVersionRestoreResponse{
		Success:           true,
		Message:           "Query restored successfully",
		QueryID:           query.ID,
		RestoredToVersion: version.Version,
	}, nil
}

// CompareVersions compares two versions and returns the differences
func (s *QueryVersionService) CompareVersions(versionID1, versionID2 string, userID string) (*models.QueryVersionDiff, error) {
	// Get both versions
	version1, err := s.GetVersion(versionID1, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version 1: %w", err)
	}

	version2, err := s.GetVersion(versionID2, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version 2: %w", err)
	}

	// Ensure versions are for the same query
	if version1.QueryID != version2.QueryID {
		return nil, errors.New("versions belong to different queries")
	}

	// Build diff
	diff := &models.QueryVersionDiff{
		Version1ID: versionID1,
		Version2ID: versionID2,
	}

	// Compare name
	diff.NameChanged = version1.Name != version2.Name
	if diff.NameChanged {
		diff.NameFrom = &version1.Name
		diff.NameTo = &version2.Name
	}

	// Compare description
	if (version1.Description == nil && version2.Description != nil) ||
		(version1.Description != nil && version2.Description == nil) ||
		(version1.Description != nil && version2.Description != nil && *version1.Description != *version2.Description) {
		diff.DescChanged = true
		diff.DescFrom = version1.Description
		diff.DescTo = version2.Description
	}

	// Compare SQL
	diff.SQLChanged = version1.SQL != version2.SQL
	if diff.SQLChanged {
		diff.SQLFrom = &version1.SQL
		diff.SQLTo = &version2.SQL
	}

	// Compare AI prompt
	if (version1.AIPrompt == nil && version2.AIPrompt != nil) ||
		(version1.AIPrompt != nil && version2.AIPrompt == nil) ||
		(version1.AIPrompt != nil && version2.AIPrompt != nil && *version1.AIPrompt != *version2.AIPrompt) {
		diff.AIPromptChanged = true
		diff.AIPromptFrom = version1.AIPrompt
		diff.AIPromptTo = version2.AIPrompt
	}

	// Compare visualization config
	config1, _ := version1.GetVisualizationConfig()
	config2, _ := version2.GetVisualizationConfig()
	diff.VisualizationChanged = !s.compareVisualizationConfigs(config1, config2)
	if diff.VisualizationChanged {
		diff.VisualizationFrom = config1
		diff.VisualizationTo = config2
	}

	// Compare tags
	tags1, _ := version1.GetTags()
	tags2, _ := version2.GetTags()
	diff.TagsChanged = !s.compareStringSlices(tags1, tags2)
	if diff.TagsChanged {
		diff.TagsAdded, diff.TagsRemoved = s.diffStringSlices(tags1, tags2)
	}

	return diff, nil
}

// DeleteVersion deletes a version
func (s *QueryVersionService) DeleteVersion(versionID string, userID string) error {
	// Get version
	version, err := s.GetVersion(versionID, userID)
	if err != nil {
		return err
	}

	// Check if user can delete
	canDelete, err := s.canDeleteVersion(userID, version.QueryID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canDelete {
		return errors.New("you do not have permission to delete this version")
	}

	// Delete version
	if err := s.db.Delete(version).Error; err != nil {
		LogError("query_version_delete_error", "Failed to delete version", map[string]interface{}{
			"version_id": versionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete version: %w", err)
	}

	LogInfo("query_version_deleted", "Query version deleted successfully", map[string]interface{}{
		"version_id": versionID,
		"user_id":    userID,
	})

	return nil
}

// Helper methods

func (s *QueryVersionService) generateChangeSummary(query *models.SavedQuery, userID string) string {
	// This would ideally track actual changes, for now return a generic message
	return fmt.Sprintf("Query: %s", query.Name)
}

func (s *QueryVersionService) cleanupAutoSaveVersions(queryID string) error {
	// Keep only the last 10 auto-save versions
	var count int64
	s.db.Model(&models.QueryVersion{}).Where("query_id = ? AND is_auto_save = ?", queryID, true).Count(&count)

	if count > 10 {
		// Delete oldest auto-save versions
		s.db.Where("query_id = ? AND is_auto_save = ?", queryID, true).
			Order("created_at ASC").
			Limit(int(count - 10)).
			Delete(&models.QueryVersion{})
	}

	return nil
}

func (s *QueryVersionService) calculateAndStoreDiff(version *models.QueryVersion, queryID string) error {
	// Get the previous version
	var previousVersion models.QueryVersion
	err := s.db.Where("query_id = ? AND version < ?", queryID, version.Version).
		Order("version DESC").First(&previousVersion).Error

	if err != nil {
		return nil // No previous version to compare with
	}

	// Calculate differences
	metadata := &models.QueryVersionMetadata{}

	// Check SQL changes
	if version.SQL != previousVersion.SQL {
		metadata.SQLChanged = true
		// Simple diff summary (first line of change)
		lines := strings.Split(version.SQL, "\n")
		if len(lines) > 0 {
			metadata.SQLDiffSummary = strings.TrimSpace(lines[0])
		}
	}

	// Check metadata changes
	if (version.Description == nil && previousVersion.Description != nil) ||
		(version.Description != nil && previousVersion.Description == nil) ||
		(version.Description != nil && previousVersion.Description != nil &&
			*version.Description != *previousVersion.Description) ||
		(version.AIPrompt == nil && previousVersion.AIPrompt != nil) ||
		(version.AIPrompt != nil && previousVersion.AIPrompt == nil) ||
		(version.AIPrompt != nil && previousVersion.AIPrompt != nil &&
			*version.AIPrompt != *previousVersion.AIPrompt) {
		metadata.MetadataChanged = true
	}

	// Check visualization config changes
	if string(version.VisualizationConfig) != string(previousVersion.VisualizationConfig) {
		metadata.ConfigChanged = true
	}

	// Check tags changes
	prevTags, _ := previousVersion.GetTags()
	currTags, _ := version.GetTags()
	if !s.compareStringSlices(prevTags, currTags) {
		metadata.TagsChanged = true
		metadata.TagsAdded, metadata.TagsRemoved = s.diffStringSlices(prevTags, currTags)
	}

	// Store metadata
	return version.SetMetadata(metadata)
}

func (s *QueryVersionService) compareVisualizationConfigs(config1, config2 map[string]interface{}) bool {
	if config1 == nil && config2 == nil {
		return true
	}
	if config1 == nil || config2 == nil {
		return false
	}

	// Simple comparison by marshaling to JSON
	json1, _ := json.Marshal(config1)
	json2, _ := json.Marshal(config2)
	return string(json1) == string(json2)
}

func (s *QueryVersionService) compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison
	mapA := make(map[string]bool)
	for _, s := range a {
		mapA[s] = true
	}

	for _, s := range b {
		if !mapA[s] {
			return false
		}
	}

	return true
}

func (s *QueryVersionService) diffStringSlices(old, new []string) (added, removed []string) {
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, s := range old {
		oldMap[s] = true
	}

	for _, s := range new {
		newMap[s] = true
		if !oldMap[s] {
			added = append(added, s)
		}
	}

	for _, s := range old {
		if !newMap[s] {
			removed = append(removed, s)
		}
	}

	return added, removed
}

// Permission checks

func (s *QueryVersionService) canCreateVersion(userID, queryID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *QueryVersionService) canViewVersions(userID, queryID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *QueryVersionService) canRestoreVersion(userID, queryID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *QueryVersionService) canDeleteVersion(userID, queryID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *QueryVersionService) isUserAdmin(userID string) (bool, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.IsSuperAdmin(), nil
}
