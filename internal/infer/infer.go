// ABOUTME: Pass 2 of the PSC type inference engine — the inference walker.
// ABOUTME: Analyze walks the CST bottom-up, annotating every node with a types.Type.

package infer

import (
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// walkUseStatement handles use_statement CST nodes. When a ProjectIndex is
// provided, it resolves the module name to a file path and triggers analysis of
// that file so its symbols are available for cross-file lookup.
//
// If idx is nil (single-file mode), or if the module cannot be resolved, this
// is a no-op. The function always returns types.Unknown because a use statement
// is not an expression.
func walkUseStatement(node *parser.Node, source []byte, idx *ProjectIndex) types.Type {
	if idx == nil {
		return types.Unknown
	}

	// The package name is held in the "package" named child of the use_statement.
	var modName string
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "package" {
			modName = child.Text(source)
			break
		}
	}

	if modName == "" {
		return types.Unknown
	}

	path, err := idx.ResolveModule(modName)
	if err != nil {
		// Module not found in lib dirs — skip silently.
		return types.Unknown
	}

	// Trigger analysis (result is cached on subsequent calls).
	_, _ = idx.AnalyzeFile(path)
	return types.Unknown
}

// Analyze runs both inference passes over the parsed tree and source and
// returns an annotation map and a slice of diagnostics.
//
// The annotation map keys are node StartByte values; each value is the
// inferred types.Type for the node starting at that byte offset.
// Diagnostics describe any type errors or arity mismatches found.
//
// idx is an optional ProjectIndex for cross-file analysis. Pass nil for
// single-file mode; use statements and fully-qualified calls are then ignored.
func Analyze(tree *parser.Tree, source []byte, idx *ProjectIndex) (map[uint32]types.Type, []Diagnostic, *SymbolTable) {
	annotations := make(map[uint32]types.Type)
	diags := make([]Diagnostic, 0)
	// classTypes maps a node's StartByte to the class name of the object
	// produced at that position (e.g. from Foo->new()). It is written by
	// inferMethodCallType and read by inferAssignmentNarrowing.
	classTypes := make(map[uint32]string)

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
	walkNode(root, source, st, annotations, &diags, idx, classTypes)

	return annotations, diags, st
}

// walkNode performs a post-order (bottom-up) traversal of the CST, computing
// a types.Type for every node that has a meaningful type and storing it in the
// annotations map keyed by the node's StartByte.
//
// idx is threaded through for cross-file analysis; it may be nil in single-file mode.
// classTypes maps node StartByte values to class names for constructor call results.
func walkNode(node *parser.Node, source []byte, st *SymbolTable, annotations map[uint32]types.Type, diags *[]Diagnostic, idx *ProjectIndex, classTypes map[uint32]string) types.Type {
	if node == nil {
		return types.Unknown
	}

	// Special-case: flow narrowing for conditional and loop statements.
	// These need condition-first walking with scoped type overrides,
	// which the generic post-order loop cannot provide.
	switch node.Kind() {
	case "conditional_statement":
		return walkConditionalStatement(node, source, st, annotations, diags, idx, classTypes)
	case "loop_statement":
		return walkLoopStatement(node, source, st, annotations, diags, idx, classTypes)
	case "for_statement":
		return walkForStatement(node, source, st, annotations, diags, idx, classTypes)
	case "subroutine_declaration_statement":
		return walkSubroutineDeclaration(node, source, st, annotations, diags, idx, classTypes)
	case "use_statement":
		return walkUseStatement(node, source, idx)
	}

	// Recurse into all children first (post-order).
	childTypes := make([]types.Type, node.ChildCount())
	for i := 0; i < node.ChildCount(); i++ {
		childTypes[i] = walkNode(node.Child(i), source, st, annotations, diags, idx, classTypes)
	}

	typ := inferNodeType(node, source, st, annotations, childTypes, diags, idx, classTypes)
	if typ != types.Unknown {
		annotations[node.StartByte()] = typ
	}
	return typ
}

// inferNodeType dispatches on node.Kind() and returns the inferred type for
// the node, emitting any diagnostics as a side effect.
//
// idx is forwarded to inferFunctionCallType for cross-file FQ call resolution;
// it may be nil in single-file mode.
// classTypes maps node StartByte values to class names for constructor call results.
func inferNodeType(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	childTypes []types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
	classTypes map[uint32]string,
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
		return inferFunctionCallType(node, source, st, annotations, diags, idx)

	case "func1op_call_expression":
		return inferFunc1opCallType(node, source, annotations, diags)

	case "func0op_call_expression":
		return inferFunc0opCallType(node, source, annotations, diags)

	// --- Method calls ---

	case "method_call_expression":
		return inferMethodCallType(node, source, st, idx, classTypes)

	// --- Assignments ---
	// Narrow the LHS variable type based on the RHS expression type.

	case "assignment_expression":
		return inferAssignmentNarrowing(node, source, st, annotations, childTypes, classTypes)

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
//
// When the name contains "::" and a ProjectIndex is available, the function
// splits the name into package and sub-name components and performs a
// cross-file symbol lookup via the index.
func inferFunctionCallType(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
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

	// Check builtins first (authoritative — prevents user subs from
	// shadowing builtin return types).
	sig, ok := types.GetBuiltin(name)
	if !ok {
		// Fully-qualified call (e.g. Foo::Bar::baz) — resolve via index.
		if idx != nil && strings.Contains(name, "::") {
			sep := strings.LastIndex(name, "::")
			pkg := name[:sep]
			funcName := name[sep+2:]
			if sym, found := idx.LookupSymbol(pkg, funcName); found && sym.ReturnType != types.Unknown {
				return sym.ReturnType
			}
		}
		// Not a builtin — check symbol table for user-defined sub with
		// an inferred return type.
		if sym, found := st.Lookup(name); found && sym.ReturnType != types.Unknown {
			return sym.ReturnType
		}
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

// inferMethodCallType handles method_call_expression nodes such as Foo->new()
// or $obj->method(). It resolves the method via the ProjectIndex when possible.
//
// CST shape: method_call_expression has an invocant (scalar or bareword), an
// anonymous "->" token, and a method name (bareword or named).
//
//   - Bareword invocant + method "new" → returns types.Object, writes className to classTypes.
//   - Bareword invocant + other method → looks up className::method in the index.
//   - Scalar invocant → looks up the variable's ClassType from the symbol table.
//     If set and idx is available → looks up classType::method in the index.
//   - Fallback → types.Any.
func inferMethodCallType(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	idx *ProjectIndex,
	classTypes map[uint32]string,
) types.Type {
	var invocantNode *parser.Node
	var methodName string

	// Walk children: invocant is the first named child, method name follows "->".
	// The grammar places the method name as the last named child.
	namedChildren := make([]*parser.Node, 0, node.ChildCount())
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil && child.IsNamed() {
			namedChildren = append(namedChildren, child)
		}
	}

	if len(namedChildren) >= 2 {
		invocantNode = namedChildren[0]
		lastChild := namedChildren[len(namedChildren)-1]
		methodName = lastChild.Text(source)
	} else if len(namedChildren) == 1 {
		// Only an invocant but no method name found yet — check anonymous children.
		invocantNode = namedChildren[0]
	}

	// If methodName is still empty, look for bareword in last position among all children.
	if methodName == "" {
		for i := node.ChildCount() - 1; i >= 0; i-- {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if child.Kind() == "bareword" || child.Kind() == "method" {
				methodName = child.Text(source)
				break
			}
		}
	}

	if invocantNode == nil || methodName == "" {
		return types.Any
	}

	invocantKind := invocantNode.Kind()

	switch invocantKind {
	case "bareword":
		// Class method call: ClassName->method(...)
		className := invocantNode.Text(source)
		if methodName == "new" {
			// Constructor call: record the class name for the calling node.
			classTypes[node.StartByte()] = className
			return types.Object
		}
		// Class method (non-constructor): resolve via index.
		if idx != nil {
			if sym, found := idx.LookupSymbol(className, methodName); found && sym.ReturnType != types.Unknown {
				return sym.ReturnType
			}
		}
		return types.Any

	case "scalar":
		// Instance method call: $obj->method(...)
		varName := sigildName("$", invocantNode, source)
		sym, found := st.Lookup(varName)
		if !found || sym.ClassType == "" {
			return types.Any
		}
		if idx != nil {
			if methodSym, ok := idx.LookupSymbol(sym.ClassType, methodName); ok && methodSym.ReturnType != types.Unknown {
				return methodSym.ReturnType
			}
		}
		return types.Any
	}

	return types.Any
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
// When the RHS is a method_call_expression that produced a class name in
// classTypes (e.g. Foo->new()), the ClassType is propagated to the LHS symbol.
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
	classTypes map[uint32]string,
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
	var rhsNode *parser.Node
	for i := node.ChildCount() - 1; i >= 0; i-- {
		child := node.Child(i)
		if child != nil && child.IsNamed() {
			if i < len(childTypes) {
				rhsType = childTypes[i]
			}
			rhsNode = child
			break
		}
	}

	if varName != "" && rhsType != types.Unknown {
		st.UpdateType(varName, rhsType)
	}

	// Propagate ClassType when the RHS is a constructor call (Foo->new()).
	if varName != "" && rhsNode != nil {
		if className, ok := classTypes[rhsNode.StartByte()]; ok && className != "" {
			st.UpdateClassType(varName, className)
		}
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
// ClassName is set for GuardIsa guards and holds the class name from the "isa" RHS bareword.
type guardResult struct {
	VarName   string
	Guard     types.GuardPattern
	Negated   bool
	ClassName string // Set for GuardIsa guards: the class name on the RHS of "isa"
	Compound  *compoundGuard
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
//	function_call_expression with known guard function name and scalar arg → per table
//	  (builtin::blessed → GuardIsa, builtin::reftype → GuardRef, builtin::is_bool → GuardBool)
//	binary_expression with "&&" or "||" between two guards → compound guard
//	lowprec_logical_expression with "and" or "or" between two guards → compound guard
//	unary_expression with "!" wrapping a recognized guard → Negated guard
//	ambiguous_function_call_expression with "not" wrapping a recognized guard → Negated guard
func extractGuardPattern(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
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

	// Pattern: builtin::blessed($x), builtin::reftype($x), builtin::is_bool($x)
	if kind == "function_call_expression" {
		return extractFunctionCallGuard(node, source, st)
	}

	// Pattern: guard1 && guard2 or guard1 || guard2 (high-precedence binary)
	// Pattern: guard1 and guard2 or guard1 or guard2 (low-precedence logical)
	// These are checked before unary_expression so that negated sub-guards
	// inside compound conditions are handled by recursive calls to extractGuardPattern.
	if kind == "binary_expression" || kind == "lowprec_logical_expression" {
		return extractCompoundGuard(node, source, st)
	}

	// Pattern: !guard (unary negation)
	if kind == "unary_expression" {
		return extractNegatedGuard(node, source, st)
	}

	// Pattern: not guard (low-precedence negation)
	if kind == "ambiguous_function_call_expression" {
		return extractNotGuard(node, source, st)
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
func extractCompoundGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
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

	leftGuard := extractGuardPattern(leftNode, source, st)
	rightGuard := extractGuardPattern(rightNode, source, st)

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
// bareword (named). ClassName is populated from the bareword RHS.
func extractIsaGuard(node *parser.Node, source []byte) *guardResult {
	var varNode *parser.Node
	var classNode *parser.Node
	var varName string
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
		switch child.Kind() {
		case "scalar":
			if varNode == nil {
				varNode = child
			}
		case "array_element_expression":
			if varNode == nil {
				varNode = child
				varName = child.Text(source)
			}
		case "bareword":
			if classNode == nil {
				classNode = child
			}
		}
	}

	if !hasIsa || varNode == nil {
		return nil
	}

	if varName == "" {
		varName = sigildName("$", varNode, source)
	}
	className := ""
	if classNode != nil {
		className = classNode.Text(source)
	}
	return &guardResult{
		VarName:   varName,
		Guard:     types.GuardPattern{Kind: types.GuardIsa},
		ClassName: className,
	}
}

// extractNegatedGuard unwraps a unary_expression with "!" to find the inner
// guard pattern. CST: unary_expression -> "!" (anon) + inner expression (named).
//
// For simple guards, the Negated flag is toggled.
// For compound guards, De Morgan's law is applied:
//   - !(A && B) becomes (!A || !B): flip Op from && to ||, negate both sub-guards
//   - !(A || B) becomes (!A && !B): flip Op from || to &&, negate both sub-guards
func extractNegatedGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
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

	result := extractGuardPattern(innerNode, source, st)
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
func extractNotGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
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

	result := extractGuardPattern(innerNode, source, st)
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

// guardFunctionTable maps known guard function base names to their guard patterns.
// Both bare names and fully-qualified names are recognized — the package prefix
// is stripped before lookup.
var guardFunctionTable = map[string]types.GuardPattern{
	"blessed": {Kind: types.GuardIsa},
	"reftype": {Kind: types.GuardRef},
	"is_bool": {Kind: types.GuardBool},
}

// extractFunctionCallGuard extracts a guard from a function_call_expression node.
// Recognizes known guard functions from builtin:: (and bare imports).
// CST: function_call_expression -> function (named) + scalar (named).
func extractFunctionCallGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
	var funcName string
	var varNode *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "function" {
			funcName = child.Text(source)
			continue
		}
		if child.Kind() == "scalar" && varNode == nil {
			varNode = child
		}
	}

	if funcName == "" || varNode == nil {
		return nil
	}

	// Strip package prefix to get the base name.
	baseName := funcName
	if idx := strings.LastIndex(funcName, "::"); idx >= 0 {
		baseName = funcName[idx+2:]
	}

	guard, ok := guardFunctionTable[baseName]
	if ok {
		varName := sigildName("$", varNode, source)
		return &guardResult{VarName: varName, Guard: guard}
	}

	// Fallback: check if this is a user-defined guard function.
	if st == nil {
		return nil
	}
	return extractUserDefinedGuard(funcName, varNode, source, st)
}

// extractUserDefinedGuard attempts to resolve a function call as a
// user-defined type guard. It looks up the sub in the symbol table,
// finds its AST node, resolves the parameter variable, extracts the
// return expression, and runs extractGuardPattern on it recursively.
// The resulting guard's VarName is remapped from the sub's parameter
// to the call-site argument.
func extractUserDefinedGuard(funcName string, argNode *parser.Node, source []byte, st *SymbolTable) *guardResult {
	sym, ok := st.Lookup(funcName)
	if !ok || sym.Kind != SymSubroutine {
		return nil
	}

	// Find the sub's AST node by walking up from argNode to the root,
	// then searching top-level children.
	root := argNode
	for root.Parent() != nil {
		root = root.Parent()
	}

	subNode := FindSubDeclNode(root, sym.StartByte, sym.EndByte)
	if subNode == nil {
		return nil
	}

	// Resolve the parameter variable.
	paramName, paramOk := ResolveSubParam(subNode, source)
	if !paramOk {
		return nil
	}

	// Find the body block.
	var bodyBlock *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			bodyBlock = child
		}
	}
	if bodyBlock == nil {
		return nil
	}

	// Extract the return expression.
	returnExpr := ExtractSubReturnExpr(bodyBlock, source)
	if returnExpr == nil {
		return nil
	}

	// Recursively extract a guard pattern from the return expression.
	// Pass nil for st to prevent infinite recursion (1-level limit).
	innerGuard := extractGuardPattern(returnExpr, source, nil)
	if innerGuard == nil {
		return nil
	}

	// Remap the guard's variable from the parameter name to the
	// call-site argument.
	callSiteArg := sigildName("$", argNode, source)
	return remapGuardVar(innerGuard, paramName, callSiteArg)
}

// extractFunc1opGuard extracts a guard from a func1op_call_expression node.
// It looks for the function keyword (defined, ref) and a scalar argument.
func extractFunc1opGuard(node *parser.Node, source []byte) *guardResult {
	var funcName string
	var varNode *parser.Node
	var varName string

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
		} else if child.Kind() == "array_element_expression" && varNode == nil {
			// Support $_[0] as a guard argument.
			varNode = child
			varName = child.Text(source)
		}
	}

	if varNode == nil {
		return nil
	}

	// Use extracted text for array_element_expression, sigildName for scalar.
	if varName == "" {
		varName = sigildName("$", varNode, source)
	}

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
//
// After all branches complete, the types of guarded variables are updated to
// the union (bitwise OR) of each branch's end-of-block type. Early-exit
// branches do not contribute to the join. When there is no else branch, the
// implicit else contributes the pre-if type.
func walkConditionalStatement(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
	classTypes map[uint32]string,
) types.Type {
	var keyword string
	var conditionNode *parser.Node
	var ifBlock *parser.Node
	var elseNode *parser.Node
	var elsifNodes []*parser.Node

	// Identify children: keyword, condition, block, and optional else/elsif.
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
			elsifNodes = append(elsifNodes, child)
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
		walkNode(conditionNode, source, st, annotations, diags, idx, classTypes)
	}

	// Extract guard pattern from the condition.
	guard := extractGuardPattern(conditionNode, source, st)

	// Compute the negate flag for the if-body. "unless" flips the narrowing
	// direction relative to the condition. Negated conditions (e.g. !defined)
	// are handled inside flattenGuards via XOR with each leaf's Negated flag,
	// so we do not flip ifBodyNegate here for guard.Negated.
	ifBodyNegate := keyword == "unless"

	// Collect the pre-if types of all guarded variables for the implicit-else
	// contribution and for the final join computation.
	preIfTypes := collectPreIfTypes(guard, st)

	hasNoElsif := len(elsifNodes) == 0
	ifBodyExits := ifBlock != nil && blockAlwaysExits(ifBlock, source)

	// Walk the if/unless body with appropriate narrowing.
	// Capture branch-end types for the join, unless the branch always exits.
	var ifBranchTypes map[string]types.Type
	if ifBlock != nil {
		if !ifBodyExits {
			ifBranchTypes = make(map[string]types.Type)
		}
		walkBlockWithGuard(ifBlock, source, st, annotations, diags, flattenGuards(guard, ifBodyNegate), ifBranchTypes, idx, classTypes)
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
	if ifBodyExits && guard != nil && guard.VarName != "" && elseNode == nil && hasNoElsif {
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
		// Early-exit handles post-if narrowing directly; skip branch merging.
		return types.Unknown
	}

	// Walk elsif branches, collecting their join types.
	var elsifJoinTypes map[string]types.Type
	for _, elsifNode := range elsifNodes {
		elsifTypes := walkElsifNode(elsifNode, source, st, annotations, diags, idx, classTypes)
		elsifJoinTypes = mergeBranchTypes(elsifJoinTypes, elsifTypes)
	}

	// Walk the else body with the opposite narrowing.
	var elseBranchTypes map[string]types.Type
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
			elseBranchTypes = make(map[string]types.Type)
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, flattenGuards(guard, !ifBodyNegate), elseBranchTypes, idx, classTypes)
		}
	}

	// Compute the join type for each guarded variable and apply it.
	// The join is the OR of:
	//   - if-branch end type (if not early-exit)
	//   - elsif chain join types (if any)
	//   - else-branch end type (if present), or pre-if type (implicit else)
	applyBranchJoin(guard, preIfTypes, ifBranchTypes, elsifJoinTypes, elseBranchTypes, elseNode, st)

	return types.Unknown
}

// collectPreIfTypes returns the current types of all variables referenced by
// the guard (leaf guards only). This records the pre-if types for the implicit
// else and for the join computation.
func collectPreIfTypes(guard *guardResult, st *SymbolTable) map[string]types.Type {
	if guard == nil {
		return nil
	}
	result := make(map[string]types.Type)
	collectLeafVarTypes(guard, st, result)
	if len(result) == 0 {
		return nil
	}
	return result
}

// collectLeafVarTypes recursively walks a guardResult tree and collects the
// current symbol-table type for each leaf variable referenced.
func collectLeafVarTypes(guard *guardResult, st *SymbolTable, out map[string]types.Type) {
	if guard == nil {
		return
	}
	if guard.Compound != nil {
		collectLeafVarTypes(guard.Compound.Left, st, out)
		collectLeafVarTypes(guard.Compound.Right, st, out)
		return
	}
	if guard.VarName != "" {
		if _, already := out[guard.VarName]; !already {
			if sym, found := st.Lookup(guard.VarName); found {
				out[guard.VarName] = sym.Type
			}
		}
	}
}

// mergeBranchTypes ORs the types from src into dst, creating dst if nil.
// Variables present only in one map are taken as-is.
func mergeBranchTypes(dst, src map[string]types.Type) map[string]types.Type {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = make(map[string]types.Type)
	}
	for name, typ := range src {
		if existing, ok := dst[name]; ok {
			dst[name] = existing | typ
		} else {
			dst[name] = typ
		}
	}
	return dst
}

// applyBranchJoin computes the post-if join type for each guarded variable and
// applies it to the outer scope via UpdateType.
//
// The join is the OR of all non-early-exit branch-end types. When there is no
// else branch, the implicit else contributes the pre-if type.
func applyBranchJoin(
	guard *guardResult,
	preIfTypes map[string]types.Type,
	ifBranchTypes map[string]types.Type,
	elsifJoinTypes map[string]types.Type,
	elseBranchTypes map[string]types.Type,
	elseNode *parser.Node,
	st *SymbolTable,
) {
	if preIfTypes == nil {
		return
	}

	for name, preTyp := range preIfTypes {
		// Start with the if-branch contribution.
		var joinType types.Type
		hasContribution := false

		if ifBranchTypes != nil {
			if t, ok := ifBranchTypes[name]; ok {
				joinType |= t
				hasContribution = true
			}
		}

		// Add elsif contributions.
		if elsifJoinTypes != nil {
			if t, ok := elsifJoinTypes[name]; ok {
				joinType |= t
				hasContribution = true
			}
		}

		// Add else contribution or implicit-else pre-if type.
		if elseNode != nil {
			if elseBranchTypes != nil {
				if t, ok := elseBranchTypes[name]; ok {
					joinType |= t
					hasContribution = true
				}
			}
		} else {
			// No else: implicit else contributes pre-if type.
			joinType |= preTyp
			hasContribution = true
		}

		if hasContribution && joinType != preTyp {
			st.UpdateType(name, joinType)
		}
	}
}

// walkElsifNode handles a single elsif node with guard-based flow narrowing.
// It mirrors walkConditionalStatement: walk condition, extract guard, walk
// block with guard, then handle the trailing else or elsif recursively.
//
// It returns the union of branch-end types for all guarded variables across
// this elsif and any trailing elsif/else blocks, for use by the caller's
// join computation.
func walkElsifNode(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
	classTypes map[uint32]string,
) map[string]types.Type {
	var conditionNode *parser.Node
	var block *parser.Node
	var elseNode *parser.Node
	var trailingElsif *parser.Node

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
			trailingElsif = child
		default:
			if conditionNode == nil {
				conditionNode = child
			}
		}
	}

	// Walk the condition to type its children.
	if conditionNode != nil {
		walkNode(conditionNode, source, st, annotations, diags, idx, classTypes)
	}

	guard := extractGuardPattern(conditionNode, source, st)

	// The elsif body negate flag starts false (no "unless" keyword in elsif).
	// Negated conditions (e.g. !defined) are handled inside flattenGuards
	// via XOR with each leaf's Negated flag, so we do not flip here.
	elsifBodyNegate := false

	// Walk the elsif body with the guard, capturing branch-end types.
	var elsifBranchTypes map[string]types.Type
	blockExits := block != nil && blockAlwaysExits(block, source)
	if block != nil {
		if !blockExits {
			elsifBranchTypes = make(map[string]types.Type)
		}
		walkBlockWithGuard(block, source, st, annotations, diags, flattenGuards(guard, elsifBodyNegate), elsifBranchTypes, idx, classTypes)
	}

	// Accumulate join types starting from this elsif's branch.
	joinTypes := elsifBranchTypes

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
			elseBranchTypes := make(map[string]types.Type)
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, flattenGuards(guard, !elsifBodyNegate), elseBranchTypes, idx, classTypes)
			joinTypes = mergeBranchTypes(joinTypes, elseBranchTypes)
		}
	}

	if trailingElsif != nil {
		trailingTypes := walkElsifNode(trailingElsif, source, st, annotations, diags, idx, classTypes)
		joinTypes = mergeBranchTypes(joinTypes, trailingTypes)
	}

	return joinTypes
}

// walkLoopStatement handles while statements with guard-based flow narrowing.
func walkLoopStatement(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
	classTypes map[uint32]string,
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
		walkNode(conditionNode, source, st, annotations, diags, idx, classTypes)
	}

	guard := extractGuardPattern(conditionNode, source, st)

	if bodyBlock != nil {
		walkBlockWithGuard(bodyBlock, source, st, annotations, diags, flattenGuards(guard, false), nil, idx, classTypes)
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
	idx *ProjectIndex,
	classTypes map[uint32]string,
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
		walkNode(iterSource, source, st, annotations, diags, idx, classTypes)
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
			walkNode(bodyBlock.Child(i), source, st, annotations, diags, idx, classTypes)
		}
		return types.Unknown
	}

	// Fallback: walk all children generically.
	for i := 0; i < node.ChildCount(); i++ {
		walkNode(node.Child(i), source, st, annotations, diags, idx, classTypes)
	}
	return types.Unknown
}

// collectExplicitReturns recursively searches the subtree rooted at node for
// all return_expression nodes. For each found, the type of the returned value
// (first named child of the return_expression) is looked up in annotations.
// A bare "return;" with no named child contributes Undef. All found types are
// ORed together and returned. Returns Unknown when no return expressions exist.
func collectExplicitReturns(node *parser.Node, source []byte, annotations map[uint32]types.Type) types.Type {
	if node == nil {
		return types.Unknown
	}

	var result types.Type

	if node.Kind() == "return_expression" {
		// Look for the first named child — the returned value.
		var valueNode *parser.Node
		for i := 0; i < node.ChildCount(); i++ {
			child := node.Child(i)
			if child != nil && child.IsNamed() {
				valueNode = child
				break
			}
		}
		if valueNode == nil {
			// bare return; → Undef
			result = types.Undef
		} else {
			if typ, ok := annotations[valueNode.StartByte()]; ok {
				result = typ
			}
		}
		return result
	}

	// Recurse into all children to find nested return_expression nodes.
	for i := 0; i < node.ChildCount(); i++ {
		result |= collectExplicitReturns(node.Child(i), source, annotations)
	}
	return result
}

// collectImplicitReturn inspects the last statement in block and returns the
// type that would be implicitly returned by falling off the end of the block.
//
// Recognized last-statement shapes:
//   - expression_statement: return the type of its first named child.
//   - conditional_statement: recurse into the if-block and else-block (if any),
//     union their implicit return types. A missing else contributes Unknown.
//   - anything else: return Unknown.
func collectImplicitReturn(block *parser.Node, source []byte, annotations map[uint32]types.Type) types.Type {
	if block == nil {
		return types.Unknown
	}

	// Find the last named child of the block (the last statement).
	var lastStmt *parser.Node
	for i := 0; i < block.ChildCount(); i++ {
		child := block.Child(i)
		if child != nil && child.IsNamed() {
			lastStmt = child
		}
	}

	if lastStmt == nil {
		return types.Unknown
	}

	switch lastStmt.Kind() {
	case "expression_statement":
		// The implicit return value is the first named child.
		for i := 0; i < lastStmt.ChildCount(); i++ {
			child := lastStmt.Child(i)
			if child != nil && child.IsNamed() {
				if typ, ok := annotations[child.StartByte()]; ok {
					return typ
				}
				return types.Unknown
			}
		}
		return types.Unknown

	case "conditional_statement":
		// Collect implicit returns from each branch of the conditional.
		// The implicit return of a conditional is the union of all branches.
		var result types.Type
		var ifBlock *parser.Node
		var elseBlock *parser.Node

		for i := 0; i < lastStmt.ChildCount(); i++ {
			child := lastStmt.Child(i)
			if child == nil {
				continue
			}
			switch child.Kind() {
			case "block":
				if ifBlock == nil {
					ifBlock = child
				}
			case "else":
				// Find the block inside the else node.
				for j := 0; j < child.ChildCount(); j++ {
					ec := child.Child(j)
					if ec != nil && ec.Kind() == "block" {
						elseBlock = ec
						break
					}
				}
			}
		}

		result |= collectImplicitReturn(ifBlock, source, annotations)
		result |= collectImplicitReturn(elseBlock, source, annotations)
		return result
	}

	return types.Unknown
}

// walkSubroutineDeclaration handles subroutine_declaration_statement nodes
// during pass 2. It walks all body statements (so expression nodes inside the
// body get type annotations), then collects all return paths — both explicit
// return expressions and the implicit return of the last statement — and stores
// their union as the subroutine's ReturnType in the symbol table.
//
// The sub is defined in the enclosing (main) scope by pass 1. After walking
// the body in a dedicated scope, ExitScope returns to main, where
// UpdateReturnType can find and update the symbol.
func walkSubroutineDeclaration(
	node *parser.Node,
	source []byte,
	st *SymbolTable,
	annotations map[uint32]types.Type,
	diags *[]Diagnostic,
	idx *ProjectIndex,
	classTypes map[uint32]string,
) types.Type {
	var subName string
	var bodyBlock *parser.Node

	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "bareword":
			subName = child.Text(source)
		case "block":
			bodyBlock = child
		}
	}

	if bodyBlock == nil {
		// No body — nothing to infer.
		return types.Code
	}

	// Enter a sub-level scope for pass-2 type narrowing (mirrors pass 1).
	st.EnterScope(subName)

	// Walk all body children so that all expression nodes are annotated.
	for i := 0; i < bodyBlock.ChildCount(); i++ {
		walkNode(bodyBlock.Child(i), source, st, annotations, diags, idx, classTypes)
	}

	// Collect explicit return types (all return_expression nodes in the body).
	explicit := collectExplicitReturns(bodyBlock, source, annotations)

	// Collect implicit return type (the last expression in the body).
	implicit := collectImplicitReturn(bodyBlock, source, annotations)

	// Union all return paths.
	returnType := explicit | implicit

	// Exit the sub scope before updating the symbol, which lives in the
	// enclosing (main) scope.
	st.ExitScope()

	if subName != "" && returnType != types.Unknown {
		st.UpdateReturnType(subName, returnType)
	}

	return types.Code
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

// remapGuardVar replaces occurrences of paramName in the guard result's
// VarName with argName. For compound guards, it recursively remaps both
// sides. Returns nil if the guard is nil.
func remapGuardVar(guard *guardResult, paramName, argName string) *guardResult {
	if guard == nil {
		return nil
	}

	result := *guard // shallow copy

	if result.Compound != nil {
		result.Compound = &compoundGuard{
			Op:    guard.Compound.Op,
			Left:  remapGuardVar(guard.Compound.Left, paramName, argName),
			Right: remapGuardVar(guard.Compound.Right, paramName, argName),
		}
		return &result
	}

	if result.VarName == paramName {
		result.VarName = argName
	}
	return &result
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
		// ClassName is preserved so that GuardIsa guards retain the class name
		// when the flattened leaf is used to shadow variables in walkBlockWithGuard.
		leaf := &guardResult{
			VarName:   guard.VarName,
			Guard:     guard.Guard,
			Negated:   guard.Negated != negate, // XOR: outer negate flips the leaf
			ClassName: guard.ClassName,
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
// If branchTypes is non-nil, the type of each guarded variable is captured
// into branchTypes just before the guard scope exits. Callers use this to
// compute the join type at the if/elsif/else merge point.
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
	branchTypes map[string]types.Type,
	idx *ProjectIndex,
	classTypes map[uint32]string,
) {
	// Build accumulated narrowed types by applying guards sequentially.
	// working tracks the current type for each variable as guards are applied,
	// whether or not each individual guard produced a change.
	// shadows collects only variables where at least one guard narrowed.
	// guardClassNames records the class name for GuardIsa guards by variable name.
	working := make(map[string]types.Type)
	shadows := make(map[string]types.Type)
	guardClassNames := make(map[string]string)

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

		// Record the class name for GuardIsa guards so the shadow symbol can
		// carry ClassType, enabling method resolution inside the block.
		if guard.Guard.Kind == types.GuardIsa && guard.ClassName != "" && !guard.Negated {
			guardClassNames[guard.VarName] = guard.ClassName
		}
	}

	scopeEntered := false
	if len(shadows) > 0 {
		st.EnterScope("guard")
		scopeEntered = true
		for name, typ := range shadows {
			className := guardClassNames[name]
			st.Define(Symbol{
				Name:      name,
				Type:      typ,
				ClassType: className,
				Kind:      SymVariable,
			})
		}
	}

	for i := 0; i < block.ChildCount(); i++ {
		walkNode(block.Child(i), source, st, annotations, diags, idx, classTypes)
	}

	// Capture branch-end types of guarded variables before the scope exits,
	// so the caller can compute the join type at the merge point.
	if branchTypes != nil && scopeEntered {
		for name := range shadows {
			if sym, found := st.Lookup(name); found {
				branchTypes[name] = sym.Type
			}
		}
	}

	if scopeEntered {
		st.ExitScope()
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

// FindSubDeclNode walks top-level children of root to find a
// subroutine_declaration_statement whose byte range matches the given
// startByte and endByte. Returns nil if no match is found.
func FindSubDeclNode(root *parser.Node, startByte, endByte uint32) *parser.Node {
	if root == nil {
		return nil
	}
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "subroutine_declaration_statement" &&
			child.StartByte() == startByte && child.EndByte() == endByte {
			return child
		}
	}
	return nil
}

// ResolveSubParam identifies the parameter variable in a single-argument
// subroutine declaration. It recognizes three styles:
//   - $_[0]: body contains an array_element_expression for @_[0]
//   - Signature: sub foo($x) { ... } — extracts first param name
//   - Shift: my $x = shift — extracts the variable name
//
// Returns the parameter name and true if recognized, or ("", false) if the
// sub doesn't match any recognized single-arg pattern.
func ResolveSubParam(subNode *parser.Node, source []byte) (string, bool) {
	if subNode == nil {
		return "", false
	}

	// Find the block (body) and optional signature.
	var bodyBlock *parser.Node
	var sigNode *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child == nil {
			continue
		}
		switch child.Kind() {
		case "block":
			bodyBlock = child
		case "signature":
			sigNode = child
		}
	}

	// Style 2: Signature parameter — sub foo($x) { ... }
	if sigNode != nil {
		paramName := extractFirstSigParam(sigNode, source)
		if paramName != "" {
			return paramName, true
		}
		return "", false
	}

	if bodyBlock == nil {
		return "", false
	}

	// Style 3: Shift — my $x = shift (first statement in body)
	if paramName := extractShiftParam(bodyBlock, source); paramName != "" {
		return paramName, true
	}

	// Style 1: $_[0] — check if the body references $_[0] anywhere.
	if bodyReferencesArg0(bodyBlock, source) {
		return "$_[0]", true
	}

	return "", false
}

// bodyReferencesArg0 checks if a block contains any reference to $_[0]
// (array_element_expression whose text is "$_[0]").
func bodyReferencesArg0(block *parser.Node, source []byte) bool {
	return nodeContainsArg0(block, source)
}

// nodeContainsArg0 recursively checks if any descendant node is an
// array_element_expression with text "$_[0]".
func nodeContainsArg0(node *parser.Node, source []byte) bool {
	if node == nil {
		return false
	}
	if node.Kind() == "array_element_expression" {
		if node.Text(source) == "$_[0]" {
			return true
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		if nodeContainsArg0(node.Child(i), source) {
			return true
		}
	}
	return false
}

// extractFirstSigParam finds the first scalar parameter in a signature node.
// Returns the parameter name (e.g. "$val") or "" if not found or if signature
// has more than one parameter (multi-arg subs are not supported).
//
// CST: signature → mandatory_parameter → scalar
func extractFirstSigParam(sigNode *parser.Node, source []byte) string {
	var params []string
	for i := 0; i < sigNode.ChildCount(); i++ {
		child := sigNode.Child(i)
		if child == nil {
			continue
		}
		// Signature params are wrapped in mandatory_parameter nodes.
		if child.Kind() == "mandatory_parameter" {
			for j := 0; j < child.ChildCount(); j++ {
				inner := child.Child(j)
				if inner != nil && inner.Kind() == "scalar" {
					params = append(params, inner.Text(source))
				}
			}
		}
	}
	// Only single-arg subs are supported.
	if len(params) == 1 {
		return params[0]
	}
	return ""
}

// extractShiftParam checks if the first statement in a block is
// "my $var = shift" or "my $var = shift @_", returning $var or "".
func extractShiftParam(block *parser.Node, source []byte) string {
	for i := 0; i < block.ChildCount(); i++ {
		child := block.Child(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		if child.Kind() != "expression_statement" {
			break // Only check the first named child.
		}
		return matchShiftAssignment(child, source)
	}
	return ""
}

// matchShiftAssignment checks if a statement node is "my $var = shift"
// and returns the variable name, or "" if it doesn't match.
//
// CST: expression_statement → assignment_expression →
//
//	variable_declaration (my + scalar) + func1op_call_expression (shift)
func matchShiftAssignment(stmt *parser.Node, source []byte) string {
	// Unwrap expression_statement to get the assignment.
	var assignNode *parser.Node
	for i := 0; i < stmt.ChildCount(); i++ {
		child := stmt.Child(i)
		if child != nil && child.Kind() == "assignment_expression" {
			assignNode = child
			break
		}
	}
	if assignNode == nil {
		return ""
	}

	var varName string
	var rhsIsShift bool
	for i := 0; i < assignNode.ChildCount(); i++ {
		child := assignNode.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "variable_declaration" {
			for j := 0; j < child.ChildCount(); j++ {
				inner := child.Child(j)
				if inner != nil && inner.Kind() == "scalar" {
					varName = inner.Text(source)
				}
			}
		}
		// Bare "shift" parses as func1op_call_expression in this grammar.
		if child.Kind() == "func1op_call_expression" {
			text := child.Text(source)
			if text == "shift" || strings.HasPrefix(text, "shift ") || strings.HasPrefix(text, "shift(@_)") {
				rhsIsShift = true
			}
		}
	}

	if varName != "" && rhsIsShift {
		return varName
	}
	return ""
}

// ExtractSubReturnExpr finds the expression that a subroutine body returns.
// It handles two cases:
//   - Explicit return: a single return_expression node in the body
//   - Implicit return: the last expression statement in the body
//
// Returns nil if the body has multiple return statements (too complex),
// or if the body is empty / has no expression.
func ExtractSubReturnExpr(block *parser.Node, source []byte) *parser.Node {
	if block == nil {
		return nil
	}

	var returnExpr *parser.Node
	returnCount := 0
	countReturns(block, &returnCount, &returnExpr)

	if returnCount > 1 {
		return nil
	}

	if returnCount == 1 && returnExpr != nil {
		for i := 0; i < returnExpr.ChildCount(); i++ {
			child := returnExpr.Child(i)
			if child != nil && child.IsNamed() {
				return child
			}
		}
		return nil
	}

	var lastExpr *parser.Node
	for i := 0; i < block.ChildCount(); i++ {
		child := block.Child(i)
		if child == nil || !child.IsNamed() {
			continue
		}
		if child.Kind() == "expression_statement" {
			lastExpr = child
		}
	}
	if lastExpr == nil {
		return nil
	}

	for i := 0; i < lastExpr.ChildCount(); i++ {
		child := lastExpr.Child(i)
		if child != nil && child.IsNamed() {
			return child
		}
	}
	return nil
}

// countReturns recursively counts return_expression nodes in a subtree.
// It stops descending into nested subroutine_declaration_statement nodes
// to avoid counting returns from inner subs.
func countReturns(node *parser.Node, count *int, found **parser.Node) {
	if node == nil {
		return
	}
	if node.Kind() == "return_expression" {
		*count++
		*found = node
		return
	}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Kind() == "subroutine_declaration_statement" {
			continue
		}
		countReturns(child, count, found)
	}
}

// ExtractArgVarName extracts a sigil-prefixed variable name from a call
// argument CST node. Returns empty string for non-variable arguments
// (literals, complex expressions, function calls).
func ExtractArgVarName(node *parser.Node, source []byte) string {
	if node == nil {
		return ""
	}
	switch node.Kind() {
	case "scalar":
		return sigildName("$", node, source)
	case "array":
		return sigildName("@", node, source)
	case "hash":
		return sigildName("%", node, source)
	}
	return ""
}
