# Type Guard Suggestions Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Attach guard suggestions to type-mismatch diagnostics so `psc check` output includes actionable hints like "Add guard: if (defined($x)) { ... }".

**Architecture:** Add a `Suggestion` field to the existing `Diagnostic` struct and a `suggestGuard` function that maps type mismatches to appropriate guard expressions using bitset operations. Populate the field inline at the two sites where `CodeTypeMismatch` diagnostics are emitted. Extend `FormatDiagnostic` to append a hint line when a suggestion is present.

**Tech Stack:** Go, existing `internal/infer` and `internal/types` packages

---

### Task 1: Add Suggestion Field to Diagnostic Struct

Add the `Suggestion string` field to the `Diagnostic` struct. Pure additive change — no behavioral change.

**Files:**
- Modify: `internal/infer/diagnostics.go`

- [ ] **Step 1: Run tests to confirm green baseline**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok` — all tests PASS

- [ ] **Step 2: Add Suggestion field to Diagnostic struct**

In `internal/infer/diagnostics.go`, add `Suggestion` after `Code`:

```go
// Diagnostic holds a single analysis finding with its source location.
type Diagnostic struct {
	StartByte  uint32
	EndByte    uint32
	Severity   Severity
	Message    string
	Code       string // machine-readable identifier, e.g. "arity-mismatch"
	Suggestion string // guard suggestion text, e.g. "Add guard: if (defined($x)) { ... }"
}
```

- [ ] **Step 3: Run tests to verify no breakage**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok` — adding a field is backward compatible

- [ ] **Step 4: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/infer/diagnostics.go
git commit -m "$(cat <<'EOF'
feat(infer): add Suggestion field to Diagnostic struct

Prepares for type guard suggestions by adding a string field to hold
actionable guard hints. No behavioral change. Part of #393.
EOF
)"
```

---

### Task 2: Implement suggestGuard Function

Add the `suggestGuard` function that maps a type mismatch (variable name, actual type, expected type) to a guard suggestion string.

**Files:**
- Modify: `internal/infer/diagnostics.go`

- [ ] **Step 1: Write failing test for suggestGuard — Undef case**

In `internal/infer/diagnostics_test.go`, add:

```go
func TestSuggestGuardDefined(t *testing.T) {
	// Scalar actual vs Int expected. Scalar includes Undef; Int does not.
	// Removing Undef from Scalar leaves Bool|Str|DualVar|Regex|Ref, which
	// still contains Int (IsSubtype(Int, narrowed) = true), so defined()
	// is suggested.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Int)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}
```

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestSuggestGuardDefined -count=1 -v`
Expected: FAIL — `SuggestGuard` does not exist yet

- [ ] **Step 2: Implement suggestGuard**

In `internal/infer/diagnostics.go`, add:

```go
// SuggestGuard returns a guard suggestion string for a type mismatch where
// actual is the variable's current type and expected is what the callee
// requires. Returns empty string when no guard can help.
//
// Each candidate guard is tried in priority order. A guard is only suggested
// if applying its narrowing to actual would make the result compatible with
// expected. Compatibility is checked two ways:
//   - types.TypeSatisfies(narrowed, expected) — standard subtype/polymorphic check
//   - types.IsSubtype(expected, narrowed) — "expected fits within narrowed",
//     handles cases where narrowed is a broad union that lost polymorphic status
//     (e.g. Scalar &^ Undef is not in polymorphicMasks but still contains Int)
//
// Priority order (first match wins):
//  1. defined() guard — narrows by removing Undef bit
//  2. ref() guard — narrows to Ref mask
//  3. builtin::is_bool() guard — narrows to Bool
//  4. No match → empty string
func SuggestGuard(varName string, actual, expected types.Type) string {
	if varName == "" {
		return ""
	}

	// Priority 1: defined() — would removing Undef make actual compatible?
	if actual&types.Undef != 0 && expected&types.Undef == 0 {
		narrowed := actual &^ types.Undef
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (defined(" + varName + ")) { ... }"
		}
	}

	// Priority 2: ref() — would narrowing to Ref make actual compatible?
	if actual&types.Ref != 0 {
		narrowed := actual & types.Ref
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (ref(" + varName + ")) { ... }"
		}
	}

	// Priority 3: builtin::is_bool() — would narrowing to Bool help?
	if actual&types.Bool != 0 {
		narrowed := actual & types.Bool
		if guardNarrowingSatisfies(narrowed, expected) {
			return "Add guard: if (builtin::is_bool(" + varName + ")) { ... }"
		}
	}

	return ""
}

// guardNarrowingSatisfies checks whether a narrowed type is compatible with
// the expected type. It uses TypeSatisfies first (handles polymorphic types
// and standard subtype checks), then falls back to checking if expected is
// a subtype of narrowed. The fallback handles cases like Scalar &^ Undef:
// a broad union that lost its polymorphic status in polymorphicMasks but
// still contains the expected type's bits (e.g. Int is a subtype of
// Scalar &^ Undef).
func guardNarrowingSatisfies(narrowed, expected types.Type) bool {
	if types.TypeSatisfies(narrowed, expected) {
		return true
	}
	return types.IsSubtype(expected, narrowed)
}
```

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestSuggestGuardDefined -count=1 -v`
Expected: PASS

- [ ] **Step 3: Write additional unit tests**

In `internal/infer/diagnostics_test.go`, add:

```go
func TestSuggestGuardRef(t *testing.T) {
	// Scalar actual, Ref expected — defined() fires first because removing
	// Undef from Scalar still contains Ref (IsSubtype(Ref, Scalar&^Undef)).
	// The defined() guard is less invasive than ref().
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Ref)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardRefDirect(t *testing.T) {
	// Ref|Int actual (no Undef), Ref expected — P1 skips (no Undef).
	// P2: narrowed = Ref. guardNarrowingSatisfies(Ref, Ref) = true.
	suggestion := infer.SuggestGuard("$x", types.Ref|types.Int, types.Ref)
	assert.Equal(t, "Add guard: if (ref($x)) { ... }", suggestion)
}

func TestSuggestGuardNoMatchStructural(t *testing.T) {
	// Scalar actual, Array expected — no guard can make Scalar satisfy Array.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Array)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardNoMatchSubtype(t *testing.T) {
	// Int actual vs Str expected — Int is already a subtype of Str,
	// so TypeSatisfies(Int, Str) is true and this wouldn't be a mismatch.
	// Use a case where it is a real mismatch: Array vs Hash.
	suggestion := infer.SuggestGuard("@arr", types.Array, types.Hash)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardEmptyVarName(t *testing.T) {
	suggestion := infer.SuggestGuard("", types.Scalar, types.Ref)
	assert.Equal(t, "", suggestion)
}

func TestSuggestGuardObject(t *testing.T) {
	// Scalar actual, Object expected — defined() fires because removing
	// Undef from Scalar still contains Object (IsSubtype(Object, narrowed)).
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Object)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardBool(t *testing.T) {
	// Scalar actual, Bool expected — defined() fires first because
	// Scalar &^ Undef still contains Bool. For a non-Scalar type that
	// includes Bool but not Undef, is_bool() would fire instead.
	suggestion := infer.SuggestGuard("$x", types.Scalar, types.Bool)
	assert.Equal(t, "Add guard: if (defined($x)) { ... }", suggestion)
}

func TestSuggestGuardBoolDirect(t *testing.T) {
	// Bool|Int actual, Bool expected — no Undef bit, so P1 skips.
	// P2 (ref): no Ref bits, skips. P3 (is_bool): Bool & (Bool|Int) = Bool,
	// guardNarrowingSatisfies(Bool, Bool) = true.
	suggestion := infer.SuggestGuard("$x", types.Bool|types.Int, types.Bool)
	assert.Equal(t, "Add guard: if (builtin::is_bool($x)) { ... }", suggestion)
}
```

- [ ] **Step 4: Run all suggestGuard tests**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestSuggestGuard -count=1 -v`
Expected: All PASS

- [ ] **Step 5: Run full infer test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 6: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/infer/diagnostics.go internal/infer/diagnostics_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add SuggestGuard for type-mismatch guard suggestions

Maps type mismatches to appropriate guard expressions using bitset
operations. Handles defined, ref, object, and bool guard patterns.
Part of #393.
EOF
)"
```

---

### Task 3: Add extractArgVarName Helper

Add a helper to extract a sigil-prefixed variable name from a call argument CST node. Returns empty string for non-variable arguments (literals, complex expressions).

**Files:**
- Modify: `internal/infer/infer.go`

- [ ] **Step 1: Write failing test**

In `internal/infer/infer_test.go`, add:

```go
func TestExtractArgVarNameScalar(t *testing.T) {
	src := []byte("push($x, 1);\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)

	// Find the function_call_expression node.
	root := tree.RootNode()
	var callNode *parser.Node
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil {
			for j := 0; j < child.ChildCount(); j++ {
				gc := child.Child(j)
				if gc != nil && gc.Kind() == "function_call_expression" {
					callNode = gc
				}
			}
		}
	}
	if callNode == nil {
		// Try direct children
		for i := 0; i < root.ChildCount(); i++ {
			child := root.Child(i)
			if child != nil && child.Kind() == "function_call_expression" {
				callNode = child
			}
		}
	}
	require.NotNil(t, callNode, "should find function_call_expression")

	name := infer.ExtractArgVarName(callNode.Child(0), src)
	// The first named child after "function" would be the first arg.
	// We just need the function to handle scalar nodes correctly.
	// This test may need adjustment based on exact CST structure.
}
```

**IMPORTANT:** This test is a sketch. The implementer should verify the exact CST structure of `push($x, 1)` using `go run ./cmd/psc parse` and adjust the test to find the actual `$x` argument node. The key assertion is:

```go
assert.Equal(t, "$x", infer.ExtractArgVarName(argNode, src))
```

- [ ] **Step 2: Implement ExtractArgVarName**

In `internal/infer/infer.go`, add:

```go
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
```

- [ ] **Step 3: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestExtractArgVarName -count=1 -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): add ExtractArgVarName for argument variable extraction

Extracts sigil-prefixed variable names from call argument CST nodes.
Returns empty string for non-variable arguments. Part of #393.
EOF
)"
```

---

### Task 4: Wire Suggestions Into Type-Mismatch Diagnostics

At both sites where `CodeTypeMismatch` diagnostics are emitted, extract the argument variable name and populate the `Suggestion` field via `SuggestGuard`.

**Files:**
- Modify: `internal/infer/infer.go`

- [ ] **Step 1: Write failing integration test**

In `internal/infer/infer_test.go`, add:

```go
func TestTypeMismatchDiagnosticSuggestionWired(t *testing.T) {
	// push expects Array as first arg. Passing Scalar $x triggers
	// a type-mismatch diagnostic. No guard can narrow Scalar to Array,
	// so Suggestion should be empty — but the field must exist and
	// be populated (empty string is valid).
	// This test verifies the wiring: SuggestGuard is called during
	// diagnostic emission. The actual suggestion logic is tested
	// by TestSuggestGuard* unit tests.
	src := []byte("my $x;\npush($x, 1);\n")
	_, diags := analyzeSource(t, src)

	require.True(t, len(diags) > 0, "should have at least one diagnostic")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			// Scalar vs Array: no guard helps, Suggestion should be empty.
			assert.Empty(t, d.Suggestion,
				"Scalar vs Array mismatch should have no suggestion")
			found = true
		}
	}
	assert.True(t, found, "should find a type-mismatch diagnostic")
}
```

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestTypeMismatchDiagnosticSuggestionWired -count=1 -v`
Expected: FAIL — Suggestion field not yet populated (test verifies the field exists via the assertion)

- [ ] **Step 2: Wire SuggestGuard into inferFunctionCallType**

In `internal/infer/infer.go`, in the `inferFunctionCallType` function (~line 352-363), modify the type-mismatch diagnostic emission:

```go
	// Validate argument types.
	for i, arg := range args {
		argType := annotations[arg.StartByte()]
		expectedType := builtinArgType(sig, i)
		if argType != types.Unknown && !types.TypeSatisfies(argType, expectedType) {
			argVarName := ExtractArgVarName(arg, source)
			*diags = append(*diags, Diagnostic{
				StartByte:  arg.StartByte(),
				EndByte:    arg.EndByte(),
				Severity:   Error,
				Code:       CodeTypeMismatch,
				Message:    typeMismatchMessage(name, i, expectedType, argType),
				Suggestion: SuggestGuard(argVarName, argType, expectedType),
			})
		}
	}
```

- [ ] **Step 3: Wire SuggestGuard into inferFunc1opCallType**

In `internal/infer/infer.go`, in the `inferFunc1opCallType` function (~line 507-519), apply the same change:

```go
	for i, arg := range args {
		argType := annotations[arg.StartByte()]
		expectedType := builtinArgType(sig, i)
		if argType != types.Unknown && !types.TypeSatisfies(argType, expectedType) {
			argVarName := ExtractArgVarName(arg, source)
			*diags = append(*diags, Diagnostic{
				StartByte:  arg.StartByte(),
				EndByte:    arg.EndByte(),
				Severity:   Error,
				Code:       CodeTypeMismatch,
				Message:    typeMismatchMessage(name, i, expectedType, argType),
				Suggestion: SuggestGuard(argVarName, argType, expectedType),
			})
		}
	}
```

- [ ] **Step 4: Run integration test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestTypeMismatchDiagnosticSuggestionWired -count=1 -v`
Expected: PASS

- [ ] **Step 5: Write test for no suggestion when no guard helps**

```go
func TestTypeMismatchDiagnosticNoSuggestionWhenNoGuardHelps(t *testing.T) {
	// push expects Array as first arg. Scalar $x triggers type-mismatch,
	// but no guard can narrow Scalar to satisfy Array — so Suggestion
	// should be empty.
	src := []byte("my $x = 1;\npush($x, 1);\n")
	_, diags := analyzeSource(t, src)

	require.True(t, len(diags) > 0, "should have at least one diagnostic")

	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			assert.Empty(t, d.Suggestion,
				"Scalar vs Array mismatch should have no suggestion")
		}
	}
}

func TestTypeMismatchDiagnosticNoSuggestionWhenClean(t *testing.T) {
	// push expects Array. Passing @arr is correct — no diagnostic at all.
	src := []byte("my @arr;\npush(@arr, 1);\n")
	_, diags := analyzeSource(t, src)

	for _, d := range diags {
		assert.NotEqual(t, infer.CodeTypeMismatch, d.Code,
			"clean code should have no type-mismatch diagnostics")
	}
}
```

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestTypeMismatchDiagnosticNoSuggestion -count=1 -v`
Expected: PASS

- [ ] **Step 6: Run full infer test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 7: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/infer/infer.go internal/infer/infer_test.go
git commit -m "$(cat <<'EOF'
feat(infer): wire guard suggestions into type-mismatch diagnostics

Both inferFunctionCallType and inferFunc1opCallType now extract the
argument variable name and populate Suggestion via SuggestGuard when
emitting type-mismatch diagnostics. Part of #393.
EOF
)"
```

---

### Task 5: Extend FormatDiagnostic to Show Suggestions

When `Diagnostic.Suggestion` is non-empty, `FormatDiagnostic` appends a hint line.

**Files:**
- Modify: `internal/infer/diagnostics.go`
- Modify: `internal/infer/diagnostics_test.go`

- [ ] **Step 1: Write failing test**

In `internal/infer/diagnostics_test.go`, add:

```go
func TestFormatDiagnosticWithSuggestion(t *testing.T) {
	source := []byte("my $x = 1;\npush($x, 1);\n")

	d := infer.Diagnostic{
		StartByte:  15,
		EndByte:    17,
		Severity:   infer.Error,
		Message:    "call to push: argument 1 expects Array, got Scalar",
		Code:       infer.CodeTypeMismatch,
		Suggestion: "Add guard: if (ref($x)) { ... }",
	}

	result := infer.FormatDiagnostic("test.pl", source, d)
	assert.Contains(t, result, "type-mismatch")
	assert.Contains(t, result, "\n  hint: Add guard: if (ref($x)) { ... }")
}

func TestFormatDiagnosticWithoutSuggestion(t *testing.T) {
	source := []byte("push();\n")

	d := infer.Diagnostic{
		StartByte: 0,
		EndByte:   6,
		Severity:  infer.Error,
		Message:   "call to push: expected at least 2 argument(s), got 0",
		Code:      infer.CodeArityMismatch,
	}

	result := infer.FormatDiagnostic("test.pl", source, d)
	assert.NotContains(t, result, "hint:")
}
```

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestFormatDiagnosticWith -count=1 -v`
Expected: FAIL — FormatDiagnostic doesn't include hint yet

- [ ] **Step 2: Modify FormatDiagnostic**

In `internal/infer/diagnostics.go`, change `FormatDiagnostic`:

```go
// FormatDiagnostic converts byte offsets to 1-based line:col positions using
// the source text and returns a human-readable diagnostic string.
//
// Output format: filename:line:col: severity: message [code]
// When Suggestion is non-empty, a hint line is appended.
func FormatDiagnostic(filename string, source []byte, d Diagnostic) string {
	line, col := byteOffsetToLineCol(source, d.StartByte)
	result := fmt.Sprintf("%s:%d:%d: %s: %s [%s]", filename, line, col, d.Severity, d.Message, d.Code)
	if d.Suggestion != "" {
		result += "\n  hint: " + d.Suggestion
	}
	return result
}
```

- [ ] **Step 3: Run tests**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -run TestFormatDiagnostic -count=1 -v`
Expected: All PASS

- [ ] **Step 4: Run full infer test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/infer/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/infer/diagnostics.go internal/infer/diagnostics_test.go
git commit -m "$(cat <<'EOF'
feat(infer): extend FormatDiagnostic to show guard suggestions

When a Diagnostic has a non-empty Suggestion field, FormatDiagnostic
appends a hint line. Backward compatible — no hint when empty.
Part of #393.
EOF
)"
```

---

### Task 6: Add psc check Integration Test

Verify that `psc check` output includes the hint line when a type-mismatch diagnostic has a suggestion.

**Files:**
- Modify: `internal/psc/check_command_test.go`

- [ ] **Step 1: Write test**

In `internal/psc/check_command_test.go`, add:

```go
func TestCheckCommandWithTypeMismatchNoHint(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "mismatch.pl")
	// push($x, 1) where $x is a scalar — triggers type-mismatch.
	// No guard can narrow Scalar to Array, so no hint line appears.
	content := "my $x;\npush($x, 1);\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})

	var stdout strings.Builder
	var stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Error(t, err, "check should report diagnostics")
	assert.Contains(t, stderr.String(), "type-mismatch", "should contain type-mismatch diagnostic")
	assert.NotContains(t, stderr.String(), "hint:", "no hint when no guard helps")
}
```

- [ ] **Step 2: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/psc/ -run TestCheckCommandWithTypeMismatchNoHint -count=1 -v`
Expected: PASS (the wiring from Task 4 and formatting from Task 5 should make this work)

- [ ] **Step 3: Run full psc test suite**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/psc/ -count=1 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 4: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/psc/check_command_test.go
git commit -m "$(cat <<'EOF'
test(psc): verify psc check shows guard suggestion hint lines

End-to-end test that push($x, 1) with scalar $x produces a
type-mismatch diagnostic with a hint line. Part of #393.
EOF
)"
```

---

### Task 7: Add LSPServer Suggestion Test

Verify that `LSPServer.Diagnostics()` returns the `Suggestion` field populated for type-mismatch diagnostics.

**Files:**
- Modify: `internal/psc/lsp_inference_test.go`

- [ ] **Step 1: Write test**

In `internal/psc/lsp_inference_test.go`, add:

```go
func TestLSPDiagnosticSuggestionField(t *testing.T) {
	server := psc.NewLSPServer()
	// push($x, 1) where $x is Scalar triggers type-mismatch.
	// Verifies the Suggestion field exists on the Diagnostic struct
	// and is accessible via the LSP API. No guard helps here, so
	// Suggestion is empty.
	source := []byte("my $x;\npush($x, 1);\n")
	err := server.OpenDocument("file:///test.pl", source)
	require.NoError(t, err)

	diags := server.Diagnostics("file:///test.pl")
	require.True(t, len(diags) > 0, "should have diagnostics")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			// Scalar vs Array: empty suggestion is correct.
			assert.Empty(t, d.Suggestion)
			found = true
		}
	}
	assert.True(t, found, "should find type-mismatch diagnostic via LSP")
}
```

Note: The implementer should check the existing imports in `lsp_inference_test.go` and add `infer` if needed.

- [ ] **Step 2: Run test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go test ./internal/psc/ -run TestLSPDiagnosticSuggestionField -count=1 -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions
git add internal/psc/lsp_inference_test.go
git commit -m "$(cat <<'EOF'
test(psc): verify LSPServer exposes guard suggestions on diagnostics

Confirms that Diagnostics() returns populated Suggestion field for
type-mismatch diagnostics. Part of #393.
EOF
)"
```

---

### Task 8: Run Full Project Test Suite

Verify the entire project still passes with all changes.

**Files:** None (verification only)

- [ ] **Step 1: Run make test**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && make test`
Expected: All packages PASS, 0 failures

- [ ] **Step 2: If any failures, fix them before proceeding**

Any failures must be investigated and fixed. Do not skip.

- [ ] **Step 3: Verify with go vet**

Run: `cd /home/perigrin/dev/pvm/.worktrees/type-guard-suggestions && go vet ./...`
Expected: No issues
