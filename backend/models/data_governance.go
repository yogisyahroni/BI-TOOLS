package models

import "time"

// DataClassification defines the sensitivity level of data
type DataClassification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"` // PII, Confidential, Public
	Description string    `json:"description"`
	Color       string    `json:"color"` // UI color hex
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name
func (DataClassification) TableName() string {
	return "data_classifications"
}

// ColumnMetadata stores additional metadata for table columns
type ColumnMetadata struct {
	ID                   uint                `gorm:"primaryKey" json:"id"`
	DatasourceID         string              `gorm:"index;type:varchar(36)" json:"datasource_id"` // UUID from Connection
	Table                string              `gorm:"column:table_name;index" json:"table_name"`
	Column               string              `gorm:"column:column_name;index" json:"column_name"`
	DataClassificationID *uint               `json:"data_classification_id"`
	DataClassification   *DataClassification `gorm:"foreignKey:DataClassificationID" json:"data_classification,omitempty"`
	Alias                string              `json:"alias"`
	Description          string              `json:"description"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// TableName specifies the table name
func (ColumnMetadata) TableName() string {
	return "column_metadata"
}

// ColumnPermission defines access rules for a specific column and role
type ColumnPermission struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	RoleID           uint           `gorm:"index" json:"role_id"`
	Role             Role           `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	ColumnMetadataID uint           `gorm:"index" json:"column_metadata_id"`
	ColumnMetadata   ColumnMetadata `gorm:"foreignKey:ColumnMetadataID" json:"column_metadata"`
	IsHidden         bool           `gorm:"default:false" json:"is_hidden"`
	MaskingType      string         `gorm:"type:varchar(20);default:'none'" json:"masking_type"` // none, full, email, last4, partial
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// TableName specifies the table name
func (ColumnPermission) TableName() string {
	return "column_permissions"
}
