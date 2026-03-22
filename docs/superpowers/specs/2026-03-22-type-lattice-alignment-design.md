# Type Lattice Alignment with Formal Paper

Date: 2026-03-22

## Problem

The PSC type system implements a uint32 bitset lattice in `internal/types/types.go`.
The formal paper at https://gist.github.com/perigrin/c4780a7511ba1421e49a4a8b385aaa3d
defines the authoritative type hierarchy for Perl. The implementation needs two things:

1. New leaf types for IEEE 754 special values (NaN, Inf) that the paper places
   in Scalar but outside the Str/Num chain.
2. Test coverage that validates the implementation against the paper's specific
   rules, edge cases, and common Perl type gotchas.

## Scope

- Add `NaN` and `Inf` as new leaf bits in the type bitset.
- Add paper-alignment tests validating subtyping, exclusion, and edge cases.
- Add Perl-gotcha tests validating that the type system can represent the
  distinctions needed to detect common Perl type errors.
- Do NOT add inference logic (that is issue 4 in the milestone).

## New Leaf Types

### NaN (1 << 17)

IEEE 754 Not-a-Number. Sits in Scalar but outside the Str/Num chain.

The paper's rationale: "NaN" passes syntactic preservation for Num
("NaN" -> NaN -> "NaN" round-trips) but fails semantic fulfillment
(NaN != NaN violates reflexivity, NaN - NaN = NaN violates subtraction
identity).

NaN is also excluded from Str. While "NaN" survives string round-tripping,
the value occupies a semantic no-man's land: it is not a meaningful string
(operations like substr, index, uc on "NaN" are semantically nonsensical)
and it is not a meaningful number. Placing NaN in its own leaf type gives
PSC the ability to generate diagnostics whenever NaN appears in either
string or numeric operations -- both are likely bugs.

### Inf (1 << 18)

IEEE 754 Infinity. Same lattice position as NaN: Scalar, not Str, not Num.

Inf passes syntactic preservation but fails semantic fulfillment
(Inf - Inf = NaN, Inf / Inf = NaN violate expected arithmetic contracts).

### Updated Scalar mask

```go
Scalar = Undef | Bool | Str | DualVar | NaN | Inf | Regex | Ref
```

### Updated Any mask

Any includes all concrete types, so it automatically picks up NaN and Inf
through the updated Scalar mask.

## Files to Modify

### internal/types/types.go

Add leaf bits:

```go
NaN Type = 1 << 17  // IEEE 754 Not-a-Number
Inf Type = 1 << 18  // IEEE 754 Infinity
```

Update Scalar mask to include NaN and Inf.

Add entries to `typeNames` map and `allLeafBits` slice (must maintain
ascending bit-position order: NaN after Glob, Inf after NaN).

Note: Changing the Scalar mask affects all consumers (infer.go, symbols.go,
etc.) because they reference `types.Scalar` symbolically. `make test` must
validate the full suite, not just `internal/types/`.

### internal/types/types_test.go

Add NaN and Inf to `TestTypeString` cases and `TestBitsetLeafTypesAreDistinct`.

### New file: internal/types/paper_alignment_test.go

Tests that validate the lattice against the paper's formal definitions.

#### Subtype chain validation

- Int <: Num <: Str <: Scalar (with comments referencing the paper's
  round-trip coercion and semantic fulfillment rationale)

#### Exclusion tests (paper-specific)

- DualVar is NOT a subtype of Str
- DualVar is NOT a subtype of Num
- DualVar IS a subtype of Scalar
- NaN is NOT a subtype of Num
- NaN is NOT a subtype of Str
- NaN IS a subtype of Scalar
- Inf is NOT a subtype of Num
- Inf is NOT a subtype of Str
- Inf IS a subtype of Scalar

#### Blessed reference unions

- `Object | HashRef` satisfies both Object and HashRef (a blessed hashref
  is simultaneously an Object and a HashRef)
- `Object | HashRef` is a subtype of Ref
- `Object | HashRef` is a subtype of Scalar

#### Bottom and top type properties

- None <: every type
- Unknown == 0, only subtype of itself
- Any contains all concrete leaf bits

### New file: internal/types/gotchas_test.go

Tests that validate the type system can represent distinctions needed
to detect common Perl type gotchas. These test type relationships and
narrowing behavior, not inference (which is issue 4).

#### Operator type confusion

- Str is NOT a subtype of Num: a Str value in numeric context
  is a type mismatch (covers `"hello" + 1`, `"foo" == "bar"`)
- Num is NOT a subtype of Int: a Num value where Int is required is
  a type mismatch
- `eq` on Int/Num values: Int is a subtype of Str (so `eq` does not
  produce a type error, but the coercion is lossy for Num)

#### Undef propagation

- Undef is NOT a subtype of Int, Num, or Str: using Undef in
  arithmetic or string context is a type mismatch
- Undef IS a subtype of Scalar: Undef can appear in scalar variables
- `defined()` guard removes Undef bit (already tested, add
  paper-rationale comments)
- Negated `defined()` guard keeps only Undef bit

#### Context narrowing for aggregates

These behaviors are already tested in `narrowing_test.go`. The gotchas
test file references them with paper rationale but does not duplicate
the assertions. New tests cover only patterns not in the existing suite:

- Ref types in scalar context pass through unchanged (a reference in
  scalar context is still a reference, not a count)

#### DualVar semantics

- DualVar is a subtype of Scalar but not Str or Num: represents values
  where string and numeric interpretations diverge

Note: DualVar context narrowing (hash-key context forcing Str, boolean
context following string component) requires new Context values or
GuardKinds. These are deferred to issue 4 (inference expansion).

#### NaN and Inf semantics

- NaN is NOT a subtype of Num: fails semantic fulfillment (NaN != NaN)
- NaN is NOT a subtype of Str: the string "NaN" is a representational
  artifact, not a meaningful string identity; string operations on it
  (substr, index, uc) are semantically nonsensical
- Inf is NOT a subtype of Num: Inf - Inf = NaN violates subtraction identity
- NaN and Inf are distinct types (different bits, can distinguish in diagnostics)
- NaN is NOT a subtype of Inf, Inf is NOT a subtype of NaN

Note: Inf arithmetic producing NaN (Inf - Inf, Inf / Inf) is an inference
concern deferred to issue 4.

#### Reference type safety

- Ref types are NOT subtypes of Str: stringifying a reference produces
  "ARRAY(0x...)" which is almost never intended
- Object is a subtype of Ref: blessed references are references
- `Object | HashRef` represents blessed hashrefs: satisfies both Object
  and HashRef requirements (method calls AND hash dereferencing)

#### Boolean edge cases

- Bool is NOT a subtype of Int: `is_bool()` primitive booleans are distinct
  from numeric 0/1
- Bool IS a subtype of Scalar
- `GuardBool` narrows Scalar to Bool, narrows Int to None (unreachable)
  (already tested in narrowing_test.go; referenced here for paper context)
- The string "0" being falsy is a runtime behavior, not a type distinction:
  the type system tracks Bool as primitive booleans only

#### Undef vs None distinction

- Undef is a concrete type (a value that exists), None is the bottom type
  (an empty set, unreachable branch sentinel)
- Undef is a subtype of Scalar; None is a subtype of everything
- These are distinct: `Undef != None`

Note: Return type inference (`return undef` producing Undef vs bare
`return` producing empty list) is deferred to issue 4.

## Implementation Order

1. Add NaN and Inf leaf bits to types.go
2. Update Scalar mask, typeNames, allLeafBits
3. Update existing tests (TestTypeString, TestBitsetLeafTypesAreDistinct,
   TestParentMasksContainDescendants)
4. Add paper_alignment_test.go
5. Add gotchas_test.go
6. Run `make test` to verify all tests pass
