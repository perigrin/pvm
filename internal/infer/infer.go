// ABOUTME: Pass 2 of the PSC type inference engine — the inference walker.
// ABOUTME: Analyze walks the CST bottom-up, annotating every node with a types.Type.

package infer

import (
	"strings"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// Analyze runs both inference passes over the parsed tree and source and
// returns an annotation map and a slice of diagnostics.
//
// The annotation map keys are node StartByte values; each value is the
// inferred types.Type for the node starting at that byte offset.
// Diagnostics describe any type errors or arity mismatches found.
func Analyze(tree *parser.Tree, source []byte) (map[uint32]types.Type, []Diagnostic) {
	annotations := make(map[uint32]types.Type)
	diags := make([]Diagnostic, 0)

	if tree == nil {
		return annotations, diags
	}
	root := tree.RootNode()
	if root == nil {
		return annotations, diags
	}

	// Pass 1: collect declarations into a symbol table.
	st := CollectDeclarations(tree, source)

	// Pass 2: bottom-up type walk.
	walkNode(root, source, st, annotations, &diags)

	return annotations, diags
}

// walkNode performs a post-order (bottom-up) traversal of the CST, computing
// a types.Type for every node that has a meaningful type and storing it in the
// annotations map keyed by the node's StartByte.
func walkNode(node *parser.Node, source []byte, st *SymbolTable, annotations map[uint32]types.Type, diags *[]Diagnostic) types.Type {
	if node == nil {
		return types.Unknown
	}

	// Recurse into all children first (post-order).
	childTypes := make([]types.Type, node.ChildCount())
	for i := 0; i < node.ChildCount(); i++ {
		childTypes[i] = walkNode(node.Child(i), source, st, annotations, diags)
	}

	typ := inferNodeType(node, source, st, annotations, childTypes, diags)
	if typ != types.Unknown {
		annotations[node.StartByte()] = typ
	}
	return typ
}

// inferNodeType dispatches on node.Kind() and returns the inferred type for
// the node, emitting any diagnostics as a side effect.
func inferNodeType(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	childTypes []types.Type,
	diags *[]Diagnostic,
) types.Type {
	kind := node.Kind()

	switch kind {

	// --- Literals ---

	case "number":
		return inferNumberType(node.Text(source))

	// --- Variables ---

	case "scalar":
		return types.Scalar

	case "array":
		return types.Array

	case "hash":
		return types.Hash

	case "arraylen":
		return types.Int

	// --- Binary expression families ---
	// The Perl grammar uses several node kinds for binary expressions,
	// each grouping operators at a similar precedence level.

	case "binary_expression",
		"equality_expression",
		"relational_expression",
		"lowprec_logical_expression":
		return inferBinaryExprType(node, source)

	// --- Unary expressions ---

	case "unary_expression":
		return inferUnaryExprType(node, source)

	// --- Function call expressions ---

	case "function_call_expression",
		"ambiguous_function_call_expression":
		return inferFunctionCallType(node, source, annotations, diags)

	case "func1op_call_expression":
		return inferFunc1opCallType(node, source, annotations, diags)

	case "func0op_call_expression":
		return inferFunc0opCallType(node, source, annotations, diags)

	// --- Method calls ---

	case "method_call_expression":
		return types.Any

	// --- Ternary / conditional ---

	case "conditional_expression":
		return types.Any
	}

	return types.Unknown
}

// inferNumberType determines whether the number text represents an integer
// or a floating-point value, returning Int or Num accordingly.
//
// Hex literals (starting with "0x" or "0X") are always Int even if they
// contain the letter 'e' or 'E'. For all other numbers, the presence of '.'
// or 'e'/'E' signals a floating-point value.
func inferNumberType(text string) types.Type {
	// Hex literals: 0x... — always integer regardless of hex digits.
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		return types.Int
	}
	// Octal/binary literals: 0b... or 0... — always integer.
	if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
		return types.Int
	}
	// Decimal float: contains '.' or 'e'/'E'.
	if strings.ContainsAny(text, ".eE") {
		return types.Num
	}
	return types.Int
}

// inferBinaryExprType finds the operator among the direct children of a
// binary-family expression node and looks it up in the type signatures table.
func inferBinaryExprType(node *parser.Node, source []byte) types.Type {
	op := findOperatorText(node, source)
	if op == "" {
		return types.Unknown
	}
	sig, ok := types.GetBinaryOp(op)
	if !ok {
		return types.Unknown
	}
	return sig.Result
}

// inferUnaryExprType finds the operator among the direct children of a
// unary_expression node and looks it up in the type signatures table.
func inferUnaryExprType(node *parser.Node, source []byte) types.Type {
	op := findOperatorText(node, source)
	if op == "" {
		return types.Unknown
	}
	sig, ok := types.GetUnaryOp(op)
	if !ok {
		return types.Unknown
	}
	return sig.Result
}

// findOperatorText returns the text of the first non-named (anonymous)
// child of node, which in the Perl grammar is the operator token.
func findOperatorText(node *parser.Node, source []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		// Anonymous nodes are punctuation/operators in tree-sitter grammars.
		if !child.IsNamed() {
			return child.Text(source)
		}
	}
	return ""
}

// inferFunctionCallType handles function_call_expression and
// ambiguous_function_call_expression nodes. It extracts the function name
// from a "function" child, looks it up as a builtin, validates arity and
// argument types, and returns the builtin's return type.
func inferFunctionCallType(
	node *parser.Node,
	source []byte,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	// Extract function name from the "function" child node.
	name := ""
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.Kind() == "function" {
			name = child.Text(source)
			break
		}
	}
	if name == "" {
		return types.Unknown
	}

	sig, ok := types.GetBuiltin(name)
	if !ok {
		return types.Unknown
	}

	// Collect argument nodes. Arguments live in a list_expression child
	// or directly as named children (for ambiguous_function_call_expression).
	args := collectCallArgs(node, source)

	// Validate arity.
	if len(args) < sig.MinArity {
		*diags = append(*diags, Diagnostic{
			StartByte: node.StartByte(),
			EndByte:   node.EndByte(),
			Severity:  Error,
			Code:      CodeArityMismatch,
			Message:   arityMessage(name, sig.MinArity, len(args)),
		})
		return sig.ReturnType
	}

	// Validate argument types.
	for i, arg := range args {
		argType := annotations[arg.StartByte()]
		expectedType := builtinArgType(sig, i)
		if argType != types.Unknown && !types.TypeSatisfies(argType, expectedType) {
			*diags = append(*diags, Diagnostic{
				StartByte: arg.StartByte(),
				EndByte:   arg.EndByte(),
				Severity:  Error,
				Code:      CodeTypeMismatch,
				Message:   typeMismatchMessage(name, i, expectedType, argType),
			})
		}
	}

	return sig.ReturnType
}

// inferFunc1opCallType handles func1op_call_expression nodes such as
// scalar(), defined(), length(), etc.
func inferFunc1opCallType(
	node *parser.Node,
	source []byte,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	// The function name is the first anonymous child (keyword token).
	name := ""
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && !child.IsNamed() {
			text := child.Text(source)
			// Skip punctuation like '(' and ')'
			if text != "(" && text != ")" {
				name = text
				break
			}
		}
	}
	if name == "" {
		return types.Unknown
	}

	sig, ok := types.GetBuiltin(name)
	if !ok {
		return types.Unknown
	}

	args := collectCallArgs(node, source)

	if len(args) < sig.MinArity {
		*diags = append(*diags, Diagnostic{
			StartByte: node.StartByte(),
			EndByte:   node.EndByte(),
			Severity:  Error,
			Code:      CodeArityMismatch,
			Message:   arityMessage(name, sig.MinArity, len(args)),
		})
		return sig.ReturnType
	}

	for i, arg := range args {
		argType := annotations[arg.StartByte()]
		expectedType := builtinArgType(sig, i)
		if argType != types.Unknown && !types.TypeSatisfies(argType, expectedType) {
			*diags = append(*diags, Diagnostic{
				StartByte: arg.StartByte(),
				EndByte:   arg.EndByte(),
				Severity:  Error,
				Code:      CodeTypeMismatch,
				Message:   typeMismatchMessage(name, i, expectedType, argType),
			})
		}
	}

	return sig.ReturnType
}

// inferFunc0opCallType handles func0op_call_expression nodes.
// These are zero-argument operators — the grammar currently doesn't emit
// them for common builtins, but we handle the node kind for completeness.
func inferFunc0opCallType(
	node *parser.Node,
	source []byte,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	name := ""
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && !child.IsNamed() {
			text := child.Text(source)
			if text != "(" && text != ")" {
				name = text
				break
			}
		}
	}
	if name == "" {
		return types.Unknown
	}
	sig, ok := types.GetBuiltin(name)
	if !ok {
		return types.Unknown
	}
	return sig.ReturnType
}

// collectCallArgs gathers the actual argument nodes for a function call.
// For calls with parentheses the arguments are inside a list_expression child;
// for ambiguous calls without parens, they may be direct children after the
// function name.
func collectCallArgs(node *parser.Node, source []byte) []*parser.Node {
	var args []*parser.Node

	// First, try to find a list_expression child.
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "list_expression" {
			// Items in a list_expression are named children separated by
			// anonymous "," tokens; collect the named children.
			for j := 0; j < child.ChildCount(); j++ {
				item := child.Child(j)
				if item != nil && item.IsNamed() {
					args = append(args, item)
				}
			}
			return args
		}
	}

	// No list_expression: collect named children that are not the function
	// name itself (i.e. skip "function" kind nodes and punctuation).
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			continue // skip punctuation
		}
		ck := child.Kind()
		if ck == "function" {
			continue // skip the function-name node
		}
		args = append(args, child)
	}
	return args
}

// builtinArgType returns the expected type for the i-th argument of a builtin,
// treating the last ArgType as variadic (repeated for all trailing arguments).
func builtinArgType(sig types.BuiltinSig, i int) types.Type {
	if len(sig.ArgTypes) == 0 {
		return types.Any
	}
	if i < len(sig.ArgTypes) {
		return sig.ArgTypes[i]
	}
	// Variadic: last element repeated.
	return sig.ArgTypes[len(sig.ArgTypes)-1]
}

// arityMessage produces a human-readable arity mismatch message.
func arityMessage(name string, min, got int) string {
	return "call to " + name + ": expected at least " +
		itoa(min) + " argument(s), got " + itoa(got)
}

// typeMismatchMessage produces a human-readable type mismatch message.
func typeMismatchMessage(name string, argIdx int, expected, actual types.Type) string {
	return "call to " + name + ": argument " + itoa(argIdx+1) +
		" expects " + expected.String() + ", got " + actual.String()
}

// itoa converts a non-negative integer to a decimal string without importing
// the strconv package.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
