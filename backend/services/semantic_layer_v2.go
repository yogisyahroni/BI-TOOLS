package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ============================================================
// Semantic Layer V2 (GAP-007)
// Adds: Hierarchies, KPIs, Perspectives, Time Intelligence
// ============================================================

// ---- V2 Models ----

// SemanticHierarchy defines a drill-down path across dimensions
// e.g., Country → Region → City → Store
type SemanticHierarchy struct {
	ID          string                   `gorm:"primaryKey" json:"id"`
	ModelID     string                   `gorm:"not null;index" json:"modelId"`
	Name        string                   `gorm:"not null" json:"name"`
	Description string                   `json:"description"`
	Levels      []SemanticHierarchyLevel `gorm:"foreignKey:HierarchyID;constraint:OnDelete:CASCADE" json:"levels"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

func (SemanticHierarchy) TableName() string { return "semantic_hierarchies" }

// SemanticHierarchyLevel is one step in a hierarchy
type SemanticHierarchyLevel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	HierarchyID string    `gorm:"not null;index" json:"hierarchyId"`
	DimensionID string    `gorm:"not null" json:"dimensionId"`
	LevelOrder  int       `gorm:"not null" json:"levelOrder"` // 0 = top, 1, 2, ...
	LabelColumn string    `json:"labelColumn,omitempty"`      // optional display column
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (SemanticHierarchyLevel) TableName() string { return "semantic_hierarchy_levels" }

// SemanticKPI is a goal-oriented metric with targets, thresholds, and trend
type SemanticKPI struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	ModelID           string    `gorm:"not null;index" json:"modelId"`
	Name              string    `gorm:"not null" json:"name"`
	Description       string    `json:"description"`
	MetricID          string    `gorm:"not null" json:"metricId"` // FK to SemanticMetric
	TargetValue       *float64  `json:"targetValue,omitempty"`
	WarningThreshold  *float64  `json:"warningThreshold,omitempty"`                  // amber zone
	CriticalThreshold *float64  `json:"criticalThreshold,omitempty"`                 // red zone
	Direction         string    `gorm:"default:'higher_is_better'" json:"direction"` // higher_is_better | lower_is_better
	TrendPeriod       string    `json:"trendPeriod,omitempty"`                       // day, week, month, quarter, year
	Unit              string    `json:"unit,omitempty"`                              // $, %, units
	Owner             string    `json:"owner,omitempty"`                             // team/person responsible
	Tags              string    `json:"tags,omitempty"`                              // comma-separated tags
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (SemanticKPI) TableName() string { return "semantic_kpis" }

// SemanticPerspective is a curated view (subset of dimensions + metrics)
// Think of it as a "saved lens" for a particular audience
type SemanticPerspective struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	ModelID      string    `gorm:"not null;index" json:"modelId"`
	Name         string    `gorm:"not null" json:"name"`
	Description  string    `json:"description"`
	DimensionIDs string    `json:"dimensionIds"` // JSON array: ["dim1","dim2"]
	MetricIDs    string    `json:"metricIds"`    // JSON array
	FilterJSON   string    `json:"filterJson"`   // default filters as JSON
	SortColumn   string    `json:"sortColumn,omitempty"`
	SortOrder    string    `gorm:"default:'asc'" json:"sortOrder"` // asc | desc
	DefaultLimit int       `gorm:"default:100" json:"defaultLimit"`
	IsPublic     bool      `gorm:"default:false" json:"isPublic"`
	CreatedBy    string    `gorm:"not null" json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (SemanticPerspective) TableName() string { return "semantic_perspectives" }

// ---- V2 DTOs ----

// KPIStatus represents the current status of a KPI
type KPIStatus struct {
	KPI          *SemanticKPI `json:"kpi"`
	CurrentValue float64      `json:"currentValue"`
	Status       string       `json:"status"` // on_track, warning, critical, no_target
	PctOfTarget  float64      `json:"pctOfTarget"`
	Trend        string       `json:"trend"` // up, down, flat
	TrendPct     float64      `json:"trendPct"`
}

// PerspectiveQuery represents a resolved perspective ready for SQL generation
type PerspectiveQuery struct {
	Dimensions []string               `json:"dimensions"`
	Metrics    []string               `json:"metrics"`
	Filters    map[string]interface{} `json:"filters"`
	SortColumn string                 `json:"sortColumn"`
	SortOrder  string                 `json:"sortOrder"`
	Limit      int                    `json:"limit"`
}

// DrillPath represents the resolved drill-down context
type DrillPath struct {
	HierarchyName string   `json:"hierarchyName"`
	CurrentLevel  int      `json:"currentLevel"`
	TotalLevels   int      `json:"totalLevels"`
	Breadcrumbs   []string `json:"breadcrumbs"` // ["USA","California","Los Angeles"]
	NextDimension string   `json:"nextDimension,omitempty"`
	PrevDimension string   `json:"prevDimension,omitempty"`
}

// TimeGrain defines a time intelligence grain
type TimeGrain string

const (
	TimeGrainDay     TimeGrain = "day"
	TimeGrainWeek    TimeGrain = "week"
	TimeGrainMonth   TimeGrain = "month"
	TimeGrainQuarter TimeGrain = "quarter"
	TimeGrainYear    TimeGrain = "year"
)

// ---- Service ----

// SemanticLayerV2Service extends the semantic layer with hierarchies, KPIs, perspectives
type SemanticLayerV2Service struct {
	db *gorm.DB
}

// NewSemanticLayerV2Service creates a new Semantic Layer V2 service
func NewSemanticLayerV2Service(db *gorm.DB) *SemanticLayerV2Service {
	return &SemanticLayerV2Service{db: db}
}

// ---- Hierarchy Operations ----

// CreateHierarchy defines a new drill-down hierarchy
func (s *SemanticLayerV2Service) CreateHierarchy(h *SemanticHierarchy) error {
	if len(h.Levels) < 2 {
		return fmt.Errorf("hierarchy must have at least 2 levels")
	}

	// Validate level ordering
	for i, level := range h.Levels {
		if level.LevelOrder != i {
			return fmt.Errorf("level order must be sequential starting from 0, got %d at index %d", level.LevelOrder, i)
		}
	}

	return s.db.Create(h).Error
}

// GetHierarchy retrieves a hierarchy with its levels
func (s *SemanticLayerV2Service) GetHierarchy(id string) (*SemanticHierarchy, error) {
	var h SemanticHierarchy
	err := s.db.Preload("Levels", func(db *gorm.DB) *gorm.DB {
		return db.Order("level_order ASC")
	}).First(&h, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// ListHierarchies lists all hierarchies for a model
func (s *SemanticLayerV2Service) ListHierarchies(modelID string) ([]SemanticHierarchy, error) {
	var hierarchies []SemanticHierarchy
	err := s.db.Preload("Levels", func(db *gorm.DB) *gorm.DB {
		return db.Order("level_order ASC")
	}).Where("model_id = ?", modelID).Find(&hierarchies).Error
	return hierarchies, err
}

// ResolveDrillPath resolves the current position in a hierarchy drill-down
func (s *SemanticLayerV2Service) ResolveDrillPath(hierarchyID string, currentLevel int, breadcrumbs []string) (*DrillPath, error) {
	h, err := s.GetHierarchy(hierarchyID)
	if err != nil {
		return nil, fmt.Errorf("hierarchy not found: %w", err)
	}

	totalLevels := len(h.Levels)
	if currentLevel < 0 || currentLevel >= totalLevels {
		return nil, fmt.Errorf("level %d out of range (0-%d)", currentLevel, totalLevels-1)
	}

	path := &DrillPath{
		HierarchyName: h.Name,
		CurrentLevel:  currentLevel,
		TotalLevels:   totalLevels,
		Breadcrumbs:   breadcrumbs,
	}

	// Resolve dimension names for next/prev
	if currentLevel+1 < totalLevels {
		nextDimID := h.Levels[currentLevel+1].DimensionID
		path.NextDimension = nextDimID
	}
	if currentLevel > 0 {
		prevDimID := h.Levels[currentLevel-1].DimensionID
		path.PrevDimension = prevDimID
	}

	return path, nil
}

// DeleteHierarchy deletes a hierarchy
func (s *SemanticLayerV2Service) DeleteHierarchy(id string) error {
	return s.db.Delete(&SemanticHierarchy{}, "id = ?", id).Error
}

// ---- KPI Operations ----

// CreateKPI defines a new KPI
func (s *SemanticLayerV2Service) CreateKPI(kpi *SemanticKPI) error {
	if kpi.MetricID == "" {
		return fmt.Errorf("KPI must reference a metric")
	}
	return s.db.Create(kpi).Error
}

// GetKPI retrieves a KPI by ID
func (s *SemanticLayerV2Service) GetKPI(id string) (*SemanticKPI, error) {
	var kpi SemanticKPI
	err := s.db.First(&kpi, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &kpi, nil
}

// ListKPIs lists all KPIs for a model
func (s *SemanticLayerV2Service) ListKPIs(modelID string) ([]SemanticKPI, error) {
	var kpis []SemanticKPI
	err := s.db.Where("model_id = ?", modelID).Find(&kpis).Error
	return kpis, err
}

// EvaluateKPIStatus evaluates a KPI's current status given the current metric value
func (s *SemanticLayerV2Service) EvaluateKPIStatus(kpi *SemanticKPI, currentValue float64, previousValue *float64) *KPIStatus {
	status := &KPIStatus{
		KPI:          kpi,
		CurrentValue: currentValue,
		Status:       "no_target",
	}

	// Calculate trend
	if previousValue != nil {
		diff := currentValue - *previousValue
		if *previousValue != 0 {
			status.TrendPct = (diff / *previousValue) * 100
		}
		switch {
		case diff > 0:
			status.Trend = "up"
		case diff < 0:
			status.Trend = "down"
		default:
			status.Trend = "flat"
		}
	} else {
		status.Trend = "flat"
	}

	// Evaluate against target
	if kpi.TargetValue != nil && *kpi.TargetValue != 0 {
		status.PctOfTarget = (currentValue / *kpi.TargetValue) * 100

		higherIsBetter := kpi.Direction != "lower_is_better"

		if higherIsBetter {
			switch {
			case kpi.CriticalThreshold != nil && currentValue < *kpi.CriticalThreshold:
				status.Status = "critical"
			case kpi.WarningThreshold != nil && currentValue < *kpi.WarningThreshold:
				status.Status = "warning"
			default:
				status.Status = "on_track"
			}
		} else {
			// Lower is better (e.g., error rate, churn)
			switch {
			case kpi.CriticalThreshold != nil && currentValue > *kpi.CriticalThreshold:
				status.Status = "critical"
			case kpi.WarningThreshold != nil && currentValue > *kpi.WarningThreshold:
				status.Status = "warning"
			default:
				status.Status = "on_track"
			}
		}
	}

	return status
}

// UpdateKPI updates an existing KPI
func (s *SemanticLayerV2Service) UpdateKPI(kpi *SemanticKPI) error {
	return s.db.Save(kpi).Error
}

// DeleteKPI deletes a KPI
func (s *SemanticLayerV2Service) DeleteKPI(id string) error {
	return s.db.Delete(&SemanticKPI{}, "id = ?", id).Error
}

// ---- Perspective Operations ----

// CreatePerspective creates a curated view
func (s *SemanticLayerV2Service) CreatePerspective(p *SemanticPerspective) error {
	// Validate JSON arrays
	if p.DimensionIDs != "" {
		var dims []string
		if err := json.Unmarshal([]byte(p.DimensionIDs), &dims); err != nil {
			return fmt.Errorf("invalid dimensionIds JSON: %w", err)
		}
	}
	if p.MetricIDs != "" {
		var mets []string
		if err := json.Unmarshal([]byte(p.MetricIDs), &mets); err != nil {
			return fmt.Errorf("invalid metricIds JSON: %w", err)
		}
	}
	return s.db.Create(p).Error
}

// GetPerspective retrieves a perspective by ID
func (s *SemanticLayerV2Service) GetPerspective(id string) (*SemanticPerspective, error) {
	var p SemanticPerspective
	err := s.db.First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPerspectives lists all perspectives for a model
func (s *SemanticLayerV2Service) ListPerspectives(modelID string, includePrivate bool, userID string) ([]SemanticPerspective, error) {
	var perspectives []SemanticPerspective
	query := s.db.Where("model_id = ?", modelID)

	if !includePrivate {
		query = query.Where("is_public = ? OR created_by = ?", true, userID)
	}

	err := query.Find(&perspectives).Error
	return perspectives, err
}

// ResolvePerspective resolves a perspective into a query specification
func (s *SemanticLayerV2Service) ResolvePerspective(perspectiveID string) (*PerspectiveQuery, error) {
	p, err := s.GetPerspective(perspectiveID)
	if err != nil {
		return nil, err
	}

	result := &PerspectiveQuery{
		SortColumn: p.SortColumn,
		SortOrder:  p.SortOrder,
		Limit:      p.DefaultLimit,
	}

	// Parse dimension IDs
	if p.DimensionIDs != "" {
		if err := json.Unmarshal([]byte(p.DimensionIDs), &result.Dimensions); err != nil {
			return nil, fmt.Errorf("invalid dimensionIds: %w", err)
		}
	}

	// Parse metric IDs
	if p.MetricIDs != "" {
		if err := json.Unmarshal([]byte(p.MetricIDs), &result.Metrics); err != nil {
			return nil, fmt.Errorf("invalid metricIds: %w", err)
		}
	}

	// Parse default filters
	if p.FilterJSON != "" {
		result.Filters = make(map[string]interface{})
		if err := json.Unmarshal([]byte(p.FilterJSON), &result.Filters); err != nil {
			return nil, fmt.Errorf("invalid filterJson: %w", err)
		}
	}

	return result, nil
}

// UpdatePerspective updates an existing perspective
func (s *SemanticLayerV2Service) UpdatePerspective(p *SemanticPerspective) error {
	return s.db.Save(p).Error
}

// DeletePerspective deletes a perspective
func (s *SemanticLayerV2Service) DeletePerspective(id string) error {
	return s.db.Delete(&SemanticPerspective{}, "id = ?", id).Error
}

// ---- Time Intelligence ----

// BuildTimeFilter generates a SQL WHERE clause for time-based filtering
func (s *SemanticLayerV2Service) BuildTimeFilter(columnName string, grain TimeGrain, periodsBack int, dialect string) string {
	if dialect == "" {
		dialect = "postgres"
	}

	switch dialect {
	case "postgres":
		return s.buildPostgresTimeFilter(columnName, grain, periodsBack)
	case "mysql":
		return s.buildMySQLTimeFilter(columnName, grain, periodsBack)
	default:
		return s.buildPostgresTimeFilter(columnName, grain, periodsBack)
	}
}

func (s *SemanticLayerV2Service) buildPostgresTimeFilter(col string, grain TimeGrain, periods int) string {
	var interval string
	switch grain {
	case TimeGrainDay:
		interval = fmt.Sprintf("%d days", periods)
	case TimeGrainWeek:
		interval = fmt.Sprintf("%d weeks", periods)
	case TimeGrainMonth:
		interval = fmt.Sprintf("%d months", periods)
	case TimeGrainQuarter:
		interval = fmt.Sprintf("%d months", periods*3)
	case TimeGrainYear:
		interval = fmt.Sprintf("%d years", periods)
	default:
		interval = fmt.Sprintf("%d days", periods)
	}
	return fmt.Sprintf("%s >= NOW() - INTERVAL '%s'", col, interval)
}

func (s *SemanticLayerV2Service) buildMySQLTimeFilter(col string, grain TimeGrain, periods int) string {
	var unit string
	count := periods
	switch grain {
	case TimeGrainDay:
		unit = "DAY"
	case TimeGrainWeek:
		unit = "WEEK"
	case TimeGrainMonth:
		unit = "MONTH"
	case TimeGrainQuarter:
		unit = "MONTH"
		count = periods * 3
	case TimeGrainYear:
		unit = "YEAR"
	default:
		unit = "DAY"
	}
	return fmt.Sprintf("%s >= DATE_SUB(NOW(), INTERVAL %d %s)", col, count, unit)
}

// BuildTimeGroupBy generates a date_trunc expression for grouping
func (s *SemanticLayerV2Service) BuildTimeGroupBy(columnName string, grain TimeGrain, dialect string) string {
	if dialect == "" || dialect == "postgres" {
		return fmt.Sprintf("DATE_TRUNC('%s', %s)", string(grain), columnName)
	}
	// MySQL
	switch grain {
	case TimeGrainDay:
		return fmt.Sprintf("DATE(%s)", columnName)
	case TimeGrainWeek:
		return fmt.Sprintf("DATE(DATE_SUB(%s, INTERVAL WEEKDAY(%s) DAY))", columnName, columnName)
	case TimeGrainMonth:
		return fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m-01')", columnName)
	case TimeGrainQuarter:
		return fmt.Sprintf("CONCAT(YEAR(%s), '-Q', QUARTER(%s))", columnName, columnName)
	case TimeGrainYear:
		return fmt.Sprintf("YEAR(%s)", columnName)
	default:
		return fmt.Sprintf("DATE(%s)", columnName)
	}
}

// ---- Comprehensive Query Builder ----

// SemanticQueryV2 represents an enhanced semantic query with v2 features
type SemanticQueryV2 struct {
	ModelID        string                 `json:"modelId"`
	Dimensions     []string               `json:"dimensions"`
	Metrics        []string               `json:"metrics"`
	Filters        map[string]interface{} `json:"filters"`
	TimeColumn     string                 `json:"timeColumn,omitempty"`
	TimeGrain      TimeGrain              `json:"timeGrain,omitempty"`
	TimePeriods    int                    `json:"timePeriods,omitempty"` // lookback periods
	DrillHierarchy string                 `json:"drillHierarchy,omitempty"`
	DrillLevel     int                    `json:"drillLevel,omitempty"`
	SortColumn     string                 `json:"sortColumn,omitempty"`
	SortOrder      string                 `json:"sortOrder,omitempty"`
	Limit          int                    `json:"limit,omitempty"`
}

// TranslateV2 translates an enhanced semantic query to SQL
func (s *SemanticLayerV2Service) TranslateV2(query *SemanticQueryV2, model *SemanticModelLite, dialect string) (string, []interface{}, error) {
	if model == nil {
		return "", nil, fmt.Errorf("model is required")
	}

	var selectParts []string
	var groupByParts []string
	var whereParts []string
	var args []interface{}

	// Add time grouping if specified
	if query.TimeColumn != "" && query.TimeGrain != "" {
		timeExpr := s.BuildTimeGroupBy(query.TimeColumn, query.TimeGrain, dialect)
		selectParts = append(selectParts, fmt.Sprintf("%s AS time_period", timeExpr))
		groupByParts = append(groupByParts, timeExpr)
	}

	// Add dimensions
	for _, dimName := range query.Dimensions {
		col, ok := model.DimMap[dimName]
		if !ok {
			return "", nil, fmt.Errorf("dimension not found: %s", dimName)
		}
		selectParts = append(selectParts, fmt.Sprintf("%s AS \"%s\"", col, dimName))
		groupByParts = append(groupByParts, col)
	}

	// Add metrics
	for _, metricName := range query.Metrics {
		formula, ok := model.MetricMap[metricName]
		if !ok {
			return "", nil, fmt.Errorf("metric not found: %s", metricName)
		}
		selectParts = append(selectParts, fmt.Sprintf("%s AS \"%s\"", formula, metricName))
	}

	if len(selectParts) == 0 {
		return "", nil, fmt.Errorf("no dimensions or metrics specified")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectParts, ", "), model.TableName)

	// Add time filter
	if query.TimeColumn != "" && query.TimePeriods > 0 {
		whereParts = append(whereParts, s.BuildTimeFilter(query.TimeColumn, query.TimeGrain, query.TimePeriods, dialect))
	}

	// Add regular filters
	for dimName, value := range query.Filters {
		col, ok := model.DimMap[dimName]
		if !ok {
			return "", nil, fmt.Errorf("filter dimension not found: %s", dimName)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", col))
		args = append(args, value)
	}

	if len(whereParts) > 0 {
		sql += " WHERE " + strings.Join(whereParts, " AND ")
	}

	if len(groupByParts) > 0 {
		sql += " GROUP BY " + strings.Join(groupByParts, ", ")
	}

	// Sort
	if query.SortColumn != "" {
		order := "ASC"
		if strings.EqualFold(query.SortOrder, "desc") {
			order = "DESC"
		}
		sql += fmt.Sprintf(" ORDER BY \"%s\" %s", query.SortColumn, order)
	} else if query.TimeColumn != "" {
		sql += " ORDER BY time_period ASC"
	}

	// Limit
	limit := query.Limit
	if limit <= 0 {
		limit = 1000
	}
	sql += fmt.Sprintf(" LIMIT %d", limit)

	return sql, args, nil
}

// SemanticModelLite is a lightweight model representation
// for query translation (avoids full GORM loading)
type SemanticModelLite struct {
	TableName string
	DimMap    map[string]string // dim name → column
	MetricMap map[string]string // metric name → formula
}

// ---- Migration Helper ----

// AutoMigrateV2 runs auto-migration for v2 models
func (s *SemanticLayerV2Service) AutoMigrateV2() error {
	return s.db.AutoMigrate(
		&SemanticHierarchy{},
		&SemanticHierarchyLevel{},
		&SemanticKPI{},
		&SemanticPerspective{},
	)
}
