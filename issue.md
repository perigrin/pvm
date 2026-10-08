---
title: The lattice has a Void type, and Maybe[T] and Optional[T] are the joins Undef|T and Void|T
state: in-progress
urgency: normal
milestone: type-libraries
blocks:
- 01a11920-411a-7cea-b78f-c1568aba6d25
- 01a1193c-44f3-7f27-bd9c-a789eb9a74e2
created: 2026-10-08T01:53:26.234239053Z
updated: 2026-10-08T02:17:16.933211588Z
sessions:
- start_sha: 214a28feeaf479afafe15410e923f54e81a1a919
  end_sha: ""
  commits: 0
  started_at: 2026-10-08T02:17:16.933211588Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T02:17:16.933211588Z
---

## Context

- paths: `internal/types/types.go` (leaf bits; `List` mask; `typeNames`; `allLeafBits`), `internal/types/join.go`, `internal/types/coercion.go`, `internal/types/narrowing.go` (`VoidCtx`), `internal/types/testdata/lattice_oracle.pl`, `internal/parse/derive.go` (`prototypeFromTypes`: `+` derives from exactly `List`), `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); `go test ./internal/types/ -count=1`
- paper: `~/dev/types-paper/2-perl-types-formal.md` "### Void" (~line 1011) and "Arity orders the top of the lattice" (~line 1620)

perigrin, 2026-10-08: add a Void type to the lattice. The paper: `Void := {()}`, arity 0, one inhabitant (the empty list) -- "returns nothing", distinct from None's "does not return". Arity orders the top: Void {0}, Scalar {1}, List {0,1,2,...}, so Void <: List, Void and Scalar are incomparable siblings, and List is their join. Scalar context coerces it: `Void ⇓^Scalar undef` (a bare `return` is Void). pvm has only the context marker `VoidCtx`, no type.

Wrapper types (perigrin, 2026-10-08; no general rule for parametric types yet): `Maybe[T]` is `Undef|T` (present, possibly undef; Moose's and Type::Tiny's meaning) and `Optional[T]` is `Void|T` (may be absent; Type::Tiny's Tuple/Dict meaning). They are names for those unions, not new types: `srand(undef)` is `Maybe[Int]` (an Int slot holding undef; measured on 5.42: warns "Use of uninitialized value" and "Integer overflow", seeds 18446744073709551615), `srand()` is `Optional[Int]`. Container types (`List[T]`, `Array[T]`, `Hash[T]`, `Ref[T]`) keep their element-type parameter: `Ref[Array]` etc. already have non-parametric names (`ArrayRef`...), but an element type has none. The paper section is `~/dev/types-paper` (canonical source; gh-pages renders it) "Absent and undefined".

In the bitset: a new leaf (bit 20 is free), `List` gains it, `typeNames` spells it `Void`. Watch: anything comparing to exactly `List` (`prototypeFromTypes`'s `+`, `Array|Hash|Scalar` literals) changes meaning when `List` widens -- find each and decide by measurement, not by keeping the old mask.

## Steps

- [ ] Write failing test `TestVoidInLattice`: `Void <: List`; `Void` and `Scalar` incomparable; `Join(Void, Scalar) == List`... (as the paper states it -- if pvm's List also holds Array and Hash, assert the paper's facts that hold, and record the difference); `Void ⊓ Scalar` is None; `ParseType("Void")` round-trips
- [ ] Write failing test `TestVoidCoercesToUndefInScalarContext`: `Void` coerces to `Undef` in scalar context; a `None` result (die) does not coerce to anything but is vacuously accepted, as now
- [ ] Write failing negative test `TestVoidIsNotNone`: `Void != None`, `None <: Void`, not `Void <: None`
- [ ] Write failing test `TestWrapperTypesAreJoins`: `Maybe[Int]` reads as exactly `Undef|Int` and `Optional[Str]` as exactly `Void|Str`, and each prints back as the union; `Maybe` or `Optional` with no parameter, or with two, is a type error naming it (negative)
- [ ] Add the leaf; widen `List`; fix every exact-`List` comparison; update the lattice oracle if it enumerates leaves; record Void in RFC 0001's type list; `make test`; commit

## Acceptance criteria

- [ ] Maybe and Optional are joins (`go test ./internal/types/ ./internal/parse/ -run '^TestWrapperTypesAreJoins$' -count=1 -v | grep -q '^--- PASS: TestWrapperTypesAreJoins'`)
- [ ] Void is in the lattice under List (`go test ./internal/types/ -run '^TestVoidInLattice$' -count=1 -v | grep -q '^--- PASS: TestVoidInLattice'`)
- [ ] scalar context coerces Void to undef (`go test ./internal/types/ -run '^TestVoidCoercesToUndefInScalarContext$' -count=1 -v | grep -q '^--- PASS: TestVoidCoercesToUndefInScalarContext'`)
- [ ] Void is not None (`go test ./internal/types/ -run '^TestVoidIsNotNone$' -count=1 -v | grep -q '^--- PASS: TestVoidIsNotNone'`)
