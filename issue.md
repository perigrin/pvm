---
title: The adjacency prerequisite check passes under the wrong declaration, in three tiers
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T21:37:00.824105072Z
updated: 2026-09-22T00:45:50.476180066Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  commits: 1
  started_at: 2026-09-21T22:29:10.522340582Z
  ended_at: 2026-09-21T23:31:55.550767895Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.522340582Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T23:31:55.550767895Z
observed_paths:
- conformance/01_literals/README.md
- conformance/02_variables/README.md
- conformance/03_context/README.md
- conformance/04_operators/README.md
- conformance/05_scoping/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/10_undeclared_callee.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/09_regex/README.md
- conformance/10_io/README.md
- conformance/11_oo/README.md
- conformance/12_packages/README.md
- conformance/13_opaque/README.md
- conformance/14_recursive/README.md
- internal/conformance/lint.go
- internal/conformance/lint_test.go
- internal/conformance/numbering_test.go
- internal/conformance/prereq_test.go
- internal/conformance/subops_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier01_test.go
- internal/conformance/tier02_test.go
- internal/conformance/tier03_test.go
- internal/conformance/tier04_test.go
- internal/conformance/tier05_test.go
- internal/conformance/tier06_test.go
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/conformance/tier08_test.go
- internal/conformance/tier09_test.go
- internal/conformance/tier10_test.go
- internal/conformance/tier11_test.go
- internal/conformance/tier12_test.go
- internal/conformance/tier13_test.go
- internal/conformance/tier14_test.go
---

Every tier's adjacency test asks: does `00_adjacency.t` emit an op that the
tier's DECLARED prerequisite introduces? The declaration is read from the
README via `readTierDeps`, never computed as N-1, which was the point.

The check is too weak, and three tiers found it independently.

MEASURED, by mutation, in each case by repointing the README's DEPENDS ON
block and watching the test stay GREEN:

    tier 12   07_subroutines -> 11_oo          green (method_named, stub)
    tier 11   08_references  -> 07_subroutines green (entersub)
    tier 04   03_context     -> 02_variables   green (aassign, aelemfast,
                                                      padav)

An adjacency file composes the tier's constructs, so it legitimately
touches SEVERAL tiers' ops. "Emits an op the declared prerequisite
introduces" is therefore satisfied by almost any plausible declaration, and
the test asserts nothing about what the README actually says.

That is worst precisely where it matters most. Tiers 09, 10, 11, 12 and 14
each declare a prerequisite that is NOT N-1, and the spec calls those out
as the interesting cases -- `09_regex` needs nothing from `08_references`,
`12_packages` needs nothing from `11_oo`. A check that passes under the
wrong declaration cannot defend the claim it was written for.

THE PER-TIER FIX, applied in tiers 04, 11 and 12: pin the declaration
itself beside the pairing, e.g.

    const declaredPrerequisite = "08_references"

That works and is proven by mutation. It is also a constant in a test file
restating what the README says -- two lists that must agree, which this
milestone has watched drift twice already.

WHAT A CORPUS-LEVEL FIX WOULD NEED: a pairing assertion strong enough that
the declaration cannot be swapped. Candidates, none obviously right:

1. Require an op the prerequisite introduces that NO OTHER reachable tier
   introduces. Fails where a prerequisite's ops are all shared.
2. Require the pairing op to be absent from the tier's own construct files,
   so the adjacency file is where the two tiers first meet. Tier 02 used a
   version of this for spellings.
3. Assert on the declaration only, and treat the op pairing as
   documentation. Weakest, but honest about what ops can prove.
4. Check the tier's files against EVERY tier's INTRODUCES set and require
   the declared prerequisite to be the LATEST one actually reached.

Option 4 is closest to the spec's intent -- the declaration should name the
latest tier the adjacency file genuinely needs -- but it interacts with
`reachable()` and wants measuring before being assumed an improvement.

Found while executing 01a0c362-0ff3, 01a0c362-0daa and 01a0c361-a317.

## Measured, across all fourteen tiers

Each candidate was run against the real corpus before anything was
written. The numbers decided it.

OPTION 1 -- an op of the prerequisite that no other reachable tier
introduces -- is VACUOUS. Measured: 103 ops across the fourteen
INTRODUCES blocks, 0 of them claimed by more than one tier. Every op
already names exactly one tier, so "no other tier introduces it" is
satisfied by every op the current check already accepts. It IS the
current check.

OPTION 2 -- the pairing op must be absent from the tier's own construct
files -- fails 10 of the 14. Only 02, 04 and 05 have an op of their
prerequisite that appears in the adjacency file alone; 03, 06, 08, 09,
11, 12, 13 and 14 share every such op with a construct file, and 07 and
10 emit NO op of their declared prerequisite at all. Adopting it would
condemn ten tiers for composing normally.

OPTION 4 -- the declared prerequisite must be the LATEST tier the
adjacency file actually reaches -- fails 4 of the 14, and two of the
four are cases where the rule CONTRADICTS a README's stated argument:

    07_subroutines  declares 06_control      reaches only 02_variables
    09_regex        declares 01_literals     reaches 04_operators (`not`)
    10_io           declares 03_context      reaches 08_references
    12_packages     declares 07_subroutines  reaches 11_oo

`09_regex` reaches tier 04 through a single `not`, which perl emits for
the `!~` in `$s !~ m{zzz}` -- an artefact of the optimiser, not a
dependency on operators. `12_packages` reaches tier 11 through
`method_named` and `stub`, emitted for `Greet->import("tag")`, which is
a plain sub call written with an arrow. That file's comment argues at
length that its prerequisite is 07 and that "pairing with tier 11
instead would assert nothing at all -- this tier needs nothing from it."
Option 4 would force the declaration the README refutes.

So all three op-based options are dead, and for one reason the corpus
had already written down: ops LINT a declared tier and cannot derive
one. The same lesson is recorded in `TestEveryTierHasAnAdjacencyFile`
and in `lintOps`.

That leaves option 3, and the measurement is what makes it the honest
choice rather than the lazy one.

## Acceptance criteria

- [x] The corpus's declared prerequisite edges are stated ONCE, beside `specTiers` in `lint_test.go`, as a hand-written list with the same documented reason `specTiers` has: a list read from the thing it checks cannot report an absence. Every one of the fourteen tiers has an entry. (`go test ./internal/conformance/ -run TestSpecPrerequisitesCoverEveryTier -count=1 -v`)
- [x] A corpus-level check fails when ANY tier's README `DEPENDS ON` block disagrees with the spec's edge for that tier, naming both the declared and the expected tier. (`go test ./internal/conformance/ -run TestTierPrerequisitesMatchTheSpec -count=1 -v`)
- [x] The check is proven by mutation against all three cases the issue measured as slipping through the op pairing -- tier 12 repointed to `11_oo`, tier 11 to `07_subroutines`, tier 04 to `02_variables` -- each caught, from a README fixture rather than by editing the real corpus. (`go test ./internal/conformance/ -run TestTierPrerequisiteMutationsAreCaught -count=1 -v`)
- [x] The check runs over the real corpus and is green, so it is a gate and not a table. (`go test ./internal/conformance/ -run 'TestTierPrerequisites|TestSpecPrerequisites' -count=1 -v`)
- [x] `lint.go` is unchanged, and no tier README is edited to make the check pass. (`git diff --stat -- internal/conformance/lint.go conformance/`)
- [x] The whole conformance suite stays green. (`go test ./internal/conformance/ -count=1`)
