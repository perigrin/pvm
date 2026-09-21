---
title: "Tier 10_io: filehandles, readline, and print as subject rather than fixture"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:52:57.867463731Z
updated: 2026-09-21T20:21:42.010031474Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T20:21:42.010031474Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:42.010031474Z
---

**Introduces:** filehandles, readline, print

`print` appears from tier 01 onward as a fixture -- every behavioural probe
uses it -- and this is where it becomes the subject. Same relationship `my`
has to tier 05.

`glob-angle` is one of the twelve `hardMarkers`, and it places at tier 13
rather than here: `<*>` is a delimiting problem, `<$fh>` is an IO one, and
they are different tiers despite the shared spelling. Files here should make
that split explicit.

## Acceptance

- [ ] Every file in 10_io/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierIoPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierIoPerlValidated'`)
- [ ] The dependency lint passes over 10_io/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierIoLint$' -count=1 -v | grep -q '^--- PASS: TestTierIoLint'`)
- [ ] The tier's adjacency file puts every construct 10_io introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierIoAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierIoAdjacency'`)
- [ ] 10_io/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierIoRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierIoRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/10_io' -count=1 -v | grep -q '^    --- PASS: TestCorpus/10_io'`)
