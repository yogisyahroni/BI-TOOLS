package models

// OptimizationSuggestion represents a single optimization suggestion
type OptimizationSuggestion struct {
	Type        string `json:"type"`     // "index", "join", "select", "where", "subquery"
	Severity    string `json:"severity"` // "high", "medium", "low"
	Title       string `json:"title"`
	Description string `json:"description"`
	Original    string `json:"original"`
	Optimized   string `json:"optimized"`
	Impact      string `json:"impact"` // Estimated performance impact
	Example     string `json:"example"`
}

// QueryAnalysisResult represents the result of query analysis
type QueryAnalysisResult struct {
	Query                string                   `json:"query"`
	Suggestions          []OptimizationSuggestion `json:"suggestions"`
	PerformanceScore     int                      `json:"performanceScore"`     // 0-100
	ComplexityLevel      string                   `json:"complexityLevel"`      // "low", "medium", "high"
	EstimatedImprovement string                   `json:"estimatedImprovement"` // e.g., "30-50%"
	CostEstimate         *CostEstimate            `json:"costEstimate,omitempty"`
	ExplainResult        *ExplainResult           `json:"explainResult,omitempty"`
}

// ExplainResult holds parsed EXPLAIN / EXPLAIN ANALYZE output
type ExplainResult struct {
	RawPlan      string                `json:"rawPlan"`
	Nodes        []ExplainNode         `json:"nodes"`
	TotalCost    float64               `json:"totalCost"`    // planner cost units
	ActualTimeMs float64               `json:"actualTimeMs"` // only with ANALYZE
	RowEstimate  int64                 `json:"rowEstimate"`
	ActualRows   int64                 `json:"actualRows"`
	Warnings     []string              `json:"warnings"`
	IndexRecs    []IndexRecommendation `json:"indexRecommendations,omitempty"`
}

// ExplainNode represents a single node in the query plan tree
type ExplainNode struct {
	NodeType     string  `json:"nodeType"` // Seq Scan, Index Scan, Hash Join, etc.
	On           string  `json:"on"`       // table / index name
	Cost         float64 `json:"cost"`     // startup..total
	Rows         int64   `json:"rows"`
	Width        int     `json:"width"`
	ActualTimeMs float64 `json:"actualTimeMs"` // only with ANALYZE
	ActualRows   int64   `json:"actualRows"`
	Filter       string  `json:"filter,omitempty"`
	Warning      string  `json:"warning,omitempty"`
}

// IndexRecommendation suggests a CREATE INDEX statement
type IndexRecommendation struct {
	Table     string `json:"table"`
	Columns   string `json:"columns"`
	Reason    string `json:"reason"`
	CreateSQL string `json:"createSql"`
	Priority  string `json:"priority"` // high, medium, low
}

// CostEstimate provides a cost model for the query
type CostEstimate struct {
	PlannerCost       float64 `json:"plannerCost"`
	EstimatedRows     int64   `json:"estimatedRows"`
	EstimatedWidth    int     `json:"estimatedWidth"`    // bytes per row
	EstimatedDataSize int64   `json:"estimatedDataSize"` // rows * width
	CostCategory      string  `json:"costCategory"`      // "cheap", "moderate", "expensive", "very_expensive"
}
