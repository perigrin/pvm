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
| A9 | A | 1 | §7.4 | The 1.3 gate is feasible but sits within run-to-run noise; a ratio of two separately timed medians can flip with no code change. See also C5. | `lexov_test.go`, `-benchtime 30x -count 3` on std5k: Tokenize 15.9/19.9/18.6 ms; Tokenize+trivia+lookup 20.2/24.0/21.4 ms; ratio ≈1.15–1.27; 4 cores, load 2.6. | stand aside | held, re-scoped; stand aside at d6cb49c7 | 083e1357 |
| A10 | A | 1 | §5.2 GRAM* | "`Recognize` takes the start symbol as a parameter" is implementation detail, and more work than needed. See also D10; B7, D6 (start forms). | `BuildLR0(start string, …)`; `Recognize` seeds `{State: 0}`; perly.y:187–257 states start forms as terminals in one rule, so one GRAM* edge needs no API change. | stand aside | released at d6cb49c7 | 083e1357 |
| A11 | A | 1 | §5.2 refusals | The refusal "checked against internal/lexer's own … operator tables" cannot read that table under §11. | `go doc -short ./internal/lexer` exports `Keywords`, `IsKeyword`, `IsFileTest`; `operators`, `bitwiseStringOps`, `positionalOps` (scan.go:1125–1143) unexported. | stand aside | released; new defect raised as A20 at d6cb49c7 | 083e1357 |
| A12 | A | 1 | §5.2 example; §5.3 Q9 | (i) The example DSL would not compile while slif.go exists. (ii) Q9 is a real defect. See also D10 (i). | (i) `type Seq int` in slif.go collides with `func Seq`. (ii) Probe: `Star` with RHS `a b` builds with err nil as `S ::= ε \| S a`; `Plus` with empty RHS panics `index out of range [0]`. | stand aside | released at d6cb49c7 | 083e1357 |
| A13 | A | 1 | §5.3 | §5.3 cites Q4 and Q7–Q11 to `## Records`, which define only Q1, Q2, Q3, Q5/Q6; the citation cannot be checked. See also D10 (sourcing). | `grep -n 'Q4\|Q7\|Q8\|Q9\|Q10\|Q11' A/body.md` → only the review line; postmortem has no Q-ids. | stand aside | released at d6cb49c7 | 083e1357 |
| A14 | A | 1 | §9.1 | "`parse.Parse` returns no syntax error" names an interface that does not exist; `rd-accepted-chart-refused 0` can be computed two ways. See also C1. | `func Parse(src []byte) *Node` has no error result; acceptance is read from `Node.Refused()`; refusal codes include `UnimplementedStatement`, `CatchAll`. | stand aside | held; superseded by A17 at d6cb49c7 | 083e1357 |
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
| C1 | C | 1 | §9.1; §9.2 | "No syntax error" is not a predicate that exists, and the gated population includes files perl rejects; the chart must accept non-Perl (against rule 5) or hold `rd-accepted-chart-refused` above 0; the gate measures RD's leniency and the lexer. See also A14; C10. | `parse.go:653` "no error return"; `zzpred` × `perl -c`: parses perl-ok/rd-accepts 743, ok/refuses 3, rej/accepts **18**, rej/refuses 5; parsent 75/3/**17**/29; `Refused()` agrees with Error/Missing on all 769 + 124 + 986 files. | block | held, re-scoped (the predicate); block at d6cb49c7 | 083e1357 |
| C2 | C | 1 | §9.2; §9.1 | "As recursive descent accepts it" is false under §9.1's own definition; a `:infix(ADD)` fixture passes before the general case works. See also A2, D5; A3, B8. | `parse.Parse` has a nil resolver (module.go:323); `use My::Lib; 1 plus 2;` → `zzpred` refused=true although `My/Lib.tpm:32` declares `sub plus :infix(ADD)`; `TestLibraryOperatorMayUseAnyLevel`. | stand aside | released (ruling) at d6cb49c7 | 083e1357 |
| C3 | C | 1 | §9.3; §10 | Long-lived as defined replays a cache warmed on the same files; the kill reads the most favourable figure, and the redefinition alone moves the line more than any fix did. See also E2. | `v1/internal/zzlong`, recognition only, median of 9 after 2 warm-ups: aggregate (763 files) per-run 19.1/21.5/29.0 ms, long-lived **10.6/9.6/8.8 ms**; std5k 45.7/58.2/48.4 vs 47.3/45.3/42.4 ms. | block (on the definition, not on the choice of long-lived) | released, against C's own bar at d6cb49c7 | 083e1357 |
| C4 | C | 1 | §9.3; §10 | "The aggregate" is undefined (corpora, statistic, std5k/edit5k inclusion, both-accept restriction); execute would define it; a kill key dominated by one generated file judges the template. | `zzspeed` rd-ms (speed-aa.txt): std5k 24–33 ms vs all 769 snippets 6–9 ms; `code_too_large.t` 484 KB of T1's 2,176 KB (22%), 4,998 near-identical lines; T1 RD total 420–470 ms. | block | held, re-scoped (one file dominates the aggregate); block at d6cb49c7 | 083e1357 |
| C5 | C | 1 | §7.4; §9.2; §9.5 | A wall-clock 1.3 gate cannot be decided under §9.3's protocol here; a correct chart/lex fails about a third of the time and a 30% regression passes about as often; ambiguous whether chart/lex's time includes Tokenize. See also A9; B6, D4. | A/A `Tokenize` vs itself: 0.65–1.32 std5k, 0.67–1.35 aggregate; thin chart/lex 0.91–1.39 / 0.90–1.44, over 1.3 in 4 of 12 trials each; interleaved A/A 0.76–1.19 (speed-aa.txt, speed-paired.txt). | block (on gating a wall-clock ratio; reporting it is fine) | core released, against C's own bar; remainder held (which size the A/A rule selects), stand aside; see C12 at d6cb49c7 | 083e1357 |
| C6 | C | 1 | §3 rule 10; §9.3 | "Benchmarks … run in `make test`'s package tests" is false; nothing in `make test` goes red on a slowdown. See also A8, D3. | A panicking benchmark passes `go test ./... -count=1` ("ok bchk") and panics only with `-bench .`; `sed -n 21,22p Makefile`; ratchet CI advisory (Makefile:55-57). | block | released, against C's own bar; consequence raised as C15 at d6cb49c7 | 083e1357 |
| C7 | C | 1 | §10; §9.3 | "Checkpoint" is undefined and the protocol lets load drift into the ratio; near 3× the verdict depends on the orchestrator's reading. See also D11. | A/A of RD alone 0.90–1.23 std5k, 0.79–1.15 aggregate (speed-aa.txt); 2f46 `speed.go` `speedRatios` times the two blocks one after the other; v1 results 2.83–3.85. | block on defining "checkpoint"; stand aside on interleaving (issue-level) | released, against C's own bar; consequence raised as C13 at d6cb49c7 | 083e1357 |
| C8 | C | 1 | §2; §9.2; §9.4 | The waypoint-1 exit cannot answer §2's question (it bounds one side), and waypoint 2 never re-asks it; a parser 5× slower once the forest exists could pass. | Read §9.2–§9.4; `parse.go` `parseWith` (parseFile then finish). | block | released at d6cb49c7 | 083e1357 |
| C9 | C | 1 | §9.2; §2 | Nothing measures completeness of `perl.go`, so the criterion passes before its work; "all of Perl" is asserted but unmeasured. | Every §9.2 key (tables-drift, refusals, superset, leo-disagree) is silent about a perly.y alternative no corpus file reaches. | block | held, re-scoped (the ledger's reference); block at d6cb49c7 | 083e1357 |
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
| E5 | E | 1 | §9.2; §9.3; §14 | The first reading the kill can act on comes after most of the port, the largest spend. | `TestScratchLadder` (toy, cold): 1 level 2.5 items/token 4.57 µs; 8: 4.8, 3.14; 25: 9.2, 6.09; 32: 10.9, 6.67. | stand aside | held, narrowed; stand aside at d6cb49c7 | 083e1357 |
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
| A17 | A | 2 | §9.1 (from A14's fix) | Under the stated definition RD "accepts" every file; the population becomes "every file perl accepts", and gated `rd-accepted-chart-refused 0` demands the chart accept files RD refuses; §9.1's "it refuses 236" contradicts its own definition. See also C1 (held), A14. | `refused.go:14` is a per-node test; `TestRootRefused`: `my $x = ;`, `foo(;`, `1 2;` give root.Refused false, subtree true; over T1's 986 files root.Refused **0**, any node **236**. | block | open | d6cb49c7 |
| A18 | A | 2 | §9.3; §9.5 (from A8's fix) | With the ratchet in every recognizer issue's criteria, `rd-accepted-chart-refused` "gated at 0" and `lex-unmapped` "equals its known list" hold only once the port is complete; every earlier issue fails, or the port is forced into one issue. See also C15. | §9.5 rows; §7 "An empty set refuses the file … (`lex-unmapped`)"; `Compare` on pu gates only `mayNotRise` against a baseline; `TestRDTotal`: one RD pass over 1,879 files 0.46–0.52 s. | block | open | d6cb49c7 |
| A19 | A | 2 | §8; §5.1 | "Static data applied at completion" names a mechanism the carried recognizer cannot host unchanged; the carried recognizer becomes a modified one the 770-file reference agreement no longer covers. See also B10, D15, E12, F1 (e). | `recognize.go:70-73` `r.lhs[s]` compacted LHS set; `:255` completes by LHS, rule identity gone; `leo.go:37-45` memo jumps by `(state, symbol)`. | stand aside | open | d6cb49c7 |
| A20 | A | 2 | §5.2 refusals (from A11's fix) | "Checked by lexing it" refuses correct literals lexed alone in term position. See also B13, D14, F4. | `TestLiterals`: alone, `**`→`*` `*`, `/`→Error, `//`→Quote, `%`→UnknownRest, `&`→FuncSigil, `<=>`→Readline, `x=`→Word `x` + `=` (15 of 31 not one Operator); after `1 ` all 31 lex as one token; `x=`, `x`, `lt`, `cmp`, `isa` are Word. | stand aside | open | d6cb49c7 |
| A21 | A | 2 | §5.3; §3 rule 9 | The CORE.tpm cross-check's difference list is mostly definitional unless the shape↔token correspondence is stated. See also D9. | `TestShapes`: `parse.CoreShapes()` 207 names (list 91, unary 84, niladic 32); `not` is list (NOTOP), `goto` unary (LOOPEX), `last`, `my`, `local` absent. | stand aside | open | d6cb49c7 |
| B10 | B | 2 | §8 | The filter, "static data applied at completion", cannot be exact for perly.y, and in waypoint 1 it can only cause false refusals; "bison's language" is undefined once rules are fitted. See also A19, D15, E12, F1. | `consist.py` over a `-Dlr.default-reduction=accepting` build: 706 states, 680 suppressed (rule, token) pairs, one reduced elsewhere (370 `indirob: scalar`, `[`); Earley + filter accepts `1..2..3` and `$a isa B isa C` (bison rejects) and rejects `$$x[0]` (bison accepts); 5,606 random sentences all accept. | block | open | d6cb49c7 |
| B11 | B | 2 | §7 | The key has no lookbehind: a lone word in a subscript `{}` stays {string, call}; waypoint 2 scope picks the call reading for declared subs and has no fact for builtins, giving WRONG sites on most hash accesses. | perl Deparse: `{ print }` → `print $_`; `$h{print}`, `$h{shift}` → string keys; `lx/subkey` on T1: 1,757 lone words in non-block `{}`, 12 in blocks with the same key. | block | open | d6cb49c7 |
| B12 | B | 2 | §5.2 | The named fitting families miss variables (17 alternatives), forced version tokens (5 alternatives) and the end-of-file `;`. See also F2; D18. | bison rule list via `r2/earley.py`; `S_tokenize_use`; perly.y comment ~720; toke.c:7670. | stand aside | open | d6cb49c7 |
| B13 | B | 2 | §5.2 refusals | "Checked by lexing it" depends on lexer state and rejects MULOP, BITANDOP, POWOP and others. See also A20, D14, F4. | `lx/dump` alone: `/`→UnknownRest, `%`→Error, `&`→FuncSigil, `**`→`*` + Word `*`, `~~`→two `~`, `x=`→Word `x` + `=`. | stand aside | open | d6cb49c7 |
| C12 | C | 2 | §9.3 (`lex-overhead`) | "Per token" does not name the count; each side over its own count inflates a correct chart/lex to ~1.7–2.0, which always fails. See also F7. | `zzspeed2`: std5k tokens all 42,172, significant 26,288 (ratio 1.60). | block | open | d6cb49c7 |
| C13 | C | 2 | §9.3 (checkpoint 1); §10 | Checkpoint 1 needs std5k and the whole superset population, close to the WP1 exit; the speed kill cannot fire until most of the budget is spent; early aggregate treatment of still-refused files unstated. See also E5. | §9.3 against Records 2f46 (kill fired one day into 15) and the postmortem's "What Worked Well". | block | open | d6cb49c7 |
| C14 | C | 2 | §9.1; §9.2 | The `lex-unmapped` exclusion lets a chart table gap leave the superset gate by filing a lexer issue. See also D7. | §7 "An empty set refuses the file at that byte"; §9.1. | stand aside | open | d6cb49c7 |
| C15 | C | 2 | §9.3; §9.5 | The absolute gates make every intermediate issue unpassable, the shape postmortem change 6 forbids. See also A18. | Criterion run over a half-done tree: superset count in the hundreds, ratchet criterion fails. | block | open | d6cb49c7 |
| D14 | D | 2 | §5.2 refusals | The check refuses correct classes, so the generator emits no tables for a correct grammar. See also A20, B13, F4. | `go test -overlay D/overlay3.json -run TestDProbe2 -v ./internal/lexer`: alone `/`→UnknownRest, `%`→Error, `&`→FuncSigil, `**`→`*` Word `*`, `//`→Quote, `%=`→Variable, `<=>`→Readline, `x=`→Word `x` `=`; after `$x ` all correct, including `Word«x=»`. | stand aside | open | d6cb49c7 |
| D15 | D | 2 | §8; §4 (Afroozeh) | The cited technique's safety property keeps the unannotated language, the opposite of "rejects exactly what bison's resolution rejects"; a worker following it accepts `1 .. 2 .. 3`, seen only in a reported key. Also: fitted rules must keep perly.y precedence (HASHBRACK 1536, and implicit on 3 alternatives); the filter has no outside reference (rule 11). See also B10, E12, F1, A19. | `perl -c -e '@a = (1 .. 2 .. 3);'` syntax error; `perl -c -e '$a + defined $b;'` OK; Afroozeh et al. SLE 2013 abstract; `internal/parse/precedence.go:242` `NonassocOperators`; Python scan: 3 alternatives with implicit precedence on fitted-family terminals. | block | open | d6cb49c7 |
| D16 | D | 2 | §9.2; §5.2 port ledger; §4 | Completeness is checked against nothing outside the ledger; a ledger missing 20 alternatives passes. See also C9 (held), F5. | The spec text; bison 3.8.2 `--report` enumerates 348 alternatives with line numbers. | stand aside | open | d6cb49c7 |
| D17 | D | 2 | §4; §9.1 | The grammar (blead 5.45) and the labels (perl 5.42) are different perls; files using newer syntax drop out and inflate `chart-accepted-perl-refused`. See also F8. | `perl -v` 5.42.0; `git -C /opt/corpora/perl5 log --oneline v5.42.0..HEAD -- perly.y` (f1a8d7d883 PPC0024, da6c78e212 multivariable foreach). | stand aside | open | d6cb49c7 |
| D18 | D | 2 | §5.2 refusals; §9.2 (re-scope of D1) | A ported rule whose terminal has no token class is dead and passes every refusal and the ledger; `PLUGIN_*_OP` is in this state for all of waypoint 1. See also F2, B12. | §5.2 and §7 text; `D/tokens.txt`, `D/allterm.txt`. | stand aside | open | d6cb49c7 |
| E9 | E | 2 | §2; §13 | §2 glosses "time to an answer" as M2's latency, but M2's gate is an edit to a re-parsed tree in < 10 ms p99; the interleaved spike could be judged on a metric M2 does not use. | `sed -n 2582,2596p docs/specs/perl-parser/06-incremental-lsp.md`; `git zhi issue show 01a17969-4736`; plan §5 M2. | stand aside | open | d6cb49c7 |
| E10 | E | 2 | §1; §2; §5.1 | A positive verdict makes the chart "a candidate engine for M2" while convergence goes unmeasured through both waypoints; the convergence fixtures' fate is unstated. | `internal/chart/testdata/convergence-edits.jsonl`, `convergence-edits-guacamole.jsonl` on pu; v1 §8. | stand aside | open | d6cb49c7 |
| E11 | E | 2 | §13 | §13 attributes this assessment's measurement to the first attempt and quotes the lookahead-on figure where lookahead 0 applies. | Minute row E1 (431 ms lookahead on, 94 ms lookahead 0); re-run 379 / 103 ms. | stand aside | open | d6cb49c7 |
| E12 | E | 2 | §8; §9.2 | §8 states two contracts ("exactly the derivations" vs "bison's accepted language"); under the first, the named mechanism cannot pass §8's own `$a * defined $b + 1`; under the second it passes whether or not the filter works. The filter also decides speed. See also B10, D15, F1, A19. | `E/r2/toy.y` (bison 3.8.2): `t*ut+t` → `(t*u((t+t)))`, matching perl Deparse; `depth1.py` parent/child filter keeps 2 readings; `t*ut` + n `+t` keeps n+1 (n = 1..8); round-1 `TestScratchLadder`: flat 751.9 items/token vs stratified 2.5. | block | open | d6cb49c7 |
| F1 | F | 2 | §8 | The filter's contract cannot be met or checked as stated: (a) undefined over fitted rules; (b) level-derived parent/child patterns miss rule 370 (follow restriction) and open-rightmost rules 182–185, 337; (c) the 667 reduce resolutions are not expressed; (d) at completion, items still grow linearly; (e) Leo composition unstated. See also B10, D15, E12, A19. | `bison --report=all`: 1,347 resolved S/R conflicts (670 shift, 667 reduce, 10 error) over 57 rules (`F/conflict-rule-text.txt`); Earley probe `E→E+E\|E*E\|a`, items/token unfiltered/completion/stratified at 160 terms 201.9/103.1/4.0. | block | open | d6cb49c7 |
| F2 | F | 2 | §5.2 fitting families | Refinement under-sizes the port: variables/indirob, NOAMP calls, method-call tokens, scope-decided `*SUB` and PRIVATEREF, phasers and keyword plugins need fitting, and none is named; §7 maps a token to a set, never a sequence. See also B12; D18. | 35 of 348 alternatives reference these terminals (`F/rules-lhs.txt`); toke.c `force_next(BAREWORD)`, `TOKEN(NOAMP)`. | block | open | d6cb49c7 |
| F3 | F | 2 | §7 | The key is underspecified: "class of the next significant token" may be (kind, text) or a terminal set; "and similar" flags do not exist; pairs such as unary/binary `-` and `x` are not decided and become superpositions. | `lexer.Token` on pu has only `OpensBlock`, `TermFollows`; its comment says expect state is deliberately not exported; litlex output. | stand aside | open | d6cb49c7 |
| F4 | F | 2 | §5.2 refusals | The generator refuses a correct grammar. See also A20, B13, D14. | `F/pu/cmd/litlex`: alone `/`→UnknownRest, `/=`→UnknownRest, `%`→Error, `%=`→Variable, `**`→`*` + Word `*`, `&`→FuncSigil; in `$a OP 1` all Operator. | stand aside | open | d6cb49c7 |
| F5 | F | 2 | §5.2 port ledger | "Every perly.y alternative" has no mechanical anchor; the out-of-range refusal needs a range nothing reads. See also C9 (held), D16. | §4 "no build step reads it"; §9.2. | stand aside | open | d6cb49c7 |
| F6 | F | 2 | §5.1; §6 | The carried recognizer takes terminal names and interns with `fmt.Sprint` per edge; "works only in ids" needs a change to `lex.Edge` and `builder.scan` that §5.1 does not name. See also H2 (Q1). | `recognize.go` `builder.scan`, `Recognizer.class`, `cacheKey` construction in `Recognize`. | stand aside | open | d6cb49c7 |
| F7 | F | 2 | §9.3 | For std5k, long-lived equals cold, so "this is the first attempt's measurement" is wrong for std5k (the postmortem's figure was warm); alternation lets one side's GC land in the other's timing; `lex-overhead`'s numerator unstated. See also C12; E2 residue. | §9.3; postmortem "34.9 → 18.3 ms warm"; speed keys `na` in pu's baseline. | stand aside | open | d6cb49c7 |
| F8 | F | 2 | §4; §9.1 | Grammar version and label oracle disagree; blead syntax 5.42 rejects is accepted and only reported; a relabel moves the gates. See also D17. | 38 perly.y commits between v5.42.0 and 94e50866 (+641/−356), including PPC0024. | stand aside | open | d6cb49c7 |
| F9 | F | 2 | §13; §7 | The §7 key needs the next token, but interleaved mode runs at lookahead 0, so a positive up-front verdict does not transfer its token table to the next spike. | §7 key; §13 "lookahead 0". | stand aside | open | d6cb49c7 |

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
