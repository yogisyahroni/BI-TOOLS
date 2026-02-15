package services

import (
	"fmt"
	"insight-engine-backend/models"
	"strings"

	"gorm.io/gorm"
)

type DataGovernanceService struct {
	DB *gorm.DB
}

func NewDataGovernanceService(db *gorm.DB) *DataGovernanceService {
	return &DataGovernanceService{DB: db}
}

// === Data Classification ===

func (s *DataGovernanceService) GetDataClassifications() ([]models.DataClassification, error) {
	var classifications []models.DataClassification
	err := s.DB.Find(&classifications).Error
	return classifications, err
}

func (s *DataGovernanceService) CreateDataClassification(c *models.DataClassification) error {
	return s.DB.Create(c).Error
}

// === Column Metadata ===

func (s *DataGovernanceService) GetColumnMetadata(datasourceID, tableName string) ([]models.ColumnMetadata, error) {
	var metadata []models.ColumnMetadata
	query := s.DB.Where("datasource_id = ?", datasourceID)

	if tableName != "" {
		query = query.Where("table_name = ?", tableName)
	}

	err := query.Preload("DataClassification").Find(&metadata).Error
	return metadata, err
}

func (s *DataGovernanceService) UpdateColumnMetadata(datasourceID, tableName, columnName string, updates map[string]interface{}) error {
	// Upsert based on unique key (datasource, table, column)
	var existing models.ColumnMetadata
	err := s.DB.Where("datasource_id = ? AND table_name = ? AND column_name = ?",
		datasourceID, tableName, columnName).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		// Ensure required fields are set in updates
		updates["datasource_id"] = datasourceID
		updates["table_name"] = tableName
		updates["column_name"] = columnName
		return s.DB.Model(&models.ColumnMetadata{}).Create(updates).Error
	} else if err != nil {
		return err
	}

	// Update existing
	// GORM Updates with map supports updating zero values (like empty string)
	return s.DB.Model(&existing).Updates(updates).Error
}

// === Column Permissions ===

func (s *DataGovernanceService) GetColumnPermissions(roleID uint) ([]models.ColumnPermission, error) {
	var permissions []models.ColumnPermission
	err := s.DB.Where("role_id = ?", roleID).
		Preload("ColumnMetadata").
		Find(&permissions).Error
	return permissions, err
}

func (s *DataGovernanceService) SetColumnPermission(p *models.ColumnPermission) error {
	var existing models.ColumnPermission
	err := s.DB.Where("role_id = ? AND column_metadata_id = ?", p.RoleID, p.ColumnMetadataID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return s.DB.Create(p).Error
	} else if err != nil {
		return err
	}

	return s.DB.Model(&existing).Updates(map[string]interface{}{
		"is_hidden":    p.IsHidden,
		"masking_type": p.MaskingType,
	}).Error
}

// === Enforcement ===

// ApplySecurity applies column-level security and masking to query results
func (s *DataGovernanceService) ApplySecurity(result *models.QueryResult, user *models.User, datasourceID string) {
	if user.IsSuperAdmin() {
		return // Admin bypass
	}

	// 1. Fetch all column metadata for this datasource
	var metadataList []models.ColumnMetadata
	s.DB.Where("datasource_id = ?", datasourceID).Find(&metadataList)

	if len(metadataList) == 0 {
		return
	}

	// Map column name to metadata ID
	colMap := make(map[string]uint)
	for _, m := range metadataList {
		colMap[m.Column] = m.ID
	}

	// 2. Fetch permissions for user's roles
	if len(user.Roles) == 0 {
		return // No roles, default allow
	}

	validRoleIDs := make([]uint, len(user.Roles))
	for i, r := range user.Roles {
		validRoleIDs[i] = r.ID
	}

	var permissions []models.ColumnPermission
	// Get permissions for ALL user roles that target the relevant columns
	s.DB.Where("role_id IN ? AND column_metadata_id IN (SELECT id FROM column_metadata WHERE datasource_id = ?)", validRoleIDs, datasourceID).
		Find(&permissions)

	// Build effective permission per column
	// Map: ColumnMetadataID -> Permission
	effectivePerms := make(map[uint]models.ColumnPermission)

	// Categorize permissions by ColumnMetadataID
	permsByColumn := make(map[uint][]models.ColumnPermission)
	for _, p := range permissions {
		permsByColumn[p.ColumnMetadataID] = append(permsByColumn[p.ColumnMetadataID], p)
	}

	for metaID, rolePerms := range permsByColumn {
		// Dispute resolution: Least restrictive wins
		// 1. If ANY role has NO restriction -> Full Access (handled later by checking role count vs perm count)
		// 2. If ALL roles have restrictions -> Compare restrictions

		// Start with the first permission as baseline
		bestPerm := rolePerms[0]

		for _, p := range rolePerms[1:] {
			// Rule 1: Not Hidden > Hidden
			if !p.IsHidden && bestPerm.IsHidden {
				bestPerm = p
				continue
			}
			if p.IsHidden && !bestPerm.IsHidden {
				continue // bestPerm is already better (visible)
			}

			// Rule 2: If both visible, lower Masking Score wins
			// Score: none (0) < partial (1) < full (2)
			newScore := getMaskingScore(p.MaskingType)
			currentScore := getMaskingScore(bestPerm.MaskingType)

			if newScore < currentScore {
				bestPerm = p
			}
		}
		effectivePerms[metaID] = bestPerm
	}

	// 3. Modifying the Result
	for colIdx, colName := range result.Columns {
		// Check if this column has metadata
		metaID, exists := colMap[colName]
		if !exists {
			continue
		}

		perm, hasRestrictions := effectivePerms[metaID]
		if !hasRestrictions {
			continue // No restrictions found for this column
		}

		// Check if user has ANY role that grants Full Access (i.e., no restriction row)
		// We count how many of the user's roles have an entry in `permissions` for this column
		rolesWithRestriction := 0
		for _, p := range permissions {
			if p.ColumnMetadataID == metaID {
				rolesWithRestriction++
			}
		}

		if rolesWithRestriction < len(validRoleIDs) {
			// At least one role has NO restriction -> Full Access
			continue
		}

		// All roles have restrictions, apply the least restrictive one (calculated above)
		if perm.IsHidden {
			// Mask entirely
			for rowIdx := range result.Rows {
				// Use specific placeholder to differentiate from NULL
				result.Rows[rowIdx][colIdx] = "[HIDDEN]"
			}
		} else if perm.MaskingType != "none" {
			// Apply masking
			for rowIdx := range result.Rows {
				val := result.Rows[rowIdx][colIdx]
				result.Rows[rowIdx][colIdx] = s.maskValue(val, perm.MaskingType)
			}
		}
	}
}

func getMaskingScore(maskType string) int {
	switch maskType {
	case "none":
		return 0
	case "email", "last4", "partial":
		return 1
	case "full":
		return 2
	default:
		return 10 // Unknown -> most restrictive
	}
}

func (s *DataGovernanceService) maskValue(val interface{}, strategy string) interface{} {
	if val == nil {
		return nil
	}
	str, ok := val.(string)
	if !ok {
		// Try to convert to string?
		str = fmt.Sprintf("%v", val)
	}

	switch strategy {
	case "email":
		// jdoe@example.com -> j***@example.com
		parts := strings.Split(str, "@")
		if len(parts) == 2 {
			if len(parts[0]) > 1 {
				return parts[0][0:1] + "***@" + parts[1]
			}
			return "***@" + parts[1]
		}
		return "***"
	case "last4":
		if len(str) > 4 {
			return strings.Repeat("*", len(str)-4) + str[len(str)-4:]
		}
		return str
	case "full":
		return "*****"
	case "partial":
		// Show first 2
		if len(str) > 2 {
			return str[0:2] + strings.Repeat("*", len(str)-2)
		}
		return str
	default:
		return str
	}
}
