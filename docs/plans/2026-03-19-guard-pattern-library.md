# Guard Pattern Library

## Problem

PSC recognizes `defined($x)`, `ref($x)`, and `$x isa Foo` as type guards but
ignores common Perl guard functions from `builtin::` (5.36+). Code like
`if (builtin::blessed($x))` should narrow `$x` to Object, and
`if (builtin::is_bool($x))` should narrow to Bool.

## Approach

Add a `function_call_expression` branch to `extractGuardPattern` that matches
known guard function names against a lookup table. Both bare (`blessed`) and
fully qualified (`builtin::blessed`) names are recognized.

### CST Shape (verified by discovery)

All guard library functions share the same CST structure:

```
function_call_expression
  function "builtin::blessed"     // or bare "blessed"
  (
  scalar "$x"
  )
```

Negation (`!builtin::blessed($x)`) wraps in `unary_expression`, handled by
existing `extractNegatedGuard`. Compound guards work via existing machinery.

### Guard Function Table

| Function | FQ Name | Guard Kind | Narrows To |
|----------|---------|-----------|-----------|
| `blessed` | `builtin::blessed` | GuardIsa (reuse) | Object |
| `reftype` | `builtin::reftype` | GuardRef (reuse) | Ref mask |
| `is_bool` | `builtin::is_bool` | GuardBool (new) | Bool |

### New Guard Kind: GuardBool

Added to `types/narrowing.go`:
- Positive: `typ & Bool` (intersection with Bool bit)
- Negated: `typ &^ Bool` (remove Bool bit)
- Empty set → None (unreachable)

### Implementation

`extractFunctionCallGuard(node, source)`:
1. Find the `function` child, get its text
2. Strip package prefix to get the base name
3. Look up base name in `guardFunctionTable`
4. Find the `scalar` child (guarded variable)
5. Return `guardResult` with the appropriate guard pattern

## Out of Scope

- `$x->isa('Foo')` method call (blocked by string grammar bug)
- Scalar::Util functions (deferred — can be added to the table later)
- Params::Util functions
- User-configurable guard function table
