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
	case "for_statement":
		return walkForStatement(node, source, st, annotations, diags)
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

// compoundGuard holds the structure of a compound boolean guard expression.
// Op is "&&" for AND conditions and "||" for OR conditions.
// Left and Right are the sub-guards, each of which may itself be compound.
type compoundGuard struct {
	Op    string // "&&" or "||"
	Left  *guardResult
	Right *guardResult
}

// guardResult holds the result of extracting a guard pattern from a condition node.
// When Compound is non-nil, this is a compound guard and VarName/Guard/Negated are unused.
type guardResult struct {
	VarName  string
	Guard    types.GuardPattern
	Negated  bool
	Compound *compoundGuard
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
//	binary_expression with "&&" or "||" between two guards → compound guard
//	lowprec_logical_expression with "and" or "or" between two guards → compound guard
//	unary_expression with "!" wrapping a recognized guard → Negated guard
//	ambiguous_function_call_expression with "not" wrapping a recognized guard → Negated guard
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

	// Pattern: guard1 && guard2 or guard1 || guard2 (high-precedence binary)
	// Pattern: guard1 and guard2 or guard1 or guard2 (low-precedence logical)
	// These are checked before unary_expression so that negated sub-guards
	// inside compound conditions are handled by recursive calls to extractGuardPattern.
	if kind == "binary_expression" || kind == "lowprec_logical_expression" {
		return extractCompoundGuard(node, source)
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

// extractCompoundGuard extracts a compound guard from a binary_expression or
// lowprec_logical_expression node. It finds the boolean operator token (&&, ||,
// and, or) and recursively extracts guards from the left and right sub-expressions.
//
// Operator normalization: "and" → "&&", "or" → "||".
//
// If both sides yield guards, a compound guardResult is returned.
// If only one side yields a guard, that single guard is returned directly
// (partial compound — the non-guard side is dropped).
// If neither side yields a guard, nil is returned.
func extractCompoundGuard(node *parser.Node, source []byte) *guardResult {
	var op string
	var leftNode *parser.Node
	var rightNode *parser.Node

	// Scan children: anonymous nodes hold the operator token;
	// named nodes are the left and right sub-expressions.
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if !child.IsNamed() {
			// Operator token — capture only the boolean operators we recognize.
			text := child.Text(source)
			switch text {
			case "&&", "||":
				op = text
			case "and":
				op = "&&"
			case "or":
				op = "||"
			}
			continue
		}
		// Named children are sub-expressions: first is left, second is right.
		if leftNode == nil {
			leftNode = child
		} else if rightNode == nil {
			rightNode = child
		}
	}

	// If we didn't find a recognized boolean operator, this is not a compound guard
	// (e.g. it could be an arithmetic binary_expression like 1 + 2).
	if op == "" {
		return nil
	}

	leftGuard := extractGuardPattern(leftNode, source)
	rightGuard := extractGuardPattern(rightNode, source)

	// Both sides are guards: return a compound guard.
	if leftGuard != nil && rightGuard != nil {
		return &guardResult{
			Compound: &compoundGuard{
				Op:    op,
				Left:  leftGuard,
				Right: rightGuard,
			},
		}
	}

	// Only one side is a guard: return it directly (partial compound).
	if leftGuard != nil {
		return leftGuard
	}
	return rightGuard
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
//
// For simple guards, the Negated flag is toggled.
// For compound guards, De Morgan's law is applied:
//   - !(A && B) becomes (!A || !B): flip Op from && to ||, negate both sub-guards
//   - !(A || B) becomes (!A && !B): flip Op from || to &&, negate both sub-guards
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
	if result == nil {
		return nil
	}

	// Apply De Morgan's law for compound guards, or toggle Negated for simple guards.
	if result.Compound != nil {
		flippedOp := "||"
		if result.Compound.Op == "||" {
			flippedOp = "&&"
		}
		// Negate each sub-guard (shallow copy and toggle Negated).
		negatedLeft := *result.Compound.Left
		negatedLeft.Negated = !negatedLeft.Negated
		negatedRight := *result.Compound.Right
		negatedRight.Negated = !negatedRight.Negated
		return &guardResult{
			Compound: &compoundGuard{
				Op:    flippedOp,
				Left:  &negatedLeft,
				Right: &negatedRight,
			},
		}
	}

	result.Negated = !result.Negated
	return result
}

// extractNotGuard unwraps an ambiguous_function_call_expression with
// function "not" to find the inner guard pattern.
// CST: ambiguous_function_call_expression -> function:"not" + inner expression (named).
//
// For simple guards, the Negated flag is toggled.
// For compound guards, De Morgan's law is applied (same as extractNegatedGuard).
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
	if result == nil {
		return nil
	}

	// Apply De Morgan's law for compound guards, or toggle Negated for simple guards.
	if result.Compound != nil {
		flippedOp := "||"
		if result.Compound.Op == "||" {
			flippedOp = "&&"
		}
		negatedLeft := *result.Compound.Left
		negatedLeft.Negated = !negatedLeft.Negated
		negatedRight := *result.Compound.Right
		negatedRight.Negated = !negatedRight.Negated
		return &guardResult{
			Compound: &compoundGuard{
				Op:    flippedOp,
				Left:  &negatedLeft,
				Right: &negatedRight,
			},
		}
	}

	result.Negated = !result.Negated
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
	hasNoElsif := true

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
			hasNoElsif = false
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

	// Compute the negate flag for the if-body. "unless" flips the narrowing
	// direction relative to the condition. Negated conditions (e.g. !defined)
	// are handled inside flattenGuards via XOR with each leaf's Negated flag,
	// so we do not flip ifBodyNegate here for guard.Negated.
	ifBodyNegate := keyword == "unless"

	// Walk the if/unless body with appropriate narrowing.
	if ifBlock != nil {
		walkBlockWithGuard(ifBlock, source, st, annotations, diags, flattenGuards(guard, ifBodyNegate))
	}

	// Early-exit narrowing: if the if-body always exits and there is no
	// else/elsif, the continuation (code after the if) knows the if-condition
	// was false. This is identical to the else-branch narrowing, so we apply
	// the same guard list that the else-body would have received.
	//
	// Note: if the if-body contains assignments before the exit (e.g.
	// $x = 1; return;), the assignment narrowing may have already mutated
	// the symbol table entry. The post-exit narrowing uses the current
	// symbol table type, so assignment narrowing takes precedence.
	//
	// Early-exit narrowing only works for simple (leaf) guards — compound
	// guards have VarName == "" which causes st.Lookup to return found=false.
	if ifBlock != nil && guard != nil && guard.VarName != "" && elseNode == nil && hasNoElsif &&
		blockAlwaysExits(ifBlock, source) {
		// The continuation gets the else-branch guard: flattenGuards(guard, !ifBodyNegate).
		// For a simple guard, this is one leaf with Negated = guard.Negated XOR !ifBodyNegate.
		elseGuards := flattenGuards(guard, !ifBodyNegate)
		for _, eg := range elseGuards {
			sym, found := st.Lookup(eg.VarName)
			if !found {
				continue
			}
			var elseType types.Type
			var narrowed bool
			if eg.Negated {
				elseType, narrowed = types.NegateGuard(sym.Type, eg.Guard)
			} else {
				elseType, narrowed = types.NarrowByGuard(sym.Type, eg.Guard)
			}
			if narrowed {
				st.UpdateType(eg.VarName, elseType)
			}
		}
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
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, flattenGuards(guard, !ifBodyNegate))
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

	// The elsif body negate flag starts false (no "unless" keyword in elsif).
	// Negated conditions (e.g. !defined) are handled inside flattenGuards
	// via XOR with each leaf's Negated flag, so we do not flip here.
	elsifBodyNegate := false

	// Walk the elsif body with the guard.
	if block != nil {
		walkBlockWithGuard(block, source, st, annotations, diags, flattenGuards(guard, elsifBodyNegate))
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
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, flattenGuards(guard, !elsifBodyNegate))
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
		walkBlockWithGuard(bodyBlock, source, st, annotations, diags, flattenGuards(guard, false))
	}

	return types.Unknown
}

// walkForStatement handles for/foreach loops with proper variable scoping.
// The loop variable is defined as Scalar in a dedicated scope that spans
// the loop body, preventing it from leaking into the enclosing scope.
//
// Note: inner "my" declarations inside the for-body have the same pass-1/pass-2
// scope limitation as walkBlockWithGuard — UpdateType for those inner variables
// will be no-ops.
func walkForStatement(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
) types.Type {
	var loopVar *parser.Node
	var iterSource *parser.Node
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
		case "scalar":
			if loopVar == nil {
				loopVar = child
			}
		case "block":
			bodyBlock = child
		default:
			// The iteration source is the other named child (array or list_expression).
			if iterSource == nil {
				iterSource = child
			}
		}
	}

	// Walk the iteration source to type it.
	if iterSource != nil {
		walkNode(iterSource, source, st, annotations, diags)
	}

	// Walk the loop body with the loop variable scoped.
	if bodyBlock != nil && loopVar != nil {
		varName := sigildName("$", loopVar, source)
		st.EnterScope("for")
		defer st.ExitScope()
		st.Define(Symbol{
			Name: varName,
			Type: types.Scalar,
			Kind: SymVariable,
		})
		// Annotate the loop variable node itself.
		annotations[loopVar.StartByte()] = types.Scalar
		for i := 0; i < bodyBlock.ChildCount(); i++ {
			walkNode(bodyBlock.Child(i), source, st, annotations, diags)
		}
		return types.Unknown
	}

	// Fallback: walk all children generically.
	for i := 0; i < node.ChildCount(); i++ {
		walkNode(node.Child(i), source, st, annotations, diags)
	}
	return types.Unknown
}

// blockAlwaysExits returns true if the block contains a top-level statement
// that unconditionally exits the current scope (return, die, exit).
// Only checks direct children of the block — nested exits inside inner
// conditions are ignored (conservative to avoid false positives).
func blockAlwaysExits(block *parser.Node, source []byte) bool {
	if block == nil {
		return false
	}
	for i := 0; i < block.ChildCount(); i++ {
		child := block.Child(i)
		if child == nil {
			continue
		}
		// Check inside expression_statement wrappers.
		target := child
		if child.Kind() == "expression_statement" && child.ChildCount() > 0 {
			target = child.Child(0)
			if target == nil {
				continue
			}
		}

		switch target.Kind() {
		case "return_expression":
			return true
		case "func1op_call_expression":
			// exit and exit N
			for j := 0; j < target.ChildCount(); j++ {
				c := target.Child(j)
				if c != nil && !c.IsNamed() && c.Text(source) == "exit" {
					return true
				}
			}
		case "ambiguous_function_call_expression", "function_call_expression":
			// die $msg, die($obj), and similar forms where die is the
			// function name. The string-argument form die("msg") produces
			// broken CST due to the gotreesitter string grammar limitation,
			// but die $var is handled here.
			for j := 0; j < target.ChildCount(); j++ {
				c := target.Child(j)
				if c != nil && c.Kind() == "function" && c.Text(source) == "die" {
					return true
				}
			}
		case "bareword":
			if target.Text(source) == "die" {
				return true
			}
		}
	}
	return false
}

// flattenGuards converts a guardResult (possibly compound) into a flat list
// of leaf guards with the Negated flag resolved per compound semantics.
//
// The negate parameter represents an outer negation to XOR into each leaf's
// own Negated flag:
//
//   - For && (negate=false): both sub-guards apply → recurse into each.
//   - For || (negate=false): neither can be guaranteed → return nil.
//   - For && (negate=true, De Morgan → ||): no narrowing → return nil.
//   - For || (negate=true, De Morgan → &&): both negated sub-guards apply → recurse.
//
// For simple (non-compound) guards: return a one-element slice with Negated
// set to guard.Negated XOR negate.
func flattenGuards(guard *guardResult, negate bool) []*guardResult {
	if guard == nil {
		return nil
	}

	if guard.Compound == nil {
		// Leaf guard: resolve the final Negated flag and return it.
		leaf := &guardResult{
			VarName: guard.VarName,
			Guard:   guard.Guard,
			Negated: guard.Negated != negate, // XOR: outer negate flips the leaf
		}
		return []*guardResult{leaf}
	}

	// Compound guard: determine whether this compound's semantics yield all-leaf
	// or no-leaf based on the operator and the outer negate.
	//
	// Effective operator after applying De Morgan for the outer negation:
	//   negate=false: keep Op as-is.
	//   negate=true:  flip Op (De Morgan: !(A&&B) = !A||!B, !(A||B) = !A&&!B).
	effectiveOp := guard.Compound.Op
	if negate {
		if effectiveOp == "&&" {
			effectiveOp = "||"
		} else {
			effectiveOp = "&&"
		}
	}

	switch effectiveOp {
	case "&&":
		// Both sub-guards apply — flatten each with the (possibly flipped) negate.
		left := flattenGuards(guard.Compound.Left, negate)
		right := flattenGuards(guard.Compound.Right, negate)
		return append(left, right...)
	case "||":
		// Either sub-guard could be the true one — no narrowing is certain.
		return nil
	}

	return nil
}

// walkBlockWithGuard walks a block node with an optional list of leaf guards
// for flow-narrowing. If guards is non-nil and non-empty, it enters a single
// "guard" scope, shadows each guarded variable with its narrowed type, walks
// the block children, then exits the scope.
//
// Each guard's Negated flag is already fully resolved by flattenGuards: when
// Negated is true, NegateGuard is applied; otherwise NarrowByGuard is applied.
//
// Guards that share the same variable are applied sequentially (the output
// type of one guard becomes the input type for the next), so that compound
// conditions on the same variable (e.g. defined($x) && ref($x)) correctly
// intersect their effects.
//
// If a variable referenced by a guard was never declared, that guard is
// silently skipped (no phantom shadow entries in the scope).
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
	guards []*guardResult,
) {
	// Build accumulated narrowed types by applying guards sequentially.
	// working tracks the current type for each variable as guards are applied,
	// whether or not each individual guard produced a change.
	// shadows collects only variables where at least one guard narrowed.
	working := make(map[string]types.Type)
	shadows := make(map[string]types.Type)

	for _, guard := range guards {
		// Determine the starting type for this variable: use the working type
		// from prior guards, or look up the declared type for the first guard.
		currentType, seen := working[guard.VarName]
		if !seen {
			sym, found := st.Lookup(guard.VarName)
			if !found {
				// Variable was never declared — skip to avoid phantom shadows.
				continue
			}
			currentType = sym.Type
		}

		var narrowedType types.Type
		var didNarrow bool
		if guard.Negated {
			narrowedType, didNarrow = types.NegateGuard(currentType, guard.Guard)
		} else {
			narrowedType, didNarrow = types.NarrowByGuard(currentType, guard.Guard)
		}

		if didNarrow {
			working[guard.VarName] = narrowedType
			shadows[guard.VarName] = narrowedType
		} else {
			// Guard did not change the type, but record the working type so
			// subsequent guards for the same variable chain from it.
			working[guard.VarName] = currentType
		}
	}

	if len(shadows) > 0 {
		st.EnterScope("guard")
		defer st.ExitScope()
		for name, typ := range shadows {
			st.Define(Symbol{
				Name: name,
				Type: typ,
				Kind: SymVariable,
			})
		}
	}

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
