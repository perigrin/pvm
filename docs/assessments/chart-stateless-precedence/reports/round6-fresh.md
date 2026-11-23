# Round 6, fresh participant

Subject: `docs/plans/2026-11-23-chart-stateless-precedence.md` at 788778ed0, blob
60327eb7d7e379443281e0623b5d5f586a382f9c. The blob is the same at the worktree's HEAD, 0d2b31610. origin/pu is
still 8c3317079.

I did not read the minute or `reports/`.

## What I checked, and what held

- **§2 figures against the indexed files.** These match:
  - 1,296 (`combined-runs/referee.txt`);
  - 25,490 and 4,395 (`toys.txt`);
  - 2,211 with 0 differences, mixed reduce pairs 6, and a maximum of 24 memo entries (`combined-runs/result-run*`,
    `combined-compare.txt`);
  - 5.62, and whole-pipeline V1/V0 of 0.740, 0.721 and 0.734 (`combined-runs/result-run1..3`);
  - 0.727 (`fresh5/result-t1.txt`: 4.1237 / 5.6701);
  - 5.49 and 43.4% (`constrain-measurement.md`).
- **The six mixed pairs.** I ran `combined.diff` in a scratch worktree with a one-off test that printed them:
  - `myterm: ( expr )` before `[`;
  - `myterm: ( )` before `[`;
  - `myterm: scalar` before `[` and before `{`;
  - `indirob: scalar` before `[` and before `{`.

  §3.1's summary is right.
- **I re-ran the round-4 generators on the `combined.diff` prototype:**
  - Seed 4 produced 20,000 random sentences plus 83 handcrafted ones, which is 20,083. 185 of the random ones
    are refused by precedence, and 2 more of the handcrafted ones are.
  - Seed 44 produced 15,000 chains, 989 of them refused by precedence.
  - The 11,574 corpus files gave 0 differences in V1 kept set, verdict, pipeline and readings, after the remapping
    in K10.
- **Symbols that exist:**
  - `forest.NewPrecedence` (precedence.go:44) and `ParseWith` (forest.go:106), which turns `prec.err` into a
    refusal (forest.go:134);
  - `chart.CountPrecedence` (chart/precedence.go:421) and `waypoint2Refusals` (chart/waypoint2_refusals_test.go:13);
  - `make chart-absence-reference` (Makefile:33, which `test` runs) and CI's "Run the absence-claim reference"
    step;
  - `speed-ratio-aggregate-longlived`, `baseline.txt` (4.7915), `TestForestDigestGolden` and
    `TestResolvedForestDigestGolden` (these are "both forest digests");
  - the README's section "The precedence constraint: `forest/precedence.go`".
- **v2 design.** §3 rule 12, rule 6 (carried over into §3), §2 and §8 say what the decision cites. The decision's
  test-based meeting of §8 ("the generated tests below are what pin it") is consistent with it.
- **Guard.** The prototype built the V1 tables for all 13 existing toys without tripping its guard (`toys.txt`
  passes), so the existing toys survive §3.2.

## Findings

**K1, §3.4 "Where it runs": as written, `make test` fails on any machine without a perl5 checkout at 94e50866.**

- **The precedent does not match.** CI's own comment on the absence reference says "Its corpus is in the tree
  (T1's in-tree default), so it needs no env". The ratchet aggregate's 562 perlgit files need `PERL5_CORPUS` or
  `/opt/corpora/perl5`, at the pin, clean (`perlgit.go` refuses otherwise).
- **The corpus is misnamed.** T1 is in the tree: `testdata/perlonjava/unit`, 986 files, the same names as the
  external copy. So only perl.git is external, not "the T1 and perl.git corpora".
- **Consequence.** pvm's CLAUDE.md makes `make test` at 100% non-negotiable. Contributors and agents without the
  corpus get a red `make test`, and the pressure is toward adding exactly the skip the decision forbids.
- **CI.** The CI step needs `PERL5_CORPUS` set. The absence step sets no env.
- **Fix.** Use the repo's existing pattern: `PVM_REQUIRE_PERL542` skips locally and fails in CI. So the target
  skips without the corpus unless `PVM_REQUIRE_...` is set, CI sets it, and CI greps for the PASS line as the
  absence step does. Alternatively, keep the aggregate out of `make test`, or say plainly that `make test` now
  requires perl5 at the pin.

**K2, §3.4 inputs: "the ratchet aggregate" is defined by timing.**

- `Protocol.Aggregate` (speed.go:93) times recursive descent on every population file. It drops std5k, edit5k and
  any file over 10% of the total (code_too_large.t at 15.2%).
- **Consequence.** A correctness test does a timing pass, and its membership is load-dependent in principle.
- **Fix.** The superset population (2,214) contains the aggregate, needs no timing and is deterministic. Name it
  instead.

**K3, §2: two figures do not reproduce as stated.**

- **"The precedence stage runs at 0.40–0.41 of today's".** The per-run V1/V0 precedence ratios are 0.392, 0.406
  and 0.423. The ratio of medians is 0.401, and fresh5 gives 0.397. It should read about 0.40 (0.39–0.42).
- **"2.2–2.4, about 43%" splices two sources.** `combined-runs` has V0 precedence at 2.18–2.23, which is 39% of
  5.62. `constrain-measurement.md` has 2.33–2.45 and 43.4%. Say which figure comes from which file.
- **Consequence.** These are cosmetic, but they sit in the section the evidence index exists to back.

**K4, §1: "V1 measured 4.12–4.16 on the machine where V2 measured 3.92" is selective.**

- The V1 run medians are 4.16, 4.03 and 4.17, and fresh5 gives 4.12.
- V2 is 4.02, 3.85 and 3.92, with a median of 3.92. But fresh5's independent run on the same machine put V2 at
  4.10, against V1's 4.12.
- The 3.92 target is perigrin's ruling, and this decision's exit does not depend on it.
- **Consequence.** "What the result decides" sets expectations for the V2 decision from a single three-run median.
  The only reproduction shows V2's increment at about 0.02. Record both readings so the V2 decision is not
  anchored on 3.92.

**K5, §4–§5: outcomes that fall between.**

- **Budget.** It is 7 days, with a stop at twice that. Nothing says what happens at 1×, between day 7 and day 14.
- **Quiet readings.** §4 requires readings where the load is at or below the core count. Nothing covers a session
  where that cannot be had. Is it a retry, a stop, or recorded with its load? The live README's convention
  (perigrin, 2026-11-15) records load rather than requiring quiet.
- **The before-state commit is not named.** It must include the stage keys, or the precedence stop has no
  before-reading. It should be the stage-key commit or the precedence change's parent.
- **Stage-key cost has no threshold or outcome.** §4 has the cost reported but says nothing if it is large. The
  clones of 2,211 forests per pass add heap, and `chart-ratchet.yml` has `timeout-minutes: 20`.

**K6, §3.4–§3.5: naming.**

- `Constrain` is a method, `(*Forest).Constrain(p *Precedence) bool` (precedence.go:322). "forest.ConstrainReference"
  reads as a package function, so say `(*Forest).ConstrainReference`.
- §3.5 lists `CountPrecedence` and the toy referees as running production. The callers that switch silently are
  these:
  - forest.go:138;
  - chart/precedence.go:327;
  - chart/precedence_test.go:95 (`constrainSource`, hence `TestWaypoint2Refusals`) and :242
    (`TestPrecedenceKeepsCorpusAccepted`);
  - forest/precedence_test.go:88 (`TestPrecedenceConstraintOrderInvariant` and the toy referees).

  That is the intended outcome, but say "every existing caller of `Constrain`".
- "The reference's own agreement with bison stays tested" needs a test that is new and named.

**K7, §3.4: the generator "reproduces the 185 and 989".**

- **Verified.** The counts are exact. 185 is of the 20,000 random sentences, not of the 20,083: the handcrafted
  sentences add 2 more.
- **Consequence if asserted as exact counts.** Any grammar change breaks the test, and the counts were measured on
  V0. Say "reproduced once, at commit", or assert a floor and print the counts.

**K8, §4: stage keys need API the decision does not name.**

- `resolve`, `countReadings` and `libraryOperators` are unexported, and the ratchet is package main. No Forest
  clone exists outside the scratch harnesses (`cloneF` and `CloneF`).
- The precedence stage's input must be the forest after `libraryOperators`, which "get no stage key".
- §6's "`internal/chart`" covers this, so it is allowed. A `forest.Bench`-like exported hook and the clone are new
  work and should say so.

**K9, INDEX: the 11,574-file run cannot be re-run from the record as written.**

- 3,627 of the 11,588 paths in `harness/fresh4/list.txt.gz` point at fresh4's deleted scratch worktree.
- Remapped to an 8c3317079 checkout, it gives 11,574 sources and 0 differences.
- The INDEX should say they are pu-8c3317079 repo paths.

## Not examined

- `reports/` and the minute, by instruction.
- `stateless.diff` and `no-map.diff` in depth.
- The `codebase-ov4` probes and `c1-report.md`.
- The recovery plan.
- The body of today's `Constrain` (whether its memo sits on `Precedence`).
- I did not re-time anything. Speed figures were checked by arithmetic only.
- I did not re-run the 2,211-file aggregate. I relied on the four recorded runs.

## Verdict: modify

K1 is the only finding that would bite on shipping. The fix is a one-line change of policy, following the
`PVM_REQUIRE_PERL542` pattern, plus correcting "T1 and perl.git". K2, K5 and K6 are wording-sized. K3 and K4 are
evidence hygiene. The substance (keys, guard, reference, inputs) held up under re-run.
