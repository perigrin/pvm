# Assess round 8: pushback lens, raiser dispositions (assess-pushback)

**Subject:** revision 8, commit 336067853, `docs/plans/2026-11-23-chart-stateless-precedence.md`. Its blob is ed91e46239b29283ad5c82cd390411621b52a2cd (`git rev-parse` at the commit). I read it as the diff from 9afff6c63 plus the changed sections.

I act only as raiser. I am also the drafter of the minute, so I am biased toward finding my own bar met.

**Recommendation, as raiser: accept.** I hold nothing and raise nothing new.

## P20: released

I release P20 against §2, "Where the time goes": "`Constrain` takes the pipeline's largest share." That agrees with the table beneath it, which gives 43% and 39%.

## New in my lens: none

I re-read every section revision 8 changed. Nothing new survived the "Y happens, because Z" test. These are the points I checked:

- **§1 and §4 together.** The exit "read as §4 defines" now has one definition. §4 defines "one measurement" as one ratchet run's printed value, and records the exit key itself (`speed-ratio-aggregate-longlived`) as well as the stage keys. The relative exit and the third stop each have a single bar.
- **§3.4's table.** Every `make test` row still needs no corpus. The perl.git row runs only in its own target and in CI, which is where P19's release rests. Each new test is now marked "new".
- **§3.4, "If a `make test` check proves slow".** It now covers every `make test` row. For a corpus row, "keeps the refused sentences and a sample" reads loosely, because a corpus row has no "refused sentences" of its own. I could not name a consequence: the referee's and the generators' refused sentences stay in `make test` either way. Dropped.
- **§5's stops.** The first stop now covers a moved digest or count key. With the middle-outcome, quiet-session, timeout and budget stops, I found no outcome that meets neither the exit nor a stop.
- **§6.** The paths outside `internal/chart` match §1's "Rule them in (Recommended)". §3's "What this does not amend" names the one departure from v2 §11.
