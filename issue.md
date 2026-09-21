---
title: "Tier 10_io: filehandles, readline, and print as subject rather than fixture"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-0a97-7c14-82fe-27d76ee5c862
created: 2026-09-21T09:52:57.867463731Z
updated: 2026-09-21T10:00:07.595334885Z
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
- [ ] 10_io/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/10_io/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierIoRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierIoRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
