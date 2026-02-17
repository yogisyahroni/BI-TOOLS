package formula_engine

import (
	"math"
	"testing"
)

func TestEvaluator_Evaluate(t *testing.T) {
	ctx := &FormulaContext{
		CellValues: map[string]interface{}{
			"A1": 10.0,
			"A2": 20.0,
			"B1": "Hello",
			"C1": true,
		},
		FieldValues: map[string]interface{}{
			"revenue": 1000.0,
			"cost":    500.0,
		},
	}

	engine := NewFormulaEngine()

	tests := []struct {
		name     string
		formula  string
		expected interface{}
		wantErr  bool
	}{
		// Arithmetic
		{"Add", "1 + 2", 3.0, false},
		{"Subtract", "10 - 4", 6.0, false},
		{"Multiply", "3 * 4", 12.0, false},
		{"Divide", "10 / 2", 5.0, false},
		{"Power", "2 ^ 3", 8.0, false},
		{"Precedence", "1 + 2 * 3", 7.0, false},
		{"Parens", "(1 + 2) * 3", 9.0, false},

		// Comparison
		{"Eq Number", "1 = 1", true, false},
		{"Neq Number", "1 <> 2", true, false},
		{"Gt", "2 > 1", true, false},
		{"Lt", "1 < 2", true, false},
		{"Gte", "1 >= 1", true, false},
		{"Lte", "1 <= 2", true, false},
		{"Eq String", `"a" = "a"`, true, false},

		// Logic
		{"String Concat", `"Hello" & " " & "World"`, "Hello World", false},

		// References
		{"Cell Ref", "A1 + A2", 30.0, false},
		{"Field Ref", "revenue - cost", 500.0, false},
		{"Missing Ref", "Z99", nil, true},

		// Functions
		{"SUM", "SUM(1, 2, 3)", 6.0, false},
		{"SUM Ref", "SUM(A1, A2)", 30.0, false},
		{"AVG", "AVG(10, 20, 30)", 20.0, false},
		{"MIN", "MIN(10, 5, 20)", 5.0, false},
		{"MAX", "MAX(10, 5, 20)", 20.0, false},
		{"IF True", "IF(1 > 0, 100, 200)", 100.0, false},
		{"IF False", "IF(1 < 0, 100, 200)", 200.0, false},
		{"Nested Func", "MAX(SUM(1,2), 5)", 5.0, false},

		// Errors
		{"Div By Zero", "1 / 0", nil, true},
		{"Unknown Func", "FOOBAR(1)", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Evaluate(tt.formula, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Special check for float equality
				if f1, ok1 := got.(float64); ok1 {
					if f2, ok2 := tt.expected.(float64); ok2 {
						if math.Abs(f1-f2) > 1e-9 {
							t.Errorf("Evaluate() = %v, want %v", got, tt.expected)
						}
						return
					}
				}
				if got != tt.expected {
					t.Errorf("Evaluate() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
