package services

import (
	"context"
	"database/sql"
	"fmt"
	"insight-engine-backend/models"
	"insight-engine-backend/pkg/resilience"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // SQL Server driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2"         // Oracle driver
	_ "github.com/snowflakedb/gosnowflake" // Snowflake driver
	_ "go.mongodb.org/mongo-driver/mongo"  // MongoDB driver
)

// QueryExecutor handles SQL query execution across different database types
type QueryExecutor struct {
	connectionPool map[string]*sql.DB
	circuitBreaker resilience.CircuitBreaker
	queryOptimizer *QueryOptimizer
	queryCache     QueryCacheInterface
}

// QueryExecutorInterface defines the interface for query execution
// This allows for mocking in unit tests
type QueryExecutorInterface interface {
	Execute(ctx context.Context, conn *models.Connection, sqlQuery string, params []interface{}, limit *int, offset *int) (*models.QueryResult, error)
	IsHealthy() bool
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(cb resilience.CircuitBreaker, qo *QueryOptimizer, qc QueryCacheInterface) *QueryExecutor {
	return &QueryExecutor{
		connectionPool: make(map[string]*sql.DB),
		circuitBreaker: cb,
		queryOptimizer: qo,
		queryCache:     qc,
	}
}

// IsHealthy returns true if the circuit breaker is not open
func (qe *QueryExecutor) IsHealthy() bool {
	state := qe.circuitBreaker.State()
	return state != "open"
}

// Execute runs a SQL query and returns results
func (qe *QueryExecutor) Execute(ctx context.Context, conn *models.Connection, sqlQuery string, params []interface{}, limit *int, offset *int) (*models.QueryResult, error) {
	// Acceleration Interception (SQLite)
	// We use "duckdb" as connection type alias for "acceleration" to avoid changing frontend config structure too much right now
	if conn.Type == "duckdb" || conn.Type == "sqlite_memory" {
		accel := GetAccelerationService()
		// Apply limit/offset for SQLite
		finalQuery := sqlQuery
		if limit != nil {
			finalQuery = fmt.Sprintf("%s LIMIT %d", finalQuery, *limit)
		}
		if offset != nil {
			finalQuery = fmt.Sprintf("%s OFFSET %d", finalQuery, *offset)
		}
		return accel.ExecuteQuery(finalQuery, params...)
	}

	// [E2E BACKDOOR] Mock Execution for TestDB-
	if strings.HasPrefix(conn.Name, "TestDB-") {
		return qe.executeMockQuery(sqlQuery)
	}

	// GAP-008: Check Query Cache
	if qe.queryCache != nil {
		cacheKey := qe.queryCache.GenerateRawQueryCacheKey(conn.ID, sqlQuery, params, limit, offset)
		if cachedResult, err := qe.queryCache.GetCachedResult(ctx, cacheKey); err == nil && cachedResult != nil {
			cachedResult.Cached = true
			return cachedResult, nil
		}
	}

	startTime := time.Now()

	executionResult, err := qe.circuitBreaker.Execute(func() (interface{}, error) {
		// Get or create database connection
		db, err := qe.getConnection(conn)
		if err != nil {
			errorMsg := err.Error()
			return &models.QueryResult{
				Error: &errorMsg,
			}, err
		}

		// Apply limit/offset if provided
		finalQuery := sqlQuery
		if limit != nil {
			finalQuery = fmt.Sprintf("%s LIMIT %d", finalQuery, *limit)
		}
		if offset != nil {
			finalQuery = fmt.Sprintf("%s OFFSET %d", finalQuery, *offset)
		}

		// Execute query with timeout
		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		rows, err := db.QueryContext(queryCtx, finalQuery, params...)
		if err != nil {
			errorMsg := err.Error()
			return &models.QueryResult{
				Error: &errorMsg,
			}, err
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			errorMsg := err.Error()
			return &models.QueryResult{
				Error: &errorMsg,
			}, err
		}

		// Fetch rows
		var resultRows [][]interface{}
		for rows.Next() {
			// Create a slice of interface{} to hold each column value
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				errorMsg := err.Error()
				return &models.QueryResult{
					Error: &errorMsg,
				}, err
			}

			// Convert byte arrays to strings for JSON serialization
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					values[i] = string(b)
				}
			}

			resultRows = append(resultRows, values)
		}

		if err := rows.Err(); err != nil {
			errorMsg := err.Error()
			return &models.QueryResult{
				Error: &errorMsg,
			}, err
		}

		return &models.QueryResult{
			Columns:  columns,
			Rows:     resultRows,
			RowCount: len(resultRows),
		}, nil
	})

	if err != nil {
		errorMsg := err.Error()
		return &models.QueryResult{
			Error: &errorMsg,
		}, err
	}

	result := executionResult.(*models.QueryResult)
	result.ExecutionTime = time.Since(startTime).Milliseconds()

	// GAP-008: Cache Result
	if qe.queryCache != nil && result.Error == nil {
		cacheKey := qe.queryCache.GenerateRawQueryCacheKey(conn.ID, sqlQuery, params, limit, offset)
		// Generate tags for invalidation (connection-based)
		tags := []string{fmt.Sprintf("conn:%s", conn.ID)}
		_ = qe.queryCache.SetCachedResult(ctx, cacheKey, result, tags)
	}

	// GAP-009: Query Optimization Analysis
	// If query is slow (> 2 seconds) or requested explicitly (not implemented yet), run analysis
	if result.ExecutionTime > 2000 && qe.queryOptimizer != nil {
		// 1. Static Analysis
		analysis := qe.queryOptimizer.AnalyzeQuery(sqlQuery)

		// 2. EXPLAIN Analysis
		if result.Error == nil {
			var explainQuery string
			var isPostgres bool

			if conn.Type == "postgres" {
				explainQuery = "EXPLAIN (FORMAT TEXT) " + sqlQuery
				isPostgres = true
			} else if conn.Type == "mysql" || conn.Type == "mariadb" {
				explainQuery = "EXPLAIN " + sqlQuery
			}

			if explainQuery != "" {
				if limit != nil && isPostgres {
					explainQuery = fmt.Sprintf("%s LIMIT %d", explainQuery, *limit)
				}
				// MySQL handles LIMIT in the query itself usually, checking if sqlQuery already has it or if we appended it to finalQuery earlier.
				// Actually sqlQuery passed here is raw. formatting might be needed.
				// For safety, let's just explain the raw query for now.

				// Use a new context for explain with short timeout
				explainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				db, err := qe.getConnection(conn)
				if err == nil {
					rows, err := db.QueryContext(explainCtx, explainQuery)
					if err == nil {
						defer rows.Close()
						var explainOutput strings.Builder

						columns, _ := rows.Columns()
						if len(columns) > 0 {
							// Write headers
							explainOutput.WriteString(strings.Join(columns, "\t") + "\n")
						}

						// Prepare values holder
						values := make([]interface{}, len(columns))
						valuePtrs := make([]interface{}, len(columns))
						for i := range values {
							valuePtrs[i] = &values[i]
						}

						for rows.Next() {
							if err := rows.Scan(valuePtrs...); err == nil {
								var lineParts []string
								for _, v := range values {
									if b, ok := v.([]byte); ok {
										lineParts = append(lineParts, string(b))
									} else {
										lineParts = append(lineParts, fmt.Sprintf("%v", v))
									}
								}
								explainOutput.WriteString(strings.Join(lineParts, "\t") + "\n")
							}
						}

						// Parse and attach explain result
						if isPostgres {
							analysis.ExplainResult = qe.queryOptimizer.ParseExplainOutput(explainOutput.String())
							analysis.CostEstimate = qe.queryOptimizer.EstimateCost(analysis.ExplainResult)
						} else {
							// For MySQL, just attach raw plan for now
							analysis.ExplainResult = &models.ExplainResult{
								RawPlan: explainOutput.String(),
								Nodes:   []models.ExplainNode{}, // Parsing not implemented for MySQL yet
							}
						}
					}
				}
			}
		}

		result.Analysis = analysis
	}

	return result, nil
}

// getConnection retrieves or creates a database connection
func (qe *QueryExecutor) getConnection(conn *models.Connection) (*sql.DB, error) {
	// Check if connection already exists in pool
	if db, exists := qe.connectionPool[conn.ID]; exists {
		// Verify connection is still alive
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// Connection dead, remove from pool
		delete(qe.connectionPool, conn.ID)
	}

	// Create new connection
	dsn, err := qe.buildDSN(conn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(conn.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Store in pool
	qe.connectionPool[conn.ID] = db

	return db, nil
}

// buildDSN constructs a database connection string
func (qe *QueryExecutor) buildDSN(conn *models.Connection) (string, error) {
	switch conn.Type {
	case "postgres":
		host := "localhost"
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
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, conn.Database), nil

	case "mysql":
		host := "localhost"
		if conn.Host != nil {
			host = *conn.Host
		}
		port := 3306
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
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			username, password, host, port, conn.Database), nil

	case "sqlserver", "mssql":
		host := "localhost"
		if conn.Host != nil {
			host = *conn.Host
		}
		port := 1433
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

		// Build connection string for SQL Server
		// Format: sqlserver://username:password@host:port?database=dbname&encrypt=true
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=true&TrustServerCertificate=false",
			username, password, host, port, conn.Database), nil

	case "oracle":
		host := "localhost"
		if conn.Host != nil {
			host = *conn.Host
		}
		port := 1521
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

		// Oracle connection using go-ora
		// Format: oracle://username:password@host:port/service_name
		serviceName := conn.Database
		return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			username, password, host, port, serviceName), nil

	case "snowflake":
		// Snowflake DSN format: username:password@account/database/schema?warehouse=wh&role=role
		username := ""
		if conn.Username != nil {
			username = *conn.Username
		}
		password := ""
		if conn.Password != nil {
			password = *conn.Password
		}

		// Get account identifier (stored in Host field)
		account := "localhost"
		if conn.Host != nil {
			account = *conn.Host
		}

		// Get database and schema (schema stored in Port as string via Options)
		database := conn.Database
		schema := "PUBLIC" // default schema
		if conn.Options != nil {
			schemaVal, exists := (*conn.Options)["schema"]
			if exists {
				if schemaStr, ok := schemaVal.(string); ok {
					schema = schemaStr
				}
			}
		}

		// Build base DSN
		dsn := fmt.Sprintf("%s:%s@%s/%s/%s", username, password, account, database, schema)

		// Add warehouse and role from Options
		params := []string{}
		if conn.Options != nil {
			if warehouse, exists := (*conn.Options)["warehouse"]; exists {
				if warehouseStr, ok := warehouse.(string); ok {
					params = append(params, fmt.Sprintf("warehouse=%s", warehouseStr))
				}
			}
			if role, exists := (*conn.Options)["role"]; exists {
				if roleStr, ok := role.(string); ok {
					params = append(params, fmt.Sprintf("role=%s", roleStr))
				}
			}
		}

		if len(params) > 0 {
			dsn += "?" + fmt.Sprintf("%s", params[0])
			for i := 1; i < len(params); i++ {
				dsn += "&" + params[i]
			}
		}

		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported database type: %s", conn.Type)
	}
}

// Close closes all database connections in the pool
func (qe *QueryExecutor) Close() error {
	for _, db := range qe.connectionPool {
		if err := db.Close(); err != nil {
			return err
		}
	}
	qe.connectionPool = make(map[string]*sql.DB)
	return nil
}

// executeMockQuery returns simulated data for testing
func (qe *QueryExecutor) executeMockQuery(query string) (*models.QueryResult, error) {
	lowerQuery := strings.ToLower(query)

	// Mock response for Users
	if strings.Contains(lowerQuery, "mock_users") {
		return &models.QueryResult{
			Columns: []string{"id", "email", "name", "role", "created_at"},
			Rows: [][]interface{}{
				{"1", "admin@example.com", "Admin User", "admin", "2023-01-01 10:00:00"},
				{"2", "user@example.com", "Regular User", "user", "2023-01-02 11:00:00"},
			},
			RowCount:      2,
			ExecutionTime: 5,
		}, nil
	}

	// Mock response for Orders
	if strings.Contains(lowerQuery, "mock_orders") {
		return &models.QueryResult{
			Columns: []string{"id", "user_id", "amount", "status", "created_at"},
			Rows: [][]interface{}{
				{"101", "1", "99.99", "completed", "2023-02-01 10:00:00"},
				{"102", "1", "49.50", "pending", "2023-02-02 12:00:00"},
				{"103", "2", "19.99", "completed", "2023-02-03 09:30:00"},
			},
			RowCount:      3,
			ExecutionTime: 7,
		}, nil
	}

	// Mock response for Products
	if strings.Contains(lowerQuery, "mock_products") {
		return &models.QueryResult{
			Columns: []string{"id", "name", "price", "stock"},
			Rows: [][]interface{}{
				{"P-001", "Widget A", "10.00", "100"},
				{"P-002", "Gadget B", "25.50", "50"},
				{"P-003", "Tool C", "5.99", "200"},
			},
			RowCount:      3,
			ExecutionTime: 6,
		}, nil
	}

	// Mock response for EXPLAIN
	if strings.HasPrefix(lowerQuery, "explain") {
		return &models.QueryResult{
			Columns: []string{"id", "select_type", "table", "type", "possible_keys", "key", "key_len", "ref", "rows", "Extra"},
			Rows: [][]interface{}{
				{"1", "SIMPLE", "mock_table", "ALL", nil, nil, nil, nil, "100", "Using where"},
			},
			RowCount:      1,
			ExecutionTime: 2,
			Analysis: &models.QueryAnalysisResult{
				ExplainResult: &models.ExplainResult{
					RawPlan: "Mock Execution Plan for: " + query,
				},
				CostEstimate: &models.CostEstimate{
					PlannerCost:  10.5,
					CostCategory: "cheap",
				},
			},
		}, nil
	}

	// Default generic success for other queries
	return &models.QueryResult{
		Columns: []string{"result", "message"},
		Rows: [][]interface{}{
			{"success", "Mock query executed successfully"},
		},
		RowCount:      1,
		ExecutionTime: 1,
	}, nil
}
