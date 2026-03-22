---
title: Align PSC type hierarchy with formal type system papers
state: pending
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:41.621506-04:00
updated: 2026-03-22T02:43:41.621506-04:00
---

Ensure PSC's internal type system matches the formal definition in the companion papers:
https://gist.github.com/perigrin/c4780a7511ba1421e49a4a8b385aaa3d

The bitset hierarchy in `internal/types/types.go` already covers much of the lattice.
Gaps to verify and address:

1. Round-trip coercion tests — type membership via syntactic preservation and semantic fulfillment
2. DualVar positioning in the lattice (proven concrete by errno `$!`)
3. Bool as primitive (`builtin::true`/`builtin::false` via `is_bool()`) vs truthy/falsy
4. NaN edge case — passes syntactic preservation for Num but fails semantic fulfillment
5. Object subtyping — blessed references and `isa` guards

## Acceptance

- [ ] Type lattice matches papers: Int <: Num <: Str <: Scalar with all leaf types
- [ ] DualVar correctly positioned as concrete Scalar subtype
- [ ] Bool primitive distinction is complete
- [ ] Tests validate subtyping relationships against paper definitions

Key files: internal/types/types.go, internal/types/narrowing.go
See also: GitHub #408
