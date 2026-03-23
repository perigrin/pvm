# PSC Inference Expansion

Date: 2026-03-23

## Problem

PSC's inference engine detects arity mismatches and container-level type
mismatches (e.g., push on a scalar) but does not detect value-level type
errors. A string used in arithmetic, an uninitialized variable used where
a defined value is required, and cross-file return types are all invisible
to the current engine.

## Scope

Four acceptance criteria from the milestone issue:

1. Non-numeric string in numeric context produces type-mismatch diagnostic
2. Undef used where defined value required produces diagnostic (with guard suggestion)
3. Cross-file return type inference works on multi-file test case
4. At least one user-defined function gets inferred parameter types

## 1. Binary Operator Operand Type Checking

Currently `inferBinaryExprType` returns the result type but does not check
whether operands match the operator's expected types. Add operand checking.

### Arithmetic operators (+, -, *, /, %, **)

Check each operand's inferred type against the signature's expected Num.
Emit `coercion-mismatch` diagnostic when:
- An operand is Str (strLeaf, not Num) — covers `"hello" + 1`
- An operand is NaN — covers NaN in arithmetic
- An operand is Inf — covers Inf in arithmetic
- An operand is Undef — covers `my $x; $x + 1`

### String operators (., x)

Check operands against expected Str/Int. Emit diagnostic when Ref types
appear in string concatenation (stringifying a reference is almost never
intended).

### Comparison operators (== vs eq)

Detect when `==` is used on Str operands (suggest `eq`) or `eq` is used
on Num/Int operands (suggest `==`). These are `coercion-mismatch` severity
Warning, not Error, because the code works — it's just surprising.

### New diagnostic code

`CodeCoercionMismatch = "coercion-mismatch"` to distinguish from existing
`CodeTypeMismatch` which fires on builtin argument types. Coercion mismatches
are about implicit type conversion at operator boundaries.

### Implementation

Modify `inferBinaryExprType` to accept the annotations map and diags slice.
After looking up the operator signature, check each operand's annotated type
against the signature's Left/Right types using `TypeSatisfies`. Emit
diagnostics for mismatches.

Key files: internal/infer/infer.go, internal/infer/diagnostics.go

## 2. Undef Propagation

### Uninitialized variables

Change `CollectDeclarations` so that `my $x;` (no initializer) is typed
as `Undef` instead of the sigil type (Scalar). The existing assignment
narrowing already upgrades the type on first assignment:

    my $x;         # type: Undef
    $x = 42;       # type: Int (narrowed by assignment)

### Undef in operators

When a binary operator or builtin receives an `Undef`-typed operand where
a defined value is expected, emit a diagnostic. Include a `defined()` guard
suggestion via the existing `SuggestGuard` function.

### Guard narrowing

The existing `defined()` guard narrowing already handles the positive case:
after `if (defined($x))`, `$x` loses the Undef bit. No changes needed to
the narrowing system.

Key files: internal/infer/symbols.go, internal/infer/infer.go

## 3. Cross-File Return Type Inference

### Return type collection

When walking `subroutine_declaration_statement`, collect the types of all
explicit `return` arguments encountered in the function body. Union them
to produce the function's return type.

Rules:
- `return 42` → Int
- `return @arr` → List (not Array — returns a list of values)
- `return %hash` → List (not Hash — returns a list of key-value pairs)
- `return` (bare) → Undef in scalar context
- Multiple returns → union of all return types

Store the inferred return type in the SymbolTable entry for the sub.

### Cross-file call resolution

When `inferFunctionCallType` resolves a cross-file call via ProjectIndex,
look up the callee's return type from its SymbolTable and use it as the
call expression's inferred type (instead of the current `types.Unknown`
fallback).

Key files: internal/infer/infer.go, internal/infer/symbols.go,
internal/infer/project.go

## 4. User-Defined Function Parameter Inference

### Parameter identification

Support three patterns for identifying parameter variables in a function body:

1. **Signatures**: `sub foo($x, $y)` — parameters are the formal_parameter
   nodes in the CST.

2. **@_ unpacking**: `my ($x, $y) = @_;` — the first assignment statement
   in the sub body where the RHS is `@_`. Variables on the LHS are parameters
   in positional order.

3. **shift chains**: `my $self = shift; my $x = shift;` — sequential
   `my $var = shift;` statements at the top of the sub body. Each shift
   produces the next positional parameter.

### Type inference from usage

Once parameter variables are identified, trace each through the function body
by examining how each variable is used:

- Operand of `+`, `-`, etc. → Num
- Passed to `push` as first arg → Array
- Passed to `keys` → Hash
- Used in string concatenation → Str
- Used in `defined()` check → may be Undef

Collect all usage constraints and intersect to produce the inferred
parameter type. Store in the SymbolTable entry for the sub.

For this issue, parameter types are stored but NOT used for call-site
validation. Call-site checking is a follow-up.

Key files: internal/infer/infer.go, internal/infer/symbols.go

## Test Plan

### E2E tests (test/e2e/testdata/check/)

New test data files:

- `coercion.pl` — string in numeric context, NaN/Inf in arithmetic,
  Ref in string context, == on strings, eq on numbers
- `undef_propagation.pl` — uninitialized variable in arithmetic and
  string context, guard narrowing eliminates the diagnostic
- Multi-file test: module with typed return values, caller that uses them

### Unit tests

- `internal/infer/infer_test.go` — operand type checking for binary ops
- `internal/infer/symbols_test.go` — return type and parameter type storage
- `internal/infer/diagnostics_test.go` — coercion-mismatch formatting

## Follow-Up Issues

- Implicit return type inference (last expression of function body)
- Call-site parameter type validation using inferred parameter types
- NaN/Inf inference from literal patterns ("NaN" + 0, 9e999)

## Files to Modify

- `internal/infer/infer.go` — operand checking, return type collection,
  parameter inference
- `internal/infer/symbols.go` — add ReturnType and ParamTypes to sub entries
- `internal/infer/diagnostics.go` — add CodeCoercionMismatch constant
- `test/e2e/testdata/check/` — new test data files
- `test/e2e/psc_check_test.go` — new E2E test cases

## What NOT to Change

- `internal/types/types.go` — no type system changes (done in issue 3)
- `internal/types/signatures.go` — operator signatures are already correct
- Existing diagnostic behavior — only add, never change existing diagnostics
