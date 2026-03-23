# PSC Inference Expansion

Date: 2026-03-23

## Problem

PSC's inference engine detects arity mismatches and container-level type
mismatches (e.g., push on a scalar) but does not detect value-level type
errors. A string used in arithmetic and an uninitialized variable used
where a defined value is required are invisible to the current engine.

Several builtin signatures in the type library are overly permissive
(e.g., `print` accepts `Any` instead of `Str`), which weakens both
existing diagnostics and the new operand checking.

## Scope

Four acceptance criteria from the milestone issue:

1. Non-numeric string in numeric context produces type-mismatch diagnostic
2. Undef used where defined value required produces diagnostic (with guard suggestion)
3. Cross-file return type inference works on multi-file test case
4. At least one user-defined function gets inferred parameter types

### What already exists

The pushback review found that significant infrastructure already exists:

- **Return type collection**: `collectExplicitReturns`, `collectImplicitReturn`,
  and `walkSubroutineDeclaration` already infer return types from explicit
  and implicit returns. Tests pass.
- **Cross-file call resolution**: `inferFunctionCallType` already looks up
  callee return types via `ProjectIndex.LookupSymbol`. Tests in
  `project_test.go` verify cross-file return type propagation.
- **Parameter identification**: `ResolveSubParam` already handles signatures,
  shift, and `$_[0]` patterns for single-parameter functions.

Acceptance criterion 3 is already met by the existing infrastructure.
This spec focuses on:
- Section 1: Binary operator operand checking (new)
- Section 2: Undef propagation (new)
- Section 3: Verify existing cross-file inference meets AC3 (validation)
- Section 4: Extend `ResolveSubParam` to multi-param, add usage-based
  type inference (extension of existing code)
- Signature audit: Fix incorrect builtin signatures (prerequisite)

## 0. Builtin Signature Audit

Audit all entries in `internal/types/signatures.go` against perldoc.
Known incorrect signatures:

- `print`: accepts `Any` but should accept `Str` (stringifies args)
- `say`: same as `print`

Audit all builtins, binary ops, and unary ops. Fix any that accept
overly permissive types. This is a prerequisite because the new operand
checking and parameter inference depend on correct signatures to produce
accurate diagnostics and constraints.

Key files: internal/types/signatures.go

## 1. Binary Operator Operand Type Checking

Currently `inferBinaryExprType` returns the result type but does not check
whether operands match the operator's expected types. Add operand checking.

### Arithmetic operators (+, -, *, /, %, **)

Check each operand's inferred type against the signature's expected Num.
Emit `coercion-mismatch` diagnostic when:
- An operand is Str (strLeaf, not Num) -- covers `"hello" + 1`
- An operand is NaN -- covers NaN in arithmetic
- An operand is Inf -- covers Inf in arithmetic
- An operand is Undef -- covers `my $x; $x + 1`

### String operators (., x)

Check operands against expected Str/Int. Emit diagnostic when Ref types
appear in string concatenation (stringifying a reference is almost never
intended).

### Comparison operators (== vs eq)

Detect when `==` is used on Str operands (suggest `eq`). Str is NOT a
subtype of Num, so this is a real type error. Emit as `coercion-mismatch`
at Warning severity because the code executes (Perl coerces) but the
result is almost certainly not what was intended.

Using `eq` on Int/Num operands is type-safe (Int <: Str) so no warning
is needed for that direction.

### New diagnostic code

`CodeCoercionMismatch = "coercion-mismatch"` to distinguish from existing
`CodeTypeMismatch` which fires on builtin argument types. Coercion mismatches
are about implicit type conversion at operator boundaries.

### Implementation

Modify `inferBinaryExprType` to accept the annotations map and diags slice.
The function is called from four node-kind cases in `inferNodeType`
(binary_expression, equality_expression, relational_expression,
lowprec_logical_expression) -- all four call sites need updating.

After looking up the operator signature, check each operand's annotated type
against the signature's Left/Right types using `TypeSatisfies`. Skip
operands with Unknown type (avoid false positives). Emit diagnostics for
mismatches.

Note: NaN and Inf operand checking is structurally ready but untestable
in this iteration because `inferNumberType` does not yet produce NaN or
Inf types. NaN/Inf literal inference is tracked as a follow-up.

Key files: internal/infer/infer.go, internal/infer/diagnostics.go

## 2. Undef Propagation

### Uninitialized variables

Change `collectVariableDeclaration` in pass 1 so that `my $x;` (no
initializer) is typed as `Undef` instead of the sigil type (Scalar).
Distinguish by checking whether the `variable_declaration` node's parent
is an `assignment_expression` -- if so, keep Scalar (pass 2 will narrow
it); if not, set Undef. The existing assignment narrowing already upgrades
the type on first assignment:

    my $x;         # type: Undef (parent is NOT assignment_expression)
    my $y = 42;    # type: Scalar in pass 1, narrowed to Int in pass 2
    $x = 42;       # type: Int (narrowed by assignment in pass 2)

### Undef in operators

When a binary operator or builtin receives an `Undef`-typed operand where
a defined value is expected, emit a `coercion-mismatch` diagnostic at
severity Error (same as other type mismatches -- Perl warns at runtime,
PSC treats it as an error). Include a `defined()` guard suggestion via
the existing `SuggestGuard` function.

Operands with Unknown type (no type information available) are silently
skipped -- no diagnostic. This avoids false positives when type info is
missing (e.g., from unresolvable cross-file calls).

### Guard narrowing

The existing `defined()` guard narrowing already handles the positive case:
after `if (defined($x))`, `$x` loses the Undef bit. No changes needed to
the narrowing system.

Key files: internal/infer/symbols.go, internal/infer/infer.go

## 3. Cross-File Return Type Inference (Validation)

This infrastructure already exists. The work here is to verify the
existing implementation meets acceptance criterion 3 and add any
missing E2E test coverage.

### What exists

- `collectExplicitReturns` (infer.go:1690) collects explicit return types
- `collectImplicitReturn` (infer.go:1733) infers from last expression
- `walkSubroutineDeclaration` (infer.go:1811) stores ReturnType
- `inferFunctionCallType` (infer.go:323) uses ReturnType for call sites
- `ProjectIndex.LookupSymbol` resolves cross-file symbols
- Tests in `project_test.go` verify cross-file propagation

### What to verify

- Existing multi-file tests cover the AC3 scenario
- If not, add an E2E test with a module that exports a function with
  a typed return, and a caller that uses the return value

### Return type precision

Array and Hash annotations are more precise than List. A function that
returns `@arr` has return type Array (not List), because Array satisfies
List via subtyping (`Array <: List`). No coercion needed -- the existing
behavior is correct and preserves precision.

Known limitation: cross-file resolution is single-hop. When AnalyzeFile
processes a dependency, it passes nil for the ProjectIndex, so the
dependency's own imports are not resolved. Multi-hop resolution is a
follow-up.

Key files: test/e2e/psc_check_test.go (verification only)

## 4. User-Defined Function Parameter Inference

### Extend parameter identification to multi-param

The existing `ResolveSubParam` handles single-parameter functions.
Extend it (or add a new `ResolveSubParams` function) to return all
parameter variables for multi-param functions.

Support three patterns:

1. **Signatures**: `sub foo($x, $y)` -- extend `extractFirstSigParam`
   to return all params, not just the first. Remove the `len(params) == 1`
   guard.

2. **@_ unpacking**: `my ($x, $y) = @_;` -- new pattern. Match an
   `assignment_expression` at the top of the sub body where the RHS is
   `@_` and the LHS is a list of variable declarations. Extract all
   variable names from the LHS in positional order. This requires
   detecting the CST shape for list assignment.

3. **shift chains**: `my $self = shift; my $x = shift;` -- extend
   `extractShiftParam` to collect all sequential `my $var = shift;`
   statements at the top of the body, not just the first.

### Type inference from usage

Once parameter variables are identified, trace each through the function
body by examining how each variable is used:

- Operand of `+`, `-`, etc. -> Num
- Passed to `push` as first arg -> Array
- Passed to `keys` -> Hash
- Used in string concatenation -> Str
- Passed to `print`/`say` -> Str (after signature fix)

Collect all usage constraints and narrow progressively: start with Any
(no information) and for each usage, intersect with the constraint type.
If `$x` is used in `$x + 1` (Num) and `push(@arr, $x)` (Any for arg 1),
the result is Num (Any & Num = Num, since Num is more specific).

If the inferred type is still Any after all constraints, do NOT store it.
Leave as Unknown to signal "no useful inference was possible." This
distinguishes "we analyzed the function and learned nothing" from
"we analyzed it and determined the parameter accepts Num." A future
strict mode can treat remaining Unknowns as validation gaps.

Store inferred parameter types as `[]types.Type` (positional, parallel
to the parameter names list) in the SymbolTable entry for the sub.
Add a `ParamTypes []types.Type` field to the `Symbol` struct.

For this issue, parameter types are stored but NOT used for call-site
validation. Call-site checking is a follow-up.

Key files: internal/infer/infer.go, internal/infer/symbols.go

## Test Plan

### E2E tests (test/e2e/testdata/check/)

New test data files:

- `coercion.pl` -- string in numeric context, Ref in string context,
  `==` on strings
- `undef_propagation.pl` -- uninitialized variable in arithmetic and
  string context, guard narrowing eliminates the diagnostic
- Verify existing multi-file test covers AC3, add if missing

### Unit tests

- `internal/infer/infer_test.go` -- operand type checking for binary ops
- `internal/infer/symbols_test.go` -- ParamTypes storage
- `internal/infer/diagnostics_test.go` -- coercion-mismatch formatting
- `internal/types/signatures_test.go` -- audit results (verify print
  expects Str, etc.)
- Parameter inference: test all three patterns (signature, @_ unpacking,
  shift chains) with multi-param functions

## Follow-Up Issues

- #410: Implicit return type inference (last expression of function body)
  -- NOTE: this already exists as `collectImplicitReturn`, so #410 may
  already be satisfied. Verify before closing.
- Call-site parameter type validation using inferred parameter types
- NaN/Inf inference from literal patterns ("NaN" + 0, 9e999)
- Multi-hop cross-file resolution (pass ProjectIndex to AnalyzeFile)
- Strict mode: flag remaining Unknown types as validation gaps

## Files to Modify

- `internal/types/signatures.go` -- fix print/say and any other incorrect
  signatures found in audit
- `internal/infer/infer.go` -- operand checking, extend parameter inference
- `internal/infer/symbols.go` -- add ParamTypes field to Symbol struct
- `internal/infer/diagnostics.go` -- add CodeCoercionMismatch constant
- `test/e2e/testdata/check/` -- new test data files
- `test/e2e/psc_check_test.go` -- new E2E test cases

## What NOT to Change

- `internal/types/types.go` -- no type system changes (done in issue 3)
- Existing diagnostic behavior -- only add, never change existing diagnostics
- Existing return type infrastructure -- already correct, only verify
