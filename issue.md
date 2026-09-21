---
title: "Tier 13_opaque: delimit without lexing the contents"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-0ff3-7f3f-b2d3-f6dbe8228f86
blocks:
- 01a0c362-14e9-70d6-8842-a5b379816a98
created: 2026-09-21T09:52:59.516182101Z
updated: 2026-09-21T10:00:07.718808823Z
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

- [ ] Every file in 13_opaque/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOpaquePerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOpaquePerlValidated'`)
- [ ] The dependency lint passes over 13_opaque/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOpaqueLint$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueLint'`)
- [ ] 13_opaque/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/13_opaque/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOpaqueRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
