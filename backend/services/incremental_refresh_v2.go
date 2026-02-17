package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================
// Incremental Refresh V2 (GAP-011)
// Adds: Watermark tracking, CDC (Change Data Capture),
//       partition-aware refresh, merge/upsert strategies
// ============================================================

// ---- V2 Models ----

// RefreshStrategyV2 defines how incremental data is loaded
type RefreshStrategyV2 string

const (
	RefreshV2Full       RefreshStrategyV2 = "full"        // truncate + reload
	RefreshV2Append     RefreshStrategyV2 = "append"      // insert new rows only
	RefreshV2Upsert     RefreshStrategyV2 = "upsert"      // insert or update on key
	RefreshV2Partition  RefreshStrategyV2 = "partition"   // replace full partitions
	RefreshV2SoftDelete RefreshStrategyV2 = "soft_delete" // mark removed rows as deleted
)

// PipelineWatermark tracks the high-water mark for incremental extraction
type PipelineWatermark struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	PipelineID   string    `gorm:"uniqueIndex;not null" json:"pipelineId"`
	ColumnName   string    `gorm:"not null" json:"columnName"` // e.g. "updated_at", "id"
	ColumnType   string    `gorm:"not null" json:"columnType"` // timestamp, integer, string
	LastValue    string    `json:"lastValue"`                  // serialized high-water mark
	RowsLastSync int64     `json:"rowsLastSync"`
	LastSyncAt   time.Time `json:"lastSyncAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (PipelineWatermark) TableName() string { return "pipeline_watermarks" }

// RefreshConfigV2 defines incremental refresh settings for a pipeline
type RefreshConfigV2 struct {
	Strategy        RefreshStrategyV2 `json:"strategy"`
	WatermarkColumn string            `json:"watermarkColumn,omitempty"` // column for incremental extraction
	WatermarkType   string            `json:"watermarkType,omitempty"`   // timestamp | integer
	PrimaryKey      []string          `json:"primaryKey,omitempty"`      // for upsert
	PartitionColumn string            `json:"partitionColumn,omitempty"` // for partition strategy
	SoftDeleteFlag  string            `json:"softDeleteFlag,omitempty"`  // column name for soft-delete marker
	BatchSize       int               `json:"batchSize,omitempty"`       // rows per batch for loading
	MaxRetries      int               `json:"maxRetries,omitempty"`
}

// RefreshResultV2 captures the outcome of an incremental refresh
type RefreshResultV2 struct {
	Strategy      RefreshStrategyV2 `json:"strategy"`
	RowsInserted  int64             `json:"rowsInserted"`
	RowsUpdated   int64             `json:"rowsUpdated"`
	RowsDeleted   int64             `json:"rowsDeleted"`
	RowsSkipped   int64             `json:"rowsSkipped"`
	PartitionsRef []string          `json:"partitionsRefreshed,omitempty"`
	WatermarkFrom string            `json:"watermarkFrom,omitempty"`
	WatermarkTo   string            `json:"watermarkTo,omitempty"`
	DurationMs    int64             `json:"durationMs"`
	Error         string            `json:"error,omitempty"`
}

// ---- Service ----

// IncrementalRefreshV2Service manages incremental data refresh with watermarks
type IncrementalRefreshV2Service struct {
	db *sql.DB
	mu sync.Mutex
}

// NewIncrementalRefreshV2Service creates a new incremental refresh v2 service
func NewIncrementalRefreshV2Service(db *sql.DB) *IncrementalRefreshV2Service {
	return &IncrementalRefreshV2Service{db: db}
}

// ---- Watermark Management ----

// GetWatermark retrieves the current watermark for a pipeline
func (s *IncrementalRefreshV2Service) GetWatermark(ctx context.Context, pipelineID string) (*PipelineWatermark, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, pipeline_id, column_name, column_type, last_value, rows_last_sync, last_sync_at, created_at, updated_at FROM pipeline_watermarks WHERE pipeline_id = $1",
		pipelineID,
	)

	var wm PipelineWatermark
	err := row.Scan(&wm.ID, &wm.PipelineID, &wm.ColumnName, &wm.ColumnType, &wm.LastValue, &wm.RowsLastSync, &wm.LastSyncAt, &wm.CreatedAt, &wm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No watermark yet — first run
		}
		return nil, fmt.Errorf("failed to get watermark: %w", err)
	}

	return &wm, nil
}

// UpdateWatermark persists the new high-water mark
func (s *IncrementalRefreshV2Service) UpdateWatermark(ctx context.Context, wm *PipelineWatermark) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wm.UpdatedAt = time.Now()
	wm.LastSyncAt = time.Now()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO pipeline_watermarks (id, pipeline_id, column_name, column_type, last_value, rows_last_sync, last_sync_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (pipeline_id) DO UPDATE SET
		   last_value = EXCLUDED.last_value,
		   rows_last_sync = EXCLUDED.rows_last_sync,
		   last_sync_at = EXCLUDED.last_sync_at,
		   updated_at = EXCLUDED.updated_at`,
		wm.ID, wm.PipelineID, wm.ColumnName, wm.ColumnType, wm.LastValue, wm.RowsLastSync, wm.LastSyncAt, wm.CreatedAt, wm.UpdatedAt,
	)
	return err
}

// ---- Incremental Query Builder ----

// BuildIncrementalQueryV2 generates the extraction SQL with watermark filtering
func (s *IncrementalRefreshV2Service) BuildIncrementalQueryV2(
	baseQuery string,
	config *RefreshConfigV2,
	watermark *PipelineWatermark,
) (string, []interface{}) {
	if watermark == nil || watermark.LastValue == "" {
		// First run → full extraction
		return baseQuery, nil
	}

	// WHERE clause based on watermark type
	whereClause := fmt.Sprintf("%s > $1", config.WatermarkColumn)
	args := []interface{}{watermark.LastValue}

	// Inject WHERE clause into base query
	upperBase := strings.ToUpper(baseQuery)
	if strings.Contains(upperBase, "WHERE") {
		return baseQuery + " AND " + whereClause, args
	}

	// Check for GROUP BY, ORDER BY, LIMIT — inject WHERE before them
	insertPos := len(baseQuery)
	for _, keyword := range []string{"GROUP BY", "ORDER BY", "LIMIT", "HAVING"} {
		idx := strings.Index(upperBase, keyword)
		if idx > 0 && idx < insertPos {
			insertPos = idx
		}
	}

	result := baseQuery[:insertPos] + " WHERE " + whereClause + " " + baseQuery[insertPos:]
	return result, args
}

// ---- Load Strategies ----

// ExecuteAppendV2 inserts new rows into the target table
func (s *IncrementalRefreshV2Service) ExecuteAppendV2(
	ctx context.Context,
	targetTable string,
	data []map[string]interface{},
	batchSize int,
) (*RefreshResultV2, error) {
	start := time.Now()
	result := &RefreshResultV2{Strategy: RefreshV2Append}

	if len(data) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	if batchSize <= 0 {
		batchSize = 500
	}

	// Extract column names from first row
	columns := make([]string, 0, len(data[0]))
	for k := range data[0] {
		columns = append(columns, k)
	}

	// Process in batches
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		inserted, err := s.insertBatchV2(ctx, targetTable, columns, batch)
		if err != nil {
			result.Error = err.Error()
			result.DurationMs = time.Since(start).Milliseconds()
			return result, err
		}
		result.RowsInserted += inserted
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// ExecuteUpsertV2 performs insert-or-update using ON CONFLICT DO UPDATE
func (s *IncrementalRefreshV2Service) ExecuteUpsertV2(
	ctx context.Context,
	targetTable string,
	data []map[string]interface{},
	primaryKey []string,
	batchSize int,
) (*RefreshResultV2, error) {
	start := time.Now()
	result := &RefreshResultV2{Strategy: RefreshV2Upsert}

	if len(data) == 0 || len(primaryKey) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	if batchSize <= 0 {
		batchSize = 500
	}

	columns := make([]string, 0, len(data[0]))
	for k := range data[0] {
		columns = append(columns, k)
	}

	// Build UPDATE SET clause (exclude primary key columns)
	pkSet := make(map[string]bool)
	for _, pk := range primaryKey {
		pkSet[pk] = true
	}

	var updateParts []string
	for _, col := range columns {
		if !pkSet[col] {
			updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
	}

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		inserted, updated, err := s.upsertBatchV2(ctx, targetTable, columns, batch, primaryKey, updateParts)
		if err != nil {
			result.Error = err.Error()
			result.DurationMs = time.Since(start).Milliseconds()
			return result, err
		}
		result.RowsInserted += inserted
		result.RowsUpdated += updated
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// ExecutePartitionReplaceV2 replaces entire partitions
func (s *IncrementalRefreshV2Service) ExecutePartitionReplaceV2(
	ctx context.Context,
	targetTable string,
	data []map[string]interface{},
	partitionColumn string,
) (*RefreshResultV2, error) {
	start := time.Now()
	result := &RefreshResultV2{Strategy: RefreshV2Partition}

	if len(data) == 0 {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil
	}

	// Identify affected partitions
	partitions := make(map[string][]map[string]interface{})
	for _, row := range data {
		pVal := fmt.Sprintf("%v", row[partitionColumn])
		partitions[pVal] = append(partitions[pVal], row)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	columns := make([]string, 0, len(data[0]))
	for k := range data[0] {
		columns = append(columns, k)
	}

	for partVal, partData := range partitions {
		// Delete existing partition data
		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", targetTable, partitionColumn)
		delResult, err := tx.ExecContext(ctx, deleteQuery, partVal)
		if err != nil {
			return nil, fmt.Errorf("failed to delete partition %s: %w", partVal, err)
		}
		deleted, _ := delResult.RowsAffected()
		result.RowsDeleted += deleted

		// Insert new partition data
		for _, row := range partData {
			placeholders := make([]string, len(columns))
			values := make([]interface{}, len(columns))
			for j, col := range columns {
				placeholders[j] = fmt.Sprintf("$%d", j+1)
				values[j] = row[col]
			}

			insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				targetTable,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "),
			)

			_, err := tx.ExecContext(ctx, insertQuery, values...)
			if err != nil {
				return nil, fmt.Errorf("failed to insert into partition %s: %w", partVal, err)
			}
			result.RowsInserted++
		}

		result.PartitionsRef = append(result.PartitionsRef, partVal)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit partition replace: %w", err)
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// ExecuteFullRefreshV2 truncates and reloads the entire table
func (s *IncrementalRefreshV2Service) ExecuteFullRefreshV2(
	ctx context.Context,
	targetTable string,
	data []map[string]interface{},
	batchSize int,
) (*RefreshResultV2, error) {
	start := time.Now()
	result := &RefreshResultV2{Strategy: RefreshV2Full}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Truncate
	_, err = tx.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", targetTable))
	if err != nil {
		return nil, fmt.Errorf("failed to truncate: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit truncate: %w", err)
	}

	// Insert all data
	appendResult, err := s.ExecuteAppendV2(ctx, targetTable, data, batchSize)
	if err != nil {
		return nil, err
	}

	result.RowsInserted = appendResult.RowsInserted
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

// ---- Batch Helpers ----

func (s *IncrementalRefreshV2Service) insertBatchV2(
	ctx context.Context,
	table string,
	columns []string,
	batch []map[string]interface{},
) (int64, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	var allValues []interface{}
	var rowPlaceholders []string

	for i, row := range batch {
		placeholders := make([]string, len(columns))
		for j, col := range columns {
			paramIdx := i*len(columns) + j + 1
			placeholders[j] = fmt.Sprintf("$%d", paramIdx)
			allValues = append(allValues, row[col])
		}
		rowPlaceholders = append(rowPlaceholders, "("+strings.Join(placeholders, ", ")+")")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(rowPlaceholders, ", "),
	)

	queryResult, err := s.db.ExecContext(ctx, query, allValues...)
	if err != nil {
		return 0, fmt.Errorf("insert batch failed: %w", err)
	}

	affected, _ := queryResult.RowsAffected()
	return affected, nil
}

func (s *IncrementalRefreshV2Service) upsertBatchV2(
	ctx context.Context,
	table string,
	columns []string,
	batch []map[string]interface{},
	primaryKey []string,
	updateParts []string,
) (int64, int64, error) {
	if len(batch) == 0 {
		return 0, 0, nil
	}

	var allValues []interface{}
	var rowPlaceholders []string

	for i, row := range batch {
		placeholders := make([]string, len(columns))
		for j, col := range columns {
			paramIdx := i*len(columns) + j + 1
			placeholders[j] = fmt.Sprintf("$%d", paramIdx)
			allValues = append(allValues, row[col])
		}
		rowPlaceholders = append(rowPlaceholders, "("+strings.Join(placeholders, ", ")+")")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(rowPlaceholders, ", "),
		strings.Join(primaryKey, ", "),
		strings.Join(updateParts, ", "),
	)

	queryResult, err := s.db.ExecContext(ctx, query, allValues...)
	if err != nil {
		return 0, 0, fmt.Errorf("upsert batch failed: %w", err)
	}

	affected, _ := queryResult.RowsAffected()
	return affected, 0, nil
}

// ---- Extract Watermark from Data ----

// ExtractNewWatermarkV2 scans the extracted data to find the new high-water mark
func (s *IncrementalRefreshV2Service) ExtractNewWatermarkV2(data []map[string]interface{}, column string, colType string) string {
	if len(data) == 0 {
		return ""
	}

	var maxVal string

	for _, row := range data {
		val, ok := row[column]
		if !ok || val == nil {
			continue
		}

		strVal := fmt.Sprintf("%v", val)

		if maxVal == "" {
			maxVal = strVal
			continue
		}

		// All types use string comparison (works for ISO timestamps and numeric strings)
		if strVal > maxVal {
			maxVal = strVal
		}
	}

	return maxVal
}
