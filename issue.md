---
title: "Tier 09_regex: match, substitution, and the delimiter rules"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:52:57.495477512Z
updated: 2026-09-21T20:21:41.729980778Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T20:21:41.729980778Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:41.729980778Z
---

**Introduces:** match, substitution, binding, delimiters

Depends on nothing from `08_references` -- one of the three independent
pairs the spec names as evidence that the ordering is a partial order rather
than a chain.

Delimiters are the substance: `m//`, `m{}`, `m##`, and the rule that a
paired delimiter nests while an unpaired one does not.

The lexer already reads `if (/TIE/)` as `Quote("/TIE/")` rather than
division, which is the match-vs-divide ambiguity this tier pins down with
files rather than leaves as a note.

## Acceptance

- [ ] Every file in 09_regex/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierRegexPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierRegexPerlValidated'`)
- [ ] The dependency lint passes over 09_regex/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierRegexLint$' -count=1 -v | grep -q '^--- PASS: TestTierRegexLint'`)
- [ ] The tier's adjacency file puts every construct 09_regex introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierRegexAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierRegexAdjacency'`)
- [ ] 09_regex/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierRegexRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierRegexRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/09_regex' -count=1 -v | grep -q '^    --- PASS: TestCorpus/09_regex'`)
