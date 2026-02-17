package formula_engine

import (
	"fmt"
	"strings"
)

// FormulaEngine is the top-level entry point for formula parsing and evaluation
type FormulaEngine struct{}

// NewFormulaEngine creates a new formula engine
func NewFormulaEngine() *FormulaEngine {
	return &FormulaEngine{}
}

// ParseFormula tokenizes and parses a formula string into an AST
func (e *FormulaEngine) ParseFormula(formula string) (FormulaNode, error) {
	// Strip leading = if present (Excel convention)
	formula = strings.TrimSpace(formula)
	if len(formula) > 0 && formula[0] == '=' {
		formula = formula[1:]
	}

	lexer := NewLexer(formula)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("tokenize error: %w", err)
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

// Evaluate evaluates a formula string against the provided context
func (e *FormulaEngine) Evaluate(formula string, ctx *FormulaContext) (interface{}, error) {
	node, err := e.ParseFormula(formula)
	if err != nil {
		return nil, err
	}
	return e.evalNode(node, ctx)
}

// Validate checks if a formula is syntactically valid
func (e *FormulaEngine) Validate(formula string) error {
	_, err := e.ParseFormula(formula)
	return err
}

// ExtractReferences extracts all cell/field references from a formula
func (e *FormulaEngine) ExtractReferences(formula string) ([]string, error) {
	node, err := e.ParseFormula(formula)
	if err != nil {
		return nil, err
	}
	var refs []string
	e.collectRefs(node, &refs)
	return refs, nil
}

func (e *FormulaEngine) collectRefs(node FormulaNode, refs *[]string) {
	switch n := node.(type) {
	case *CellRefNode:
		*refs = append(*refs, n.Ref)
		if n.RangeEnd != "" {
			*refs = append(*refs, n.RangeEnd)
		}
	case *BinaryNode:
		e.collectRefs(n.Left, refs)
		e.collectRefs(n.Right, refs)
	case *UnaryNode:
		e.collectRefs(n.Operand, refs)
	case *FuncCallNode:
		for _, arg := range n.Args {
			e.collectRefs(arg, refs)
		}
	}
}

// EvaluateSeries evaluates a formula against a dataset (list of rows)
// It parses the formula once and evaluates it for each row
func (e *FormulaEngine) EvaluateSeries(formula string, data []map[string]interface{}) ([]interface{}, error) {
	// 1. Parse formula once
	node, err := e.ParseFormula(formula)
	if err != nil {
		return nil, err
	}

	results := make([]interface{}, len(data))

	// 2. Iterate and evaluate
	for i, row := range data {
		ctx := &FormulaContext{
			FieldValues: row,
		}

		val, err := e.evalNode(node, ctx)
		if err != nil {
			// Check if we should fail fast or return error for this row?
			// For now, return error string as value or nil?
			// Excel returns #ERROR. Let's return error message or nil.
			// Better: return nil and logic layer handles it, or return specific Error type.
			// Let's return the error as a string value for now to keep alignment.
			results[i] = fmt.Sprintf("#ERROR: %v", err)
		} else {
			results[i] = val
		}
	}

	return results, nil
}
