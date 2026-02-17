package formula_engine

import (
	"fmt"
	"strings"
)

// FormulaLexer tokenizes a formula string
type FormulaLexer struct {
	input  string
	pos    int
	tokens []Token
}

// NewLexer creates a new lexer instance
func NewLexer(input string) *FormulaLexer {
	return &FormulaLexer{input: input}
}

// Tokenize breaks the input into tokens
func (l *FormulaLexer) Tokenize() ([]Token, error) {
	l.tokens = nil
	l.pos = 0

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.pos++
			continue
		}

		// Numbers
		if ch >= '0' && ch <= '9' {
			l.readNumber()
			continue
		}

		// Strings (double-quoted)
		if ch == '"' {
			if err := l.readString(); err != nil {
				return nil, err
			}
			continue
		}

		// Identifiers / Keywords / Cell refs
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' {
			l.readIdentOrCellRef()
			continue
		}

		// Operators & punctuation
		switch ch {
		// Bracketed identifiers (e.g., [Sales], [Field Name])
		case '[':
			if err := l.readBracketedIdent(); err != nil {
				return nil, err
			}

		case '+':
			l.emit(TokPlus, "+")
		case '-':
			l.emit(TokMinus, "-")
		case '*':
			l.emit(TokStar, "*")
		case '/':
			l.emit(TokSlash, "/")
		case '%':
			l.emit(TokPercent, "%")
		case '^':
			l.emit(TokCaret, "^")
		case '&':
			l.emit(TokAmpersand, "&")
		case '(':
			l.emit(TokLParen, "(")
		case ')':
			l.emit(TokRParen, ")")
		case ',':
			l.emit(TokComma, ",")
		case ':':
			l.emit(TokColon, ":")
		case '!':
			l.emit(TokBang, "!")
		case '=':
			l.emit(TokEq, "=")
		case '<':
			if l.peek() == '>' {
				l.pos++
				l.emit(TokNeq, "<>")
			} else if l.peek() == '=' {
				l.pos++
				l.emit(TokLte, "<=")
			} else {
				l.emit(TokLt, "<")
			}
		case '>':
			if l.peek() == '=' {
				l.pos++
				l.emit(TokGte, ">=")
			} else {
				l.emit(TokGt, ">")
			}
		default:
			return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, l.pos)
		}
	}

	l.tokens = append(l.tokens, Token{Kind: TokEOF, Value: "", Pos: l.pos})
	return l.tokens, nil
}

func (l *FormulaLexer) emit(kind TokenKind, value string) {
	l.tokens = append(l.tokens, Token{Kind: kind, Value: value, Pos: l.pos})
	l.pos++
}

func (l *FormulaLexer) peek() byte {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *FormulaLexer) readNumber() {
	start := l.pos
	hasDot := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '.' && !hasDot {
			hasDot = true
			l.pos++
		} else if ch >= '0' && ch <= '9' {
			l.pos++
		} else {
			break
		}
	}
	l.tokens = append(l.tokens, Token{Kind: TokNumber, Value: l.input[start:l.pos], Pos: start})
}

func (l *FormulaLexer) readString() error {
	l.pos++ // skip opening "
	start := l.pos
	for l.pos < len(l.input) {
		if l.input[l.pos] == '"' {
			val := l.input[start:l.pos]
			l.pos++ // skip closing "
			l.tokens = append(l.tokens, Token{Kind: TokString, Value: val, Pos: start - 1})
			return nil
		}
		l.pos++
	}
	return fmt.Errorf("unterminated string starting at position %d", start-1)
}

func (l *FormulaLexer) readIdentOrCellRef() {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '.' {
			l.pos++
		} else {
			break
		}
	}

	word := l.input[start:l.pos]
	upper := strings.ToUpper(word)

	// Boolean keywords
	if upper == "TRUE" {
		l.tokens = append(l.tokens, Token{Kind: TokBool, Value: "TRUE", Pos: start})
		return
	}
	if upper == "FALSE" {
		l.tokens = append(l.tokens, Token{Kind: TokBool, Value: "FALSE", Pos: start})
		return
	}

	// Check if it looks like a cell reference (e.g., A1, AB123, $A$1)
	if isCellRef(word) {
		l.tokens = append(l.tokens, Token{Kind: TokCellRef, Value: upper, Pos: start})
		return
	}

	l.tokens = append(l.tokens, Token{Kind: TokIdent, Value: upper, Pos: start})
}

func isCellRef(s string) bool {
	s = strings.ToUpper(s)
	// Check for strict cell ref format: [A-Z]+[0-9]+
	// Supporting absolute refs like $A$1 is optional based on spec, but let's keep it robust
	i := 0
	// Skip $ prefix
	if i < len(s) && s[i] == '$' {
		i++
	}
	// Must start with A-Z
	if i >= len(s) || s[i] < 'A' || s[i] > 'Z' {
		return false
	}
	for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
		i++
	}
	// Skip $ before row
	if i < len(s) && s[i] == '$' {
		i++
	}
	// Must end with digits
	if i >= len(s) || s[i] < '0' || s[i] > '9' {
		return false
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	// If we consumed the whole string, it's a cell ref
	return i == len(s)
}

func (l *FormulaLexer) readBracketedIdent() error {
	l.pos++ // skip opening [
	start := l.pos
	for l.pos < len(l.input) {
		if l.input[l.pos] == ']' {
			val := l.input[start:l.pos]
			l.pos++ // skip closing ]
			// Emit as TokIdent so parser treats it as a reference/identifier
			l.tokens = append(l.tokens, Token{Kind: TokIdent, Value: val, Pos: start - 1})
			return nil
		}
		l.pos++
	}
	return fmt.Errorf("unterminated bracketed identifier starting at position %d", start-1)
}
