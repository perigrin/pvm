---
title: "Tier 11_oo: bless and class, where the adjacency file catches the live bug"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-0929-795d-ac6b-c1740240e49e
blocks:
- 01a0c362-0ff3-7f3f-b2d3-f6dbe8228f86
created: 2026-09-21T09:52:58.28262933Z
updated: 2026-09-21T10:00:07.631025967Z
---

**Introduces:** bless AND class/field/method/ADJUST, indirect new

Comes BEFORE packages, reversing an earlier draft. Measured under 5.42.0:

    class Foo { method m { 1 } }     syntax OK, no `package` anywhere
    bless {}, "Foo"                  needs a REF and a STRING, nothing else

`class Foo {}` IS a package declaration -- a different spelling of tier 12's
construct, not a consumer of it -- so the old order was circular.
`__PACKAGE__` stays `main` outside the class body, so `class` introduces its
own scope rather than inheriting one.

## The adjacency file is what catches the live bug

Measured 2026-09-21: `class`, `field`, `method`, `ADJUST`, `:param` and
`:isa` each parse clean IN ISOLATION. **True per construct and false per
file.** `class Foo { ADJUST { 1 } method m { 2 } }` is ONE Unknown
swallowing both, and `t2_test.go:105-111` records 41 Unknowns across seven
`class/*.t` files with nothing in the chain owning ADJUST.

So this tier's construct files would ALL PASS over a live bug. Its adjacency
file is the one that fails. Write the adjacency files first here -- this is
the tier that proves why that third kind of file exists.

`indirect-new` is one of the twelve `hardMarkers` and places here.

## Acceptance

- [ ] The ADJUST-plus-method adjacency file fails while the construct files pass, which is the case that proves adjacency files earn their place (`go test ./internal/conformance/ -run '^TestTierOoAdjacencyCatchesAdjust$' -count=1 -v | grep -q '^--- PASS: TestTierOoAdjacencyCatchesAdjust'`)
- [ ] Every file in 11_oo/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOoPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOoPerlValidated'`)
- [ ] The dependency lint passes over 11_oo/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOoLint$' -count=1 -v | grep -q '^--- PASS: TestTierOoLint'`)
- [ ] 11_oo/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/11_oo/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOoRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOoRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
