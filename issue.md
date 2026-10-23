---
title: "Tier 08_references: deref, arrow, anonymous constructors"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
blocks:
- 01a0c362-0daa-7422-b787-7650ce684c4b
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:57.12910003Z
updated: 2026-10-23T17:09:33.823043333Z
sessions:
- start_sha: b4af7e5155e5d0a16b7b29683f2f2731eb3074c8
  end_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  commits: 1
  started_at: 2026-09-21T20:48:38.443049149Z
  ended_at: 2026-09-21T21:10:55.64673177Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:48:38.443049149Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:10:55.64673177Z
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

**Introduces:** `\`, deref, arrow, anonymous constructors, `${ }`, `@{ }`

After subroutines, following *Modern Perl*: dereference syntax is harder
than a sub declaration even though `\@a` alone looks simpler.

Two `hardMarkers` place here -- deref-brace and deref-at.

## This tier is a prerequisite for 11_oo, which is the non-obvious edge

`bless` takes a REFERENCE and a string. Every blessed-hash object is this
tier plus a string, which is why the object tier depends on this one and NOT
on packages. That dependency is real and satisfied three tiers earlier.

A known defect to cover rather than discover: a brace before `->` is a
hashref, and the tree currently misreads two anonhash cases (M1 issue
01a0bae2).

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/08_references` -> `TestCorpus` over the tier-08 topics (adjacency-08_references.md, code-references.md, dereferencing.md, taking-references.md)). Left unchanged: `TestTierReferencesAdjacency`, because no current test holds that property.

## Acceptance criteria

- [ ] Every file in 08_references/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierReferencesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesPerlValidated'`)
- [ ] The dependency lint passes over 08_references/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierReferencesLint$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesLint'`)
- [ ] The tier's adjacency file puts every construct 08_references introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestTierReferencesAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesAdjacency'`)
- [ ] 08_references/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierReferencesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-08_references|code-references|dereferencing|taking-references)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
