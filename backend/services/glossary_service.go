package services

import (
	"insight-engine-backend/models"

	"gorm.io/gorm"
)

type GlossaryService struct {
	db *gorm.DB
}

func NewGlossaryService(db *gorm.DB) *GlossaryService {
	return &GlossaryService{db: db}
}

// CreateTerm creates a new business term
func (s *GlossaryService) CreateTerm(term *models.BusinessTerm) error {
	return s.db.Create(term).Error
}

// GetTerm retrieves a term by ID with its mappings
func (s *GlossaryService) GetTerm(id string) (*models.BusinessTerm, error) {
	var term models.BusinessTerm
	if err := s.db.Preload("RelatedColumns").First(&term, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &term, nil
}

// ListTerms retrieves all terms for a workspace
func (s *GlossaryService) ListTerms(workspaceID string) ([]models.BusinessTerm, error) {
	var terms []models.BusinessTerm
	err := s.db.Where("workspace_id = ?", workspaceID).Find(&terms).Error
	return terms, err
}

// UpdateTerm updates a business term
func (s *GlossaryService) UpdateTerm(term *models.BusinessTerm) error {
	return s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(term).Error
}

// DeleteTerm deletes a business term
func (s *GlossaryService) DeleteTerm(id string) error {
	return s.db.Delete(&models.BusinessTerm{}, "id = ?", id).Error
}

// AddMapping adds a mapping between a term and a column/metric
func (s *GlossaryService) AddMapping(mapping *models.TermColumnMapping) error {
	return s.db.Create(mapping).Error
}

// RemoveMapping removes a mapping
func (s *GlossaryService) RemoveMapping(id string) error {
	return s.db.Delete(&models.TermColumnMapping{}, "id = ?", id).Error
}

// AutoGenerateGlossary uses AI to generate glossary terms from schema metadata (Placeholder for NL feature)
func (s *GlossaryService) AutoGenerateGlossary(workspaceID string) error {
	// TODO: Integrate with AIService to scan schema and propose terms
	return nil
}
