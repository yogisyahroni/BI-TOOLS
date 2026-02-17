package formula_engine

import (
	"testing"
)

func TestLexer_Tokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenKind
		values   []string // optional, checks values if provided
	}{
		{
			name:     "Simple Arithmetic",
			input:    "1 + 2 * 3",
			expected: []TokenKind{TokNumber, TokPlus, TokNumber, TokStar, TokNumber, TokEOF},
		},
		{
			name:     "Function Call",
			input:    "SUM(A1, 10)",
			expected: []TokenKind{TokIdent, TokLParen, TokCellRef, TokComma, TokNumber, TokRParen, TokEOF},
		},
		{
			name:     "Cell Range",
			input:    "A1:B5",
			expected: []TokenKind{TokCellRef, TokColon, TokCellRef, TokEOF},
		},
		{
			name:     "String and Bool",
			input:    `"hello" & TRUE`,
			expected: []TokenKind{TokString, TokAmpersand, TokBool, TokEOF},
		},
		{
			name:     "Complex Expression",
			input:    "IF(A1 > 0, B1 * 2, 0)",
			expected: []TokenKind{TokIdent, TokLParen, TokCellRef, TokGt, TokNumber, TokComma, TokCellRef, TokStar, TokNumber, TokComma, TokNumber, TokRParen, TokEOF},
		},
		{
			name:     "Comparison Operators",
			input:    "A <> B <= C >= D",
			expected: []TokenKind{TokIdent, TokNeq, TokIdent, TokLte, TokIdent, TokGte, TokIdent, TokEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, tok := range tokens {
				if tok.Kind != tt.expected[i] {
					t.Errorf("token %d: expected kind %s, got %s (value: %s)", i, tt.expected[i], tok.Kind, tok.Value)
				}
			}
		})
	}
}

func TestLexer_Error(t *testing.T) {
	input := `"unterminated string`
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Error("expected error for unterminated string, got nil")
	}
}
