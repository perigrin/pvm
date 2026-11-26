# chain-review round 4: refinement repair on chart-perly-w4

## Scope

This round answers the three blocking findings that were still open:
- FR3-4, promoted by alignment;
- FR3-7, promoted by pushback;
- r4-fresh4 FR4-1.

Residual findings are left as they are: r4-fresh3 FR4-1, r4-fresh4 FR4-2 to FR4-4, and the rest. All edits went through the `git zhi` CLI (git-zhi 0.5.2). The repository, git config and branches were not touched. No issue was added or split, and no edge, title or label changed. The milestone still has 19 issues.

## How each finding was answered

| Finding | Answer | Issues |
|---|---|---|
| FR3-4 | **The new rule.** `lex-unmapped` is decided by V1's rule, applied to the run V1 actually made: the un-narrowed, no-precedence rerun. If that rerun's death byte is the first unmapped token's (`at == unmapped`), the reason is `lex-unmapped`; otherwise it is `grammar`. When the first chart itself dies at the first unmapped token, the un-narrowed run must die there too, so the reason is `lex-unmapped` with no rerun.<br>**Where it is written.** d166 states the rule for narrowing. 3262 adds the precedence branch.<br>**Criteria.** Two were added, each with a fixture measured at e3432c0f in a scratch copy:<br>• d166, `TestRefusalRerunScopeThenUnmappedIsLexUnmapped`: `1 plus 2;\nprint <<EOT;\nno terminator\n` must give `lex-unmapped` at byte 16. On its own, `1 plus 2;\n` is `scope` at byte 2, so this fixture really has a scope fault first.<br>• 3262, `TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped`: `1 .. 2 .. 3;\nprint <<EOT;\nno terminator\n` must give `lex-unmapped` at byte 19.<br>**Description change.** d166's `TestRefusalLexUnmappedNotRerun` criterion said a lex-unmapped file "is never rerun", which is no longer true in general. It now reads "a file whose narrowed chart dies at its first unmapped token … without a rerun". Its command is unchanged. | 01a21447-d166, 01a21463-3262 |
| FR3-7 | 3282's prose, step and `TestForestEqualityJudgedRecognizer` criterion now cover only the toys that `prectab.Feasible` passes. The five toys `Feasible` refuses are excepted and judged by `Constrain` over the no-precedence mode, as d218 states. The command is unchanged and still red today. | 01a21463-3282 |
| r4-fresh4 FR4-1 | **d199 owns the refusal.** Following design §4.1, the judged constructor's table generation returns `Feasible`'s refusal as an error, and `Recognize` returns that error, unwrapped, for every input. The no-precedence constructor still recognizes those grammars, which is the path the referees use.<br>**d199 criterion.** `TestRecognizerDoesNotJudgeInfeasibleGrammar` was replaced by `TestJudgedRecognizerRefusesInfeasibleGrammar`, with one subtest per toy: emptyred, nullableShift, MarkFirstFilter, empty and scancollide. A TDD step was added.<br>**d1eb criterion.** `TestExpectedUnjudgedForInfeasibleGrammar` was replaced by `TestExpectedOnInfeasibleGrammarIsNoPrecedence`, with one subtest per toy. On such a grammar there is no judged chart, because the judged constructor refuses the grammar; `Expected` on the no-precedence chart is V1's.<br>**3262.** 3262 now states that `parseStream` turns the generation error into a refusal of every file, so this is not a way round its `ParseWith` guard. It adds `TestParseRefusesInfeasibleGrammar`: `Parse`, with the judged default, refuses every input on each of the five toys with the generation error, in the shape of `TestPrecedenceGenerationErrorRefusesEveryFile`.<br>**Unchanged.** d218's referee path for these toys (`TestInfeasibleToysJudgedByConstrain`) and d135's `NewPrecedence` not calling `Feasible` stay as they are. | 01a21447-d199, 01a21447-d1eb, 01a21463-3262 |

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21447-d166 | **Prose:** "`lex-unmapped` is decided as today" became the rule, with its measured example.<br>**Criterion description:** `TestRefusalLexUnmappedNotRerun` reworded; command unchanged.<br>**Step:** the rerun-tests step now names the scope-then-unmapped case.<br>**Criterion added:** `TestRefusalRerunScopeThenUnmappedIsLexUnmapped`. | FR3-4 |
| 01a21463-3262 | **Prose:** the rerun rule's first branch now says `lex-unmapped` or `grammar`, with its measured example.<br>**Criterion added:** `TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped`. | FR3-4 |
| 01a21463-3262 | **Prose:** a sentence on infeasible grammars, added after the `ParseWith` guard.<br>**Criterion added:** `TestParseRefusesInfeasibleGrammar`, with 5 subtests. | r4-fresh4 FR4-1 |
| 01a21447-d199 | **Prose:** the infeasible-grammar refusal, and that d199 owns it.<br>**Step:** a new TDD step.<br>**Criterion replaced:** `TestRecognizerDoesNotJudgeInfeasibleGrammar` became `TestJudgedRecognizerRefusesInfeasibleGrammar`. | r4-fresh4 FR4-1 |
| 01a21447-d1eb | **Criterion replaced:** `TestExpectedUnjudgedForInfeasibleGrammar` became `TestExpectedOnInfeasibleGrammarIsNoPrecedence`. | r4-fresh4 FR4-1 |
| 01a21463-3282 | **Prose, step and criterion description:** limited to the toys `Feasible` passes, with the infeasible-toy exception. | FR3-7 |
| Milestone Records | Added the entry "Chain-review round 4 repair". No mapping row changed: no residual id moved between issues. | all three |

Every other line of each edited body is byte-identical to the live body read at the start of this round. That includes Negative Scenarios, the README-first steps, the round-3 first steps and the close steps. All 5 stored bodies and the milestone body were re-read and match their source files. Labels and the resolution command are unchanged.

## Verification

- All 7 new or changed criterion commands were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and all are red. 3282's command is unchanged, but it was re-run anyway.
- Each fixture's V1 reason was measured in the scratch copy at `repair/scratch/`, not in the repository:

  | Fixture | V1 refusal |
  |---|---|
  | `1 plus 2;\nprint <<EOT;\nno terminator\n` | lex-unmapped at 16 |
  | `1 .. 2 .. 3;\nprint <<EOT;\nno terminator\n` | lex-unmapped at 19 |
  | `1 plus 2;\n` | scope at 2 |

- All five toy names exist or have an owner: `emptyred` and `empty` are in `toyNames`, `nullableShiftToy` and the `TestPrecedenceMarkFirstFilter` toy are in `precedence_pinned_test.go`, and d05d creates `scancollideToy`.
- `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.

## Unchanged

- **Issue order and critical chain.** The critical chain is 15 issues long: d02f → d05d → d089 → d0ba → d0e1 → d135 → d166 → d199 → d1c2 → 3282 → 3262 → d218 → d247 → d2d1 → bb9b.
- **Open question (r4-fresh3 FR4-1, residual).** The judged `New` must not raise `ParsePrecedence`'s "not the committed grammar's tables" error for level-less toy tables. This round's text addresses only grammars that fail `Feasible`, and leaves the non-committed-tables case to that residual.
