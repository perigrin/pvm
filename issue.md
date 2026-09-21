---
title: "Tier 07_subroutines: call forms, and the parenless argument question"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-047b-76a4-b0f8-9043eee66baa
blocks:
- 01a0c362-0929-795d-ac6b-c1740240e49e
created: 2026-09-21T09:52:56.634555553Z
updated: 2026-09-21T10:00:07.488093661Z
---

**Introduces:** declaration, call forms, @_, return, signatures

## The tier the smallest useful corpus cannot skip

The two largest T2 Unknown buckets are both parenless call argument
position. A parenless call to an unknown callee is where the parser must
decide how much to consume, and it is what makes tiers 01-04 alone
insufficient as a gate.

`signature` is one of the twelve `hardMarkers` and places here.

Call FORMS are the substance: `f(1)`, `f 1`, `&f`, `$code->()` and
method-call syntax each decide their arguments differently.

## Acceptance

- [ ] Every file in 07_subroutines/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint passes over 07_subroutines/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] 07_subroutines/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/07_subroutines/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierSubroutinesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
