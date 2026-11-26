# chain-review round 7, review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`: 19 issues, edges unchanged. I diffed every body against my round-6 snapshot (`f2r7-diff.txt`). Only d199, d218, d278, 3262 and the milestone Records changed. I ran my checks before I read `repair-r6.md`.

## 1. Recommendation

**Modify.** My FR4-1, which I declared blocking in round 6, was not answered. `repair-r6.md` does not list it, and the round-6 Records entry does not name it. d218's text on it is unchanged. Everything else I hold is residual.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 to FR2-8, FR3-1, FR5-1 | **released** (earlier rounds) | Their fixes are untouched this round. I re-checked FR5-1: 3262's `TestParseRefusesInfeasibleGrammar` still compares `Feasible`'s text and rejects "not the committed grammar". It now runs through `parseStream` inside `forest`, which the new `ParseWith` guard does not intercept. That is consistent. |
| FR3-2 | **held, residual** | 3262 grew again, with the three-way list and two more guard subtests. FR6-8 says the same about session size. |
| FR4-1 (d218) | **held, blocking** | **What is unchanged.** d218's "Tests this change moves" still reads "the production-versus-reference differentials, which this issue repoints at the recognizer's forest", and its step still says "repoint the existing production-versus-reference differentials". Neither names a test, says what happens on an input the judged chart refuses, or carries a criterion for them.<br>**What the repair did and did not cover.** This round's d218 edit settles only the toys: the candidate is the explicit form over each toy's own precedence, then Build. That covers `TestForestEqualityToys`. It does not reach the `Constrain`-against-`ConstrainReference` tests that 3262 pinned to the no-precedence constructor: forest and ratchet `precedence_reference_test.go`, `precedence_perlgit_test.go` and `TestWaypoint2RefusalsMatchReference`.<br>**Why it still blocks.** The last of these errors "recognition leaves no forest" on every row the judged chart refuses: its precedence rows, and its scope rows once the lattice is narrowed, which r6-fresh5 also measured. 3262's joint criterion runs it at milestone verify. Followed literally, d218 turns a closed issue's criterion red. Followed loosely, design §5's "'Production' now means the recognizer's forest" is unmet. d218's first step removes `Constrain` from `parseStream` and cannot surface either outcome, because none of these tests calls `Parse`.<br>**Fix, unchanged from round 6.** One paragraph in d218. Either name each test it repoints and give refused inputs d247's verdict-only rule, with one criterion that runs them. Or state that they stay `Constrain` referees on the no-precedence mode, reconcile that with design §5, and drop "repoint" from the step. |
| FR6-1 | **held, residual** | **What still happens.** Near-perly tables, with perly.y's count and LHSs but one right-hand side different, still fall into "judge nothing". The new `ParseWith` guard compares against what `New(tab)` would derive, and for those tables that is the empty judgement, so `Parse` accepts on them unjudged after d218. Today it refuses them with "not the committed grammar's tables".<br>**Others say the same.** r6-fresh3 FR6-1, r6-fresh4 FR6-2, r6-fresh6 FR6-7 and r6-pushback N15 raise this point. All are residual.<br>**Why it stays residual.** No production caller builds such tables, and `TestWriteTablesAndDrift` catches drift. |

## 3. Promotions

None. I read every residual row in r1-* to r6-*.

- **r6-fresh6 FR6-4 comes closest.** 3262 pins `forestdigest_test.go`, but its forest comes from `resolve_test.go`'s `built()`, which calls `recognize.New`. Meanwhile 3262 now lists `resolve_test.go` under "Judge from here". Pinning one means changing a helper the other shares. A builder can still pin the digest path alone, and 3262's first step reports a red golden before anything is edited, so it can be followed without a wrong outcome. Residual.
- **r6-fresh6 FR6-5:** d2a2 omits `main_test.go`'s `resetKeys`, so `TestResetKeySet` goes red at d2a2. The edit that fixes it is the natural one, removing the same key. Residual.
- **The rest:** each is true and residual on its own terms. That covers FR6-3, FR6-6 to FR6-9 and N15.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR7-1 | residual | 01a21447-d218 | Over "the leveled toys of `testdata/precedence/toys` that `prectab.Feasible` passes (13 toys, less `empty` and `emptyred`)", the candidate is the explicit form over the toy's own precedence. | **It assumes the other 11 toys pass `Feasible`.** That is asserted, not measured. The design names only `emptyred`, `nullableShift` and `empty` as failures, and the minute records counter 6 at 0 on "all 13 toys". The other counters' results on `nullpred`, `nullsuf`, `prefix` and the rest are not recorded anywhere I could find.<br>**If one fails,** the explicit form returns `Feasible`'s refusal for every input on it, and `TestForestEqualityToys` is red at d218. d218's first step would report it, so this is a stop, not a wrong build. | d218 Context. `toyNames` (`forest/precedence_test.go:23`). Design §4.1's examples. Minute rows F6-C2 and F6-C1. | Have d135's `TestFeasibleRefusesFailingToys`, or a sibling, assert `Feasible` passes on each of the 11, by name. The exception list is then measured once, before d218 relies on it. |

### Repair re-derived and found sound

- **The `ParseWith` guard.** A recognizer's precedence must equal what `New(tab)` derives. So an empty explicit precedence over the committed tables, or over a copy of them, panics. There is a subtest for each, and their criterion is red today. The rerun, the referees and `TestParseRefusesInfeasibleGrammar` reach the recognizer outside `ParseWith`, so none trips the guard. The speed harness and `Parse` use `New`, so they pass.
- **3262's three-way list.** The cost-bound rule ("report the number, change nothing without perigrin's ruling") answers N7, FR3-3 and AL17 without inventing a stop.
- **d278.** `TestBruteForceJudgedOverLeveledToy` requires `refused derivations: N` with N > 0 over a feasible leveled toy. That answers PB8 / FR6-2. It is a new test, so the criterion is red before the work.
- **Records and mapping.** The new Records entry matches the edits, and the mapping is unchanged.

## 5. What I did not re-examine

- I ran no criterion and did not measure `Feasible` on the 13 toys, which is FR7-1's question.
- I did not re-read the issues this round left unchanged: d02f to d166, d1c2, 3282, d1eb, d247, d2a2, d2d1 and bb9b.
- I did not check that each test in 3262's new "Judge from here, expected to hold" list really runs over perly.y's committed tables.
