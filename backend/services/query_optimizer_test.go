package services_test

import (
	"insight-engine-backend/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryOptimizer_AnalyzeQuery(t *testing.T) {
	qo := services.NewQueryOptimizer()

	tests := []struct {
		name           string
		query          string
		expectedScore  int
		expectedIssues []string
	}{
		{
			name:           "Perfect Query",
			query:          "SELECT id, name FROM users WHERE id = 1",
			expectedScore:  100,
			expectedIssues: []string{},
		},
		{
			name:           "SELECT *",
			query:          "SELECT * FROM users WHERE id = 1",
			expectedScore:  90, // Medium severity penalty (10)
			expectedIssues: []string{"SELECT *"},
		},
		{
			name:           "Missing WHERE",
			query:          "SELECT id FROM users",
			expectedScore:  80, // High severity penalty (20)
			expectedIssues: []string{"Missing WHERE clause"},
		},
		{
			name:           "OR in WHERE",
			query:          "SELECT id FROM users WHERE id = 1 OR id = 2",
			expectedScore:  90, // Medium
			expectedIssues: []string{"OR in WHERE clause"},
		},
		{
			name:           "Leading Wildcard",
			query:          "SELECT id FROM users WHERE name LIKE '%john'",
			expectedScore:  90, // Medium
			expectedIssues: []string{"LIKE with leading wildcard"},
		},
		{
			name:           "Multiple Issues",
			query:          "SELECT * FROM users", // SELECT * (10) + Missing WHERE (20)
			expectedScore:  70,
			expectedIssues: []string{"SELECT *", "Missing WHERE clause"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qo.AnalyzeQuery(tt.query)
			assert.Equal(t, tt.expectedScore, result.PerformanceScore)

			// Check suggestions
			var suggestionTitles []string
			for _, s := range result.Suggestions {
				suggestionTitles = append(suggestionTitles, s.Title)
			}

			for _, issue := range tt.expectedIssues {
				assert.Contains(t, suggestionTitles, issue)
			}
		})
	}
}

func TestQueryOptimizer_ParseExplain(t *testing.T) {
	qo := services.NewQueryOptimizer()

	rawPlan := `
Seq Scan on users  (cost=0.00..1.02 rows=2 width=100)
  Filter: (id = 1)
`
	result := qo.ParseExplainOutput(rawPlan)
	assert.NotNil(t, result)
	assert.Equal(t, 1.02, result.TotalCost)
	assert.Equal(t, int64(2), result.RowEstimate)
	assert.Len(t, result.Nodes, 1)
	assert.Equal(t, "Seq Scan", result.Nodes[0].NodeType)
	assert.Equal(t, "users", result.Nodes[0].On)
}
