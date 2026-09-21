---
title: "Tier 03_context: scalar vs list, and where the inference gap becomes a number"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
blocks:
- 01a0c361-a317-7807-bb9e-15f299562244
created: 2026-09-21T09:52:19.065451383Z
updated: 2026-09-21T10:00:07.345821057Z
---

**Introduces:** scalar vs list context, the discriminating pairs

Context sits immediately before operators and no earlier. The moment
operators arrive you cannot proceed without it: `@a + 1` is arithmetic on a
count and `"$x" . @a` is not, and the operator decides which.

## This tier is where a known inference gap becomes measurable

`types.NarrowByContext` is called at `infer.go:524` with NO source of
context: the CST has no `Wantarray` node and no call-site propagation. The
discriminating pairs are the files that make that gap a number rather than a
note.

chromatic opens *Modern Perl* with context, before any syntax, because a
READER needs the frame immediately. A PARSER needs it when operators arrive.
Same fact, different consumers -- and this corpus is ordered for the parser.

## Acceptance

- [ ] Every file in 03_context/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierContextPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierContextPerlValidated'`)
- [ ] The dependency lint passes over 03_context/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierContextLint$' -count=1 -v | grep -q '^--- PASS: TestTierContextLint'`)
- [ ] 03_context/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/03_context/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierContextRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierContextRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
