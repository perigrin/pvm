# Guard Pattern Library

## Problem

PSC recognizes `defined($x)`, `ref($x)`, and `$x isa Foo` as type guards but
ignores common Perl guard functions from Scalar::Util. Code like
`if (blessed($x))` should narrow `$x` to Object, and
`if (looks_like_number($x))` should narrow to Num.

## Approach

Add a `function_call_expression` branch to `extractGuardPattern` that matches
known guard function names against a lookup table. Both bare (`blessed`) and
fully qualified (`Scalar::Util::blessed`) names are recognized.

### CST Shape (verified by discovery)

All guard library functions share the same CST structure:

```
function_call_expression
  function "blessed"              // or "Scalar::Util::blessed"
  (
  scalar "$x"
  )
```

Negation (`!blessed($x)`) wraps in `unary_expression`, which the existing
`extractNegatedGuard` already handles. Compound guards
(`blessed($x) && looks_like_number($y)`) work via the existing compound
guard extraction.

### Guard Function Table

| Function | FQ Name | Guard Kind | Narrows To |
|----------|---------|-----------|-----------|
| `looks_like_number` | `Scalar::Util::looks_like_number` | GuardNumeric | Num mask |
| `blessed` | `Scalar::Util::blessed` | GuardBlessed | Object |
| `reftype` | `Scalar::Util::reftype` | GuardRef (plain) | Ref mask |

### New Guard Kind

Add `GuardNumeric` to `types/narrowing.go`:
- Positive: `typ & Num` (intersection with Num mask — keeps Int and Num bits)
- Negated: `typ &^ Num` (remove all numeric bits)

`blessed` reuses `GuardIsa` (narrows to Object, same semantics).
`reftype` reuses `GuardRef` (narrows to Ref mask, same as `ref()`).

### Extraction

New function `extractFunctionCallGuard(node, source)`:
1. Find the `function` child, get its text
2. Strip any package prefix to get the base name
3. Look up base name in the guard function table
4. Find the `scalar` child (the guarded variable)
5. Return `guardResult` with the appropriate guard pattern

Add to `extractGuardPattern`:
```go
if kind == "function_call_expression" {
    return extractFunctionCallGuard(node, source)
}
```

### What Changes

**types/narrowing.go:** Add `GuardNumeric` constant and handling in
`NarrowByGuard` / `NegateGuard`. Positive: `typ & Num`. Negative: `typ &^ Num`.

**infer/infer.go:** Add `extractFunctionCallGuard` and the guard function
lookup table. Add `function_call_expression` case to `extractGuardPattern`.

**No changes to:** `walkBlockWithGuard`, `walkConditionalStatement`, or any
other walker function. The guard extraction feeds into existing machinery.

## Out of Scope

- `$x->isa('Foo')` method call (blocked by string grammar bug)
- Params::Util functions
- User-configurable guard function table
- `reftype($x) eq 'HASH'` (blocked by string grammar bug)

## Implementation Order

1. Add `GuardNumeric` to narrowing.go + tests
2. Add `extractFunctionCallGuard` to infer.go + tests
3. Verify negation and compound interactions work for free
