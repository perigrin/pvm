---
title: "Tier 06_control: statements, which are a different code path from tier 04"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:55.931602782Z
updated: 2026-09-21T20:21:39.117321671Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T20:21:39.117321671Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:39.117321671Z
---

**Introduces:** if/unless, while/until, for/foreach, postfix forms, goto

Statement-level parsing, handled by hand-written recursive descent
(`parse.go` dispatching to `control.go`) rather than by the Pratt expression
parser. That makes this tier a measurement of a different code path from
tier 04's, which is why it is its own tier.

Postfix forms (`EXPR if COND`) most often reveal a statement/expression
boundary bug, since the same construct parses both ways depending on
position.

`goto` is one of the twelve `hardMarkers` and places here.

## Acceptance

- [ ] Every file in 06_control/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierControlPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierControlPerlValidated'`)
- [ ] The dependency lint passes over 06_control/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierControlLint$' -count=1 -v | grep -q '^--- PASS: TestTierControlLint'`)
- [ ] The tier's adjacency file puts every construct 06_control introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierControlAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierControlAdjacency'`)
- [ ] 06_control/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierControlRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierControlRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/06_control' -count=1 -v | grep -q '^    --- PASS: TestCorpus/06_control'`)
