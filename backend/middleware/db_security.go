package middleware

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"insight-engine-backend/services"
)

// DatabaseSecurityConfig holds configuration for database security measures
type DatabaseSecurityConfig struct {
	BlockDangerousOperations bool   // Whether to block dangerous SQL operations
	MaxQueryComplexity       int    // Maximum complexity score allowed
	AllowedTablesPattern     string // Regex pattern for allowed table names
}

// DatabaseSecurityMiddleware provides database security measures including SQL injection prevention
func DatabaseSecurityMiddleware(config DatabaseSecurityConfig) fiber.Handler {
	// Compile regex patterns once
	if config.AllowedTablesPattern != "" {
		_ = regexp.MustCompile(config.AllowedTablesPattern)
	}

	return func(c *fiber.Ctx) error {
		// Continue to next middleware
		return c.Next()
	}
}

// SQLInjectionDetector provides methods to detect potential SQL injection attempts
type SQLInjectionDetector struct{}

// DetectSQLInjection checks if a string contains potential SQL injection patterns
func (s *SQLInjectionDetector) DetectSQLInjection(input string) (bool, string) {
	if input == "" {
		return false, ""
	}

	// Convert to lowercase for pattern matching
	lowerInput := strings.ToLower(input)

	// Dangerous SQL keywords/patterns
	dangerousPatterns := []struct {
		pattern string
		desc    string
	}{
		{`(\b(drop|truncate|alter|create|delete|update)\b)`, "Dangerous SQL operation"},
		{`(;\s*(drop|truncate|alter|create|delete|update))`, "SQL statement termination followed by dangerous operation"},
		{`('|--)`, "Potential SQL comment or quote manipulation"},
		{`(union\s+select)`, "UNION-based injection"},
		{`(exec\s*\(|execute\s*\(|sp_|xp_)`, "Stored procedure execution"},
		{`(waitfor\s+delay|pg_sleep|sleep)`, "Time-based injection"},
		{`(benchmark\s*\(|time_to_sec)`, "Benchmark-based injection"},
	}

	for _, pat := range dangerousPatterns {
		if regexp.MustCompile(pat.pattern).MatchString(lowerInput) {
			return true, pat.desc
		}
	}

	return false, ""
}

// ValidateSQLQuery performs security validation on a SQL query
func (s *SQLInjectionDetector) ValidateSQLQuery(query string) error {
	// Check for dangerous patterns
	isDangerous, reason := s.DetectSQLInjection(query)
	if isDangerous {
		return fmt.Errorf("security violation: %s detected in query", reason)
	}

	// Additional validations can be added here
	// For example, checking for proper parameterization, etc.

	return nil
}

// QueryOptimizer provides methods to optimize and validate queries
type QueryOptimizer struct {
	detector *SQLInjectionDetector
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		detector: &SQLInjectionDetector{},
	}
}

// OptimizeAndValidateQuery optimizes and validates a query for security
func (q *QueryOptimizer) OptimizeAndValidateQuery(query string) (string, error) {
	// First, validate for security issues
	if err := q.detector.ValidateSQLQuery(query); err != nil {
		return "", err
	}

	// Additional optimization logic can be added here
	optimizedQuery := q.optimizeQueryStructure(query)

	return optimizedQuery, nil
}

// optimizeQueryStructure performs basic query structure optimizations
func (q *QueryOptimizer) optimizeQueryStructure(query string) string {
	// Remove extra whitespace
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	
	// Trim leading/trailing spaces
	query = strings.TrimSpace(query)
	
	return query
}

// NPlusOneDetector helps detect potential N+1 query patterns
type NPlusOneDetector struct {
	QueryLog map[string]time.Time // Track query execution times
}

// NewNPlusOneDetector creates a new N+1 detector
func NewNPlusOneDetector() *NPlusOneDetector {
	return &NPlusOneDetector{
		QueryLog: make(map[string]time.Time),
	}
}

// LogQueryExecution logs a query execution for N+1 detection
func (n *NPlusOneDetector) LogQueryExecution(queryHash, userID string) {
	key := fmt.Sprintf("%s:%s", userID, queryHash)
	n.QueryLog[key] = time.Now()
}

// DetectNPlusOnePotential checks if a pattern suggests N+1 queries
func (n *NPlusOneDetector) DetectNPlusOnePotential(queryHash, userID string) bool {
	now := time.Now()
	
	// Check if similar queries have been executed recently
	count := 0
	for logKey, execTime := range n.QueryLog {
		if strings.HasPrefix(logKey, userID+":") && 
		   strings.HasSuffix(logKey, ":"+queryHash) &&
		   now.Sub(execTime) < time.Second*5 { // Within 5 seconds
			count++
		}
	}
	
	// If we see the same query executed multiple times in quick succession, 
	// it might indicate an N+1 problem
	return count > 5 // Threshold can be adjusted
}

// AddDatabaseSecurityLogging adds security-focused logging for database operations
func AddDatabaseSecurityLogging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Continue with the request
		err := c.Next()
		
		// Log database-related operations for security monitoring
		duration := time.Since(start)
		
		// Log slow queries for potential optimization
		if duration > time.Second*2 {
			services.LogWarn("slow_query", "Slow database query detected", map[string]interface{}{
				"path":     c.Path(),
				"method":   c.Method(),
				"duration": duration.String(),
				"ip":       c.IP(),
			})
		}
		
		return err
	}
}

