package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// PaginationService handles cursor-based pagination logic
type PaginationService struct{}

// NewPaginationService creates a new pagination service
func NewPaginationService() *PaginationService {
	return &PaginationService{}
}

// CursorData represents the data encoded in the cursor
type CursorData struct {
	Values []interface{} `json:"v"` // Values of the sort columns
	Sorts  []SortConfig  `json:"s"` // Sort configuration to validate against current query
}

// SortConfig represents a sort column and direction
type SortConfig struct {
	Column    string `json:"c"`
	Direction string `json:"d"` // ASC or DESC
}

// EncodeCursor creates a base64 encoded cursor string from the last row's values
func (ps *PaginationService) EncodeCursor(values []interface{}, sorts []SortConfig) (string, error) {
	if len(values) == 0 {
		return "", nil
	}

	data := CursorData{
		Values: values,
		Sorts:  sorts,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// DecodeCursor decodes a cursor string into values
func (ps *PaginationService) DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}

	jsonData, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var data CursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor data: %w", err)
	}

	return &data, nil
}

// ApplyCursorPagination rewrites the SQL query to include cursor-based pagination
// This is a simplified implementation that assumes:
// 1. The query is a simple SELECT
// 2. We can inject WHERE clauses
// 3. We are sorting by unique columns (or set of columns)
func (ps *PaginationService) ApplyCursorPagination(originalSQL string, cursor string) (string, []interface{}, error) {
	if cursor == "" {
		return originalSQL, nil, nil
	}

	cursorData, err := ps.DecodeCursor(cursor)
	if err != nil {
		return "", nil, err
	}

	if len(cursorData.Values) != len(cursorData.Sorts) {
		return "", nil, fmt.Errorf("cursor values count mismatch sort columns count")
	}

	// Build the WHERE clause for keyset pagination
	// For (a, b) > (val_a, val_b)
	// Equivalent to: (a > val_a) OR (a = val_a AND b > val_b)

	// However, standard SQL tuple comparison is (a, b) > (val_a, val_b)
	// supported by Postgres, MySQL 8+, but maybe not all.
	// We'll use the expanded form for maximum compatibility.

	// conditions := []string{}
	// args := []interface{}{}

	// TODO: Implementing full SQL parsing is complex.
	// For now, we will interpret this as a "Optimization Hint"
	// If the query builder provided the SQL, we should modify the QueryBuilder logic instead.

	// For raw SQL, injecting into WHERE is risky without parsing.
	// Current strategy: This service provides the encoding/decoding.
	// The implementation of applying it should be in QueryExecutor or QueryBuilder where structure is known.

	return originalSQL, nil, nil
}
