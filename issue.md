---
title: Tier READMEs, because the dependency lint needs an INTRODUCES set that cannot drift
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-6e43-7c0d-8828-ef7659402e39
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:50:20.295669503Z
updated: 2026-09-21T10:00:07.234431253Z
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

## Acceptance

- [ ] The lint reads its per-tier INTRODUCES set from the READMEs rather than a hand-maintained Go table (`go test ./internal/conformance/ -run '^TestIntroducesReadFromReadmes$' -count=1 -v | grep -q '^--- PASS: TestIntroducesReadFromReadmes'`)
- [ ] Every tier directory has a README stating what it introduces and what it depends on (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] A construct in a README that no file in that tier introduces is reported (`go test ./internal/conformance/ -run '^TestAspirationalIntroducesReported$' -count=1 -v | grep -q '^--- PASS: TestAspirationalIntroducesReported'`)
- [ ] The spec's open-questions section records both questions as answered (`grep -qi 'grouping' docs/plans/2026-09-21-graded-conformance-corpus.md`)
