package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"insight-engine-backend/database"
	"insight-engine-backend/models"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// PipelineExecutor handles real data pipeline execution
type PipelineExecutor struct {
	db         *sql.DB
	mu         sync.RWMutex
	activeRuns map[string]*ExecutionContext
}

// ExecutionContext tracks a running pipeline execution
type ExecutionContext struct {
	ExecutionID string
	PipelineID  string
	Status      string
	Progress    int
	Cancel      context.CancelFunc
	StartedAt   time.Time
}

// ExecutionResult contains the outcome of a pipeline run
type ExecutionResult struct {
	RowsProcessed     int
	BytesProcessed    int64
	QualityViolations int
	Logs              []models.ExecutionLog
	Error             error
	DurationMs        int
}

// QualityViolation tracks a single quality rule violation
type QualityViolation struct {
	RuleID   string `json:"ruleId"`
	Column   string `json:"column"`
	RuleType string `json:"ruleType"`
	RowIndex int    `json:"rowIndex"`
	Value    string `json:"value"`
	Message  string `json:"message"`
}

// Global executor instance
var GlobalPipelineExecutor *PipelineExecutor

// NewPipelineExecutor creates a new pipeline executor
func NewPipelineExecutor() *PipelineExecutor {
	return &PipelineExecutor{
		activeRuns: make(map[string]*ExecutionContext),
	}
}

// InitPipelineExecutor initializes the global executor
func InitPipelineExecutor() {
	GlobalPipelineExecutor = NewPipelineExecutor()
	LogInfo("pipeline_executor_init", "Pipeline executor initialized", nil)
}

// GetActiveRun returns the execution context for a running pipeline
func (pe *PipelineExecutor) GetActiveRun(executionID string) *ExecutionContext {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.activeRuns[executionID]
}

// Execute runs a complete pipeline: extract → transform → load
func (pe *PipelineExecutor) Execute(pipelineID string, executionID string) *ExecutionResult {
	result := &ExecutionResult{
		Logs: make([]models.ExecutionLog, 0, 64),
	}

	startTime := time.Now()

	// Create cancellable context with 30-minute timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Register active run
	execCtx := &ExecutionContext{
		ExecutionID: executionID,
		PipelineID:  pipelineID,
		Status:      "PROCESSING",
		Progress:    0,
		Cancel:      cancel,
		StartedAt:   startTime,
	}
	pe.mu.Lock()
	pe.activeRuns[executionID] = execCtx
	pe.mu.Unlock()

	defer func() {
		pe.mu.Lock()
		delete(pe.activeRuns, executionID)
		pe.mu.Unlock()
	}()

	// Step 1: Load pipeline configuration
	result.appendLog("INFO", "INIT", "Loading pipeline configuration", "")
	pe.updateProgress(executionID, 5, "PROCESSING")

	var pipeline models.Pipeline
	if err := database.DB.Preload("QualityRules").First(&pipeline, "id = ?", pipelineID).Error; err != nil {
		result.Error = fmt.Errorf("pipeline not found: %w", err)
		result.appendLog("ERROR", "INIT", "Pipeline not found", err.Error())
		return pe.finalizeResult(result, startTime)
	}

	// Step 2: Parse source configuration
	result.appendLog("INFO", "INIT", fmt.Sprintf("Pipeline '%s' loaded, source type: %s", pipeline.Name, pipeline.SourceType), "")
	pe.updateProgress(executionID, 10, "EXTRACTING")

	var sourceConfig models.SourceConfig
	if err := json.Unmarshal([]byte(pipeline.SourceConfig), &sourceConfig); err != nil {
		result.Error = fmt.Errorf("invalid source config: %w", err)
		result.appendLog("ERROR", "INIT", "Failed to parse source configuration", err.Error())
		return pe.finalizeResult(result, startTime)
	}

	// Step 3: Extract data from source
	result.appendLog("INFO", "EXTRACT", fmt.Sprintf("Connecting to %s source...", pipeline.SourceType), "")

	data, rowCount, bytesRead, err := pe.extractData(ctx, &pipeline, &sourceConfig)
	if err != nil {
		result.Error = fmt.Errorf("extraction failed: %w", err)
		result.appendLog("ERROR", "EXTRACT", "Data extraction failed", err.Error())
		return pe.finalizeResult(result, startTime)
	}

	result.RowsProcessed = rowCount
	result.BytesProcessed = bytesRead
	result.appendLog("INFO", "EXTRACT", fmt.Sprintf("Extracted %d rows (%d bytes)", rowCount, bytesRead), "")
	pe.updateProgress(executionID, 40, "TRANSFORMING")

	// Step 4: Apply transformation steps
	if pipeline.TransformationSteps != nil && *pipeline.TransformationSteps != "" && *pipeline.TransformationSteps != "null" {
		result.appendLog("INFO", "TRANSFORM", "Applying transformation steps...", "")

		var steps []models.TransformStep
		if err := json.Unmarshal([]byte(*pipeline.TransformationSteps), &steps); err != nil {
			result.Error = fmt.Errorf("invalid transformation steps: %w", err)
			result.appendLog("ERROR", "TRANSFORM", "Failed to parse transformation steps", err.Error())
			return pe.finalizeResult(result, startTime)
		}

		// Sort by order
		sort.Slice(steps, func(i, j int) bool { return steps[i].Order < steps[j].Order })

		for i, step := range steps {
			select {
			case <-ctx.Done():
				result.Error = fmt.Errorf("execution cancelled or timed out")
				result.appendLog("ERROR", "TRANSFORM", "Execution cancelled", "")
				return pe.finalizeResult(result, startTime)
			default:
			}

			result.appendLog("INFO", "TRANSFORM", fmt.Sprintf("Step %d/%d: %s", i+1, len(steps), step.Type), "")
			data, err = pe.applyTransform(data, &step)
			if err != nil {
				result.Error = fmt.Errorf("transform step '%s' failed: %w", step.Type, err)
				result.appendLog("ERROR", "TRANSFORM", fmt.Sprintf("Transform step '%s' failed", step.Type), err.Error())
				return pe.finalizeResult(result, startTime)
			}

			progressPct := 40 + ((i + 1) * 20 / len(steps))
			pe.updateProgress(executionID, progressPct, "TRANSFORMING")
		}

		result.RowsProcessed = len(data)
		result.appendLog("INFO", "TRANSFORM", fmt.Sprintf("Transformations complete. Rows after transforms: %d", len(data)), "")
	} else {
		result.appendLog("INFO", "TRANSFORM", "No transformation steps configured, skipping", "")
	}

	pe.updateProgress(executionID, 65, "LOADING")

	// Step 5: Validate quality rules
	if len(pipeline.QualityRules) > 0 {
		result.appendLog("INFO", "VALIDATE", fmt.Sprintf("Running %d quality rules...", len(pipeline.QualityRules)), "")
		violations := pe.validateQualityRules(data, pipeline.QualityRules)
		result.QualityViolations = len(violations)

		if len(violations) > 0 {
			result.appendLog("WARN", "VALIDATE", fmt.Sprintf("%d quality violations found", len(violations)), "")

			// Check for FAIL severity violations
			for _, rule := range pipeline.QualityRules {
				if rule.Severity == "FAIL" {
					for _, v := range violations {
						if v.RuleType == rule.RuleType && v.Column == rule.Column {
							result.Error = fmt.Errorf("quality rule violation (severity=FAIL): %s on column '%s'", rule.RuleType, rule.Column)
							result.appendLog("ERROR", "VALIDATE", "Pipeline aborted due to FAIL-severity quality violation", result.Error.Error())
							return pe.finalizeResult(result, startTime)
						}
					}
				}
			}
		} else {
			result.appendLog("INFO", "VALIDATE", "All quality rules passed", "")
		}
	}

	pe.updateProgress(executionID, 75, "LOADING")

	// Step 6: Load to destination
	result.appendLog("INFO", "LOAD", fmt.Sprintf("Loading %d rows to destination (%s)...", len(data), pipeline.DestinationType), "")

	if err := pe.loadToDestination(ctx, &pipeline, data); err != nil {
		result.Error = fmt.Errorf("load failed: %w", err)
		result.appendLog("ERROR", "LOAD", "Failed to load data to destination", err.Error())
		return pe.finalizeResult(result, startTime)
	}

	result.appendLog("INFO", "LOAD", fmt.Sprintf("Successfully loaded %d rows to destination", len(data)), "")
	pe.updateProgress(executionID, 100, "COMPLETED")

	result.appendLog("INFO", "COMPLETE", fmt.Sprintf("Pipeline execution completed in %dms", int(time.Since(startTime).Milliseconds())), "")

	return pe.finalizeResult(result, startTime)
}

// extractData connects to the source and retrieves data
func (pe *PipelineExecutor) extractData(ctx context.Context, pipeline *models.Pipeline, config *models.SourceConfig) ([]map[string]interface{}, int, int64, error) {
	switch pipeline.SourceType {
	case "POSTGRES":
		return pe.extractFromPostgres(ctx, pipeline, config)
	case "MYSQL":
		return pe.extractFromMySQL(ctx, pipeline, config)
	default:
		return nil, 0, 0, fmt.Errorf("unsupported source type: %s", pipeline.SourceType)
	}
}

// extractFromPostgres extracts data from a PostgreSQL source
func (pe *PipelineExecutor) extractFromPostgres(ctx context.Context, pipeline *models.Pipeline, config *models.SourceConfig) ([]map[string]interface{}, int, int64, error) {
	// Resolve connection credentials
	host, port, dbName, username, password, sslMode, err := pe.resolveConnectionCredentials(pipeline, config)
	if err != nil {
		return nil, 0, 0, err
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=30",
		host, port, username, password, dbName, sslMode)

	sourceDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer sourceDB.Close()

	sourceDB.SetMaxOpenConns(5)
	sourceDB.SetMaxIdleConns(1)
	sourceDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sourceDB.PingContext(ctx); err != nil {
		return nil, 0, 0, fmt.Errorf("PostgreSQL connection ping failed: %w", err)
	}

	query := pe.resolveQuery(pipeline, config)
	if query == "" {
		return nil, 0, 0, fmt.Errorf("no source query configured")
	}

	// Apply row limit
	rowLimit := pipeline.RowLimit
	if rowLimit <= 0 {
		rowLimit = 100000
	}
	limitedQuery := fmt.Sprintf("SELECT * FROM (%s) AS _sub LIMIT %d", query, rowLimit)

	return pe.executeQuery(ctx, sourceDB, limitedQuery)
}

// extractFromMySQL extracts data from a MySQL source
func (pe *PipelineExecutor) extractFromMySQL(ctx context.Context, pipeline *models.Pipeline, config *models.SourceConfig) ([]map[string]interface{}, int, int64, error) {
	host, port, dbName, username, password, _, err := pe.resolveConnectionCredentials(pipeline, config)
	if err != nil {
		return nil, 0, 0, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=30s&parseTime=true",
		username, password, host, port, dbName)

	sourceDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer sourceDB.Close()

	sourceDB.SetMaxOpenConns(5)
	sourceDB.SetMaxIdleConns(1)
	sourceDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sourceDB.PingContext(ctx); err != nil {
		return nil, 0, 0, fmt.Errorf("MySQL connection ping failed: %w", err)
	}

	query := pe.resolveQuery(pipeline, config)
	if query == "" {
		return nil, 0, 0, fmt.Errorf("no source query configured")
	}

	rowLimit := pipeline.RowLimit
	if rowLimit <= 0 {
		rowLimit = 100000
	}
	limitedQuery := fmt.Sprintf("SELECT * FROM (%s) AS _sub LIMIT %d", query, rowLimit)

	return pe.executeQuery(ctx, sourceDB, limitedQuery)
}

// resolveConnectionCredentials gets source connection details from either ConnectionID or inline config
func (pe *PipelineExecutor) resolveConnectionCredentials(pipeline *models.Pipeline, config *models.SourceConfig) (host string, port int, dbName string, username string, password string, sslMode string, err error) {
	// Priority: ConnectionID (existing Connection record) > inline SourceConfig
	if pipeline.ConnectionID != nil && *pipeline.ConnectionID != "" {
		var conn models.Connection
		if dbErr := database.DB.First(&conn, "id = ?", *pipeline.ConnectionID).Error; dbErr != nil {
			return "", 0, "", "", "", "", fmt.Errorf("connection not found: %w", dbErr)
		}

		if conn.Host != nil {
			host = *conn.Host
		}
		if conn.Port != nil {
			port = *conn.Port
		}
		dbName = conn.Database
		if conn.Username != nil {
			username = *conn.Username
		}
		if conn.Password != nil {
			password = *conn.Password
		}
		sslMode = "disable"
		if conn.Options != nil {
			if ssl, ok := (*conn.Options)["sslMode"]; ok {
				sslMode = fmt.Sprintf("%v", ssl)
			}
		}
		return host, port, dbName, username, password, sslMode, nil
	}

	// Fallback: use inline source config
	host = config.Host
	port = config.Port
	dbName = config.Database
	username = config.Username
	password = config.Password
	sslMode = config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	if host == "" {
		return "", 0, "", "", "", "", fmt.Errorf("no host configured in source config or connection")
	}

	return host, port, dbName, username, password, sslMode, nil
}

// resolveQuery determines which SQL query to execute
func (pe *PipelineExecutor) resolveQuery(pipeline *models.Pipeline, config *models.SourceConfig) string {
	// Priority: Pipeline.SourceQuery > SourceConfig.Query
	if pipeline.SourceQuery != nil && *pipeline.SourceQuery != "" {
		return *pipeline.SourceQuery
	}
	return config.Query
}

// executeQuery runs a SQL query and returns results as maps
func (pe *PipelineExecutor) executeQuery(ctx context.Context, db *sql.DB, query string) ([]map[string]interface{}, int, int64, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to get columns: %w", err)
	}

	var result []map[string]interface{}
	var totalBytes int64

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, 0, fmt.Errorf("scan failed: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				val = string(b)
				totalBytes += int64(len(b))
			} else {
				// Rough estimate for non-byte types
				totalBytes += 8
			}
			row[col] = val
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return result, len(result), totalBytes, nil
}

// applyTransform applies a single transformation step to the data
func (pe *PipelineExecutor) applyTransform(data []map[string]interface{}, step *models.TransformStep) ([]map[string]interface{}, error) {
	switch step.Type {
	case "FILTER":
		return pe.transformFilter(data, step.Config)
	case "RENAME":
		return pe.transformRename(data, step.Config)
	case "CAST":
		return pe.transformCast(data, step.Config)
	case "DEDUPLICATE":
		return pe.transformDeduplicate(data, step.Config)
	case "AGGREGATE":
		return pe.transformAggregate(data, step.Config)
	default:
		return data, nil // Unknown transform types are silently skipped
	}
}

// transformFilter removes rows that don't match the condition
func (pe *PipelineExecutor) transformFilter(data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	column, _ := config["column"].(string)
	operator, _ := config["operator"].(string)
	value := config["value"]

	if column == "" || operator == "" {
		return data, fmt.Errorf("filter requires 'column' and 'operator'")
	}

	var filtered []map[string]interface{}
	for _, row := range data {
		cellVal := row[column]
		if pe.evaluateCondition(cellVal, operator, value) {
			filtered = append(filtered, row)
		}
	}

	return filtered, nil
}

// transformRename renames columns in the data
func (pe *PipelineExecutor) transformRename(data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	mappings, ok := config["mappings"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("rename requires 'mappings' object")
	}

	for _, row := range data {
		for oldName, newNameVal := range mappings {
			newName, _ := newNameVal.(string)
			if newName == "" {
				continue
			}
			if val, exists := row[oldName]; exists {
				row[newName] = val
				delete(row, oldName)
			}
		}
	}

	return data, nil
}

// transformCast converts column values to specified types
func (pe *PipelineExecutor) transformCast(data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	casts, ok := config["casts"].(map[string]interface{})
	if !ok {
		return data, fmt.Errorf("cast requires 'casts' object")
	}

	for _, row := range data {
		for col, targetTypeVal := range casts {
			targetType, _ := targetTypeVal.(string)
			if val, exists := row[col]; exists {
				row[col] = pe.castValue(val, targetType)
			}
		}
	}

	return data, nil
}

// transformDeduplicate removes duplicate rows based on specified columns
func (pe *PipelineExecutor) transformDeduplicate(data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	columnsRaw, ok := config["columns"].([]interface{})
	if !ok || len(columnsRaw) == 0 {
		return data, fmt.Errorf("deduplicate requires 'columns' array")
	}

	var columns []string
	for _, c := range columnsRaw {
		if s, ok := c.(string); ok {
			columns = append(columns, s)
		}
	}

	seen := make(map[string]bool)
	var result []map[string]interface{}

	for _, row := range data {
		var keyParts []string
		for _, col := range columns {
			keyParts = append(keyParts, fmt.Sprintf("%v", row[col]))
		}
		key := strings.Join(keyParts, "||")

		if !seen[key] {
			seen[key] = true
			result = append(result, row)
		}
	}

	return result, nil
}

// transformAggregate groups data and applies aggregate functions
func (pe *PipelineExecutor) transformAggregate(data []map[string]interface{}, config map[string]interface{}) ([]map[string]interface{}, error) {
	groupByRaw, _ := config["groupBy"].([]interface{})
	aggregatesRaw, _ := config["aggregates"].([]interface{})

	if len(groupByRaw) == 0 {
		return data, fmt.Errorf("aggregate requires 'groupBy' array")
	}

	var groupBy []string
	for _, g := range groupByRaw {
		if s, ok := g.(string); ok {
			groupBy = append(groupBy, s)
		}
	}

	// Group rows by key
	groups := make(map[string][]map[string]interface{})
	groupOrder := make([]string, 0)

	for _, row := range data {
		var keyParts []string
		for _, col := range groupBy {
			keyParts = append(keyParts, fmt.Sprintf("%v", row[col]))
		}
		key := strings.Join(keyParts, "||")
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], row)
	}

	// Build result with aggregates
	var result []map[string]interface{}
	for _, key := range groupOrder {
		groupRows := groups[key]
		resultRow := make(map[string]interface{})

		// Copy groupBy columns from first row
		for _, col := range groupBy {
			resultRow[col] = groupRows[0][col]
		}

		// Apply aggregates
		for _, aggRaw := range aggregatesRaw {
			aggMap, ok := aggRaw.(map[string]interface{})
			if !ok {
				continue
			}
			col, _ := aggMap["column"].(string)
			fn, _ := aggMap["function"].(string)
			alias, _ := aggMap["alias"].(string)
			if alias == "" {
				alias = fmt.Sprintf("%s_%s", fn, col)
			}

			resultRow[alias] = pe.computeAggregate(groupRows, col, fn)
		}

		result = append(result, resultRow)
	}

	return result, nil
}

// computeAggregate calculates an aggregate value for a group
func (pe *PipelineExecutor) computeAggregate(rows []map[string]interface{}, column string, fn string) interface{} {
	switch strings.ToUpper(fn) {
	case "COUNT":
		return len(rows)
	case "SUM":
		var sum float64
		for _, row := range rows {
			sum += pe.toFloat64(row[column])
		}
		return sum
	case "AVG":
		if len(rows) == 0 {
			return 0.0
		}
		var sum float64
		for _, row := range rows {
			sum += pe.toFloat64(row[column])
		}
		return sum / float64(len(rows))
	case "MIN":
		if len(rows) == 0 {
			return nil
		}
		minVal := pe.toFloat64(rows[0][column])
		for _, row := range rows[1:] {
			v := pe.toFloat64(row[column])
			if v < minVal {
				minVal = v
			}
		}
		return minVal
	case "MAX":
		if len(rows) == 0 {
			return nil
		}
		maxVal := pe.toFloat64(rows[0][column])
		for _, row := range rows[1:] {
			v := pe.toFloat64(row[column])
			if v > maxVal {
				maxVal = v
			}
		}
		return maxVal
	default:
		return nil
	}
}

// validateQualityRules checks data against quality rules
func (pe *PipelineExecutor) validateQualityRules(data []map[string]interface{}, rules []models.QualityRule) []QualityViolation {
	var violations []QualityViolation

	for _, rule := range rules {
		for i, row := range data {
			val := row[rule.Column]

			var violated bool
			switch rule.RuleType {
			case "NOT_NULL":
				violated = val == nil || fmt.Sprintf("%v", val) == ""
			case "UNIQUE":
				// Check uniqueness across all rows (O(n) per rule)
				for j, otherRow := range data {
					if j != i && fmt.Sprintf("%v", otherRow[rule.Column]) == fmt.Sprintf("%v", val) {
						violated = true
						break
					}
				}
			case "RANGE":
				if rule.Value != nil {
					var rangeConfig struct {
						Min float64 `json:"min"`
						Max float64 `json:"max"`
					}
					if err := json.Unmarshal([]byte(*rule.Value), &rangeConfig); err == nil {
						numVal := pe.toFloat64(val)
						violated = numVal < rangeConfig.Min || numVal > rangeConfig.Max
					}
				}
			case "REGEX":
				// Simple string match for now
				if rule.Value != nil && val != nil {
					valStr := fmt.Sprintf("%v", val)
					violated = !strings.Contains(valStr, *rule.Value)
				}
			}

			if violated {
				violations = append(violations, QualityViolation{
					RuleID:   rule.ID,
					Column:   rule.Column,
					RuleType: rule.RuleType,
					RowIndex: i,
					Value:    fmt.Sprintf("%v", val),
					Message:  fmt.Sprintf("Row %d, column '%s': %s rule violated", i, rule.Column, rule.RuleType),
				})
			}
		}
	}

	return violations
}

// loadToDestination writes processed data to the target
func (pe *PipelineExecutor) loadToDestination(ctx context.Context, pipeline *models.Pipeline, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil // Nothing to load
	}

	switch pipeline.DestinationType {
	case "INTERNAL_RAW":
		return pe.loadToInternalRaw(ctx, pipeline, data)
	case "POSTGRES", "MYSQL":
		return pe.loadToExternalDB(ctx, pipeline, data)
	default:
		return fmt.Errorf("unsupported destination type: %s", pipeline.DestinationType)
	}
}

// loadToInternalRaw writes data to the application's own database as a raw table
func (pe *PipelineExecutor) loadToInternalRaw(ctx context.Context, pipeline *models.Pipeline, data []map[string]interface{}) error {
	// Generate table name from pipeline name
	tableName := fmt.Sprintf("pipeline_data_%s", strings.ReplaceAll(strings.ToLower(pipeline.Name), " ", "_"))
	tableName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, tableName)

	// Truncate to 63 chars (Postgres identifier limit)
	if len(tableName) > 63 {
		tableName = tableName[:63]
	}

	// Determine columns from first row
	firstRow := data[0]
	var columns []string
	for col := range firstRow {
		columns = append(columns, col)
	}
	sort.Strings(columns) // Deterministic column order

	// Create the table (DROP + CREATE for OVERWRITE mode, or CREATE IF NOT EXISTS for APPEND)
	writeMode := "OVERWRITE"
	if pipeline.DestinationConfig != nil {
		var destConf models.DestConfig
		if err := json.Unmarshal([]byte(*pipeline.DestinationConfig), &destConf); err == nil && destConf.WriteMode != "" {
			writeMode = destConf.WriteMode
		}
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}

	if writeMode == "OVERWRITE" {
		dropSQL := fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, tableName)
		if _, err := sqlDB.ExecContext(ctx, dropSQL); err != nil {
			return fmt.Errorf("failed to drop existing table: %w", err)
		}
	}

	// Build CREATE TABLE with TEXT columns (raw ingestion pattern)
	var colDefs []string
	for _, col := range columns {
		colDefs = append(colDefs, fmt.Sprintf(`"%s" TEXT`, col))
	}
	createSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s)`, tableName, strings.Join(colDefs, ", "))
	if _, err := sqlDB.ExecContext(ctx, createSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Batch insert data (chunks of 500 rows)
	batchSize := 500
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		if err := pe.insertBatch(ctx, sqlDB, tableName, columns, batch); err != nil {
			return fmt.Errorf("batch insert failed at row %d: %w", i, err)
		}
	}

	return nil
}

// loadToExternalDB writes data to an external database
func (pe *PipelineExecutor) loadToExternalDB(ctx context.Context, pipeline *models.Pipeline, data []map[string]interface{}) error {
	if pipeline.DestinationConfig == nil {
		return fmt.Errorf("destination config required for external DB target")
	}

	var destConf models.DestConfig
	if err := json.Unmarshal([]byte(*pipeline.DestinationConfig), &destConf); err != nil {
		return fmt.Errorf("invalid destination config: %w", err)
	}

	if destConf.ConnectionID == "" {
		return fmt.Errorf("connectionId required in destination config for external DB")
	}

	// Get destination connection
	var conn models.Connection
	if err := database.DB.First(&conn, "id = ?", destConf.ConnectionID).Error; err != nil {
		return fmt.Errorf("destination connection not found: %w", err)
	}

	// Build DSN
	var dsn string
	var driver string
	host := ""
	if conn.Host != nil {
		host = *conn.Host
	}
	port := 5432
	if conn.Port != nil {
		port = *conn.Port
	}
	username := ""
	if conn.Username != nil {
		username = *conn.Username
	}
	password := ""
	if conn.Password != nil {
		password = *conn.Password
	}

	switch pipeline.DestinationType {
	case "POSTGRES":
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=30",
			host, port, username, password, conn.Database)
	case "MYSQL":
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=30s&parseTime=true",
			username, password, host, port, conn.Database)
	default:
		return fmt.Errorf("unsupported destination type: %s", pipeline.DestinationType)
	}

	destDB, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to destination: %w", err)
	}
	defer destDB.Close()

	destDB.SetMaxOpenConns(5)
	destDB.SetConnMaxLifetime(5 * time.Minute)

	if err := destDB.PingContext(ctx); err != nil {
		return fmt.Errorf("destination ping failed: %w", err)
	}

	tableName := destConf.TableName
	if tableName == "" {
		tableName = fmt.Sprintf("pipeline_%s", pipeline.ID)
	}

	// Determine columns
	var columns []string
	for col := range data[0] {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	// Batch insert
	batchSize := 500
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		if err := pe.insertBatch(ctx, destDB, tableName, columns, batch); err != nil {
			return fmt.Errorf("external batch insert failed at row %d: %w", i, err)
		}
	}

	return nil
}

// insertBatch inserts a batch of rows into a table
func (pe *PipelineExecutor) insertBatch(ctx context.Context, db *sql.DB, tableName string, columns []string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// Build parameterized INSERT
	var quotedCols []string
	for _, col := range columns {
		quotedCols = append(quotedCols, fmt.Sprintf(`"%s"`, col))
	}

	var valueRows []string
	var args []interface{}
	paramIdx := 1

	for _, row := range rows {
		var placeholders []string
		for _, col := range columns {
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
			val := row[col]
			if val == nil {
				args = append(args, nil)
			} else {
				args = append(args, fmt.Sprintf("%v", val))
			}
			paramIdx++
		}
		valueRows = append(valueRows, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
	}

	insertSQL := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES %s`,
		tableName,
		strings.Join(quotedCols, ", "),
		strings.Join(valueRows, ", "))

	_, err := db.ExecContext(ctx, insertSQL)
	return err
}

// Helper: updateProgress updates the execution context progress
func (pe *PipelineExecutor) updateProgress(executionID string, progress int, status string) {
	pe.mu.Lock()
	if ctx, ok := pe.activeRuns[executionID]; ok {
		ctx.Progress = progress
		ctx.Status = status
	}
	pe.mu.Unlock()

	// Also update DB
	database.DB.Model(&models.JobExecution{}).
		Where("id = ?", executionID).
		Updates(map[string]interface{}{
			"progress": progress,
			"status":   status,
		})
}

// Helper: finalizeResult calculates duration and serializes logs
func (pe *PipelineExecutor) finalizeResult(result *ExecutionResult, startTime time.Time) *ExecutionResult {
	result.DurationMs = int(time.Since(startTime).Milliseconds())
	return result
}

// Helper: appendLog adds a log entry
func (r *ExecutionResult) appendLog(level, step, message, details string) {
	r.Logs = append(r.Logs, models.ExecutionLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Step:      step,
		Details:   details,
	})
}

// Helper: evaluateCondition checks if a value matches a filter condition
func (pe *PipelineExecutor) evaluateCondition(cellVal interface{}, operator string, filterVal interface{}) bool {
	cellStr := fmt.Sprintf("%v", cellVal)
	filterStr := fmt.Sprintf("%v", filterVal)

	switch operator {
	case "eq", "=", "==":
		return cellStr == filterStr
	case "neq", "!=", "<>":
		return cellStr != filterStr
	case "contains":
		return strings.Contains(strings.ToLower(cellStr), strings.ToLower(filterStr))
	case "not_contains":
		return !strings.Contains(strings.ToLower(cellStr), strings.ToLower(filterStr))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(cellStr), strings.ToLower(filterStr))
	case "gt", ">":
		return pe.toFloat64(cellVal) > pe.toFloat64(filterVal)
	case "gte", ">=":
		return pe.toFloat64(cellVal) >= pe.toFloat64(filterVal)
	case "lt", "<":
		return pe.toFloat64(cellVal) < pe.toFloat64(filterVal)
	case "lte", "<=":
		return pe.toFloat64(cellVal) <= pe.toFloat64(filterVal)
	case "is_null":
		return cellVal == nil || cellStr == "" || cellStr == "<nil>"
	case "is_not_null":
		return cellVal != nil && cellStr != "" && cellStr != "<nil>"
	default:
		return true
	}
}

// Helper: toFloat64 converts an interface value to float64
func (pe *PipelineExecutor) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}

// Helper: castValue converts a value to the target type
func (pe *PipelineExecutor) castValue(val interface{}, targetType string) interface{} {
	if val == nil {
		return nil
	}

	switch strings.ToUpper(targetType) {
	case "STRING", "TEXT":
		return fmt.Sprintf("%v", val)
	case "INT", "INTEGER":
		return int(pe.toFloat64(val))
	case "FLOAT", "DOUBLE", "DECIMAL":
		return pe.toFloat64(val)
	case "BOOL", "BOOLEAN":
		s := strings.ToLower(fmt.Sprintf("%v", val))
		return s == "true" || s == "1" || s == "yes"
	default:
		return val
	}
}
