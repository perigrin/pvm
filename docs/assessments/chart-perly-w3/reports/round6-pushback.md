# Assess round 6: pushback lens, raiser dispositions (assess-pushback)

**Subject:** revision 6, commit 788778ed0, `docs/plans/2026-11-23-chart-stateless-precedence.md`. Its blob is 60327eb7d7e379443281e0623b5d5f586a382f9c (`git rev-parse` at the commit). I read it via the diff from 2782fd181 plus the changed sections.

I act only as raiser. I am also the drafter of the minute, so I am biased toward finding my own bar met.

## P18: released

I release it against two passages.

§1:

> "'Target for V1 then V2 (Recommended)': 3.92 is where this decision and the later V2 decision should land together. It is not this decision's exit. V1 measured 4.12–4.16 on the machine where V2 measured 3.92."

§5, "What the result decides":

> "the V2 decision comes next, toward 3.92 for V1 and V2 together."

The figure now has a stated relation to this decision's exit, and it is attributed to V2's machine.

## New in my lens

### P19 (medium-high; §3.4 "Where it runs"): `make test` would fail on every machine without the perl.git corpus

**What the revision says.** §3.4 puts the aggregate differential in `make chart-precedence-reference`. It says `make test` runs it after the suite, "following the precedent of `make chart-absence-reference`". It adds: "When its corpora are missing, the target fails; it never skips." §3.4 also says the aggregate "needs the T1 and perl.git corpora".

**If it ships:** `make test` fails on any contributor's or agent's machine that has no perl5 tree at `PERL5_CORPUS` or `/opt/corpora/perl5`. pvm's CLAUDE.md requires `make test` to pass 100% before every commit, so every commit there is blocked. The other outcome is that someone softens the target to a skip mid-issue, which is §5's "criterion change" stop.

**Because:**

- **The precedent does not carry over.** The Makefile's `test` target is `go test ./... -count=1`, then `$(MAKE) chart-absence-reference`. That target needs no external corpus. CI says so: "Its corpus is in the tree (T1's in-tree default), so it needs no env" (`.github/workflows/ci.yml`, the absence-reference step).
- **The perl.git half needs an external tree.** It is read from `PERL5_CORPUS`, falling back to `/opt/corpora/perl5` (`cmd/ratchet/oracle.go`, `perlgit.go` `perl5Root`).
- **The repo's convention is the opposite.** CI says: "Each corpus-backed ratchet skips when its corpus is absent, which is correct for a contributor and useless as a gate … This job supplies the corpora and fails if any ratchet skips."
- **CI's matrix job is not exposed.** It runs `go test ./...`, not `make test`, so the failure lands on local `make test` only.

**Modify:** follow the repo's split. In `make test`, the target skips with a printed notice when the corpora are absent. The corpus-backed Ratchets job runs it and fails if it skipped, as the "Fail if a corpus-backed ratchet skipped" step does today. Alternatively, keep the aggregate part out of `make test` and run it only in that job.

**What would change my mind:** perigrin rules that every development machine must hold the perl.git corpus. In that case this reduces to stating that requirement.
