<!-- ABOUTME: Minute of the crochet:assess session on chart precedence V2, judging precedence in the recognizer. -->
<!-- ABOUTME: Covers rounds 1 to 8 and revisions 2 to 9, ending in a fixed point at round 8: the ledger, cross-references, growth checks, per-participant records, rulings, outcomes and the residual list for refinement. -->

## Assessment: Chart precedence V2: judge precedence in the recognizer

Recommendation: accept: fixed point at round 8 (all ten participants accept; no blocking finding held or raised), on revision 9 (bc9b4abd, 6d4dacab)

Before the fixed point: perigrin ruled accept during round 6, on revision 7, with seven of eight accepting and assess-fresh6 outstanding. assess-fresh6's report then arrived with a modify and three blocking findings, and perigrin ruled "Revise to rev 8, run round 7 (Recommended)", holding the acceptance. Round 7 released those three and assess-codebase held one new blocking finding, C22; revision 9 answered it and round 8 followed. Both rulings are recorded below, under "Round 6" and "Round 6, continued".

By round: round 1, three modify; round 2, four modify; round 3, accept from assess-fresh3 and modify from the other four; round 4, modify from assess-fresh4 and accept from the other five; round 5, modify from assess-fresh5 and accept from the other six; round 6, accept from all seven reporting participants, with assess-fresh6's report outstanding when perigrin ruled, and modify from assess-fresh6 when it arrived; round 7, modify from assess-codebase and accept from the other eight; round 8, accept from all ten.

- **Subject:** `docs/plans/2026-11-25-chart-precedence-v2-design.md`. Beside it, in `docs/assessments/chart-precedence-v2/`: `rulings.md`, `context-split.md`, `feasibility.md` and `feasibility-code/`. PR perigrin/pvm#646.
- **Revisions:**

  | revision | commit | blob | path | title |
  |---|---|---|---|---|
  | 2 | 488e1c7e | 487d7c51 | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 3 | a8808dd2 | 71d5f7bb | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 4 | 1fab4a89 | 68ee2dd6 | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 5 | de604896 | 3a89343d | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 6 | bcd26304 | a9c54728 | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 7 | 28da56fa | 5cb3ab81 | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 8 | 2d9611db | d893c1f0 | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |
  | 9 | bc9b4abd | 6d4dacab | `docs/plans/2026-11-25-chart-precedence-v2-design.md` | Chart precedence V2: judge precedence in the recognizer |

  Revision 1 was not assessed. The subject's status line calls itself revision 2.
- **Revision assessed in round 1:** commit 488e1c7e, blob 487d7c51b7503e9db4fe621b330f2281ad391782. Each participant names that blob in its report and says it checked it with `git hash-object`.
- **Revision assessed in round 2:** commit a8808dd2, blob 71d5f7bb4a1f535d49b68258f6f1d715f95197d7 (revision 3). Each of the four participants names that blob and says it checked it with `git hash-object`. The commit changes three docs files and no code. The minute was at a8808dd2 (round 1 only).
- **Revision assessed in round 3:** commit 1fab4a89, blob 68ee2dd6e860876138d6e88cb99ca2983c7a3dd5 (revision 4). Each of the five participants names that blob and says it checked it with `git hash-object`. The commit changes three docs files and no code.
- **Revision assessed in round 4:** commit de604896, blob 3a89343d52c3b1bb4533af51b92ac77743587749 (revision 5). Each of the six participants names that blob and says it checked it with `git hash-object`. The commit changes the design and `rulings.md` and no code. The worktree's HEAD during the round was 0df159fb, the minute's commit through round 3; the design there is the same blob.
- **Revision assessed in round 5:** commit bcd26304, blob a9c54728581ceddd1823d1ccaab9519d18c6d2ab (revision 6). Each of the seven participants names that blob and says it checked it with `git hash-object`. The commit changes the design only: the status line and one paragraph of §4.5. The worktree's HEAD during the round was 824de021, the minute's commit through round 4; the design there is the same blob.
- **Revision assessed in round 6:** commit 28da56fa, blob 5cb3ab812cfe20097c3255f47060a79b3f42ec97 (revision 7). Each of the seven reporting participants names that blob and says it checked it with `git hash-object`. The commit changes the design only: the status line, §4.1's last paragraph and two bullets of §5. The worktree's HEAD during the round was 3457b82c, the minute's commit through round 5; the design there is the same blob.
- **Revision assessed in round 7:** commit 2d9611db, blob d893c1f0c665cdc8f8816435cc67fa5caa95bf0a (revision 8). Each of the nine participants names that blob and says it checked it with `git hash-object`. The range from revision 7 changes the design and `rulings.md` and no code. The minute in the worktree during the round was at 5fc3b8bf and did not yet hold the F6 rows; the brief pointed participants to `r6-fresh6.md` for them.
- **Revision assessed in round 8:** commit bc9b4abd, blob 6d4dacabc89db25a77c11ef0015ca653337ff777 (revision 9). Each of the ten participants names that blob and says it checked it with `git hash-object`. The commit changes two lines of the design, the status line and one phrase in §4.3, and no code: assess-codebase reports `git diff --stat 488e1c7e bc9b4abd -- internal` empty, so no code changed across the whole assessment. The minute in the worktree during the round ran through round 6's late report; participants read round 7's rows from the `r7-*.md` reports.
- **Decision-maker:** perigrin.
- **Drafter:** assess-pushback, from round 2 on. The drafter records and holds no view. In round 1 it was the pushback participant and raised P1 to P12; those rows are the raiser's, not the drafter's. The round-1 minute was drafted after the round closed, from the three reports.
  - In round 2, at the coordinator's direction, it also acted as raiser of P1 to P12: as raiser it released or held them, and it raised P13 to P16. Those round-2 P rows are the raiser's, not the drafter's.
  - In round 3 it acted as raiser of the P rows still held, and raised P17 and P18.
  - In round 4 it acted as raiser of the P rows still held, and raised nothing.
  - In round 5 it acted as raiser of the P rows still held, and raised nothing.
  - In round 6 it acted as raiser of the P rows still held, and raised P19.
  - In round 7 it acted as raiser of the P rows still held, released P19 and raised P20.
  - In round 8 it acted as raiser of the P rows still held, released P20 and raised nothing.
  - From round 2 on it does not count as an independent participant. The independent participants are assess-architecture and assess-codebase (rounds 1 to 3), assess-fresh2 (rounds 2 to 5), assess-fresh3 (rounds 3 to 5), assess-fresh4 (rounds 4 to 6), assess-fresh5 (rounds 5 to 7), assess-fresh6 (rounds 6 and 7; its round-6 report was outstanding when perigrin ruled) assess-fresh7 (rounds 7 and 8) and assess-fresh8 (round 8).
- **Sources:** round 1: `r1-architecture.md`, `r1-codebase.md` and `r1-pushback.md`. Round 2: `r2-architecture.md`, `r2-codebase.md`, `r2-pushback.md` and `r2-fresh2.md`; the author's `response-r1.md`; the rulings recorded in `rulings.md` under "after assess round 1". Round 3: `r3-architecture.md`, `r3-codebase.md`, `r3-pushback.md`, `r3-fresh2.md` and `r3-fresh3.md`; the author's `response-r2.md`; the ruling recorded under "after assess round 2". Round 4: `r4-architecture.md`, `r4-codebase.md`, `r4-pushback.md`, `r4-fresh2.md`, `r4-fresh3.md` and `r4-fresh4.md`; the author's `response-r3.md`; the entry in `rulings.md` under "during assess round 3". Round 5: `r5-architecture.md`, `r5-codebase.md`, `r5-pushback.md`, `r5-fresh2.md`, `r5-fresh3.md`, `r5-fresh4.md` and `r5-fresh5.md`; the author's `response-r4.md`. Round 6: `r6-architecture.md`, `r6-codebase.md`, `r6-pushback.md`, `r6-fresh2.md`, `r6-fresh3.md`, `r6-fresh4.md` and `r6-fresh5.md`; the author's `response-r5.md`; the entry in `rulings.md` under "during assess round 6"; then `r6-fresh6.md` and the entry under "after assess round 6". Round 7: nine reports, `r7-architecture.md`, `r7-codebase.md`, `r7-pushback.md` and `r7-fresh2.md` to `r7-fresh7.md`; the author's `response-r6.md`. Round 8: ten reports, `r8-architecture.md`, `r8-codebase.md`, `r8-pushback.md` and `r8-fresh2.md` to `r8-fresh8.md`; the author's `response-r7.md`. All reports are in the session scratchpad.

IDs keep each participant's own prefix:

- A: assess-architecture, the architecture and direction lens;
- C: assess-codebase, the codebase lens;
- P: assess-pushback, the spec-quality lens;
- F2: assess-fresh2, a fresh participant in round 2. It says it read neither the minute, nor any `r1-*`, `r2-*` or `response-*` file, nor `brief-r2.md`, before writing. It was resumed in round 3;
- F3: assess-fresh3, a fresh participant in round 3. It says it read the design, `rulings.md`, `feasibility.md`, V1's decision and the code, and opened neither the minute, nor any `r1-*`, `r2-*`, `r3-*` or `response-*` file, nor `brief-r2.md` or `brief-r3.md`. It was resumed in round 4, and read the minute then;
- F4: assess-fresh4, a fresh participant in round 4. It says it read the design, `rulings.md`, `feasibility.md`, `context-split.md`, the feasibility code and the code, and opened neither the minute, nor any `r1-` to `r4-` or `response-` file, nor `f2-scratch` or `fresh3-code`. It was resumed in round 5, and read the minute then;
- F5: assess-fresh5, a fresh participant in round 5. It says it did not open the minute, the prior reports and responses, `f2-scratch`, `fresh3-code`, `fresh4-code` or briefs r2 to r5. It discloses that a repository-wide `git grep` for the governing design's filename printed four rows of the minute (A3, A23, F2-4, F3-8), and says it raises nothing on that point. It was resumed in round 6;
- F6: assess-fresh6, a fresh participant in round 6. Its ids carry a letter: F6-C for the rows it declares blocking and F6-R for its residual rows. Its report arrived after perigrin's round-6 ruling. It says it did not open the minute, any `r1-*` to `r6-*` or `response-*` file, or `f2-scratch`, `fresh3-code`, `fresh4-code` or `fresh5-code`. It was resumed in round 7;
- F7: assess-fresh7, a fresh participant in round 7. Its ids carry a letter (F7-P, F7-C, F7-A). It says it did not open the minute, any `r1-*` to `r7-*` or `response-*` file, or any `*-code` or `*-scratch` directory but its own, and that `rulings.md` names finding ids on which it relied on none. It was resumed in round 8;
- F8: assess-fresh8, a fresh participant in round 8. Its ids carry a letter (F8-P, F8-C, F8-A). It says it did not open the minute, any `r1-*` to `r8-*` or `response-*` file, or any `*-code` or `*-scratch` directory. Its disclosure is recorded under "Round 8".

Each row keeps its raiser's classification as declared. After round 1 every status was `held`: that round aired views and nothing had been answered. Status changes in rounds 2 to 8 carry the round number, and only a row's raiser changed its status. A status cell keeps its history, oldest first. A row with no entry for a round was held unchanged in that round. Each row's section reference is to the revision named in its revision column.

## Ledger

| id | raiser | round | blocking or residual | section | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A1 | assess-architecture | 1 | blocking | §4.2 | The subject's central departure ("perigrin ruled that rule 2 is lifted for the precedence value only") rests on a ruling that is not on record. §9 says rulings are recorded the day they are made. | `rulings.md` has six entries: Direction, the §6 assumption, Library-operator narrowing, Precedence refusals, Rule 12, Guards. None mentions rule 2 or a value on items. `grep -n "rule 2" docs/assessments/chart-precedence-v2/rulings.md` prints nothing. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A2 | assess-architecture | 1 | blocking | §4.2, §4.4 | Both rule citations are wrong, and the rule-12 lift does not cover the change to items. Refinement cannot check a lift against a rule it cannot find. | The README has no numbered rules; its forest bullet cites "spec §3 rule 12, §10". Rule 2 is `2026-10-29-chart-parser-design.md` §3 rule 2; the subject's quoted wording ("items hold no context facts") appears in no file. Rule 12 is the governing design's §3 rule 12: "The recognizer stays Boolean and its items stay (state, distance)". §4.2 makes items (state, distance, context), yet the rule-12 lift is "for these refusal records only". | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A3 | assess-architecture | 1 | blocking | header, §2, whole | Review finds live text and an accepted ruling reversed with no record. V1's minute raised this four times (A1, A14, A26, H3) and carried it here. | The subject does not name the governing design (`docs/plans/2026-10-30-chart-parser-v2-design.md`, on `origin/docs/chart-parser-v2-design`), nor what it reverses: §8 ("precedence leaves recognition (perigrin, 2026-10-30)"), rule 6 ("after the chart"), rule 12's "Scope … collapses what remains, after the chart", and the §10 kill "Facts in items". It also reverses V1's own "What this does not amend". `TestPrecedenceIgnoredByRecognition` encodes §8 and is not in §4.4's list. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A4 | assess-architecture | 1 | blocking | §2 anti-goals, §5, §7 | The `set-reuse` stop ("may fall … may not rise") fires on the good direction and allows any fall in the bad one. A fall is the harm rule 2 exists to prevent, and it is what a context in items would cause. | `cmd/ratchet/verdicts.go`, `measureVerdicts`: `ratio("set-reuse", hits, positions)`. README key table: "share of positions whose Earley set was already interned". The ruling says only "move by design" and gives no direction. V1's first-round stop was "set-reuse falls" (`reports/round1-pushback.md`). | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A5 | assess-architecture | 1 | blocking | §2, §7 | The stop "`speed-stage-recognize-aggregate-longlived` rises above 1.0" cannot be read. It names no statistic or run, and the threshold sits inside the noise of a recogniser that did not change. | V1's six verdict runs, recognition identical in all: 0.8609, 0.8710, 0.9043, 0.9162, 0.8857, 0.9629 (`docs/assessments/chart-perly-w3/verdict/run*.stdout`). One internal measurement reached 1.02 (`run1-before.stderr`). The combined prototype read the unchanged recogniser at 0.99 and 0.96. V1 §4 defined "one measurement" and judged medians. `rulings.md` has no ruling for 1.0. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A6 | assess-architecture | 1 | blocking | §5 against §7 | Two different references for one bar: §5's 27 keys "equal 8c3317079's", §7's "unchanged" from "pu with V1, after the baseline retake". No key record exists for 8c3317079, so that side can only be checked by re-running a 10-minute ratchet on a pre-V1 commit. | `grep -rln 8c3317079 docs internal` finds plans, reports and the chain, but no ratchet output. `baseline.txt`'s 29 non-timing keys equal V1's after-run (`diff` against `verdict/run6-after.stdout` prints nothing; 29 lines each side). V1's before-state was 789949d0, not 8c3317079 (`run1-before.meta`). | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A7 | assess-architecture | 1 | blocking | §5, §7 | The fourth §3 counter cannot be zero as a grammar test. As a corpus count it is not a grammar test, and perl.git `t/` was never measured for it. | `feasibility.md` §4: "After narrowing … exactly **1** set still does: set 1626, `{HASHBRACK, PERLY_BRACE_OPEN}`, for `r347 indirob ::= scalar`". The subject does not say what the recogniser does when that set follows `indirob ::= scalar`; the feasibility note says it "would have to carry 'allowed for HASHBRACK only' there or narrow by parent". | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A8 | assess-architecture | 1 | blocking | §4.1 against §4.4 | Followed literally, the guards ("no precedence off-switch, and every guard runs with precedence on") run a mode in which §4.1 has no lookahead terminal `t`. The same gap covers `Expected`, and a recogniser called without narrowing. | `recognize.Options.NoLookahead` is one of `TestSetDigestGolden`'s five option sets, and `TestReferenceRecognizer` runs "under every `Options` combination". `Chart.Expected` re-closes with `b.complete(nil)` (`recognize/expected.go`). The ratchet's recognition keys call `recognize.Recognize` on a raw `lex.Linear` lattice with no scope (`cmd/ratchet/verdicts.go`). On such lattices `feasibility.md` §4 measured 117 tokens with differing reduce verdicts. V1 deferred V2 "once its effect on `Expected` and on refusal reasons is designed for"; §4.4 gives `Expected` one sentence. | held in part (r2): `NoLookahead` and `Expected` are now defined; the un-narrowed-lattice clause of §4.1 is held; released (r3, against revision 4) | 2 (488e1c7e, 487d7c51) |
| A9 | assess-architecture | 1 | blocking | §6 | A ratchet key (`speed-stage-precedence-aggregate-longlived`) is retired, or repointed at test-only code, by the builder and without perigrin. | README, output contract: "no key is added, renamed or retired after it", departed from only by perigrin's rulings (D2 and D5). The governing design's §9.5 says the same. The raiser calls it a question for perigrin, not a builder's choice. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A10 | assess-architecture | 1 | blocking, for perigrin | §7 | The exit is the relative one only, offered as "a proposal for assessment to settle". It passes on any gain past noise, about 0.05 on 3.53, bought with three lifted rules and a spent kill criterion. V1's rulings say V2 sets a target. | V1 §1: "The V2 decision sets its own after this decision's verdict reading". V1 §5: "the V2 decision comes next, setting its own target". V1's after-range is 3.48–3.56. The one measured V2 increment is 0.5% (subject §3). §10 expects "about 0.68, plus about 16% less forest", which nobody has measured. The speed kill is 3×. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| A11 | assess-architecture | 1 | residual | §7 | The bar is weaker than V1's with no reason given. Without a stop for "checks pass but the exit is unmet", a V2 that is correct and not faster has no disposition. | V1 §5 has four things §7 lacks: "A stop overrides the exit"; a stop when the checks pass but the exit is unmet; the clause for a session that cannot be quiet; a budget with a stop at twice the budget. | held (r2, "residual, stands"); the raiser notes the "checks pass but the exit is unmet" stop landed | 2 (488e1c7e, 487d7c51) |
| A12 | assess-architecture | 1 | residual | §4.1, §4.4 | Hidden scope. The tables or their builder must cross a package boundary, and the reference recogniser must read them while sharing no code. | `recognize` imports only `grammar`; `forest` imports `recognize`. The tables are built by `forest.NewPrecedence` and held unexported. `libraryOperators` lives in `forest` and reads a `*Forest`. `TestReferenceSharesNoCode` holds the reference apart. The fuzz toy grammars declare no precedence (V1 minute, F3). | held (r2, "residual, stands") | 2 (488e1c7e, 487d7c51) |
| A13 | assess-architecture | 1 | residual | §4.1, direction | The next decision (interleaved lexing, lookahead 0) inherits a recogniser whose precedence needs lookahead, with nothing on record weighing it. V1's round-1 A11 raised the same trade. | `2026-10-29-chart-parser-design.md` §5: "interleaved means lookahead 0". The governing design's §13 keeps interleaving as future work. | held (r2, "residual, stands"); held (r3, "residual, stands, narrower") | 2 (488e1c7e, 487d7c51) |
| A14 | assess-architecture | 1 | residual | §8, §9 | Two postmortem items are dropped or reversed without a word. | Postmortem, "What Will We Change": "Gate the generation cost. Add a `NewPrecedence` per-process timing key before V2 adds generation work"; absent from §8. "Carry the follow-ups into the next refinement" lists the `shifts` defect and the G1 limit; §2 puts both out, on an assumption `rulings.md` marks "not a ruling". Pruning counts, test dedup and Q12 are not mentioned. | held (r2, "residual, stands") | 2 (488e1c7e, 487d7c51) |
| A15 | assess-architecture | 1 | residual | §4.3, §4.4, §5, §9 | The decision carries milestone content. Each item is a line someone must keep true through execution. | §4.4 names files (`sets.go`, `cache.go`, `leo.go`, `leo_memo.go`, `expected.go`) and what their keys and values hold, which §6 then lists as left to the builder. "Each regeneration is its own reviewed commit" and "The guards are written before the recognizer changes" are chain ordering. §9 is all process. §5's seed numbers and row lists are test detail. | held (r2, "residual, stands") | 2 (488e1c7e, 487d7c51) |
| A16 | assess-architecture | 1 | residual | §4.4 | Two readings of where the refusal records live ("beside the sets"; set interning's "values include the refusal records"). Only one keeps the README's property that sets differing only by side data share one `*Set`. | README, Leo memos: "Memos live beside the sets, never in a set, so two nodes whose sets differ only in their tops share one set (`TestLeoTopOutsideSetIdentity`)". | held (r2, "residual, stands") | 2 (488e1c7e, 487d7c51) |
| A17 | assess-architecture | 1 | residual | §2, §4.4 | The bound on the rule-2 lift ("one of V1's 46 spine tops, or `none`") is prose. Once §10's "Facts in items" kill is spent, nothing mechanical stops a second fact entering items. | The governing design's §10 was the guard. No stop or test in the subject replaces it. | released (r2, against §7's stop on item facts) | 2 (488e1c7e, 487d7c51) |
| A18 | assess-architecture | 1 | residual | §2 anti-goals | The anti-goals read with the force of rulings. Two of them are the orchestrator's: the 1.0 line and "may not rise". | `rulings.md` has neither. The anti-goals are not marked as proposals, the way §7's exit is. | released (r2: both anti-goals it named are gone) | 2 (488e1c7e, 487d7c51) |
| A19 | assess-architecture | 1 | residual | §4.4, §9 | Live text is false from the first recogniser issue until the last ("a final docs issue"). V1 changed the live documents first. | V1 §3: "The live documents change first, in the same pull request". README statements that become false: "Items are an LR(0) state and the distance"; "No fact is added to an Earley item … recognition stays Boolean"; "Leo never changes acceptance"; the key-table rows for `speed-stage-precedence`, `speed-stage-resolve` and `alloc-per-token-chart`; "stay recognition's" (`TestRecognitionKeysUnchangedByPipeline`). | held (r2, "residual, stands") | 2 (488e1c7e, 487d7c51) |
| C1 | assess-codebase | 1 | blocking | §4.4 (library operators; precedence refusals), §2 | V1's refusal reason is decided by stage order (grammar, then scope, then precedence); V2's by position. A file V1 refuses for `grammar` becomes `precedence` at an earlier byte whenever a precedence-dead prefix comes first. The subject does not say this, and "only because" is not observable in a pruned chart without the unpruned recognizer §4.5 removes from production. `refused-where.txt` (336 rows) and `TestParseRefusalMatchesRecognition` know only `table` and `grammar`. | `TestZZProbeNullableAndReasons` on V1: `"$a <=> $b <=> $c;\n"` → `{Byte:0 Reason:precedence}`; `"$a <=> $b <=> $c; }\n"` → `{Byte:18 Reason:grammar}`; `"$a <=> $b;\n"` → accepted. Under §4.4 the chart dies at the second `<=>` (byte 10) in both refused files. `1 .. 2 .. 3; )` → `{Byte:13 Reason:grammar}`. `TestParseRefusalMatchesRecognition` maps `{"table","grammar"}` only. | released (r2, against revision 3); the raiser raises C16 on the rule that replaced the old text | 2 (488e1c7e, 487d7c51) |
| C2 | assess-codebase | 1 | blocking | §2 anti-goals, §7 Stops | The recognize-stage stop can fire from noise, and it would kill a V2 that lowers the total. No rationale is given for 1.0. | `verdict/run{1..6}-*.stdout`: recognize 0.8609, 0.8710, 0.9043, 0.9162, 0.8857, 0.9629. `run6-after.stderr`: `measurement 1: median 0.9807 min 0.823 max 1.048`; single measurements reached a max of 1.08. `git log -- internal/chart/recognize/` shows no recognizer change in V1. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| C3 | assess-codebase | 1 | blocking | §2, §5, §7 | The reason given for "may fall, may not rise" covers `items-per-token` only; fewer items says nothing about `set-reuse`'s direction. Both keys are taken over files recognition accepts, and that population changes under V2. Either key can rise from the population alone, and the Stop then halts on nothing. | README:552 (`set-reuse` definition). `verdicts.go:87`. `c1/c1-report.md` (a precedence value on items): `set-reuse: 0.501 0.502`, `distinct-sets: 10500 11481`. `TestZZFeasCorpus`: `libops-refused=3`. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| C4 | assess-codebase | 1 | blocking | §3 table row 4, §5 "Feasibility as tests", §7 Stops | Counter 4 is a corpus count (0); its static form is already 1. As a static test it cannot pass; as a corpus test it does not guard the grammar. §4 never says what the recognizer does at that set. | `TestZZM4StaticSets`: `after narrowing to <=1 PLUGIN_* (any choice)=1 … r347 indirob ::= scalar before HASHBRACK vs PERLY_BRACE_OPEN`. `TestZZProbe1626`, 26 inputs: after a scalar the `{` token is always set `1850{PERLY_BRACE_OPEN}` with `indirob-scalar-done-before=true`; set 1626 appears after `print STDERR` with `indirob-scalar-done-before=false`. The raiser could not reach the case through the lexer. | released (r2, against revision 3); the raiser raises C14 on the behaviour sentence added in §4.1 | 2 (488e1c7e, 487d7c51) |
| C5 | assess-codebase | 1 | residual | §3 "Contexts rarely meet", §10 item 3, §5 | All 145 conflicting nodes pass under one top and fail entirely under another; no corpus node keeps two different non-empty alternative sets. So V1's union shortcut is never exercised by the corpus, and the corpus forest-equality differential cannot show V2 reproduces it. The instrument also compares `eval`'s first-terminal sets (`x.out`), not alternatives. | `TestZZProbeLocal`: `context split: {Nodes:1410324 Multi:3425 ByOut:145 ByAlts:145 ByAltsNonEmpty:0}`. `context-split.md` code: `o := slices.Clone(x.out)`. | held (r2, residual) | 2 (488e1c7e, 487d7c51) |
| C6 | assess-codebase | 1 | residual | §4.4 "Refusal records beside the sets" | "Never pairs a completion with a parent the recognizer refused" is not sufficient when followed literally. An alternative completed under a context whose parent chain later dies would be added to a node reached under another context; V1 does not keep it. Build has to walk by context. The raiser says this is implementation detail in the decision. | `forest/forest.go`: `build`, `split`, `alt`, `item` read facts as (symbol, origin, rule) with no context. `TestZZProbeLocal`: `marked nodes=1282773 dropped alts on marked nodes=20336 of them passing in some eval-visited context=0`. Reasoned for inputs outside the corpus, not measured. | held (r2, residual); held (r7, residual): §4.3's tie covers the parent side; Build walking by context is still unsaid | 2 (488e1c7e, 487d7c51) |
| C7 | assess-codebase | 1 | residual | §4.1 | The recognizer's lookahead is a terminal set, and §4.1 defines no verdict when the set's terminals disagree. Without narrowing that is 117 corpus tokens, and many callers recognize without narrowing; narrowing needs the scope and the source, which those callers do not all hold. | `TestZZFeasCorpus`: `M4 unnarrowed: … reduce-diff=117`. `TestZZM4StaticSets`: `differing reduce verdict unnarrowed=63`. Callers: `cmd/ratchet/verdicts.go:279`, `speed.go` (4 sites), `bench.go`, `chart/precedence.go:245`, `forest/bench.go:20`. `grep -rn "recognize\.New(" internal \| wc -l` → 36. | held (r2, residual); the raiser says §4.1's new rule for un-narrowed lattices is defective, C15; released (r3: §4.1 defines the verdict over a set and says what an un-narrowed lattice gets; "who narrows for the other callers is P4's row") | 2 (488e1c7e, 487d7c51) |
| C8 | assess-codebase | 1 | residual (ruling conflict, for perigrin) | §4.4 guards, rulings "Regenerate everything" | None of the guards knows precedence today, and their grammars declare no levels, so regenerating them tests nothing about precedence. Giving fuzzed grammars levels meets the limit that §4's mechanism is exact only where the four counters are zero. A recognizer over a grammar with no levels prunes nothing, which is the old recognizer in effect, beside "no off-switch" and §4.5's test-only copy. | `grep -n "Prec\|Level\|prec" recognize/fuzz_test.go recognize/reference_test.go recognize/setdigest_test.go forest/forestdigest_test.go` → no output. `forest/precedence.go: stateFree` returns `a precedence key is state-dependent`. | held (r2, residual, for perigrin); held (r3, residual, for perigrin), "narrowed but not gone"; held (r4, residual), "narrower again"; the raiser's round-4 entry no longer marks it for perigrin | 2 (488e1c7e, 487d7c51) |
| C9 | assess-codebase | 1 | residual | §4.4 "the forest digest … regenerated", §5 | `resolved-forest-digest.golden` digests `Parse`'s output for 438 files. If V2's forest equals V1's it must not change by one byte; regenerating it would hide the failure it exists to catch. `forest-digest.golden` (Build's output) does change. | `forest/forestdigest_test.go: TestResolvedForestDigestGolden` calls `Parse(tab, src, Scope{Imports: …})`; `forestText`: "Nothing in it depends on pointers, allocation or Alt order". | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| C10 | assess-codebase | 1 | residual | §4.4, §6 (hidden scope) | Not listed in §4.4, and each forced by the code: (a) `TestRecognitionKeysUnchangedByPipeline` fatals under V2, and README's "stay recognition's" claim goes with it; (b) `Chart.Expected` has no lookahead terminal to judge a reduction by and feeds `leo-disagree`; (c) `Options.NoLookahead` promises "Neither changes acceptance"; (d) the two `TestItemIsStateDistance*` tests; (e) the tables' move out of `forest`; (f) the referee needs the old Build as well as the old recognizer. | `cmd/ratchet/verdicts_test.go:190-197`; `recognize/expected.go:39-45`; `recognize/recognize.go:24-33`; `recognize/recognize_test.go:258-285`; `leodisagree.go:34`. | held (r2, residual): (b) answered, (c) is now C15, (a), (d), (e), (f) unchanged; held (r3, residual): (b) and (c) answered, (a), (d), (e) stand, (f) "changes shape"; held (r4, residual): (a), (d), (e) stand; (f) stands, with F3-4; held (r5, residual): (a), (d), (e), (f) stand | 2 (488e1c7e, 487d7c51) |
| C11 | assess-codebase | 1 | residual | §4.3, §10 | 0.01% is how often two contexts at one node disagree, not how often an item carries a context other than `none`. The effect on set interning and the transition cache is not measured in §3. The raiser says this bears on C2. | `TestZZProbeLocal`: `shift records=574274 with a non-none canonical top=73513` (12.8%). `c1/c1-report.md` lines 1–5 and 238–246: distinct sets +9.3%, cache hits −0.5 points. | held (r2, residual); held (r7, residual); the raiser says §4.3 now gives a second reason an item keeps a context | 2 (488e1c7e, 487d7c51) |
| C12 | assess-codebase | 1 | residual | §3, §5 | The corpus has no precedence refusal and no kept library operator. Refusal parity and kept-operator behaviour rest on toys, referee sentences and the static check alone. | `TestZZProbeLocal`: `pipeline reasons: map[:1837 grammar:30 lex-unmapped:9 scope:5]`, `precedence-refused=0`. `TestZZFeasCorpus`: `plugin-after=0`. | held (r2, residual) | 2 (488e1c7e, 487d7c51) |
| C13 | assess-codebase | 1 | residual | §3 "The tables do not cover … Leo memos" | Leo is partly covered already, for the reduce side: the combined prototype refused the whole chain when any link might be refused. The shift side and the context on the memo are new. | `chart-perly-w3/measurement/combined-compare.txt` line 2 (`refused Leo links 0` over 2,211 files); `prototypes/combined.diff.gz` (`recognize/leo.go`, `leo_memo.go`). | held (r2, residual) | 2 (488e1c7e, 487d7c51) |
| P1 | assess-pushback | 1 | blocking | §2 anti-goal "A file's verdict changes"; §4.4; §5 "The comparison covers the verdict" | "Verdict" does not say whether the refusal reason is part of it. If it is, the anti-goal and the §5 differential cannot pass. If it is not, the differential's "verdict" must say so, and `refused-where.txt` rows change reason as well as byte, which §4.4 does not say. §4.4's "died only because of a precedence refusal" cannot reproduce V1's reason in general. | Scratch test, `Parse(tab, src, Scope{})`: `"$a <=> $b <=> $c;"` → `{Byte:0 Reason:precedence}`; `"$a <=> $b <=> $c +;"` → `{Byte:18 Reason:grammar}`; control `"$a <=> $b +;"` → `{Byte:11 Reason:grammar}`. V1's order is `forest.parseStream`'s. `refused-where.txt`: 282 `grammar`, 54 `table`, 0 `precedence`. Not measured: how many of the 282 rows move. | released (r2, against revision 3); the raiser raises P14 on the rule that replaced the old text | 2 (488e1c7e, 487d7c51) |
| P2 | assess-pushback | 1 | blocking | §2 anti-goal "The recognize stage rises above 1.0"; §7 stop | The threshold names no statistic, run count or protocol, and the unchanged recognizer already reads within 4% of it. V2 fails or stops on a reading noise can produce, or the builder picks the statistic that passes. V2 adds work to this stage by design. | `git diff --stat 8c3317079 HEAD -- internal/chart/recognize` is empty. `verdict/run*.stdout`: 0.8609, 0.8710, 0.9043, 0.9162, 0.8857, 0.9629. Per-measurement maxima in `run*.stderr` reach 1.042, 1.048, 1.056, 1.08. `baseline.txt` holds 0.9209. The subject does not say which stage the narrowing's cost lands in. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| P3 | assess-pushback | 1 | blocking | §5 "Feasibility as tests"; §7 stop; §3 last table row; §4.1 | The fourth test is red on the day it is written, or is written over the corpus, where it cannot see a grammar change. On an input that reaches the case the recognizer's behaviour is unspecified, and forest equality with V1 is not argued. | `feasibility.md` §4: "exactly **1** set still does: set 1626". Corpus: 102 tokens carry the set, 0 after the completion. perl.git (the raiser's run): 78 tokens, 0 after it. 14 inputs tried, none reached it; reachability unknown, not disproved. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| P4 | assess-pushback | 1 | residual | §4.4 "Library operators, before recognition" | "Applied to the lattice" has no representation today for an operator that is in scope, and the subject does not say who narrows for callers that have no scope. Unstated work: new terminal sets in the generated tables or a different lattice input. `leo-disagree` and the recognize stage key would time or judge a different input from production's. | Scratch test `TestZZPBNarrowedSets`: `sets=1880 withPlugin=62 narrowed-to-none missing=0 narrowed-to-one total=682 missing=682`. README: "Input is a lattice of terminal-set ids". Direct `recognize.Recognize(…, lex.Linear(s))` callers outside tests: `speed.go` (3), `verdicts.go`, `chart/bench.go` (2), `chart/precedence.go`, `forest/bench.go`, `recognize/leodisagree.go`. | held (r2, residual); held (r5, residual), with new evidence: 795 of the 1,296 referee sentences need a single-terminal set the tables lack | 2 (488e1c7e, 487d7c51) |
| P5 | assess-pushback | 1 | residual | §4.1; §4.5; §6 last bullet; §4.4 "The guards" | Scope the subject implies but does not size. Refinement finds it mid-chain, which is §7's "amendment of a started issue" stop. | (a) `forest` imports `recognize`; `Precedence` is in `forest/precedence.go` (844 lines). (b) 1,212 core items against V1's 1,224, so an item-id mapping is needed. (c) The test-only old recognizer is 1,223 lines over eight files, plus today's `Build` pairing and the forest-side `libraryOperators` prune. (d) `reference_test.go` (239 lines) "gains precedence". (e) §6's "build-tagged path" under "no precedence off-switch". | held in part (r2, residual): (e) released, (a) to (d) held; held in part (r3, residual): (c) released, (a), (b), (d) held | 2 (488e1c7e, 487d7c51) |
| P6 | assess-pushback | 1 | residual (conflict for perigrin) | §4.4 "the set digest, the forest digest and the brute-force outside reference are regenerated" | There are two forest goldens and the subject names one. Read with "Regenerate everything", `resolved-forest-digest.golden` is regenerated too, and the one in-tree check that already states the goal is erased by the change it should judge. | `forest/forestdigest_test.go`: `TestForestDigestGolden` (Build's forest) and `TestResolvedForestDigestGolden` (`Parse`'s final forest). Both doc comments say "Never regenerate it to make a change pass". | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| P7 | assess-pushback | 1 | residual | §2 anti-goal and §7 stop on `items-per-token` and `set-reuse` | The direction is asserted, not argued; for `set-reuse` a fall is fewer cache hits and a rise is more. The population also changes. A correct V2 in which pruning makes two sets intern equal trips a stop, while an unbounded fall passes. | `measureVerdicts` sums only files with `r.Recognized = c.Accepted()`. In V2 the 3 corpus files `LibraryOperators` refuses leave that population. README key table: both are "reported", not gated. The raiser did not show a rise occurs. | released (r2, against revision 3) | 2 (488e1c7e, 487d7c51) |
| P8 | assess-pushback | 1 | residual | §7 exit; §5 "Keys"; §2 "27 of the 29 non-timing keys" | (a) "Before" pins no commit, and §5 pins the keys to `8c3317079`, a pre-V1 commit. (b) Nothing in the exit asserts the pass left production or that the shift side prunes in the recognizer. (c) "Non-timing" silently leaves out the six `alloc-*` and `gc-*` keys, which V2's larger items will move. | `baseline.txt` has 47 keys: 12 timing, 6 alloc and gc, 29 others. `git diff 8c3317079 HEAD -- …/baseline.txt` changes only speed, stage and gc lines. Exit list, §7. | held in part (r2, residual): (a) released, (b) and (c) held | 2 (488e1c7e, 487d7c51) |
| P9 | assess-pushback | 1 | residual | §5 "Forest equality": "the reachable node set" | The comparison is made before `resolve` and over reachable nodes, but `resolve` reads every node. A file with a package, class or sub declaration node that only a forbidden derivation holds would resolve differently in V2 with the forest differential green. Measured: no effect on either population. | `resolver.declarations` loops `for _, nd := range f.Nodes[1:]`. Scratch test `TestZZPBOrphans`: corpus `judged=1839 … unreachable-nodes=156526 files-with-unreachable-decl-nodes=0 resolve-differs=0`; perl.git `judged=617 … unreachable-nodes=172971 files-with-unreachable-decl-nodes=0 resolve-differs=0`. The differing case was not constructed. | held (r2, residual) | 2 (488e1c7e, 487d7c51) |
| P10 | assess-pushback | 1 | residual (evidence for §3) | §3 "The tables do not cover … perl.git `t/`"; §5 "The perl.git reference" | The perl.git gap closes clean when measured. The failure path stays thin: the perl.git differential is behind a build tag and a make target, so a perl.git-only forest difference is found late and by hand; the subject's answer is "stop", with no fallback named. | The committed feasibility code with the loader swapped for `perlGit(perl5Root(), srcs)`: `files=620 recognized(forest)=617 libops-refused=0 judged=617`; `core items differ: 0`; `needing >=2 verdict vectors=769 (cross-core: 0)`; `over-keep=0`; static-model misses 0; `M4 narrowed: reduce-diff=0`; `M4 unnarrowed: reduce-diff=432`. Output kept at `pb-perlgit.out` in the session scratchpad. | held (r2, residual) | 2 (488e1c7e, 487d7c51) |
| P11 | assess-pushback | 1 | residual | whole subject; §1 "V2 changes how precedence is applied, not what it decides" | One decision, but two observable behaviour changes (refusal reason and byte; `Expected`) sit under a sentence that denies any. The narrowing move is a precursor that could be judged alone, with V1's pass still on. | §4.4 "Expected". `leo-disagree` compares `Expected` at every position (README key table). `feasibility.md` Verdict item 3 makes the narrowing a precondition of the reduce side. | held (r2, residual); held (r3, residual), "smaller" | 2 (488e1c7e, 487d7c51) |
| P12 | assess-pushback | 1 | residual | §4.2 against §4.3 | Two collapse rules: §4.2's (a top that forbids no shift is `none`) and §4.3's (`none` "wherever its context cannot change a verdict reachable from it"), which has no stated way to compute it and is not in §6. Both chart-shape keys depend on which is built; the cost side has no check. | §4.2 and §4.3 of the subject. `feasibility.md` §3.2 computes reachability in scratch code only. | held (r2, residual); held (r7, residual, "changed shape": the correctness half is removed; what stays is that the rule names no way to compute it and is not in §6) | 2 (488e1c7e, 487d7c51) |
| A20 | assess-architecture | 2 | blocking | §4.4 "Refusal reasons are positional" | The standard precedence refusal reports `grammar`, and "Keep 'precedence'" keeps almost nothing: a reduce-side refusal removes its item in an earlier set than the one that dies. | `print $fh [];` is a `waypoint2Refusals` row. The refused completion is `indirob ::= scalar` before `[`, in the set after `$fh`; V1's minute row H1 measured the chart dying at `{Byte:11 …}`, the `]`, two sets later. No precedence refusal happens while that set is built. The standing ruling's words are "died only because of a precedence refusal". Rests on V1's recorded measurement and on reading §4.4; nothing was run. | held; released (r3, by ruling "Diagnose by rerun"); the raiser raises A26 on the byte claim that came with the fix | 3 (a8808dd2, 71d5f7bb) |
| A21 | assess-architecture | 2 | residual | header | One more departure from the governing design is unlisted in the supersession. | Governing design, rule 12: "Scope … collapses what remains, after the chart in waypoint 2". §4.4 narrows library operators before recognition. The ruling "Narrow before recognition" covers it; the header list does not name it. | held | 3 (a8808dd2, 71d5f7bb) |
| A22 | assess-architecture | 2 | residual | §2, last line | The key's retirement has no home in §4 or §7, and it touches `baseline.txt` inside the milestone, beside the non-goal on retaking it. | README, exit rule: the ratchet exits 2 on a baseline that "names a key the ratchet does not print". Also touched: `TestPipelineStagePrecedenceTimesConstrain` and the README key table. | held | 3 (a8808dd2, 71d5f7bb) |
| A23 | assess-architecture | 2 | residual | header | A reader cannot open the document the decision supersedes parts of. | `ls docs/plans/2026-10-30-chart-parser-v2-design.md`: "No such file or directory". It is on `origin/docs/chart-parser-v2-design`. V1's header named the branch. | held | 3 (a8808dd2, 71d5f7bb) |
| A24 | assess-architecture | 2 | residual | header, rule-6 bullet | Rule 2 is quoted inside the rule-6 bullet with no statement of whether it is superseded, so a reader cannot tell if it still binds. | The ruling lists §8, rule 6 and rule 12 only. Rule 2 is about dynamic context (imports, pragmas); the spine top is a grammar fact. | held | 3 (a8808dd2, 71d5f7bb) |
| A25 | assess-architecture | 2 | residual | §4.1 | "With narrowed input, the only set where this can happen is set 1626" is true for linear lattices only. | README, class keys: "Where edges with different sets leave one node, its lookahead is their union". §3 already says the tables do not cover branched lattices. | held | 3 (a8808dd2, 71d5f7bb) |
| C14 | assess-codebase | 2 | blocking | §4.1 reduce side; §3 row 4 | The "every terminal in the set" rule is false when read literally, and common. A token's set holds readings that cannot follow the completed rule's LHS, and V1's table allows the reduction before those by default. So the forbidden completion is kept and the forest differs from V1's on ordinary code. §3's count of 1 was measured over terminals that can follow the LHS. | Probe `TestZZR2Corpus`, narrowed corpus: `multi-terminal with a set=73547; a chart completion forbidden before a following terminal and allowed before another terminal of the token: any=17100, one that can also follow=0`. Top case: `15590 r168 listexpr ::= term forbidden before ASSIGNOP, allowed before PERLY_EQUAL_SIGN`. Static, `TestZZR2`: `any terminal=224 sets; one that can also follow=0 sets`. `TestZZM4StaticSets` tests `inFollow` on both terminals. Shown with V1's tables, not with V2 running. | held; released (r3, against revision 4) | 3 (a8808dd2, 71d5f7bb) |
| C15 | assess-codebase | 2 | blocking | §4.1, second sub-bullet | "Where there is no lookahead, the completion is not judged on the reduce side" is wrong three ways: (a) the lookahead set always exists; (b) the recognizer cannot tell a narrowed lattice from an un-narrowed one; (c) skipping the reduce side changes acceptance, which breaks `Options`' contract, `TestLookaheadKeepsAcceptance`, `TestExpectedIgnoresLookahead`, the fuzz option matrix and the set digest's `NoLookahead` line, and means the ratchet's `Recognized` accepts what production refuses. | `recognize/recognize.go: Recognize` (`la, laBits := r.lookahead(out[pos])` before the option is read); `Options` doc. `TestZZR2`, V1's evaluator with the reduce check skipped: `"1 .. 2 .. 3;\n" V1 pipeline={Byte:0 Reason:precedence}; root passes with both sides=false, with the shift side only=true`; the same for `"$a <=> $b <=> $c;\n"`; `"$a <=> $b;\n"` passes both ways. `recognize/fuzz_test.go:298`; `recognize/setdigest_test.go:32`. | held; released (r3, against revision 4) | 3 (a8808dd2, 71d5f7bb) |
| C16 | assess-codebase | 2 | blocking | §4.4 "Refusal reasons are positional" | The rule names three of the four refusal reasons and one of the two refusal sites. (a) `lex-unmapped` is missing, so 9 corpus files and the `table` rows of `refused-where.txt` become `grammar`. (b) A file can be refused with no dead set (the file ends too soon), and the rule gives it no reason. (c), which the raiser calls residual beside these: the `scope` clause never fires on the corpus, and the first ruling's "keeps its reason `scope`" still stands in `rulings.md`. | `forest/forest.go: parseStream` (`at := len(src)`, `Reason: "lex-unmapped"`); `pipelineReason` has five entries. `TestZZR2Corpus`: `reasons map[:1837 grammar:30 lex-unmapped:9 scope:5]`; `libops-refused files=3; tokens narrowing leaves with no terminal in them=0`. | held; released (r3, by ruling "Diagnose by rerun"); the raiser raises C19 on the byte evidence | 3 (a8808dd2, 71d5f7bb) |
| C17 | assess-codebase | 2 | residual | §3, after the table | "in the corpus or in 14 constructed inputs" cites evidence a reader cannot find: no committed file holds 14 constructed inputs. | `grep -rn "constructed inputs" docs/assessments/chart-precedence-v2/` → no hit outside the design. `git diff --stat 488e1c7e a8808dd2` → three files, none under `feasibility-code/`. The raiser's own round-1 probe ran 26. | held | 3 (a8808dd2, 71d5f7bb) |
| P13 | assess-pushback | 2 | blocking | §4.1 third sub-bullet, against §4.1's first sentence and §4.4 "The guards" | (1) Under `NoLookahead` the recognizer applies the shift side only, so it accepts files the default mode refuses and builds a larger forest; that breaks three contracts in the tree and makes `NoLookahead` a reduce-side off-switch beside "no precedence off-switch". (2) The same bullet calls an un-narrowed lattice "no lookahead", against the every-terminal rule in §4.1's first sentence; the recognizer cannot tell the two kinds of lattice apart without a flag. | Scratch run, `Precedence.reducesFree` forced true. Normal: `"$a <=> $b <=> $c;"` and `"1 .. 2 .. 3;"` → `Reason:precedence`. Reduce side off: both accepted (31 nodes/25 alternatives; 26/20); `"1 + 2 * 3 - 4;"` 35/28 against 32/24; control `"1 + 2;"` 20/16 both ways. A stand-in for V2 under `NoLookahead`; none was built. Contracts: `recognize.Options` doc; `TestLookaheadKeepsAcceptance`; `forest/forest_test.go` `optionSets`. | held; released (r3, against revision 4) | 3 (a8808dd2, 71d5f7bb) |
| P14 | assess-pushback | 2 | blocking | §4.4 "Refusal reasons are positional", first sub-bullet | Followed literally: (a) `lex-unmapped` no longer exists as a reason; (b) the sentence is false of files `resolve` refuses, which have a forest and no dead set; (c) `scope` from narrowing can arise at one terminal set only, and there it is the condition V1 reads as `lex-unmapped`; (d) "while that set was built" does not say which set a refused completion belongs to. | (a) `pipelineReason`; 54 `table` rows in `refused-where.txt`; `forest.parseStream`. (b) `parseStream`: `resolve(…) >= 0` → `Reason: "scope"` after Build. (c) Scratch test: `sets=1880 withPlugin=62 ofThemOnlyPlugin=1`. (d) The nonassoc inputs of P13 are refused through the reduce side. The `lex-unmapped` key is computed from lexing (`cmd/ratchet/lex.go`), not from `Refusal`. | held; released (r3, by ruling "Diagnose by rerun") | 3 (a8808dd2, 71d5f7bb) |
| P15 | assess-pushback | 2 | residual | §7 stop "any input reaches set 1626 after `indirob ::= scalar`"; §5 | The stop has no named observer. §5's feasibility tests are "over the tables" and cannot see an input; production keeps the completion silently. | §5 "Feasibility as tests"; the committed `zz_feas_ratchet_test.go.txt` holds the counter (`n1626done`) as scratch code only. | held | 3 (a8808dd2, 71d5f7bb) |
| P16 | assess-pushback | 2 | residual | §2 (the retired key; the non-goal on retaking `baseline.txt`); §8 | Retiring the key removes a line from `baseline.txt` and from the key list after §8's retake. That is not a retake, but the non-goal reads as "baseline.txt is not touched". | `cmd/ratchet/compare.go`: `return fmt.Errorf("%s is missing", k)`; `main.go` key list holds the key. | held | 3 (a8808dd2, 71d5f7bb) |
| F2-1 | assess-fresh2 | 2 | blocking (false claim; a step that does the wrong thing literally) | §4.1 reduce side | V1's `reduceAt` is 0 for a terminal that cannot follow the rule's LHS, so a set holding one forbidden FOLLOW terminal and one non-FOLLOW terminal is "mixed" and kept. The first corpus file with `=` then binary `-` differs from V1, which is a §7 stop. The claim that only set 1626 can mix verdicts is false under the subject's own wording. | Scratch test `TestZZF2Literal`, PLUGIN terminals dropped: 2,350 (set, rule) pairs over 163 sets are kept by the literal rule where every FOLLOW terminal forbids; 0 multi-terminal pairs are refused by it. `TestZZF2Example`: `r199 termbinop ::= term ASSIGNOP term` before `ADDOP`: `reduceAt=1`, in FOLLOW; before `PERLY_MINUS`: `reduceAt=0`, not in FOLLOW; the `-` token's set is `{ADDOP, PERLY_MINUS}`. V1 on `$x = 1 - 2;` prunes 30 reachable nodes to 27. With the FOLLOW restriction the count is 1 (set 1626). | held; released (r3, against revision 4) | 3 (a8808dd2, 71d5f7bb) |
| F2-2 | assess-fresh2 | 2 | blocking (a criterion that cannot pass; §3 and §5 contradict) | §5 "The existing differentials stay"; §3 "holds only for perly.y's current shape"; §7 "a forest differs from V1's on any input" | The toy differentials run on the recognizer's forest, and toy `emptyred` breaks the property §4.1's shift rule rests on: both its forbidden shift keys have a nonterminal after `at`'s dot. By the raiser's reading the V2 recognizer accepts `VAR LB NUM RB LB NUM RB`, which bison, `Constrain` and the reference refuse. Either the toy check fails and trips a stop, or the fix needs an item fact beyond the spine top, also a stop. | Scratch test `TestZZF2Toys`: `emptyred`: forbidden shifts own-terminal=0, other=2; M3 cross-core=1, needs copies=1, over-keep=1. `nullableShift` toy: other=1. The other 11 toys: other=0, over-keep=0. `TestZZF2EmptyRed`: of 7 sentences, `Constrain` refuses 1 the chart recognizes. The V2 behaviour is reasoned, not run. | held; released (r3, against revision 4); the raiser says the fix left F2-6 and F2-10 open | 3 (a8808dd2, 71d5f7bb) |
| F2-3 | assess-fresh2 | 2 | residual | §4.1 "Where there is no lookahead, the completion is not judged on the reduce side" | With the reduce side off and the shift side on, a nonassoc chain such as `1 .. 2 .. 3` keeps its left-nested reading and is accepted, against `Options`' "Neither changes acceptance". The option, `TestLookaheadKeepsAcceptance`'s premise and the reference recognizer's precedence need one stated rule. | `recognize/recognize.go` `Options`; `recognize/recognize_test.go` `TestLookaheadKeepsAcceptance`. Reasoned from `reducesIn`/`shifts`, not run. | held; released (r3) | 3 (a8808dd2, 71d5f7bb) |
| F2-4 | assess-fresh2 | 2 | residual | header, "Governed by" | Three more clauses of the governing design are reversed and not named: §2 ("conflicts are kept as ambiguity and resolved in the forest"), §6's data flow, and rule 12's "Scope … collapses what remains, after the chart". The rulings cover the substance; the list is short. | `git show origin/docs/chart-parser-v2-design:docs/plans/2026-10-30-chart-parser-v2-design.md`, lines 50, 105–111, 251. V1 named "v2 §2, §8 and rule 6". | held | 3 (a8808dd2, 71d5f7bb) |
| F2-5 | assess-fresh2 | 2 | residual | §4.4 refusal reasons | (a) No reason is defined when every set survives and the file is still not accepted. (b) A V1 `scope` refusal whose narrowed token keeps a non-PLUGIN terminal becomes `grammar`; `rulings.md`'s library-operator entry still reads "A scope refusal keeps its reason `scope`". | `forest.parseStream`; `TestZZF2Literal`: "only PLUGIN_*=1, with other terminals=61"; `rulings.md`, both entries. | held; released (r3, by ruling "Diagnose by rerun"); (a) "shrinks to one case, raised as F2-11" | 3 (a8808dd2, 71d5f7bb) |
| F2-6 | assess-fresh2 | 2 | residual; blocking from round 3 (promoted by its raiser) | §5 "Feasibility as tests" | §4.1's shift rule also rests on every forbidden shift being of `at`'s own next terminal. It is in §3's table row but not clearly among the tests. | `feasibility.md` §3.0; F2-2's toy counts. | held; held; **promoted to blocking by its raiser, assess-fresh2, in round 3**, which now holds it as promoter; released (r4, by its raiser and promoter, against §5's fourth zero counter) | 3 (a8808dd2, 71d5f7bb) |
| F2-7 | assess-fresh2 | 2 | residual | §4.1 "a recognizer given an un-narrowed lattice"; §4.4 "There is no precedence off-switch" | The ratchet's direct `recognize.Recognize` callers pass un-narrowed lattices and so run with the reduce side off, which is not what ships. How a recognizer tells the two kinds of lattice apart is not said. | `cmd/ratchet/verdicts.go` `chartInputs`; `cmd/ratchet/speed.go`; `precedence.go` `builtForest`. | held; held (r3, residual, "narrowed"); held (r5, residual, "narrowed again": `chart.builtForest` leaves the list) | 3 (a8808dd2, 71d5f7bb) |
| F2-8 | assess-fresh2 | 2 | residual | §4.1 "Both verdicts come from V1's generated tables, read as they are" | The tables live in `forest`, which imports `recognize`; they must move or be passed in. A recognizer built without a precedence is an off-switch in fact. | `forest/precedence.go` `Precedence`; `recognize.New`. | held; held in part (r3, residual): the package boundary stands; the other half is now F2-9 | 3 (a8808dd2, 71d5f7bb) |
| A26 | assess-architecture | 3 | blocking | §4.4 "Refusal reasons are diagnosed by rerun"; ruling "Diagnose by rerun" | "Every other reason and byte matches V1's" is false for a `scope` refusal from narrowing: its byte moves, and the first test that compares it goes red, which is §7's "criterion change" stop mid-milestone. The two rulings also disagree: the first says a scope refusal's "byte may move", the last says "Only a `precedence` refusal's byte moves". | Scratch test at `1fab4a89`: V1's `forest.Parse` under `Scope{}`, against where recognition dies on the same tokens with every `PLUGIN_*` terminal removed. Three move: `foo plus 2 plus 3;` V1 `scope` byte 4, rerun dies at 11; `foo bar 1 plus 2;` 4 against 10; `f $x plus 2 plus 3;` 5 against 12. Three match: `1 plus 2;` 2 and 2; `print $x plus 2 plus 3;` 16 and 16; `my $x = 1 plus 2 plus 3;` 10 and 10. Control `1 + 2;` accepted by both. Cause given: `prune` returns "the byte of the first dropped reading reached from the root". Not probed: a file with a library operator in scope. | held; released (r4, against revision 5 and the note in `rulings.md`) | 4 (1fab4a89, 68ee2dd6) |
| A27 | assess-architecture | 3 | blocking | §4.4 "The guards" against §4.5 and §5 | "No caller can parse with precedence off, since the no-precedence mode only classifies refusals" is false as soon as §4.5 is built, and the ruling's "used only to classify refusals" is stretched without a word. Refinement must choose between the two. | §4.5: "the referees use it and no separate copy is kept". §5: two toys "stay judged by `Constrain` over the no-precedence mode". The differentials sit in three packages (`forest`, `chart`, `cmd/ratchet`); a Go test in `chart` or `cmd/ratchet` reaches `recognize` only through exported names. §6 has the claim that can hold: "no caller of `Parse` can turn precedence off". | held; released (r4, against revision 5; the raiser judges the referees' test use fits the ruling) | 4 (1fab4a89, 68ee2dd6) |
| A28 | assess-architecture | 3 | residual | §4.4, direction | "The cost falls on refused files only": in the editing loop the long view aims at, refused input is the usual case, and each refusal now costs up to three recognitions. Nothing weighs it. | Governing design §13: "mid-edit code is incomplete". The raiser says the speed exit cannot see it, the aggregate being drawn from files both parsers accept (README, "Recursive descent has two sides"). | held | 4 (1fab4a89, 68ee2dd6) |
| A29 | assess-architecture | 3 | residual | §4.1, reduce side | Where no terminal in the set can follow the LHS, "every" is vacuously true and the completion is refused. Under `NoLookahead` that acts as a FOLLOW filter, which the option exists to switch off. Acceptance is the same, but the sets differ. | `recognize.Options.NoLookahead`: "completes every complete item, whatever follows". §4.1: "`Options.NoLookahead` withholds it only from the FOLLOW filter". | held | 4 (1fab4a89, 68ee2dd6) |
| C18 | assess-codebase | 3 | blocking (touches a ruling, for perigrin) | §4.4 "The guards", against §4.5 and §5 | Two other parts of the subject parse with precedence off: §4.5's referees and §5's two failing toys. The referees live in three packages and the rerun itself is in `forest`, so the mode has to be selectable from outside `recognize`. The ruling's own words are "used only to classify refusals"; the referee use goes past them. | Subject §4.4, §4.5, §5. `rulings.md`, "after assess round 2". `forest/forest.go: parseStream` calls `recognize.Recognize`; differentials in `forest/precedence_reference_test.go`, `chart/precedence.go: builtForest`, `cmd/ratchet/precedence_reference_test.go`. §6's "no caller of `Parse`" is "the true form". | held; released (r4, against revision 5; the raiser takes the `rulings.md` note as settling the conflict it flagged) | 4 (1fab4a89, 68ee2dd6) |
| C19 | assess-codebase | 3 | residual | §4.4 "Every other reason and byte matches V1's" | V1's `scope` byte comes from a forest walk; the rerun's comes from where the narrowed chart dies. The two are defined differently and agree on all the corpus has: 3 files. A test over more `scope` refusals should pin it. | `forest/resolve.go: prune`. `TestZZR3Corpus`: the three guacamole files at bytes 22, 24, 23, equal both ways; `narrowed-recognition verdict differs from V1 recognize+LibraryOperators on 0` of 1,881 files. | held; released (r4: §4.4 no longer claims the `scope` byte matches) | 4 (1fab4a89, 68ee2dd6) |
| P17 | assess-pushback | 3 | blocking (two parts contradict; a conflict with the ruling's words, for perigrin) | §4.4 "The guards" against §4.5, §5 "existing differentials" and §4.1's last paragraph | The same revision names three other uses of the mode, each of which parses with precedence off: the referees (§4.5), the two toys (§5), and any grammar whose tables are refused (§4.1), which covers fuzzed or toy grammars that fail §3's checks, so those guards do not run "with precedence on". The mode has to be reachable through `recognize`'s exported surface. | Subject §4.4, §4.5, §5, §4.1 last paragraph, §6 last bullet. `rulings.md`, "after assess round 2", last sub-bullet. `cmd/ratchet/precedence_reference_test.go: comparePrecedence`, `chart/precedence.go: builtForest`, the `forest` package's tests. `recognize` imports only `grammar` and `lex`. | held; released (r4, against revision 5; the raiser judges the referees' use fits the ruling) | 4 (1fab4a89, 68ee2dd6) |
| P18 | assess-pushback | 3 | residual | §4.4 "diagnosed by rerun"; §7 exit | The rerun's cost and state are unstated. A refused file is recognized up to three times inside the timed pipeline, and `ParseWith` takes the caller's long-lived `*Recognizer`, whose intern table and cache would then serve both modes. | `cmd/ratchet/speed.go: pipelineSide` runs `pipelineParse` on each path and accepts any `pipelineReason`. Corpus reasons: 30 `grammar`, 9 `lex-unmapped`, 5 `scope`. `forest.ParseWith` doc. | held | 4 (1fab4a89, 68ee2dd6) |
| F2-9 | assess-fresh2 | 3 | residual (wording against a ruling, for perigrin) | §4.4 "The guards"; §4.5; rulings "Diagnose by rerun" | §4.5 and §5 have the referees recognize in the no-precedence mode and Build its chart, which is parsing with precedence off. The ruling's words are "used only to classify refusals" and "not a way to parse". The mode must be exported from `recognize`. | §4.4, §4.5, §6 of the subject; `rulings.md`, "after assess round 2", last bullet. §6's "no caller of `Parse`" is "the reading that reconciles them". | held; released (r4, against revision 5) | 4 (1fab4a89, 68ee2dd6) |
| F2-10 | assess-fresh2 | 3 | residual | §4.1 last paragraph | If the table-generation refusal sits in `NewPrecedence`, `Constrain` has no tables for `emptyred` or `nullableShift`; then the toy referee cannot load `emptyred` and `TestPrecedenceNullableShiftKnownLimit` cannot stay green, against §2 and §5. The check has to be the recognizer's own. | `forest/precedence_test.go` `loadToy` and `forest/precedence_pinned_test.go` `pinnedToy` both `t.Fatal` on a `NewPrecedence` error. | held; held (r6, residual), "a little wider": three named toys now fail | 4 (1fab4a89, 68ee2dd6) |
| F2-11 | assess-fresh2 | 3 | residual | §4.4 "Every other reason and byte matches V1's" | Nothing shows V1's `scope` byte and the narrowed chart's death byte are always equal. The raiser could not make them differ. "The byte where the chart died" is also undefined when a file is refused with every set alive. | Scratch test `TestZZF3ScopeByte`: 29 of 29 constructed library-operator refusals agree, over three runs with different input lists (only the last list is saved). | held; released (r4), "with one crumb left": the death byte is undefined for a file refused with every set alive; r5: the raiser lists the crumb as held, residual; r6: the raiser says "The minute's record of F2-11 is right": the row is released, and the crumb is a separate residual it still holds | 4 (1fab4a89, 68ee2dd6) |
| F3-1 | assess-fresh3 | 3 | residual | §4.3 | The collapse rule looks forward only. A completion must return only to parents that predicted its context; an item collapsed mid-lineage is indistinguishable from one that got there through a shift its real context forbade. Followed literally the builder loses a derivation or keeps one V1 drops; the forest differential would catch it. The raiser says the rule is implementation and belongs in §6. | Reasoning from `forest/precedence.go` `judgement.seq` and §4.2. Not measured. | held; released (r7, against revision 8's §4.3) | 4 (1fab4a89, 68ee2dd6) |
| F3-2 | assess-fresh3 | 3 | residual | §4.4 "Refusal records" | Build must also keep alternatives apart by item context. A context-blind Build keeps an alternative V1 drops; the differential would catch it. Build's change is larger than the bullet says. | `forest/forest.go` `builder.facts` is keyed (symbol, origin, rule) with no context. Reasoning; not measured. | held; released (r7: §4.3 makes context part of which parent a completion is allowed for; the raiser's release "rests on reading the record as per completing item") | 4 (1fab4a89, 68ee2dd6) |
| F3-3 | assess-fresh3 | 3 | residual | §4.4 "Every other reason and byte matches V1's" | A `scope` byte could move on some input: V1's comes from `resolver.prune`, the rerun's from where the chart died. Only 3 files pin it. | `TestZZF3Narrow`: 3 of 3 equal in-tree, none in perl.git. The raiser did not construct a differing case. | held; released (r4: §4.4 now says that byte "may differ from V1's") | 4 (1fab4a89, 68ee2dd6) |
| F3-4 | assess-fresh3 | 3 | residual | §4.5, §5 | The reference forest now comes from the new recognizer and the new Build, and the set digest is regenerated with precedence on, so nothing pins the mode's sets. A fault common to both modes passes the differential. | `forest/forestdigest_test.go`; `recognize/testdata/set-digest.golden`; `TestParseRefusalMatchesRecognition`. What stays independent: `resolved-forest-digest.golden` and `refused-where.txt`. | held; held (r5, residual); the raiser notes the new referee paragraph adds a check against bison on the referee sentences only | 4 (1fab4a89, 68ee2dd6) |
| F3-5 | assess-fresh3 | 3 | residual | §5 "Feasibility as tests" | Two properties the design rests on are 0 today and are not among the four counters, so a grammar change could break either with no test failing. | `TestZZF3Empty`: `nullable-RHS rules=29 forbidden (rule,t in FOLLOW)=0`. `TestZZF3PartialScan`: 76 states hold two or more core items scanning one terminal, none with differing verdicts; 93 synthetic states counted. | held; held (r6, residual, "narrower"): revision 7 answers part of its first property; the second is still not among the counters; released (r7: the fifth counter is its first property and the sixth its second) | 4 (1fab4a89, 68ee2dd6) |
| F3-6 | assess-fresh3 | 3 | residual | §4.1 | When no terminal in the set can follow, the reduce rule is vacuously true. Under `NoLookahead` that refuses exactly what the FOLLOW filter would. No change to acceptance or the forest; the sentence can be read two ways. | `recognize.go` `builder.complete`: a nil `la` completes every rule. | held | 4 (1fab4a89, 68ee2dd6) |
| F3-7 | assess-fresh3 | 3 | residual | §4.1, §6 | Unlisted scope: moving or injecting the tables (built in `forest`, which imports `recognize`), and an exported switch beside `Options`, whose comment says no option changes acceptance. | `forest/precedence.go` `NewPrecedence`; `forest/forest.go` `parseStream`; `recognize.Options`. | held | 4 (1fab4a89, 68ee2dd6) |
| F3-8 | assess-fresh3 | 3 | residual | header | A reader cannot open the document the decision supersedes; the file is not in the tree. | `ls docs/plans`; `git show origin/docs/chart-parser-v2-design:docs/plans/2026-10-30-chart-parser-v2-design.md`. V1 named the branch. | held | 4 (1fab4a89, 68ee2dd6) |
| F3-9 | assess-fresh3 | 3 | residual | §3, §10 | §3 does not measure what a context in every item costs set interning and the transition cache. V2 could be correct and no faster; §7 already makes that a Stop. | `recognize.go` `cacheKey` and `intern`; rulings, "Relative only". | held | 4 (1fab4a89, 68ee2dd6) |
| A30 | assess-architecture | 4 | residual | §4.4 "The guards" | "Every guard over perly.y runs with precedence on" is followed by a list that opens with `FuzzRecognize`, which is not over perly.y. Its grammars are toys and decoded bytes that declare no levels, so "gains precedence" promises a check that tests nothing unless the grammars change. The raiser says it is the same subject as C8. | README, recogniser section: "each input byte is the singleton set id of one terminal of the toy statement grammar … with one, its bytes decode to a small grammar". §4.1: generating the tables "refuses any other grammar" that fails §3's checks. | held | 5 (de604896, 3a89343d) |
| F3-10 | assess-fresh3 | 4 | residual | §4.4 "The guards" | The qualifier "over perly.y" is new in revision 5 and the list under it was not re-sorted: `FuzzRecognize` does not run over perly.y. Refinement must decide what it "gains" on grammars with no levels, which is C8's open question. Nothing built from the sentence does the wrong thing; the raiser classes it residual "because the false part is the grouping, not a behaviour". | `recognize/fuzz_test.go` ABOUTME: "on stmtGrammar and fuzzed small grammars". `grep -n "perly\|committedTables" recognize/fuzz_test.go` prints nothing. The reference recognizer does run over perly.y: `recognize/agree_test.go` calls `refRun(committedTables(t), …)`. | held | 5 (de604896, 3a89343d) |
| F4-1 | assess-fresh4 | 4 | blocking | §5 "The existing differentials stay … 'Production' now means the recognizer's forest"; §4.5 "`Constrain` … test-only"; §2 anti-goal `precedence-disagree` | The two cannot both hold for the bison referee as it is built. `referee.judge` (non-test code, reached from the ratchet binary via `chart.CountPrecedence`) runs `builtForest` on the lexer's own lattice, then `narrow` on the forest, then `f.Constrain`. Read one way ("production = the recognizer's forest", §4.1's rule on the un-narrowed lattice, then the referee's narrow), 233 sentences disagree with bison, 3 of them accepted where bison refuses, so `precedence-disagree` leaves 0, an anti-goal. Read the other way, `judge` stays on `Constrain`: then `Constrain` is not test-only, and the only key that referees against bison stops refereeing what production runs. The raiser says the design has to say which: (a) the referee narrows the lattice to the sentence's own terminals before recognition, or (b) the referee stays on the no-precedence mode plus `Constrain`, and §4.5 and §2 say so. | Scratch worktree at `de604896`, `go test ./internal/chart -run '^TestZZF4Referee$' -v`: `v1: kept=1296 disagree=0`; `setrule: kept=1296 disagree=233 (with a PLUGIN_ terminal: 233; chart accepts what bison refuses: 3)`; control `exact-first: disagree=71 (PLUGIN_: 71; accepts-refused: 2)`. 588 of the 1,296 kept sentences carry a `PLUGIN_*` terminal. Example: `s0028 "use My::Ops; not qw(a) plus qw(a) ;"`. The "setrule" figure is V1's evaluator with §4.1's reduce rule substituted, not a recognizer; none was built. Code in `fresh4-code/` in the session scratchpad. | held; released (r5, by its raiser, against §4.5's new last paragraph and its own measurement: `lattice: kept=1296 disagree=0 (accepts what bison refuses: 0; refuses what bison accepts: 0)`) | 5 (de604896, 3a89343d) |
| F4-2 | assess-fresh4 | 4 | residual (follows F4-1) | §5 "Forest equality … every referee sentence"; "V1's forest comes from the old recognizer, then Build, then `LibraryOperators`, then `Constrain`" | That is not V1's path for a referee sentence (Build, `narrow`, `Constrain`; no `LibraryOperators`). Under `LibraryOperators` with no operator in scope every `PLUGIN_*` reading is dead, so the 588 `PLUGIN_*` sentences exercise none of the eleven levels, and the corpus does not either. V2's judgement of the `PLUGIN_*` levels would be compared with V1 nowhere but the libops fixtures. The raiser says answering F4-1 with option (a) closes this too. | `forest/resolve.go: libraryOperators` (the `keep := -1` loop); `chart/precedence.go: referee.narrowed`; `feasibility.md` §4, "Tokens carrying any PLUGIN_* terminal: 1,110 → 0". | held; held (r5, residual, "narrower") | 5 (de604896, 3a89343d) |
| F4-3 | assess-fresh4 | 4 | residual | §4.4 "applied to the lattice before recognition" | A narrowed token needs a terminal-set id. Production code never appends to `tab.TerminalSets` today, and the committed tables are shared between recognizers. Whether every narrowed set already exists among the 1,880 was measured neither by the design nor by the raiser. For refinement. | `grep "TerminalSets = append"` over `internal/chart`: test files only. `recognize.go: grammarTables` comment. | held; held (r5, residual, "wider": it now also carries the referee's narrowed sets); new evidence from four reports is under "Round 5" | 5 (de604896, 3a89343d) |
| F4-4 | assess-fresh4 | 4 | residual | §4.1, §4.4 "Both verdicts come from V1's generated tables, read as they are" | The tables live in package `forest`, which imports `recognize`. A package move is implied and unlisted in §6. For refinement. | `forest/precedence.go: Precedence`; `forest/forest.go` imports `recognize`. | held | 5 (de604896, 3a89343d) |
| F4-5 | assess-fresh4 | 4 | residual | §2 "`items-per-token` and `set-reuse` … reported, not gated" | True. But `verdicts.go` takes both from a bare `recognize.Recognize` on the un-narrowed lattice, so under V2 they describe a recognizer run production never makes. Reported only, so nothing is held open. | `cmd/ratchet/verdicts.go: Chart` closure; `compare.go` has no rule for either key. | held | 5 (de604896, 3a89343d) |
| C20 | assess-codebase | 5 | residual (evidence for F4-3) | §4.5 "narrows the lattice to those terminals"; §4.4 "applied to the lattice" | 19 of the single-terminal sets the sentence fixture needs are not among the committed tables' 1,880 sets. The referee, like production's narrowing, needs a way to name a set the tables lack. The raiser says this puts a number on F4-3 and is for refinement. | `TestZZR5Referee`: `singleton sets missing from the tables=19`. | held | 6 (bcd26304, a9c54728) |
| F5-1 | assess-fresh5 | 5 | blocking | §4.1 (last paragraph), §5 ("Feasibility as tests", "The existing differentials stay"), §3 | "Two bison toys fail: `emptyred` and `nullableShift`" is false for toy `empty`. It passes all four counters, yet bison's answer on it needs the reduce verdict of an empty rule (`opt: %prec LOW` before `LB`). The recognizer has no completion event for an empty rule, and V1's shift tables hold no forbidden shift for this toy, so neither of §4.1's two judging sites can drop the derivation. `TestPrecedenceToyReferees/empty` cannot pass through the recognizer as specified: `VAR LB NUM RB` comes out with two derivations where bison has one. For perly.y the same hole is open but empty today, and no pinned counter stops a grammar change from giving an empty rule a non-allow verdict. The raiser's fix: a fifth counter in §3 and §5 (no rule whose symbols are all nullable has a non-allow reduce verdict before a FOLLOW terminal; perly.y reads 0), and the toy exception stated by that rule, not by a count of two. | `fresh5-code/forest_zz_f5toys_test.go.txt`: `toy empty PASSES §3: reduce-conflict 0, parent-conflict 0, context cross-core 0 over-keep 0, forbidden shifts 0`. `forest_zz_f5empty4_test.go.txt`: set 1 holds state 4 with `[s: VAR opt . LB NUM RB]` in the kernel; `reduceAt[opt ::= (empty), LB] = 1`. `forest_zz_f5empty3_test.go.txt`, V1's judgement with empty-rule verdicts never applied: `toy empty "VAR LB NUM RB": chart accept ambiguous (2 derivations), bison accept 4:1-4,1:0-4`; with them applied: `0 of 10 sentences differ`. Bracket for the proposed counter: toy `empty` 1, a built toy 1, the other twelve toys 0, perly.y 0 of 13 empty rules; corpus: 44,328 empty-rule reductions judged, 0 forbidden. No V2 recognizer was built. | held; released (r6, by its raiser, against revision 7); "where it falls short of the letter" is recorded by the raiser as F5-6; r7: "stays released, now on the bar as written"; the raiser corrects its round-6 release as made "against something that landed nearby, not my own bar" | 6 (bcd26304, a9c54728) |
| F5-2 | assess-fresh5 | 5 | residual | §4.1, §5 | The exception is given as "the two toys". A third toy grammar fails counter 4: `TestPrecedenceMarkFirstFilter`'s pinned toy, which is compared through `judgeBoth`, a production-against-reference differential. Stating the exception as a rule covers it. | `forest_zz_f5toys_test.go.txt`: `pinned MarkFirstFilter FAILS §3 ... off-at 1`. The other pinned toys and eleven of the thirteen named toys pass; `emptyred` and `nullableShiftToy` fail. | held; released (r6: §5's exception is now "every toy that fails §3's checks") | 6 (bcd26304, a9c54728) |
| F5-3 | assess-fresh5 | 5 | residual | §4.4 "Refusal reasons" | `refused-where.txt` is also read by `internal/chart/refusals_test.go` through `recognizeSource`, which calls `recognize.Recognize` directly. Under V2 that is precedence-on recognition, while the forest test gets the rerun's byte. A snippet where precedence kills the chart before the grammar does would need two different bytes on one line. The raiser found no such snippet and could not measure it without V2. | `refusals_test.go:refusedWhere`, `port_lex_core_test.go:recognizeSource`. `chart_zz_f5b_test.go.txt`: of 424 port snippets and 856 guacamole files recognition accepts, Parse refuses 0 for precedence. | held | 6 (bcd26304, a9c54728) |
| F5-4 | assess-fresh5 | 5 | residual | §2, §4.5 | The ratchet's stage harness is unnamed scope: `stageInputs` builds the resolve stage's input as Build, `LibraryOperators`, `Constrain`, and `measureStages` times recognition on un-narrowed lattices. Removing the retired key edits `keys` and a `baseline.txt` line inside V2, because `checkBaseline` requires the exact key set. | `cmd/ratchet/speed.go`: `pipelineStages`, `stageInputs`, `chartStages`; `cmd/ratchet/main.go`: `keys`, `check`. | held | 6 (bcd26304, a9c54728) |
| F5-5 | assess-fresh5 | 5 | residual | §4.5 | 19 of the singleton terminal sets the referee's narrowing needs are not in the committed tables, so the referee appends to `tab.TerminalSets` (the recognizer's `extend` allows it). "Builder's matter." | `chart_zz_f5_test.go.txt`: `singleton sets appended to the tables 19`. | held; stood aside (r6): "It is the same fact as C20 and belongs to F4-3, which carries it." | 6 (bcd26304, a9c54728) |
| A31 | assess-architecture | 6 | residual | §5 "Feasibility as tests", fifth counter | The pin is narrower than the hole it covers. A rule whose symbols are all nullable also derives the empty string inside the nullable closure, with no completion event, so a grammar change could give one a non-allow verdict and no test would fail. Nothing is wrong today. | The fix F5-1's raiser proposed was wider: "no rule whose symbols are all nullable has a non-allow reduce verdict before a FOLLOW terminal" (minute, row F5-1). §4.1: "the nullable closure completes it inside the state". The raiser did not measure it and says whether the narrower wording answers F5-1 is its raiser's to say. | held; released (r7, against revision 8) | 7 (28da56fa, 5cb3ab81) |
| C21 | assess-codebase | 6 | residual | §5 "empty rules with a non-allow reduce verdict before any terminal"; §3 | Two gaps, both 0 today. (a) "Empty rule" reads as a rule with an empty right-hand side; the closure also completes a rule whose right-hand side is non-empty and all nullable, over an empty span, with no completion event. perly.y has 16 such rules, and a `%prec` on one would slip past a counter that counts only the 13. (b) §5 calls these "§3's counters", but §3 and `feasibility.md` report neither the fourth nor the fifth; the perly.y value of the fifth is stated nowhere in the committed evidence. | `TestZZR6Empty`: `empty-RHS rules=13 with a non-allow verdict=0; non-empty all-nullable-RHS rules=16 with a non-allow verdict=0`, the same over every terminal and over FOLLOW(LHS) only; toy `empty`: `empty-RHS rules=2 non-allow=1 [opt ::= (empty)]`. `recognize/recognize.go: builder.complete`. `grep -n -i "empty rule" …/feasibility.md` → no hit. | held; released (r7) on (a); on (b) the raiser says "F3-11 and F5-6 carry that; I stand aside on it" | 7 (28da56fa, 5cb3ab81) |
| P19 | assess-pushback | 6 | residual (bears on F5-1's bar, for its raiser) | §5 "Feasibility as tests", fifth counter; §4.1 last paragraph | The recognizer skips every zero-width completion, which is every rule whose right-hand side is all nullable, not only empty rules. The counter as worded watches 13 of perly.y's 29 such rules. The one all-nullable rule that carries a precedence level, `optlistexpr ::= empty`, is among the 16 it does not watch. All 29 read 0 today; a grammar change that gave one of the 16 a verdict would pass the test. | `builder.complete`: `if origin == b.pos && b.r.tab.States[state].Predicted { continue }`. Scratch test `TestZZPB6EmptyRules`: `rules=350; empty right-hand side=13 (non-allow verdict before some terminal: 0); non-empty right-hand side, all nullable=16 (non-allow: 0)`. `TestZZPB6Levels`: `r310 optlistexpr ::= empty: level 1, right-hand side length 1`, the only all-nullable rule with a level. No toy was built where the two wordings give different counts. | held; released (r7, against revision 8) | 7 (28da56fa, 5cb3ab81) |
| F2-12 | assess-fresh2 | 6 | residual | §5, fifth counter | The recognizer skips every complete item of a predicted state at its own origin, not only rules with an empty right-hand side. perly.y has 13 empty rules and 29 rules with an all-nullable right-hand side, so the counter as worded leaves 16 unpinned. All 29 read 0 today. The raiser says it overlaps F3-5 and that whether the fifth counter meets F5-1's bar is its raiser's to say. | `recognize.builder.complete`: `if origin == b.pos && … Predicted { continue }`. Its round-2 `TestZZF2Literal`: "rules with an all-nullable RHS and a non-zero reduce verdict: 0". The minute's F3-5 row; the response's "13 empty rules". No test was re-run this round. | held; released (r7, against revision 8) | 7 (28da56fa, 5cb3ab81) |
| F3-11 | assess-fresh3 | 6 | residual | §4.1 "passes §3's checks"; §5 "§3's counters become tests" | The fifth counter is called one of §3's, but §3 has no row for it and `feasibility.md` does not measure it. Its figure for perly.y, 0 over 13 empty rules, appears only in the author's response file. A reader cannot find the evidence for a check the design now gates on. The figure itself is right. | `grep -n -i "empty rule"` over the design matches only §4.1 and §5. Its round-3 `TestZZF3Empty` printed `empty rules=13 forbidden-or-mixed (rule,t in FOLLOW)=0`. | held; held (r7, residual, "wider": six counters are called §3's and §3's table has four rows) | 7 (28da56fa, 5cb3ab81) |
| F4-6 | assess-fresh4 | 6 | residual | §5 "Feasibility as tests", fifth counter; §4.1 "The recognizer has no completion event for an empty rule" | The counter pins rules with an empty right-hand side only. The nullable closure also completes, with no completion event, any longer rule whose symbols are all nullable when it spans no tokens. perly.y has 16 such rules, all allow-only today. A grammar change that gave one a non-allow verdict would keep a derivation V1 drops with all five counters still at 0. For refinement, or for F5-1's raiser to weigh. | Scratch worktree at `28da56fa`, `TestZZF6Empty`: `perly.y empty rules=13 non-allow=0; non-empty all-nullable rules=16 non-allow=0`; the failing case, `empty empty rules=2 non-allow=1; non-empty all-nullable rules=0 non-allow=0`. Code: `fresh4-code/zz_f6_test.go.txt`. It rests on the tables and `builder.complete`, not on an input. | held; released (r7, against revision 8) | 7 (28da56fa, 5cb3ab81) |
| F5-6 | assess-fresh5 | 6 | residual | §3 table, §5 "Feasibility as tests" | §3's table and `feasibility.md` carry no row for the fifth counter; its perly.y value and its bracket exist only in this session's scratch code. The counter is also narrower than the class with no completion event: it leaves out all-nullable non-empty rules. "A sharper sentence and one table row; for refinement." | `forest_zz_f6_test.go.txt`: `perly.y: 13 empty rules; non-allow before any terminal 0 ...; all-nullable non-empty rules 16, their non-allow pairs 0`; `toy empty: non-allow before any terminal 1`. `forest_zz_f6b_test.go.txt`: a unit-empty toy (`opt ::= e %prec LOW`, `e ::=`) passes all counters, reads 0 on both wordings, and V1 keeps 2 derivations with or without the unit rule's verdict. | held; held (r7, residual, "narrower": the class half is answered; the missing §3 rows remain) | 7 (28da56fa, 5cb3ab81) |
| F6-C1 | assess-fresh6 | 6 | blocking | §4.1 last paragraph; §5 "Feasibility as tests", fifth counter | "The recognizer has no completion event for an empty rule" holds for every rule whose right-hand side is all nullable, and the counter does not count those. A grammar passes all five counters and the recognizer accepts what V1 refuses. On perly.y the counter is "close to vacuous": its 13 empty rules carry no level, and the one leveled rule that derives the empty string, r310 `optlistexpr ::= empty %prec PREC_LOW`, is in the class the counter skips. So §4.1's "the recognizer judges precedence only for a grammar that passes §3's checks" does not hold as the checks are worded. | Toy `s: X a T Z \| X b; a: n %prec LOW; b: n T R; n: ε \| Q`, levels `nonassoc LOW`, `left T`. `NewPrecedence` accepts it. Counters: M1=0, M2=0, empty-RHS non-allow=0, forbidden shifts=0; yet `reduceAt[a ::= n, T]` = 1. `X T Z`: the unconstrained chart accepts, `Constrain` and `ConstrainReference` refuse. `a: n .` sits only in predicted state 2, which `complete` skips. Bracket: `X T R` accepted by all; perly.y wide count 0 of 16; toy `empty` trips the narrow counter. `fresh6-code/zz_f6d_test.go.txt`, `zz_f6e_test.go.txt`, `zz_f6_test.go.txt`. Argued from `complete` and the automaton's states; no V2 recognizer exists; the toy was not checked against bison. | held; released (r7, by its raiser, against revision 8): its toy `nulwrap` now trips the fifth counter while passing the other five | 7 (28da56fa, 5cb3ab81) |
| F6-C2 | assess-fresh6 | 6 | blocking | §4.1 "Shift side"; §5 counters 3 and 4 | An item is a state, and one transition carries every core item of the state that has the terminal next, including dot-0 items V1 never checks. No counter pins that they all get the same verdict. A grammar passes all five counters (and the committed M3 static check) while one scan must be both refused and allowed: refusing loses a sentence V1 accepts, allowing keeps one V1 refuses. On perly.y the count is 0, "so this is the guard, not today's behaviour". | Toy `s: P c \| P a \| P d X W; c: X Q; a: n X R; d: n %prec HIGH; n: ε \| M`, levels `left X`, `left HIGH`. Generation accepts. Counters: M1=0, M2=0, empty=0, forbidden shifts 1, all "terminal is `at`'s next symbol"; `ZZM3Static`: needCopiesCross=0, overKeep=0. Predicted state 2 under top `{s: P . a}` scans X for `[c: . X Q]` (not forbidden) and `[a: n . X R]` (forbidden). V1: `P X Q` accept, `P X R` refuse. Bracket: perly.y and all 13 toys count 0, and perly.y has 0 forbidden scans by an item of a predicted state. `fresh6-code/zz_f6g_test.go.txt`, `zz_f6f_test.go.txt`. | held; released (r7, by its raiser, against revision 8): its toy `scancollide` is caught by the sixth counter; "the wording leaves one thing open, raised as F6-R3" | 7 (28da56fa, 5cb3ab81) |
| F6-C3 | assess-fresh6 | 6 | blocking | §4.3 first sentence; §4.1 "Where each side is judged"; §4.4 refusal records | §4.3's collapse rule, followed literally, discards the context a completion needs to find the parent it was predicted under: no verdict "reachable from" the item uses it at that point, so the rule drops it, and §4.1 names two judgement sites, neither of which says a completion advances only the parents of its own context. The first corpus file with this shape differs from V1's forest, which is a §7 Stop; `context-split.md` counts 145 such nodes. The raiser says the rule is implementation detail and the answer is to move it to §6, or to say the context is kept from prediction to completion. | `foo -$x * $y;` on the committed tables. Token 1 is `{ADDOP, PERLY_MINUS}`, so two predictors meet at set 2. V1 on node `termbinop` tokens 2-5 before `;`: under `{termbinop: term ADDOP . term}` it passes r203 `term MULOP term`; under `{termunop: PERLY_MINUS . term}` it passes nothing. The only context-dependent verdict is the scan of MULOP at token 3. With `none` after it, the item completes into both parents waiting at set 2, including unary minus. Bracket: `foo -$x ** $y;` and `foo -$x;` have no node whose tops differ. `fresh6-code/zz_f6h_test.go.txt`. | held; released (r7, by its raiser, against revision 8's §4.3); the remaining wording is raised as F6-R4 | 7 (28da56fa, 5cb3ab81) |
| F6-R1 | assess-fresh6 | 6 | residual | §4.4 "Library operators, before recognition"; §5 forest equality inputs | When a library operator is kept, the narrowed set is not in the tables. Production must mint sets at parse time, which only tests do today, or the tables grow. No corpus token keeps a `PLUGIN_*` reading after narrowing, so no corpus differential exercises that path. | `fresh6-code/zz_f6c_test.go.txt`: "sets=1880; holding a PLUGIN_*=62; narrowed to no PLUGIN_*: empty=1, absent=0; narrowed to one: 682, absent=682". `grep "TerminalSets = append"` finds test files only. | held; stood aside (r7): "Same fact as F4-3 and P4, which carry it" | 7 (28da56fa, 5cb3ab81) |
| F6-R2 | assess-fresh6 | 6 | residual | §4.4 "Leo"; §5 "Leo" | "A Leo chain must not skip a completion that precedence refuses" is not an edge case: 51 of the 65 Leo-eligible rules have a non-allow reduce verdict before some FOLLOW terminal, so a memo's top depends on the lookahead at each use. A cost risk for the §7 exit, already covered by its stop. The promised test is easy to build on perly.y (`$a = $b + $c`). | `fresh6-code/zz_f6b_test.go.txt`: "Leo-eligible rules=65; with a non-allow reduce verdict before a FOLLOW terminal=51". | held | 7 (28da56fa, 5cb3ab81) |
| C22 | assess-codebase | 7 | blocking (wording; one phrase) | §4.3, second sentence, against §4.2 and the example | "A completion advances only the parent items whose context predicted it" states the rule in a way the example beside it contradicts. §4.2 defines an item's context as the spine top it carries. In the example, the parent that must advance does not carry the completion's context; it is that context. Read literally (advance the parents whose context equals the completion's), neither parent advances and nothing parses. The raiser's fix: "A completion advances only the parent items that predicted it", leaving the two cases to §4.2. It says the rule itself "is right and needed", that the builder would notice a literal misreading at the first test, and that it calls the row blocking "because it is a statement of mechanism that does the wrong thing when followed as written". | `TestZZR7Contexts` on V1's evaluator: `shift of ADDOP at token 1 after [termbinop: term . ADDOP term] (rule begun at 0): the rule's own context = {none}`; `shift of MULOP at token 3 after [termbinop: term . MULOP term] (rule begun at 2): the rule's own context = {termbinop: term ADDOP . term}; forbidden=false`; the same under `{termunop: PERLY_MINUS . term}`: `forbidden=true`. `forest/precedence.go: judgement.seq`: `rtop = at` when `m > 1`. `builder.complete` today advances every item of the origin set that awaits the completed symbol, with no filter. | held; released (r8, by its raiser, against revision 9's phrase "that predicted it") | 8 (2d9611db, d893c1f0) |
| P20 | assess-pushback | 7 | residual | §4.3, new first paragraph | "A completion advances only the parent items whose context predicted it" can be read as "whose context equals the completion's". By §4.2 that is the wrong test for every child that is not a left corner. The example states the right outcome and the forest differential would catch the wrong reading, so nothing ships wrong; the sentence is sharper as "the parent items that gave it its context (§4.2)". It is also mechanism, which F6-C3's raiser proposed moving to §6. | §4.2: "A child predicted at a later position takes the predicting item as its spine top." `feasibility.md` §3.3: "Each predicted item's context is a function of the predicting kernel *state* K, not of K's own context." Reasoned from the text. | held; released (r8, against revision 9) | 8 (2d9611db, d893c1f0) |
| A32 | assess-architecture | 7 | residual | §4.3 against §4.2 | Read alone, "An item carries `none` only where its context neither changes a verdict reachable from it nor selects among the parents it can complete into" seems to deny `none` to an item whose top forbids no shift but which shares an origin with a predictor under a real top. A builder who read it that way would carry a non-canonical top, which is outside the ruling. §4.2 prevents that reading; the two sections say it in two places. The raiser says it is the same subject as P12. | §4.2: "Any spine top that forbids no shift collapses to `none` … So a context is either one of V1's 46 spine tops or `none`." Not measured. | held | 8 (2d9611db, d893c1f0) |
| F3-12 | assess-fresh3 | 7 | residual | §4.3, first paragraph | A kernel parent item carries its own rule's context, which is not the context it gives the children it predicts. Read as "the parent's carried context equals the completion's", the sentence picks the wrong parents. The sentence before it, §4.2 and the minus example point to the right reading, and the forest differential would catch the wrong one. A sharper sentence: "only the parents that predicted it under that context". | §4.2: "A child predicted at a later position takes the predicting item as its spine top." `feasibility.md` §3.3. | held; released (r8, against revision 9) | 8 (2d9611db, d893c1f0) |
| F6-R3 | assess-fresh6 | 7 | residual | §5, sixth counter | V1 gives no shift verdict to a dot-0 item. A counter that compares only items V1 checks does not see the toy the counter was added for. The sentence should say an unchecked item counts as allowed; reading V1's table for a dot-0 item returns "not forbidden", so the literal reading is the right one. | `fresh6-code/zz_f6_test.go.txt` (dot > 0 only) on `scancollide`: `sameState/top/terminal two-at disagree=0`. `zz_f6f_test.go.txt` (dot-0 items included): `scans where core items disagree=1`. perly.y: 0 both ways. | held | 8 (2d9611db, d893c1f0) |
| F6-R4 | assess-fresh6 | 7 | residual | §4.3, second sentence and heading | For a parent at a later rule position, the child's context is the parent item itself, not the parent's own context. Read word for word, "whose context" could be taken as the parents' own contexts, which do not tell them apart. The first sentence and the example fix the meaning, so nothing built from it goes wrong. The rule is also a third judgement site filed under "Collapsing contexts" rather than §4.1. | §4.2; the F6-C3 probe: the two tops are the parent items `termbinop: term ADDOP . term` and `termunop: PERLY_MINUS . term`. | held; held in part (r8, residual): the wording half is released; the placement half stands (the pairing rule is a third judgement site filed under "Collapsing contexts" and not in §4.1) | 8 (2d9611db, d893c1f0) |
| F7-P1 | assess-fresh7 | 7 | residual | §4.4 "The guards" | `FuzzRecognize` is not a guard over perly.y, and its grammars declare no levels, so "gains precedence" changes nothing there. The reference recognizer does run over perly.y, so that half is real. The raiser says refinement should say what the fuzzer does, "or perigrin may want to drop it from the list". | `recognize/fuzz_test.go` ABOUTME: "on stmtGrammar and fuzzed small grammars". `grep -n "Level\|%prec\|Prec" recognize/fuzz_test.go recognize/reference_test.go` prints nothing. | held; stood aside (r8): "The same fact as A30 and F3-10, both held residual and tied to C8. Their raisers hold it; a third row adds nothing." | 8 (2d9611db, d893c1f0) |
| F7-C1 | assess-fresh7 | 7 | residual | §4.1, §6 | The tables are generated in package `forest`, and `forest` imports `recognize`. Moving the generation, or injecting the tables, is work §6 does not list. Implementation, for refinement. | `forest/forest.go:19` imports `internal/chart/recognize`; `grep -rn 'chart/forest"' recognize/*.go` prints nothing. | held; stood aside (r8): "The same fact as A12, F2-8, F3-7 and F4-4 (tables built in `forest`, which imports `recognize`). Held residual by four raisers." | 8 (2d9611db, d893c1f0) |
| F7-A1 | assess-fresh7 | 7 | residual | §5 "Feasibility as tests" | "Context conflicts" does not say which check it is. §4.2's rule needs two things: no predictor with two canonical tops, and no rule begun under one predictor with two real tops that disagree. A test that checks only §3's per-state measure would pass a grammar change that breaks §4.2. | The raiser's count of the first: 0 predictors, 0 canonical start items (`TestZZF7Proto`, "static:" line). `feasibility.md` §3.2 "Per predictor K" measured the second. | held | 8 (2d9611db, d893c1f0) |
| F7-C2 | assess-fresh7 | 7 | residual | §4.5 | Today's `narrow` also drops every derivation that uses a slipper `;`. Narrowing the lattice cannot do that, because the slipper runs inside recognition. Followed literally, the V2 referee keeps those derivations. No verdict moves on the current fixture, so `precedence-disagree` stays 0 today. | `precedence.go` `narrow`: `nd.To == nd.From+1`. Bracket: `TestPrecedenceReferee` passes unmodified, and passes with zero-width leaves kept alive. | held; stood aside (r8): "The minute's round-5 evidence already measures it: assess-codebase's lattice-narrowed referee equals today's on all 1,296 sentences, with the slipper inserting a `;` in every one. My own bracket in round 7 agrees." | 8 (2d9611db, d893c1f0) |
| F7-C3 | assess-fresh7 | 7 | residual | §4.1, §5 | At a node before `}` or at the end of input the recognizer's set also holds `;`, and V1 judges the derivation through the slipper before `;`. The two agree only because `;`, `}` and EOF have no level. That holds on perly.y and no §5 counter states it. | `recognize.go` `extend`: `if closes && g.semi >= 0 { b.set(g.semi) }`; `buildGrammar`: `g.eofBits.set(g.semi)`. `feasibility.md` §4 added `;` by hand. | held | 8 (2d9611db, d893c1f0) |
| F7-A2 | assess-fresh7 | 8 | residual | §4.3, second sentence | Read with no context, "A completion advances only the parent items that predicted it" lets every item awaiting the symbol in the origin set count as having "predicted it", which is plain Earley and keeps what V1 drops. The sentence before it and the example rule that reading out, so nothing ships wrong; "under its context (§4.2)" would close it. The raiser puts it in the same family as A32, P20, F3-12 and F6-R4. | `F7_NOMATCH=1`: 26 over-keeps, first `s0187 "foo qw(a) .. qw(a) ;"` keeps `term: UNIOPSUB term [0,4)`. With the rule as the example means it: 0. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-P1 | assess-fresh8 | 8 | residual | §4.5 | Lattice narrowing cannot remove the slipper's zero-width `;`, which today's `narrow` removes. Measured harmless: verdicts and groupings do not move. Refinement should know the referee's forests will carry those leaves. | Scratch worktree, `narrow` changed to keep zero-width leaves: `TestPrecedenceReferee` ok before and after. 1,296 of 1,299 kept sentences have a zero-width leaf; 0 become ambiguous or cut. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-C1 | assess-fresh8 | 8 | residual | §3 | "in 14 constructed inputs": the 14 inputs are in neither `feasibility.md` nor `context-split.md`, so the claim cannot be re-run from the committed evidence. | `grep -rn "14 constructed" docs/assessments/chart-precedence-v2/` finds nothing. The raiser's own 44 inputs agree with the claim. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-C2 | assess-fresh8 | 8 | residual | §4.1, §5 | "perly.y passes" all six counters: counters 5 and 6 have no committed measurement. The claim is true, but the evidence lives only in reports until §5's tests exist. | `rules=350 allNullable=29 rulesWithNonAllow=71 counter5=0 tops=46 triples=674 notOwnNext=0 counter6=0 (scans judged 20746)`. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-A1 | assess-fresh8 | 8 | residual | header | The governing design is not a file in the tree. It lives on branch `origin/docs/chart-parser-v2-design`. V1 names the branch; V2 does not. | `git ls-files \| grep chart-parser` lists only the 10-29 design. `git show origin/docs/chart-parser-v2-design:docs/plans/2026-10-30-chart-parser-v2-design.md` works. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-P2 | assess-fresh8 | 8 | residual | §4.4 | `FuzzRecognize` runs on `toyStmt` and fuzzed toy grammars, which declare no precedence, so "gain precedence" has no meaning there until refinement gives it one. The reference recognizer does run on a perly.y corpus sample, so it becomes a second implementation of the judgement: scope the design names but does not size. | `recognize/fuzz_test.go` `FuzzRecognize`, `fuzzGrammar`; `recognize/agree_test.go` ABOUTME. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-P3 | assess-fresh8 | 8 | residual | §2, §4.1, §7 | An input that reaches set 1626 after `indirob ::= scalar` keeps a completion V1 drops. During the milestone that is a Stop. After delivery nothing reports it, so production would over-keep silently. | §4.1's own rule ("the completion is kept"); `feasibility.md` §4. | held | 9 (bc9b4abd, 6d4dacab) |
| F8-P4 | assess-fresh8 | 8 | residual | §4.5, §5 | §4.5 gives the old pipeline as "the unpruned recognizer, then Build, then `Constrain`"; §5 gives it as recognizer, Build, `LibraryOperators`, `Constrain`. The two lists differ by one stage, and §5's is the right one. "A sharper sentence, nothing more." | The two paragraphs. | held | 9 (bc9b4abd, 6d4dacab) |

After round 1: 44 rows (A1–A19, C1–C13, P1–P12); by the raisers' own declarations, 17 blocking (A1–A10, C1–C4, P1–P3) and 27 residual, all held.

After round 2: 66 rows. Round 2 added 22 (A20–A25, C14–C17, P13–P16, F2-1 to F2-8): 8 declared blocking (A20, C14, C15, C16, P13, P14, F2-1, F2-2) and 14 residual. Statuses are in the ledger; the round-2 tally is under "Round 2".

After round 3: 86 rows. Round 3 added 20 (A26–A29, C18, C19, P17, P18, F2-9 to F2-11, F3-1 to F3-9): 4 declared blocking (A26, A27, C18, P17) and 16 residual. One residual row, F2-6, was promoted to blocking by its raiser. The round-3 tally is under "Round 3".

After round 4: 93 rows. Round 4 added 7 (A30, F3-10, F4-1 to F4-5): 1 declared blocking (F4-1) and 6 residual. The round-4 tally is under "Round 4".

After round 5: 99 rows. Round 5 added 6 (C20, F5-1 to F5-5): 1 declared blocking (F5-1) and 5 residual. The round-5 tally is under "Round 5".

After round 6, with seven of eight participants reporting: 106 rows. Round 6 added 7 (A31, C21, P19, F2-12, F3-11, F4-6, F5-6), all residual. The round-6 tally is under "Round 6".

After assess-fresh6's report, which arrived after the ruling: 111 rows. It added 5 (F6-C1, F6-C2, F6-C3, F6-R1, F6-R2): 3 declared blocking and 2 residual. See "Round 6, continued".

After round 7: 122 rows. Round 7 added 11 (C22, P20, A32, F3-12, F6-R3, F6-R4, F7-P1, F7-C1, F7-A1, F7-C2, F7-C3): 1 declared blocking (C22) and 10 residual. The round-7 tally is under "Round 7".

After round 8: 130 rows. Round 8 added 8 (F7-A2, F8-P1 to F8-P4, F8-C1, F8-C2, F8-A1), all residual. The final tally is under "Round 8".

## Round 1

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

Named by the coordinator:

- **A5 / C2 / P2:** the recognize-stage stop at 1.0.
- **A4 / C3 / P7:** the direction of `set-reuse` (C3 and P7 also name the changed population; A4 and P7 are declared at different levels, blocking and residual).
- **A7 / C4 / P3:** counter 4 and set 1626.
- **C1 / P1:** positional refusal reasons against V1's stage order.
- **A8 / C7:** the completion verdict where there is no single lookahead terminal.
- **C9 / P6:** `resolved-forest-digest.golden` held, not regenerated (declared residual by both; P6 marks it for perigrin).

Found while drafting:

- **A8 / C7 / P4:** callers that recognize a raw `lex.Linear` lattice with no scope and so no narrowing.
- **A8 / C10(b) / P11:** `Chart.Expected` under precedence.
- **A8 / C10(c):** `Options.NoLookahead`.
- **A12 / C10(e) / P5(a):** V1's tables live in `forest`, which imports `recognize`.
- **C10(f) / P5(c):** the referee needs the old Build as well as the old recognizer.
- **C8 / P5(d), P5(e):** the reference recogniser and fuzzer gaining precedence; "no off-switch" beside a path that prunes nothing.
- **A6 / P8(a):** the two references for the 27 answer keys (declared blocking by A, residual by P).
- **A10 / P8(b):** what the exit does not require (A10: a magnitude; P8(b): that the pass left production).
- **A18 / A5, A4:** A18 says the two anti-goals behind A5 and A4 are the orchestrator's, not rulings.
- **A2 / assess-codebase's architecture axis:** the codebase report's axis text (not a row) also says "README rule 12" is the spec's and that rule 2 is pinned by `TestItemIsStateDistanceR2` and `TestItemIsStateDistanceF2`; the latter is row C10(d).
- **A19 / C10(a):** README's "stay recognition's" and `TestRecognitionKeysUnchangedByPipeline`.
- **C12 / C1 / P1:** the corpus holds no precedence refusal, and `refused-where.txt` no `precedence` row.
- **C5 / C6 / a P discard:** V1's union shortcut. assess-pushback discarded it as the codebase lens's to confirm; C5 and C6 are that lens's rows.
- **C13 / P10:** each covers one of §3's three stated gaps (Leo; perl.git `t/`). Different gaps, same sentence of the subject.
- **A15 / C6:** both say mechanism in §4.4 belongs outside the decision.
- **C11 / C2:** the raiser links them itself.

Reports disagree on one count, recorded without resolving it: `speed.go` call sites are 4 in C7 and 3 in P4.

### Per-participant record

**assess-architecture.** Recommendation: modify. It read the blob and ran no test, ratchet or corpus run; every row is from reading, `grep` and `diff`. It raised A1–A19: ten declared blocking (A1–A10, with A10 marked for perigrin) and nine residual (A11–A19). Its stated grounds: the subject reverses rulings and rules of the governing design without naming them (A1, A2, A3); part of its bar cannot be read or reads the wrong way (A4, A5, A6, A7); §4.1 is undefined for a recogniser mode §4.4 says must run with precedence on (A8). It says the direction is perigrin's by ruling, that it does not reopen it, and that every blocking row "is fixed by a sentence, a ruling or a deletion". It declared blocks, not stand-asides.

**assess-codebase.** Recommendation: modify. It worked in a detached scratch worktree at 488e1c7e, re-ran the committed feasibility code and reproduced §3's corpus numbers and `context-split.md`'s. It raised C1–C13: four declared blocking (C1–C4) and nine residual (C5–C13, with C8 marked a ruling conflict for perigrin). It says "the direction survives the code" and that §4's mechanism can produce V1's forest in principle under two conditions §4 does not state (the union shortcut, C6; the narrowing order, C1). It did not run perl.git. It declared blocks, not stand-asides.

**assess-pushback.** Recommendation: modify. It ran `paad:pushback` v2.1.0 non-interactively and worked in a detached scratch worktree at 488e1c7e. It raised P1–P12: three declared blocking (P1–P3) and nine residual (P4–P12, with P6 marked a conflict for perigrin and P10 marked evidence for §3), and reported five discards. It ran the committed feasibility code over perl.git `t/` (P10). It judged the subject one decision, not several (P11). It says the direction "survives everything I measured". It declared blocks, not stand-asides.

No participant recommended reject or accept. No participant stood aside on any row.

### Outcome

Three modify, 17 rows declared blocking, every row held. No fixed point. **Another round is owed**, on a revision that answers the held rows or on the raisers' positions after reading each other's.

### Questions for perigrin

These rows are marked for perigrin by their raisers, or turn on a ruling. They are put as questions; this minute gives no answers. perigrin's answers are recorded under "Round 2", "Rulings after round 1".

1. **A1.** Did you rule that rule 2 is lifted for the precedence value on items? §4.2 says so and `rulings.md` has no such entry. If you did, what were the words, so they can be recorded?
2. **A3.** The subject reverses the governing design's §8 ("precedence leaves recognition (perigrin, 2026-10-30)"), rule 6, part of rule 12 and the §10 kill "Facts in items", and V1's "What this does not amend", without naming them. Do you reverse them, and should the subject name each one it reverses?
3. **A9.** Is `speed-stage-precedence-aggregate-longlived` retired, kept to time V1's pass as a reference, or something else? §6 leaves it to the builder; the README's output contract says no key is retired except by your ruling.
4. **A10.** What is V2's exit? §7 offers the relative bar only (after-median below before-minimum) as a proposal. V1's rulings say the V2 decision sets its own target. Is there a number, and what is it?
5. **C8.** "Regenerate everything" was ruled for guards whose grammars declare no precedence levels. Do the reference recogniser and the fuzzer get grammars with levels, given that `NewPrecedence` refuses a grammar with a state-dependent key? And is a recogniser over a grammar with no levels, which prunes nothing, compatible with "no off-switch"?
6. **P6** (with C9). Does "Regenerate everything" cover `resolved-forest-digest.golden`, or is that golden held unchanged as a check on the goal? Both forest goldens carry "Never regenerate it to make a change pass"; does the ruling replace that contract for `forest-digest.golden`?

## Round 2

Round 2 assessed revision 3 (a8808dd2, blob 71d5f7bb). The three round-1 participants were resumed, and assess-fresh2 read the subject fresh.

### Rulings after round 1

perigrin made four rulings after round 1. `rulings.md` records them verbatim under "2026-11-25, after assess round 1" and says the questions came from rows A1, A3, A9, A10, C1, P1, C9 and P6. In the order of this minute's round-1 questions:

- **Questions 1 and 2 (A1, A3):** "Supersede, bounded (Recommended)".
- **Question 3 (A9) and question 6 (P6, with C9):** "Hold golden, retire key (Recommended)".
- **Question 4 (A10):** "Relative only (Recommended)".
- **C1 and P1** (not among the six questions; put to perigrin by the coordinator): "Accept/refuse only (Recommended)".
- **Question 5 (C8):** `rulings.md` records no answer. Its raiser holds C8, still marked for perigrin.

### Drafter: growth check

Made by the drafter before the round went ahead, and reported at the top of `r2-pushback.md`. It holds no view on the subject. It compared `git diff 488e1c7e a8808dd2 -- docs/plans` (one file, 192 lines to 254) with the blocking rows `response-r1.md` claims and with the four rulings.

- **Hunks that trace to a blocking row or a ruling:** the header's "Governed by" block (A1, A2, A3; "Supersede, bounded"); §2's anti-goal changes (A4, C3; C1, P1; A5, C2, P2) and the retired key (A9); §3's fourth table row and the sentence after it (A7, C4, P3); §4.1's lookahead-set rule and its three sub-bullets (A7, C4, P3, A8); §4.2's last line (A1); §4.4's refusal-record citation (A2), the positional-reasons bullet (C1, P1), the `Expected` clause (A8) and the two goldens ("Hold golden"; rows C9 and P6, both declared residual); §5's "accept or refuse", feasibility tests and keys (C1, P1; A7, C4, P3; A6; A4, C3); §6's deleted key bullet (A9); §7's exit and stops (A10; A3; A4, C3; A5, C2, P2; A7, C4, P3; C1, P1).
- **Hunks that trace only to a residual row** (the response says residuals are not answered in the subject): §6's deletion of "a copy, or a build-tagged path" (P5(e), C8); and the words "a recognizer given an un-narrowed lattice" in §4.1, which is the subject of residual C7 and P4 as well as blocking A8.
- **Hunks that answer nothing** (wording or layout, no content): the "Follows" line; §1's list and paragraphs; §2's goals split, non-goals sub-list and regrouped key names; §3's round-5 table, ceiling list and "Contexts rarely meet" lead-in; §4.1's "come from"; §4.2's re-split paragraphs; §4.3's "of nodes"; §4.4's "Library operators" sub-bullets; §4.5's second paragraph; §5's reworded "existing differentials" and re-laid lists; §8; §9's chain-export bullet; §10's sub-lists. Seventeen hunks.
- **Growth:** no new section, mechanism or criterion appears that no row or ruling asked for. The file grew by 62 lines, about two thirds of the changed lines being the layout hunks. One sentence imports a participant's evidence into §3 ("in 14 constructed inputs", from P3's row); C4's row reported 26 inputs for the same check.

### Status changes in round 2

Only the raiser changed a row's status.

| raiser | released | held in part | held | stood aside |
|---|---|---|---|---|
| assess-architecture | A1, A2, A3, A4, A5, A6, A7, A9, A10, A17, A18 | A8 (blocking) | A11, A12, A13, A14, A15, A16, A19 (residual) | none |
| assess-codebase | C1, C2, C3, C4, C9 | none | C5, C6, C7, C8, C10, C11, C12, C13 (residual) | none |
| assess-pushback, as raiser | P1, P2, P3, P6, P7 | P5 and P8 (residual) | P4, P9, P10, P11, P12 (residual) | none |

Of the 17 rows declared blocking in round 1, 16 are released and 1 (A8) is held in part. Of the 27 declared residual, 5 are released (A17, A18, C9, P6, P7), 2 are held in part and 20 are held. No row was promoted: each resumed participant wrote "none" under promotions.

Three releases came with a new row on the text that answered the old one: C1 with C16, C4 with C14, P1 with P14. assess-architecture released A3 and raised residual A21 for the one item of its list the header does not name.

Round 2 added 22 rows, all held: 8 declared blocking (A20, C14, C15, C16, P13, P14, F2-1, F2-2) and 14 residual (A21–A25, C17, P15, P16, F2-3 to F2-8).

**After round 2: 66 rows; 21 released; 9 held rows are declared blocking** (A8 in part, A20, C14, C15, C16, P13, P14, F2-1, F2-2); 36 held rows are residual (2 of them held in part).

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

Named by the coordinator:

- **C14 / F2-1:** §4.1's "every terminal in the set" reduce rule, read over terminals that cannot follow the rule's LHS. Both declared blocking; both propose the same restriction to terminals that can follow.
- **C15 / P13 / F2-3:** `NoLookahead` as "reduce side not judged" changes acceptance. C15 and P13 are declared blocking; F2-3 is declared residual.
- **C16 / P14 / F2-5:** the positional rule for refusal reasons. C16 and P14 are declared blocking; F2-5 is declared residual.

Found while drafting:

- **A20 / P14(d) / C16 / F2-5:** A20 is the fourth row on the positional rule. A20 and P14(d) name the same point: a completion refused in one set, the chart dying in a later one.
- **A8 (held part) / C15(b) / P13(2) / F2-7 / C7 / P4:** the un-narrowed lattice in §4.1's no-lookahead bullet, and the callers that recognize without narrowing. A8, C15 and P13 are declared blocking; F2-7, C7 and P4 residual.
- **C16(b) / F2-5(a):** a file refused with every set alive. **C16(c) / P14(c) / F2-5(b):** the `scope` clause; P14 and F2-5 both count 1 of the 62 `PLUGIN_*`-bearing sets as holding only `PLUGIN_*` terminals. **C16(a) / P14(a):** `lex-unmapped` missing from the rule. **C16(c) / F2-5(b):** both note the first ruling's "keeps its reason `scope`" still stands in `rulings.md`.
- **A21 / F2-4:** the header's supersession list omits rule 12's "Scope … collapses what remains, after the chart". F2-4 names two further clauses (§2 and §6's data flow).
- **A22 / P16:** retiring the stage key edits `baseline.txt` inside the milestone, beside the non-goal on retaking it.
- **C17 / the growth check:** §3's "14 constructed inputs". The counts in the reports are 14 (P3), 26 (C4) and 38 more (assess-fresh2's codebase axis).
- **F2-8 / A12 / C10(e) / P5(a):** V1's tables live in `forest`, which imports `recognize`.
- **F2-2 / F2-6 / C8:** grammars other than perly.y under V2's recognizer: the bison toys (F2-2), the shift-shape assumption as a test (F2-6), the guards' grammars with no levels (C8).
- **A13 / C15 / P13:** the reduce side unjudged at lookahead 0. A13 (residual, held) says revision 3 sharpens it for interleaved mode.
- **A25 / §3's own limit:** branched lattices against "the only set where this can happen is set 1626".
- **P15 / A7 (released):** the set-1626 stop. A7's release says the two stops agree; P15 says the stop has no observer.
- **A24 / A2 (released):** whether rule 2 itself is superseded.

Classifications that differ across raisers for the same defect, recorded without resolving them: `NoLookahead` (C15 and P13 blocking, F2-3 residual); refusal reasons (A20, C16 and P14 blocking, F2-5 residual).

### Per-participant record

**assess-architecture.** Recommendation: modify. Blob 71d5f7bb, checked. It says it read revision 3 whole and ran its checks before reading the author's response, and did not read the diff line by line. It ran no test. It released A1–A7, A9, A10, A17 and A18, held A8 in part, and let A11–A16 and A19 stand as residual. It raised A20 (blocking) and A21–A25 (residual). Its stated grounds for modify: A8's held part and A20, "each … a one-sentence fix". Promotions: none. It declared a block (A8 in part, A20), not a stand-aside.

**assess-codebase.** Recommendation: modify. Blob 71d5f7bb, checked. It re-ran the committed feasibility code in a scratch worktree at a8808dd2 and says it reproduces round 1 exactly. It released C1–C4 and C9 and held C5–C8 and C10–C13 as residual. It raised C14, C15 and C16 (blocking) and C17 (residual). Its stated grounds: "Three sentences that revision 3 added to answer them are wrong against the code", each with a one-sentence fix. It says C14 and C15 are shown with V1's tables and evaluator, not with V2 running. Promotions: none. It declared blocks, not stand-asides.

**assess-pushback, as raiser.** Recommendation: modify. Blob 71d5f7bb, checked. It re-ran its round-1 measurements in a scratch worktree at a8808dd2. It released P1, P2, P3, P6 and P7, held P5 and P8 in part, and held P4 and P9–P12 as residual. It raised P13 and P14 (blocking) and P15 and P16 (residual), and reported two discards. Promotions: none. It declared blocks, not stand-asides. It is not an independent participant this round.

**assess-fresh2.** Recommendation: modify. Blob 71d5f7bb, checked. A fresh read. It raised F2-1 and F2-2 (blocking) and F2-3 to F2-8 (residual). It says the direction "survives the code and the decisions in force", that it "found no reason the design cannot work on perly.y", and that each blocker is a one-sentence fix needing no ruling. It ran scratch tests over the committed tables and the toy fixtures, kept in `f2-scratch/` in the session scratchpad; it made no corpus run. It says F2-2's V2 behaviour is reasoned, not run. It declared blocks, not stand-asides.

No participant recommended reject or accept. No participant stood aside on any row.

### Outcome

Four modify. Nine held rows are declared blocking: A8 (in part), A20, C14, C15, C16, P13, P14, F2-1 and F2-2. No fixed point. **Another round is owed.**

### Rows that turn on a ruling

Put as questions; this minute gives no answers.

1. **C8** (held, for perigrin; unanswered from round 1). Do the reference recogniser and the fuzzer get grammars with precedence levels? Is a recogniser over a grammar with no levels compatible with "no off-switch"? F2-2 and F2-8 bear on the same question for the toys and for a recogniser built without a precedence.
2. **A20.** Its raiser offers two repairs: a sentence defining the reason by cause, "or a ruling that `grammar` is acceptable here". Is `grammar` acceptable for a file whose only cause of death is a precedence refusal in an earlier set?
3. **C16(c), F2-5(b).** The first ruling on library operators says "A scope refusal keeps its reason `scope`". Under the positional rule those refusals become `grammar` wherever the narrowed token keeps a terminal. Does "Accept/refuse only" narrow that earlier ruling too, as it says it narrows "Precedence refusals"?

## Round 3

Round 3 assessed revision 4 (1fab4a89, blob 68ee2dd6). The four round-2 participants were resumed, and assess-fresh3 read the subject fresh.

### Ruling after round 2

perigrin made one ruling after round 2. `rulings.md` records it verbatim under "2026-11-25, after assess round 2" and says the question came from rows A20, C16 and P14: **"Diagnose by rerun (Recommended)"**.

Of round 2's "Rows that turn on a ruling": questions 2 (A20) and 3 (C16(c), F2-5(b)) are answered by it. Question 1 (C8) has no recorded answer; its raiser holds C8, "narrowed but not gone", still marked for perigrin.

### perigrin's words during round 3

Said after round 3's reports began arriving, recorded verbatim: "The goal is alignment and consensus".

### Drafter: growth check

Made by the drafter and reported at the top of `r3-pushback.md`. It holds no view on the subject. It compared `git diff a8808dd2 1fab4a89 -- docs/plans` (one file, 254 lines to 258; 14 lines removed, 18 added) with `response-r2.md` and the ruling "Diagnose by rerun".

- **Hunks that trace to a blocking row or the ruling:** §4.1's "every terminal in the set that can follow the rule's LHS" (C14, F2-1); the deleted "no lookahead" sub-bullet and "The lookahead set always exists" (C15, P13, A8's held part); the un-narrowed-lattice sentence (A8, C15(b), P13(2); also the subject of residual C7, P4 and F2-7); §4.1's new last paragraph and §5's exception for the two toys (F2-2); §4.4's "diagnosed by rerun" bullet and §2's anti-goal sentence (A20, C16, P14; the ruling); §4.4's "No caller can parse with precedence off" (the ruling's last sub-bullet).
- **Hunks named by the response as consequences of the ruling, which no blocking row asked for:** §4.5 (the referees use the no-precedence mode; "no separate copy is kept") and §6's last bullet. They change the ground under residual P5(c) and C10(f).
- **Hunks that answer nothing:** the status line's revision number only. Revision 4 has no layout-only hunks.
- **Growth:** three things are new in substance: the production no-precedence mode and the rerun on refusal (the ruling); the same mode as the referees' source of V1's forest (the ruling's words are "used only to classify refusals" and "not a way to parse"); and a table-generation refusal for a grammar that fails §3's checks, which goes beyond the sentence F2-2's own repair proposed. One clause left the subject: "`refused-where.txt` is regenerated", which the round-1 ruling "Accept/refuse only" lists.

### Status changes in round 3

Only the raiser changed a row's status.

| raiser | released in round 3 | held in part | held | promoted |
|---|---|---|---|---|
| assess-architecture | A8, A20 | none | A11–A16, A19, A21–A25 (residual; A13 "narrower") | none |
| assess-codebase | C14, C15, C16, C7 | none | C5, C6, C8, C10, C11, C12, C13, C17 (residual) | none |
| assess-pushback, as raiser | P13, P14 | P5 ((c) released in round 3) and P8 (residual) | P4, P9, P10, P11, P12, P15, P16 (residual) | none |
| assess-fresh2 | F2-1, F2-2, F2-3, F2-5 | F2-8 (residual) | F2-4, F2-7 (residual); F2-6 | F2-6, residual to blocking |

Every one of the 9 rows held as blocking after round 2 is released: A8, A20, C14, C15, C16, P13, P14, F2-1 and F2-2. Three further rows declared residual are released: C7, F2-3 and F2-5.

Releases that came with a new row on the text that answered the old one: A20 with A26; C16 with C19; F2-2 with F2-6 (promoted) and F2-10; F2-5 with F2-11.

**The promotion.** assess-fresh2 promoted its own F2-6 from residual to blocking and holds it as promoter. Its stated grounds: revision 4 turned the list of checks into a gate, and the two places that state it disagree. §4.1 says `nullableShift` fails "§3's checks"; §5 lists three zero counters and the set-1626 list; by the raiser's measurement (`TestZZF2Toys`) `nullableShift` has reduce conflicts 0, parent conflicts 0 and no context needing a copy or over-keeping, and its only failure is "forbidden shifts by a non-holding item = 1". So under §5's list it passes and the recognizer judges it. `emptyred` also fails the context counter. Its proposed fix is a fourth zero counter in §5.

Round 3 added 20 rows, all held: 4 declared blocking (A26, A27, C18, P17) and 16 residual (A28, A29, C19, P18, F2-9 to F2-11, F3-1 to F3-9).

**After round 3: 86 rows; 33 released; 5 held rows are blocking** (A26, A27, C18, P17, and F2-6 by promotion); 48 held rows are residual (3 of them held in part).

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

Named by the coordinator:

- **C18 / P17 / F2-9:** §4.4's "No caller can parse with precedence off" against the referees' and the toys' use of the no-precedence mode, and against the ruling's "used only to classify refusals". C18 and P17 are declared blocking; F2-9 is declared residual.

Found while drafting:

- **A27** is the fourth row on the same defect, declared blocking. All four name §6's "no caller of `Parse`" as the wording the rest of the subject supports, and C18, P17 and F2-9 mark the ruling's words for perigrin. P17 alone adds a third use: a grammar whose tables are refused. **F3-7** names the exported switch the mode needs, as unlisted scope.
- **A26 / C19 / F2-11 / F3-3:** "Every other reason and byte matches V1's" for a `scope` refusal from narrowing. A26 is declared blocking; C19, F2-11 and F3-3 residual. The reports' measurements differ and are recorded without resolving them: A26 reports three constructed inputs whose byte moves (`foo plus 2 plus 3;` 4 against 11; `foo bar 1 plus 2;` 4 against 10; `f $x plus 2 plus 3;` 5 against 12) and three that match; C19 and F3-3 report 3 of 3 corpus files equal; F2-11 reports 29 of 29 constructed inputs equal; assess-pushback's release of P14 reports 3 corpus files and 16 constructed inputs equal. A26 also says the first ruling ("its byte may move") and the last ("Only a `precedence` refusal's byte moves") disagree.
- **A29 / F3-6:** the reduce rule is vacuous where no terminal in the set can follow the LHS. The same point appears, not as a row, in F2-3's release ("one crumb, not a hold") and P13's release ("one loose end, not a row").
- **F2-6 (promoted) / F3-5 / F2-10:** which checks decide that a grammar passes. F2-6: the shift-shape check is not among §5's tests. F3-5: two further properties are not among them. F2-10: where the table-generation refusal sits. assess-codebase, in its report and not as a row, says it checked the toy claim and found it true: `emptyred … a nonterminal=2`, `nullableShift … a terminal=1, a nonterminal=1`, every other toy 0, perly.y 674 and 0.
- **A28 / P18:** the rerun's cost. The two reports say different things about the timed population, recorded without resolving them: A28 says the speed aggregate is drawn from files both parsers accept; P18 says the speed side runs the pipeline on every aggregate file, refused ones included.
- **F3-4 / C10(f) / P5(c):** the referee and production now share the recognizer and Build. P5(c) is released on the same change (no separate copy).
- **A23 / F3-8:** the governing design is not a file in the tree.
- **F3-7 / F2-8 / A12 / C10(e) / P5(a):** V1's tables live in `forest`, which imports `recognize`.
- **F3-2 / C6:** Build must walk by context. **F3-1 / P12:** §4.3's collapse rule. **F3-9 / C11:** the unmeasured cost of a context on interning and the cache.
- **P4 / assess-fresh3's codebase axis** (not a row): assess-fresh3 ran a recognizer on narrowed lattices and reports 0 narrowed sets missing from the committed tables over the corpus and perl.git, "so narrowing needs no table change". P4, held, reports 682 of 682 one-`PLUGIN_*` sets missing, for an operator that is in scope; no corpus token keeps a `PLUGIN_*` reading after narrowing (C12).
- **Set 1626:** the constructed-input counts in the reports are now 14 (P3), 26 (C4), 38 (assess-fresh2, round 2) and 32 (assess-fresh3); perl.git `t/` reads 78 tokens and 0 hits in P10 and in assess-fresh3's run. §3 still says 14 (C17) and still lists perl.git as uncovered (P10).

Classifications that differ across raisers for the same defect, recorded without resolving them: the no-precedence mode (A27, C18 and P17 blocking; F2-9 residual); the `scope` byte (A26 blocking; C19, F2-11 and F3-3 residual).

### Per-participant record

**assess-architecture.** Recommendation: modify. Blob 68ee2dd6, checked. It released A8 and A20 and let its twelve residual rows stand. It raised A26 and A27 (blocking) and A28 and A29 (residual). It says both blocks "come from the rerun ruling's text meeting the rest of the subject", that each is a one-sentence fix, and that "the direction and the mechanism stand". The only code it ran is the A26 probe. Promotions: none. It declared blocks, not stand-asides.

**assess-codebase.** Recommendation: modify. Blob 68ee2dd6, checked. It released C14, C15, C16 and C7 and held eight residual rows. It raised C18 (blocking, touching a ruling, for perigrin) and C19 (residual). It says C18 "came in with the round-2 ruling's consequences" and that "its fix is one word". It re-ran its checks in a scratch worktree, including the restricted reduce rule over the corpus (0 mixed tokens, against 17,100 for the unrestricted rule in round 2). Promotions: none; it names P15 as the nearest to blocking and leaves it with its raiser. It declared a block, not a stand-aside.

**assess-pushback, as raiser.** Recommendation: modify, "on one new row (P17)". Blob 68ee2dd6, checked. It released P13 and P14, released P5(c), and held its other residual rows. It raised P17 (blocking) and P18 (residual), and reported two discards. Promotions: none. It declared a block, not a stand-aside. It is not an independent participant.

**assess-fresh2.** Recommendation: modify. Blob 68ee2dd6, checked. It released F2-1, F2-2, F2-3 and F2-5, promoted F2-6 to blocking, and held F2-4, F2-7 and F2-8 (in part). It raised F2-9, F2-10 and F2-11 (residual). It says "Nothing else I found holds the subject open" besides F2-6, whose fix "is one line in §5". It declared a block, not a stand-aside.

**assess-fresh3.** Recommendation: **accept**. Blob 68ee2dd6, checked. A fresh read. "I found no blocker in revision 4." It raised F3-1 to F3-9, all residual, and says none holds the decision open. It answered the three axes "matches", "matches" and "matches". It says it tried to break the reduce rule, the shift rule, set 1626 and narrowing before recognition by measurement, and it ran a recognizer on narrowed lattices over the corpus and perl.git `t/` (accept/refuse differs from V1 on 0 files in each). Its scratch tests are in `fresh3-code/` in the session scratchpad. It did not examine whether `emptyred` and `nullableShift` fail §3's checks.

No participant recommended reject. No participant stood aside on any row.

### Outcome

One accept and four modify. Five held rows are blocking: A26, A27, C18, P17, and F2-6 by promotion. No fixed point. **Another round is owed.**

### Rows that turn on a ruling

Put as questions; this minute gives no answers.

1. **A27, C18, P17, F2-9.** The ruling says the no-precedence mode is "used only to classify refusals" and "is not a way to parse". §4.5 has the referees build V1's forest from it, and §5 keeps two toys on it. Is the referees' use within the ruling? If it is, does §4.4 take §6's wording, "no caller of `Parse`"?
2. **A26.** The first ruling on library operators says a scope refusal's "byte may move". The last says "Only a `precedence` refusal's byte moves". Which holds for a `scope` refusal from narrowing?
3. **C8** (held, for perigrin; unanswered since round 1). Do the guards' grammars, which declare no precedence levels, get levels?

## Round 4

Round 4 assessed revision 5 (de604896, blob 3a89343d). The five round-3 participants were resumed, and assess-fresh4 read the subject fresh.

### What was recorded in `rulings.md` during round 3

`rulings.md` gained a section "2026-11-25, during assess round 3" in the same commit as revision 5. It records:

- perigrin's words "The goal is alignment and consensus", said "while round 3's reports were arriving, after the orchestrator suggested that a contradiction three participants had raised (C18, P17 and A27) might need perigrin's call", with what follows from them: "A finding the subject can answer goes back to its raisers to settle. Only views that do not reconcile go to perigrin."
- a note on the round-2 entry: two of its clauses, "Only a `precedence` refusal's byte moves" and "Used only to classify refusals", quote the option perigrin chose; perigrin did not say them.

`response-r3.md` says none of round 3's blocking rows was taken to perigrin as a question. Of round 3's "Rows that turn on a ruling": the raisers of question 1 (A27, C18, P17, F2-9) and question 2 (A26) settled them by release, each against the note. Question 3 (C8) has no recorded answer; its raiser holds C8 as residual and no longer marks it for perigrin.

### Drafter: growth check

Made by the drafter and reported at the top of `r4-pushback.md`. It holds no view on the subject. It compared `git diff 1fab4a89 de604896 -- docs/plans` (one file, 258 lines to 269; 4 lines removed, 15 added) with `response-r3.md`.

- **Hunks that trace to a blocking row:** §4.4's rerun sub-bullet split into four, with "A `scope` refusal from narrowing … may differ from V1's byte" (A26); §4.4's "No caller of `Parse` can turn precedence off" and the mode's two uses (C18, A27, P17; also residual F2-9); "Every guard over perly.y" in place of "Every guard" (P17's third use, named by the response); §5's fourth zero counter (F2-6, blocking by promotion).
- **Hunks that trace only to a residual row:** none.
- **Hunks that answer nothing:** the status line's revision number only. No layout-only hunks.
- **Growth:** no new section, mechanism or criterion beyond the fourth zero counter, which is the repair F2-6's raiser proposed. The response's table covers every content hunk. Outside `docs/plans`, the commit adds the `rulings.md` section described above, on which the response relies for A26 and for C18, A27 and P17.

### Status changes in round 4

Only the raiser changed a row's status.

| raiser | released in round 4 | held in part | held |
|---|---|---|---|
| assess-architecture | A26, A27 | none | A11–A16, A19, A21–A25, A28, A29 (residual) |
| assess-codebase | C18, C19 | none | C5, C6, C8, C10, C11, C12, C13, C17 (residual) |
| assess-pushback, as raiser | P17 | P5 and P8 (residual) | P4, P9, P10, P11, P12, P15, P16, P18 (residual) |
| assess-fresh2 | F2-6 (its own promotion), F2-9, F2-11 | F2-8 (residual) | F2-4, F2-7, F2-10 (residual) |
| assess-fresh3 | F3-3 | none | F3-1, F3-2, F3-4 to F3-9 (residual) |

Every one of the 5 rows held as blocking after round 3 is released: A26, A27, C18, P17 and F2-6. Four further residual rows are released: C19, F2-9, F2-11 and F3-3. No row was promoted in round 4: each resumed participant wrote "none".

Three raisers were asked by the response to judge whether the referees' use of the no-precedence mode fits the ruling. assess-architecture, assess-codebase and assess-pushback each recorded that it does, in releasing A27, C18 and P17.

Round 4 added 7 rows, all held: 1 declared blocking (F4-1) and 6 residual (A30, F3-10, F4-2 to F4-5).

**After round 4: 93 rows; 42 released; 1 held row is blocking (F4-1);** 50 held rows are residual (3 of them held in part).

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

- **A30 / F3-10 / C8:** `FuzzRecognize` listed under "Every guard over perly.y". Both new rows are declared residual and both tie it to C8, whose raiser makes the same point in holding it. assess-pushback notes it in releasing P17 and leaves it to C8.
- **F4-1 / C7, P4, F2-7 (the referee as an un-narrowed caller) / C18, P17 (the referee's call sites):** earlier rows name `chart/precedence.go: builtForest` among the callers that recognize an un-narrowed lattice, and among the referees outside `recognize`. F4-1 is the first row on what the bison referee does after that: it narrows the forest, not the lattice, and calls `Constrain` from non-test code.
- **F4-2 / C12:** the corpus holds no kept library operator, so the `PLUGIN_*` levels are not exercised there.
- **F4-3 / P4:** terminal-set ids for narrowed tokens. F4-3 says the existence of every narrowed set was not measured. The ledger holds two measurements: P4, 0 of 62 narrowed-to-none sets missing and 682 of 682 narrowed-to-one sets missing; assess-fresh3, 0 sets missing over the corpus and perl.git `t/`, re-run in round 4.
- **F4-4 / A12 / C10(e) / P5(a) / F2-8 / F3-7:** V1's tables live in `forest`, which imports `recognize`. Six rows.
- **F4-5 / F2-7 / C3 (released) / P7 (released):** the two chart-shape keys are taken from an un-narrowed `Recognize`.
- **F3-4 / C10(f):** the referee and production share the recognizer and Build. assess-architecture weighed promoting F3-4 against the governing design's rule 11 and left it residual, naming a pin for refinement (the no-precedence mode's Build forest equal to today's `forest-digest.golden`).
- **F3-5 / F2-6 (released):** assess-fresh3 keeps F3-5 residual and notes that F2-6, "a gap of the same kind", was promoted by its raiser.
- **F2-10:** held by its raiser. assess-codebase, weighing it for promotion, says §5's "stay judged by `Constrain`" already fixes the reading.
- **Set 1626:** assess-fresh4 adds 35 constructed inputs, none reaching the set; the counts in the reports are now 14, 26, 38, 32 and 35.

### Per-participant record

**assess-architecture.** Recommendation: **accept**. Blob 3a89343d, checked. It released A26 and A27, holds no blocking row and raised A30 (residual). Fourteen residual rows of its own stand. It weighed F3-4 for promotion and left it. It ran no test this round. Promotions: none.

**assess-codebase.** Recommendation: **accept**. Blob 3a89343d, checked. It released C18 and C19, holds no blocking row and raised nothing. Eight residual rows stand. It re-ran the committed static feasibility tests; the fourth zero counter reads 0 on perly.y (`a terminal=674, a nonterminal=0`). It weighed F2-10 and P15 for promotion and left both. Promotions: none.

**assess-pushback, as raiser.** Recommendation: **accept**. Blob 3a89343d, checked. It released P17, holds no blocking row and raised nothing. Ten residual rows stand, two of them in part. Promotions: none. It is not an independent participant.

**assess-fresh2.** Recommendation: **accept**. Blob 3a89343d, checked. It released F2-6 (which it had promoted), F2-9 and F2-11, and holds F2-4, F2-7, F2-8 (in part) and F2-10 as residual. It raised nothing and re-ran nothing, the code being unchanged. Promotions: none.

**assess-fresh3.** Recommendation: **accept**. Blob 3a89343d, checked. It re-ran its round-3 measurements and reports every number unchanged. It released F3-3, holds eight residual rows and raised F3-10 (residual). It weighed P9, C5 and C8 for promotion and left each. Promotions: none.

**assess-fresh4.** Recommendation: **modify**. Blob 3a89343d, checked. A fresh read. It raised F4-1 (blocking) and F4-2 to F4-5 (residual). It says "The direction survives; one sentence of §5 and one of §4.5 do not, and one answer key is left without an owner." Its axes: the codebase "matches, except for the bison referee"; architecture "matches the decisions in force"; direction "sound". It re-ran the committed static feasibility tests and ran the four counters over every bison toy ("exactly the two the design names fail"). Its scratch code is in `fresh4-code/` in the session scratchpad. It declared a block, not a stand-aside.

No participant recommended reject. No participant stood aside on any row.

### Outcome

Five accept and one modify. Every row that was blocking going into the round is released by its raiser. **No fixed point:** a fresh reader, assess-fresh4, raised a new blocking finding, F4-1, and it is held. **Another round is owed.**

## Round 5

Round 5 assessed revision 6 (bcd26304, blob a9c54728). The six round-4 participants were resumed, and assess-fresh5 read the subject fresh. `rulings.md` is unchanged since round 4; the brief says perigrin made no ruling in between. `response-r4.md` says the author "took the raiser's first fix" for F4-1 and offered it to the raiser to settle.

### Drafter: growth check

Made by the drafter and reported at the top of `r5-pushback.md`. It holds no view on the subject. It compared `git diff de604896 bcd26304 -- docs/plans` (one file, 269 lines to 275; 1 line changed, 6 added) with `response-r4.md`.

- **The hunk that traces to a blocking row:** §4.5's new last paragraph. The bison referee (`referee.judge`, behind `precedence-disagree`) today narrows the forest after Build and calls `Constrain`; under V2 it narrows the lattice to the sentence's terminals before recognition and judges the recognizer's forest; "So `precedence-disagree` checks the recognizer, and `Constrain` stays test-only." It traces to F4-1 and is option (a) of the two repairs its raiser offered.
- **Hunks that trace only to a residual row:** none. The response leaves the representation of the narrowed sets to residual F4-3, and the diff adds nothing on it.
- **Hunks that answer nothing:** the status line's revision number only.
- **Growth:** one paragraph is new. It adds a second narrowing before recognition, the referee's, beside the library-operator narrowing of §4.4. §5 is untouched, including the sentence residual F4-2 quotes.

### The release of F4-1

assess-fresh4 released F4-1, its only blocking row. It says "the measurement, not the paragraph, is what I release on". The measurement, emulated at bcd26304 with today's recognizer on singleton terminal sets, then Build, then V1's `Constrain`:

- `lattice: kept=1296 disagree=0 (accepts what bison refuses: 0; refuses what bison accepts: 0)`;
- `tokens whose set lacks bison's terminal=0`;
- with the forest narrow added back, `lattice+forest-narrow: disagree=0`;
- its round-4 proof re-run on the same tree: `v1: disagree=0`, `exact-first: disagree=71`, `setrule: disagree=233`, unchanged.

Three other reports measured the same move, none as a row:

- assess-codebase: `kept sentences=1296; today's referee disagrees with bison on 0; lattice-narrowed referee equals today's on 1296, differs on 0; lattice-narrowed disagrees with bison on 0; sentences where the slipper inserted a ; in the narrowed run=1296`. Its comparison includes the grouping. Its failing bracket: with a sentence's last two terminals swapped, a difference on 200 of 200.
- assess-fresh5: `kept sentences 1296; precedence-disagree today 0; with lattice narrowing 0`. Its bracket, with the narrowing removed: 610 disagreements.
- assess-pushback: of the 1,296, 501 can be narrowed with the tables as committed, and accept or refuse differs from today's order on 0 of them. It did not compare groupings or test the other 795.

Each of the four says no V2 recognizer exists, so the measurements show the narrowing move, not V2's agreement with bison.

### Status changes in round 5

Only the raiser changed a row's status.

| raiser | released in round 5 | held, with a change recorded | held unchanged |
|---|---|---|---|
| assess-architecture | none | none | A11–A16, A19, A21–A25, A28–A30 (residual) |
| assess-codebase | none | C10 ((f) listed as standing) | C5, C6, C8, C11, C12, C13, C17 (residual) |
| assess-pushback, as raiser | none | P4 (new evidence) | P5 and P8 (in part), P9–P12, P15, P16, P18 (residual) |
| assess-fresh2 | none | F2-7 ("narrowed again") | F2-4, F2-8 (in part), F2-10 (residual) |
| assess-fresh3 | none | F3-4 (a note) | F3-1, F3-2, F3-5 to F3-10 (residual) |
| assess-fresh4 | F4-1 | F4-2 ("narrower"), F4-3 ("wider") | F4-4, F4-5 (residual) |

assess-fresh2's round-5 table lists "F2-11 crumb" as held, residual. F2-11 was released in round 4 "with one crumb left" (the death byte is undefined for a file refused with every set alive). The row's status keeps both entries; this minute counts the row as released and the crumb as held by its raiser.

No row was promoted: each resumed participant wrote "none". Round 5 added 6 rows, all held: 1 declared blocking (F5-1) and 5 residual (C20, F5-2 to F5-5).

**After round 5: 99 rows; 43 released (F2-11 with a crumb its raiser holds); 1 held row is blocking (F5-1);** 55 held rows are residual (3 of them held in part).

### New evidence on held residual rows

**P4 and F4-3: terminal-set ids for narrowed tokens.** Both stay residual by their raisers' declaration. Five reports add figures:

- assess-pushback, on P4: over the committed referee sentences, `tokens=7567 already-that-one-terminal=6323 need-narrowing=1244 of-them-no-singleton-set-in-tables=979`; `sentences with a narrowed set missing=795 (PLUGIN_ sentences among them: 588); distinct missing singletons=19`. Its round-1 figure re-run: `narrowed-to-none missing=0 narrowed-to-one total=682 missing=682`.
- assess-codebase, as row C20, and assess-fresh5, as row F5-5: 19 singleton sets missing from the tables, each appended in the test process.
- assess-fresh3, on F4-3: the committed tables have 1,880 terminal sets over 119 terminals; 90 terminals have a one-terminal set and 29 do not. The 29 include all eleven `PLUGIN_*` terminals and also `PERLY_MINUS`, `PERLY_PLUS`, `PERLY_STAR`, `PERLY_EQUAL_SIGN`, `PERLY_COLON`, `PREINC`, `PREDEC`, `POSTINC`, `POSTDEC`, `NOAMP`, `METHCALL` and `YADAYADA`.
- assess-fresh4, on F4-3: to emulate §4.5 it "had to append a singleton set for each terminal to a copy of the tables".
- A way through named in three reports: the recognizer already accepts a set added to the tables after `New` (`grammarTables.extend`; `TestRecognizeSetAddedAfterNew`).

The 19 and the 29 measure different things: 19 is the missing singletons the sentence fixture uses; 29 is every terminal with no one-terminal set.

**assess-fresh3's correction.** In round 3 it wrote that "every narrowed set already has an id, so narrowing needs no table change". In round 5 it says: "That was too broad." Its measurement covered the corpus and perl.git, where no library operator is ever kept, so every narrowed set was the token's set less all its `PLUGIN_*` terminals; "It says nothing about a kept operator, nor about the referee's narrowing that revision 6 adds." This bears on the round-3 cross-reference between P4 and that statement.

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

- **C20 / F5-5 / P4 / F4-3 / F4's and F3's evidence:** the single-terminal sets the referee's narrowing needs. All residual.
- **F5-1 / F3-5:** F3-5 (residual, held) names, as a property not among §5's counters, that no rule with an all-nullable right-hand side is forbidden before a terminal in FOLLOW of its LHS. F5-1 (blocking) is on the same property: it shows a toy, `empty`, where the property fails while all four counters pass, and proposes the property as a fifth counter. The reports' counts for perly.y differ and are recorded without resolving them: assess-fresh3, "29 nullable-RHS rules, 0 forbidden"; assess-fresh5, "0 of 13 empty rules". assess-codebase's round-1 axis text has "29 rules derive only the empty string; none has a non-zero `reduceAt`".
- **F5-1 / F5-2 / F2-2 (released) / F2-6 (released) / F2-10:** which toy grammars fall outside the recognizer, and how the subject says so. Earlier reports found that exactly the two named toys fail the four counters (assess-fresh2, round 3; assess-codebase, round 3, on the fourth counter; assess-fresh4, rounds 4 and 5). F5-1 does not dispute that: `empty` passes the four and still needs a verdict the recognizer cannot apply. F5-2 adds a pinned toy (`TestPrecedenceMarkFirstFilter`'s) that fails counter 4 and is not one of the two.
- **F5-4 / A22 / P16:** retiring the stage key edits `baseline.txt` and the key list inside the milestone. **F5-4 / F2-7 / F4-5:** the stage harness and the chart-shape keys recognize un-narrowed lattices.
- **F5-3 / C7 (released) / P4 / F2-7:** another direct `recognize.Recognize` caller, `recognizeSource`, this one behind `refused-where.txt`. **F5-3 / C12:** no corpus or port snippet is refused for precedence.
- **F4-2:** held by its raiser as "narrower". assess-codebase and assess-architecture each say, not as a row, that §5's sentence on V1's forest for "every referee sentence" is still inexact and that it is F4-2.
- **Set 1626:** assess-fresh5 adds 21 constructed inputs, none reaching the set; the counts in the reports are now 14, 26, 38, 32, 35 and 21.

### Per-participant record

**assess-architecture.** Recommendation: **accept**. Blob a9c54728, checked. It holds no blocking row and raised nothing. Fifteen residual rows stand. It checked the new paragraph against V1 §3.5 and the governing design's rule 11 and says it "fits the decisions in force". It ran no test. Promotions: none.

**assess-codebase.** Recommendation: **accept**. Blob a9c54728, checked. It holds no blocking row. It measured the referee move (above) and raised C20 (residual, evidence for F4-3). Eight residual rows stand. Promotions: none.

**assess-pushback, as raiser.** Recommendation: **accept**. Blob a9c54728, checked. It holds no blocking row and raised nothing. Ten residual rows stand, two in part; P4 carries new evidence. It weighed promoting P4 and F4-3 and left both. Promotions: none. It is not an independent participant.

**assess-fresh2.** Recommendation: **accept**. Blob a9c54728, checked. It holds no blocking row and raised nothing. It re-ran no test. It re-checked its released rows against revision 6 and found each fix in place. Promotions: none.

**assess-fresh3.** Recommendation: **accept**. Blob a9c54728, checked. It re-ran its measurements and reports every number unchanged. It holds nine residual rows, raised nothing, corrected its round-3 statement (above) and added evidence to F4-3. It says of F4-1: "I missed it in rounds 3 and 4." Promotions: none.

**assess-fresh4.** Recommendation: **accept**. Blob a9c54728, checked. It released F4-1 on its own measurement and holds F4-2 to F4-5 as residual. It raised nothing. It weighed C6, F3-1, F3-2, P9, F2-10 and P15 for promotion and left each. Promotions: none.

**assess-fresh5.** Recommendation: **modify**. Blob a9c54728, checked. A fresh read. It raised F5-1 (blocking) and F5-2 to F5-5 (residual). It says "One blocking finding (F5-1), with a small fix. Everything else I checked holds." Its axes: the codebase "matches, with one exception" (the toys); architecture "matches the decisions in force as far as I can check"; direction "sound for perly.y". Its scratch code is in `fresh5-code/` in the session scratchpad. It declared a block, not a stand-aside.

No participant recommended reject. No participant stood aside on any row.

### Outcome

Six accept and one modify. The row that was blocking going into the round, F4-1, is released by its raiser. **No fixed point:** a fresh reader, assess-fresh5, raised a new blocking finding, F5-1, and it is held. **Round 6 is owed. The bound is 8.**

## Round 6

Round 6 assessed revision 7 (28da56fa, blob 5cb3ab81). The seven round-5 participants were resumed, and assess-fresh6 was dispatched to read the subject fresh. Seven reports were in hand when this section was drafted. `response-r5.md` says the author "took the raiser's fix" for F5-1 and offered it to the raiser to settle.

### perigrin's ruling during round 6

Recorded in `rulings.md` under "2026-11-25, during assess round 6", under the heading "Acceptance". perigrin's words, verbatim: "I think if 7 of 8 participants agree and there are no blockers we could move forward yeah".

`rulings.md` records with it:

- perigrin said this while round 6's last report, from the new fresh reader assess-fresh6, was still out;
- at that point seven participants recommended accept (assess-codebase, assess-architecture, assess-pushback, assess-fresh2, assess-fresh3, assess-fresh4 and assess-fresh5), and no blocking finding was held;
- "Revision 7 is accepted.";
- "assess-fresh6's report is recorded in the minute when it arrives. If it raises a blocking finding, the finding goes to perigrin and is not answered silently."

### Drafter: growth check

Made by the drafter and reported at the top of `r6-pushback.md`. It holds no view on the subject. It compared `git diff bcd26304 28da56fa -- docs/plans` (one file, 275 lines to 278; 4 lines removed, 7 added) with `response-r5.md`.

- **Hunks that trace to the blocking row:** §4.1's last paragraph ("Two bison toys fail" replaced by "A bison toy that fails any of the checks is judged as §5 says", with three toys as examples and the reason `empty` fails); §5's "five counters must be 0", the fifth being "empty rules with a non-allow reduce verdict before any terminal"; §5's exception, "the two toys" becoming "every toy that fails §3's checks". All trace to F5-1. The response says the last also covers residual F5-2's third toy "without listing it".
- **Hunks that trace only to a residual row:** none.
- **Hunks that answer nothing:** the status line's revision number only.
- **Growth:** one counter is new, and the toy exception changed from a count to a rule; both are the repair F5-1's raiser proposed. Two facts about how the repair landed, stated without a view. F5-1's proposed fix reads "add a fifth counter to §3 and §5"; the diff adds it to §5, and §3's table still has four rows. F5-1's proposed counter reads "no rule whose symbols are all nullable has a non-allow reduce verdict before a FOLLOW terminal"; the subject's reads "empty rules with a non-allow reduce verdict before any terminal".

### Status changes in round 6

Only the raiser changed a row's status.

| raiser | released in round 6 | stood aside | held, with a change recorded | held unchanged |
|---|---|---|---|---|
| assess-architecture | none | none | none | A11–A16, A19, A21–A25, A28–A30 (residual) |
| assess-codebase | none | none | none | C5, C6, C8, C10, C11, C12, C13, C17, C20 (residual) |
| assess-pushback, as raiser | none | none | none | P4, P5 and P8 (in part), P9–P12, P15, P16, P18 (residual) |
| assess-fresh2 | none | none | F2-10 ("a little wider") | F2-4, F2-7, F2-8 (in part) (residual) |
| assess-fresh3 | none | none | F3-5 ("narrower") | F3-1, F3-2, F3-4, F3-6 to F3-10 (residual) |
| assess-fresh4 | none | none | none | F4-2 to F4-5 (residual) |
| assess-fresh5 | F5-1, F5-2 | F5-5 | none | F5-3, F5-4 (residual) |

- **F5-1, released by its raiser.** It names its bar ("a fifth counter in §3 and §5, over rules whose symbols are all nullable, and the toy exception stated as a rule") and scores revision 7 against it. Landed: the exception as a rule in §4.1 and §5; five zero counters in §5. Re-run at 28da56fa: toy `empty` passes the first four counters and reads 1 on the fifth; perly.y reads 0 on the fifth as revision 7 words it. "Where it falls short of the letter": the counter covers empty rules, not all-nullable non-empty rules, and §3's table has no fifth row. The raiser says "Neither holds the decision open", gives its test of the narrower wording (a unit-empty toy on which bison's conflict sits on the inner empty rule, which the counter covers), and records the remainder as F5-6, residual.
- **F5-2, released by its raiser:** the rule form covers `TestPrecedenceMarkFirstFilter`'s toy without listing it.
- **F5-5, stood aside by its raiser.** Its grounds, in full: "It is the same fact as C20 and belongs to F4-3, which carries it." This is the first stand-aside in the session.
- **F2-11.** assess-fresh2 says "The minute's record of F2-11 is right": the row is released as of round 4, and the crumb is a separate residual it still holds.
- assess-fresh4 re-ran F4-1's proof on revision 7's tree and reports it unchanged; the row "stays released".

No row was promoted: each of the seven wrote "none". Round 6 added 7 rows, all declared residual and all held: A31, C21, P19, F2-12, F3-11, F4-6 and F5-6.

**After round 6, with seven of eight participants reporting: 106 rows; 45 released; 1 stood aside (F5-5); no held row is blocking;** 60 held rows are residual (3 of them held in part).

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

Named by the coordinator:

- **C21 / P19 / F2-12 / F3-11 / F4-6: the zero-width, all-nullable gap in the fifth counter.** Five raisers (assess-codebase, assess-pushback, assess-fresh2, assess-fresh3 and assess-fresh4) each found, in its own round-6 report, that the counter says "empty rules" while `builder.complete` skips every zero-width completion, so that the counter covers 13 of perly.y's 29 all-nullable rules. Each classed it residual. Where each raiser's statement sits:
  - C21(a), P19, F2-12 and F4-6 are rows on the gap itself;
  - assess-fresh3 states the gap in its round-6 entry for F3-5 (held, "narrower"), with a measurement (`TestZZF3HiddenReduces`: predicted states reduce 13 rules with an empty right-hand side and 16 with a non-empty one). Its new row, F3-11, is on a neighbouring point: §3 and `feasibility.md` carry no row or figure for the fifth counter.

Found while drafting:

- **A31 and F5-6 name the same gap**, both residual, which makes seven raisers. A31 says it did not measure it and rests on the minute's F5-1 row; F5-6 is the remainder its raiser recorded on releasing F5-1.
- **The reports' figures agree:** 13 rules with an empty right-hand side and 16 more that are all nullable, 29 in all, every one reading 0 (C21, P19, F4-6, F5-6, and assess-fresh3's F3-5 entry; F2-12 cites the counts). This reconciles the two counts recorded without resolution in round 5 (assess-fresh5's 13 and assess-fresh3's and assess-codebase's 29): they count different classes. P19 and assess-fresh3 each report that `optlistexpr ::= empty` is the only one of the 29 that carries a precedence level, and that it is among the 16.
- **C21(b) / F3-11 / F5-6: no row in §3 for the fifth counter.** C21(b) adds that §3 reports the fourth counter no more than the fifth. The growth check records the same fact.
- **Four raisers refer the gap to F5-1's raiser:** A31, P19, F2-12 and F4-6 each say whether the narrower wording meets F5-1's bar is its raiser's to judge. Its raiser released F5-1 and recorded F5-6.
- **F5-5 (stood aside) / C20 / F4-3 / P4:** the single-terminal sets the referee's narrowing needs. assess-fresh3 notes that C20 and F5-5 report 19 sets the sentence fixture needs and lacks, and it reported 29 terminals with no one-terminal set at all.
- **F2-10:** three raisers weigh it in round 6 without changing its class. assess-codebase says `loadToy` builds all 13 toys without error today; assess-fresh5 says revision 7 adds `empty` to the toys it matters for.
- **F5-3:** assess-codebase, assess-fresh3 and assess-fresh4 each weighed it for promotion and left it residual; none found a snippet where the two bytes differ.
- **Which toys fail:** assess-codebase reports exactly three of the 13 named toys and `nullableShiftToy` failing the two counters it can run over toy tables (`empty`, `emptyred`, `nullableShift`). assess-architecture's report says "the minute records two failing §3's checks (`emptyred`, `empty`)"; the minute's rows name `emptyred`, `nullableShift`, `empty`, and a pinned toy of `TestPrecedenceMarkFirstFilter` (F5-2). Recorded without resolving it.

### Per-participant record

**assess-architecture.** Recommendation: **accept**. Blob 5cb3ab81, checked. It holds no blocking row. It raised A31 (residual). Fifteen residual rows stand. It ran no test. Promotions: none.

**assess-codebase.** Recommendation: **accept**. Blob 5cb3ab81, checked. It holds no blocking row. It measured the fifth counter against the tables: 0 on perly.y, 1 on toy `empty`, and not redundant with the fourth. It raised C21 (residual). Nine residual rows stand. Promotions: none.

**assess-pushback, as raiser.** Recommendation: **accept**. Blob 5cb3ab81, checked. It holds no blocking row. It raised P19 (residual, flagged for F5-1's raiser). Ten residual rows stand, two in part. Promotions: none. It is not an independent participant.

**assess-fresh2.** Recommendation: **accept**. Blob 5cb3ab81, checked. It holds no blocking row. It raised F2-12 (residual). It re-ran no test and re-checked its released rows against revision 7. Promotions: none.

**assess-fresh3.** Recommendation: **accept**. Blob 5cb3ab81, checked. It re-ran its measurements and reports every earlier number unchanged. It holds nine residual rows, F3-5 "narrower", and raised F3-11 (residual). Promotions: none.

**assess-fresh4.** Recommendation: **accept**. Blob 5cb3ab81, checked. It holds no blocking row. It re-ran F4-1's proof, checked the fifth counter as a reader (perly.y 0 over 13 empty rules; toy `empty` 1 of 2), and raised F4-6 (residual). Four residual rows stand. Promotions: none.

**assess-fresh5.** Recommendation: **accept**. Blob 5cb3ab81, checked. It released F5-1 and F5-2, stood aside on F5-5, holds F5-3 and F5-4 as residual, and raised F5-6 (residual). It re-ran its round-5 experiments on toy `empty`. Promotions: none; it names C6, F3-2, F2-10 and P15 as nearest to blocking and leaves each.

**assess-fresh6.** A fresh participant dispatched in round 6. Its report was outstanding when perigrin ruled and when this section was drafted. It will be recorded here when it arrives. By `rulings.md`, a blocking finding in it "goes to perigrin and is not answered silently".

No reporting participant recommended reject or modify. One row, F5-5, was stood aside by its raiser.

### Outcome

**Accepted by perigrin's ruling during round 6, within the bound of eight rounds.** Revision 7 (28da56fa, blob 5cb3ab81) is the accepted revision.

- All seven reporting participants recommend accept.
- No blocking finding is held. Every row ever declared or promoted blocking is released by its raiser: 32 rows across rounds 1 to 5 (17, 8, 4 and one promotion, 1, 1).
- 60 residual rows are held, for refinement and chain-review. Seven of them, raised in round 6, are on the accepted revision's fifth counter.
- This is not recorded as a fixed point. One participant, assess-fresh6, had not reported. The acceptance rests on perigrin's ruling, quoted above.

## Round 6, continued: assess-fresh6's report and the ruling after it

Everything under "Round 6" above was drafted with seven reports in hand and stands as the record of that moment. This section records what arrived afterwards.

### assess-fresh6's report

**assess-fresh6.** Recommendation: **modify**. It read revision 7 (28da56fa), blob 5cb3ab812cfe20097c3255f47060a79b3f42ec97, checked with `git hash-object`. A fresh read. It raised F6-C1, F6-C2 and F6-C3 (blocking) and F6-R1 and F6-R2 (residual).

- It says: "On perly.y the direction holds. Every static assumption I could test against the committed tables comes out 0, including two the subject does not pin." And of its three blocking rows: "None is about direction, and each has a one-sentence fix."
- Its axes: the codebase claims "reproduce" (it re-ran the committed static feasibility tests and reports §3's figures); architecture, "no conflict between a ruling and the subject"; direction, "Sound for perly.y".
- For F6-C1 and F6-C2 it proposes two more counters, both 0 on perly.y today: rules with an all-nullable right-hand side, and core items of one state that scan one terminal with different verdicts. For F6-C3 it proposes moving §4.3's rule to §6, or saying the context is kept from prediction to completion.
- It says no V2 recognizer exists, so each "the recognizer cannot" is argued from `complete`, `scan` and the automaton's states, not run, and that it did not check either toy against bison.
- Its scratch tests are in `fresh6-code/` in the session scratchpad. It declared blocks, not stand-asides.

### perigrin's ruling after round 6

Recorded in `rulings.md` under "2026-11-25, after assess round 6", under the heading "The late findings". perigrin's words, verbatim: "Revise to rev 8, run round 7 (Recommended)".

`rulings.md` records with it:

- assess-fresh6's report arrived after the acceptance ruling, recommended modify and raised three blocking findings; "F6-C3 is a defect for perly.y itself";
- "Revision 8 answers all three.";
- "Round 7 runs within the bound of eight, to confirm consensus.";
- "The acceptance of revision 7 is held until round 7 is done."

Revision 8 is commit 2d9611db, blob d893c1f0.

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

- **F6-C1 / A31, C21, P19, F2-12, F4-6, F5-6, and assess-fresh3's F3-5 entry:** the fifth counter covers empty rules and not every rule whose right-hand side is all nullable. Seven raisers classed it residual in round 6; assess-fresh6 declares it blocking. Classifications differ and are recorded without resolving them. F6-C1 adds what the others did not have: a toy grammar that passes all five counters while V1 refuses a sentence the unconstrained chart accepts. P19's row says no such toy was built; F5-6's unit-empty toy read 0 on both wordings.
- **F6-C2 / F3-5's second property:** no state scans one terminal for two core items with different verdicts. F3-5 (residual, held) measured it as 0 over 76 states on perly.y and says it is not among the counters; F6-C2 (blocking) gives a toy that passes every counter and needs two answers for one scan.
- **F6-C3 / F3-1 / P12 / C6, F3-2 / A15:** §4.3's collapse rule. F3-1 (residual, held) says the rule "looks forward only" and that a completion must return only to parents that predicted its context; it was reasoned, not measured, and its raiser said it belongs in §6. F6-C3 (blocking) gives a perly.y input, `foo -$x * $y;`. P12 is on the two collapse rules; C6 and F3-2 on Build walking by context; A15 on mechanism in the decision.
- **F6-R1 / P4 / F4-3:** the narrowed terminal sets for a kept operator; the same figure, 682 of 682 absent, in P4 and F6-R1.
- **F6-R2 / C13:** Leo under precedence.

### Tally and state

**After assess-fresh6's report: 111 rows; 45 released; 1 stood aside (F5-5); 3 held rows are blocking (F6-C1, F6-C2, F6-C3);** 62 held rows are residual (3 of them held in part).

Round 6 ends with seven accept and one modify. The acceptance perigrin ruled during round 6 is held, by perigrin's later ruling, until round 7 is done. Round 7 assesses revision 8, within the bound of eight.

## Round 7

Round 7 assessed revision 8 (2d9611db, blob d893c1f0). The eight earlier participants were resumed, and assess-fresh7 read the subject fresh. Nine reports. `response-r6.md` says revision 8 takes the raiser's fix for F6-C1 and F6-C2 and "the raiser's second option" for F6-C3.

### Drafter: growth check

Made by the drafter and reported at the top of `r7-pushback.md`. It holds no view on the subject. It compared `git diff 28da56fa 2d9611db -- docs/plans` (one file, 278 lines to 281; 5 lines removed, 8 added) with `response-r6.md`. The range spans two commits: 10a844da ("accept the V2 design at revision 7", which touched the status line) and 2d9611db.

- **Hunks that trace to a blocking row:** §4.1's "no completion event for … a rule whose right-hand side is all nullable, when it spans no tokens" and §5's fifth counter reworded to "rules whose right-hand side is all nullable" (F6-C1; the same wording is the repair asked for by residual A31, C21(a), P19, F2-12, F4-6, F5-6 and assess-fresh3's F3-5 entry); §5's "six" and the new sixth counter (F6-C2; also F3-5's second property); §4.3's new first paragraph and the tightened collapse sentence (F6-C3; bears on residual F3-1 and P12).
- **Hunks that trace only to a residual row:** none.
- **Hunks that answer nothing:** the status line's revision number. Its three added sentences record perigrin's two rulings after round 5; no finding asked for them.
- **Growth:** one counter is new. §4.3 gained a paragraph of mechanism (which parents a completion advances). F6-C3's raiser offered two repairs, moving the collapse rule to §6 or saying the context is kept from prediction to completion; the diff takes the second, so §4.3 grew where the first would have shrunk it. §3's table still has four rows, against six zero counters and the set list in §5.

### Status changes in round 7

Only the raiser changed a row's status.

| raiser | released in round 7 | stood aside | held, with a change recorded | held unchanged |
|---|---|---|---|---|
| assess-architecture | A31 | none | none | A11–A16, A19, A21–A25, A28–A30 (residual) |
| assess-codebase | C21 (on (a)) | C21(b) only | C6, C11 | C5, C8, C10, C12, C13, C17, C20 (residual) |
| assess-pushback, as raiser | P19 | none | P12 ("changed shape") | P4, P5 and P8 (in part), P9–P11, P15, P16, P18 (residual) |
| assess-fresh2 | F2-12 | none | none | F2-4, F2-7, F2-8 (in part), F2-10, the F2-11 crumb (residual) |
| assess-fresh3 | F3-1, F3-2, F3-5 | none | F3-11 ("wider") | F3-4, F3-6 to F3-10 (residual) |
| assess-fresh4 | F4-6 | none | none | F4-2 to F4-5 (residual) |
| assess-fresh5 | none | F5-5, as before | F5-6 ("narrower") | F5-3, F5-4 (residual) |
| assess-fresh6 | F6-C1, F6-C2, F6-C3 | F6-R1 | none | F6-R2 (residual) |

- **F6-C1, F6-C2 and F6-C3, released by their raiser.** assess-fresh6 re-ran every round-6 probe at 2d9611db. F6-C1: its toy `nulwrap` now trips the fifth counter (`nullable-nonempty non-allow any=1`) while passing the other five; perly.y reads 0 over all 29 such rules. F6-C2: its toy `scancollide` is caught by the sixth counter and passes the other five; perly.y and the 13 toys read 0; one thing left open is F6-R3. F6-C3: "Yes, the rule as now written keeps the unary-minus parent out", by the two new sentences of §4.3; the remaining wording is F6-R4. It says it is biased toward finding its own bar met, and that where the wording leaves room it raises residuals "rather than count it as met".
- **F6-R1, stood aside by its raiser:** "Same fact as F4-3 and P4, which carry it". The second stand-aside in the session.
- **The rows on the fifth counter's wording, released by their raisers:** A31, C21(a), P19, F2-12 and F4-6, with F3-5 (both properties).
- **Two raisers corrected their own earlier calls.**
  - assess-fresh3: "F6-C1, F6-C2 and F6-C3 are the substance of my F3-5 (both halves) and F3-1. I raised them in round 3 and held them as residual for four rounds … My classification was too lenient, and for F3-1 I had the reasoning without the construction."
  - assess-fresh5: its round-6 release of F5-1 rested on a unit-empty toy; assess-fresh6's toy put the conflict on the all-nullable non-empty rule, "so my release was against something that landed nearby, not my own bar. Revision 8 now meets the bar I wrote in round 5." F5-1 "stays released, now on the bar as written".
- assess-fresh4 re-ran F4-1's proof at 2d9611db and reports it unchanged.

No row was promoted: each resumed participant wrote "none". Round 7 added 11 rows, all held: 1 declared blocking (C22) and 10 residual.

**After round 7: 122 rows; 56 released; 2 stood aside (F5-5, F6-R1); 1 held row is blocking (C22);** 63 held rows are residual (3 of them held in part).

### assess-fresh7's prototype measurement

assess-fresh7 built a scratch prototype of §4.1 to §4.3 on the committed tables and ran it against V1. This is the first measurement in the session made with a recognizer that applies the design's rules. Its description: "A plain LR(0)-state Earley over `tab.States`, with the slipper, no Leo, no cache", with items (state, origin, context), the predictor's one canonical top, the scan and completion refusals of §4.1, §4.3's parent matching, and library operators' dead pairs removed from the terminal sets first. It is compared with V1 (`Recognize`, `Build`, `libraryOperators`, `Constrain`) on accept or refuse and on the set of (rule, from, to) lying on a full parse.

| Run | Inputs | V1 accepted | `precedence` refusals | Accept differs | Over-keep | Over-prune |
|---|---|---|---|---|---|---|
| Referee sentences, guacamole and T1 | 3,175 | 2,532 | 13 | 0 | 0 | 0 |
| Random source, seed 7 | 1,000,000 | 99,137 | 148 | 0 | 0 | 0 |
| Random source, seed 23 | 600,000 | 59,468 | 93 | 0 | 0 | 0 |
| Random source with six library operators in scope, seed 11 | 400,000 | 56,672 | 85 | 0 | 0 | 0 |

- With precedence off, the prototype accepts exactly what the committed recognizer accepts (`noprec-differs=0` in every run).
- **Its brackets.** With every item carrying `none` (`F7_ONECTX=1`): 13 accept differences and 694 over-keeps. With a completion advancing every parent (`F7_NOMATCH=1`): 26 over-keeps. "So both the context and §4.3's parent matching are load-bearing".
- During every run it counted four of the cases §5 pins as never arising (a completion with a mixed verdict across its lookahead set, a scan whose core items disagree, a parent state whose awaiting items disagree, two rules of one LHS with different verdicts): 0 on all 2,003,175 inputs.
- Set 1626 after `indirob ::= scalar`: 736,233 generated `print …` prefixes, 39,588 holding a token with set 1626, 0 with the completion before it. It gives a structural reason: a `scalar` ends in a Variable token or a `}`, both in `notHashBrace`.
- Toys: on the eleven that pass every counter, the prototype equals V1 and bison on all 4,378 sentences; on `emptyred` it differs from V1 on 1 of 7 sentences; on `empty` no sentence differs, but the counter fires.
- Its stated limits: no Leo memo, no set interning, transition cache or `Expected`; it compared (rule, from, to), not split points, rule-position nodes or alternative order; forests were compared only for inputs of at most 600 tokens; std5k, edit5k and perl.git `t/` were not run. On Build walking per context, it tried a Build that pairs by §4.4's parent records only: 0 over-keeps on the corpus run and on 600,000 random inputs; "I could not construct a failing input, and I did not prove none exists."
- Its code is in `fresh7-code/` in the session scratchpad.

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

Named by the coordinator:

- **C22 / P20:** §4.3's sentence "A completion advances only the parent items whose context predicted it". C22 is declared blocking; P20 residual.

Found while drafting:

- **F3-12 and F6-R4 are on the same sentence**, both residual. Four raisers (assess-codebase, assess-pushback, assess-fresh3, assess-fresh6) each say the sentence can be read as comparing the parent's own context with the completion's, that §4.2 and the example give the right reading, and that the rule itself is right. Classifications differ and are recorded without resolving them: C22 blocking; P20, F3-12 and F6-R4 residual. C22's raiser says: "If the other participants read the sentence as loosely as its author meant it, I would not hold the decision on it." The four proposed rewordings: "the parent items that predicted it" (C22); "the parent items that gave it its context (§4.2)" (P20); "only the parents that predicted it under that context" (F3-12); F6-R4 proposes none.
- Three other reports read the sentence without raising a row. assess-architecture: the context a parent hands its child "is a function of the parent's state and the parent's own context". assess-fresh2: it "selects whole predictor states and needs no fact beyond the spine top". assess-fresh7's prototype implements it as "a completion advances only parents whose context predicted it" and reports 0 differences.
- **A32 / P12:** §4.3's collapse sentence against §4.2's canonical collapse. assess-fresh4 says P12's two rules "are now one condition stated twice".
- **F6-R4 / P20 / A15 / F3-1:** the rule as mechanism in the decision; F6-R4 adds that it is a third judgement site filed under "Collapsing contexts".
- **F6-R3:** assess-codebase and assess-fresh4 each say, not as a row, that their count for the sixth counter includes dot-0 items.
- **F7-P1 / A30 / F3-10 / C8:** `FuzzRecognize` under "Every guard over perly.y". Four rows, all residual.
- **F7-C1 / A12 / C10(e) / P5(a) / F2-8 / F3-7 / F4-4:** the tables live in `forest`, which imports `recognize`. Seven rows, all residual.
- **F7-C2:** the slipper under the referee's lattice narrowing. In round 5 assess-codebase reported the slipper inserting a `;` in all 1,296 narrowed sentences and changing nothing, and assess-fresh4 reported `lattice+forest-narrow: disagree=0`; neither was a row.
- **F7-C3:** `feasibility.md` §4 added `;` to closing sets by hand; F7-C3 says no §5 counter states the property.
- **F7-A1 / F3-11, F5-6, C21(b):** what the counters are and where their evidence is. F3-11 and F5-6 are on the missing §3 rows; F7-A1 is on what "context conflicts" names.
- **C11 / F3-9:** five reports (assess-codebase, assess-architecture, assess-fresh2, assess-fresh4, assess-fresh5) say §4.3's "0.01% of nodes" no longer describes how many items keep a context, and each leaves it with C11 and F3-9 and §7's stop.
- **F6-R2 / C13:** Leo. assess-fresh7 adds, under what it did not examine, that a reduce verdict in the middle of a chain breaks the argument `builder.leo` rests on today.

### Per-participant record

**assess-codebase.** Recommendation: **modify**, "by one phrase". Blob d893c1f0, checked. It measured the fifth counter (29 rules, 0; toy `empty` 1) and the sixth (`(state, terminal) with >=2 core items scanning it=467 … differing verdicts=0`; three synthetic states counted) and says both hold. It says §4.3's new rule "is right and needed" and its example real. It released C21 on (a), stood aside on C21(b), and raised C22 (blocking). It says: "I would release on a rewording and need no further round for it." Promotions: none. It declared a block, not a stand-aside.

**assess-architecture.** Recommendation: **accept**. Blob d893c1f0, checked. It answers the question put to it, whether §4.3's new rule stays within "Items may carry the canonical spine top and nothing else": "Yes. The rule changes what the carried fact is used for and when it may be dropped. It does not add a fact to an item." It released A31 and raised A32 (residual). Fifteen residual rows stand. It ran no test. Promotions: none.

**assess-pushback, as raiser.** Recommendation: **accept**. Blob d893c1f0, checked. It released P19 and raised P20 (residual). Ten residual rows stand, two in part. Promotions: none. It is not an independent participant.

**assess-fresh2.** Recommendation: **accept**. Blob d893c1f0, checked. It released F2-12, holds five residuals including the F2-11 crumb, and raised nothing. It says §4.3's example checks out and agrees with §4.2. Promotions: none.

**assess-fresh3.** Recommendation: **accept**. Blob d893c1f0, checked. It re-ran its measurements unchanged and added `TestZZF3MinusExample`. It released F3-1, F3-2 and F3-5, holds F3-11 "wider" and five other residuals, and raised F3-12 (residual). It corrected its own classification (above). Promotions: none.

**assess-fresh4.** Recommendation: **accept**. Blob d893c1f0, checked. It released F4-6, re-ran F4-1's proof, and measured the sixth counter on perly.y over every top (`states scanning one terminal for >=2 core items=76; (state, top, terminal) with mixed verdicts=0; forbidden scans seen=2082`). Four residual rows stand. It raised nothing. Promotions: none.

**assess-fresh5.** Recommendation: **accept**. Blob d893c1f0, checked. It corrected its round-6 release of F5-1 (above), holds F5-3, F5-4 and F5-6 ("narrower"), keeps F5-5 stood aside, and raised nothing. It checked the fifth and sixth counters and §4.3's example (`TestZZF7Counters`, `TestZZF7Example`). Promotions: none.

**assess-fresh6.** Recommendation: **accept**. Blob d893c1f0, checked. It released F6-C1, F6-C2 and F6-C3, stood aside on F6-R1, holds F6-R2, and raised F6-R3 and F6-R4 (residual). Promotions: none.

**assess-fresh7.** Recommendation: **accept**. Blob d893c1f0, checked. A fresh read. "No blocking finding. Five residuals." Its axes: codebase "matches", architecture "matches", direction "matches". It built and ran the prototype recorded above. It raised F7-P1, F7-C1, F7-A1, F7-C2 and F7-C3, all residual.

No participant recommended reject. Two rows have been stood aside by their raisers.

### Outcome

Eight accept and one modify. The three rows that were blocking going into the round, F6-C1, F6-C2 and F6-C3, are released by their raiser. assess-codebase raised and holds one new blocking finding, C22, on the wording of one sentence in §4.3. No fixed point.

**C22 was held, so revision 9 was made (commit bc9b4abd, blob 6d4dacab), and round 8 is owed. Round 8 is the bound.** The acceptance perigrin ruled during round 6 remains held.

## Round 8

Round 8, the bound, assessed revision 9 (bc9b4abd, blob 6d4dacab). The nine round-7 participants were resumed, and assess-fresh8 read the subject fresh. Ten reports. The brief says perigrin made no ruling after round 7. `response-r7.md` says revision 9 takes "the raiser's phrase" for C22, that it also answers P20, and that "No other text changed" apart from the status line.

### Drafter: growth check

Made by the drafter and reported at the top of `r8-pushback.md`. It holds no view on the subject. It compared `git diff 2d9611db bc9b4abd -- docs/plans` (one file, 281 lines before and after; 2 lines changed) with `response-r7.md`.

- **The hunk that traces to the blocking row:** §4.3's second sentence, "A completion advances only the parent items whose context predicted it" becoming "A completion advances only the parent items that predicted it". It traces to C22 and is the phrase C22's raiser proposed. Residual P20, F3-12 and F6-R4 are on the same sentence.
- **Hunks that trace only to a residual row:** none.
- **Hunks that answer nothing:** the status line's revision number. Its last clause, "held until round 7 is done" becoming "held until the assessment closes", follows from round 7 not closing the assessment; no finding asked for it.
- **Growth:** none. The line count is unchanged, one phrase was shortened, and no sentence, counter or mechanism was added.

### Status changes in round 8

Only the raiser changed a row's status.

| raiser | released in round 8 | stood aside in round 8 | held in part | held unchanged |
|---|---|---|---|---|
| assess-architecture | none | none | none | A11–A16, A19, A21–A25, A28–A30, A32 (residual) |
| assess-codebase | C22 | none | none | C5, C6, C8, C10, C11, C12, C13, C17, C20 (residual) |
| assess-pushback, as raiser | P20 | none | P5, P8 | P4, P9–P12, P15, P16, P18 (residual) |
| assess-fresh2 | none | none | F2-8 | F2-4, F2-7, F2-10, the F2-11 crumb (residual) |
| assess-fresh3 | F3-12 | none | none | F3-4, F3-6 to F3-11 (residual) |
| assess-fresh4 | none | none | none | F4-2 to F4-5 (residual) |
| assess-fresh5 | none | none (F5-5 as before) | none | F5-3, F5-4, F5-6 (residual) |
| assess-fresh6 | none | none (F6-R1 as before) | F6-R4 | F6-R2, F6-R3 (residual) |
| assess-fresh7 | none | F7-P1, F7-C1, F7-C2 | none | F7-A1, F7-C3 (residual) |

- **C22, released by its raiser.** assess-codebase: the sentence now reads "the phrase I named as my bar, so I checked it the other way: does it fit the example and §4.2, and does anything else in the subject still carry the old reading? It fits." Its re-run `TestZZR8Contexts` on `foo -$x * $y;`: the binary-minus parent's rule carries `{none}`; `$x * $y` is `allowed` under `{termbinop: term ADDOP . term}` and `FORBIDDEN` under `{termunop: PERLY_MINUS . term}`. It says it looked for a weak point at the left-corner parent and that "predicted it" covers that case too.
- **P20 and F3-12, released by their raisers.** Each says the reading it flagged is gone because the sentence no longer names the parent's context, and that its own suggested wording was not the one used.
- **F6-R4, held in part.** assess-fresh6 releases the wording half and holds the placement half as residual: the pairing rule "is a third judgement site filed under 'Collapsing contexts' and not in §4.1".
- **assess-fresh7's three stand-asides,** each because the row duplicates rows already in the ledger. Its grounds, in full:
  - F7-P1: "The same fact as A30 and F3-10, both held residual and tied to C8. Their raisers hold it; a third row adds nothing."
  - F7-C1: "The same fact as A12, F2-8, F3-7 and F4-4 (tables built in `forest`, which imports `recognize`). Held residual by four raisers."
  - F7-C2: "The minute's round-5 evidence already measures it: assess-codebase's lattice-narrowed referee equals today's on all 1,296 sentences, with the slipper inserting a `;` in every one. My own bracket in round 7 agrees."
- **Three participants corrected their round-7 reading of the old phrase.** assess-architecture: "I read the old phrase as 'the context the parent gives its child' and did not flag it. C22 was right that the words allowed the wrong reading." assess-fresh2: "my reading then was more charitable than the text." assess-fresh4: "C22 is right that the old words, compared field against field, pick no parent in the paragraph's own example."
- assess-fresh3, assess-fresh4, assess-fresh5 and assess-fresh6 each re-ran the measurements behind rows they released in earlier rounds and report every number unchanged.

No row was promoted. Round 8 added 8 rows, all declared residual and all held: F7-A2 and F8-P1 to F8-P4, F8-C1, F8-C2 and F8-A1.

### assess-fresh7's measurement of the two phrases

assess-fresh7 re-ran its round-7 prototype at bc9b4abd with the new wording: "A parent at a later rule position advances when the context it *gives* its children equals the completion's. A left-corner parent advances when it carries the same context."

| Run | Inputs | V1 accepted | Accept differs | Over-keep | Over-prune |
|---|---|---|---|---|---|
| Referee sentences, guacamole and T1 | 3,175 | 2,532 | 0 | 0 | 0 |
| Random source, seed 31 | 400,000 | 39,498 | 0 | 0 | 0 |
| Random source with library operators in scope, seed 37 | 200,000 | 28,319 | 0 | 0 | 0 |

So revision 9's phrase, as the prototype implements it, matches V1 on 603,175 inputs.

- **Revision 8's phrase, read literally** (`F7_OWNCTX=1`: a parent advances when the context it carries equals the completion's): on the corpus run the prototype accepts 702 inputs where V1 accepts 2,532, a loss of 1,830 acceptances. First: `s0003 "qw(a) < qw(a) < qw(a) ;"`, V1 accepts, the prototype dies at token 3. Its words: "So C22 was a real defect, and revision 9 removes it."
- **The other bracket** (`F7_NOMATCH=1`: every awaiting parent advances): 26 over-keeps, as in round 7. "So the rule still has to select parents". This is the evidence for its new residual row F7-A2.
- The four "never arises" cases of §5 stayed 0 in each run; the static recounts are unchanged.

### assess-fresh8's disclosure

assess-fresh8 lists two "leaks I did not rely on":

- "`brief-r8.md` names C22 and P20 as the reason for revision 9's §4.3 phrase. I read that brief before noticing it was for resumed participants."
- "One `grep -l` listed `docs/assessments/chart-precedence-v2.md` as a file that mentions the governing design. I saw the file name only, no rows."

It says of its seven rows: "I cannot say which of the rows above repeat earlier findings."

### Cross-references

Rows that name the same defect. Cross-referencing does not merge rows; each raiser keeps its own.

- **C22 / P20 / F3-12 / F6-R4 (wording half):** all four released against the same phrase. **F7-A2 / A32 / F6-R4 (placement half) / P12:** what still stands on §4.3, all residual: the new sentence read with no context (F7-A2); the collapse sentence against §4.2 (A32); where the rule is filed (F6-R4); how the collapse is computed (P12).
- **F8-P1 / F7-C2 (stood aside):** the slipper's zero-width `;` under the referee's lattice narrowing. Four reports measured it as moving no verdict: assess-codebase and assess-fresh4 (round 5), assess-fresh7 (round 7) and assess-fresh8.
- **F8-C1 / C17:** "14 constructed inputs" has no committed file behind it.
- **F8-C2 / F3-11 / F5-6:** counters five and six have no committed measurement and no row in §3.
- **F8-A1 / A23 / F3-8:** the governing design is cited without its branch.
- **F8-P2 / A30 / F3-10 / C8, and F7-P1 (stood aside):** `FuzzRecognize` under "Every guard over perly.y". F8-P2 adds that the reference recognizer becomes a second implementation of the judgement, which P5(d) also names.
- **F8-P3 / P15:** set 1626. P15: the stop has no observer during the milestone. F8-P3: after delivery nothing reports it.
- **F8-P4 / F4-2:** §5's sentence on V1's old pipeline. F4-2: it is not the referee's path. F8-P4: §4.5 lists the pipeline one stage shorter.
- **F7-C1 (stood aside) / A12 / C10(e) / P5(a) / F2-8 / F3-7 / F4-4:** the tables live in `forest`.

### Per-participant record

**assess-architecture.** Recommendation: **accept**. Blob 6d4dacab, checked. It holds no blocking row and raised nothing. Sixteen residual rows stand. It says the new sentence "stays within the ruling" and "names no new fact". It ran no test. Promotions: none.

**assess-codebase.** Recommendation: **accept**. Blob 6d4dacab, checked. It released C22 and raised nothing. Nine residual rows stand. Promotions: none.

**assess-pushback, as raiser.** Recommendation: **accept**. Blob 6d4dacab, checked. It released P20 and raised nothing. Ten residual rows stand, two in part. Promotions: none. It is not an independent participant.

**assess-fresh2.** Recommendation: **accept**. Blob 6d4dacab, checked. It holds four residual rows (one in part) and the F2-11 crumb, and raised nothing. Promotions: none.

**assess-fresh3.** Recommendation: **accept**. Blob 6d4dacab, checked. It released F3-12, holds seven residual rows and raised nothing. It weighed F7-C2 and F7-C3 for promotion and left both. Promotions: none.

**assess-fresh4.** Recommendation: **accept**. Blob 6d4dacab, checked. It holds four residual rows and raised nothing. It weighed F7-C2, F6-R3, F7-C3, A32, P12, C11 and F3-9 for promotion and left each. Promotions: none.

**assess-fresh5.** Recommendation: **accept**. Blob 6d4dacab, checked. It holds three residual rows, keeps F5-5 stood aside, and raised nothing. Promotions: none.

**assess-fresh6.** Recommendation: **accept**. Blob 6d4dacab, checked. Its three blocking rows stay released on re-run. It holds F6-R2 and F6-R3, holds F6-R4 in part, keeps F6-R1 stood aside, and raised nothing. Promotions: none.

**assess-fresh7.** Recommendation: **accept**. Blob 6d4dacab, checked. It stood aside on F7-P1, F7-C1 and F7-C2, holds F7-A1 and F7-C3, and raised F7-A2 (residual). It weighed F3-2, A25 and F6-R2 for promotion and left each. Promotions: none.

**assess-fresh8.** Recommendation: **accept**. Blob 6d4dacab, checked. A fresh read. "I found no blocking finding in revision 9." Its axes: codebase "matches", architecture "matches, by ruling", direction "sound". It raised F8-P1 to F8-P4, F8-C1, F8-C2 and F8-A1, all residual: "None holds the subject open." It re-derived the totals (46 tops, 674 triples), and counters four, five and six on perly.y, and tried 44 constructed inputs against set 1626.

No participant recommended reject or modify.

### Outcome

**A fixed point at round 8, within the bound of eight.**

- All ten participants recommend accept, on revision 9 (bc9b4abd, blob 6d4dacab).
- No blocking finding is held. The one row blocking going into the round, C22, is released by its raiser.
- No blocking finding was raised in the round, including by the fresh reader, assess-fresh8.
- Every resumed participant reports having read the residual rows, and each wrote "none" under promotions.

**Final tally: 130 rows.**

| | rows |
|---|---|
| released by their raisers | 59 |
| stood aside by their raisers | 5 (F5-5, F6-R1, F7-P1, F7-C1, F7-C2) |
| held, residual | 66 (4 of them held in part: P5, P8, F2-8, F6-R4) |
| held, blocking | 0 |

Of the 130, 36 were declared or promoted blocking by their raisers across the eight rounds (17 in round 1; 8 in round 2; 4 and one promotion in round 3; 1 in round 4; 1 in round 5; 3 in round 6's late report; 1 in round 7). All 36 are released.

The acceptance perigrin ruled during round 6, on revision 7, was held by the ruling "Revise to rev 8, run round 7 (Recommended)" until the assessment closed. The assessment closes here, on revision 9.

## Residual rows for refinement

The 66 rows held as residual at the close, one line each, in the raisers' terms. The full row is in the ledger. One further item is not a row: the F2-11 crumb.

**assess-architecture (16)**

- A11: §7 lacks "a stop overrides the exit", a quiet-session clause and a budget.
- A12: the tables are "read as they are" from a package `recognize` cannot import.
- A13: the reduce side needs lookahead, which interleaved mode does not have; unweighed.
- A14: the postmortem's `NewPrecedence` timing key and follow-ups are not carried.
- A15: the decision carries milestone and builder's detail (§4.3, §4.4's file names, §5's lists, §9).
- A16: refusal records "beside the sets" against "their values include the refusal records".
- A19: the live README is false from the first recogniser issue until a final docs issue.
- A21: rule 12's scope clause is not in the header's supersession list.
- A22: the retired stage key has no home in §4 or §7 and touches `baseline.txt` in the milestone.
- A23: the governing design is cited without its branch.
- A24: rule 2 is quoted without saying whether it still binds.
- A25: "the only set where this can happen is set 1626" holds for linear lattices only.
- A28: the rerun's cost on refused input, the usual case in an editing loop, is unweighed.
- A29: the reduce rule is vacuously true where no terminal in the set can follow the LHS.
- A30: `FuzzRecognize` is listed under "every guard over perly.y".
- A32: §4.3's collapse sentence, read alone, seems to deny `none` where §4.2 grants it.

**assess-codebase (9)**

- C5: the corpus never exercises V1's union shortcut; no fixture is named.
- C6: Build must walk by context; §4.4's refusal-record bullets do not say so.
- C8: the guards' grammars declare no precedence levels.
- C10: unlisted scope: (a) `TestRecognitionKeysUnchangedByPipeline`; (d) the `TestItemIsStateDistance*` tests; (e) the tables' move; (f) referee and production share recognizer and Build.
- C11: the cost of a context on set interning and the cache is unmeasured.
- C12: the corpus has no precedence refusal and no kept library operator.
- C13: Leo's shift side and the context on the memo are new and uncovered.
- C17: §3's "14 constructed inputs" has no committed file.
- C20: 19 single-terminal sets the referee fixture needs are not in the tables.

**assess-pushback (10)**

- P4: narrowed terminal sets have no representation for a kept operator (682 of 682 absent) or for 795 of 1,296 referee sentences.
- P5, in part: (a) the tables live in `forest`; (b) an item-id mapping is needed; (d) the reference recognizer gains precedence.
- P8, in part: (b) nothing in the exit asserts the pass left production; (c) the six `alloc-*` and `gc-*` keys are neither held nor exempted.
- P9: `resolve` reads unreachable nodes, which the differential does not compare; measured 0 effect.
- P10: §3 still lists perl.git `t/` as uncovered; the feasibility counters read clean there.
- P11: §1's "same answers as V1" beside a changed `Expected` and two bytes that may move.
- P12: §4.3's collapse rule names no way to compute it and is not in §6.
- P15: the set-1626 stop has no named observer.
- P16: retiring the stage key edits `baseline.txt` beside the non-goal on retaking it.
- P18: the rerun's cost and its use of the caller's long-lived recognizer are unstated.

**assess-fresh2 (4, and the crumb)**

- F2-4: the header omits three reversed clauses of the governing design.
- F2-7: the ratchet's direct `Recognize` callers pass un-narrowed lattices.
- F2-8, in part: the tables live in `forest`, which imports `recognize`.
- F2-10: if the table-generation refusal sits in `NewPrecedence`, `Constrain` has no tables for the failing toys.
- The F2-11 crumb (the row itself is released): "the byte where the chart died" is undefined for a file refused with every set alive.

**assess-fresh3 (7)**

- F3-4: nothing pins the no-precedence mode's sets; referee and production share code.
- F3-6: the reduce rule is vacuous where no terminal can follow.
- F3-7: moving or injecting the tables, and an exported switch, are unlisted scope.
- F3-8: the governing design is not in the tree.
- F3-9: the cost of a context in every item is unmeasured.
- F3-10: `FuzzRecognize` under "every guard over perly.y".
- F3-11: §3's table has four rows for §5's six counters.

**assess-fresh4 (4)**

- F4-2: §5 words V1's forest for referee sentences through `LibraryOperators`, which is not the referee's path.
- F4-3: a narrowed token needs a terminal-set id the tables lack.
- F4-4: the tables must cross from `forest` to `recognize`.
- F4-5: the two chart-shape keys come from a recognizer run production never makes.

**assess-fresh5 (3)**

- F5-3: `refused-where.txt` is read by two tests that would recognize differently under V2.
- F5-4: the ratchet's stage harness and the `baseline.txt` key removal are unnamed scope.
- F5-6: §3 and `feasibility.md` have no row for the fifth or sixth counter.

**assess-fresh6 (3)**

- F6-R2: 51 of 65 Leo-eligible rules can be refused; a cost risk for the exit.
- F6-R3: the sixth counter does not say that a dot-0 item counts as allowed.
- F6-R4, in part: the pairing rule is a third judgement site filed under "Collapsing contexts", not §4.1.

**assess-fresh7 (3)**

- F7-A1: "context conflicts" does not say which check it is.
- F7-C3: agreement at closing sets rests on `;`, `}` and EOF having no level, which no counter states.
- F7-A2: "that predicted it", read with no context, is plain Earley; "under its context (§4.2)" would close it.

**assess-fresh8 (7)**

- F8-P1: the referee's forests will carry the slipper's zero-width `;` leaves.
- F8-C1: "14 constructed inputs" cannot be re-run from the committed evidence.
- F8-C2: counters five and six have no committed measurement.
- F8-A1: the governing design is cited without its branch.
- F8-P2: "gain precedence" has no meaning for `FuzzRecognize`; the reference recognizer becomes a second implementation, unsized.
- F8-P3: after delivery nothing reports an input that reaches set 1626.
- F8-P4: §4.5 and §5 list the old pipeline one stage apart.

Rows that share a subject are grouped in the cross-reference sections of each round. The five rows stood aside by their raisers (F5-5, F6-R1, F7-P1, F7-C1, F7-C2) are each carried by a held row named in their grounds.
