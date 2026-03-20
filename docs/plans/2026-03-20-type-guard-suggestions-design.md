# Type Guard Suggestions Design

## Goal

When PSC emits a type-mismatch diagnostic (e.g., "expected Int, got Scalar"),
attach a suggestion for the appropriate guard expression that would narrow the
variable's type to satisfy the callee. Suggestions appear as hint lines in
`psc check` output and are available programmatically via the `Diagnostic`
struct's `Suggestion` field.

## Architecture

Inline in the Analyze walker. At the two sites where type-mismatch diagnostics
are emitted (builtin call argument validation), add a `Suggestion` field
populated by a new `suggestGuard(varName, actualType, expectedType)` function.
No new pass, no new types.

### Diagnostic Struct Change

Add a `Suggestion string` field to `infer.Diagnostic`. Empty when no guard
applies. Backward compatible — existing consumers ignore unknown fields.

### Guard Selection Rules

`suggestGuard` examines the bitset difference between actual and expected types
to pick the right guard. First match wins:

| Priority | Condition | Suggestion |
|----------|-----------|------------|
| 1 | `actual` has Undef, `expected` does not | `if (defined($var)) { ... }` |
| 2 | `expected` within Ref mask, `actual` broader | `if (ref($var)) { ... }` |
| 3 | `expected` is Object | `if (ref($var)) { ... }` |
| 4 | `expected` is Bool, `actual` broader | `if (builtin::is_bool($var)) { ... }` |
| — | No clear guard maps | No suggestion |

When multiple bits differ (e.g., Scalar includes Undef+Bool+Int+Num+Str), the
function picks the single most useful guard. No compound guard suggestions.

### Variable Name Extraction

The argument CST node is already available at the diagnostic site. Extract the
variable name using existing helpers (sigildName for scalar/array/hash nodes).
Non-variable arguments (e.g., literal expressions) produce no suggestion.

### Integration Points

Two sites in `infer.go` emit `CodeTypeMismatch`:

1. Builtin function call argument validation (~line 356)
2. Method/function call argument validation (~line 511)

Both follow the same pattern: iterate args, check TypeSatisfies, emit
diagnostic. The suggestion is added as one extra field on the Diagnostic
literal.

### Output Format

`FormatDiagnostic` appends a hint line when `Suggestion` is non-empty:

```
file.pl:3:5: error: call to push: argument 1 expects Array, got Scalar [type-mismatch]
  hint: Add guard: if (ref($x)) { ... }
```

No hint line when Suggestion is empty — fully backward compatible.

### Scope

- Built-in guards only (defined, ref, isa, builtin::is_bool)
- type-mismatch diagnostics only
- No user-defined guard awareness (they wrap built-ins anyway)
- No full code rewrite — guard name + wrapping if-snippet

## Testing

### Unit Tests for suggestGuard

- Scalar vs Int → `defined($x)` (Undef is the extra bit)
- Scalar vs Ref → `ref($x)`
- Scalar vs Object → `ref($x)`
- Scalar vs Bool → `builtin::is_bool($x)`
- Int vs Str → no suggestion (structural mismatch)
- Array vs Hash → no suggestion
- Empty var name → no suggestion

### Integration Tests

- `push($x, 1)` where `$x` is Scalar → diagnostic with ref suggestion
- Guard already in place → no diagnostic
- Types match → no diagnostic

### FormatDiagnostic Tests

- Diagnostic with suggestion → hint line present
- Diagnostic without suggestion → no hint line

### LSPServer Tests

- Document with type mismatch → Diagnostics() returns Suggestion field

## Depends On

- Existing type inference (Analyze, annotations, diagnostics)
- Existing guard pattern types (GuardDefined, GuardRef, GuardIsa, GuardBool)
- User-Defined Type Guard Functions (#392) — completed

## References

- Issue: #393
- Design predecessor: #392 User-Defined Type Guard Functions
