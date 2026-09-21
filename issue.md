---
title: "Tier 05_scoping: my, our, state, local, blocks"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:52:40.407130019Z
updated: 2026-09-21T13:43:01.911441851Z
---

**Introduces:** my, our, state, local, blocks

`my` appears at tier 01 in `my $x = .5`, four tiers before the tier that
introduces it. That is the partial order working as documented, not a
violation: tier 01 USES `my` as a fixture and this tier is where its
semantics are the subject.

`local` is the one worth care -- dynamic scope, not lexical, and a file that
treats it as a spelling variant of `my` measures nothing.

## Acceptance

- [ ] Every file in 05_scoping/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierScopingPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierScopingPerlValidated'`)
- [ ] The dependency lint passes over 05_scoping/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierScopingLint$' -count=1 -v | grep -q '^--- PASS: TestTierScopingLint'`)
- [ ] The tier's adjacency file puts every construct 05_scoping introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierScopingAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierScopingAdjacency'`)
- [ ] 05_scoping/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierScopingRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierScopingRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/05_scoping' -count=1 -v | grep -q '^    --- PASS: TestCorpus/05_scoping'`)
