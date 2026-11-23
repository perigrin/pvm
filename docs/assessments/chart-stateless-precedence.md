<!-- ABOUTME: Minute of the crochet:assess session on the chart parser's precedence design, now precedence without parser state. -->
<!-- ABOUTME: Covers rounds 1 to 8 and revisions 1 to 8 (renamed at 4), ending in non-convergence at the bound: the ledger, cross-references, declarations, rulings, outcomes, measurements and experiments. -->

## Assessment: Chart parser: precedence without parser state

Recommendation: accept, by perigrin's ruling at the bound (round 1: three modify; round 2: four modify; round 3: four modify; round 4: four modify; round 5: accept with edits from assess-architecture, modify from the other three; round 6: accept with edits from assess-architecture, modify from the other three; round 7: accept from assess-architecture and assess-codebase, assess-pushback holding nothing, modify from assess-fresh7; round 8: accept from assess-architecture, assess-codebase and assess-pushback, modify from assess-fresh8; no fixed point within the bound of eight; perigrin ruled "Apply fixes, accept (Recommended)", and revision 9 applies M1–M5)

- **Subject:** `docs/plans/2026-11-23-chart-stateless-precedence.md`. Revisions 1 to 3 were `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`, and this minute was `docs/assessments/chart-precedence-in-recognizer.md` until revision 4 renamed both. Earlier rows cite the old path as written.
- **Revisions:**

  | revision | commit | blob | path | title |
  |---|---|---|---|---|
  | 1 | 7b264fa9c | 52c82af1 | `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` | Chart parser: precedence in the recognizer |
  | 2 | c270bbfa3 | 68371487 | `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` | Chart parser: bison's reduce-side precedence in the recognizer |
  | 3 | 3e4d0c43b | a8960d05 | `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` | Chart parser: the precedence constraint without maps |
  | 4 | 6dc1d9558 | 123012c1 | `docs/plans/2026-11-23-chart-stateless-precedence.md` (renamed) | Chart parser: precedence without parser state |
  | 5 | 2782fd181 | d97415a1 | `docs/plans/2026-11-23-chart-stateless-precedence.md` | Chart parser: precedence without parser state |
  | 6 | 788778ed0 | 60327eb7 | `docs/plans/2026-11-23-chart-stateless-precedence.md` | Chart parser: precedence without parser state |
  | 7 | 9afff6c63 | ce616721 | `docs/plans/2026-11-23-chart-stateless-precedence.md` | Chart parser: precedence without parser state |
  | 8 | 336067853 | ed91e462 | `docs/plans/2026-11-23-chart-stateless-precedence.md` | Chart parser: precedence without parser state |

- **Revision assessed in round 1:** design commit 7b264fa9c, blob 52c82af16a738eca6e03cdb050e9c167903212c7. Each participant read that blob, untracked in worktree `cr-close`, and checked it with `git hash-object`. It is byte-identical to the blob at 7b264fa9c.
- **Revision assessed in round 2:** commit c270bbfa3, blob 68371487a68e3f8c1661a80ea830356228a3a163 (revision 2, retitled "bison's reduce-side precedence in the recognizer").
- **Revision assessed in round 3:** commit 3e4d0c43b, blob a8960d05a84284fb0f9fcec803d101949fd9c405 (revision 3, retitled "the precedence constraint without maps"). The file keeps its name (see A20).
- **Revision assessed in round 4:** commit 6dc1d9558, blob 123012c1a5765883ded36de157c8d0e8ac764f0a (revision 4, renamed to `docs/plans/2026-11-23-chart-stateless-precedence.md`).
- **Revision assessed in round 5:** commit 2782fd181, blob d97415a147b923df589b27bed3cad0317f6bbfe2 (revision 5: the forest-side judgement only, "V1 now"). The minute was at 157b0f820.
- **Revision assessed in round 6:** commit 788778ed0, blob 60327eb7d7e379443281e0623b5d5f586a382f9c (revision 6). The minute was at 0d2b31610.
- **Revision assessed in round 7:** commit 9afff6c63, blob ce616721c9ed717dc119cf4c5f3329daf802f4cd (revision 7). The minute was at 1ff6e3f6a.
- **Revision assessed in round 8:** commit 336067853, blob ed91e46239b29283ad5c82cd390411621b52a2cd (revision 8). The minute was at 146411199.
- **Decision-maker:** perigrin.
- **Drafter:** assess-pushback, from round 2 on. The drafter records and holds no view. In round 1 it was the pushback participant and raised P1 to P12. From round 2 on it does not count as an independent participant.
  - In round 2, at the coordinator's direction, it also acted as raiser of P1 to P12. As raiser it released, held or carried them, and it raised P13 to P15.
  - Those round-2 P rows are the raiser's, not the drafter's.
  - In round 3 it acted as raiser of P13 to P15 only, and raised P16.
  - In round 4 it acted as raiser of P16 and of P1, P2, P3 and P10, and raised P17.
  - In round 5 it acted as raiser of P17 and raised P18.
  - In round 6 it acted as raiser of P18 and raised P19.
  - In round 7 it acted as raiser of P19 and raised P20.
  - In round 8 it acted as raiser of P20 and raised nothing.
  - The independent participants are:
    - assess-codebase and assess-architecture, rounds 1 to 8;
    - assess-fresh, round 2 only;
    - assess-fresh3, round 3 only;
    - assess-fresh4, round 4 only;
    - assess-fresh5, round 5 only;
    - assess-fresh6, round 6 only;
    - assess-fresh7, round 7 only;
    - assess-fresh8, round 8.

Each row's section reference is to the revision named in its revision column. IDs keep each participant's own prefix:

- A: assess-architecture, the architecture and direction lens;
- C: assess-codebase, the codebase lens;
- P: assess-pushback, the spec-quality lens;
- F: assess-fresh, a fresh participant in round 2. It read the subject without the prior findings, and it wrote its report before reading this minute. It was not resumed in round 3, so its rows are marked raiser absent.
- G: assess-fresh3, a fresh participant in round 3. Of this minute it read only "Measurement after round 2". It was not resumed in round 4, so its rows are marked raiser absent.
- H: assess-fresh4, a fresh participant in round 4. Of this minute it read only "Experiments after round 3". It was not resumed in round 5, so its rows are marked raiser absent.
- J: assess-fresh5, a fresh participant in round 5. It read neither this minute nor the evidence directory's `reports/`. It was not resumed in round 6, so its rows are marked raiser absent. assess-architecture lists A32 as "not used", so there is no row A32.
- K: assess-fresh6, a fresh participant in round 6. It read neither this minute nor `reports/`. Its K1 mentions "the remapping in K10"; the remapping is described in its K9, and there is no K10. It was not resumed in round 7, so its rows are marked raiser absent.
- L: assess-fresh7, a fresh participant in round 7. It read neither this minute nor `reports/` before writing. It was not resumed in round 8, so its rows are marked raiser absent.
- M: assess-fresh8, a fresh participant in round 8. It read the minute and `reports/` only after writing.

"C1" in a ledger row is assess-codebase's finding. The chart-recovery spike is always written "the C1 spike".

The full reports and evidence are in the committed evidence directory, `docs/assessments/chart-stateless-precedence/`. Its `INDEX.md` maps every file to its original path. The original scratchpad paths, relative to the session scratchpad `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/`, are given in parentheses.

- `reports/round1-architecture.md` (was `assess-w3/round1-architecture.md`)
- `reports/round1-codebase.md` (was `assess-w3/round1-codebase.md`)
- `reports/round1-pushback.md` (was `assess-w3/round1-pushback.md`)
- `reports/round2-architecture.md` (was `assess-w3/round2-architecture.md`)
- `reports/round2-codebase.md` (was `assess-w3/round2-codebase.md`)
- `reports/round2-pushback.md` (was `assess-w3/round2-pushback.md`)
- `reports/round2-fresh.md` (was `assess-w3/round2-fresh.md`)
- `reports/round3-architecture.md` (was `assess-w3/round3-architecture.md`)
- `reports/round3-codebase.md` (was `assess-w3/round3-codebase.md`)
- `reports/round3-pushback.md` (was `assess-w3/round3-pushback.md`)
- `reports/round3-fresh.md` (was `assess-w3/round3-fresh.md`), with its prototype at `prototypes/no-map.diff.gz` (was `assess-w3/fresh3/prototype.diff`) and harness under `harness/fresh3/` (was `assess-w3/fresh3/exp_test.go.txt`)
- `reports/round4-architecture.md` (was `assess-w3/round4-architecture.md`)
- `reports/round4-codebase.md` (was `assess-w3/round4-codebase.md`), with its probe files under `harness/codebase-ov4/` (was `assess-w3/codebase/ov4/`)
- `reports/round4-pushback.md` (was `assess-w3/round4-pushback.md`)
- `reports/round4-fresh.md` (was `assess-w3/round4-fresh.md`), with its tests and generators under `harness/fresh4/` (was `assess-w3/fresh4/`)
- `reports/round5-architecture.md` (was `assess-w3/round5-architecture.md`)
- `reports/round5-codebase.md` (was `assess-w3/round5-codebase.md`)
- `reports/round5-pushback.md` (was `assess-w3/round5-pushback.md`)
- `reports/round5-fresh.md` (was `assess-w3/round5-fresh.md`), with its reproduction under `harness/fresh5/` (was `assess-w3/fresh5/out/`)
- `reports/round6-architecture.md`, `reports/round6-codebase.md`, `reports/round6-pushback.md` and `reports/round6-fresh.md` (were `assess-w3/round6-*.md`)
- `reports/round7-architecture.md`, `reports/round7-codebase.md`, `reports/round7-pushback.md` and `reports/round7-fresh.md` (were `assess-w3/round7-*.md`)
- Round 8's reports are not yet in the directory. They are at `assess-w3/round8-architecture.md`, `assess-w3/round8-codebase.md`, `assess-w3/round8-pushback.md` and `assess-w3/round8-fresh.md` in the session scratchpad.

The prototypes behind "Experiments after round 3" are `prototypes/stateless.diff.gz` (was `stateless/stateless.diff`) and `prototypes/combined.diff.gz` (was `combined/combined.diff`). Their measurements are under `measurement/` (was `measure/`, `stateless/` and `combined/`). The C1 spike's report is `c1/c1-report.md` (was `chart-recovery/c1-report.md`).

## Ledger

Rows from round 1 were raised against `7b264fa9c (52c82af1)`, rows from round 2 against `c270bbfa3 (68371487)`, rows from round 3 against `3e4d0c43b (a8960d05)`, rows from round 4 against `6dc1d9558 (123012c1)`, rows from round 5 against `2782fd181 (d97415a1)`, rows from round 6 against `788778ed0 (60327eb7)`, rows from round 7 against `9afff6c63 (ce616721)`, and rows from round 8 against `336067853 (ed91e462)`.

Statuses as each raiser declared them:

- **released**: the raiser names the text it releases against.
- **held**: the raiser keeps the finding open.
- **carried**: carried, by revision 2's §6, to the later design of the value on items. The raiser agrees in each case.
- **moot**: the raiser declares that the row's subject was retired by a later revision.
- **open**: no disposition yet.
- **raiser absent**: the raiser was not resumed, so no one can release the row. Where a retargeting may bear on such a row, the drafter says what the row's subject is and where it stands in the latest revision. That note is not a disposition.

A **carried to refinement** status is the raiser's own words for a row it calls issue-level. An **editorial, for refinement** status is the raiser's own words for a row it says refinement can make without another round. An **optional, for refinement** status is the raiser's own words for a row it says is not a condition of acceptance. A **view, no change** status is the raiser's own label for a row it says needs no change. Statuses marked r2 to r8 give the round of the latest disposition.

| id | raiser | round | section | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|
| A1 | assess-architecture | 1 | §1, §4.1, §4.2 | Refinement and review judge the issues against v2. They find §4.1 reversing v2 §8 with no ruling behind it, or they let an accepted ruling be reversed silently. | v2 §2 ("conflicts are kept as ambiguity and resolved in the forest"); v2 §8 (perigrin 2026-10-30, "precedence leaves recognition"); v2 §3 rule 12; v2 §10 kill "Facts in items". The §1 ruling is not in chart-perly-w2 Records or -inputs. Options B and C are undefined. | released (r2, against §1, §2); residue raised as A14 | 7b264fa9c (52c82af1) |
| A2 | assess-architecture | 1 | header, §7 | A reader following "Follows" reads superseded rules, and A1's check cannot be made from the document. | "Follows" names the 2026-10-29 first attempt, which v2 supersedes. §7's "design §13" exists only in v2. v2 is on branch `docs/chart-parser-v2-design`, not on pu. | released (r2, against the header) | 7b264fa9c (52c82af1) |
| A3 | assess-architecture | 1 | §2, §4.2 | `items-per-token` and `set-reuse` may barely move for left-associative chains (`.`, `+`, `&&`, `||`), and the exit fails for a reason the design did not anticipate. | Reasoned, not measured. Prediction is unchanged, and the inner sub-chains of `$a . $b . $c . $d` are legal and reduced before `.`, so their items survive. The 2016 approach passes the constraint down at prediction (from memory, not checked). | released as overtaken (r2, against §3 citing C1 spike); descendant A15 | 7b264fa9c (52c82af1) |
| A4 | assess-architecture | 1 | §6 | A key-set change with no ruling, and the precedence stage key's end state is undefined once `Constrain` is removed. Until then the timed pipeline pays for precedence three times. | README output contract: "no key is added, renamed or retired after it". Each earlier exception names a ruling. "A number may not become `na`". `forest.ParseWith` calls `parseStream`, then `f.Constrain`. | released (r2, against §1 "Rule them in", §5, §7) | 7b264fa9c (52c82af1) |
| A5 | assess-architecture | 1 | §2, §6 | A number with no derivation becomes the waypoint's verdict, read by an unstated protocol, and it re-gates speed against D1 without saying so. | The stage keys sum to 2.107, leaving 0.39 at 2.5. Stop 3 allows recognize up to 1.0. §2's 2.7 is an inference from medians (README speed protocol). chart-recovery grounded its exit in a projection. D1 extended 2026-11-15 ("dropping the speed gate was universal"). | released (r2, against §1 goal, §5); A17 for perigrin | 7b264fa9c (52c82af1) |
| A6 | assess-architecture | 1 | §6 (stops) | The stops detect noise or nothing, and only after the work has merged. | README rules: "Timing keys … are never compared with the baseline". v2 §9.3: never compared strictly. Stop 4 is set against a single run (4.7915); checkpoint 3 read 5.18 and 5.11. The speed stops fire only "once both tables are in". | released (r2, against §6 stops); residue raised as A16 | 7b264fa9c (52c82af1) |
| A7 | assess-architecture | 1 | §6 (held keys) | A gain is claimed from a smaller population, and gated keys are left out of the stops. | `items-per-token` and `set-reuse` are measured "over files recognition accepts", and that population changes when recognition refuses precedence errors. Precedent: the `speed-ratio-aggregate-cold` redefinition. The list omits `collapse-over-limit`, `tables-drift`, `ledger-gaps` and `lex-unmapped`. `oracle-exact-chart` is made a floor. | released (r2, against §5) | 7b264fa9c (52c82af1) |
| A8 | assess-architecture | 1 | §5 | The suite goes red, or the recognizer loses its outside oracle. The `Expected` contract (v2 §13) changes undecided. | `recognize/reference_test.go`, a textbook Earley reference with no precedence, used by `TestReferenceRecognizer`, `TestReferenceAgreesOnTruncatedFiles`, `TestWarmExpectedMatchesReference` and the fuzz oracle. v2 §3 rule 11. | released (r2, against §4.2, §4.3); wording raised as A13 | 7b264fa9c (52c82af1) |
| A9 | assess-architecture | 1 | §4.3, §5.5 | The equivalence proof is partial until the last issue, and that issue changes the prover. | `referee.judge` (`internal/chart/precedence.go`) builds through `builtForest`, then `f.Constrain`, so an under-filtering table is invisible while `Constrain` stays. "Reporting and not removing" is not in the key's definition. "Every corpus" names no key. | released (r2, against §2, §7) | 7b264fa9c (52c82af1) |
| A10 | assess-architecture | 1 | §4, §8 | Issues inherit a mechanism with no invariant to test it against. | §4 names functions, `uint32` and a (state, LHS) key. §3 finds every pair but one global. The README records an earlier chart-side keying that over-refused. §5.3's choice is missing from §8. | released (r2, against §4, §4.5) | 7b264fa9c (52c82af1) |
| A11 | assess-architecture | 1 | §7 | The speed exit is reached, the next decision (M2, interleaving) becomes less reachable, and nothing on record weighs the trade. | v2 §2: what a positive verdict enables. First attempt's rule 2 (set interning). v2 §13 convergence "on state equality". v2 §11 isolation is not restated. | released (r2, against §6, §2, §4.3, §7) | 7b264fa9c (52c82af1) |
| C1 | assess-codebase | 1 | §4.2 | `++ ++$x`, which bison accepts, is refused, and no referee notices. | (rules 256, 257) × POSTINC/POSTDEC/POSTJOIN resolve as error (one `%nonassoc` level). Toy `testdata/precedence/toys/prefix.txt` records `accept … PREINC PREINC NUM`. README: "`++ ++ $x` … is kept". No such sentence in `testdata/precedence/sentences.txt`. `shapes.pl`: every reduce or error shift item is `X: Z . t …`. | moot (r3, raiser: the value on items is retired, revision 3 §5); stays in the minute for any later design | 7b264fa9c (52c82af1) |
| C2 | assess-codebase | 1 | §4.1 | As literally written, the filter never fires on `$a = $a = …` or `$a + $b * $c`. One of the two fixes turns it into an LALR(1) lookahead filter that §3 did not count. | `TestScratchAssessSets`: `=` is {ASSIGNOP, PERLY_EQUAL_SIGN}, `*`, `+`, `-` and `++` similar. 2,595 of 7,455 operator tokens have more than one terminal. The second terminal is in no resolution pair. `TestScratchAssessFollow`; `la.pl`. | released (r2, against §4.5; caveat on "there" in §4) | 7b264fa9c (52c82af1) |
| C3 | assess-codebase | 1 | §4.1 (and §5) | Acceptance differs between lookahead on and off. The fuzz target and `TestLookaheadKeepsAcceptance` fail or are weakened. At lookahead 0 (interleaved mode) half of precedence vanishes. | `recognize.Options` doc: "Neither changes acceptance". Fuzz `oracleDisagreement` compares `optionCombos` against `refRun`. `NoLookahead` sets `follows = nil` (recognize.go l.276-280). `Expected` is built "as recognition without lookahead". | released (r2, against §4.3) | 7b264fa9c (52c82af1) |
| C4 | assess-codebase | 1 | §4.2, §5.1 | Stored values split sets for no reason. Without re-propagation, §4.2's "cannot lose a derivation" is false, and the failure shows as false refusals. | The value is read only where a completed child at `pos` completes into its parent at `pos`, so it can stay builder-local. `builder.complete` visits each item once, and `isNew` drops a duplicate (state, origin). | moot (r3, raiser: the value on items is retired, revision 3 §5); stays in the minute for any later design | 7b264fa9c (52c82af1) |
| C5 | assess-codebase | 1 | §5.2, §8.2 | The shift side loses the value on most of its traffic, because unit links go through Leo memos. | `TestScratchAssessLeo`: operator chains form no multi-link memos (only `else` 14 and `optlistexpr` 239). Single-link memos include termbinop 1,652, termeqop 187, termrelop 97 and termunop 110. `complete` adds the memo's `top` and `continue`s before `awaiting`. | moot (r3, raiser: the value on items is retired, revision 3 §5); stays in the minute for any later design | 7b264fa9c (52c82af1) |
| C6 | assess-codebase | 1 | §4.3 | A blocked reduce-side derivation is packed into the forest anyway, and "apply the tables when packing" cannot be done for each parent separately. | `forest.build` makes a fact for every final item, including completions the recognizer skipped. Example: `$x = $y + $z + $w`. Child nodes (sym, from, to) are shared by parents. 35 operator-completing states hold one reduce and no non-final kernel item (`mix.pl`). | released (r2, against §4) | 7b264fa9c (52c82af1) |
| C7 | assess-codebase | 1 | §4.1, §5.4 | The reduce table's keying is undecided. Keying by chart state needs a chart-to-bison state map that does not exist. | Chart state 373 matches bison 193 (the conflict) and 232 matches 153 (allowed). `NewPrecedence` uses bison's LR(0) states, "never the chart's split states". `resolve(r, t)` covers 669 of 670 pairs. | moot (r3, raiser: the recognizer filter is retired; `Constrain`'s state-keyed indirob handling is unchanged) | 7b264fa9c (52c82af1) |
| C8 | assess-codebase | 1 | §4.2, §8.1 | The value does not fit its type, and the identity element contradicts itself. | 32 levels plus a reserved bit is 33 bits. "Empty set's identity, a reserved bit" against "an empty result is zero". PLUGIN_* is already among the 32 (perly.y l.150-182). About 20 levels are possible on top (`xr.pl`). | moot (r3, raiser: the value on items is retired, revision 3 §5); stays in the minute for any later design | 7b264fa9c (52c82af1) |
| C9 | assess-codebase | 1 | §4.2 | `$a + $b < $c` and `$a and $b and $c` group wrongly if `%prec` unit rules set PREC_LOW. | `termbinop: termrelop %prec PREC_LOW`, `termbinop: termeqop %prec PREC_LOW`, `termrelop: relopchain %prec PREC_LOW`, `listexpr: term %prec PREC_LOW` and `expr: listexpr %prec PREC_LOW`. bison uses their precedence only in reduce-side pairs (rule 236 against CHRELOP, state 68). | moot (r3, raiser: the value on items is retired, revision 3 §5); stays in the minute for any later design | 7b264fa9c (52c82af1) |
| C10 | assess-codebase | 1 | §5.5, §4.3 | There is no equivalence proof, and the residue that gates `Constrain`'s removal has no key. | `precedence-covered` 1,304 of 1,347. 43 resolutions are uncovered (reviewed), and C1's case is absent. `precedence-disagree` reads only the generated fixture. The 13 toy referees run against `Constrain` only. | moot as framed (r3, raiser); its core recurs against revision 3 §3.1 as C17 | 7b264fa9c (52c82af1) |
| C11 | assess-codebase | 1 | §4.2, §4.1 | Raised as support: checking the top operator alone is sound only alongside the reduce side. Any reduce-side residue (C2, C3) also opens a shift-side hole. | `pred.pl`: in 43 of 46 (R, state-before-operand) pairs, R alone predicts the operand. The other 3 cannot begin a left-operand rule. The reduce side forces inner spine levels. | released (r2, moot: §2, §7) | 7b264fa9c (52c82af1) |
| P1 | assess-pushback | 1 | §4.2 | The implementer chooses between one value per item, which is wrong where a state holds several dotted rules, and one value per dotted rule, which is a different design with a different cost to set identity. | 51 of 772 states complete more than one rule (`states.txt`). The recovery plan (branch `docs/chart-recovery-plan`, §C, C1): "A single value per item could not kill one rule inside a state". | moot (r4, raiser: §6 puts the value on items out of scope) | 7b264fa9c (52c82af1) |
| P2 | assess-pushback | 1 | §4.2 | A derivation the table admits is lost after its item has already completed, and shows as false refusals (`rd-accepted-chart-refused` above 0). | `builder.add` (recognize.go:376) deduplicates with `isNew` and never revisits. Recovery plan C1: "An item whose value grows must complete again". | moot (r4, raiser: §6 puts the value on items out of scope) | 7b264fa9c (52c82af1) |
| P3 | assess-pushback | 1 | §4.2 | Under one reading, `$a << ($b < $c)` survives recognition. | `%prec` unit rules (perly.go:354, 360, 440, 442, 464) are both "unit" and "operator" rules. SHIFTOP/CHRELOP resolves as reduce (26 CHRELOP reduce pairs, bison run). §3 counts (rule, token) pairs, not level expressibility. | moot (r4, raiser: §6 puts the value on items out of scope; its analogue, whether the spine-top key sees through `%prec` unit rules, is left to the differential oracle and the first stop) | 7b264fa9c (52c82af1) |
| P4 | assess-pushback | 1 | §4.2, §5.4, §5.5 | A resolution the generator cannot map is dropped or mis-mapped silently. 43 resolutions' entries are unrefereed. | bison 373 rules against the chart's 350. One reduce resolution names POSTJOIN, which perly.go:796 omits. Referee coverage 1,304 of 1,347. No stated refuse, review-list or drift-check path. | released (r2, against §4.4, §2) | 7b264fa9c (52c82af1) |
| P5 | assess-pushback | 1 | §6, §8.4 | The waypoint cannot finish, or the exit is lowered partway through. | 4.79 = 0.655 + 0.790 + 0.662 + about 2.68. Lex is untouched, and resolve and readings are out of scope (§7). Reaching 2.5 needs the post-forest share at or below about 1.1, mostly from removing `Constrain`, which no key times. The 115 ms probe is std5k only. | released (r2, against §1 goal, §6); see P13 | 7b264fa9c (52c82af1) |
| P6 | assess-pushback | 1 | §4.3, §6 (stop 4) | Stop 4 is a false alarm or is never checked. | `Constrain` runs until the last issue, so no fall below 4.79 is expected "once both tables are in". No reading protocol, against quiet repeats of 5.18 and 5.11, 8.56 and 8.39. Stop 3's causal clause cannot be measured. | released (r2, against §5 ¶2, §6 third stop) | 7b264fa9c (52c82af1) |
| P7 | assess-pushback | 1 | §4.3, §8.3 | The last issue has no runnable criterion, or it removes `Constrain` on evidence from generated sentences alone. | `precedence-disagree` is defined over the generated sentences (README key table). `Constrain` has no report-only mode (precedence.go:322 returns a bool and unlinks). "The waypoint's last issue" names no milestone. | released (r2, dissolved by §7) | 7b264fa9c (52c82af1) |
| P8 | assess-pushback | 1 | §1, §5.1 | The merged code contradicts the live README and the binding design, and no superseding record says so. | README: "No fact is added to an Earley item (spec §3 rule 12, §10)". The recovery plan says a successor design must amend rule 12 and §10, with C1's numbers as evidence. The design does not cite C1. "Option B/C" appear in no committed document. "Follows" cites v1. | released (r2, against header, §1, §2) | 7b264fa9c (52c82af1) |
| P9 | assess-pushback | 1 | §5.2, §8.2 | With Leo on, the reduce-side filter is skipped for intermediate links, and `leo-disagree` (acceptance only) does not notice. | `complete` calls `leo()` before `awaiting`. The lookahead case relies on "FOLLOW only shrinks up a chain" (leo.go), and no precedence analogue is stated. Recovery plan C1: the memo stores no value, plus a guard on values. | released (r2, against §4.5) | 7b264fa9c (52c82af1) |
| P10 | assess-pushback | 1 | §4.2, §8.1 | The "no operator" bit collides with a level, or the type changes partway through implementation. | perly.go:744-806 declares 32 levels, PLUGIN_* among them. Adding the reserved bit makes 33. | moot (r4, raiser: §6 puts the value on items out of scope) | 7b264fa9c (52c82af1) |
| P11 | assess-pushback | 1 | §5.3 | Two designs are left open. If neither is built, a derivation through a blocked terminal reaches the forest. | §5.3: "either a check at the next scan or a residue for the forest's tables". | released (r2, against §4.5, §2) | 7b264fa9c (52c82af1) |
| P12 | assess-pushback | 1 | §3 | §3's figures cannot be reproduced from pu. | `internal/chart/offline/precedence/` is untracked and has no expected-output file. The figures reproduce with `count.sh /opt/corpora/perl5`. | released (r2, against commit 7b264fa9c tracking the scripts, and §3); expected-output file noted as a nit | 7b264fa9c (52c82af1) |
| A12 | assess-architecture | 2 | §4.1, §6 (stop 1) | The first issue either trips a stop by design, or regenerates a golden the test forbids regenerating. | `forest-digest.golden` hashes "the forest Build makes" (`forestDigest` → `built`), before the constraint. §4: "The forest build refuses the same completions". `TestForestDigestGolden`: "Never regenerate it to make a change pass". | moot (r3, raiser: approach retired); raiser says revision 3 §3.1 is correct | c270bbfa3 (68371487) |
| A13 | assess-architecture | 2 | §4.2 | A fuzz oracle written to the stated equality fails on the first `%nonassoc` chain, or it is weakened ad hoc. | `1 .. 2 .. 3`: the filter refuses `1 .. 2` before `..`, but `1 .. (2 .. 3)` is blocked only on the shift side. `++$x++` behaves the same. README "Forbid": error resolutions fail both sides. The oracle sits in `recognize`, and `forest` imports `recognize`. | moot (r3, raiser: approach retired) | c270bbfa3 (68371487) |
| A14 | assess-architecture | 2 | §2 | The same unrecorded reversal as A1, smaller: review at the end finds live text contradicting the build. | v2 §3 rule 6: "refuse a span … after the chart (§8)". Rule 12: "as precedence does (§8)". v2 is on an unmerged branch. Live README statements that become false: `Options`' "Neither changes acceptance", Status's "the precedence constraint over the forest", "reads only the forest". | moot (r3, raiser: approach retired; revision 3 amends nothing in v2) | c270bbfa3 (68371487) |
| A15 | assess-architecture | 2 | §3 | A small gain could be misread as "precedence in recognition doesn't pay" when it means "the reduce side was the small half". | Reasoned, not measured. C1 spike's 47.9% counts every alternative at ambiguous operator nodes. Same-level left-associative chains (`.`, `+`, `&&`, `\|\|`) are blocked only on the shift side. | released (r3, raiser: answered by the measurement after round 2) | c270bbfa3 (68371487) |
| A16 | assess-architecture | 2 | §5 | Stop 3 compares readings across sessions, and a change of kernel or load decides it. | The before-reading is on pu before any change, and the after-reading is up to 7 days later. chart-recovery: "Compare only interleaved pairs from one session". "Quiet" is undefined in §5; the README's working definition is a 1-minute load at or below the CPU count. | released (r3, against §4's reading protocol) | c270bbfa3 (68371487) |
| A17 | assess-architecture | 2 | §1, §6 | The exit of 5 may hold on the unchanged tree and say nothing about B, or be unreachable within scope. Raised for perigrin. | `baseline.txt` 4.7915 (one run); checkpoint 3 5.18 and 5.11; before-reading not yet taken. §7 excludes speed work on resolve and readings. | released (r3, raiser: answered by the ruling "Relative and ≤ 5", §1, §4) | c270bbfa3 (68371487) |
| C12 | assess-codebase | 2 | §4.1, §6 | The exit is unreachable, or the stop gets "interpreted" away. Declared "Blocking". | `forest-digest.golden` hashes `built()`, before the constraint (resolve_test.go l.304-314, forestdigest_test.go l.77-79). `$x = $y + $z;` loses the `(x=y)+z` alternative. | moot (r3, raiser: `Build` is no longer changed) | c270bbfa3 (68371487) |
| C13 | assess-codebase | 2 | §4.2 | The oracle fails on the first such input, or is weakened during execution without a decision. Declared "Blocking". | Fixture sentences `s0568 qw(a) .. qw(a) .. qw(a) ;` and `s0395 ++ qw(a) ++ ;`, both "refuse" in bison. Each keeps one grouping through recognition. `refRun` builds no forest. `oracleDisagreement` requires equality (fuzz_test.go l.317 and after). | moot (r3, raiser: the reference contract is untouched) | c270bbfa3 (68371487) |
| C14 | assess-codebase | 2 | §6 | The exit's speed clause can be met with no recognizer change. Raised as a question for perigrin. | Baseline 4.7915; 4.95 under VERDICT (postmortem). Only the stop "not below the before-reading's minimum" tests the decision. | released (r3, against §1 "Relative and ≤ 5" and §4's exit) | c270bbfa3 (68371487) |
| C15 | assess-codebase | 2 | §3 | The share of groupings is overstated. Low. | The C1 spike report: 47.9% "come from a rule that is an operator rule or carries %prec", which includes unit `%prec` rules such as `term: scalar %prec PERLY_PAREN_OPEN`. | moot (r3, raiser: the 47.9% citation is gone from §2) | c270bbfa3 (68371487) |
| C16 | assess-codebase | 2 | §4.1 | A refusal's reason and byte may move from `precedence` (Byte 0) to `grammar`. Low, issue-level. | Wherever every chart derivation dies on the reduce side. No case found among the obvious ones (`$a <=> $b <=> $c` keeps its right grouping). | moot (r3, raiser: recognition is unchanged) | c270bbfa3 (68371487) |
| P13 | assess-pushback | 2 | §6 | "Exit met" can be recorded with zero gain, alongside a fired no-gain stop. | Baseline 4.7915 ≤ 5. Checkpoint 3 read 5.18 and 5.11. A stop is "reported", not a failure of the exit. | released (r3, against §1 "Relative and ≤ 5" and §4's exit) | c270bbfa3 (68371487) |
| P14 | assess-pushback | 2 | §4.1, §6 (stop 1) | The first stop fires on the intended change, or a "never regenerate" golden is regenerated without a ruling. | `TestForestDigestGolden` (forestdigest_test.go:101) hashes Build's forest before `Constrain`. `TestSetDigestGolden` (setdigest_test.go:110) also moves under default options. Only `TestResolvedForestDigestGolden` (line 120) is expected to stay fixed. | moot (r3, raiser: approach retired, §5) | c270bbfa3 (68371487) |
| P15 | assess-pushback | 2 | §4.2 | A fuzz oracle written to "exactly" reports disagreements, and is weakened in an unplanned way or the issue stalls. | `1 .. 2 .. 3`: the reduce side blocks `1 .. 2` before `..`; `1 .. (2 .. 3)` is refused only on the shift side (§7 out of scope). Containment holds; equality does not. | moot (r3, raiser: approach retired, §5) | c270bbfa3 (68371487) |
| F1 | assess-fresh | 2 | §6 | The exit can pass with B doing nothing, or fail because of drift. | Baseline 4.7915. `testdata/ratchet/checkpoint3.txt` 5.1765 and 5.1051 (c4eb1e46). Records: recognize about 0.77 and about 0.90 on a5498fa4; stage sum 3.638 to 3.13 after a host move. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (the exit) persists in revision 4 §5, under the ruling "Relative and ≤ 5" | c270bbfa3 (68371487) |
| F2 | assess-fresh | 2 | §3, §6 | Seven days buys a few percent, and §6's question about C is answered only through noise. | Probe on 1,676 files (not the aggregate; load 0.7–1.1; single runs): B refuses 7.0% of packed alternatives with dependents, and `Constrain` drops 10.5%. `Constrain` on the B-pruned forest takes ×0.87–0.97 of its time; on its own output, ×0.94–1.00. `Constrain` is about 46% of `ParseWith` time. Estimate: 4.79 to about 4.6. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (the speed premise of B) returns in revision 4, which reinstates B as §3.3 and measures it (V2 against V1, 0.69 against 0.73) | c270bbfa3 (68371487) |
| F3 | assess-fresh | 2 | §4.2 | An implementer writes the equality test, it goes red on the first nonassoc chain, and the contract is rewritten during execution. | `$a <=> $b <=> $c;` (`TestParseAppliesPrecedence`) and `1 .. 2 + 3 .. 4;`: the root stays live after B. Fuzz toy grammars declare no precedence. `recognize` cannot import `forest`. v2 §3 rule 11. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (the reference contract) returns in revision 4 §3 ("Recognition's contract"), stated as containment | c270bbfa3 (68371487) |
| F4 | assess-fresh | 2 | §4.1, §4.3, §5 | §6's "a held key moves" stop becomes ambiguous, and goldens are regenerated ad hoc. | Sets change. Tests holding sets or acceptance: `TestReferenceRecognizer` (acceptance, `Inserted`, `Expected` under every Options combination), `TestReferenceAgreesOnTruncatedFiles`, `TestSetDigestGolden`, `FuzzRecognize`. `Expected` re-closes via `b.complete(nil)` over pruned origin sets. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (tests whose contracts move with a recognizer filter) returns with revision 4 §3.3; see C23, C27 and H2 | c270bbfa3 (68371487) |
| F5 | assess-fresh | 2 | §4.1 | Raised as support: exactness holds only while the table is computed on `NewPrecedence`'s automaton, not copied from bison's report. | Probe: 0 B-refused alternatives that `Constrain` keeps, and 0 verdict flips over 1,558,495 alternatives. Inserted-`;` contexts not modelled. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (exactness of the reduce-side filter) returns with revision 4 §3.3, under its differential test | c270bbfa3 (68371487) |
| F6 | assess-fresh | 2 | §3, §4.4 | The figures are bison's, but the table will be the chart's. Minor. | Chart automaton: 629 shift and 8 error pairs, against bison's 670 and 10; the difference is about the 44 POSTJOIN lines in `bison-conflicts.txt`. §3 omits the 10 error resolutions. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (bison's counts against the chart's table) returns with revision 4 §3.1 and §3.3 | c270bbfa3 (68371487) |
| F7 | assess-fresh | 2 | §6 (stops) | The recognize-stage stop is likely to trip for reasons unrelated to B. | C1 spike: items barely move, and B adds a lookup per completion. Records: recognize drifts about 15% between sessions on one commit. Same-session before and after is not stated. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (a recognize-stage stop) is not among revision 4 §5's stops | c270bbfa3 (68371487) |
| F8 | assess-fresh | 2 | §4.5 | Granularity left to the implementer. Issue-level. | `complete` works per (state, LHS) (`g.lhs[state]`); the forest drops facts per rule. Leo memos are keyed by (node, symbol), independent of `la`. At a slipper node `la` holds the inserted `;`. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (the filter's granularity, Leo, slipper) returns with revision 4 §3.3; see C28, C29 and H5 | c270bbfa3 (68371487) |
| F9 | assess-fresh | 2 | §5 | The precedence and resolve stage keys would time mutated forests. Issue-level. | `Constrain` and resolve mutate the forest in place, so forests must be cloned outside the timed span, as `measureStages` does for lattices and charts. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (fresh inputs for the stage keys) is addressed in revision 4 §4's text: "a fresh copy of the forest made outside the span" | c270bbfa3 (68371487) |
| F10 | assess-fresh | 2 | §2 | Rule 6 is amended without being listed. Minor. | v2 §3 rule 6 places the constraint "after the chart". The 2026-11-14 ruling on the spec's standing (chart-perly-w2-inputs Records) permits the amendment. | open; raiser absent (assess-fresh not resumed after round 2). Drafter's note, not a disposition: the subject (amending v2 rule 6) returns with revision 4 §3.3, which has no amendment section; see A26, C24 and H3 | c270bbfa3 (68371487) |
| C17 | assess-codebase | 3 | §3.1 | A changed kept set on a file outside the sample passes every listed check. | `forest-digest.golden` hashes `Build`'s forest before the constraint, so this change cannot move it. `resolved-forest-digest.golden` covers 438 sample files against a 2,211-file aggregate, and resolve's `prune` can hide an over-kept alternative. The referees compare a verdict and one grouping. Proposed: keep today's `Constrain` as a test-only oracle and compare verdict and `kept` per file. Shared memo slices (`seq` returns `right`) are the realistic corruption path. | released (r4, against §3: "Today's `Constrain` is kept as a test-only oracle. A differential test compares verdict and exact kept set per file over the aggregate, the referee and the toy referees.") | 3e4d0c43b (a8960d05) |
| C18 | assess-codebase | 3 | §3.2, §3.3 | §3.2 forbids every design that meets §3.3, and the overflow path a wrong answer could hide in goes untested. | Contexts per node are unknown before the pass (theoretical bound 687 × 238). Any per-node storage with overflow grows. Proposed: allow amortized growth of a pooled array but not hash-map growth, and add a test hook forcing inline capacity 1, patterned on the interner's `hashMask` (sets.go). Measured maxima, 12 eval and 32 mark contexts, mean the overflow never runs on the corpora. | released (r4, against §3.2: "Per-node lists may grow. A test forces several entries onto one node") | 3e4d0c43b (a8960d05) |
| C19 | assess-codebase | 3 | §3.4, §4 | Every timed pass after the first times an already-constrained or already-pruned forest, and both §4 stops read the precedence key. | `Constrain` unlinks in place, and resolve's `prune` prunes in place. The README's stage protocol prepares inputs "once outside the span" (l.638-641). `Forest` has no copy method. §2 measured an already-constrained forest at about 0.88 of the cost. | released (r4, against §4: "Each timed pass works on a fresh copy of the forest made outside the span") | 3e4d0c43b (a8960d05) |
| C20 | assess-codebase | 3 | §3.2 | A literal reading forbids `NewPrecedence`'s construction-time maps. Low. | Construction uses maps (`sym`, `level`, `in`, `id`, `kernels`). Only `lrState.next` is read per derivation (`walk`): a dense goto table of 687 × 237 entries. `tab.EOF` = 237, so terminal-indexed tables need 238 columns. | released (r4, against §3.2: "The maps `NewPrecedence` builds once per grammar stay.") | 3e4d0c43b (a8960d05) |
| C21 | assess-codebase | 3 | §4 | "Its first property holding" can be read as not requiring §3.2, the decision's own subject. Low. | §4: "§3 built, its first property holding". | released (r4, against §5's exit, which lists all parts) | 3e4d0c43b (a8960d05) |
| C22 | assess-codebase | 3 | §3.2 | Reusable buffers hung on `Precedence` would race under concurrent `Parse`. Low, issue-level. | `Precedence` is shared via `grammar.Derived(tab, buildParsePrecedence)` (readings.go). | released (r4, against §3.2: "Per-call state lives on the call, never on the shared, cached tables.") | 3e4d0c43b (a8960d05) |
| A18 | assess-architecture | 3 | §2 | Refinement or review misreads the stage figures as a regression, or anchors the ≤ 5 half of the exit to the wrong number. | §2 omits the minute's caveat ("Compare these stages with each other, not with the ratchet's keys") and the whole-pipeline 5.49 (5.48–5.64). Baseline `speed-stage-recognize` 0.7903 and `-forest` 0.6620, against §2's 0.94 and 0.73. | released (r4, against §2's caveat: "compare them with each other, not with `baseline.txt`") | 3e4d0c43b (a8960d05) |
| A19 | assess-architecture | 3 | §5 | A later reader treats "about 5%" as a permanent ceiling and never revisits precedence in recognition after the map work changes the cost shape. | (1 − 0.877) × 2.38 / 5.49 ≈ 5.3% is `Constrain`'s share alone. It omits build, resolve and readings on a forest up to 16.17% smaller, and was measured on the map-based implementation. | moot (r4, raiser: the "about 5%" claim is removed) | 3e4d0c43b (a8960d05) |
| A20 | assess-architecture | 3 | header | A search by file name finds the wrong answer: the file records a decision not to put precedence in the recognizer. Low. | The plan and the minute keep the name `chart-precedence-in-recognizer`. Line 6 explains this. Assessments are named name-for-name (`chart-waypoint.md`). Proposed: `git mv` at merge, or perigrin rules the name stays. | released (r4, against the rename and the design's header); the minute's residue raised as A29 | 3e4d0c43b (a8960d05) |
| A21 | assess-architecture | 3 | §3.4 | Every close gate that runs the ratchet gets slower, against a budget already overrun. | Arithmetic, not a run: about 33 × 0.546 × (1 + 2.38) ≈ 61 s for the precedence key, plus about 21–22 s each for resolve and readings, so roughly +100 s. README "Budget": 300 s budgeted, 543 s on pu, 664 and 683 s in KEYS's runs (input 01a1ec3c). Verify's 1,800 s per-criterion timeout. | released (r4, against §4: "Their cost to a ratchet run is reported with the commit that adds them.") | 3e4d0c43b (a8960d05) |
| A22 | assess-architecture | 3 | §3.4, §4 | Refinement writes an exit criterion that cannot run, or that compares one ratchet run with `baseline.txt`, which the README forbids. The live README goes stale. | README key table, rules, the `speed-stage sum` stderr line, and "There is no per-stage collector key". Recovery plan §6: the verdict reading is "recorded in the milestone's Records", not a resolution command. | released (r4, against §3's live-documents list and §4's reading recorded in Records) | 3e4d0c43b (a8960d05) |
| A23 | assess-architecture | 3 | §4, §5 | The next decision starts from no recorded intent. | Revision 2's §6 "what the result decides" sentence is dropped in revision 3. Candidates after the map work: recognize 0.94, build 0.73, the rest of `Constrain`, or nothing. | released (r4, against §5's "What the result decides") | 3e4d0c43b (a8960d05) |
| A24 | assess-architecture | 3 | §3.2 | The literal reading may exclude a bounded fallback that meets §3.3. Raised for perigrin. | Recovery plan R2: "find duplicates by scanning, switch to a map at 24 items". §3.2 forbids any map, and §3.3 asks that a slow node stay correct. | released (r4, raiser: answered by ruling and measurement, §3.2) | 3e4d0c43b (a8960d05) |
| A25 | assess-architecture | 3 | §4 | An issue whose precedence stage key does not move merges unreported. Optional. | §4 reads before and after "at the end". Recovery plan M: "reported before it merges". The w2 postmortem: "stopped the waypoint at FOREST's gate rather than after its merge". | released (r4, raiser: optional and not adopted) | 3e4d0c43b (a8960d05) |
| P16 | assess-pushback | 3 | §3.2 | Whoever implements first decides whether §3.2 is met: a hand-rolled hash table and a dense indexed table would both be defensible readings. Low. | §3.2: "No hash map is read or written". It does not say whether that means Go's `map` or any hashed lookup. | released (r4, against §3.2: "No hash table, built-in or hand-written, is read or written per derivation. Per-node lists may grow. … The maps `NewPrecedence` builds once per grammar stay.") | 3e4d0c43b (a8960d05) |
| G1 | assess-fresh3 | 3 | §1, §4 | The exit cannot tell perigrin whether the work paid for itself. | `baseline.txt` (retaken at 83fdd50f4) reads 4.7915 with `Constrain` already in `ParseWith`. The measurement reads 5.49 on another host and is "not comparable". The README records the key moving across a host change. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (on which host the ≤ 5 half is read) persists; revision 4 §2 says its ratios are "measured on one machine … compare them with each other, not with `baseline.txt`", and §4 reads speed in one session | 3e4d0c43b (a8960d05) |
| G2 | assess-fresh3 | 3 | §4 | If §3 is built with identical output and a faster stage but the ratio stays above 5, nothing stops the work before twice the budget. | §4's exit and stops have no case for "built, identical, stage key improved, exit not met". | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (no stop for the middle outcome) is addressed in revision 4 §5's text by the stop "the precedence stage is faster but `speed-ratio-aggregate-longlived` stays above 5" | 3e4d0c43b (a8960d05) |
| G3 | assess-fresh3 | 3 | §2 | A reader expects a size hint to recover much of the cost. In fresh3's experiment it does nothing measurable. | Experiment on 859 forests: maps with a size hint 111/91/104 ms against 111/97/107 ms; dense goto 0–7%; the no-map variant 40/36/35 ms. Recursion is 810 calls deep against a 64-frame CPU profile limit (`maxCPUProfStack`), so profile shares are approximate. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (profile shares against wall-clock evidence) persists; revision 4 §2 cites the prototypes' wall-clock ratios and no profile shares | 3e4d0c43b (a8960d05) |
| G4 | assess-fresh3 | 3 | §3.1 | A change in which alternatives survive on an aggregate file outside the sample passes every named instrument. | The goldens cover about 440 sample files. The count keys and referee keys count verdicts and disagreements. Proposed (issue-level): a differential check against today's `Constrain` over the 2,211 files, comparing node heads and `Alt.Next`. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (a differential check over the aggregate) is addressed in revision 4 §3's text by the named differential test | 3e4d0c43b (a8960d05) |
| G5 | assess-fresh3 | 3 | §3.2 | "Allocated once per call, not grown entry by entry" forbids the design that measured fastest, and pushes toward an arena where appending to a returned sub-slice overwrites its neighbour. | The prototype grows per-node slices by append for the 14% of nodes with more than one context. `eval` returns memoized slices directly, and `seq` returns `right` unchanged. Proposed: "no hash map; no table indexed by node, context or state grows during the call". | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (§3.2 as a property) is addressed in revision 4 §3.2's text: "Per-node lists may grow" | 3e4d0c43b (a8960d05) |
| G6 | assess-fresh3 | 3 | §3 | The issue author has to rediscover the memo semantics. Not a defect in the decision. | mark records before recursing; mark's key includes `first` (-1 means any); eval memoizes after computing, with no in-progress sentinel (safe because the forest is acyclic); memoized slices are shared; `kept []bool` is by alternative. The prototype kept all of these. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (memo semantics a rewrite must keep) is not stated in revision 4; the raiser called it issue-level | 3e4d0c43b (a8960d05) |
| G7 | assess-fresh3 | 3 | §3.4 | The issues have to land in a fixed order. Issue-level. | The three stage keys do not exist on pu: `measureStages` has lex, recognize and forest (`Build` alone). They need their own commit before any change to `Constrain`, and the interleaved before must be taken on it. This touches `main.go`, `main_test.go`, the README key table and the baseline. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (the stage keys' commit order) is addressed in revision 4 §4's text: "in a commit of their own on `pu`, before any change to precedence" | 3e4d0c43b (a8960d05) |
| G8 | assess-fresh3 | 3 | §4 (direction) | The exit's thresholds will not separate outcomes, and perigrin lacks the table of where the time goes next. | If the guacamole speedup holds, `Constrain` falls from 2.38 to about 0.8–0.9 and the pipeline from about 5.5 to about 4.0. Proposed: the exit report gives the seven-stage table, before and after, in one session. | open; raiser absent (assess-fresh3 not resumed in round 4). Drafter's note, not a disposition: the subject (the seven-stage table at the exit) is addressed in revision 4 §4 ("every stage key's median, minimum and maximum, before and after") and §5 ("The before-and-after stage table goes to perigrin") | 3e4d0c43b (a8960d05) |
| C23 | assess-codebase | 4 | §3 ("Identical output") | Declared "Blocking". The promised identical refusal reason is false for inputs whose every derivation dies on the reduce side. | Patched-prototype probe `TestZZAssess4`: `"print $fh [];"` recognizes under V0 but not V2, and the pipeline reason moves from "precedence" to "grammar". Ten other probes unchanged. With the filter on in `recognizeSource`, `TestPrecedenceIgnoredByRecognition` fails ("refused at byte 11"). No aggregate file has this shape. | moot (r5, raiser: V2 is deferred; recognition is unchanged) | 6dc1d9558 (123012c1) |
| C24 | assess-codebase | 4 | §3.3 (no amendment section) | Refinement meets the conflict mid-issue: §5's "amendment of a started issue" stop. | Revision 2's "What this decision amends" (v2 §8, v2 §2) is dropped; `grep -i amend` finds only the recovery-plan citation. `TestPrecedenceIgnoredByRecognition` encodes v2 §8. | released (r5, against §3's "What this does not amend") | 6dc1d9558 (123012c1) |
| C25 | assess-codebase | 4 | §3.1 | §3.1's key list is incomplete, and the guard's failure path is not the one named. Low, issue-level. | The shift side also keys on the item at the rule position, `item(r, m-1)`, and on the shifted terminal f, the first terminal of the right sibling. The prototype panics in `grammar.Derived`, where §3.1 promises "refused when the tables are built". The existing `prec.err` path is available. | released (r5, against §3.1's shift-side keys and §3.2's error path) | 6dc1d9558 (123012c1) |
| C26 | assess-codebase | 4 | §3.3 | "The forest build refuses the same completions" overstates what was built. Low. | `build` filters with `c.Refuses(r, next)`, which is `Never` only. It exempts every fact with origin j, including slipper kernel items that recognition filters. The raiser judges this sound, because the forest filter judges what Build keeps. | moot (r5, raiser: concerns V2) | 6dc1d9558 (123012c1) |
| C27 | assess-codebase | 4 | §3 ("Recognition's contract"), §6 | The contract has no named oracle, and the §6 test extension checks nothing precedence-related. | `TestOracleCatchesSkewUnderOneOption` asserts 3 `Options` fields on `toyStmt`, which has no precedence. `oracleDisagreement` requires equality with `refRun`. It is undecided whether the filter is an `Options` field (zero value off, as in the prototype) or always on. | moot (r5, raiser: concerns V2) | 6dc1d9558 (123012c1) |
| C28 | assess-codebase | 4 | §3.3 (Leo) | The parent-dependent `indirob` completion could be made through Leo. Low, issue-level. | The check must skip the top link, and must refuse on a parent-tabled link, not only on `Never`. `indirob ::= scalar` is a Leo item. | moot (r5, raiser: concerns V2) | 6dc1d9558 (123012c1) |
| C29 | assess-codebase | 4 | §3.3 (parent table) | An implementer keying by a single item, or by bison's state, gets the parent table wrong. Low, issue-level. | The key is the chart parent state, and refusal requires every item there awaiting the LHS to be forbidden. The path scans `src.Items()` with `goTo`, bypassing `awaiting()`. | moot (r5, raiser: concerns V2) | 6dc1d9558 (123012c1) |
| A26 | assess-architecture | 4 | §3.3 | Marked "must fix". The same silent reversal of an accepted ruling that A1 raised. Refinement and review read v2 §8 as the only statement of precedence and find §3.3 outside it. | v2 §8: "So precedence leaves recognition (perigrin, 2026-10-30)"; "This section is the only statement of how precedence works". Also v2 §2, rule 6 ("after the chart") and rule 12 ("as precedence does"). Revision 2's §2 named the first two; revision 4 has no amendment section. v2 is on an unmerged branch. | moot (r5, raiser: V2 is deferred, §3, §6); the raiser says the amendment carries to the V2 decision | 6dc1d9558 (123012c1) |
| A27 | assess-architecture | 4 | §3 ("Recognition's contract") | An over-refusing filter on input outside the aggregate goes unseen by any recognizer-level test, and "`Expected` is unchanged" rests on an unfixed detail. | No holder of the lower bound is named, and the fuzz oracle sits in `recognize`. `Expected` re-closes with `complete(nil)` while lookahead is on (`expected.go`). §6's note on `TestOracleCatchesSkewUnderOneOption` implies an unnamed new option. | moot here (r5, raiser); the raiser says it carries to the V2 decision | 6dc1d9558 (123012c1) |
| A28 | assess-architecture | 4 | §5, §6 | §6 contradicts §3, and "faster" in §5's new stop is undefined. Low. | §6: "The forest build applies the shift side". §3.3 has Build apply the reduce side, and §3.1 the shift side. Proposed: define "faster" as stop 3's complement. | released (r5, against §5's fourth stop and §6's wording) | 6dc1d9558 (123012c1) |
| A29 | assess-architecture | 4 | the minute | A reader arriving from the design's "Assessment:" link finds a subject path that is gone, under a title naming the retired approach. | At 6dc1d9558 the minute's ABOUTME, title and Subject line named the retired approach and the old path, and no line recorded the rename or revision 4. The drafter edited these in round 4 at the coordinator's direction (see Round 4). | released (r5, against the minute's ABOUTME, title, Subject and revision table); residue raised as A34 | 6dc1d9558 (123012c1) |
| A30 | assess-architecture | 4 | §2 | Once the scratchpad is cleared, the decision's measured basis cannot be reproduced or audited, and the minute's links die. | `stateless/stateless.diff`, `combined/combined.diff` and every round's report exist only in the session scratchpad. w1 postmortem: "Keep evidence in the Records, not the scratchpad." The C1 spike kept its commit on a local branch. | released (r5, against the plan's header naming the evidence directory, which the raiser verified); residue raised as A34 | 6dc1d9558 (123012c1) |
| A31 | assess-architecture | 4 | §3 | If `judge` keeps calling the oracle, the referee keys stay at 0 while measuring nothing about production. | `chart.CountPrecedence`'s `referee.judge` calls `f.Constrain` (`internal/chart/precedence.go`), and `TestPrecedenceToyReferees` runs against `Constrain`. §3 does not say which path these keys judge afterwards. | released (r5, against §3.5) | 6dc1d9558 (123012c1) |
| P17 | assess-pushback | 4 | §3 ("Identical output", "Recognition's contract") | A refusal reason can change with every named check green, or the first stop fires on output the recognition contract allows. | "Identical output" promises verdict, refusal reason, kept set and readings count; the differential test "compares verdict and exact kept set". When recognition refuses first, `parseStream` gives "grammar" with a byte instead of "precedence" at Byte 0. `print $fh []` is a candidate shape. | released (r5, against §3's heading "Recognition and the built forest are unchanged" and §3.4's five compared outputs over the refused fixtures) | 6dc1d9558 (123012c1) |
| H1 | assess-fresh4 | 4 | §3 ("Identical output"), §3.3 | The differential test as specified does not compare the reason, so the stop never fires and the change ships silently. This breaks the decision's promise, the README's refusal classes and `TestPrecedenceIgnoredByRecognition`. | `print $fh [];`: V0 `{Byte:0 Reason:precedence}`, V2 `{Byte:11 Reason:grammar}`. Also `sort $f [];`, `{ print $fh [] }`, `print $fh [] if 1;` and `my @x = (print $fh []);`, and 57 of 15,000 chain sentences. `combined/toys.txt` reports "recognition acceptance differs 30". 11,574 files gave zero V0 precedence refusals. | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject (refusal reason under the recognizer filter) is V2's, which revision 5 §6 defers; revision 5 §2 cites it as a reason for the deferral | 6dc1d9558 (123012c1) |
| H2 | assess-fresh4 | 4 | §3 ("`Expected` … is unchanged") | Error text and refusal bytes change, and refused-where and port fixtures may move. | `print $fh [ ) ;`, `sort $f [ ) ;` and `{ print $fh [ } ;`: V2's `Expected` lacks `PERLY_BRACKET_CLOSE`. `print $fh [ ] ) ;` and `print $fh [] ]`: V2 dies at node 4 against 5. `expected.go`: "a pruned completion is never needed by a later set" holds for follow-set lookahead, not for a precedence refusal. | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject (`Expected` under the recognizer filter) is V2's, which revision 5 §6 defers; revision 5 §2 cites it as a reason for the deferral | 6dc1d9558 (123012c1) |
| H3 | assess-fresh4 | 4 | §3.3 | The record does not say that §3.3 amends v2 §8. | v2 §8: "precedence leaves recognition … recognition uses perly.y's rules as written, is Boolean" (perigrin, 2026-10-30). README: "Recognition ignores both (spec §8)". | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject (amending v2 §8) is addressed in revision 5's text: "What this does not amend" says precedence is still applied after the chart | 6dc1d9558 (123012c1) |
| H4 | assess-fresh4 | 4 | §3.1 | "Refused when the tables are built, never mis-parsed" overclaims. The oracle is the real guard, and the generation failure is untested. | The guard checks only that keys agree across bison's LR(0) states. Evaluator logic (`canon`, m==1 pass-through, the empty-right-child shift exemption, `rootTop`) is covered only by the oracle. No test trips generation failure. `ReduceSideOf` does not fail on mixed keys. The slipper is safe only because `;`, `}` and EOF are in no Never set (0, 0, 0), and nothing guards that. | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject (the guard's limits and an untested failure path) is addressed in revision 5 §3.2's text: "It does not prove the evaluator supplies the right key … A toy grammar that makes generation fail is added" | 6dc1d9558 (123012c1) |
| H5 | assess-fresh4 | 4 | §3.3 (Leo) | The Leo branch is dead or unreached code with a promise attached. | `RefusedLinks` was 0 on the aggregate and 0 on 15,000 chains. Proposed: a forcing test, as §3.2 has for multi-entry memos, or drop the branch. | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject (the Leo branch under the recognizer filter) is V2's, which revision 5 §6 defers | 6dc1d9558 (123012c1) |
| H6 | assess-fresh4 | 4 | §5 | perigrin cannot weigh about 4% against a contract change, and the first stop is only as wide as the differential test. | V1 to V2 is 0.73 to 0.69, and every output change found comes from §3.3. Proposed: report V1 beside V2, or §3.3 as its own commit and reading; widen the test to `waypoint2Refusals`, the referee's refused sentences and the port refused fixtures, comparing reason, byte and `Expected`. | open; raiser absent (assess-fresh4 not resumed in round 5). Drafter's note, not a disposition: the subject has two parts: reporting V1 apart from V2 (V2 is deferred by revision 5 §6) and the test's width (revision 5 §3.4 widens it to reason, byte, `Expected` and refused fixtures) | 6dc1d9558 (123012c1) |
| A33 | assess-architecture | 5 | §1, §5 | Two readings of "done" coexist: the exit is met at about 4.1 while "3.92 first" is unmet, and the next decision starts from a disagreement nobody wrote down. Raised for perigrin, or a sentence. | §1: "Let's get ourselves to 3.92 first then look". In the combined prototype, 3.92 is V2's ParseWith and V1's is 4.16 (V0 5.62). §2 gives V1 as 0.73; 5.62 × 0.73 ≈ 4.1. | released (r6, against §1 "Target for V1 then V2 … It is not this decision's exit" and §5); meaning raised as A36 | 2782fd181 (d97415a1) |
| A34 | assess-architecture | 5 | the minute; the evidence directory | Once the scratchpad is cleared, the minute's links die, and some figures can be read but not reproduced. | The minute's header still points at the scratchpad (`assess-w3/…`, `stateless/`, `combined/`). `no-map.diff.gz` was `fresh3/prototype.diff`, verified by `cmp`, with no recorded mapping. Not copied: `fresh3/exp_test.go.txt`, `codebase/ov4/`, `chart-recovery/c1-report.md`. §2's 11,574, 20,083 and 15,000 appear only in `reports/round4-fresh.md`, with no generator. | released for the subject (r6, against `INDEX.md` and the evidence directory, which the raiser checked); the raiser assigns the minute's scratchpad pointers to the drafter (see Round 6, housekeeping) | 2782fd181 (d97415a1) |
| A35 | assess-architecture | 5 | §3.4 | The differential may push the chart suite past its 50 s budget, or land nowhere that runs regularly. Low. | The aggregate needs `T1_CORPUS` and `/opt/corpora/perl5`, and the prototype's preparation took 17 s. Precedents: `make chart-absence-reference` (a build tag, run by `make test` and CI's corpus-backed job) and the ratchet (not in `make test`). | released (r6, against §3.4 "Where it runs"); question raised as A38 | 2782fd181 (d97415a1) |
| C30 | assess-codebase | 5 | §3.4 | The new evaluator's refused path is checked on about 30 inputs. | `testdata/port/refused-where.txt` has 336 refused cases (282 `grammar`, 54 `table`) and none `precedence`. The aggregate has no precedence refusals. What remains is 22 referee refuse sentences, the toy referees and 10 `waypoint2Refusals` rows. The 185 and 989 precedence-refused sentences were generated in scratch and are not committed. | released (r6, against §3.4's inputs: the port's refused cases dropped, the seeded generator added) | 2782fd181 (d97415a1) |
| C31 | assess-codebase | 5 | §3.4, §3.5 | "Test-only oracle" and "differential over the aggregate" cannot both be true as written. | A `forest` `_test.go` file is invisible to `chart` tests and `cmd/ratchet`. The aggregate is built in `cmd/ratchet` (`aggregateFiles`, package `main`), and the perly referee runs in `chart`. The oracle needs `Precedence`'s unexported automaton. Both prototypes used non-test `zz_*.go` files. Options: (a) an exported reference function; (b) everything inside `forest`'s tests, reading testdata paths. | released (r6, against §3.4: `forest.ConstrainReference`, documented as used only by tests) | 2782fd181 (d97415a1) |
| C32 | assess-codebase | 5 | §3.1 | The referee and the toys could end up on different tables from production, with a lock or map lookup on every `Constrain` call. Low, issue-level. | The tables belong inside `NewPrecedence`, stored on `Precedence`. stateless used a global `map[*Precedence]` under a mutex; combined used `grammar.Derived` keyed by tables. | released (r6, against §3.1 "stored on the `Precedence` value" and §3.3) | 2782fd181 (d97415a1) |
| C33 | assess-codebase | 5 | §3.1, §3.3 | Evaluator details that are easy to break silently change root or memo judgements. Low, issue-level. | Root spine tops `item(r, 0)` and `buildShift`'s dot-0 rule; canonicalisation of `pend` and `pi`; `pi = -1` allowed on mixed pairs (`reduces`: `pi < 0 ||`), safe on perly only by grammar shape; the empty-right-child exemption; cached slices shared by `seq`; about 72 MB of memo heads on 3-million-node forests. | carried to refinement (r6 and r7, raiser: issue-level) | 2782fd181 (d97415a1) |
| P18 | assess-pushback | 5 | §1, §5 | Refinement reads 3.92 as this work's target, reached on the prototype machine only by the out-of-scope V2, or the result reads as a miss. | §1 records "Let's get ourselves to 3.92 first then look". Combined prototype: ParseWith V0 5.62, V1 4.16, V2 3.92. §5's exit is the relative exit and ≤ 5, with no mention of 3.92. | released (r6, against §1 "3.92 is where this decision and the later V2 decision should land together. It is not this decision's exit …" and §5) | 2782fd181 (d97415a1) |
| J1 | assess-fresh5 | 5 | §2, header | perigrin cannot audit the number the ≤ 5 projection rests on, and the refusal-path evidence exists only in agent reports. The raiser calls this "a record defect, not a wrong number". | `combined-compare.txt` has no RESULT lines; timing needs `CB_TIME`, and its output was not kept. `constrain-measurement.md` reads 5.49, not 5.62. No harness or output for the 11,574, 20,083 and 15,000 populations. The raw files listed in `constrain-measurement.md` are absent. "alts built 2566149" against 2565144 elsewhere. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6's header lists raw timing runs and harnesses in the evidence directory, and `INDEX.md` addresses the 5.49 against 5.62 and the 2,566,149 against 2,565,144 counts | 2782fd181 (d97415a1) |
| J2 | assess-fresh5 | 5 | §1 | The milestone Records carry contradictory rulings, and an executor may treat 3.92 as the target or V2 as in scope. | "Do just B first" refers to the V2 filter this decision defers. "V2: both (Recommended)" is superseded by "V1 now, V2 measured after", and nothing marks it so. "3.92" is used nowhere in §5. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §1 is "Rulings in force" and names what "V1 now, V2 measured after" supersedes; it states that 3.92 is not this decision's exit | 2782fd181 (d97415a1) |
| J3 | assess-fresh5 | 5 | §3.4 | Comparing items that cannot differ needs production-side machinery that a test-only oracle should not need. | `Expected` is a property of the recognizer's chart (`recognize/expected.go`). A precedence refusal's Byte is always 0 (`forest.go` `parseStream`). The reason is "precedence" exactly when the verdict is false. The prototype used a package global, `ConstrainVariant`. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §3.4 compares verdict, kept set and readings at the constraint's boundary, and says recognition, the byte and `Expected` cannot differ | 2782fd181 (d97415a1) |
| J4 | assess-fresh5 | 5 | §3.4, §3.5 | The strongest comparison may run only on the author's machine, and a skip reads as green. | `Constrain` is called from `forest` (the toys), `chart` (`precedence.go`, non-test; `precedence_test.go`; `waypoint2Refusals`) and `cmd/ratchet` (`corpus()`, `perlGit`, `aggregateFiles`). `make test` skips the perlgit tests without the perl5 tree (`perlgit_test.go`). `chart-ratchet` is advisory. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §3.4 names `forest.ConstrainReference` and `make chart-precedence-reference`; where the aggregate part runs is the subject of P19, A38, C34 and K1 | 2782fd181 (d97415a1) |
| J5 | assess-fresh5 | 5 | §5 | One outcome meets neither the exit nor any stop. | Not covered: the stage key improves and the ratio is ≤ 5, but the ratio's after-median is not below its before-minimum. Proposed: a stop for "exit not met for any other reason". The raiser's run suggests the gap will not bite (V1's worst round is below V0's best). | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §5 adds the stop "§3 is built and its checks pass, but the exit is otherwise not met" | 2782fd181 (d97415a1) |
| J6 | assess-fresh5 | 5 | §4 | "Never compared with the baseline" misdescribes the mechanics, and the stage table will not sum to ParseWith. Low. | `checkBaseline` (`cmd/ratchet/compare.go`) requires every printed key once in `baseline.txt`. Adding a key touches `baseline.txt`, `main.go`'s keys slice (which cites a ruling per addition) and `main_test.go`. Timing keys pass compare either way (`TestWaypoint2KeysCompareRules`). libops has no stage key. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §4: "printed, listed in `baseline.txt` and never gated", and "Library operators get no stage key" | 2782fd181 (d97415a1) |
| J7 | assess-fresh5 | 5 | §3.2 | The guard narrows `NewPrecedence`'s contract without saying so, and its test is promised, not shown. | A guard failure on perly.y refuses every file via `prec.err` (forest.go:134), which shows in `rd-accepted-chart-refused`. Neither prototype has a toy that makes generation fail. The guard enumerates every LR(0) state holding a key, so it may refuse a grammar `Constrain` handles. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §3.2 says the guard "is stricter than today's `Constrain`, so it may refuse a grammar that `Constrain` handles", and calls the failing toy "new work" | 2782fd181 (d97415a1) |
| J8 | assess-fresh5 | 5 | §3.4 | "The port's refused cases" names no fixture. | `port_relational_unary_test.go` defers the four chains to "the issue that completes the port". | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §3.4's inputs no longer name the port's refused cases | 2782fd181 (d97415a1) |
| J9 | assess-fresh5 | 5 | §3.4 | The 185 and 989 precedence-refused sentences that §2 cites are one-shot evidence. | `stateless-compare.txt`: "oracle refuses 0" on the aggregate. Refusal coverage rests on the referee, the toys and `waypoint2Refusals`. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: revision 6 §3.4 adds a generator "committed with a fixed seed" for the refused sentences | 2782fd181 (d97415a1) |
| J10 | assess-fresh5 | 5 | §3 | Raised in support: the decision fits v2. | Rule 12 untouched. v2 §8's operator context at both edges still holds. Rule 11 is kept by §3.5. The six parent-item pairs match the README's history (T1/array.t; the twostates and indirobsort toys). Ablations: mixed-forbid flips 50 files, noshift 625, noreduce 1,099. | open; raiser absent (assess-fresh5 not resumed in round 6). Drafter's note, not a disposition: raised in support; revision 6 keeps §3.1's keys and §3.5 | 2782fd181 (d97415a1) |
| A36 | assess-architecture | 6 | §1, §5 | The V2 decision inherits a target it cannot check without restating it, or claims success or failure against the wrong instrument. Raised for perigrin, or a sentence. | 3.92 is the combined harness's V2 ParseWith ratio (`combined-runs/result-run3.txt`: `RESULT V2-parsewith ratio 3.9205`; three runs 3.85–4.02), with V0 at 5.62. The ratchet's key reads 4.79 on pu for the same code. On the ratchet, 0.69–0.70 of V0 would read about 3.3–3.4. | moot (r7, raiser: by ruling, §1 "Drop the number … This supersedes the 3.92 target") | 788778ed0 (60327eb7) |
| A37 | assess-architecture | 6 | §1, §2 | The direction "toward 3.92" looks settled by measurement, while the only reproduction found V2 adding almost nothing. The V2 decision starts from an overstated expectation. | `reports/round5-fresh.md`: "V2's ParseWith came out at 0.723 of V0's, not 0.69". Combined runs: V2 3.85–4.02 (median 3.92); V1 4.03–4.17 (median 4.16); fresh5 V1 4.12. §2 cites fresh5 for V1 only. | released (r7, against §2's V2 paragraph: "Its gain over V1 did not reproduce", and V1's figures given per source) | 788778ed0 (60327eb7) |
| A38 | assess-architecture | 6 | §3.4 | The `make test` gate gains a machine-specific corpus dependency that CLAUDE.md's 100% rule cannot accommodate. Raised as a question before refinement. | §3.4: the aggregate "needs the T1 and perl.git corpora"; `make test` runs the target; "it never skips". Makefile `test:` lines 21–23 run `chart-absence-reference`, whose perl.git use the raiser did not check. | released (r7, raiser: by ruling "Split it" and §3.4's table, "Corpus needed: none" for every `make test` row) | 788778ed0 (60327eb7) |
| A39 | assess-architecture | 6 | §2, `INDEX.md` | Small figure drifts in the section the index exists to back. Low. | Per process, V1 over V0 precedence is 0.423, 0.406 and 0.392, against §2's "0.40–0.41". `harness/fresh4/list.txt.gz` has 11,588 lines against §2's 11,574 files, unexplained in the index. | released (r7, against §2 "0.39–0.42 of today's, per run" and `INDEX.md`'s account of the 11,588 lines) | 788778ed0 (60327eb7) |
| C34 | assess-codebase | 6 | §3.4 | A contributor or agent without `/opt/corpora/perl5` can no longer commit; `make test` absorbs a full population pass; the CI step lacks `PERL5_CORPUS`. | ci.yml: "Its corpus is in the tree (T1's in-tree default), so it needs no env." `perlGit` refuses a root that is missing, dirty or not at 94e50866 (`perlgit.go`). `aggregateFiles` calls `supersetPopulation`. The Ratchets job sets `PERL5_CORPUS` only on its own `go test` step. Options: (a) keep it out of `make test`; (b) split it. | released (r7, against §3.4's split table and "A developer without perl.git runs `make test`") | 788778ed0 (60327eb7) |
| C35 | assess-codebase | 6 | §3.4 | Comparing readings from `cmd/ratchet` needs an export or a production test switch. Low. | `countReadings` is unexported, and `ParseWith` calls it after resolve (readings.go l.32). The prototype used a production global, `ConstrainVariant`. Equal kept sets imply equal readings. | released (r7, against §3.4: "Equal kept sets give equal readings") | 788778ed0 (60327eb7) |
| C36 | assess-codebase | 6 | §3.4 | Asserting 185 and 989 makes the test a change detector that a legitimate table or lexer change breaks. Low. | The counts belong to today's grammar and lexer. Seeds 4 and 44 in `harness/fresh4/` (`rand.NewSource`) reproduce the sentences. Runtime of 35,083 parses in the ordinary suite is unmeasured. | released (r7, against §3.4: it "does not assert the 185 and 989 of today's grammar", and "If it proves slow") | 788778ed0 (60327eb7) |
| C37 | assess-codebase | 6 | §3.1 | The generation cost adds to every cold start and every test that builds a `Precedence`. Low, not a stop. | `NewStateless` takes about 90 ms per build on the committed grammar. Today's `buildParsePrecedence` takes about 310 ms, including grammar loading (`TestZZAssess6GenTime`, overlay). | released (r7, against §2: "Building the tables costs about 90 ms, once per process") | 788778ed0 (60327eb7) |
| P19 | assess-pushback | 6 | §3.4 | `make test` fails on every machine without the perl.git corpus, blocking commits under pvm's CLAUDE.md, or the target is softened to a skip mid-issue. | §3.4: "When its corpora are missing, the target fails; it never skips." The Makefile's `test` runs `chart-absence-reference`, whose corpus is in the tree (ci.yml). perl.git comes from `PERL5_CORPUS` or `/opt/corpora/perl5` (`oracle.go`, `perlgit.go`). ci.yml: corpus-backed ratchets skip when the corpus is absent, "correct for a contributor". CI's matrix runs `go test ./...`. | released (r7, against §3.4's "inputs, and where each runs" table and "The perl.git target is not run by `make test` … A developer without perl.git runs `make test`") | 788778ed0 (60327eb7) |
| K1 | assess-fresh6 | 6 | §3.4 | Contributors and agents without the corpus get a red `make test`, with pressure toward exactly the skip the decision forbids. | ci.yml on the absence reference: "Its corpus is in the tree … so it needs no env". The aggregate's 562 perlgit files need the pinned, clean perl5 tree. T1 is in the tree (`testdata/perlonjava/unit`, 986 files), so only perl.git is external. Proposed: the `PVM_REQUIRE_PERL542` pattern (skip locally, fail in CI). | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §3.4's table runs only corpus-free inputs in `make test` and puts the perl.git files in their own target; it no longer says "the T1 and perl.git corpora" | 788778ed0 (60327eb7) |
| K2 | assess-fresh6 | 6 | §3.4 | A correctness test does a timing pass, and its membership is load-dependent in principle. | `Protocol.Aggregate` (speed.go:93) times recursive descent on every population file and drops std5k, edit5k and any file over 10% of the total (code_too_large.t at 15.2%). The superset population (2,214) contains the aggregate and needs no timing. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §3.4: "Population is the superset population, which is deterministic, not the timing-chosen aggregate" | 788778ed0 (60327eb7) |
| K3 | assess-fresh6 | 6 | §2 | Two figures do not reproduce as stated. Cosmetic. | Per-run V1/V0 precedence 0.392, 0.406, 0.423 (ratio of medians 0.401; fresh5 0.397), against "0.40–0.41". "2.2–2.4, about 43%" splices `combined-runs` (2.18–2.23, 39% of 5.62) with `constrain-measurement.md` (2.33–2.45, 43.4%). | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §2 gives 2.33–2.45 (43%) and 2.18–2.23 (39%) per source, and "0.39–0.42 of today's, per run" | 788778ed0 (60327eb7) |
| K4 | assess-fresh6 | 6 | §1 | "What the result decides" anchors the V2 decision on a single three-run median. | V1 run medians 4.16, 4.03, 4.17, fresh5 4.12. V2 4.02, 3.85, 3.92, fresh5 4.10. The only reproduction shows V2's increment at about 0.02. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §2 records V2 at 4.10 against V1's 4.12 in round 5 ("Its gain over V1 did not reproduce"), and §1 records "Drop the number" | 788778ed0 (60327eb7) |
| K5 | assess-fresh6 | 6 | §4, §5 | Outcomes that fall between the stated cases have no instruction. | Nothing says what happens between day 7 and day 14 of the budget, or when no quiet session can be had (the README's 2026-11-15 convention records load). The before-state commit is not named. The stage keys' cost has no threshold; `chart-ratchet.yml` has `timeout-minutes: 20`. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §4 says "That commit is the before-state"; §5 adds stops for no quiet session and for the CI timeout, and a status report at the budget itself | 788778ed0 (60327eb7) |
| K6 | assess-fresh6 | 6 | §3.4, §3.5 | The name reads as a package function, and the callers that switch silently are not listed. | `Constrain` is a method, `(*Forest).Constrain` (precedence.go:322). Callers: forest.go:138; chart/precedence.go:327; chart/precedence_test.go:95 and :242; forest/precedence_test.go:88. "The reference's own agreement with bison stays tested" names no test. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §3.4 names `(*Forest).ConstrainReference`, lists the switching callers, and says "A new test holds the reference's own agreement with bison" | 788778ed0 (60327eb7) |
| K7 | assess-fresh6 | 6 | §3.4 | Asserted as exact counts, any grammar change breaks the test, and the counts were measured on V0. | Verified exact: seed 4 gives 185 refused of the 20,000 random sentences, and the 83 handcrafted add 2 more; seed 44 gives 989 of 15,000 chains. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §3.4: the test "does not assert the 185 and 989 of today's grammar" | 788778ed0 (60327eb7) |
| K8 | assess-fresh6 | 6 | §4 | The stage keys need API the decision does not name. | `resolve`, `countReadings` and `libraryOperators` are unexported, and the ratchet is package `main`. No `Forest` clone exists outside the scratch harnesses. The precedence stage's input is the forest after `libraryOperators`. §6's `internal/chart` scope allows the new work. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: revision 7 §4 lists new work: an exported forest copy and exported access to `libraryOperators`, resolve and the readings count | 788778ed0 (60327eb7) |
| K9 | assess-fresh6 | 6 | evidence directory (`INDEX.md`) | The 11,574-file run cannot be re-run from the record as written. | 3,627 of 11,588 paths in `harness/fresh4/list.txt.gz` point at fresh4's deleted scratch worktree. Remapped to an 8c3317079 checkout, they give 11,574 sources and 0 differences. | open; raiser absent (assess-fresh6 not resumed in round 7). Drafter's note, not a disposition: the subject is the evidence index; assess-architecture quotes `INDEX.md` as now mapping the 11,588 lines onto the 11,574 files | 788778ed0 (60327eb7) |
| A40 | assess-architecture | 7 | §2 | A reader takes "most" to mean a majority and overestimates what this decision can save. Low. | §2: "`Constrain` takes most of the pipeline's time". Its own table gives 39% and 43%. | released (r8, against §2: "`Constrain` takes the pipeline's largest share.") | 9afff6c63 (ce616721) |
| A41 | assess-architecture | 7 | §3.4 | The population is mis-stated, or `make test` grows with no stated recourse. Low. | The in-tree row reads "(T1, Guacamole, std5k)". The README's in-tree corpora are guacamole, std5k, edit5k and T1. The generators have a slow-path fallback; the in-tree part (about 1,650 files judged twice) has none. ABSENTFAST held the chart suite to 50 s, and the prototype prepared 2,211 forests in 17–19 s. | released (r8, against §3.4's in-tree row naming edit5k and parsent, and the fallback covering every `make test` row) | 9afff6c63 (ce616721) |
| A42 | assess-architecture | 7 | §5 | A moved digest or count key has no stop to report it, so regenerating is left to a worker's judgement, against the golden's own comment. Marked "should fix". | §3.4: "What stays unchanged: both forest digests and every ratchet count key." The first stop is only "a §3.4 check differs from the reference". Revision 3's first stop was "any output in §3.1 changes". | released (r8, against §5's first stop: "or a forest digest or ratchet count key moves") | 9afff6c63 (ce616721) |
| A43 | assess-architecture | 7 | §1 | Whether allocation work may start while V2 is in execution is left to interpretation. Low; raised for perigrin, or a word. | §1: "wait until the V2 decision is made"; §5: "wait". "Made" can mean decided or delivered. | released (r8, raiser: answered by ruling, §1 "V2's verdict reading (Recommended)" and §5) | 9afff6c63 (ce616721) |
| C38 | assess-codebase | 7 | §3.4 | An implementer building the in-tree part from the row, not from `corpus()`, silently drops files from the differential. Low, editorial. | `corpus()` (lex.go l.40-70) also yields `internal/psc/testdata/latency/edit5k.pl` (labelled `accept`, `in-tree.txt` l.908) and Guacamole's `parsent/` files (124 labelled). | released (r8, against §3.4's in-tree row: "T1, Guacamole parses and parsent, std5k, edit5k") | 9afff6c63 (ce616721) |
| P20 | assess-pushback | 7 | §2 | A reader takes `Constrain` to be more than half the pipeline, overstating what V1 can buy. Low; the raiser does not hold acceptance on it. | §2: "takes most of the pipeline's time"; the table beneath gives 43% and 39%. | released (r8, against §2: "`Constrain` takes the pipeline's largest share.") | 9afff6c63 (ce616721) |
| L1 | assess-fresh7 | 7 | §4, §5 | The exit and the precedence-stage stop each have two defensible bars, so the verdict can be disputed after the fact. | §5's exit is judged on `speed-ratio-aggregate-longlived`, which is not a stage key, and §4 records only stage keys. `Protocol.Ratio` (speed.go:68) already returns the median of 3 measurements of 2+9 pairs (v2 §9.3's sense). In `combined-runs`, today's run medians are 5.59–5.68 and its round minimum 4.99. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §4 defines "One measurement is one ratchet run's printed value of a key" and records `speed-ratio-aggregate-longlived` with the stage keys | 9afff6c63 (ce616721) |
| L2 | assess-fresh7 | 7 | §4 | The before-state commit either times the copying or changes the protocol mid-issue, which is a §5 stop. | `Protocol.pass` (speed.go:54–65) starts the clock and then calls `side()`; no per-pass setup runs outside the span. The alternative, pre-built copies for 3×11 passes per mutating key, inflates the heap, which `constrain-measurement.md` blames for 5.49 against 4.79. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §4 lists as new work "A per-pass preparation hook in the ratchet's timing protocol … holding one pass's copies at a time" | 9afff6c63 (ce616721) |
| L3 | assess-fresh7 | 7 | §1, §6 | If the ruling at refinement is no, §3.4's sixth row has nowhere to run. | §1 lists "Split it" as in force; §6 defers the `Makefile` and CI paths "for perigrin to rule in at refinement". v2 §11: "Only internal/chart changes". The "does not amend" paragraph omits §11. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §1 records "Rule them in (Recommended)" for the `Makefile` target and the CI step, and §3 lists the v2 §11 departure | 9afff6c63 (ce616721) |
| L4 | assess-fresh7 | 7 | §2 | The table-build figure comes from a different prototype. Small: the cost is once per process. | "About 90 ms" is from stateless.diff (105, 122 and 85 ms; 86), whose tables are maps. The raiser timed combined.diff's `NewStateless` five times at load 0.16: 105–137 ms; `NewPrecedence` a further 297–343 ms. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §2 gives "105–137 ms once per process (round 7)" for V1's dense tables | 9afff6c63 (ce616721) |
| L5 | assess-fresh7 | 7 | §3.4 | An implementer adds `libraryOperators` to the referee comparison and changes what it compares. | The referee judges `builtForest` (Build only), narrowed to the sentence's terminals, with no `libraryOperators` (chart/precedence.go `judge`). "After `libraryOperators`" holds only for the population and generator rows. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §3.4 compares "the forest its caller passes to the constraint today", and for the referee "`Build`'s forest narrowed to the sentence's terminals" | 9afff6c63 (ce616721) |
| L6 | assess-fresh7 | 7 | §3.4, §4 | The in-tree differential depends on §4's exports without saying so, and the in-tree list is incomplete. | The superset population is computed in `cmd/ratchet` (package `main`), so the in-tree differential is a `cmd/ratchet` test needing §4's forest copy and exported `libraryOperators`. §4 justifies those by the stage keys alone. `corpus()` includes edit5k.pl. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §3.4's in-tree row names edit5k and parsent and says "new test in `cmd/ratchet`, using §4's new API" | 9afff6c63 (ce616721) |
| L7 | assess-fresh7 | 7 | §3.4 | New work is not marked as new. | "The seeded generators of §2 (committed)" exist only as evidence in `harness/fresh4/*.go.txt`, which call the prototype's `fresh4Run` and `CStats`. Turning them into tests, `make chart-precedence-reference` and its build tag are new work; §3.2 and §4 mark new work explicitly. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §3.4's table marks the generators' tests, the in-tree and perl.git tests and the target as new | 9afff6c63 (ce616721) |
| L8 | assess-fresh7 | 7 | §3.4 | The standing checks drop the corpus refusals that §2's evidence covered. | The superset population excludes every file perl refuses, where real-code precedence refusals live. §2's 11,574-file run included such files. Proposed: compare on every recognized corpus file. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §3.4's corpus rows cover "every recognized file … perl-refused files included" | 9afff6c63 (ce616721) |
| L9 | assess-fresh7 | 7 | §2, `INDEX.md` | Two claims are not backed by the committed evidence. | "Round 6 re-ran both generators and the 11,574 files": no output of that re-run is committed under `measurement/` or `harness/`. `INDEX.md` says "the decision quotes 5.62", but revision 7 no longer does. | open; raiser absent (assess-fresh7 not resumed in round 8). Drafter's note, not a disposition: revision 8 §2 says of round 6's re-run "the outputs themselves were not kept"; `INDEX.md` now says "Revisions 4 to 6 quoted 5.62" | 9afff6c63 (ce616721) |
| C39 | assess-codebase | 8 | §3.4 | Only the pipeline's own scope reproduces the pipeline's input to the constraint. Low, editorial. | `libraryOperators` (resolve.go l.630-656) calls `r.prune(f)` when `scope.LibraryOperators` marks a plugin terminal dead. In the ratchet the scope is `rd(path).scope` from `chartInputs` (verdicts.go). The raiser also corrects its own round-3 and round-6 claim that `libraryOperators` "only reads the forest". | editorial, for refinement (r8, raiser: "Refinement can carry this; it is not a round finding") | 336067853 (ed91e462) |
| A44 | assess-architecture | 8 | §4 | If one run's aggregate differed, a before-against-after comparison could reflect a different file set rather than the change. Low. | README: "The share moves with load, and so does which files the rule excludes … Re-read it rather than trust it." The ratchet prints each excluded file on stderr. Proposed: add "and each run's excluded files" to "What it records". | optional, for refinement (r8, raiser: not a condition of acceptance) | 336067853 (ed91e462) |
| A45 | assess-architecture | 8 | §3 ("What this does not amend") | The live record of path rulings falls behind; at worst a reader concludes the `Makefile` target was never ruled in. Low. | The README's isolation paragraph (lines 370–382) records every ruling admitting a path outside `internal/chart` (e.g. MAPDISPATCH's `Makefile` and `ci.yml`). The decision's live-documents list names only the precedence-constraint section and the key table. | optional, for refinement (r8, raiser: not a condition of acceptance) | 336067853 (ed91e462) |
| M1 | assess-fresh8 | 8 | §5 | The last issue can close on the exit with a stop live, or halt despite the exit, at the implementer's choice. | The whole-ratio exit can be met while the precedence stage's median is not below its before-minimum; the stage keys are timed in separate passes, so noise can do this. Nothing orders exit and stops. Proposed: "a stop overrides the exit". Raiser: decision-level. | open | 336067853 (ed91e462) |
| M2 | assess-fresh8 | 8 | §5 | An outcome that is neither exit nor stop, whose natural repair is the tuning the stops forbid: reseeding, dropping a sentence, loosening an assertion. | The exit requires "the checks of §3.4 pass"; stop 1 covers only "differs from the reference". Checks that can fail without differing: the reference-against-bison test, "at least one sentence is refused", the failing toy grammar, the perl.git target timing out or skipping. Proposed: "a §3.4 check fails, or …". Raiser: decision-level. | open | 336067853 (ed91e462) |
| M3 | assess-fresh8 | 8 | §3.4 | If the corpus bulk moves behind `chart-precedence-reference`'s tag, the in-tree agreement check runs only in CI, undoing "Split it" for in-tree corpora. | "Its bulk moves behind the build tag" names no tag. `make test` already runs the tagged `chart-absence-reference` (Makefile:23); `chart-precedence-reference` is explicitly not run by `make test`. "Slow" has no threshold. Raiser: mostly the implementing issue, one clause in the decision. | open | 336067853 (ed91e462) |
| M4 | assess-fresh8 | 8 | §3 ("What this does not amend") | A reader checking v2 §9.5 finds an unrecorded departure. Low. | v2 §9.5: keys "may be added or retired when the reset lands, and never afterwards". The decision adds three keys under "Rule them in" (as Recovery Plan D5 ruled its own), but the list names only §11. Raiser: decision-level. | open | 336067853 (ed91e462) |
| M5 | assess-fresh8 | 8 | §2 | Some figures cannot be reproduced from the kept outputs. The raiser says nothing rests on them. Low. | 105–137 ms (round 7); 20,083 with 185 refused and 15,000 with 989 refused; round 6's re-run. `INDEX.md` has no round-7 harness or output entry. Proposed: label them "reported, not kept", as §2 does for round 6, or index round 7's output. | open | 336067853 (ed91e462) |
| M6 | assess-fresh8 | 8 | §1, §5 | None: raised as a view. "≤ 5" is near-vacuous on the baseline machine, so the relative half carries the exit. | Baseline 4.79; the harness read 5.5–5.6; at ×0.73 either is under 5. Raiser: "consistent with 'Relative and ≤ 5', and no change is needed." | view, no change (r8, raiser) | 336067853 (ed91e462) |
| M7 | assess-fresh8 | 8 | §4 | The stage keys' handling of refused files is undefined. Raiser: issue-level. | A file `libraryOperators` or the constraint refuses has no forest for the next stage. `chartStages` documents its equivalent for `Build`. Resolve's copies must be taken after the constraint. | open | 336067853 (ed91e462) |
| M8 | assess-fresh8 | 8 | §3.4 | About 1,650 files, judged twice with a forest copy each, run inside `go test ./...` on three OSes. Raiser: issue-level. | The CI matrix runs `go test ./...` on every OS; the "proves slow" escape is subject to M3. | open | 336067853 (ed91e462) |

### Cross-references

The drafter lists overlaps and does not merge them. Each row keeps its own status.

**From round 1:**

| cluster | rows | shared subject |
|---|---|---|
| value per item; re-propagation | P1, P2, C4 | what the value attaches to, and widening after completion |
| which rules set the value | P3, C9, C1 | `%prec` unit rules; prefix rules |
| governing decision and amendment | P8, A1, A2 | rule 12, the §10 kill, v2 §8; v1 cited as "Follows"; Options B and C undefined; the C1 spike not cited |
| value width | P10, C8 | 33 bits in a `uint32`; PLUGIN_* levels already counted |
| exit | P5, A5 | 2.5 underived; no reading protocol |
| stops | P6, A6 | timing compared to a single run; stops fire late or on predicted behaviour |
| removal gate and referee | P7, C10, A9 | `precedence-disagree` covers generated sentences only; no report-only mode; no residue key |
| Leo | P9, C5 | the Leo interaction in §5.2 |
| reference recognizer and Options | A8, C3 | the reference recognizer, `Expected`, the lookahead-independence contract |
| table generation and keying | P4, C7, A10 | rule mapping, keying by state or by rule, error path |
| the §5.3 residue and which terminals the filter tests | P11, C2, A10, C11 | ambiguous lexemes |

**From round 2:**

| cluster | rows | shared subject |
|---|---|---|
| forest digest | P14, A12, C12 (and F4 on the set digest) | `forest-digest.golden` hashes Build's forest before the constraint; §4.1 and stop 1 hold it unchanged |
| reference contract | P15, A13, C13, F3 | "exactly" against containment; where the oracle can live |
| exit | P13, A17, C14, F1 | the ≤ 5 exit against the 4.7915 baseline and run-to-run drift |
| tests and evidence for §4.1 | C10, F4 | which tests and keys evidence exactness over the corpora; which tests' contracts move |
| what B can remove; speed share | A15, F2 (and C15 on the 47.9% figure) | the share of the forest and of `Constrain`'s cost that the reduce side can touch |
| rule 6 | A14, F10 | v2 §3 rule 6's "after the chart" clause, not listed in §2 |
| session and reading protocol | A16, F7 | before and after in one session, interleaved; "quiet" undefined |
| indirob and what "there" keys on | C7, C2 (caveat), A10 (minor), F8 | a rule-global table against a per-state exception; (state, LHS) granularity |
| exactness argument | F5, C10 | the subset claim and the automaton it is computed on |
| tables are the chart's | F6, P4 | the chart's automaton against bison's counts; POSTJOIN |

**From round 3:**

| cluster | rows | shared subject |
|---|---|---|
| differential oracle for §3.1 | C17, G4 (and C10's core) | today's `Constrain` as a test-only oracle, compared per file over the aggregate |
| what "no map" forbids and allows | C18, G5, P16, A24 (and C20, C22) | §3.2 against §3.3; growth, overflow, construction-time maps, hashed structures |
| shared memo slices | C17, G5, G6 | memoized slices are shared and must not be mutated |
| stage-key protocol | C19, G7, A21, A22 | fresh forests per pass; commit order; ratchet run time; README and Records |
| §2's evidence | A18, G3 | the dropped caveat and 5.49; profile shares against wall-clock evidence |
| exit and what it decides | G1, G2, G8, A23 (and C21, A25) | the ≤ 5 half on which host; the middle outcome; the seven-stage table; the next target |
| the retirement bound | A19 | "about 5%" as `Constrain`'s share on the map-based implementation |
| the file name | A20 | the retained name `chart-precedence-in-recognizer` |

**From round 4:**

| cluster | rows | shared subject |
|---|---|---|
| refusal reason under the recognizer filter | C23, P17, H1 (and H2 on `Expected` and the byte; round 2's C16, declared moot in round 3) | `print $fh []` shapes refused by recognition with "grammar" and a byte, where today it is "precedence" at Byte 0 |
| amendment of v2 §8 | C24, A26, H3 (and round 1's A1 and round 2's A14 and F10) | §3.3 puts the reduce side in recognition with no amendment section |
| recognition contract and its oracle | C27, A27 (and round 2's A13, C13, P15 and F3, declared moot in round 3 or raiser absent) | containment's lower bound has no named holder; the `Options` question |
| width of the differential test | H6, P17 (and C23) | the test compares verdict and kept set only; widen to reason, byte, `Expected`, refused fixtures |
| the generation-time guard | C25, H4 | keys §3.1 omits; panic against refusal; what the guard does not cover |
| Leo under the filter | C28, H5 (and round 2's P9 and C5) | refusing a parent-tabled link; the untested branch |
| evidence location | A30 | prototypes and reports only in the scratchpad |
| the minute's leftovers | A29 (and A20) | title, ABOUTME and Subject after the rename |
| what the referee keys judge | A31 | production path against the oracle |
| §3.3's forest-side claim and §6's wording | C26, A28 | what Build actually refuses; reduce side against shift side |

**From round 5:**

| cluster | rows | shared subject |
|---|---|---|
| 3.92 | A33, P18, J2 | a ruling's figure, V2's on the prototype machine, against an exit that does not name it; J2 also marks superseded rulings |
| evidence | A34, J1 (and A30, released) | what the evidence directory holds; the minute's scratchpad pointers; missing generators and outputs |
| refused-path fixtures | C30, J8, J9 | "the port's refused cases" holds no precedence refusal; the 185 and 989 refused sentences are not committed |
| oracle home and where the differential runs | C31, J4, A35 | a test-only oracle invisible to `chart` and `cmd/ratchet`; skips on machines without the corpora; the suite's budget |
| exit and stop gap | J5 | an outcome that meets neither the exit nor any stop |
| baseline wording | J6 | "never compared with the baseline" against `checkBaseline`; libops has no stage key |
| items compared that cannot differ | J3 (and P17, H6, both earlier) | `Expected`, byte and reason under V1 |
| the generation guard | J7 (and C25, H4, both earlier) | the test is promised, not shown; a narrowed `NewPrecedence` contract |
| evaluator details | C32, C33 | where the tables are built; root, canonicalisation, `pi = -1`, shared slices |
| carried to the V2 decision | A26, A27 (and H1, H2, H5, C23, C26–C29) | rows whose raisers declare them moot here or whose subject is V2's |

**From round 6:**

| cluster | rows | shared subject |
|---|---|---|
| where the aggregate check runs | P19, A38, C34, K1 (and A35, J4, C31, earlier) | `make test` running a target that needs the perl.git corpus and never skips |
| 3.92 and V2's unreproduced gain | A36, K4, A37 (and A33, P18, J2, earlier) | the figure's instrument; V2 at 0.723 in the only reproduction |
| figure ranges | A39, K3 | 0.39–0.42 against "0.40–0.41"; 2.2–2.4 and 43% from two sources; 11,588 against 11,574 |
| readings at the boundary | C35 | an export or a production switch to compare readings |
| don't assert 185 and 989 | C36, K7 | counts that belong to today's grammar |
| aggregate against superset population | K2 | a correctness input defined by a timing protocol |
| uncovered outcomes | K5 | budget between 1× and 2×; no quiet session; before-state commit; stage-key cost |
| method name and silent caller switch | K6 | `(*Forest).ConstrainReference`; the callers of `Constrain` |
| new API for stage keys | K8 | unexported stages; no `Forest` clone |
| INDEX paths | K9 (and A34, J1, earlier) | fresh4's file list points at a deleted worktree |
| generation cost | C37 | about 90 ms per `Precedence` build |

**From round 7:**

| cluster | rows | shared subject |
|---|---|---|
| "most" against 39–43% | P20, A40 | one word in §2 against its own table |
| edit5k, the fallback, and the in-tree part's dependence on §4's API | A41, C38, L6 | the in-tree row omits edit5k (and, per C38, Guacamole `parsent/`); no slow-path fallback; the in-tree differential needs §4's exports |
| digest and count-key stop | A42 | the first stop names only the reference comparison |
| "wait until" | A43 | "made" as decided or delivered |
| exit key and "measurement" | L1 | the reading omits the exit's key; three senses of median and minimum |
| out-of-span copies | L2 (and K8, earlier) | no `Protocol` hook for per-pass setup |
| paths deferred against "Split it" | L3 | §6 defers the `Makefile` and CI paths that §1's ruling needs |
| table-build cost | L4 (and C37, earlier) | 90 ms from the map-based prototype; V1's dense tables untimed |
| referee boundary | L5 | the referee has no `libraryOperators` |
| new work unmarked | L7 | the generators, the target and its build tag |
| corpus refusals | L8 (and K2, earlier) | the superset population drops perl-refused files |
| evidence claims | L9 (and K9, J1, earlier) | the round-6 re-run's output; `INDEX.md`'s stale 5.62 line |

**From round 8:**

| cluster | rows | shared subject |
|---|---|---|
| exit and stops | M1, M2 (and J5, A42, earlier) | how the exit and the stops order, and a failing check that differs from no reference |
| where the bulk moves if slow | M3, M8 (and A41, C36, earlier) | which build tag; whether `make test` keeps the in-tree check; cost on three OSes |
| departures from v2 | M4, A45 (and A1, A14, A26, C24, H3, L3, earlier) | v2 §9.5's key rule and the README isolation paragraph |
| figures without kept outputs | M5 (and L9, K9, J1, earlier) | round 6's and round 7's figures, recorded only in reports |
| the verdict reading's record | A44 (and L1, earlier) | each run's excluded files |
| the constraint's input scope | C39 (and L5, earlier) | which scope the corpus checks pass to `libraryOperators` |
| "≤ 5" | M6 (and P13, A17, C14, F1, G1, earlier) | the absolute half of the exit |
| stage keys and refused files | M7 (and C19, K8, L2, earlier) | what the next stage times when a file is refused |

### Distribution by section

Counted by the first section each row names. A row naming several sections counts once. Each round's rows use that round's revision's numbering, and the numberings differ.

**Round 1 (revision 1):**

| section | rows |
|---|---|
| §1 | A1, P8 |
| header | A2 |
| §2 | A3, A5 |
| §3 | P12 |
| §4 / §4.1 | A10, C2, C3, C7 |
| §4.2 | C1, C4, C8, C9, C11, P1, P2, P3, P4, P10 |
| §4.3 | A9, C6, P6, P7 |
| §5 | A8 |
| §5.2 | C5, P9 |
| §5.3 | P11 |
| §5.5 | C10 |
| §6 | A4, A6, A7, P5 |
| §7 | A11 |

**Round 2 (revision 2):**

| section | rows |
|---|---|
| §1 | A17 |
| §2 | A14, F10 |
| §3 | A15, C15, F2, F6 |
| §4.1 | A12, C12, C16, P14, F4, F5 |
| §4.2 | A13, C13, P15, F3 |
| §4.5 | F8 |
| §5 | A16, F9 |
| §6 | C14, P13, F1, F7 |

**Round 3 (revision 3):**

| section | rows |
|---|---|
| header | A20 |
| §1 | G1 |
| §2 | A18, G3 |
| §3 | G6 |
| §3.1 | C17, G4 |
| §3.2 | C18, C20, C22, A24, P16, G5 |
| §3.4 | C19, A21, A22, G7 |
| §4 | C21, A23, A25, G2, G8 |
| §5 | A19 |

**Round 4 (revision 4):**

| section | rows |
|---|---|
| the minute | A29 |
| §2 | A30 |
| §3 | C23, C27, A27, A31, P17, H1, H2 |
| §3.1 | C25, H4 |
| §3.3 | C24, C26, C28, C29, A26, H3, H5 |
| §5 | A28, H6 |

**Round 5 (revision 5):**

| section | rows |
|---|---|
| the minute | A34 |
| §1 | A33, P18, J2 |
| §2 | J1 |
| §3 | J10 |
| §3.1 | C32, C33 |
| §3.2 | J7 |
| §3.4 | A35, C30, C31, J3, J4, J8, J9 |
| §4 | J6 |
| §5 | J5 |

**Round 6 (revision 6):**

| section | rows |
|---|---|
| §1 | A36, A37, K4 |
| §2 | A39, K3 |
| §3.1 | C37 |
| §3.4 | A38, C34, C35, C36, P19, K1, K2, K6, K7 |
| §4 | K5, K8 |
| evidence directory | K9 |

**Round 7 (revision 7):**

| section | rows |
|---|---|
| §1 | A43, L3 |
| §2 | A40, P20, L4, L9 |
| §3.4 | A41, C38, L5, L6, L7, L8 |
| §4 | L1, L2 |
| §5 | A42 |

**Round 8 (revision 8):**

| section | rows |
|---|---|
| §1 | M6 |
| §2 | M5 |
| §3 | A45, M4 |
| §3.4 | C39, M3, M8 |
| §4 | A44, M7 |
| §5 | M1, M2 |

## Round 1

Dispatched: assess-architecture, assess-codebase and assess-pushback. All three reported. Each read blob 52c82af1 and confirmed it with `git hash-object`.

### Recommendations, as declared

- **assess-architecture: modify.** Its report: "The idea is plausible, and §6's keys-first ordering is right. But the document reverses accepted decisions it does not name (A1, A2). Its proof of equivalence and its oracles do not survive the change as described (A8, A9). Its exit and stops break the ratchet's own rules on timing and keys, and are not grounded the way chart-recovery's were (A4, A5, A6)."
  - Its own severity column marks these as "must fix": A1, A2, A3 ("must address"), A4, A5, A6, A8 and A9.
  - It marks these as "should fix": A7, A10 and A11.
- **assess-codebase: modify.** Its report: "The mechanism fits this recognizer, and my measurements back the design's core bet."
  - It lists C1 under "Blocking before acceptance".
  - It lists C2, C3, C4, C6 and C10 as "Should be decided in the decision text".
  - It lists C5, C7, C8 and C9 as "Issue-level, not for the decision".
- **assess-pushback: modify.** The direction is sound and the evidence checked is accurate. Three things block acceptance: the value semantics (P1, P2), the exit and stops (P5–P7), and the unamended rule (P8).

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | No block or stand-aside declared in those terms. Its severity labels ("must fix", "must address", "should fix") are quoted above and are not read as blocks. |
| assess-codebase | "Blocking before acceptance: C1", in its own words. No stand-aside declared. |
| assess-pushback | No block or stand-aside declared in those terms. Its report says P1, P2, P5–P7 and P8 "block acceptance". |

None of these is converted into a block or a release by the drafter. Each participant confirms its row in round 2.

### What each participant says remains

At the end of round 1, all 34 rows remained open. No participant has released a finding, and no finding has been rebutted.

### Outcome

**Another round is owed.** Three participants recommend modify, and every finding is open. Round 2 assesses the revision that answers them, and goes to the participants as the diff plus this ledger.

## Round 2

**Revision:** c270bbfa3, blob 68371487a68e3f8c1661a80ea830356228a3a163. It follows perigrin's rulings of 2026-11-23, made after round 1:

- "Do just B first and let's re-measure";
- "Rule them in" (the three stage keys);
- the goal: `speed-ratio-aggregate-longlived` at or below 5, with a budget of 7 days.

Revision 2's §6 proposes carrying round 1's value-only findings (C1, C4, C5, C8, C9, P1, P2, P3, P10) to the later design of the value on items.

**Dispatched:** assess-architecture, assess-codebase, assess-pushback (as raiser of P1–P12 only; see the drafter note) and assess-fresh. All four reported. Each confirmed the blob:

- assess-architecture, assess-codebase and assess-fresh with `git hash-object` on the live file;
- assess-pushback by reading `git show c270bbfa3:` and confirming the blob id with `git rev-parse`.

assess-architecture, assess-codebase and assess-pushback each say they read the subject first and the diff last. assess-fresh says it wrote its report before reading this minute and had still not read it. Its report does not mention reading the diff.

### Recommendations, as declared

- **assess-architecture: modify.** "The remaining changes are small, and none of them changes the direction of the decision."
  - It released A1–A11.
  - It marks as "must fix": A12, A13.
  - It marks as "should fix": A14, A15, A16.
  - It marks as "for perigrin": A17.
- **assess-codebase: modify.**
  - It names two blocking findings, "each fixable in a sentence": C12 and C13.
  - It holds C7 and C10.
  - It released C2, C3, C6, C11.
  - It carried, with agreement: C1, C4, C5, C8, C9.
  - It raised for perigrin: C14.
  - It rates as low: C15, C16.
  - In its words: "The narrowing to the reduce side itself fits the code well."
- **assess-pushback (as raiser): modify.**
  - It released P4–P9, P11, P12.
  - It carried, with agreement: P1, P2, P3, P10.
  - It raised P14 (high), P15 (medium) and P13 (medium-low).
- **assess-fresh: modify.** "The mechanism is sound, and it fits the code." It calls F2 "the central finding". It says F1–F3 are "design-level problems, and I would fix them before refinement, not in issues".

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | No block or stand-aside declared in those terms. Its labels: "must fix" (A12, A13), "should fix" (A14–A16), "for perigrin" (A17). |
| assess-codebase | "Blocking", in its own words: C12, C13. Held: C7, C10. No stand-aside declared. |
| assess-pushback | No block or stand-aside declared in those terms. Its severity labels are as above. |
| assess-fresh | No block or stand-aside declared in those terms. It asks for F1–F3 to be fixed "before refinement, not in issues". |

The drafter converts none of these into a block or a release.

### Dispositions of round-1 findings, by their raisers

| raiser | released | held | carried (raiser agrees) |
|---|---|---|---|
| assess-architecture | A1 (residue A14), A2, A3 (as overtaken; descendant A15), A4, A5 (see A17), A6 (residue A16), A7, A8 (wording A13), A9, A10, A11 | none | none of its own. It agrees that the nine in §6's list are value-only. |
| assess-codebase | C2, C3, C6, C11 | C7, C10 (narrowed) | C1, C4, C5, C8, C9 |
| assess-pushback | P4, P5, P6, P7, P8, P9, P11, P12 | none | P1, P2, P3, P10 |

Each release names the revision-2 text it rests on; see the ledger's status column and the reports. All three raisers note their own bias toward finding their bar met.

### What each participant says remains

- **assess-architecture:** A12 and A13 (must fix), A14–A16 (should fix), A17 (for perigrin). Minor points with no row:
  - the recovery plan's path (A2);
  - `precedence-covered` not listed (A7);
  - one clause on `Expected`'s re-closure (A8);
  - what "there" keys on (A10).
- **assess-codebase:** C12 and C13 (blocking), C7 and C10 (held), C14 (for perigrin), C15 and C16 (low).
- **assess-pushback:** P13, P14, P15.
- **assess-fresh:** F1–F10. It asks for F1–F3 before refinement and calls F8 and F9 issue-level. F5 is raised as support.

### Rulings after the round's reports

Recorded verbatim, as rulings and not as findings. The gloss after each is the coordinator's statement of its meaning.

1. **On the exit**, asked after P13, A17 and C14:

   > "Relative and ≤ 5 (Recommended)"

   That is: exit when the after-reading's median is below the before-reading's minimum, taken interleaved in one session, and at or below 5.

2. **On direction**, asked after F2's measurement that `Constrain`'s cost barely depends on forest size:

   > "Measure first, then decide"

   That is: before any milestone, confirm the bound with a quiet ratchet-aggregate measurement, record it, and return to perigrin.

The drafter does not say which rows these rulings answer. Each row's raiser disposes of it.

### Outcome

**Not a fixed point.**

- The subject is held pending the measurement named in the direction ruling, and perigrin's decision after it.
- There is no revision 3 yet.
- Every round-2 row (A12–A17, C12–C16, P13–P15, F1–F10) is open, and C7 and C10 are held.

## Author's statement (recorded, not a finding)

After round 1, the author reported the following. It is the author's statement, recorded as given. The drafter has not checked it, and it is not a ledger row.

> chart-recovery's C1 spike already measured precedence values on items. Its report is at scratchpad `chart-recovery/c1-report.md`. Records 2026-11-13, and perigrin's 2026-11-14 D7 ruling: "I think the successor design is something we can revisit after §8 as written." The design does not cite it.

That report is now `c1/c1-report.md` in the evidence directory (was `chart-recovery/c1-report.md` in the session scratchpad).

Revision 2's header (lines 11-12) cites "the chart-recovery spike C1 (`chart-recovery` Records, 2026-11-13)", and its §1 quotes D7. The drafter records that as a fact about the text, not as a disposition of anything.

## Measurement after round 2 (recorded, not a finding)

Taken under perigrin's "Measure first, then decide", by the measure-constrain agent on `pu` 8c3317079. This section is recorded by the coordinator from the agent's report. That report is now `measurement/constrain-measurement.md` in the evidence directory (was `measure/report.md` in the session scratchpad), with its raw runs under `measurement/constrain-runs/`.

**Population and protocol**

- The population is the ratchet's own aggregate: `aggregateFiles` with `productionProtocol` gives 2,211 files (T1 827, guacamole 822, perlgit 562), the same list in every run.
- Machine: 4 cores, kernel 6.18.44-fc-v726, go1.24.4. The 1-minute load was 0.27–1.39.
- Three processes, each with 2 warm-up rounds and 9 timed rounds, long-lived, with every stage and variant interleaved within each round. Figures are the median of the three run medians, with the minimum and maximum in brackets.

**Stage time, as a ratio over recursive descent's time (546 ms)**

| Stage | Ratio |
|---|---|
| ParseWith, the whole pipeline | 5.49 (5.48–5.64) |
| lex | 0.66 |
| recognize | 0.94 |
| build | 0.73 |
| library operators | 0.09 |
| Constrain | 2.38 (2.33–2.45) |
| resolve | 0.20 |
| readings | 0.15 |

Constrain is 43.4% of ParseWith. The whole-pipeline ratio differs from the baseline's 4.79: the machine is different and the harness holds every forest in memory. Compare these stages with each other, not with the ratchet's keys.

**Constrain on three forests**

| Forest | Time, as a fraction of the as-built forest's | Packed alternatives |
|---|---|---|
| As built | 1 | 2,565,144 |
| Reduce-side pruned | 0.891 (0.861–0.917) | 2,421,356 (−5.61%) |
| Everything Constrain drops already removed | 0.877 (0.847–0.921) | 2,150,267 (−16.17%) |

**Exactness of the reduce-side filter:** 0 alternatives it removes are kept by Constrain, and 0 verdicts flip.

**Profile:** about 57% of Constrain's own CPU is map operations keyed by edge context. `mapassign` takes 30% (7.4 s of its 12.4 s is growth of maps made with no size hint), eval's lookups 18%, and `Precedence.walk`'s goto map 9%.

**Context counts:** one Constrain pass over the same aggregate, by the coordinator, on an instrumented scratch copy.

- 3,069,300 nodes.
- 2,594,592 eval contexts over 2,221,013 nodes: 1,900,091 of those nodes have exactly one context, and none has more than 12.
- 2,740,656 mark contexts: none more than 32 per node.
- bison's automaton: 687 states, 237 symbols.

**perigrin's ruling on it:** "We need to eliminate the map".


## Round 3

**Revision:** 3e4d0c43b, blob a8960d05a84284fb0f9fcec803d101949fd9c405. It follows the measurement after round 2 and perigrin's ruling on it, "We need to eliminate the map".

Revision 3:

- retitles the subject "the precedence constraint without maps";
- decides that the constraint keeps no hash map on its evaluation path, with its output unchanged;
- retires revisions 1 and 2's approach (§5).

**Dispatched:** assess-codebase, assess-architecture, assess-pushback (as raiser of P13–P15 only) and assess-fresh3, a new fresh participant. All four reported. assess-fresh, round 2's fresh participant, was not resumed.

How each confirmed the revision:

- assess-codebase and assess-architecture used `git hash-object` on the live file.
- assess-fresh3 says the live file matches the committed blob.
- assess-pushback read the text with `git show 3e4d0c43b:`. Its report does not record a separate blob check.

assess-fresh3 read only "Measurement after round 2" of this minute.

**assess-fresh3's prototype (recorded from its report).** In a scratch worktree at pu 8c3317079, since removed, it built a variant of `Constrain` with no map. It kept the diff, now `prototypes/no-map.diff.gz` in the evidence directory (was `assess-w3/fresh3/prototype.diff` in the session scratchpad). The variant uses:

- per-node `[]{s,t,out}` for eval and per-node `[][3]int32` for mark, by linear scan;
- a dense goto table.

Its report says the following:

- every `Alts[].Next` and `Nodes[].Alt` head was identical to the control's;
- both digest goldens pass;
- the median `Constrain` time was 40/36/35 ms against 111/97/107 ms, 2.6–3.0× faster;
- the population was 859 forests from guacamole plus std5k.pl, with load 0.4–0.9 on 4 cores.

The drafter has not run it.

### Recommendations, as declared

- **assess-codebase: modify.** "The retarget fits the code." Its wording: "The decision text needs four changes". It names C17, C18, C19 and C21, and calls C20 and C22 issue-level.
- **assess-architecture: modify.** "The direction is sound and well evidenced." It says:
  - "Nothing blocks";
  - "A18, A19, A21, A22 and A23 are each a sentence or two";
  - "A20 and A24 are perigrin's to settle";
  - A25 is optional.
- **assess-pushback (as raiser): modify on P16 only.** P13 is released, P14 and P15 are moot, and P16 is low.
- **assess-fresh3: modify.** "The direction is right, and a wall-clock prototype with identical output supports it more strongly than the profile does. Fix the exit (G1, G2, G8), restate §2's evidence (G3) and §3.2 as a property (G5). G4, G6 and G7 go into the issues."

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-codebase | No block or stand-aside declared in those terms. Its wording is quoted above. |
| assess-architecture | "Nothing blocks." No stand-aside declared. |
| assess-pushback | None declared. |
| assess-fresh3 | No block or stand-aside declared in those terms. Its wording is quoted above. |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released | moot (subject retired) | held | not re-disposed |
|---|---|---|---|---|
| assess-codebase | C14 | C1, C4, C5, C7, C8, C9, C10 (as framed; its core recurs as C17), C12, C13, C15, C16 | none | none |
| assess-architecture | A15 (by measurement), A16, A17 (by ruling) | A12, A13, A14 | none | none |
| assess-pushback | P13 | P14, P15 | none | P1, P2, P3, P10 (carried in round 2; round 3 asked the raiser only about P13–P15) |

### Raiser absent

assess-fresh was not resumed in round 3, so F1–F10 cannot be released, held or declared moot by their raiser. They stay open. The drafter decides none of them.

For each row, the ledger's status says what the row's subject is and where it stands in revision 3. On the drafter's reading of the text:

- **Subject belongs to the approach revision 3 §5 retires:** F2, F3, F4, F5, F6, F8, F10.
- **Subject not in revision 3's §4:** F7.
- **Subject persists in revision 3:** F1 (the exit, §4) and F9 (stage-key inputs, §3.4). F9's subject also appears in C19 and G7.

Whether any F row is moot is a disposition only its raiser, or a ruling, can make.

### What each participant says remains

- **assess-codebase:** C17, C18, C19 and C21 in the decision text; C20 and C22 issue-level.
- **assess-architecture:**
  - A18, A19, A21, A22 and A23, "each a sentence or two";
  - A20 and A24 for perigrin;
  - A25 optional.
- **assess-pushback:** P16. It has not re-disposed P1, P2, P3 and P10.
- **assess-fresh3:**
  - G1, G2, G3, G5 and G8 in the decision;
  - G4, G6 and G7 in the issues.

### Pending: questions put to perigrin after the round

Two questions were put to perigrin after the round's reports:

1. what "no map" forbids (A24; see also C18, C20, G5 and P16);
2. whether to rename the files (A20).

His response to both, verbatim:

> "Let's discuss."

This is recorded as pending, not as a ruling. Neither question is answered yet.

### Outcome

**Not a fixed point.** Revision 4 is owed after perigrin's answers.

- Open:
  - C17–C22;
  - A18–A25;
  - P16;
  - G1–G8;
  - F1–F10 (raiser absent).
- P1, P2, P3 and P10 stand as carried in round 2 and are not re-disposed.

## Experiments after round 3 (recorded, not findings)

Both experiments were run by agents in scratch worktrees on `pu` 8c3317079, under perigrin's "Measure first, then decide", and are recorded here by the coordinator from their reports. Their sources are in the evidence directory: `prototypes/stateless.diff.gz` (was `stateless/stateless.diff` in the session scratchpad) and `prototypes/combined.diff.gz` (was `combined/combined.diff`), with their outputs under `measurement/`.

**State-free filter (stateless-proto).** It gave identical output to `Constrain` on:

- the 1,296 referee sentences;
- 30,588 toy forests;
- 4,395 toy sentences checked against bison;
- all 2,211 aggregate files.

What it needs:

- *Reduce side:* the rule and the terminal after the span decide everywhere, except 6 pairs (`indirob` and `myterm` before `[` and `{`). The parent item settles those, leaving 0 of 257,293 keys that differ between states.
- *Shift side:* 32 keys of the form `X: term . OP term` differ between states. The spine-top item settles them, leaving 0 of 460,291 keys that differ.

The simpler "parent is the resolved rule" form agrees on perly and the aggregate. It fails on 4 toys (21 sentences, 252 forests), in each case accepting where bison refuses.

Each piece of context is needed. Without it:

| Removed | Effect |
|---|---|
| Parent item | 25 aggregate files flip to refused (e.g. T1/array.t) |
| Conflict check | 1,520 verdicts flip |
| Shift check | 625 files differ |
| Reduce check | 1,099 files differ |

With its memo still a Go map, it took 0.77 of `Constrain`'s time.

**Combined prototype (combined-proto).** Three variants against today's code (V0):

- V1: the state-free filter, with its memo in per-node lists, dense tables and no map on the evaluation path;
- V2: V1 plus the reduce-side filter at completion in recognition and in `Build`.

One disagreement was found and fixed while building, before any figure below: filtering empty-span facts in `Build` emptied a node on the toy "empty". Empty spans are now exempt.

Differences from V0:

| Check | V1 | V2 |
|---|---|---|
| Referee: verdict, grouping, kept set | 0 | 0 |
| 25,490 toy forests | 0 | 0 |
| 4,395 toy sentences vs bison | 0 | 0 |
| Aggregate: verdict, kept set, refusal reason, readings | 0 | 0 |

Recognition, V2 against V0:

- acceptance is identical on all 2,211 aggregate files;
- items per token fall from 4.70 to 4.47;
- packed alternatives fall by 5.65%.

Timing, as a ratio over recursive descent (550 ms). The run was quiet (load 1.5–2.5 against nproc 4), long-lived, interleaved with a rotating order, 3 processes:

| Stage | V0 | V1 | V2 |
|---|---|---|---|
| ParseWith | 5.62 | 4.16 | 3.92 |
| Precedence | 2.21 | 0.88 | 0.87 |
| Recognize | 0.99 | 0.96 | 0.92 |
| Build | 0.75 | 0.72 | 0.69 |

The same per process, against V0:

| Stage | V1 | V2 |
|---|---|---|
| ParseWith | 0.73 (0.72–0.74) | 0.69 (0.69–0.72) |
| Precedence | 0.41 | 0.40 |

V1's memo holds at most 10 eval contexts per node, and at most 24 across eval and mark.

**perigrin's ruling on it:** "V2: both (Recommended)". The files were renamed to `chart-stateless-precedence` at revision 4.


## Round 4

**Revision:** 6dc1d9558, blob 123012c1a5765883ded36de157c8d0e8ac764f0a. It follows the experiments after round 3 and perigrin's ruling on them, "V2: both (Recommended)".

Revision 4:

- renames the design to `docs/plans/2026-11-23-chart-stateless-precedence.md` and this minute to `docs/assessments/chart-stateless-precedence.md`;
- decides precedence without parser state (§3.1), with no hashing on the evaluation path (§3.2);
- applies bison's reduce side before the forest (§3.3).

**Dispatched:** assess-codebase, assess-architecture, assess-pushback (as raiser of P16, P1, P2, P3 and P10 only) and assess-fresh4, a new fresh participant. All four reported. assess-fresh3 was not resumed.

How each confirmed the revision:

- assess-codebase and assess-architecture used `git hash-object`.
- assess-pushback used `git rev-parse` at the commit.
- assess-fresh4 names the blob.

Of this minute:

- assess-fresh4 read only "Experiments after round 3".
- assess-architecture read its header and "Experiments after round 3".
- assess-codebase read "Experiments after round 3".

**What the participants ran (recorded from their reports; the drafter has run none of it):**

- **assess-codebase** applied `combined.diff` through `go test -overlay` and probed V0 against V2 on eleven inputs. It also ran the chart package's Port, Precedence, Waypoint and Refusal tests with the filter switched on.
- **assess-fresh4** ran the prototype's `PrepareCombined` harness, which compares V1 and V2 with V0, over three input sets:
  - 11,574 files under `/opt/corpora` and the repo;
  - 20,083 operator sentences;
  - 15,000 Leo-chain and nullable-context sentences.

  It reports that V1 never disagreed with `Constrain`, and that V2 changed the refusal reason, the byte and `Expected` on `print $fh [ … ]` shapes (H1, H2).

### Recommendations, as declared

- **assess-codebase: modify.** "The design matches the prototypes, and the prototypes are exact on everything measured. Two things must change in the text":
  - C23 (declared "Blocking");
  - C24 (classed "Missing").

  "Also in the text": C27. Issue-level: C25, C26, C28, C29.
- **assess-architecture: modify.** "The direction is sound, and the evidence is the strongest this assessment has had. The blocking issue is that revision 4 puts precedence back into recognition without the amendment text revision 2 had (A26)."
  - It marks A26 "must fix".
  - It marks A27, A29, A30 and A31 "should fix", and A28 "low".
- **assess-pushback (as raiser): modify on P17.** P16 is released; P1, P2, P3 and P10 are moot.
- **assess-fresh4: modify.** Its list, in its own words:
  1. "perigrin rules on one of two paths": allow precedence refusals to become grammar refusals, or "§3.3 does not ship. V1 alone gives most of the gain, at 0.73."
  2. "Record the amendment of design §8."
  3. "Widen the differential test as H6 describes."
  4. "State H4's limits, and add a toy that trips the generation failure and a test that forces a Leo refusal."
  5. "Report V1 and V2 separately in the final reading."

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-codebase | C23, "Blocking", in its own words. No stand-aside declared. |
| assess-architecture | It calls A26 "the blocking issue" and marks it "must fix". No stand-aside declared. |
| assess-pushback | None declared. |
| assess-fresh4 | No block or stand-aside declared in those terms. Its list is quoted above. |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released | moot |
|---|---|---|
| assess-codebase | C17, C18, C19, C20, C21, C22 | none |
| assess-architecture | A18, A20 (residue A29), A21, A22, A23, A24 (by ruling and measurement), A25 (optional, not adopted) | A19 (claim removed) |
| assess-pushback | P16 | P1, P2, P3, P10 (§6) |

### Raiser absent

- **assess-fresh3 was not resumed**, so G1–G8 stay open with no one to release them.
  - For each, the ledger's status says where its subject stands in revision 4's text. On the drafter's reading, revision 4's text speaks to the subjects of G2, G4, G5, G7 and G8.
  - G1 and G3 persist in part. G6 is not stated in revision 4, and its raiser called it issue-level.
  - The drafter decides none of them.
- **F1–F10 remain raiser absent.** Their notes are updated to revision 4. Revision 4 reinstates the reduce-side filter, so the subjects of F2, F3, F4, F5, F6, F8 and F10 are in scope again.

### The drafter's edit for A29

At the coordinator's direction, the drafter changed the minute's header in round 4. It changed:

- the title, to "Assessment: Chart parser: precedence without parser state";
- both ABOUTME lines;
- the Subject line, to the renamed path;
- and it added the revision table, which gives old and new paths.

This is housekeeping on the record, not a disposition of A29. A29 stays open for its raiser.

### What each participant says remains

- **assess-codebase:** C23 and C24, and C27, in the text; C25, C26, C28 and C29 issue-level.
- **assess-architecture:** A26 (must fix); A27, A29, A30 and A31 (should fix); A28 (low).
- **assess-pushback:** P17.
- **assess-fresh4:** H1–H6, through the five steps quoted above.

### Ruling after the round's reports

Recorded verbatim, as a ruling and not a finding. Asked whether revision 5 decides V1 or V2, given H1 and H2, perigrin chose:

> "V1 now, V2 measured after (Recommended)"

The coordinator's statement of its meaning: ship the exact state-free filter first. The recognizer filter becomes a follow-up decision with its own measurement, once its `Expected` and refusal-reason effects are designed for.

The drafter does not say which rows this ruling answers. Each row's raiser disposes of it.

### Outcome

**Not a fixed point.** Revision 5 is owed.

- Open:
  - C23–C29;
  - A26–A31;
  - P17;
  - H1–H6.
- Raiser absent: F1–F10 and G1–G8.


## Round 5

**Revision:** 2782fd181, blob d97415a147b923df589b27bed3cad0317f6bbfe2. It follows perigrin's ruling after round 4, "V1 now, V2 measured after (Recommended)". Revision 5:

- decides only the forest-side judgement, with no parser state and no hashing;
- leaves recognition and the built forest unchanged;
- defers the recognizer's reduce-side filter (V2) by ruling (§6);
- names an evidence directory, `docs/assessments/chart-stateless-precedence/`.

**Dispatched:** assess-architecture, assess-codebase, assess-pushback (as raiser of P17 only) and assess-fresh5, a new fresh participant. All four reported. assess-fresh4 was not resumed.

How each confirmed the revision:

- assess-architecture and assess-codebase used `git hash-object`.
- assess-pushback used `git rev-parse` at the commit.
- assess-fresh5 names the blob, and reports it is the same at 2782fd181 and at 157b0f820.

assess-fresh5 read neither this minute nor `reports/`.

**What the participants checked or ran (recorded from their reports; the drafter has run none of it):**

- **assess-architecture** compared the 15 files under `reports/` with the scratchpad originals using `cmp`, and found all identical. It found that the three gzipped diffs pass `gzip -t` and apply to `git archive 8c3317079 internal/chart`. It also found `no-map.diff.gz` byte-identical to `fresh3/prototype.diff`.
- **assess-codebase:**
  - found the two prototype diffs byte-identical to those it reviewed in round 4;
  - generated V1's mixed-pair list on the committed tables, with six pairs, as §3.1 says;
  - sketched a toy grammar that makes generation fail, but did not run it.
- **assess-fresh5** applied `combined.diff` and ran the combined harness over the 2,211-file aggregate:
  - V1 against V0 differs in 0 verdicts, 0 kept sets, 0 refusal reasons and 0 readings;
  - one process, with load ≤ 1.5 on 4 cores;
  - V1's ParseWith is 0.727 of V0's, and its precedence stage 0.40;
  - V2's ParseWith came out at 0.723 of V0's, not 0.69.

### Recommendations, as declared

- **assess-architecture: accept, with edits.** "Accept, with three small edits that do not change the decision. A33 needs perigrin's reading of one ruling. A34 and A35 are each a sentence. None of them blocks."
- **assess-codebase: modify.** "Small changes; nothing blocking."
  - C30 and C31 in the text.
  - C32 and C33 issue-level.
  - "§3.1, §3.2 and §3.5 are accurate against the prototype. §3.4 is achievable."
- **assess-pushback (as raiser): modify on P18.** P17 is released.
- **assess-fresh5: modify.** "The engineering claim holds. I reproduced exact agreement and the 0.73 independently. Before acceptance, fix J1, J2, J4 and J5, and trim J3; the rest is wording." J10 is raised in support.

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | "None of them blocks." No stand-aside declared. |
| assess-codebase | "Nothing blocking." No stand-aside declared. |
| assess-pushback | None declared. |
| assess-fresh5 | No block or stand-aside declared in those terms. It asks for J1, J2, J4 and J5 "before acceptance". |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released | moot |
|---|---|---|
| assess-architecture | A28, A29 (residue A34), A30 (residue A34), A31 | A26, A27. The raiser says both carry to the V2 decision. |
| assess-codebase | C24, C25 | C23, C26, C27, C28, C29 (V2 deferred) |
| assess-pushback | P17 | none |

### Raiser absent

assess-fresh4 was not resumed, so H1–H6 stay open with no one to release them. For each, the ledger's status says where its subject stands in revision 5. On the drafter's reading:

- The subjects of H1, H2 and H5 are V2's, which revision 5 §6 defers.
- Revision 5's text speaks to the subjects of H3 ("What this does not amend") and H4 (§3.2).
- H6 has two parts: one concerns V2, and the other concerns the test's width, which §3.4 widens.

The drafter decides none of them. F1–F10 and G1–G8 remain raiser absent as recorded in round 4.

### What each participant says remains

- **assess-architecture:** A33 (for perigrin, or a sentence), A34 (should fix), A35 (low).
- **assess-codebase:** C30 and C31 in the text; C32 and C33 issue-level.
- **assess-pushback:** P18.
- **assess-fresh5:**
  - J1, J2, J4 and J5 before acceptance, and J3 trimmed;
  - J6–J9 as wording;
  - J10 in support.

### Ruling after the round's reports

Recorded verbatim, as a ruling and not a finding. Asked how revision 5 treats "3.92" (A33, P18), perigrin chose:

> "Target for V1 then V2 (Recommended)"

The coordinator's statement of its meaning:

- 3.92 is where V1 and the later V2 decision should land together;
- this decision's exit stays the relative ≤ 5 rule;
- allocation work waits until both are in.

The drafter does not say which rows this ruling answers. Each row's raiser disposes of it.

### Outcome

**Not a fixed point.** Revision 6 is owed.

- Open:
  - A33, A34, A35;
  - C30–C33;
  - P18;
  - J1–J10.
- Raiser absent: F1–F10, G1–G8 and H1–H6.


## Round 6

**Revision:** 788778ed0, blob 60327eb7d7e379443281e0623b5d5f586a382f9c. Revision 6:

- answers round 5;
- lists only the rulings in force in §1;
- names the reference `forest.ConstrainReference` and the target `make chart-precedence-reference` in §3.4;
- completes the evidence directory with `INDEX.md`, raw timing runs and harnesses.

**Dispatched:** assess-architecture, assess-codebase, assess-pushback (as raiser of P18 only) and assess-fresh6, a new fresh participant. All four reported. assess-fresh5 was not resumed.

How each confirmed the revision:

- assess-architecture and assess-codebase used `git hash-object`.
- assess-pushback used `git rev-parse` at the commit.
- assess-fresh6 names the blob and says it is the same at 0d2b31610.

assess-fresh6 read neither this minute nor `reports/`.

**What the participants checked or ran (recorded from their reports; the drafter has run none of it):**

- **assess-architecture** listed the evidence directory (89 files) and recomputed V1 and V2 over V0 per process from `combined-runs/result-run{1,2,3}.txt`. It also found the seeds and the 15,000 loop in `harness/fresh4/`.
- **assess-codebase** ran one timing probe through `go test -overlay`, giving generation at about 90 ms against about 310 ms.
- **assess-fresh6** re-ran the round-4 generators on the combined prototype:
  - seed 4 gives 185 precedence-refused sentences of 20,000 random, plus 2 of the 83 handcrafted;
  - seed 44 gives 989 of 15,000;
  - the 11,574 corpus files give 0 differences after remapping paths.

  It also printed the six mixed pairs.

### Drafter housekeeping for A34's residue

At the coordinator's direction, the drafter repointed the minute's paths for reports, prototypes, harnesses, measurements and the C1 spike report to `docs/assessments/chart-stateless-precedence/`. Each original scratchpad path is kept in parentheses. This was committed at 2c70194e6.

- Ledger rows still cite the paths their raisers used.
- The author's quoted statement stays verbatim, with the new location given after it.

This is housekeeping on the record, not a disposition. assess-architecture released A34 in this round for the subject, and assigned the minute's pointers to the drafter.

### Recommendations, as declared

- **assess-architecture: accept, with edits.** "The decision is ready. Its exit, stops, keys and scope are settled, and the evidence directory is now an adequate audit record."
  - A36 and A37 concern the next decision.
  - "A38 is a question that needs an answer before refinement."
  - A39 is cosmetic.
- **assess-codebase: modify.** "One text change: **C34.**"
  - Low: C35, C36, C37.
  - C33 carried to refinement.
  - "The rest of §3 is accurate and workable on the code as it stands."
- **assess-pushback (as raiser): modify on P19.** P18 is released.
- **assess-fresh6: modify.**
  - "K1 is the only finding that would bite on shipping."
  - "K2, K5 and K6 are wording-sized. K3 and K4 are evidence hygiene."
  - "The substance (keys, guard, reference, inputs) held up under re-run."

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | No block or stand-aside declared. A38 "needs an answer before refinement". |
| assess-codebase | No block or stand-aside declared. |
| assess-pushback | None declared. |
| assess-fresh6 | No block or stand-aside declared. |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released | carried to refinement |
|---|---|---|
| assess-architecture | A33 (meaning raised as A36), A34 (for the subject), A35 (question raised as A38) | none |
| assess-codebase | C30, C31, C32 | C33 (issue-level) |
| assess-pushback | P18 | none |

### Raiser absent

assess-fresh5 was not resumed, so J1–J10 stay open with no one to release them.

- For each, the ledger's status says where revision 6's text stands. On the drafter's reading, revision 6's text speaks to the subject of every J row.
- J4's question of where the aggregate part runs is the subject of P19, A38, C34 and K1.
- J10 was raised in support.

The drafter decides none of them. F1–F10, G1–G8 and H1–H6 remain raiser absent as recorded earlier.

### What each participant says remains

- **assess-architecture:** A36 (for perigrin), A37 (should fix), A38 (question), A39 (low).
- **assess-codebase:** C34 (one text change); C35, C36 and C37 (low); C33 carried to refinement.
- **assess-pushback:** P19.
- **assess-fresh6:**
  - K1;
  - K2, K5 and K6 (wording);
  - K3 and K4 (evidence hygiene);
  - K7, K8 and K9.

### Rulings after the round's reports

Recorded verbatim, as rulings and not as findings. The gloss after each is the coordinator's statement of its meaning.

1. **On where the aggregate check runs**, asked after P19, A38 and C34:

   > "Split it (Recommended)"

   That is:
   - `make test` runs the in-tree part: T1, Guacamole, std5k, the referee, the toys and the seeded refused sentences.
   - The perl.git part runs in `make chart-precedence-reference` and in CI's corpus-backed Ratchets job, with `PERL5_CORPUS` set. There, a missing corpus fails.

2. **On 3.92**, asked after A36:

   > "Drop the number"

   That is: there is no numeric V1-plus-V2 target, and the V2 decision sets its own after V1's verdict reading.

The drafter does not say which rows these rulings answer. Each row's raiser disposes of it.

### Outcome

**Not a fixed point.** Revision 7 is owed.

- Open:
  - A36–A39;
  - C34–C37;
  - P19;
  - K1–K9.
- C33 is carried to refinement by its raiser.
- Raiser absent: F1–F10, G1–G8, H1–H6 and J1–J10.

## Round 7

**Revision:** 9afff6c63, blob ce616721c9ed717dc119cf4c5f3329daf802f4cd. Revision 7 answers round 6. It:

- records "Drop the number" and "Split it" in §1;
- tabulates §2's figures by source;
- gives §3.4 a table of inputs and where each runs;
- names §4's before-state and new work;
- adds stops for a session that cannot be quiet, for the CI timeout and for the budget itself.

**Dispatched:** assess-architecture, assess-codebase, assess-pushback (as raiser of P19 only) and assess-fresh7, a new fresh participant. All four reported. assess-fresh6 was not resumed.

How each confirmed the revision:

- assess-architecture and assess-codebase used `git hash-object`.
- assess-pushback used `git rev-parse` at the commit.
- assess-fresh7 names the blob and says it is the same at 1ff6e3f6a.

assess-fresh7 read neither this minute nor `reports/` before writing.

Round 7 is the seventh of at most eight rounds.

**What the participants checked or ran (recorded from their reports; the drafter has run none of it):**

- **assess-architecture** recomputed every §2 figure from the evidence files and confirmed `reports/round6-architecture.md` byte for byte. It also traced in `main.go` that the stage keys run inside CI's 20-minute timeout.
- **assess-codebase** traced `corpus()`, `supersetPopulation`, the committed labels and every caller of `.Constrain(`. It found the in-tree part buildable without perl.git, and the caller list "complete and correct".
- **assess-fresh7** applied combined.diff and timed `NewStateless` five times: 105–137 ms. It also confirmed the six mixed pairs by a run.

### Recommendations, as declared

- **assess-architecture: accept.** "Nothing of mine remains held. Four new findings are low or textual, and none blocks."
  - "A42 is one clause and worth making before refinement."
  - "A40, A41 and A43 are wording."
  - "None of the four changes what the decision builds, measures or stops on."
- **assess-codebase: accept.** "C38 is an editorial fix that refinement can carry. C33 goes to the issues. From the codebase lens the assessment has reached its fixed point." It says it "raised nothing new that needs another assessment round".
- **assess-pushback (as raiser):** "holding nothing". It released P19 and raised P20, which it rates low and does not hold acceptance on. Its report states no accept or modify, and the drafter infers none.
- **assess-fresh7: modify.** "The direction is sound and the figures hold. L1–L3 are internal contradictions or under-specifications that would surface mid-milestone as stops. L2 is the one that would."

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | "Held: none." "None blocks." No stand-aside declared. |
| assess-codebase | "Nothing remains held." No stand-aside declared. |
| assess-pushback | Holding nothing. P20 is not held against acceptance. |
| assess-fresh7 | No block or stand-aside declared in those terms. Its verdict is modify. |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released | moot | carried to refinement |
|---|---|---|---|
| assess-architecture | A37, A38 (by ruling and text), A39 | A36 (by ruling) | none |
| assess-codebase | C34, C35, C36, C37 | none | C33 |
| assess-pushback | P19 | none | none |

### Raiser absent

assess-fresh6 was not resumed, so K1–K9 stay open with no one to release them.

- For each, the ledger's status says where revision 7's text stands. On the drafter's reading, revision 7's text speaks to the subject of every K row.
- K9's subject is the evidence index, and the note quotes assess-architecture's reading of it.

The drafter decides none of them. F1–F10, G1–G8, H1–H6 and J1–J10 remain raiser absent as recorded earlier.

### What each participant says remains

- **assess-architecture:** A42 ("should fix"); A40, A41 and A43 (wording); none held.
- **assess-codebase:** C38 (editorial, for refinement) and C33 (carried to refinement); none held.
- **assess-pushback:** P20 (low; not held against acceptance).
- **assess-fresh7:** L1–L9, with L1–L3 named as the ones that "would surface mid-milestone as stops".

### Rulings after the round's reports

Recorded verbatim, as rulings and not as findings. The gloss after each is the coordinator's statement of its meaning.

1. **On L3**, the paths §6 defers:

   > "Rule them in (Recommended)"

   That is: `Makefile` (the `chart-precedence-reference` target) and `.github/workflows/ci.yml` (the Ratchets job step) are ruled in for the issue that adds that target.

2. **On A43**, when allocation and the other stages may start:

   > "V2's verdict reading (Recommended)"

   That is: allocation and the other stages wait until V2 is built and measured.

The drafter does not say which rows these rulings answer. Each row's raiser disposes of it.

### Outcome

**Not a fixed point.** Round 7 raised new findings. Revision 8 is owed, for round 8, the last within the bound of eight.

- Open:
  - A40–A43;
  - P20;
  - L1–L9.
- Editorial, for refinement: C38.
- Carried to refinement: C33.
- Raiser absent: F1–F10, G1–G8, H1–H6, J1–J10 and K1–K9.
- No participant holds a finding against acceptance:
  - assess-architecture says "Held: none";
  - assess-codebase says "Nothing remains held";
  - assess-pushback says it holds nothing.
  - assess-fresh7 declares modify and names L1–L3. It declares no block.

## Round 8

**Revision:** 336067853, blob ed91e46239b29283ad5c82cd390411621b52a2cd. Revision 8 answers round 7. It:

- adds the rulings "V2's verdict reading" and "Rule them in (Recommended)" to §1;
- restates §2 by section;
- gives §3.4 the boundary each caller uses, the in-tree row with edit5k and parsent, and "new" on each new test;
- defines §4's measurement and adds the per-pass preparation hook;
- adds the moved digest or count key to §5's first stop;
- lists the v2 §11 departure.

**Dispatched:** assess-architecture, assess-codebase, assess-pushback (as raiser of P20 only) and assess-fresh8, a new fresh participant. All four reported. assess-fresh7 was not resumed.

How each confirmed the revision:

- assess-architecture and assess-codebase used `git hash-object`.
- assess-pushback used `git rev-parse` at the commit.
- assess-fresh8 names the blob and says it is the same at 146411199.

assess-fresh8 read the minute and `reports/` only after writing. Round 8 is the last round within the bound of eight.

**What the participants checked or ran (recorded from their reports; the drafter has run none of it):**

- **assess-codebase** traced `Protocol.pass`, `Ratio`, `measureStages`, `judge` and `narrow`, and `libraryOperators` against revision 8's text. It found each statement accurate. It corrected its own earlier claim that `libraryOperators` only reads the forest.
- **assess-architecture** traced the new 105–137 ms figure to `reports/round7-fresh.md`. It checked the verdict reading against the README and chart-recovery.
- **assess-fresh8** recomputed every §2 figure from the indexed outputs and found them matching. It checked that every named symbol exists or is marked new.

### Drafter housekeeping: round 7's rows restored

The round-7 minute as committed at 146411199 had lost three things:

- the round-7 ledger rows (A40–A43, C38, P20, L1–L9);
- the round-7 cross-reference table;
- the round-7 distribution table.

The round-7 section and the round-7 status updates were present. The drafter restored the 15 rows and both tables from its round-7 draft, with their text unchanged. Their statuses now carry round 8's dispositions. This is housekeeping on the record, not a disposition.

### Recommendations, as declared

- **assess-architecture: accept.** "I hold nothing. From my lens this revision is a fixed point. A44 and A45 are new and low. Refinement can carry them, and neither is a block." It says: "Fixed point: confirmed from the architecture lens."
- **assess-codebase: accept.** "I hold nothing." "I **confirm a fixed point** from the codebase lens." C39 is "an editorial note for refinement", which "is not a round finding".
- **assess-pushback (as raiser): accept.** "I hold nothing and raise nothing new." P20 is released.
- **assess-fresh8: modify.** "M1 and M2 are one-sentence edits that close the exit/stop gap. M3 and M4 are one clause each. The rest belongs in issues or needs no change."
  - M1–M4 are marked for the decision.
  - M5 is marked for the decision, low.
  - M6 is a view with no change.
  - M7 and M8 are marked for the implementing issue.

### Blocks and stand-asides, as declared

| participant | declared |
|---|---|
| assess-architecture | "Held: none." A44 and A45 are "not conditions of acceptance". No stand-aside declared. |
| assess-codebase | "I hold nothing." No stand-aside declared. |
| assess-pushback | Holds nothing. |
| assess-fresh8 | No block or stand-aside declared in those terms. Its verdict is modify, with M1–M4 for the decision. |

The drafter converts none of these into a block or a release.

### Dispositions of earlier findings, by their raisers

| raiser | released |
|---|---|
| assess-architecture | A40, A41, A42, A43 (by ruling) |
| assess-codebase | C38 |
| assess-pushback | P20 |

C33 stays carried to refinement, as its raiser declared in rounds 6 and 7.

### Raiser absent

assess-fresh7 was not resumed, so L1–L9 stay open with no one to release them.

- For each, the ledger's status says where revision 8's text stands.
- On the drafter's reading, revision 8's text speaks to the subject of every L row.

The drafter decides none of them. F1–F10, G1–G8, H1–H6, J1–J10 and K1–K9 remain raiser absent as recorded earlier.

### What each participant says remains

- **assess-architecture:** A44 and A45, optional, for refinement. None held.
- **assess-codebase:** C39, editorial, for refinement, and C33, carried to refinement. None held.
- **assess-pushback:** nothing.
- **assess-fresh8:** M1–M4 for the decision; M5 for the decision, low; M6, no change; M7 and M8 for the implementing issue.

## Outcome of the session

**No fixed point within the bound of eight.** Round 8 was the last round the bound allows.

- **Three participants confirm a fixed point from their lenses:**
  - assess-architecture: "Fixed point: confirmed from the architecture lens";
  - assess-codebase: "I **confirm a fixed point** from the codebase lens";
  - assess-pushback holds nothing and raises nothing new.
- **One participant does not.** assess-fresh8 recommends modify and raised new findings it places in the decision. A fixed point needs every participant to report nothing held and nothing new, so the session did not converge.

This non-convergence is itself a finding for perigrin. The minute records it rather than resolving it.

**What remains open at the bound:**

| kind | rows | as its raiser places it |
|---|---|---|
| decision-level | M1, M2, M3, M4 | M1 and M2 "one-sentence edits"; M3 and M4 "one clause each" (M3 "mostly the implementing issue") |
| decision text, low | M5 | label figures "reported, not kept", or index round 7's output |
| issue-level | M7, M8 | the implementing issue |
| for refinement | C39 (editorial), A44 and A45 (optional), C33 (carried) | not conditions of acceptance, in their raisers' words |
| view, no change | M6 | "no change is needed" |
| raiser absent | F1–F10, G1–G8, H1–H6, J1–J10, K1–K9, L1–L9 | no participant can release them; the drafter's notes say where the latest revision's text stands on each |

**Who decides the close.** It is perigrin's decision. The coordinator has put it to him. The minute records no ruling on it yet, and the drafter neither proposes nor judges one.

## perigrin's ruling on the close

The question was put to perigrin after round 8's outcome. Three participants accepted and held nothing. Round 8's fresh participant raised M1–M4, which it described as one-sentence decision fixes, and the M5 label.

perigrin's answer, verbatim:

> "Apply fixes, accept (Recommended)"

The coordinator states what this means; the drafter judges nothing. Revision 9 applies the fixes as follows:

| Finding | What revision 9 does |
|---|---|
| M1 | "A stop overrides the exit: if both hold, the milestone stops." |
| M2 | The first stop now fires when any §3.4 check fails, including the reference disagreeing with bison, a failed assertion, or the perl.git target failing or timing out in CI. |
| M3 | A slow check's bulk moves behind a build tag that `make test` still runs, never to the perl.git target. |
| M4 | "What this does not amend" adds v2 §9.5's no-new-keys rule, departed from for the three stage keys by ruling. |
| M5 | The generator counts and the table-build cost are labelled "reported, not kept". |

perigrin accepted the decision at revision 9.

Still open, for refinement and the implementing issues:

- the issue-level findings M7, M8 and C33;
- C39, editorial;
- A44 and A45, optional.

The raiser-absent rows stay as recorded.
