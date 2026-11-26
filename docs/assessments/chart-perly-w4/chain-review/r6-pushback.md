# chain-review round 6: pushback (plan quality) on chart-perly-w4

## Drafter: growth check

I diffed the live chain against my round-5 copy (`pb5/`) before reading `repair-r5.md`. Three issue bodies changed: d199 (22 lines), 3262 (6) and d218 (2). The milestone gained one Records entry, "Chain-review round 5 repair". No edge, title or label changed, and there are still 19 issues.

| Edit | Answers |
|---|---|
| d199: `paths:` gains `prectab/`; the precedence-source interface (explicit and derived forms); the infeasible-grammar paragraph routed through the explicit form; a TDD step; the `TestJudgedRecognizerRefusesInfeasibleGrammar` criterion rewritten to compare text and to reject "not the committed grammar"; three new criteria (`TestJudgedOverLevellessTablesRecognizesAsBefore`, `TestJudgedJudgesCopyOfCommittedTables`, `TestJudgedRefusesFailedDerivationOnPerlyTables`) | r4-fresh3 FR4-1 (blocking; my promotion), which subsumes r5-fresh4 FR5-1 and r5-fresh5 FR5-1 |
| 3262: `New` becomes the derived form; the infeasible refusal goes through `ParseWith` with explicit precedence; a note on the "Not expected to move" list; `TestParseRefusesInfeasibleGrammar` rewritten with the same text check | FR4-1 (mirror). The text check also answers fresh2/fresh3's residual FR5-1 and my N12, as `repair-r5.md` says. |
| d218: the fate of `TestPrecedenceGenerationErrorRefusesEveryFile` names the derived form's failed derivation | FR4-1 (mirror) |
| Milestone Records | FR4-1 |

**Verdict.** Every edit answers FR4-1 or mirrors it. Folding the residual text check (FR5-1, N12) into the same criteria added no new criterion; it tightened two that already existed. So nothing grew beyond its finding.

The repair adds three criteria to d199, all of them red today. `criteria-lint.sh` exits 0 over all 19 issues.

## 1. Recommendation

**Modify.** My promotion of r4-fresh3 FR4-1 is released. I verified the content rule against the tree: perly.y's tables match `grammar.Committed()` rule for rule, including right-hand sides. The rule leaves no off-switch on perly.y.

Hunting for "a toy suite it still breaks", as the brief asked, I found one: **N14, blocking.**

- The derived form judges nothing on any tables that are not perly.y's, including the 13 levelled toys in `testdata/precedence/toys/`.
- d218 then tells the production differentials over "the toys" to compare "production", meaning the recognizer's forest with no `Constrain`, against V1.
- Production's constructor over a toy is now, by d199's own rule, unjudged. So the toy differential would go red on correct code, and that red is a false §7 stop.

The fix is one sentence in d218.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1–PB7, PB9–PB11, PB16, PB17 | **released** | Unaffected. On PB17: the derived form does not reopen it. The `ParseWith` guard still rejects the no-precedence constructor. On perly.y's tables, the derived form either judges or refuses every input, and it never judges nothing (`TestJudgedRefusesFailedDerivationOnPerlyTables`). |
| PB8, PB12–PB15, PB18–PB20 | residual | Unchanged. |
| N1, N5, N6 | **released** | Unchanged. |
| N2–N4, N7–N11, N13 | residual | Unchanged. |
| N12 | **released** | 3262's `TestParseRefusesInfeasibleGrammar` now builds a judged recognizer over each toy's own precedence through `ParseWith`, compares the refusal as text with `prectab.Feasible`'s error, and its criterion fails if the output contains "not the committed grammar". It can no longer go green on the wrong error. |
| FR2-8, FR3-7 (my promotions) | **released** | Released in earlier rounds; d2d1 and 3282 are unchanged. |
| **r4-fresh3 FR4-1** (my promotion) | **released** | Settled point by point below. |

**FR4-1: what the fix had to decide, and where it landed.**

1. *How the judged constructor receives a grammar's precedence.* This is now the **explicit form**: a `*prectab.Precedence` beside the tables. The toys that need it are listed: bison, pinned, `Feasible`-failing, the judged differentials' toys, and the fuzz toy.
2. *What `New` does on tables that are not perly.y's.* It judges nothing, "by fact", because tables carry no levels. `TestJudgedOverLevellessTablesRecognizesAsBefore` pins this over the `recognize` toys and the fuzz seeds, so the ~113 unqualified `New(` calls stay unmoved. That answers my point (c).
3. *Which error `Parse` returns.* On perly.y's tables, a failed derivation is returned for every input (`TestJudgedRefusesFailedDerivationOnPerlyTables`, mirrored in d218's fate for `TestPrecedenceGenerationErrorRefusesEveryFile`). On a failing toy, the error goes through `ParseWith` and the explicit form, compared as text with `Feasible`'s error.

The d199 criterion from my point (b), `TestJudgedRecognizerRefusesInfeasibleGrammar`, is now passable: it uses the explicit form and rejects "not the committed grammar".

**Re-run (the flagged call).** The content match compares right-hand sides, which is stricter than today's count-and-LHS check. I measured it against the tree through a `go test -overlay` test file, so the working tree was untouched (`git status` clean). Result: `G1=350 tables=350 seq=0 rhs-mismatch=0 refused=4`.

Generated perly.y tables match `grammar.Committed()`'s G1 rule for rule, RHS names included. There are no sequence rules, whose BNF expansion would have broken the one-to-one match. So the stricter check judges the real tables, and the check that would catch a mismatch still exists: 3262's `TestDefaultRecognizerJudgesPrecedence` and d199's `copy` and `regenerated` subtests would each go red if perly.y silently stopped being judged.

**Off-switch hunt.** On perly.y there is no off-switch:
- `Parse` and `ParseWith` are only ever called with perly.y tables: every caller in `forest/*_test.go`, `cmd/ratchet` and production passes committed or regenerated tables.
- Levels come from `grammar.Committed()`, not from the tables, so no tables value can carry different levels.
- A failed derivation refuses, never judges nothing.

The judge-nothing branch does exist for non-perly tables. But no production path reaches it, and V1 had no precedence for such tables either: it refused them with an error. That one behaviour change is N15 (residual).

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| r5-fresh5 FR5-2 (3262's restated rule omits d166's no-rerun shortcut for lex-unmapped) | not promoted | An always-rerun build gives the same reason at the same byte: the un-narrowed run dies at the first unmapped token too. Only `TestRefusalLexUnmappedNotRerun`'s "never rerun" assertion would move, and 3262's first step reports that with its failure text. Residual. |
| r5-fresh5 FR5-3 (`TestWaypoint2RefusalsMatchReference` is pinned at 3262; d218 "repoints the production-versus-reference differentials") | not promoted | The two are reconcilable: the waypoint2 Constrain-against-reference check stays a referee on the no-precedence forest, and d218's repoint concerns the forest-equality differentials. d218's first step surfaces a misreading. Residual, but N14's fix could name both in one sentence. |
| r5-fresh5 FR5-4 (d218's hook check against the infeasible toys) | not promoted | d218 already excepts those toys in prose and criterion. Residual. |
| r5-fresh5 FR5-5 (3262 size) | not promoted | The repair put the interface in d199, not 3262, as FR5-5 asked. 3262 gained no criterion. |
| All earlier residual rows | not promoted | Unchanged, and none makes proceeding wrong. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N14 | **blocking** | d218 (with d199, d0ba) | d218: "From here the production differentials (toys, seeds, the union shortcut, waypoint2 and the corpus) compare production with V1 for real … with 'production' now the recognizer's forest, except every toy `prectab.Feasible` refuses". `TestForestEqualityCandidateNeverConstrains` requires the toy candidate never to enter `Constrain`, and "Every production differential is green … over toys" requires it to equal V1. | d199's derived form, which is what `recognize.New` and so "production" become at 3262, judges nothing on tables whose rules are not perly.y's. `testdata/precedence/toys/` holds 13 toys with levels, and `Parse` cannot reach their levels. A toy candidate built the production way at d218 is therefore unjudged, has no `Constrain`, and keeps every alternative V1's `Constrain` drops. `TestForestEqualityToys` then reports "a forest differs from V1's", which is a §7 stop on correct V2 code. Only the explicit form over the toy's own precedence gives the right candidate. d199's explicit-form list names "the judged differentials' toys" (3282's), not d218's production differentials, and d218 does not say it. This is a knock-on of the round-5 repair: before it, `New` over a toy was undefined; now it is defined as unjudged. | d199's derived-form paragraph ("If the rules are not perly.y's … the judgement is empty"); d218's Context and criteria 4 and 6; `ls testdata/precedence/toys` (13 files with precedence declarations) | One sentence in d218, and the same one added to d199's explicit-form list: "Over toys, the production differential's candidate is the judged recognizer in the explicit form over the toy's own precedence, then Build, with no `Constrain`. `Parse` has no toy precedence, so 'production' over a toy means that pipeline." |
| N15 | residual | d199, 3262 | The derived form over non-perly tables "judges nothing, by fact" | V1's `Parse` over toy tables refused every input with "not the committed grammar's tables". V2's `Parse` over the same tables accepts them unjudged. No test or production caller relies on the V1 refusal; I checked every `Parse(` call in `forest/*_test.go`, and all pass perly.y tables. So nothing breaks. But it is a stated behaviour change that no issue names. | `forest/readings.go:121`; the grep above | State it in 3262's "Rewritten to V2's answer" list, or keep `parseStream`'s refusal for tables that are not perly.y's. |

## 5. What I did not re-examine

- I ran no V2 code. N14 is reasoned from d199's stated rule and d218's text. The only measurement this round is the rule-for-rule overlay test.
- I did not open d0ba's eventual toy-loading code, which does not exist yet. N14 assumes that the toys are the levelled `testdata/precedence/toys` set that `loadToy` reads, as d0ba's "toy forest" criteria and `precedence_reference_test.go` imply.
- I did not re-read the 16 unchanged bodies beyond confirming they are byte-identical to round 5.
- I did not re-check the residual mapping. The repairer reports no row moved, and the milestone diff shows only a Records entry.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
