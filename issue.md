---
title: "Tier 09_regex: match, substitution, and the delimiter rules"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:52:57.495477512Z
updated: 2026-10-23T19:18:30.13412368Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  commits: 1
  started_at: 2026-09-21T20:21:41.729980778Z
  ended_at: 2026-09-21T20:47:29.636384609Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:41.729980778Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:47:29.636384609Z
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

**Introduces:** match, substitution, binding, delimiters

Depends on nothing from `08_references` -- one of the three independent
pairs the spec names as evidence that the ordering is a partial order rather
than a chain.

Delimiters are the substance: `m//`, `m{}`, `m##`, and the rule that a
paired delimiter nests while an unpaired one does not.

The lexer already reads `if (/TIE/)` as `Quote("/TIE/")` rather than
division, which is the match-vs-divide ambiguity this tier pins down with
files rather than leaves as a note.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/09_regex` -> `TestCorpus` over the tier-09 topics (adjacency-09_regex.md, read-position.md, regex-delimiters.md, regex-matching.md, regex-pattern-positions.md, regex-quote-likes.md)). Left unchanged: `TestTierRegexAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierRegexAdjacency` -> `TestAdjacencyCoversItsTier/09_regex`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Every file in 09_regex/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierRegexPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierRegexPerlValidated'`)
- [ ] The dependency lint passes over 09_regex/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierRegexLint$' -count=1 -v | grep -q '^--- PASS: TestTierRegexLint'`)
- [ ] The tier's adjacency topic puts every construct 09_regex introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^09_regex$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/09_regex'`)
- [ ] 09_regex/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierRegexRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierRegexRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-09_regex|read-position|regex-delimiters|regex-matching|regex-pattern-positions|regex-quote-likes)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
