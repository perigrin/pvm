---
title: "Tier 03_context: scalar vs list, and where the inference gap becomes a number"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
blocks:
- 01a0c361-a317-7807-bb9e-15f299562244
created: 2026-09-21T09:52:19.065451383Z
updated: 2026-10-23T17:09:33.256391975Z
sessions:
- start_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  end_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  commits: 1
  started_at: 2026-09-21T20:48:38.350029802Z
  ended_at: 2026-09-21T21:10:55.499016614Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:48:38.350029802Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:10:55.499016614Z
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

**Introduces:** scalar vs list context, the discriminating pairs

Context sits immediately before operators and no earlier. The moment
operators arrive you cannot proceed without it: `@a + 1` is arithmetic on a
count and `"$x" . @a` is not, and the operator decides which.

## This tier is where a known inference gap becomes measurable

`types.NarrowByContext` is called at `infer.go:524` with NO source of
context: the CST has no `Wantarray` node and no call-site propagation. The
discriminating pairs are the files that make that gap a number rather than a
note.

chromatic opens *Modern Perl* with context, before any syntax, because a
READER needs the frame immediately. A PARSER needs it when operators arrive.
Same fact, different consumers -- and this corpus is ordered for the parser.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/03_context` -> `TestCorpus` over the tier-03 topics (adjacency-03_context.md, aggregates-in-context.md, context-from-inside.md, list-operators-in-context.md, range.md)). Left unchanged: `TestTierContextAdjacency`, because no current test holds that property.

## Acceptance criteria

- [ ] Every file in 03_context/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierContextPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierContextPerlValidated'`)
- [ ] The dependency lint passes over 03_context/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierContextLint$' -count=1 -v | grep -q '^--- PASS: TestTierContextLint'`)
- [ ] The tier's adjacency file puts every construct 03_context introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierContextAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierContextAdjacency'`)
- [ ] 03_context/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierContextRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierContextRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-03_context|aggregates-in-context|context-from-inside|list-operators-in-context|range)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
