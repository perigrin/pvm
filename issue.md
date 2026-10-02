---
title: "Types paper: parametric containers, tuples, lvalue lists, and multi and context-indexed function types"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:09.165064836Z
updated: 2026-10-02T01:42:09.165064836Z
---

perigrin (2026-10-01): parametric types go into the paper (~/dev/types-paper) first, so pvm and perl5-son extend one lattice. RFC 0001's open questions 1, 3, 4 and 5 wait on it:

- `List[T]` (values in flight, covariant) against `Array[T]` (a container, invariant); `Hash[T]` flattening to `List[Str|T]`; `ArrayRef[T]` from `Array[T]`; whether `T <: List[T]`.
- Lvalue lists: `chomp`/`chop` modify their arguments in place.
- Tuples: `each` returns `(Str, T)` for a hash.
- Function types with several signatures, selected by arity, argument type or context.

Membership in the paper is defined through interpretation (`v ∈ T` iff interpreting `v` through `T` is lossless and its contracts hold), and the parametric section should keep that definition.

## Acceptance criteria

- [ ] The paper has a parametric-types section (`sh -c 'grep -qi "^#\+ .*parametric" $HOME/dev/types-paper/2-perl-types-formal.md'`)
