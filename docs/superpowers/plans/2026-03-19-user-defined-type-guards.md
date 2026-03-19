# User-Defined Type Guard Functions Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Detect user-defined single-arg subs whose body matches a recognized guard pattern, and apply that guard's narrowing when the sub is called in a condition.

**Architecture:** Add a fallback path in `extractFunctionCallGuard` that looks up the called function in the symbol table, finds the sub's AST node by byte range, resolves the parameter variable (three styles), extracts the return expression, runs `extractGuardPattern` recursively on it, and remaps the result to the call-site argument. Thread `*SymbolTable` through `extractGuardPattern` and its callees to enable symbol lookup. Extend existing guard extractors to accept `array_element_expression` nodes so `$_[0]` works as a guard argument.

**Tech Stack:** Go, tree-sitter CST via `internal/parser`, existing type inference in `internal/infer`

**Critical CST facts** (verified against the gotreesitter Perl grammar):
- `$_[0]` parses as `array_element_expression`, NOT `array_access_expression` or `scalar`
- Signature params nest inside `mandatory_parameter`: `signature` → `mandatory_parameter` → `scalar`
- Bare `shift` parses as `func1op_call_expression`, NOT `func0op_call_expression`
- `ref($_[0]) eq 'HASH'` with quoted strings inside brace-delimited sub bodies **corrupts the parse tree** (tree-sitter treats `'` as namespace separator). All tests must use bare `ref($_[0])` (→ `types.Ref`) or `defined($_[0])` (→ removes Undef) instead.

---

### Task 1: Thread SymbolTable Through extractGuardPattern

The `extractGuardPattern` function and its callees currently take only `(node, source)`. To look up user-defined subs, `extractFunctionCallGuard` needs access to the symbol table. Thread `*SymbolTable` through the call chain.

**Files:**
- Modify: `internal/infer/infer.go`

- [ ] **Step 1: Run tests to confirm green baseline**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok` — all tests PASS

- [ ] **Step 2: Add `st *SymbolTable` parameter to extractGuardPattern and callees**

Change signatures (no behavioral change — pure refactor):

```go
// extractGuardPattern: add st parameter
func extractGuardPattern(node *parser.Node, source []byte, st *SymbolTable) *guardResult {

// extractCompoundGuard: add st parameter, pass to recursive calls
func extractCompoundGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
    // Inside: extractGuardPattern(leftNode, source, st)
    // Inside: extractGuardPattern(rightNode, source, st)

// extractNegatedGuard: add st parameter, pass to recursive call
func extractNegatedGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
    // Inside: extractGuardPattern(innerNode, source, st)

// extractNotGuard: add st parameter, pass to recursive call
func extractNotGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
    // Inside: extractGuardPattern(innerNode, source, st)

// extractFunctionCallGuard: add st parameter
func extractFunctionCallGuard(node *parser.Node, source []byte, st *SymbolTable) *guardResult {
```

Update dispatch calls inside `extractGuardPattern`:
```go
return extractFunctionCallGuard(node, source, st)   // was: (node, source)
return extractCompoundGuard(node, source, st)        // was: (node, source)
return extractNegatedGuard(node, source, st)         // was: (node, source)
return extractNotGuard(node, source, st)             // was: (node, source)
// extractFunc1opGuard and extractIsaGuard are unchanged — they don't recurse
```

Update the three external call sites:
- `walkConditionalStatement` (line ~1176): `extractGuardPattern(conditionNode, source, st)`
- `walkElsifNode` (line ~1433): `extractGuardPattern(conditionNode, source, st)`
- `walkLoopStatement` (line ~1513): `extractGuardPattern(conditionNode, source, st)`

- [ ] **Step 3: Run tests to confirm all still pass**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok` — all tests PASS (pure refactor, no behavior change)

- [ ] **Step 4: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go
git commit -m "$(cat <<'EOF'
refactor(infer): thread SymbolTable through extractGuardPattern

Preparation for user-defined type guard functions (#392). The symbol
table is needed to look up sub declarations when a function call in a
condition doesn't match the known guard function table.
EOF
)"
```

---

### Task 2: Extend Guard Extractors to Accept array_element_expression

The existing `extractFunc1opGuard` and `extractIsaGuard` only recognize `scalar` nodes as guard arguments. `$_[0]` parses as `array_element_expression`, so these functions need to also accept that node kind. This is needed for user-defined guard functions that use `$_[0]` in their body.

**Files:**
- Modify: `internal/infer/infer.go`
- Test: `internal/infer/infer_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestFlowNarrowingDefinedGuardArrayElement(t *testing.T) {
	// defined($_[0]) inside a sub body uses array_element_expression.
	// Verify that extractFunc1opGuard handles this node type.
	src := []byte("my @arr;\nif (defined($arr[0])) {\n    my $y = $arr[0];\n}\n")
	// This tests that array_element_expression is recognized by
	// extractFunc1opGuard. If it works, the guard extracts correctly.
	annotations, _ := analyzeSource(t, src)
	_ = annotations // Mainly verifying no crash; array element guard is new behavior.
}
```

- [ ] **Step 2: Run test to verify baseline**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestFlowNarrowingDefinedGuardArrayElement -count=1 -v`
Expected: PASS (no crash, but no narrowing yet)

- [ ] **Step 3: Extend extractFunc1opGuard to accept array_element_expression**

In `extractFunc1opGuard`, change the child detection loop to also accept `array_element_expression`:

```go
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
```

- [ ] **Step 4: Extend extractIsaGuard to accept array_element_expression**

In `extractIsaGuard`, change the child detection to also accept `array_element_expression`:

```go
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
		if child.Kind() == "scalar" && varNode == nil {
			varNode = child
		} else if child.Kind() == "array_element_expression" && varNode == nil {
			varNode = child
			varName = child.Text(source)
		} else if child.Kind() == "bareword" && classNode == nil {
			classNode = child
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
```

- [ ] **Step 5: Run tests to confirm all pass**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok` — all tests PASS

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): extend guard extractors to accept array_element_expression

extractFunc1opGuard and extractIsaGuard now recognize
array_element_expression nodes (e.g. $_[0]) in addition to scalar
nodes. This is needed for user-defined guard functions whose body
references $_[0]. Part of #392.
EOF
)"
```

---

### Task 3: Add findSubDeclNode Helper

Locate a subroutine's AST node given its byte range from the symbol table. Walk top-level children of the source file to find a `subroutine_declaration_statement` matching the symbol's `StartByte`/`EndByte`.

**Files:**
- Modify: `internal/infer/infer.go` (add new function)
- Test: `internal/infer/infer_test.go` (add test)

- [ ] **Step 1: Write the failing test**

```go
func TestFindSubDeclNode(t *testing.T) {
	src := []byte("my $x = 1;\nsub is_ref { ref($_[0]) }\nmy $y = 2;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	// Find the sub node using its byte range.
	// "sub is_ref { ref($_[0]) }" starts at byte 11.
	var subStart, subEnd uint32
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil && child.Kind() == "subroutine_declaration_statement" {
			subStart = child.StartByte()
			subEnd = child.EndByte()
		}
	}
	require.NotZero(t, subEnd, "should find a sub in the source")

	subNode := infer.FindSubDeclNode(root, subStart, subEnd)
	require.NotNil(t, subNode, "should find the sub declaration node")
	assert.Equal(t, "subroutine_declaration_statement", subNode.Kind())
	assert.Contains(t, subNode.Text(src), "is_ref")
}

func TestFindSubDeclNodeNotFound(t *testing.T) {
	src := []byte("my $x = 1;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	subNode := infer.FindSubDeclNode(root, 99, 199)
	assert.Nil(t, subNode, "should return nil when no matching sub found")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestFindSubDeclNode -count=1 -v`
Expected: FAIL — `FindSubDeclNode` undefined

- [ ] **Step 3: Write minimal implementation**

Add to `internal/infer/infer.go`:

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestFindSubDeclNode -count=1 -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add FindSubDeclNode helper for locating sub AST nodes

Walks top-level children of the source file to find a
subroutine_declaration_statement matching a given byte range.
Needed for user-defined type guard function resolution (#392).
EOF
)"
```

---

### Task 4: Add resolveSubParam Helper — $_[0] Style

Extract the parameter variable from a subroutine's AST node. Start with `$_[0]` detection only (TDD: one style at a time).

**Files:**
- Modify: `internal/infer/infer.go` (add new functions)
- Test: `internal/infer/infer_test.go` (add test)

- [ ] **Step 1: Write the failing test for $_[0] style**

```go
func TestResolveSubParamArrayElement(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve parameter")
	assert.Equal(t, "$_[0]", paramName)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParam -count=1 -v`
Expected: FAIL — `ResolveSubParam` undefined

- [ ] **Step 3: Write minimal implementation (just $_[0] support)**

Add to `internal/infer/infer.go`:

```go
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

// extractFirstSigParam is a stub — implemented in Task 5.
func extractFirstSigParam(sigNode *parser.Node, source []byte) string {
	return ""
}

// extractShiftParam is a stub — implemented in Task 6.
func extractShiftParam(block *parser.Node, source []byte) string {
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamArrayElement -count=1 -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add ResolveSubParam with $_[0] detection

Resolves the parameter variable in single-arg subs. First style:
$_[0] via array_element_expression detection. Signature and shift
stubs to be filled in next. Part of #392.
EOF
)"
```

---

### Task 5: Add Signature Parameter Resolution

Fill in `extractFirstSigParam` to handle `sub foo($x) { ... }` style.

**Files:**
- Modify: `internal/infer/infer.go`
- Test: `internal/infer/infer_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestResolveSubParamSignature(t *testing.T) {
	src := []byte("sub is_defined ($val) { defined($val) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve signature parameter")
	assert.Equal(t, "$val", paramName)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamSignature -count=1 -v`
Expected: FAIL — `extractFirstSigParam` returns "" (stub)

- [ ] **Step 3: Implement extractFirstSigParam**

Replace the stub. CST structure is: `signature` → `mandatory_parameter` → `scalar`.

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamSignature -count=1 -v`
Expected: PASS

- [ ] **Step 5: Write test for multi-arg rejection**

```go
func TestResolveSubParamMultiArgRejected(t *testing.T) {
	src := []byte("sub check ($x, $y) { defined($x) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	_, ok := infer.ResolveSubParam(subNode, src)
	assert.False(t, ok, "multi-arg sub should not be recognized as guard")
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamMultiArgRejected -count=1 -v`
Expected: PASS (extractFirstSigParam returns "" for 2 params)

- [ ] **Step 7: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 8: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add signature parameter resolution for guard functions

extractFirstSigParam walks through mandatory_parameter children of
signature nodes to find the first scalar param. Multi-arg subs are
rejected. Part of #392.
EOF
)"
```

---

### Task 6: Add Shift Parameter Resolution

Fill in `extractShiftParam` to handle `sub foo { my $x = shift; ... }` style.

**Files:**
- Modify: `internal/infer/infer.go`
- Test: `internal/infer/infer_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestResolveSubParamShift(t *testing.T) {
	src := []byte("sub is_ref { my $val = shift; ref($val) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve shift parameter")
	assert.Equal(t, "$val", paramName)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamShift -count=1 -v`
Expected: FAIL — `extractShiftParam` returns "" (stub)

- [ ] **Step 3: Implement extractShiftParam and matchShiftAssignment**

Replace the stub. CST for `my $val = shift`:
`expression_statement` → `assignment_expression` → (`variable_declaration` → `scalar`) + `func1op_call_expression` (text "shift")

```go
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
//        variable_declaration (my + scalar) + func1op_call_expression (shift)
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestResolveSubParamShift -count=1 -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add shift parameter resolution for guard functions

extractShiftParam matches my $var = shift as the first statement in
a sub body. Bare shift parses as func1op_call_expression in the
gotreesitter Perl grammar. Part of #392.
EOF
)"
```

---

### Task 7: Add ExtractSubReturnExpr Helper

Extract the return expression from a subroutine body. Handles explicit `return expr` and implicit last-expression return.

**Files:**
- Modify: `internal/infer/infer.go` (add new function)
- Test: `internal/infer/infer_test.go` (add tests)

- [ ] **Step 1: Write the failing test for explicit return**

```go
func TestExtractSubReturnExprExplicit(t *testing.T) {
	src := []byte("sub is_ref { return ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	require.NotNil(t, expr, "should find return expression")
	assert.Contains(t, expr.Text(src), "ref")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestExtractSubReturnExpr -count=1 -v`
Expected: FAIL — `ExtractSubReturnExpr` undefined

- [ ] **Step 3: Write implementation**

Add to `internal/infer/infer.go`:

```go
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

	// Count return statements and find the single return expression.
	var returnExpr *parser.Node
	returnCount := 0
	countReturns(block, &returnCount, &returnExpr)

	// Multiple returns = too complex for guard analysis.
	if returnCount > 1 {
		return nil
	}

	// Single explicit return: extract the expression child of the return node.
	if returnCount == 1 && returnExpr != nil {
		for i := 0; i < returnExpr.ChildCount(); i++ {
			child := returnExpr.Child(i)
			if child != nil && child.IsNamed() {
				return child
			}
		}
		return nil
	}

	// Implicit return: last expression_statement in the block.
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

	// Unwrap expression_statement to get the actual expression.
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestExtractSubReturnExprExplicit -count=1 -v`
Expected: PASS

- [ ] **Step 5: Write test for implicit return**

```go
func TestExtractSubReturnExprImplicit(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	require.NotNil(t, expr, "should find implicit return expression")
	assert.Contains(t, expr.Text(src), "ref")
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestExtractSubReturnExpr -count=1 -v`
Expected: Both tests PASS

- [ ] **Step 7: Write test for multiple returns**

```go
func TestExtractSubReturnExprMultipleReturns(t *testing.T) {
	src := []byte("sub check { if ($x) { return 1 } return 0 }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	assert.Nil(t, expr, "multiple returns should be rejected")
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestExtractSubReturnExpr -count=1 -v`
Expected: All 3 tests PASS

- [ ] **Step 9: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 10: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add ExtractSubReturnExpr for guard function body analysis

Extracts the return expression from a sub body. Handles explicit
return and implicit last-expression. Rejects bodies with multiple
return statements. Part of #392.
EOF
)"
```

---

### Task 8: Add remapGuardVar Helper

Given a `guardResult` where `VarName` is the parameter name (e.g. `$val` or `$_[0]`), remap it to the call-site argument variable name. Tested via internal test (same package has access to unexported types).

**Files:**
- Modify: `internal/infer/infer.go` (add new function)
- Test: `internal/infer/remap_test.go` (new internal test file — same package)

- [ ] **Step 1: Write the failing test**

Create `internal/infer/remap_test.go` (internal test, same package — can access unexported guardResult):

```go
package infer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/types"
)

func TestRemapGuardVarSimple(t *testing.T) {
	guard := &guardResult{
		VarName: "$val",
		Guard:   types.GuardPattern{Kind: types.GuardRef},
	}
	remapped := remapGuardVar(guard, "$val", "$x")
	require.NotNil(t, remapped)
	assert.Equal(t, "$x", remapped.VarName)
	assert.Equal(t, types.GuardRef, remapped.Guard.Kind)
}

func TestRemapGuardVarNoMatch(t *testing.T) {
	guard := &guardResult{
		VarName: "$other",
		Guard:   types.GuardPattern{Kind: types.GuardDefined},
	}
	remapped := remapGuardVar(guard, "$val", "$x")
	require.NotNil(t, remapped)
	assert.Equal(t, "$other", remapped.VarName, "should not remap non-matching var")
}

func TestRemapGuardVarCompound(t *testing.T) {
	guard := &guardResult{
		Compound: &compoundGuard{
			Op: "&&",
			Left: &guardResult{
				VarName: "$_[0]",
				Guard:   types.GuardPattern{Kind: types.GuardDefined},
			},
			Right: &guardResult{
				VarName: "$_[0]",
				Guard:   types.GuardPattern{Kind: types.GuardRef},
			},
		},
	}
	remapped := remapGuardVar(guard, "$_[0]", "$x")
	require.NotNil(t, remapped)
	require.NotNil(t, remapped.Compound)
	assert.Equal(t, "$x", remapped.Compound.Left.VarName)
	assert.Equal(t, "$x", remapped.Compound.Right.VarName)
}

func TestRemapGuardVarNil(t *testing.T) {
	assert.Nil(t, remapGuardVar(nil, "$val", "$x"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestRemapGuardVar -count=1 -v`
Expected: FAIL — `remapGuardVar` undefined

- [ ] **Step 3: Write implementation**

Add to `internal/infer/infer.go`:

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestRemapGuardVar -count=1 -v`
Expected: All 4 tests PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/remap_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add remapGuardVar for call-site argument substitution

Replaces the parameter variable name in a guard result with the
call-site argument name. Handles compound guards recursively.
Part of #392.
EOF
)"
```

---

### Task 9: Wire Up User-Defined Guard Resolution

The main integration task. When `extractFunctionCallGuard` doesn't find a match in the guard function table, fall back to looking up the sub in the symbol table, analyzing its body, and returning a remapped guard result.

**Files:**
- Modify: `internal/infer/infer.go` (modify `extractFunctionCallGuard`, add `extractUserDefinedGuard`)
- Test: `internal/infer/infer_test.go` (add first integration test)

- [ ] **Step 1: Write the failing test — $_[0] style defined guard**

```go
func TestUserDefinedGuardFuncDefined(t *testing.T) {
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x = undef;\nif (is_defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "need at least 4 $x occurrences, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, ifBodyTyp, "if-body $x should have Undef removed")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncDefined -count=1 -v`
Expected: FAIL — no narrowing happens

- [ ] **Step 3: Implement user-defined guard resolution**

Modify `extractFunctionCallGuard` — add fallback after guard table miss. Add new `extractUserDefinedGuard` function.

```go
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

	// Check known guard function table first.
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

	// Find the sub's AST node by walking up from argNode to root,
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

	// Extract the return expression. For shift-style subs, skip the
	// shift assignment when looking for the return expression.
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
```

**Key design note:** Passing `nil` for `st` in the recursive `extractGuardPattern` call enforces the 1-level recursion limit. A user-defined guard calling another user-defined guard won't be recognized.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncDefined -count=1 -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): user-defined type guard function resolution (#392)

When a function call in a condition doesn't match the known guard
table, look up the sub in the symbol table, analyze its body for a
recognized guard pattern, and remap to the call-site argument.
Enforces 1-level recursion limit by passing nil SymbolTable to the
inner extractGuardPattern call.
EOF
)"
```

---

### Task 10: Add Remaining Integration Tests

Add tests for all guard function styles and edge cases from the design.

**Files:**
- Modify: `internal/infer/infer_test.go`

- [ ] **Step 1: Write test for $_[0] ref guard**

```go
func TestUserDefinedGuardFuncRefArrayAccess(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\nmy $x = undef;\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after user-defined ref guard")
}
```

- [ ] **Step 2: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncRefArrayAccess -count=1 -v`
Expected: PASS

- [ ] **Step 3: Write test for signature-style defined guard**

```go
func TestUserDefinedGuardFuncSignature(t *testing.T) {
	src := []byte("sub is_defined ($val) { defined($val) }\nmy $x = undef;\nif (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, typ, "if-body $x should have Undef removed after signature-style guard")
}
```

- [ ] **Step 4: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncSignature -count=1 -v`
Expected: PASS

- [ ] **Step 5: Write test for shift-style ref guard**

```go
func TestUserDefinedGuardFuncShift(t *testing.T) {
	src := []byte("sub is_ref { my $val = shift; ref($val) }\nmy $x = undef;\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after shift-style guard")
}
```

- [ ] **Step 6: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncShift -count=1 -v`
Expected: PASS

- [ ] **Step 7: Write test for isa guard function**

```go
func TestUserDefinedGuardFuncIsa(t *testing.T) {
	src := []byte("sub is_foo { $_[0] isa Foo }\nmy $x = undef;\nif (is_foo($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, typ, "if-body $x should be Object after isa guard")
}
```

- [ ] **Step 8: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncIsa -count=1 -v`
Expected: PASS

- [ ] **Step 9: Write test for negated call site**

```go
func TestUserDefinedGuardFuncNegatedDefined(t *testing.T) {
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x = undef;\nunless (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// unless(is_defined($x)) means body gets negated guard → Undef.
	unlessBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[unlessBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}
```

- [ ] **Step 10: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncNegatedDefined -count=1 -v`
Expected: PASS

- [ ] **Step 11: Write test for compound guard in body**

```go
func TestUserDefinedGuardFuncCompound(t *testing.T) {
	src := []byte("sub is_defined_ref { defined($_[0]) && ref($_[0]) }\nmy $x = undef;\nif (is_defined_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after compound guard (defined+ref)")
}
```

- [ ] **Step 12: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncCompound -count=1 -v`
Expected: PASS

- [ ] **Step 13: Write test for non-guard sub (no false positives)**

```go
func TestUserDefinedGuardFuncNonGuardIgnored(t *testing.T) {
	src := []byte("sub do_stuff { 42 }\nmy $x = undef;\nif (do_stuff($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	if ok {
		assert.Equal(t, types.Scalar, typ, "if-body $x should stay Scalar (non-guard sub)")
	}
}
```

- [ ] **Step 14: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFuncNonGuardIgnored -count=1 -v`
Expected: PASS

- [ ] **Step 15: Write test for complex argument (no narrowing)**

```go
func TestUserDefinedGuardFuncComplexArgNoCrash(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\nmy @arr;\nif (is_ref($arr[0])) {\n    my $y = 1;\n}\n")
	annotations, _ := analyzeSource(t, src)
	_ = annotations // Just verify no panic.
}
```

- [ ] **Step 16: Run all user-defined guard tests**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -run TestUserDefinedGuardFunc -count=1 -v`
Expected: All tests PASS

- [ ] **Step 17: Run full test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 18: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards
git add internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
test(infer): comprehensive tests for user-defined type guard functions

Tests all three parameter styles ($_[0], signatures, shift), defined/ref/isa
guard types, negated call sites, compound guards, non-guard rejection,
and complex argument handling. Part of #392.
EOF
)"
```

---

### Task 11: Run Full Project Test Suite

Verify the entire project still passes with all changes.

**Files:** None (verification only)

- [ ] **Step 1: Run make test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && make test`
Expected: All packages PASS, 0 failures

- [ ] **Step 2: If any failures, fix them before proceeding**

Any failures must be investigated and fixed. Do not skip.

- [ ] **Step 3: Verify with go vet**

Run: `cd /home/perigrin/dev/pvm/.worktrees/user-defined-type-guards && go vet ./...`
Expected: No issues
