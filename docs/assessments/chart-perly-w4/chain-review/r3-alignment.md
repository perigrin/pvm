# chain-review chart-perly-w4, round 3: review-alignment (resumed)

## 1. Recommendation

**Modify.** One blocking finding is held: I promote PB17 (section 3). Everything I raised in rounds 1 and 2 as blocking is released, including AL12, AL13 and my PB9 promotion.

The PB17 fix is small:
- make `ParseWith` refuse a no-precedence recognizer;
- add that case to `TestParseHasNoPrecedenceSwitch`.

If that lands, or if the raiser stands aside with the decision's §4.4 in view, I expect to recommend accept.

**How this round was done.** I read the live chain first: 19 issues at `e3432c0f`, edges taken from `issue show`'s `blocked_by`. I ran the checks before reading `repair-r2.md`.

**Edges.**
- Stage 1 runs in order: d02f → d05d → d089 → d0ba.
- Stage 2 then fans out: d0ba → {d0e1, d10d} → d135.
- Stage 3 continues: d135 → d166 → d199 → d1c2 → 3282 → 3262 → {d1eb, d218}.
- After d218: d218 → {d247, d278, d2a2}, with d1eb → d278.
- The chain closes: {d1eb, d247, d278, d2a2} → d2d1 → bb9b.
- This is the order the milestone states. The stage sequence from AL3 is intact.

**Criteria.** I re-ran all 190 criteria against `e3432c0f`, each passed to `sh -c` unmodified. All 190 are red, so none is green before its work.

**Residual mapping.** The 66 rows plus the crumb are all present. Carried ids survived the reorder:
- d199 carries A16, A29, F3-6, C13 (reduce half) and F6-R2.
- 3262 carries F5-3, C12 (precedence half), F3-7 (switch half) and C10 (a).
- d1c2 and 3282 are as in round 2.

The only mismatches are the round-1 residual AL6 rows (A19 ×4, C20).

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1 | released (held from r2) | d166's rule is unchanged from round 2. 3262 now states the same rule, and adds "No interim state reports a precedence-only death as `scope`". That claim is true now, because `recognize.New` stays unjudged until 3262. |
| AL2 | released (r2) | No change bears on it. |
| AL3 | released (r2) | Edges re-derived above. |
| AL4 | released (r2) | 190 of 190 criteria are red. |
| AL5–AL11 | residual, held | Unchanged and outside the repair's scope. I re-ran the AL6 cross-check: same 5 misses. |
| AL12 | **released** | **What the fix had to change as well:** the criterion's literal, the prose's count, and any other "1,299".<br>**The criterion:** d0e1's is now `compared 1296 of 1296`. Its prose says "the 1,296 kept sentences (the rows of sentences.txt with a rendering…)". I re-measured: 1,296 rows have a rendering, and `TestPrecedenceReferee` logs "1296 kept sentences".<br>**Left over:** "1,296 of 1,299" survives in d0e1's and d247's F8-P1 quote. It no longer reaches any criterion, so I folded it into AL14. |
| AL13 | **released** | **What the fix had to change as well:** an owner for both tests; their new expected answers; the interim `scope` mislabel; and the edge order, so the switch happens after the shift side.<br>**Owners and answers:** 3262 is now the single switch-over, after 3282. It names `TestParseAppliesPrecedence` (expects `precedence` away from byte 0) and `TestRatchetVerdictIsPipeline` (expects `prec.pl` not recognized). It also names `TestPipelineStageKeysRefusedFiles`, which neither of us had found. All three are asserted in one criterion.<br>**Interim mislabel:** gone, because the default stays unjudged until 3262.<br>**Unlisted breakage:** 3262 has an inventory step. It runs `make test` with `New` judged, before any edit, and sends any failure not on its list to perigrin.<br>**The fixtures' premise:** I measured it at `e3432c0f` (scratch copy, `forest/zz_al3_test.go`). `print $fh [];`, `$a <=> $b <=> $c;` and `1 .. 2 .. 3;` are each `recognized=true`, and `Parse` gives `{Byte:0 Reason:precedence}`. This is the premise of d0ba's new V1-side test and 3262's default-judges test, and it holds.<br>**Bias check:** I wrote the bar. Against that bias, I re-read 3262 whole, not just the named tests. |
| AL14 | residual, held | d247 still asserts `carrying zero-width ; leaf: 1296` without defining "carry". My round-2 measurements: 0 reachable, 1,277 present in `f.Nodes` on the V1 path, and 1,296 slipper insertions per the minute. d0e1's and d247's prose still quote "1,296 of 1,299". |
| PB9 (my r2 promotion) | **released** | d166 now says `terminals-per-token` is summed over the un-narrowed `lex.Lex` tokens. `TestTerminalsPerTokenCountsUnnarrowedTokens` has a dead-`PLUGIN_*` subtest that logs the narrowed sum beside the un-narrowed one, plus a corpus subtest. It is written as a guard before the caller changes. The criterion is red today because the test does not exist. A step also runs the resolution command where perl.git is present. |

**The repairer's two flagged calls.**

- **3262's list of tests that move comes from reading the code.** I accept that it comes with an inventory step, and that the step's failure path goes to perigrin rather than to an edit. That is the right disposition for a list nobody has run.
- **`TestJudgedRecognizerIsExplicit` assumes the reduce side refuses something on the set-digest sample.** I did not build a judged recognizer, so this is unmeasured. But the sample is corpus files (`setdigest_test.go`, "over the corpus sample"). Any `a + b * c` gives a completion of `term ADDOP term` before `*`, which V1's reduce verdict forbids before every following terminal of that set. The design measures reduce-side pruning at 5.6% of packed alternatives over the corpus. I judge it near-certain to pass and leave it unpromoted. Recording the count in the test's log would make it checkable.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| PB17 (pushback, held residual in r1 and r2) | residual | **blocking** (I hold it) | **The repair made it concrete.** Round 2 chose "the two modes are chosen by constructor" (d199). That gives `recognize` an exported no-precedence constructor, which the rerun in `forest` and the referee in `chart` need.<br>**Why that is an off-switch:** `ParseWith(tab, r *recognize.Recognizer, src, scope)` accepts any recognizer. After d218, `ParseWith` no longer calls `Constrain`. So `ParseWith(tab, <no-precedence recognizer>, …)` returns an unjudged forest.<br>**What the decision says:** design §4.4: "No caller of `Parse` can turn precedence off." §6: the mode is selected "so that no caller of `Parse` can turn precedence off". Ruling "Diagnose by rerun": the mode "classifies refusals and is not a way to parse".<br>**Why nothing catches it:** `TestParseHasNoPrecedenceSwitch` (3262) reflects over `Parse`/`ParseWith`'s "exported parameters and option fields". A `*Recognizer` value built by another constructor passes it, so the criterion is green while the guarantee is false.<br>**Which reason applies:** a claim the subject asserts is false, and its criterion cannot detect it. This is the reason I give for promoting it.<br>**Fix (3262):** the recognizer records its mode, and `ParseWith` refuses (or panics on) a no-precedence one. Extend `TestParseHasNoPrecedenceSwitch` to call `ParseWith` with the no-precedence constructor and require that. Or keep the no-precedence constructor out of reach of `ParseWith`'s callers, and say how. |

**Rows read and not promoted.** I read every residual row from r1 and r2, for both lenses and fresh2:
- **PB8.** It is now sharper, since the brute force is pinned no-precedence until d278, but it is still an absent criterion rather than a wrong one.
- **PB12–PB15, PB18–PB20, N2, N4.**
- **N3.** My round-2 measurement (0 reachable zero-width leaves) makes the dedup case moot.
- **FR2-6.** d199 is now 14 criteria without the pins.
- **FR2-7.** Answered: d05d adds `scancollideToy`, and d135 keeps the tests in `forest`.
- **FR2-8.** d2d1's `.meta` files are still unnamed in Steps. That is a bookkeeping trap, not a wrong claim, so it stays residual.
- **AL5–AL11, AL14.**

## 4. New findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| AL15 | residual | 01a21447-d199 | `TestJudgedRecognizerIsExplicit`: "the judged constructor's [sets] differ from [the default's] in at least one set" on the perly.y sample. | The test is unmeasured (repairer's flag). If the sample had no reduce-side refusal, the criterion could never pass. My reading says it will pass (section 2), but the test logs no count, so a pass cannot be told from a lucky one. | d199 criterion. The sample is corpus files (`setdigest_test.go` ABOUTME). | Log the number of differing sets and refused completions, and require a non-zero number of each. |
| AL16 | residual | milestone Records | The "Chain-review round 1 repair" entry still says d199 "keeps … the guards pinned to it" and that 3262 "classifies precedence refusals by rerun". | Both statements are true of round 1 and false now: the pins and the switch moved to 3262. The round-2 entry beneath corrects them, so a reader who reads both is not misled. | Records lines 91–95 against 96–100. | Leave as is: Records are dated and append-only. Optionally add "(superseded by the round-2 entry)". |

## 5. What I did not re-examine

- **3262's size.** It holds the switch-over, the rerun, both readers, three rewrites and two retirements. Sizing is pushback's. I checked only that it carries the ids and that its criteria are red.
- **The tests named in 3262's "Pinned" and "Not expected to move" buckets.** I did not check them against the code one by one. The inventory step covers that.
- **Whether a `none` context prints as `{State Dist}`** (PB1's statement in d1c2).
- **Unchanged bodies.** d089, d1eb, d218, d278, d2a2, d2d1 and bb9b are byte-identical to round 2, and I did not re-read them beyond the criterion re-run.
- **perl.git targets.** Their criteria are gated red today.
