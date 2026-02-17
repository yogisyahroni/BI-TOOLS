package formula_engine

import (
	"fmt"
	"strconv"
)

// FormulaParser parses tokens into an AST
type FormulaParser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser instance
func NewParser(tokens []Token) *FormulaParser {
	return &FormulaParser{tokens: tokens, pos: 0}
}

// Parse entry point
func (p *FormulaParser) Parse() (FormulaNode, error) {
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.current().Kind != TokEOF {
		return nil, fmt.Errorf("unexpected token '%s' at position %d", p.current().Value, p.current().Pos)
	}
	return node, nil
}

func (p *FormulaParser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TokEOF}
	}
	return p.tokens[p.pos]
}

func (p *FormulaParser) advance() Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *FormulaParser) expect(kind TokenKind) (Token, error) {
	tok := p.current()
	if tok.Kind != kind {
		return tok, fmt.Errorf("expected token type %s but got '%s' at position %d", kind.String(), tok.Value, tok.Pos)
	}
	p.pos++
	return tok, nil
}

// Expression precedence (lowest to highest):
// comparison  ( = <> < > <= >= )
// concat      ( & )
// additive    ( + - )
// multiplicative ( * / % )
// power       ( ^ )
// unary       ( -x +x )
// primary     ( number, string, bool, cellRef, funcCall, grouped )

func (p *FormulaParser) parseExpression() (FormulaNode, error) {
	return p.parseComparison()
}

func (p *FormulaParser) parseComparison() (FormulaNode, error) {
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}

	for {
		kind := p.current().Kind
		if kind == TokEq || kind == TokNeq || kind == TokLt || kind == TokGt || kind == TokLte || kind == TokGte {
			p.advance()
			right, err := p.parseConcat()
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{Op: kind, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *FormulaParser) parseConcat() (FormulaNode, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for p.current().Kind == TokAmpersand {
		p.advance()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Op: TokAmpersand, Left: left, Right: right}
	}
	return left, nil
}

func (p *FormulaParser) parseAdditive() (FormulaNode, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for {
		kind := p.current().Kind
		if kind == TokPlus || kind == TokMinus {
			p.advance()
			right, err := p.parseMultiplicative()
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{Op: kind, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *FormulaParser) parseMultiplicative() (FormulaNode, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}

	for {
		kind := p.current().Kind
		if kind == TokStar || kind == TokSlash || kind == TokPercent {
			p.advance()
			right, err := p.parsePower()
			if err != nil {
				return nil, err
			}
			left = &BinaryNode{Op: kind, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *FormulaParser) parsePower() (FormulaNode, error) {
	base, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	if p.current().Kind == TokCaret {
		p.advance()
		exp, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &BinaryNode{Op: TokCaret, Left: base, Right: exp}, nil
	}
	return base, nil
}

func (p *FormulaParser) parseUnary() (FormulaNode, error) {
	if p.current().Kind == TokMinus || p.current().Kind == TokPlus {
		op := p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{Op: op.Kind, Operand: operand}, nil
	}
	return p.parsePrimary()
}

func (p *FormulaParser) parsePrimary() (FormulaNode, error) {
	tok := p.current()

	switch tok.Kind {
	case TokNumber:
		p.advance()
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", tok.Value)
		}
		return &NumberNode{Value: val}, nil

	case TokString:
		p.advance()
		return &StringNode{Value: tok.Value}, nil

	case TokBool:
		p.advance()
		return &BoolNode{Value: tok.Value == "TRUE"}, nil

	case TokCellRef:
		p.advance()
		ref := tok.Value
		rangeEnd := ""
		if p.current().Kind == TokColon {
			p.advance()
			endTok, err := p.expect(TokCellRef)
			if err != nil {
				return nil, fmt.Errorf("expected cell reference after ':' at position %d", p.current().Pos)
			}
			rangeEnd = endTok.Value
		}
		return &CellRefNode{Ref: ref, RangeEnd: rangeEnd}, nil

	case TokIdent:
		// Could be a function call
		name := tok.Value
		p.advance()
		if p.current().Kind == TokLParen {
			return p.parseFuncCall(name)
		}
		// Otherwise treat as a named reference (column/field name)
		// Or maybe just an identifier. For now, we reuse CellRefNode which handles named refs too.
		return &CellRefNode{Ref: name}, nil

	case TokLParen:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokRParen); err != nil {
			return nil, fmt.Errorf("expected ')' at position %d", p.current().Pos)
		}
		return expr, nil

	default:
		return nil, fmt.Errorf("unexpected token '%s' at position %d", tok.Value, tok.Pos)
	}
}

func (p *FormulaParser) parseFuncCall(name string) (FormulaNode, error) {
	p.advance() // skip (
	var args []FormulaNode

	if p.current().Kind != TokRParen {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		for p.current().Kind == TokComma {
			p.advance()
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	if _, err := p.expect(TokRParen); err != nil {
		return nil, fmt.Errorf("expected ')' after function arguments at position %d", p.current().Pos)
	}

	return &FuncCallNode{Name: name, Args: args}, nil
}
