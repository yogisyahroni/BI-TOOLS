package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"insight-engine-backend/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// VersionService handles dashboard versioning operations
type VersionService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
}

// NewVersionService creates a new version service
func NewVersionService(db *gorm.DB, notificationSvc *NotificationService) *VersionService {
	return &VersionService{
		db:              db,
		notificationSvc: notificationSvc,
	}
}

// CreateVersion creates a new version for a dashboard
func (s *VersionService) CreateVersion(dashboardID string, userID string, req *models.DashboardVersionCreateRequest) (*models.DashboardVersion, error) {
	// Get dashboard with cards
	var dashboard models.Dashboard
	if err := s.db.Where("id = ?", dashboardID).Preload("Cards").First(&dashboard).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dashboard not found")
		}
		return nil, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	// Check permissions - only owner or admin can create versions
	if dashboard.UserID.String() != userID {
		canCreate, err := s.canCreateVersion(userID, dashboardID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canCreate {
			return nil, errors.New("you do not have permission to create versions for this dashboard")
		}
	}

	// Get next version number for this dashboard
	var nextVersion int
	s.db.Model(&models.DashboardVersion{}).Where("dashboard_id = ?", dashboardID).
		Select("COALESCE(MAX(version), 0) + 1").Scan(&nextVersion)

	// Create change summary if not provided and not auto-save
	changeSummary := req.ChangeSummary
	if changeSummary == "" && !req.IsAutoSave {
		changeSummary = s.generateChangeSummary(&dashboard, userID)
	}

	// Serialize dashboard data
	version, err := models.NewDashboardVersion(&dashboard, userID, changeSummary, req.IsAutoSave)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}
	version.Version = nextVersion

	// Set metadata if provided
	if req.Metadata != nil {
		metadata := &models.DashboardVersionMetadata{
			CardCount:   len(dashboard.Cards),
			FilterCount: 0,
		}
		if dashboard.Filters != nil && *dashboard.Filters != "" {
			var filters []map[string]interface{}
			if err := json.Unmarshal([]byte(*dashboard.Filters), &filters); err == nil {
				metadata.FilterCount = len(filters)
			}
		}
		version.SetMetadata(metadata)
	}

	// Save version
	if err := s.db.Create(version).Error; err != nil {
		LogError("version_create_error", "Failed to create version", map[string]interface{}{
			"dashboard_id": dashboardID,
			"user_id":      userID,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	// Load relationships
	s.db.Preload("CreatedByUser").First(version, "id = ?", version.ID)

	// Clean up old auto-save versions if needed
	if req.IsAutoSave {
		if err := s.cleanupAutoSaveVersions(dashboardID); err != nil {
			LogError("auto_save_cleanup_error", "Failed to cleanup old auto-save versions", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	LogInfo("version_created", "Dashboard version created successfully", map[string]interface{}{
		"version_id":   version.ID,
		"dashboard_id": dashboardID,
		"version":      nextVersion,
		"is_auto_save": req.IsAutoSave,
		"user_id":      userID,
	})

	return version, nil
}

// AutoSaveVersion automatically saves a version if there are changes
func (s *VersionService) AutoSaveVersion(dashboardID string, userID string) (*models.DashboardVersion, error) {
	// Check if there are recent auto-saves (within 5 minutes)
	var recentAutoSave models.DashboardVersion
	err := s.db.Where("dashboard_id = ? AND created_by = ? AND is_auto_save = ? AND created_at > ?",
		dashboardID, userID, true, time.Now().Add(-5*time.Minute)).
		Order("created_at DESC").First(&recentAutoSave).Error

	if err == nil {
		// Recent auto-save exists, check if content changed
		var dashboard models.Dashboard
		s.db.Where("id = ?", dashboardID).Preload("Cards").First(&dashboard)

		currentCardsJSON, _ := json.Marshal(dashboard.Cards)
		if string(currentCardsJSON) == recentAutoSave.CardsJSON {
			// No changes, don't create a new version
			return &recentAutoSave, nil
		}
	}

	// Create new auto-save version
	req := &models.DashboardVersionCreateRequest{
		ChangeSummary: "Auto-save",
		IsAutoSave:    true,
	}
	return s.CreateVersion(dashboardID, userID, req)
}

// GetVersions retrieves all versions for a dashboard with pagination
func (s *VersionService) GetVersions(dashboardID string, userID string, filter *models.DashboardVersionFilter) ([]models.DashboardVersion, int64, error) {
	// Check permissions
	var dashboard models.Dashboard
	if err := s.db.Where("id = ?", dashboardID).First(&dashboard).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("dashboard not found")
		}
		return nil, 0, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	// Check if user can view versions
	if dashboard.UserID.String() != userID {
		canView, err := s.canViewVersions(userID, dashboardID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canView {
			return nil, 0, errors.New("you do not have permission to view versions for this dashboard")
		}
	}

	// Build query
	query := s.db.Model(&models.DashboardVersion{}).Where("dashboard_id = ?", dashboardID)

	// Apply filters
	if filter != nil {
		if filter.IsAutoSave != nil {
			query = query.Where("is_auto_save = ?", *filter.IsAutoSave)
		}
		if filter.CreatedBy != nil {
			query = query.Where("created_by = ?", *filter.CreatedBy)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
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
	query = query.Order(orderBy)

	// Apply pagination
	if filter != nil {
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	// Fetch versions with user info
	var versions []models.DashboardVersion
	if err := query.Preload("CreatedByUser").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch versions: %w", err)
	}

	return versions, total, nil
}

// GetVersion retrieves a specific version
func (s *VersionService) GetVersion(versionID string, userID string) (*models.DashboardVersion, error) {
	var version models.DashboardVersion
	if err := s.db.Preload("CreatedByUser").Preload("Dashboard").First(&version, "id = ?", versionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("version not found")
		}
		return nil, fmt.Errorf("failed to fetch version: %w", err)
	}

	// Check permissions
	if version.Dashboard != nil && version.Dashboard.UserID.String() != userID {
		canView, err := s.canViewVersions(userID, version.DashboardID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to check permissions: %w", err)
		}
		if !canView {
			return nil, errors.New("you do not have permission to view this version")
		}
	}

	return &version, nil
}

// RestoreVersion restores a dashboard to a specific version
func (s *VersionService) RestoreVersion(versionID string, userID string) (*models.DashboardVersionRestoreResponse, error) {
	// Get version
	version, err := s.GetVersion(versionID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user can restore
	canRestore, err := s.canRestoreVersion(userID, version.DashboardID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canRestore {
		return nil, errors.New("you do not have permission to restore this version")
	}

	// Get dashboard
	var dashboard models.Dashboard
	if err := s.db.Where("id = ?", version.DashboardID).First(&dashboard).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch dashboard: %w", err)
	}

	// Parse cards from version
	var versionCards []models.DashboardVersionCard
	if err := json.Unmarshal([]byte(version.CardsJSON), &versionCards); err != nil {
		return nil, fmt.Errorf("failed to parse version cards: %w", err)
	}

	// Update dashboard using transaction
	tx := s.db.Begin()

	// Update dashboard metadata
	updates := map[string]interface{}{
		"name":        version.Name,
		"description": version.Description,
		"filters":     version.FiltersJSON,
	}

	if err := tx.Model(&dashboard).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update dashboard: %w", err)
	}

	// Delete existing cards
	if err := tx.Where("dashboard_id = ?", dashboard.ID).Delete(&models.DashboardCard{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing cards: %w", err)
	}

	// Create new cards from version
	for _, cardData := range versionCards {
		card := models.DashboardCard{
			DashboardID: dashboard.ID,
		}

		if cardData.ID != "" {
			if parsedID, err := uuid.Parse(cardData.ID); err == nil {
				card.ID = parsedID
			}
		}
		// Fix: Parse QueryID from *string to *uuid.UUID
		if cardData.QueryID != nil {
			if parsedQueryID, err := uuid.Parse(*cardData.QueryID); err == nil {
				card.QueryID = &parsedQueryID
			}
		}
		if cardData.Title != nil {
			card.Title = cardData.Title
		}
		if cardData.Position != nil {
			card.Position = datatypes.JSON(cardData.Position)
		}
		if cardData.VisualizationConfig != nil {
			card.VisualizationConfig = datatypes.JSON(cardData.VisualizationConfig)
		}

		if err := tx.Create(&card).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create card: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit restore: %w", err)
	}

	// Create a new version to record the restore action
	restoreReq := &models.DashboardVersionCreateRequest{
		ChangeSummary: fmt.Sprintf("Restored from version %d", version.Version),
		IsAutoSave:    false,
		Metadata: map[string]interface{}{
			"restored_from_version":    version.Version,
			"restored_from_version_id": version.ID,
		},
	}
	// Fixed: dashboard.ID is UUID, need string
	if _, err := s.CreateVersion(dashboard.ID.String(), userID, restoreReq); err != nil {
		LogError("restore_version_create_error", "Failed to create restore version", map[string]interface{}{
			"error": err.Error(),
		})
	}

	LogInfo("version_restored", "Dashboard version restored successfully", map[string]interface{}{
		"version_id":   version.ID,
		"dashboard_id": dashboard.ID,
		"restored_to":  version.Version,
		"user_id":      userID,
	})

	// Send notification
	if s.notificationSvc != nil {
		s.notificationSvc.SendNotification(
			userID,
			"Dashboard Restored",
			fmt.Sprintf("Dashboard '%s' was restored to version %d", dashboard.Name, version.Version),
			"success",
			fmt.Sprintf("/dashboards/%s", dashboard.ID),
			map[string]interface{}{
				"dashboard_id":   dashboard.ID,
				"version_id":     version.ID,
				"version_number": version.Version,
			},
		)
	}

	return &models.DashboardVersionRestoreResponse{
		Success:           true,
		Message:           "Dashboard restored successfully",
		DashboardID:       dashboard.ID.String(),
		RestoredToVersion: version.Version,
	}, nil
}

// CompareVersions compares two versions and returns the differences
func (s *VersionService) CompareVersions(versionID1, versionID2 string, userID string) (*models.DashboardVersionDiff, error) {
	// Get both versions
	version1, err := s.GetVersion(versionID1, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version 1: %w", err)
	}

	version2, err := s.GetVersion(versionID2, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version 2: %w", err)
	}

	// Ensure versions are for the same dashboard
	if version1.DashboardID != version2.DashboardID {
		return nil, errors.New("versions belong to different dashboards")
	}

	// Build diff
	diff := &models.DashboardVersionDiff{
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

	// Compare filters
	if (version1.FiltersJSON == nil && version2.FiltersJSON != nil) ||
		(version1.FiltersJSON != nil && version2.FiltersJSON == nil) ||
		(version1.FiltersJSON != nil && version2.FiltersJSON != nil && *version1.FiltersJSON != *version2.FiltersJSON) {
		diff.FiltersChanged = true
		diff.FiltersFrom = version1.FiltersJSON
		diff.FiltersTo = version2.FiltersJSON
	}

	// Compare cards
	cards1, _ := version1.GetCards()
	cards2, _ := version2.GetCards()
	diff.CardsDiff = s.compareCards(cards1, cards2)

	return diff, nil
}

// DeleteVersion deletes a version
func (s *VersionService) DeleteVersion(versionID string, userID string) error {
	// Get version
	version, err := s.GetVersion(versionID, userID)
	if err != nil {
		return err
	}

	// Check if user can delete
	canDelete, err := s.canDeleteVersion(userID, version.DashboardID.String())
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !canDelete {
		return errors.New("you do not have permission to delete this version")
	}

	// Delete version
	if err := s.db.Delete(version).Error; err != nil {
		LogError("version_delete_error", "Failed to delete version", map[string]interface{}{
			"version_id": versionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete version: %w", err)
	}

	LogInfo("version_deleted", "Dashboard version deleted successfully", map[string]interface{}{
		"version_id": versionID,
		"user_id":    userID,
	})

	return nil
}

// Helper methods

func (s *VersionService) generateChangeSummary(dashboard *models.Dashboard, userID string) string {
	// This would ideally track actual changes, for now return a generic message
	cardCount := len(dashboard.Cards)
	return fmt.Sprintf("Dashboard with %d cards", cardCount)
}

func (s *VersionService) cleanupAutoSaveVersions(dashboardID string) error {
	// Keep only the last 10 auto-save versions
	var count int64
	s.db.Model(&models.DashboardVersion{}).Where("dashboard_id = ? AND is_auto_save = ?", dashboardID, true).Count(&count)

	if count > 10 {
		// Delete oldest auto-save versions
		s.db.Where("dashboard_id = ? AND is_auto_save = ?", dashboardID, true).
			Order("created_at ASC").
			Limit(int(count - 10)).
			Delete(&models.DashboardVersion{})
	}

	return nil
}

func (s *VersionService) compareCards(cards1, cards2 []models.DashboardVersionCard) models.DashboardCardsDiff {
	result := models.DashboardCardsDiff{
		Added:     []models.DashboardVersionCard{},
		Removed:   []models.DashboardVersionCard{},
		Modified:  []models.DashboardCardChange{},
		Unchanged: []models.DashboardVersionCard{},
	}

	// Build maps for comparison
	cardMap1 := make(map[string]models.DashboardVersionCard)
	cardMap2 := make(map[string]models.DashboardVersionCard)

	for _, card := range cards1 {
		if card.ID != "" {
			cardMap1[card.ID] = card
		}
	}

	for _, card := range cards2 {
		if card.ID != "" {
			cardMap2[card.ID] = card
		}
	}

	// Find added, removed, and modified
	for id, card := range cardMap2 {
		if oldCard, exists := cardMap1[id]; exists {
			// Card exists in both, check if modified
			changes := s.getCardChanges(oldCard, card)
			if len(changes) > 0 {
				result.Modified = append(result.Modified, models.DashboardCardChange{
					Before:  oldCard,
					After:   card,
					Changes: changes,
				})
			} else {
				result.Unchanged = append(result.Unchanged, card)
			}
		} else {
			// Card only in version 2 (added)
			result.Added = append(result.Added, card)
		}
	}

	// Find removed (in version 1 but not in version 2)
	for id, card := range cardMap1 {
		if _, exists := cardMap2[id]; !exists {
			result.Removed = append(result.Removed, card)
		}
	}

	return result
}

func (s *VersionService) getCardChanges(card1, card2 models.DashboardVersionCard) []string {
	var changes []string

	// Compare title
	if (card1.Title == nil && card2.Title != nil) ||
		(card1.Title != nil && card2.Title == nil) ||
		(card1.Title != nil && card2.Title != nil && *card1.Title != *card2.Title) {
		changes = append(changes, "title")
	}

	// Compare query ID
	if (card1.QueryID == nil && card2.QueryID != nil) ||
		(card1.QueryID != nil && card2.QueryID == nil) ||
		(card1.QueryID != nil && card2.QueryID != nil && *card1.QueryID != *card2.QueryID) {
		changes = append(changes, "query")
	}

	// Compare position
	if string(card1.Position) != string(card2.Position) {
		changes = append(changes, "position")
	}

	// Compare visualization config
	if string(card1.VisualizationConfig) != string(card2.VisualizationConfig) {
		changes = append(changes, "visualization")
	}

	return changes
}

// Permission checks

func (s *VersionService) canCreateVersion(userID, dashboardID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *VersionService) canViewVersions(userID, dashboardID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *VersionService) canRestoreVersion(userID, dashboardID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *VersionService) canDeleteVersion(userID, dashboardID string) (bool, error) {
	return s.isUserAdmin(userID)
}

func (s *VersionService) isUserAdmin(userID string) (bool, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.IsSuperAdmin(), nil
}
