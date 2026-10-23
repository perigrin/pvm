---
title: "Tier 12_packages: use and require, heavier than the object system"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-073a-78c2-ba73-227fb48a26c4
created: 2026-09-21T09:52:58.866189072Z
updated: 2026-10-23T19:18:30.561708117Z
sessions:
- start_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  end_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  commits: 1
  started_at: 2026-09-21T20:48:38.508464179Z
  ended_at: 2026-09-21T21:10:55.861543164Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:48:38.508464179Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:10:55.861543164Z
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

**Introduces:** package, use, require, imports

Later than the object system because these are the heavier construct: they
locate and load a FILE, then call `import`, which needs subroutines (tier
07) to mean anything. A tier list that puts them before the object system is
ordering by how a program is usually WRITTEN rather than by what the grammar
depends on.

Depends on nothing from `10_io` -- another of the independent pairs.

## The authoring constraint this tier explains

Of the 35 T1 files matching tier 04, 31 `use Test::More`. Every extracted
tier-04 file would contain this tier's construct and fail the dependency
check by construction. That is why everything is authored and T1 is a source
of CASES, cited by filename and pinned revision, never copied.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/12_packages` -> `TestCorpus` over the tier-12 topics (adjacency-12_packages.md, compile-phase.md, loading.md, packages.md)). Left unchanged: `TestTierPackagesAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierPackagesAdjacency` -> `TestAdjacencyCoversItsTier/12_packages`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Every file in 12_packages/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierPackagesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesPerlValidated'`)
- [ ] The dependency lint passes over 12_packages/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierPackagesLint$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesLint'`)
- [ ] The tier's adjacency topic puts every construct 12_packages introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^12_packages$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/12_packages'`)
- [ ] 12_packages/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierPackagesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-12_packages|compile-phase|loading|packages)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
