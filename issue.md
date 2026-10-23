---
title: "Tier 02_variables: sigils, containers, element access"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c361-7479-7275-ace5-cd9a6dd2bd6c
- 01a0c362-0929-795d-ac6b-c1740240e49e
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:14.225829534Z
updated: 2026-10-23T23:40:17.416486923Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  commits: 1
  started_at: 2026-09-21T20:21:38.998307187Z
  ended_at: 2026-09-21T20:47:21.291091025Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:38.998307187Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:47:21.291091025Z
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

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/02_variables` -> `TestCorpus` over the tier-02 topics (adjacency-02_variables.md, aggregate-operators.md, arrays.md, hashes.md, names.md)). Left unchanged here: `TestTierVariablesAdjacency`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestTierVariablesAdjacency` -> `TestAdjacencyCoversItsTier/02_variables`. These commands pass once branch ccx/adj merges.

Criteria updated 2026-10-23 (review round 1, agent:ccx-rev1): `TestAdjacencyCoversItsTier` no longer checks the prerequisite pairing, which any plausible declaration satisfied. The pairing claim moved to its own criterion, run by `TestTierPrerequisitesMatchTheSpec`, which fails when a tier's DEPENDS ON is repointed.

## Acceptance criteria

- [ ] Every file in 02_variables/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierVariablesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesPerlValidated'`)
- [ ] The dependency lint passes over 02_variables/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierVariablesLint$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesLint'`)
- [ ] The tier's adjacency topic puts every construct 02_variables introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^02_variables$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/02_variables'`)
- [ ] The tier's declared prerequisite is the spec's edge rather than an assumed tier N-1 (`go test ./internal/conformance/ -run '^TestTierPrerequisitesMatchTheSpec$' -count=1 -v | grep -q '^--- PASS: TestTierPrerequisitesMatchTheSpec'`)
- [ ] 02_variables/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierVariablesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierVariablesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-02_variables|aggregate-operators|arrays|hashes|names)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
