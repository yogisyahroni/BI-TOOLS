package services

import (
	"fmt"
	"insight-engine-backend/models"
	"regexp"
	"strings"
)

// QueryOptimizer provides SQL query optimization suggestions
type QueryOptimizer struct {
	// Common optimization patterns
	patterns []OptimizationPattern
}

// OptimizationPattern represents a query optimization pattern
type OptimizationPattern struct {
	Name        string
	Description string
	Severity    string // "high", "medium", "low"
	Pattern     *regexp.Regexp
	Suggestion  string
	Example     string
}

// OptimizationSuggestion moved to models

// QueryAnalysisResult moved to models

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		patterns: []OptimizationPattern{
			{
				Name:        "SELECT *",
				Description: "Avoid SELECT * - specify only needed columns",
				Severity:    "medium",
				Pattern:     regexp.MustCompile(`(?i)SELECT\s+\*`),
				Suggestion:  "Replace SELECT * with specific column names to reduce data transfer and improve performance",
				Example:     "SELECT id, name, email FROM users",
			},
			{
				Name:        "Missing WHERE clause",
				Description: "Query without WHERE clause may scan entire table",
				Severity:    "high",
				Pattern:     regexp.MustCompile(`(?i)SELECT.*FROM\s+\w+\s*(?:;|$)`),
				Suggestion:  "Add WHERE clause to filter data and reduce rows scanned",
				Example:     "SELECT * FROM users WHERE created_at > '2024-01-01'",
			},
			{
				Name:        "OR in WHERE clause",
				Description: "OR conditions can prevent index usage",
				Severity:    "medium",
				Pattern:     regexp.MustCompile(`(?i)WHERE.*\sOR\s`),
				Suggestion:  "Consider using UNION or IN clause instead of OR for better index utilization",
				Example:     "SELECT * FROM users WHERE id IN (1, 2, 3)",
			},
			{
				Name:        "Function on indexed column",
				Description: "Functions on indexed columns prevent index usage",
				Severity:    "high",
				Pattern:     regexp.MustCompile(`(?i)WHERE\s+\w+\([^)]+\)\s*=`),
				Suggestion:  "Avoid functions on indexed columns in WHERE clause",
				Example:     "Use WHERE created_at >= DATE instead of WHERE DATE(created_at) = DATE",
			},
			{
				Name:        "LIKE with leading wildcard",
				Description: "LIKE with leading % prevents index usage",
				Severity:    "medium",
				Pattern:     regexp.MustCompile(`(?i)LIKE\s+'%`),
				Suggestion:  "Avoid leading wildcards in LIKE patterns. Consider full-text search for better performance",
				Example:     "Use LIKE 'prefix%' or full-text search instead of LIKE '%search%'",
			},
			{
				Name:        "Subquery in SELECT",
				Description: "Subqueries in SELECT can be slow",
				Severity:    "medium",
				Pattern:     regexp.MustCompile(`(?i)SELECT.*\(SELECT`),
				Suggestion:  "Consider using JOINs instead of subqueries in SELECT clause",
				Example:     "Use LEFT JOIN instead of correlated subquery",
			},
			{
				Name:        "NOT IN with subquery",
				Description: "NOT IN with subquery can be very slow",
				Severity:    "high",
				Pattern:     regexp.MustCompile(`(?i)NOT\s+IN\s*\(SELECT`),
				Suggestion:  "Use NOT EXISTS or LEFT JOIN with NULL check instead of NOT IN",
				Example:     "SELECT * FROM users u WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)",
			},
			{
				Name:        "DISTINCT without necessity",
				Description: "DISTINCT can be expensive, ensure it's necessary",
				Severity:    "low",
				Pattern:     regexp.MustCompile(`(?i)SELECT\s+DISTINCT`),
				Suggestion:  "Verify if DISTINCT is necessary. Consider using GROUP BY if aggregating",
				Example:     "Use GROUP BY for aggregations instead of DISTINCT",
			},
		},
	}
}

// AnalyzeQuery analyzes a SQL query and provides optimization suggestions
func (qo *QueryOptimizer) AnalyzeQuery(query string) *models.QueryAnalysisResult {
	query = strings.TrimSpace(query)
	suggestions := []models.OptimizationSuggestion{}

	// Check each pattern
	for _, pattern := range qo.patterns {
		if pattern.Pattern.MatchString(query) {
			suggestion := models.OptimizationSuggestion{
				Type:        qo.categorizePattern(pattern.Name),
				Severity:    pattern.Severity,
				Title:       pattern.Name,
				Description: pattern.Description,
				Original:    qo.extractMatchedPart(query, pattern.Pattern),
				Optimized:   pattern.Example,
				Impact:      qo.estimateImpact(pattern.Severity),
				Example:     pattern.Suggestion,
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// Calculate performance score (100 - penalties)
	score := 100
	for _, s := range suggestions {
		switch s.Severity {
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}

	// Determine complexity level
	complexity := qo.determineComplexity(query)

	// Estimate improvement
	improvement := qo.estimateImprovement(suggestions)

	return &models.QueryAnalysisResult{
		Query:                query,
		Suggestions:          suggestions,
		PerformanceScore:     score,
		ComplexityLevel:      complexity,
		EstimatedImprovement: improvement,
	}
}

// categorizePattern categorizes the pattern type
func (qo *QueryOptimizer) categorizePattern(name string) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "index") {
		return "index"
	}
	if strings.Contains(name, "join") {
		return "join"
	}
	if strings.Contains(name, "select") {
		return "select"
	}
	if strings.Contains(name, "where") {
		return "where"
	}
	if strings.Contains(name, "subquery") {
		return "subquery"
	}
	return "general"
}

// extractMatchedPart extracts the matched part of the query
func (qo *QueryOptimizer) extractMatchedPart(query string, pattern *regexp.Regexp) string {
	match := pattern.FindString(query)
	if match == "" {
		return query
	}
	return match
}

// estimateImpact estimates the performance impact
func (qo *QueryOptimizer) estimateImpact(severity string) string {
	switch severity {
	case "high":
		return "50-80% improvement possible"
	case "medium":
		return "20-50% improvement possible"
	case "low":
		return "5-20% improvement possible"
	default:
		return "Minor improvement"
	}
}

// determineComplexity determines query complexity
func (qo *QueryOptimizer) determineComplexity(query string) string {
	query = strings.ToLower(query)

	// Count complexity indicators
	complexity := 0

	if strings.Contains(query, "join") {
		complexity += 2
	}
	if strings.Contains(query, "subquery") || strings.Contains(query, "(select") {
		complexity += 3
	}
	if strings.Contains(query, "group by") {
		complexity += 1
	}
	if strings.Contains(query, "having") {
		complexity += 2
	}
	if strings.Contains(query, "union") {
		complexity += 2
	}

	if complexity >= 6 {
		return "high"
	} else if complexity >= 3 {
		return "medium"
	}
	return "low"
}

// estimateImprovement estimates overall improvement potential
func (qo *QueryOptimizer) estimateImprovement(suggestions []models.OptimizationSuggestion) string {
	if len(suggestions) == 0 {
		return "Query is already optimized"
	}

	highCount := 0
	mediumCount := 0

	for _, s := range suggestions {
		if s.Severity == "high" {
			highCount++
		} else if s.Severity == "medium" {
			mediumCount++
		}
	}

	if highCount >= 2 {
		return "60-80% improvement possible"
	} else if highCount == 1 {
		return "40-60% improvement possible"
	} else if mediumCount >= 2 {
		return "20-40% improvement possible"
	} else if mediumCount == 1 {
		return "10-20% improvement possible"
	}
	return "5-10% improvement possible"
}

// OptimizeQuery attempts to automatically optimize a query
func (qo *QueryOptimizer) OptimizeQuery(query string) string {
	optimized := query

	// Apply automatic optimizations
	// 1. Replace SELECT * with specific columns (requires schema knowledge)
	// For now, we just suggest - actual optimization needs schema context

	// 2. Add LIMIT if missing (for safety)
	if !strings.Contains(strings.ToLower(optimized), "limit") {
		optimized = strings.TrimSuffix(optimized, ";")
		optimized = fmt.Sprintf("%s LIMIT 1000;", optimized)
	}

	return optimized
}

// ============================================================
// EXPLAIN Integration (GAP-009)
// ============================================================

// ExplainResult, ExplainNode, IndexRecommendation, CostEstimate moved to models

// ParseExplainOutput parses raw EXPLAIN text output into structured ExplainResult
func (qo *QueryOptimizer) ParseExplainOutput(rawPlan string) *models.ExplainResult {
	result := &models.ExplainResult{
		RawPlan: rawPlan,
		Nodes:   []models.ExplainNode{},
	}

	lines := strings.Split(rawPlan, "\n")

	// Regex patterns for EXPLAIN output parsing
	nodePattern := regexp.MustCompile(`(?i)(Seq Scan|Index Scan|Index Only Scan|Bitmap Heap Scan|Bitmap Index Scan|Hash Join|Merge Join|Nested Loop|Sort|Aggregate|Hash|Materialize|Limit|Append|Subquery Scan|CTE Scan|Gather|Gather Merge)\s+(?:on\s+)?(\S+)?`)
	costPattern := regexp.MustCompile(`cost=(\d+\.?\d*)\.\.(\d+\.?\d*)\s+rows=(\d+)\s+width=(\d+)`)
	actualPattern := regexp.MustCompile(`actual time=(\d+\.?\d*)\.\.(\d+\.?\d*)\s+rows=(\d+)`)
	filterPattern := regexp.MustCompile(`Filter:\s+(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Planning") || strings.HasPrefix(line, "Execution") {
			continue
		}

		nodeMatch := nodePattern.FindStringSubmatch(line)
		if nodeMatch == nil {
			continue
		}

		node := models.ExplainNode{
			NodeType: nodeMatch[1],
		}
		if len(nodeMatch) > 2 {
			node.On = nodeMatch[2]
		}

		// Parse cost
		costMatch := costPattern.FindStringSubmatch(line)
		if costMatch != nil {
			cost, _ := parseFloat(costMatch[2])
			rows, _ := parseInt(costMatch[3])
			width, _ := parseInt(costMatch[4])

			node.Cost = cost
			node.Rows = int64(rows)
			node.Width = width

			// Track total cost from root node
			if result.TotalCost == 0 {
				result.TotalCost = cost
				result.RowEstimate = int64(rows)
			}
		}

		// Parse actual time (if ANALYZE was used)
		actualMatch := actualPattern.FindStringSubmatch(line)
		if actualMatch != nil {
			actualTime, _ := parseFloat(actualMatch[2])
			actualRows, _ := parseInt(actualMatch[3])
			node.ActualTimeMs = actualTime
			node.ActualRows = int64(actualRows)

			if result.ActualTimeMs == 0 {
				result.ActualTimeMs = actualTime
				result.ActualRows = int64(actualRows)
			}
		}

		// Parse filter
		filterMatch := filterPattern.FindStringSubmatch(line)
		if filterMatch != nil {
			node.Filter = filterMatch[1]
		}

		// Detect warnings
		if strings.EqualFold(node.NodeType, "Seq Scan") && node.Rows > 1000 {
			node.Warning = fmt.Sprintf("Sequential scan on %s with %d estimated rows — consider adding an index", node.On, node.Rows)
			result.Warnings = append(result.Warnings, node.Warning)
		}

		result.Nodes = append(result.Nodes, node)
	}

	return result
}

// RecommendIndexes analyzes a SQL query and suggests indexes based on WHERE, JOIN, and ORDER BY columns
func (qo *QueryOptimizer) RecommendIndexes(query string) []models.IndexRecommendation {
	var recs []models.IndexRecommendation
	upperQ := strings.ToUpper(query)
	lowerQ := strings.ToLower(query)

	// Extract table names from FROM clause
	fromPattern := regexp.MustCompile(`(?i)FROM\s+(\w+)`)
	fromMatches := fromPattern.FindAllStringSubmatch(query, -1)
	tableSet := make(map[string]bool)
	for _, m := range fromMatches {
		tableSet[m[1]] = true
	}

	// Extract JOIN tables
	joinPattern := regexp.MustCompile(`(?i)JOIN\s+(\w+)`)
	joinMatches := joinPattern.FindAllStringSubmatch(query, -1)
	for _, m := range joinMatches {
		tableSet[m[1]] = true
	}

	// Extract WHERE column references (simple heuristic)
	wherePattern := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:GROUP|ORDER|HAVING|LIMIT|;|$)`)
	whereMatch := wherePattern.FindStringSubmatch(query)
	if whereMatch != nil {
		whereClause := whereMatch[1]
		colPattern := regexp.MustCompile(`(\w+)\s*(?:=|>|<|>=|<=|<>|!=|LIKE|IN|BETWEEN|IS)`)
		colMatches := colPattern.FindAllStringSubmatch(whereClause, -1)
		seen := make(map[string]bool)
		for _, cm := range colMatches {
			colName := strings.ToLower(cm[1])
			if isReservedWord(colName) || seen[colName] {
				continue
			}
			seen[colName] = true

			// Find the most likely table for this column
			table := guessTableForColumn(colName, tableSet)
			recs = append(recs, models.IndexRecommendation{
				Table:     table,
				Columns:   colName,
				Reason:    fmt.Sprintf("Column '%s' used in WHERE clause", colName),
				CreateSQL: fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_%s ON %s (%s);", table, colName, table, colName),
				Priority:  "high",
			})
		}
	}

	// Extract ORDER BY columns
	orderPattern := regexp.MustCompile(`(?i)ORDER\s+BY\s+(\w+(?:\s*,\s*\w+)*)`)
	orderMatch := orderPattern.FindStringSubmatch(query)
	if orderMatch != nil {
		cols := strings.Split(orderMatch[1], ",")
		for _, col := range cols {
			col = strings.TrimSpace(strings.ToLower(col))
			col = strings.Fields(col)[0] // Remove ASC/DESC
			if isReservedWord(col) {
				continue
			}
			table := guessTableForColumn(col, tableSet)
			recs = append(recs, models.IndexRecommendation{
				Table:     table,
				Columns:   col,
				Reason:    fmt.Sprintf("Column '%s' used in ORDER BY clause", col),
				CreateSQL: fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_%s ON %s (%s);", table, col, table, col),
				Priority:  "medium",
			})
		}
	}

	// Check for JOIN conditions (foreign key indexes)
	joinCondPattern := regexp.MustCompile(`(?i)ON\s+(\w+)\.(\w+)\s*=\s*(\w+)\.(\w+)`)
	joinCondMatches := joinCondPattern.FindAllStringSubmatch(query, -1)
	for _, jm := range joinCondMatches {
		for i := 1; i <= 3; i += 2 {
			tbl := strings.ToLower(jm[i])
			col := strings.ToLower(jm[i+1])
			recs = append(recs, models.IndexRecommendation{
				Table:     tbl,
				Columns:   col,
				Reason:    fmt.Sprintf("Column '%s.%s' used in JOIN condition", tbl, col),
				CreateSQL: fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_%s ON %s (%s);", tbl, col, tbl, col),
				Priority:  "high",
			})
		}
	}

	// Suppress noise
	_ = upperQ
	_ = lowerQ

	return deduplicateIndexRecs(recs)
}

func (qo *QueryOptimizer) EstimateCost(explainResult *models.ExplainResult) *models.CostEstimate {
	if explainResult == nil {
		return &models.CostEstimate{CostCategory: "unknown"}
	}

	est := &models.CostEstimate{
		PlannerCost:   explainResult.TotalCost,
		EstimatedRows: explainResult.RowEstimate,
	}

	// Calculate width from first node
	if len(explainResult.Nodes) > 0 {
		est.EstimatedWidth = explainResult.Nodes[0].Width
	}
	est.EstimatedDataSize = est.EstimatedRows * int64(est.EstimatedWidth)

	// Categorize cost
	switch {
	case est.PlannerCost < 10:
		est.CostCategory = "cheap"
	case est.PlannerCost < 100:
		est.CostCategory = "moderate"
	case est.PlannerCost < 10000:
		est.CostCategory = "expensive"
	default:
		est.CostCategory = "very_expensive"
	}

	return est
}

// ---- Helpers ----

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func isReservedWord(s string) bool {
	reserved := map[string]bool{
		"select": true, "from": true, "where": true, "and": true, "or": true,
		"not": true, "in": true, "like": true, "between": true, "is": true,
		"null": true, "true": true, "false": true, "as": true, "on": true,
		"join": true, "left": true, "right": true, "inner": true, "outer": true,
		"group": true, "order": true, "by": true, "having": true, "limit": true,
		"offset": true, "union": true, "all": true, "distinct": true, "case": true,
		"when": true, "then": true, "else": true, "end": true, "exists": true,
	}
	return reserved[strings.ToLower(s)]
}

func guessTableForColumn(col string, tables map[string]bool) string {
	// If only one table, use it
	if len(tables) == 1 {
		for t := range tables {
			return strings.ToLower(t)
		}
	}
	// Otherwise return first table (imprecise but useful for suggestions)
	for t := range tables {
		return strings.ToLower(t)
	}
	return "unknown_table"
}

func deduplicateIndexRecs(recs []models.IndexRecommendation) []models.IndexRecommendation {
	seen := make(map[string]bool)
	var unique []models.IndexRecommendation
	for _, r := range recs {
		key := r.Table + "." + r.Columns
		if !seen[key] {
			seen[key] = true
			unique = append(unique, r)
		}
	}
	return unique
}
