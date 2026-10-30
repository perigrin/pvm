<!-- ABOUTME: Minute of the crochet:assess session on the chart-parser second-attempt design (v2). -->
<!-- ABOUTME: Covers round 1: the ledger, clusters by section, each participant's declared position, and what was not checked. -->

# Assessment: chart-parser second-attempt design

**Subject:** `docs/plans/2026-10-30-chart-parser-v2-design.md`
**Revision assessed:** round 1 at 083e1357 (083e135759ca88c95bcee786c1da24d81df41a06). Branch `docs/chart-parser-v2-design`. All five participants name this revision as the one they read.
**Author:** the orchestrating agent. It holds a view on the subject and does not draft.
**Drafter:** a recording role that holds no view. It raises no finding.
**Round 1 dispatched:** 2026-10-30.

**Participants and lenses**

| participant | lens |
|---|---|
| A | codebase and architecture: the subject's claims against `origin/pu` (4fffe4d6) |
| B | grammar fit: perly.y over `internal/lexer` |
| C | measurement and kill design: §2, §7.4, §9, §10 |
| D | spec quality: paad:pushback taxonomy, plus crochet:assess's rule that implementation detail belongs in issues |
| E | architecture and direction |

Section references are to the subject at 083e1357. Rules (§3 rule N) are under §3.

## Outcome of round 1

### Recommendations, as declared

- **A: modify.** "I am not questioning any of perigrin's listed choices." A5 refines the (kind, text) table "without rejecting it"; "if perigrin meant (kind, text) literally, then I am questioning that choice."
- **B: modify.** "The machinery choice is unaffected by all of this. The grammar-over-lexer interface needs one more page of design (what is re-fitted, what the key holds, what precedence must preserve) before refinement can size it." Questions *how* the precedence rewrite is specified (B3) and the table's key (B4), not the choice of a rewrite or of a compiled table.
- **C: modify.** "The shape is right." "Several gates as written measure something other than what they are meant to … Each fix is a sentence or a deterministic check, not a redesign." C3 questions the definition of long-lived, "not perigrin's choice to kill on it".
- **D: modify.** States what would change its mind on D1, D5 and D6 (see Per participant). D2 questions perigrin's "precedence rewrite" choice "only its scope, not the choice".
- **E: modify.** "The blocks are about what the spike can *conclude*, not about how it is built." E1 questions a perigrin choice ("a kill there killing only up-front mode"); E2 questions "part of a perigrin choice".

### Blocks, as declared

| participant | blocks |
|---|---|
| A | A1, A2, A5, A8 |
| B | B3, B4, B5 |
| C | C1, C3 (on the definition, not on the choice of long-lived), C4, C5 (on gating a wall-clock ratio; reporting it is fine), C6, C7 (on defining "checkpoint"), C8, C9 |
| D | D1, D2, D3, D4 ("resolvable with one reported or gated key"), D5, D10 (on the gate defined via Q4). Conditional: D6 ("block otherwise", i.e. if D5's resolution does not move the start forms out); D8 ("block on the budget only if perigrin sizes from them") |
| E | E1, E2, E3 |

### Stand-asides, as declared, with grounds

Where a participant gave no grounds beyond the finding itself, the table says so.

| id | grounds given by the participant |
|---|---|
| A3, A4, A6, A7, A9–A16 | None stated beyond each row's "what would resolve it". |
| B1, B2, B6, B7, B8, B9 | "They are corrections or clarifications, and B6–B8 may move to issues." |
| C2 | "Name the entry point in the decision; `:tighter`/`:equiv` scope goes to B and the issue." |
| C7 (interleaving) | Issue-level. |
| C10 | None stated. |
| C11 | Issue-level. |
| D6 | "Stand aside if D5's resolution moves them out." |
| D7 | "One sentence fixes it." |
| D8 (the figures) | None stated beyond the row. |
| D9 | None stated. |
| D10 (all but the Q4 gate) | "Move it, do not re-specify." |
| D11, D12, D13 | None stated. |
| E4 | "It is future work, but it is the design's only stated link to M2." |
| E5 | None stated; the row says its resolution "changes the order of the work, not its scope". |
| E6 | None stated. |
| E7 | None stated; the row says the open question "belongs in waypoint 2's refinement". |
| E8 | "I am not questioning the choice of a precedence rewrite, only the claim that its stated schemes are complete." |

### Measurements that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **Precedence lines in perly.y:150–182.** A4, B1 and D8 measure 32 directives (line 152 blank). E6 reports "33 precedence lines (`150-182`)" among the counts it checked as correct.
- **Nonterminals and alternatives.** B1 measures 348 alternatives over 114 nonterminals (bison 3.8.2). D8 measures 114 and "about 348" by heuristic. E6 did not check this count. A did not check it.
- **`%prec` uses that are not operator levels.** B3: "about 20 of the 39". D2: "33 of the 39" (21 `PERLY_PAREN_OPEN`, 12 `PREC_LOW`).
- **perly.y terminals.** B: 131 (sampled 31). D1: about 125. E6: 124 distinct token names.
- **`lex-overhead` on a thin chart/lex.** A9: 1.15–1.27 on std5k (`-benchtime 30x -count 3`). C5: 0.91–1.39 on std5k and 0.90–1.44 on the aggregate under §9.3's protocol, with an A/A spread of 0.65–1.35.
- **Long-lived against cold on std5k.** C3 (v1 tree): about equal. E2 (pu): 36.98 ms against 36.47 ms, ratio 1.01. Both report the aggregate's replay figure about half of first sight.

### Round outcome

**Another round owed.** Every participant declared at least one block; nothing has been released. 57 findings are open.

## Ledger

Status uses the participant's round-1 word. "See also" links rows whose raisers describe the same defect; rows are not merged or ranked.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A1 | A | 1 | §5.3 Decouple; §5.4 | §5.3 sizes "decouple" as two functions; under that scope §5.4's "single visible deletion" does not compile, and the preparation issue's criterion passes while the coupling remains. | `BuildLR0` takes `Rule`/`Symbol`/`Seq`… defined in `slif.go`; `BuildTables`/`GenerateTables` call `Probes`/`FindConflicts`; `lex.New(*grammar.CutGrammar)`/`lex.Guacamole` is a third loader; `grep -cE 'ReadSLIF\|ApplyCut\|Guacamole…'` over carried tests: recognize_test 18, expect_test 16, lr0_test 11, leo_test 4, lattice_test 2, fuzz_test 1. | block | released at d6cb49c7 | 083e1357 |
| A2 | A | 1 | §7.1; §9.2; §11 | Waypoint 1's ".tpm-declared infix operator in scope is accepted" cannot be met without changing `internal/parse`, which §11 forbids; waypoint 1 cannot exit. See also C2, D5. | `go doc -short ./internal/parse` exports no operator symbol/class accessor; probe `parse.DeclaredSignatures("sub ⊕ :infix(ADD) …")` returns no keys; `grep -rln ':infix(' internal/parse/declarations \| grep -v CORE.tpm` finds nothing. | block | released (ruling) at d6cb49c7 | 083e1357 |
| A3 | A | 1 | §7.1 | "Adds an entry … to the `PLUGIN_*_OP` terminal of its class" is undefined for some operators pvm accepts; a worker invents a mapping that can disagree with RD. See also B8, D5. | pvm has 15 classes (`classAnchors`); perl has 11 `INFIX_PREC_*` (`perl.h:6117-6127`) mapped by `tokentype_for_plugop`; `:tighter`/`:looser`/`:equiv` give levels with no PLUGIN token. | stand aside | released (ruling) at d6cb49c7 | 083e1357 |
| A4 | A | 1 | §4; §5.2 | Two counts or citations taken at a glance; one repeats an error the code warns about. See also B1, D8 (count); B2 (citation); E6 (reports 33 as checked). | `grep -cE '^%(left\|right\|nonassoc)' perly.y` → 32; `go doc ./internal/parse LevelsAccountedFor` warns of an earlier "33"; library-operator emit is at `scan.go:1269–1272`, not 1227. | stand aside | released at d6cb49c7 | 083e1357 |
| A5 | A | 1 | §7.1; §7.3; §3 rule 7 | The (kind, text) key throws away decisions the lexer already makes; §7.3 calls `{` block-vs-hash a fact pvm lacks, but pvm has it; the chart may take the reading the lexer rejected. See also B4, D1. | `lexer.Token` carries `OpensBlock`, `TermFollows`; probe: `map {` ob=true, `my $h = {` ob=false, `sub f {` ob=true, `print $fh -1` tf=true on `$fh`. | block | released at d6cb49c7 | 083e1357 |
| A6 | A | 1 | §3 rule 12; §7.3 | Rule 12 sorts every collapsing fact into grammar or scope; some toke.c decisions depend on the next character, and waypoint 2's resolve has no scope fact to settle them. See also B4 (ii). | `perl -MO=Deparse,-p -e 'print ($x), 1;'` → `(print($x), '???')`; `print` → {LSTOP, FUNC} yields two readings. | stand aside | released at d6cb49c7 | 083e1357 |
| A7 | A | 1 | §5.1; §9.5 | "cmd/ratchet carried over" overstates what carries; §9.5's keys and gate rules cannot run on the carried comparator without changing the ratchet's own rule. | `Compare`: "a key is never removed or renamed"; §9.5 renames and drops keys; `Compare` has only `mayNotRise`/`mayNotFall`, no absolute threshold; `main.go` (448 lines) imports `chart/lex`, Guacamole paths, `parseoracle`. | stand aside | released at d6cb49c7 | 083e1357 |
| A8 | A | 1 | §3 rule 10; §9.3 | "Speed is gated from the first recognizer issue" has no mechanism that fails; repeats the first attempt's failure. See also C6, D3. | `make test` is `go test ./... -count=1`, no `-bench`; a panicking benchmark leaves `go test ./internal/chart/probe_a/ -count=1` at `ok`; ratchet not in `make test`; `chart-ratchet.yml` `continue-on-error: true`; no `Benchmark` functions in internal/chart on pu. | block | released; consequence raised as A18 at d6cb49c7 | 083e1357 |
| A9 | A | 1 | §7.4 | The 1.3 gate is feasible but sits within run-to-run noise; a ratio of two separately timed medians can flip with no code change. See also C5. | `lexov_test.go`, `-benchtime 30x -count 3` on std5k: Tokenize 15.9/19.9/18.6 ms; Tokenize+trivia+lookup 20.2/24.0/21.4 ms; ratio ≈1.15–1.27; 4 cores, load 2.6. | stand aside | held, re-scoped; stand aside at d6cb49c7; released into A22 at 460adf6f | 083e1357 |
| A10 | A | 1 | §5.2 GRAM* | "`Recognize` takes the start symbol as a parameter" is implementation detail, and more work than needed. See also D10; B7, D6 (start forms). | `BuildLR0(start string, …)`; `Recognize` seeds `{State: 0}`; perly.y:187–257 states start forms as terminals in one rule, so one GRAM* edge needs no API change. | stand aside | released at d6cb49c7 | 083e1357 |
| A11 | A | 1 | §5.2 refusals | The refusal "checked against internal/lexer's own … operator tables" cannot read that table under §11. | `go doc -short ./internal/lexer` exports `Keywords`, `IsKeyword`, `IsFileTest`; `operators`, `bitwiseStringOps`, `positionalOps` (scan.go:1125–1143) unexported. | stand aside | released; new defect raised as A20 at d6cb49c7 | 083e1357 |
| A12 | A | 1 | §5.2 example; §5.3 Q9 | (i) The example DSL would not compile while slif.go exists. (ii) Q9 is a real defect. See also D10 (i). | (i) `type Seq int` in slif.go collides with `func Seq`. (ii) Probe: `Star` with RHS `a b` builds with err nil as `S ::= ε \| S a`; `Plus` with empty RHS panics `index out of range [0]`. | stand aside | released at d6cb49c7 | 083e1357 |
| A13 | A | 1 | §5.3 | §5.3 cites Q4 and Q7–Q11 to `## Records`, which define only Q1, Q2, Q3, Q5/Q6; the citation cannot be checked. See also D10 (sourcing). | `grep -n 'Q4\|Q7\|Q8\|Q9\|Q10\|Q11' A/body.md` → only the review line; postmortem has no Q-ids. | stand aside | released at d6cb49c7 | 083e1357 |
| A14 | A | 1 | §9.1 | "`parse.Parse` returns no syntax error" names an interface that does not exist; `rd-accepted-chart-refused 0` can be computed two ways. See also C1. | `func Parse(src []byte) *Node` has no error result; acceptance is read from `Node.Refused()`; refusal codes include `UnimplementedStatement`, `CatchAll`. | stand aside | held; superseded by A17 at d6cb49c7; released (with A17) at 460adf6f | 083e1357 |
| A15 | A | 1 | §8 | The cross-check against "internal/parse's binding powers" cannot read them. See also B3 (cross-check). | `bindingPowers`, `deriveBindingPowers` unexported; `parse.LevelsAccountedFor() map[int]string` exported. | stand aside | released at d6cb49c7 | 083e1357 |
| A16 | A | 1 | §7; §9.2 | The superset holds partly because quotes are opaque; interpolated code and `s///e` bodies are never recognized; not stated. See also B4 (i), B5, D1 (ii). | Probe: `"a $x b"`, `s/a/$b/g`, `qw(a b)` each lex as one `Quote`; 23 perly.y lines mention SUBLEXSTART, PMFUNC, QWLIST, FORMLBRACK, PERLY_SNAIL or HASHBRACK. | stand aside | released at d6cb49c7 | 083e1357 |
| B1 | B | 1 | §4 | The port is ~15% more alternatives and ~70% more nonterminals than stated, so the 15-day budget rests on the wrong size; "33 lines" contradicts §8's 32-level cross-check. See also A4, D8; E6 (reports 33 as checked). | `bison -Wnone --report=all` (3.8.2): 373 rules = 348 + 24 mid-rule + `$accept`; 114 nonterminals; `grep -n '^%\(left\|right\|nonassoc\)'` → 32. | stand aside | released at d6cb49c7 | 083e1357 |
| B2 | B | 1 | §5.2 | Stale citation. See also A4. | `git show origin/pu:internal/lexer/scan.go`: 1227 is `bitwiseStringOps`; symbol→Operator at 1269. | stand aside | released at d6cb49c7 | 083e1357 |
| B3 | B | 1 | §8 | The superset gate fails on std5k.pl and on `$a + defined $b`-type code, or the implementer invents an unspecified rewrite mid-milestone. See also D2, E8; A15 (cross-check). | Stratified rewrite (`strat.py`) refuses `$a + defined $b`, `$a * defined $b + 1`, `$a + not $b, 1`, `my $s = join ",", @a`; perl Deparse and RD (`rdprobe`, 0 refused) accept; 110 `= listop` lines in std5k.pl. | block | released (ruling); the filter's feasibility raised as B10 at d6cb49c7 | 083e1357 |
| B4 | B | 1 | §7.1–7.3; §3 rule 8 | (i) Whole-kind Variable and Quote make every variable and string ambiguous; (ii) FUNC/LSTOP, FUNC1/UNIOP, NOTOP/FUNC1, autoquote and `{` are lexical, so waypoint 2's resolve cannot choose and gives WRONG sites; (iii) OB, TF and signature mode are thrown away. See also A5, D1; A6 (ii); A16 (i). | Table of 31 sampled tokens (rows 1, 3, 5, 6, 10, 18, 20, 28); `lx/dump`; `lx/amb` on T1: per-statement p99 16, >2^10 in 19 of 44,307 statements (0.04%), a floor. | block | released; remainder raised as B11 at d6cb49c7 | 083e1357 |
| B5 | B | 1 | §5.2; §3 rule 7 | A cited rule-per-line port over internal/lexer refuses `{ 1 }`, `$h{a}`, `+{a=>1}` and end-of-file statements without `;`, and cannot express formats, labels, attributes, interpolation or postfix deref; the citation refusal pushes toward transcription. See also D1; D12 (HeredocBody trivia); A16. | `toke.c:6897` (synthetic `PERLY_SEMICOLON` before `}`), `toke.c:7670`; perly.y 1328, 1710, 854, 906, 1058, 1659; `lx/dump`. | block | released, against B's own bar; omission raised as B12 at d6cb49c7 | 083e1357 |
| B6 | B | 1 | §7.4; §9.5 | `lex-overhead` passes while superposition inflates recognizer work; the gate measures the cheap half. See also D4. | Argued from §7.2/§7.3: lookup cost is constant per token; extra terminals cost items. Not measured. | stand aside | released at d6cb49c7 | 083e1357 |
| B7 | B | 1 | §5.2; §9.2 | Non-program start forms lex in the wrong state; §11 forbids changing the lexer. See also D6; A10. | perly.y:189–259 GRAM* mid-rule actions set `parser->expect` to XTERM/XBLOCK; `checkpoint.go` exports only statement-start states. | stand aside | released at d6cb49c7 | 083e1357 |
| B8 | B | 1 | §5.2; §7.1; §9.2 | "Accepted, as recursive descent accepts it" holds only for infix library operators in perl's 11 classes; RD's 15 classes, relation placement and prefix/postfix fixity are not expressible; waypoint 2 parity differs. See also A3, C2, D5. | `internal/parse/operators.go` `classAnchors`, `operatorArity`; `toke.c:9264-9289`; perl.h:6123 vs perly.y:168 on PLUGIN_REL_OP. | stand aside | released (ruling) at d6cb49c7 | 083e1357 |
| B9 | B | 1 | §5.2 refusals | "A precedence token is used by no rule": if "used" means a right-hand side, the generator refuses perly.y's own UMINUS and PREC_LOW. | `grep -n UMINUS perly.y` → 175, 1497, 1499 (`%prec` only). | stand aside | released at d6cb49c7 | 083e1357 |
| C1 | C | 1 | §9.1; §9.2 | "No syntax error" is not a predicate that exists, and the gated population includes files perl rejects; the chart must accept non-Perl (against rule 5) or hold `rd-accepted-chart-refused` above 0; the gate measures RD's leniency and the lexer. See also A14; C10. | `parse.go:653` "no error return"; `zzpred` × `perl -c`: parses perl-ok/rd-accepts 743, ok/refuses 3, rej/accepts **18**, rej/refuses 5; parsent 75/3/**17**/29; `Refused()` agrees with Error/Missing on all 769 + 124 + 986 files. | block | held, re-scoped (the predicate); block at d6cb49c7; released, against C's own bar at 460adf6f | 083e1357 |
| C2 | C | 1 | §9.2; §9.1 | "As recursive descent accepts it" is false under §9.1's own definition; a `:infix(ADD)` fixture passes before the general case works. See also A2, D5; A3, B8. | `parse.Parse` has a nil resolver (module.go:323); `use My::Lib; 1 plus 2;` → `zzpred` refused=true although `My/Lib.tpm:32` declares `sub plus :infix(ADD)`; `TestLibraryOperatorMayUseAnyLevel`. | stand aside | released (ruling) at d6cb49c7 | 083e1357 |
| C3 | C | 1 | §9.3; §10 | Long-lived as defined replays a cache warmed on the same files; the kill reads the most favourable figure, and the redefinition alone moves the line more than any fix did. See also E2. | `v1/internal/zzlong`, recognition only, median of 9 after 2 warm-ups: aggregate (763 files) per-run 19.1/21.5/29.0 ms, long-lived **10.6/9.6/8.8 ms**; std5k 45.7/58.2/48.4 vs 47.3/45.3/42.4 ms. | block (on the definition, not on the choice of long-lived) | released, against C's own bar at d6cb49c7 | 083e1357 |
| C4 | C | 1 | §9.3; §10 | "The aggregate" is undefined (corpora, statistic, std5k/edit5k inclusion, both-accept restriction); execute would define it; a kill key dominated by one generated file judges the template. | `zzspeed` rd-ms (speed-aa.txt): std5k 24–33 ms vs all 769 snippets 6–9 ms; `code_too_large.t` 484 KB of T1's 2,176 KB (22%), 4,998 near-identical lines; T1 RD total 420–470 ms. | block | held, re-scoped (one file dominates the aggregate); block at d6cb49c7; released, against C's own bar at 460adf6f | 083e1357 |
| C5 | C | 1 | §7.4; §9.2; §9.5 | A wall-clock 1.3 gate cannot be decided under §9.3's protocol here; a correct chart/lex fails about a third of the time and a 30% regression passes about as often; ambiguous whether chart/lex's time includes Tokenize. See also A9; B6, D4. | A/A `Tokenize` vs itself: 0.65–1.32 std5k, 0.67–1.35 aggregate; thin chart/lex 0.91–1.39 / 0.90–1.44, over 1.3 in 4 of 12 trials each; interleaved A/A 0.76–1.19 (speed-aa.txt, speed-paired.txt). | block (on gating a wall-clock ratio; reporting it is fine) | core released, against C's own bar; remainder held (which size the A/A rule selects), stand aside; see C12 at d6cb49c7; remainder released, against C's own bar at b85dc97d | 083e1357 |
| C6 | C | 1 | §3 rule 10; §9.3 | "Benchmarks … run in `make test`'s package tests" is false; nothing in `make test` goes red on a slowdown. See also A8, D3. | A panicking benchmark passes `go test ./... -count=1` ("ok bchk") and panics only with `-bench .`; `sed -n 21,22p Makefile`; ratchet CI advisory (Makefile:55-57). | block | released, against C's own bar; consequence raised as C15 at d6cb49c7 | 083e1357 |
| C7 | C | 1 | §10; §9.3 | "Checkpoint" is undefined and the protocol lets load drift into the ratio; near 3× the verdict depends on the orchestrator's reading. See also D11. | A/A of RD alone 0.90–1.23 std5k, 0.79–1.15 aggregate (speed-aa.txt); 2f46 `speed.go` `speedRatios` times the two blocks one after the other; v1 results 2.83–3.85. | block on defining "checkpoint"; stand aside on interleaving (issue-level) | released, against C's own bar; consequence raised as C13 at d6cb49c7 | 083e1357 |
| C8 | C | 1 | §2; §9.2; §9.4 | The waypoint-1 exit cannot answer §2's question (it bounds one side), and waypoint 2 never re-asks it; a parser 5× slower once the forest exists could pass. | Read §9.2–§9.4; `parse.go` `parseWith` (parseFile then finish). | block | released at d6cb49c7 | 083e1357 |
| C9 | C | 1 | §9.2; §2 | Nothing measures completeness of `perl.go`, so the criterion passes before its work; "all of Perl" is asserted but unmeasured. | Every §9.2 key (tables-drift, refusals, superset, leo-disagree) is silent about a perly.y alternative no corpus file reaches. | block | held, re-scoped (the ledger's reference); block at d6cb49c7; released, against C's own bar at 460adf6f | 083e1357 |
| C10 | C | 1 | §9.1 | The corpus table's counts and labels mix verdicts; the denominator decides what the gates mean. See also C1, C4. | perl 5.42 accepts 78 of 124 parsent; RD refuses 3 perl-accepted parses and 3 parsent; `zzpred …/unit` files=986 refused=236 (24%). | stand aside | released at d6cb49c7 | 083e1357 |
| C11 | C | 1 | §9.3 | Until Q3 lands, the cold figure measures `recognize.New`'s setup; the harm is misreading a reported key. See also E2 (cold). | longlived.txt: fresh recognizer per file over 763 snippets 1,097–1,283 ms vs 9–11 ms long-lived (~1.5 ms/file setup). | stand aside | released at d6cb49c7 | 083e1357 |
| D1 | D | 1 | §3 rule 7; §5.2; §7.1; §7.3 | The port is a re-derivation presented as a transcription; rules restated in lexer vocabulary carry a `Cite` for a rule they do not transcribe; sigils cannot be keyed; OB/TF are discarded and §7.3 re-decides `{`. See also A5, B4, B5; A16. | `go test -overlay … -run TestDProbe ./internal/lexer` → `Variable«$x»`…`«$#x»`, `Quote«"a $x @{[ f() ]}"»`, `Word«foo» Operator«:»`, `Operator«{»^B` vs `Operator«{»`, `Variable«$fh»^T`; ~125 perly.y terminals vs 20 lexer kinds. | block | released, against D's bar; remainder re-scoped as D18 at d6cb49c7 | 083e1357 |
| D2 | D | 1 | §8; §5.2 | The rewrite is defined only for binary operators; most `%prec` uses resolve LALR conflicts with no stratified counterpart; the implementer invents semantics or the grammar stays ambiguous, against §5.2's "unambiguous expression grammar". See also B3, E8. | `sed -n 150,182p perly.y` → 32 directives; `grep '%prec' … \| sort \| uniq -c` → 21 `PERLY_PAREN_OPEN`, 12 `PREC_LOW`, 3 `UNIOP`, 2 `UMINUS`, 1 `LSTOP`. | block | released (ruling); consequence raised as D15 at d6cb49c7 | 083e1357 |
| D3 | D | 1 | §3 rule 10; §9.3; §12 | "Speed is gated by benchmarks from the first recognizer issue" gates nothing; speed again goes unmeasured until the checkpoint. See also A8, C6. | `Makefile:21-22`; overlaid `b.Fatal` benchmark: `go test -count=1 ./internal/lexer` → `ok`; `-run '^$' -bench DProbe` → `FAIL`. | block | released, against D's bar at d6cb49c7 | 083e1357 |
| D4 | D | 1 | §7.4; §3 rule 12 | The overhead gate measures the step v2 made cheap, not where v2 moved the cost; wider terminal sets show only in the single speed kill. See also B6; C5. | Argued from §7.2, §7.3 and postmortem figures (mapping 110 of 150 ms). Not measured. | block ("resolvable with one reported or gated key") | released at d6cb49c7 | 083e1357 |
| D5 | D | 1 | §9.2; §7.1; §11 | A waypoint-1 exit criterion cannot go green under §11; unclassed and relation-placed operators leave a hole in the superset; the GRAM*/PLUGIN_* scope is not in perigrin's recorded decisions. See also A2, C2; A3, B8. | `grep -n '^func [A-Z]' internal/parse/{operators,declarations,precedence}.go` → no class accessor; `operators.go:15-30`; Records DECISIONS 2026-10-30 names neither; scope added in 399245b6. | block | released (ruling) at d6cb49c7 | 083e1357 |
| D6 | D | 1 | §5.2 GRAM*; §9.2; §2 | Non-program start forms cannot be lexed "exactly as for recursive descent" up front; the start-form test passes only with samples chosen so the difference never shows. See also B7; A10. | `perly.y:186-268` expect per GRAM form; `checkpoint.go:38-48` `StartState*` all `expect: XState`; `{ a => 1 }` as an expression lexes as a block. | stand aside if D5's resolution moves them out; block otherwise | released, against D's stated condition at d6cb49c7 | 083e1357 |
| D7 | D | 1 | §9.2 | Superset and `lex-unmapped` gates disagree over a listed file RD accepts. | `internal/chart/testdata/lex-unmapped-known.txt` has 3 entries; whether RD accepts them under v2 is not measured. | stand aside | released at d6cb49c7 | 083e1357 |
| D8 | D | 1 | §4; §14 | Counts taken at a glance; the 15-day budget is sized against a grammar about 70% larger than stated. See also A4, B1; E6 (reports 33 as checked). | `grep -cE '^[a-z_][a-z_0-9]*\s*:' perly.y` → 67; `awk 'NR>=184' … \| grep -cE '^[a-z_]…\s*$'` → 46; Python strip → 114 LHS, ~348 alternatives; lines 150–182 hold 32 directives. | stand aside on the figures; block on the budget only if perigrin sizes from them | released at d6cb49c7 | 083e1357 |
| D9 | D | 1 | §3 rule 9; §5.2; §5.4 | Two sources of truth for builtin shape; drift surfaces as `rd-accepted-chart-refused` or as a silent misparse. | `internal/parse/declarations.go:711-723`: `CoreShapes` exists "for the chart's CORE.tpm cross-check"; §5.4 discards that cross-check. | stand aside | released at d6cb49c7 | 083e1357 |
| D10 | D | 1 | §5.2; §5.3; §5.4; §7.1–7.2; §9.3; §12 | Implementation detail in a decision makes the accepted decision false on any rename; §9.2's `leo-disagree` gate is defined via a Q4 with no checkable source. See also A13 (sourcing); A10, A12 (i). | `grep -c 'Q9' D/records.md` → 0; no Q4 match; Records line 195 is the review summary. | block on the gate defined via Q4; stand aside on the rest | released at d6cb49c7 | 083e1357 |
| D11 | D | 1 | §10 | "At any checkpoint" is undefined; whoever measures decides when a kill fires. See also C7. | The spec text. | stand aside | released at d6cb49c7 | 083e1357 |
| D12 | D | 1 | §5.2 | Which kinds are trivia is unstated; mapping `HeredocBody` to a terminal set refuses every heredoc file. See also B5 (HeredocBody). | `lexer.go` kind comments (HeredocBody "begin after the line the marker sits on"). | stand aside | released at d6cb49c7 | 083e1357 |
| D13 | D | 1 | §9.4 | The scope record can be corrupted by readings that die. See also E7. | The spec text, plus Earley semantics. | stand aside | released at d6cb49c7 | 083e1357 |
| E1 | E | 1 | §2; §10; §13 | Both waypoint-1 outcomes lead to interleaved mode, so the verdict decides nothing about the approach; the kill is unfalsifiable for the approach as a whole. | `TestScratchExpectedEveryToken` on std5k: `Expected` at all 26,289 nodes median 430.97 ms (lookahead) / 94.24 ms (no lookahead) vs recognition 36.98 ms. | block | released (ruling); disagreement with a consequence recorded at d6cb49c7 | 083e1357 |
| E2 | E | 1 | §9.3; §10 | The long-lived kill figure replays a cache warmed on the same inputs; v1 and v2 define "cold" differently, so neither v2 figure compares with v1's 3.9–5.0. See also C3; C11 (cold). | `TestScratchAggregateColdVsLongLived`: second half (382 files) seen 4.98 ms vs unseen 11.03 ms; 763 files cold 2.90 s vs long-lived 9.65 ms; `TestScratchLadder` replay vs cold; std5k cold/long-lived 1.01. | block | released, against E's own bar at d6cb49c7 | 083e1357 |
| E3 | E | 1 | §1; §2 | The question no longer asks why pvm would want a chart; a positive verdict names no decision it unlocks. | `grep -n 'prior-art\|A1.3' docs/plans/2026-10-30-chart-parser-v2-design.md` → nothing; `git zhi milestone list`: m2-incremental open at 30/38. | block | released; residue raised as E10 at d6cb49c7 | 083e1357 |
| E4 | E | 1 | §13; §5.3 Q2; §9.4 | §13's M2 payoff contradicts spec ch. 6 §6.3.1 and is the stated justification for Q2 and the scope-record interface. | `sed -n 881,925p docs/specs/perl-parser/06-incremental-lsp.md`; `sed -n 975,1000p docs/specs/perl-parser/03-lexer-feedback.md`; `internal/parse/incremental.go:564` `lexer.NewTokenCache`. | stand aside | released; slip raised as E11 at d6cb49c7 | 083e1357 |
| E5 | E | 1 | §9.2; §9.3; §14 | The first reading the kill can act on comes after most of the port, the largest spend. | `TestScratchLadder` (toy, cold): 1 level 2.5 items/token 4.57 µs; 8: 4.8, 3.14; 25: 9.2, 6.09; 32: 10.9, 6.67. | stand aside | held, narrowed; stand aside at d6cb49c7; released at 460adf6f | 083e1357 |
| E6 | E | 1 | §3; §5; §7; §9; §10 | Postmortem items mostly acted on; item 1 (component fit) is only partly acted on and left to lens (a). Reports "33 precedence lines" as checked. See also A4, B1, D8 (count 32). | Table of the seven items; `awk 'NR<=182 && /^%(token\|left\|right\|nonassoc)/' perly.y \| … \| sort -u \| wc -l` → 124 token names vs 20 lexer kinds. | stand aside | released; E corrects its count to 32 at d6cb49c7 | 083e1357 |
| E7 | E | 1 | §9.4; §13 | Waypoint 2's scope record is designed now for a mode out of scope, with an open question about which reading's completion updates it. See also D13. | The spec text. | stand aside | released at d6cb49c7 | 083e1357 |
| E8 | E | 1 | §8 | §8's schemes do not cover the non-operator roles perly.y's precedence lines play; an inexpressible level leads to the twice-amended-definition stop. See also B3, D2. | `sed -n 148,184p /opt/corpora/perl5/perly.y`; precedence split between `expr` (perly.y:1256-1260) and `term`. | stand aside | released, superseded by the ruling; new problem raised as E12 at d6cb49c7 | 083e1357 |

## Clusters

57 findings, at 083e1357. Counted two ways from the "section of the subject" column. A rule (§3 rule N) counts under §3.

### By the first section each finding names

| section | count | findings |
|---|---|---|
| §1 | 1 | E3 |
| §2 | 2 | C8, E1 |
| §3 | 7 | A6, A8, C6, D1, D3, D9, E6 |
| §4 | 3 | A4, B1, D8 |
| §5 | 14 | A1, A7, A10, A11, A12, A13, B2, B5, B7, B8, B9, D6, D10, D12 |
| §7 | 9 | A2, A3, A5, A9, A16, B4, B6, C5, D4 |
| §8 | 4 | A15, B3, D2, E8 |
| §9 | 14 | A14, C1, C2, C3, C4, C9, C10, C11, D5, D7, D13, E2, E5, E7 |
| §10 | 2 | C7, D11 |
| §13 | 1 | E4 |

Within §5: §5.1 1 (A7), §5.2 11 (A10, A11, A12, B2, B5, B7, B8, B9, D6, D10, D12), §5.3 2 (A1, A13).
Within §7: §7 unsubdivided 1 (A16), §7.1 4 (A2, A3, A5, B4), §7.4 4 (A9, B6, C5, D4).
Within §9: §9.1 3 (A14, C1, C10), §9.2 5 (C2, C9, D5, D7, E5), §9.3 4 (C3, C4, C11, E2), §9.4 2 (D13, E7).

### By every section a finding names

| section | count |
|---|---|
| §1 | 1 |
| §2 | 5 |
| §3 | 11 |
| §4 | 3 |
| §5 | 20 |
| §7 | 15 |
| §8 | 4 |
| §9 | 30 |
| §10 | 7 |
| §11 | 2 |
| §12 | 2 |
| §13 | 3 |
| §14 | 2 |

§6 and §15 are named by no finding.

### By raiser

A 16, B 9, C 11, D 13, E 8.

## Per participant

### A (codebase and architecture)

- **Recommendation:** modify.
- **Blocks on:** A1 (decouple scoped as a list that misses the generator's input type, the lattice package and every test fixture), A2 (a waypoint-1 exit criterion unreachable under §11; "the choice belongs to perigrin now"), A5 (the token key discards facts the lexer already decides, against rule 7), A8 (rule 10's gate has no failing mechanism).
- **Says remains:** all sixteen findings open.
- **Did not check:** whether perly.y can be ported rule by rule, or the counts; recognizer speed or memory on a perly-sized grammar, and intern/cache growth over the corpus; word-shaped library operators; edit5k.pl, perl.git `t/`, T1's RD verdicts, the budgets, Leo on the new grammar, §13. Timings are from one loaded 4-core machine; "the ratio, not the milliseconds, is the claim."
- **Brackets given:** A2's empty result could come from the wrong source form (the form used is `coreops_test.go`'s); A9's noise could be machine load; A8 shows Go's default, and only the Makefile was checked for `-bench` (not CI beyond `ci.yml`'s `make test`).

### B (grammar fit)

- **Recommendation:** modify.
- **Blocks on:** B3, B4, B5. "Each is a feasibility claim at the core of waypoint 1 that I could falsify with a command, and each is a form of faithful-not-fitted."
- **Says remains:** all nine findings open.
- **Did not check:** the chart-waypoint postmortem, the milestone `## Records` and the first-attempt design; `internal/parse`'s binding-power table against chapter 4; construct frequency in the Guacamole snippets; all 131 perly.y terminals (sampled 31); the real cost of superposition in the carried recognizer.
- **Brackets given:** perl probes ran on system perl 5.42.0, not blead 5.45; `lx/amb` is a heuristic; `strat.py` is B's reading of §8.

### C (measurement and kill design)

- **Recommendation:** modify.
- **Blocks on:** C1, C3 (definition only), C4, C5 (gating a wall-clock ratio), C6, C7 (defining "checkpoint"), C8, C9.
- **Says remains:** all eleven findings open.
- **Did not check:** perl's verdict on T1; whether a perly.y chart over pvm's tokens accepts or refuses each of the 35 RD-lenient files ("B's feasibility area"); whether a long-lived recognizer's heap slows RD in-process via GC; CPU against wall time; the "~302 / 67 / 39" counts and perl.git `t/`; the Leo, `set-reuse` and `tables-drift` keys beyond noting they pass today.
- **Brackets given:** 4 cores at load ~1.9 with other participants running; v1 timings use Guacamole and a recognizer without Q1/Q3, so "what carries over is the shape, not the size."

### D (spec quality)

- **Recommendation:** modify.
- **Blocks on:** D1, D2, D3, D4, D5, D10 (the gate defined via Q4). D6 blocks unless D5's resolution moves the start forms out; D8 blocks on the budget only if perigrin sizes from the figures.
- **What would change its mind, as stated:** D1 shrinks to "say so, and require the correspondence list" if "hand-ported" means restated in lexer vocabulary with a pointer; "the list stays non-negotiable". D5 and D6 reduce to the dependency problem if perigrin ruled GRAM*/PLUGIN_* in scope off the record.
- **Also recorded:** a scope-shape note that the spec bundles separable work, and six discards (listed in D's report).
- **Says remains:** all thirteen findings open.
- **Did not check:** D4's cost (no v2 table built); whether RD accepts the 3 known-list files (D7); "302 alternatives" exactly; the first design doc beyond the postmortem and subject; spec chapters 2, 3 and 5; toke.c beyond `tokentype_for_plugop` (`perl.c:2701`, `pp_ctl.c:4337`, `toke.c:14216-14642` unchecked); whether a full-grammar run fits the old generator.
- **Brackets given:** the lexer probe ran at 083e1357's lexer; perly.y counts depend on heuristics; the `make test` result is standard `go test` behaviour.

### E (architecture and direction)

- **Recommendation:** modify.
- **Blocks on:** E1, E2, E3.
- **Alignment checks reported as aligned:** rules 1–6 carry over; WRONG 0, the 2^10 collapse and oracle grading kept in waypoint 2; §11 matches the parser milestones; §2's "exactly as for recursive descent" holds and §7.4's denominator is fair.
- **Says remains:** all eight findings open.
- **Did not check:** the "302 / 67" count and whether toke.c's token types fit a (kind, text) table ("lens (a), for others"); RD's timing on the same file sets (E2 compares the chart with itself); any perly.y-derived grammar; whether interleaved mode needs a new lexer; spec chapters 4, 5 and 7 in depth; the rounds of v1's assessment; what Q2's memo does to E1's numbers.
- **Brackets given:** all timings use pu's recognizer with Guacamole tables and `fmt.Sprint` keys; load ~2.

## Rulings during round 1 (perigrin, 2026-10-30)

- **Precedence (B3, D2, E8):** "Conflict filter". `perly.y`'s rules stay as written; the generator derives a filter that rejects exactly what bison's conflict resolution rejects (§8).
- **Library operators (A2, A3, B8, C2, D5):** "Defer to waypoint 2". The `PLUGIN_*` grammar is ported; the `.tpm` extension, its test and any `internal/parse` export are waypoint 2's decision (§9.4).
- **What a verdict means (E1, E3):** "interleaved does more work but can return an answer _sooner_". An up-front kill closes up-front mode only; interleaved is judged on time to an answer (§2, §13).

## Author's response to round 1 (revision after 083e1357)

The author is the orchestrating agent. Each row answers one finding; whether it is released is the raiser's to say in round 2. The subject shrank from 504 to 399 lines: implementation detail moved here (next section), not re-specified.

| id | response | where |
|---|---|---|
| A1 | The decision now states only the requirement: the discard is a single deletion leaving the kept packages compiling and tested. The coupling A1 found (types in `slif.go`, `BuildTables`→probes, `lex.New`/`lex.Guacamole`, kept tests on Guacamole tables) is handed to refinement. | §5.1; hand-off H1 |
| A2 | Ruling: library operators deferred to waypoint 2, so waypoint 1 has no criterion needing an `internal/parse` export. | §9.4 |
| A3 | Same ruling; the 15-class/relation mapping is waypoint 2's decision. | §9.4 |
| A4 | Counts corrected (32 lines; 348 alternatives over 114 nonterminals); the `scan.go` citation removed. | §4, §5.2 |
| A5 | The key now includes the lexer's computed flags (`OpensBlock`, `TermFollows`); `{` is no longer re-decided. | §7 |
| A6 | The key includes the next token's class; next-character ambiguities are not left to scope. | §7, rule 12 |
| A7 | The ratchet's carry-over is restated: keys reset with a new baseline, absolute gates added. | §5.1, §9.5 |
| A8 | "Run in `make test`" removed. The ratchet is in every recognizer issue's acceptance criteria, so its gated keys block each issue; benchmarks run by the ratchet. | §9.3 |
| A9 | `lex-overhead` is measured alternately, per token, at a size where `Tokenize` against itself agrees within 10%. | §9.3 |
| A10 | The `Recognize` parameter is removed from the decision; start forms are start rules, mechanism left to the issue. | §5.2 |
| A11 | The refusal checks a literal by lexing it with `internal/lexer`, not by reading unexported tables. | §5.2 |
| A12 | The code example is removed; Q9 is in the hand-off list. | hand-off H2 |
| A13 | The Q-identifiers are removed from the decision; each item is described in the hand-off list, sourced to the review gate's quality lens. | hand-off H2 |
| A14 | "Accepts" is defined as `Node.Refused()` false on `parse.Parse`'s root. | §9.1 |
| A15 | The cross-check uses chapter 4's table and the exported `parse.LevelsAccountedFor()`. | §8 |
| A16 | Stated as a parity limit: quote-like bodies are single tokens, as for recursive descent today. | §2 |
| B1 | Counts corrected; waypoint 1's proposed budget raised from 15 to 20 days for the larger port. | §4, §14 |
| B2 | Citation removed. | §5.2 |
| B3 | Ruling: conflict filter; B3's cases are the filter's named tests. | §8 |
| B4 | (i) sigil and quote-prefix classes in the key; (ii) next-token class in the key, so FUNC/LSTOP, FUNC1/UNIOP and autoquote are decided lexically; (iii) the lexer's flags kept. | §7 |
| B5 | Fitted citations defined, and the rule families needing fitting named (braces and inserted `;`, quote-likes, formats, labels, attributes, postfix deref, heredocs as trivia). | §5.2 |
| B6 | `terminals-per-token` and `items-per-token` reported. | §7, §9.5 |
| B7 | Non-program start forms are tested only where they lex the same from statement start; the limit is recorded. | §5.2 |
| B8 | Ruling: waypoint 2. | §9.4 |
| B9 | "Used" includes `%prec`. | §5.2 |
| C1 | The superset population is files both perl and RD accept; RD-accepted, perl-refused files are reported (`rd-accepted-perl-refused`). | §9.1, §9.5 |
| C2 | Ruling: waypoint 2. | §9.4 |
| C3 | Long-lived is one recognizer per pass over every file once; no pass replays a cache warmed on its own inputs. | §9.3 |
| C4 | The aggregate is defined: summed time over the superset population excluding std5k.pl and edit5k.pl, with per-corpus shares reported. | §9.3 |
| C5 | See A9; time is per token and alternation removes the separate-median problem; whether chart/lex's time includes `Tokenize` is now stated (overhead is chart/lex's time over `Tokenize`'s). | §9.3 |
| C6 | See A8. | §9.3 |
| C7 | Checkpoints defined (three); chart and RD timed alternately within a measurement. | §9.3 |
| C8 | Waypoint 2 re-measures speed with the forest (checkpoint 3). | §9.3, §9.4 |
| C9 | The port ledger makes completeness a checked criterion. | §5.2, §9.2 |
| C10 | The corpus table now gives measured labels (perl accepts 78 of 124 `parsent`; RD refuses 236 of T1). | §9.1 |
| C11 | The cold figure is reported, not gated; FOLLOW moving to the generator is in the hand-off list. | §9.5; hand-off H2 |
| D1 | Fitted citations and the widened key. | §5.2, §7 |
| D2 | Ruling: conflict filter, which covers `%prec` on non-operator rules. | §8 |
| D3 | See A8. | §9.3 |
| D4 | Superposition's cost is reported per token (B6). | §7, §9.5 |
| D5 | Ruling: waypoint 2; GRAM*/PLUGIN_* scope is now recorded in this minute's rulings and in §5.2. | §5.2, §9.4 |
| D6 | See B7. | §5.2 |
| D7 | Files on the `lex-unmapped` known list are outside the superset population. | §9.1 |
| D8 | Counts corrected; budget raised. | §4, §14 |
| D9 | The CORE.tpm cross-check is replaced, not dropped: the port's builtin classes against `parse.CoreShapes()`. | §5.3, rule 9 |
| D10 | Implementation detail moved here (code example, table layout, preparation checklist, test names); `leo-disagree` is defined in the decision itself (acceptance and `Expected` agree). | §9.2; hand-off |
| D11 | See C7. | §9.3 |
| D12 | Heredoc bodies are trivia. | §5.2 |
| D13 | The scope record's design moves wholly to waypoint 2; the decision keeps only that scope is never held in items. | rule 12, §9.4 |
| E1 | Ruling: the verdict's meaning is stated; interleaved is judged on time to an answer. | §2 |
| E2 | See C3; long-lived now matches the first attempt's measurement. | §9.3 |
| E3 | §1 restates why a chart, and §2 what a positive verdict unlocks. | §1, §2 |
| E4 | §13's claim corrected: chapter 6 re-lexes from lexer checkpoints; chart-based resumption is a hypothesis for the interleaved spike. | §13 |
| E5 | Every recognizer issue reports speed before checkpoint 1. | §9.3 |
| E6 | Component fit is answered by the widened key and fitted rules. | §7, §5.2 |
| E7 | See D13. | §9.4 |
| E8 | See D2. | §8 |

## Handed to refinement

Moved out of the decision; refinement turns them into issues. Sources: the milestone review gate's quality and over-engineering lenses (chart-waypoint, 2026-10-30) and round 1 above.

- **H1, separation of kept from discarded code (A1):** the generator's input types live in `slif.go`; `BuildTables`/`GenerateTables` call the probe and conflict tooling; chart/lex's `New` takes `*grammar.CutGrammar` and `lex.Guacamole` is a third loader; the kept recognizer and Leo tests and the fuzz target run on Guacamole's tables and need a toy or new grammar; `cmd/ratchet/main.go` imports chart/lex and Guacamole paths.
- **H2, fixes to the kept code from the review gate:** token and lookahead class keys built once per token as integer ids, not `fmt.Sprint` per node (quality Q1); `Expected(pos)` memoized per interned set and the origin sets it read (Q2); FIRST and FOLLOW computed by the generator and stored on the tables (Q3); `leo-disagree` comparing acceptance and `Expected` (Q4); `Recognize` validating its lattice (Q7); `Recognizer` documented as not safe for concurrent use (Q8); `BuildLR0` refusing a sequence rule whose right-hand side is not one symbol (Q9; A12 reproduced the panic); `WriteTables` deleting only the table files it owns (Q10); unmapped files excluded from refusal counts (Q11); and the over-engineering lens's cheap deletions (the unread `Transition.Items`, the `postdot`/`after` duplicate, the `chart_noleo` tag, ratchet rules for never-printed keys).
- **H3, mechanisms:** how start forms are selected (A10); the table's layout and lookup form; benchmark and test names; how the port ledger is stored.

## Round 2

**Revision assessed:** d6cb49c7 (d6cb49c71eb8caeae30a6936e4f222e660e6e708). All six participants name this revision as the one they read. Code read at `origin/pu` 4fffe4d6; perl at 94e5086608.
**Participants:** A–E resumed, with the round-1 lenses. **F** joins as a cold reader: it read the subject without the prior findings and did not open this minute or any other report.
**Round 2 dispatched:** 2026-10-30.

The round-1 rows' status column above now carries each raiser's round-2 word. New findings are below. Section references are to the subject at d6cb49c7.

### Recommendations, as declared

- **A: modify.** Blocks on A17 and A18, "both are one-sentence fixes".
- **B: modify.** "The rulings stand. I record no disagreement with them. B10 concerns how §8 states the filter, not the choice to have one."
- **C: modify.** "The revision fixed most of what I raised … What remains is definitional and each fix is one sentence: C1's predicate, C4's dominant file, C9's reference, and C12's denominator. C13 and C15 are new problems in the gate schedule." Records no disagreement with the three rulings.
- **D: modify.** "D1–D13 are all released. Rulings are perigrin's. I disagree with no ruling. D15 is about how the ruling is cited and checked."
- **E: modify.** E12 is about "the contract and mechanism the text attaches to" the conflict-filter ruling, not the ruling.
- **F: modify.** "The question, the kill criteria, the isolation rule and the source counts hold up when measured. Two mechanisms do not hold up as written: the precedence filter (§8), and the list of fitting families (§5.2)."

### Blocks open after round 2, as declared

| participant | blocks |
|---|---|
| A | A17, A18 |
| B | B10, B11 |
| C | C1 (held, re-scoped), C4 (held, re-scoped), C9 (held, re-scoped), C12, C13, C15 |
| D | D15 |
| E | E12 |
| F | F1, F2 |

Fourteen blocks are open.

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A9 (re-scoped) | "I leave the block to C5; a size that satisfies this precondition cannot be shown to exist." |
| A19 | "§9.2's `leo-disagree 0` and rule 11 would catch a Leo bypass, which is why I stand aside." |
| A20 | "A check that used `Checkpoints`/`Resume` from a captured operator-position state would avoid the problem, which is why I stand aside." |
| A21 | None stated beyond the row. |
| B12 | "The port ledger would catch all three, so this is sizing." |
| B13 | "This is implementation; hand it to refinement (H3)." |
| B4 residual (`sub {` as two readings) | Recorded by B inside B4's release: "Rule 12's scope collapse covers it in waypoint 2." |
| C5 remainder | "Say which size the rule selects (the aggregate) rather than leave it to the measurer." |
| C14 | "This is D7's row, and refinement can carry it." |
| D14, D16, D17, D18 | None stated beyond each row. |
| E5 (held, narrowed) | None stated beyond the row. |
| E9, E10, E11 | None stated beyond each row. |
| F3–F9 | None stated beyond each row. |

### Releases the participant said were against its own bar

Each resumed participant declared a general bias toward finding its own bar met. These are the releases where it said more than that, in its words:

- **B5:** "The ledger would discover a missing family, so the omission is sizing, not correctness. I release against my own bar here."
- **C3, C5 (core), C6, C7:** each "against my own bar". C: "I release C3, C5's core, C6, C7 and C8 against that bar" (the header), though C8's row says only "Released".
- **D1:** "In round 1 I asked for a per-*terminal* correspondence list. What exists is a per-*alternative* ledger. That covers rules but not terminals, and the gap is D18."
- **D3:** "I asked for a gated per-issue speed key. Per-issue speed is only *reported* (§9.5), so a 6× issue blocks nothing until checkpoint 1. I accept that because of C5's A/A evidence."
- **D6:** "released, against my condition". "I said 'stand aside if moved out'. The forms stayed in, but §5.2 limits their tests … That is honest, so I accept it."
- **E2:** "That is the resolution I asked for. I am releasing against my own bar here." Residue noted, not held: for std5k.pl alone, long-lived and cold "measure the same thing" (see F7).

A did not mark any single release as against its bar; its header says "Where I release a finding, I re-ran its proof and say what else the fix would have had to change."

### Disagreements with a ruling's consequences, as recorded

- **E1:** "Both verdicts still lead to an interleaved spike, so waypoint 1 cannot stop the approach as a whole. That is the ruling's intent, so it is not a finding." E does not hold E1 on it.
- **C (library operators):** "the superset population at WP1 has no library-operator files, because `parse.Parse` has no loader. So nothing at WP1 exercises the `PLUGIN_*` rules that were ported." Recorded as a consequence; C records no disagreement with the ruling.
- A, B, D and F record no disagreement with a ruling.

### Measurements that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **Lexing an operator literal on its own.** A20: `/` gives Error, `%` gives UnknownRest; 15 of 31 literals not one Operator. B13, D14 and F4: `/` gives UnknownRest, `%` gives Error. D14 lists eight literals, F4 six, B13 six. All four agree that every literal lexes correctly after a term.
- **The `Tokenize` A/A precondition.** A9: at k = 1, 4 and 16 copies of std5k, one trial in four falls outside 10%, so "no size reliably meets the precondition here". C5: on the aggregate, A/A stays within 10% in 10 of 10 trials; on std5k, outside 10% in 4 of 10.
- **T1's size and RD's refusals.** C1: under a full-tree walk, RD refuses 236 of 986, and the population holds 691 T1 files; "the gate now covers the files RD itself refuses (221 in T1 alone)". A17: root-only 0, any node 236. F: "996 directory entries against 986 files claimed" (not checked further).
- **Seen against unseen replay.** E2 re-run: 1.5× (round 1: 2.2×); std5k cold against replay 1.29 (round 1: 1.01). C3 re-run on the v1 tree: per-run 25.0/20.1/20.8 ms against replay 8.6/10.3/9.3 ms.
- **The round-1 count of precedence lines.** E6 now reports 32, correcting its round-1 report of 33. The round-1 "Measurements that differ" entry on this is settled by E's own correction.
- **Afroozeh et al. (2013).** D15 relies on the abstract. E12 and F1 write from memory and did not read the paper.

### Round-2 ledger

New findings. Columns as in the round-1 ledger.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A17 | A | 2 | §9.1 (from A14's fix) | Under the stated definition RD "accepts" every file; the population becomes "every file perl accepts", and gated `rd-accepted-chart-refused 0` demands the chart accept files RD refuses; §9.1's "it refuses 236" contradicts its own definition. See also C1 (held), A14. | `refused.go:14` is a per-node test; `TestRootRefused`: `my $x = ;`, `foo(;`, `1 2;` give root.Refused false, subtree true; over T1's 986 files root.Refused **0**, any node **236**. | block | released at 460adf6f | d6cb49c7 |
| A18 | A | 2 | §9.3; §9.5 (from A8's fix) | With the ratchet in every recognizer issue's criteria, `rd-accepted-chart-refused` "gated at 0" and `lex-unmapped` "equals its known list" hold only once the port is complete; every earlier issue fails, or the port is forced into one issue. See also C15. | §9.5 rows; §7 "An empty set refuses the file … (`lex-unmapped`)"; `Compare` on pu gates only `mayNotRise` against a baseline; `TestRDTotal`: one RD pass over 1,879 files 0.46–0.52 s. | block | released at 460adf6f | d6cb49c7 |
| A19 | A | 2 | §8; §5.1 | "Static data applied at completion" names a mechanism the carried recognizer cannot host unchanged; the carried recognizer becomes a modified one the 770-file reference agreement no longer covers. See also B10, D15, E12, F1 (e). | `recognize.go:70-73` `r.lhs[s]` compacted LHS set; `:255` completes by LHS, rule identity gone; `leo.go:37-45` memo jumps by `(state, symbol)`. | stand aside | released (ruling) at 460adf6f | d6cb49c7 |
| A20 | A | 2 | §5.2 refusals (from A11's fix) | "Checked by lexing it" refuses correct literals lexed alone in term position. See also B13, D14, F4. | `TestLiterals`: alone, `**`→`*` `*`, `/`→Error, `//`→Quote, `%`→UnknownRest, `&`→FuncSigil, `<=>`→Readline, `x=`→Word `x` + `=` (15 of 31 not one Operator); after `1 ` all 31 lex as one token; `x=`, `x`, `lt`, `cmp`, `isa` are Word. | stand aside | released at 460adf6f | d6cb49c7 |
| A21 | A | 2 | §5.3; §3 rule 9 | The CORE.tpm cross-check's difference list is mostly definitional unless the shape↔token correspondence is stated. See also D9. | `TestShapes`: `parse.CoreShapes()` 207 names (list 91, unary 84, niladic 32); `not` is list (NOTOP), `goto` unary (LOOPEX), `last`, `my`, `local` absent. | stand aside | released at 460adf6f | d6cb49c7 |
| B10 | B | 2 | §8 | The filter, "static data applied at completion", cannot be exact for perly.y, and in waypoint 1 it can only cause false refusals; "bison's language" is undefined once rules are fitted. See also A19, D15, E12, F1. | `consist.py` over a `-Dlr.default-reduction=accepting` build: 706 states, 680 suppressed (rule, token) pairs, one reduced elsewhere (370 `indirob: scalar`, `[`); Earley + filter accepts `1..2..3` and `$a isa B isa C` (bison rejects) and rejects `$$x[0]` (bison accepts); 5,606 random sentences all accept. | block | released (ruling) at 460adf6f | d6cb49c7 |
| B11 | B | 2 | §7 | The key has no lookbehind: a lone word in a subscript `{}` stays {string, call}; waypoint 2 scope picks the call reading for declared subs and has no fact for builtins, giving WRONG sites on most hash accesses. | perl Deparse: `{ print }` → `print $_`; `$h{print}`, `$h{shift}` → string keys; `lx/subkey` on T1: 1,757 lone words in non-block `{}`, 12 in blocks with the same key. | block | held; block at 460adf6f; released, against B's own bar on glob slots (B15) at 8e3aa4b5 | d6cb49c7 |
| B12 | B | 2 | §5.2 | The named fitting families miss variables (17 alternatives), forced version tokens (5 alternatives) and the end-of-file `;`. See also F2; D18. | bison rule list via `r2/earley.py`; `S_tokenize_use`; perly.y comment ~720; toke.c:7670. | stand aside | released at 460adf6f | d6cb49c7 |
| B13 | B | 2 | §5.2 refusals | "Checked by lexing it" depends on lexer state and rejects MULOP, BITANDOP, POWOP and others. See also A20, D14, F4. | `lx/dump` alone: `/`→UnknownRest, `%`→Error, `&`→FuncSigil, `**`→`*` + Word `*`, `~~`→two `~`, `x=`→Word `x` + `=`. | stand aside | released at 460adf6f | d6cb49c7 |
| C12 | C | 2 | §9.3 (`lex-overhead`) | "Per token" does not name the count; each side over its own count inflates a correct chart/lex to ~1.7–2.0, which always fails. See also F7. | `zzspeed2`: std5k tokens all 42,172, significant 26,288 (ratio 1.60). | block | released, against C's own bar at 460adf6f | d6cb49c7 |
| C13 | C | 2 | §9.3 (checkpoint 1); §10 | Checkpoint 1 needs std5k and the whole superset population, close to the WP1 exit; the speed kill cannot fire until most of the budget is spent; early aggregate treatment of still-refused files unstated. See also E5. | §9.3 against Records 2f46 (kill fired one day into 15) and the postmortem's "What Worked Well". | block | released, against C's own bar; residue raised as C18 at 460adf6f | d6cb49c7 |
| C14 | C | 2 | §9.1; §9.2 | The `lex-unmapped` exclusion lets a chart table gap leave the superset gate by filing a lexer issue. See also D7. | §7 "An empty set refuses the file at that byte"; §9.1. | stand aside | superseded: re-scoped into C17 at 460adf6f | d6cb49c7 |
| C15 | C | 2 | §9.3; §9.5 | The absolute gates make every intermediate issue unpassable, the shape postmortem change 6 forbids. See also A18. | Criterion run over a half-done tree: superset count in the hundreds, ratchet criterion fails. | block | released, against C's own bar at 460adf6f | d6cb49c7 |
| D14 | D | 2 | §5.2 refusals | The check refuses correct classes, so the generator emits no tables for a correct grammar. See also A20, B13, F4. | `go test -overlay D/overlay3.json -run TestDProbe2 -v ./internal/lexer`: alone `/`→UnknownRest, `%`→Error, `&`→FuncSigil, `**`→`*` Word `*`, `//`→Quote, `%=`→Variable, `<=>`→Readline, `x=`→Word `x` `=`; after `$x ` all correct, including `Word«x=»`. | stand aside | released at 460adf6f | d6cb49c7 |
| D15 | D | 2 | §8; §4 (Afroozeh) | The cited technique's safety property keeps the unannotated language, the opposite of "rejects exactly what bison's resolution rejects"; a worker following it accepts `1 .. 2 .. 3`, seen only in a reported key. Also: fitted rules must keep perly.y precedence (HASHBRACK 1536, and implicit on 3 alternatives); the filter has no outside reference (rule 11). See also B10, E12, F1, A19. | `perl -c -e '@a = (1 .. 2 .. 3);'` syntax error; `perl -c -e '$a + defined $b;'` OK; Afroozeh et al. SLE 2013 abstract; `internal/parse/precedence.go:242` `NonassocOperators`; Python scan: 3 alternatives with implicit precedence on fitted-family terminals. | block | released (superseded by the ruling) at 460adf6f | d6cb49c7 |
| D16 | D | 2 | §9.2; §5.2 port ledger; §4 | Completeness is checked against nothing outside the ledger; a ledger missing 20 alternatives passes. See also C9 (held), F5. | The spec text; bison 3.8.2 `--report` enumerates 348 alternatives with line numbers. | stand aside | released at 460adf6f | d6cb49c7 |
| D17 | D | 2 | §4; §9.1 | The grammar (blead 5.45) and the labels (perl 5.42) are different perls; files using newer syntax drop out and inflate `chart-accepted-perl-refused`. See also F8. | `perl -v` 5.42.0; `git -C /opt/corpora/perl5 log --oneline v5.42.0..HEAD -- perly.y` (f1a8d7d883 PPC0024, da6c78e212 multivariable foreach). | stand aside | released, against D's bar at 460adf6f | d6cb49c7 |
| D18 | D | 2 | §5.2 refusals; §9.2 (re-scope of D1) | A ported rule whose terminal has no token class is dead and passes every refusal and the ledger; `PLUGIN_*_OP` is in this state for all of waypoint 1. See also F2, B12. | §5.2 and §7 text; `D/tokens.txt`, `D/allterm.txt`. | stand aside | released at 460adf6f | d6cb49c7 |
| E9 | E | 2 | §2; §13 | §2 glosses "time to an answer" as M2's latency, but M2's gate is an edit to a re-parsed tree in < 10 ms p99; the interleaved spike could be judged on a metric M2 does not use. | `sed -n 2582,2596p docs/specs/perl-parser/06-incremental-lsp.md`; `git zhi issue show 01a17969-4736`; plan §5 M2. | stand aside | released, partly against E's own bar at 460adf6f | d6cb49c7 |
| E10 | E | 2 | §1; §2; §5.1 | A positive verdict makes the chart "a candidate engine for M2" while convergence goes unmeasured through both waypoints; the convergence fixtures' fate is unstated. | `internal/chart/testdata/convergence-edits.jsonl`, `convergence-edits-guacamole.jsonl` on pu; v1 §8. | stand aside | released; small residue noted at 460adf6f | d6cb49c7 |
| E11 | E | 2 | §13 | §13 attributes this assessment's measurement to the first attempt and quotes the lookahead-on figure where lookahead 0 applies. | Minute row E1 (431 ms lookahead on, 94 ms lookahead 0); re-run 379 / 103 ms. | stand aside | released at 460adf6f | d6cb49c7 |
| E12 | E | 2 | §8; §9.2 | §8 states two contracts ("exactly the derivations" vs "bison's accepted language"); under the first, the named mechanism cannot pass §8's own `$a * defined $b + 1`; under the second it passes whether or not the filter works. The filter also decides speed. See also B10, D15, F1, A19. | `E/r2/toy.y` (bison 3.8.2): `t*ut+t` → `(t*u((t+t)))`, matching perl Deparse; `depth1.py` parent/child filter keeps 2 readings; `t*ut` + n `+t` keeps n+1 (n = 1..8); round-1 `TestScratchLadder`: flat 751.9 items/token vs stratified 2.5. | block | released (superseded by the ruling); residue raised as E15 at 460adf6f | d6cb49c7 |
| F1 | F | 2 | §8 | The filter's contract cannot be met or checked as stated: (a) undefined over fitted rules; (b) level-derived parent/child patterns miss rule 370 (follow restriction) and open-rightmost rules 182–185, 337; (c) the 667 reduce resolutions are not expressed; (d) at completion, items still grow linearly; (e) Leo composition unstated. See also B10, D15, E12, A19. | `bison --report=all`: 1,347 resolved S/R conflicts (670 shift, 667 reduce, 10 error) over 57 rules (`F/conflict-rule-text.txt`); Earley probe `E→E+E\|E*E\|a`, items/token unfiltered/completion/stratified at 160 terms 201.9/103.1/4.0. | block | released, against F's own bar at 460adf6f | d6cb49c7 |
| F2 | F | 2 | §5.2 fitting families | Refinement under-sizes the port: variables/indirob, NOAMP calls, method-call tokens, scope-decided `*SUB` and PRIVATEREF, phasers and keyword plugins need fitting, and none is named; §7 maps a token to a set, never a sequence. See also B12; D18. | 35 of 348 alternatives reference these terminals (`F/rules-lhs.txt`); toke.c `force_next(BAREWORD)`, `TOKEN(NOAMP)`. | block | released at 460adf6f | d6cb49c7 |
| F3 | F | 2 | §7 | The key is underspecified: "class of the next significant token" may be (kind, text) or a terminal set; "and similar" flags do not exist; pairs such as unary/binary `-` and `x` are not decided and become superpositions. | `lexer.Token` on pu has only `OpensBlock`, `TermFollows`; its comment says expect state is deliberately not exported; litlex output. | stand aside | released; re-scoped into F10 at 460adf6f | d6cb49c7 |
| F4 | F | 2 | §5.2 refusals | The generator refuses a correct grammar. See also A20, B13, D14. | `F/pu/cmd/litlex`: alone `/`→UnknownRest, `/=`→UnknownRest, `%`→Error, `%=`→Variable, `**`→`*` + Word `*`, `&`→FuncSigil; in `$a OP 1` all Operator. | stand aside | released at 460adf6f | d6cb49c7 |
| F5 | F | 2 | §5.2 port ledger | "Every perly.y alternative" has no mechanical anchor; the out-of-range refusal needs a range nothing reads. See also C9 (held), D16. | §4 "no build step reads it"; §9.2. | stand aside | released at 460adf6f | d6cb49c7 |
| F6 | F | 2 | §5.1; §6 | The carried recognizer takes terminal names and interns with `fmt.Sprint` per edge; "works only in ids" needs a change to `lex.Edge` and `builder.scan` that §5.1 does not name. See also H2 (Q1). | `recognize.go` `builder.scan`, `Recognizer.class`, `cacheKey` construction in `Recognize`. | stand aside | released at 460adf6f | d6cb49c7 |
| F7 | F | 2 | §9.3 | For std5k, long-lived equals cold, so "this is the first attempt's measurement" is wrong for std5k (the postmortem's figure was warm); alternation lets one side's GC land in the other's timing; `lex-overhead`'s numerator unstated. See also C12; E2 residue. | §9.3; postmortem "34.9 → 18.3 ms warm"; speed keys `na` in pu's baseline. | stand aside | held, re-scoped (GC between alternating runs); stand aside at 460adf6f; held; stand aside at 8e3aa4b5; released at b85dc97d | d6cb49c7 |
| F8 | F | 2 | §4; §9.1 | Grammar version and label oracle disagree; blead syntax 5.42 rejects is accepted and only reported; a relabel moves the gates. See also D17. | 38 perly.y commits between v5.42.0 and 94e50866 (+641/−356), including PPC0024. | stand aside | released at 460adf6f | d6cb49c7 |
| F9 | F | 2 | §13; §7 | The §7 key needs the next token, but interleaved mode runs at lookahead 0, so a positive up-front verdict does not transfer its token table to the next spike. | §7 key; §13 "lookahead 0". | stand aside | released at 460adf6f | d6cb49c7 |

### Round-2 clusters

Counted by the first section each finding names, as in round 1. A rule (§3 rule N) counts under §3.

**Round-2 new findings (31, at d6cb49c7)**

| section | count | findings |
|---|---|---|
| §1 | 1 | E10 |
| §2 | 1 | E9 |
| §4 | 2 | D17, F8 |
| §5 | 10 | A20, A21, B12, B13, D14, D18, F2, F4, F5, F6 |
| §7 | 2 | B11, F3 |
| §8 | 5 | A19, B10, D15, E12, F1 |
| §9 | 8 | A17, A18, C12, C13, C14, C15, D16, F7 |
| §13 | 2 | E11, F9 |

Within §5: §5.1 1 (F6), §5.2 8 (A20, B12, B13, D14, D18, F2, F4, F5), §5.3 1 (A21).
Within §9: §9.1 2 (A17, C14), §9.2 1 (D16), §9.3 5 (A18, C12, C13, C15, F7).

By raiser: A 5, B 4, C 4, D 5, E 4, F 9.

**Rows open at d6cb49c7 (38: the 31 round-2 findings plus 7 round-1 rows held)**

The held round-1 rows are A9, A14, C1, C4, C5 (remainder), C9 and E5. Their sections are from the round-1 ledger.

| section | count | findings |
|---|---|---|
| §1 | 1 | E10 |
| §2 | 1 | E9 |
| §4 | 2 | D17, F8 |
| §5 | 10 | A20, A21, B12, B13, D14, D18, F2, F4, F5, F6 |
| §7 | 4 | A9, B11, C5, F3 |
| §8 | 5 | A19, B10, D15, E12, F1 |
| §9 | 13 | A14, A17, A18, C1, C4, C9, C12, C13, C14, C15, D16, E5, F7 |
| §13 | 2 | E11, F9 |

**Open blocks by section (14):** §5 1 (F2), §7 1 (B11), §8 4 (B10, D15, E12, F1), §9 8 (A17, A18, C1, C4, C9, C12, C13, C15).

### Per participant, round 2

#### A (codebase and architecture)

- **Recommendation:** modify.
- **Released:** A1, A2 (ruling), A3 (ruling), A4, A5, A6, A7, A8, A10, A11, A12, A13, A15, A16.
- **Held:** A9, re-scoped, stand aside; A14, "held, superseded by A17".
- **New:** A17 (block), A18 (block), A19, A20, A21 (stand aside).
- **Checks reported:** §7's widened key can be built against pu, with two qualifications (only two flags exist; `x`, `x=`, `lt`, `isa` arrive as Word). The port ledger can be built: no two of the 348 alternatives start on the same line; five have their first symbol on a later line than their `:`/`|`. The CORE.tpm cross-check can be built (A21).
- **Did not re-examine:** the conflict filter's correctness against bison (B's lens); perl's verdicts behind "78 of 124 parsent", §13's `Expected` timings, the budgets, chapter 6's checkpoint design; whether a real-size token table keeps chart/lex within 1.3.

#### B (grammar fit)

- **Recommendation:** modify.
- **Released:** B1, B2, B3, B4 (remainder raised as B11; `sub {` residual standing aside), B5 (against its own bar; omission raised as B12), B6, B7, B8 (ruling), B9.
- **Held:** none.
- **New:** B10 (block), B11 (block), B12, B13 (stand aside).
- **Re-run:** every round-1 command gives identical output at d6cb49c7.
- **Did not re-examine:** §9.3 and §9.1's corpus figures; §13; H1 and H2; whether a parent-item filter is exact for perly.y ("I showed only that the global form is not"); real perl t/ files through the toke.c-level oracle.

#### C (measurement and kill design)

- **Recommendation:** modify.
- **Released:** C2 (ruling), C3, C6, C7 (each against its own bar), C5's core (against its own bar), C8, C10, C11.
- **Held:** C1 (re-scoped to the predicate, block), C4 (re-scoped to the dominant file, block), C9 (re-scoped to the ledger's reference, block), C5's remainder (which size the A/A rule selects, stand aside).
- **New:** C12 (block), C13 (block), C14 (stand aside), C15 (block).
- **Did not re-examine:** §7's key and §8's filter (B's lens); why T1's 74 perl rejections happen; chart-side aggregate shares; GC cross-effects of long-lived heaps on alternating runs; whether `parse.LevelsAccountedFor()` exists on pu.

#### D (spec quality)

- **Recommendation:** modify.
- **Released:** D1 (against its bar; remainder re-scoped as D18), D2 (ruling), D3 (against its bar), D4, D5 (ruling), D6 (against its stated condition), D7–D13.
- **Held:** none.
- **New:** D15 (block); D14, D16, D17, D18 (stand aside). Three discards and one note (a deterministic `items-per-token` may-not-rise ratchet as a per-issue speed signal, "not a finding").
- **On the hand-off list:** "Mostly no" over-move; H1–H3 moved correctly; two places under-specified (D15, D16).
- **Did not re-examine:** the full text of Afroozeh et al. (abstract only); whether a bison-exact completion filter is achievable for every conflict; the speed protocol's numbers; §1–§2's verdict semantics, §13, the per-row responses for A, B, C and E, and the budget; every toke.c citation beyond `tokentype_for_plugop`.

#### E (architecture and direction)

- **Recommendation:** modify.
- **Released:** E1 (ruling; disagreement with a consequence recorded), E2 (against its own bar; std5k residue noted, not held), E3 (residue raised as E10), E4 (slip raised as E11), E6 (with its correction to 32), E7, E8 (superseded by the ruling; new mechanism's problem raised as E12).
- **Held:** E5, narrowed, stand aside. "One sentence saying what happens when an early-warning ratio on std5k.pl is above 3×."
- **New:** E9, E10, E11 (stand aside), E12 (block).
- **Did not re-examine:** §7's key feasibility (lens a); the port ledger's mechanism and the 348/114 count; §9.1's corpus labels; whether `parse.LevelsAccountedFor()` is exported; the `lex-overhead` protocol; Afroozeh et al.; any perly.y-derived measurement; whether perly.y's resolution removes sentences outside `%nonassoc`.

#### F (cold reader)

- **Recommendation:** modify.
- **Released / held:** none; F had no prior findings.
- **New:** F1, F2 (block); F3–F9 (stand aside).
- **Requirement classification, as given:** Blocking: §8 precedence filter (F1). Partial: §3 rules 7–12 (F6, F5), §5.1 (F6), §5.2 fitted grammar (F2), §5.2 refusals (F4), §5.2 port ledger (F5), §7 (F3), §9.1 (F8), §9.3 (F7), §13 (F9). Ready: §2, §4, §5.2 start forms and PLUGIN_*, §9.2, §9.5, §10, §11, §12, §14 ("perigrin's call").
- **Three axes, as given:** codebase "mostly aligned"; architecture: rules 1–6 and 12 consistent, rule 4's 2^10 collapse and waypoint 2's 1% kill depend on F1; direction "aligned with M2 and ch. 6's latency framing", with risks in F9 and F1 (d).
- **Did not check:** Afroozeh et al.'s coverage of deep and open-rightmost conflicts (no access to the paper); `items-per-token` on real perly.y and expression-length distribution; whether split LR(0)ε states map onto bison's LALR states; how 2f46 timed std5k; the T1 file count and RD's 236; budget realism, chapter 4's table against the 32 lines, the CORE.tpm cross-check's content.

### Round outcome

**Another round owed.** Fourteen blocks are open: A17, A18, B10, B11, C1, C4, C9, C12, C13, C15, D15, E12, F1, F2. No participant has released them. 38 rows are open in all.

## Rulings during round 2 (perigrin, 2026-10-30)

- **Precedence (B10, D15, E12, F1, A19):** "Option 2". Precedence leaves recognition entirely. Recognition uses `perly.y`'s rules as written and is ambiguous on operator chains. In waypoint 2, precedence and associativity are a Boolean constraint evaluated over the forest (semiring-style), with safe semantics (Afroozeh et al.). Bison is the referee, not the mechanism; perl is the oracle. This supersedes round 1's "conflict filter" ruling.
  - **Evidence given before the ruling (author's probe, 2026-10-30, not a participant's).** Recursive descent's trees counted maximal chains of binary, `?:` and comparison operators without grouping parens; list commas were excluded, since `listexpr` is unambiguous (`perly.y:1266-1273`).

    | Corpus | Chains | Mean | p99 | Max |
    |---|---|---|---|---|
    | std5k.pl | 1,217 | 1.45 | 5 | 5 |
    | Guacamole snippets | 123 | 1.11 | 3 | 4 |
    | T1 (1,508 files) | 26,034 | 1.26 | 3 | 10 |
    | perl's dual-life modules (947) | 60,537 | 1.29 | 4 | 22 |
    | Debian `/usr/share/perl5` (212) | 8,545 | 1.24 | 4 | 15 |

    Σn²/Σn is 1.3–2.1. The longest chains are `and`/`or` cascades, `&&`/`||` guard ladders, `.` assembly and `?:` ladders. Mainstream CPAN (Moose, Mojolicious, DBI) could not be fetched: the container has no CPAN access.
- **Recursive descent follows bison (asked, answered):** RD's Pratt loop comes from `perly.y`'s precedence block (chapter 4), refuses `%nonassoc` chains as perl does (`internal/parse/expr.go:11-21`), and is graded against perl by the oracle.

## Author's response to round 2 (revision after d6cb49c7)

The author is the orchestrating agent. Release is each raiser's to give.

The subject grew from 3,247 to about 3,640 words. Most of the growth is the ruling-driven rewrite of §8 and its consequences in rule 6 and §6. The rest answers findings that asked for a definition (A17, C12, C13, C4), not for added mechanism. Mechanics moved to H4 and H5 below.

| id | response | where |
|---|---|---|
| A17 / C1 | "Accepts" now means no node in the tree reports a refusal, the measure behind the 236. | §9.1 |
| A18 / C15 | The absolute values are the waypoint's exit, met by its final issue. Every earlier issue holds only the ratchet's no-regression rule. The key table now says "may not rise; X at the exit". | §9.2, §9.3, §9.5 |
| A19, B10, D15, E12, F1 | Ruling: option 2. There is no filter in recognition; the recognizer stays Boolean; the precedence contract is safe semantics over the forest, with bison as referee. | §8, rule 6, §6 |
| A20, B13, D14, F4 | A literal is checked where its class can stand (an operator between two terms, a keyword at statement start); the exact checks are in H5. | §5.2; H5 |
| A21 / D9 | The CORE.tpm cross-check states a shape-to-token-type correspondence; its table is in H5. | §5.3; H5 |
| B11 | The key adds the previous significant token's class, for a lone word in a subscript's `{}`. | §7 |
| B12, F2, D18 | Token-to-sequence is ruled out: rules are fitted to the single token. The families are in H4 (variables and indirob, the inserted `;` at `}` and EOF, forced version tokens, NOAMP, method calls, scope-decided `*SUB`/`PRIVATEREF` as superpositions, phasers, keyword plugins excluded). A terminal with no token class is now a refusal, with `PLUGIN_*` excepted as waypoint 2's. | §5.2; H4, H5 |
| C4 | The aggregate excludes any file taking more than 10% of RD's time (`code_too_large.t`), and reports it. | §9.3 |
| C9, D16, F5 | The port ledger is anchored by a committed reference list of `perly.y`'s alternatives at the pinned revision, which is also what citations are checked against. | §5.2 |
| C12, F7 (numerator) | `lex-overhead` is a ratio of totals on the same files (chart/lex including its `Tokenize` call, over `Tokenize`), so token counts don't enter. | §9.3 |
| C13, E5 | Checkpoint 1 is "as soon as std5k.pl is recognized", with the aggregate over the recognized part and that fraction reported. The kill may fire at any checkpoint. | §9.3 |
| C14 | A known-list entry must name an open lexer issue and show a lexer error token in that file. | §9.2 |
| D17, F8 | The label perl's version is recorded beside the grammar's revision; differences show in `chart-accepted-perl-refused`. | §9.1 |
| E9 | The interleaved spike is judged on M2's measure: an edit to a re-parsed tree under 10 ms at p99. | §13 |
| E10 | The convergence edits are kept as fixtures; a positive verdict makes the chart an M2 candidate only once convergence is measured. | §5.1, §13 |
| E11 | Re-attributed to this assessment, with the lookahead-off figure (94 ms). | §13 |
| F3 | A neighbour's class is defined (kind plus text for words and operators, never a terminal set). The flags are named exactly. Unary/binary `-` and `x` are decided by position in the grammar. | §7 |
| F6 | Covered by H2 (Q1, integer class ids). | H2 |
| F7 (std5k) | For std5k.pl, long-lived and cold coincide; the first attempt's figure was warmer. | §9.3 |
| F9 | Interleaved mode needs its own token key, since §7's reads the next token. | §13 |

## Handed to refinement (additions)

- **H4, rule families needing fitting:**
  - variables and indirect objects (the lexer's single `Variable` token for `PERLY_DOLLAR` plus a name; 17 alternatives);
  - braces, and the `;` toke.c inserts before `}` and at end of file;
  - quote-likes and interpolation (single tokens);
  - formats, labels, attributes and postfix dereference;
  - heredocs (bodies are trivia);
  - version tokens toke.c forces (5 alternatives);
  - `NOAMP` calls and method-call tokens;
  - scope-decided `*SUB` and `PRIVATEREF`, as superpositions;
  - phasers;
  - keyword plugins, excluded, with that reason in the ledger.
- **H5, the generator's exact checks:**
  - a literal is lexed in a short context where its class stands (for example `$a OP $b`), or through the lexer's checkpoint and resume from an operator-position state (A20);
  - every terminal a rule uses has a token class, with `PLUGIN_*` excepted (D18);
  - citations fall within the reference list;
  - the CORE.tpm correspondence: niladic ↔ `FUNC0`, unary ↔ `UNIOP`/`FUNC1`, list ↔ `LSTOP` (A21);
  - unused precedence tokens include `%prec`-only uses (B9).

## Round 3

**Revision assessed:** 460adf6f (460adf6f3b23abe7b2353df1e757a1c43ad1a8c2). All seven participants name this revision as the one they read. Code read at `origin/pu` 4fffe4d6; perl at 94e5086608.
**Participants:** A–F resumed (F is no longer a cold reader). **G** joins as this round's cold reader: it read the subject without the prior findings, did not open this minute or any other report, and could not see H1–H5.
**Round 3 dispatched:** 2026-10-30.

The status column of earlier rows above now carries each raiser's round-3 word where the raiser gave one. New findings are below. Section references are to the subject at 460adf6f.

### Recommendations, as declared

- **A: modify.** Blocks on A22 and A23. "Both have one-line resolutions."
- **B: modify.** "The rulings stand, including option 2. I record no disagreement. Test 1 supports option 2 for waypoint 1."
- **C: modify.** "Everything I blocked on in round 2 is fixed. Two blocks remain, both introduced by this revision's fixes, and each is a one-line change."
- **D: modify.** "All three blocks need about one sentence each, and none challenges a ruling."
- **E: accept.** "Every block I raised is now released." "None of the four stand-asides changes the build."
- **F: modify.** "I hold no block. Two new findings stand aside … Each is a sentence to fix."
- **G: modify.** "Block on G1 and G2: both are small edits to §5.2 and §9.5. G4 blocks only if the hand-off doesn't already cover it."

### Blocks open after round 3, as declared

| participant | blocks |
|---|---|
| A | A22, A23 |
| B | B11 (held from round 2) |
| C | C16, C17 |
| D | D19 ("editorial, but it must happen before acceptance freezes the text"), D20, D21 |
| E | none |
| F | none |
| G | G1, G2. Conditional: G4, "Stand aside if the hand-off (H4) already says this; I couldn't check. Otherwise block." |

Ten blocks are open: A22, A23, B11, C16, C17, D19, D20, D21, G1, G2. G4's condition is G's to settle. For the record, H4 above reads "heredocs (bodies are trivia)"; G4 also asks that neighbour classes skip `HeredocBody` and names `DataSection`.

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A24 | None stated beyond the row. |
| B14 | "Since the fix is mechanical and the ledger already requires a reason per alternative. It should be named so nobody ports these 'as written'." |
| C18 | "Low severity: a kill is a stop reported to perigrin, not automatic." |
| D22, D23 | None stated beyond each row. D23's bracket: "the superset gate is unaffected. This is a statement defect, not a gate defect." |
| D24 | "C's lens owns the numbers." |
| E13 | None stated beyond the row; bracket: "The class may be empty in practice." |
| E14, E16 | None stated beyond each row. E: "None of the four stand-asides changes the build." |
| E15 | "This is implementation detail, so I would accept it as a hand-off." |
| F7 (re-scoped) | "A protocol sentence … or an H3 item if the author calls it mechanics. I would release it on either." |
| F10, F11 | "Each is a sentence to fix." |
| G3 | "Risk to budget, not to correctness." |
| G5–G10 | None stated beyond each row. G: G4, G6, G7 and G8 "may already be answered" in the hand-off it could not read. |

### Releases the participant said were against its own bar

- **C1, C4, C9, C12, C13, C15:** "I release C1, C4, C9, C12, C13 and C15 against my own bar, and say so below."
- **D17:** "Against my bar: this is disclosed rather than closed, which is what I asked for."
- **E9:** "released, partly against my own bar". "§2 still glosses the ruling as 'time to an answer, M2's latency'. Read together with §13, that is now consistent. Read alone, it is looser than §13."
- **F1:** "I am releasing F1 against my own bar. The ruling removed the mechanism F1 attacked, and the probe below says the cost F1(d) warned about is small at the chain lengths the minute records. I am the participant most inclined to find that convenient, so the numbers are given in full."

A and B mark no single release as against their bar. Residues noted with a release but not held: E10 (§2's positive path still lacks the convergence condition; the 20-edit Guacamole list's fate unstated); E12 (residue raised as E15); D15 (implicit precedence on 3 fitted alternatives "matters little" now); A20 and D14 (H5's checkpoint-and-resume alternative is infeasible; refinement should use the `$a OP $b` form).

### Earlier rows not addressed in round 3

- **C5 (remainder, held at d6cb49c7):** C's round-3 report does not mention C5. Its status stands as recorded at d6cb49c7.

### Disagreements with a ruling's consequences, as recorded

- **E:** "under option 2, waypoint 1 cannot fail on precedence and its speed figure is a lower bound. So the spike's decisive evidence now arrives in waypoint 2, after 20 more days. That is the ruling's trade-off, not a finding."
- **C:** no disagreement; notes that "a waypoint-1 pass therefore says even less about §2's question than before. That is stated … so it is not a finding."
- **F:** "My disagreement with the ruling's consequences is recorded as none."
- **B, D:** no disagreement. A states none. G was not shown the rulings.

### Measurements and views that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **What decides a lone word in `{}`.** B11 (held): a neighbour's class should include its `OpensBlock` bit, since `{ print }` and `$h{print}` differ only in OB. A23, D20 and F10: OB is false on both `$h{` and `+{`, so it does not separate a subscript from an anonymous hash; the deciding fact is the token *before* the `{` (F10 cites toke.c `yyl_leftcurly`, `PL_expect == XOPERATOR`). G5: put OB in the neighbour class, and leave subscript against anonymous hash to a fitted subscript-key rule.
- **The error-token condition on the known list.** A24 and C17: the three files on pu's known list carry 0 Error/UnknownRest tokens (C17: 0 of the 1,509-file population too), so the condition makes real lexer defects unlistable. F (hunt note, no finding): "Feasible"; the three chapter-6 lexer defects it re-lexed are fixed on pu, so it found no live case. G: the three known-list files cannot be listed under v2's rule, "passable but worth knowing". A24 adds that under v2's key the `$#{…}` files map, "so the list empties on its own".
- **`lex-overhead` noise.** A22: alternating A/A medians across runs ±7–17% (k=1: 0.83–1.06). C16: ten alternating runs of a stand-in, 1.073–1.253, with 7 of 9 "rising" against the first; A/A alternating 0.95–1.05, sequential 0.89–1.20. D24 did not re-measure.
- **Sentences perly.y removes beyond `%nonassoc`.** E13: "I did not find a single sentence that perly.y's resolution removes beyond `%nonassoc`." B14, D21 and G2: seven alternatives refuse in their *actions* (perly.y 1467, 1469, 1485, 1487, 1542, 1547, 1554). B: the refusal of `isa`, `<=>` and `cmp` chains "is done by these error actions plus a shift resolution, not by `%nonassoc`". D21: `NCRELOP` and `NCEQOP` are `%left` (perly.y:166-167).
- **Cost of ambiguous recognition.** A (toy on the carried recognizer): items per token grow linearly with chain length; "with the assessment's chain data … the aggregate cost is small … No finding." F (plain Earley): at the recorded lengths the flat grammar costs 0.66–1.4× the ladder's items, and at k=20 less completion work; "chain ambiguity is not a plausible cause of a 3× kill". G3: a flat grammar "may be *faster*", so waypoint 1's ratio is "not predictive". E16 (lexer proxy): prefix-operator sites, Σk²/Σk 2.0–3.9, "the same order as the author's chains". Round-2 rows E12 (toy: flat 751.9 items/token cold) and F1 (d) are earlier figures on different probes.
- **T1's 236.** A and C reproduce it under the full-tree walk. G: "Not confirmed. My walk over all 986 was still running."
- **The level cross-check.** E14, F11 (i) and G9 each find §12's "level cross-check" undefined at 460adf6f.

### Round-3 ledger

New findings. Columns as in the earlier ledgers.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A22 | A | 3 | §9.3; §9.5 (from A18's fix) | `lex-overhead` is a wall-clock ratio under "may not rise" in every recognizer issue's criteria; the ratchet compares strictly, so unchanged code fails about half the time, and honest work raises it by design. See also C16, D24; A9. | `compare.go:88` strict `>` with no tolerance; `TestAA` (alternating, 9 pairs after 2 warm-up, load 2.9) medians k=1 0.83/1.03/1.02/1.06, k=4 0.99/0.99/0.96/1.07, k=16 0.94/0.93/1.02/0.99. | block | released at 8e3aa4b5 | 460adf6f |
| A23 | A | 3 | §7 | Under §7's neighbour class (kind, plus text), a hash subscript and an anonymous hash give the word identical keys; `$h{foo}` stays {string, call}; scope cannot settle it (perl autoquotes even when `foo` is declared), so waypoint 2 grades WRONG on bareword hash keys. See also B11, D20, F10, G5. | `perl -MO=Deparse` with `sub foo{1}`: `$h{foo}`→`$h{'foo'}`, `$r->{foo}`→`$r->{'foo'}`, `+{ foo }`→`{foo()}`; `TestNeighbours`: previous `{` (ob=false), next `}` identical in both; the deciding fact is the token before `{`. | block | released at 8e3aa4b5 | 460adf6f |
| A24 | A | 3 | §9.2 (`lex-unmapped`) | The error-token listing condition would exclude every entry on today's known list; a lexer defect without an error token cannot be listed, counts in `rd-accepted-chart-refused`, and the exit waits on a v0.1 lexer fix §11 puts outside the milestone. See also C17. | `TestKnown`: 0 Error/UnknownRest tokens in each of the three known-list files (01a1882d-d8a4); `TestDsharp`: `$#{$foo}`, `$#{ $foo->{'bar'} }` each one `Variable`. | stand aside | released at 8e3aa4b5 | 460adf6f |
| B14 | B | 3 | §5.2 ("bison actions … are not ported") | perly.y's seven unconditional-error alternatives, ported without actions, accept; no gate catches it, and rule 5 breaks permanently ("An editor gets a 'parse' of `$a < $b <;`"). See also D21, G2; E13, G1. | Lines 1467, 1469, 1485, 1487, 1542, 1547, 1554; `r2/earley.py nofilter`: `my $f = sub;`, `$a < $b isa C <;`, `$a isa B <;`, `$a <=> $b ==;` all ACCEPT; perl 5.42 refuses each. | stand aside | released at 8e3aa4b5 | 460adf6f |
| C16 | C | 3 | §9.5 (`lex-overhead` "may not rise"); §9.3 | "May not rise" on a wall-clock ratio fails code that has not changed; most intermediate issues fail the ratchet; alternation is not clearly required for chart/lex over `Tokenize`. See also A22, D24. | `lexov-r3.txt` alt-thin/tok: 1.098, 1.184, 1.210, 1.186, 1.080, 1.147, 1.136, 1.125, 1.253, 1.073 (7 of 9 rise against the first); A/A alternating 0.95–1.05, sequential 0.89–1.20; compare.go:89. | block | released, against C's own bar; residue (state that `lex-overhead` alternates) held, stand aside at 8e3aa4b5; residue released, against C's own bar at b85dc97d | 460adf6f |
| C17 | C | 3 | §9.2 (known list); §9.1; §11 | The error-token condition makes real lexer defects unlistable, so the exit can depend on another milestone (postmortem change 6's shape). See also A24; C14. | `zzerr`: `withErrorOrUnknownRest=0` over the three known-list files and over the 1,509-file population. | block | released, against C's own bar at 8e3aa4b5 | 460adf6f |
| C18 | C | 3 | §9.3 (checkpoint 1) | The early aggregate has no minimum fraction, and whether the 10% exclusion is computed over the recognized part or the whole population is unstated; over a small part ordinary files cross 10%. | §9.3; `top-r3.txt`: after the outlier, shares are flat at ~1.5%, so over a 5% part any file is ~30%. | stand aside | released, against C's own bar at 8e3aa4b5 | 460adf6f |
| D19 | D | 3 | §8 ¶1–2; §5.2; §12 | The contract is stated more than once and the statements disagree: "removing each derivation the declared levels forbid" refuses `$a + defined $b`, "safe … never a sentence the grammar accepts" keeps it; non-operator `%prec` is "for §8" (§5.2), answered by fitted rules (§8), and tested as part of the constraint (§12). See also F11 (ii); E12. | The spec text; `perl -c -e '$a + defined $b;'` OK; `%prec` histogram 21 `PERLY_PAREN_OPEN`, 12 `PREC_LOW`. | block ("editorial, but it must happen before acceptance freezes the text") | released at 8e3aa4b5 | 460adf6f |
| D20 | D | 3 | §7 | The key cannot make a subscript word a string as §7 says; waypoint 2 gets WRONG sites on accessors named like their hash key (`sub name { $_[0]{name} }`). See also A23, B11, F10, G5. | `go test -overlay D/overlay4.json -run TestDProbe3 -v ./internal/lexer`: `$h{foo}`, `+{ foo }`, `$h->{foo}` give the same previous/next tokens for `foo`; `perl -MO=Deparse,-p -e 'sub foo {1} $h{foo}; $r = +{ foo };'` → `$h{'foo'}`, `{foo()}`. | block | held, re-scoped (the stated rule); block at 8e3aa4b5; released at b85dc97d | 460adf6f |
| D21 | D | 3 | §8; §5.2; §3 rule 5 | The chart accepts code perl and RD both refuse, and no gate sees it: action-encoded refusals (`yyerror; YYERROR`) are dropped, and `NCRELOP`/`NCEQOP` are `%left`, so `($a <=> $b) <=> $c` survives safe semantics. See also B14, G2; E13, G1. | `grep -n 'YYERROR' perly.y` (1467, 1469, 1485, 1487, 1542, 1547, 1554); perly.y:166-167; `perl -c` refuses `$a <=> $b <=> $c`, `$a isa B isa C`, `$x = sub foo {1}`; RD refuses all three (`TestDProbe5`). | block | held; block at 8e3aa4b5; released, against D's bar; remainder re-scoped as D26 at b85dc97d | 460adf6f |
| D22 | D | 3 | §5.2 refusals; H5 | The no-token-class refusal fires on `UMINUS` and `PREC_LOW` (`%prec`-only) and, if start forms are terminals, on the seven `GRAM*`; the two refusals read "used" differently. | `grep -n UMINUS perly.y` → 175, 1497, 1499; §5.2's exemption list names only `PLUGIN_*`. | stand aside | held; stand aside at 8e3aa4b5; released, against D's bar at b85dc97d | 460adf6f |
| D23 | D | 3 | §2 parity limits; H4 | A third parity limit (keyword plugins excluded) is stated only in a hand-off list; it has no ruling, unlike library operators. | `D/probe4.out`: three `Future::AsyncAwait` snippets give `refused-anywhere=true` under `parse.Parse`; `grep -rn ':keyword(' --include=*.tpm`. | stand aside | released at 8e3aa4b5 | 460adf6f |
| D24 | D | 3 | §9.5; §9.3 | A wall-clock ratio gated at "may not rise" on every recognizer issue fails on noise. See also A22, C16. | §9.5 table; C5 and A9 (minute, round 1). Not re-measured. | stand aside | released at 8e3aa4b5 | 460adf6f |
| E13 | E | 3 | §8; §3 rule 5 | §8 knowingly accepts sentences bison removes beyond `%nonassoc`, against rule 5; chapter 7 has no marker sites in a file perl refuses, so WRONG 0 never sees them. See also G1; B14, D21, G2. | §3 and §8 text; chapter 7 §7.1.5; minute, round-2 rulings ("Recursive descent follows bison"). | stand aside | released; concrete case raised as E17 at 8e3aa4b5 | 460adf6f |
| E14 | E | 3 | §12; §8 | §12 lists "the level cross-check", whose definition left with the old §8; an implementer cannot tell what it compares, when, or whether it gates. See also F11 (i), G9. | `grep -n -i 'cross-check\|LevelsAccounted\|chapter 4'` → lines 109, 175–176, 280, 377–378, none defining a level check. | stand aside | released at 8e3aa4b5 | 460adf6f |
| E15 | E | 3 | §8 | One bit per forest node cannot decide precedence: a shared subtree is valid in one use and invalid in another, to unbounded depth; refined from the word "Boolean", the first implementation misses these cases. | `E/r2/depth1.py` re-run (2 readings; n+1 readings for n = 1..8); `perl -MO=Deparse,-p -e '$x = $a * defined $b + 1'` → `($x = ($a * defined(($b + 1))))`. | stand aside | released at 8e3aa4b5 | 460adf6f |
| E16 | E | 3 | §2; §8; §9.2 | What waypoint 1 decides has shifted: token fit, an ambiguous recognizer's speed as a lower bound, and an easier superset; a waypoint-1 pass will be over-read. See also G3. | `TestScratchPrefixAmbiguity`: prefix sites with k ≥ 1, std5k 168 (Σk²/Σk 3.9), T1 5,865 (2.0), perl5 `lib` 116 (2.3), dual-life 8,791 (2.9, max 31). | stand aside | released at 8e3aa4b5 | 460adf6f |
| F10 | F | 3 | §7 | The previous-token class cannot make a subscript word a string; toke.c decides at the `{` from the token before it (two back from the word), and `{-foo}` is also covered there; `$h{time}` is a WRONG reading wherever `time` resolves as a builtin. See also A23, B11, D20, G5; F3. | `F/pu/cmd/brace`: `$h{foo}` and `+{foo}` give identical `{` flags; `{foo};` differs only by OB; toke.c `yyl_leftcurly` (6650–6676). | stand aside | held, re-scoped (the stated rule); stand aside at 8e3aa4b5; released, against F's own bar at b85dc97d | 460adf6f |
| F11 | F | 3 | §12; §5.2; §8 | Two orphans from the §8 rewrite: (i) §12's "level cross-check" has no definition; (ii) whether the constraint reads non-operator `%prec` is stated both ways. See also E14, G9 (i); D19 (ii). | `grep -n "level cross-check"` → line 377 only; §5.2 line 144, §8 lines 236–238, §12 lines 383–385. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G1 | G | 3 | §3 rule 5; §9.2; §9.4; §9.5 | Nothing in either waypoint gates on rule 5's refusal half; an over-accepting fitted rule, a dropped `%nonassoc` or a ported error production passes every gate. See also E13; B14, D21, G2. | §9.5: `chart-accepted-perl-refused` is "reported"; first design's exit required "every `parsent` case refused" (~l.400); 46 perl-refused `parsent` files exist; perl refuses `1..2..3`, `++$x++`, `(1 <=> 2 <=>)`. | block | released, against G's own bar at 8e3aa4b5 | 460adf6f |
| G2 | G | 3 | §5.2; §8 | Alternatives whose action is the refusal, ported as written, turn syntax errors into accepted sentences; the doc never names the category. See also B14, D21; G1. | perly.y:1466–1469, 1484–1487, 1542, 1547, 1554 (unconditional); 669, 1107/1111, 1193 (conditional); perl 5.42 refuses `(1 <=> 2 <=>);` and `my $f = sub foo {1};`. | block ("cheap fix") | released; defect raised as G13 at 8e3aa4b5 | 460adf6f |
| G3 | G | 3 | §2; §8; §9.3 | Waypoint 1's speed figure does not measure precedence work; a flat grammar may be faster than any that does precedence in recognition; a pass is "necessary but not predictive", and v1's 1.5–1.9× was on a stratified grammar. See also E16. | §8 "Waypoint 1 does no precedence"; §9.3 checkpoint 3 first includes constraints; bison: 1,347 conflicts. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G4 | G | 3 | §7; §6 | `HeredocBody` is not trivia in pvm's lexer and sits inside expressions; without a stated rule every file with a heredoc argument in a list refuses or needs an undocumented special case, and neighbour classes change. See also H4 ("heredocs (bodies are trivia)"); D12. | `lexer.IsTrivia(HeredocBody)` false (`trivia_test.go`); `foo(<<A,\nbody\nA\n $x);` → `… HeredocOpen Operator[,] HeredocBody Variable[$x] …`; RD special-cases it at `parse.go:1090`, `expr.go:613`, `incremental.go:376`. | "Stand aside if the hand-off (H4) already says this; I couldn't check. Otherwise block." | released (G: 'the condition is met') at 8e3aa4b5 | 460adf6f |
| G5 | G | 3 | §7 neighbour class | The neighbour class drops the flags the subscript-string case needs; `$h{foo}`, `+{ foo }` and `map { foo } @x` key `foo` identically, and the superposition reaches resolve, which collapses by scope. See also A23, B11, D20, F10. | gprobe: `$h{foo}` no flag; `map { foo }` → `Operator[{]+B Word CloseBracket+B`; `+{ foo }` no flag. | stand aside | released; wording defect raised as G14 at 8e3aa4b5 | 460adf6f |
| G6 | G | 3 | §5.2 ("a set, never a sequence") | Only one direction of the mismatch is named; pvm also emits several tokens where toke.c emits one (`sub f :prototype($)` → `Operator[:] Word Prototype` against `COLONATTR THING`). | gprobe output; perly.y `subattrlist`. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G7 | G | 3 | §5.1; §7 | "Carried over" understates the recognizer work: edges carry terminal names, `class()` keys with `fmt.Sprint`, and `Recognize` imports `chart/lex`. See also F6; H2 (Q1). | `recognize.go:117–170, 316–340` on origin/pu. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G8 | G | 3 | §5.2 refusal ("where the class stands") | It is not visible how the generator would check this; the lexer exports only statement-start states. See also A20, D14 (H5's checkpoint alternative). | `lexer` exports `Tokenize`, `TokenizeBarewords`, `IsTrivia` and token flags; no per-position expect state. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G9 | G | 3 | §12 | "The level cross-check" is used once and defined nowhere. See also E14, F11 (i). | grep "level": §12 is the only use. | stand aside | released at 8e3aa4b5 | 460adf6f |
| G10 | G | 3 | §9.2 builtins | The builtins gate passes once it runs, whatever it finds. "Noted only." | §9.2 wording. | stand aside | released at 8e3aa4b5 | 460adf6f |

### Round-3 clusters

Counted by the first section each finding names, as in earlier rounds. A rule (§3 rule N) counts under §3.

**Round-3 new findings (29, at 460adf6f)**

| section | count | findings |
|---|---|---|
| §2 | 3 | D23, E16, G3 |
| §3 | 1 | G1 |
| §5 | 6 | B14, D22, G2, G6, G7, G8 |
| §7 | 5 | A23, D20, F10, G4, G5 |
| §8 | 4 | D19, D21, E13, E15 |
| §9 | 7 | A22, A24, C16, C17, C18, D24, G10 |
| §12 | 3 | E14, F11, G9 |

Within §5: §5.1 1 (G7), §5.2 5 (B14, D22, G2, G6, G8).
Within §9: §9.2 3 (A24, C17, G10), §9.3 2 (A22, C18), §9.5 2 (C16, D24).

By raiser: A 3, B 1, C 3, D 6, E 4, F 2, G 10.

**Rows open at 460adf6f (32: the 29 round-3 findings plus B11, C5's remainder and F7)**

| section | count | findings |
|---|---|---|
| §2 | 3 | D23, E16, G3 |
| §3 | 1 | G1 |
| §5 | 6 | B14, D22, G2, G6, G7, G8 |
| §7 | 7 | A23, B11, C5, D20, F10, G4, G5 |
| §8 | 4 | D19, D21, E13, E15 |
| §9 | 8 | A22, A24, C16, C17, C18, D24, F7, G10 |
| §12 | 3 | E14, F11, G9 |

**Open blocks by section (10):** §3 1 (G1), §5 1 (G2), §7 3 (A23, B11, D20), §8 2 (D19, D21), §9 3 (A22, C16, C17). G4 (§7) is conditional and not counted.

### Per participant, round 3

#### A (codebase and architecture)

- **Recommendation:** modify.
- **Released:** A9 ("into A22"), A14/A17, A18, A19 (ruling), A20, A21. "All other rows were released in round 2"; their proofs still reproduce on 4fffe4d6.
- **Held:** none.
- **New:** A22 (block), A23 (block), A24 (stand aside).
- **Checks reported:** §9.1's whole-tree acceptance gives root 0 and tree 236 over T1. The exit-versus-issue gates are sound for `rd-accepted-chart-refused`, `lex-unmapped`, `tables-drift` and `leo-disagree`, not for `lex-overhead`. The ledger's reference list can be built (348 over 114; five alternatives need an anchor line chosen). The next-token half of §7's key works; the previous-token half does not deliver the subscript claim. Recognition without precedence: on a toy, items per token grow linearly with chain length and time faster than quadratically; at the recorded chain data, "no finding".
- **Did not re-examine:** the safe-precedence constraint against bison, and non-operator `%prec` (B's lens); §13's numbers, perl's verdict counts, and whether `code_too_large.t` is the only file over 10%; H4's family list.

#### B (grammar fit)

- **Recommendation:** modify.
- **Released:** B10 (ruling), B12, B13. B1–B9 remain released; the counts re-run as 348, 114, 32, 39.
- **Held:** B11, block. "One clause in §7: a neighbour's class includes its OB flag."
- **New:** B14 (stand aside).
- **Tests reported:** (1) perly.y as written, unfiltered: Earley accepts all 21 targeted inputs bison accepts and all 5,606 random sentences; it also accepts `1..2..3` and `$a isa B isa C`, which bison refuses. (2) The previous-token class does not settle the 1,757 `$h{word}` sites. (3) H4 is complete "for token reshaping", checked against signatures and postfix deref; one non-reshaping family is missing (B14).
- **Did not re-examine:** §9.1–§9.5's new mechanics; the author's chain-length probe; §13, H1, H2; waypoint 2's constraint on the 21 `%prec PERLY_PAREN_OPEN` uses; signature `optcolon` lexing.

#### C (measurement and kill design)

- **Recommendation:** modify.
- **Released:** C1, C4, C9, C12, C13, C15, each against its own bar.
- **Held:** none declared. C14 is "re-scoped into C17 (block)". C5's remainder is not mentioned.
- **New:** C16 (block), C17 (block), C18 (stand aside).
- **Re-runs reported:** T1 any-node refused 236; `code_too_large.t` 48.6–51.7% of RD's time, next largest `operations.t` 1.5%; stand-in `lex-overhead` 1.07–1.25 alternating, under 1.3 in 10 of 10.
- **Did not re-examine:** §8's forest constraint and §7's previous-token key; chain lengths in practice or ambiguity's cost; the chart side of any timing; H4 and H5 beyond the `lex-unmapped` and ledger points.

#### D (spec quality)

- **Recommendation:** modify.
- **Released:** D14, D15 (superseded by the ruling), D16, D17 (against its bar), D18. D1–D13 were released in round 2.
- **Held:** none.
- **New:** D19, D20, D21 (block); D22, D23, D24 (stand aside). Three discards.
- **On the questions asked:** the contract is "stated in four places … Two of them conflict" (D19). H5 is implementation; H4 is sizing except "keyword plugins, excluded", a parity limit (D23). The growth (399 to 431 lines) is "mostly decision", with about 10 lines of creep in §13 and §5.1, "not raising it as a finding".
- **Did not re-examine:** whether a safe Boolean constraint can be evaluated over a shared packed forest ("needs a finite attribute per node … waypoint 2's design"); the chain-length probe and `items-per-token`; blead 5.45's verdicts on D21's cases; the 2016 Afroozeh paper; §9.3's aggregate rule and checkpoint-1 arithmetic; D20's frequency in the corpora.

#### E (architecture and direction)

- **Recommendation:** accept.
- **Released:** E5, E9 (partly against its own bar), E10 (small residue), E11, E12 (superseded by the ruling; residue raised as E15).
- **Held:** none.
- **New:** E13, E14, E15, E16 (stand aside).
- **On the coordinator's checks:** option 2 is consistent with rules 2, 6 and 12; rule 5 is "the exception" (E13). Consistent with chapter 7, with the gap that perl-refused files are never graded. Consistent with chapter 4 in substance, but the bridge is undefined (E14). Waypoint 1 is still worth its budget, "for a different reason than §2 gives" (E16).
- **Did not re-examine:** §7's previous-token key; the error-token rule; §9.1's walk; the 10% exclusion and `lex-overhead`; the ledger's reference list; H4 and H5; either Afroozeh paper; E13's size; any perly.y-derived measurement.

#### F (resumed; cold in round 2)

- **Recommendation:** modify. "I hold no block."
- **Released:** F1 (against its own bar), F2, F3 (re-scoped into F10), F4, F5, F6, F8, F9.
- **Held:** F7, re-scoped (GC between alternating runs), stand aside.
- **New:** F10, F11 (stand aside).
- **Probe reported:** flat against ladder at k = 2 and k = 20 (table in F's report); at the recorded chain lengths the flat grammar costs 0.66–1.4× the ladder's items; 5.6× at 22 operators. Bracket: plain Earley, not pu's split LR(0)ε recognizer; no `term` alternatives beyond `a`; chain lengths are the author's; `set-reuse` not modelled.
- **Hunt notes (no finding):** the error-token rule is feasible on today's lexer; `OpensBlock` is set correctly on every brace form tried; safe precedence leaves waypoint 1's gates unaffected; the 2^10 count is taken after the constraint.
- **Did not re-examine:** whether safe semantics reaches through perly.y's unit rules; the probe on pu's recognizer or perly.y, and `set-reuse`; the author's chain probe; §9.3's exclusion and checkpoint 1; whether `TermFollows` is false on `$x` in `print $x [1]` ("inferred it from the comment"); H4's 17 and 5 counts.

#### G (cold reader)

- **Recommendation:** modify.
- **Released / held:** none; G had no prior findings.
- **New:** G1, G2 (block); G4 (conditional, see above); G3, G5–G10 (stand aside).
- **Claims measured:** 348/114, 32/39, seven `GRAM*` and eleven `PLUGIN_*` confirmed; 78 of 124 `parsent` confirmed; with precedence stripped, bison reports 1,347 S/R and 0 R/R in 57 states, five of them not on operator tokens (four on `PERLY_COMMA`, one on `PERLY_BRACKET_OPEN`); the named exports exist; T1's 236 not confirmed (run unfinished).
- **Classification, as given:** codebase Partial (G7, G4, G5); architecture Partial ("Rule 5 lost its only gate", G1, G2); direction Ready. Per requirement: port ledger Ready; token interface Partial (G4, G5, G6); generator refusals Partial (G8); superset gate Ready; rule 5 enforcement Missing (G1); error productions Blocking (G2); speed protocol Partial (G3); isolation Ready.
- **On the document's size:** "About right for a decision. It defers detail to a hand-off I wasn't allowed to read … If so, the decision should cite which H-item answers each."
- **Did not check:** the minute and H1–H5; T1's 236; §13's `Expected` figures; whether safe precedence over a shared packed forest is a plain Boolean semiring ("finite but not Boolean … waypoint 2's problem"); chain lengths; any timing.

### Round outcome

**Another round owed.** Ten blocks are open: A22, A23, B11, C16, C17, D19, D20, D21, G1, G2. G4 is a conditional block whose condition G has not settled. No participant has released any of them. 32 rows are open in all. E recommends accept; A, B, C, D, F and G recommend modify.

## Author's response to round 3 (revision after 460adf6f)

The author is the orchestrating agent. Release is each raiser's to give. No ruling was needed: every block was answered under rules already in force.

The subject grew from about 3,640 to about 4,030 words. The growth is the requirements this round found missing: a refusal gate for rule 5 (G1, E13), alternatives refused in perly.y's actions (B14, D21, G2), the safe contract stated precisely (D19, E15), and waypoint 1's speed stated as a lower bound (G3, E16). §13's future-work detail was cut, as D judged it creep.

| id | response | where |
|---|---|---|
| A23, B11, D20, F10, G5 | The key reads the **two** previous significant tokens. The token before the `{` decides `$h{foo}` (string) against `+{ foo }` (call); `OpensBlock` is false on both braces, as A23/D20/F10 measured, so a flag alone could not. The flags also enter neighbour classes. | §7 |
| A22, C16, D24 | Timing keys are never compared strictly between issues. `lex-overhead` is reported per issue and gated only at the exit, as the median of three measurements. | §9.3, §9.5 |
| C17, A24, C14 | The known-list rule is now: an entry names an open lexer issue and the byte where the chart refuses, and the gate review confirms the construct there is that issue's. The error-token condition is dropped. | §9.2 |
| D19, E15, F11 (precedence half) | §8 is the only statement of the contract. "Safe" is defined: between two derivations of one nonterminal over one span, the declared preference keeps one; a span's only derivation is never removed; `%nonassoc` chains are the sole refusal. The carried value is each derivation's exposed operator context at both edges, not one bit. `%prec` on non-operator rules is not part of the constraint; the oracle finds differences and fitted rules answer them. | §8 |
| D21, B14, G2 | An alternative whose action refuses (unconditional `yyerror`/`YYERROR`, e.g. the non-chaining comparisons at perly.y:1466-1487, an anonymous `sub` without a body) is not ported as accepting. The ledger marks it "refused in action", and a test checks the chart refuses it. | §5.2, §12 |
| G1, E13 | A refusal gate: `chart-accepted-both-refused 0` at the exit means the chart accepts no file that both perl refuses with a syntax error and RD refuses. Other perl-refused acceptances stay reported. | §9.2, §9.5 |
| G3, E16 | Waypoint 1's speed figure is a lower bound: it can kill but not pass; the verdict on speed is checkpoint 3. | §2 |
| E14, F11, G9 | The level cross-check is defined in waypoint 1's exit: the grammar's declared levels against chapter 4's table and `parse.LevelsAccountedFor()`, with differences listed. | §9.2 |
| C18 | At checkpoint 1 the kill reads std5k, and the aggregate only once at least half the population is recognized; the 10% exclusion is always over the whole population. | §9.3 |
| D22 | The refusal exemptions now cover `%prec`-only precedence tokens and the `GRAM*` start selectors. | §5.2 |
| D23 | Keyword plugins are the third parity limit, stated in §2. | §2 |
| G4 | Heredoc bodies are attached to their heredoc's opening token, and are not neighbours or significant tokens. | §7 |
| G6 | Where the lexer emits several tokens for one perly.y terminal, the rule is fitted to the tokens. | §5.2 |
| G7, F6 | §5.1 states the recognizer's input changes from names to set ids (H2). | §5.1 |
| G8, A20 | The literal check is by lexing inside a short context string for the class; the operator-position checkpoint option is dropped (A20: checkpoints exist only at statement start). | §5.2; H5 |

## Round 4

**Revision assessed:** 8e3aa4b5 (8e3aa4b52d05790971d88badec3d431ead39fdc5). All eight participants name this revision as the one they read. Code read at `origin/pu` 4fffe4d6; perl at 94e5086608; perl 5.42.0 and bison 3.8.2 on the box.
**Participants:** A–G resumed (G was round 3's cold reader). **H** joins as this round's cold reader: it read the subject without the prior findings and did not open this minute or any other report.
**Round 4 dispatched:** 2026-10-30.

The status column of earlier rows above now carries each raiser's round-4 word where the raiser gave one. New findings are below. Section references are to the subject at 8e3aa4b5.

### Recommendations, as declared

- **A: "Accept as far as the codebase lens goes."** "I hold no blocks. A25 and A26 are stand-asides with one-line resolutions, and each fails loudly or has partial cover … rather than silently."
- **B: accept,** "with B15 and B16 as stand-asides. Neither touches feasibility."
- **C: modify.** "My round-3 blocks are fixed. The new refusal gate is the right idea, and it is paired correctly with the superset gate." Blocks on C19, C20, C21: "Each fix is a sentence plus a committed label list."
- **D: modify.** "None of the open items challenges a ruling."
- **E: modify.** "Both remaining findings come from one false premise in §8: that bison removes sentences only at `%nonassoc`."
- **F: modify.** "One block, F12: the new refused-in-action test contradicts §8."
- **G: accept,** "with G11–G14 as stand-asides … I hold no block."
- **H: modify.** "One finding blocks (H1). The rest stand aside for refinement."

### Blocks open after round 4, as declared

| participant | blocks |
|---|---|
| A | none |
| B | none |
| C | C19, C20, C21 |
| D | D20 (held, re-scoped; "one phrase"), D21 (held; "one sentence in §8, one in §12") |
| E | E17 |
| F | F12 ("a criterion that cannot go green, written into the text acceptance freezes") |
| G | none |
| H | H1 |

Eight blocks are open: C19, C20, C21, D20, D21, E17, F12, H1.

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A25, A26 | "Each fails loudly or has partial cover (§12's test, and the superset gate) rather than silently." |
| B15, B16 | "Neither touches feasibility. Both are a sentence each in §7 and §5.2, or in H4." B16: "The superset gate would catch a dropped try/catch on those 3 T1 files, which is why I stand aside rather than block." |
| C16 residue | "Say `lex-overhead` alternates." None further. |
| D22 (held) | None stated beyond the row. |
| D25 | None stated beyond the row. |
| E18 | "Waypoint 2 is refined later and the fix is a clause, but it belongs in the contract, not in implementation." |
| F7 (held) | "I said I would release on either fix; neither landed. It does not block." |
| F10 (re-scoped) | None stated beyond the row; "The key's information is sufficient … Only the stated rule is wrong." |
| G11 | "One clause, and the gate fails loudly either way." |
| G12 | None stated; "Today this is latent." |
| G13 | "One word plus one sentence; a gate catches the wrong reading." |
| G14 | None stated; "waypoint 2's oracle is positioned to catch it." |
| G15 | "H3 owns the layout, so this is implementation." |
| H2–H8 | "The rest stand aside for refinement." |

### Releases the participant said were against its own bar

- **B11:** "released this round, against my own bar on glob slots (B15)". "The key does not decide the glob case in principle."
- **C16, C17, C18:** "I release C16, C17 and C18 against my own bar."
- **G1:** "Released, against my own bar." "Residual: the population is small (all of it is parsent), and its predicate is undefined (G11). The §8 text now contradicts the gate (G12). Neither reopens G1."

A, D, E and F mark no single release as against their bar. Residues noted with a release, not held: C16 (state that `lex-overhead` alternates; held as a stand-aside residue); G2 (wording catches the conditional try/catch refusal, raised as G13); G3 ("doesn't add the early probe of constraint cost I offered … I don't hold for it"); G5 (wording, raised as G14); E13 (concrete case raised as E17); D19 ("one consistency gap remains … I carry it there [D21] rather than holding D19 for it").

### Earlier rows not addressed in round 4

- **C5 (remainder, held at d6cb49c7):** not mentioned in C's round-3 or round-4 report. Its status stands.
- **F7:** F notes "The minute has no response row for F7 in round 3."

### Disagreements with a ruling's consequences, as recorded

No participant records a disagreement with a ruling this round. D: "None of the open items challenges a ruling." F (hunt note, no finding) records a divergence from the postmortem, not a ruling: with timing keys never compared strictly, `lex-overhead` is gated only at the exit, "where the postmortem's change 2 asked for it 'from the first recognizer issue'".

### Measurements and views that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **The refusal gate's population.**
  - C19: 49 files perl and RD both refuse: 22 with "syntax error" (all Guacamole), 12 other toke.c/perly.y compile refusals (Guacamole), 15 environmental (T1: 13 missing modules, 2 `-T`).
  - G11 (parsent only): RD refuses 29 of the 46 perl-refused; 10 have "syntax error" on perl's first line, 20 anywhere.
  - H3 (parsent only): perl says "syntax error" on 29 of 124, non-syntax-refused 17, accepts 78; RD refuses 32; 20 are both.
  - A25: the codebase's only classifier, `parseoracle.Classify`, is message-based ("syntax error" in stderr).
- **T1 and syntax errors.** G: of 987 top-level files, 913 compile, 74 fail, and only `README.md` says "syntax error", so "T1 contributes no Perl file". H: 912 OK, 74 other, no syntax errors among 986. C21: of the 74, 68 "Can't locate", 5 `-T`, 1 other. C19: 15 of the both-refused are T1, all environmental.
- **Glob-slot sites.** B15: about 96 (after `Operator«*»`). F10: "words such as `*NAME{CODE}` about 40". Both count `Operator«->»` 759, `Variable` 655, `CloseBracket«}»` 124, `CloseBracket«]»` 123.
- **Which comparison chains perl refuses.** C (not a row): `1 == 2 == 3;` and `$a lt $b lt $c;` are `syntax OK` on 5.42, "because equality and relational ops chain since 5.32". D21: RD accepts `$a == $b == $c`, as perl does. E17: `$a < $b < $c` accepted. E17, F12 and H1 list the refused ones (`<=>`, `cmp`, `isa`, and mixes with `==`).
- **`lex-overhead` noise.** A (re-run at load 0.2, k=16): medians 1.05/1.00/1.00/0.94, "stable enough for a single exit gate". C16 (load 1.2): alternating A/A 0.964–1.129, sequential A/A up to 1.406; stand-in 0.995–1.278 alternating.
- **The §8 mechanism on a toy.** E15 (released): §8's edge-context contract keeps exactly bison's reading on all 7,665 bison-accepted toy strings. E18: on bison-refused strings, the chart accepts about a quarter (109, 104, 107 per seed).

### Round-4 ledger

New findings. Columns as in the earlier ledgers.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A25 | A | 4 | §9.2 refusal gate; §5.2 refused-in-action | "Perl refuses with a syntax error" can only be read with the message-based classifier, and perl's refused-in-action alternatives do not say "syntax error", so the gate cannot see them; a ledger miss is uncovered. See also C19, G11. | `parseoracle.Classify` (`corpus.go:468-493`); `perl -c` + `grep -c "syntax error"`: `sub;` 0, `sub f(@a = 1) {}` 0, `try {1} catch {2}` 0; `my $x = 1 2;`, `sub {`, `(1;`, `1..2..3` each 1. | stand aside | released at b85dc97d | 8e3aa4b5 |
| A26 | A | 4 | §5.2 refused-in-action | The rule is unconditional; try/catch refuses only when `catch_paren` is empty, so as written it refuses every `try/catch` (loudly, through `rd-accepted-chart-refused`). See also B16, G13, F12 (c). | `perly.y:660-671` (`if(!$scalar) { yyerror(…); YYERROR; }`); `catch_paren` at 812-820; perl refuses `catch { 2 }`, accepts `catch ($e) { 2 }`; slurpy-default rules (1103-1110) use `yyerror` without `YYERROR`. | stand aside | released at b85dc97d | 8e3aa4b5 |
| B15 | B | 4 | §7 | Glob slots need a third previous token: `*STDOUT{foo}` (string) and `print STDOUT { foo }` (anonymous hash) share the key. Also: §7's prose ("a term before the `{`") is wrong for `->`, and `+{ foo }` is a call only if `foo` is declared. See also D20, F10, G14. | `lx/subkey2` grouping of the 1,757 sites by key (table in B's report); Deparse with `sub foo`: `+{foo()}`; T1 glob slots all use slot names. | stand aside | released, against B's own bar at b85dc97d | 8e3aa4b5 |
| B16 | B | 4 | §5.2 | "`yyerror` and `YYERROR`" over-selects 670 (dropping every try/catch) and under-selects 1105-1112 (`yyerror` without `YYERROR`), which then ports as accepting, a rule-5 over-acceptance no gate sees. See also A26, G13, F12 (c). | `grep -n YYERROR perly.y` → 8 lines; `r2/tc.txt`: `try {} catch {}` ACCEPT in bison (actions stripped) and Earley, refused by perl; `sub f (@x = 1) {}` → "A slurpy parameter may not have a default value"; 3 T1 files use `try {`. | stand aside | released at b85dc97d | 8e3aa4b5 |
| C19 | C | 4 | §9.2 (refusal gate) | "Perl refuses with a syntax error" is not a defined predicate; the implementer picks the population, which decides whether refused-in-action alternatives are ever corpus-gated (C1's defect in the refusal direction). See also A25, G11; H3. | `bothrej-msgs.txt`: 22 "syntax error" lines, 12 others (e.g. "Illegal declaration of anonymous subroutine" on `Statements-SubStatement-002/3/4.pl`), 15 environmental. | block | released, against C's own bar at b85dc97d | 8e3aa4b5 |
| C20 | C | 4 | §8; §9.2 | The waypoint-1 refusal gate makes §8's "waypoint 1's gates are unaffected" false: a `%nonassoc` chain is in the gated population and waypoint 1 accepts it; the exit depends on corpus contents; §8 calls the gate a report, §9.5 gates it at 0. See also F12 (b), G12. | `na/`: `1..2..3;` → perl "syntax error … near "2.."", RD refused=true; pvm lexes it `Number Operator(..) Number Operator(..) Number`; no `..` chain in the Guacamole files. | block | released, against C's own bar at b85dc97d | 8e3aa4b5 |
| C21 | C | 4 | §9.1 (T1's labels) | T1's perl labels are mostly about this machine's `@INC`; the superset population and its counts change with installed modules, so a "may not rise" count can move with no code change. See also H4. | `t1perl.txt`'s 74 rejects: 68 env-module, 5 env-taint, 1 other. | block | released, against C's own bar at b85dc97d | 8e3aa4b5 |
| D25 | D | 4 | §7; §5.2 | Two attachment rules for heredoc bodies: §5.2 attaches trivia to the next significant token, §7 attaches bodies to their opening token; one sentence is false whichever is built. See also G4. | The spec text; `lexer.go` comment: HeredocBody "begin[s] after the line the marker sits on". | stand aside | released at b85dc97d | 8e3aa4b5 |
| E17 | E | 4 | §5.2; §8; §12 | Refused-in-action comparisons are still derivable by other rules and are `%left`, so the chart accepts them at both waypoints; the §12 test cannot pass; §8's "changes acceptance only at `%nonassoc`" is false. See also D21, F12 (a), H1. | `E/r4/eqop.py`: `T NCEQOP T CHEQOP T`, `T NCEQOP T NCEQOP T`, `T NCRELOP T NCRELOP T` each 2 derivations; perl (`use v5.36`) refuses `$a <=> $b == $c`, `$a == $b <=> $c`, `$a <=> $b <=> $c`, `$a isa X isa Y`; RD refuses all four (`rd_scratch_test.go`); perly.y:166-167, 1438. | block | released, partly against E's own bar; residue carried into E18 at b85dc97d; E: 'its case comes back at this revision'; raised as E21 at 5fe8dc4b | 8e3aa4b5 |
| E18 | E | 4 | §8 | "Never remove the only derivation" rescues forbidden readings after a `%nonassoc` refusal: the chart accepts `1 .. 2 + 3 .. 4` as `(1 .. 2) + (3 .. 4)`; perl and RD refuse it. | `E/r4/safe.py` over `toy2.y`: of ~445 bison-refused strings per seed, the chart accepts 109/104/107; the variant `safe_v.py` accepts 0 and still agrees on every accepted string; perl refuses `1 .. 2 + 3 .. 4`, `1 .. 2 ** 3 .. 4`. | stand aside | held, re-scoped (shielded chains); block at b85dc97d; released (ruling) at 5fe8dc4b | 8e3aa4b5 |
| F12 | F | 4 | §5.2; §12; §8; §9.2 | (a) The refused-in-action test for non-chaining comparisons can never go green; (b) §8's "report" contradicts the gate at 0, and `1..2..3` breaks waypoint 1's gate (latent); (c) "an alternative whose action refuses" is not well defined (1107/1111, 669, and scope-conditional 579, 614, 1193). See also D21, E17, H1 (a); C20, G12 (b); A26, B16, G13 (c). | `F/probe/ncchain.py`: `t NCEQOP t NCEQOP t`, `t CHEQOP t NCEQOP t`, `t NCRELOP t NCRELOP t` all True; `perl -c` "syntax error" on `$a <=> $b <=> $c`, `$a == $b <=> $c`; bison rules 225, 244, 245, 268; `grep -n "yyerror\|YYERROR\|croak" perly.y`. | block | held, re-scoped (shielded chains; parts (a) and (c) released); block at b85dc97d; released at 5fe8dc4b | 8e3aa4b5 |
| G11 | G | 4 | §9.2 refusal gate | "Perl refuses with a syntax error" is not a defined predicate (10 or 20 parsent files by reading), and either reading excludes perly.y's own `yyerror` refusals with custom text. See also A25, C19. | `parsent-refused2.txt`: 10 against 20; `perl -c parsent/Statements-SubStatement-002.pl` → "Illegal declaration of anonymous subroutine". | stand aside | released at b85dc97d | 8e3aa4b5 |
| G12 | G | 4 | §8 (two sentences) | §8 contradicts the new gate: (a) "waypoint 1's gates are unaffected" while a `%nonassoc`-only both-refused file holds the gate above 0; (b) a bison-removed sentence "shows in the refusal gate's report" while the gate is 0 at exit. Latent today. See also C20, F12 (b). | §8 l.264–271 against §9.2/§9.5; gprobe of the 20 population files (none a `%nonassoc` case); RD refuses `1..2..3`. | stand aside | released; superseded by G16 at b85dc97d | 8e3aa4b5 |
| G13 | G | 4 | §5.2 "refused in action" | Read literally, the exception refuses every try/catch (no "unconditional" in the subject), and misses the slurpy-default `yyerror` without `YYERROR`. See also A26, B16, F12 (c). | perly.y:660–681, 812–820, 1103–1112; `rdcount`: both T1 try/catch files accepted by RD and perl. | stand aside | released at b85dc97d | 8e3aa4b5 |
| G14 | G | 4 | §7 example rule | "A term before the `{`" / "an operator before it" is wrong for `->{`, the most common subscript; silent in waypoint 1, caught only by waypoint 2's WRONG at a marked site. See also D20, F10, B15. | gprobe on `$h->{foo}`, `f()->{foo}`; perl deparses `$r->{foo}` as `$r->{'foo'}`. | stand aside | released at b85dc97d | 8e3aa4b5 |
| G15 | G | 4 | §7 "Compiled … constant-time" | With the token plus three neighbour classes carrying text and flags, a dense table (~10^10) is out of reach; the lookup must be sparse. Implementation; noted so H3 does not read "precomputed" as dense. See also H6. | Arithmetic on §7's key; H3. | stand aside | released at b85dc97d | 8e3aa4b5 |
| H1 | H | 4 | §5.2; §8; §12 | The refused-in-action test for non-chaining comparisons can never pass; perl's refusal comes from bison's precedence (`%prec PREC_LOW`, shift, `yyerror`); the chart accepts the family at both waypoints, and §8's "only at `%nonassoc`" is false. See also D21, E17, F12 (a). | bison report: "Conflict between rule 224 and token NCRELOP resolved as shift (PREC_LOW < NCRELOP)", same for rule 236; `%nonassoc` errors only for DOTDOT, POSTINC/POSTDEC/POSTJOIN and three PLUGIN classes; perl 5.42 refuses `$a <=> $b <=> $c`, `$x isa A isa B`, `$x isa A < 1`, `1 < $x isa A`, `$a cmp $b == 1`, `$a == $b cmp 1`. | block | released, against H's own bar; continues as H9 at b85dc97d | 8e3aa4b5 |
| H2 | H | 4 | §8 | "A `%nonassoc` chain is refused", applied literally, matches nested prefix operators of one level, which perl accepts; waypoint 2 would refuse valid Perl. | perl accepts `defined ref $x`, `print sort @x`, `require require Foo`, `last last`, `-e -f $x`; refuses `++$x++`, `1 .. 2 .. 3`. | stand aside | released at b85dc97d | 8e3aa4b5 |
| H3 | H | 4 | §9.2 refusal gate; §9.1 | At waypoint 1 the gate is graded on about 20 mostly lexical one-line snippets and cannot detect grammar over-acceptance; a green gate is weak evidence. See also C19, G11; G1. | parsent: perl "syntax error" 29, non-syntax 17, accepts 78; RD refuses 32; 20 both (`.e1`, `@@foo`, `$foo()`, `sort {} 1`); T1: no syntax errors among 986 (912 OK, 74 other). | stand aside | released at b85dc97d | 8e3aa4b5 |
| H4 | H | 4 | §9.1 labels | The labels depend on `@INC`; the superset population changes with the machine that labels it (here the T1 part is 691). See also C21. | `H/t1-perl.txt` (run from `unit/`, 10 s timeout). | stand aside | released at b85dc97d | 8e3aa4b5 |
| H5 | H | 4 | §9.5 | If the reset lands before the first grammar slice, every recognizer key is `na`, the issue that first prints them fails the ratchet, and re-baselining is forbidden. | `cmd/ratchet/compare.go:84` fails `na → number` with "baseline still na; take a baseline". | stand aside | held, narrowed; stand aside at b85dc97d; released at 5fe8dc4b | 8e3aa4b5 |
| H6 | H | 4 | §7 "Compiled" | A dense product table would have ~10^10 cells, so the lookup must be sparse or staged; no notation is given for a context-dependent token class, though rule 9 makes the grammar file the single source. See also G15. | Probe output; `perly.y:1536` (toke.c's inserted `;` before `}`). | stand aside | released at b85dc97d | 8e3aa4b5 |
| H7 | H | 4 | §5.2 start forms | Six of seven `GRAM*` start forms have no consumer in pvm; port and ledger work answering no question. | `perly.y:187-268`; nothing in `internal/parse` or `psc` parses those as entry points. | stand aside | released at b85dc97d | 8e3aa4b5 |
| H8 | H | 4 | §9.2 `leo-disagree` | "`Expected` at every node", but a recognizer has positions, not nodes. | Wording. | stand aside | released at b85dc97d | 8e3aa4b5 |

### Round-4 clusters

Counted by the first section each finding names, as in earlier rounds.

**Round-4 new findings (24, at 8e3aa4b5)**

| section | count | findings |
|---|---|---|
| §5 | 7 | A26, B16, E17, F12, G13, H1, H7 |
| §7 | 5 | B15, D25, G14, G15, H6 |
| §8 | 4 | C20, E18, G12, H2 |
| §9 | 8 | A25, C19, C21, G11, H3, H4, H5, H8 |

Within §5: all seven are §5.2.
Within §9: §9.1 2 (C21, H4), §9.2 5 (A25, C19, G11, H3, H8), §9.5 1 (H5).

By raiser: A 2, B 2, C 3, D 1, E 2, F 1, G 5, H 8.

**Rows open at 8e3aa4b5 (31: the 24 round-4 findings plus C5's remainder, C16's residue, D20, D21, D22, F7 and F10)**

| section | count | findings |
|---|---|---|
| §5 | 8 | A26, B16, D22, E17, F12, G13, H1, H7 |
| §7 | 8 | B15, C5, D20, D25, F10, G14, G15, H6 |
| §8 | 5 | C20, D21, E18, G12, H2 |
| §9 | 10 | A25, C16, C19, C21, F7, G11, H3, H4, H5, H8 |

**Open blocks by section (8):** §5 3 (E17, F12, H1), §7 1 (D20), §8 2 (C20, D21), §9 2 (C19, C21).

Two groups account for six of the eight blocks. D21, E17, F12 and H1 concern the refused-in-action comparisons and §8's "only at `%nonassoc`" (H1 is a cold reader's). C19 and C20 concern the new refusal gate. D20 and C21 stand alone.

### Per participant, round 4

#### A (codebase and architecture)

- **Recommendation:** "Accept as far as the codebase lens goes."
- **Released:** A22, A23, A24. All earlier rows stay released; their proofs re-ran unchanged (the 348-alternative parse not re-run).
- **Held:** none.
- **New:** A25, A26 (stand aside).
- **Other checks:** heredoc bodies pair with openers by FIFO, feasible; `return { foo }` has `OpensBlock` set and Deparse prints `return {foo()}`, "I did not establish which reading perl takes, so this is not a finding"; `chart-accepted-both-refused` passes trivially before work but is paired with `rd-accepted-chart-refused 0`, so it cannot pass vacuously at the exit.
- **Did not re-examine:** §8's contract, checkpoint 1's floor, the 10% exclusion and §2's lower bound; H4/H5 in full; the 348-alternative parse.

#### B (grammar fit)

- **Recommendation:** accept.
- **Released:** B11 (against its own bar on glob slots), B14. B1–B10, B12, B13 stay released.
- **Held:** none.
- **New:** B15, B16 (stand aside).
- **Re-runs:** `lx/subkey` 1,757 / 12; `r2/bisonp` and `earley.py` on the error-action inputs; `grep YYERROR` 8 lines.
- **Did not re-examine:** §9.2's new gate, the level cross-check and the known-list review; §9.3/§9.5 timing keys and the checkpoint-1 floor; §2's lower bound; §8 beyond its first paragraph; heredoc attachment against §5.2's trivia rule ("Those two may say different things, and I did not check them"); Deparse ran on 5.42, not blead.

#### C (measurement and kill design)

- **Recommendation:** modify.
- **Released:** C16, C17, C18, each against its own bar.
- **Held:** C16's residue (state that `lex-overhead` alternates), stand aside. C5's remainder is not mentioned.
- **New:** C19, C20, C21 (block).
- **On the new gate:** "Can it pass before its work?" Trivially 0 before the port, but meaningful because the exit pairs it with superset 0 and a complete port, "so there is no finding here". Two observations not raised as rows: the "non-chaining comparisons at `perly.y:1466-1487`" are narrower than the phrase suggests (equality and relational ops chain since 5.32); `sprintf {$fh} $foo;` needs a fitted `LSTOP` split H4 does not list.
- **Did not re-examine:** §7's key; §8's constraint; the level cross-check; whether perl-accepted T1 files would fail with modules installed; the chart side of any timing.

#### D (spec quality)

- **Recommendation:** modify.
- **Released:** D19, D23, D24.
- **Held:** D20 (re-scoped to the stated rule, block), D21 (block), D22 (stand aside; adds `PLUGEXPR`/`PLUGSTMT`).
- **New:** D25 (stand aside). Two discards.
- **On the questions asked:** §8's contract is stated "once: yes. Consistently: not yet" (D21). The ~10% growth is "Decision", with two small leftovers not raised: an implementation line in §5.1, and the timing rule stated in three places in §9.3/§9.5 ("One fact in three places is how D19 started").
- **Did not re-examine:** whether the safe contract reproduces perl on the B3 cases; T1 for `%nonassoc` or non-chaining chains; blead's verdicts; §9.3 arithmetic and §2's lower bound; other rows and H1–H5 beyond D22.

#### E (architecture and direction)

- **Recommendation:** modify.
- **Released:** E13 (concrete case raised as E17), E14, E15 ("after the §8 mechanism matched bison on all 7,665 accepted toy strings"), E16.
- **Held:** none.
- **New:** E17 (block), E18 (stand aside).
- **Did not re-examine:** §7's key; heredoc attachment; the checkpoint-1 floor; the timing-key rule; the known-list rule; H1–H5; other `%prec` uses beyond the comparison rules ("Others may remove sentences in the same way as E17"); any perly.y-derived measurement.

#### F

- **Recommendation:** modify.
- **Released:** F11 (both halves).
- **Held:** F7 (stand aside), F10 (re-scoped to the stated rule, stand aside).
- **New:** F12 (block).
- **Hunt notes (no finding):** §2's lower bound consistent with §10; checkpoint-1 floor cannot pass before its work; `GRAM*` exemptions correct; heredoc attachment consistent with the lexer; round-2 evidence re-runs unchanged (1,347 conflicts, litlex, 35 alternatives).
- **Did not re-examine:** §8's edge-context algorithm and unit rules; how "perl refuses with a syntax error" is classified in practice, and perl.git `t/`; other participants' rows; §9.3 arithmetic; H4/H5 beyond refused-in-action; `{-name}` keys in CPAN-style code.

#### G (resumed; cold in round 3)

- **Recommendation:** accept, "with G11–G14 as stand-asides".
- **Released:** G1 (against its own bar), G2 (defect raised as G13), G3, G4 ("the condition is met": H4 and §7), G5 (wording raised as G14), G6, G7, G8, G9, G10.
- **Held:** none.
- **New:** G11, G12, G13, G14, G15 (stand aside).
- **Measurements added:** the refusal gate's parsent population (10 or 20); T1 contributes no Perl file to it; the 20 population files fail on token-level or structural points, none a `%nonassoc` chain ("my judgment from the tokens"); two superset files use try/catch.
- **Did not re-examine:** other participants' rows; §8's feasibility over a shared forest ("I accept it as stated"); checkpoint-1 floor and timing keys; whether `perl -c` on T1 hides syntax errors behind failing `use` lines; any timing.

#### H (cold reader)

- **Recommendation:** modify.
- **Released / held:** none; H had no prior findings.
- **New:** H1 (block); H2–H8 (stand aside).
- **Claims measured:** 348/114/32 confirmed; `PERL_VERSION 45` confirmed; `OpensBlock` false on `$h{`, `+{` and `$h->{`, true on `map {`, `sub {` and statement `{`; `TermFollows` confirmed; RD refuses 236 of 986 confirmed; perl accepts 78 of 124 confirmed; chains short (std5k p99 4, max 6; T1 p99 2, max 8; an upper bound); the exports exist.
- **Classification, as given:** Blocking: refused-in-action alternatives (H1). Partial: precedence as a forest constraint (H1, H2), token interface (H6), refusal gate (H3), ratchet reset (H5). Ready: port and ledger, superset gate, speed protocol, `lex-overhead ≤ 1.3` ("Not measured"), isolation.
- **Axes, as given:** codebase "aligns"; architecture aligns "with the first design's rules 1–6 and with spec chapter 4's 32 levels. The exception is rule 5"; direction "aligns".
- **Did not check:** whether the lexer stream can be fitted to all 348 alternatives (sampled braces and `;` only); Leo on ambiguous right-recursive chains; `lex-overhead` and speed; whether the carried `lex.Lattice`/`lex.State` generics separate by a single deletion; perl.git `t/`; bison's other 1,347 resolutions beyond the error and `yyerror` paths.

### Round outcome

**Another round owed.** Eight blocks are open: C19, C20, C21, D20, D21, E17, F12, H1. No participant has released any of them. 31 rows are open in all. A ("as far as the codebase lens goes"), B and G recommend accept; C, D, E, F and H recommend modify.

## Author's response to round 4 (revision 51c59f61)

The author is the orchestrating agent. Release is each raiser's to give. No ruling was needed. The subject grew from about 4,030 to about 4,280 words, mostly in §8 as rewritten below.

| id | response | where |
|---|---|---|
| D21, E17, F12, H1, C20, E18, G12 | §8 now draws one line: **recognition decides what is Perl; the forest decides which reading.** Every precedence-driven refusal is a fitted rule enforced in recognition. That covers `%nonassoc` chains as bison fires them (infix after infix, prefix with postfix, but not prefix after prefix: H2) and the non-chaining comparisons bison refuses through `%prec PREC_LOW` on `%left` rules (`$a <=> $b <=> $c`, `$a <=> $b == $c`, `$x isa A < 1`). The forest only disambiguates and never removes a span's only derivation, so `1 .. 2 + 3 .. 4` is refused in recognition, not rescued in the forest. Waypoint 1's acceptance is therefore perl's, refusals included. Which exact forms perl refuses is settled by §12's tests against perl, not by the document. | §8, §12 |
| B16, A26, G13 | Only an *unconditional* action refusal is ported as a refusal. A conditional one (a `catch` without a variable, a slurpy default) becomes a fitted rule excluding that case. | §5.2 |
| C19, C21, A25, G11, H4 | Perl's labels are pinned. A committed label file records how it was produced (binary and version, cwd, `PERL5LIB`, timeout). Each refusal carries a reviewed category: **syntax** (any compile-time refusal of the construct, whatever the message, including perly.y's action refusals and toke.c's fatal lexing errors) or **environment** (a missing module, `-T`, a runtime failure). The refusal gate counts syntax refusals that RD also refuses. | §9.1, §9.2 |
| H3 | The refusal gate is stated as weak at waypoint 1, because the corpora hold few such files. §12's refusal tests, each paired with an accepted near neighbour, carry rule 5 until perl.git `t/`. | §9.2, §12 |
| D20, F10, B15, G14 | §7 no longer says "term before, operator before". The two previous tokens decide `$h{foo}`, `$h->{foo}` and `f()->{foo}` against `+{ foo }`; which pairs mean what is the grammar's to state and the table's to hold. The glob-slot case (`*STDOUT{foo}`, three tokens back) is left to the superposition and the oracle, and recorded as known. | §7 |
| G15, H6 | The table is sparse: the token's own class first, refined by a neighbour only where a rule needs it. Layout and notation are in H3. | §7; H3 |
| D25 | One attachment rule: trivia attaches to the next significant token, except heredoc bodies, which attach to their opening token. | §5.2, §7 |
| D22 | `PLUGEXPR`/`PLUGSTMT` are added to the exemptions (`%prec`-only tokens were already covered). | §5.2 |
| H5 | A key's baseline is taken by the first issue that makes it print a number. | §9.2 |
| H7 | The six non-program start forms stay ported, with limited testing as stated. perigrin approved porting all seven, and `fullstmt` is recursive descent's `Repair` unit for later work. | §5.2 |
| H8 | "Every position". | §9.2 |
| F7, C's stand-aside on alternation | A GC runs before each timed run, and `lex-overhead` is timed alternately. | §9.3 |
| H2 | Folded into §8's `%nonassoc` statement. | §8 |

## Round 5

**Revision assessed:** b85dc97d (b85dc97da41dd1030e72c13f56bf7819436aab15). All nine participants name this revision as the one they read. Code read at `origin/pu` 4fffe4d6; perl at 94e5086608; perl 5.42.0 and bison 3.8.2 on the box.
**Participants:** A–H resumed (H was round 4's cold reader). **I** joins as this round's cold reader: it read the subject without the prior findings and did not open this minute or any other report.
**Round 5 dispatched:** 2026-10-30.

The status column of earlier rows above now carries each raiser's round-5 word where the raiser gave one. New findings are below. Section references are to the subject at b85dc97d.

### Ruling during round 5 (perigrin, 2026-10-30)

- **The shielded-chain question** (a refused chain with a tighter operator between its two operators, e.g. `1 .. 2 + 3 .. 4`; raised in this round by B17, D26, E18, F12, G16 and H9): "Forest refuses (bison-exact)".

The ruling was given after the round's reports were written. No participant has answered it, and no finding is released by it until its raiser says so.

### Recommendations, as declared

- **A: "Accept on the codebase lens."** "I hold no blocks. A27 is a stand-aside with a one-line resolution."
- **B: modify,** "B17 block. Everything else in my lens is met."
- **C: accept.** "My blocks are all released. C22 is a latent definitional gap, and I stand aside on it."
- **D: modify.** "D26 challenges no ruling."
- **E: modify.** "E18 … needs perigrin's choice among options 1–3."
- **F: modify.** "One block remains: F12, re-scoped to the mechanism the revision chose."
- **G: modify. Block on G16.** "The fix is either a sentence of partial stratification with perigrin's sign-off, or a return to forest-side refusal with waypoint 1's over-acceptance stated."
- **H: modify.** "One finding blocks: H9, which is new."
- **I: modify.** "The design is coherent and its factual claims hold up when measured. Two changes block: I1 … and I3."

### Blocks open after round 5, as declared

| participant | blocks |
|---|---|
| A | none |
| B | B17 |
| C | none |
| D | D26 ("it decides what is built, and the stated tests cannot catch the wrong build") |
| E | E18 (held, re-scoped; now block) |
| F | F12 (held, re-scoped; "§8's claim that waypoint 1's acceptance is perl's is false as written, and rule 5 is a carried design rule") |
| G | G16 |
| H | H9 |
| I | I1, I3 |

Eight blocks are open: B17, D26, E18, F12, G16, H9, I1, I3. Six of them (B17, D26, E18, F12, G16, H9) concern shielded chains, the question perigrin's round-5 ruling addresses.

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A27 | "No corpus file trips it at waypoint 1; the category definition is the decision's." |
| C22 | "A latent definitional gap." "Low today." |
| D27 | None stated beyond the row; "How often this happens is unknown, and probably rare." |
| H5 (held, narrowed) | None stated beyond the row. |
| H10, H11 | None stated beyond each row. |
| I2 | "Waypoint 2 is refined only after waypoint 1's verdict, but the sentence is false as written." |
| I4, I5, I6 | "The other findings stand aside but should be fixed in the text." |

### Releases the participant said were against its own bar

- **B15:** "I release against my bar: the glob case is not decided by the key." Nit noted: §7 still calls `+{ foo }` "a call".
- **C5 (remainder), C16 (residue), C19, C20, C21:** "Every release below is against my own bar."
- **D21:** "Against my bar: I proposed putting the refusal in the forest constraint, and the author put it in recognition. That is equally valid, *if* the fitted rule can be written. D26 says what that requires."
- **D22:** "released, against my bar … the decision's own wording stays looser than I asked."
- **E17:** "Released, partly against my own bar." "The fix covers *adjacent* comparisons only … carried [into E18]."
- **F10:** "released, against my own bar". Residue, not held: `$h{-foo}` (deciding token three back) "is not in the known list beside the glob slot".
- **H1:** "I am biased toward my own bar, so I note that this release is against it." "The defect behind H1 continues … as H9."

A and G mark no single release as against their bar.

### Earlier rows not addressed in round 5

None. Every row open at 8e3aa4b5 was given a round-5 word by its raiser.

### Disagreements with a ruling's consequences, as recorded

These were written before the round-5 ruling.

- **E (E18):** "On this evidence, 'safe' semantics buys nothing for perl and costs rule 5. This is recorded as disagreement, not as a finding."
- **B (B17):** its option (b) "touches a ruling; I raise it as a consequence, not as disagreement."
- **G (G16):** "the choice may be perigrin's, since it bears on round 2's ruling."
- **H (H9):** option (b) "changes round 2's ruling ('never refuses'), so it is perigrin's call."
- **F (F12):** option (2) "breaks 'the forest never refuses'".
- **D:** "D26 challenges no ruling. Option 2's ruling covers ordering in the forest, while refusals in recognition are the author's design."
- A, C and I record none.

### Measurements and views that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **How much a local fitted refusal misses.**
  - B17 (`fitted.py`, head-only rules, toke.c token streams): 19 accepted controls kept; 8 shallow refusals refused; 8 deep refusals still accepted; 5,606 random sentences identical to bison.
  - E18 (toy, 3,000 strings × 3 seeds, of ~445 bison-refused): §8 as written (`split.py`) accepts 338/331/319; with a local stratification of the nonassoc level (`fitted.py`) 251/241/238; with the forest refusing a span whose every derivation is forbidden (`strict.py`) 0/0/0. All variants agree with bison on every accepted string. "The share … rose from about 24% to about 75%" against round 4.
  - D26 (`shield.py`), F12 (`fitted_refusal.py`): the shielded forms accepted by a top/operand rule; perl refuses them.
  - G16: "No chart was built … I didn't observe it on a running recognizer."
  - I1 (differential of perly.y under LALR against the as-written grammar under GLR, 1.08M + 3 × 7.9–9.8M strings of ≤ 5 tokens, plus 82k random derivations): both named families confirmed; two further differences found (`PLUGIN_LOW_OP` after a comma list; `LSTOP PERLY_DOLLAR BAREWORD [ ]`).
- **§9.2's baseline sentence against §9.5's `na` sentence.** H5 (held): "The two sentences contradict each other." I6: a "minor wording conflict", consistent with `compare.go` "once the issue commits the baseline". F and G (hunt notes, no finding): consistent with `compare.go:84–85`.
- **The forced GC before each timed run.** F7 released on it. H10: it moves the collection cost of allocation-heavy code out of the timed region, favouring the chart. G (hunt note, not raised): it moves out only the previous run's garbage; "a run that allocates past the heap target still collects inside its own timing".
- **When `lex-overhead` gates.** I3: gate it from the first issue, per postmortem change 2 and rule 10. A22, C16 and D24 were released in round 4 on its being reported per issue and gated at the exit; F's round-4 hunt note recorded the divergence from the postmortem.
- **The slurpy-default refusal.** A26 (released, note), G13 (released, note) and H11 each say `perly.y:1103-1112` refuses unconditionally, though §5.2 calls it conditional. All three say either treatment refuses it.
- **Op-check refusals in the syntax category.** A27 and C22 make the same point; A27 proposes a category "check", C22 "semantic".
- **RD on prefix-with-postfix.** C20: RD *accepts* `++$x++` and `++$x--`, so they are outside `chart-accepted-both-refused`. E18 and H list `++$x++` among perl's refusals.
- **T1 labels and working directory.** A (run from each file's directory): 68 environmental, 6 `-T`. C21 (round 4): 68 / 5 / 1. H4: 912 OK, 74 environment. A: "PerlOnJava's tests expect the repository root as cwd."
- **`code_too_large.t`'s size.** I: 495 KB of 3.13 MB. C (round 1): 484 KB of 2,176 KB.
- **The `isa` example.** C20: a syntax error only with the `isa` feature on; without it, "Bareword found where operator expected".

### Round-5 ledger

New findings. Columns as in the earlier ledgers. E18 and F12 were re-scoped in this round; their rows are in the round-4 ledger.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A27 | A | 5 | §9.1 categories; §9.2 refusal gate | The **syntax** category also covers perl's op-tree checks (arity, prototypes, lvalue); a file so labelled that RD also refuses makes `chart-accepted-both-refused 0` unreachable, pushing toward a mislabel or a false "facts in items" kill. Latent at waypoint 1. See also C22. | `TestArity` + `perl -c`: `sub g($) { 1 } g(1, 2);` "Too many arguments", RD refuses, perly.y accepts; `length(@a, 1)` RD refuses; `$x++ ++` RD accepts; scan of perl-refused parses/parsent finds one such file (`ExprNameNot-002.pl`), RD accepts it. | stand aside | released at 5fe8dc4b | b85dc97d |
| B17 | B | 5 | §8 | Head-only fitted refusals catch shallow chains only; a deep chain (`1 .. -2 .. 3`) is accepted through a reading the fitted rule leaves as the operand's only derivation, so the forest keeps it; rule 5 broken; §12's tests pass before the work. See also D26, E18, F12, G16, H9; I1. | `bisonp2` (7 unconditional actions restored) against `fitted.py`: 19 controls accepted; 8 shallow refused; 8 deep (`1 .. 2 + 3 .. 4`, `1 .. -2 .. 3`, `$a <=> !$b <=> $c`, `not $a <=> $b <=> $c` …) accepted by fitted Earley, refused by perl. | block | released (ruling), against B's bar on a reading rather than a measurement at 5fe8dc4b | b85dc97d |
| C22 | C | 5 | §9.1 ("syntax … whatever perl's message says") | The syntax category also covers op.c's semantic checks no grammar can refuse; such a file labelled syntax and refused by RD makes the gate unpassable. Latent until perl.git `t/`. See also A27. | `na/r21.pl` (`$x++ ++`): perl "Can't modify postincrement", RD refused=false; round 1's `lenient-reasons.txt`: "Can't modify logical and", "Invalid CODE attribute", both RD-accepted. | stand aside | released, against C's own bar at 5fe8dc4b | b85dc97d |
| D26 | D | 5 | §8; §12 refusal tests | A fitted refusal over an operand's *top* refuses `1..2..3` but accepts shielded chains (`1 .. 2 + 3 .. 4`, `$a <=> $b <=> $c . 1`); the forest keeps them; every §12 test uses an unshielded form, so the wrong build passes. See also B17, E18, F12, G16, H9. | `python3 -I D/shield.py`: `n .. n .. n` False, `n .. n + n .. n` and `n + n .. n .. n` True; `perl -c` refuses `@a=(1 .. 2 + 3 .. 4)`, `@a=(1 + 2 .. 3 .. 4)`, `$r = $a <=> $b <=> $c . 1`, `$r = $a <=> $b <=> $c \|\| 1`; RD refuses all four (`TestDProbe5`, `D/probe5b.out`). | block | released at 5fe8dc4b | b85dc97d |
| D27 | D | 5 | §9.1; §9.2; §11 | A file perl refuses for a lexing error pvm's lexer misses is labelled syntax; if RD refuses it for another reason, the refusal gate fails for a reason no `internal/chart` issue can fix (postmortem change 6's shape); `lex-unmapped`'s known list covers only refused files. See also C17, A24. | The spec text of §9.1, §9.2 and §11. | stand aside | held; stand aside at 5fe8dc4b | b85dc97d |
| G16 | G | 5 | §8 | A `%nonassoc` refusal is defined relative to the ordering, so local fitted rules in an unordered recognizer leave a derivation for shielded forms (`1 .. 2 + 3 .. 4`, `++$x->{a}++`); the over-acceptance is silent at waypoint 1 while §8 says it cannot happen. See also B17, D26, E18, F12, H9. | perl 5.42 refuses `1 .. 2 + 3 .. 4`, `1 .. 2 \|\| 3 .. 4`, `$a <=> $b + 1 <=> $c`, `++$x->{a}++`; accepts `(1 .. 2) \|\| (3 .. 4)`, `++$x ** 2`, `defined ref $x`; dodging derivations listed; round 4: 0 of 20 population files hold these forms. | block | released at 5fe8dc4b | b85dc97d |
| H9 | H | 5 | §8; §12 | Local fitted rules kill only side-by-side derivations; a tighter operator between leaves a reading (`(1..2) + (3..4)`) accepted at both waypoints; RD refuses every such case, so a corpus occurrence turns the refusal gate red with no rule to fix it; §8's opening paragraph rules out the needed mechanism. See also B17, D26, E18, F12, G16; H1. | perl 5.42 "syntax error" on `1 .. 2 + 3 .. 4`, `1 .. 2 * 3 .. 4`, `1 .. 2 \|\| 3 .. 4`, `1 .. 2 && 3 .. 4`, `1 .. $a == $b .. 2`, `$a <=> $b + 1 <=> $c`, `$a <=> $b . 1 == $c`, `$a <=> $b x 2 <=> $c`, `$a == $b + 1 <=> $c`, `$a < $b + 1 isa A`; RD refuses `H/r5/c5`–`c10`. | block | released (ruling) at 5fe8dc4b | b85dc97d |
| H10 | H | 5 | §9.3 ("a GC before each timed run") | Forcing GC moves the collection cost of allocation-heavy code out of the timing; the chart allocates more, and allocation was the first attempt's hot spot, so the kill could under-read it. See also F7. | Postmortem "Recognizer hot spots" (`fmt.Sprint` keys 74%); §9.3's "as an editor holds one". Not measured. | stand aside | held; stand aside at 5fe8dc4b | b85dc97d |
| H11 | H | 5 | §5.2 (conditional action refusals) | Wording: the slurpy default is two alternatives that refuse unconditionally; following the example, a worker writes a fitted rule where the ledger should exclude, and ledger and grammar disagree. See also A26, G13. | `perly.y:1104-1111`; catch case `perly.y:665-671` (`if(!$scalar)`). | stand aside | released at 5fe8dc4b | b85dc97d |
| I1 | I | 5 | §8; §12 | §8's refusal list is incomplete: the differential confirms both named families and finds two more (`PLUGIN_LOW_OP` after a comma list; `LSTOP PERLY_DOLLAR BAREWORD [ ]` from `%prec PREC_LOW` on `indirob: scalar`); the chart accepts them and no waypoint-1 gate notices. See also B17, D26, E18, F12, G16, H9. | `mk2.py`: `prec` (perly.y with precedence, 0 conflicts) against `noprec` (`%glr-parser` + `%merge`, 1,343 S/R) over 1.08M strings of length 1–4, 3 × 7.9–9.8M of length 5, 82k random derivations; `D*.txt`; perl 5.42 refuses `$fh=\*STDOUT; print $fh []`. | block | released, against I's bar; PLUGIN part re-scoped as I8 at 5fe8dc4b | b85dc97d |
| I2 | I | 5 | §8; §12 | Waypoint 2's constraint can keep the wrong grouping and nothing catches it: `parseoracle.Markers` has five kinds, none changed by `$a + $b * $c` grouped wrongly; perl cannot be the oracle for a toy grammar (rule 11). See also D15. | `parseoracle.Markers` on pu: srefgen, rv2hv, match, readline, anonhash; `perl -MO=Deparse,-p` returns full grouping; a bison-built perly.y parser in I's scratch (~100 lines). | stand aside | held, re-scoped (two stale sentences, §12 and §4); stand aside at 5fe8dc4b | b85dc97d |
| I3 | I | 5 | §9.3; §9.5 (`lex-overhead`) | The token interface's cost, which killed chart-waypoint, is first enforced at the final issue, repeating the first attempt's failure. See also A22, C16, D24. | Postmortem change 2 ("from its first recognizer issue, not a single speed checkpoint"); rule 10; an absolute threshold, median of three, can fail in any issue. | block | released at 5fe8dc4b | b85dc97d |
| I4 | I | 5 | §9.2 ("excluded with a reason") | Exclusion reasons have no limit; a partial port can pass the exit and "scope is all of Perl" goes unverified. | §5.2 names a closed set of excepted terminals but no closed set of reasons for excluding alternatives. Not measured how many of the 348 the corpora exercise. | stand aside | released, against I's bar at 5fe8dc4b | b85dc97d |
| I5 | I | 5 | §9.4; §10 | If readings are counted before the precedence constraint, an 8-operator chain gives Catalan(8) = 1430 > 2^10 and the collapse kill fires on precedence alone. | §6, §9.4 list the constraint before the count but do not say the count follows it; first design's rule 4 counts "after `chart/resolve`"; `chain.pl`: longest chain 5 in std5k, 8 in T1, 32 in perl `t/op/64bitint.t`. | stand aside | released at 5fe8dc4b | b85dc97d |
| I6 | I | 5 | §9.2; §9.5 | Minor wording conflict between "an `na` key that starts printing a number fails it" and the first-issue baseline. See also H5. | `cmd/ratchet/compare.go` "baseline still na; take a baseline". | stand aside | released at 5fe8dc4b | b85dc97d |

### Round-5 clusters

Counted by the first section each finding names, as in earlier rounds.

**Round-5 new findings (15, at b85dc97d)**

| section | count | findings |
|---|---|---|
| §5 | 1 | H11 |
| §8 | 6 | B17, D26, G16, H9, I1, I2 |
| §9 | 8 | A27, C22, D27, H10, I3, I4, I5, I6 |

Within §9: §9.1 3 (A27, C22, D27), §9.2 2 (I4, I6), §9.3 2 (H10, I3), §9.4 1 (I5).

By raiser: A 1, B 1, C 1, D 2, E 0, F 0, G 1, H 3, I 6. E18 and F12 were re-scoped, not raised anew.

**Rows open at b85dc97d (18: the 15 round-5 findings plus E18, F12 and H5)**

F12 is counted at its re-scoped section (§8); its round-4 row names §5.2 first.

| section | count | findings |
|---|---|---|
| §5 | 1 | H11 |
| §8 | 8 | B17, D26, E18, F12, G16, H9, I1, I2 |
| §9 | 9 | A27, C22, D27, H5, H10, I3, I4, I5, I6 |

**Open blocks by section (8):** §8 7 (B17, D26, E18, F12, G16, H9, I1), §9 1 (I3).

### Per participant, round 5

#### A (codebase and architecture)

- **Recommendation:** "Accept on the codebase lens."
- **Released:** A25, A26. All earlier rows stay released.
- **Held:** none.
- **New:** A27 (stand aside).
- **Checks reported:** pinned labels extend an existing mechanism (`perl-rejects.txt`, `chart.RejectsDisagree`); `Classify` would have called 12 both-refused files "Other" that are syntax under §9.1 (3 parses, 9 parsent); refusal totals reproduce §9.1 (23 of 769; 46 of 124); cwd matters and §9.1 records it; under §9.1's categories T1 contributes no syntax refusals; §8's stated refusals and acceptances match perl 5.42.0. Note on A26: the slurpy default refuses unconditionally, though §5.2 calls it conditional.
- **Did not re-examine:** §7's sparse table beyond its wording (`f()->{foo}` not probed); §8's fitted rules as grammar; GC and the baseline rule beyond reading; T1 from the repository-root cwd.

#### B (grammar fit)

- **Recommendation:** modify, B17 block.
- **Released:** B15 (against its own bar), B16. B1–B14 stay released.
- **Held:** none.
- **New:** B17 (block), with three options: deep refusal in recognition, the forest may refuse, or a recorded limit.
- **Did not re-examine:** §9.3 and §9.1's categories; §7's sparse table; heredoc attachment; whether option (a) stays small; lexer-level token streams for the deep cases.

#### C (measurement and kill design)

- **Recommendation:** accept.
- **Released:** C5's remainder and C16's residue, C19, C20, C21, each against its own bar.
- **Held:** none.
- **New:** C22 (stand aside).
- **Re-runs:** `lex-overhead` stand-in 1.124–1.267 alternating (10 of 10 under 1.3), A/A alternating 0.907–1.021, sequential A/A to 1.153 and sequential stand-in to 1.354; §8's refusal list and neighbours checked against perl and RD (`na/r1–r22.pl`, `na/i.pl`).
- **Did not re-examine:** §7; heredoc attachment and §5.2's exemptions; §12's refusal tests beyond the verdicts; the forest constraint's feasibility; the chart side of any timing.

#### D (spec quality)

- **Recommendation:** modify.
- **Released:** D20, D21 (against its bar; remainder re-scoped as D26), D22 (against its bar), D25.
- **Held:** none.
- **New:** D26 (block), D27 (stand aside). One discard.
- **On consistency:** §8 agrees with §5.2, §9.2 and §12 "on their face". "One inconsistency is of substance": a fitted refusal over an operand's top cannot refuse a shielded chain; refusing it needs the exposed edge, "and constraining edges is ordering" (D26).
- **Did not re-examine:** the cost of edge-constrained operands; shielded forms of prefix-with-postfix refusals; §9.1 beyond D27 and §9.3's GC; the glob-slot case; blead's verdicts.

#### E (architecture and direction)

- **Recommendation:** modify.
- **Released:** E17 (partly against its own bar; residue carried into E18).
- **Held:** E18, re-scoped, now block, "needs perigrin's choice among options 1–3": record a rule-5 exception; let the forest refuse a span whose every derivation is forbidden; or edge-aware stratification in recognition.
- **New:** none.
- **Disagreement recorded:** "'safe' semantics buys nothing for perl and costs rule 5."
- **Did not re-examine:** §7; §9.1's labels; §9.3's baselines and GC; conditional action refusals; other `%prec` uses; any perly.y-derived measurement.

#### F

- **Recommendation:** modify.
- **Released:** F7, F10 (against its own bar; `$h{-foo}` residue not held).
- **Held:** F12, re-scoped, block: parts (a) and (c) released, part (b) replaced by the shielded-chain mechanism.
- **New:** none.
- **Hunt notes (no finding):** baseline and `na` sentences consistent with `compare.go`; pinned labels feasible and cannot pass before their work; the refusal gate's restriction consistent with §12 carrying rule 5; the sparse table consistent with rule 8 if refinement depth is bounded per class; `PLUGEXPR`/`PLUGSTMT` exemptions consistent with §2.
- **Did not re-examine:** whether a partial ladder fits the speed budget; other forms bison refuses through precedence (prefix-with-postfix split by a tighter operator); §9.1 categories against real T1 refusals, perl.git `t/`, other rows; H4 and H5.

#### G

- **Recommendation:** modify, block on G16.
- **Released:** G11, G12 (superseded by G16), G13, G14, G15.
- **Held:** none.
- **New:** G16 (block).
- **Hunt notes (no finding):** §9.2/§9.5 baseline sentences consistent with `compare.go:84–85`; GC before each run moves only the previous run's garbage ("C's lens; I don't raise it"); glob slots left to superposition, accepted.
- **Did not re-examine:** rows outside its own; whether partial stratification changes the speed picture; §9.3's floor and GC; the other `%nonassoc` levels (`UNIOP`, `LSTOP`, `KW_REQUIRE`, `LOOPEX`); no chart built.

#### H

- **Recommendation:** modify.
- **Released:** H1 (against its own bar), H2, H3, H4, H6, H7 (H7: "a ruling, not a finding"), H8.
- **Held:** H5, narrowed, stand aside.
- **New:** H9 (block), H10, H11 (stand aside).
- **Classification changes, as given:** named refusals Ready; refusals with a tighter operator between Blocking (H9); forest constraint Partial; refusal gate and labels Ready; token interface Ready at decision level; ratchet reset Partial (H5); speed protocol Partial (H10).
- **Did not re-examine:** H9 option (a)'s interaction with Leo or LR(0) state cost; bison's other `resolved as` entries ("for example through `%prec UNIOP` on `KW_DO`/`KW_LOCAL`"); heredoc attachment beyond one example; `PLUGEXPR`/`PLUGSTMT`, §9.3's arithmetic, §13; other rows.

#### I (cold reader)

- **Recommendation:** modify.
- **Released / held:** none; I had no prior findings.
- **New:** I1, I3 (block); I2, I4, I5, I6 (stand aside).
- **Method (I1):** perly.y as bison builds it against the same rules without precedence, built as GLR with `%merge`, over enumerated token strings of up to 5 tokens and 82k random derivations. Limits: token strings, not Perl source; short strings, mostly under `GRAMEXPR`.
- **Classification, as given:** Ready: §2, §3 rules 7–12, §4, §5.2 grammar and ledger, §5.2 start forms, §5.3, §9.1, §9.4/§10/§11 apart from I2 and I5. Partial: §5.1, §7, §8 forest constraint (I2, I5), §9.2 (I4), §9.3 (I3), §12. Blocking: §8 refusals in recognition (I1).
- **Axes, as given:** codebase aligned; architecture aligned, "Rule 11 is only partly met (I1, I2)"; direction aligned. "It says too little in one place: §8 names its refusals rather than saying how they are found."
- **Did not check:** toke.c's refusals on pvm's tokens, including whether `TermFollows` reproduces `print $fh []`; statement-level and longer contexts; how many of the 348 alternatives the corpora exercise; the cost of §7's key or superposition; the 20-day budget; RD's verdicts and T1's 236.

### Round outcome

**Another round owed.** Eight blocks are open: B17, D26, E18, F12, G16, H9, I1, I3. No participant has released any of them. Six concern shielded chains, on which perigrin has since ruled "Forest refuses (bison-exact)"; their raisers have not yet answered the ruling. 18 rows are open in all. A ("on the codebase lens") and C recommend accept; B, D, E, F, G, H and I recommend modify.

## Author's response to round 5 (revisions aa27c6aa and this commit)

The author is the orchestrating agent. Release is each raiser's to give. perigrin ruled on the shielded-chain question after the reports: **"Forest refuses (bison-exact)"**. The subject grew from about 4,280 to about 4,550 words, mostly in §8.

| id | response | where |
|---|---|---|
| B17, D26, E18, F12, G16, H9 | Under the ruling, refusal by precedence leaves recognition entirely; no fitted rule tries it. Recognition keeps every derivation `perly.y`'s rules allow, shielded chains included. In waypoint 2's forest, precedence first keeps the derivation bison would build, then forbids what bison would never build — `%nonassoc` chains as bison fires them and the non-chaining comparisons, including chains split by a tighter operator (`1 .. 2 + 3 .. 4`, `$a <=> $b + 1 <=> $c`). A span whose every derivation is forbidden is refused. The value carried is each derivation's exposed operator context at both edges, so a shielded chain is seen through the tighter operator. Waypoint 1 over-accepts these forms, stated, and their tests are waypoint 2's. | §8, §12 |
| I1 | The constraint covers every conflict bison resolves by precedence, `%prec` on non-operator rules included (`indirob: scalar %prec PREC_LOW`, `print $fh []`). The list is found rather than written: tests come from I's differential, `perly.y` with precedence against `perly.y` without (GLR). The `PLUGIN_LOW_OP` family is out of scope with the keyword plugins. | §8 |
| I2 | A bison-built parser of the pinned `perly.y` is the referee for grouping and acceptance on generated operator sentences; perl is the oracle on the corpora. | §8 |
| I3 | `lex-overhead` is gated at an absolute 1.3, median of three, in every issue that touches chart/lex or the token table, from the first. | §9.3, §9.5 |
| A27, C22 | The categories are now **grammar**, **precedence**, **semantic** (arity, prototypes, op.c's "Can't modify") and **environment**. Waypoint 1's refusal gate counts grammar refusals only; precedence joins at waypoint 2; semantic and environment are reported, never gated. Feature-gated forms (`isa`) are labelled under the preamble's feature set. | §9.1, §9.2 |
| D27 | The refusal gate excludes files on `lex-unmapped`'s known list, which is confirmed at review. | §9.2 |
| H10 | GC is forced between runs, outside the timed span, so each side pays for the collection its own run triggers. | §9.3 |
| H11, A26 | The slurpy default is no longer given as a conditional example; only `catch` without a variable remains. | §5.2 |
| I4 | Exclusion from the port is allowed only for a reason on a closed list (keyword plugin, unconditional action refusal, start selector with no consumer), and the count is reported. | §9.2 |
| I5 | Readings are counted after precedence and scope have acted. | §9.4 |
| I6, H5 | An `na` key that starts printing a number fails the ratchet unless that issue takes its baseline, in that issue's commit. | §9.2, §9.5 |

## Round 6

**Revision assessed:** 5fe8dc4b (5fe8dc4b4232e1a6a72b73306339eed43e770eec). All ten participants name this revision as the one they read. Code read at `origin/pu` 4fffe4d6; perl at 94e5086608; perl 5.42.0 and bison 3.8.2 on the box.
**Participants:** A–I resumed (I was round 5's cold reader). **J** joins as this round's cold reader: it read the subject without the prior findings and did not open this minute or any other report.
**Round 6 dispatched:** 2026-10-30.

The status column of earlier rows above now carries each raiser's round-6 word where the raiser gave one. New findings are below. Section references are to the subject at 5fe8dc4b.

### Recommendations, as declared

- **A: modify.** "I block on A28 only: the referee's place, pinned outside the build or an explicit exception to §4 and the pure-Go rule."
- **B: accept,** "with B18 and B19 as stand-asides".
- **C: accept.** "The ruling's consequences are carried through to the gates and the labels. No block of mine is open."
- **D: "accept, with stand-asides."** "No finding of mine blocks … D29 and D31 are consistency edits that should land before acceptance freezes the text. D28 and D30 decide whether two gates mean what they say."
- **E: modify.** Block on E21.
- **F: "accept, with one stand-aside (F13)."** "No block. F12 is released."
- **G: accept.** "G17 should land before refinement, because it decides where the referee's toolchain lives."
- **H: "accept, with three stand-aside rows (H10, H12, H13)."** H14 is also a stand-aside in its ledger.
- **I: modify.** "Block: I7. The fix is one sentence in §12 and §9.1."
- **J: modify.** "The core claims hold up … One finding blocks (J7). The others stand aside."

### Blocks open after round 6, as declared

| participant | blocks |
|---|---|
| A | A28 ("it contradicts a stated line of the decision and a repository rule; the fix is one sentence") |
| B, C, D, F, G, H | none |
| E | E21 ("A required waypoint-1 test cannot pass. The fix is a sentence.") |
| I | I7 |
| J | J7 |

Four blocks are open: A28, E21, I7, J7. E21 and I7 describe the same defect.

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A29 | None stated beyond the row. |
| B18, B19 | None stated beyond each row. |
| C23 | Latent: "no current corpus file is affected." |
| D27 (held) | None stated beyond the row. |
| D28–D31 | "Each needs about one sentence." |
| D32 | "C's lens owns the numbers." |
| E19 | "The referee test is the backstop." |
| E20 | None stated beyond the row. |
| F13 | None stated beyond the row. |
| G17, G19 | "Each is one sentence." |
| G18 | "Editorial." |
| H10 (held) | None stated beyond the row; "I have not measured it, since no v2 code exists." |
| H12, H13, H14 | None stated beyond each row. |
| I2 (held, re-scoped) | None stated beyond the row. |
| I8 | "Waypoint 2 is refined after waypoint 1's verdict, but the sentence decides the matter now, and wrongly." |
| I9, I10 | None stated beyond each row. |
| J1–J6, J8–J10 | "The others stand aside." |

### Releases the participant said were against its own bar

- **B17:** "released (ruling), against my bar on a reading rather than a measurement". "I did not build the forest, so this is a reading of §8, not a measurement."
- **C22:** "Every release below is against my own bar."
- **I1:** "Released, against my bar. The definition in §8's 'Second' step still names only the operator cases. The `indirob` case is said to be 'resolved the same way', but nothing defines how an edge context sees a refusal that depends on lookahead."
- **I4:** "released, against my bar … The other half of my resolution was a key for alternatives the corpora never exercise. It was not adopted, and I do not hold for it."

Bias notes without an "against my bar" release: H9 ("I am the participant most likely to see that bar as met, so I note it"); I3 ("This is my bar met, and I am biased toward seeing it met"; residue noted, not held: a partial table may time only a prefix of a file). A, D, E, F and G mark no single release as against their bar.

### Earlier rows not addressed in round 6

None. Every row open at b85dc97d was given a round-6 word by its raiser.

E adds a word on a row it released in round 5: E17 "Released in round 5, but its case comes back at this revision; raised as E21". I7 says the same of F12 (a); F's round-6 word on F12 is "released".

### Disagreements with a ruling's consequences, as recorded

None. E: "The ruling chose the variant the evidence supported." C, D and I: no disagreement with "Forest refuses (bison-exact)". G and H: the ruling took the option each offered in round 5. I8 "is about a sentence that goes beyond the ruling, not about the ruling itself."

### Measurements and views that differ between reports

Recorded as reported. The drafter does not reconcile them.

- **`lex-overhead` at an absolute 1.3 per issue.**
  - A (stand-in with §7's full key, aggregate): medians 1.142 / 1.080 / 1.193, per-pair 0.88–1.44; std5k medians 1.02 / 1.09 / 1.28, per-pair 0.34–2.16.
  - C: thin stand-in 1.065–1.293 (10 of 10 under 1.3); a naive full key (one map keyed by a struct of four strings) 1.34–1.57, failing every trial: "the gate does its job".
  - J: stand-in 1.05× on std5k, "a lower bound".
  - D32 (not re-measured): near its threshold on every issue, citing round 1's A9 and C5. G (hunt note): "Near 1.3 it may flap."
- **The bison referee's feasibility and place.** A28, G17, H12 and J6: bison and a C compiler at test time against the pure-Go rule; A28 blocks, the others stand aside. F (hunt note): "feasible for operator sentences … Mapping … onto lexer tokens for the chart's side is implementation." J6: "A Go LR driver over committed tables is feasible." D28: the referee and the chart do not share an alphabet for fitted families.
- **The precedence label.** B19, C23, D30 and I10 (stand-asides): a reviewer's judgment that can take a file out of the waypoint-1 gate; three propose a mechanical check by the differential. F (hunt note): "acceptable for the few such files at waypoint 1". G (hunt note): "reviewed when the file is committed, which is enough". H: "workable … a wrong label shows up there [waypoint 2] rather than staying hidden." J7 also says the label "could be mechanical".
- **The probe §8 cites.** E20: it is E's `strict.py`, a six-level toy; "name the probe's scope". E re-ran `strict.py`: 0 of 447/443/445 bison-refused strings accepted. I9: "I found no probe of the round-5 definition's refusals" in the record (it cites E15 and E18 under the earlier definition). J8: the probe is not visible from the decision.
- **A literal reading of §8's forbidden list.** E19 (`literal.py`): accepts 270/254/254 of bison-refused strings; `strict.py` 0.
- **The slurpy default.** D: "unconditional (perly.y:1105-1112)"; `perl -c` gives "A slurpy parameter may not have a default value". F (hunt note): "The `yyerror` at `perly.y:1107/1111` has no `YYERROR`, so whether the action refuses depends on perl's error count, not the grammar." B, G, H: unconditional.
- **The forced GC outside the timed span.** H10 (held): allocation under the trigger is never charged. C (not a finding): "a small bias in the chart's favour, and it was already there in the first attempt". I (hunt note): "by at most one GC cycle per run … This is H10's row."
- **The refusal gate's population.** J: parsent 29 both refused; T1 74 perl refusals, 15 also RD-refused, all environment, so 0 grammar refusals in T1 ("an upper bound"). G (round 4): RD refuses 29 of the 46 perl-refused parsent, 20 with "syntax error". J2 calls the population thin.
- **`code_too_large.t`.** J: 50.3% of RD time over a 691-file T1 superset population. C (round 3): 48.6–51.7% over three runs.

### Round-6 ledger

New findings. Columns as in the earlier ledgers.

| id | raiser | round | section of the subject | consequence if it ships | evidence | block/stand-aside as declared | status | revision |
|---|---|---|---|---|---|---|---|---|
| A28 | A | 6 | §8 ("Bison is the referee"); §4; rule 11 | The referee needs bison and a C compiler, contradicting §4 ("no build step reads it") and the repository's pure-Go rule; in the ratchet or `make test` CI needs a C toolchain, and if run nowhere reproducible the outside reference is a scratch artifact. See also G17, H12, J6. | `ci.yml:2` "Pure Go build — no CGO"; `Makefile:79-82` `CGO_ENABLED=0`; `perly.y:34` `%define api.pure` and named references, bison-specific; bison and gcc present in this container only. | block | open | 5fe8dc4b |
| A29 | A | 6 | §9.2 closed exclusion list; §5.2 | The closed list has no reason for an alternative removed by fitting a conditional refusal; `catch_paren: empty` is one, so the exit criterion fails on §5.2's own treatment. See also H14, J5. | `perly.y:812-814` (`catch_paren: empty`, "not really valid grammar"); refusal in the try alternative's mid-rule action (660-671); perl refuses `try {1} catch {2}`, accepts `catch ($e) {2}`. | stand aside | open | 5fe8dc4b |
| B18 | B | 6 | §8 | The non-operator `%prec` constraint depends on the parent context, not only exposed edges; a forest rule keyed on (rule, next token) refuses `print $$fh[0]`, which perl accepts; disagreement-driven generation never produces it, so the tests pass before the work. See also F13, J4. | `consist.py`: (rule 370 `indirob: scalar`, `[`) suppressed in state 193 (LSTOP), reduced in state 153 (inside `$`); `bisonp2` ACCEPT on `print $$fh[0]` (`r6.tok`); perl refuses `print $fh []`, accepts `print $fh [1]`, `print {$fh} []`, `print STDOUT []`, `print $fh +[]`. | stand aside | open | 5fe8dc4b |
| B19 | B | 6 | §9.2; §9.1 | The precedence label decides which perl refusals the waypoint-1 gate counts, and perl's message cannot tell it from grammar; a wrongly accepted grammar refusal can be cleared by relabelling. See also C23, D30, I10. | perl 5.42: `1 .. 2 + 3 .. 4` (precedence) and `$a isa B <;` (grammar) both "syntax error". | stand aside | open | 5fe8dc4b |
| C23 | C | 6 | §9.1 (precedence category); §8 | Whether a refusal is precedence or grammar is a reviewer's judgment though §8 builds the tool that decides it; a mislabel either makes the waypoint-1 gate unpassable or excludes a real grammar refusal. Latent. See also B19, D30, I10. | `na/p.pl`, `na/r8.pl`, `na/r9.pl`: perl syntax error, RD refused=true; `perly.y:1462-1487`. | stand aside | open | 5fe8dc4b |
| D28 | D | 6 | §8 ("same acceptance and same grouping"); §5.2; H4 | For fitted families the referee (perly.y tokens) and the chart (lexer tokens) do not share an alphabet; "same grouping" is undefined, and the outside reference is missing where fitting could be wrong. | §8's own example `indirob: scalar %prec PREC_LOW` is in H4's variables-and-indirob family; `perl -c -e 'print $fh [];'` syntax error. | stand aside | open | 5fe8dc4b |
| D29 | D | 6 | §3 rule 6; §6; §12 | Rule 6, §6 and §12 restate §8 more narrowly than §8 now says (no refusal; "the oracle"; perl as oracle on toy grammars). See also G18, I2. | Spec text, lines 82–84, 205–206, 466–467. | stand aside | open | 5fe8dc4b |
| D30 | D | 6 | §9.1 categories; §9.2 refusal gate | The precedence label decides whether a file is gated and nothing fixes it before chart work starts; a chart grammar defect can be "explained" by relabelling. See also B19, C23, I10. | §9.1 and §9.2 text; compare §9.2's `lex-unmapped` review confirmation. | stand aside | open | 5fe8dc4b |
| D31 | D | 6 | §8 against §2 and §9.4 | `PLUGIN_*` tokens are library operators, in scope at waypoint 2 by ruling, but §8 puts their precedence refusals out of scope with keyword plugins; a chained `PLUGIN_REL_OP` would be accepted with no test. See also G19, H13, I8. | perly.y:153-178 (`PLUGIN_LOW_OP`, `PLUGIN_REL_OP` `%nonassoc`); §2 lines 60–65; §9.4. | stand aside | open | 5fe8dc4b |
| D32 | D | 6 | §9.3; §9.5 (D24 partly reopened) | A wall-clock gate near its threshold on every chart/lex issue; a correct issue can fail and be re-run until it passes. See also I3, A22, C16. | §9.3; minute A9, C5. Not re-measured. | stand aside | open | 5fe8dc4b |
| E19 | E | 6 | §8 | Read as a closed list of chain shapes, "forbidden" is not bison-exact: `t:t^t:t+t` keeps `(t:(((t^t):t)+t))`. See also F13. | `E/r6/literal.py` (3,000 strings × 3 seeds): every bison-accepted string gets bison's reading; of bison-refused, the chart accepts 270/447, 254/443, 254/445; `strict.py` 0. | stand aside | open | 5fe8dc4b |
| E20 | E | 6 | §8 | §8 cites "the precedence probe recorded in the assessment" for exactness; it was a six-level toy, so the sentence reads as evidence for Perl. See also I9, J8. | The probe is E's `strict.py` (one `%nonassoc` binary level, two prefix operators, three other binary levels). | stand aside | open | 5fe8dc4b |
| E21 | E | 6 | §5.2; §9.1; §9.2; §12 | The four comparison `yyerror` alternatives are "refused in action" and labelled grammar, but recognition keeps every derivation, so waypoint 1 accepts their sentences: the §12 waypoint-1 test cannot pass, and a grammar-labelled file holds the gate above 0. E17's failure returned. See also I7; E17, F12, D21. | `grep -n YYERROR perly.y` (1467, 1469, 1485, 1487); `E/r4/eqop.py` re-run: 2 derivations each for `T NCEQOP T NCEQOP T`, `T NCEQOP T CHEQOP T`, `T NCRELOP T CHRELOP T`; perl and RD refuse `$a <=> $b <=> $c`, `$a <=> $b == $c`. | block | open | 5fe8dc4b |
| F13 | F | 6 | §8 ("found, not listed"); §12 | The forbidden set has no completeness measure; the only differential on record is ≤ 5 tokens under `GRAMEXPR`; a conflict with a longer or statement-level witness leaves the forest accepting what bison refuses, with green tests. See also B18, E19, J4; I1. | `bison --report=all perly.y`: 1,347 resolutions, 1,347 distinct (rule, token) pairs, 57 states; perl refuses `print $fh []`, accepts `print $fh [1]`. | stand aside | open | 5fe8dc4b |
| G17 | G | 6 | §8 "Bison is the referee"; §12 | The referee is a C program and the doc does not say where it runs; wired into `make test` it fails on a pure-Go machine, never re-run it goes stale. See also A28, H12, J6. | §8, §12; minute I1, I2 (scratch, "~100 lines"); CLAUDE.md "No CGO, no C compiler". | stand aside | open | 5fe8dc4b |
| G18 | G | 6 | §4 Afroozeh row; §12 "Precedence (waypoint 2)" | Two sentences left by the rewrite contradict §8: §4 "without removing sentences"; §12 "perl as the oracle" on toy grammars. See also I2, D29. | §4 table; §12 bullet 5 against §8 paragraph 4. | stand aside | open | 5fe8dc4b |
| G19 | G | 6 | §8; §2; §9.4 | `PLUGIN_*` is conflated with keyword plugins; a library operator at `PLUGIN_LOW_OP` would be accepted with no refusal anywhere. See also D31, H13, I8. | §2 l.60–65; §8; §9.4; §5.2 lists `PLUGIN_*` and `PLUGEXPR`/`PLUGSTMT` separately. | stand aside | open | 5fe8dc4b |
| H12 | H | 6 | §8 ("tested against a bison-built parser") | A referee running bison and a C compiler at test time breaks the pure-Go rule, or is skipped, and the waypoint-2 precedence tests silently stop testing. See also A28, G17, J6. | `Makefile:79-82` `CGO_ENABLED=0`; nothing under `Makefile` or `.github` mentions bison. | stand aside | open | 5fe8dc4b |
| H13 | H | 6 | §8 against §2 and §9.4 | `PLUGIN_*` are infix plugins, not keyword plugins; three `PLUGIN_*` levels are `%nonassoc`, so waypoint 2 would accept and group chains perl refuses, with no test. See also D31, G19, I8. | `perly.y:153,174,180`; bison report lines 26850, 28414, 28980 ("resolved as an error (%nonassoc PLUGIN_…)"); §2 l.64-65; §9.4. | stand aside | open | 5fe8dc4b |
| H14 | H | 6 | §9.2 closed exclusion list; §5.2 start forms | "A start selector with no consumer" is a listed exclusion reason while §5.2 ports all seven `GRAM*` forms; a worker could exclude six. See also J5, A29. | §5.2 l.159-163; §9.2's exclusion bullet. | stand aside | open | 5fe8dc4b |
| I7 | I | 6 | §12; §9.1; §5.2 | The four non-chaining comparison alternatives, reached by bison only through precedence, are still derived without it; their waypoint-1 refused-in-action tests cannot go green, and §9.1 files them under both grammar and precedence. F12 (a) returned. See also E21; F12, D21. | `./prec` 0 / `./noprec` 1 on `THING NCRELOP THING NCRELOP THING`, `… NCRELOP … CHRELOP …`, `… NCEQOP … NCEQOP …`, `… CHEQOP … NCEQOP …`; the three anonymous-sub alternatives 0 / 0; no chained non-chaining comparison in parses/parsent. | block | open | 5fe8dc4b |
| I8 | I | 6 | §8 ("outside scope with the keyword plugins") | Contradicts §2 and §9.4: waypoint 2 would map a `.tpm` operator to `PLUGIN_*` and accept forever what perl refuses for it. See also D31, G19, H13; I1. | toke.c: `PLUGIN_*_OP` from `PL_infix_plugin` (9368, 9478), `PLUGEXPR`/`PLUGSTMT` from `PL_keyword_plugin` (9344); three `%nonassoc` PLUGIN error resolutions; `prec` 0 / `noprec` 1 on `THING PERLY_COMMA THING PLUGIN_LOW_OP THING`, `THING PLUGIN_REL_OP THING PLUGIN_REL_OP THING`. | stand aside | open | 5fe8dc4b |
| I9 | I | 6 | §8 ("The precedence probe recorded in the assessment … refused exactly what bison refuses") | States a feasibility result the record does not contain; a reader takes bison-exactness by edge context as shown. See also E20, J8. | Minute E15 (7,665 toy strings, earlier definition), E18 (about a quarter of bison-refused accepted, earlier definition); no probe of the round-5 definition's refusals found. | stand aside | open | 5fe8dc4b |
| I10 | I | 6 | §9.1 (precedence label) | The label is a new way to list a defect away; a grammar gap can be labelled precedence and drop out of the refusal gate. See also B19, C23, D30. | §9.2's `lex-unmapped` confirmation has no counterpart for the label; the differential can produce a confirming token string. | stand aside | open | 5fe8dc4b |
| J1 | J | 6 | §7 | `$h{-foo}` and `+{-foo}` give `foo` the same key; scope collapses `$h{-foo}` to a call when `foo` is declared (WRONG); contradicts §7's "ambiguities the neighbouring tokens decide are covered by the key". See also F10 (residue), B15. | `lexer.Tokenize`: both `Operator{`(no B) `Operator-` · next `}`; Deparse `$h{'-foo'}` vs `{-foo()}`; 0 lines in std5k, 1 file in T1, 7 in perl5 t/lib. | stand aside | open | 5fe8dc4b |
| J2 | J | 6 | §9.2; §12 | Waypoint 1 can pass with terminal sets too broad: the refusal gate's population is 29 parsent snippets plus the action tests, T1 contributes 0, and `terminals-per-token` is reported only. See also H3, G1. | parsent 29 both refused; T1 74 perl refusals, 15 RD-refused, all environment. | stand aside | open | 5fe8dc4b |
| J3 | J | 6 | §5.2; §9.1 | Scope- or feature-conditional action refusals (`croak_kw_unless_class`, "Experimental subroutine signatures not enabled") are assigned nowhere and fit no label category. See also F12 (c). | perly.y:579, 614, 1193. | stand aside | open | 5fe8dc4b |
| J4 | J | 6 | §8; §10 | The forest definition is shown only on the operator ladder; ~120 of 1,347 resolutions are non-ladder, some refusals come from shifting into a later error; if one is not expressible, no stop applies. Wording: "where bison shifts into the conflict" should be "where bison's action is error". See also B18, F13. | bison report: 670 shift, 667 reduce, 10 error; ~120 involve UNIOP, UNIOPSUB, KW_REQUIRE, LOOPEX, PREC_LOW. | stand aside | open | 5fe8dc4b |
| J5 | J | 6 | §9.2; §5.2 | The exclusion list's "start selector with no consumer" contradicts §5.2's porting of all seven; `leo-disagree`'s population is unstated. See also H14, A29. | Text of the two sections. | stand aside | open | 5fe8dc4b |
| J6 | J | 6 | §8; §12 | A bison-built referee implies bison and a C compiler at test time against the pure-Go rule; whether it is in `make test` is unstated. See also A28, G17, H12. | `bison --xml` on perly.y: 706-state automaton (2.3 MB); a Go LR driver over committed tables is feasible. | stand aside | open | 5fe8dc4b |
| J7 | J | 6 | §3 rule 11; §7; §12 | The token table, the component attempt 1 died on, has no outside reference: tested on toy grammars and against pvm's own lexer, never against toke.c's per-token decision; superposition masks errors in the corpus gates. See also J1, J2; B19, C23, D30, I10 (the label). | §12's Generator bullet; the box's perl lacks `-DDEBUGGING` (`perl -DT` refuses); the pinned source can build one. | block | open | 5fe8dc4b |
| J8 | J | 6 | §1; §8; §13 | Load-bearing claims cite evidence a reader of the decision cannot see (recognition 1.5–1.9×, the precedence and chain-length probes, H2/H4/H5). See also E20, I9. | Postmortem text: no 1.5–1.9 figure. | stand aside | open | 5fe8dc4b |
| J9 | J | 6 | §3 | "Rules 1–6 carry over" is followed by a waypoint that deliberately breaks rule 5 for precedence; a reader of §3 alone gets the wrong contract. | §3 against §8 ("over-accepts"). | stand aside | open | 5fe8dc4b |
| J10 | J | 6 | §9.2; §11 | `lex-unmapped`'s known list exempts only empty-set tokens; a lexer defect producing a wrong but mapped token must be fitted (encoding the defect) or fail the superset gate. See also D27. | §9.2 gate text. | stand aside | open | 5fe8dc4b |

### Round-6 clusters

Counted by the first section each finding names, as in earlier rounds. A rule (§3 rule N) counts under §3.

**Round-6 new findings (34, at 5fe8dc4b)**

| section | count | findings |
|---|---|---|
| §1 | 1 | J8 |
| §3 | 3 | D29, J7, J9 |
| §4 | 1 | G18 |
| §5 | 2 | E21, J3 |
| §7 | 1 | J1 |
| §8 | 15 | A28, B18, D28, D31, E19, E20, F13, G17, G19, H12, H13, I8, I9, J4, J6 |
| §9 | 10 | A29, B19, C23, D30, D32, H14, I10, J2, J5, J10 |
| §12 | 1 | I7 |

Within §9: §9.1 3 (C23, D30, I10), §9.2 6 (A29, B19, H14, J2, J5, J10), §9.3 1 (D32).

By raiser: A 2, B 2, C 1, D 5, E 3, F 1, G 3, H 3, I 4, J 10.

Within §8, four groups account for 11 of the 15: the referee's toolchain (A28, G17, H12, J6), `PLUGIN_*` against keyword plugins (D31, G19, H13, I8), the probe §8 cites (E20, I9), and coverage beyond the listed shapes (B18, E19, F13, J4; E19 overlaps F13).

**Rows open at 5fe8dc4b (37: the 34 round-6 findings plus D27, H10 and I2)**

I2 is counted at its re-scoped section (§12, "two stale sentences" in §12 and §4); its round-5 row names §8 first.

| section | count | findings |
|---|---|---|
| §1 | 1 | J8 |
| §3 | 3 | D29, J7, J9 |
| §4 | 1 | G18 |
| §5 | 2 | E21, J3 |
| §7 | 1 | J1 |
| §8 | 15 | A28, B18, D28, D31, E19, E20, F13, G17, G19, H12, H13, I8, I9, J4, J6 |
| §9 | 12 | A29, B19, C23, D27, D30, D32, H10, H14, I10, J2, J5, J10 |
| §12 | 2 | I2, I7 |

**Open blocks by section (4):** §3 1 (J7), §5 1 (E21), §8 1 (A28), §12 1 (I7).

### Per participant, round 6

#### A (codebase and architecture)

- **Recommendation:** modify.
- **Released:** A27. All earlier rows stay released (T1 root 0 / whole-tree 236 re-run).
- **Held:** none.
- **New:** A28 (block), A29 (stand aside).
- **Other checks:** §8's new examples match perl 5.42.0; `lex-overhead` absolute ≤ 1.3 "feasible with a modest margin" (aggregate median of three 1.14; bracket: the stand-in allocates, which rule 8 forbids); the baseline and `na` rules consistent with `compare.go` once extended; §9.4 has no code to check.
- **Did not re-examine:** the forest constraint's bison-exactness; other non-operator `%prec` refusals beyond the example; §7, §13 and the budgets.

#### B (grammar fit)

- **Recommendation:** accept, with B18 and B19 as stand-asides.
- **Released:** B17 (ruling; against its bar, "on a reading rather than a measurement"). B1–B16 stay released.
- **Held:** none.
- **New:** B18, B19 (stand aside).
- **Re-runs:** `r5.tok`: 19 accepted by bison and the unfitted recognizer; 16 refused by bison and accepted unfitted ("waypoint 1's over-acceptance, now intended"); `grep -c YYERROR` 8; `lx/subkey` 1,757 / 12. B traced `1 .. -2 .. 3` through §8's two steps to a refusal, matching perl.
- **Did not re-examine:** §9.3; §9.4; the closed exclusion list; the author's bison-exact probe; an implementation of §8's constraint; blead's verdicts.

#### C (measurement and kill design)

- **Recommendation:** accept.
- **Released:** C22 (against its own bar). C1, C3–C21 and C5's remainder still released.
- **Held:** none.
- **New:** C23 (stand aside).
- **Checks reported:** the per-issue `lex-overhead` gate can pass (thin stand-in 1.07–1.29), is not free (naive full key 1.34–1.57), and cannot pass before its work; §8's new examples are syntax errors in perl and refused by RD; no current corpus file is precedence-only. The forced GC: "a small bias in the chart's favour … Not a finding." Records no disagreement with the ruling.
- **Did not re-examine:** §8's forest mechanism; the referee and the GLR generator; §5.2's closed list and the slurpy change; §9.4; the chart side of timing beyond the stand-ins.

#### D (spec quality)

- **Recommendation:** accept, with stand-asides.
- **Released:** D26. D checked a chain split by a looser operator (`$a <=> $b <=> $c || 1`) is refused through step 1 then step 2, "so §8's two-step order matters and is stated."
- **Held:** D27 (stand aside): the known-list exclusion covers files the chart refuses; D27's case is a file the chart accepts.
- **New:** D28, D29, D30, D31, D32 (stand aside). Two discards.
- **Did not re-examine:** the author's bison-exact probe and how the forest encodes lookahead-dependent `%prec`; whether "keep bison's derivation, else refuse" is computable from edge contexts for all 1,347 resolutions; §9.4 and §9.3's GC; blead's verdicts.

#### E (architecture and direction)

- **Recommendation:** modify.
- **Released:** E18 (under the ruling; `strict.py` re-run: 0 bison-refused strings accepted on three seeds). E17's case "comes back at this revision", raised as E21.
- **Held:** none.
- **New:** E19, E20 (stand aside), E21 (block).
- **Hunt notes (no finding):** §9.4 consistent with the 2^10 collapse; §2 unchanged and consistent; "prefix after prefix accepted" and `++$x++` refused still hold.
- **Did not re-examine:** whether bison and GLR parsers of perly.y can be driven on token sentences (actions stripped, `YYERROR` kept); how a reviewer tells precedence from grammar; `indirob` (`print $fh []`); §9.3; §7; H1–H5; any perly.y-derived measurement.

#### F

- **Recommendation:** accept, with one stand-aside (F13).
- **Released:** F12. F1–F11 still hold on re-reading.
- **Held:** none.
- **New:** F13 (stand aside).
- **Hunt notes (no finding):** the bison referee is feasible for operator sentences; the precedence label is "acceptable for the few such files at waypoint 1"; the closed exclusion list bounds "excluded"; absolute `lex-overhead` per issue "restores the postmortem's change 2" and cannot pass vacuously; baseline sentences agree; §9.4 consistent; the slurpy-default `yyerror` lacks `YYERROR` ("I did not test whether perl refuses `sub f ($x, @y = 1) {}`").
- **Did not re-examine:** the edge-context algorithm and whether it expresses rule 370's follow restriction; the `PLUGIN_LOW_OP` exclusion, I's other rows, §9.1's label file against T1; H4, H5, §7's sparsity, the budget.

#### G

- **Recommendation:** accept.
- **Released:** G16. G1–G15 re-read for regressions; "Nothing regressed."
- **Held:** none.
- **New:** G17, G18, G19 (stand aside).
- **Hunt notes (no finding):** I1's enumeration (≤ 5 tokens) cannot reach a 7-token split chain, but §12 names split chains, "so the family is covered"; the unused "start selector" reason is "harmless"; the precedence label's review at commit "is enough"; `lex-overhead` absolute bound does not conflict with "never compared strictly".
- **Did not re-examine:** whether exposed edge context decides bison-exactness locally ("I rely on I1's differential"); §9.4 and collapse thresholds; other rows; speed beyond the notes; T1 label counts under the four categories.

#### H

- **Recommendation:** accept, with stand-asides.
- **Released:** H5, H9 (the ruling took its option (b); bias noted), H11. H1–H4, H6–H8 remain released.
- **Held:** H10 (stand aside): the GC fix addresses cross-charging, "which was not my point".
- **New:** H12, H13, H14 (stand aside). A note, not a row, on §12's "perl as the oracle" bullet, left to I2.
- **Feasibility notes:** label categories workable; `lex-overhead` gate "cannot be dodged" but "can flake under load"; forest refusal taken from I1's differential, not reproduced; §9.4 consistent.
- **Did not re-examine:** I's differential and the round-5 probe; whether exposed edges express the 667 reduce resolutions and non-operator `%prec`; §9.3 arithmetic, §13, the token table; other rows.

#### I

- **Recommendation:** modify.
- **Released:** I1 (against its bar; PLUGIN part re-scoped as I8), I3 (residue noted), I4 (against its bar), I5, I6.
- **Held:** I2, re-scoped to two stale sentences (§12's precedence bullet; §4's Afroozeh row), stand aside.
- **New:** I7 (block); I8, I9, I10 (stand aside).
- **Re-runs:** `prec`/`noprec` differential on the round-5 inputs unchanged (150 length-≤4 disagreements); split chains and the four comparison chains 0/1; anonymous-sub alternatives 0/0; toke.c's `PL_infix_plugin` and `PL_keyword_plugin` are separate hooks.
- **Did not re-examine:** whether an edge-context forest constraint can be bison-exact ("I9 is about what the record contains, not a counter-example"); §7, §9.4's other text, §10, §11, §13, §14; other rows and H1–H5; `TermFollows` on `print $fh []`; statement-level contexts; speed.

#### J (cold reader)

- **Recommendation:** modify.
- **Released / held:** none; J had no prior findings.
- **New:** J7 (block); J1–J6, J8–J10 (stand aside).
- **Measurements:** 348/114, 32/39, PLUGIN lines and classes, RD's 236, perl's 78, `code_too_large.t` at 50.3% all hold; `lex-overhead` stand-in 1.05× on std5k (lower bound); bison resolutions 670 shift, 667 reduce, 10 error; §8's examples hold on perl 5.42; refusal gate population parsent 29, T1 0 grammar; superset T1 population 691; no Error/UnknownRest tokens in the corpora.
- **Classification, as given:** Ready: the port and ledger, the compiled table and `lex-overhead`, the superset gate, the speed protocol and kill, start forms, the builtins and levels cross-check, isolation (with J6/J10 caveats). Partial: the token key (J1), precedence in the forest (J4, J6), action refusals (J3), the refusal gate (J2), exclusions (J5). Missing: an outside reference for the token table (J7).
- **Did not check:** the minute, the milestone Records and H2/H4/H5; the recognition cost of ambiguity and chain lengths; edit5k.pl, perl.git `t/`, the Guacamole `parses` labels, whether std5k.pl compiles under perl; toke.c beyond the probes; chapter 4's table and §14's budget; the carried recognizer's speed on the new grammar.

### Round outcome

**Another round owed.** Four blocks are open: A28, E21, I7, J7. No participant has released any of them. E21 and I7 describe the same defect. 37 rows are open in all. B, C, D, F, G and H recommend accept, B, D, F and H with stand-asides named; A, E, I and J recommend modify.

## Author's response to round 6 (revision b3398943)

The author is the orchestrating agent. Release is each raiser's to give. No ruling was needed. The subject grew from about 4,520 to about 5,110 words.

| id | response | where |
|---|---|---|
| A28, G17, H12, J6 | The bison referee runs offline. A script pinned to the `perly.y` revision writes generated sentences (as terminal strings), bison's verdict and bison's grouping to committed fixtures; `make test` and the ratchet read the fixtures and need no C toolchain. The stripped grammar keeps its `YYERROR` actions. §4 says the referee is generated offline. | §4, §8 |
| E21, I7 | The four comparison-chain alternatives (`perly.y:1466-1487`) are marked "refused by precedence", not "refused in action"; their refusal is §8's and their tests are waypoint 2's. §12's waypoint-1 refusal tests cover only "refused in action". | §5.2, §8, §12 |
| J7 | The token table gets an outside reference: a `-DDEBUGGING` build of the pinned perl records `-DT`'s token type per token over the corpora to committed fixtures, and a test checks membership in the chart's terminal set and reports set sizes. | §12 |
| B19, C23, D30, I10 | The precedence label is shown by evidence: perl's token stream (from the same `-DT` build) is accepted by `perly.y` without precedence and refused with it. | §9.1 |
| B18, F13 | Generation covers every conflict bison resolves by precedence, with sentences bison accepts and refuses for each (`print $fh []`, `print $$fh[0]`); coverage is reported. | §8 |
| D28 | The referee's sentences are terminal strings, so no re-lexing is involved. | §8 |
| E19, J4 (wording) | The listed shapes are stated as examples; the definition is bison's resolution of each conflict. "Shifts into the conflict" is now "bison's resolution is an error action". | §8 |
| J4 (stop) | New stop: a precedence resolution not stateable over the forest from a span's derivations and neighbouring tokens goes to perigrin. | §10 |
| E20, I9, J8 | §8 cites E's probe by row and says it covers a six-level ladder only. §1's recognition figure now cites the postmortem's 34.9 → 18.3 ms against RD's 16–31 ms. §4 says where H1–H5 are listed. | §1, §4, §8 |
| G19, H13, D31, I8 | `PLUGIN_*` operators' precedence refusals are waypoint 2's with the library operators; only `PLUGEXPR`/`PLUGSTMT` are excluded. | §8 |
| A29, H14, J5 | The closed list drops "a start selector with no consumer" and adds "the alternative a fitted conditional refusal needs" (`catch_paren: empty`). `leo-disagree`'s population is unchanged and not restated here. | §5.2, §9.2 |
| J3 | Refusals in actions under scope or features (`method`/`ADJUST` outside a class, signatures not enabled) are a **scope** label and waypoint 2's. | §5.2, §9.1 |
| G18, I2 | §4's Afroozeh row says this design departs from their safety property. §12's precedence bullet uses the referee's fixtures, and bison builds of toy grammars, not perl. | §4, §12 |
| D29 | Rule 6 now says a constraint may refuse a span where no reading survives. | §3 |
| J9 | §3 states rule 5's waypoint-1 exception for precedence refusals. | §3 |
| J1 | A `{` neighbour's class carries subscript-or-anon-hash, derived from its own predecessor, so `$h{-foo}` and `+{-foo}` key differently; still a class, so the key does not recurse. | §7 |
| J2 | `terminals-per-token` is now may-not-rise. | §9.5 |
| J10, D27 | The known list also takes a file whose chart verdict traces to a wrong-but-mapped lexer token, under the same review; it leaves both gates' populations. | §11 |
| H10 | Within a timed run, collection is left to the runtime; GC is forced only between runs; allocation per token is reported for both sides. | §9.3 |
| D32 | Not changed. A's and C's measurements put a single lookup at 1.07–1.29 and a §7-shaped stand-in at 1.14; the gate is meant to push toward the sparse table. | — |
