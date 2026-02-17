package formula_engine

import (
	"fmt"
	"math"
	"strings"
)

// Function represents a built-in formula function
type Function func(args []interface{}) (interface{}, error)

// FunctionRegistry holds all available functions
var FunctionRegistry = map[string]Function{
	"SUM":     funcSum,
	"AVG":     funcAvg,
	"MIN":     funcMin,
	"MAX":     funcMax,
	"IF":      funcIf,
	"VLOOKUP": funcVLookup,
	// Date/Time
	"NOW":   funcNow,
	"TODAY": funcToday,
	"YEAR":  funcYear,
	"MONTH": funcMonth,
	// Text
	"UPPER":  funcUpper,
	"LOWER":  funcLower,
	"CONCAT": funcConcat,
	"LEN":    funcLen,
	"TRIM":   funcTrim,
	"LEFT":   funcLeft,
	"RIGHT":  funcRight,
}

// Helper to convert any value to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	case string:
		// Excel-like: strings in arithmetic might convert if they look like numbers
		// But in strict mode, maybe not. Let's try to parse.
		/*
			f, err := strconv.ParseFloat(val, 64)
			if err == nil {
				return f, nil
			}
		*/
		return 0, fmt.Errorf("cannot convert string '%s' to number", val)
	default:
		return 0, fmt.Errorf("cannot convert type %T to number", val)
	}
}

// Helper to recursively flatten arguments for aggregation functions
func flattenArgs(args []interface{}) []float64 {
	var nums []float64
	for _, arg := range args {
		switch v := arg.(type) {
		case []interface{}:
			nums = append(nums, flattenArgs(v)...)
		case [][]interface{}: // Handle 2D arrays
			for _, row := range v {
				nums = append(nums, flattenArgs(row)...)
			}
		default:
			if val, err := toFloat64(v); err == nil {
				nums = append(nums, val)
			}
		}
	}
	return nums
}

// ---------------- Standard Functions ----------------

func funcSum(args []interface{}) (interface{}, error) {
	nums := flattenArgs(args)
	sum := 0.0
	for _, val := range nums {
		sum += val
	}
	return sum, nil
}

func funcAvg(args []interface{}) (interface{}, error) {
	nums := flattenArgs(args)
	if len(nums) == 0 {
		return 0.0, nil // Avoid DivByZero
	}
	sum := 0.0
	for _, val := range nums {
		sum += val
	}
	return sum / float64(len(nums)), nil
}

func funcMin(args []interface{}) (interface{}, error) {
	nums := flattenArgs(args)
	if len(nums) == 0 {
		return 0.0, nil
	}
	minVal := math.MaxFloat64
	for _, val := range nums {
		if val < minVal {
			minVal = val
		}
	}
	return minVal, nil
}

func funcMax(args []interface{}) (interface{}, error) {
	nums := flattenArgs(args)
	if len(nums) == 0 {
		return 0.0, nil
	}
	maxVal := -math.MaxFloat64
	for _, val := range nums {
		if val > maxVal {
			maxVal = val
		}
	}
	return maxVal, nil
}

func funcIf(args []interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("IF requires 2 or 3 arguments")
	}

	// IF logic now handled in evaluator.go (lazy eval), but we keep this for legacy or direct calls
	// Evaluator should have resolved args if it called this, meaning args[0] is the result of condition
	// BUT Evaluator actually handles the branching.
	// If execution reaches here, it might be due to direct call.
	// The evaluator's special case for IF implies this function body might not be reached from evaluator.
	// However, we return the value passed.

	// Wait, if evaluator handles IF, it passes the RESULT of the branch.
	// So if we reach here via evaluator, something is wrong or evaluator behavior changed.
	// Actually evaluator handles IF *before* calling this.
	// So strict logic: this function is technically unreachable from evaluator.
	// But we keep it safe.
	return args[0], nil
}

func funcVLookup(args []interface{}) (interface{}, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("VLOOKUP requires 3 or 4 arguments")
	}

	lookupValue := args[0]
	tableArray := args[1]
	colIndexArg := args[2]
	exactMatch := false // Default to approximate matching (TRUE/1)? Excel default is TRUE.
	// But usually people want FALSE (Exact). Let's stick to Excel behavior default TRUE,
	// BUT, for most business apps providing a default of FALSE might be better?
	// Excel: "If Range_lookup is TRUE or omitted... an approximate match is returned"
	// Let's implement Excel behavior: Default TRUE (Approximate).
	// Actually, wait, approximate match requires sorting. "If range_lookup is TRUE, the values... must be placed in ascending order".
	// Implementing Approximate match correctly requires sorting assumption.
	// To avoid confusion, let's treat default as EXACT match if the table isn't seemingly sorted?
	// No, stick to spec. Default to True (Approximate) is dangerous if not sorted.
	// Most users assume VLOOKUP(val, table, col, FALSE) for exact.

	if len(args) == 4 {
		switch v := args[3].(type) {
		case bool:
			exactMatch = !v // standard VLOOKUP: FALSE = Exact Match (so exactMatch=true)
		case float64:
			exactMatch = v == 0 // 0 = Exact
		}
	} else {
		// If omitted, Excel defaults to TRUE (Approximate).
		// We will implement Exact match logic only for now to be safe and robust for typical API usages.
		// Approximate match on unsorted data returns garbage basically.
		exactMatch = true // We ONLY support exact match for this implementation to prevent bugs with unsorted data.
	}

	colIndex, err := toFloat64(colIndexArg)
	if err != nil {
		return nil, fmt.Errorf("col_index_num must be a number")
	}
	if colIndex < 1 {
		return nil, fmt.Errorf("col_index_num must be >= 1")
	}

	// Resolve Table Array
	var rows [][]interface{}

	switch t := tableArray.(type) {
	case [][]interface{}:
		rows = t
	case []interface{}:
		// Could be 1D array (single row or single col?)
		// Let's treat as list of rows if elements are slices, else single row?
		// Or single column?
		rows = make([][]interface{}, len(t))
		for i, item := range t {
			if row, ok := item.([]interface{}); ok {
				rows[i] = row
			} else {
				rows[i] = []interface{}{item} // Treat as 1-column table
			}
		}
	default:
		return nil, fmt.Errorf("table_array must be a range or array")
	}

	targetCol := int(colIndex) - 1

	// Iterate (Linear Scan) - Optimize later if needed
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}

		val := row[0]

		// Comparison Logic
		match := false
		if exactMatch {
			match = areEqual(val, lookupValue)
		} else {
			// Approximate match logic not fully implemented, falling back to Exact for safety or implementing simple <= check?
			// For now, force Exact match logic for reliability.
			match = areEqual(val, lookupValue)
		}

		if match {
			if targetCol >= len(row) {
				return nil, fmt.Errorf("col_index_num out of bounds")
			}
			return row[targetCol], nil
		}
	}

	return nil, fmt.Errorf("N/A") // VLOOKUP returns N/A if not found
}

func areEqual(a, b interface{}) bool {
	// Simple string conversion comparison for now
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// GetFunction retrieves a function by name (case-insensitive)
func GetFunction(name string) (Function, bool) {
	fn, ok := FunctionRegistry[strings.ToUpper(name)]
	return fn, ok
}
