package formula_engine

import (
	"testing"
	"time"
)

func TestExtendedFunctions(t *testing.T) {
	engine := NewFormulaEngine()

	tests := []struct {
		name    string
		formula string
		want    interface{} // Use appropriate type matching logic esp. for float64
	}{
		// Text Functions
		{"UPPER", "UPPER(\"hello\")", "HELLO"},
		{"LOWER", "LOWER(\"HELLO\")", "hello"},
		{"CONCAT", "CONCAT(\"Hello\", \" \", \"World\")", "Hello World"},
		{"LEN", "LEN(\"hello\")", 5.0},
		{"TRIM", "TRIM(\"  hello  \")", "hello"},
		{"LEFT", "LEFT(\"hello\", 2)", "he"},
		{"RIGHT", "RIGHT(\"hello\", 2)", "lo"},

		// Date Functions (Mocking time is hard, so we test logic logic like YEAR/MONTH with fixed strings)
		{"YEAR", "YEAR(\"2023-10-05\")", 2023.0},
		{"MONTH", "MONTH(\"2023-10-05\")", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Evaluate(tt.formula, nil)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if wantStr, ok := tt.want.(string); ok {
				if got != wantStr {
					t.Errorf("got %v, want %v", got, wantStr)
				}
			} else if wantFloat, ok := tt.want.(float64); ok {
				gotFloat, err := toFloat64(got)
				if err != nil {
					t.Errorf("got result %v which is not float64 convertable", got)
				}
				if gotFloat != wantFloat {
					t.Errorf("got %v, want %v", gotFloat, wantFloat)
				}
			}
		})
	}
}

func TestDateNow(t *testing.T) {
	// Custom test for NOW() to just check no error and returns time.Time
	engine := NewFormulaEngine()
	got, err := engine.Evaluate("NOW()", nil)
	if err != nil {
		t.Fatalf("NOW() error = %v", err)
	}
	if _, ok := got.(time.Time); !ok {
		t.Errorf("NOW() returned %T, want time.Time", got)
	}
}
