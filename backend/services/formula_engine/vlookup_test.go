package formula_engine

import (
	"testing"
)

func TestEvaluate_VLookup(t *testing.T) {
	engine := NewFormulaEngine()

	// Prepare data: 2D array
	// ID | Name      | Age
	// 1  | "Alice"   | 30
	// 2  | "Bob"     | 25
	// 3  | "Charlie" | 35
	data := [][]interface{}{
		{1.0, "Alice", 30.0},
		{2.0, "Bob", 25.0},
		{3.0, "Charlie", 35.0},
	}

	ctx := &FormulaContext{
		FieldValues: map[string]interface{}{
			"Data": data,
		},
	}

	tests := []struct {
		formula string
		want    interface{}
		wantErr bool
	}{
		{"VLOOKUP(1, Data, 2, FALSE)", "Alice", false},
		{"VLOOKUP(2, Data, 3, 0)", 25.0, false}, // 0 works as FALSE
		{"VLOOKUP(3, Data, 2, FALSE)", "Charlie", false},
		{"VLOOKUP(99, Data, 2, FALSE)", nil, true}, // Not Found
		{"VLOOKUP(1, Data, 1, FALSE)", 1.0, false},
		{"VLOOKUP(2, Data, 4, FALSE)", nil, true}, // Col Index out of bounds
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			got, err := engine.Evaluate(tt.formula, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluate_NestedFlattening(t *testing.T) {
	// Verify that SUM still works with new flattening logic
	engine := NewFormulaEngine()
	ctx := &FormulaContext{
		FieldValues: map[string]interface{}{
			"Row1": []interface{}{1, 2, 3},
			"Row2": []interface{}{4, 5, 6},
			"Matrix": [][]interface{}{
				{10, 20},
				{30, 40},
			},
		},
	}

	tests := []struct {
		formula string
		want    interface{}
	}{
		{"SUM(Row1)", 6.0},
		{"SUM(Row1, Row2)", 21.0},
		{"SUM(Matrix)", 100.0},
		{"MAX(Matrix)", 40.0},
		{"MIN(Matrix)", 10.0},
		{"AVG(Matrix)", 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			got, err := engine.Evaluate(tt.formula, ctx)
			if err != nil {
				t.Errorf("Evaluate() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
