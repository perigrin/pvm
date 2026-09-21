---
title: "Tier 06_control: statements, which are a different code path from tier 04"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-c7d7-7ea4-9c93-b8bcc80a108e
blocks:
- 01a0c362-073a-78c2-ba73-227fb48a26c4
created: 2026-09-21T09:52:55.931602782Z
updated: 2026-09-21T10:00:07.451764201Z
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
- [ ] 06_control/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/06_control/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierControlRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierControlRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
