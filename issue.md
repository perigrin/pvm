---
title: A tier claims an op once is unenforced; twelve collisions were resolved by hand
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:03.717841424Z
updated: 2026-09-21T21:40:49.587869779Z
sessions:
- start_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T21:40:49.587869779Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:40:49.587869779Z
---

`TestCorpusLints` checks that every op a README CLAIMS is emitted by some
file in that tier, and that no file emits an op claimed by a LATER tier.
It does not check that an op is claimed by only ONE tier.

So two tiers can both claim the same op and both pass. That is not
hypothetical: writing the fourteen tiers at 0ce515cb produced twelve ops
claimed by two or more tiers each, all of which had to be resolved by hand:

    sassign               02 and 04
    multideref            02 and 08
    rv2av rv2hv           02 and 08
    av2arylen             02 and 08
    gv                    02 and 10
    and or                04 and 06
    not                   04 and 09
    enterloop leaveloop   05, 06 and 11
    anoncode              07 and 08
    entersub leavesub     07 and 11
    ref                   08 and 11
    shift                 02 and 11

The rule applied was "the tier that emits it first owns it", and each
later tier's README now explains why it REACHES that op without
INTRODUCING it -- which turned out to be the more interesting prose in
several cases (tier 06 reaching tier 04's `and`/`or` for `if`/`unless`
is the clearest).

Nothing enforces the rule. A fifteenth tier, or an edit to an existing
README, could silently re-claim an op and the suite would stay green.

The check is cheap: read every tier's INTRODUCES set, assert the sets are
pairwise disjoint, and name both tiers when they are not. It belongs
beside `TestOpTableDerivedFromReadmes`, which already reads all fourteen
sets for its own purposes.
