---
title: "Tier 07_subroutines: call forms, and the parenless argument question"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c362-0ff3-7f3f-b2d3-f6dbe8228f86
created: 2026-09-21T09:52:56.634555553Z
updated: 2026-09-21T13:42:10.369106471Z
---

**Introduces:** declaration, `@_`, return, signatures

**Call forms are NOT here.** They are split into their own issue, because
the spec names "tiers 01-04 plus a call-form slice from tier 07" as the
smallest useful corpus, and burying that slice behind declaration and
signatures delays the milestone's one stated delivery target. This issue is
the rest of the tier.

`signature` is one of the twelve `hardMarkers` and places here.

## What is left once call forms are gone

    sub f { }             declaration, named
    sub { }               anonymous
    sub f ($x, $y) { }    signatures, 5.36+ and stable in 5.42
    @_                    the implicit argument array
    return / implicit     the last-expression rule

`@_` is the one worth care: it aliases, so `$_[0] = 1` modifies the
caller's variable. A file that treats it as a plain array measures nothing
about what makes it different.

## Acceptance

- [ ] Every file in 07_subroutines/ that this issue writes is validated against pinned perl 5.42.0 (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint passes over 07_subroutines/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] The tier's adjacency file puts every construct it introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestTierSubroutinesAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesAdjacency'`)
- [ ] `@_` aliasing is measured against perl rather than assumed, since that is what distinguishes it from an array (`go test ./internal/conformance/ -run '^TestTierSubroutinesArgAliasing$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesArgAliasing'`)
- [ ] 07_subroutines/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierSubroutinesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/07_subroutines' -count=1 -v | grep -q '^    --- PASS: TestCorpus/07_subroutines'`)
