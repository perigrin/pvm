---
title: "Tier 13_opaque: delimit without lexing the contents"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c362-14e9-70d6-8842-a5b379816a98
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:59.516182101Z
updated: 2026-10-23T19:18:30.954993175Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  commits: 1
  started_at: 2026-09-21T20:21:42.117864833Z
  ended_at: 2026-09-21T20:47:29.925665969Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:42.117864833Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:47:29.925665969Z
observed_paths:
- conformance/02_variables/00_adjacency.t
- conformance/02_variables/02_array_element.t
- conformance/02_variables/05_braced_name.t
- conformance/02_variables/06_hash.t
- conformance/02_variables/07_hash_element_expr.t
- conformance/02_variables/09_hash_slice.t
- conformance/02_variables/08_hash_exists_delete.t
- conformance/02_variables/12_package_scalar.t
- conformance/02_variables/10_package_array.t
- conformance/02_variables/11_package_hash.t
- conformance/02_variables/README.md
- conformance/07_subroutines/00_adjacency.t
- conformance/07_subroutines/07_args_alias.t
- conformance/07_subroutines/README.md
- conformance/09_regex/00_adjacency.t
- conformance/09_regex/08_hash_delimiter.t
- conformance/09_regex/09_nesting_delimiters.t
- conformance/09_regex/README.md
- conformance/10_io/02_readline_scalar.t
- conformance/10_io/03_readline_list.t
- conformance/10_io/README.md
- conformance/13_opaque/00_adjacency.t
- conformance/13_opaque/01_heredoc_interpolating.t
- conformance/13_opaque/02_heredoc_literal.t
- conformance/13_opaque/03_heredoc_indented.t
- conformance/13_opaque/06_format_write.t
- conformance/13_opaque/08_data_section.t
- conformance/13_opaque/09_format_picture.t
- conformance/13_opaque/10_pod_not_perl.t
- conformance/13_opaque/11_data_not_perl.t
- conformance/13_opaque/README.md
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier02_test.go
- internal/conformance/tier05_test.go
- internal/conformance/tier06_test.go
- internal/conformance/tier07_test.go
- internal/conformance/tier09_test.go
- internal/conformance/tier10_test.go
- internal/conformance/tier13_test.go
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

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/13_opaque` -> `TestCorpus` over the tier-13 topics (adjacency-13_opaque.md, formats.md, heredocs.md, pod-and-data.md, quote-like-terms.md)). Left unchanged: `TestTierOpaqueAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierOpaqueAdjacency` -> `TestAdjacencyCoversItsTier/13_opaque`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Each of the six delimiter kinds is delimited correctly without its contents being lexed (`go test ./internal/conformance/ -run '^TestTierOpaqueDelimitsWithoutLexing$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueDelimitsWithoutLexing'`)
- [ ] Every file in 13_opaque/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOpaquePerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOpaquePerlValidated'`)
- [ ] The dependency lint passes over 13_opaque/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOpaqueLint$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueLint'`)
- [ ] The tier's adjacency topic puts every construct 13_opaque introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^13_opaque$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/13_opaque'`)
- [ ] 13_opaque/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOpaqueRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOpaqueRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-13_opaque|formats|heredocs|pod-and-data|quote-like-terms)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
