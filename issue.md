---
title: "Tier 02_variables: sigils, containers, element access"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c361-7479-7275-ace5-cd9a6dd2bd6c
- 01a0c362-0929-795d-ac6b-c1740240e49e
created: 2026-09-21T09:52:14.225829534Z
updated: 2026-09-21T13:40:24.099721265Z
---

**Introduces:** sigils, scalars, arrays, hashes, element access,
delete/exists, `$::`

Tier 01 gives literals; this tier gives the things that hold them.
Self-contained in the same way: `my $x = 1` needs no notion of context, so
nothing here depends on tier 03 or later.

Three of the twelve `hardMarkers` place here -- `delete`, `exists` and
`pkg-colon` -- which makes this tier a real measurement rather than a
formality.

`$::` is the case worth writing early: package-qualified names are lexically
distinctive, so the token stream classifies them and a declaration would be
checking itself.

## Acceptance

- [ ] Every file in 02_variables/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierVariablesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesPerlValidated'`)
- [ ] The dependency lint passes over 02_variables/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierVariablesLint$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesLint'`)
- [ ] 02_variables/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/02_variables/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierVariablesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
