# User-Defined Type Guard Functions

GitHub Issue: #392

## Problem

PSC recognizes built-in guard patterns (`defined($x)`, `ref($x) eq 'HASH'`,
`$x isa Foo`, etc.) when they appear directly in conditions. But when a user
wraps a guard pattern in a subroutine:

```perl
sub is_hashref { return ref($_[0]) eq 'HASH' }
if (is_hashref($x)) {
    # $x should narrow to HashRef here, but doesn't
}
```

PSC cannot see through the function call to recognize the guard.

## Solution

Add a fallback path in `extractFunctionCallGuard` that resolves user-defined
single-argument subs. When a function call in a condition doesn't match the
known guard function table, look up the sub in the symbol table, analyze its
body for a recognized guard pattern, and remap the result to the call-site
argument.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Sub arity | Single-argument only | Covers the common case; multi-arg adds complexity for little gain |
| Parameter styles | `$_[0]`, signatures, `shift` | All three are common Perl idioms for single-arg subs |
| File scope | Same-file only | Simpler first increment; cross-file via ProjectIndex can follow |
| Analysis timing | Lazy (on call-site encounter) | Avoids analyzing subs never used as guards |
| Caching | None | Few call sites per guard func in same-file scope; cost is negligible |
| Recursion | 1-level limit | Guard func calling another guard func is too exotic for now |

## Architecture

No new types, no new fields on `Symbol`, no new data structures. The change is
a new code path in the existing `extractGuardPattern` dispatch.

### Integration Point

In `extractFunctionCallGuard` (or called from the same dispatch):

1. Existing path: check known guard function table -> if hit, return
2. **New path**: extract function name, look up in symbol table as `SymbolSub`
3. Find the sub's AST node via `StartByte`/`EndByte`
4. Call `extractUserGuardFunc(node, source)` (new helper)
5. Extract call-site argument (must be a plain scalar variable)
6. Remap guard result's `VarName` from parameter name to call-site argument

### Parameter Variable Resolution

Three styles, checked in order:

**Style 1 — `$_[0]`:** The guard body directly references `$_[0]`. Look for
`array_access_expression` on `@_` with index `0`. The param name is literally
`$_[0]`.

**Style 2 — Signature:** The sub has a `signature` node. Extract the first
parameter name. The guard body references that name.

**Style 3 — `shift`:** The first statement is `my $var = shift` (or
`my $var = shift @_`). Treat `$var` as the parameter name.

If none match, bail out — the sub isn't recognizable as a guard.

### Return Expression Extraction

**Explicit return:** Walk the body for a `return_expression` node. If there are
multiple return statements, bail out (too complex).

**Implicit return:** If no return found, take the last expression statement in
the body.

**Statement count heuristic:** A guard function body should be: optionally a
`my $x = shift` statement, followed by one return expression (explicit or
implicit). Bodies with more complexity are skipped.

### Call-Site Argument Handling

Only handle plain scalar variable arguments (`$x`). Complex expressions
(`$hash{key}`, function calls) are not narrowable — bail out gracefully.

## Examples

```perl
# $_[0] style
sub is_hashref { ref($_[0]) eq 'HASH' }
if (is_hashref($x)) { ... }  # $x narrows to HashRef

# Signature style
sub is_defined($val) { defined($val) }
if (is_defined($y)) { ... }  # $y: Undef removed

# Shift style
sub is_foo { my $obj = shift; $obj isa Foo }
if (is_foo($z)) { ... }  # $z narrows to Object

# Compound guard in body
sub is_defined_ref { defined($_[0]) && ref($_[0]) }
if (is_defined_ref($x)) { ... }  # $x: Undef removed AND narrowed to Ref

# Negated at call site
unless (is_hashref($x)) { ... }  # else-branch narrows to HashRef
```

## Test Plan

1. `$_[0]` style guard — `ref($_[0]) eq 'HASH'` narrows to HashRef
2. Signature style guard — `ref($val) eq 'HASH'` same narrowing
3. Shift style guard — `my $val = shift; ref($val) eq 'HASH'` same narrowing
4. Defined guard function — `defined($_[0])` removes Undef
5. Isa guard function — `$_[0] isa Foo` narrows to Object
6. Negated call site — `unless` / `if (!...)` narrows in else-branch
7. Compound guard in body — `defined($_[0]) && ref($_[0])` both apply
8. Non-guard sub ignored — no false positives
9. Multi-arg sub ignored — bail out gracefully
10. Complex argument ignored — `is_hashref($hash{key})` no narrowing

## Dependencies

- Guard Pattern Library (done — `extractGuardPattern` and friends)
- Method Return Type Inference (done — sub body analysis infrastructure)

## Future Work

- Cross-file guard resolution via ProjectIndex
- Multi-arg guard functions with designated guard parameter
- Guard functions with multiple return paths
- Caching if performance becomes a concern
