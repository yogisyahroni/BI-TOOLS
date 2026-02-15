package services

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"insight-engine-backend/models"

	_ "github.com/glebarez/go-sqlite" // Pure Go SQLite driver
)

// AccelerationService manages the in-memory SQLite instance for query acceleration
type AccelerationService struct {
	db *sql.DB
	mu sync.RWMutex
}

var (
	accelerationInstance *AccelerationService
	accelerationOnce     sync.Once
)

// GetAccelerationService returns the singleton instance of AccelerationService
func GetAccelerationService() *AccelerationService {
	accelerationOnce.Do(func() {
		// Initialize in-memory SQLite
		// cache=shared allows multiple connections to access the same in-memory DB
		// mode=memory ensures it's purely in RAM
		dsn := "file:acceleration?mode=memory&cache=shared"
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			fmt.Printf("Failed to initialize Acceleration Service (SQLite): %v\n", err)
			return
		}

		// Verify connection
		if err := db.Ping(); err != nil {
			fmt.Printf("Failed to ping Acceleration Service (SQLite): %v\n", err)
			return
		}

		// Set connection limits for in-memory shared cache
		db.SetMaxOpenConns(10) // Limit concurrency to avoid locking issues in SQLite
		db.SetMaxIdleConns(5)

		accelerationInstance = &AccelerationService{
			db: db,
		}
		fmt.Println("Acceleration Service (SQLite Pure Go) Initialized")
	})
	return accelerationInstance
}

// LoadJSON loads JSON data into a SQLite table
// This parses the JSON and creates a table dynamically based on the first record
func (s *AccelerationService) LoadJSON(tableName string, jsonData []interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("acceleration service not initialized")
	}

	if len(jsonData) == 0 {
		return nil // Nothing to load
	}

	// 1. Infer Schema from first record
	firstRecord, ok := jsonData[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid json data format: expected []map[string]interface{}")
	}

	// Drop table if exists
	dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err := s.db.Exec(dropQuery)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	// Build CREATE TABLE query
	createParams := ""
	for key, val := range firstRecord {
		sqlType := "TEXT"
		switch val.(type) {
		case float64:
			sqlType = "REAL"
		case int, int64:
			sqlType = "INTEGER"
		case bool:
			sqlType = "INTEGER" // SQLite uses 0/1 for boolean
		}
		if createParams != "" {
			createParams += ", "
		}
		createParams += fmt.Sprintf("\"%s\" %s", key, sqlType)
	}

	createQuery := fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, createParams)
	_, err = s.db.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// 2. Insert Data
	// Optimally we should use a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare insertion statement
	// INSERT INTO table (col1, col2) VALUES (?, ?)
	columnNames := make([]string, 0, len(firstRecord))
	placeholders := ""
	for key := range firstRecord {
		columnNames = append(columnNames, key)
		if placeholders != "" {
			placeholders += ", "
		}
		placeholders += "?"
	}
	colsString := ""
	for _, col := range columnNames {
		if colsString != "" {
			colsString += ", "
		}
		colsString += fmt.Sprintf("\"%s\"", col)
	}

	insertSQL := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s)", tableName, colsString, placeholders)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, rowItem := range jsonData {
		rowMap, ok := rowItem.(map[string]interface{})
		if !ok {
			continue // Skip invalid rows
		}

		values := make([]interface{}, 0, len(columnNames))
		for _, col := range columnNames {
			val := rowMap[col]
			if bVal, isBool := val.(bool); isBool {
				if bVal {
					values = append(values, 1)
				} else {
					values = append(values, 0)
				}
			} else {
				values = append(values, val)
			}
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ExecuteQuery executes a SQL query against the in-memory SQLite
func (s *AccelerationService) ExecuteQuery(query string, params ...interface{}) (*models.QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("acceleration service not initialized")
	}

	start := time.Now()
	rows, err := s.db.Query(query, params...)
	if err != nil {
		return &models.QueryResult{Error: func() *string { s := err.Error(); return &s }()}, err
	}
	defer rows.Close()

	// Parse columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Parse rows (generic)
	var resultRows [][]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Post-processing for bytes -> string and SQLite typings
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
			// SQLite driver might require type assertion if returning interface{} directly
		}
		resultRows = append(resultRows, values)
	}

	return &models.QueryResult{
		Columns:       columns,
		Rows:          resultRows,
		RowCount:      len(resultRows),
		ExecutionTime: time.Since(start).Milliseconds(),
	}, nil
}

// Close closes the SQLite connection
func (s *AccelerationService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
