---
title: "Tier 11_oo: bless and class, where the adjacency file catches the live bug"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c362-0929-795d-ac6b-c1740240e49e
blocks:
- 01a0c436-39e4-7560-9bf0-81d647e96aae
created: 2026-09-21T09:52:58.28262933Z
updated: 2026-10-23T23:40:18.034047784Z
sessions:
- start_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  end_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  commits: 1
  started_at: 2026-09-21T21:11:14.221446628Z
  ended_at: 2026-09-21T21:36:34.519402668Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:11:14.221446628Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:36:34.519402668Z
observed_paths:
- conformance/04_operators/00_adjacency.t
- conformance/04_operators/05_logical.t
- conformance/04_operators/README.md
- conformance/11_oo/00_adjacency.t
- conformance/11_oo/03_method_named.t
- conformance/11_oo/README.md
- internal/conformance/tier04_test.go
- internal/conformance/tier11_test.go
---

**Introduces:** bless AND class/field/method/ADJUST, indirect new

Comes BEFORE packages, reversing an earlier draft. Measured under 5.42.0:

    class Foo { method m { 1 } }     syntax OK, no `package` anywhere
    bless {}, "Foo"                  needs a REF and a STRING, nothing else

`class Foo {}` IS a package declaration -- a different spelling of tier 12's
construct, not a consumer of it -- so the old order was circular.
`__PACKAGE__` stays `main` outside the class body, so `class` introduces its
own scope rather than inheriting one.

## The adjacency file is what catches the live bug

Measured 2026-09-21: `class`, `field`, `method`, `ADJUST`, `:param` and
`:isa` each parse clean IN ISOLATION. **True per construct and false per
file.** `class Foo { ADJUST { 1 } method m { 2 } }` is ONE Unknown
swallowing both, and `t2_test.go:105-111` records 41 Unknowns across seven
`class/*.t` files with nothing in the chain owning ADJUST.

So this tier's construct files would ALL PASS over a live bug. Its adjacency
file is the one that fails. Write the adjacency files first here -- this is
the tier that proves why that third kind of file exists.

`indirect-new` is one of the twelve `hardMarkers` and places here.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/11_oo` -> `TestCorpus` over the tier-11 topics (adjacency-11_oo.md, bless-dispatch.md, class-feature.md, tie.md, ungated.md)). Left unchanged here: `TestTierOoAdjacency`, `TestTierOoAdjacencyCatchesAdjust`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestTierOoAdjacency` -> `TestAdjacencyCoversItsTier/11_oo`; `TestTierOoAdjacencyCatchesAdjust` -> `TestCorpusRatchet` plus the case's clean line in corpus.ratchet. These commands pass once branch ccx/adj merges.

Obsolete by perigrin's ruling 2026-10-23: `TestTierOoAdjacencyCatchesAdjust` asserted that the tier-11 adjacency case FAILS on the ADJUST-plus-sibling pair. The parser fixed that bug, so the case parses clean and the ratchet records it clean; the criterion now checks that instead.

Criteria updated 2026-10-23 (review round 1, agent:ccx-rev1): `TestAdjacencyCoversItsTier` no longer checks the prerequisite pairing, which any plausible declaration satisfied. The pairing claim moved to its own criterion, run by `TestTierPrerequisitesMatchTheSpec`, which fails when a tier's DEPENDS ON is repointed.

## Acceptance criteria

- [ ] The ADJUST-plus-sibling adjacency case now parses clean, and the corpus ratchet records it clean (`go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1 -v | grep -q '^--- PASS: TestCorpusRatchet' && grep -qxF -e '- adjacency-11_oo.md/The whole tier in one body' internal/conformance/testdata/corpus.ratchet`)
- [ ] Every file in 11_oo/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOoPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOoPerlValidated'`)
- [ ] The dependency lint passes over 11_oo/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOoLint$' -count=1 -v | grep -q '^--- PASS: TestTierOoLint'`)
- [ ] The tier's adjacency topic puts every construct 11_oo introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^11_oo$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/11_oo'`)
- [ ] The tier's declared prerequisite is the spec's edge rather than an assumed tier N-1 (`go test ./internal/conformance/ -run '^TestTierPrerequisitesMatchTheSpec$' -count=1 -v | grep -q '^--- PASS: TestTierPrerequisitesMatchTheSpec'`)
- [ ] 11_oo/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOoRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOoRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-11_oo|bless-dispatch|class-feature|tie|ungated)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
