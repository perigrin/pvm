# chain-review chart-perly-w4, round 4: review-alignment (resumed)

## 1. Recommendation

**Modify.** PB17, the finding I held, is released. AL14 is released too.

I now promote r3-fresh3's FR3-4 to blocking. The refusal rule that d166 and 3262 state gives `grammar` where V1 gives `lex-unmapped`. This is the same class of defect as AL1, and the fix is one clause and one subtest. If it lands I expect to recommend accept.

**How this round was done.** I read the live chain first: 19 issues at `e3432c0f`, with edges taken from `issue show`'s `blocked_by`. I ran the checks before reading `repair-r3.md`.

- **Edges:** unchanged from round 3. Stage 1 runs d02f → d05d → d089 → d0ba. Stage 2 fans out to {d0e1, d10d} → d135. Stage 3 runs d166 → d199 → d1c2 → 3282 → 3262 → {d1eb, d218}. Then come d218 → {d247, d278, d2a2}, with d1eb → d278. Everything converges on d2d1, and d2d1 → bb9b.
- **Criteria:** I re-ran all 191 against `e3432c0f`, each passed to `sh -c` unmodified. All 191 are red.
- **Residual mapping:** the 66 rows and the crumb are all present. The only mismatches are the AL6 ones: A19 ×4 and C20.
- **Carries:** each issue's `Carries:` line is unchanged from round 3.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1–AL4, AL12, AL13 | released (earlier rounds) | **AL1, AL13:** re-checked against this round's 3262 text. Its rule still says the un-narrowed rerun "dies at all → `grammar`". That is right for AL1's case. FR3-4 below shows the one case the rule still misses.<br>**AL12:** d0e1's `compared 1296 of 1296` is unchanged.<br>**AL4:** 191 of 191 criteria are red. |
| PB17 (my r3 promotion) | **released** | **What the fix had to change as well:**<br>(a) the recognizer records how it was built;<br>(b) `ParseWith` refuses a no-precedence recognizer;<br>(c) a criterion proves both the refusal and the acceptance;<br>(d) no existing `ParseWith` caller passes a no-precedence recognizer;<br>(e) the guard cannot land while `New` is still the no-precedence recognizer;<br>(f) the rerun must not go through `ParseWith`.<br>**What I found, point by point:**<br>(a) and (b) are stated in 3262.<br>(c) is `TestParseHasNoPrecedenceSwitch`, with three named subtests: `no-parameter-or-option`, `parsewith-refuses-no-precedence-recognizer` and `parsewith-accepts-default`. It is red today.<br>(d) The only `ParseWith` caller in the tree, outside `Parse`, is `TestPipelineStagesReproduceParseWith` (`cmd/ratchet/stages_test.go:275`), and it passes `recognize.New`, which judges from 3262 on.<br>(e) The guard step comes after the switch inside 3262, and the prose says why.<br>(f) The prose says the rerun is "reached only … inside `parseStream`", so it bypasses `ParseWith`.<br>**Why a panic is enough:** design §4.4 says no caller of `Parse` can turn precedence off, and §6 says the builder chooses how. A guard at runtime meets that.<br>**Bias check:** I wrote this bar. Against that, I grepped every `ParseWith(` call site rather than trusting the prose. |
| AL5–AL11 | residual, held | Unchanged. The AL6 cross-check gives the same 5 misses. |
| AL14 | **released** | **The count is defined.** d247 says a sentence "carries a leaf" when the slipper inserted a zero-width `;` into its narrowed-lattice chart. The criterion now requires `slipper ; insertions: [1-9][0-9]* of 1296`, so the count must be non-zero and is not pinned to the literal. The prose also quotes my 1,277-present and 0-reachable figures.<br>**The stale quote is gone.** d0e1's "1,296 of 1,299" now reads "1,296 of the 1,296 kept sentences". |
| AL15 | residual, held | The transient part was answered by N5/FR3-1: the clause "`recognize.New` is still unjudged" left the criterion. Still unanswered: `TestJudgedRecognizerIsExplicit` logs no count of differing sets or refused completions, so a pass cannot be checked. Pushback (N9) and I both judge it low risk. |
| AL16 | residual, held | The Records entries are dated and append-only. The round-3 entry is added beneath the others. |

**The repairer's four flagged calls.**

- **PB17:** accepted, as above.
- **`TestPrecedenceGenerationErrorRefusesEveryFile` (d218):** accepted. A fate is stated: the error is raised where the judged recognizer reads `prectab`, and the `grammar.Derived` hook follows the type there. d218 also has a run-first step. If the stated fate is wrong, that step shows it and the issue reports it.
- **d247's count:** accepted, as in AL14.
- **Lists reasoned from code, not run:** accepted. d166, d218, d247 and 3262 each first apply the change alone, run `make test`, and send any red test missing from the list to the orchestrator before anything is edited. That is the correct disposition for an unmeasured list. It also covers FR3-3 (`TestCompletionProbesBounded`, with 0.6% headroom), which is why I do not promote that row.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| FR3-4 (r3-fresh3) | residual | **blocking** (I hold it) | **V1's rule.** In `parseStream`, the un-narrowed recognition dies at byte `at`. If `at == unmapped`, the byte of the first token no terminal admits, the reason is `lex-unmapped`, and otherwise it is `grammar`.<br>**The chain's rule.** 3262 says: "First the rerun is un-narrowed: if it dies at all, the reason is `grammar` at its byte". d166 says "`lex-unmapped` is decided as today". Its criterion `TestRefusalLexUnmappedNotRerun` reads "A lex-unmapped file … is never rerun".<br>**Where they part.** Take a file whose narrowed or judged chart dies early, at an out-of-scope library operator or a precedence refusal, and which has an unmapped token later on. Here `lex-unmapped` is decided on the early death byte. That byte is not `unmapped`, so the file is not `lex-unmapped`. Next the un-narrowed rerun dies at the unmapped token, and the rule says `grammar`. V1 says `lex-unmapped` at that same byte.<br>**Why this is a defect.** It breaks design §4.4's "Every reason matches V1's" and the ruling "Diagnose by rerun". No subtest of `TestRefusalRerunOrder` or of d166's rerun tests builds this two-fault file, so nothing would detect it.<br>**Classification.** This is a stated rule that gives the wrong answer when followed literally, the same class as AL1. I derived it from `forest.go` lines 116–129 and the issues' text. I did not run it. |

**Fix (d166, carried in 3262).**
- When the un-narrowed rerun dies, the reason is `lex-unmapped` if its death byte is the first unmapped token's byte, and `grammar` otherwise.
- Reword the "never rerun" criterion to "not rerun when the chart itself dies at the unmapped token".
- Add a two-fault subtest: an out-of-scope library operator followed by an unmapped token must give `lex-unmapped` at the later byte.

**Rows read and not promoted.** I read every row in r1–r3 from all participants.

- **FR3-7.** 3282's toy differential says nothing about the infeasible toys. They are new tests, a test that leaves them out still matches the criterion's grep, and d218 states the exception. So this is a gap in what the issue says, not a criterion that cannot pass. It stays residual.
- **FR3-3.** Covered by 3262's run-first step.
- **N7 and N8.** N7 is covered by 3262's run-first step. N8 (3262's size) is pushback's lens.
- **FR3-2 (fresh2).** Size, pushback's lens.
- **PB8, PB12–PB15, PB18–PB20, N2, N3, N9.** Unchanged.
- **FR2-8 / FR3-6.** Answered: d2d1's `drive.sh` now writes the `.meta` files with named keys.
- **AL5–AL11, AL15, AL16.** Unchanged.

## 4. New findings

None.

## 5. What I did not re-examine

- **The FR3-4 case, run.** I have not built a file with an early scope fault and a later unmapped token and run it at `e3432c0f`. The finding rests on the code path and the issues' wording. A one-file measurement, with V1's reason against the rule as written, would confirm it.
- **The 3262 "pinned" and "rewritten" lists against the code.** Each issue's run-first step covers that.
- **Whether `scope.pl` is accepted by every recognizer after d166's narrowing.** This is the premise of the rewritten `TestRecognitionKeysUnchangedByPipeline`. Its refusal comes from resolve's feature scope, not a library operator, so I expect it to hold. I did not run it.
- **Unchanged bodies.** d02f, d05d, d089, d0ba, d10d, d135, d1c2, d1eb, d278, d2a2, bb9b and 3282 have no diff from round 3. I did not re-read them beyond the criterion re-run.
- **perl.git targets.** Their criteria are gated red today.
