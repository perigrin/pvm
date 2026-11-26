# chart-perly-w4 chain-review, round 7: review-fresh5 (resumed)

I read the live chain first and diffed it against my round-6 dump (`fr6x/` against `fr7x/`), ran my checks, and only then read `repair-r6.md`.

The repair edited 3262, d218, d199 and d278, and appended one entry to Records. Nothing else moved:
- no dependency edge, title or label changed;
- no residual-mapping row changed (the Records diff is append-only);
- the criterion count went from 197 to 198.

Every cheap criterion is still red today. The working tree is clean.

## 1. Recommendation

**Modify.** Both of my own blockers, FR5-1 and FR6-1, are settled and released.

One blocking finding is still unanswered: my FR5-3, which r6-fresh2 merged into its FR4-1 and now holds as blocking. I concur, and I hold FR5-3 as blocking beside it. The round-6 repair did not touch it, and d218's "repoint the existing production-versus-reference differentials" sentence is unchanged.

## 2. Status of my findings

I am biased toward finding my own bars met. I scored each fix against what I wrote, not against what landed near it.

### FR5-1 (`ParseWith` off-switch): released

**What I predicted the fix would also have to change.** If `ParseWith` refuses explicit precedence, two things break unless they change too:
- 3262's `TestParseRefusesInfeasibleGrammar`, which passed `ParseWith` an explicit toy precedence;
- every `ParseWith` caller that holds a non-`New` recognizer.

**What I checked.**
- **Callers.** I grepped every `ParseWith` caller. There are three: `forest.Parse` (through `New`), `stages_test.go:275` (through `New`) and `speed.go`'s `pipelineParse` (through `New`). None is broken.
- **The infeasible-toy test.** It now runs through `parseStream` inside `forest`, not through `ParseWith`.
- **The guard's rule.** For tables whose rules are perly.y's, a copy included, the recognizer must carry the derived precedence. An explicit precedence over toy tables stays allowed, and there is nothing there to bypass.
- **A recognizer over the wrong tables.** A recognizer built over toy tables and passed with perly.y's `tab` carries a precedence that is not `tab`'s derived one, so the guard catches it.
- **The criterion.** New subtests refuse an empty explicit precedence over both the committed tables and a copy of them.

My bar was "`ParseWith` checks that the recognizer's precedence is `tab`'s own". That bar is met.

### FR6-1 (3262's false claim about the ~113 in-package calls): released

**What the fix had to do.** It had to split 3262's list, name the bounds with almost no headroom, and stop routing a cost move to "a forest differs". It does all three:
- the list is now three classes: level-less toys, perly.y-backed tests that judge from here, and cost bounds;
- `TestCompletionProbesBounded` (0.6362 against a bound of 0.64) and `TestLeoMemoEntriesBounded` (0.0395 against 0.040) are named with their numbers;
- a red cost bound is reported to perigrin, and the bound is not changed without his ruling.

**Coverage check.** I listed every test in package `recognize` that reads perly.y's tables directly. There are 18. Every one is in a named class, or in a file pinned above (`agree_test.go`, `setdigest_test.go`). All 15 names the list adds exist in the tree.

### FR5-2: held, residual

3262's rerun prose is unchanged. It still does not restate d166's no-rerun shortcut for `TestRefusalLexUnmappedNotRerun`. It can still be built to keep that shortcut, so it stays residual.

### FR5-3: held, blocking

I hold this beside r6-fresh2 FR4-1, which absorbed it. d218 still repoints "the existing production-versus-reference differentials" without naming them. `TestWaypoint2RefusalsMatchReference` errors "recognition leaves no forest" on rows the judged chart refuses. 3262 pins that test and runs it in a criterion that verify re-runs at close.

Read literally, d218 turns 3262's criterion red after 3262 has closed. Read loosely, §5's "production now means the recognizer's forest" goes unchecked. I accept fresh2's promotion and its grounds.

**Fix.** Have d218 name each existing reference test and say which it repoints. For a repointed test, say what happens to inputs the judged chart refuses (verdict only, as d247 does for its referee test). Add a criterion.

### FR5-4: released

d218's never-constrains criterion now reads "toys, through the explicit form over each toy's own precedence". A toy that `Feasible` refuses cannot go through the explicit form without error, so the check covers only the feasible toys. The infeasible toys stay on `Constrain`, as d218 says. The contradiction is gone.

### FR5-5

Released in round 6. It is unchanged.

## 3. Promotions

I read every residual row in r1 to r6 across all participants, including r6's FR6-1 to FR6-9 (fresh3, fresh4 and fresh6) and pushback's N15.

I promote none of them. Fresh2's FR4-1 is already blocking, and I hold it jointly through FR5-3.

The nearest call is r6-fresh3 FR6-1 and N15: a leveled grammar that is not perly.y's would judge nothing. Production only ever builds `grammar.Generate(grammar.Committed())`. Toys with levels must use the explicit form. And `ParseWith` now rejects explicit precedence over perly.y's rules. So no production or test path silently loses precedence. It stays residual.

## 4. New findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR7-1 | residual | 3262 (with d218, d199) | `ParseWith` "panics unless that precedence is the one `New(tab)` would derive". | Over perly.y's tables, `New(tab)` can derive an *error*. `TestPrecedenceGenerationErrorRefusesEveryFile` forces exactly that through `grammar.Derived`, then calls `Parse`, which calls `ParseWith`. A guard written as "the recognizer holds a non-nil precedence equal to the derived one" would panic there. That breaks d218's stated fate ("`Parse` still refuses every file with the generation error"). The guard can be built correctly, so this is residual. | 3262 guard paragraph. `forest/precedence_test.go:370-393`. `forest.go:101`. | Say that the guard accepts `New(tab)`'s recognizer when its derivation failed, and lets `Recognize`'s error become the refusal. Add a subtest to `TestParseHasNoPrecedenceSwitch`. |
| FR7-2 | residual | d218 (with d199) | The feasible leveled toys are "13 toys, less `empty` and `emptyred`". | Nobody has measured `prectab.Feasible` on the other 11 toys: `follow`, `indirob`, `indirobsort`, `ladder`, `lalr`, `mixed`, `nonassoc`, `nullpred`, `nullsuf`, `prefix` and `twostates`. If one of them fails (`nullpred` and `nullsuf` are the likely ones under counter 5), the general exception "every toy Feasible refuses" still covers it. But the parenthetical count would be false, and a builder who follows it would read the resulting error as a §7 stop. | The repair's d218 text. `feasibility.md` names no toy in `toyNames` other than `empty` and `emptyred`. `forest/precedence_test.go:24`. | Drop the count, or have d135's `TestFeasibleRefusesFailingToys` log a Feasible verdict for every toy in `toyNames`. |

## 5. Not re-examined

- I ran no `go test` criteria in this round.
- I did not re-trace the issues the repair left byte-identical: d02f to d135, d166, d1c2, 3282, d1eb, d247, d2a2, d2d1 and bb9b.
- I did not measure `Feasible` on the 11 toys (FR7-2).
- I did not re-check the 66-row mapping row by row. The Records diff is append-only.
