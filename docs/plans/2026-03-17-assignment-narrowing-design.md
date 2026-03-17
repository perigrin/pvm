# Assignment-Based Type Narrowing Design

## Problem

PSC's type inference assigns every scalar variable the type `Scalar` based on
its sigil, regardless of what value is assigned to it. This means `my $x = 42`
gives `$x` the type `Scalar`, and `chr($x)` passes the type check because
`Scalar` satisfies `Int` polymorphically. The engine cannot distinguish
`my $x = 42` from `my $x = "hello"` — both are just `Scalar`.

This prevents the engine from catching value-level anti-patterns like
`my $s = "hello"; chr($s)` (chr expects Int, got Str) or
`my $n = 42; keys($n)` (keys expects Hash, got Int).

## Solution

Two changes:

1. **Assignment-based narrowing** — when the inference walk encounters an
   assignment, update the variable's type in the symbol table to match the
   RHS expression's type.

2. **Non-permissive Unknown** — change `TypeSatisfies(Unknown, X)` from `true`
   to `false`, matching TypeScript's `unknown` semantics. The engine becomes
   honest about what it cannot determine.

## Assignment Narrowing

### Where It Happens

Narrowing occurs in Pass 2 (the inference walk), not Pass 1. Pass 1 collects
declarations with sigil-derived types. Pass 2, during its bottom-up walk,
refines variable types when it encounters assignments.

The walk is post-order: by the time we visit an assignment node, the RHS
expression's type is already computed and in the annotation map.

### Algorithm

For each assignment node encountered during Pass 2:

1. Extract the RHS type from the annotation map
2. Extract the LHS variable name (with sigil)
3. Look up the variable in the symbol table
4. Update its type to the RHS type

For variable references, check the symbol table first for a refined type
before falling back to the sigil type.

### Reassignment

Each assignment updates the variable's type. After `$x = 42`, the type is
Int. After `$x = "hello"`, the type becomes Str. The type at any reference
point reflects the most recent assignment in source order.

This is sound for straight-line code. Branching (if/else assigning different
types) is not handled in this design — that requires the full flow narrowing
from the future PRD.

### Unknown RHS

When the RHS type is Unknown (unrecognized expression, user-defined function
call), the variable's type becomes Unknown. Combined with the non-permissive
Unknown change, this means subsequent uses of the variable produce diagnostics.

This is intentional: Unknown diagnostics expose coverage gaps in the inference
engine and serve as a forcing function for improving expression type coverage.

## Non-Permissive Unknown

### TypeScript Analogy

- **`any`** — disables type checking. Assignable to and from everything.
- **`unknown`** — safe top type. Anything can be assigned to `unknown`, but
  `unknown` cannot be used where a specific type is expected.

Our `Any` behaves like TypeScript's `any` (polymorphic, passes everything).
Our `Unknown` currently behaves like `any` too. This change makes `Unknown`
behave like TypeScript's `unknown`.

### Implementation

In `TypeSatisfies`, remove the permissive Unknown check:

```go
// Before:
if actual == Unknown {
    return true  // permissive
}

// After: remove this block entirely
// Unknown falls through to the normal checks and returns false
```

### Impact

- Unresolved expressions used as builtin arguments produce diagnostics
- Variables assigned from unresolved expressions have type Unknown, and
  using them where a specific type is expected produces diagnostics
- This is strict by default — a future PR will add a `--permissive` flag
  for projects that want to suppress Unknown-related diagnostics

## What This Enables

| Pattern | Before | After |
|---|---|---|
| `my $x = 42; chr($x)` | No diagnostic | No diagnostic ($x is Int, chr wants Int) |
| `my $s = "hello"; chr($s)` | No diagnostic | **Diagnostic** ($s is Str, chr wants Int) |
| `my $n = 42; keys($n)` | No diagnostic | **Diagnostic** ($n is Int, keys wants Hash) |
| `my $x = foo(); chr($x)` | No diagnostic | **Diagnostic** ($x is Unknown, strict) |

The last case is the accepted noise: unknown function returns produce
diagnostics. Each such diagnostic is a TODO for improving engine coverage.

## Changes Required

### internal/types/types.go
- Remove the `actual == Unknown` permissive check from `TypeSatisfies`

### internal/infer/infer.go
- In the inference walk, detect assignment nodes
- Extract RHS type from annotation map
- Update variable type in symbol table via `SymbolTable.UpdateType()`
- For variable reference nodes, consult symbol table for refined type

### internal/infer/symbols.go
- Add `UpdateType(name string, typ types.Type)` method to SymbolTable
- Updates the type of an existing symbol in the current scope chain

### Test updates
- Invert `TestTypeSatisfiesUnknown` (Unknown no longer passes permissively)
- Add tests for assignment narrowing: `my $x = 42` → $x is Int
- Add tests for value-level anti-patterns: `my $s = "hello"; chr($s)`
- Update existing tests that relied on Unknown being permissive
- Add E2E test data for value-level diagnostics

## Future Work

- **Strictness toggle** — `psc check --permissive` flag to restore
  Unknown-as-permissive behavior for projects with many unresolved types
- **Branch-aware narrowing** — track types through if/else branches,
  compute union types at join points
- **Expression coverage** — reduce Unknown noise by handling more
  expression types: string literals, method calls, anonymous subs,
  ternary expressions
