---
title: "Tier 14_recursive: re-enter Perl inside a delimiter"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-127c-73c1-8bb8-a4bf8b66c735
created: 2026-09-21T09:53:00.137066905Z
updated: 2026-10-23T23:40:18.221227332Z
sessions:
- start_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  end_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  commits: 1
  started_at: 2026-09-21T20:48:38.565352142Z
  ended_at: 2026-09-21T21:10:56.024886019Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:48:38.565352142Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:10:56.024886019Z
observed_paths:
- conformance/03_context/00_adjacency.t
- conformance/03_context/05_sort.t
- conformance/03_context/README.md
- conformance/08_references/00_adjacency.t
- conformance/08_references/06_deref_at.t
- conformance/08_references/10_deref_at_sigil.t
- conformance/08_references/11_deref_postfix.t
- conformance/08_references/README.md
- conformance/12_packages/00_adjacency.t
- conformance/12_packages/03_use_pragma.t
- conformance/12_packages/05_use_empty_list.t
- conformance/12_packages/04_use_import.t
- conformance/12_packages/06_require_bareword.t
- conformance/12_packages/07_require_expression.t
- conformance/12_packages/08_import_call.t
- conformance/12_packages/README.md
- conformance/14_recursive/05_hostile_contents.t
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier03_test.go
- internal/conformance/tier08_test.go
- internal/conformance/tier12_test.go
- internal/conformance/tier14_test.go
---

**Introduces:** s///e, (?{ }), qr// with embedded code

Where the lexer must RE-ENTER Perl inside a delimiter, which is harder than
tier 13's delimit-only problem and structurally different. That difference
is why these are two tiers rather than one.

The last tier, and the one most likely to refuse wholesale. That is the
expected outcome, not a failure: the corpus states where the parser is, and
a tier that refuses entirely is a true statement about where it is.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/14_recursive` -> `TestCorpus` over the tier-14 topics (adjacency-14_recursive.md, embedded-code.md, eval-replacement.md)). Left unchanged here: `TestTierRecursiveAdjacency`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestTierRecursiveAdjacency` -> `TestAdjacencyCoversItsTier/14_recursive`. These commands pass once branch ccx/adj merges.

Criteria updated 2026-10-23 (review round 1, agent:ccx-rev1): `TestAdjacencyCoversItsTier` no longer checks the prerequisite pairing, which any plausible declaration satisfied. The pairing claim moved to its own criterion, run by `TestTierPrerequisitesMatchTheSpec`, which fails when a tier's DEPENDS ON is repointed.

## Acceptance criteria

- [ ] Every file in 14_recursive/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierRecursivePerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierRecursivePerlValidated'`)
- [ ] The dependency lint passes over 14_recursive/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierRecursiveLint$' -count=1 -v | grep -q '^--- PASS: TestTierRecursiveLint'`)
- [ ] The tier's adjacency topic puts every construct 14_recursive introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^14_recursive$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/14_recursive'`)
- [ ] The tier's declared prerequisite is the spec's edge rather than an assumed tier N-1 (`go test ./internal/conformance/ -run '^TestTierPrerequisitesMatchTheSpec$' -count=1 -v | grep -q '^--- PASS: TestTierPrerequisitesMatchTheSpec'`)
- [ ] 14_recursive/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierRecursiveRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierRecursiveRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-14_recursive|embedded-code|eval-replacement)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
