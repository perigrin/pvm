# chain-review chart-perly-w4, round 2: review-alignment (resumed)

## 1. Recommendation

**Modify.**

All four of my round-1 blocking findings are answered: three are released and one is released on new evidence. But this round raises two new blocking findings, and I promote one of pushback's residual rows:

- **AL12:** a criterion fixes a count of 1,299 referee sentences. There are 1,296.
- **AL13:** two existing tests on the `$a <=> $b <=> $c;` fixture break when the recognizer starts judging, and no issue owns them.
- **PB9, promoted:** narrowing in the ratchet's caller can move `terminals-per-token`, an answer key.

I read the chain live from the worktree at `e3432c0f`, 19 issues, before reading `repair-r1.md`.

## 2. Status of my round-1 findings

| id | status | grounds |
|---|---|---|
| AL1 | **released** | **What the fix had to change as well:** the rule in d166, a two-fault criterion, d199's or 3262's wording of the same rule, and `refused-where.txt` handling.<br>**What I found:** d166 now says the un-narrowed rerun "dies at all, wherever it dies" → `grammar`, and "only if the rerun accepts" → `scope`. It adds `TestRefusalRerunLaterGrammarFaultIsGrammar`, which is red today. 3262 states the same rule, so the d166/d199 contradiction is gone.<br>**`refused-where.txt` claim:** "only `table` and `grammar` rows" is true. I measured 282 `grammar` rows and 54 `table` rows.<br>**Bias check:** I am inclined to find my own bar met. Against that, I re-read d166's whole paragraph, not just the changed sentence. |
| AL2 | **released**, on a correction to my own premise | **Round-1 claim:** referee forests differ on about 1,296 reachable nodes.<br>**What the measurement shows:** I tested this in a scratch copy of `e3432c0f` (`al2-scratch/internal/chart/zz_al2_test.go`, `zz_al2b_test.go`; nothing in the repository was touched). Over the 1,296 kept sentences, a zero-width `;` leaf exists in `f.Nodes` for 1,277 forests. It is reachable from the root in **0** forests, both in the raw Build forest and after narrowing with zero-width leaves kept. The walk does reach zero-span nonterminal nodes (`stmtseq` 1,297, `empty` 1,885), so it can see zero-width nodes. Keeping the leaves alive changes 0 verdicts, 0 reachable nodes and 0 alternatives.<br>**Why the V2 forest has none either:** narrowing and precedence only remove derivations, so the V2 forest cannot reach a leaf the un-narrowed forest does not reach. The reachable-node comparison would therefore have been green without the repair. This matches F7-C2's stand-aside.<br>**The repair:** the exclusion is harmless, narrow, and pinned by three subtests. It does not need perigrin. My round-1 consequence was wrong, and I withdraw it. |
| AL3 | **released** | **Edges, checked against `blocked_by` from `issue show`:** d02f ← nothing; d05d ← d02f; d089 ← d05d; d0ba ← d089; d0e1, d10d ← d0ba; d135 ← all five; then a straight line d166 → d199 → 3262 → d1c2 → 3282 → {d1eb, d218}.<br>**Rationale text:** the "disjoint files" sentence is replaced with the real reason.<br>**What else had to change:** the shared `feasibility_test.go` is now created by d02f and extended serially. The Makefile line is edited by d089, then d10d, also serially.<br>**Tool note:** the `blocks` field in `issue list --format json` reads in the opposite direction from round 1's output. `issue show`'s `blocked_by` is the authority I used. |
| AL4 | **released** | I re-ran all 185 criteria against `e3432c0f`, passing each command to `sh -c` unmodified. All 185 are red; none passes before its work. d199's criterion is now gated on `TestPrecedenceIgnoredByRecognition` being gone, and bb9b's on the verdict citation. |
| AL5 | residual, held | Unchanged. 8 rows are still labelled "not carried" while bb9b's `Carries:` line names them. |
| AL6 | residual, held | The mapping cross-check, re-run, still shows "A19 → d166, d199, d1c2, d218" and "C20 → d247" missing from those issues' `Carries:` lines.<br>**Nothing lost in the splits:** every id that d199 and d1c2 carried before the split is now carried by d199, 3262, d1c2 or 3282, and the mapping points at the right half. The F8-P1 row was correctly extended to d0e1. |
| AL7–AL11 | residual, held | Unchanged. None was in the repair's scope. |

**On the repairer's flagged call: keeping the two retired tests in d199.** I agree with it. The repairer's own grounds apply: the fixture stops recognizing when the default recognizer judges. But those same grounds expose AL13 below.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| PB9 (pushback) | residual | **blocking** (I now hold it) | **The instruction:** d166 tells the ratchet's direct `Recognize` callers that hold a scope to pass narrowed lattices. It names only the chart-shape keys (F2-7, F4-5).<br>**What else sits in that caller:** the one caller with a scope, `chartInputs` in `cmd/ratchet/verdicts.go:271-277`, also sums `terminals-per-token` from each token's terminal set. That key is one of the 27 answer keys, and it sits in the same loop.<br>**Why it moves:** narrowing exists to drop `PLUGIN_*` readings, and §3 counts 1,110 corpus tokens carrying them. If the count is taken from narrowed sets, the key moves, which is a §7 stop.<br>**Why nothing catches it in time:** no criterion before d2d1 or the milestone resolution compares answer keys.<br>**Classification:** the step does the wrong thing when followed literally, and it fails late. That makes it a defect, not a residual.<br>**Fix:** d166 states that `terminals-per-token` is counted from the un-narrowed token stream. Add a criterion that the ratchet's `terminals-per-token` on a fixture with a dead `PLUGIN_*` reading equals `e3432c0f`'s, or a unit test on `chartInputs`. |

I read every other residual row: PB8, PB10 to PB20, and AL5 to AL11. PB10 and PB11 are answered by the AL3 edges. I promote none of the rest.

## 4. New findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| AL12 | blocking | 01a21447-d0e1 | Negative criterion: "compares every referee sentence it loads, 1,299 as F8-P1 measured", checked as `compared 1299 of 1299`. | The referee has 1,296 kept sentences, so a differential that loads every kept sentence logs `compared 1296 of 1296`, and the criterion cannot pass. Fixing it means changing the criterion, which is a §7 stop. This was present in round 1 and I missed it. | `sentences.txt` has 1,341 rows, 1,296 of them with a rendering. `TestPrecedenceReferee` logs "1296 kept sentences agree with bison". Every other measurement in the minute says 1,296: F4-1, the round-5 lines 556–565, and F7-C2. Only F8-P1's prose says "1,299". | Change the count to 1296, or assert on `compared N of N` with N equal to the sentences that have a rendering. |
| AL13 | blocking | 01a21447-d199 (and 01a21463-3262) | d199: "Two existing tests stop holding the moment recognition judges", naming `TestPrecedenceIgnoredByRecognition` and `TestRecognitionKeysUnchangedByPipeline`. d199 also says a precedence-only death is classified by d166's rule because "no corpus file or port snippet has such a death". | **Two more tests on the same fixture.**<br>`TestRatchetVerdictIsPipeline` (`cmd/ratchet/verdicts_test.go`) requires `prec.pl` = `$a <=> $b <=> $c;` to be recognized by `recognize.New(tab, Options{})`. That breaks at the same moment.<br>`TestParseAppliesPrecedence` (`forest/readings_test.go:191`) requires `Refusal{Reason: "precedence"}` at Byte 0. Under the interim rule it reads `scope`, because the no-precedence rerun accepts. After 3262 it reads `precedence` at a non-zero byte, which 3262's own `precedence-byte-not-zero` requires.<br>**Who owns them:** no issue names either test, so d199 or 3282 cannot close `make test` green without unlisted edits. The interim `scope` label is a reason V1 never gives for that input. | `grep` over all 19 bodies and the milestone finds 0 mentions of either test. `TestRecognitionKeysUnchangedByPipeline` and `TestRatchetVerdictIsPipeline` share the `prec.pl` fixture. `TestParseAppliesPrecedence` is `Parse(…, "$a <=> $b <=> $c;", …) != Refusal{Reason: "precedence"}`. | Name both tests in the issue whose change breaks them. Measure which issue that is: either d199 (reduce side) or 3282 (shift side). The left-nested reading dies on the reduce verdict and the right-nested one on the shift verdict, so it may be 3282. Say what each test becomes. Have 3262 own `TestParseAppliesPrecedence`'s byte. |
| AL14 | residual | 01a21447-d247 | The slipper pin asserts `carrying zero-width ; leaf: 1296`. | "Carry" is undefined. My measurement gives 0 for reachable leaves. For leaves present in `f.Nodes` it gives 1,277 on the V1 Build path. assess-codebase counted 1,296 ";-insertions" in the narrowed recognizer. A builder who counts the natural thing may get a number other than 1,296 and hit a stop on a correct change. | `zz_al2_test.go`: "forests with a zero-width ; node: 1277 … reachable-zero 0". Minute line 563: "sentences where the slipper inserted a ; in the narrowed run=1296". | Define the count as slipper insertions in the chart. Or drop the literal and assert that every kept sentence is compared and 0 are ambiguous or cut. |

## 5. What I did not re-examine

- **Pushback's own blocking fixes, PB1 to PB7, on their merits.** For PB1, that means whether the `none` context really prints as `{State Dist}`. For PB6 and PB7, the sizing. I checked only that the splits lost no carried id, kept the dependency order, and left every criterion red.
- **Which issue the `$a <=> $b <=> $c;` fixture stops recognizing at** (AL13). I did not measure it. The finding holds either way.
- **The remaining residual mapping rows in depth.** I re-ran only the id-set and `Carries:` cross-checks.
- **perl.git targets** (`make chart-precedence-reference`, `make chart-ratchet`). Their criteria are gated red today.
- **Unchanged issue bodies.** d02f, d05d, d089, d0ba, d10d, d135, d1eb and d2d1 are byte-identical to round 1. I did not re-read them on their own terms beyond the criterion re-run. That is where AL12 hid last round.
