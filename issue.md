---
title: "Tier 07 call-form slice: the smallest useful corpus gate"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-a317-7807-bb9e-15f299562244
created: 2026-09-21T13:41:10.741737124Z
updated: 2026-10-23T23:40:18.268369832Z
sessions:
- start_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  end_sha: b41801021927f68fcbf799c60d6b947393d983de
  commits: 1
  started_at: 2026-09-21T21:39:35.811982145Z
  ended_at: 2026-09-21T22:10:47.924455531Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:39:35.811982145Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T22:10:47.924455531Z
observed_paths:
- conformance/02_variables/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/00_adjacency.t
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/09_prototype_extent.t
- conformance/07_subroutines/11_code_ref_call.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/11_oo/README.md
- conformance/13_opaque/README.md
- conformance/README.md
- internal/conformance/claimonce_test.go
- internal/conformance/conformance_test.go
- internal/conformance/file.go
- internal/conformance/hardmarkers_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/parse/easy_test.go
- internal/parse/t2_test.go
---

**Introduces:** call forms -- `f(1)`, `f 1`, `&f`, `$code->()`, method-call
syntax

## Why this is split out from the rest of tier 07

The spec names the delivery milestone: **"the smallest useful corpus is
tiers 01-04 plus a call-form slice from tier 07"**. Call forms are the
slice. Declaration, `@_`, `return` and signatures are the rest, and they are
not on the path to that milestone.

The reason the slice is what matters: **the two largest T2 Unknown buckets
are both parenless call argument position.** A parenless call to an unknown
callee is where the parser must decide how much to consume, and it is the
thing that makes a gate on tiers 01-04 alone insufficient.

The spec's own words on why 01-03 is not enough: "tier 01 is one landed
issue away rather than satisfied, which means a gate on tiers 01-03 says
almost nothing. Tier 04 is the first tier where a parse can be WRONG rather
than Unknown, and the two largest T2 buckets are parenless call argument
position, which needs the 07 slice."

## What each form decides differently

    f(1)          parenthesised, unambiguous
    f 1           parenless -- how much does it consume?
    &f            no argument list at all, inherits @_
    $code->()     the callee is a value, not a name
    $obj->m(1)    the callee is resolved at run time

A corpus file per form, plus the adjacency file that puts them side by side.

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestTierCallFormsParenlessExtent` -> the greedy and `($)`-prototype cases of `argument-extent.md`, under `TestCorpus`; `TestTierCallFormsPerlValidated` -> `TestTierSubroutinesPerlValidated`; `TestTierCallFormsLint` -> `TestTierSubroutinesLint`; `TestTierCallFormsRefusalsCited` -> `TestTierSubroutinesRefusalsCited`). Left unchanged here: `TestTierCallFormsAdjacency`; the note below maps each to the test that now holds it.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, and that each case's main program is multi-statement; the declared prerequisite is pinned by `TestTierPrerequisitesMatchTheSpec` instead. Old -> new: `TestTierCallFormsAdjacency` -> `TestAdjacencyCoversItsTier/07_subroutines`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] A parenless call to an unknown callee has its argument extent measured against perl rather than assumed (`go test ./internal/conformance/ -run '^TestCorpus$/^argument-extent\.md$/(greedy|prototype_cuts)' -count=1 -v | grep -c '^    --- PASS: TestCorpus/argument-extent\.md/.*\(greedy\|prototype_cuts\)' | grep -qx 2`)
- [ ] Every file in 07_subroutines/ that this issue writes is validated against pinned perl 5.42.0 (`go test ./internal/conformance/ -run '^TestTierSubroutinesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesPerlValidated'`)
- [ ] The dependency lint passes over the call-form files (`go test ./internal/conformance/ -run '^TestTierSubroutinesLint$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesLint'`)
- [ ] The adjacency topic places the call forms side by side in one body, per the spec's rule (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^07_subroutines$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/07_subroutines'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierSubroutinesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierSubroutinesRefusalsCited'`)
- [ ] The smallest-useful-corpus gate is met: tiers 01-04 plus this slice all green (`go test ./internal/conformance/ -run '^TestSmallestUsefulCorpus$' -count=1 -v | grep -q '^--- PASS: TestSmallestUsefulCorpus'`)
