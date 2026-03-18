# Extended Flow Narrowing Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend PSC's flow narrowing to support elsif guard narrowing, early-exit narrowing (guard clauses with return/die/exit), and for-loop variable scoping.

**Architecture:** All three features modify `internal/infer/infer.go` (the inference walker). Feature 2 also adds a `Negated` field to `guardResult` and new logic in `extractGuardPattern` to unwrap `!`/`not` wrappers. Feature 3 adds a new `walkForStatement` handler. No changes to the type system (`internal/types/`) are needed.

**Tech Stack:** Go, gotreesitter (pure-Go tree-sitter runtime), existing `infer/` and `types/` packages.

**Spec:** `docs/superpowers/specs/2026-03-18-extended-flow-narrowing-design.md`

---

## Task 1: elsif Guard Narrowing

**Files:**
- Modify: `internal/infer/infer.go:668-675` (replace elsif handling in `walkConditionalStatement`)
- Test: `internal/infer/infer_test.go` (add new tests at end of file)

### Step 1: Write failing test — elsif with defined() guard

- [ ] Add test to `internal/infer/infer_test.go`:

```go
func TestFlowNarrowingElsifWithGuard(t *testing.T) {
	// elsif (defined($x)) should narrow $x to Scalar in the elsif body.
	src := []byte("my $x = undef;\nif ($x isa Foo) {\n    my $y = $x;\n} elsif (defined($x)) {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=if-condition, [2]=if-body, [3]=elsif-condition, [4]=elsif-body
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated")
	assert.Equal(t, types.Scalar, elsifBodyTyp, "elsif-body $x should be Scalar (defined guard)")
}
```

### Step 2: Run test to verify it fails

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingElsifWithGuard -v`
- Expected: FAIL — elsif-body $x is Scalar (sigil type), not from guard narrowing. Note: This test may pass accidentally because `defined` on `Scalar` narrows to `Scalar`. If it passes, check the annotation value is correct for the right reason by also adding:

```go
func TestFlowNarrowingElsifWithRefGuard(t *testing.T) {
	// elsif (ref($x)) should narrow $x to Ref in the elsif body.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} elsif (ref($x)) {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated")
	assert.Equal(t, types.Ref, elsifBodyTyp, "elsif-body $x should be Ref (ref guard in elsif)")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingElsifWithRefGuard -v`
- Expected: FAIL — actual is Scalar (2), expected Ref (10).

### Step 3: Write failing test — 3-branch elsif chain

- [ ] Add test:

```go
func TestFlowNarrowingElsifChainThreeBranches(t *testing.T) {
	// Three branches: if + elsif + elsif. Each should get its own guard.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} elsif ($x isa Foo) {\n    my $c = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=if-cond, [2]=if-body, [3]=elsif1-cond, [4]=elsif1-body, [5]=elsif2-cond, [6]=elsif2-body
	require.True(t, len(offsets) >= 7, "should find at least 7 occurrences of $x, got %d", len(offsets))

	elsif1BodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "first elsif-body $x should be annotated")
	assert.Equal(t, types.Ref, elsif1BodyTyp, "first elsif-body $x should be Ref")

	elsif2BodyTyp, ok2 := annotations[offsets[6]]
	assert.True(t, ok2, "second elsif-body $x should be annotated")
	assert.Equal(t, types.Object, elsif2BodyTyp, "second elsif-body $x should be Object (isa guard)")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingElsifChainThreeBranches -v`
- Expected: FAIL

### Step 4: Write failing test — elsif with else

- [ ] Add test:

```go
func TestFlowNarrowingElsifWithElse(t *testing.T) {
	// elsif (ref($x)) with else: elsif-body → Ref, else-body → Scalar (negated ref)
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=if-cond, [2]=if-body, [3]=elsif-cond, [4]=elsif-body, [5]=else-body
	require.True(t, len(offsets) >= 6, "should find at least 6 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated")
	assert.Equal(t, types.Ref, elsifBodyTyp, "elsif-body $x should be Ref (ref guard)")

	elseBodyTyp, ok2 := annotations[offsets[5]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar, elseBodyTyp, "else-body $x should be Scalar (negated ref guard)")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingElsifWithElse -v`
- Expected: FAIL

### Step 5: Implement `walkElsifNode`

- [ ] In `internal/infer/infer.go`, add `walkElsifNode` function before `walkLoopStatement`:

```go
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
```

### Step 6: Update `walkConditionalStatement` to call `walkElsifNode`

- [ ] In `internal/infer/infer.go`, replace the elsif case in the scanning loop (lines 668-675):

Replace:
```go
	case "elsif":
		// Walk the elsif subtree without guard narrowing (deferred).
		// The elsif node has the same structure as a conditional_statement
		// (condition, block, optional else) so walking its children gives
		// them type annotations even though we don't apply guard narrowing.
		for j := 0; j < child.ChildCount(); j++ {
			walkNode(child.Child(j), source, st, annotations, diags)
		}
```

With:
```go
	case "elsif":
		walkElsifNode(child, source, st, annotations, diags)
```

### Step 7: Run tests to verify they pass

- [ ] Run: `go test ./internal/infer/ -v`
- Expected: ALL PASS (including the existing `TestFlowNarrowingElsifBranchGetsAnnotations` which should continue working since it tests a non-guard condition in elsif)

### Step 8: Commit

- [ ] ```bash
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "feat(infer): add guard narrowing for elsif branches

walkElsifNode extracts guard patterns from elsif conditions and
applies scoped type overrides, matching the if/unless behavior.
Supports arbitrary-length elsif chains via recursion."
```

---

## Task 2: Negation Unwrapping in `extractGuardPattern`

**Files:**
- Modify: `internal/infer/infer.go:525-558` (`guardResult` struct and `extractGuardPattern`)
- Modify: `internal/infer/infer.go:636-718` (`walkConditionalStatement` — use `Negated` field)
- Test: `internal/infer/infer_test.go`

### Step 1: Write failing test — negated defined guard in if

- [ ] Add test:

```go
func TestExtractGuardPatternNegatedDefined(t *testing.T) {
	// if (!defined($x)): the condition is a negated defined guard.
	// The if-body should get the negated guard (Undef), not the positive (Scalar).
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	// if (!defined($x)): if-body should have negated-defined = Undef
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Undef, ifBodyTyp, "if-body $x should be Undef (negated defined guard)")

	// else-body should have positive defined = Scalar
	elseBodyTyp, ok2 := annotations[offsets[3]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar, elseBodyTyp, "else-body $x should be Scalar (positive defined guard)")
}
```

### Step 2: Run test to verify it fails

- [ ] Run: `go test ./internal/infer/ -run TestExtractGuardPatternNegatedDefined -v`
- Expected: FAIL — `!defined($x)` is a `unary_expression` which `extractGuardPattern` doesn't recognize, so no guard is extracted and no narrowing is applied. Both branches get Scalar (sigil type).

### Step 3: Write failing test — `not` keyword

- [ ] Add test:

```go
func TestExtractGuardPatternNotKeyword(t *testing.T) {
	// if (not ref($x)): "not" is an ambiguous_function_call_expression wrapping ref.
	src := []byte("my $x = undef;\nif (not ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	// if (not ref($x)): if-body should have negated ref = Scalar
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar, ifBodyTyp, "if-body $x should be Scalar (negated ref guard)")

	// else-body should have positive ref = Ref
	elseBodyTyp, ok2 := annotations[offsets[3]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Ref, elseBodyTyp, "else-body $x should be Ref (positive ref guard)")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestExtractGuardPatternNotKeyword -v`
- Expected: FAIL

### Step 4: Add `Negated` field to `guardResult`

- [ ] In `internal/infer/infer.go`, modify the `guardResult` struct (line ~526):

```go
// guardResult holds the result of extracting a guard pattern from a condition node.
type guardResult struct {
	VarName string
	Guard   types.GuardPattern
	Negated bool
}
```

### Step 5: Add negation unwrapping to `extractGuardPattern`

- [ ] In `internal/infer/infer.go`, add two new branches to `extractGuardPattern` before the `return nil`:

```go
	// Pattern: !guard (unary negation)
	if kind == "unary_expression" {
		return extractNegatedGuard(node, source)
	}

	// Pattern: not guard (low-precedence negation)
	if kind == "ambiguous_function_call_expression" {
		return extractNotGuard(node, source)
	}
```

- [ ] Add the two helper functions after `extractIsaGuard`:

```go
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
```

### Step 6: Update `walkConditionalStatement` to use `Negated`

- [ ] In `walkConditionalStatement`, replace:

```go
	isUnless := keyword == "unless"

	// Walk the if/unless body with appropriate narrowing.
	if ifBlock != nil {
		walkBlockWithGuard(ifBlock, source, st, annotations, diags, guard, isUnless)
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
			walkBlockWithGuard(elseBlock, source, st, annotations, diags, guard, !isUnless)
		}
	}
```

With:

```go
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
```

### Step 7: Run tests to verify they pass

- [ ] Run: `go test ./internal/infer/ -v`
- Expected: ALL PASS (all existing tests plus the two new negation tests)

### Step 8: Commit

- [ ] ```bash
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "feat(infer): unwrap negated guards (!guard, not guard)

extractGuardPattern now recognizes unary_expression with ! and
ambiguous_function_call_expression with not, recursing into the
inner expression and setting Negated on the guardResult.
walkConditionalStatement uses the Negated flag to flip narrowing."
```

---

## Task 3: Early-Exit Detection

**Files:**
- Modify: `internal/infer/infer.go` (add `blockAlwaysExits` function)
- Test: `internal/infer/infer_test.go`

### Step 1: Write failing test — blockAlwaysExits with return

- [ ] Add test. Since `blockAlwaysExits` is a helper, test it indirectly through end-to-end behavior. The test for post-if narrowing (Task 4) will be the real validation. For now, write a unit-level test:

```go
func TestFlowNarrowingEarlyReturnNarrowsDefined(t *testing.T) {
	// if (!defined($x)) { return; } — after the if, $x should be Scalar (non-undef).
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=post-if reference
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar, postIfTyp, "post-if $x should be Scalar (defined guard after early return)")
}
```

### Step 2: Write failing test — block does NOT exit (negative test)

- [ ] Add test:

```go
func TestFlowNarrowingNoEarlyExitNoNarrowing(t *testing.T) {
	// if (!defined($x)) { $x = 1; } — block does not exit, so no post-if narrowing.
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    $x = 1;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=if-body assignment LHS, [3]=post-if ref
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar, postIfTyp, "post-if $x should remain Scalar (no early exit, no narrowing)")
}
```

### Step 3: Write failing test — early die

- [ ] Add test:

```go
func TestFlowNarrowingEarlyDieNarrowsRef(t *testing.T) {
	// if (!ref($x)) { die; } — after the if, $x should be Ref.
	src := []byte("my $x = undef;\nif (!ref($x)) {\n    die;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Ref, postIfTyp, "post-if $x should be Ref (ref guard after early die)")
}
```

### Step 4: Write failing test — unless with early exit

- [ ] Add test:

```go
func TestFlowNarrowingUnlessEarlyReturn(t *testing.T) {
	// unless (defined($x)) { return; } — after the unless, $x should be Scalar.
	src := []byte("my $x = undef;\nunless (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-unless $x should be annotated")
	assert.Equal(t, types.Scalar, postIfTyp, "post-unless $x should be Scalar (defined after early return)")
}
```

### Step 5: Write failing test — non-negated guard with early exit

- [ ] Add test:

```go
func TestFlowNarrowingNonNegatedEarlyReturn(t *testing.T) {
	// if (defined($x)) { return; } — after the if, $x should be Undef
	// (NegateGuard applied because the exiting branch proved defined).
	src := []byte("my $x = undef;\nif (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Undef, postIfTyp, "post-if $x should be Undef (negated defined guard after early return)")
}
```

- [ ] Run all five new tests: `go test ./internal/infer/ -run "TestFlowNarrowingEarly|TestFlowNarrowingNoEarly|TestFlowNarrowingUnlessEarly|TestFlowNarrowingNonNegated" -v`
- Expected: FAIL for the four early-exit tests (no post-if narrowing yet), PASS for the negative test (no narrowing expected anyway — might pass accidentally)

### Step 6: Implement `blockAlwaysExits`

- [ ] Add to `internal/infer/infer.go` before `walkBlockWithGuard`:

```go
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
		case "bareword":
			if target.Text(source) == "die" {
				return true
			}
		}
	}
	return false
}
```

### Step 7: Wire post-if narrowing into `walkConditionalStatement`

- [ ] In `walkConditionalStatement`, after walking the if-block and before walking the else body, add:

```go
	// Early-exit narrowing: if the if-body always exits and there is no
	// else/elsif, apply the else-branch guard to the rest of the scope.
	if ifBlock != nil && guard != nil && elseNode == nil && hasNoElsif &&
		blockAlwaysExits(ifBlock, source) {
		sym, found := st.Lookup(guard.VarName)
		if found {
			var elseType types.Type
			var narrowed bool
			if ifBodyNegate {
				// If-body had negated guard, so else-branch is the positive guard.
				elseType, narrowed = types.NarrowByGuard(sym.Type, guard.Guard)
			} else {
				// If-body had positive guard, so else-branch is the negated guard.
				elseType, narrowed = types.NegateGuard(sym.Type, guard.Guard)
			}
			if narrowed {
				st.UpdateType(guard.VarName, elseType)
			}
		}
	}
```

- [ ] You also need to track whether there was an elsif node. Add a `hasNoElsif` flag in the scanning loop. In the variable declarations at the top of `walkConditionalStatement`, add:

```go
	var hasNoElsif = true
```

And in the scanning loop's `"elsif"` case, add:

```go
	case "elsif":
		hasNoElsif = false
		walkElsifNode(child, source, st, annotations, diags)
```

### Step 8: Run tests to verify they pass

- [ ] Run: `go test ./internal/infer/ -v`
- Expected: ALL PASS

### Step 9: Commit

- [ ] ```bash
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "feat(infer): add early-exit narrowing for guard clauses

After if (!guard) { return/die/exit; }, the rest of the scope gets
the else-branch type via st.UpdateType. blockAlwaysExits detects
return_expression, bare die, and exit."
```

---

## Task 4: For-Loop Variable Scoping

**Files:**
- Modify: `internal/infer/infer.go:44-57` (add `for_statement` to `walkNode` switch)
- Test: `internal/infer/infer_test.go`

### Step 1: Write failing test — for-loop variable is Scalar inside body

- [ ] Add test:

```go
func TestFlowNarrowingForLoopVariable(t *testing.T) {
	// for my $item (@arr) { $item; } — $item should be Scalar inside the body.
	src := []byte("my @arr;\nfor my $item (@arr) {\n    $item;\n}\n")
	annotations, _ := analyzeSource(t, src)

	itemOffset := findLastVarOffset(src, "$item")
	typ, ok := annotations[itemOffset]
	assert.True(t, ok, "for-body $item should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $item should be Scalar")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingForLoopVariable -v`
- Expected: FAIL or PASS (depends on whether the generic walk annotates the scalar node). Check the result — if it already passes because the scalar node gets its sigil type, add a scope-leak test instead.

### Step 2: Write failing test — for-loop variable does not leak

- [ ] Add test:

```go
func TestFlowNarrowingForLoopVariableNoLeak(t *testing.T) {
	// After the for loop, $item should not be annotated (it's out of scope).
	// We verify by checking that a post-loop reference to $item gets the
	// sigil type Scalar (from the scalar node), not from a lingering scope entry.
	src := []byte("my @arr;\nfor my $item (@arr) {\n    $item;\n}\n$item;\n")
	_, _, st := analyzeSourceFull(t, src)

	// $item should not be in the symbol table after the for loop exits.
	_, ok := st.Lookup("$item")
	assert.False(t, ok, "$item should not be in symbol table after for loop")
}
```

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingForLoopVariableNoLeak -v`
- Expected: Likely FAIL — without `walkForStatement`, $item from pass 1 may be in a block scope that was exited, so `Lookup` should already return false. If it passes, the test still serves as a regression guard.

### Step 3: Write failing test — for-loop over list expression

- [ ] Add test:

```go
func TestFlowNarrowingForLoopOverList(t *testing.T) {
	// for my $n (1, 2, 3) { $n; } — $n should be Scalar inside the body.
	src := []byte("for my $n (1, 2, 3) {\n    $n;\n}\n")
	annotations, _ := analyzeSource(t, src)

	nOffset := findLastVarOffset(src, "$n")
	typ, ok := annotations[nOffset]
	assert.True(t, ok, "for-body $n should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $n should be Scalar")
}
```

### Step 4: Write failing test — for without `my`

- [ ] Add test:

```go
func TestFlowNarrowingForLoopWithoutMy(t *testing.T) {
	// for $item (@arr) { $item; } — same behavior, no my keyword.
	src := []byte("my @arr;\nfor $item (@arr) {\n    $item;\n}\n")
	annotations, _ := analyzeSource(t, src)

	itemOffset := findLastVarOffset(src, "$item")
	typ, ok := annotations[itemOffset]
	assert.True(t, ok, "for-body $item should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $item should be Scalar")
}
```

### Step 5: Implement `walkForStatement`

- [ ] In `internal/infer/infer.go`, add to the `walkNode` switch (after the `loop_statement` case):

```go
	case "for_statement":
		return walkForStatement(node, source, st, annotations, diags)
```

- [ ] Add the `walkForStatement` function before `walkBlockWithGuard`:

```go
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
			if iterSource == nil && child.Kind() != "scalar" {
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
```

### Step 6: Run tests to verify they pass

- [ ] Run: `go test ./internal/infer/ -v`
- Expected: ALL PASS

### Step 7: Commit

- [ ] ```bash
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "feat(infer): add for-loop variable scoping

walkForStatement defines the loop variable as Scalar in a dedicated
scope that spans the loop body. Handles for/foreach with and without
my, and both array and list_expression iteration sources."
```

---

## Task 5: C-Style For Verification

**Files:**
- Test: `internal/infer/infer_test.go`

### Step 1: Write test — C-style for initializer narrows variable

- [ ] Add test:

```go
func TestFlowNarrowingCStyleForInitializer(t *testing.T) {
	// for (my $i = 0; ...) — $i should be narrowed to Int by assignment.
	src := []byte("for (my $i = 0; $i < 10; $i++) {\n    $i;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// The $i reference inside the body should be Int (narrowed by my $i = 0).
	offsets := findAllVarOffsets(src, "$i")
	// Find the $i inside the block body (last occurrence).
	bodyOffset := offsets[len(offsets)-1]
	typ, ok := annotations[bodyOffset]
	assert.True(t, ok, "for-body $i should be annotated")
	assert.Equal(t, types.Int, typ, "for-body $i should be Int (narrowed by assignment)")
}
```

### Step 2: Run test

- [ ] Run: `go test ./internal/infer/ -run TestFlowNarrowingCStyleForInitializer -v`
- Expected: PASS (assignment narrowing via the generic walk already handles `my $i = 0`)
- If it fails, investigate and fix.

### Step 3: Commit

- [ ] ```bash
git add internal/infer/infer_test.go
git commit -m "test(infer): verify C-style for initializer narrows via assignment"
```

---

## Task 6: Full Test Suite Verification

### Step 1: Run full test suite

- [ ] Run: `make test`
- Expected: ALL PASS, 0 failures

### Step 2: Final commit if any cleanup needed

- [ ] If any test needed fixing, commit the fixes.

### Step 3: Update the extractGuardPattern doc comment

- [ ] Update the comment on `extractGuardPattern` to list all recognized shapes:

```go
// extractGuardPattern examines a condition expression node and returns the
// guard pattern if the condition matches a recognized form. Returns nil if
// no guard pattern is recognized.
//
// Recognized CST shapes:
//
//	func1op_call_expression with keyword "defined" and scalar child → GuardDefined
//	func1op_call_expression with keyword "ref" and scalar child → GuardRef
//	relational_expression with "isa" operator, scalar LHS, bareword RHS → GuardIsa
//	unary_expression with "!" wrapping a recognized guard → Negated guard
//	ambiguous_function_call_expression with "not" wrapping a recognized guard → Negated guard
```

- [ ] ```bash
git add internal/infer/infer.go
git commit -m "docs(infer): update extractGuardPattern comment with negation shapes"
```
