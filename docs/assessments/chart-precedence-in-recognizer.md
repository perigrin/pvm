<!-- ABOUTME: Minute of the crochet:assess session on the design for applying precedence inside the chart recognizer. -->
<!-- ABOUTME: Covers round 1: the finding ledger, cross-references, recommendations and declarations, the outcome, and the author's statement. -->

## Assessment: Chart parser: precedence in the recognizer

Recommendation: pending (round 1: three modify)

- **Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.
- **Revision assessed in round 1:** design commit 7b264fa9c, blob 52c82af16a738eca6e03cdb050e9c167903212c7. Each participant read that blob, untracked in worktree `cr-close`, and checked it with `git hash-object`. It is byte-identical to the blob at 7b264fa9c.
- **Decision-maker:** perigrin.
- **Drafter:** assess-pushback, from round 2 on. The drafter records and holds no view. In round 1 it was the pushback participant and raised P1 to P12. From here on it does not count as an independent participant. assess-codebase and assess-architecture remain independent.

Section references are to the revision above. IDs keep each participant's own prefix:

- A: assess-architecture, the architecture and direction lens;
- C: assess-codebase, the codebase lens;
- P: assess-pushback, the spec-quality lens.

The full reports are at the paths below, in the session scratchpad (`/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/`):

- `assess-w3/round1-architecture.md`
- `assess-w3/round1-codebase.md`
- `assess-w3/round1-pushback.md`

## Ledger

The revision for every row is `7b264fa9c (52c82af1)`.

| id | raiser | round | section | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|
| A1 | assess-architecture | 1 | §1, §4.1, §4.2 | Refinement and review judge the issues against v2. They find §4.1 reversing v2 §8 with no ruling behind it, or they let an accepted ruling be reversed silently. | v2 §2 ("conflicts are kept as ambiguity and resolved in the forest"); v2 §8 (perigrin 2026-10-30, "precedence leaves recognition"); v2 §3 rule 12; v2 §10 kill "Facts in items". The §1 ruling is not in chart-perly-w2 Records or -inputs. Options B and C are undefined. | open | 7b264fa9c (52c82af1) |
| A2 | assess-architecture | 1 | header, §7 | A reader following "Follows" reads superseded rules, and A1's check cannot be made from the document. | "Follows" names the 2026-10-29 first attempt, which v2 supersedes. §7's "design §13" exists only in v2. v2 is on branch `docs/chart-parser-v2-design`, not on pu. | open | 7b264fa9c (52c82af1) |
| A3 | assess-architecture | 1 | §2, §4.2 | `items-per-token` and `set-reuse` may barely move for left-associative chains (`.`, `+`, `&&`, `||`), and the exit fails for a reason the design did not anticipate. | Reasoned, not measured. Prediction is unchanged, and the inner sub-chains of `$a . $b . $c . $d` are legal and reduced before `.`, so their items survive. The 2016 approach passes the constraint down at prediction (from memory, not checked). | open | 7b264fa9c (52c82af1) |
| A4 | assess-architecture | 1 | §6 | A key-set change with no ruling, and the precedence stage key's end state is undefined once `Constrain` is removed. Until then the timed pipeline pays for precedence three times. | README output contract: "no key is added, renamed or retired after it". Each earlier exception names a ruling. "A number may not become `na`". `forest.ParseWith` calls `parseStream`, then `f.Constrain`. | open | 7b264fa9c (52c82af1) |
| A5 | assess-architecture | 1 | §2, §6 | A number with no derivation becomes the waypoint's verdict, read by an unstated protocol, and it re-gates speed against D1 without saying so. | The stage keys sum to 2.107, leaving 0.39 at 2.5. Stop 3 allows recognize up to 1.0. §2's 2.7 is an inference from medians (README speed protocol). chart-recovery grounded its exit in a projection. D1 extended 2026-11-15 ("dropping the speed gate was universal"). | open | 7b264fa9c (52c82af1) |
| A6 | assess-architecture | 1 | §6 (stops) | The stops detect noise or nothing, and only after the work has merged. | README rules: "Timing keys … are never compared with the baseline". v2 §9.3: never compared strictly. Stop 4 is set against a single run (4.7915); checkpoint 3 read 5.18 and 5.11. The speed stops fire only "once both tables are in". | open | 7b264fa9c (52c82af1) |
| A7 | assess-architecture | 1 | §6 (held keys) | A gain is claimed from a smaller population, and gated keys are left out of the stops. | `items-per-token` and `set-reuse` are measured "over files recognition accepts", and that population changes when recognition refuses precedence errors. Precedent: the `speed-ratio-aggregate-cold` redefinition. The list omits `collapse-over-limit`, `tables-drift`, `ledger-gaps` and `lex-unmapped`. `oracle-exact-chart` is made a floor. | open | 7b264fa9c (52c82af1) |
| A8 | assess-architecture | 1 | §5 | The suite goes red, or the recognizer loses its outside oracle. The `Expected` contract (v2 §13) changes undecided. | `recognize/reference_test.go`, a textbook Earley reference with no precedence, used by `TestReferenceRecognizer`, `TestReferenceAgreesOnTruncatedFiles`, `TestWarmExpectedMatchesReference` and the fuzz oracle. v2 §3 rule 11. | open | 7b264fa9c (52c82af1) |
| A9 | assess-architecture | 1 | §4.3, §5.5 | The equivalence proof is partial until the last issue, and that issue changes the prover. | `referee.judge` (`internal/chart/precedence.go`) builds through `builtForest`, then `f.Constrain`, so an under-filtering table is invisible while `Constrain` stays. "Reporting and not removing" is not in the key's definition. "Every corpus" names no key. | open | 7b264fa9c (52c82af1) |
| A10 | assess-architecture | 1 | §4, §8 | Issues inherit a mechanism with no invariant to test it against. | §4 names functions, `uint32` and a (state, LHS) key. §3 finds every pair but one global. The README records an earlier chart-side keying that over-refused. §5.3's choice is missing from §8. | open | 7b264fa9c (52c82af1) |
| A11 | assess-architecture | 1 | §7 | The speed exit is reached, the next decision (M2, interleaving) becomes less reachable, and nothing on record weighs the trade. | v2 §2: what a positive verdict enables. First attempt's rule 2 (set interning). v2 §13 convergence "on state equality". v2 §11 isolation is not restated. | open | 7b264fa9c (52c82af1) |
| C1 | assess-codebase | 1 | §4.2 | `++ ++$x`, which bison accepts, is refused, and no referee notices. | (rules 256, 257) × POSTINC/POSTDEC/POSTJOIN resolve as error (one `%nonassoc` level). Toy `testdata/precedence/toys/prefix.txt` records `accept … PREINC PREINC NUM`. README: "`++ ++ $x` … is kept". No such sentence in `testdata/precedence/sentences.txt`. `shapes.pl`: every reduce or error shift item is `X: Z . t …`. | open | 7b264fa9c (52c82af1) |
| C2 | assess-codebase | 1 | §4.1 | As literally written, the filter never fires on `$a = $a = …` or `$a + $b * $c`. One of the two fixes turns it into an LALR(1) lookahead filter that §3 did not count. | `TestScratchAssessSets`: `=` is {ASSIGNOP, PERLY_EQUAL_SIGN}, `*`, `+`, `-` and `++` similar. 2,595 of 7,455 operator tokens have more than one terminal. The second terminal is in no resolution pair. `TestScratchAssessFollow`; `la.pl`. | open | 7b264fa9c (52c82af1) |
| C3 | assess-codebase | 1 | §4.1 (and §5) | Acceptance differs between lookahead on and off. The fuzz target and `TestLookaheadKeepsAcceptance` fail or are weakened. At lookahead 0 (interleaved mode) half of precedence vanishes. | `recognize.Options` doc: "Neither changes acceptance". Fuzz `oracleDisagreement` compares `optionCombos` against `refRun`. `NoLookahead` sets `follows = nil` (recognize.go l.276-280). `Expected` is built "as recognition without lookahead". | open | 7b264fa9c (52c82af1) |
| C4 | assess-codebase | 1 | §4.2, §5.1 | Stored values split sets for no reason. Without re-propagation, §4.2's "cannot lose a derivation" is false, and the failure shows as false refusals. | The value is read only where a completed child at `pos` completes into its parent at `pos`, so it can stay builder-local. `builder.complete` visits each item once, and `isNew` drops a duplicate (state, origin). | open | 7b264fa9c (52c82af1) |
| C5 | assess-codebase | 1 | §5.2, §8.2 | The shift side loses the value on most of its traffic, because unit links go through Leo memos. | `TestScratchAssessLeo`: operator chains form no multi-link memos (only `else` 14 and `optlistexpr` 239). Single-link memos include termbinop 1,652, termeqop 187, termrelop 97 and termunop 110. `complete` adds the memo's `top` and `continue`s before `awaiting`. | open | 7b264fa9c (52c82af1) |
| C6 | assess-codebase | 1 | §4.3 | A blocked reduce-side derivation is packed into the forest anyway, and "apply the tables when packing" cannot be done for each parent separately. | `forest.build` makes a fact for every final item, including completions the recognizer skipped. Example: `$x = $y + $z + $w`. Child nodes (sym, from, to) are shared by parents. 35 operator-completing states hold one reduce and no non-final kernel item (`mix.pl`). | open | 7b264fa9c (52c82af1) |
| C7 | assess-codebase | 1 | §4.1, §5.4 | The reduce table's keying is undecided. Keying by chart state needs a chart-to-bison state map that does not exist. | Chart state 373 matches bison 193 (the conflict) and 232 matches 153 (allowed). `NewPrecedence` uses bison's LR(0) states, "never the chart's split states". `resolve(r, t)` covers 669 of 670 pairs. | open | 7b264fa9c (52c82af1) |
| C8 | assess-codebase | 1 | §4.2, §8.1 | The value does not fit its type, and the identity element contradicts itself. | 32 levels plus a reserved bit is 33 bits. "Empty set's identity, a reserved bit" against "an empty result is zero". PLUGIN_* is already among the 32 (perly.y l.150-182). About 20 levels are possible on top (`xr.pl`). | open | 7b264fa9c (52c82af1) |
| C9 | assess-codebase | 1 | §4.2 | `$a + $b < $c` and `$a and $b and $c` group wrongly if `%prec` unit rules set PREC_LOW. | `termbinop: termrelop %prec PREC_LOW`, `termbinop: termeqop %prec PREC_LOW`, `termrelop: relopchain %prec PREC_LOW`, `listexpr: term %prec PREC_LOW` and `expr: listexpr %prec PREC_LOW`. bison uses their precedence only in reduce-side pairs (rule 236 against CHRELOP, state 68). | open | 7b264fa9c (52c82af1) |
| C10 | assess-codebase | 1 | §5.5, §4.3 | There is no equivalence proof, and the residue that gates `Constrain`'s removal has no key. | `precedence-covered` 1,304 of 1,347. 43 resolutions are uncovered (reviewed), and C1's case is absent. `precedence-disagree` reads only the generated fixture. The 13 toy referees run against `Constrain` only. | open | 7b264fa9c (52c82af1) |
| C11 | assess-codebase | 1 | §4.2, §4.1 | Raised as support: checking the top operator alone is sound only alongside the reduce side. Any reduce-side residue (C2, C3) also opens a shift-side hole. | `pred.pl`: in 43 of 46 (R, state-before-operand) pairs, R alone predicts the operand. The other 3 cannot begin a left-operand rule. The reduce side forces inner spine levels. | open | 7b264fa9c (52c82af1) |
| P1 | assess-pushback | 1 | §4.2 | The implementer chooses between one value per item, which is wrong where a state holds several dotted rules, and one value per dotted rule, which is a different design with a different cost to set identity. | 51 of 772 states complete more than one rule (`states.txt`). The recovery plan (branch `docs/chart-recovery-plan`, §C, C1): "A single value per item could not kill one rule inside a state". | open | 7b264fa9c (52c82af1) |
| P2 | assess-pushback | 1 | §4.2 | A derivation the table admits is lost after its item has already completed, and shows as false refusals (`rd-accepted-chart-refused` above 0). | `builder.add` (recognize.go:376) deduplicates with `isNew` and never revisits. Recovery plan C1: "An item whose value grows must complete again". | open | 7b264fa9c (52c82af1) |
| P3 | assess-pushback | 1 | §4.2 | Under one reading, `$a << ($b < $c)` survives recognition. | `%prec` unit rules (perly.go:354, 360, 440, 442, 464) are both "unit" and "operator" rules. SHIFTOP/CHRELOP resolves as reduce (26 CHRELOP reduce pairs, bison run). §3 counts (rule, token) pairs, not level expressibility. | open | 7b264fa9c (52c82af1) |
| P4 | assess-pushback | 1 | §4.2, §5.4, §5.5 | A resolution the generator cannot map is dropped or mis-mapped silently. 43 resolutions' entries are unrefereed. | bison 373 rules against the chart's 350. One reduce resolution names POSTJOIN, which perly.go:796 omits. Referee coverage 1,304 of 1,347. No stated refuse, review-list or drift-check path. | open | 7b264fa9c (52c82af1) |
| P5 | assess-pushback | 1 | §6, §8.4 | The waypoint cannot finish, or the exit is lowered partway through. | 4.79 = 0.655 + 0.790 + 0.662 + about 2.68. Lex is untouched, and resolve and readings are out of scope (§7). Reaching 2.5 needs the post-forest share at or below about 1.1, mostly from removing `Constrain`, which no key times. The 115 ms probe is std5k only. | open | 7b264fa9c (52c82af1) |
| P6 | assess-pushback | 1 | §4.3, §6 (stop 4) | Stop 4 is a false alarm or is never checked. | `Constrain` runs until the last issue, so no fall below 4.79 is expected "once both tables are in". No reading protocol, against quiet repeats of 5.18 and 5.11, 8.56 and 8.39. Stop 3's causal clause cannot be measured. | open | 7b264fa9c (52c82af1) |
| P7 | assess-pushback | 1 | §4.3, §8.3 | The last issue has no runnable criterion, or it removes `Constrain` on evidence from generated sentences alone. | `precedence-disagree` is defined over the generated sentences (README key table). `Constrain` has no report-only mode (precedence.go:322 returns a bool and unlinks). "The waypoint's last issue" names no milestone. | open | 7b264fa9c (52c82af1) |
| P8 | assess-pushback | 1 | §1, §5.1 | The merged code contradicts the live README and the binding design, and no superseding record says so. | README: "No fact is added to an Earley item (spec §3 rule 12, §10)". The recovery plan says a successor design must amend rule 12 and §10, with C1's numbers as evidence. The design does not cite C1. "Option B/C" appear in no committed document. "Follows" cites v1. | open | 7b264fa9c (52c82af1) |
| P9 | assess-pushback | 1 | §5.2, §8.2 | With Leo on, the reduce-side filter is skipped for intermediate links, and `leo-disagree` (acceptance only) does not notice. | `complete` calls `leo()` before `awaiting`. The lookahead case relies on "FOLLOW only shrinks up a chain" (leo.go), and no precedence analogue is stated. Recovery plan C1: the memo stores no value, plus a guard on values. | open | 7b264fa9c (52c82af1) |
| P10 | assess-pushback | 1 | §4.2, §8.1 | The "no operator" bit collides with a level, or the type changes partway through implementation. | perly.go:744-806 declares 32 levels, PLUGIN_* among them. Adding the reserved bit makes 33. | open | 7b264fa9c (52c82af1) |
| P11 | assess-pushback | 1 | §5.3 | Two designs are left open. If neither is built, a derivation through a blocked terminal reaches the forest. | §5.3: "either a check at the next scan or a residue for the forest's tables". | open | 7b264fa9c (52c82af1) |
| P12 | assess-pushback | 1 | §3 | §3's figures cannot be reproduced from pu. | `internal/chart/offline/precedence/` is untracked and has no expected-output file. The figures reproduce with `count.sh /opt/corpora/perl5`. | open | 7b264fa9c (52c82af1) |

### Cross-references

The drafter lists overlaps and does not merge them. Each row stays open in its own right.

| cluster | rows | shared subject |
|---|---|---|
| value per item; re-propagation | P1, P2, C4 | what the value attaches to, and widening after completion |
| which rules set the value | P3, C9, C1 | `%prec` unit rules; prefix rules |
| governing decision and amendment | P8, A1, A2 | rule 12, the §10 kill, v2 §8; v1 cited as "Follows"; Options B and C undefined; chart-recovery C1 not cited |
| value width | P10, C8 | 33 bits in a `uint32`; PLUGIN_* levels already counted |
| exit | P5, A5 | 2.5 underived; no reading protocol |
| stops | P6, A6 | timing compared to a single run; stops fire late or on predicted behaviour |
| removal gate and referee | P7, C10, A9 | `precedence-disagree` covers generated sentences only; no report-only mode; no residue key |
| Leo | P9, C5 | the Leo interaction in §5.2 |
| reference recognizer and Options | A8, C3 | the reference recognizer, `Expected`, the lookahead-independence contract |
| table generation and keying | P4, C7, A10 | rule mapping, keying by state or by rule, error path |
| the §5.3 residue and which terminals the filter tests | P11, C2, A10, C11 | ambiguous lexemes |

### Distribution by section

Counted by the first section each row names. A row naming several sections counts once.

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

All 34 rows remain open. No participant has released a finding, and no finding has been rebutted.

### Outcome

**Another round is owed.** Three participants recommend modify, and every finding is open. Round 2 assesses the revision that answers them, and goes to the participants as the diff plus this ledger.

## Author's statement (recorded, not a finding)

After round 1, the author reported the following. It is the author's statement, recorded as given. The drafter has not checked it, and it is not a ledger row.

> chart-recovery's C1 spike already measured precedence values on items. Its report is at scratchpad `chart-recovery/c1-report.md`. Records 2026-11-13, and perigrin's 2026-11-14 D7 ruling: "I think the successor design is something we can revisit after §8 as written." The design does not cite it.
