---
title: "Tier 10_io: filehandles, readline, and print as subject rather than fixture"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:52:57.867463731Z
updated: 2026-10-23T19:18:30.22968557Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  commits: 1
  started_at: 2026-09-21T20:21:42.010031474Z
  ended_at: 2026-09-21T20:47:29.775868789Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:42.010031474Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:47:29.775868789Z
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

**Introduces:** filehandles, readline, print

`print` appears from tier 01 onward as a fixture -- every behavioural probe
uses it -- and this is where it becomes the subject. Same relationship `my`
has to tier 05.

`glob-angle` is one of the twelve `hardMarkers`, and it places at tier 13
rather than here: `<*>` is a delimiting problem, `<$fh>` is an IO one, and
they are different tiers despite the shared spelling. Files here should make
that split explicit.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/10_io` -> `TestCorpus` over the tier-10 topics (adjacency-10_io.md, handles.md, symbol-table.md, writing.md)). Left unchanged: `TestTierIoAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierIoAdjacency` -> `TestAdjacencyCoversItsTier/10_io`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Every file in 10_io/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierIoPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierIoPerlValidated'`)
- [ ] The dependency lint passes over 10_io/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierIoLint$' -count=1 -v | grep -q '^--- PASS: TestTierIoLint'`)
- [ ] The tier's adjacency topic puts every construct 10_io introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^10_io$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/10_io'`)
- [ ] 10_io/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierIoRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierIoRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-10_io|handles|symbol-table|writing)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
