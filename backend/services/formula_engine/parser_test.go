package formula_engine

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // String representation of AST
	}{
		{
			name:     "Simple Addition",
			input:    "1 + 2",
			expected: "(1 + 2)",
		},
		{
			name:     "Precedence Multiplication",
			input:    "1 + 2 * 3",
			expected: "(1 + (2 * 3))",
		},
		{
			name:     "Precedence Parentheses",
			input:    "(1 + 2) * 3",
			expected: "((1 + 2) * 3)",
		},
		{
			name:     "Function Call",
			input:    "SUM(A1, 10)",
			expected: "SUM(A1, 10)",
		},
		{
			name:     "Nested Function Call",
			input:    "MAX(MIN(1, 2), 3)",
			expected: "MAX(MIN(1, 2), 3)",
		},
		{
			name:     "String Concatenation",
			input:    `"Hello" & " " & "World"`,
			expected: `(("Hello" & " ") & "World")`,
		},
		{
			name:     "Comparison",
			input:    "A1 > 10",
			expected: "(A1 > 10)",
		},
		{
			name:     "Unary Operator",
			input:    "-5 + 2",
			expected: "((-5) + 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			parser := NewParser(tokens)
			node, err := parser.Parse()
			if err != nil {
				t.Fatalf("parser error: %v", err)
			}

			if node.String() != tt.expected {
				t.Errorf("expected AST %s, got %s", tt.expected, node.String())
			}
		})
	}
}

func TestParser_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Missing Closing Paren", "SUM(1, 2"},
		{"Unexpected Token", "1 + * 2"},
		{"Invalid Function Args", "SUM(1, )"}, // Empty arg not supported yet
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, _ := lexer.Tokenize() // ignore lexer errors for parser test
			parser := NewParser(tokens)
			_, err := parser.Parse()
			if err == nil {
				t.Error("expected parser error, got nil")
			}
		})
	}
}
