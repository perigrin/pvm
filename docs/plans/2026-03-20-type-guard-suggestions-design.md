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

`suggestGuard` tries each guard candidate in priority order. A guard is only
suggested if applying its narrowing to actual would produce a type that
satisfies expected (checked via `types.TypeSatisfies`). This prevents
suggesting `defined()` when the real problem is structural (e.g., Scalar vs
Array). First match wins:

| Priority | Guard | Narrowing |
|----------|-------|-----------|
| 1 | `defined($var)` | `actual &^ Undef` |
| 2 | `ref($var)` | `actual & Ref` |
| 3 | `builtin::is_bool($var)` | `actual & Bool` |
| — | No suggestion | — |

Each guard is only suggested when `TypeSatisfies(narrowedType, expected)`
returns true. This means suggestions fire when the actual type is close to
expected and a single guard resolves the mismatch.

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
file.pl:3:5: error: call to bless: argument 1 expects Ref, got Undef [type-mismatch]
  hint: Add guard: if (defined($x)) { ... }
```

No hint line when Suggestion is empty — fully backward compatible.

### Scope

- Built-in guards only (defined, ref, isa, builtin::is_bool)
- type-mismatch diagnostics only
- No user-defined guard awareness (they wrap built-ins anyway)
- No full code rewrite — guard name + wrapping if-snippet

## Testing

### Unit Tests for suggestGuard

- Scalar vs Int → `defined($x)` (Undef removal leaves union containing Int)
- Scalar vs Ref → `defined($x)` (Undef removal leaves union containing Ref)
- Scalar vs Object → `defined($x)` (Undef removal leaves union containing Object)
- Scalar vs Bool → `defined($x)` (Undef removal leaves union containing Bool)
- Ref|Int vs Ref → `ref($x)` (no Undef, ref guard fires directly)
- Bool|Int vs Bool → `builtin::is_bool($x)` (no Undef, is_bool fires directly)
- Scalar vs Array → no suggestion (structural mismatch)
- Array vs Hash → no suggestion
- Empty var name → no suggestion

### Integration Tests

- `push($x, 1)` where `$x` is Scalar → diagnostic with no suggestion (structural)
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
