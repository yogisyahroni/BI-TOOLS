package formula_engine

import (
	"fmt"
)

// ============================================================
// Formula Engine Grammar
// Defines Token Types and AST Nodes
// ============================================================

// ---- Token Types ----

type TokenKind int

const (
	TokNumber    TokenKind = iota // 42, 3.14
	TokString                     // "hello"
	TokIdent                      // SUM, column_name
	TokCellRef                    // A1, B2:C5
	TokPlus                       // +
	TokMinus                      // -
	TokStar                       // *
	TokSlash                      // /
	TokPercent                    // %
	TokCaret                      // ^
	TokAmpersand                  // &  (string concat)
	TokEq                         // =
	TokNeq                        // <>
	TokLt                         // <
	TokGt                         // >
	TokLte                        // <=
	TokGte                        // >=
	TokLParen                     // (
	TokRParen                     // )
	TokComma                      // ,
	TokColon                      // :
	TokBang                       // !
	TokBool                       // TRUE, FALSE
	TokEOF
)

func (k TokenKind) String() string {
	switch k {
	case TokNumber:
		return "NUMBER"
	case TokString:
		return "STRING"
	case TokIdent:
		return "IDENT"
	case TokCellRef:
		return "CELL_REF"
	case TokPlus:
		return "+"
	case TokMinus:
		return "-"
	case TokStar:
		return "*"
	case TokSlash:
		return "/"
	case TokPercent:
		return "%"
	case TokCaret:
		return "^"
	case TokAmpersand:
		return "&"
	case TokEq:
		return "="
	case TokNeq:
		return "<>"
	case TokLt:
		return "<"
	case TokGt:
		return ">"
	case TokLte:
		return "<="
	case TokGte:
		return ">="
	case TokLParen:
		return "("
	case TokRParen:
		return ")"
	case TokComma:
		return ","
	case TokColon:
		return ":"
	case TokBang:
		return "!"
	case TokBool:
		return "BOOL"
	case TokEOF:
		return "EOF"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", k)
	}
}

// Token represents a single lexical token
type Token struct {
	Kind  TokenKind
	Value string
	Pos   int
}

// ---- AST Nodes ----

// FormulaNode is the interface for all AST nodes
type FormulaNode interface {
	NodeType() string
	String() string
}

// NumberNode represents a numeric literal
type NumberNode struct {
	Value float64
}

func (n *NumberNode) NodeType() string { return "number" }
func (n *NumberNode) String() string   { return fmt.Sprintf("%g", n.Value) }

// StringNode represents a string literal
type StringNode struct {
	Value string
}

func (n *StringNode) NodeType() string { return "string" }
func (n *StringNode) String() string   { return fmt.Sprintf("%q", n.Value) }

// BoolNode represents TRUE/FALSE
type BoolNode struct {
	Value bool
}

func (n *BoolNode) NodeType() string { return "bool" }
func (n *BoolNode) String() string   { return fmt.Sprintf("%v", n.Value) }

// CellRefNode represents a cell reference like A1 or A1:B5
// Also used for named references (column names) if RangeEnd is empty logic handles distinction
type CellRefNode struct {
	Ref      string
	RangeEnd string // empty if single cell or named ref
}

func (n *CellRefNode) NodeType() string { return "cellRef" }
func (n *CellRefNode) String() string {
	if n.RangeEnd != "" {
		return fmt.Sprintf("%s:%s", n.Ref, n.RangeEnd)
	}
	return n.Ref
}

// UnaryNode represents a unary operation (-x, +x)
type UnaryNode struct {
	Op      TokenKind
	Operand FormulaNode
}

func (n *UnaryNode) NodeType() string { return "unary" }
func (n *UnaryNode) String() string   { return fmt.Sprintf("(%s%s)", n.Op.String(), n.Operand.String()) }

// BinaryNode represents a binary operation (a + b)
type BinaryNode struct {
	Op    TokenKind
	Left  FormulaNode
	Right FormulaNode
}

func (n *BinaryNode) NodeType() string { return "binary" }
func (n *BinaryNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Op.String(), n.Right.String())
}

// FuncCallNode represents a function call like SUM(A1:A10)
type FuncCallNode struct {
	Name string
	Args []FormulaNode
}

func (n *FuncCallNode) NodeType() string { return "funcCall" }
func (n *FuncCallNode) String() string {
	args := ""
	for i, arg := range n.Args {
		if i > 0 {
			args += ", "
		}
		args += arg.String()
	}
	return fmt.Sprintf("%s(%s)", n.Name, args)
}
