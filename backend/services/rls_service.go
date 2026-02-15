package services

import (
	"fmt"
	"insight-engine-backend/models"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// RLSService handles Row-Level Security policy enforcement
type RLSService struct {
	db *gorm.DB
}

// NewRLSService creates a new RLS service
func NewRLSService(db *gorm.DB) *RLSService {
	return &RLSService{db: db}
}

// ApplyRLSToQuery modifies a SQL query to enforce RLS policies
func (s *RLSService) ApplyRLSToQuery(query string, userCtx models.UserContext, connectionID string) (string, error) {
	// Extract table names from query
	tableNames := s.extractTableNames(query)
	if len(tableNames) == 0 {
		// No tables found, return original query
		return query, nil
	}

	// Batch get applicable policies for all tables (Fix N+1)
	policiesMap, err := s.GetPoliciesForTables(tableNames, connectionID, userCtx.Roles)
	if err != nil {
		return "", fmt.Errorf("failed to get RLS policies: %w", err)
	}

	var allConditions []string
	for _, tableName := range tableNames {
		policies := policiesMap[tableName]
		if len(policies) == 0 {
			continue
		}

		// Evaluate and combine policies
		tableConditions, err := s.evaluatePolicies(policies, userCtx)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate policies for table '%s': %w", tableName, err)
		}

		if tableConditions != "" {
			allConditions = append(allConditions, fmt.Sprintf("(%s)", tableConditions))
		}
	}

	// No policies to apply
	if len(allConditions) == 0 {
		return query, nil
	}

	// Combine all conditions
	combinedCondition := strings.Join(allConditions, " AND ")

	// Inject RLS conditions into query
	modifiedQuery, err := s.injectWhereClause(query, combinedCondition)
	if err != nil {
		return "", fmt.Errorf("failed to inject RLS conditions: %w", err)
	}

	LogInfo("rls_applied", "Applied RLS policies to query", map[string]interface{}{"tables": tableNames, "conditions": combinedCondition})
	return modifiedQuery, nil
}

// GetPoliciesForTables retrieves applicable RLS policies for multiple tables in minimal queries
func (s *RLSService) GetPoliciesForTables(tableNames []string, connectionID string, userRoles []string) (map[string][]models.RLSPolicy, error) {
	policyMap := make(map[string][]models.RLSPolicy)
	for _, t := range tableNames {
		policyMap[t] = []models.RLSPolicy{}
	}

	// 1. Fetch exact matches for ALL tables in one query
	var exactPolicies []models.RLSPolicy
	if err := s.db.Where("connection_id = ? AND enabled = ? AND table_name IN ?", connectionID, true, tableNames).
		Order("priority DESC").Find(&exactPolicies).Error; err != nil {
		return nil, err
	}

	// 2. Fetch wildcard policies (once for the connection)
	// Optimization: checking if ANY table needs wildcard might be premature optimization, just fetch them.
	// We assume wildcard policies are relatively few.
	var wildcardPolicies []models.RLSPolicy
	if err := s.db.Where("connection_id = ? AND enabled = ? AND table_name LIKE ?", connectionID, true, "%*%").
		Order("priority DESC").Find(&wildcardPolicies).Error; err != nil {
		return nil, err
	}

	// 3. Distribute policies to the map
	// Add exact matches
	for _, p := range exactPolicies {
		policyMap[p.Table] = append(policyMap[p.Table], p)
	}

	// Add wildcard matches to relevant tables
	for _, wp := range wildcardPolicies {
		pattern := strings.ReplaceAll(wp.Table, "*", ".*")
		// Anchor the regex
		regexPattern := fmt.Sprintf("^%s$", pattern)
		re, err := regexp.Compile(regexPattern)
		if err != nil {
			LogWarn("rls_invalid_regex", "Invalid wildcard policy pattern ignored", map[string]interface{}{"pattern": wp.Table})
			continue
		}

		for _, t := range tableNames {
			if re.MatchString(t) {
				policyMap[t] = append(policyMap[t], wp)
			}
		}
	}

	// 4. Filter by Roles for each table list
	finalMap := make(map[string][]models.RLSPolicy)
	for t, policies := range policyMap {
		finalMap[t] = s.filterPoliciesByRole(policies, userRoles)
	}

	return finalMap, nil
}

// GetPoliciesForTable retrieves applicable RLS policies for a single table (Wrapper for GetPoliciesForTables)
func (s *RLSService) GetPoliciesForTable(tableName, connectionID string, userRoles []string) ([]models.RLSPolicy, error) {
	policyMap, err := s.GetPoliciesForTables([]string{tableName}, connectionID, userRoles)
	if err != nil {
		return nil, err
	}
	return policyMap[tableName], nil
}

// filterPoliciesByRole filters a list of policies based on user roles
func (s *RLSService) filterPoliciesByRole(policies []models.RLSPolicy, userRoles []string) []models.RLSPolicy {
	var filteredPolicies []models.RLSPolicy
	for _, policy := range policies {
		if policy.RoleIDs == nil || len(*policy.RoleIDs) == 0 {
			// Policy applies to all users
			filteredPolicies = append(filteredPolicies, policy)
			continue
		}

		// Check if user has any matching role
		hasRole := false
		for _, userRole := range userRoles {
			for _, policyRole := range *policy.RoleIDs {
				if userRole == policyRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		if hasRole {
			filteredPolicies = append(filteredPolicies, policy)
		}
	}
	return filteredPolicies
}

// evaluatePolicies combines multiple policies into a single WHERE clause
func (s *RLSService) evaluatePolicies(policies []models.RLSPolicy, userCtx models.UserContext) (string, error) {
	if len(policies) == 0 {
		return "", nil
	}

	var conditions []string
	for _, policy := range policies {
		condition, err := s.evaluateCondition(policy.Condition, userCtx)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate policy '%s': %w", policy.Name, err)
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", condition))
	}

	// Combine based on mode (first policy's mode is used if mixed)
	mode := policies[0].Mode
	if mode == "" {
		mode = "AND"
	}

	separator := " AND "
	if mode == "OR" {
		separator = " OR "
	}

	return strings.Join(conditions, separator), nil
}

// evaluateCondition replaces template variables in a condition string
func (s *RLSService) evaluateCondition(condition string, userCtx models.UserContext) (string, error) {
	result := condition

	// Define allowed template variables
	templates := map[string]string{
		"{{current_user.id}}":    userCtx.UserID,
		"{{current_user.email}}": userCtx.Email,
	}

	// Add roles as comma-separated list
	if len(userCtx.Roles) > 0 {
		rolesStr := "'" + strings.Join(userCtx.Roles, "','") + "'"
		templates["{{current_user.roles}}"] = rolesStr
	}

	// Add team IDs as comma-separated list
	if len(userCtx.TeamIDs) > 0 {
		teamsStr := "'" + strings.Join(userCtx.TeamIDs, "','") + "'"
		templates["{{current_user.team_ids}}"] = teamsStr
	}

	// Add custom attributes
	for key, value := range userCtx.Attributes {
		templateKey := fmt.Sprintf("{{current_user.attributes.%s}}", key)
		templates[templateKey] = fmt.Sprintf("%v", value)
	}

	// Replace all templates
	for template, value := range templates {
		// ESCAPE THE VALUE TO PREVENT SQL INJECTION
		escapedValue := s.escapeSQLString(value)
		result = strings.ReplaceAll(result, template, escapedValue)
	}

	// Security check: Ensure no unreplaced templates
	if strings.Contains(result, "{{") {
		return "", fmt.Errorf("condition contains unreplaced template variables: %s", result)
	}

	return result, nil
}

// escapeSQLString escapes a string to be safely used in a SQL query.
// It escapes single quotes by doubling them (standard SQL escaping).
func (s *RLSService) escapeSQLString(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

// extractTableNames extracts table names from a SQL query
func (s *RLSService) extractTableNames(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))

	var tables []string

	// Simple regex patterns for common SQL patterns
	patterns := []string{
		`from\s+([a-z_][a-z0-9_]*)\s*`,          // FROM table
		`join\s+([a-z_][a-z0-9_]*)\s*`,          // JOIN table
		`update\s+([a-z_][a-z0-9_]*)\s*`,        // UPDATE table
		`insert\s+into\s+([a-z_][a-z0-9_]*)\s*`, // INSERT INTO table
		`delete\s+from\s+([a-z_][a-z0-9_]*)\s*`, // DELETE FROM table
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(query, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tableName := match[1]
				// Deduplicate
				if !sliceContainsString(tables, tableName) {
					tables = append(tables, tableName)
				}
			}
		}
	}

	return tables
}

// injectWhereClause adds RLS conditions to a SQL query
func (s *RLSService) injectWhereClause(query, condition string) (string, error) {
	if condition == "" {
		return query, nil
	}

	query = strings.TrimSpace(query)
	lowerQuery := strings.ToLower(query)

	// Check if query already has WHERE clause
	whereIdx := strings.Index(lowerQuery, " where ")

	if whereIdx != -1 {
		// Query has WHERE - inject condition with AND
		// Find the position after WHERE clause
		beforeWhere := query[:whereIdx+7] // +7 for " WHERE "
		afterWhere := query[whereIdx+7:]

		return fmt.Sprintf("%s(%s) AND (%s)", beforeWhere, condition, afterWhere), nil
	}

	// Query doesn't have WHERE - add it
	// Find position before ORDER BY, GROUP BY, HAVING, LIMIT, etc.
	endClauses := []string{" order by ", " group by ", " having ", " limit ", " offset "}
	insertPos := len(query)

	for _, clause := range endClauses {
		idx := strings.Index(lowerQuery, clause)
		if idx != -1 && idx < insertPos {
			insertPos = idx
		}
	}

	before := strings.TrimSpace(query[:insertPos])
	after := ""
	if insertPos < len(query) {
		after = " " + strings.TrimSpace(query[insertPos:])
	}

	return fmt.Sprintf("%s WHERE (%s)%s", before, condition, after), nil
}

// CreatePolicy creates a new RLS policy
func (s *RLSService) CreatePolicy(policy *models.RLSPolicy) error {
	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	return s.db.Create(policy).Error
}

// UpdatePolicy updates an existing RLS policy
func (s *RLSService) UpdatePolicy(policy *models.RLSPolicy) error {
	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	return s.db.Save(policy).Error
}

// DeletePolicy deletes an RLS policy
func (s *RLSService) DeletePolicy(policyID string) error {
	return s.db.Delete(&models.RLSPolicy{}, "id = ?", policyID).Error
}

// GetPolicy retrieves a single policy by ID
func (s *RLSService) GetPolicy(policyID string) (*models.RLSPolicy, error) {
	var policy models.RLSPolicy
	err := s.db.First(&policy, "id = ?", policyID).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies retrieves all policies for a user
func (s *RLSService) ListPolicies(userID string) ([]models.RLSPolicy, error) {
	var policies []models.RLSPolicy
	err := s.db.Where("user_id = ?", userID).Order("priority DESC").Find(&policies).Error
	return policies, err
}

// validatePolicy validates policy fields
func (s *RLSService) validatePolicy(policy *models.RLSPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}
	if policy.ConnectionID == "" {
		return fmt.Errorf("connection ID is required")
	}
	if policy.Table == "" {
		return fmt.Errorf("table name is required")
	}
	if policy.Condition == "" {
		return fmt.Errorf("condition is required")
	}
	if policy.Mode != "AND" && policy.Mode != "OR" {
		return fmt.Errorf("mode must be 'AND' or 'OR'")
	}

	// Validate template syntax
	if !strings.Contains(policy.Condition, "{{") {
		LogWarn("rls_no_template_vars", "RLS policy has no template variables, will apply same filter to all users", map[string]interface{}{"policy_name": policy.Name})
	}

	return nil
}

// TestPolicy tests a policy against sample data (for UI preview)
func (s *RLSService) TestPolicy(policyID string, userCtx models.UserContext, sampleQuery string) (string, error) {
	policy, err := s.GetPolicy(policyID)
	if err != nil {
		return "", err
	}

	// Evaluate the policy condition
	evaluatedCondition, err := s.evaluateCondition(policy.Condition, userCtx)
	if err != nil {
		return "", err
	}

	// Apply to sample query
	modifiedQuery, err := s.injectWhereClause(sampleQuery, evaluatedCondition)
	if err != nil {
		return "", err
	}

	return modifiedQuery, nil
}

// Helper function
func sliceContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
