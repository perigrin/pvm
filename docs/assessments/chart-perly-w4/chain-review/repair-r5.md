# chain-review round 5: refinement repair on chart-perly-w4

## Scope

This repair answers one blocking finding: **r4-fresh3 FR4-1**. It was promoted by alignment, pushback, fresh2 and fresh3, and raised again as r5-fresh4 FR5-1 and r5-fresh5 FR5-1. The repair also covers fresh2's residual FR5-1, that the tests must compare the error text, because constraint 5 requires the same thing.

No other residual finding is answered.

- All edits went through the `git zhi` CLI (git-zhi 0.5.2).
- The repository, git config and branches were not touched.
- No issue, edge, title or label changed. There are still 19 issues.
- 3262 gained no new criterion, which respects FR5-5. The interface lives in d199, and 3262 and d218 mirror it.

## The interface, stated in d199

**The gap.** `grammar.Tables` carries no levels. `ParsePrecedence`'s derivation, which d135 moves to `prectab`, errors on any tables value that is not perly.y's. The judged constructor therefore has two forms.

**Explicit form.** The judged constructor takes a `*prectab.Precedence` beside the tables. A toy grammar with levels reaches the judgement only this way, over its own levels:
- the bison toys and the pinned toys;
- the failing toys of the `Feasible` refusal;
- the judged differentials' toys;
- `FuzzRecognize`'s precedence toy.

`Feasible` runs on the precedence it is given, and its refusal is the error.

**Derived form.** This is the judged constructor with no precedence, which `recognize.New` becomes at 3262. It decides **by content**, not by pointer identity or provenance:
- **The rules are perly.y's, rule for rule.** That means the same rule count, and each rule's LHS and RHS symbol names equal to `grammar.Committed()`'s G1. This extends the count-and-LHS check that `buildParsePrecedence` makes today. The derived form derives perly.y's precedence through `prectab`, once per tables value, and judges with it. A copy, or a set of tables regenerated from `grammar.Committed()`, is judged the same as the committed value.
- **The derivation fails** on perly.y's tables, including when `Feasible` refuses. `Recognize` then returns the error for every input. It never judges nothing.
- **The rules are not perly.y's.** The tables declare no levels, so the judgement is empty because of that fact, not as a fallback. Nothing is forbidden, so sets, `Expected`, Leo and the cache are exactly the no-precedence recognizer's.

**Why content, not provenance.** `grammar.Tables` has no level or origin field. `grammar.Derived` already gives a copy its own derived slot, and adding a field would change the generated tables and `tables-drift`. Content is what the code supports, and `buildParsePrecedence` already uses it.

## The five constraints

| # | Constraint | Where it is met | Criterion (red today) |
|---|---|---|---|
| 1 | A level-less toy grammar recognizes exactly as before. | d199, derived form | d199 `TestJudgedOverLevellessTablesRecognizesAsBefore`, with subtests `recognize-toys` and `fuzz-seeds`. The judged constructor with no precedence gives no error and the same sets, `Expected` and acceptance as the no-precedence constructor. The name avoids the existing `TestNewOverAnyTablesValue` in `shared_test.go`, which passes today. |
| 2 | A toy with levels gets its precedence explicitly. | d199, explicit form | Exercised by d199's infeasible-grammar criterion and 3262's `ParseWith` criterion, both over each toy's own precedence. No separate criterion was asked for. |
| 3 | No off-switch keyed on the tables value. | d199, derived form decides by content | d199 `TestJudgedJudgesCopyOfCommittedTables`, with subtests `copy` and `regenerated`. Each is judged with perly.y's precedence: its sets on the perly.y sample equal the committed value's judged sets and differ from the no-precedence sets. |
| 4 | A failed derivation on perly.y's tables refuses. | d199, derived form | d199 `TestJudgedRefusesFailedDerivationOnPerlyTables`. The derivation is forced to fail through its `grammar.Derived` hook, the way `TestPrecedenceGenerationErrorRefusesEveryFile` does it, and `Recognize` returns the error for every input. |
| 5 | The `Feasible`-refusal tests assert `Feasible`'s own text. | d199, 3262 | d199 `TestJudgedRecognizerRefusesInfeasibleGrammar` and 3262 `TestParseRefusesInfeasibleGrammar` both build the judged recognizer over each toy's own precedence. They compare the refusal as text with `prectab.Feasible`'s error. Each criterion also fails if the output contains "not the committed grammar". 3262's test now uses `ParseWith` with that recognizer, because `Parse`'s `New` derives perly.y's precedence and cannot know a toy's levels. |

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21447-d199 | **Prose:** added the precedence-source interface (explicit and derived forms, content rule, failure and empty cases). The infeasible-grammar paragraph now says the refusal goes through the explicit form and carries `Feasible`'s own text.<br>**Paths:** gains `internal/chart/prectab/`.<br>**Steps:** added a precedence-source TDD step.<br>**Criteria:** the `TestJudgedRecognizerRefusesInfeasibleGrammar` criterion is rewritten (explicit precedence, text compared, "not the committed grammar" excluded). Three criteria added, for constraints 1, 3 and 4. | r4-fresh3 FR4-1 (constraints 1–5) |
| 01a21463-3262 | **Mirror:** `recognize.New` becomes d199's derived form, with its three outcomes.<br>**Prose:** the infeasible-toy refusal goes through `ParseWith` with explicit precedence.<br>**List note:** the "Not expected to move" list says it was built from the 25 qualified call sites, and that the roughly 113 unqualified `New(` calls in package `recognize` are covered by constraint 1's test.<br>**Criterion:** `TestParseRefusesInfeasibleGrammar` is rewritten (`ParseWith`, explicit precedence, text compared). | r4-fresh3 FR4-1 (constraint 5, mirror, 3262 list) |
| 01a21447-d218 | **Mirror:** the fate of `TestPrecedenceGenerationErrorRefusesEveryFile` now names the derived form's failed derivation on perly.y's tables, pinned by `TestJudgedRefusesFailedDerivationOnPerlyTables`. | r4-fresh3 FR4-1 (mirror) |
| Milestone Records | Added "Chain-review round 5 repair". No mapping row changed. | r4-fresh3 FR4-1 |

Every other line of each edited body is byte-identical to the live body read at the start of this round. The three stored bodies and the milestone body were re-read, and each matches its source file. Labels and the resolution command are unchanged.

## Verification

- **New and changed criteria are red.** All 5 were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and all 5 exit 1.
- **The named tests exist.** `TestNewOverAnyTablesValue` exists today, so I did not reuse its name. `buildParsePrecedence`'s current checks are at `forest/readings.go:116-133`. `grammar.Tables` has no level field (`grammar/lr0.go:58`).
- **The lint passes.** `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.

## Unchanged

The critical chain is still 15 issues:

d02f → d05d → d089 → d0ba → d0e1 → d135 → d166 → d199 → d1c2 → 3282 → 3262 → d218 → d247 → d2d1 → bb9b
