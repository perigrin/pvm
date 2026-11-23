# Round 5, fresh participant: chart-stateless-precedence

Subject: docs/plans/2026-11-23-chart-stateless-precedence.md, blob d97415a147b923df589b27bed3cad0317f6bbfe2. The blob is the same at 2782fd181 and at the worktree's HEAD, 157b0f820. I did not read the minute or reports/.

## What I did

- Checked every named function, test, fixture and key against pu 8c3317079: Constrain, NewPrecedence, CountPrecedence, waypoint2Refusals, the toy referees, both forest digests, precedence-disagree/covered, speed-ratio-aggregate-longlived, and the README precedence section.
- Read v2 §3 and §8.
- Applied combined.diff in a scratch worktree, which I have since removed. I ran TestZZCombined over the 2,211-file aggregate:
  - **Correctness:** reproduced exactly. V1 against V0 differs in 0 verdicts, 0 kept sets, 0 refusal reasons and 0 readings. The tables have 6 mixed pairs and 46 spine tops.
  - **Timing:** one process, 2 warm-up and 9 timed rounds, load ≤1.5 on 4 cores.

| Stage | V0 | V1 | V2 |
|---|---|---|---|
| ParseWith | 5.67 (5.39–6.19) | 4.12 (3.84–4.32) | 4.10 |
| precedence | 2.33 | 0.92 | 0.90 |

  V1's ParseWith is 0.727 of V0's, which confirms the 0.73. V1's precedence stage is 0.40 of V0's, which confirms the 0.41. V1's worst round is below V0's best. V2's ParseWith came out at 0.723 of V0's, not 0.69; that bears on the deferred decision, not on this one.

## Findings

**J1 §2 and header: the evidence directory does not hold what §2 says it holds.** §2 says "all of it is in the evidence directory". That is not true of:
- the timing figures 0.73, 0.72–0.74, 0.41 and 0.69. combined-compare.txt has no RESULT lines; the timing part of the harness needs CB_TIME, and its output was not kept;
- the 5.62. constrain-measurement.md reads 5.49;
- the 11,574 corpus files, the 20,083 operator sentences and the 15,000 Leo/nullable sentences. No harness or output for these is in the directory;
- the raw files that constrain-measurement.md lists (env-*, log-*, pprof, run.sh, harness/).

There is also an unexplained count: stateless-compare.txt says "alts built 2566149", and every other file says 2565144.

*If it ships:* perigrin cannot audit the number the ≤5 projection rests on, and the refusal-path evidence exists only in agent reports. My run confirms the V1 timing, so this is a record defect, not a wrong number. *Fix:* commit the RESULT output and the generators for the three missing populations, or say they are in reports.

**J2 §1: the rulings carry text from broader revisions.**
- "Do just B first" refers to B, the reduce-side-pruned forest in constrain-measurement.md. That is the V2 filter this decision defers.
- "V2: both (Recommended)" is superseded by "V1 now, V2 measured after", and nothing marks it as superseded.
- "Let's get ourselves to 3.92 first" names a number used nowhere in §5, whose exit is relative and ≤5.

*If it ships:* the milestone Records carry contradictory rulings, and an executor may treat 3.92 as the target or V2 as in scope. *Fix:* mark the superseded rulings, and say what 3.92 is: the exit, an expectation, or nothing.

**J3 §3.4: three of the five "identical output" items cannot differ under V1.**
- Expected is a property of the recognizer's chart (recognize/expected.go), and recognition does not change.
- A precedence refusal's Byte is always 0 (the shortcut in forest.go parseStream).
- The reason is "precedence" exactly when the verdict is false.

These items are left over from the V2 version. Comparing them requires running the whole pipeline with the oracle swapped in. The prototype did that with a package global, ConstrainVariant, which is production-side machinery that a "test-only oracle" should not need. *Fix:* compare the verdict, kept set and readings at Constrain's level, and say that Expected and Byte are unchanged by construction.

**J4 §3.4 and §3.5: the "test-only oracle" has no named home, and the aggregate differential has no named runner.** Today Constrain is called from three packages:
- forest (the toys);
- chart: precedence.go, which is not a test file (the referee and CountPrecedence), plus precedence_test.go and waypoint2Refusals;
- cmd/ratchet (the aggregate, through corpus(), perlGit and aggregateFiles).

An oracle in forest's _test.go is invisible to the last two, so it must be exported non-test code or live in a helper package. The aggregate also needs /opt/corpora:
- `make test` skips the perlgit tests when the perl5 tree is absent (perlgit_test.go);
- chart-ratchet is advisory and not part of `make test`.

*If it ships:* the strongest comparison may run only on the author's machine, and a skip reads as green. *Fix:* name the oracle's symbol and package, and name the test that runs the aggregate comparison and its skip behaviour. Alternatively make it a ratchet count key gated at 0.

**J5 §5: the exit and the stops leave a gap.**
- The exit needs the ratio key's after-median to be below its before-minimum, and at or below 5.
- The stops test the stage key's relative change, and the ratio being above 5.
- Not covered: the stage key improves and the ratio is ≤5, but the ratio's after-median is not below its before-minimum. Neither exit nor stop applies.

The "≤5" is also machine-relative. The committed baseline already reads 4.79, and this machine reads 5.5–5.7. My run suggests the gap will not bite (V1's worst round is below V0's best), but the executor should not need to rely on that. *Fix:* add a stop for "exit not met for any other reason".

**J6 §4: "never compared with the baseline" is inaccurate.**
- checkBaseline in cmd/ratchet/compare.go requires every printed key to appear exactly once in baseline.txt. Adding a key therefore also touches baseline.txt, the keys slice in main.go (which cites a ruling per addition) and main_test.go. Timing keys pass compare in either direction (TestWaypoint2KeysCompareRules), so "never gated" holds.
- libops has no stage key, so the stage table sent to perigrin will not sum to ParseWith.

Low severity. *Fix:* reword as "in baseline.txt, reported, never gated", and name libops or say why it is omitted.

**J7 §3.2: the guard is sound in shape but its test is promised, not shown.**
- The prototype panics in NewStateless. The decision routes the failure through NewPrecedence. ParseWith turns prec.err into a refusal of every file (forest.go:134), so a guard failure on perly.y would show loudly in rd-accepted-chart-refused. That is good.
- Neither prototype contains a toy grammar that makes generation fail. That is new work; it is feasible.
- The guard enumerates every LR(0) state that holds the key. It is therefore strict, and may refuse a grammar that Constrain handles. That narrows NewPrecedence's contract, and the decision should state it.

**J8 §3.4: "the port's refused cases" is not a fixture.** port_relational_unary_test.go defers the four chains to "the issue that completes the port". *Fix:* name the set or drop it.

**J9 §3.4: the aggregate exercises no refusal.** stateless-compare.txt says "oracle refuses 0". Refusal coverage therefore rests on the referee, the toys and waypoint2Refusals. The 185 and 989 precedence-refused sentences that §2 cites as evidence are not in §3.4's test list, so that evidence is one-shot. *Fix:* keep a seeded generator for those sentences, or say why not.

**J10, in support: architecture and direction.** The decision fits v2:
- rule 12 is untouched;
- v2 §8 says the value carried up the forest is the operator context at both edges, and that still holds;
- rule 11 (outside reference) is kept by §3.5.

The six parent-item pairs answer exactly the README's history: "keying on the rule and the token alone refused" T1/array.t, and the twostates and indirobsort toys. The ablations in stateless-compare.txt show that each key component can fail the test:
- mixed-forbid flips 50 files, T1/array.t among them;
- noshift flips 625;
- noreduce flips 1,099.

If the guard holds, every key has one decision. The oracle then only has to check that the evaluator supplies the right key, and the decision is honest that it does only that.

## Not examined

- The no-map and stateless prototypes (not run), and the toy and referee tests under the prototype.
- The three missing populations; only one timing process, not three.
- The recovery plan.
- The cost to cold start of building the tables (about 86 ms once per grammar, per stateless-compare).
- Whether "the ratchet key table" is the only live document affected beyond the README.

## Verdict: modify

The engineering claim holds. I reproduced exact agreement and the 0.73 independently. Before acceptance, fix J1, J2, J4 and J5, and trim J3; the rest is wording.
