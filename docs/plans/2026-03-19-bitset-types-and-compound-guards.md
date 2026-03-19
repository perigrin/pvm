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

Replace `type Type int` with `type Type uint32` where each concrete type gets a
unique bit. Parent types are the bitwise OR of their children. Unions are native
OR operations. Intersections are native AND.

### Bit Assignment

22 types fit in 32 bits with 10 bits of headroom. Perl's type lattice is stable
— the only future additions would be user-defined class names, which are
subtypes of Object and can share Object's bit or get a name tag.

```
Bit  Type        Parent mask
---  ----        -----------
0    Undef       Undef
1    Bool        Bool
2    Int         Int
3    Num         Int | Num
4    Str         Int | Num | Str
5    DualVar     DualVar
6    Regex       Regex
7    ScalarRef   ScalarRef
8    ArrayRef    ArrayRef
9    HashRef     HashRef
10   CodeRef     CodeRef
11   GlobRef     GlobRef
12   Object      Object
13   Array       Array
14   Hash        Hash
15   Code        Code
16   Glob        Glob
17   None        (special: zero or all-bits, see below)

Derived masks:
  Ref     = ScalarRef | ArrayRef | HashRef | CodeRef | GlobRef | Object
  Scalar  = Undef | Bool | Str | Num | Int | DualVar | Regex | Ref
  List    = Array | Hash
  Any     = Scalar | List | Code | Glob
```

Note: `Num` includes the `Int` bit because Int <: Num. `Str` includes both
`Int` and `Num` bits because Num <: Str. Each parent mask contains all its
descendants' bits. This makes `IsSubtype(child, parent)` a single mask
operation: `parent & child == child`.

### Special Types

- **Unknown** = `0` (no bits set). The zero value. Comparisons like
  `typ != Unknown` become `typ != 0`.
- **None** = bottom type. `IsSubtype(None, anything)` returns true. Represented
  as a sentinel constant, not a bitmask — no value inhabits None.
- **Any** = all concrete type bits OR'd together. Top of the lattice.

### Operations

```go
// Union: what types could this value be?
union := typeA | typeB           // Scalar | Ref

// Intersection: what types satisfy both constraints?
inter := typeA & typeB           // narrows

// Subtraction: remove a type from a union
without := typ &^ Undef          // remove Undef possibility

// Subtype check: is child contained within parent?
IsSubtype(child, parent) = parent & child == child

// Could-be check: could this value be Undef?
couldBeUndef := typ & Undef != 0

// Least upper bound: smallest type containing both
LUB(a, b) = a | b               // just OR — the bitset IS the LUB
```

### What Changes

**types.go:** `Type` becomes `uint32`. Constants become bit positions.
`parentMap` is replaced by the mask definitions. `IsSubtype` becomes a one-line
mask check. `TypeSatisfies` and `polymorphicTypes` adapt to bitmasks. The
`String()` method prints the canonical name for single-bit types or `A|B` for
unions.

**narrowing.go:** `NarrowByGuard` and `NegateGuard` use bit operations.
`NarrowByGuard(GuardDefined)` becomes `typ &^ Undef` (remove Undef bit).
`NegateGuard(GuardDefined)` becomes `typ & Undef` (keep only Undef bit).
`refTypeMap` stays — it maps strings to specific bitmask values.

**signatures.go:** No structural change. Signature constants use the new bit
values. `GetBuiltin`, `GetBinaryOp`, `GetUnaryOp` unchanged.

**infer.go:** Annotation map stays `map[uint32]Type`. Guard results stay the
same. Branch merging added at join points (see below).

**symbols.go:** `Symbol.Type` is already `types.Type` — no change needed.

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
In the else-body, at least one guard is false — apply neither (conservative).

Actually, the else-body semantics for `&&` are: `!(A && B)` = `!A || !B`. We
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
independently to its own variable. `walkBlockWithGuard` must handle a list of
guards, entering a scope that shadows multiple variables.

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
my $x = undef;           # $x: Scalar (sigil type)
if (defined($x)) {
    # $x: Scalar (non-Undef, narrowed by defined guard)
    my $y = $x;
} elsif (ref($x)) {
    # $x: Ref
    my $z = $x;
} else {
    # $x: Scalar (negated ref on already-narrowed type)
    my $w = $x;
}
# join point: Scalar | Ref | Scalar = Scalar (Ref is a subtype of Scalar)
```

With bitsets, the union computation is a single OR across branch types. If the
result equals the pre-if type, no update is needed.

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

1. **Bitset type system** — rewrite `types.go`, update `narrowing.go` and
   `signatures.go`, fix all tests. No changes to `infer.go` needed — the
   annotation map value type stays `types.Type`, it just holds bitmasks now.
2. **Compound guard extraction** — add `compoundGuard` struct, extend
   `extractGuardPattern` for `binary_expression` / `lowprec_logical_expression`,
   update `walkBlockWithGuard` to handle compound guards.
3. **Branch merging** — extend `walkConditionalStatement` and `walkElsifNode`
   to track branch types and compute union at join points.
