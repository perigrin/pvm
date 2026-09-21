---
title: Tier READMEs, because the dependency lint needs an INTRODUCES set that cannot drift
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-6e43-7c0d-8828-ef7659402e39
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:50:20.295669503Z
updated: 2026-09-21T15:40:37.240541208Z
---

Listed in the spec's tooling list and left as one of two open questions:
"Does each tier get a README stating what it introduces and why it sits
there? Recommended yes. The spike wrote no README and did not miss one,
which is weak evidence at three files."

## The argument that settles it, which is not the documentation argument

A tier README is not primarily for readers. **The dependency lint needs a
machine-readable INTRODUCES set per tier**, and that set has to live
somewhere. In a Go table, the tier's contents and its declared contents are
two lists that drift. In the tier directory, drift is visible in the same
diff.

So the README is where the lint's input lives, and the prose comes free.
That also answers "why does this tier sit here" -- check 2 says a tier sits
where it does because the next cannot proceed without it, and that claim
should be written where it can be argued against.

## The second open question, folded in

"Does `04_operators` need internal grouping by precedence level?" It is the
largest tier by a distance. **Do not decide now.** Decide it when the tier
is written and the directory is actually unreadable; a grouping invented
before the files exist will be the wrong grouping. The deferral is recorded
against the tier 04 issue rather than left in the spec's open questions.

## This issue also enforces adjacency across every tier

The spec's rule is per tier -- "every construct the tier introduces, in one
body, each adjacent to another" -- and its argument is that a
one-construct-per-file corpus CANNOT reach the ADJUST-class bug at all:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

Each tier issue carries its own adjacency criterion, but a per-tier
criterion can be forgotten one tier at a time. This issue owns the generic
check, because it already owns the INTRODUCES set the check needs. Where a
tier declares a prerequisite the adjacency file pairs with THAT, not with
tier N-1 -- pairing 12_packages with 11_oo would assert nothing, since the
spec says packages needs nothing from the tier before it.

## Acceptance

- [ ] The lint reads its per-tier INTRODUCES set from the READMEs rather than a hand-maintained Go table (`go test ./internal/conformance/ -run '^TestIntroducesReadFromReadmes$' -count=1 -v | grep -q '^--- PASS: TestIntroducesReadFromReadmes'`)
- [ ] Every tier directory has a README stating what it introduces and what it depends on (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] A construct in a README that no file in that tier introduces is reported (`go test ./internal/conformance/ -run '^TestAspirationalIntroducesReported$' -count=1 -v | grep -q '^--- PASS: TestAspirationalIntroducesReported'`)
- [ ] The spec's open-questions section is gone, both questions having been answered rather than merely asked (`! grep -q '^## Open questions' docs/plans/2026-09-21-graded-conformance-corpus.md`)
- [ ] Every tier directory holds an adjacency file covering its full INTRODUCES set, checked generically so no tier can omit one (`go test ./internal/conformance/ -run '^TestEveryTierHasAnAdjacencyFile$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAnAdjacencyFile'`)
- [ ] A tier README declaring a prerequisite at a later tier number fails the lint, which is check 2 given teeth (`go test ./internal/conformance/ -run '^TestTierNecessityCheck$' -count=1 -v | grep -q '^--- PASS: TestTierNecessityCheck'`)
