---
title: "Tier 07_subroutines: call forms, and the parenless argument question"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
blocks:
- 01a0c362-0ff3-7f3f-b2d3-f6dbe8228f86
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:56.634555553Z
updated: 2026-10-23T19:18:29.952567002Z
sessions:
- start_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  end_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  commits: 1
  started_at: 2026-09-21T20:21:39.183806878Z
  ended_at: 2026-09-21T20:47:21.803800155Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:39.183806878Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:47:21.803800155Z
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

**Introduces:** declaration, `@_`, return, signatures

**Call forms are NOT here.** They are split into their own issue, because
the spec names "tiers 01-04 plus a call-form slice from tier 07" as the
smallest useful corpus, and burying that slice behind declaration and
signatures delays the milestone's one stated delivery target. This issue is
the rest of the tier.

`signature` is one of the twelve `hardMarkers` and places here.

## What is left once call forms are gone

    sub f { }             declaration, named
    sub { }               anonymous
    sub f ($x, $y) { }    signatures, 5.36+ and stable in 5.42
    @_                    the implicit argument array
    return / implicit     the last-expression rule

`@_` is the one worth care: it aliases, so `$_[0] = 1` modifies the
caller's variable. A file that treats it as a plain array measures nothing
about what makes it different.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/07_subroutines` -> `TestCorpus` over the tier-07 topics (adjacency-07_subroutines.md, argument-extent.md, arguments.md, subroutines.md); `TestTierSubroutinesArgAliasing` -> the `arguments.md` case on `@_` aliasing the caller's variables, under `TestCorpus`). Left unchanged: `TestTierSubroutinesAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierSubroutinesAdjacency` -> `TestAdjacencyCoversItsTier/07_subroutines`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Every file in 07_subroutines/ that this issue writes is validated against pinned perl 5.42.0 (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint passes over 07_subroutines/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] The tier's adjacency topic puts every construct it introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^07_subroutines$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/07_subroutines'`)
- [ ] `@_` aliasing is measured against perl rather than assumed, since that is what distinguishes it from an array (`go test ./internal/conformance/ -run '^TestCorpus$/^arguments\.md$/aliases_the_caller' -count=1 -v | grep -q '^    --- PASS: TestCorpus/arguments\.md/.*aliases_the_caller'`)
- [ ] 07_subroutines/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierSubroutinesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-07_subroutines|argument-extent|arguments|subroutines)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
