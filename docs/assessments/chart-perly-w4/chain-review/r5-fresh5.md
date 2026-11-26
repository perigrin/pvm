# chart-perly-w4 chain-review, round 5: fresh reader (review-fresh5)

**Recommendation: modify.** I found one blocking finding and four residuals.

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`, and it has 19 issues. I also read the design (revision 9), the rulings, the minute's "Residual rows for refinement", and the code each issue touches. I did not read any prior findings, any `r1-*` to `r5-*` file, any `repair-*` file, or the Records' chain-review notes. My dumps are in `review/fr5x/`.

## Findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR5-1 | **blocking** | 3262 (also d135, d199, d278) | After the switch, `recognize.New(tab, opts)` is the judged recognizer. It "reads `prectab`'s tables, which d135 caches as a `grammar.Derived` value". A derivation error then refuses every input (d199: "Recognize returns that error, unwrapped, for every input"; 3262/d218: `TestPrecedenceGenerationErrorRefusesEveryFile`'s shape). In the same issue, 3262 says "the recognizer-internal tests over toy grammars without levels" are "not expected to move. If one of them moves, a forest differs from V1's". d278 also says the level-less fuzzed grammars "judge nothing". No issue says where a judged `New` gets precedence for a tables value that is not perly.y's. | The derivation that d135 moves, `buildParsePrecedence`, returns an error for any tables value that is not the committed grammar's: "not the committed grammar's tables". `grammar.Tables` carries no levels, so for a toy, judged `New` has nothing to derive from. That leaves two outcomes. **(1)** At 3262, every `New(toy)` gets the generation error. The recognizer suite, `FuzzRecognize`'s seeds, the slipper and Leo toy tests, and the forest toy differential all go red. The chain has already labelled that red a §7 stop ("a forest differs"). The intermediate state cannot close green. **(2)** The builder invents "tables that are not perly.y's mean no levels, so judge nothing". That is a precedence off-switch chosen by which tables value is passed, which d199 forbids for infeasible grammars and §4.4 forbids for callers. It is also indistinguishable from a perly.y derivation failure, which must refuse. The `ParseWith` guard checks only the constructor's identity, so a judged recognizer whose precedence came from elsewhere would pass it. This is precedence quietly not applied, and a behaviour change at 3262 that tests elsewhere do not expect. | `internal/chart/forest/readings.go` `buildParsePrecedence` (the rule-count and LHS checks against `grammar.Committed()`). `grammar/lr0.go` `Tables` has no level field. `recognize/lattice_test.go` `buildToy` builds an empty `grammar.Grammar{}`. In-package level-less `New(toy…)` sites include `shared_test.go` (`TestNewOverAnyTablesValue`, `TestCopyTablesBesideLaterNew`, `TestRecognizersOverDistinctTablesShareNothing`), `slipper_test.go` `agree`, `fuzz_test.go:127,464`, `cache_test.go:80,160,190`. In all, 113 unqualified `New(` calls sit in package `recognize`'s tests. 3262's list was "built by reading the 25 test call sites of `recognize.New`", which are the qualified ones, so none of these 113 were read. | In d135 or d199, before 3262, state the rule for where a judged recognizer's precedence comes from. Two options: (a) a tables value carries, or is given, its grammar's levels, and a grammar that declares none judges nothing by a fact rather than by fallback; or (b) 3262 pins every level-less toy site to the no-precedence constructor and lists them. Either way, add two criteria. The first: judged `New` over a level-less toy recognizes as before (for example, `TestNewOverAnyTablesValue` stays green). The second: judged `New` over a perly.y tables value whose derivation fails refuses with that error and never judges nothing. Also say whether `ParseWith` checks that the recognizer's precedence is `tab`'s own. |
| FR5-2 | residual | d166 → 3262 | d166 pins `TestRefusalLexUnmappedNotRerun`: lex-unmapped "without a rerun". 3262 restates the rule as "First the rerun is un-narrowed: if it dies at all…" and does not repeat the no-rerun shortcut. It also does not list that test among the tests that move. | If 3262 is built literally, so that it always reruns, a d166 criterion goes red at 3262 under a test the chain does not expect to move. | Text of d166 Negative Scenarios and 3262 Context. | In 3262, say that the "narrowed chart dies at the first unmapped token, so no rerun" shortcut stands. |
| FR5-3 | residual | 3262 / d218 | 3262 pins `TestWaypoint2RefusalsMatchReference` to the no-precedence mode, and its criterion's prose says so. d218 then says it will "repoint the existing production-versus-reference differentials at the recognizer's forest". That test errors on any row that has no recognized forest. Its `waypoint2ScopeRefusals` rows die in a narrowed chart. | If d218 repoints this test, it goes red. If d218 repoints none of the tests, design §5's "'Production' now means the recognizer's forest" has no criterion, for this test or for `TestPrecedenceCorpusMatchesReference` and `TestPrecedencePerlGitMatchesReference`. | `waypoint2_refusals_test.go:109-122`. `cmd/ratchet/precedence_reference_test.go` `comparePrecedence`. | Have d218 name which existing reference differentials it repoints and which stay on the no-precedence forest. |
| FR5-4 | residual | d218 | `TestForestEqualityCandidateNeverConstrains` asserts that the toy candidate never enters `Constrain`. The same issue keeps the infeasible toys "judged by `Constrain` over the no-precedence mode". | A builder who keeps the infeasible toys inside `TestForestEqualityToys` trips the hook. | d218 Context and Negative Scenarios. | Say that the hook check covers only the toys that `Feasible` passes. |
| FR5-5 | residual | 3262 | 3262 is large: 26 paths, the rerun rule, the readers' agreement, the `ParseWith` guard, five rewrites and one retirement. FR5-1's fix adds to it if it lands there. | It is a session-size risk. It is not a blocker on its own. | 3262 Context and Steps. | Land FR5-1's rule in d135 or d199, not in 3262. |

## Checked and clean

- **The two gap kinds, traced across every issue.** I checked whether each later issue could falsify a criterion of an earlier one:
  - the goldens, the `set-digest-noprecedence` copy, and the "only commits touching the golden" criterion;
  - the `refused-where.txt` byte identity;
  - the known-limit pins;
  - `TestJudgedRecognizerIsExplicit` after the switch;
  - the d1c2 collapse tests;
  - the union-shortcut fixture in d0ba against d3282, where I found no contradiction, because V1 keeps the union at a shared node;
  - the 27-key grep, which yields exactly 27 on `baseline.txt`;
  - the `baseline.txt` one-line diff;
  - Expected against the set digest, which 3262 pins to no-precedence;
  - the answer keys: they are computed from `Accepted`, the pipeline, not from `Recognized`, and `lex-unmapped` comes from `lex.Lex` alone, so the narrowing and judging recognition in `chartInputs` moves no gated key.
- **Off-switch paths.**
  - `Parse`, the speed side through `ParseWith`, `facts.go` and `chartInputs` all follow `recognize.New`.
  - The rerun, the referees and the pinned guards are the only no-precedence users.
  - The only open path is FR5-1's.
- **Cheap criteria today.** I ran all 56 non-`go test`/`make` criterion commands against the tree, and every one is red today.
- **Residual rows.** All 66 rows, plus the crumb, are carried in an issue's `Carries:` or declined in the body with a reason. The declined rows are A14, A15, A25's code half and P9.
- **Order.** The stage order holds (counters → differentials → recognizer). Every code issue precedes d2d1, and d2d1 precedes bb9b.

## Not examined

- I ran no `go test` criterion and no `make test`, so nothing in this report establishes which `go test` criteria are red or green today.
- I did not read the bodies of the forest toy tests (`precedence_statefree_test.go`, `precedence_pinned_test.go`) beyond their helpers.
- I did not check the speed-protocol driver or the `chain/` export process.
- The installed git-zhi's `verify --dry-run` reads only issues that are done, so it is empty today, and I relied on my own extraction of the 194 criterion commands.
