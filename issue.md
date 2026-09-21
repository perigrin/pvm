---
title: "Tier 13_opaque: delimit without lexing the contents"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c362-14e9-70d6-8842-a5b379816a98
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:59.516182101Z
updated: 2026-09-21T20:21:42.117864833Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T20:21:42.117864833Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:42.117864833Z
---

**Introduces:** format, qx, `<*>`, heredocs, POD, `__END__`/`__DATA__`

Where the lexer must DELIMIT without lexing the contents. The M0 text names
this set already: heredocs, POD, `__END__`, `q{}` with nested delimiters
"all to be *recognised* -- not understood, but delimited correctly."

Four `hardMarkers` place here: heredoc, format, qx, glob-angle.

Heredocs are a known refusal: a heredoc in statement position does not yet
parse, body and all (M1 issue 01a0c13f-aaf8). Files here will refuse, and
that is the corpus working rather than a blocker -- a refusal is recorded
with its code and skips until the parser catches up.

## Acceptance

- [ ] Each of the six delimiter kinds is delimited correctly without its contents being lexed (`go test ./internal/conformance/ -run '^TestTierOpaqueDelimitsWithoutLexing$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueDelimitsWithoutLexing'`)
- [ ] Every file in 13_opaque/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOpaquePerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOpaquePerlValidated'`)
- [ ] The dependency lint passes over 13_opaque/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOpaqueLint$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueLint'`)
- [ ] The tier's adjacency file puts every construct 13_opaque introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierOpaqueAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueAdjacency'`)
- [ ] 13_opaque/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOpaqueRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus/13_opaque' -count=1 -v | grep -q '^    --- PASS: TestCorpus/13_opaque'`)
