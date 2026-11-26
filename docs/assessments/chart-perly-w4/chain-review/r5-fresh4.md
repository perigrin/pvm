# chain-review round 5: review-fresh4 (holding r4-fresh4 FR4-1 to FR4-4)

I read the live chain, 19 issues, from `/home/user/pvm/.claude/worktrees/v2-refine`. The copy is in `fr5/issues.json` and `fr5/ms.json`. I diffed it against my round-4 copy in `fr4/`, and the diff is in `fr5/diff.txt`. Only then did I read `repair-r4.md`.

What changed:
- The bodies of d166, d199, d1eb, 3262 and 3282, and an appended entry in the milestone's Records.
- No `blocked_by` edge changed.
- No title or label changed.
- No row of the residual mapping changed.

## 1. Recommendation: modify

There is one blocking finding, FR5-1. It also promotes r4-fresh3 FR4-1. The round-4 repair of my FR4-1 states its intent correctly, but it rests on a judged `recognize.New(tab, Options)` that has no source of precedence for any tables value except the committed grammar's. The chain never says what that constructor does with toy tables. Each of the two readings the text allows breaks something the chain asserts. One of the new criteria, 3262's `TestParseRefusesInfeasibleGrammar`, either cannot pass or passes before its work exists.

## 2. Status of my findings

| Finding | Status | Grounds |
|---|---|---|
| r4-fresh4 FR4-1 (an infeasible grammar is recognized unjudged) | **held, narrowed** | **The d199 half is released.** d199 now owns the refusal. `TestJudgedRecognizerRefusesInfeasibleGrammar` replaces the criterion that had pinned the opposite behaviour, and d1eb's criterion follows it. That criterion can pass, provided d199's judged constructor accepts the toy's `Precedence`. It is red today. **The 3262 half is held.** The end-to-end claim, that "`Parse` over a failing toy refuses everything with the generation error", cannot be realised through `Parse` as the chain defines it. See FR5-1. |
| r4-fresh4 FR4-2 (`TestCompletionProbesBounded`, `TestLeoMemoEntriesBounded` near their bounds) | **held, residual**. I decided not to promote it. | Re-run today, both are unchanged: 22217/34924 = 0.6362 against 0.64, and 1410/35718 = 0.0395 against 0.040. I keep it residual for three reasons. (1) Which way either count moves is not measured. Pruning removes items, which lowers both counts; contexts on Leo memos could raise one of them. (2) A bound crossing is a judgement about cost, not a wrong answer. 3262's first step sends it to the orchestrator with its failure text, and then on to perigrin. That is the right owner for a cost bound, and pushback, alignment, fresh2 and fresh3 reached the same view on N7/FR3-3. (3) Promoting it would ask the chain to pick a bound in advance without a measurement, and that is worse than a reported stop. One cheap edit is still worth making whenever 3262 is next touched: name the two tests in 3262's Context, so the report is expected rather than a surprise. |
| r4-fresh4 FR4-3 (`make test` never compiles the `precedencereference`-tagged perl.git test) | **held, residual** | Not addressed, and the repair log says so. 3262 and d218 still add no `make chart-precedence-reference` step. CI's Ratchets job is the backstop. |
| r4-fresh4 FR4-4 (d1c2 and d1eb have no run-first step; `intern_test.go` uses unkeyed `Item` literals) | **held, residual** | Not addressed. d1eb was edited this round, but only its infeasible-grammar criterion. It still has no run-first step and no "Tests this change moves" paragraph. |

## 3. Promotions

**r4-fresh3 FR4-1 is promoted to blocking and merged into FR5-1.** In round 4 it was an open question: does the judged `New` swallow the "not the committed grammar's tables" error? The repair of my FR4-1 answered it the wrong way by default. d199 now says "The judged constructor's table generation returns `Feasible`'s refusal as an error, and `Recognize` returns that error, unwrapped, for every input". 3262 says "`parseStream` turns that error into the refusal of every file". Nothing separates a toy that has no levels from an infeasible grammar. So the literal reading now turns every `New(toyTables(...))` recognizer test red at 3262. The repairer's own log predicts that 3262's first `make test` "will report a batch of toy-grammar recognizer tests".

Read literally, the chain now forces that outcome; it is no longer a risk. A red batch of that size, outside 3262's "Not expected to move" class, sends 3262 back to the orchestrator before any edit. Resolving it needs a decision the chain has not made. Adding that decision once 3262 has started is an amendment of a started issue, which design §7 lists as a stop.

I read every residual row from r1 to r4: AL5–AL16, PB8–PB20, N2–N11, FR2-5–FR2-8, r3-fresh2 FR3-2, r3-fresh3 FR3-3 to FR3-7, r4-fresh2 FR4-1 and r4-fresh3 FR4-1. I promote nothing else.
- **N7, FR3-3 and my FR4-2.** Residual, for the reasons above.
- **PB8**, no criterion that the brute force runs judged. A missing check, not a wrong one.
- **N10 and N11.** Covered by the run-first steps.
- **r4-fresh2 FR4-1.** Its subject overlaps my FR4-3. It is still residual, because the goal holds transitively through forest equality.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR5-1 | **blocking** | 01a21463-3262, 01a21447-d199 (r4-fresh3 FR4-1 promoted into it) | 3262 makes `recognize.New(tab, Options)` the judged constructor. d199 makes the judged constructor's table-generation error come back from `Recognize` "for every input". 3262 adds `TestParseRefusesInfeasibleGrammar`: "Parse with the judged default refuses every input on each grammar prectab.Feasible refuses, with the generation error". Its "Not expected to move" class still holds "the recognizer-internal tests over toy grammars without levels". | **A toy's precedence is not in its `Tables`.** Its levels and `%prec` come from `loadToy` and `pinnedToy` through `NewPrecedence(tab, levels, rulePrec, nil)`. The only derivation from `tab` alone is `buildParsePrecedence`, which errors for every tables value that is not the committed grammar's. `New(tab, Options)` and `Parse(tab, src, scope)` have no other way to receive it. So the judged default, given a toy, can only do one of two things:<br>**(a) Error.** Then every `New(toyTables...)` call in 11 `recognize` test files goes red at 3262, against the "Not expected to move" class. That is r4-fresh3 FR4-1, and the run-first step turns it into a stop.<br>**(b) Judge nothing.** Then `Parse` over a non-committed tables value accepts with no precedence. V1 refuses every recognized file there with the generation error. This is the off-switch my FR4-1 closed, chosen this time by the tables value instead of by feasibility.<br>**Either way, 3262's new criterion is defective.** `Parse` over emptyred, nullableShift, MarkFirstFilter, empty or scancollide can never see the toy's levels, so it cannot refuse with `Feasible`'s refusal. Read instead as "refuses every input with some generation error", it can already pass for the wrong reason: V1's `Parse` already refuses every recognized input on toy tables, with "not the committed grammar's tables". | `forest/readings.go:116-133` (`buildParsePrecedence`: `len(tab.Rules) != len(g.G1)` returns "not the committed grammar's tables"). `grammar/lr0.go:58` (`Tables` carries no levels or `%prec`). `forest/precedence_test.go:40-70` and `precedence_pinned_test.go:17-25` (toy precedence built outside `Tables`). `toyTables`, `toyStmt` and `buildToy` are used in 11 `recognize/*_test.go` files: agree, cache, fuzz, lattice, leo, leodisagree, recognize, reference, setids, shared and slipper. | **State the rule once, in d199, and mirror it in 3262.**<br>**(1) The judged constructor that tests use takes the precedence explicitly**, for example `NewJudged(tab, *prectab.Precedence, Options)`. `Feasible` refuses through that constructor, which is what d199's `TestJudgedRecognizerRefusesInfeasibleGrammar` already needs.<br>**(2) `New(tab, Options)` derives precedence the way `ParsePrecedence` does.** For tables that are not the committed grammar's, it judges nothing, so the toy recognizer tests do not move. For the committed grammar's tables, a real generation error is returned unwrapped.<br>**(3) `parseStream` keeps refusing every file with the derivation's error**, as it does today, so `Parse` over non-committed tables keeps V1's answer and no off-switch opens.<br>**Then replace 3262's `TestParseRefusesInfeasibleGrammar`.** Over the toys, use a test that the judged constructor given each toy's `Precedence` refuses with `Feasible`'s message, and that `Parse` over the same toy tables still refuses every recognized input with the derivation's error, as at e3432c0f. Name the toy-tables recognizer tests in 3262's "Not expected to move" under rule (2). |

## Checked and found sound this round

- **Edges.** No dependency edge changed, and the critical chain is unchanged.
- **Records.** No mapping row moved, and the added Records entry is appended under the others.
- **Carried ids.** No `Carries:` line changed, so nothing was lost in a split; there were no splits.
- **New criteria are red today.** The names `TestJudgedRecognizerRefusesInfeasibleGrammar`, `TestExpectedOnInfeasibleGrammarIsNoPrecedence`, `TestRefusalRerunScopeThenUnmappedIsLexUnmapped` and `TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped` appear nowhere in the tree. The fixture bytes for FR3-4's two cases come from the repairer's scratch measurement. I did not re-measure them.
- **3282's FR3-7 edit** limits the judged differential to the toys `Feasible` passes, which matches d218. Under FR5-1 rule (1), that differential gets each toy's `Precedence` through the explicit constructor. As written, it inherits the same unstated mechanism.

## 5. Not re-examined

- I did not re-measure the FR3-4 fixtures (bytes 16 and 19).
- I did not check d1eb's new criterion beyond its dependence on FR5-1.
- I did not re-read d02f to d10d, d135, d218, d247, d278, d2a2, d2d1 or bb9b. Their bodies are byte-identical to round 4's, which I reviewed then.
- I ran no V2 code. The re-run of my two bound tests is the only test execution in this round.
