---
title: The adjacency prerequisite check passes under the wrong declaration, in three tiers
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T21:37:00.824105072Z
updated: 2026-09-21T21:37:00.824105072Z
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
