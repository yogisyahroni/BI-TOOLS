package formula_engine

import (
	"testing"
)

func TestEvaluateSeries(t *testing.T) {
	engine := NewFormulaEngine()

	data := []map[string]interface{}{
		{"Sales": 100.0, "Cost": 50.0, "Region": "US"},
		{"Sales": 200.0, "Cost": 150.0, "Region": "EU"},
		{"Sales": 50.0, "Cost": 60.0, "Region": "US"},
	}

	tests := []struct {
		name    string
		formula string
		want    []interface{}
	}{
		{
			name:    "Simple Arithmetic",
			formula: "[Sales] - [Cost]",
			want:    []interface{}{50.0, 50.0, -10.0},
		},
		{
			name:    "Multiplication",
			formula: "[Sales] * 2",
			want:    []interface{}{200.0, 400.0, 100.0},
		},
		{
			name:    "Conditional Logic",
			formula: "IF([Sales] > 100, \"High\", \"Low\")",
			want:    []interface{}{"Low", "High", "Low"},
		},
		{
			name:    "Unknown Field",
			formula: "[Profit]",
			want:    []interface{}{"#ERROR: unresolved reference: Profit", "#ERROR: unresolved reference: Profit", "#ERROR: unresolved reference: Profit"},
		},
		{
			name:    "String Comparison",
			formula: "IF([Region] = \"US\", [Sales] * 0.1, [Sales] * 0.2)",
			want:    []interface{}{10.0, 40.0, 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.EvaluateSeries(tt.formula, data)
			if err != nil {
				t.Fatalf("EvaluateSeries() error = %v", err)
			}

			if len(got) != len(tt.want) {
				t.Errorf("EvaluateSeries() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Row %d: got %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
