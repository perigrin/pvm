---
title: Expand PSC inference to cover more Perl type semantics
state: pending
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:47.968515-04:00
updated: 2026-03-22T02:43:47.968515-04:00
---

Expand PSC's type inference beyond arity and container-level type mismatches to
cover value-level type checking as described in the formal papers.

## Expansion Areas

1. String-to-number coercion detection — flag `my $x = "hello"; $x + 1`
2. Undef propagation — track undef through expressions
3. Return type inference across files — verify against real modules
4. User-defined function signatures — infer from usage patterns
5. Context-sensitive inference — scalar vs list context
6. Pattern matching guards — regex matches providing type information

## Acceptance

- [ ] Non-numeric string in numeric context produces type-mismatch diagnostic
- [ ] Undef used where defined value required produces diagnostic (without guard)
- [ ] Cross-file return type inference works on multi-file test case
- [ ] At least one user-defined function gets inferred parameter types

Key files: internal/infer/infer.go, internal/types/signatures.go
See also: GitHub #409
