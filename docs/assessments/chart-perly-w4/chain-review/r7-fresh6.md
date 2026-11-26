# chain-review round 7: review-fresh6, resumed

I read the subject live from `/home/user/pvm/.claude/worktrees/v2-refine` at `e3432c0f`. It has 19 issues.

**Method.** I read the chain first and diffed every issue body and the milestone body against my round-6 dump. I read `repair-r6.md` after that. The diff matches the log, with nothing else:

- edits to the bodies of 3262, d218, d199 and d278;
- one new entry in the milestone Records;
- no change to any edge, label, title or mapping row.

The dependency graph is unchanged: d1c2 → 3282 → 3262 → {d1eb, d218} → … → d2d1 → bb9b.

## 1. Recommendation: accept

I hold no blocking finding. Both of my blocking findings landed, and I re-ran what proved them. Neither repair broke an edge, a carried id or a criterion's red-before state. The new residuals are below.

## 2. Status of my round-6 findings

| id | status | grounds |
|---|---|---|
| FR6-1 | **released** | **What the fix had to reach.** The guard; a criterion that exercises the hole; every caller of `ParseWith` that passes a recognizer; the infeasible-toy test that used to rely on `ParseWith` accepting the explicit form; and the failed-derivation path. **Each one, as checked:** 3262 now says the recognizer records the precedence it judges with, and `ParseWith` panics unless that is the precedence `New(tab)` would derive. For perly.y's rules, rule for rule, "an explicit precedence over those tables, such as an empty one built with `prectab.NewPrecedence(tab, nil, …)`, panics". The criterion `TestParseHasNoPrecedenceSwitch` gains two required subtests: empty explicit precedence over the committed tables, and over a copy of them. Both are red today, because the test does not exist yet. The callers are `forest.go:101` (`Parse` → `New`), `speed.go:147/286/380` (`New`) and `stages_test.go:274` (`New`), and all of them pass the derived form. `TestParseRefusesInfeasibleGrammar` moved to `parseStream` inside `forest`. In any case the explicit form stays legal over tables that are not perly.y's, which is safe because those have no derived levels to bypass. A failed derivation reaches `ParseWith` only through `Parse`'s own `New`, and that is "the one New(tab) would derive". The existing `TestPrecedenceGenerationErrorRefusesEveryFile` would turn red if the guard panicked there. |
| FR6-2 | **released** | d278 now names the gap: `forestToys` and `slipperToys` are level-less and judge nothing. It adds a TDD step over a §3-feasible leveled toy, and a criterion `TestBruteForceJudgedOverLeveledToy` that requires a PASS, no SKIP, `refused derivations: [1-9]…`, and the "judges nothing" sentence in `forest_test.go`. The criterion is red before the work. `forest_test.go` is in d278's paths. The ruling "Regenerate everything" is now met in substance. Two follow-on residuals are FR7-1 and FR7-2. |
| FR6-3 | held, residual | Unchanged. The corpus candidates in d199 and 3282 still don't say the lattice is narrowed. The round-6 repair said "explicit form … then Build" for d218's *toys*, but toys carry no library operators, so that does not touch this finding. |
| FR6-4 | held, residual (sharper) | 3262 now puts `resolve_test.go` under "Judge from here, expected to hold" and keeps `forestdigest_test.go` under "Pinned to the no-precedence constructor". `forestdigest_test.go` has no recognizer call of its own. It digests `built()` (`resolve_test.go:304`), which the list now says judges. Left judged, `TestForestDigestGolden` moves at 3262, because the Build forest loses the alternatives V1 dropped only at `Constrain`. The list then calls it pinned. The builder has to split `built` or give the digest its own builder, and the issue says neither. Step 1's dry run reports the red test, so I still class this residual. `perlForest` (`forest_test.go:516`) is still in no group. |
| FR6-5 | held, residual | Unchanged. d2a2 still does not name `main_test.go`, where `resetKeys` and `TestResetKeySet` will turn red when the key leaves `keys`. This is an instance of r4-fresh4 FR4-4: d2a2 has no dry-run step. |
| FR6-6 | held, residual | Unchanged. No criterion covers a long-lived recognizer across files with narrowed set ids. |
| FR6-7 | held, residual | Unchanged. r6-fresh2 FR6-1, r6-fresh3 FR6-1, r6-fresh4 FR6-2 and r6-pushback N15 raise the same subject. I do not promote it (see section 3). |
| FR6-8 | held, residual | 3262 grew again: it gains a cost-bound group and a reporting rule. It is cohesive, but it is the largest issue in the chain. |
| FR6-9 | held, residual | d247 still reads "the entry-counting hook the switch issue added". That hook is d218's. |

## 3. Promotions

None. I read every residual row in r1-* through r6-*: alignment, pushback, and fresh2 to fresh6.

The rows on the derived form judging nothing over tables that are not perly.y's are r6-fresh2 FR6-1, r6-fresh3 FR6-1, r6-fresh4 FR6-2, r6-pushback N15 and my FR6-7. I keep these residual.

- Production and every anti-goal input use the committed tables, whose generation `tables-drift` holds to perly.y.
- A tables value that differs from perly.y by one rule is another grammar. `grammar.Tables` carries no levels for it, so no precedence of perly.y's is being bypassed.
- The round-6 `ParseWith` guard is consistent with this: it allows the explicit form only over tables that are not perly.y's.

What is left is undocumented behaviour, for the docs issue to state, not an off-switch.

Other rows I re-weighed and leave residual:

- **PB8.** It is now answered by d278's brute-force criterion.
- **r4-fresh4 FR4-4.** It reappears as FR6-5, and each instance is caught at close by `make test`.
- **FR3-3 / AL17.** These are answered by 3262's named cost bounds and the rule to report the number to perigrin.
- **r5-fresh5 FR5-3 and FR5-4.** The current text resolves both:
  - `TestWaypoint2RefusalsMatchReference` stays pinned as a `Constrain`-versus-reference check;
  - the toy candidate check excepts the toys `Feasible` refuses.

## 4. New findings

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR7-1 | residual | 01a21447-d278 | The brute force "gains a §3-feasible toy with levels, reusing the fuzz oracle's toy". | The fuzz oracle's toy is written in `recognize/fuzz_test.go`, a test file of package `recognize`. The brute force and `ConstrainReference` are in package `forest`'s tests. A `_test.go` symbol of another package cannot be imported. So the issue relies on an interface that does not exist, and the builder will duplicate the toy or move it. If the two copies drift apart, "reuse" means nothing. | `recognize/fuzz_test.go:171` and the fuzz seeds; `forest/forest_test.go:84-124` (`brute`); d278's Context. | Name where the shared toy lives, for example a grammar file under `testdata/precedence/toys`, or a non-test constructor in `prectab`. Both tests then read the one definition. |
| FR7-2 | residual | 01a21447-d278 | The brute-force criterion compares "the derivations `ConstrainReference` keeps over the no-precedence forest" with the judged recognizer's Build forest. | V2 is held equal to V1's `Constrain`, not to `ConstrainReference`. The two are pinned to diverge (`TestPrecedenceReferenceKnownDivergence`, the "shifts defect", out of scope per design §2). On a toy that reaches that divergence, the criterion would be red and the builder would read a §7 stop. The pinned divergence needs an empty rule with a non-allow verdict, which counter 5 makes infeasible, so a feasible toy probably avoids it. "Probably" is the whole guarantee. | `forest/precedence_pinned_test.go:27-46`; design §2 non-goals; d278's new criterion. | Judge the brute force's derivations with `Constrain`, which is V1, the thing V2 must equal. Or keep `ConstrainReference` and state that a feasible toy cannot reach the pinned divergence, with counter 5 as the reason. |

## 5. What I did not re-examine

- I did not run any criterion. I re-read the five new or changed criteria and the guard text, and checked them against the call sites in the tree.
- I did not re-check d0e1, d10d, d135, d166, d1c2, 3282, d1eb, d247, d2a2, d2d1 or bb9b beyond confirming that their bodies are byte-identical to round 6.
- I did not re-check the numeric fixtures, or the 66-row residual mapping, which is unchanged.
- I did not check whether the cost bounds in 3262 (0.64, 0.040) hold once the default judges. That is now a stated report-to-perigrin path, not a claim.
