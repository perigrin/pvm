# chain-review round 6: review-fresh4

I read the live chain (19 issues) from `/home/user/pvm/.claude/worktrees/v2-refine` into `fr6/`, and diffed it against my round-5 copy (`fr6/diff.txt`). Only then did I read `repair-r5.md`.

What changed since round 5:
- The bodies of d199, d218 and 3262, and an appended entry in the milestone's Records.
- No edge, title or label changed, and no issue was added.
- No row of the residual mapping changed.

## 1. Recommendation: modify

The repair settles the toy-suite problem. The derived form judges perly.y's precedence by content, level-less toy tables recognize as before, a failed derivation on perly.y's tables refuses every input, and the `Feasible` refusal tests compare `Feasible`'s own text.

I was asked to hunt for an off-switch, and I found one that the repair created. FR6-1 is blocking. 3262's criterion now requires `ParseWith` to accept a judged recognizer built in the **explicit form**, with whatever `*prectab.Precedence` the caller supplies. `ParseWith` is exported, and 3262's guard panics only on the no-precedence constructor. So a caller holding perly.y's tables can pass a judged recognizer whose precedence forbids nothing. After d218, `ParseWith` then parses perly.y with no precedence applied. That is an off-switch, which the ruling forbids.

The fix is small, and it needs no change to d199's explicit form.

## 2. Status of my findings

| Finding | Status | Grounds |
|---|---|---|
| r4-fresh4 FR4-1 (narrowed: `Parse` over an infeasible grammar) | **released** | **Infeasible grammars.** d199's `TestJudgedRecognizerRefusesInfeasibleGrammar` now builds the judged recognizer over each failing toy's own precedence. It compares the refusal as text with `prectab.Feasible`'s error and fails on "not the committed grammar". So an infeasible grammar is refused and never recognized unjudged.<br>**The `Parse` half.** I held it because `Parse` cannot see a toy's levels. 3262 now moves that half onto `ParseWith` with the explicit form, so it is no longer unpassable.<br>**Not vacuous.** It cannot pass vacuously: every new test name is absent from the tree, and the text comparison excludes the error V1 gives today.<br>The remaining off-switch concern is not about feasibility. It moves to FR6-1. |
| r5-fresh4 FR5-1 (the judged `New` has no precedence source for toy tables) | **released**, with one part carried to FR6-2 as residual | My suggested rules (1) and (2) landed in d199:<br>• an explicit form for toys with levels;<br>• a derived form that uses perly.y's precedence for perly.y's tables, judged by content, and judges nothing for other tables.<br>3262's "Not expected to move" class now accounts for the roughly 113 unqualified `New(` calls, and `TestJudgedOverLevellessTablesRecognizesAsBefore` covers them. A toy suite no longer goes red at the switch.<br>I checked the mutated-tables helpers (`without`, `withExpected`, `copyTables`, `addSet`, `ownTables`):<br>• the first four mutate toy tables, so their tests stay level-less, unchanged;<br>• `ownTables` and `TestNewOverAnyTablesValue` copy perly.y's tables, so under the content rule both sides are judged alike.<br>My rule (3), that `Parse` over non-perly.y tables keeps refusing as V1 does, was not adopted. That is FR6-2, which I rate residual. |
| r4-fresh4 FR4-2 (bound tests near their limits) | **held, residual** | Re-run today and unchanged: 0.6362 against 0.64, and 0.0395 against 0.040. My grounds from round 5 stand: the direction is unmeasured, it is a cost bound, and 3262's run-first step sends it to the orchestrator and perigrin. |
| r4-fresh4 FR4-3 (`make test` never compiles the `precedencereference` test) | **held, residual** | Unchanged. It overlaps r5-fresh5 FR5-3: d218 still does not name which reference differentials it repoints. |
| r4-fresh4 FR4-4 (no run-first step on d1c2 and d1eb; unkeyed `Item` literals) | **held, residual** | Unchanged. |

## 3. Promotions

I promote none. I read every residual row from r1 to r5, including these from round 5:
- r5-fresh2 FR5-1, r5-fresh3 FR5-1 and r5-pushback N12. The rewritten 3262 criterion answers them.
- r5-fresh5 FR5-2 to FR5-5, r5-pushback N13, and the earlier rows I weighed in round 5.

**r5-fresh5 FR5-3 and FR5-4** (d218 does not name which reference differentials it repoints, and the scope of the toy hook). Both are real ambiguities. In each, the failure mode is a test that goes red, which the builder can see; nothing passes silently. So both stay residual. They should land with my FR4-3 if d218 is edited for FR6-1.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR6-1 | **blocking** | 01a21463-3262 (with d199 and d218) | 3262: "No argument or option of `Parse` or `ParseWith` turns precedence off (F3-7), **and no recognizer passed to `ParseWith` does either**." The guard is "`ParseWith` panics when it is handed a no-precedence recognizer". The rewritten `TestParseRefusesInfeasibleGrammar` *requires* `ParseWith` to accept "a judged recognizer over each failing toy's own precedence". d199 defines that explicit form: "The judged constructor takes a `*prectab.Precedence` beside the tables". | The explicit constructor must be exported, because `forest`'s tests build it from package `recognize`, so any caller can reach it. Through the explicit form, `ParseWith(perlyTab, judged(perlyTab, p))` judges with whatever `p` the caller built. `prectab.NewPrecedence(perlyTab, nil, …)` gives a precedence that forbids nothing. After d218 removes `Constrain`, the call parses perly.y input with no precedence and no error. The no-precedence guard does not fire, because the recognizer is "judged". `TestParseHasNoPrecedenceSwitch` still passes, because its three subtests check parameters and option fields, the no-precedence constructor and the default, and none of them checks the explicit form. So the criterion passes without the property it names. The ruling "Guards: Regenerate everything" says "There is no precedence off-switch". 3262 extends that ruling to every recognizer handed to `ParseWith`. | `forest/forest.go:104-118`: `ParseWith` is exported, takes any `*recognize.Recognizer`, and is documented for callers that keep a recognizer across files. Before d218, `ParseWith` always judged with `ParsePrecedence(tab)`, derived from the tables, whatever recognizer it was given. After d218, the judgement is whatever the recognizer carries. The `forest` tests are in `package forest` (`forest_test.go:4`), so they can call `parseStream` directly. | **In 3262**, make the `ParseWith` guard accept a recognizer only when its precedence is the derived form's precedence for `tab`, the same `prectab` value `New(tab)` would use. Otherwise it panics, as it does for the no-precedence constructor.<br>**Move `TestParseRefusesInfeasibleGrammar`** off `ParseWith`, onto `parseStream` inside package `forest`, which is a test in the same package and not a public path. Or drop it as duplicate: d199's `TestJudgedRecognizerRefusesInfeasibleGrammar` already pins the refusal text at `Recognize`.<br>**Add a subtest** to `TestParseHasNoPrecedenceSwitch`, `parsewith-refuses-explicit-precedence`: `ParseWith` panics on an explicit-form recognizer whose precedence is not the derived one for `tab`. |
| FR6-2 | residual | 01a21447-d199, 3262, d218 | d199 says that for tables "whose rules are not perly.y's, the tables declare no levels and the judgement is empty by that fact, not by fallback". | Tables never declare levels: `grammar.Tables` has no level field, so a toy *with* levels and a level-less toy look the same to `New`. At `e3432c0f`, `Parse` over any non-perly.y tables refuses every recognized file with "not the committed grammar's tables". After d218, the same call accepts with no precedence. That includes a toy with levels, and tables equal to perly.y in count and LHS but differing in one RHS, which V1 judged with perly.y's precedence and the stricter content rule now drops. The ruling's off-switch concern is perly.y input, and only tests reach `Parse` with such tables. Production always builds its tables from the same `grammar.Committed()`, and the copy and regenerated subtests pin the content match. So this is a changed answer on an input class nothing uses, not a stop. | `forest/readings.go:116-133`; `grammar/lr0.go:58`; `grammar/refuse.go:30` (production tables come from the embedded grammar, so they always match). | My round-5 rule (3): let the recognizer judge nothing over non-perly.y tables, which keeps the toy tests unmoved, but have `parseStream` keep V1's refusal of every file with the derivation's error when `tab` is not perly.y's. Then "judge nothing" is never a `Parse` outcome. |

## Checked and found sound

- **New criteria are red.** All five new or changed test names are absent from the tree.
- **The content rule fails loudly, not silently.** If the comparison misread perly.y's committed tables, the derived form would judge nothing in production. `TestJudgedJudgesCopyOfCommittedTables` would then fail its "different from the no-precedence constructor's" clause, and so would 3262's `TestDefaultRecognizerJudgesPrecedence`. So a too-strict RHS comparison shows up as a red test, not as a silent off-switch on production.
- **Precedence tests that use committed-shape tables stay on perly.y's path.** `TestPrecedenceGenerationErrorRefusesEveryFile` uses `grammar.Generate(grammar.Committed())`, which matches by content. So the derived form takes the forced error through its hook, as d218's fate now says.
- **Toy-table helpers.** Those in `recognize` mutate only toys, or copy perly.y's tables, so none of them crosses the content rule unexpectedly.
- **Unchanged structure.** Edges, the residual mapping and the `Carries:` lines are unchanged.

## 5. Not re-examined

- Issues other than d199, d218 and 3262. Their bodies are byte-identical to round 5's.
- How `grammar.Committed()`'s G1 represents its RHS, against `Tables.Rules`. The comparison is the builder's. I relied on the copy and regenerated criteria to catch a wrong one.
- I ran no V2 code. The only tests I ran were my two bound tests.
