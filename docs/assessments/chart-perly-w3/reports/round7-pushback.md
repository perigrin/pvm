# Assess round 7: pushback lens, raiser dispositions (assess-pushback)

**Subject:** revision 7, commit 9afff6c63, `docs/plans/2026-11-23-chart-stateless-precedence.md`. Its blob is ce616721c9ed717dc119cf4c5f3329daf802f4cd (`git rev-parse` at the commit). I read it as the diff from 788778ed0 plus the changed sections.

I act only as raiser. I am also the drafter of the minute, so I am biased toward finding my own bar met.

## P19: released

I release P19 against §3.4's "The inputs, and where each runs" table:

> "The superset population's perl.git files | `make chart-precedence-reference`, behind its own build tag | perl.git at 94e50866"

and its bullets:

> "The perl.git target is not run by `make test`. CI's corpus-backed Ratchets job runs it with `PERL5_CORPUS` set, and fails unless it passes … A developer without perl.git runs `make test`. The perl.git part is checked in CI."

Every input `make test` runs is marked "Corpus needed: none". `make test` no longer depends on perl.git, and CI still fails on a missing corpus. That matches the repo's split and perigrin's "Split it".

## New in my lens

### P20 (low; §2): "most of the pipeline's time" contradicts the table beneath it

**The contradiction.** §2's first bullet says `Constrain` "takes most of the pipeline's time". The table directly below gives its share as 43% (`constrain-runs/`) and 39% (`combined-runs/`).

**If it ships:** a reader of the decision alone takes `Constrain` to be more than half the pipeline. That overstates what V1 can buy, in the section the evidence index exists to back.

**Fix:** use one word such as "the largest share", or "about two-fifths".

**What would change my mind:** nothing. It is unconditional, but it is one word and I do not hold acceptance on it.

### Considered and dropped

- **"If it proves slow, the bulk moves behind the build tag."** This is a requirement with no measure of "slow". I dropped it because the fallback is stated: `make test` keeps the refused sentences and a sample. The refused path stays in `make test` whichever way an implementer judges it, so I could not name a Y.
- **The superset population's in-tree part needs perl labels.** `supersetPopulation` errors on a file with no perl label (`population.go`). Whether those labels are committed for T1, Guacamole and std5k belongs to the codebase lens, and I did not establish a Y.
