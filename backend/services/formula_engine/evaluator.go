package formula_engine

import (
	"fmt"
	"math"
	"strings"
)

// FormulaContext provides values for cell references and field lookups
type FormulaContext struct {
	// CellValues maps cell refs (e.g. "A1") to values
	CellValues map[string]interface{}
	// FieldValues maps field/column names to values (for row-level formulas)
	FieldValues map[string]interface{}
	// RangeResolver resolves a range (e.g., "A1:A10") into a list of values
	RangeResolver func(start string, end string) ([]interface{}, error)
}

// Evaluate evaluates a formula AST node against the provided context
func (e *FormulaEngine) evalNode(node FormulaNode, ctx *FormulaContext) (interface{}, error) {
	switch n := node.(type) {
	case *NumberNode:
		return n.Value, nil
	case *StringNode:
		return n.Value, nil
	case *BoolNode:
		return n.Value, nil
	case *CellRefNode:
		return e.resolveRef(n, ctx)
	case *UnaryNode:
		return e.evalUnary(n, ctx)
	case *BinaryNode:
		return e.evalBinary(n, ctx)
	case *FuncCallNode:
		return e.evalFunc(n, ctx)
	default:
		return nil, fmt.Errorf("unknown node type: %T", node)
	}
}

func (e *FormulaEngine) resolveRef(n *CellRefNode, ctx *FormulaContext) (interface{}, error) {
	if ctx == nil {
		return nil, fmt.Errorf("no context provided for reference '%s'", n.Ref)
	}

	// Range resolution
	if n.RangeEnd != "" {
		if ctx.RangeResolver != nil {
			return ctx.RangeResolver(n.Ref, n.RangeEnd)
		}
		return nil, fmt.Errorf("range resolution not supported in this context")
	}

	// Try cell values first
	if ctx.CellValues != nil {
		if val, ok := ctx.CellValues[n.Ref]; ok {
			return val, nil
		}
	}

	// Try field values
	if ctx.FieldValues != nil {
		if val, ok := ctx.FieldValues[n.Ref]; ok {
			return val, nil
		}
		// Case-insensitive lookup
		for k, v := range ctx.FieldValues {
			if strings.EqualFold(k, n.Ref) {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf("unresolved reference: %s", n.Ref)
}

func (e *FormulaEngine) evalUnary(n *UnaryNode, ctx *FormulaContext) (interface{}, error) {
	operand, err := e.evalNode(n.Operand, ctx)
	if err != nil {
		return nil, err
	}

	num, err := toFloat64(operand)
	if err != nil {
		return nil, fmt.Errorf("unary operator requires numeric operand: %w", err)
	}

	if n.Op == TokMinus {
		return -num, nil
	}
	return num, nil
}

func (e *FormulaEngine) evalBinary(n *BinaryNode, ctx *FormulaContext) (interface{}, error) {
	left, err := e.evalNode(n.Left, ctx)
	if err != nil {
		return nil, err
	}
	right, err := e.evalNode(n.Right, ctx)
	if err != nil {
		return nil, err
	}

	// String concatenation
	if n.Op == TokAmpersand {
		return fmt.Sprintf("%v%v", left, right), nil
	}

	// Comparison operators (work on both numbers and strings)
	switch n.Op {
	case TokEq:
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
	case TokNeq:
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right), nil
	}

	// Numeric operations
	lNum, lErr := toFloat64(left)
	rNum, rErr := toFloat64(right)

	// Comparison with numeric coercion
	switch n.Op {
	case TokLt:
		if lErr == nil && rErr == nil {
			return lNum < rNum, nil
		}
		return fmt.Sprintf("%v", left) < fmt.Sprintf("%v", right), nil
	case TokGt:
		if lErr == nil && rErr == nil {
			return lNum > rNum, nil
		}
		return fmt.Sprintf("%v", left) > fmt.Sprintf("%v", right), nil
	case TokLte:
		if lErr == nil && rErr == nil {
			return lNum <= rNum, nil
		}
		return fmt.Sprintf("%v", left) <= fmt.Sprintf("%v", right), nil
	case TokGte:
		if lErr == nil && rErr == nil {
			return lNum >= rNum, nil
		}
		return fmt.Sprintf("%v", left) >= fmt.Sprintf("%v", right), nil
	}

	// Arithmetic — both must be numeric
	if lErr != nil {
		return nil, fmt.Errorf("left operand is not numeric: %v", left)
	}
	if rErr != nil {
		return nil, fmt.Errorf("right operand is not numeric: %v", right)
	}

	switch n.Op {
	case TokPlus:
		return lNum + rNum, nil
	case TokMinus:
		return lNum - rNum, nil
	case TokStar:
		return lNum * rNum, nil
	case TokSlash:
		if rNum == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return lNum / rNum, nil
	case TokPercent:
		if rNum == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return math.Mod(lNum, rNum), nil
	case TokCaret:
		return math.Pow(lNum, rNum), nil
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", n.Op.String())
	}
}

func (e *FormulaEngine) evalFunc(n *FuncCallNode, ctx *FormulaContext) (interface{}, error) {
	fn, ok := GetFunction(n.Name)
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", n.Name)
	}

	// Evaluate arguments using 'e' reference to call evalNode
	// But Wait, evalNode is a method on e.
	// We need to evaluate all arguments before passing to function
	// Except for IF, which might be lazy?
	// For standard functions, strict evaluation is easier.
	// For IF, we should specialize or let the function handle raw nodes?
	// Standard excel evaluates args first usually, but IF is special.
	// CURRENT ARCHITECTURE: Evaluate args first. (Simplification)
	// TODO: Support lazy evaluation for control flow functions by passing nodes + context to function?
	// For now, Eager evaluation.

	// Special case for IF to support lazy eval if we wanted, but let's stick to eager for simplicity unless broken.
	// Actually IF needs lazy eval to avoid DivByZero in dead branches.
	// e.g. IF(B1=0, 0, A1/B1). If eager, A1/B1 crashes.
	// Let's special case IF here.

	if n.Name == "IF" {
		if len(n.Args) < 2 || len(n.Args) > 3 {
			return nil, fmt.Errorf("IF requires 2 or 3 arguments")
		}
		cond, err := e.evalNode(n.Args[0], ctx)
		if err != nil {
			return nil, err
		}

		condBool := false
		switch v := cond.(type) {
		case bool:
			condBool = v
		case float64:
			condBool = v != 0
		}

		if condBool {
			return e.evalNode(n.Args[1], ctx)
		}
		if len(n.Args) == 3 {
			return e.evalNode(n.Args[2], ctx)
		}
		return false, nil
	}

	args := make([]interface{}, len(n.Args))
	for i, argNode := range n.Args {
		val, err := e.evalNode(argNode, ctx)
		if err != nil {
			return nil, err
		}

		// If the argument evaluated to a slice (range), flatten it?
		// Or pass it as is?
		// Functions like SUM take ranges. Functions like SQRT take single numbers.
		// Let's pass as is, function logic handles it.
		// Wait, resolveRef for range returns []interface{}.

		// Flatten logic for variadic functions like SUM(A1, B1:B5)?
		// The `funcSum` iterates. If it encounters a slice, it should flatten.
		// Let's allow `funcSum` to handle slices.

		args[i] = val
	}

	// Post-processing for Aggregates:
	// If a function receives a slice as an arg, it needs to handle it.
	// Our `funcSum` expects `interface{}` but looks for float64.
	// We need to update `registry.go` to handle slices in SUM/AVG etc. if we pass them here.

	// Let's verify `funcSum` in `registry.go`...
	// It treats `arg` as `interface{}` and calls `toFloat64`.
	// `toFloat64` fails on slice.
	// Fix required in `registry.go` or here.
	// Better to fix in `registry.go` or make `toFloat64` handle flattening?
	// Flattening here is safer for the generic engine.

	flatArgs := make([]interface{}, 0, len(args))
	for _, arg := range args {
		if slice, ok := arg.([]interface{}); ok {
			flatArgs = append(flatArgs, slice...)
		} else {
			flatArgs = append(flatArgs, arg)
		}
	}

	// Only flatten for aggregation functions?
	// VLOOKUP expects a Range as a range, not flattened.
	// Complex... for now let's pass `args` and let function decide?
	// But `funcSum` in registry.go is simple.
	// Let's use `flatArgs` for now, assuming mostly Aggregations.
	// TODO: Add metadata to FunctionRegistry about whether it accepts ranges or scalars.
	// For 100% parity, we need robust handling.

	// Assumption: All implemented functions (SUM, AVG, MIN, MAX) handle flat lists.
	// VLOOKUP isn't fully implemented yet.

	return fn(flatArgs)
}
