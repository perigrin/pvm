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
| A1 | A | 1 | §5.3 Decouple; §5.4 | §5.3 sizes "decouple" as two functions; under that scope §5.4's "single visible deletion" does not compile, and the preparation issue's criterion passes while the coupling remains. | `BuildLR0` takes `Rule`/`Symbol`/`Seq`… defined in `slif.go`; `BuildTables`/`GenerateTables` call `Probes`/`FindConflicts`; `lex.New(*grammar.CutGrammar)`/`lex.Guacamole` is a third loader; `grep -cE 'ReadSLIF\|ApplyCut\|Guacamole…'` over carried tests: recognize_test 18, expect_test 16, lr0_test 11, leo_test 4, lattice_test 2, fuzz_test 1. | block | open | 083e1357 |
| A2 | A | 1 | §7.1; §9.2; §11 | Waypoint 1's ".tpm-declared infix operator in scope is accepted" cannot be met without changing `internal/parse`, which §11 forbids; waypoint 1 cannot exit. See also C2, D5. | `go doc -short ./internal/parse` exports no operator symbol/class accessor; probe `parse.DeclaredSignatures("sub ⊕ :infix(ADD) …")` returns no keys; `grep -rln ':infix(' internal/parse/declarations \| grep -v CORE.tpm` finds nothing. | block | open | 083e1357 |
| A3 | A | 1 | §7.1 | "Adds an entry … to the `PLUGIN_*_OP` terminal of its class" is undefined for some operators pvm accepts; a worker invents a mapping that can disagree with RD. See also B8, D5. | pvm has 15 classes (`classAnchors`); perl has 11 `INFIX_PREC_*` (`perl.h:6117-6127`) mapped by `tokentype_for_plugop`; `:tighter`/`:looser`/`:equiv` give levels with no PLUGIN token. | stand aside | open | 083e1357 |
| A4 | A | 1 | §4; §5.2 | Two counts or citations taken at a glance; one repeats an error the code warns about. See also B1, D8 (count); B2 (citation); E6 (reports 33 as checked). | `grep -cE '^%(left\|right\|nonassoc)' perly.y` → 32; `go doc ./internal/parse LevelsAccountedFor` warns of an earlier "33"; library-operator emit is at `scan.go:1269–1272`, not 1227. | stand aside | open | 083e1357 |
| A5 | A | 1 | §7.1; §7.3; §3 rule 7 | The (kind, text) key throws away decisions the lexer already makes; §7.3 calls `{` block-vs-hash a fact pvm lacks, but pvm has it; the chart may take the reading the lexer rejected. See also B4, D1. | `lexer.Token` carries `OpensBlock`, `TermFollows`; probe: `map {` ob=true, `my $h = {` ob=false, `sub f {` ob=true, `print $fh -1` tf=true on `$fh`. | block | open | 083e1357 |
| A6 | A | 1 | §3 rule 12; §7.3 | Rule 12 sorts every collapsing fact into grammar or scope; some toke.c decisions depend on the next character, and waypoint 2's resolve has no scope fact to settle them. See also B4 (ii). | `perl -MO=Deparse,-p -e 'print ($x), 1;'` → `(print($x), '???')`; `print` → {LSTOP, FUNC} yields two readings. | stand aside | open | 083e1357 |
| A7 | A | 1 | §5.1; §9.5 | "cmd/ratchet carried over" overstates what carries; §9.5's keys and gate rules cannot run on the carried comparator without changing the ratchet's own rule. | `Compare`: "a key is never removed or renamed"; §9.5 renames and drops keys; `Compare` has only `mayNotRise`/`mayNotFall`, no absolute threshold; `main.go` (448 lines) imports `chart/lex`, Guacamole paths, `parseoracle`. | stand aside | open | 083e1357 |
| A8 | A | 1 | §3 rule 10; §9.3 | "Speed is gated from the first recognizer issue" has no mechanism that fails; repeats the first attempt's failure. See also C6, D3. | `make test` is `go test ./... -count=1`, no `-bench`; a panicking benchmark leaves `go test ./internal/chart/probe_a/ -count=1` at `ok`; ratchet not in `make test`; `chart-ratchet.yml` `continue-on-error: true`; no `Benchmark` functions in internal/chart on pu. | block | open | 083e1357 |
| A9 | A | 1 | §7.4 | The 1.3 gate is feasible but sits within run-to-run noise; a ratio of two separately timed medians can flip with no code change. See also C5. | `lexov_test.go`, `-benchtime 30x -count 3` on std5k: Tokenize 15.9/19.9/18.6 ms; Tokenize+trivia+lookup 20.2/24.0/21.4 ms; ratio ≈1.15–1.27; 4 cores, load 2.6. | stand aside | open | 083e1357 |
| A10 | A | 1 | §5.2 GRAM* | "`Recognize` takes the start symbol as a parameter" is implementation detail, and more work than needed. See also D10; B7, D6 (start forms). | `BuildLR0(start string, …)`; `Recognize` seeds `{State: 0}`; perly.y:187–257 states start forms as terminals in one rule, so one GRAM* edge needs no API change. | stand aside | open | 083e1357 |
| A11 | A | 1 | §5.2 refusals | The refusal "checked against internal/lexer's own … operator tables" cannot read that table under §11. | `go doc -short ./internal/lexer` exports `Keywords`, `IsKeyword`, `IsFileTest`; `operators`, `bitwiseStringOps`, `positionalOps` (scan.go:1125–1143) unexported. | stand aside | open | 083e1357 |
| A12 | A | 1 | §5.2 example; §5.3 Q9 | (i) The example DSL would not compile while slif.go exists. (ii) Q9 is a real defect. See also D10 (i). | (i) `type Seq int` in slif.go collides with `func Seq`. (ii) Probe: `Star` with RHS `a b` builds with err nil as `S ::= ε \| S a`; `Plus` with empty RHS panics `index out of range [0]`. | stand aside | open | 083e1357 |
| A13 | A | 1 | §5.3 | §5.3 cites Q4 and Q7–Q11 to `## Records`, which define only Q1, Q2, Q3, Q5/Q6; the citation cannot be checked. See also D10 (sourcing). | `grep -n 'Q4\|Q7\|Q8\|Q9\|Q10\|Q11' A/body.md` → only the review line; postmortem has no Q-ids. | stand aside | open | 083e1357 |
| A14 | A | 1 | §9.1 | "`parse.Parse` returns no syntax error" names an interface that does not exist; `rd-accepted-chart-refused 0` can be computed two ways. See also C1. | `func Parse(src []byte) *Node` has no error result; acceptance is read from `Node.Refused()`; refusal codes include `UnimplementedStatement`, `CatchAll`. | stand aside | open | 083e1357 |
| A15 | A | 1 | §8 | The cross-check against "internal/parse's binding powers" cannot read them. See also B3 (cross-check). | `bindingPowers`, `deriveBindingPowers` unexported; `parse.LevelsAccountedFor() map[int]string` exported. | stand aside | open | 083e1357 |
| A16 | A | 1 | §7; §9.2 | The superset holds partly because quotes are opaque; interpolated code and `s///e` bodies are never recognized; not stated. See also B4 (i), B5, D1 (ii). | Probe: `"a $x b"`, `s/a/$b/g`, `qw(a b)` each lex as one `Quote`; 23 perly.y lines mention SUBLEXSTART, PMFUNC, QWLIST, FORMLBRACK, PERLY_SNAIL or HASHBRACK. | stand aside | open | 083e1357 |
| B1 | B | 1 | §4 | The port is ~15% more alternatives and ~70% more nonterminals than stated, so the 15-day budget rests on the wrong size; "33 lines" contradicts §8's 32-level cross-check. See also A4, D8; E6 (reports 33 as checked). | `bison -Wnone --report=all` (3.8.2): 373 rules = 348 + 24 mid-rule + `$accept`; 114 nonterminals; `grep -n '^%\(left\|right\|nonassoc\)'` → 32. | stand aside | open | 083e1357 |
| B2 | B | 1 | §5.2 | Stale citation. See also A4. | `git show origin/pu:internal/lexer/scan.go`: 1227 is `bitwiseStringOps`; symbol→Operator at 1269. | stand aside | open | 083e1357 |
| B3 | B | 1 | §8 | The superset gate fails on std5k.pl and on `$a + defined $b`-type code, or the implementer invents an unspecified rewrite mid-milestone. See also D2, E8; A15 (cross-check). | Stratified rewrite (`strat.py`) refuses `$a + defined $b`, `$a * defined $b + 1`, `$a + not $b, 1`, `my $s = join ",", @a`; perl Deparse and RD (`rdprobe`, 0 refused) accept; 110 `= listop` lines in std5k.pl. | block | open | 083e1357 |
| B4 | B | 1 | §7.1–7.3; §3 rule 8 | (i) Whole-kind Variable and Quote make every variable and string ambiguous; (ii) FUNC/LSTOP, FUNC1/UNIOP, NOTOP/FUNC1, autoquote and `{` are lexical, so waypoint 2's resolve cannot choose and gives WRONG sites; (iii) OB, TF and signature mode are thrown away. See also A5, D1; A6 (ii); A16 (i). | Table of 31 sampled tokens (rows 1, 3, 5, 6, 10, 18, 20, 28); `lx/dump`; `lx/amb` on T1: per-statement p99 16, >2^10 in 19 of 44,307 statements (0.04%), a floor. | block | open | 083e1357 |
| B5 | B | 1 | §5.2; §3 rule 7 | A cited rule-per-line port over internal/lexer refuses `{ 1 }`, `$h{a}`, `+{a=>1}` and end-of-file statements without `;`, and cannot express formats, labels, attributes, interpolation or postfix deref; the citation refusal pushes toward transcription. See also D1; D12 (HeredocBody trivia); A16. | `toke.c:6897` (synthetic `PERLY_SEMICOLON` before `}`), `toke.c:7670`; perly.y 1328, 1710, 854, 906, 1058, 1659; `lx/dump`. | block | open | 083e1357 |
| B6 | B | 1 | §7.4; §9.5 | `lex-overhead` passes while superposition inflates recognizer work; the gate measures the cheap half. See also D4. | Argued from §7.2/§7.3: lookup cost is constant per token; extra terminals cost items. Not measured. | stand aside | open | 083e1357 |
| B7 | B | 1 | §5.2; §9.2 | Non-program start forms lex in the wrong state; §11 forbids changing the lexer. See also D6; A10. | perly.y:189–259 GRAM* mid-rule actions set `parser->expect` to XTERM/XBLOCK; `checkpoint.go` exports only statement-start states. | stand aside | open | 083e1357 |
| B8 | B | 1 | §5.2; §7.1; §9.2 | "Accepted, as recursive descent accepts it" holds only for infix library operators in perl's 11 classes; RD's 15 classes, relation placement and prefix/postfix fixity are not expressible; waypoint 2 parity differs. See also A3, C2, D5. | `internal/parse/operators.go` `classAnchors`, `operatorArity`; `toke.c:9264-9289`; perl.h:6123 vs perly.y:168 on PLUGIN_REL_OP. | stand aside | open | 083e1357 |
| B9 | B | 1 | §5.2 refusals | "A precedence token is used by no rule": if "used" means a right-hand side, the generator refuses perly.y's own UMINUS and PREC_LOW. | `grep -n UMINUS perly.y` → 175, 1497, 1499 (`%prec` only). | stand aside | open | 083e1357 |
| C1 | C | 1 | §9.1; §9.2 | "No syntax error" is not a predicate that exists, and the gated population includes files perl rejects; the chart must accept non-Perl (against rule 5) or hold `rd-accepted-chart-refused` above 0; the gate measures RD's leniency and the lexer. See also A14; C10. | `parse.go:653` "no error return"; `zzpred` × `perl -c`: parses perl-ok/rd-accepts 743, ok/refuses 3, rej/accepts **18**, rej/refuses 5; parsent 75/3/**17**/29; `Refused()` agrees with Error/Missing on all 769 + 124 + 986 files. | block | open | 083e1357 |
| C2 | C | 1 | §9.2; §9.1 | "As recursive descent accepts it" is false under §9.1's own definition; a `:infix(ADD)` fixture passes before the general case works. See also A2, D5; A3, B8. | `parse.Parse` has a nil resolver (module.go:323); `use My::Lib; 1 plus 2;` → `zzpred` refused=true although `My/Lib.tpm:32` declares `sub plus :infix(ADD)`; `TestLibraryOperatorMayUseAnyLevel`. | stand aside | open | 083e1357 |
| C3 | C | 1 | §9.3; §10 | Long-lived as defined replays a cache warmed on the same files; the kill reads the most favourable figure, and the redefinition alone moves the line more than any fix did. See also E2. | `v1/internal/zzlong`, recognition only, median of 9 after 2 warm-ups: aggregate (763 files) per-run 19.1/21.5/29.0 ms, long-lived **10.6/9.6/8.8 ms**; std5k 45.7/58.2/48.4 vs 47.3/45.3/42.4 ms. | block (on the definition, not on the choice of long-lived) | open | 083e1357 |
| C4 | C | 1 | §9.3; §10 | "The aggregate" is undefined (corpora, statistic, std5k/edit5k inclusion, both-accept restriction); execute would define it; a kill key dominated by one generated file judges the template. | `zzspeed` rd-ms (speed-aa.txt): std5k 24–33 ms vs all 769 snippets 6–9 ms; `code_too_large.t` 484 KB of T1's 2,176 KB (22%), 4,998 near-identical lines; T1 RD total 420–470 ms. | block | open | 083e1357 |
| C5 | C | 1 | §7.4; §9.2; §9.5 | A wall-clock 1.3 gate cannot be decided under §9.3's protocol here; a correct chart/lex fails about a third of the time and a 30% regression passes about as often; ambiguous whether chart/lex's time includes Tokenize. See also A9; B6, D4. | A/A `Tokenize` vs itself: 0.65–1.32 std5k, 0.67–1.35 aggregate; thin chart/lex 0.91–1.39 / 0.90–1.44, over 1.3 in 4 of 12 trials each; interleaved A/A 0.76–1.19 (speed-aa.txt, speed-paired.txt). | block (on gating a wall-clock ratio; reporting it is fine) | open | 083e1357 |
| C6 | C | 1 | §3 rule 10; §9.3 | "Benchmarks … run in `make test`'s package tests" is false; nothing in `make test` goes red on a slowdown. See also A8, D3. | A panicking benchmark passes `go test ./... -count=1` ("ok bchk") and panics only with `-bench .`; `sed -n 21,22p Makefile`; ratchet CI advisory (Makefile:55-57). | block | open | 083e1357 |
| C7 | C | 1 | §10; §9.3 | "Checkpoint" is undefined and the protocol lets load drift into the ratio; near 3× the verdict depends on the orchestrator's reading. See also D11. | A/A of RD alone 0.90–1.23 std5k, 0.79–1.15 aggregate (speed-aa.txt); 2f46 `speed.go` `speedRatios` times the two blocks one after the other; v1 results 2.83–3.85. | block on defining "checkpoint"; stand aside on interleaving (issue-level) | open | 083e1357 |
| C8 | C | 1 | §2; §9.2; §9.4 | The waypoint-1 exit cannot answer §2's question (it bounds one side), and waypoint 2 never re-asks it; a parser 5× slower once the forest exists could pass. | Read §9.2–§9.4; `parse.go` `parseWith` (parseFile then finish). | block | open | 083e1357 |
| C9 | C | 1 | §9.2; §2 | Nothing measures completeness of `perl.go`, so the criterion passes before its work; "all of Perl" is asserted but unmeasured. | Every §9.2 key (tables-drift, refusals, superset, leo-disagree) is silent about a perly.y alternative no corpus file reaches. | block | open | 083e1357 |
| C10 | C | 1 | §9.1 | The corpus table's counts and labels mix verdicts; the denominator decides what the gates mean. See also C1, C4. | perl 5.42 accepts 78 of 124 parsent; RD refuses 3 perl-accepted parses and 3 parsent; `zzpred …/unit` files=986 refused=236 (24%). | stand aside | open | 083e1357 |
| C11 | C | 1 | §9.3 | Until Q3 lands, the cold figure measures `recognize.New`'s setup; the harm is misreading a reported key. See also E2 (cold). | longlived.txt: fresh recognizer per file over 763 snippets 1,097–1,283 ms vs 9–11 ms long-lived (~1.5 ms/file setup). | stand aside | open | 083e1357 |
| D1 | D | 1 | §3 rule 7; §5.2; §7.1; §7.3 | The port is a re-derivation presented as a transcription; rules restated in lexer vocabulary carry a `Cite` for a rule they do not transcribe; sigils cannot be keyed; OB/TF are discarded and §7.3 re-decides `{`. See also A5, B4, B5; A16. | `go test -overlay … -run TestDProbe ./internal/lexer` → `Variable«$x»`…`«$#x»`, `Quote«"a $x @{[ f() ]}"»`, `Word«foo» Operator«:»`, `Operator«{»^B` vs `Operator«{»`, `Variable«$fh»^T`; ~125 perly.y terminals vs 20 lexer kinds. | block | open | 083e1357 |
| D2 | D | 1 | §8; §5.2 | The rewrite is defined only for binary operators; most `%prec` uses resolve LALR conflicts with no stratified counterpart; the implementer invents semantics or the grammar stays ambiguous, against §5.2's "unambiguous expression grammar". See also B3, E8. | `sed -n 150,182p perly.y` → 32 directives; `grep '%prec' … \| sort \| uniq -c` → 21 `PERLY_PAREN_OPEN`, 12 `PREC_LOW`, 3 `UNIOP`, 2 `UMINUS`, 1 `LSTOP`. | block | open | 083e1357 |
| D3 | D | 1 | §3 rule 10; §9.3; §12 | "Speed is gated by benchmarks from the first recognizer issue" gates nothing; speed again goes unmeasured until the checkpoint. See also A8, C6. | `Makefile:21-22`; overlaid `b.Fatal` benchmark: `go test -count=1 ./internal/lexer` → `ok`; `-run '^$' -bench DProbe` → `FAIL`. | block | open | 083e1357 |
| D4 | D | 1 | §7.4; §3 rule 12 | The overhead gate measures the step v2 made cheap, not where v2 moved the cost; wider terminal sets show only in the single speed kill. See also B6; C5. | Argued from §7.2, §7.3 and postmortem figures (mapping 110 of 150 ms). Not measured. | block ("resolvable with one reported or gated key") | open | 083e1357 |
| D5 | D | 1 | §9.2; §7.1; §11 | A waypoint-1 exit criterion cannot go green under §11; unclassed and relation-placed operators leave a hole in the superset; the GRAM*/PLUGIN_* scope is not in perigrin's recorded decisions. See also A2, C2; A3, B8. | `grep -n '^func [A-Z]' internal/parse/{operators,declarations,precedence}.go` → no class accessor; `operators.go:15-30`; Records DECISIONS 2026-10-30 names neither; scope added in 399245b6. | block | open | 083e1357 |
| D6 | D | 1 | §5.2 GRAM*; §9.2; §2 | Non-program start forms cannot be lexed "exactly as for recursive descent" up front; the start-form test passes only with samples chosen so the difference never shows. See also B7; A10. | `perly.y:186-268` expect per GRAM form; `checkpoint.go:38-48` `StartState*` all `expect: XState`; `{ a => 1 }` as an expression lexes as a block. | stand aside if D5's resolution moves them out; block otherwise | open | 083e1357 |
| D7 | D | 1 | §9.2 | Superset and `lex-unmapped` gates disagree over a listed file RD accepts. | `internal/chart/testdata/lex-unmapped-known.txt` has 3 entries; whether RD accepts them under v2 is not measured. | stand aside | open | 083e1357 |
| D8 | D | 1 | §4; §14 | Counts taken at a glance; the 15-day budget is sized against a grammar about 70% larger than stated. See also A4, B1; E6 (reports 33 as checked). | `grep -cE '^[a-z_][a-z_0-9]*\s*:' perly.y` → 67; `awk 'NR>=184' … \| grep -cE '^[a-z_]…\s*$'` → 46; Python strip → 114 LHS, ~348 alternatives; lines 150–182 hold 32 directives. | stand aside on the figures; block on the budget only if perigrin sizes from them | open | 083e1357 |
| D9 | D | 1 | §3 rule 9; §5.2; §5.4 | Two sources of truth for builtin shape; drift surfaces as `rd-accepted-chart-refused` or as a silent misparse. | `internal/parse/declarations.go:711-723`: `CoreShapes` exists "for the chart's CORE.tpm cross-check"; §5.4 discards that cross-check. | stand aside | open | 083e1357 |
| D10 | D | 1 | §5.2; §5.3; §5.4; §7.1–7.2; §9.3; §12 | Implementation detail in a decision makes the accepted decision false on any rename; §9.2's `leo-disagree` gate is defined via a Q4 with no checkable source. See also A13 (sourcing); A10, A12 (i). | `grep -c 'Q9' D/records.md` → 0; no Q4 match; Records line 195 is the review summary. | block on the gate defined via Q4; stand aside on the rest | open | 083e1357 |
| D11 | D | 1 | §10 | "At any checkpoint" is undefined; whoever measures decides when a kill fires. See also C7. | The spec text. | stand aside | open | 083e1357 |
| D12 | D | 1 | §5.2 | Which kinds are trivia is unstated; mapping `HeredocBody` to a terminal set refuses every heredoc file. See also B5 (HeredocBody). | `lexer.go` kind comments (HeredocBody "begin after the line the marker sits on"). | stand aside | open | 083e1357 |
| D13 | D | 1 | §9.4 | The scope record can be corrupted by readings that die. See also E7. | The spec text, plus Earley semantics. | stand aside | open | 083e1357 |
| E1 | E | 1 | §2; §10; §13 | Both waypoint-1 outcomes lead to interleaved mode, so the verdict decides nothing about the approach; the kill is unfalsifiable for the approach as a whole. | `TestScratchExpectedEveryToken` on std5k: `Expected` at all 26,289 nodes median 430.97 ms (lookahead) / 94.24 ms (no lookahead) vs recognition 36.98 ms. | block | open | 083e1357 |
| E2 | E | 1 | §9.3; §10 | The long-lived kill figure replays a cache warmed on the same inputs; v1 and v2 define "cold" differently, so neither v2 figure compares with v1's 3.9–5.0. See also C3; C11 (cold). | `TestScratchAggregateColdVsLongLived`: second half (382 files) seen 4.98 ms vs unseen 11.03 ms; 763 files cold 2.90 s vs long-lived 9.65 ms; `TestScratchLadder` replay vs cold; std5k cold/long-lived 1.01. | block | open | 083e1357 |
| E3 | E | 1 | §1; §2 | The question no longer asks why pvm would want a chart; a positive verdict names no decision it unlocks. | `grep -n 'prior-art\|A1.3' docs/plans/2026-10-30-chart-parser-v2-design.md` → nothing; `git zhi milestone list`: m2-incremental open at 30/38. | block | open | 083e1357 |
| E4 | E | 1 | §13; §5.3 Q2; §9.4 | §13's M2 payoff contradicts spec ch. 6 §6.3.1 and is the stated justification for Q2 and the scope-record interface. | `sed -n 881,925p docs/specs/perl-parser/06-incremental-lsp.md`; `sed -n 975,1000p docs/specs/perl-parser/03-lexer-feedback.md`; `internal/parse/incremental.go:564` `lexer.NewTokenCache`. | stand aside | open | 083e1357 |
| E5 | E | 1 | §9.2; §9.3; §14 | The first reading the kill can act on comes after most of the port, the largest spend. | `TestScratchLadder` (toy, cold): 1 level 2.5 items/token 4.57 µs; 8: 4.8, 3.14; 25: 9.2, 6.09; 32: 10.9, 6.67. | stand aside | open | 083e1357 |
| E6 | E | 1 | §3; §5; §7; §9; §10 | Postmortem items mostly acted on; item 1 (component fit) is only partly acted on and left to lens (a). Reports "33 precedence lines" as checked. See also A4, B1, D8 (count 32). | Table of the seven items; `awk 'NR<=182 && /^%(token\|left\|right\|nonassoc)/' perly.y \| … \| sort -u \| wc -l` → 124 token names vs 20 lexer kinds. | stand aside | open | 083e1357 |
| E7 | E | 1 | §9.4; §13 | Waypoint 2's scope record is designed now for a mode out of scope, with an open question about which reading's completion updates it. See also D13. | The spec text. | stand aside | open | 083e1357 |
| E8 | E | 1 | §8 | §8's schemes do not cover the non-operator roles perly.y's precedence lines play; an inexpressible level leads to the twice-amended-definition stop. See also B3, D2. | `sed -n 148,184p /opt/corpora/perl5/perly.y`; precedence split between `expr` (perly.y:1256-1260) and `term`. | stand aside | open | 083e1357 |

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
