# chain-review chart-perly-w4, round 5: review-alignment (resumed)

## 1. Recommendation

**Modify.** FR3-4, the finding I held, is released. I measured the fix's fixtures at `e3432c0f`.

I promote r4-fresh3 FR4-1 to blocking, as the brief asked me to decide:
- The chain never says where the judged default, `recognize.New(tab, opts)`, gets its precedence.
- `Tables` carry no levels, so for any tables other than the committed grammar's there is nothing to judge with.
- Round 4's repair made a generation error propagate from `Recognize` for every input. That points the literal implementation at turning every toy-grammar recognizer test red at 3262.
- The same gap makes 3262's new criterion `TestParseRefusesInfeasibleGrammar` passable for a reason unrelated to `Feasible`.

The fix is a stated rule plus one rewritten criterion (section 3).

**Checks this round.** I read the chain live first: 19 issues at `e3432c0f`, edges taken from `blocked_by`. I read `repair-r4.md` only after running the checks.
- **Edges:** unchanged from rounds 3 and 4.
- **Criteria:** I re-ran all 194 against `e3432c0f`, each passed to `sh -c` unmodified. All 194 are red.
- **Residual mapping:** 66 rows plus the crumb. The only misses are the 5 from AL6.
- **Carries:** each issue's `Carries:` line is unchanged.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1–AL4, AL12, AL13, AL14, PB17 | released (earlier rounds) | I re-ran the criterion check: 194 of 194 are red.<br>The PB17 guard text in 3262 is unchanged.<br>3262's rule now also names `lex-unmapped` first, which is what AL1's rule needed. |
| FR3-4 (my r4 promotion) | **released** | **What the fix had to change as well:**<br>- The `lex-unmapped` test must be decided on the un-narrowed rerun's byte, in both d166 and 3262.<br>- The "never rerun" criterion must be narrowed.<br>- A two-fault fixture is needed for scope, and another for precedence.<br>- The rule's order must still put `lex-unmapped`/`grammar` before `scope` before `precedence`.<br>**Checked in the chain:**<br>- d166: "When the un-narrowed rerun dies, its death byte decides … `lex-unmapped` … otherwise `grammar`". It also keeps "narrowed chart dies at the first unmapped token → `lex-unmapped` with no rerun". That shortcut is sound, because the un-narrowed run is a superset and must die there too.<br>- 3262 states the same rule for the judged chart.<br>- `TestRefusalLexUnmappedNotRerun` now covers only that shortcut.<br>- New criteria: `TestRefusalRerunScopeThenUnmappedIsLexUnmapped` (d166) and `TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped` (3262). Both are red today.<br>**Measured at `e3432c0f`** (scratch copy, `forest/zz_al5_test.go`, V1's `Parse` with an empty scope):<br>- `1 plus 2;\nprint <<EOT;\nno terminator\n` → `{Byte:16 Reason:lex-unmapped}`<br>- `1 plus 2;\n` → `{Byte:2 Reason:scope}`<br>- `1 .. 2 .. 3;\nprint <<EOT;\nno terminator\n` → `{Byte:19 Reason:lex-unmapped}`<br>The issues' numbers are right, and both fixtures really have two faults.<br>**Bias check:** I wrote the bar. Against that bias, I ran the fixtures rather than read them. |
| AL5–AL11, AL15, AL16 | residual, held | Unchanged. |

**What the repair also answered, checked:**
- **FR3-7.** 3282 now restricts the judged differential to the toys `Feasible` passes, and says the others are judged by `Constrain` over the no-precedence mode. That matches design §5.
- **r4-fresh4 FR4-1.** In d199 the judged constructor refuses an infeasible grammar instead of recognizing it unjudged. That matches design §4.1, "Generating its tables refuses any other grammar", better than round 4's "sets equal the no-precedence mode's". d1eb's criterion follows.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| r4-fresh3 FR4-1 | residual | **blocking** (I hold it) | See the four parts below the table. |

**The gap.**
- `recognize.New(tab, opts)` takes only `*grammar.Tables`.
- `grammar.Tables` carries no precedence levels or `%prec` (`grammar/lr0.go:58`).
- Today precedence comes from outside the tables:
  - for perly.y, `buildParsePrecedence` reads `grammar.Committed()`, and it errors for any other tables: "not the committed grammar's tables" (`forest/readings.go:116-133`);
  - for bison toys, `loadToy` passes each toy's own `level` lines to `NewPrecedence`.
- From 3262 on, `New` judges, so it needs a source for precedence. No issue states one, and none says what `New` does with tables that are not the committed grammar's.

**Why it is blocking.** There are three consequences.

- **(a) A stop, predictable now.**
  - The likely implementation reuses the committed-grammar derivation.
  - Round 4's repair says in d199 that the judged constructor's generation error is returned by `Recognize` "for every input".
  - d218 says the generation error is "raised where the judged recognizer reads `prectab`'s tables".
  - Followed literally, every toy-grammar recognizer test errors at 3262. `recognize`'s tests have about 113 `New(` call sites, many on `toyStmt` or `toyTables`.
  - 3262 lists these tests as "Not expected to move".
  - Its run-first step would report them, and amending a started issue is a §7 stop.
- **(b) A criterion that passes for the wrong reason.**
  - 3262's `TestParseRefusesInfeasibleGrammar` says `Parse` "refuses every input on each grammar `prectab.Feasible` refuses, with the generation error".
  - `Parse(tab, src, scope)` cannot see a toy's levels, so it cannot run `Feasible` on emptyred and the others.
  - If `New` errors on non-committed tables, `Parse` over a failing toy refuses everything with the not-committed error. The criterion goes green, and the prose's "design §4.1" reason is false.
  - A feasible toy would be refused just the same.
- **(c) An interaction with PB17's guard.**
  - Suppose instead that `New` over toy tables "judges nothing".
  - Then whatever marks that recognizer must not be the no-precedence mark.
  - Otherwise `ParseWith`, and so `Parse` on toy tables, panics.

**Classification.** This is an unstated rule whose literal reading does the wrong thing. It also produces a criterion that passes before its claim is true. Both are defects in what the subject asserts.

**Fix, in d199 and mirrored in 3262 and d218:**

- **State the judged default's precedence source.**
  - **The committed grammar's tables:** `New` derives precedence from `grammar.Committed()`, cached once per tables value as d135 already says. A real generation error, `Feasible`'s included, is returned unwrapped by `Recognize`.
  - **Any other tables:** `New` judges nothing, because there is no precedence to judge. It still counts as the default for the `ParseWith` guard.
  - **A judged recognizer over a toy:** only the explicit judged constructor builds one, taking the toy's precedence as an argument.
  - This is r4-fresh3's own suggested sentence, extended for the guard.
- **Rewrite `TestParseRefusesInfeasibleGrammar`, or replace it** with a test that cannot pass on the not-committed error:
  - one option is a committed-shape `Derived` hook injecting `Feasible`'s refusal, as `TestPrecedenceGenerationErrorRefusesEveryFile` does;
  - the other is to rely on d199's judged-constructor test, which already covers the toys.
- **Add one control:** `New` over a level-less toy grammar recognizes exactly as the no-precedence constructor does.

**Rows read and not promoted.** I read every row in r1 to r4 from every participant.

- **r4-fresh4 FR4-2 (residual, high).** 3262's list was built from 25 call sites, where `recognize` alone has more than 100. FR4-1 above is the concrete form of that worry. Once the rule is stated, the run-first step handles the rest.
- **FR4-3.** `make test` does not compile the `precedencereference` tag. It is residual: those tests gain criteria of their own that run the `make` target.
- **FR4-4.** Not every behaviour-changing issue has the run-first step. That includes d1c2's change to `Item`. It is residual: d1c2 states that no test moves, and its criteria run the set digest.
- **r4-fresh2 FR4-1.** It names the repointed differentials only as a class. Residual.
- **N10 and N11 (pushback).** Residual.
- **PB8, PB12–PB15, PB18–PB20, N2, N3, N7, N9.** Unchanged.
- **AL5–AL11, AL15, AL16.** Unchanged.

## 4. New findings

None besides the promotion.

## 5. What I did not re-examine

- **Which toy-grammar tests actually go red under a judged `New`.** I did not build one. Consequence (a) rests on `buildParsePrecedence`'s error path, `Tables` having no levels, and the repair's "returned for every input" wording. The rest of the case is read from the code, not run.
- **Unchanged bodies.** d02f, d05d, d089, d0ba, d0e1, d10d, d135, d1c2, d218, d247, d278, d2a2, d2d1 and bb9b have no diff from round 4. I did not re-read them beyond the criterion re-run.
- **perl.git targets.** Their criteria are gated red today.
