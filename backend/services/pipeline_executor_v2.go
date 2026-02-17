package services

import (
	"fmt"
	"sort"
	"strings"
)

// ============================================================
// Pipeline Executor V2 Transforms (GAP-010)
// Adds: JOIN, UNION, PIVOT, DEDUPLICATE
// These extend the existing PipelineExecutor.applyTransform
// ============================================================

// TransformJoinConfig configures a join operation
type TransformJoinConfig struct {
	JoinType string `json:"joinType"` // inner, left, right, full
	LeftKey  string `json:"leftKey"`
	RightKey string `json:"rightKey"`
	Prefix   string `json:"prefix,omitempty"` // prefix for right-side columns to avoid collisions
}

// TransformPivotConfig configures a pivot operation
type TransformPivotConfig struct {
	GroupBy     string `json:"groupBy"`     // column to group by
	PivotColumn string `json:"pivotColumn"` // column whose values become new columns
	ValueColumn string `json:"valueColumn"` // column to aggregate
	AggFunc     string `json:"aggFunc"`     // sum, avg, count, min, max, first, last
}

// TransformDeduplicateConfig configures deduplication
type TransformDeduplicateConfig struct {
	Columns    []string `json:"columns"`              // columns to check for duplicates
	KeepFirst  bool     `json:"keepFirst"`            // true = keep first occurrence, false = keep last
	SortColumn string   `json:"sortColumn,omitempty"` // optional sort before dedup
	SortOrder  string   `json:"sortOrder,omitempty"`  // asc or desc
}

// TransformUnionConfig configures a union operation
type TransformUnionConfig struct {
	Distinct bool `json:"distinct"` // true = remove duplicates after union
}

// ---- Join Transform ----

// TransformJoin performs a join between two datasets
func TransformJoin(
	leftData []map[string]interface{},
	rightData []map[string]interface{},
	config TransformJoinConfig,
) ([]map[string]interface{}, error) {
	if config.LeftKey == "" || config.RightKey == "" {
		return nil, fmt.Errorf("join requires leftKey and rightKey")
	}

	joinType := strings.ToLower(config.JoinType)
	if joinType == "" {
		joinType = "inner"
	}

	prefix := config.Prefix
	if prefix == "" {
		prefix = "right_"
	}

	// Build right-side index
	rightIndex := make(map[string][]map[string]interface{})
	for _, rRow := range rightData {
		keyVal := fmt.Sprintf("%v", rRow[config.RightKey])
		rightIndex[keyVal] = append(rightIndex[keyVal], rRow)
	}

	var result []map[string]interface{}
	rightMatched := make(map[string]bool)

	for _, lRow := range leftData {
		leftKeyVal := fmt.Sprintf("%v", lRow[config.LeftKey])
		rRows, found := rightIndex[leftKeyVal]

		if found {
			rightMatched[leftKeyVal] = true
			for _, rRow := range rRows {
				merged := mergeRows(lRow, rRow, prefix, config.RightKey)
				result = append(result, merged)
			}
		} else if joinType == "left" || joinType == "full" {
			// Left/Full: include unmatched left rows with null right columns
			merged := make(map[string]interface{})
			for k, v := range lRow {
				merged[k] = v
			}
			result = append(result, merged)
		}
		// Inner: skip unmatched left rows
	}

	// Right/Full join: include unmatched right rows
	if joinType == "right" || joinType == "full" {
		for _, rRow := range rightData {
			rightKeyVal := fmt.Sprintf("%v", rRow[config.RightKey])
			if !rightMatched[rightKeyVal] {
				merged := make(map[string]interface{})
				for k, v := range rRow {
					if k == config.RightKey {
						merged[k] = v
					} else {
						merged[prefix+k] = v
					}
				}
				result = append(result, merged)
			}
		}
	}

	return result, nil
}

func mergeRows(left, right map[string]interface{}, prefix string, rightKey string) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy left columns
	for k, v := range left {
		merged[k] = v
	}

	// Copy right columns with prefix (skip the join key to avoid duplication)
	for k, v := range right {
		if k == rightKey {
			continue // Already present from left side
		}
		merged[prefix+k] = v
	}

	return merged
}

// ---- Union Transform ----

// TransformUnion combines two datasets
func TransformUnion(
	topData []map[string]interface{},
	bottomData []map[string]interface{},
	config TransformUnionConfig,
) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(topData)+len(bottomData))
	result = append(result, topData...)
	result = append(result, bottomData...)

	if config.Distinct {
		result = deduplicateRows(result, nil)
	}

	return result
}

// ---- Pivot Transform ----

// TransformPivot pivots data from rows to columns
func TransformPivot(
	data []map[string]interface{},
	config TransformPivotConfig,
) ([]map[string]interface{}, error) {
	if config.GroupBy == "" || config.PivotColumn == "" || config.ValueColumn == "" {
		return nil, fmt.Errorf("pivot requires groupBy, pivotColumn, and valueColumn")
	}

	aggFunc := strings.ToLower(config.AggFunc)
	if aggFunc == "" {
		aggFunc = "sum"
	}

	// Phase 1: Collect unique pivot values and group data
	type groupAcc struct {
		values map[string][]float64 // pivot_value -> list of numeric values
	}

	groups := make(map[string]*groupAcc)
	pivotValues := make(map[string]bool)
	groupOrder := make([]string, 0)

	for _, row := range data {
		groupKey := fmt.Sprintf("%v", row[config.GroupBy])
		pivotVal := fmt.Sprintf("%v", row[config.PivotColumn])
		pivotValues[pivotVal] = true

		if _, exists := groups[groupKey]; !exists {
			groups[groupKey] = &groupAcc{
				values: make(map[string][]float64),
			}
			groupOrder = append(groupOrder, groupKey)
		}

		numVal := toFloat64Loose(row[config.ValueColumn])
		groups[groupKey].values[pivotVal] = append(groups[groupKey].values[pivotVal], numVal)
	}

	// Phase 2: Build result rows
	// Sort pivot values for deterministic column order
	sortedPivots := make([]string, 0, len(pivotValues))
	for pv := range pivotValues {
		sortedPivots = append(sortedPivots, pv)
	}
	sort.Strings(sortedPivots)

	var result []map[string]interface{}
	for _, groupKey := range groupOrder {
		acc := groups[groupKey]
		row := map[string]interface{}{
			config.GroupBy: groupKey,
		}

		for _, pivotVal := range sortedPivots {
			vals := acc.values[pivotVal]
			row[pivotVal] = aggregateValues(vals, aggFunc)
		}

		result = append(result, row)
	}

	return result, nil
}

func aggregateValues(vals []float64, aggFunc string) float64 {
	if len(vals) == 0 {
		return 0
	}

	switch aggFunc {
	case "sum":
		s := 0.0
		for _, v := range vals {
			s += v
		}
		return s
	case "avg", "average":
		s := 0.0
		for _, v := range vals {
			s += v
		}
		return s / float64(len(vals))
	case "count":
		return float64(len(vals))
	case "min":
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return m
	case "max":
		m := vals[0]
		for _, v := range vals[1:] {
			if v > m {
				m = v
			}
		}
		return m
	case "first":
		return vals[0]
	case "last":
		return vals[len(vals)-1]
	default:
		// Default to sum
		s := 0.0
		for _, v := range vals {
			s += v
		}
		return s
	}
}

// ---- Deduplicate Transform ----

// TransformDeduplicate removes duplicate rows
func TransformDeduplicate(
	data []map[string]interface{},
	config TransformDeduplicateConfig,
) []map[string]interface{} {
	// Optional pre-sort
	if config.SortColumn != "" {
		sortOrder := strings.ToLower(config.SortOrder)
		sort.SliceStable(data, func(i, j int) bool {
			vi := fmt.Sprintf("%v", data[i][config.SortColumn])
			vj := fmt.Sprintf("%v", data[j][config.SortColumn])
			if sortOrder == "desc" {
				return vi > vj
			}
			return vi < vj
		})
	}

	if len(config.Columns) > 0 {
		return deduplicateRows(data, config.Columns)
	}

	// Deduplicate on all columns
	return deduplicateRows(data, nil)
}

func deduplicateRows(data []map[string]interface{}, columns []string) []map[string]interface{} {
	seen := make(map[string]bool)
	var result []map[string]interface{}

	for _, row := range data {
		key := buildRowKey(row, columns)
		if !seen[key] {
			seen[key] = true
			result = append(result, row)
		}
	}

	return result
}

func buildRowKey(row map[string]interface{}, columns []string) string {
	if len(columns) == 0 {
		// Use all columns, sorted for determinism
		columns = make([]string, 0, len(row))
		for k := range row {
			columns = append(columns, k)
		}
		sort.Strings(columns)
	}

	var sb strings.Builder
	for i, col := range columns {
		if i > 0 {
			sb.WriteByte('|')
		}
		sb.WriteString(col)
		sb.WriteByte('=')
		sb.WriteString(fmt.Sprintf("%v", row[col]))
	}
	return sb.String()
}

// ---- Unpivot Transform (bonus) ----

// TransformUnpivotConfig configures an unpivot (melt) operation
type TransformUnpivotConfig struct {
	IDColumns    []string `json:"idColumns"`    // columns to keep as identifiers
	ValueColumns []string `json:"valueColumns"` // columns to unpivot
	VarName      string   `json:"varName"`      // name for the variable column (default: "variable")
	ValName      string   `json:"valName"`      // name for the value column (default: "value")
}

// TransformUnpivot converts columns into rows (reverse of pivot)
func TransformUnpivot(
	data []map[string]interface{},
	config TransformUnpivotConfig,
) ([]map[string]interface{}, error) {
	if len(config.ValueColumns) == 0 {
		return nil, fmt.Errorf("unpivot requires at least one valueColumn")
	}

	varName := config.VarName
	if varName == "" {
		varName = "variable"
	}
	valName := config.ValName
	if valName == "" {
		valName = "value"
	}

	var result []map[string]interface{}

	for _, row := range data {
		for _, valCol := range config.ValueColumns {
			newRow := make(map[string]interface{})

			// Copy ID columns
			for _, idCol := range config.IDColumns {
				newRow[idCol] = row[idCol]
			}

			// Add variable and value
			newRow[varName] = valCol
			newRow[valName] = row[valCol]

			result = append(result, newRow)
		}
	}

	return result, nil
}

// ---- helpers ----

func toFloat64Loose(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case string:
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			return f
		}
		return 0
	case nil:
		return 0
	default:
		return 0
	}
}
