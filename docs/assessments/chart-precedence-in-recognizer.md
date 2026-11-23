<!-- ABOUTME: Minute of the crochet:assess session on the design for applying precedence inside the chart recognizer. -->
<!-- ABOUTME: Covers rounds 1 to 3: the finding ledger, cross-references, recommendations and declarations, rulings and pending questions, outcomes, the measurement and the author's statement. -->

## Assessment: Chart parser: precedence in the recognizer

Recommendation: pending (round 1: three modify; round 2: four modify; round 3: four modify; revision 4 owed after perigrin's answers)

- **Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.
- **Revision assessed in round 1:** design commit 7b264fa9c, blob 52c82af16a738eca6e03cdb050e9c167903212c7. Each participant read that blob, untracked in worktree `cr-close`, and checked it with `git hash-object`. It is byte-identical to the blob at 7b264fa9c.
- **Revision assessed in round 2:** commit c270bbfa3, blob 68371487a68e3f8c1661a80ea830356228a3a163 (revision 2, retitled "bison's reduce-side precedence in the recognizer").
- **Revision assessed in round 3:** commit 3e4d0c43b, blob a8960d05a84284fb0f9fcec803d101949fd9c405 (revision 3, retitled "the precedence constraint without maps"). The file keeps its name (see A20).
- **Decision-maker:** perigrin.
- **Drafter:** assess-pushback, from round 2 on. The drafter records and holds no view. In round 1 it was the pushback participant and raised P1 to P12. From round 2 on it does not count as an independent participant.
  - In round 2, at the coordinator's direction, it also acted as raiser of P1 to P12. As raiser it released, held or carried them, and it raised P13 to P15.
  - Those round-2 P rows are the raiser's, not the drafter's.
  - In round 3 it acted as raiser of P13 to P15 only, and raised P16.
  - The independent participants are assess-codebase and assess-architecture (rounds 1 to 3), assess-fresh (round 2 only; not resumed in round 3) and assess-fresh3 (round 3).

Each row's section reference is to the revision named in its revision column. IDs keep each participant's own prefix:

- A: assess-architecture, the architecture and direction lens;
- C: assess-codebase, the codebase lens;
- P: assess-pushback, the spec-quality lens;
- F: assess-fresh, a fresh participant in round 2. It read the subject without the prior findings, and it wrote its report before reading this minute. It was not resumed in round 3, so its rows are marked raiser absent.
- G: assess-fresh3, a fresh participant in round 3. Of this minute it read only "Measurement after round 2".

"C1" in a ledger row is assess-codebase's finding. The chart-recovery spike is always written "the C1 spike".

The full reports are at the paths below, in the session scratchpad (`/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/`):

- `assess-w3/round1-architecture.md`
- `assess-w3/round1-codebase.md`
- `assess-w3/round1-pushback.md`
- `assess-w3/round2-architecture.md`
- `assess-w3/round2-codebase.md`
- `assess-w3/round2-pushback.md`
- `assess-w3/round2-fresh.md`
- `assess-w3/round3-architecture.md`
- `assess-w3/round3-codebase.md`
- `assess-w3/round3-pushback.md`
- `assess-w3/round3-fresh.md`, with its prototype at `assess-w3/fresh3/prototype.diff` and harness at `assess-w3/fresh3/exp_test.go.txt`

## Ledger

Rows from round 1 were raised against `7b264fa9c (52c82af1)`, rows from round 2 against `c270bbfa3 (68371487)`, and rows from round 3 against `3e4d0c43b (a8960d05)`.

Statuses as each raiser declared them:

- **released**: the raiser names the text it releases against.
- **held**: the raiser keeps the finding open.
- **carried**: carried, by revision 2's §6, to the later design of the value on items. The raiser agrees in each case.
- **moot**: the raiser declares that the row's subject was retired by a later revision.
- **open**: no disposition yet.
- **raiser absent**: the raiser was not resumed, so no one can release the row. Where a retargeting may bear on such a row, the drafter says what the row's subject is and where it stands in revision 3. That note is not a disposition.

Statuses marked r2 or r3 give the round of the latest disposition.

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
| P1 | assess-pushback | 1 | §4.2 | The implementer chooses between one value per item, which is wrong where a state holds several dotted rules, and one value per dotted rule, which is a different design with a different cost to set identity. | 51 of 772 states complete more than one rule (`states.txt`). The recovery plan (branch `docs/chart-recovery-plan`, §C, C1): "A single value per item could not kill one rule inside a state". | carried (r2); not re-disposed by the raiser in round 3. Revision 3 §5 puts the value on items out of scope | 7b264fa9c (52c82af1) |
| P2 | assess-pushback | 1 | §4.2 | A derivation the table admits is lost after its item has already completed, and shows as false refusals (`rd-accepted-chart-refused` above 0). | `builder.add` (recognize.go:376) deduplicates with `isNew` and never revisits. Recovery plan C1: "An item whose value grows must complete again". | carried (r2); not re-disposed by the raiser in round 3. Revision 3 §5 puts the value on items out of scope | 7b264fa9c (52c82af1) |
| P3 | assess-pushback | 1 | §4.2 | Under one reading, `$a << ($b < $c)` survives recognition. | `%prec` unit rules (perly.go:354, 360, 440, 442, 464) are both "unit" and "operator" rules. SHIFTOP/CHRELOP resolves as reduce (26 CHRELOP reduce pairs, bison run). §3 counts (rule, token) pairs, not level expressibility. | carried (r2); not re-disposed by the raiser in round 3. Revision 3 §5 puts the value on items out of scope | 7b264fa9c (52c82af1) |
| P4 | assess-pushback | 1 | §4.2, §5.4, §5.5 | A resolution the generator cannot map is dropped or mis-mapped silently. 43 resolutions' entries are unrefereed. | bison 373 rules against the chart's 350. One reduce resolution names POSTJOIN, which perly.go:796 omits. Referee coverage 1,304 of 1,347. No stated refuse, review-list or drift-check path. | released (r2, against §4.4, §2) | 7b264fa9c (52c82af1) |
| P5 | assess-pushback | 1 | §6, §8.4 | The waypoint cannot finish, or the exit is lowered partway through. | 4.79 = 0.655 + 0.790 + 0.662 + about 2.68. Lex is untouched, and resolve and readings are out of scope (§7). Reaching 2.5 needs the post-forest share at or below about 1.1, mostly from removing `Constrain`, which no key times. The 115 ms probe is std5k only. | released (r2, against §1 goal, §6); see P13 | 7b264fa9c (52c82af1) |
| P6 | assess-pushback | 1 | §4.3, §6 (stop 4) | Stop 4 is a false alarm or is never checked. | `Constrain` runs until the last issue, so no fall below 4.79 is expected "once both tables are in". No reading protocol, against quiet repeats of 5.18 and 5.11, 8.56 and 8.39. Stop 3's causal clause cannot be measured. | released (r2, against §5 ¶2, §6 third stop) | 7b264fa9c (52c82af1) |
| P7 | assess-pushback | 1 | §4.3, §8.3 | The last issue has no runnable criterion, or it removes `Constrain` on evidence from generated sentences alone. | `precedence-disagree` is defined over the generated sentences (README key table). `Constrain` has no report-only mode (precedence.go:322 returns a bool and unlinks). "The waypoint's last issue" names no milestone. | released (r2, dissolved by §7) | 7b264fa9c (52c82af1) |
| P8 | assess-pushback | 1 | §1, §5.1 | The merged code contradicts the live README and the binding design, and no superseding record says so. | README: "No fact is added to an Earley item (spec §3 rule 12, §10)". The recovery plan says a successor design must amend rule 12 and §10, with C1's numbers as evidence. The design does not cite C1. "Option B/C" appear in no committed document. "Follows" cites v1. | released (r2, against header, §1, §2) | 7b264fa9c (52c82af1) |
| P9 | assess-pushback | 1 | §5.2, §8.2 | With Leo on, the reduce-side filter is skipped for intermediate links, and `leo-disagree` (acceptance only) does not notice. | `complete` calls `leo()` before `awaiting`. The lookahead case relies on "FOLLOW only shrinks up a chain" (leo.go), and no precedence analogue is stated. Recovery plan C1: the memo stores no value, plus a guard on values. | released (r2, against §4.5) | 7b264fa9c (52c82af1) |
| P10 | assess-pushback | 1 | §4.2, §8.1 | The "no operator" bit collides with a level, or the type changes partway through implementation. | perly.go:744-806 declares 32 levels, PLUGIN_* among them. Adding the reserved bit makes 33. | carried (r2); not re-disposed by the raiser in round 3. Revision 3 §5 puts the value on items out of scope | 7b264fa9c (52c82af1) |
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
| F1 | assess-fresh | 2 | §6 | The exit can pass with B doing nothing, or fail because of drift. | Baseline 4.7915. `testdata/ratchet/checkpoint3.txt` 5.1765 and 5.1051 (c4eb1e46). Records: recognize about 0.77 and about 0.90 on a5498fa4; stage sum 3.638 to 3.13 after a host move. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (the exit) persists in revision 3 §4, under the ruling "Relative and ≤ 5" | c270bbfa3 (68371487) |
| F2 | assess-fresh | 2 | §3, §6 | Seven days buys a few percent, and §6's question about C is answered only through noise. | Probe on 1,676 files (not the aggregate; load 0.7–1.1; single runs): B refuses 7.0% of packed alternatives with dependents, and `Constrain` drops 10.5%. `Constrain` on the B-pruned forest takes ×0.87–0.97 of its time; on its own output, ×0.94–1.00. `Constrain` is about 46% of `ParseWith` time. Estimate: 4.79 to about 4.6. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (the speed premise of B) belongs to revisions 1–2; revision 3 §5 retires B, citing the measurement after round 2 | c270bbfa3 (68371487) |
| F3 | assess-fresh | 2 | §4.2 | An implementer writes the equality test, it goes red on the first nonassoc chain, and the contract is rewritten during execution. | `$a <=> $b <=> $c;` (`TestParseAppliesPrecedence`) and `1 .. 2 + 3 .. 4;`: the root stays live after B. Fuzz toy grammars declare no precedence. `recognize` cannot import `forest`. v2 §3 rule 11. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (the reference contract) belongs to revisions 1–2, which revision 3 §5 retires | c270bbfa3 (68371487) |
| F4 | assess-fresh | 2 | §4.1, §4.3, §5 | §6's "a held key moves" stop becomes ambiguous, and goldens are regenerated ad hoc. | Sets change. Tests holding sets or acceptance: `TestReferenceRecognizer` (acceptance, `Inserted`, `Expected` under every Options combination), `TestReferenceAgreesOnTruncatedFiles`, `TestSetDigestGolden`, `FuzzRecognize`. `Expected` re-closes via `b.complete(nil)` over pruned origin sets. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (tests whose contracts move with a recognizer filter) belongs to revisions 1–2, which revision 3 §5 retires | c270bbfa3 (68371487) |
| F5 | assess-fresh | 2 | §4.1 | Raised as support: exactness holds only while the table is computed on `NewPrecedence`'s automaton, not copied from bison's report. | Probe: 0 B-refused alternatives that `Constrain` keeps, and 0 verdict flips over 1,558,495 alternatives. Inserted-`;` contexts not modelled. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (exactness of the reduce-side filter) belongs to revisions 1–2, which revision 3 §5 retires | c270bbfa3 (68371487) |
| F6 | assess-fresh | 2 | §3, §4.4 | The figures are bison's, but the table will be the chart's. Minor. | Chart automaton: 629 shift and 8 error pairs, against bison's 670 and 10; the difference is about the 44 POSTJOIN lines in `bison-conflicts.txt`. §3 omits the 10 error resolutions. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (the filter's table counts) belongs to revisions 1–2, which revision 3 §5 retires | c270bbfa3 (68371487) |
| F7 | assess-fresh | 2 | §6 (stops) | The recognize-stage stop is likely to trip for reasons unrelated to B. | C1 spike: items barely move, and B adds a lookup per completion. Records: recognize drifts about 15% between sessions on one commit. Same-session before and after is not stated. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (the recognize-stage stop) is not in revision 3 §4 | c270bbfa3 (68371487) |
| F8 | assess-fresh | 2 | §4.5 | Granularity left to the implementer. Issue-level. | `complete` works per (state, LHS) (`g.lhs[state]`); the forest drops facts per rule. Leo memos are keyed by (node, symbol), independent of `la`. At a slipper node `la` holds the inserted `;`. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (§4.5's filter granularity) belongs to revisions 1–2, which revision 3 §5 retires | c270bbfa3 (68371487) |
| F9 | assess-fresh | 2 | §5 | The precedence and resolve stage keys would time mutated forests. Issue-level. | `Constrain` and resolve mutate the forest in place, so forests must be cloned outside the timed span, as `measureStages` does for lattices and charts. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (fresh inputs for the precedence and resolve stage keys) persists in revision 3 §3.4 | c270bbfa3 (68371487) |
| F10 | assess-fresh | 2 | §2 | Rule 6 is amended without being listed. Minor. | v2 §3 rule 6 places the constraint "after the chart". The 2026-11-14 ruling on the spec's standing (chart-perly-w2-inputs Records) permits the amendment. | open; raiser absent (assess-fresh not resumed in round 3). Drafter's note, not a disposition: the subject (amending v2 rule 6) belongs to revision 2's §2; revision 3 amends nothing in v2 | c270bbfa3 (68371487) |
| C17 | assess-codebase | 3 | §3.1 | A changed kept set on a file outside the sample passes every listed check. | `forest-digest.golden` hashes `Build`'s forest before the constraint, so this change cannot move it. `resolved-forest-digest.golden` covers 438 sample files against a 2,211-file aggregate, and resolve's `prune` can hide an over-kept alternative. The referees compare a verdict and one grouping. Proposed: keep today's `Constrain` as a test-only oracle and compare verdict and `kept` per file. Shared memo slices (`seq` returns `right`) are the realistic corruption path. | open | 3e4d0c43b (a8960d05) |
| C18 | assess-codebase | 3 | §3.2, §3.3 | §3.2 forbids every design that meets §3.3, and the overflow path a wrong answer could hide in goes untested. | Contexts per node are unknown before the pass (theoretical bound 687 × 238). Any per-node storage with overflow grows. Proposed: allow amortized growth of a pooled array but not hash-map growth, and add a test hook forcing inline capacity 1, patterned on the interner's `hashMask` (sets.go). Measured maxima, 12 eval and 32 mark contexts, mean the overflow never runs on the corpora. | open | 3e4d0c43b (a8960d05) |
| C19 | assess-codebase | 3 | §3.4, §4 | Every timed pass after the first times an already-constrained or already-pruned forest, and both §4 stops read the precedence key. | `Constrain` unlinks in place, and resolve's `prune` prunes in place. The README's stage protocol prepares inputs "once outside the span" (l.638-641). `Forest` has no copy method. §2 measured an already-constrained forest at about 0.88 of the cost. | open | 3e4d0c43b (a8960d05) |
| C20 | assess-codebase | 3 | §3.2 | A literal reading forbids `NewPrecedence`'s construction-time maps. Low. | Construction uses maps (`sym`, `level`, `in`, `id`, `kernels`). Only `lrState.next` is read per derivation (`walk`): a dense goto table of 687 × 237 entries. `tab.EOF` = 237, so terminal-indexed tables need 238 columns. | open | 3e4d0c43b (a8960d05) |
| C21 | assess-codebase | 3 | §4 | "Its first property holding" can be read as not requiring §3.2, the decision's own subject. Low. | §4: "§3 built, its first property holding". | open | 3e4d0c43b (a8960d05) |
| C22 | assess-codebase | 3 | §3.2 | Reusable buffers hung on `Precedence` would race under concurrent `Parse`. Low, issue-level. | `Precedence` is shared via `grammar.Derived(tab, buildParsePrecedence)` (readings.go). | open | 3e4d0c43b (a8960d05) |
| A18 | assess-architecture | 3 | §2 | Refinement or review misreads the stage figures as a regression, or anchors the ≤ 5 half of the exit to the wrong number. | §2 omits the minute's caveat ("Compare these stages with each other, not with the ratchet's keys") and the whole-pipeline 5.49 (5.48–5.64). Baseline `speed-stage-recognize` 0.7903 and `-forest` 0.6620, against §2's 0.94 and 0.73. | open | 3e4d0c43b (a8960d05) |
| A19 | assess-architecture | 3 | §5 | A later reader treats "about 5%" as a permanent ceiling and never revisits precedence in recognition after the map work changes the cost shape. | (1 − 0.877) × 2.38 / 5.49 ≈ 5.3% is `Constrain`'s share alone. It omits build, resolve and readings on a forest up to 16.17% smaller, and was measured on the map-based implementation. | open | 3e4d0c43b (a8960d05) |
| A20 | assess-architecture | 3 | header | A search by file name finds the wrong answer: the file records a decision not to put precedence in the recognizer. Low. | The plan and the minute keep the name `chart-precedence-in-recognizer`. Line 6 explains this. Assessments are named name-for-name (`chart-waypoint.md`). Proposed: `git mv` at merge, or perigrin rules the name stays. | open; question put to perigrin (see Pending) | 3e4d0c43b (a8960d05) |
| A21 | assess-architecture | 3 | §3.4 | Every close gate that runs the ratchet gets slower, against a budget already overrun. | Arithmetic, not a run: about 33 × 0.546 × (1 + 2.38) ≈ 61 s for the precedence key, plus about 21–22 s each for resolve and readings, so roughly +100 s. README "Budget": 300 s budgeted, 543 s on pu, 664 and 683 s in KEYS's runs (input 01a1ec3c). Verify's 1,800 s per-criterion timeout. | open | 3e4d0c43b (a8960d05) |
| A22 | assess-architecture | 3 | §3.4, §4 | Refinement writes an exit criterion that cannot run, or that compares one ratchet run with `baseline.txt`, which the README forbids. The live README goes stale. | README key table, rules, the `speed-stage sum` stderr line, and "There is no per-stage collector key". Recovery plan §6: the verdict reading is "recorded in the milestone's Records", not a resolution command. | open | 3e4d0c43b (a8960d05) |
| A23 | assess-architecture | 3 | §4, §5 | The next decision starts from no recorded intent. | Revision 2's §6 "what the result decides" sentence is dropped in revision 3. Candidates after the map work: recognize 0.94, build 0.73, the rest of `Constrain`, or nothing. | open | 3e4d0c43b (a8960d05) |
| A24 | assess-architecture | 3 | §3.2 | The literal reading may exclude a bounded fallback that meets §3.3. Raised for perigrin. | Recovery plan R2: "find duplicates by scanning, switch to a map at 24 items". §3.2 forbids any map, and §3.3 asks that a slow node stay correct. | open; question put to perigrin (see Pending) | 3e4d0c43b (a8960d05) |
| A25 | assess-architecture | 3 | §4 | An issue whose precedence stage key does not move merges unreported. Optional. | §4 reads before and after "at the end". Recovery plan M: "reported before it merges". The w2 postmortem: "stopped the waypoint at FOREST's gate rather than after its merge". | open | 3e4d0c43b (a8960d05) |
| P16 | assess-pushback | 3 | §3.2 | Whoever implements first decides whether §3.2 is met: a hand-rolled hash table and a dense indexed table would both be defensible readings. Low. | §3.2: "No hash map is read or written". It does not say whether that means Go's `map` or any hashed lookup. | open; question put to perigrin (see Pending) | 3e4d0c43b (a8960d05) |
| G1 | assess-fresh3 | 3 | §1, §4 | The exit cannot tell perigrin whether the work paid for itself. | `baseline.txt` (retaken at 83fdd50f4) reads 4.7915 with `Constrain` already in `ParseWith`. The measurement reads 5.49 on another host and is "not comparable". The README records the key moving across a host change. | open | 3e4d0c43b (a8960d05) |
| G2 | assess-fresh3 | 3 | §4 | If §3 is built with identical output and a faster stage but the ratio stays above 5, nothing stops the work before twice the budget. | §4's exit and stops have no case for "built, identical, stage key improved, exit not met". | open | 3e4d0c43b (a8960d05) |
| G3 | assess-fresh3 | 3 | §2 | A reader expects a size hint to recover much of the cost. In fresh3's experiment it does nothing measurable. | Experiment on 859 forests: maps with a size hint 111/91/104 ms against 111/97/107 ms; dense goto 0–7%; the no-map variant 40/36/35 ms. Recursion is 810 calls deep against a 64-frame CPU profile limit (`maxCPUProfStack`), so profile shares are approximate. | open | 3e4d0c43b (a8960d05) |
| G4 | assess-fresh3 | 3 | §3.1 | A change in which alternatives survive on an aggregate file outside the sample passes every named instrument. | The goldens cover about 440 sample files. The count keys and referee keys count verdicts and disagreements. Proposed (issue-level): a differential check against today's `Constrain` over the 2,211 files, comparing node heads and `Alt.Next`. | open | 3e4d0c43b (a8960d05) |
| G5 | assess-fresh3 | 3 | §3.2 | "Allocated once per call, not grown entry by entry" forbids the design that measured fastest, and pushes toward an arena where appending to a returned sub-slice overwrites its neighbour. | The prototype grows per-node slices by append for the 14% of nodes with more than one context. `eval` returns memoized slices directly, and `seq` returns `right` unchanged. Proposed: "no hash map; no table indexed by node, context or state grows during the call". | open | 3e4d0c43b (a8960d05) |
| G6 | assess-fresh3 | 3 | §3 | The issue author has to rediscover the memo semantics. Not a defect in the decision. | mark records before recursing; mark's key includes `first` (-1 means any); eval memoizes after computing, with no in-progress sentinel (safe because the forest is acyclic); memoized slices are shared; `kept []bool` is by alternative. The prototype kept all of these. | open | 3e4d0c43b (a8960d05) |
| G7 | assess-fresh3 | 3 | §3.4 | The issues have to land in a fixed order. Issue-level. | The three stage keys do not exist on pu: `measureStages` has lex, recognize and forest (`Build` alone). They need their own commit before any change to `Constrain`, and the interleaved before must be taken on it. This touches `main.go`, `main_test.go`, the README key table and the baseline. | open | 3e4d0c43b (a8960d05) |
| G8 | assess-fresh3 | 3 | §4 (direction) | The exit's thresholds will not separate outcomes, and perigrin lacks the table of where the time goes next. | If the guacamole speedup holds, `Constrain` falls from 2.38 to about 0.8–0.9 and the pipeline from about 5.5 to about 4.0. Proposed: the exit report gives the seven-stage table, before and after, in one session. | open | 3e4d0c43b (a8960d05) |

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

Revision 2's header (lines 11-12) cites "the chart-recovery spike C1 (`chart-recovery` Records, 2026-11-13)", and its §1 quotes D7. The drafter records that as a fact about the text, not as a disposition of anything.

## Measurement after round 2 (recorded, not a finding)

Taken under perigrin's "Measure first, then decide", by the measure-constrain agent on `pu` 8c3317079. This section is recorded by the coordinator from the agent's report.

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

**assess-fresh3's prototype (recorded from its report).** In a scratch worktree at pu 8c3317079, since removed, it built a variant of `Constrain` with no map. It kept the diff at `assess-w3/fresh3/prototype.diff`. The variant uses:

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
