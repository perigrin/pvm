---
title: "Tier 07 call-form slice: the smallest useful corpus gate"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-a317-7807-bb9e-15f299562244
created: 2026-09-21T13:41:10.741737124Z
updated: 2026-09-21T13:41:43.995310787Z
---

**Introduces:** call forms -- `f(1)`, `f 1`, `&f`, `$code->()`, method-call
syntax

## Why this is split out from the rest of tier 07

The spec names the delivery milestone: **"the smallest useful corpus is
tiers 01-04 plus a call-form slice from tier 07"**. Call forms are the
slice. Declaration, `@_`, `return` and signatures are the rest, and they are
not on the path to that milestone.

The reason the slice is what matters: **the two largest T2 Unknown buckets
are both parenless call argument position.** A parenless call to an unknown
callee is where the parser must decide how much to consume, and it is the
thing that makes a gate on tiers 01-04 alone insufficient.

The spec's own words on why 01-03 is not enough: "tier 01 is one landed
issue away rather than satisfied, which means a gate on tiers 01-03 says
almost nothing. Tier 04 is the first tier where a parse can be WRONG rather
than Unknown, and the two largest T2 buckets are parenless call argument
position, which needs the 07 slice."

## What each form decides differently

    f(1)          parenthesised, unambiguous
    f 1           parenless -- how much does it consume?
    &f            no argument list at all, inherits @_
    $code->()     the callee is a value, not a name
    $obj->m(1)    the callee is resolved at run time

A corpus file per form, plus the adjacency file that puts them side by side.

## Acceptance

- [ ] A parenless call to an unknown callee has its argument extent measured against perl rather than assumed (`go test ./internal/conformance/ -run '^TestTierCallFormsParenlessExtent$' -count=1 -v | grep -q '^--- PASS: TestTierCallFormsParenlessExtent'`)
- [ ] Every file in 07_subroutines/ that this issue writes is validated against pinned perl 5.42.0 (`go test ./internal/conformance/ -run '^TestTierCallFormsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierCallFormsPerlValidated'`)
- [ ] The dependency lint passes over the call-form files (`go test ./internal/conformance/ -run '^TestTierCallFormsLint$' -count=1 -v | grep -q '^--- PASS: TestTierCallFormsLint'`)
- [ ] The adjacency file places the call forms side by side in one body, per the spec's rule (`go test ./internal/conformance/ -run '^TestTierCallFormsAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierCallFormsAdjacency'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierCallFormsRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierCallFormsRefusalsCited'`)
- [ ] The smallest-useful-corpus gate is met: tiers 01-04 plus this slice all green (`go test ./internal/conformance/ -run '^TestSmallestUsefulCorpus$' -count=1 -v | grep -q '^--- PASS: TestSmallestUsefulCorpus'`)
