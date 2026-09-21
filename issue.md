---
title: A tier claims an op once is unenforced; twelve collisions were resolved by hand
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:43:03.717841424Z
updated: 2026-09-21T21:44:13.397884326Z
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

## Why a duplicate claim is not cosmetic

`reachable` builds `claimedBy[op] = name` as it walks the tiers, so the
second claimant overwrites the first and WHICH tier the lint reports as
owner depends on Go's map iteration order. Worse, a duplicate claim
widens the earlier tier: an op claimed by both 02 and 08 is allowed from
02 on, so a file that genuinely needs 08 passes its lint at 02. The
collision is invisible by construction, and it loosens the lint.

## Acceptance

- [ ] No op in the corpus is introduced by more than one tier (`go test ./internal/conformance/ -run '^TestOpClaimedByExactlyOneTier$' -count=1 -v | grep -q '^--- PASS: TestOpClaimedByExactlyOneTier'`)
- [ ] A collision in a fixture corpus is caught, and the failure names the op and BOTH claiming tiers rather than one of them (`go test ./internal/conformance/ -run '^TestOpCollisionsNamesBothTiers$' -count=1 -v | grep -q '^--- PASS: TestOpCollisionsNamesBothTiers'`)
- [ ] Every collision is reported in one run, not the first -- twelve at a time is the realistic case, and a check naming one per run makes the fix twelve runs long (`go test ./internal/conformance/ -run '^TestOpCollisionsReportsAllAtOnce$' -count=1 -v | grep -q '^--- PASS: TestOpCollisionsReportsAllAtOnce'`)
- [ ] The check reads the tier READMEs through `readTierOps`, the same input the lint uses, so it cannot drift from a second hand-kept list (`grep -q 'readTierOps(corpusDir)' internal/conformance/claimonce_test.go`)
- [ ] The rest of the conformance suite is unaffected (`go test ./internal/conformance/ -count=1`)
