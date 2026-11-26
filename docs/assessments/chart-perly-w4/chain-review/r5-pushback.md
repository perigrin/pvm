# chain-review round 5: pushback (plan quality) on chart-perly-w4

## Drafter: growth check

I diffed the live chain against my round-4 copy (`pb4/`) before reading `repair-r4.md`.

- **Issues changed:** 5 bodies — 3262, 3282, d166, d199 and d1eb.
- **Milestone:** one new Records entry, "Chain-review round 4 repair".
- **Unchanged:** every edge, title and label, and the issue count, which is still 19.

Each changed line maps to a `repair-r4.md` row, and each row answers a blocking finding:

| Edit | Answers |
|---|---|
| d166: the lex-unmapped rule's prose, one criterion reworded, the step, and `TestRefusalRerunScopeThenUnmappedIsLexUnmapped` | FR3-4, promoted by alignment |
| 3262: the precedence branch of the rule, and `TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped` | FR3-4 |
| 3262: the infeasible-grammar sentence, and `TestParseRefusesInfeasibleGrammar` | r4-fresh4 FR4-1 |
| d199: the infeasible-grammar refusal's prose, its TDD step, and `TestJudgedRecognizerRefusesInfeasibleGrammar` replacing `TestRecognizerDoesNotJudgeInfeasibleGrammar` | r4-fresh4 FR4-1 |
| d1eb: `TestExpectedOnInfeasibleGrammarIsNoPrecedence` replacing `TestExpectedUnjudgedForInfeasibleGrammar` | r4-fresh4 FR4-1. d1eb had to change: its old criterion asserted unjudged recognition of a grammar that d199 now refuses. |
| 3282: prose, step and criterion description limited to the toys `Feasible` passes | FR3-7, promoted by me |
| Milestone Records | all three |

**Verdict.** No edit answers nothing. Nothing grew beyond its finding.

- d1eb's change is a knock-on that the d199 change required, and it was made. The other knock-on, 3282's exception, also landed.
- `criteria-lint.sh` exits 0 over all 19 issues, so every criterion is still red at `e3432c0f`, the new ones included.

## 1. Recommendation

**Modify.** FR3-7 is released. Its fix landed, and it agrees with d199's new refusal and d218's exception.

I am promoting **r4-fresh3 FR4-1** to blocking (section 3). Round 4's FR4-1 repair turned an open question into a contradiction. d199 now has the judged constructor refuse a grammar through table generation, "for every input". But the chain never says where the judged constructor gets a non-committed grammar's precedence. Today the only source is `ParsePrecedence`, which errors on every tables value other than the committed grammar's. As written:

- the d199 infeasible-toy criterion cannot get `Feasible`'s refusal;
- 3262's "not expected to move" toy tests would all go red at the switch-over.

The fix is a stated interface. Nothing needs re-planning.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1–PB7, PB9–PB11, PB16, PB17 | **released** | None of these is touched by this round's edits. I re-checked the bodies they depend on (d1c2, d218, d247, d278, d2a2), and they are byte-identical to round 4. The new `ParseWith` text in 3262 is still consistent with PB17: the only caller of `ParseWith` outside `forest.go` uses `recognize.New`. |
| PB8, PB12–PB15, PB18–PB20 | residual | Unchanged. |
| N1, N5, N6 | **released** | Released in earlier rounds. N6's 3262 criterion is unchanged. |
| N2–N4, N7–N11 | residual | Unchanged. N7 is now subsumed by the promotion: see section 3. |
| FR2-8 (my promotion) | **released** | Released in round 4. d2d1 is unchanged. |
| **FR3-7** (my promotion) | **released** | **What the fix had to change:** 3282's prose, its step and its criterion; consistency with d218's exception; and a stated fate for the excepted toys. **Checked:** 3282 now says the judged-recognizer differential runs "over the toys `prectab.Feasible` passes". The five refused toys are named and "judged by `Constrain` over the no-precedence mode", which is d218's wording. This also agrees with d199's new rule: the judged constructor refuses those toys outright, so the excepted toys could not be compared judged anyway. The command did not need to change, because it greps for PASS. The test's toy list is the worker's, and the criterion now names it. The criterion is still red today. |

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| **r4-fresh3 FR4-1** | **blocking** | **The chain now asserts things that cannot all hold.**<br>**(a) There is no precedence source for a non-committed grammar.** `recognize.New(tab, Options{})` takes only a tables value. The only path from a tables value to its precedence is `forest.ParsePrecedence` / `buildParsePrecedence`, which d135 moves to `prectab`. It reads `grammar.Committed()` and returns "N table rules for perly.y's M: not the committed grammar's tables" for any other value (`forest/readings.go:116-133`).<br>**(b) d199's new criterion depends on that source.** `TestJudgedRecognizerRefusesInfeasibleGrammar` requires the judged constructor's table generation to return **`Feasible`'s refusal** on emptyred, nullableShift, MarkFirstFilter, empty and scancollide. Those are toy grammars with their own levels. Through the only existing source, the error would be "not the committed grammar's tables", not `Feasible`'s refusal. So the criterion is unreachable unless the judged constructor can be given a toy's levels, and no issue says how. The same gap sits under other tests: d199's "reduce-side test on perly.y and a toy", 3282's feasible-toy differential, and d278's levelled fuzz toy. Each needs a judged recognizer over a toy's own precedence.<br>**(c) 3262 contradicts it.** Round 4 added to d199 that a generation error makes `Recognize` return that error "for every input". From 3262, `New` is the judged constructor, and 3262 files "the recognizer-internal tests over toy grammars without levels" under "Not expected to move". `recognize/*_test.go` holds 99 `New(...)` call sites, most over `toyTables`. If judged `New` derives precedence the only way the tree knows, every one of those errors on every input. In that case 3262's first step halts on dozens of unlisted reds, and the worker has to make the interface decision alone. That is the situation the standing first step exists to prevent, not to absorb.<br>**Why it is blocking now, not residual.** In round 4 this was an open question about behaviour. Round 4's repair committed d199 to an error path, and that turned the question into two parts of the chain that contradict each other ((b) against the tree, (c) against 3262). It also adds a criterion, (b), that cannot pass as specified. The repairer flagged it, and predicted the batch of reds.<br>**What a fix must decide, in d199, with 3262 mirroring it:**<br>(1) how the judged constructor receives a grammar's precedence, for example a constructor that takes a `*prectab.Precedence` the caller builds, from the committed grammar or a toy's levels;<br>(2) what `New` does once it judges, on tables that are not the committed grammar's. It could judge nothing, or refuse, but the choice must be consistent with 3262's "not expected to move" and with `TestParseRefusesInfeasibleGrammar` / `TestPrecedenceGenerationErrorRefusesEveryFile`;<br>(3) which error `Parse` returns on a non-committed or infeasible grammar. V1's `Parse` refuses toy tables with the "not the committed grammar" error today.<br>I would stand aside on none of these alone, because (b) is a criterion that cannot pass. |
| N7 / r3-fresh3 FR3-3 (probe bound) | not promoted, folded | The same 3262 "not expected to move" list is now also the subject of the promotion above. The probe-bound headroom (0.6362 against 0.64) is still a cost question that the first step routes to the orchestrator with its text. It stays residual. |
| r4-fresh4 FR4-2 – FR4-4 and all earlier residual rows (AL5–AL16, FR2-5–FR2-7, FR3-2–FR3-5, PB/N residuals) | not promoted | None changed this round, and none makes proceeding wrong on its own. FR3-4 and FR3-5 are answered or have fates. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N12 | residual | 3262 | "`Parse` over a failing toy refuses everything (design §4.1)" with "the generation error", and `TestParseRefusesInfeasibleGrammar` requires that for each of the five toys | `Parse` takes tables only, so it reaches a toy's levels only through the same missing source as in FR4-1. Today it refuses those toys with "not the committed grammar's tables", which is already "refuses everything". So the criterion can go green with the wrong error unless the test checks the message. | `forest/forest.go` `parseStream`; `readings.go:121` | Resolve with FR4-1 point (3). Have the test assert the error text, not only the refusal. |
| N13 | residual | d166, 3262 | The lex-unmapped rule (FR3-4's fix) | I checked it for consistency: d166's "narrowed chart dies at the first unmapped token → lex-unmapped, no rerun" and 3262's rerun branch give V1's reason on both measured fixtures. The rule's only remaining gap is a file whose judged chart dies exactly at the unmapped token because of precedence. The unmapped token kills the un-narrowed run there too, so `lex-unmapped` is still V1's answer. No issue found. | d166 and 3262 prose; repair-r4's measured table | None. Recorded so the next reader need not repeat it. |

## 5. What I did not re-examine

- I ran no V2 code. FR4-1's consequence is reasoned from `readings.go` and the call-site count, not observed. That count is a grep for `New(` in `recognize/*_test.go`, 99 hits, not a count of tests.
- I did not re-measure the fixtures behind FR3-4 (lex-unmapped at 16 and 19, scope at 2). I took them from `repair-r4.md`.
- I did not re-read the 14 unchanged bodies beyond confirming they are byte-identical to round 4.
- I did not re-check the residual mapping. The repairer reports no mapping row moved, and the milestone diff confirms only a Records entry was added.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
