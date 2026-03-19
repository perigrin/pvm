# Bitset Types, Compound Guards, and Branch Merging

## Problem

PSC's type system represents types as integer enums. This prevents union types
(`Scalar | Ref`), makes intersection impossible, and forces join points after
if/else to discard narrowing information. Three features need this fixed:

1. After `if (ref($x)) { ... } else { ... }`, the post-if type of `$x` should
   reflect both branches, not just revert to the pre-if type.
2. `if (defined($x) && ref($x))` should apply both guards — intersection.
3. `if (defined($x) || ref($x))` should negate both guards in the else block.

All three require types that can express "this value is one of these types."

## Approach: Bitset Encoding

Replace `type Type int` with `type Type uint32` where each leaf type gets a
unique bit. Parent types are bitmasks — the OR of their descendant leaf bits.
Unions are native OR operations. Intersections are native AND.

### Leaf-Only Bit Assignment

Only leaf types (types with no children in the hierarchy) get their own bit.
Parent types like `Scalar`, `Ref`, `List`, and `Any` are derived masks — the
OR of all their descendant leaf bits. A concrete value of type `Num` holds only
the `Num` bit. A variable declared as `my $x` (type `Scalar`) holds the full
`Scalar` mask, meaning it could be any scalar subtype.

This distinction matters: a single bit means "this value IS this type." A
multi-bit mask means "this value COULD BE any of these types." The bitset
representation unifies both — `Num` is just a mask that happens to have one bit.

```
Bit  Type        Notes
---  ----        -----
0    Undef
1    Bool
2    Int
3    Num         Leaf: 3.14 is Num but not Int
4    Str         Leaf: "hello" is Str but not Num
5    DualVar
6    Regex
7    ScalarRef
8    ArrayRef
9    HashRef
10   CodeRef
11   GlobRef
12   Object
13   Array
14   Hash
15   Code
16   Glob

Derived parent masks (OR of descendant bits):
  Num     includes Int (because Int <: Num), so: NumBit | IntBit
  Str     includes Num, Int: StrBit | NumBit | IntBit
  Ref     = ScalarRef | ArrayRef | HashRef | CodeRef | GlobRef | Object
  Scalar  = Undef | Bool | Str | Num | Int | DualVar | Regex | Ref (all bits)
  List    = Array | Hash
  Any     = Scalar | List | Code | Glob
```

Wait — `Num` and `Str` are NOT leaf types in the hierarchy (Int <: Num <: Str).
But they DO need their own bits because a value can be concretely `Num` (3.14)
without being `Int`. So every type that a concrete value can inhabit gets a bit,
even if it has subtypes. The "parent masks" are the convenience aliases that
include all descendant bits for subtype checking.

To clarify the dual role:
- `Num` as a **concrete type** (the result of `3.14`) = just bit 3
- `Num` as a **type constraint** (accepts Int too) = bit 3 | bit 2 = `NumMask`

We define both:
```go
const (
    Num     Type = 1 << 3   // concrete: this value is a float
    NumMask Type = Num | Int // constraint: accepts Num or Int
)
```

`IsSubtype(child, parent)` uses the mask form: `parentMask & child == child`.
`IsSubtype(Int, NumMask)` → `(Num|Int) & Int == Int` → true.
`IsSubtype(Num, Int)` → `Int & Num == Num` → false (Int doesn't contain Num).

### Special Types

- **Unknown** = `0` (no bits set). The zero value sentinel. `typ != Unknown`
  becomes `typ != 0`.
- **None** = bottom type, represented as a named sentinel constant (e.g.,
  `1 << 31`). No concrete value inhabits None. `IsSubtype(None, anything)`
  returns true via special case. When guard narrowing produces the empty set
  (e.g., `defined()` on a value known to be only `Undef`), the result is `None`
  — the branch is unreachable.
- **Any** = all concrete type bits OR'd together. Top of the lattice.

### Operations

```go
// Union: what types could this value be?
union := typeA | typeB           // e.g., Int | Str

// Intersection: what types satisfy both constraints?
inter := typeA & typeB           // narrows

// Subtraction: remove a type from a union
without := typ &^ Undef          // remove Undef possibility

// Subtype check: is child contained within parent mask?
IsSubtype(child, parent) = parentMask & child == child

// Could-be check: could this value be Undef?
couldBeUndef := typ & Undef != 0

// Least upper bound: smallest type containing both
LUB(a, b) = a | b               // just OR — the bitset IS the LUB
```

### String Display

`String()` prints the canonical name for known masks (e.g., `Scalar`, `Ref`,
`NumMask` displays as `Num`) or `A|B` for arbitrary unions. Bits are printed
in bit-position order for deterministic output.

### What Changes

**types.go:** `Type` becomes `uint32`. Leaf constants use `1 << N`. Parent
masks defined as OR of children. `parentMap` removed — replaced by mask
constants. `IsSubtype` becomes a one-line mask check (with `None` special case).
`TypeSatisfies` adapts: for polymorphic types (`Any`, `Scalar`, `List`), keep
the existing reverse-subtype check (`IsSubtype(required, actual)`) rather than
the simpler `actual & required != 0`. The simpler check would make
`TypeSatisfies(Str, Int)` return true (since `Str`'s mask contains `Int`'s
bit), but `Str` should NOT satisfy `Int` — a string might hold non-numeric
data. Only the three top-level container types are polymorphic. `String()`
checks a name table for known masks, falls back to `A|B` display.

**narrowing.go:** `NarrowByGuard` and `NegateGuard` use bit operations.
`NarrowByGuard(GuardDefined)` removes the Undef bit: `typ &^ Undef`. If the
result is zero (empty set), return `(None, true)` — the branch is unreachable.
`NegateGuard(GuardDefined)` keeps only the Undef bit: `typ & Undef`. Same
empty-set → None rule. `refTypeMap` stays — maps strings to bitmask values.

**NarrowByContext with unions:** When a union contains bits that narrow
differently in context, apply context narrowing per-bit and OR the results.
For `Array | Str` in scalar context: Array → Int, Str → Str, result = `Int | Str`.

**signatures.go:** No structural change. Signature constants use the new bit
values. `GetBuiltin`, `GetBinaryOp`, `GetUnaryOp` unchanged.

**infer.go:** Annotation map stays `map[uint32]Type`. Guard results stay the
same. Branch merging added at join points (see below).

**symbols.go:** `Symbol.Type` is already `types.Type` — no change needed.

**lsp.go:** `TypeAtByte` returns `types.Type` — no change. The `String()`
method handles display.

### Effort Note

The bitset rewrite is ~60% of the total implementation effort. It is mechanical
— every constant keeps its name, just changes its value — but `NarrowByGuard`
and `NegateGuard` have hand-written logic per guard kind that must be carefully
rewritten as bit operations. All existing tests compare by constant name
(`types.Int`), not numeric value, so they adapt automatically.

## Compound Guards

### CST Shapes (verified by discovery)

```
// defined($x) && ref($x)
binary_expression
  func1op_call_expression "defined($x)"    // left guard
  && (anon)                                 // operator
  func1op_call_expression "ref($x)"        // right guard

// defined($x) and ref($x)
lowprec_logical_expression
  func1op_call_expression "defined($x)"
  and (anon)
  func1op_call_expression "ref($x)"

// Same structure for || and or
```

### Guard Extraction

`extractGuardPattern` currently returns a single `*guardResult`. For compound
guards, return a slice: `[]*guardResult`. A new function
`extractCompoundGuard` handles `binary_expression` and
`lowprec_logical_expression` nodes:

1. Find the operator (`&&`, `||`, `and`, `or`)
2. Recursively call `extractGuardPattern` on the left and right children
3. Return both results tagged with the operator

The `guardResult` struct gains a field to carry compound information:

```go
type guardResult struct {
    VarName  string
    Guard    types.GuardPattern
    Negated  bool
    // Compound guard: if non-nil, this result is a compound of Left op Right.
    // VarName/Guard/Negated are unused when Compound is set.
    Compound *compoundGuard
}

type compoundGuard struct {
    Op    string         // "&&" or "||"
    Left  *guardResult
    Right *guardResult
}
```

### Narrowing Semantics

**`&&` / `and` (conjunction):** Both guards hold in the if-body. Apply both
narrowings — intersection. With bitsets: `narrowByGuard(A) & narrowByGuard(B)`.
In the else-body, at least one guard is false — `!(A && B)` = `!A || !B`. We
cannot apply both negated guards (that would be `!A && !B`, too strong). The
conservative choice is to apply no narrowing in the else-body of a `&&` guard.

**`||` / `or` (disjunction):** At least one guard holds in the if-body. We
cannot apply either guard alone (either could be the true one). No narrowing in
the if-body. In the else-body, both guards are false — apply both negated
guards: `negateGuard(A) & negateGuard(B)`.

Summary:

| Operator | if-body | else-body |
|----------|---------|-----------|
| `&&`     | Both guards (intersection) | No narrowing |
| `\|\|`   | No narrowing | Both negated guards (intersection of negations) |

### Multi-variable guards

`defined($x) && ref($y)` guards two different variables. Each guard applies
independently to its own variable. `walkBlockWithGuard` takes a list of guards
(`[]*guardResult`) and enters one scope that shadows all guarded variables. The
signature changes from a single `*guardResult` to `[]*guardResult`; all callers
are updated. Single-guard callers wrap in a one-element slice.

### Interaction with Negated flag

`!` wrapping a compound guard distributes via De Morgan:
- `!(A && B)` = `!A || !B` — flip operator and negate both sides
- `!(A || B)` = `!A && !B` — flip operator and negate both sides

`extractNegatedGuard` already recurses into `extractGuardPattern`. When the
inner result is compound, flip the operator and toggle Negated on both children.

## Branch Merging at Join Points

### Mechanism

After an if/elsif/else chain where all branches complete (no universal early
exit), the type at the join point is the union of all branch outcomes.

In `walkConditionalStatement`, after walking all branches:

1. Record the type of the guarded variable after each branch exits its scope
2. Compute the union: `branchType1 | branchType2 | ...`
3. Apply via `st.UpdateType`

If any branch has an early exit (detected by `blockAlwaysExits`), that branch
does not contribute to the join — its type is unreachable.

If there is no else branch, the "implicit else" contributes the pre-if type
(the value was not narrowed — no branch executed).

### Example

```perl
my $x = undef;           # $x: Scalar (full mask — all scalar subtypes)
if (defined($x)) {
    # $x: Scalar &^ Undef (Scalar minus Undef bit)
    my $y = $x;
} elsif (ref($x)) {
    # $x: Ref mask
    my $z = $x;
} else {
    # $x: negated ref on already-narrowed type
    my $w = $x;
}
# join point: (Scalar &^ Undef) | Ref | (narrowed else type)
# With bitsets this is a single OR across all branch types.
```

### elsif Chain Merging

`walkElsifNode` currently does not return information about branch types.
Extend it to return the type of the guarded variable after its branch and any
trailing else/elsif branches. `walkConditionalStatement` collects these and
ORs them with the if-branch type.

## Out of Scope

- User-defined type guard functions (TypeScript-style `x is Foo`)
- Guard pattern library (`looks_like_number`, `blessed`, `reftype`)
- Cross-file analysis
- `ref($x) eq 'HASH'` CST extraction (blocked by gotreesitter string grammar)
- Exhaustiveness checking for if/elsif chains
- Post-loop type merging (while/for)

## Implementation Order

1. **Bitset type system** (~60% of effort) — rewrite `types.go` with leaf-bit
   constants and parent masks, update `narrowing.go` (bit operations, empty-set
   → None), update `signatures.go` constants, fix all tests. No changes to
   `infer.go` needed — the annotation map value type stays `types.Type`.
2. **Compound guard extraction** — add `compoundGuard` struct, extend
   `extractGuardPattern` for `binary_expression` / `lowprec_logical_expression`,
   change `walkBlockWithGuard` to accept `[]*guardResult` and shadow multiple
   variables in one scope.
3. **Branch merging** — extend `walkConditionalStatement` and `walkElsifNode`
   to track branch types and compute union (OR) at join points.
