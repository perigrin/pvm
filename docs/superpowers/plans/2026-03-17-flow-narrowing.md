# Flow Narrowing Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add guard-based flow narrowing to the PSC type inference engine so that variables inside if/unless/while blocks receive refined types based on condition guards.

**Architecture:** Special-case the inference walker for `conditional_statement` and `loop_statement` nodes. Walk the condition first, extract a guard pattern, shadow the guarded variable with a narrowed type in a new scope, walk the block body, then exit the scope. For else blocks, apply the negated guard.

**Tech Stack:** Go, tree-sitter (gotreesitter), existing `types/narrowing.go` and `infer/` packages.

---

## CST Reference

These tree-sitter node structures were verified by parsing sample Perl:

```
# if (defined($x)) { ... }
conditional_statement
  if (anonymous) "if"          ← keyword: "if" or "unless"
  ( (anonymous)
  func1op_call_expression      ← condition
    defined (anonymous)        ← guard function name
    ( (anonymous)
    scalar                     ← guarded variable
      $ (anonymous)
      varname "x"
    ) (anonymous)
  ) (anonymous)
  block                        ← guarded body
    { ... }

# if (...) { ... } else { ... }
conditional_statement
  if (anonymous) "if"
  ( ... condition ... )
  block                        ← if-body
  else                         ← else child node
    else (anonymous) "else"
    block                      ← else-body

# while (defined($x)) { ... }
loop_statement
  while (anonymous) "while"
  ( ... condition ... )
  block                        ← loop body
```

Key facts:
- `if` and `unless` share node kind `conditional_statement`; the keyword text distinguishes them
- `else` is a named child of the `conditional_statement` with its own `block`
- `while` uses `loop_statement` with the same condition/block structure
- `defined()` and `ref()` both produce `func1op_call_expression` with the keyword as first anonymous child
- `isa` is NOT supported by the current tree-sitter Perl grammar (parse error) — deferred
- `ref($x) eq 'TYPE'` requires string literals which have a known gotreesitter bug — deferred

## Scope of This Plan

**In scope:** GuardDefined and GuardRef (plain) in `if`, `unless`, `if/else`, and `while`.

**Deferred:** GuardRef with RefType (needs string fix), GuardIsa (needs grammar support), elsif chains, nested guards, early-exit narrowing.

---

## Task 1: NegateGuard in types/narrowing.go

Add a function that computes the negated type for a guard pattern.

**Files:**
- Modify: `internal/types/narrowing.go`
- Modify: `internal/types/narrowing_test.go`

- [ ] **Step 1: Write failing tests for NegateGuard**

Add to `internal/types/narrowing_test.go`:

```go
// TestNegateGuardDefined verifies that negating a defined() guard narrows to Undef.
func TestNegateGuardDefined(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardDefined}

	// Scalar negated under defined → Undef (variable is not defined)
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	// Any negated under defined → Undef
	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	// Undef negated under defined → Undef (already undef)
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.True(t, ok, "Undef negates under defined guard")
	assert.Equal(t, types.Undef, narrowed)

	// Int negated under defined → no useful narrowing
	narrowed, ok = types.NegateGuard(types.Int, guard)
	assert.False(t, ok, "Int does not negate meaningfully under defined guard")
	assert.Equal(t, types.Int, narrowed)
}

// TestNegateGuardRef verifies that negating a ref() guard keeps Scalar (non-ref).
func TestNegateGuardRef(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef}

	// Scalar negated under ref → Scalar (not a reference)
	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.True(t, ok, "Scalar negates under ref guard")
	assert.Equal(t, types.Scalar, narrowed)

	// Any negated under ref → Scalar
	narrowed, ok = types.NegateGuard(types.Any, guard)
	assert.True(t, ok, "Any negates under ref guard")
	assert.Equal(t, types.Scalar, narrowed)

	// Undef negated under ref → Undef (already non-ref, no useful narrowing)
	narrowed, ok = types.NegateGuard(types.Undef, guard)
	assert.False(t, ok, "Undef does not negate meaningfully under ref guard")
	assert.Equal(t, types.Undef, narrowed)
}

// TestNegateGuardRefWithType verifies that ref eq 'TYPE' negation produces no narrowing.
func TestNegateGuardRefWithType(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardRef, RefType: "HASH"}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.False(t, ok, "ref eq TYPE negation is not useful")
	assert.Equal(t, types.Scalar, narrowed)
}

// TestNegateGuardIsa verifies that isa negation produces no narrowing.
func TestNegateGuardIsa(t *testing.T) {
	guard := types.GuardPattern{Kind: types.GuardIsa}

	narrowed, ok := types.NegateGuard(types.Scalar, guard)
	assert.False(t, ok, "isa negation is not useful")
	assert.Equal(t, types.Scalar, narrowed)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/types/ -run TestNegateGuard -v`
Expected: FAIL — `NegateGuard` does not exist yet.

- [ ] **Step 3: Implement NegateGuard**

Add to `internal/types/narrowing.go`, after `NarrowByGuard`:

```go
// NegateGuard returns the type that typ narrows to when a guard expression is
// known to be FALSE. The second return value is true when useful narrowing
// occurred.
//
// Rules:
//   - GuardDefined: If typ could be undef (Scalar, Undef, Any), narrows to Undef.
//     Concrete non-undef types produce no useful narrowing.
//   - GuardRef (plain): Narrows to Scalar (non-reference).
//   - GuardRef with RefType: No useful narrowing ("not a HashRef" could be anything).
//   - GuardIsa: No useful narrowing ("not a Foo" could be anything).
func NegateGuard(typ Type, guard GuardPattern) (Type, bool) {
	switch guard.Kind {
	case GuardDefined:
		if typ == Scalar || typ == Undef || typ == Any {
			return Undef, true
		}
		return typ, false

	case GuardRef:
		if guard.RefType != "" {
			// "not ref eq TYPE" is not useful
			return typ, false
		}
		if typ == Scalar || typ == Any || typ == Ref {
			return Scalar, true
		}
		return typ, false

	case GuardIsa:
		return typ, false

	default:
		return typ, false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/types/ -run TestNegateGuard -v`
Expected: PASS (all 4 tests).

- [ ] **Step 5: Run full types test suite**

Run: `go test ./internal/types/ -v`
Expected: All existing tests still pass.

- [ ] **Step 6: Commit**

```bash
git add internal/types/narrowing.go internal/types/narrowing_test.go
git commit -m "feat(types): add NegateGuard for negated flow narrowing"
```

---

## Task 2: extractGuardPattern in infer.go

Add a function that reads a condition CST node and returns the guard pattern and variable name.

**Files:**
- Modify: `internal/infer/infer.go`
- Modify: `internal/infer/infer_test.go`

- [ ] **Step 1: Write failing tests for guard extraction**

These tests call `infer.Analyze` on Perl snippets and check the annotation map for narrowed types inside guarded blocks. Since the walker doesn't handle guards yet, the variables will have their default sigil type.

Add to `internal/infer/infer_test.go`:

```go
// --- Flow narrowing tests ---

func TestFlowNarrowingDefinedGuard(t *testing.T) {
	// if (defined($x)): if-body $x → Scalar (non-Undef), else-body $x → Undef.
	// Note: "my $x = undef" does NOT narrow $x to Undef because the undef
	// keyword produces Unknown type, so $x stays at sigil type Scalar.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=declaration, [1]=condition defined($x), [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar, ifBodyTyp, "if-body $x should be Scalar (defined guard on Scalar)")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingRefGuard(t *testing.T) {
	// Inside if (ref($x)), $x should be narrowed to Ref.
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// Find the $x inside "my $y = $x" (the if-body).
	// Source: "my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\n"
	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x at offset %d should be annotated", ifBodyXOffset)
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after ref() guard")
}

func TestFlowNarrowingUnlessDefinedGuard(t *testing.T) {
	// Inside unless (defined($x)), $x should be Undef (negated defined).
	src := []byte("my $x = undef;\nunless (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingWhileDefinedGuard(t *testing.T) {
	// Inside while (defined($x)), $x should be Scalar (non-Undef).
	src := []byte("my $x = undef;\nwhile (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	whileBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[whileBodyXOffset]
	assert.True(t, ok, "while-body $x should be annotated")
	assert.Equal(t, types.Scalar, typ, "while-body $x should be Scalar (defined guard)")
}

func TestFlowNarrowingIfElseRefGuard(t *testing.T) {
	// if (ref($x)) → Ref in if-body, Scalar in else-body
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// Find both $x references inside the blocks.
	// "my $y = $x" is in the if-body, "my $z = $x" is in the else-body.
	offsets := findAllVarOffsets(src, "$x")
	// offsets[0] = declaration "my $x", offsets[1] = defined($x) condition,
	// offsets[2] = if-body $x, offsets[3] = else-body $x
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, ifBodyTyp, "if-body $x should be Ref after ref() guard")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar, elseBodyTyp, "else-body $x should be Scalar (negated ref guard)")
}

func TestFlowNarrowingScopeRestoration(t *testing.T) {
	// After the if block, $x should revert to its pre-narrowed type.
	// "my $x = undef" leaves $x at sigil type Scalar (undef keyword → Unknown,
	// so no assignment narrowing occurs).
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\nmy $z = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition ref($x), [2]=if-body, [3]=after-if
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x")

	afterIfTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar, afterIfTyp, "post-if $x should revert to Scalar (pre-narrowed sigil type)")
}

// findLastVarOffset returns the byte offset of the last occurrence of varName in source.
func findLastVarOffset(source []byte, varName string) uint32 {
	needle := []byte(varName)
	last := -1
	for i := 0; i <= len(source)-len(needle); i++ {
		if string(source[i:i+len(needle)]) == string(needle) {
			last = i
		}
	}
	return uint32(last)
}

// findAllVarOffsets returns all byte offsets where varName appears in source.
func findAllVarOffsets(source []byte, varName string) []uint32 {
	needle := []byte(varName)
	var offsets []uint32
	for i := 0; i <= len(source)-len(needle); i++ {
		if string(source[i:i+len(needle)]) == string(needle) {
			offsets = append(offsets, uint32(i))
		}
	}
	return offsets
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infer/ -run TestFlowNarrowing -v`
Expected: FAIL — flow narrowing is not implemented yet. Variables will have their assignment-narrowed or sigil types, not guard-narrowed types.

- [ ] **Step 3: Implement extractGuardPattern**

Add to `internal/infer/infer.go`:

```go
// guardResult holds the result of extracting a guard pattern from a condition node.
type guardResult struct {
	VarName string
	Guard   types.GuardPattern
}

// extractGuardPattern examines a condition expression node and returns the
// guard pattern if the condition matches a recognized form (defined($x) or
// ref($x)). Returns nil if no guard pattern is recognized.
//
// Recognized CST shapes:
//
//   func1op_call_expression with keyword "defined" and scalar child → GuardDefined
//   func1op_call_expression with keyword "ref" and scalar child → GuardRef
func extractGuardPattern(node *parser.Node, source []byte) *guardResult {
	if node == nil {
		return nil
	}

	kind := node.Kind()

	// Pattern: defined($x) or ref($x)
	if kind == "func1op_call_expression" {
		return extractFunc1opGuard(node, source)
	}

	return nil
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
```

- [ ] **Step 4: Run tests (still failing — walker not wired yet)**

Run: `go test ./internal/infer/ -run TestFlowNarrowing -v`
Expected: Still FAIL — `extractGuardPattern` exists but `walkNode` doesn't call it yet.

- [ ] **Step 5: Commit extraction logic**

```bash
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "feat(infer): add extractGuardPattern and flow narrowing tests (red)"
```

---

## Task 3: Wire flow narrowing into walkNode

Special-case `conditional_statement` and `loop_statement` in `walkNode` to walk children in guard-aware order with scoped type overrides.

**Files:**
- Modify: `internal/infer/infer.go`

- [ ] **Step 1: Verify helpers compile**

Run: `go build ./internal/infer/`
Expected: Compiles without error (extractGuardPattern from Task 2 is present).

- [ ] **Step 2: Add walkConditionalStatement helper**

Add to `internal/infer/infer.go`:

```go
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

	return types.Unknown
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
		currentType := types.Scalar // default for undeclared
		if sym, ok := st.Lookup(guard.VarName); ok {
			currentType = sym.Type
		}

		var narrowedType types.Type
		var narrowed bool
		if negate {
			narrowedType, narrowed = types.NegateGuard(currentType, guard.Guard)
		} else {
			narrowedType, narrowed = types.NarrowByGuard(currentType, guard.Guard)
		}

		if narrowed {
			st.EnterScope("guard")
			st.Define(Symbol{
				Name: guard.VarName,
				Type: narrowedType,
				Kind: SymVariable,
			})
			for i := 0; i < block.ChildCount(); i++ {
				walkNode(block.Child(i), source, st, annotations, diags)
			}
			st.ExitScope()
			return
		}
	}

	// No guard or no narrowing: walk normally.
	for i := 0; i < block.ChildCount(); i++ {
		walkNode(block.Child(i), source, st, annotations, diags)
	}
}
```

- [ ] **Step 3: Verify helpers compile**

Run: `go build ./internal/infer/`
Expected: Compiles without error.

- [ ] **Step 4: Add special cases to walkNode**

Modify `walkNode` in `internal/infer/infer.go`. Insert a check before the generic child recursion to intercept `conditional_statement` and `loop_statement`:

```go
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
```

- [ ] **Step 5: Run flow narrowing tests**

Run: `go test ./internal/infer/ -run TestFlowNarrowing -v`
Expected: PASS (all 6 flow narrowing tests).

- [ ] **Step 6: Run full infer test suite**

Run: `go test ./internal/infer/ -v`
Expected: All existing tests still pass.

- [ ] **Step 7: Run full project tests**

Run: `make test`
Expected: All tests pass across all packages.

- [ ] **Step 8: Commit**

```bash
git add internal/infer/infer.go
git commit -m "feat(infer): wire flow narrowing for if/unless/while with scoped type overrides"
```

---

## Task 4: Edge case tests

Add tests for non-guard conditions, unless+ref, and nested guards.

**Files:**
- Modify: `internal/infer/infer_test.go`

- [ ] **Step 1: Write edge case tests**

Add to `internal/infer/infer_test.go`:

```go
func TestFlowNarrowingNonGuardCondition(t *testing.T) {
	// if ($x > 0) is not a recognized guard; $x should keep its type.
	src := []byte("my $x = 42;\nif ($x > 0) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition, [2]=if-body
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x")

	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Int, ifBodyTyp, "if-body $x should keep Int (no guard narrowing)")
}

func TestFlowNarrowingUnlessRefWithElse(t *testing.T) {
	// unless (ref($x)): body $x → Scalar (negated ref), else $x → Ref (positive ref)
	src := []byte("my $x = undef;\nunless (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=condition, [2]=unless-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	unlessBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Scalar, unlessBodyTyp, "unless-body $x should be Scalar (negated ref guard)")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Ref, elseBodyTyp, "else-body $x should be Ref (positive ref guard)")
}
```

- [ ] **Step 2: Run edge case tests**

Run: `go test ./internal/infer/ -run "TestFlowNarrowing(NonGuard|UnlessRef)" -v`
Expected: PASS.

- [ ] **Step 3: Run full test suite**

Run: `make test`
Expected: All pass.

- [ ] **Step 4: Commit**

```bash
git add internal/infer/infer_test.go
git commit -m "test(infer): add edge case tests for non-guard conditions and unless+ref"
```
