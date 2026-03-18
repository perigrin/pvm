// ABOUTME: Pass 2 of the PSC type inference engine — the inference walker.
// ABOUTME: Analyze walks the CST bottom-up, annotating every node with a types.Type.

package infer

import (
	"strconv"
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
func Analyze(tree *parser.Tree, source []byte) (map[uint32]types.Type, []Diagnostic, *SymbolTable) {
	annotations := make(map[uint32]types.Type)
	diags := make([]Diagnostic, 0)

	if tree == nil {
		return annotations, diags, NewSymbolTable()
	}
	root := tree.RootNode()
	if root == nil {
		return annotations, diags, NewSymbolTable()
	}

	// Pass 1: collect declarations into a symbol table.
	st := CollectDeclarations(tree, source)

	// Pass 2: bottom-up type walk.
	walkNode(root, source, st, annotations, &diags)

	return annotations, diags, st
}

// walkNode performs a post-order (bottom-up) traversal of the CST, computing
// a types.Type for every node that has a meaningful type and storing it in the
// annotations map keyed by the node's StartByte.
func walkNode(node *parser.Node, source []byte, st *SymbolTable, annotations map[uint32]types.Type, diags *[]Diagnostic) types.Type {
	if node == nil {
		return types.Unknown
	}

	// Special-case: flow narrowing for conditional and loop statements.
	// These need condition-first walking with scoped type overrides,
	// which the generic post-order loop cannot provide.
	switch node.Kind() {
	case "conditional_statement":
		return walkConditionalStatement(node, source, st, annotations, diags)
	case "loop_statement":
		return walkLoopStatement(node, source, st, annotations, diags)
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

	case "string_literal", "interpolated_string_literal":
		return types.Str

	// --- Variables ---
	// If the symbol table has a narrowed type for the variable, use that;
	// otherwise fall back to the sigil type.

	case "scalar":
		return lookupNarrowedType(node, source, st, "$", types.Scalar)

	case "array":
		return lookupNarrowedType(node, source, st, "@", types.Array)

	case "hash":
		return lookupNarrowedType(node, source, st, "%", types.Hash)

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

	// --- Assignments ---
	// Narrow the LHS variable type based on the RHS expression type.

	case "assignment_expression":
		return inferAssignmentNarrowing(node, source, st, annotations, childTypes)

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
		strconv.Itoa(min) + " argument(s), got " + strconv.Itoa(got)
}

// typeMismatchMessage produces a human-readable type mismatch message.
func typeMismatchMessage(name string, argIdx int, expected, actual types.Type) string {
	return "call to " + name + ": argument " + strconv.Itoa(argIdx+1) +
		" expects " + expected.String() + ", got " + actual.String()
}

// lookupNarrowedType checks the symbol table for a narrowed type for the
// variable represented by a scalar/array/hash node.  If a symbol is found
// and its type is not Unknown, the refined type is returned; otherwise the
// fallback sigil type is returned.
func lookupNarrowedType(node *parser.Node, source []byte, st *SymbolTable, sigil string, fallback types.Type) types.Type {
	name := sigildName(sigil, node, source)
	if sym, ok := st.Lookup(name); ok && sym.Type != types.Unknown {
		return sym.Type
	}
	return fallback
}

// inferAssignmentNarrowing handles assignment_expression nodes. It extracts
// the LHS variable name and the RHS type, then updates the symbol table to
// narrow the variable's type based on the assigned value.
//
// Tree structure:
//
//	assignment_expression
//	  variable_declaration (or scalar/array/hash for plain assignments)
//	  = (anonymous)
//	  <rhs expression>
func inferAssignmentNarrowing(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	childTypes []types.Type,
) types.Type {
	// Find the LHS variable name and the RHS type.
	var varName string
	var rhsType types.Type

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		ck := child.Kind()

		switch ck {
		case "variable_declaration":
			// my $x = ... — extract the sigil-prefixed name from the child
			varName = extractVarNameFromDecl(child, source)

		case "scalar":
			// $x = ... (plain assignment, LHS only)
			if varName == "" {
				varName = sigildName("$", child, source)
			}
		case "array":
			if varName == "" {
				varName = sigildName("@", child, source)
			}
		case "hash":
			if varName == "" {
				varName = sigildName("%", child, source)
			}
		}
	}

	// The RHS is the last named child (after the = operator).
	// Its type is the corresponding entry in childTypes.
	for i := node.ChildCount() - 1; i >= 0; i-- {
		child := node.Child(i)
		if child != nil && child.IsNamed() {
			if i < len(childTypes) {
				rhsType = childTypes[i]
			}
			break
		}
	}

	if varName != "" && rhsType != types.Unknown {
		st.UpdateType(varName, rhsType)
	}

	return rhsType
}

// guardResult holds the result of extracting a guard pattern from a condition node.
type guardResult struct {
	VarName string
	Guard   types.GuardPattern
	Negated bool
}

// extractGuardPattern examines a condition expression node and returns the
// guard pattern if the condition matches a recognized form. Returns nil if
// no guard pattern is recognized.
//
// Recognized CST shapes:
//
//	func1op_call_expression with keyword "defined" and scalar child → GuardDefined
//	func1op_call_expression with keyword "ref" and scalar child → GuardRef
//	relational_expression with "isa" operator, scalar LHS, bareword RHS → GuardIsa
func extractGuardPattern(node *parser.Node, source []byte) *guardResult {
	if node == nil {
		return nil
	}

	kind := node.Kind()

	// Pattern: defined($x) or ref($x)
	if kind == "func1op_call_expression" {
		return extractFunc1opGuard(node, source)
	}

	// Pattern: $x isa Foo
	if kind == "relational_expression" {
		return extractIsaGuard(node, source)
	}

	// Pattern: !guard (unary negation)
	if kind == "unary_expression" {
		return extractNegatedGuard(node, source)
	}

	// Pattern: not guard (low-precedence negation)
	if kind == "ambiguous_function_call_expression" {
		return extractNotGuard(node, source)
	}

	return nil
}

// extractIsaGuard extracts a guard from a relational_expression node of the
// form "$x isa Foo". The CST has children: scalar (named), "isa" (anonymous),
// bareword (named).
func extractIsaGuard(node *parser.Node, source []byte) *guardResult {
	var varNode *parser.Node
	hasIsa := false

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			if child.Text(source) == "isa" {
				hasIsa = true
			}
			continue
		}
		if child.Kind() == "scalar" && varNode == nil {
			varNode = child
		}
	}

	if !hasIsa || varNode == nil {
		return nil
	}

	varName := sigildName("$", varNode, source)
	return &guardResult{VarName: varName, Guard: types.GuardPattern{Kind: types.GuardIsa}}
}

// extractNegatedGuard unwraps a unary_expression with "!" to find the inner
// guard pattern. CST: unary_expression -> "!" (anon) + inner expression (named).
func extractNegatedGuard(node *parser.Node, source []byte) *guardResult {
	hasNot := false
	var innerNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			if child.Text(source) == "!" {
				hasNot = true
			}
			continue
		}
		if innerNode == nil {
			innerNode = child
		}
	}

	if !hasNot || innerNode == nil {
		return nil
	}

	result := extractGuardPattern(innerNode, source)
	if result != nil {
		result.Negated = !result.Negated
	}
	return result
}

// extractNotGuard unwraps an ambiguous_function_call_expression with
// function "not" to find the inner guard pattern.
// CST: ambiguous_function_call_expression -> function:"not" + inner expression (named).
func extractNotGuard(node *parser.Node, source []byte) *guardResult {
	hasNot := false
	var innerNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "function" && child.Text(source) == "not" {
			hasNot = true
			continue
		}
		if child.IsNamed() && innerNode == nil {
			innerNode = child
		}
	}

	if !hasNot || innerNode == nil {
		return nil
	}

	result := extractGuardPattern(innerNode, source)
	if result != nil {
		result.Negated = !result.Negated
	}
	return result
}

// extractFunc1opGuard extracts a guard from a func1op_call_expression node.
// It looks for the function keyword (defined, ref) and a scalar argument.
func extractFunc1opGuard(node *parser.Node, source []byte) *guardResult {
	var funcName string
	var varNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			text := child.Text(source)
			if text != "(" && text != ")" {
				funcName = text
			}
			continue
		}
		if child.Kind() == "scalar" {
			varNode = child
		}
	}

	if varNode == nil {
		return nil
	}

	varName := sigildName("$", varNode, source)

	switch funcName {
	case "defined":
		return &guardResult{VarName: varName, Guard: types.GuardPattern{Kind: types.GuardDefined}}
	case "ref":
		return &guardResult{VarName: varName, Guard: types.GuardPattern{Kind: types.GuardRef}}
	}

	return nil
}

// walkConditionalStatement handles if/unless statements with guard-based flow
// narrowing. It walks the condition first, extracts a guard pattern, then walks
// the if-body and else-body with appropriate scoped type overrides.
//
// For "if", the if-body gets the positive guard narrowing and the else-body
// gets the negated guard. For "unless", the narrowing is flipped.
func walkConditionalStatement(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	var keyword string
	var conditionNode *parser.Node
	var ifBlock *parser.Node
	var elseNode *parser.Node

	// Identify children: keyword, condition, block, and optional else.
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			text := child.Text(source)
			if text == "if" || text == "unless" {
				keyword = text
			}
			continue
		}
		switch child.Kind() {
		case "block":
			if ifBlock == nil {
				ifBlock = child
			}
		case "else":
			elseNode = child
		case "elsif":
			walkElsifNode(child, source, st, annotations, diags)
		default:
			// The condition is the only named non-block, non-else child.
			// Enclosing parentheses are anonymous nodes in the CST, so the
			// condition expression (e.g. func1op_call_expression) appears
			// directly as a named child of conditional_statement.
			if conditionNode == nil {
				conditionNode = child
			}
		}
	}

	// Walk the condition node to type its children.
	if conditionNode != nil {
		walkNode(conditionNode, source, st, annotations, diags)
	}

	// Extract guard pattern from the condition.
	guard := extractGuardPattern(conditionNode, source)

	// Compute the negate flag for the if-body. "unless" flips it,
	// and a negated condition (e.g. !defined) flips it again.
	ifBodyNegate := keyword == "unless"
	if guard != nil && guard.Negated {
		ifBodyNegate = !ifBodyNegate
	}

	// Walk the if/unless body with appropriate narrowing.
	if ifBlock != nil {
		walkBlockWithGuard(ifBlock, source, st, annotations, diags, guard, ifBodyNegate)
	}

	// Walk the else body with the opposite narrowing.
	if elseNode != nil {
		var elseBlock *parser.Node
		for i := 0; i < elseNode.ChildCount(); i++ {
			child := elseNode.Child(i)
			if child != nil && child.Kind() == "block" {
				elseBlock = child
				break
			}
		}
		if elseBlock != nil {
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, guard, !ifBodyNegate)
		}
	}

	return types.Unknown
}

// walkElsifNode handles a single elsif node with guard-based flow narrowing.
// It mirrors walkConditionalStatement: walk condition, extract guard, walk
// block with guard, then handle the trailing else or elsif recursively.
func walkElsifNode(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) {
	var conditionNode *parser.Node
	var block *parser.Node
	var elseNode *parser.Node
	var elsifNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			continue
		}
		switch child.Kind() {
		case "block":
			if block == nil {
				block = child
			}
		case "else":
			elseNode = child
		case "elsif":
			elsifNode = child
		default:
			if conditionNode == nil {
				conditionNode = child
			}
		}
	}

	// Walk the condition to type its children.
	if conditionNode != nil {
		walkNode(conditionNode, source, st, annotations, diags)
	}

	guard := extractGuardPattern(conditionNode, source)

	// Walk the elsif body with the guard.
	if block != nil {
		walkBlockWithGuard(block, source, st, annotations, diags, guard, false)
	}

	// Handle trailing else or elsif.
	if elseNode != nil {
		var elseBlock *parser.Node
		for i := 0; i < elseNode.ChildCount(); i++ {
			child := elseNode.Child(i)
			if child != nil && child.Kind() == "block" {
				elseBlock = child
				break
			}
		}
		if elseBlock != nil {
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, guard, true)
		}
	}

	if elsifNode != nil {
		walkElsifNode(elsifNode, source, st, annotations, diags)
	}
}

// walkLoopStatement handles while statements with guard-based flow narrowing.
func walkLoopStatement(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	var conditionNode *parser.Node
	var bodyBlock *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			continue
		}
		switch child.Kind() {
		case "block":
			bodyBlock = child
		default:
			if conditionNode == nil {
				conditionNode = child
			}
		}
	}

	if conditionNode != nil {
		walkNode(conditionNode, source, st, annotations, diags)
	}

	guard := extractGuardPattern(conditionNode, source)

	if bodyBlock != nil {
		walkBlockWithGuard(bodyBlock, source, st, annotations, diags, guard, false)
	}

	return types.Unknown
}

// walkBlockWithGuard walks a block node with an optional guard-based type
// override. If guard is non-nil, it enters a new "guard" scope, shadows the
// guarded variable with the narrowed type, walks the block children, then exits
// the scope. If negate is true, the negated guard type is applied instead.
//
// Note: inner "my" declarations inside the block were processed by pass 1
// (CollectDeclarations) in block scopes that no longer exist by pass 2.
// UpdateType calls for those inner variables will be no-ops. This is a known
// limitation — guard narrowing applies to the guarded variable, not to new
// declarations inside the block.
func walkBlockWithGuard(
	block *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	guard *guardResult,
	negate bool,
) {
	if guard != nil {
		// Look up the variable's current type for narrowing.
		// If the variable was never declared, skip narrowing entirely
		// to avoid creating phantom shadow entries in the guard scope.
		sym, found := st.Lookup(guard.VarName)
		if !found {
			for i := 0; i < block.ChildCount(); i++ {
				walkNode(block.Child(i), source, st, annotations, diags)
			}
			return
		}
		currentType := sym.Type

		var narrowedType types.Type
		var narrowed bool
		if negate {
			narrowedType, narrowed = types.NegateGuard(currentType, guard.Guard)
		} else {
			narrowedType, narrowed = types.NarrowByGuard(currentType, guard.Guard)
		}

		if narrowed {
			st.EnterScope("guard")
			defer st.ExitScope()
			st.Define(Symbol{
				Name: guard.VarName,
				Type: narrowedType,
				Kind: SymVariable,
			})
			for i := 0; i < block.ChildCount(); i++ {
				walkNode(block.Child(i), source, st, annotations, diags)
			}
			return
		}
	}

	// No guard or no narrowing: walk normally.
	for i := 0; i < block.ChildCount(); i++ {
		walkNode(block.Child(i), source, st, annotations, diags)
	}
}

// extractVarNameFromDecl pulls the sigil-prefixed variable name from a
// variable_declaration node (e.g. "my $x" -> "$x").
func extractVarNameFromDecl(node *parser.Node, source []byte) string {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "scalar":
			return sigildName("$", child, source)
		case "array":
			return sigildName("@", child, source)
		case "hash":
			return sigildName("%", child, source)
		}
	}
	return ""
}
