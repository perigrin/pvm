<!-- ABOUTME: Minute of the crochet:assess session on the chart-parser research spike design. -->
<!-- ABOUTME: Covers rounds 1 to 6: declared outcomes, rulings, author calls, the findings ledger, clusters, and what was not checked. -->

# Assessment: chart-parser research spike design

**Subject:** `docs/plans/2026-10-29-chart-parser-design.md`
**Revisions assessed:** round 1 at 4ec0955a (4ec0955a58637fc7b69fd41414fe53c7c8cc0665); round 2 at 47951e0a; round 3 at b5b10428 (b5b10428f51ff8a2caf4f3d5b65410eb8700a165); round 4 at e34f8233 (e34f8233a2582c9f6749b03f25448e661083a10f); round 5 at 7a430aa4 (7a430aa43f21a012ae01c3c0955005fc0522ea88); round 6 at 723440f1 (723440f115a34f1d4ef0bde74796ccfc555270a5). Round 7 is briefed on f82e7dd9. Branch `design/chart-parser-spike`.
**Author:** the orchestrating AI session. **Decision-maker:** perigrin.
**Drafter:** a recording role that holds no view. It raises no finding.

Section and phase references in each row use the numbering of the revision that row was raised against.
- At 47951e0a, R1 renumbered the phases: phase 2 became the early measures, phase 3 the M4–M6 probes, and phase 4 full parity.
- At b5b10428, the round-2 response renumbered them again: 0 measure, 1 Standard Perl, 2 probes, 3 parity, 4 incremental.
- At e34f8233, the round-3 response says "Phases are gone": §8 runs a waypoint, then the M4–M6 steps, then a checkpoint, then the steps to parity, then incremental.

Participant F (round 4) uses IDs F-1 and up, which are distinct from C's F1 to F13. Participant G (round 5) uses IDs G-1 and up.

## Outcome of round 1

### Recommendations, as declared

- **A (codebase and architecture): modify.** "The document can be fixed without changing its shape." The parity contract (§7), which every phase gate depends on, "measures much less than it claims". The phase-1 corpus is not "valid" in the sense the gate needs. Two claims about pvm's lexer don't hold against the code.
- **B (parsing theory and feasibility): modify.** "The overall shape holds up." Three load-bearing technical claims are wrong or rest on an unstated precondition, and each affects phase 1 directly. "None of these is fatal to the spike."
- **C (spec and plan quality): modify.** The question (§2), the rules (§3) and the isolation boundary (§5) are sound. "The plan around them is not bounded." The fixes are "to the plan's structure, not to its idea".

### Blocks, as declared

| participant | blocks |
|---|---|
| A | A1, A2, A3, A4, A5 (on the isolation claim only), A7 |
| B | B1, B2, B3 |
| C | F1, F2 (resolved by the F1 reorder), F3, F4, F5, F6 (on parts (i) infrastructure accounting and (iv) undefined thresholds only), F7, F8, F13 (dissolved by the F1 restructure) |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A5 (perly.y detail) | Phase 1's to measure. |
| A6 | The terminal encoding belongs to the implementing issue. The CORE.tpm dependency should be one sentence in §5. |
| A8 | Cost, and the issues own it. §8 should still add require/use loading, and §5 should drop "existing ratchet harness" for anything except the oracle. |
| A9 | Fine for the spike. §2 Output should say the M6 verdict excludes re-emission, or §8 should cost a chart canon. |
| A10 | §11 already says it. §2 should be reworded to "the parse half of M2". |
| A11 | Small. The conclusion still stands. Better to state the RD baseline. |
| A12 | Implementation. The issue should say where the chart's ratchets do run. |
| A13 | No consequence if the spec states it. It supports the spike. |
| A14 | Doesn't block: the bracket diff can run on about 99% of phase 1. The 9 should be listed as one-parser cases. |
| A15 | The label map belongs to the `chart/compare` issue. §10 should name `internal/lexer` and `chart/compare`'s import of `internal/parse`. |
| B4 | Stand-aside, but correct the premise before phase 1. |
| B5 | Define the class and the comparison. |
| B6 | Add a §10 test that fails if Leo is a no-op. The side-table mechanism goes to the implementing issue. |
| B7 | State "post-hoc, by search over `pl[]`" as a constraint. The intermediate-node memo goes to the implementing issue. |
| B8 | Make Aycock–Horspool states a phase-3 measured option, not a phase-1 commitment. Belongs to the implementing issue. |
| B9 | Restate as a hypothesis for phase 3 to test. |
| B10 | Define "region" and the counting method. |
| B11 | Specify the normalisation (label map, unary-chain collapse, list flattening, byte spans) before phase 1. |
| B12 | Belongs to the implementing issue, but "terminals are predicates" is a design statement. |
| B13 | Wording only. No further grounds given. |
| F6 (ii), (iii), (v), (vi) | The implementing issue's concern. |
| F9 | A human verdict is legitimate for a research spike (§2 "the verdict is perigrin's"). It should still be written down before the numbers arrive. |
| F10 | A one-word fix: "T1 at or above its ratchet". |
| F11 | Belongs to the implementing issue. The spec only needs to say the extracted corpus is checked in and reviewed. |
| F12 | The how belongs to the implementing issue. The spec should state the property: phase gates enforced on the spike branch or tag, invisible to mainline `make test`. |

### Round outcome

**Another round owed.** Blocks are held by A, B and C.

## Rulings

Rulings are perigrin's, as recorded or relayed by the author. They are not findings and they release nothing. Only the raiser of a finding can release it.

The round-1 response says perigrin "ruled on five questions on 2026-10-29", but it lists six rulings. All six are recorded here.

### From the round-1 response (R1–R6)

- **R1, phase order.** After phase 1: early measures (phase 2) and the M4–M6 probes (phase 3). Full parity (phase 4) runs only if both verdicts are positive.
- **R2, seed.** Guacamole v0.008's grammar directly, not Chalk's perl.bnf@00399607.
- **R3, arbiter.** `perl -MO=Deparse,-p` decides structural disagreements.
- **R4, lexer.** Refactor `internal/lexer` so scanners take their expectation as a parameter. This is mainline work.
- **R5, notation.** The subset of Marpa SLIF that Guacamole uses replaces Chalk's notation. Guacamole has no SLIF prioritised (`||`) rules; its two `||` occurrences are lexeme strings.
- **R6, labels.** `parsent` means "not Standard Perl", so its labels are correct by definition. Only a `parses` case perl rejects is an erratum.

### From the round-2 response (R7–R10, 2026-10-29)

- **R7, probes.** Each phase-2 probe builds the Standard-Perl exclusions it needs, charged to that probe. The milestone-corpus head-to-head moves to phase 3 (parity). The response gives this as its answer to A16, C-N1 and D3.
- **R8, verdicts.** A positive M4–M6 verdict alone unlocks parity. A negative incremental verdict means M2 keeps its snapshot design and phase 4 is skipped. The response gives this as its answer to D6.
- **R9, phase 0.** A cheap phase 0 measures first: recognition only, over Guacamole's G1, on tokens from today's production lexer. The lexer refactor (R4) and interleaved lexing come in phase 1, only if phase 0 passes. The response gives this as its answer to B-N3, D1, A17 and D19.
- **R10, token cut.** G1 rules below pvm's token level are collapsed into token predicates, listed and reviewed. The response gives this as its answer to B-N2 and B12.

### After round 3 (R11–R12, 2026-10-29)

perigrin made these after round 3 was complete. They are given in his words as relayed by the author. They were made after the round, so no participant assessed them in it.

- **R11.** Standard Perl stays only as an interstitial state, the first waypoint. The spike is one ratchet towards parity, with speed measured at every step. The speed kill: the aggregate chart-to-recursive-descent ratio, end to end, on every file both parsers accept, above 3× at any ratchet step. The trend is recorded throughout.
- **R12.** Ratchet steps are the probes. Each step is costed on both parsers. M4–M6 cases are scheduled as early steps. The verdicts are read off step costs, with a checkpoint after the first M4–M6 steps.

The round-3 response restates R11 and R12, and says §8 was rewritten around them. Round 4 assessed e34f8233, which carries that rewrite.

### After round 4 (R13–R15, 2026-10-29)

perigrin made these after round 4 was complete, and they are relayed by the author. They were made after the round, so no participant assessed them in it.

- **R13.** Before the checkpoint, speed and convergence also run on a generated Standard-Perl file of several thousand lines. pvm's existing fixture generator makes it, with Standard Perl's excluded features written out. It is validated by both parsers accepting it and by perl compiling it. The ratchet order is unchanged.
- **R14.** The Deparse arbiter is dropped. perigrin's reasoning is that pvm is a linter and LSP, not an interpreter, so pure operator-precedence grouping that affects only evaluated values doesn't matter. Parity gates on the oracle. Other span disagreements are recorded, not gated.
- **R15.** The oracle gains a sixth marker for argument structure, taken from perl's op tree: the argument count per call, and filehandle-or-not per print/say. It is mainline oracle work, done before the M4–M6 steps, and it grades both parsers.

The open rows that name §7.4, the arbiter R14 drops, are A26, D4, E18, E20, F-1, F-2, F-3 and F-4.

The round-4 response restates R13–R15. In it, R13 names the generated file std5k.pl, from `latencyFixture`. It also repeats the author's correction about call sites.

### During round 5 (R16, 2026-10-29)

perigrin made R16 during round 5. B, C, D and E had already raised the repetition finding when it came in.

- **R16.** std5k.pl's repetition (one template about 70 times, raised by B's N15, C's C-N23, D, and E's E25) is accepted. perigrin's reasoning: the bias is caught later, when perl.git `t/` is parsed, because the speed kill fires at any step including the parity steps. §8 records std5k.pl's ratio as optimistic (a lower bound) until real files join. The author named a residual cost and perigrin accepted it: a false pass at the checkpoint costs the parity steps built before `t/` exposes it, bounded by the per-step budgets.

R16 is recorded against B-N15, C-N23, D-N7 and E25 as a ruling made after those rows were raised. It is not a release, because only a raiser can release a row. F-16 and G-4 also raise std5k.pl's repetition, but R16 does not name them.

### Author calls (not rulings)

An author call is the author's decision on a point perigrin has not ruled on. perigrin may overturn it. It is recorded here so the record is complete, and it is not a ruling.

- **Argument-marker parity (round-5 response).** The author wrote: "on the argument marker, parity is the chart's WRONG at or below recursive descent's, not WRONG = 0. Recursive descent has known attachment defects, so WRONG = 0 would make parity stricter than the floor it is measured against (C-N26, A30)." Round 6 assessed this call at 723440f1. The author's withdrawal of it is recorded under "Outcome of round 6".

### Facts the author checked

- **On D4's example (round-3 response).** On origin/pu ad7faf79, recursive descent parses `${$h->{k}}` as `unary` over `index($h, k)`. Issue 01a0ad52, which described the childless leaf, is done. The author says this "bears on D4's example, not on its argument". It is recorded on the D4 row and is not a release.
- **perly.y (round-3 response).** `/opt/corpora/perl5/perly.y` has 18 occurrences of `parser->expect`: 14 writes and 4 `==` tests. §12 says so at e34f8233.
- **Behind R15.** The oracle's call_sites protocol carries only the five existing markers (srefgen, match, anonhash, readline, rv2hv), graded per statement against perl's op tree (`internal/parseoracle/README.md`, `compare_facts.go`). It does not grade argument count or print filehandles. The author had said earlier that call sites covered argument attachment, and the author now records that this was wrong.

## Outcome of round 2

### Recommendations, as declared

- **A: modify.** "The revision closes most of round 1." Three things still stand: the arbiter has no comparison procedure (A2), the lattice rejoin rule ignores lexer state (A7), and the R1 reordering broke §2's evidence 2 (A16).
- **B: modify.** "Most of round 1 is answered properly." Two of the fixes create new defects, and both block (B-N1, B-N2). A third, B-N3, is an internal contradiction about what drives the default up-front lexer.
- **C: modify, now narrow.** "The restructure (R1) is the right shape." Three blocking defects remain, "all three are short fixes to the text" (F2, F9, C-N1).
- **D: modify.** "The question is worth asking, and most of the pvm facts the doc cites check out." Four gates cannot be evaluated as written or would mislead, and the default lexing mode drops the property §1 gives as the spike's reason. "Fixable by editing the document, with no change to its direction."

### Blocks held, as declared

| participant | blocks |
|---|---|
| A | A2 (re-raised), A7 (re-raised), A16 |
| B | B-N1 (re-raising B1), B-N2 (re-raising B12), B-N3 |
| C | F2 (re-raised), F9 (re-raised; its hot-spot part is a stand-aside), C-N1 |
| D | D1, D2, D3, D4 (the missing component and the undecided ceiling only), D5, D6 |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A4 (re-raised) | Name the bucket for the 38 module-missing cases: stub modules on `@INC`, or "accepted and graded by span diff only". |
| A6 (re-raised) | Say that the generator imports `internal/parse` at generate time, or say how CORE.tpm is read otherwise. |
| A17 | Say what drives up-front classes. Ordering the R4 refactor after phase 1's class comparison is perigrin's call, since R4 is his ruling. |
| A18 | Name the phase-2 inputs now, with RD timed on the same files. |
| A19 | Small. A cited count a reader would re-run. |
| A20 | Add the trivia attachment rule to the normalisation fixed before phase 1. |
| B10 → B-N7 | Pick the counting point and write it down. |
| B-N4 | The implementing issue's to get right. |
| B-N5 | No further grounds given. |
| B-N6 | No further grounds given. |
| F4 (re-raised) | Give the direction of "does not count towards parity" and a cap. |
| F6 (ii), (iii), (v), (vi) | Stay with the implementing issue, as agreed. F6 is otherwise released. |
| F7 (re-raised) | "Continue" must record a new budget so the 2× rule applies again; the calendar estimate governs where the two disagree. |
| F9 (hot-spot part) | Secondary: "a kill unless removing the named hot spots brings the ratio under 3×". |
| C-N2 | Restore the definitions of measures (b) and (c). |
| C-N3 | Reword the question, or state that phase 3 answers only the M4–M6 and incremental clauses. |
| C-N4 | perigrin's verdict is legitimate; restore "a fired kill ends the spike unless perigrin records an override and its reason". |
| C-N5 | State that the refactor counts in phase 1's budget, and either that it stands on its own merits or that a kill reverts it. |
| C-N6 | Say "resolvable by facts perl had; some are obtainable by pvm only by execution, and those stay ambiguous". |
| C-N7 | State the disposition of the module-dependent cases. |
| C-N8 | Say "at or above recursive descent's measured score on the same extracted corpus, recorded when it is extracted". |
| C-N9 | No further grounds given. |
| D7 | Say the generator, at go-generate time, may import `internal/parse`. |
| D8–D11, D13–D17, D19, D20 | No further grounds given. |
| D12 | How belongs to the implementing issue. |
| D18 | Branch mechanics belong to the implementing issue. |

### Round outcome

**Another round owed.** Blocks are held by A, B, C and D.

## Outcome of round 3

### Recommendations, as declared

- **A: modify, narrowly.** "Round 2 left three blocks: A2, A7 and A16. All three are now met." The revision opens one new block, A21. "The rest is stand-asides."
- **B: modify.** "The rejoin rule, the up-front driver and the collapse itself now hold." The revision introduced two problems: the phase-0 fixed input is not Standard Perl (B-N11, block), and the token cut drops Guacamole's lexeme priorities (B-N10, stand-aside "but it must land before phase 0").
- **C: modify.** "Most of round 2 has landed." Three blocks remain or are new: F2, C-N10 and C-N11.
- **D: modify.** "14 of my 20 rows are released." Three blocks stand in narrower form (D4, D5, D1) and the revision introduced one more (D-N1). "Each needs a paragraph of edits. None changes the spike's direction."
- **E: modify.** "The question is worth asking, and most of the document's references hold up." Two parts cannot do their jobs as written: the arbiter (§7.4) and phase 0. "All of these can be fixed inside the current shape."

### Blocks held, as declared

| participant | blocks |
|---|---|
| A | A21 |
| B | B-N11 |
| C | F2 (held, changed form), C-N10, C-N11 |
| D | D1 (re-raised, narrowed), D4 (re-raised), D5 (re-raised), D-N1 |
| E | E1, E2, E3, E4, E5, E6 |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A22 | Add a re-time of the full cold parse at phase 1's gate, against the same 3×. |
| A23 | One line in the token-cut list. |
| A24 | Drop the parenthesis or replace it with "new: Deparse of canon against Deparse of source". |
| A25 | Harmless now. The counts should stand on their own procedure. |
| B-N7 (re-raised) | Count over a fixed unit, e.g. the product of readings over all packed nodes inside one statement node. |
| B-N8 | Implementing issue. |
| B-N9 | No further grounds given. |
| B-N10 | Stand-aside, but it must land in the cut's list before phase 0. |
| B-N12 | No further grounds given. |
| C-N12 | State that a positive incremental verdict without a positive M4–M6 verdict ends at the recommendation. |
| C-N13 | perigrin may accept the tilt. If so, say it; otherwise report probe cost with and without the feature-building share. |
| C-N14 | Say "after optimising the named hot spots, re-timed". |
| C-N15 | Restore a hard cap that collapses to an Error node, independent of decidability. |
| C-N16 | No further grounds given. |
| C-N17 | Say where the code lives, and that `make test-chart` runs on any change under `internal/lexer`. |
| D3 (residue) | Released. The residual bias needs no action: it is conservative, and perigrin ruled it. |
| D14 (re-raised) | Recount 68/N on the phase-1 corpus as defined, or drop the number and keep the argument. |
| D-N2 | "Stubs declare a `$VERSION` satisfying every request." |
| D-N3 | State that phase 4 requires phase 3, or give phase 4 a Standard-Perl corpus. |
| D-N4 | Belongs to the implementing issue for the token cut. |
| E7 | Add error recovery to phase 1 Build. |
| E8 | Matters at phase 3. |
| E9, E12, E15, E16, E17 | No further grounds given. |
| E10 | Partly an implementing issue. |
| E11 | Reword rule 3 to admit perl's deterministic lexer heuristics. |
| E13 | Implementing issue. List it as a third isolation exception. |
| E14 | Implementing issue. |

### Round outcome

**Another round owed.** Blocks are held by A, B, C, D and E. Rulings R11 and R12 came after this round, and no participant has assessed them.

## Outcome of round 4

### Recommendations, as declared

- **A: modify.** "My open items, A21 to A25, are met. The restructure into one ratchet fits better than phases did." Two things block: A26 and A27.
- **B: modify.** "The token-cut priority rule, the rejoin state and the conflict list now hold, and I release them." Before the checkpoint every input is too short for either failure mode to appear (B-N13, block, re-raising B-N11). "The per-statement cap as written still collapses whole subs" (B-N14).
- **C: modify, with two narrow blocks.** "Most of the structure is sounder for it." The two blocks are "both short text fixes": F2 and C-N18.
- **D: modify.** The restructure "settles D1 and D-N1 properly, and the two-sided arbiter is the right shape". Two blocks remain, "both narrower than before": D4 and D5. "Each fix is a paragraph."
- **E: modify.** "The revision fixes what I blocked on." The restructure "has moved two problems rather than removed them, and has added one": E18 and E19 (blocks), plus stand-asides.
- **F: modify.** "The question is a good one", and the isolation and kill structure are sound. "But the two instruments that decide the outcome do not work as written": the arbiter (§7.4) and the checkpoint (§8).

### Blocks held, as declared

| participant | blocks |
|---|---|
| A | A26, A27 (A27 carries A18, re-raised) |
| B | B-N13 (re-raising B-N11) |
| C | F2 (held, changed form), C-N18 (re-raising C-N13) |
| D | D4 (held, narrowed), D5 (re-raised, narrowed) |
| E | E18, E19 (E19 carries E12, re-raised) |
| F | F-1, F-2, F-3 ("folds into F-1/F-2 rewrite"), F-6, F-7 |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A28 | R11 is perigrin's later ruling, so this is his trade. The document should say R11 supersedes R1's early kill. |
| A29 | State it as intended, or keep chart unit tests behind the tag. |
| B-N14 (re-raising B-N7) | Count over the innermost statement, and use an exact sum-product. |
| C-N19 | Also record each step's ratio on the previous step's file set. |
| C-N20 | Draw the edits from a multi-statement file, or state that the incremental verdict is provisional until parity. |
| C-N21 | Give the waypoint sub-estimates per component, and give the checkpoint a cumulative estimate. |
| C-N22 | Implementing issue: say how the PR set is chosen, and record the list. |
| D19 (re-raised) | Let the speed kill fire as soon as the recognizer runs on production tokens over D5's long input, within the waypoint. |
| D-N5 | D5's long input also serves the convergence measure. |
| D-N6 | Name the runner, and say the tax is accepted. |
| E1 (residue) | Released. Not re-raised: say whether "both right" shape differences count in the denominator. |
| E8 (held) | Name a measure for covered-ness in ambiguous statements, or say that the M4–M6 verdict does not rest on it. |
| E17 (held) | "Places touched" is undefined; grammar lines are compared with Go lines; who judges "a rule no Perl construct justifies" is unassigned. |
| E20 | Implementing issue for the emitter; (b) to be listed as a known-undecidable class. |
| E21 | Add set reuse to the every-step list. |
| E22 | No further grounds given. |
| E23 | Restrict the exclusion to a call's first segment. |
| F-4 | The how belongs to the implementing issue; the what (an emitter whose output perl compiles to the same op tree) belongs in the design. |
| F-5 | For a multi-reading statement, check that some live reading matches perl, otherwise record WRONG. |
| F-8 | Fix with F-7's file. |
| F-9 | Count only the states whose conflicts are in force, and repeat as conflicts join. |
| F-10 | Harmless for scoring, since both kinds are excluded. |
| F-11 | A minimum count, or a count instead of a ratio; the product definition belongs to the implementing issue. |
| F-12 | No further grounds given. |
| F-13 | Implementing issue that introduces interleaving; the doc should say which side owns those facts. |
| F-14 | Sizing belongs to refinement. Split the waypoint into a recognizer-only speed probe, then the full stack. |

### Round outcome

**Another round owed.** Blocks are held by A, B, C, D, E and F. Rulings R13, R14 and R15 came after this round, and no participant has assessed them.

## Outcome of round 5

### Recommendations, as declared

- **A: accept, with stand-asides.** "None of my rows block." A26 is moot under R14, and A27 is answered by R13's std5k.pl. A30–A32 ask the text to state what R14 and R15 cost.
- **B: modify (narrowly).** "The blocks I raised are answered." B-N15 is "a measurement caveat with a one-sentence remedy, not a structural defect. If the author records that remedy, I would accept."
- **C: modify, with one narrow block.** "R13, R14 and R15 are implemented faithfully, with two omissions in §10's isolation list." The block is C-N23.
- **D: accept.** "I hold no blocks." D raises four stand-asides, D-N7 to D-N10, "each a sentence or two of text".
- **E: accept, with stand-asides.** "None of my rows remains a block. … None of these stops the spike from producing the evidence §2 asks for."
- **F: modify.** "One thing still blocks (F-15)."
- **G: modify.** "The architecture is coherent, and I found no conflict with pvm's decisions. The first step can start from this document." Two measures are not yet well defined, and the realistic input flatters the chart. "Each needs a paragraph, not a redesign."

### Blocks held, as declared

| participant | blocks |
|---|---|
| A | none |
| B | none |
| C | C-N23 (re-raising C-N20) |
| D | none |
| E | none |
| F | F-15 |
| G | G-1, G-2 ("folds into G-1"), G-3, G-4 |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A30 | Say the marker can be hedged; RD lands wider, not WRONG, where it can't decide; put its mainline cost in the pre-checkpoint ratchet budget. The definitional points belong to the marker's own issue. |
| A31 | Add it to §10's exceptions, or say the template is copied. |
| A32 | One sentence in §7 or §2 Output: parity and the M4–M6 verdict do not cover grouping-only differences. |
| B-N15 | Record transition-cache hit rate and set reuse beside std5k.pl's ratio and time once with the cache disabled, or vary the blocks. |
| F2 (re-raised, narrowed) | Downgraded because the std5k.pl arm is well defined and the override exists. Define the aggregate as a ratio of total times, or kill on std5k.pl alone before the checkpoint. |
| C-N21 (held, narrowed) | Record a whole-leg estimate for the parity leg at the checkpoint, under the same 2× rule. |
| C-N24 | Add both to the exceptions, as ordinary mainline work. |
| C-N25 | Bring the argument marker forward to the waypoint, or say the waypoint exit is graded on about 9% of cases plus acceptance. |
| C-N26 | Implementing issue: the kind lands with RD's measured count as its ratchet. The spec should say whether WRONG = 0 or "at or above RD" applies to the new kind. |
| D-N7 | Write std5k.pl from several templates, or also report its ratio with the transition cache off, or on its first block alone. |
| D-N8 | List five exceptions, or move std5k.pl's generator out of `psc`. |
| D-N9 | Define the count as direct operands, with a nested list counting once; give sites a per-call key; RD WRONG from the new marker is mainline defect yield. How belongs to the R15 issue. |
| D-N10 | One clause in rule 5: "as graded by the oracle (§7.1); elsewhere, disagreements are recorded (§7.3)". |
| E11 (re-raised) | A one-word fix flagged in round 4 and not taken. |
| E17 (held, minor) | One word fixes it: sum or each kind. |
| E24 | List both, or say std5k's generator is copied into `internal/chart`. |
| E25 | Record the transition-cache hit rate beside std5k's ratio, or vary block shapes. |
| E26 | Say which; if surplus, scope the sentence to the five existing markers. |
| F-5 (held) | One clause in rule 4 tying WRONG to §7.1's grading. |
| F-16 | Several block templates or a ratio beside one unique block; chart acceptance of std5k.pl an exit criterion of the first step. Generator details belong to the implementing issue. |
| F-17, F-18 | No further grounds given. |
| G-5, G-6, G-7, G-9, G-10, G-11 | No further grounds given. |
| G-8 | Implementing-issue detail. |

### Round outcome

**Another round owed.** Blocks are held by C, F and G. R16 came during the round, after the repetition rows were raised, and it released none of them.

## Outcome of round 6

The round-6 brief asked each participant for one status per open row, any new rows, and, optionally, a view on the author call. It had a 20-minute time box. Rulings were not to be reopened, though participants could say whether the text implements them faithfully.

### Recommendations, as declared

- **A: accept.** "My three open rows (A30–A32) are met." A33 is a view on the author call, given as a stand-aside. "The revision breaks nothing I can find in my lens."
- **B: accept.** "My lens has nothing left that blocks." B-N16 is a view on the author call.
- **C: modify, with no blocks.** "Nothing left in my series stops the spike from starting." The argument-marker definition has two gaps, C-N27 and C-N28, which are stand-asides because "§10's test cases show what is intended".
- **D: accept.** "All five of my round-5 stand-asides are carried out in the text." D-N11 and D-N12 are stand-asides. "I hold no blocks."
- **E: accept, with stand-asides.** "All my open rows are released." The marker definition has three defects, E27–E29: "None of these blocks."
- **F: accept.** "None of my open rows blocks any more." Four stand-asides remain.
- **G: modify (narrowly).** "Ten of my eleven rows are released." G-1 is re-raised in narrower form as G-12 and G-13. "Two sentences fix it."

### Blocks held, as declared

| participant | blocks |
|---|---|
| A | none |
| B | none |
| C | none |
| D | none |
| E | none |
| F | none |
| G | G-12, G-13 (with G-1 re-raised through them) |

### Stand-asides, as declared, with grounds

| id | grounds given by the participant |
|---|---|
| A33 | perigrin's to rule on; A would take "subset of RD's WRONG sites" over "count at or below". |
| B-N16 | A view on the author call; perigrin's to rule. |
| F2 (held) | Define the aggregate as total chart time over total RD time, or apply only std5k.pl's arm before the checkpoint. |
| C-N21 (held) | Record a whole-leg estimate at the checkpoint, under the same 2× rule. |
| C-N27 | Add "and each call of a builtin that has a special first slot" to the site list. |
| C-N28 | Grade a subject site with no perl counterpart only where perl cannot have folded it; matching does not depend on order alone. How belongs to the implementing issue. |
| C-N29 | The author's call; C recommends "every site the chart gets WRONG, recursive descent also gets WRONG". |
| C-N30 | One sentence: the convergence figure on std5k.pl is optimistic, and the incremental verdict is provisional until re-measured after parity. |
| C-N31 | Scope it as "WRONG > 0 on the five older kinds, or above recursive descent's on argument structure". |
| D-N11 | "Less `pushmark` or `padrange` and the callee; nulled (`ex-`) children count"; add a lexicals-only case to §10; say the invocant counts. |
| D-N12 | "One site per sub, `&`- or method call, and per `print`, `say`, `sort`, `map` and `grep`". |
| E27 | Add "and each `print`, `say`, `sort`, `map`, `grep`" to the site set. |
| E28 | Implementing issue: count tree children including nulled list ops, less `pushmark` or `padrange` and the callee; add a padrange case to §10. |
| E29 | Scope the waypoint's "no WRONG site" to the five older kinds, or state the order. |
| F-16 (held in part) | Make the chart's acceptance of std5k.pl an exit criterion of the first step. |
| F-18 (re-raised) | Add "and the builtins `print`, `say`, `sort`, `map`, `grep`" to the site list. |
| F-19 | The what belongs here (a missing or extra site is WRONG); matching and the invocant convention belong to the implementing issue. |
| F-20 | Scope §9's "WRONG > 0" and §8's "no WRONG site" to the five older kinds, or to sites where RD is not WRONG. |
| G-14 | An author call; perigrin's to make. |

### Views on the author call, as given

Each participant's view is recorded in its own words. They are not tallied.

- **A (A33):** "I'd take 'subset of RD's WRONG sites' over 'count at or below'." A cites `TestSiteRatchetDetectsSwapAtEqualTotals`.
- **B (B-N16):** a set rule: "every site that is WRONG for the chart is WRONG for recursive descent too, and any other chart WRONG is a finding".
- **C (C-N29):** "accept the relaxation, but make it per site": "every site the chart gets WRONG, recursive descent also gets WRONG".
- **D (in its report, not as a row):** "every site the chart gets WRONG must also be WRONG for recursive descent, with the chart's count at or below recursive descent's".
- **E (in its report, not as a row):** E agrees with the direction, and would keep the count gate with one sentence added: "every site WRONG on the chart and right on RD is listed and reviewed at each step".
- **F (in its report, not as a row):** F agrees, with one tightening: "Chart WRONG ≤ RD WRONG, and no site WRONG for the chart that RD gets right".
- **G (G-14):** relaxing to the floor is right, but "the site-level form costs nothing extra: the chart's WRONG sites are a subset of RD's, per file and site".

### The author's action after round 6

The author's round-6 response withdraws the author call. It says: "A, B, D and G independently proposed judging parity site by site, and the revision takes that: on the argument kind, every site that is WRONG for the chart must be WRONG for recursive descent too, per file and site. Any other chart WRONG is a finding, and an RD WRONG the chart gets right is defect yield." The author says this "narrows the earlier call rather than reversing a ruling", and that "perigrin may still rule otherwise".

The response moves the subject from 723440f1 to f82e7dd9, which goes from 495 to 505 lines. It also lists text changes against G-12, G-13, D-N11, E28, E29, F-20 and F-16, among others. The participants judge the withdrawal and those changes in round 7, and nothing here is released by them. The round-7 brief has a 15-minute time box.

### Round outcome

**Another round owed.** G holds blocks G-12 and G-13.

## Ledger

Round-1 status uses the participant's round-1 word. Round-2 status uses the raiser's round-2 word, with its reason.

### Round-1 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | round-1 status | round-2 status | revision |
|---|---|---|---|---|---|---|---|---|---|
| A1 | A | 1 | §7.1; §3 rule 4; §6 | The oracle never consults subject surplus, so over-generation (the chart's main failure mode) is invisible: a `/` read as match where perl divides scores exact, and a forest reporting the union of its readings' sites scores exact, not wider, skipping 07 §7.1.5's exact-floor guard. The harness WRONG is weaker than rule 4's WRONG, and the spec never says how a forest maps to `SubjectFacts`. See also F3, F8. | "WRONG = 0" passes for the pest-style silent misparse prior-art §A1.3 calls the most important lesson; decisiveness has no metric. | Read `internal/parseoracle/compare_facts.go` `compareMarkers` ("A subject surplus … is never consulted"), `decideMarker`; `verdict.go` `BucketWider`; 07 §7.1.3; oracle README. | open, block | **released** (the mapping half): §7.1 maps partial readings to unresolved, checked against `decideMarker`; the exact floor is kept. The asserted-reading half is re-raised under A2. | 4ec0955a, 47951e0a |
| A2 | A | 1 | §7.3; §5 `chart/compare` | Only 68 of 730 perl-valid phase-1 `parses` snippets carry any marker site (68 sites), so the oracle is vacuous on about 91% of phase 1. 07 §7.5.2 calls a shape difference where both match perl "both fine", contradicting "a difference is a finding". Canonical re-emission, pvm's grouping check, is withheld from the chart. See also F4, B11, D4, D14. | The phase-1 gate "every disagreement decided by the oracle" can't be met and gets waived; the bracket diff becomes a comparison against RD's own mistakes with no arbiter. | Ran `parse_facts.pl` over `pre/*.parses.pl` → `perl ok: 730 with >=1 marker site: 68` (`facts.txt`). Read 07 §7.5.2; conformance plan §5 M3; `internal/parse/canon.go` (1812 lines, RD-only). | open, block | **re-raised (block)**, changed: R3 names the arbiter, but no procedure turns the chart's grouping into something comparable with Deparse. Deparse rewrites about 44% of snippets (409/730 survive rough normalisation), so without a chart-side emitter every disagreement is undecided, and undecided has no limit. Ways out: (a) decide RD's side only via Deparse of RD's canon against Deparse of the source; (b) a fully-parenthesised emitter for single-reading regions. Either way, cap the undecided share or make it part of the gate. See also D4, F4. | 4ec0955a, 47951e0a |
| A3 | A | 1 | §7.2 | Lossless round-trip from the forest's tokens holds by construction, because `lexer.Tokenize` already covers every byte exactly once; it tests nothing. The real property is §10's span consistency of every packed node. | One of the three parity conditions is always green. | Read `internal/lexer/lexer.go` `Tokenize` doc; `docs/plans/2026-09-21-cst-is-the-source-of-truth.md`. | open, block | **released**: §7.2 and §10 require each packed node's yield to equal its span and the root's yield to be the source. Loose end in A20. | 4ec0955a, 47951e0a |
| A4 | A | 1 | §8 phase 1 corpus and gate; §2 rule 5 | Run dynamically, the 38 files yield 791 `parses` / 126 `parsent` (not 546/87). Without a feature preamble 172 `parses` fail `perl -c`; with `use feature ":5.36"` 61 still fail (38 missing modules, 23 real, e.g. `$#`, `$*`, `<<$fh>>`, `package Foo 1.4.0;`, `sub foo :attr {1}`, `no v5.10`). 79 of 126 `parsent` cases are valid Perl. The corpus needs `say` as builtin, which §1 lists as a context defect. See also F11, B2. | The gate forces the chart to accept code perl rejects, against rule 5; WRONG is unscorable on those cases; phase 1 bakes in an undeclared feature state. | Ran snippet extraction; `perl -c` with and without preamble (`perlc.txt`, `perlc2.txt`). Perl 5.42.0. | open, block | **re-raised (stand-aside)**, narrowed: preamble, dynamic counts, gate on perl-accepted cases and errata all landed. The 38 module-missing `parses` cases are not errata, don't compile, and have no stated bucket. See also C-N7, D2, D14. | 4ec0955a, 47951e0a |
| A5 | A | 1 | §5 `chart/lex`; §10 Isolation | The lexer does run standalone and up front (`newParser` calls `lexer.TokenizeBarewords`), but its scanners can't be reused without its state machine: every scanner goes through `emit`, which updates `expect` and about 30 other fields; 17 `l.expect` branches in `scan.go`; scanners are unexported. Reuse means refactoring or exporting `internal/lexer`, which §10 omits. Also, perly.y writes `PL_expect` (e.g. `XTERM` in `for(;;)`, `XATTRBLOCK`, `XBLOCK→XOPERATOR` after `->()`). See also B4. | Either about 13k lines of lexer are forked (two lexers drifting), or the lexer `internal/parse` runs on changes; either breaks "cannot block mainline work". The up-front versus interleaved choice rests on a premise wrong at the margins. | Read `internal/parse/parse.go` `newParser`/`newParserOn`; `internal/lexer/lexer.go` `lexer`, `emit`; `grep -c 'l\.expect'` → scan.go 17, proto.go 13; `/opt/corpora/perl5/perly.y` l.189–259, 354, 359, 1210, 1380–1402. | open, block (isolation claim); perly.y detail open, stand-aside | **released**: §5 and §10 name the `internal/lexer` refactor and correct the perl premise. Noted: "18 places" is 14 assignments (A19); "takes its expectation as a parameter" understates the refactor, since `emit` advances about 30 fields (see D12, C-N5); ordering is A17. | 4ec0955a, 47951e0a |
| A6 | A | 1 | §5 `chart/grammar`; §5 dependency claim | The lexer has about 20 coarse kinds (brackets are `Operator`, every quote-like is `Quote`), while the seed's terminals are keyword and delimiter regexes (345 lines mention `%PATTERN`/`WS_OPT`), so terminals must be (kind, text) pairs. Builtin shapes come from CORE.tpm, read by `internal/parse` (`coreShapes`, `coreBarewordOperators`). See also B12. | The chart restates CORE.tpm (a second source of truth) or imports `internal/parse`, breaking the stated dependency. | Read `internal/lexer/lexer.go` Kind consts; `internal/parse/declarations.go` `coreBarewordOperators`, `deriveShapes`; `git -C chalk show 00399607:grammar/perl.bnf`. | open, stand-aside | **re-raised (stand-aside)**, narrowed: token predicates landed, but `internal/parse` reads CORE.tpm by parsing it (`readCore` → `newParser(src, nil, true)`), so the `chart/grammar` generator imports `internal/parse` or carries a second `.tpm` reader, while §5 says only `chart/compare` imports it. See also D7, D8. | 4ec0955a, 47951e0a |
| A7 | A | 1 | §5 "Token branches"; `chart/recognize` | Token branches of different lengths break a token-indexed chart: undeclared word then `/` tokenizes as `/` plus `2`… for division but one `Quote` to the next `/` for a pattern. The chart must be indexed by byte offset over a token lattice; relative distances, Leo's "d tokens back" and interning need defining on it. The lexer's context tracking assumes one linear history. See also B1. | The first thing phase 1 builds hits a design the spec doesn't describe; interning across a branch is undefined. | Read `internal/lexer/expect.go` `after`, `case Word`; `lexer.go` `emit`, `constSubs`. Extent mismatch inferred from `scanBarePattern` emitting one Quote. | open, block | **re-raised (block)**, changed: the byte-offset lattice meets the position bar, but paths rejoin "at the first byte offset both paths reach" while in different lexer states (after `/,/` XOPERATOR, after `/` and `,` a term state; brackets, heredoc queue, carries can differ). pvm's `TokenCache.resyncsAt` and `State.Equal` resync only where offset and lexer state match; the rejoin condition should be the same. See also B-N1, D12. | 4ec0955a, 47951e0a |
| A8 | A | 1 | §8 phase 2 gate; §5 "existing oracle and ratchet harness" | Only `parseoracle.Run` takes an arbitrary `Subject`; `TestPerlGitTRatchet`, `TestParseRatchet` and `conformance` are wired to `internal/parse`. 619/619 is measured through `parse.ParseFileFrom` with a require/use loader, and the 185→278 change came from require resolution, which phase 2's ratchet order lacks. A forest→`call_sites` adapter is also needed (RD's `sites.go` is 1843 lines). See also F1. | Phase 2's parity gate needs a module loader and new chart ratchets not in the plan; cost, duration and drift-and-stall risk are understated. | Read `internal/parse/perlgitt_test.go`; `testdata/perlgitt.ratchet` "619 files, 619 clean"; `internal/conformance/ratchet.go:337`, `run.go:139`; `internal/parseoracle/subject.go`; `wc -l internal/parse/sites.go`. | open, stand-aside | **released**: phase 4 names require/use loading; §6 routes only to `parseoracle`. | 4ec0955a, 47951e0a |
| A9 | A | 1 | §7 last paragraph; §2 M4–M6 | The conformance plan gates canonical re-emission at 100% at M3 and again at M6; the chart has no canon, so the M6 head-to-head compares a route meeting M6's gate with one that can't be measured against it. | The M6 verdict rests implicitly on a weaker gate than M6 itself. | Read `docs/plans/2026-09-05-parser-conformance-plan.md` §5 M3 and M6 tables. | open, stand-aside | **released**: §2 item 2 and §7 say the M6 verdict excludes re-emission. | 4ec0955a, 47951e0a |
| A10 | A | 1 | §2 question; §8 phases 3–4; §11 | perigrin's 2026-10-27 ruling defines M2 as "a 1-char edit … reaches a re-parsed tree AND its published analysis in < 10 ms p99"; the chart can't feed analysis, so the incremental verdict covers only the parse half. §11 says so; §2 doesn't. See also D11. | A favourable incremental verdict could be read as an M2 verdict. | Read conformance plan §5 "Ruling on the step-4 numbers". | open, stand-aside | **released**: §2 and §11 say "the parse half of M2". | 4ec0955a, 47951e0a |
| A11 | A | 1 | §8 phase 3 "about 11 ms for 5 kLOC" | 0.18 s / 67K lines × 5,054 lines ≈ 13.6 ms, not about 11. RD's cold parse on edit5k.pl is 23.89 ms p50, 33.42 ms p99, so the "≈3×" kill line sits around 70–100 ms. The conclusion still stands. See also B9. | Small; the RD baseline the kill criterion is relative to should be stated. | Read plan §5 "Measured at step 4"; subject §12 YAEP numbers. Arithmetic A's. | open, stand-aside | **released**: §9 gives 13.6 ms and RD's 23.89/33.42 ms as hypotheses. That baseline is on a non-Standard file (A18). | 4ec0955a, 47951e0a |
| A12 | A | 1 | §10 Ratchets | pvm's ratchets fail in both directions (`TestRatchetFailsOnImprovement`); an ungated ratchet doesn't fail, which the repo calls failing open. Mainline changes to shared `internal/lexer` will move chart results with no signal. See also F12, D18. | Spike numbers drift between phase gates unnoticed. | Read `internal/parseoracle/ratchet_test.go` `TestRatchetFailsOnImprovement`; crochet development-workflow "Writing a check that can be trusted". | open, stand-aside | **released**: chart ratchets fail in both directions and gate the spike branch. A did not check what runs them on that branch. | 4ec0955a, 47951e0a |
| A13 | A | 1 | §1 RD choice rationale | The shipped parser doesn't let the lexer ask the parser: its lexer is up front with its own `expect` approximation, hedging where it can't know. "That choice has held" is true of RD, not of A1.3's lexer-asks-parser property; interleaved mode would bring it in for the first time. A notes this supports the spike. See also D1. | None if stated; otherwise readers think the chart gives up a property pvm has. | Read `parse.go` `newParser`; `expect.go` Word case; prior-art §A1.3. | open, stand-aside | **released**: §1 corrected and matches `parse.go` `newParser`. | 4ec0955a, 47951e0a |
| A14 | A | 1 | brief (c), §8 phase 1 corpus | With the `:5.36` preamble, `parse.Parse` produces no refused node for 721 of 730 perl-valid `parses` snippets. It refuses 9: `0_.4p1`, `while () {}`, `package q;`/`qq`/`qw`/`qx`/`qr`, `package Foo v1.4.0;`, `package Foo v1.4.0 {1}` (apparent RD defects). Round-trip held on all 917. "Accepted" means `Node.Refused()` false, not a correct tree. | The 9 need listing as one-parser cases. | Ran `PRE='use feature ":5.36"; ' go run ./internal/parse/cmd/guac ../snips` in a scratch archive copy → `rd.txt`; joined with `perlc2.txt` → `joined.txt`. | open, stand-aside | **released**: not a spec finding. Re-run gave the same 9 refusals; A has not checked whether they are filed. | 4ec0955a, 47951e0a |
| A15 | A | 1 | brief (d), §10 isolation | For `internal/parse`, `lower`, `infer`, `psc` the no-change claim looks true in production code (inferred), but `chart/compare` exporting RD trees as labelled spans imports `internal/parse`, against "depends only on `internal/lexer`", and needs a hand-maintained label map between RD's `parse.Kind` and Guacamole's 133 nonterminals. See also F5, B11. | The isolation claim holds for what it names and misses what it doesn't. | Read CST doc; `git show 00399607:grammar/perl.bnf \| grep -E '^[A-Za-z]+ ->' \| awk '{print $1}' \| sort -u \| wc -l` → 133. | open, stand-aside | **released**: §5 header and §10 name `chart/compare`'s import; the label map stays with the implementing issue. | 4ec0955a, 47951e0a |
| B1 | B | 1 | §5 chart/lex "Token branches"; §8 phase 1 branch case | The two readings of the phase-1 branch case produce token sequences of different lengths that don't rejoin until later. YAEP's recogniser and tree walk assume one linear token array (`build_pl` over `toks[tok_curr]`, `pl[]` by token number, `make_parse` `pl_ind--`). The design never says what a position is under a branch; byte positions would destroy interning (inferred); a non-rejoining branch forces multiple lexer states. See also A7. | Phase 1's interface item can't be built on YAEP's recogniser as described; it gets rebuilt mid-phase or faked. | Read yaep.c `build_pl` (l.4638 on), `make_parse` (l.5417 on). Ran `perl -MO=Deparse -e 'foo / 2; # /;'` → `'foo' / 2;`; with `sub foo;` → `foo(/ 2; # /);` (4 tokens against 3 before rejoin). | open, block | **re-raised as B-N1 (block)**: position under a branch is now defined, so that bar is met, but rejoining is not sound. Consequences for `pl_ind--` and fork-node lookahead are in B-N4. | 4ec0955a, 47951e0a |
| B2 | B | 1 | §5 chart/grammar "Seed" | The seed accepts programs perl refuses: `StatementList -> Statement WS_OPT StatementList` allows statements with no `;`; Guacamole splits `Statement`/`BlockStatement`. Removing the rule naively rejects `if ($x) { 1 } $y = 2;`. The seed merges four `ExprValue*` classes into one and has no `NonBrace*` (Guacamole has 84). "The last full Guacamole-derived grammar" is false: clone removal predates 00399607 (28c6029e, d534a328, 39b9a112). It has 374 alternatives over 134 LHS nonterminals, not 227. See also A4. | Phase-1 gate "every `parsent` refused" fails at once; the fix becomes a re-merge of Guacamole into a diverged grammar. | Ran `seedrec.py perl.bnf`: `'1 2'`, `'$x $y'`, `'$x = 1 $y = 2'` ACCEPT (perl -c refuses all); rule removed: `'1 2'` reject, `'if ($x) { 1 } $y = 2;'` reject. Read perl.bnf l.195-198, 374-377; Guacamole.pm l.17-30, 150-204; `git log -- grammar/perl.bnf`. | open, block | **released**: `g1count.py` over Guacamole G1 gives 0 derivations for `NumberDec NumberDec`, `$x $y`, `$x = 1 $y = 2`, and exactly 1 for each of a further set (21 inputs in all, own extractor, not Marpa). Holds only for G1 as written, which B-N2 puts in doubt. G1 refuses `;` and `{1};`, which phase 4 must add. | 4ec0955a, 47951e0a |
| B3 | B | 1 | §5 seed right recursion; §4 YAEP borrowings; §5 up-front vs interleaved | Right recursion interns well only with one-token lookahead, which needs token k+1 before set k+1 is built; interleaved lexing can't supply it, so it forces lookahead 0, where YAEP's right-recursive list goes quadratic. YAEP does not use Leo; right recursion plus lookahead stands in for it. Lexing mode, lookahead, list shape and Leo are one coupled choice, not independent ones. See also D15. | If phase 1 picks interleaved lexing, statement lists go quadratic unless Leo works from day one, and the 11 ms extrapolation loses about 2×. | Ran `rec`, 20 000 statements = 100 000 tokens: R la=1 0.007 s, 11 unique sets; R la=0 87.59 s, 20 007 unique sets, 200 070 012 start situations; L la=0/1 0.016–0.017 s, 20 007 sets. Read `yaep/doc/Internals.txt` items 8, 9, 11; yaep.c l.4660-4664. | open, block | **released**, note in B-N5: §5 "One coupled choice" states the circularity and attributes the probe. Re-run: R la=1 100k tokens 0.009 s; R la=0 25k tokens 3.0 s; L la=0 100k tokens 0.025 s. | 4ec0955a, 47951e0a |
| B4 | B | 1 | §5 "Perl's own `PL_expect` runs standalone" | False for perl: perly.y writes `parser->expect` in 18 places (e.g. `$subref->()` XBLOCK→XOPERATOR), and toke.c sets PL_expect from the symbol table (`yyl_constant_op`, `PL_expect = XTERM`). Grammar-derived classes may still be drivable standalone, but perl is not evidence of it. See also A5, A19, D13. | Phase 1 measures against a false baseline; agreement with perl's 11 states would be misread as evidence for standalone lexing. | Read perly.y l.189-259, 1380-1402 (`grep -c 'parser->expect\s*='` → 18); toke.c l.7971-7979, `yyl_just_a_word` l.8167/8174; perl.h l.5972-5984 (11 states). | open, stand-aside | **released**: §5 says perly.y writes `parser->expect` in 18 places and that toke.c reads the symbol table; both match B's count. (A19 and D13 count 14 writes.) | 4ec0955a, 47951e0a |
| B5 | B | 1 | §5 "Expectation classes"; §8 phase-1 comparison with 11 states | Compression is not defined well enough to measure: the lexer sees the union of expected terminals per Earley set, bounded by distinct set cores. The meaningful measure is a projection onto lexically conflicting token pairs (`/`, `{`, `<`, `%` `&` `*` `?` `-X` `<<` `x`). Comparing to perl's 11 states needs an explicit map from each state to the conflicts it settles. Where the projection says "both", up-front mode must fork (back to B1). | The comparison gate yields an uninterpretable number; "decides by measurement" becomes judgement. | Inferred from Earley set semantics. Read Chalk earley-parser.md "Terminal Clustering". | open, stand-aside | **released**, note in B-N6: the projection, the map from the 11 states and the mismatch count are now stated. | 4ec0955a, 47951e0a |
| B6 | B | 1 | §4 "Leo memos stored by reference"; §10 "Leo on and off give identical forests" | No prior art in either source: YAEP doesn't use Leo; Chalk's Leo covers one chain level (chain extension disabled), so its equivalence test passes for a near no-op Leo. A reference memo keeps sets internable (inferred), but resolving the top walks the chain in O(depth) unless cached in a per-position side table outside set identity; transition-cache hits must rebuild it; chain-end sets never intern. | A Leo that passes "identical forests" while as slow as no Leo, as happened in Chalk. | Read Chalk earley-parser.md "Chain Extension (Disabled)"; Earley.pm l.1363-1372; YAEP Internals.txt item 11. | open, stand-aside | **released**: §10 test fails if Leo is a no-op (`S → a S` at la=0 linear); resolved tops live outside set identity. Side-table rebuild and per-node keying go to the implementing issue, accepted with bias noted. | 4ec0955a, 47951e0a |
| B7 | B | 1 | §5 chart/forest "Scott SPPF built from the chart"; §3 rule 2 | Interning and a binarised SPPF are compatible only if the SPPF is built after recognition, as YAEP does (`make_parse` walks back via `pl_ind - dists[...]`, memoises on (rule, orig, pl_ind) in `parse_state_tab`); a Scott SPPF also needs an intermediate (rule, pos, orig, end) memo. Scott 2008 builds the SPPF during recognition with node pointers in items (recalled), which defeats interning; the design doesn't rule that variant out. | An implementer following Scott's paper puts node pointers in items and quietly loses interning. | Read yaep.c l.4766-4830 (`parse_state`), l.5600-5690. Scott 2008 recalled, not reread. | open, stand-aside | **released**: §4 and §5 say "after recognition" and "items never hold forest pointers". Walking the lattice is in B-N4. | 4ec0955a, 47951e0a |
| B8 | B | 1 | §4 table; §5 recognize: Aycock–Horspool automaton plus interning | Compatible (inferred), but the benefit is unproven: YAEP already computes each core's closure once and reuses it about 20× on average, and its author declined Aycock–Horspool ("only a few percent speedup"). YAEP's tree walk matches on (rule, pos), so states must be expanded back to items. libmarpa moved back to per-item AH items (recalled). | Extra complexity on forest and Leo paths for an unmeasured speed gain. | Read Internals.txt items 3, 5-7, 10; yaep.c `set_core` l.1466-1499. | open, stand-aside | **released**: a measured option in phase 2. | 4ec0955a, 47951e0a |
| B9 | B | 1 | §8 phase 3 "A cold parse cannot meet 10 ms" | A measurement result stated as settled: 11 ms is from 1.43 s / 635 K lines = 2.25 µs/line on a 2015 2.6 GHz Xeon; the 67 K-line run gives 13.4 ms (parse) or 20.9 ms (overall) for 5 K lines. Pre-tokenised, lookahead on, a smaller mostly left-recursive C grammar, no forest, C lines; the 67 K file is `gen.c` ×10 (inflated cache reuse, inferred). See also A11. | Phase 3's gate is decided before phase 3 runs; M2's verdict is shaped by an extrapolation. | Read YAEP README "Speed comparison", Internals.txt item 9. B's arithmetic. Toy probe: 100 K tokens in 7 ms (no evidence for Perl). | open, stand-aside | **released**: now a hypothesis with the RD baseline; the 10 ms sentence is gone. 13.6 ms is close to B's 13.4 ms. | 4ec0955a, 47951e0a |
| B10 | B | 1 | §9 "more than 2^10 readings in one region" | Reading counts are cheap to compute exactly by sum-product over packed nodes (inferred), but "region" is undefined (symbol node, statement, branch span). Independent branches multiply: k undeclared words then `/` in one statement give 2^k, so the cap is k ≈ 10, plausible in generated code. Seed spurious ambiguity (B2) also inflates counts. | The kill criterion can't be applied as written. | Inferred. Seed rules l.374-377. | open, stand-aside | **re-raised as B-N7 (stand-aside)**: "region = a symbol node" makes the root a region, whose count multiplies every independent ambiguity in the file. | 4ec0955a, 47951e0a |
| B11 | B | 1 | §5 chart/compare; §7.3 "identical labelled spans" | Not well defined: labels differ (RD kinds vs `ExprAddU`…, seven ambiguity-class labels cover few kinds); Guacamole precedence yields same-span chains RD lacks; right-recursive `StatementSeq` nests suffix brackets where RD is flat; token spans compare only if token boundaries match, so bytes are safer. PARSEVAL normalises (recalled). See also A2, A15, F4. | Parity condition 3 is never met, or met by ad-hoc normalisation chosen after results are visible. | Read seed and Guacamole rule shapes; inferred. | open, stand-aside | **released**: normalisation fixed before phase 1; label-map contents to the implementing issue. See also A20. | 4ec0955a, 47951e0a |
| B12 | B | 1 | §5 chart/grammar "map `%PATTERN_n%` regexes to token kinds" | Patterns don't map one-to-one: keyword alternations overlap (`read`, `write`, `join`, `split` in PATTERN_2 and PATTERN_62; `last` in 68, 70, 71); no `\b`, so `iffy;` matches `if`. A terminal must be a predicate over a token (kind plus text). Trivia rules go when trivia moves to the lexer; dropping `WS_OPT` creates a duplicate `Program -> StatementList` and removes the self-ambiguous `WS_OPT -> WS WS_OPT WS_OPT`. All whitespace sensitivity sits inside regex terminals. See also A6. | The mapping hides real decisions; duplicate rules add spurious ambiguity. | Ran Python analysis of perl.bnf: "duplicate rules after dropping WS_OPT: 2 (Program, WS_OPT)", "rules with adjacent symbols lacking WS_OPT: 0"; `seedrec.py`: `'iffy;' ACCEPT`. | open, stand-aside | **re-raised as B-N2 (block)**: predicates were the bar, but with the new seed G1's terminals are often smaller than pvm's tokens, so a predicate per lexeme isn't enough. | 4ec0955a, 47951e0a |
| B13 | B | 1 | §5 "Where even perl cannot decide statically" | Perl always decides from compile-time facts (`foo / 2` is division unless `sub foo` seen). A branch exists where pvm lacks a fact perl would have; the packed node's condition is "foo is a declared sub here", which rule 3 lets resolve use. See also C-N6, D10. | Wording only, but it hides that every branch has a resolvable condition. | Ran the B1 Deparse probe. Read toke.c l.8174. | open, stand-aside | **released**: rule 3 now says this, with the toke.c citation. | 4ec0955a, 47951e0a |
| F1 | C | 1 | §8 Phase 2; §2 | Phase 2 is a rewrite labelled a spike: its gate is full parity ("T2 56/56, perl.git `t/` 619/619, T1, and the corpus and site ratchets no lower"), the Guacamole exclusion list is only the start, and no answer arrives before it; speed and incremental come only after. See also A8. | The spike costs what the current parser cost or more before any verdict; verdicts may come months late or never (see F7). | `internal/parse` 25,416 non-test and 40,505 test lines, 862 commits on origin/pu since 2026-09-15; `docs/plans/2026-09-26-perl-git-t-parses.md` baseline "185 of 620 clean", T2 "22/56". | open, block | **released**: §8 orders Standard Perl → early measures → probes → parity, parity only if both verdicts are positive; phases 2 and 3 are bounded. What the restructure broke is in C-N1–C-N3 and C-N6. | 4ec0955a, 47951e0a |
| F2 | C | 1 | §8 order; §9 Speed row; "Throughout" | The cheapest kill signals fire last: speed kill scoped "At phase 3", incremental non-convergence measured in phase 4, both measurable on Standard Perl after phase 1. The edit5k.pl tripwire may time error recovery in phase 1 if edit5k.pl is not Standard Perl (not checked). | A spike a 10× cold-parse gap or Wiren-style non-convergence would kill in week two runs all of phase 2 first. | §8 phase 3 "Gate: perigrin rules on the numbers"; §8 "Throughout"; §3 rule 5. | open, block (resolved by the F1 reorder) | **re-raised (block)**, changed form: measures moved to phase 2, but no Standard-Perl source file is named; the one-line snippet corpus measures per-call setup; the only baseline, edit5k.pl, contains heredocs (edit5k.pl:30 `my $text = <<"END";`, also 102, 174), which are phase-4 work. Bar: name the phase-2 speed input and its RD baseline. See also A18, D5. | 4ec0955a, 47951e0a |
| F3 | C | 1 | §8 phase 1 gate; §9 Silent-misparse row; §7.1 | WRONG = 0 is close to vacuous: the oracle scores five marker kinds, and a subject reporting no site of a kind scores `no-answer`, so a chart reporting no `call_sites` reaches WRONG = 0 by construction; most snippets have no marker op. The gate rests on the bracket diff, graded against RD. See also A1, A2. | Phase 1 can pass with every precedence or structure error invisible; the silent-misparse kill fires only on the five markers. | `internal/parseoracle/README.md` "Measuring another implementation"; Guacamole `t/Statements/Block.t:6-20`, `t/Statements/RequireStatement.t:6-15`. | open, block | **released**: §7.1 reports the forest as a subject, states 68/730 and "necessary, not sufficient"; §7.4 adds a structural arbiter. | 4ec0955a, 47951e0a |
| F4 | C | 1 | §8 phase 1 gate; §7.3 | "Every disagreement decided by the oracle" has no decider for most disagreements: a same-span label difference (precedence, attachment) or block vs hash with no surviving `anonhash` produces no marker. See also A2, B11, D4. | The gate never closes, or whoever types first decides. | Oracle README ("Perl's optree is post-peephole and only ever loses marker ops"). | open, block | **re-raised (stand-aside)**, changed form: the decider is named (§7.4), but the bar also had a cap. "Does not count towards parity" reads both ways; if excluded and uncapped, parity can be reached with any number undecided. See also A2, D4. | 4ec0955a, 47951e0a |
| F5 | C | 1 | §10 Isolation vs §8 phase 2 probes, §7.3 | Contradiction: §10 says nothing in `internal/parse`, `lower`, `infer` or `psc` changes, but §8 implements probes on both sides and §7.3 makes RD-side findings fixable; an open M4 issue on the RD side is a change to `internal/parse`. See also A15, A5. | Either the A/B loses its RD half or the isolation claim, and "cannot block mainline work", is false. | §10 last bullet; §8 phase 2 "On the recursive-descent side…". | open, block | **released**: §10 carves out RD probe work and the lexer refactor as mainline work; §8 phase 3 agrees. The lexer carve-out raises C-N5. | 4ec0955a, 47951e0a |
| F6 | C | 1 | §8 phase 2 probes; §9 "Probes show no advantage" | The A/B is not fair or cheap: (i) maturity bias, with `chart/resolve` infrastructure unallocated; (ii) learning bias survives alternation; (iii) implementer unstated, so review cycles may measure the reviewer; (iv) "clearly less" and "grammar hack" undefined; (v) closed-issue RD cost is historical vs fresh chart cost; (vi) open-issue rule pulls M4–M6 work into the spike. | The headline verdict turns on undefined terms and accounting chosen after the data. | §8 "at least half the probes cost clearly less on the chart and none needs a grammar hack"; §2 item 3. | open, block on (i) and (iv); open, stand-aside on (ii), (iii), (v), (vi) | **released**: (iv) numeric ("at most half the lines and half the places touched"; "a grammar rule written for a single probe"); (i) infrastructure charged equally over a fixed sample. (ii), (iii), (v), (vi) stay stand-asides with the implementing issue. | 4ec0955a, 47951e0a |
| F7 | C | 1 | §9 Drift and stall row | The stall kill ("A gate open far longer than estimated", "Each phase gate's date") refers to estimates and dates the document does not contain. See also D20. | The one guard against an "80–90% migration that drifts" can't fire; phase 2 can stay open indefinitely. | Grep of subject for estimat/week/day/budget finds only this row and the cold-open budget; `/home/user/perigrin/chalk/CLAUDE.md:293`. | open, block | **re-raised (stand-aside)**, changed form: budgets exist, but (a) after "continue" nothing re-arms the rule, and (b) refinement can inflate the issue count; the calendar estimate is the robust one. See also D20. | 4ec0955a, 47951e0a |
| F8 | C | 1 | §9 Silent-misparse row | Unfalsifiable: "Any WRONG that cannot be fixed by refusing" never fires, since refusal is always available; a chart refusing its way to WRONG = 0 then misses parity on coverage, a phase-2 gate miss, not a kill. See also A1, D17. | Rule 5's headline safety property has no kill attached. | §3 rule 5; §9. | open, block | **released**: §9 adopts C's wording; it can fire, since RD refuses only 9 of 730 snippets. The consequence of firing is weakened by C-N4. (D17 holds the revised wording cannot be evaluated.) | 4ec0955a, 47951e0a |
| F9 | C | 1 | §8 phase 3 and 4 gates; §9 Speed and Expectation-class rows | Gates without pass conditions: "perigrin rules on the numbers", "the evidence goes into the recommendation", "more than about 3×… with no credible route… perigrin rules", "A large mismatch is investigated". Only "about 3×" is a number, hedged by "credible route". See also D16. | These phases can't fail, so they inform only the write-up, not continuation. | The quoted text. | open, stand-aside | **re-raised (block)**, changed form: the incremental half reads "sets that rarely match again within one statement of the edit"; "rarely" has no number, and since R1 that verdict gates phase 4. Secondary, stand-aside: "no single hot spot accounting for half the gap" is an unbounded escape (a 10× gap with one hot spot at half still leaves 5.5×). See also D16. | 4ec0955a, 47951e0a |
| F10 | C | 1 | §8 phase 2 gate | "T1" has no target, while every other floor names a number. T1 is 986 files and RD is not clean on it. | Read as T1 ≥ RD ratchet or another bar, depending on who checks. | `docs/specs/perl-parser/07-conformance.md:517` (986 files); `docs/plans/2026-09-26-perl-git-t-parses.md` T1 "554 clean (56.2%)". | open, stand-aside | **released**: phase 4 gate "T1 and the corpus and site ratchets at or above recursive descent's". | 4ec0955a, 47951e0a |
| F11 | C | 1 | §8 phase 1 corpus | Snippet extraction and grading are unstated work: arguments are Perl string literals (`'...'`, `q{...}`, at least 32 multi-line or quote-form calls) to evaluate, write as per-file corpus files and keep in sync. Some `parsent` cases are valid Perl (e.g. `parsent('do { 1, 2 }')`); others carry `# should it?`; the gate treats them alike. See also A4. | An unreviewed extraction step is the corpus, and decides what phase 1 passes on. | Guacamole `lib/Guacamole/Test.pm:22-58`, `t/Statements/Block.t:11-15`; `internal/parseoracle/subject.go`. | open, stand-aside | **released**: corpus "extracted… checked in, and reviewed", with a preamble and erratum rule. Residue in C-N7. | 4ec0955a, 47951e0a |
| F12 | C | 1 | §10 Ratchets | "Reported, but not `make test` gates" needs a mechanism touching shared harness code (`internal/conformance/ratchet_test.go`, `internal/parseoracle/ratchet_test.go`); phase gates use the same ratchets, so "not a gate" must mean "not a mainline gate". See also A12, D18. | Chart regressions turn mainline `make test` red, or nothing enforces phase gates; pvm CLAUDE.md requires 100% `make test` and forbids workarounds. | §10; pvm CLAUDE.md "Test Failure Protocol". | open, stand-aside | **released**: §10 states the property. (D18 raises where `internal/chart` lives.) | 4ec0955a, 47951e0a |
| F13 | C | 1 | §2 Output; §1 | No intermediate deliverable: §2 item 1 (defect yield) has no phase, and there is no point after phase 1 where the spike can stop and still answer part of §2. | A kill or stall mid-phase-2 leaves experimental code and no verdict. | §2 "each measured in a phase of §8"; §8 phase 5. | open, block (dissolved by the F1 restructure if the probe phase yields the M4–M6 verdict) | **released**: a recommendation with both verdicts comes after phase 3. Fit with §2's question is C-N3. | 4ec0955a, 47951e0a |

### Round-2 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A16 | A | 2 | §2 evidence 2; §8 phase 3 | The reordering broke the head-to-head: T1, `t/comp/proto.t`, `t/op/signatures.t`, T4 regex files, CPAN and `t/op` now run in phase 3, before phase 4 adds heredocs, bareword filehandles, paren-less calls and prototype-driven parses. Only 42 of 227 `t/op` files avoid all three of heredocs, bareword-filehandle prints and paren-less `is`/`ok` calls (heuristic). Evidence 1 has the same problem. See also C-N1, D3. | The M4–M6 verdict is backed by a head-to-head that is mostly chart refusals, read as "the chart loses on real code" when it is "not built yet". | `git diff 4ec0955a 47951e0a` moves the head-to-head to phase 3; grep over `/opt/corpora/perl5/t/op/*.t`: 69 heredoc, 15 bareword-filehandle print, 162 paren-less call, 42 none. | open, block | 47951e0a |
| A17 | A | 2 | §5 chart/lex; §8 phase 1 Build | Up-front lexing is the default, but expectation classes are defined over Earley sets, which don't exist up front. It doesn't say what drives the up-front lexer: the existing `expect` machine or something smaller than the set. Phase 1 builds the R4 refactor before the measurement that decides the mode; if the existing machine drives the default, the refactor may not be needed. See also B-N3, D1, D19, C-N5. | A production lexer change for a mode the design expects not to use, or an up-front mode with no defined class source. | Read subject §5 "Expectation classes", "One coupled choice"; §8 phase 1 Build; `internal/lexer/lexer.go` `emit`. | open, stand-aside | 47951e0a |
| A18 | A | 2 | §9 Speed row; §8 phase 2 | The speed baseline is RD on edit5k.pl (23.89/33.42 ms), but edit5k.pl has 70 heredocs, outside Standard Perl until phase 4. Phase 2 measures one-line snippets (setup, not throughput) and unnamed "Standard-Perl source files"; the 3× kill compares "on the same input". See also F2, D5. | The speed kill could fire or not on files picked after the numbers arrive, undoing F9's "thresholds written before the numbers". | `grep -cE "<<['\"~A-Z]" internal/psc/testdata/latency/edit5k.pl` → 70; subject §8 phase 2, §9. | open, stand-aside | 47951e0a |
| A19 | A | 2 | §5 chart/lex "18 places" | perly.y has 14 `parser->expect` assignments; the other 4 matches are `== XBLOCK` tests. A's own round-1 citation mixed the two. See also D13, B4. | Small; a cited count a reader would re-run. | `grep -c "expect\s*=[^=]" /opt/corpora/perl5/perly.y` → 14; with `==` included, 18. | open, stand-aside | 47951e0a |
| A20 | A | 2 | §5 grammar `:discard`; §6 "trivia attached"; §7.2 yield; §5 compare "byte spans" | With trivia `:discard`ed from the grammar, a node's yield equals its byte span only once trivia attaches to tokens, and the attachment rule (leading, trailing, which side of a newline) also decides the spans `chart/compare` diffs against RD's `Start`/`End`. See also A3, B11. | The yield property and span normalisation depend on an unfixed rule; a mismatch shows as a bracket disagreement on every node. | Read subject §5, §6, §7.2; `internal/lexer` emits trivia as separate tokens (`IsTrivia`). | open, stand-aside | 47951e0a |
| B-N1 | B | 2 | §5 "Input model" | Rejoining at "the first byte offset both paths reach" is unsound: paths can share a token boundary in different lexer states. `<<` as shift or heredoc: in `foo <<EOF;` the shift path is `foo` `<<` `EOF` `;`, the heredoc path `foo` `<<EOF` `;`; the first shared offset is after `EOF`, where only one path has a heredoc body pending. The rejoin condition must be (byte offset, lexer state equal), including at least the pending-heredoc queue. Re-raises B1. See also A7, D12. | Wrong tokens after every heredoc branch; the forest carries a reading perl never takes (rule 5). | Ran `perl -MO=Deparse h1.pl` → `'foo' << 'EOF'; 'EOF'->bar;`; with `sub foo;`, h2.pl → `foo("bar\n");`. Token boundaries inferred. | open, block | 47951e0a |
| B-N2 | B | 2 | §4 "G1 rules used as written"; §5 chart/grammar "each lexeme name maps to a token predicate" | G1 builds single pvm tokens out of several lexemes (`VarScalar ::= SigilScalar VarIdentExpr ElemSeq0`, `VarIdentExpr ::= VarIdentExpr PackageSep VarIdent`, `LitString ::= SingleQuote NonSingleOrEscapedQuote_Many SingleQuote`, `SubNameExpr`, `Label ::= IdentComp Colon`, `Diamond ::= '<' VarScalar '>'`), while pvm emits `$x`, `'a b'`, `Foo::Bar` as one token. Either sub-token rules are rewritten (contradicting "as written" and reopening B2) or the chart lexer splits tokens (contradicting scanner reuse). Say which, and where the cut falls; `VarScalar` holds `ElemSeq0`. Re-raises B12. See also D8. | The grammar reader can't be written as specified, and the shipped grammar differs from the one checked. | Read Guacamole.pm G1, extracted to `r2/g1.txt`; G1 terminals include `SigilScalar`, `VarIdent`, `IdentComp`, `PackageSep`, `DoubleQuote`, `SingleQuote`, `NonDoubleOrEscapedQuote_Many`. | open, block | 47951e0a |
| B-N3 | B | 2 | §5 "Expectation classes" vs "One coupled choice … up front … is the default" | Classes are defined from the Earley set, which an up-front lexer has no access to, so the default mode can't use classes as defined. Either the production `expect` machine drives the chart's lexer, or there is an undescribed standalone driver. If the former, §1's motivation is not delivered by the default and phase 1 measures the production machine, not grammar classes. See also A17, D1, D19. | Phase 1 builds `chart/lex` in a mode whose driver is unspecified; §1's premise is quietly dropped. | Read §1 para 1; §5 Expectation classes; §5 One coupled choice. | open, block (name the up-front driver) | 47951e0a |
| B-N4 | B | 2 | §5 recognize and forest, over the lattice | YAEP machinery assumes a linear index: `make_parse` steps back with `pl_ind--`, `check_cached_transition_set` reads `pl[place+1-dist]`, and the (set, token, lookahead) key has one next token. At a fork node the next token isn't unique and a predecessor is a token edge, not node−1; within a branch that probably means lookahead 0 and no cache, which should be said. | Small; the implementing issue's to get right. | Read yaep.c l.4614-4628, 5560-5600. | open, stand-aside | 47951e0a |
| B-N5 | B | 2 | §5 "One coupled choice" | "List shape" is in the heading but fixed in both bullets; at la=0 a left-recursive list is linear without Leo (0.025 s on 100k tokens), though "G1 as written" keeps right-recursive `StatementSeq`. "11 unique sets" holds only at la=1; with Leo at la=0 statement-boundary sets stop interning, so phase-2 set reuse will depend on lexing mode. See also B3. | Phase 2 reads low set reuse under interleaving as a defect rather than as expected. | Ran `rec L 0 20000` → 0.025 s; round-1 L la=0 had 20,007 unique sets. Leo inferred. | open, stand-aside | 47951e0a |
| B-N6 | B | 2 | §5 Expectation classes, the `{` pair | In G1 block and hash are the same lexeme (`LBrace`), told apart by `NonBrace*`, so `{` is not a lexical conflict for this grammar. In perl XSTATE and XTERM differ mostly in how `{` is taken, so the map from the 11 states registers grammar artefacts against the ">2 of 11" threshold. See also B5, C-N9. | The investigate-before-phase-2 trigger may fire, or be argued away, for a reason unrelated to grammar correctness. | Read G1 `BlockNonEmpty ::= LBrace StatementSeq RBrace`, `LitHashNonEmpty ::= LBrace Expression RBrace`. | open, stand-aside | 47951e0a |
| B-N7 | B | 2 | §9 Forest row | A sum-product count at any symbol node includes the root, which multiplies independent ambiguities. Count at an ambiguous (packed) node stopping at the next independent ambiguity, or take the maximum over statements; pick one and write it down. Re-raises B10. | False forest kills on large files with many small independent ambiguities. | Inferred from the definition of sum-product counting. | open, stand-aside | 47951e0a |
| C-N1 | C | 2 | §2 item 2; §8 phase 3 | The phase-3 head-to-head needs parity-level coverage: §2 assigns evidence 2 (M4 `t/comp/proto.t`, `t/op/signatures.t`, T1; M5 T4 regex files and CPAN; M6 `t/op`) to phase 3, where the chart has only the Standard grammar plus per-probe context; it will refuse most of those files. See also A16, D3. | Head-to-head numbers dominated by refusals, read as "chart far worse at M4–M6" or waved away. | §8 phase 3 "built on the Standard grammar plus only the context pieces it needs"; §8 phase 4 exclusion list; edit5k.pl:30. | open, block | 47951e0a |
| C-N2 | C | 2 | §8 phase 5; §9 | Stale cross-reference: phase 5 says "Measures (b) and (c)", defined at 4ec0955a as "rebel's measures a to c"; that sentence was removed and only (a) is now defined (phase 2). | Phase 5's measures can't be identified from the document. | 47951e0a lines 309, 350; 4ec0955a line 326. | open, stand-aside | 47951e0a |
| C-N3 | C | 2 | §2 question vs §2 Output | The question opens "Can an Earley chart parser reach parity with the floor pvm stands on", but parity is phase 4, run only on two positive verdicts, while the recommendation comes after phase 3. See also D6. | A positive recommendation reads as "can reach parity" when untested; a negative one leaves the question formally open. | §2 blockquote; §2 Output; §8 phase 4 heading. | open, stand-aside | 47951e0a |
| C-N4 | C | 2 | §9 preamble | Kills no longer kill: 4ec0955a said "A kill ends the spike early and the recommendation says why"; 47951e0a says only "Fixed before work starts.", and the phase-2 gate is "perigrin's ruling, against thresholds". Nothing states what happens when a §9 criterion fires. See also F8, D17. | Every kill criterion is advisory, and thresholds can be overridden after the numbers with nothing recorded. | Diff at §9's opening; 4ec0955a line 346. | open, stand-aside | 47951e0a |
| C-N5 | C | 2 | §5 `chart/lex`; §10 Isolation | Phase 1 puts the R4 production refactor on the critical path with no exit: the lexer is 5,724 non-test lines; `expect` is read in scan.go (17), proto.go (13), lexer.go, brackets.go, heredoc.go, quote.go, pod.go; the token cache resyncs on lexer `State`. It is charged to no budget or independent justification, and its fate on a kill is unstated. See also A17, D12, D19. | A killed spike leaves an M2-critical mainline refactor justified only by the spike, or its cost is hidden outside phase 1's budget so the F7 drift guard undercounts. | §5 "That is a production change: it goes through mainline gates"; grep counts; `internal/lexer/tokencache_test.go` ABOUTME. | open, stand-aside | 47951e0a |
| C-N6 | C | 2 | §3 rule 3 vs §9 Forest row and §12 Kegel | Rule 3 says "Perl always decides from such facts… so every branch has a resolvable condition", while §9 kills on "where perl's reading is not decidable by compile-time facts" and §12 cites Kegel ("Perl cannot be statically parsed (BEGIN plus prototypes)"); the parseoracle README calls a prototype installed by string `eval` in `BEGIN` "the undecidability case". See also D10, B13. | Either the forest kill can never fire, or rule 3 is false for BEGIN-time code. | §3 rule 3; §9 Forest row; `internal/parseoracle/README.md` opening. | open, stand-aside | 47951e0a |
| C-N7 | C | 2 | §8 phase 1 corpus and gate | 38 cases fall between the gate's categories: of 61 `parses` failing `perl -c` with the preamble, 38 are missing modules and only 23 are named errata; the gate covers "every `parses` case perl accepts" and "every erratum", and the 38 are neither. 791 − 61 = 730, matching §7.1. See also A4, D2, D14. | Their disposition is decided by whoever extracts, and §7.1's 730 denominator silently depends on it. | Ledger row A4; §8 phase 1 corpus. | open, stand-aside | 47951e0a |
| C-N8 | C | 2 | §8 phase 1 gate vs §7.1 | The gate requires all four parts of §7, and §7.1 requires coverage and exactness "at or above the recursive-descent ratchet", but no RD ratchet exists over the Guacamole snippets. | The comparison point is chosen ad hoc at the gate. | §7.1; §8 phase 1 gate; pvm ratchets are T1, corpus, sites. | open, stand-aside | 47951e0a |
| C-N9 | C | 2 | §8 phase 1 gate vs §9 classes row | Units mismatch: the gate records "the count of conflict decisions where the generated classes differ"; §9 triggers on "More than 2 of the 11 states' conflict decisions unmatched", which counts states. See also D17, B-N6. | Two readings of one threshold. | §8 phase 1; §9. | open, stand-aside | 47951e0a |
| D1 | D | 2 | §1 vs §5 "One coupled choice" vs §8 phase 1 | §1 justifies the spike by the interleaved lexer-asks-parser property, but §5 makes up-front lexing the default while defining classes over Earley sets, which don't exist up front; whether grammar classes can run up front is unknown until phase 1, yet phase 1 must already build the recognizer and choose lookahead and Leo, and the mode is recorded only at phase 1's gate. See also A17, B-N3, D19, A13. | Phase 1 builds both modes, or a default that may reduce to the existing `expect` approximation over a hedged token stream; nothing covers up front failing to use grammar classes. | Read §1 para 1; §5; §8 phase 1 gate; `internal/parse/parse.go:679-681` (`TokenizeTyped`/`TokenizeBarewords`); prior-art.md l.914. | open, block | 47951e0a |
| D2 | D | 2 | §8 phase 1 corpus and gate; §3 rule 5 | The gate requires every erratum (23 under perl 5.42.0) refused with an Error node. (a) How "perl rejects" is decided is unstated: `perl -c` with the preamble over 769 unique `parses` gives 61, and 23 needs excluding 38 `Can't locate`. (b) At least 16 of the 23 are not syntax errors: 10 `Invalid CODE attribute`, 5 `no 5.010`, 1 `Can't modify logical and`. (c) 07 §7.5.2 says perl-failed input is NO ANSWER, excluded from scoring. (d) G1 "as written" accepts these cases, and rule 5 refuses only outside the accepted language. See also A4, C-N7. | The gate requires refusing valid Perl (e.g. `sub foo :attr {1}`, `no 5.010`), needing semantic checks or grammar edits, contradicts pvm's conformance rule, and can't be reproduced. | Ran recorder stub over Guacamole `t/`: 791/126 (769/124 unique); `perl -c` → `D/rejects.txt`: 38 Can't locate, 10 Invalid CODE attr, 5 too modern, 2 `<<STDIN>>`, 2 dotted version, 1 bare `<<`, 1 `&&` assignment, `$#`, `$*`. Read 07 §7.5.2. | open, block | 47951e0a |
| D3 | D | 2 | §8 phase 3 sample and threshold; §2 evidence 2 | Most candidate probes and M4/M5 head-to-head corpora need constructs Standard Perl excludes (prototypes, paren-less calls, unusual delimiters: standard.pm "All subroutines must use parentheses", "Prototypes do not change the parsing rules", "Limited delimiters"); `t/comp/proto.t` and CPAN are full of them. A probe needing an excluded feature needs exactly "a grammar rule written for a single probe". See also A16, C-N1. | The M4–M6 verdict is negative by construction or measured on a chart that refuses most of its corpus; phase 3 is not the promised early answer. | Read Guacamole `lib/standard.pm` "Things not supported" (l.81–355); §8 phase 3 candidates; §8 phase 4 list. | open, block | 47951e0a |
| D4 | D | 2 | §7 part 4; §5 components; §8 phase 1 Build | No component builds the Deparse arbiter and phase 1's Build list omits it. pvm's Deparse channel (07 §7.1 Channel 4) compares re-emitted source, closed to a chart with no canon; matching `-p` output to two trees' spans is unspecified. Deparse folds constants (`1 + 2 * 3` → `7`), so precedence disagreements come out undecided, with no ceiling. See also A2, F4. | Part 4 can't be evaluated in phase 1, and once built has an unbounded escape hatch. | Ran `perl -MO=Deparse,-p -e 'my $y = 1 + 2 * 3;'` → `(my($y) = 7);`; `map { $_ => 1 }` → `map({($_, 1);} @x)`. Read 07-conformance.md l.125–140, 7.1.2 table. | open, block (the missing component and the undecided ceiling; how it matches spans belongs to the implementing issue) | 47951e0a |
| D5 | D | 2 | §8 phase 2 gate; §9 Speed | The 3× kill is "on the same input", with baseline edit5k.pl (23.89 ms p50), which has 70 heredocs and is not Standard Perl; phase 2 names no "Standard-Perl source files". "Cold parse" is unscoped: `parse.Parse` includes declaration reading and site building, and the chart side's inclusion of forest and resolve is unstated. See also A18, F2. | The one numeric phase-2 kill has no fixed input or scope, contradicting "thresholds written here before the numbers arrive". | Ran `grep -c '<<' internal/psc/testdata/latency/edit5k.pl` → 70. Read conformance plan l.620 (`BenchmarkLatencyParse` = `parse.Parse`); standard.pm l.102. | open, block | 47951e0a |
| D6 | D | 2 | §2 Output; §8 phase 4 heading | "Two verdicts that may differ", yet phases 4 and 5 run only if both are positive; phase 4 is what the M4–M6 route needs, so a negative incremental verdict also stops a positive M4–M6 route. See also C-N3. | A negative M2 result vetoes a positive M4–M6 result; the design's split is defeated. | Read §2 Output "Phases 4 and 5 run only if perigrin rules both positive"; §8 phase 4. | open, block (one-line fix: either verdict alone unlocks phase 4) | 47951e0a |
| D7 | D | 2 | §5 dependency rule; §5 chart/grammar | "Builtin shapes … come from CORE.tpm, the one source `internal/parse` already reads", but "only `chart/compare` also imports `internal/parse`". pvm reads CORE.tpm by parsing it with `internal/parse` (`coreDeclarations` → `newParser`, then `deriveShapes`); `coreShapes`, `coreBarewordOperators` are unexported, and the lexer takes bareword operators from a caller that has read CORE.tpm. See also A6, D8. | chart/grammar imports `internal/parse` (breaks §5), or code moves out (breaks §10), or the derivation is reimplemented (the forbidden second table). | Read `internal/parse/declarations.go` `readCore`, `coreDeclarations`, `coreShapes`, `deriveShapes`; `internal/lexer/lexer.go:222-225`. | open, stand-aside | 47951e0a |
| D8 | D | 2 | §4 Guacamole; §5 chart/grammar | G1 already encodes builtin shapes in about 200 per-keyword rules (e.g. `OpKeywordAbsExpr ::= OpKeywordAbs OpUnaryKeywordArg`, `OpKeywordTimeExpr ::= OpKeywordTime LParen RParen`), so "G1 rules are used as written" and "builtin shapes come from CORE.tpm, not from a second table" cannot both hold. See also A6, D7, B-N2. | Two shape tables that disagree (e.g. `time` without parens), with no rule for which wins. | Ran `grep -c` of `^Op…KeywordExpr ::=` in `lib/Guacamole.pm` → 201; l.802, 1225, 1269. | open, stand-aside | 47951e0a |
| D9 | D | 2 | §3 rule 6 vs §5 chart/lex | Rule 6 says "Weights may order, never decide", but §5 says "Guacamole's lexeme priorities and perl's tiebreaks decide"; a Marpa lexeme priority is a weight that discards the losing lexemes. | Rule 6 is broken by the first component built under it, or needs a scope ("in the forest"). | Read §3.6; §5 chart/lex para 2; Guacamole.pm l.1884ff `:lexeme ~ X priority => 1`. | open, stand-aside | 47951e0a |
| D10 | D | 2 | §3 rule 3 | "An ambiguity in pvm is a fact pvm lacks, so every branch has a resolvable condition"; spec 03 §3.0 calls lexer ← runtime "a dependency on an interpreter that has already run. A static parser cannot reproduce it in general". Resolvable by perl, not necessarily by pvm; §9's forest kill assumes the case rule 3 denies. See also C-N6, B13. | Rule 3 is stated as settled while §9 and spec 03 say otherwise; readers may expect `chart/resolve` to close every branch. | Read 03-lexer-feedback.md §3.0 item 2; §9 forest row. | open, stand-aside | 47951e0a |
| D11 | D | 2 | §2 Q; §8 phase 5; §1 "parse half of M2" | M2's ruled gate is "Incremental == full" and a 1-char edit to edit5k.pl reaching "a re-parsed tree AND its published analysis in < 10 ms p99"; the spike measures convergence and equivalence only, and its only speed bar (3× cold) allows about 70 ms. See also A10. | "A better chance at the parse half of M2" can be ruled positive on a parser that can't meet M2's latency bar even before analysis. | Read conformance plan §5 M2 table and 2026-10-27 ruling (l.580–640). | open, stand-aside | 47951e0a |
| D12 | D | 2 | §5 chart/lex, input model | "A scanner takes its expectation as a parameter" replaces only `Expect`; other token-deciding state includes the heredoc `pending` queue, bracket stack, `constSubs`, `sawSubWord`/`expectPrototype`, `attrArgDepth` and format state, each forking per lattice path and leaking into what can be shared. See also A5, A7, B-N1, C-N5. | The refactor and the interning-near-a-branch question are larger than stated; "one lexer" may become one per path. | Read `internal/lexer/lexer.go` l.258–430, l.620–632; `l.expect` 44 times across 9 non-test files. | open, stand-aside | 47951e0a |
| D13 | D | 2 | §5 "PL_expect is not evidence" | "perly.y writes `parser->expect` in 18 places": 14 are writes and 4 are reads (`if (parser->expect == XBLOCK)`, l.1380–1401); the reads support the doc's point more strongly. See also A19, B4. | Minor factual error in a claim stated as measured. | Ran `grep -c 'parser->expect' perly.y` → 18; `grep -c 'parser->expect *= *X'` → 14. | open, stand-aside | 47951e0a |
| D14 | D | 2 | §7 part 1 | "Only 68 of 730 snippets carry a marker site": the phase-1 corpus is 791 `parses` minus 23 errata (768), or 769 unique; 730 matches neither and has no source. See also A2, A4, C-N7 (A's re-run reproduces 730 as the preamble-passing count; C-N7 notes 791 − 61 = 730). | An unreconciled number in the sentence that justifies "necessary, not sufficient". | Ran recorder counts; read §8 phase 1 corpus. | open, stand-aside | 47951e0a |
| D15 | D | 2 | §5 coupled choice; §9 Speed | The YAEP probe figures (0.007 s / 11 sets; 87.6 s / 20,007 sets) and the 13.6 ms extrapolation have no file, commit or script, and the grammar was a toy list, not Guacamole; they justify the default mode. See also B3, B-N5. | The default is chosen on numbers nobody can rerun. | Read §5, §9; nothing matching `87.6` or `20,007` under pvm, Guacamole or Chalk. | open, stand-aside | 47951e0a |
| D16 | D | 2 | §8 phase 2 gate | "Sets that rarely match again within one statement of the edit": "rarely" has no number and no edit population is named. See also F9. | The incremental verdict is a judgment, not a threshold fixed in advance. | Read §8 phase 2. | open, stand-aside | 47951e0a |
| D17 | D | 2 | §9 Silent misparse; §9 Expectation classes | "A WRONG … whose only fix is a refusal" needs knowledge of every possible fix, so can't be evaluated; the class row's "investigated before phase 2" is an action, not a kill, and counts states while the phase-1 gate counts decisions. See also F8, C-N4, C-N9. | Two of the eight kill rows cannot fire. | Read §9 table. | open, stand-aside | 47951e0a |
| D18 | D | 2 | §10 Ratchets; Isolation | `make test` is `go test ./... -count=1`, so any `internal/chart` test on mainline runs in it, while the lexer refactor and RD probe work land on mainline. Where `internal/chart` lives and how a long-lived branch tracks a changing lexer are unstated; `go generate` tables are run by no Makefile target (CLAUDE.md "Use the Makefile for ALL builds"). See also F12, A12. | Chart ratchets end up in mainline `make test`, against the doc, or the spike lives on a drifting branch. | Read `Makefile` l.21–22; CLAUDE.md. | open, stand-aside | 47951e0a |
| D19 | D | 2 | §8 overall | The first evaluable kill (speed, phase 2) comes after phase 1, the largest phase (grammar reader and generator, production lexer refactor, recognizer, forest, compare, oracle adapter, Deparse arbiter, parity on about 768 cases). A recognizer-only speed and set-reuse measurement over Guacamole's grammar on today's tokens could come first. See also A17, B-N3, C-N5, D1. | The most expensive work, including a production change, is spent before any kill signal. | Inferred from §8 Build lists and §9. | open, stand-aside | 47951e0a |
| D20 | D | 2 | §8 Budgets | "A phase open beyond twice its budget forces a decision" names no owner and no check, and the issue-count half has no clock. See also F7. | An obligation nobody will check. | Read §8 Budgets. | open, stand-aside | 47951e0a |

### Round-3 status of rows open at 47951e0a

Each status is in the raiser's own word, with its reason, as declared at b5b10428. Rows released in round 1 or round 2 are not repeated here.

| id | raiser | round-3 status | reason given |
|---|---|---|---|
| A2 | A | **released** (with it, A1's asserted-reading half) | §7.4 compares Deparse of RD's canon with Deparse of the source, with at most 10% undecided as a §9 kill row. A re-ran it over the 730 snippets: 722 equal, 1 differs (`$# { $foo->{'bar'} }`), 7 canons fail to compile (`1e_+_1`, `1e_+__1`, `07p-2`, `0xfp-1`, `0x_ap+2`, `0xap+2_`, `while () {}` → `while (){}})`), so 8/730 (1.1%) are undecidable. Phase-3 rates are unmeasured. Remaining error: "pvm's existing Faithful check" is wrong (A24). See also D4, C-N10, E1–E3. |
| A4 | A | **released** | Stubs on `@INC`; errata excluded per 07 §7.5.2. Re-run: bare stubs make 20 of 38 compile, and the other 18 need an `our $VERSION` set high enough. That is an implementation detail for the corpus issue. See also D-N2. |
| A6 | A | **released** | Shapes come from G1's per-keyword rules; CORE.tpm is a reviewed cross-check; the generator's import of `internal/parse` is stated. See also E13. |
| A7 | A | **released** | §5 rejoins only on equal byte offset and full lexer state, citing `resyncsAt`/`State.Equal`, with the `foo <<EOF;` test in §10 and no transition cache inside a branch. Not re-run, since no lattice exists. The rule applies only where a lexer can branch (A21). |
| B1 | B | **released** via B-N1 | See B-N1. |
| B10 | B | **re-raised (stand-aside)** via B-N7 | See B-N7. |
| B12 | B | **released for the collapse; re-raised as B-N10 for predicate overlap**, via B-N2 | See B-N2. |
| F2 | C | **held, changed form (block)** | `edit5k-standard.pl` rewrites only heredocs. Guacamole `lib/standard.pm:88-100` bans auto-quoting, and edit5k.pl has 281 lines with a bareword before `=>` (e.g. l.13-16) and 350 with a `{bareword}` subscript (e.g. l.22), so G1 refuses it at line 13. Secondary: edit5k.pl is generated ("do not edit by hand"), so a hand copy drifts. Bar: the whole file made Standard Perl (or generated as a variant), with phase 0's recognizer accepting it before timing counts. See also D5, E4, B-N11, A23. |
| F4 | C | **released** | §7.4 caps undecided at 10% of all disagreements, with a §9 row: direction and cap. Released against C's own bar; it misfires with the arbiter's asymmetry (C-N10). |
| F7 | C | **released** | The calendar governs, "continue" records a new budget and re-arms, checked at every issue close. Released against C's own bar. |
| F9 | C | **released** | Phase 0: fewer than 80% of 200 edits converging within the statement plus 8 tokens. Hot-spot clause close to C's wording; residue C-N14; the mode is C-N11. See also E12. |
| A16 | A | **released** | Head-to-heads run in phase 3 "once the chart parses those files"; phase-2 probes build the features they need, charged to the probe. |
| A17 | A | **released**, but superseded by A21 | Up front uses the production lexer with no refactor; the refactor comes only if phase 0 keeps interleaving viable. |
| A18 | A | **released** | `edit5k-standard.pl` is fixed, with RD timed on the same file. A's re-run: heredocs are the only part G1 lacks a rule for. POD is under A23. (C, D, E and B hold that auto-quoting is also excluded: F2, D5, E4, B-N11.) |
| A19 | A | **released** | §12: 14 places. (E15 counts 18.) |
| A20 | A | **released** | §5 compare: a trivia attachment rule is part of the normalisation fixed before phase 1. |
| B-N1 | B | **released** | `State` is the whole lexer minus cursor and output, so `pending`, `brackets`, `expect`, `pendingFormat`, declared names and carries all take part in `Equal` (checkpoint.go); `resyncsAt` also requires `peekEnd <= pos`. Side effects in B-N8, B-N9. |
| B-N2 | B | **released for the collapse; re-raised as B-N10 for predicate overlap** | `collapse.py` leaves no sub-token lexemes; `1 2`, `$x $y`, `$x = 1 $y = 2` give 0 derivations, and a further ten inputs give exactly 1. This is B's own reading of the cut. |
| B-N3 | B | **released** | §5: up front means the production `expect` machine, no classes; §1's property is assigned to interleaved mode only. Knock-on in B-N12. |
| B-N4 | B | **released** | "Lookahead 0 and no cache inside a branch" is stated; the lattice-edge walk is accepted as the implementing issue's. |
| B-N5 | B | **released** | "Lower set reuse by construction"; phase 0 re-measures. |
| B-N6 | B | **released**, with a sibling in B-N9 | `{` removed, for the right reason; `<<` fails the same test. |
| B-N7 | B | **re-raised (stand-aside)**, against B's own suggestion | "Stopping at nested packed nodes" counts one node's local fan-out, so 2^10 needs over 1024 packed children; the kill has gone from over-sensitive to blind. Count over a fixed unit, e.g. one statement node. See also C-N15. |
| C-N1 | C | **released** | Evidence 3 is "phase 3, once the chart parses those files". |
| C-N2 | C | **released** | Phase 4 defines (a), (b), (c) inline. |
| C-N3 | C | **released** | §2 asks "Is an Earley chart parser worth taking to parity". |
| C-N4 | C | **released** | "A fired kill ends the spike unless perigrin records an override and its reason." |
| C-N5 | C | **released** | The refactor counts in phase 1's budget, is done only if phase 0 passes, and is kept only on its own merits. |
| C-N6 | C | **released** | Rule 3 now says facts obtainable only by running code stay ambiguous. Side effect on §9's forest row: C-N15. |
| C-N7 | C | **released** | Every case has a disposition. Note for the implementing issue: a stub with no prototypes or exports can change perl's own parse. |
| C-N8 | C | **released** | §7.1: "at or above recursive descent's measured score on the same corpus". |
| C-N9 | C | **released** | The phase-1 gate and §9 both count states. |
| D1 | D | **re-raised (narrowed), block** | The circularity is gone, but phase 1 builds interleaved `chart/lex` only "if phase 0 kept interleaving viable", and phase 0 has no viability criterion; it never runs recognition at lookahead 0 with Leo on G1. Fix: run that on `edit5k-standard.pl` under the same 3× bar as the viability test, and name it in §5. See also C-N11, E6, B-N12, D-N1. |
| D2 | D | **released** | Perl-rejected cases are excluded from scoring per 07 §7.5.2. Re-run: with stubs defining `$VERSION = 999`, exactly 23 of 769 unique (23 of 791) are rejected. See D-N2. |
| D3 | D | **released** | R7 makes the verdict evaluable. Residual bias, stand-aside with no action required: a chart probe pays the whole feature while a closed RD issue pays only its PR; and "no Perl construct justifies" is a weak hack test. See also C-N13, E17. |
| D4 | D | **re-raised, block** | (a) False about pvm: `parse.Faithful` (`internal/parse/fidelity.go:45`) compares tokens and never invokes perl; no production code in `internal/parse` or `internal/parseoracle` runs Deparse. (b) Circular: where canon is token-equal to the source, the two Deparse outputs are trivially equal, so the arbiter decides for RD wherever RD is Faithful (347/347 M3 files). D ran `parse.Faithful` and got true for `my $y = ${$h->{k}};` (cited as a known one-childless-leaf defect, issue 01a0ad52) and for `my $y = 1 + 2 * 3;`. The arbiter decides on literal snippets by textual identity, so the 10% ceiling is easy. Fix: compare Deparse,-p of the source with a fully parenthesised emission of each tree, or Concise op shapes, not resting on Faithful; update phase-1 Build, §9's undecided row, and §10's arbiter test. See also A24, C-N10, E1, E2, E3. **Author's note (not a release; only D can release D4):** on origin/pu ad7faf79, recursive descent parses `${$h->{k}}` as `unary` over `index($h, k)`, and issue 01a0ad52, which described the leaf, is closed (done). |
| D5 | D | **re-raised, block** | Heredocs are not edit5k.pl's only exclusion: about 350 bareword `{key}` subscripts and about 281 `word =>` pairs, which `standard.pm` forbids. It is also generated ("do not edit by hand"). Fix: rewrite under named rules with an acceptance test (Guacamole's grammar or the phase-0 recognizer accepts it), preferably generated behind a flag; update §9's Speed row and phase 0's convergence corpus. See also F2, E4, B-N11, A23. |
| D6 | D | **released (block); residue is D-N3** | R8: a positive M4–M6 verdict alone unlocks phase 3. |
| D7 | D | **released** | The generator may import `internal/parse` at `go generate` time. See also E13. |
| D8 | D | **released** | Shapes from G1's per-keyword rules; CORE.tpm a reviewed cross-check. |
| D9 | D | **released** | Rule 6 scoped to ordering readings; lexeme priority restated as a deterministic tiebreak. See also B-N10. |
| D10 | D | **released** | Rule 3 now admits facts obtainable only by execution. See also E11. |
| D11 | D | **released** | §2: "no result here establishes M2's latency gate"; phase 4 reports the parse share. |
| D12 | D | **released** | Lattice nodes carry full lexer state; `checkpoint.go:102` uses `reflect.DeepEqual` over the cloned lexer. See also B-N8. |
| D13 | D | **released** | 14 writes, re-verified. (E15 counts 18.) |
| D14 | D | **re-raised, stand-aside** | 730 is 791 minus 61 rejections without stubs; under the revision's own stub rule the count is 768 with repeats or 746 unique, so 68 is measured on a corpus the doc no longer uses. "Participant A's count" should not be cited in the design. See also E9, A25. |
| D15 | D | **released** | §5 labels the probe a toy and phase 0 re-measures, acceptable given D1's fix. |
| D16 | D | **released** | Numeric threshold. Residue for the implementing issue: how the 200 edits are drawn. See also E12. |
| D17 | D | **released** | The silent-misparse kill fires on an observed fix; class rows count states. Under D4 as written, the arbiter signal is biased toward RD. |
| D18 | D | **released** | `chart` build tag, `make test-chart`, generation through a Makefile target. See also C-N17, E14. |
| D19 | D | **released** | Phase 0 comes first. See D-N1 for what it no longer covers. |
| D20 | D | **released** | The calendar governs, checked at every issue close, re-armed on continue. |

### Round-3 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A21 | A | 3 | §5 chart/lex "Two modes", "Ties"; §8 phase 1 Build; rule 3 | Up-front mode cannot emit token branches: it is the production lexer, a single linear machine that picks term for `foo / 2` and hedges ("That one stays a hedge"). Branches, the lattice and the rejoin rule exist only in interleaved mode, yet phase 1 lists "one token-branch case" whatever phase 0 decides. If phase 0 picks up front, rule 3's lexical conditions disappear and the chart inherits the production lexer's guesses. See also E6, C-N11, D1, A7. | Phase 2's lexical probes (named unary vs list operator with unknown prototype, `sort $coderef @list`) are measured on a chart that can't represent their ambiguity, so the M4–M6 verdict is about a weaker design than §3 describes. | Read subject §5; `internal/lexer/expect.go` `case Word`; `internal/lexer/lexer.go` `tokenize` (one `l.step(scanOne)` loop, one state). | open, block | b5b10428 |
| A22 | A | 3 | §8 Fixed inputs, Phase 0 gate | The speed kill only ever measures recognition: the chart's "cold parse" is defined as lexing, recognition and forest construction, but the only speed gate is phase 0's "recognition alone above 3×"; forest construction and interleaved lexing are never gated, and §9 cites phase 0 only. See also D-N1, E5, C-N11. | A chart that recognizes at 2× but parses at 5× passes every gate to phase 3. | Read §8 "Fixed inputs", phase 0 gate, §9 Speed row. | open, stand-aside | b5b10428 |
| A23 | A | 3 | §5 "The token cut" | pvm emits `Pod`, `DataSection` and `FormatBody` tokens; G1 has no rule for them and `:discard` covers only whitespace and `#` comments. `edit5k-standard.pl` keeps edit5k.pl's 70 POD blocks. The cut must declare these kinds trivia. See also F2, D5, E4, B-N11. | Phase 0's fixed input fails on its first POD block, or is rewritten again unrecorded. | `grep -c "^=" internal/psc/testdata/latency/edit5k.pl` → 140; Guacamole.pm:1871–1875. | open, stand-aside | b5b10428 |
| A24 | A | 3 | §7.4 "(pvm's existing Faithful check)" | Misattributed: `Faithful` is a token comparison of canon against source ("forgives exactly two differences") and never runs perl; the conformance plan's description of `TestCanonFidelityRatchet` as Deparse-based is stale. `Faithful` = true on 6 snippets (e.g. `0xfp-1`, tokenized as `0xfp - 1`) where Deparse shows RD's tree is wrong. Byproduct: 7 RD defects to file. See also D4, E3. | A reader reuses `Faithful` as the arbiter, which then passes the wrong trees it exists to catch. | `r3/faithful.txt`, `r3/arb.txt`; `internal/parse/fidelity.go` `Faithful`; `fidelity_test.go` `TestCanonFidelityRatchet`. | open, stand-aside | b5b10428 |
| A25 | A | 3 | §7.1 "participant A's count"; §10 "the round-1 ambiguity probes" | Review residue in a live design: the counts should stand on their own reproducible procedure. See also E16, D14, E9. | Harmless now; the design will be read without the minute, and these phrases point into an archive the reader lacks. | Subject lines 269 and 418. | open, stand-aside | b5b10428 |
| B-N8 | B | 3 | §5 Input model, rejoin on `State.Equal` | `State` includes the production `expect` machine; in interleaved mode expectation should come from the Earley set, so two paths perl's grammar would merge can differ in `expect`, never rejoin, and be refused at the length bound. Sound, but costs coverage. Drop `expect` from the compared state, or measure how often the bound refuses. See also D12, B-N1. | Avoidable refusals on branch cases. | Read checkpoint.go: `State` is the whole lexer, `expect` included. | open, stand-aside | b5b10428 |
| B-N9 | B | 3 | §5 conflict list; §10 "`foo <<EOF;` does not rejoin early" | In G1 `<<` is only `OpShift` and `DoubleDiamond` (`<<>>`), and Standard Perl excludes heredocs, so `<<` is not a lexical conflict, as with `{`. The §10 lattice test can't arise under G1 in phase 1; it needs a toy grammar or phase 3. `?` is likewise only `OpTriThen`. See also B-N6, A7. | A class component with nothing in it, and a phase-1 test that can't be written against the real grammar. | Read Guacamole.pm l.346-347, 1596, 1604; no G1 heredoc rule. | open, stand-aside | b5b10428 |
| B-N10 | B | 3 | §5 "The token cut" | Lexeme priority is part of the cut and is not encoded: G0 `SubNameNonQLike` and `SubName` match any identifier, and Marpa's `latm` with `:lexeme ~ OpKeywordX priority => 1` lets the keyword win. Letting keyword text satisfy `SubNameNonQLike` gives 2 derivations for `print(1);`, `time();`, `length($x);`; forbidding it refuses `$x->print()`. Statically fixable (keyword text satisfies `SubName` but not `SubNameNonQLike`), but the cut must say so; "Ties" covers interleaved only. Re-raises B-N2. See also D9, E10. | Phase 0 recognises every parenthesised builtin call twice, inflating speed and set numbers; phase 1's forest carries a spurious reading at each. | Ran collapse.py: `print(1)` 2; `time()` 2, or 1 keyword-only; `$x->print()` 1 or 0. Read Guacamole.pm l.405-425, `:lexeme … priority => 1` from l.1884. | open, stand-aside (must land in the cut's list before phase 0) | b5b10428 |
| B-N11 | B | 3 | §8 "Fixed inputs"; phase 0 gate | `edit5k-standard.pl` is not Standard Perl: only the 70 heredocs are rewritten, but edit5k.pl auto-quotes throughout (`name => 'block0'` at l.13; `$self->{name}`, `$_[0]->{limit}`); 351 lines match `\{[a-z_]+\}` or `\b[a-z_]+ =>`. standard.pm excludes auto-quoting and collapsed G1 refuses both forms. Everything the fixture needs must be rewritten and validated: G1 accepts the whole file and perl -c still passes. See also F2, D5, E4, A23. | The speed kill, the spike's first gate, fires or passes on a meaningless number. | Read edit5k.pl l.1-40, `grep -cE` → 351; standard.pm l.88-100. Ran collapse.py: `$x->{WORD};` 0, `$x->{'a'};` 1, `my %h = (WORD => 'a');` 0, `('a' => 'a')` 1. | open, block | b5b10428 |
| B-N12 | B | 3 | §8 phase 0 gate; §5 "One coupled choice" | Phase 0 lexes up front (lookahead 1 available) and decides whether interleaving "stays viable", but interleaving means lookahead 0 with Leo, a different cost regime (toy: 0.009 s vs 87.6 s). The kill doesn't say which lookahead; recognition over up-front tokens can run at lookahead 0 with Leo to emulate interleaving's recogniser cost, and the gate should require that figure before phase 1. See also D1, C-N11, E6. | Phase 0 passes on lookahead-1 numbers, and phase 1 builds an unmeasured lookahead-0 recogniser. | Read §5, §8 phase 0; ran rec probe in round 2. | open, stand-aside | b5b10428 |
| C-N10 | C | 3 | §7.4; §9 "Undecided" row; §8 phase 2 | The arbiter is one-sided: it decides only where RD's canon Deparses like the source, and then the differing chart span is wrong; no branch finds the chart right. Phase 2's probes come from open "perl disagrees" issues, where RD is wrong, so every disagreement there is undecided and the 10% kill fires at the phase-2 gate whether the chart is right or not. See also E1, E2, D4, F4. | The phase-2 gate kills the spike on its own probe corpus, or probes are scored elsewhere unstated; defect yield can't be verified and the M4–M6 evidence goes ungraded. | §7.4; §2 evidence 2; §8 phase 2 "drawn from open issues"; `internal/parse/fidelity.go:1-2`. | open, block | b5b10428 |
| C-N11 | C | 3 | §5 "Two modes"; §8 phase 0 and phase 1 | Speed and convergence are measured once, up front, in phase 0; phase 1 may switch to interleaved, which §5 says is quadratic without Leo and reuses fewer sets. Phase 0 has no "viable" criterion, phase 1 has no speed gate, and the 47951e0a every-step tripwire is gone. See also D1, D-N1, E6, A22, B-N12, A21. | Both the speed kill and the incremental verdict are decided on up-front numbers; §1's headline ("Whether that is affordable is the first thing this spike measures") has no pass condition. | §5 l.169-188; §8 phase 0 and 1 build lists; diff 47951e0a→b5b10428 (tripwire removed). | open, block | b5b10428 |
| C-N12 | C | 3 | §2 Output; §8 phase 4 | Under R8, M4–M6 negative with incremental positive means phase 3 never runs, so phase 4 ("over phase 3's corpora") has no corpus and no parity-level chart to fuzz. See also D-N3. | Whoever reaches that state decides what "positive incremental" leads to. | §2 Output; phase 3 and 4 headings; phase 4's first sentence. | open, stand-aside | b5b10428 |
| C-N13 | C | 3 | §8 phase 2 | Charging excluded features to each probe ("prototype-driven parses or paren-less calls for `first { ... } @list`") against a bar of "at most half the lines and half the places touched" of a mature RD biases the verdict and may decide it in advance; the equal split spreads cost only over probes that share it. See also D3, E17, F6. | The M4–M6 verdict can be negative for maturity reasons, not architectural ones; then the spike can't answer its question. | §8 phase 2, R7; round-1 F6(i). | open, stand-aside | b5b10428 |
| C-N14 | C | 3 | §8 phase 0 speed kill | "Unless removing the named hot spots brings it under 3×": if "removing" means subtracting profile time, any gap escapes by naming enough hot spots; if it means optimising and re-timing, it is sound. See also E5, F9. | The one cheap kill can be argued away after the numbers. | Phase 0 gate text. | open, stand-aside | b5b10428 |
| C-N15 | C | 3 | §9 Forest row; §3 rule 4 | The forest kill fires only "where perl's reading is decidable from facts pvm has", which is a resolver bug, not growth; readings ambiguous for lack of facts have no cap, and 4ec0955a rule 4's "A cap collapses runaway regions to an Error node" is removed. See also B-N7. | Many uses of an unknown imported sub can grow the forest without bound, with no kill or cap. | §9 row; 4ec0955a §3 rule 4 vs b5b10428 §3 rule 4. | open, stand-aside | b5b10428 |
| C-N16 | C | 3 | §10 Fuzzing | Stale phase reference: "recognizer and forest invariants from phase 0", but `chart/forest` is built in phase 1 and phase 0 is "recognition only". | Minor; whoever wires phase 0's CI looks for a forest fuzzer that can't exist yet. | §10; §8 phase 0 and 1. | open, stand-aside | b5b10428 |
| C-N17 | C | 3 | §10 Ratchets | `internal/chart` builds behind a `chart` tag so mainline `make test` doesn't run it, yet the lexer refactor it depends on is mainline work, so a mainline lexer change can break the chart unnoticed; where the chart code lives is unstated. See also E14, D18. | The chart rots silently, or untested tagged code sits beside pvm's 100% `make test` rule, close to the workaround CLAUDE.md forbids. | §10; pvm CLAUDE.md "Build/Test Commands", "Problem-Solving Philosophy". | open, stand-aside | b5b10428 |
| D-N1 | D | 3 | §8 Fixed inputs, phase 0 gate; §9 Speed | The revision deleted the every-step tripwire (47951e0a: "A cold-parse tripwire on Standard Perl inputs runs at every step, as a warning"). Speed is gated only in phase 0, on recognition alone over up-front tokens; the defined full cold parse is never used by a gate, and forest construction, interleaved lexing and `chart/resolve` are never measured against a threshold. See also A22, C-N11, E5, D1. | The configuration delivering §1's property could be 10× slower than RD and still reach parity with no kill. | Read `git diff 47951e0a b5b10428`; §8 Fixed inputs; phase 0 gate; §9 Speed row. | open, block | b5b10428 |
| D-N2 | D | 3 | §8 phase 1 corpus | Stub contents are unspecified: stubs without `$VERSION` give 41 errata (`use Module 4.10` fails the version check), stubs with `$VERSION = 999` give 23. See also A4, C-N7, D2. | The errata set, and so the scored corpus, depends on an unstated choice. | Ran: `D/rejects3.txt` (41) vs `D/rej999.txt` (23). | open, stand-aside | b5b10428 |
| D-N3 | D | 3 | §2 Output vs §8 phase 4 | R8 made M4–M6 independent of the incremental verdict but not the reverse: phase 4 runs over phase 3's corpora, and FuzzIncremental on the chart needs parity-level coverage, so incremental positive with M4–M6 negative can't run phase 4. See also C-N12, D6. | One-way coupling the doc doesn't state. | Read §2 Output; phase 4 first sentence. | open, stand-aside | b5b10428 |
| D-N4 | D | 3 | §5 token cut; §8 phase 1 gate | Several Standard Perl exclusions are lexical (limited delimiters such as `q@...@`, no space before a delimiter such as `q <foo>`, no nested delimiters), which pvm's lexer accepts; "every `parsent` case refused" then depends on the reviewed predicates encoding those limits, which the doc never requires. See also B-N10, E10. | Some parsent cases are accepted and the failure looks like a grammar fault, not a predicate list. | Read `standard.pm` "Limited delimiters", "No spaces ... before delimiters". | open, stand-aside | b5b10428 |
| E1 | E | 3 | §7.4; §9 "Undecided disagreements" | The 10% cap is incoherent, since a decided disagreement always means the chart is wrong. (a) If chart-lost disagreements are fixed, every remaining one is undecided, so the cap is really "zero disagreements". (b) If not fixed (phase 1 needs only that an arbiter exists; §9 kills only on fix-by-refusal), every extra chart error lowers the undecided ratio. See also C-N10, E2, D4. | Under (a) a hidden zero-tolerance rule nobody chose; under (b) it rewards misparses; either way 10% decides nothing. | Read §7.4, §8 phase 1 gate, §9 Silent misparse row; arithmetic inferred. | open, block | b5b10428 |
| E2 | E | 3 | §7.4; §9 | The arbiter only ever rules against the chart: it decides only where RD's canon is perl-equivalent, and then RD wins; where RD is wrong the case is undecided, so every chart-right/RD-wrong case pushes toward the kill. See also C-N10, E1, D4. | In phase 3 and on M4–M6 material the "more than 10% undecided" kill fires in proportion to how much better the chart is; the chart can never be credited a win. | Read §7.4; inferred. | open, block | b5b10428 |
| E3 | E | 3 | §7.4 | (1) "pvm's existing Faithful check" is a token comparison, not Deparse; Deparse,-p of canon exists only in a few tests. (2) "Where Deparse loses the form" goes to undecided, but the test is equality, which can't see a lost form: both sides fold equal and the case is decided for RD. (3) Granularity is unstated: Faithful and Deparse are per file, a disagreement is a span, so one unfaithful construct makes a whole file undecided. See also A24, D4. | Wrong-associativity cases are decided for RD; the implementer must choose between two checks, and the choice changes the verdict. | Read `fidelity.go` `Faithful`; `heredocsubless_test.go:39-51`, `canon_test.go:431`. Ran perl 5.44.0 Deparse,-p: `print((1+2)+3)` and `print 1+(2+3)` both → `print(6);`; `1+2*3` and `1+(2*3)` both → `print(7);`. Read `docs/measurements/2026-10-23-precedence-aware-canon/README.md`: perl.git t/ is 342/619 faithful. | open, block | b5b10428 |
| E4 | E | 3 | §8 Fixed inputs, Phase 0 | `edit5k-standard.pl` (heredocs rewritten) is not Standard Perl: the file auto-quotes throughout and uses brace-less prefix deref (`%$s`), both excluded; a G1 recognizer rejects it within 15 lines. See also F2, D5, B-N11, A23. | Phase 0's numbers time a recognizer that dies early, which is meaningless and looks "fast". | Ran on edit5k.pl: 281 bareword-before-`=>` lines, 350 bareword `{key}` lines, `%$s` at l.62, 134, 206…; l.13 `name    => 'block0',`. Read standard.pm ("`$foo{key}` # not ok", "`( foo => 'bar' )` # not ok", "`@$foo` # not ok"). Marpa::R2 not installed. | open, block (fix: "rewritten until it is Standard Perl", gated on the chart recognising the whole file) | b5b10428 |
| E5 | E | 3 | §8 Phase 0 gate, Fixed inputs; §9 Speed | The speed kill compares chart recognition alone with RD's `parse.Parse` (lex + parse + tree); the defined chart cold parse (lex + recognition + forest) is used by no gate, no later phase has a speed gate, and "unless removing the named hot spots…" can't be evaluated (measured after removal, or projected). See also A22, D-N1, C-N11, C-N14. | The comparison favours the chart, and the full cold path, including Leo recovery, is never held to a threshold. | Read §8, §9; RD's 23.89/33.42 ms is `BenchmarkLatencyParse` = `parse.Parse` (`docs/plans/2026-09-05-parser-conformance-plan.md:620`). | open, block (fix: re-run the same 3× gate on the defined cold parse at phase 1's gate) | b5b10428 |
| E6 | E | 3 | §5 chart/lex "Two modes, chosen by phase 0"; §8 phase 0/1 | Nothing in phase 0's gate chooses the mode, and "viable" has no criterion. Three things hang on it: (a) the mainline `internal/lexer` refactor; (b) phase 1's token-branch case, since the up-front lexer emits a flat stream, not a lattice; (c) phase 1's expectation-class gate, moot up front. §1 calls this "the first thing this spike measures". See also D1, C-N11, A21, B-N12. | Whether the refactor happens, and what phase 1 builds, becomes an unrecorded judgement. | Read §1 last sentence; §5 "Up front … No expectation classes are involved"; §8 phase 0 gate; phase 1 Build. | open, block | b5b10428 |
| E7 | E | 3 | §8 Phase 1 gate; §6 | The gate requires "every `parsent` case refused, with an Error node", but no phase builds error recovery (only §6's diagram: "`error` symbol, minimal skip"), and G1 has no `error` symbol, so adding one is an unowned grammar extension. | The gate can't be met with the listed components. | Read §8 phase 0/1 Build; ran `grep -n -i error lib/Guacamole.pm` → only `CHILD_ERROR_NATIVE`. | open, stand-aside | b5b10428 |
| E8 | E | 3 | §3 rule 4; §7.1; §7.3 | Rule 4 says "§7 says how each is measured", but the span diff and arbiter apply only where both produce one reading; inside ambiguous regions, "perl's reading is among the live ones" is checked only at marker sites ("necessary, not sufficient"), and the yield check is met by any well-formed forest. | Phase 1 is mostly unambiguous (Guacamole's `parses` asserts one tree), but in phase 3 and M4 "covered" is unmeasured except at marker sites. | Read §7.1, §7.3; Guacamole `Test.pm` `parses`: `is(scalar(@trees), 1, "parsed unambiguously")`. | open, stand-aside | b5b10428 |
| E9 | E | 3 | §7.1; §8 phase 1 Corpus | §7.1's 730 and its 68 marker sites were counted without the module stubs §8 prescribes (38 of 61 rejections are missing modules), and cite "participant A's count", which a reader can't look up. See also D14, A25. | §7.1's baseline and denominator differ from the corpus phase 1 scores. | Ran recorder with stub `Guacamole::Test` (and stub `Test::Fatal` for three files): 791/126, 769/124 unique, matching §8; `perl -c` under `:5.36`: parses 730/61, of which 38 "Can't locate"; parsent 80 accepted / 46 rejected. | open, stand-aside | b5b10428 |
| E10 | E | 3 | §4; §5 token cut, chart/grammar | "G1 rules are used as written above pvm's token level" is untrue of the doc's own example: `VarScalar ::= SigilScalar VarIdentExpr ElemSeq0` straddles the cut (`$x` one `Variable` token; `ElemSeq0 ::= Element*` the subscripts), so it must be rewritten, and the doc doesn't say which others straddle. The listed SLIF subset omits sequence rules (`X ::= Y*`, used twice), `:default ::= action =>` and `name =>` (4 uses). See also B-N2, B-N10, D-N4. | "As written" hides a set of grammar edits from review. | Read Guacamole.pm l.355, 384, 1512, 15; `internal/lexer/lexer.go` `Variable`. Ran a count of SLIF adverbs. | open, stand-aside | b5b10428 |
| E11 | E | 3 | §3 rule 3 vs §5 Ties | Rule 3 says "only compile-time facts may prune" and "Perl decides every case from such facts", but §5 Ties resolves with "perl's own heuristics (`{` and `intuit_more`)", which are textual heuristics; rule 6 covers lexeme priorities, not these. See also C-N6, D10. | A rule §5 breaks invites reading rule 3 as decorative. | Read §3 rules 3 and 6, §5 Ties; spec 03 §3.3; `internal/lexer/intuit.go`. | open, stand-aside | b5b10428 |
| E12 | E | 3 | §8 phase 0 convergence | "80% of 200 single-statement edits converge within the statement + 8 tokens", but the edit set is not fixed in advance and dominates the outcome (a token-count-preserving rename converges trivially); the metric takes the first matching set, so later divergence (e.g. right-recursive completion cascade at `}` with Leo off) is invisible until phase 4's measure (b). See also F9, D16. | The verdict can be steered by the choice of edits. | Read §8 phase 0, phase 4 (a)–(c); inferred. | open, stand-aside | b5b10428 |
| E13 | E | 3 | §5 Builtin shapes; §10 Isolation | The CORE.tpm cross-check imports `internal/parse`, but the shape table `coreShapes()` is unexported, so the check needs an `internal/parse` API change, contrary to §10's no-change clause. See also D7, A6. | A small, unlisted mainline change. | Read `internal/parse/declarations.go:712` `func coreShapes() map[string]Shape`. | open, stand-aside | b5b10428 |
| E14 | E | 3 | §10 Ratchets | The `chart` build tag keeps the code out of `make test` and also out of the pre-commit `go-vet-custom`, `staticcheck-linting` and `quick-go-check` hooks, which pass no tags. See also C-N17, D18. | Spike code escapes hooks CLAUDE.md calls non-negotiable. | Read `.pre-commit-config.yaml` hook ids; Makefile `test: go test ./... -count=1`. | open, stand-aside | b5b10428 |
| E15 | E | 3 | §12 | "`perly.y` writes `parser->expect` in 14 places" does not match E's count of 18 assignment lines at 94e5086608. Spec 03's eleven states are correct. See also A19, D13, B4 (A and D count 14 writes and 4 `==` reads). | A stale reference; it doesn't affect the spike. | Ran `grep -n "parser->expect" /opt/corpora/perl5/perly.y` → 18. | open, stand-aside | b5b10428 |
| E16 | E | 3 | Status line; §4; §7; §11; §12 | Dead weight answering critics: the Status line pointing at `docs/assessments/`; "participant A's count"; the Chalk `perl.bnf` lineage (28c6029e); "Chalk's re-emission lowers its SoN IR … does not serve here"; §11's B::SoN schema detail; "Kegel … Cited from memory". See also A25. | The document grows and goes stale as reviews accumulate (E9, E15 are symptoms). | Read those passages. | open, stand-aside | b5b10428 |
| E17 | E | 3 | §8 phase 2 | "Clearly cheaper" (at most half the lines and half the places touched) compares grammar-rule lines in a generated-table system with hand-written Go; "places touched" is undefined; who judges "a rule … that no Perl construct justifies" is unassigned. See also C-N13, D3, F6. | With 6–9 probes, a definitional choice can decide the M4–M6 verdict. | Read §8 phase 2. | open, stand-aside | b5b10428 |

### Round-4 status of rows open at b5b10428, and of released rows re-raised

Each status is in the raiser's own word, with its reason, as declared at e34f8233. Rows still released and not mentioned again are not repeated here.

| id | raiser | round-4 status | reason given |
|---|---|---|---|
| A18 | A | **re-raised** inside A27 (was released in round 3) | `edit5k-standard.pl` was removed (R11). Until the chart accepts edit5k.pl, which comes after the checkpoint, every speed ratio is on one-line snippets. |
| A21 | A | **released** | §5: up front "cannot express a lexical ambiguity; it resolves each one by its own default", and interleaved mode "arrives at the first ratchet step whose feature needs a token branch", timed by the every-step measure. Consequence accepted: at the waypoint both parsers share that lexer, so a disagreement from its defaults can't show there. |
| A22 | A | **released** | Speed end to end at every step. |
| A23 | A | **released** | `Pod`, `DataSection`, `FormatBody` are trivia. |
| A24 | A | **released** | The `Faithful` attribution is gone. Byproduct still open: the 7 round-3 RD defects, which A hasn't seen filed. |
| A25 | A | **released** | Review references removed. |
| B10 | B | **re-raised (stand-aside)** via B-N7 → B-N14 | See B-N14. |
| B-N7 | B | **re-raised as B-N14 (stand-aside)** | The cap is now per statement, as asked, but the product is an upper bound once packed nodes nest, and "inside" includes nested statements. |
| B-N8 | B | **released** | The compared state excludes the production `expect` field; the rest of `State` still takes part, so B-N1's soundness argument holds. |
| B-N9 | B | **released** | Conflict list is `/`, `<`, `%`, `&`, `*`, `-X`, `x`; `{`, `<<`, `?` join as the ratchet adds them; rejoin tested on a toy grammar. |
| B-N10 | B | **released** | collapse.py under the static rule gives 1 derivation each for `print(1);`, `time();`, `length($x);`, `return(1);`, `shift;`, `print 1, 2;`, `$x->print();`, `$x->print;`, `sub print {}`, `print->foo();`, `foo(1);`; §10 tests `print(1)`. (E23 raises the rule as too strict.) |
| B-N11 | B | **re-raised as B-N13 (block)** | The fixed input is gone, which was the bar, but nothing reaches file scale before the checkpoint. |
| B-N12 | B | **released**, remainder folded into B-N13 | Interleaved mode is timed when in force, at lookahead 0 with Leo; on what input is B-N13. |
| F2 | C | **held, changed form (block)** | Speed runs "on every file both parsers accept", but edit5k.pl needs auto-quoting (281 + 350 lines) and heredocs (70), both on the parity list after the checkpoint, so the pre-checkpoint speed set is one-liners measuring per-call setup. "Aggregate" is undefined (total time, mean of ratios, geometric mean). Bar: define the aggregate; add one realistic-size Standard-Perl file from the waypoint on. See also A27, B-N13, D5, E19, F-7, C-N19. |
| C-N10 | C | **released** | §7.4 is two-sided, with a branch deciding for the chart counted as defect yield, and §10 tests both directions. Note: the chart needs its own fully parenthesised emitter, and "the renaming changes the parse" needs a detection method. |
| C-N11 | C | **released** | The every-step speed measure times whichever mode is in force. |
| C-N12 | C | **released** | §2 Output: a positive incremental verdict without a positive M4–M6 verdict ends at the recommendation. |
| C-N13 | C | **re-raised as C-N18 (block)** | Both comparisons are reported and RD is costed from its PR, but the verdict rule doesn't say which comparison it uses. |
| C-N14 | C | **released** | The hot-spot clause is removed; plain 3×. Input problem is F2. |
| C-N15 | C | **released** | Rule 4's per-statement cap is restored, and §9 fires on more than 1% of statements collapsed, independent of decidability. |
| C-N16 | C | **released** | Forest invariants from the step that builds the forest. |
| C-N17 | C | **released** | `internal/chart` is mainline code in `make test` and the hooks; numeric ratchets run under `make chart-ratchet`. Accepted cost: chart tests can block mainline lexer work. See also A29, D-N6. |
| D1 | D | **released** | Interleaving arrives at the first step needing a branch and is timed every step against the 3× kill, which is the viability criterion asked for; its value depends on D5's fix. |
| D4 | D | **held, block (narrowed)** | Works: the renamed `my $y = $k1 + $k2 * $k3;` case decides for the chart. Fails: (a) `Deparse,-p` erases `my $x = @a` vs `my ($x) = @a` (Concise shows `aassign` only for the second), a silent "both legitimate" pass; (b) full parenthesisation changes meaning (`$x x 3` vs `(($x) x 3)`), flooding undecided, and "and so is the source" needs a reading; (c) statement scope loses context (`first { $_ } @l` with and without `use List::Util "first"`). Fix: Deparse the whole file with the disputed statement replaced; parens only where perl builds the same op tree, tested by Concise round-trip; compare with plain Deparse plus Concise op names; "the renamed source". Minor: string literals are not renamed. D accepts that the `${$h->{k}}` example is stale (author's check on 01a0ad52) and says it no longer matters, because §7.4 no longer rests on Faithful. See also A26, E18, E20, F-1, F-2, F-3, F-4. |
| D5 | D | **re-raised (narrowed), block** | Until parity the speed kill and convergence see only one-liners and M4–M6 case files; Earley's failure modes show only on long input; "aggregate ratio" undefined. Fix: at the waypoint, one long Standard-Perl input (concatenated accepted cases to at least 5 kLOC, or edit5k.pl rewritten until accepted), and aggregate = total chart time over total RD time on the same bytes; also change the convergence corpus (D-N5) and §9's Speed row. See also F2, A27, B-N13, E19, F-7. |
| D14 | D | **released** | The 730 figure and "participant A's count" are removed. |
| D19 | D | **re-raised, stand-aside** (was released in round 3) | R11 removed phase 0, so the first speed ratio comes after the waypoint, the largest step. Fix: let §9's speed kill fire once the recognizer runs on production tokens over D5's long input, within the waypoint. See also A28, F-14, C-N21. |
| D-N1 | D | **released** | End-to-end speed at every step in either mode; hot-spot escape gone. Input problem is D5. |
| D-N2 | D | **released** | §8 stub rule reproduces 23. |
| D-N3 | D | **released** | §2 Output states the one-way coupling. |
| D-N4 | D | **released** | Deferred to the token-cut implementing issue, where D placed it. |
| E1 | E | **released** | Four outcomes keep decided and both-legitimate entries in the denominator; "no disagreement decided against the chart" removes the reward reading. Residual (stand-aside, not re-raised): whether legitimate differences count in the denominator. |
| E2 | E | **released** | Two-sided; chart wins count as defect yield; §10 tests each direction. |
| E3 | E | **released** as raised; residue in E18 and E20 | Faithful attribution gone. Re-ran renamed: `print(($k1+$k2)+$k3)` vs `print($k1+($k2+$k3))` differ, the source matches the left form; `-` and `**` likewise distinguished. Per-statement granularity creates E18. |
| E4 | E | **released**, superseded by E19 | Fixed input gone, but edit5k.pl can't join before the recommendation. |
| E5 | E | **released**, superseded by E19 | End to end defined, every step, hot-spot clause removed. |
| E6 | E | **released** | The trigger is defined. Noted, not blocked: an M4 case needs a branch, so the mainline lexer refactor lands before the checkpoint while `m2-incremental` is in flight (30/38); that is perigrin's ruling to make. |
| E7 | E | **released** | An `error` terminal added by the generator, built at the waypoint. |
| E8 | E | **held (stand-aside)** | Under R12 the ambiguous M4 case (named unary vs list operator, unknown prototype) comes before the checkpoint, so "covered but undecided" is evaluated with only marker sites. See also F-5. |
| E9 | E | **released** | 730 and "participant A" removed; stubs declare a satisfying `$VERSION`. |
| E10 | E | **released** | Straddling rules split (`VarScalar ::= ScalarVar ElemSeq0`); SLIF subset covers sequences and `name =>`. |
| E11 | E | **released**, with a correction | Rule 3 admits deterministic lexer rules, but the example is wrong: `S_intuit_more` decides subscript `[`/`{` after a variable mainly inside patterns; the block-vs-hash heuristic is `yyl_leftcurly` (spec 03 §3.3). A one-word fix. |
| E12 | E | **re-raised**, folded into E19 | Edits are fixed in advance, but drawn from one-statement snippets (mean 13 bytes), so there are no tokens after the edit. |
| E13 | E | **released** | Listed as the third isolation exception in §5 and §10. |
| E14 | E | **released** | Build tag gone; `make chart-ratchet` separate. |
| E15 | E | **released**: E's round-3 count was wrong | `grep -c`: 18 occurrences, 14 writes, 4 `==` tests; E's round-3 pattern matched `==`. §12 is correct. |
| E16 | E | **released** | Status line, "participant A", Chalk lineage, re-emission sentence, B::SoN detail and Kegel removed. (F-12 counts 7 remaining ruling attributions.) |
| E17 | E | **held (stand-aside)** | "Places touched" undefined; grammar lines vs Go lines; who judges "a rule no Perl construct justifies" unassigned. See also C-N18, C-N22, F-6. |

### Round-4 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A26 | A | 4 | §7.4, the two-sided arbiter | "Emitted fully parenthesised" and "the renaming changes the parse" are undefined, and the obvious readings are unsound: (a) parens are not neutral (`(my $x) = …` list assignment, `(1) x 3`, print's first slot), and at statement level `Deparse,-p` maps `my $x = E` and `(my $x) = E` alike, so an over-parenthesising emitter passes; (b) no test for a renaming-caused parse change, and renaming can make a literal a filehandle (`print 1 -2` vs `print $k1 -$k2`); (c) renaming depends on pvm's literal boundaries (it splits `v1.4.0`); (d) strings are not renamed; (e) statement emitted alone or in place is unstated. See also D4, E18, E20, F-1, F-2, F-3, F-4. | Disagreements get "decided" when they shouldn't be, feeding defect yield and the silent-misparse kill. | Ran `cmd/rename` over 242 renamed `parses` snippets (225 perl-accepted): 3 stop compiling (`package Foo v1.$k1;`, its `{1}` form, `use $k1`); `print $k1 -$k2;` → `print($k1 (-$k2))`; `my $x = "a" . "b" x 3;` → `(my($x) = 'abbb')`; `while ((my $line) = (<>))` deparses as list assignment vs source's `defined((my $line = readline(ARGV)))`; `my $x = $k1;`, `my ($x) = $k1;`, `((my $x) = ($k1));` all → `(my($x) = $k1);`, Concise `aassign` for the last two. Perl 5.42.0. | open, block (parens around operator expressions only; compare with plain Deparse too; a mechanical `B::Concise,-exec` op-sequence check for renaming; never rename version/`use`/`require`/`package`/prototype positions; emit in place) | e34f8233 |
| A27 | A | 4 | §8 "Every step measures"; §8 Checkpoint; §9 speed and convergence rows | The checkpoint's incremental verdict has no multi-statement input: convergence edits come from the corpora's statements, which before the checkpoint are one-statement snippets and M4–M6 cases, so an edit reaches end of input and holds trivially or is undefined; speed times setup. edit5k.pl joins only after heredocs, after the checkpoint. Re-raises A18. See also B-N13, F2, D5, E19, F-7, F-8, C-N20, D-N5. | One of the spike's two verdicts is decided on inputs that can't fail; the 3× kill judged on setup cost. | Subject §8; response row F2/B-N11/D5/E4; rounds 2–3: edit5k.pl has 70 heredocs. | open, block (name a pre-checkpoint multi-statement input, e.g. restore `edit5k-standard.pl` as a corpus) | e34f8233 |
| A28 | A | 4 | §8 waypoint, against R1 | The first speed number now comes after the waypoint builds the recognizer, forest, error recovery, `chart/compare`, the arbiter and the oracle adapter; the recognition-only early kill is gone. The document should say R11 supersedes R1's early kill. See also D19, F-14, C-N21. | The cheapest exit comes after most of the old phase 1 is built. | Diff: "Phase 0: measure first" removed; §8 waypoint "The first step builds …". | open, stand-aside | e34f8233 |
| A29 | A | 4 | §10, first paragraph | `internal/chart` tests now run in mainline `make test`; under CLAUDE.md's 100% rule a mainline `internal/lexer` change that breaks a chart test blocks mainline, reversing the round-1 "cannot block mainline work" property unstated. See also D-N6, C-N17. | Coupling by surprise during the lexer refactor and later lexer work. | Subject §10; pvm CLAUDE.md "All test files MUST pass 100% before committing". | open, stand-aside | e34f8233 |
| B-N13 | B | 4 | §8 "Every step measures"; §8 Checkpoint; §9 Speed and convergence | Before the checkpoint every input is snippet-sized (Guacamole `parses` strings; M4–M6 steps add constructs, not files; edit5k.pl, `t/op`, CPAN only on parity steps). (a) Snippet speed measures constant cost, while quadratic right recursion at lookahead 0 needs long statement lists (25k tokens to reach 3 s). (b) Convergence on one-statement snippets holds by construction, so the incremental verdict can't say "keep the snapshot design". (c) §9 cites edit5k.pl figures no pre-checkpoint step measures. Fix: a validated file-scale Standard-Perl fixture, or a generated one. Re-raises B-N11. See also A27, F2, D5, E19, F-7. | Both checkpoint verdicts come from inputs where their failure modes can't appear. | Read §8; ran round-3 collapse.py: `$x->{WORD}` and `WORD =>` give 0; round-2 rec R la=0 25k tokens 3.0 s, 100k tokens 87.6 s. | open, block | e34f8233 |
| B-N14 | B | 4 | §3 rule 4; §9 Forest row | The per-statement count collapses whole subs and overcounts: (a) statements nest, so a `SubStatement` or `Condition` multiplies every ambiguity in its body and collapses at 2^10 though no inner statement nears the cap — count over the innermost statement; (b) the product of fan-outs is an upper bound (k nested two-way levels: k+1 readings vs 2^k); use an exact sum-product. Re-raises B-N7. See also E22, F-11, C-N15. | Spurious Error nodes over large constructs, and a 1% kill that may fire on a counting artefact. | Read §3 rule 4, §9; G1 `BlockStatement ::= … SubStatement \| Condition \| BlockNonEmpty`, `BlockNonEmpty ::= LBrace StatementSeq RBrace`. Arithmetic inferred. | open, stand-aside | e34f8233 |
| C-N18 | C | 4 | §8 Checkpoint | The verdict reports cost "with and without the reusable-infrastructure share" but doesn't say which decides "at least half these steps are clearly cheaper"; M4–M6 steps build exactly the share parity reuses, so the two can disagree on the steps in question. Re-raises C-N13. See also F-6, E17. | Whoever reads the table first picks the number; the parity commitment is decided after the data. | §8 Checkpoint; "On to parity" first sentence; "Every step measures" cost bullet. | open, block (name the deciding comparison; C suggests "with the share") | e34f8233 |
| C-N19 | C | 4 | §8 "Every step measures"; §9 Speed | "Every file both parsers accept" grows each step, so each aggregate is over a different set; a step admitting heavier files shifts the ratio by composition, and the 3× kill can fire on a step that made nothing slower. See also F2. | A kill or override rests on a number with no like-for-like baseline. | §8 speed bullet; §9 Speed row. | open, stand-aside | e34f8233 |
| C-N20 | C | 4 | §8 convergence; Checkpoint | The incremental verdict is vacuous at the checkpoint (one-statement files always converge because the file ends) and decides nothing there, since it only gates incremental work after parity, where it is re-measured. See also A27, D-N5, F-8, E19. | The 80% threshold can't fire where it is evaluated; low harm, but §2 calls it one of two verdicts. | §8 convergence bullet; Checkpoint; §2 Output; §8 Incremental heading. | open, stand-aside | e34f8233 |
| C-N21 | C | 4 | §8 Budgets; §8 Waypoint | Budgets are per step, with no cap on the whole and an oversized waypoint (recognizer, forest, error recovery, `chart/compare`, arbiter, oracle adapter, generator, token cut) under one estimate; about 6–9 M4–M6 and 9 parity steps follow, each estimated just before it starts, so the 2× rule may never trip while the total grows. See also F-14, D19, A28. | The "80–90% migration" shape: every step on budget, the spike never ends. | §8 Budgets, Waypoint; /home/user/perigrin/chalk/CLAUDE.md:293. | open, stand-aside | e34f8233 |
| C-N22 | C | 4 | §8 "Every step measures", cost bullet | RD's historical cost "from the PR that added them" is ill-defined for features built over many PRs (862 commits on `internal/parse` since 2026-09-15); "the PR" undercounts RD. See also E17, C-N18. | The comparison varies with which PRs are counted. | §8 cost bullet; `git log origin/pu -- internal/parse \| wc -l` = 862. | open, stand-aside | e34f8233 |
| D-N5 | D | 4 | §8 convergence; checkpoint | Convergence over "200 single-statement edits per corpus, chosen ... from the corpus's statements" on one-statement Guacamole snippets is satisfied trivially within "the edited statement plus 8 tokens" because the file ends. (D's own id: N1.) See also C-N20, F-8, A27, E12. | The checkpoint's incremental verdict is vacuously positive; harmless only because incremental waits for parity. | Read §8; ran round-2 recorder: cases are single statements. | open, stand-aside | e34f8233 |
| D-N6 | D | 4 | §10 | With `internal/chart` in mainline `make test` and the hooks, every `internal/lexer` change must keep the spike's tests green; `make chart-ratchet` "runs on any change under `internal/chart` or `internal/lexer`", but nothing says what runs it. (D's own id: N2.) See also A29, C-N17. | Spike maintenance taxes mainline lexer work, and the ratchet is an obligation with no runner. | Read §10 para 2; CLAUDE.md "Test Failure Protocol". | open, stand-aside | e34f8233 |
| E18 | E | 4 | §7.4 | The arbiter deparses "the statement a disagreement falls in" alone, losing the compile-time context (prototypes, imports, declared subs, features) that decides the M4 cases, so it can rule for the wrong parser; "renaming changes the parse" doesn't catch it, since isolation changes the parse, not renaming. Carrying context means substituting into the file or carrying preceding declarations, then declaring `$kN` under `use strict`, and stating the renaming test. See also F-3, D4, A26. | Spurious "decided against the chart" entries in a parity rule requiring zero, and spurious defect yield. | Ran Deparse,-p: `first { $_ } @l;` → indirect method, with `use List::Util 'first';` → `&first(sub { $_; } , @l)`; `f $k1, $k2;` → `($k1->f, $k2)`, after `sub f($);` → `(&f($k1), $k2)`; `use strict; print $k1+1;` → `Global symbol "$k1" requires explicit package name`. | open, block | e34f8233 |
| E19 | E | 4 | §8 "Every step measures", Checkpoint; §9 Speed, No incremental convergence | Before the checkpoint "every file both parsers accept" is only Guacamole snippets and M4–M6 repros (edit5k.pl and T1 need auto-quoting and heredocs, after the checkpoint); on 13-byte inputs the 3× ratio measures fixed overhead, not Leo at lookahead 0, right-recursive `StatementSeq` or interning, and convergence is degenerate (E12). Adding a Standard-Perl rewrite of edit5k.pl (auto-quotes, `%$s`, heredocs) satisfies R11; moving the incremental verdict to parity would fit §2. Re-raises E4, E5, E12. See also A27, B-N13, F2, D5, F-7. | The verdicts that fund parity rest on toy timings and toy edits; a slow-asymptote chart is found only after parity is paid for. | Ran: recorded `parses` corpus, 791 cases, mean 13.1 bytes, max 70. Read §8 "On to parity". Round 3: edit5k.pl 281 + 350 lines. | open, block | e34f8233 |
| E20 | E | 4 | §7.4 | Residual blind spots: (a) only literals outside quote-likes are renamed, so strings, `use constant` subs and `__LINE__` still fold; (b) Deparse,-p is not injective (`my $x = @a` vs `(my $x) = @a`); (c) "fully parenthesised" is unsafe (`($x) x 3` vs `$x x 3`), and neither the chart's emitter nor a fully-parenthesising RD emitter exists (RD's `Canon` parenthesises by binding power); §5 lists neither. See also D4, A26, F-1, F-4. | Real disagreements waved through as "legitimate"; a naive emitter makes correct readings lose. | Ran: `print 'a'.('b'.'c')` and `print(('a'.'b').'c')` both → `print('abc')`; `use constant {A=>1,B=>2,C=>3}; print A+(B+C)` and `print((A+B)+C)` both → `print(6)`; `my $x = @a` and `(my $x) = @a` both → `(my($x) = @a)`; `($x) x 3` → `(($x) x 3)`, `$x x 3` → `($x x 3)`. Read `internal/parse/canon.go`. | open, stand-aside | e34f8233 |
| E21 | E | 4 | §9 Context leaking row vs §8 | §9's signal "Set reuse falling as steps add features" refers to a measure §8 no longer takes (every-step measures are speed, cost, yield, convergence); the old phase 0 measured set reuse. | The context-leak risk has no signal. | Read §8 bullets; §9 row; diff. | open, stand-aside | e34f8233 |
| E22 | E | 4 | §3 rule 4; §9 Forest blow-up | "The product over the packed nodes inside it" overcounts nested ambiguity (a two-way node with a two-way node in each alternative has 4 readings, product 8); round 3's sum-product wording was correct. Whether counted before or after `chart/resolve` is unstated. See also B-N14, F-11. | Early collapse under the 2^10 cap; the 1% kill fires on a counting artefact. | Read §3 rule 4, §9; arithmetic inferred. | open, stand-aside | e34f8233 |
| E23 | E | 4 | §5 token cut | "Keyword text satisfies `SubName` … but not `SubNameNonQLike`" is stricter than Marpa's priority, which arbitrates only among lexemes acceptable at a position: after `PackageSep` in `SubNameCallExpr ::= SubNameCallExpr PackageSep SubNameNonQLike` no keyword lexeme is acceptable, so Guacamole accepts `Time::HiRes::time()` and `POSIX::exit(0)` and the static rule rejects them; Guacamole's suite has no such case. See also B-N10. | Package-qualified calls to keyword-named subs are refused from the waypoint on. | Read Guacamole.pm l.400–431; 217 `OpKeyword… ~ '…'` lexemes. Ran grep of the recorded corpus for `::name(`. Guacamole not run. | open, stand-aside | e34f8233 |
| F-1 | F | 4 | §7.4 | Deparse,-p text can't tell apart readings perl builds differently: scalar and list assignment deparse alike, so "the parser whose emission deparses as the source does is right" calls both right. See also D4, A26, E20. | A real misparse on either side is classed a legitimate shape difference; WRONG undercounted. | Ran: `my @a; my $x = @a;` and `my @a; ((my $x) = @a);` both → `(my($x) = @a);`; Concise `padsv_store` vs `aassign`; at runtime `$x` gets 3 and `($y)` gets 5. `internal/parseoracle/README.md` compares op trees for this reason. | open, block | e34f8233 |
| F-2 | F | 4 | §7.4 | Renaming literals changes perl's own parse of the disputed ambiguity, and no procedure detects "the renaming changes the parse" (the original folds, so deparses can't be compared). See also A26, D4. | The arbiter decides against the correct parser: a wrong "against the chart" fires the silent-misparse row; a wrong "for the chart" inflates defect yield. | Ran: `sub f { { 1, 2 } }` → `+{1, 2}`, renamed `{ $k1, $k2 }` → bare block; `map { 1, 2 } @x` a syntax error, renamed parses. Inferred: `use Module 4.10` → `use Module $k1`, `require 5.006` → `require $k1`; autoquoted barewords unclear. | open, block | e34f8233 |
| F-3 | F | 4 | §7.4 | The arbiter's scope is "the statement a disagreement falls in", but perl's reading depends on the rest of the file, exactly in the context-dependent M4 cases. Fix: whole file with one statement substituted. See also E18, D4, A26. | The arbiter adjudicates a different program; M4 evidence arbitrated wrongly. | Ran: `my ($x,$y); my @r = (foo $x, $y);` → `($x->foo, $y)`; with `sub foo($){}` → `(foo($x), $y)`; with `sub foo{}` → `foo($x, $y)`; `first { … } @l` without/with `use List::Util qw(first)`. | open, block ("folds into F-1/F-2 rewrite") | e34f8233 |
| F-4 | F | 4 | §7.4; §5 | The arbiter needs an emitter per parser that writes a reading as Perl without changing its meaning; neither exists and §5 lists neither. Parens aren't neutral (`(my $x) =`, `($k1) x $k2`, `print (…) .`), and block vs hash and indirect objects are not grouping questions, so "fully parenthesised" is the wrong specification. See also E20, A26, D4. | An emitter bug makes a parser "wrong"; two new trusted emitters tested only by §10's hand-built trees. | Read `internal/parse/canon.go` (parens "only where the source had them or binding power requires", README.md:134-137); grep of `internal/` found no fully-parenthesised mode. Ran `(($k1) x $k2)` stays list repetition. | open, stand-aside | e34f8233 |
| F-5 | F | 4 | §7.3; §7.4; rule 4 | Disagreements are diffed only "wherever both produce one reading"; when the chart is ambiguous, nothing checks perl's reading is among the live ones except the oracle's five marker kinds. See also E8. | An ambiguous chart escapes the arbiter; rule 4's WRONG is unmeasured outside marker sites. | Read §7.1 "necessary, not sufficient"; §7.3. | open, stand-aside | e34f8233 |
| F-6 | F | 4 | §8 Checkpoint | The M4–M6 verdict has two inputs (with and without the reusable share) and one threshold, and the doc doesn't say which counts; the chart builds exclusions RD already has, so the readings may give opposite verdicts. See also C-N18, E17. | The main verdict is not decidable from the document; a post-data judgment. | Read §8 Checkpoint and M4–M6 paragraph; inferred: `first { … } @list` alone needs paren-less user-sub calls plus `&`-prototype parsing on the chart side. | open, block | e34f8233 |
| F-7 | F | 4 | §8; §9 Speed | No real file is measured for speed before the checkpoint: edit5k.pl has 70 heredocs, which come after the checkpoint; until then fixed overhead dominates. Fix: a fixed realistic Standard-Perl file (e.g. edit5k with heredocs lowered to strings, generated once and checked in) from the waypoint on. See also F2, A27, B-N13, D5, E19. | The 3× kill fires or not on noise; the YAEP 13.6 ms extrapolation remains the only speed number at the decision point. | Ran `grep -cE "<<['\"A-Z~]" internal/psc/testdata/latency/edit5k.pl` → 70. Read §8 parity list; 23.89/33.42 ms match conformance-plan:620. | open, block | e34f8233 |
| F-8 | F | 4 | §8 convergence | Convergence is degenerate on pre-checkpoint corpora (single statements); edits are "chosen before the first step" but M4–M6 cases are "fixed before the waypoint exits", later; "keep the snapshot design" is undefined. See also C-N20, D-N5, A27, E12. | The checkpoint's incremental verdict is trivially positive or undefined. | Ran recorder with stub `Guacamole::Test`: every case one statement or fragment. Read §8 bullets. | open, stand-aside | e34f8233 |
| F-9 | F | 4 | §8 waypoint exit | The expectation-class check is undefined for states whose conflicts aren't in the grammar yet: five of the eleven states (XBLOCK, XATTRBLOCK, XATTRTERM, XTERMBLOCK, XBLOCKTERM) settle `{` or attribute questions, and `{` is deferred, so either ">2" fires by construction or the check covers about six states. | A gate that always fires or can't. | Read spec 03 §3.1.2; §5. | open, stand-aside | e34f8233 |
| F-10 | F | 4 | §8 waypoint | "A `parses` case perl still rejects is a Guacamole erratum" mislabels most: rejections are mostly compile-time semantics or BEGIN execution; the real errata are where perl parses differently (`<<$fh>>`, `<<STDIN>>` are heredocs to perl). A `parses` case perl accepts with a different tree is caught only by §7. See also A4, D2. | Harmless for scoring; Guacamole's labels are weaker evidence than implied. | Ran each unique case under `use feature ':5.36'`, perl 5.44.0: 708 of 769 `parses` accepted; 61 rejected = 38 missing module + 23 (12 invalid `:attr`, `no 5.010`, `Can't modify logical and`, `$#`/`$*`, `1.4.0` version format, `<<STDIN>>` heredoc readings); 79 of 124 `parsent` compile; 791/126 confirmed. | open, stand-aside | e34f8233 |
| F-11 | F | 4 | §9 | The 10% undecided threshold sits on a small denominator at the checkpoint (1 of 5 kills the spike), and "the product over the packed nodes inside it" overcounts nested ambiguity. See also B-N14, E22. | The cap's 1% is computed on an inflated count. | Inferred from §8 corpus sizes. | open, stand-aside | e34f8233 |
| F-12 | F | 4 | §1–§12 (whole document) | Dead weight: "(perigrin's ruling, 2026-10-29)" 7 times; §4's YAEP/Chalk inventory and §11 B::SoN written for a reviewer; five of §9's nine rows restate §8 thresholds. See also E16, A25. | Drift between §8 and §9; readers spend effort on provenance. | Read the section texts. | open, stand-aside | e34f8233 |
| F-13 | F | 4 | §5 chart/lex | Interleaved mode doesn't say who maintains the context facts the lexer consumes (declared subs, constant subs, bareword ops): today the lexer derives them; in interleaved mode the chart recognises declarations, so facts flow chart → lexer, which rule 2 doesn't address. | Rule 2's "two places only" is violated or carried silently by the lexer. | Read `internal/lexer/checkpoint.go` `State.Clone` (clones `barewordOps`, `lexSubs`, `constSubs`). | open, stand-aside | e34f8233 |
| F-14 | F | 4 | §8 first step | Not startable at a measurable size: the first step builds the grammar reader, the token cut (389 G0 `~` rules; the SLIF subset omits G0 rules and character classes), generator, recognizer, forest, error recovery, compare, arbiter and oracle adapter before any number exists; the only control is the 2× budget. See also C-N21, D19, A28. | The first evidence arrives after the most expensive step. | Ran `grep -cE '^\s*[A-Za-z0-9_]+\s+~' lib/Guacamole.pm` → 389; `::=` lines 440; `OpKeyword*Expr ::=` 197. | open, stand-aside | e34f8233 |

### Round-5 status of rows open at e34f8233, and of released rows re-raised

Each status is in the raiser's own word, with its reason, as declared at 7a430aa4.

| id | raiser | round-5 status | reason given |
|---|---|---|---|
| A18 | A | **released** with A27 | Same evidence as A27. |
| A26 | A | **released** (moot) | R14 drops the arbiter; `grep -n -i "arbiter\|Deparse"` on 7a430aa4 finds no remaining reference. Its cost is A32. |
| A27 | A | **released** | §8 "A realistic input from the start": std5k.pl from `latencyFixture`, with speed, set reuse and convergence measured on it, and the kill on "the aggregate ratio, or std5k.pl's". Re-checked the template: heredoc, POD (trivia), `printf {$out}` (in G1); from the parts read, the heredoc is the only exclusion (inferred). Caveat in A31. (G-4 reads the template as also using autoquoted keys and paren-less calls.) |
| A28 | A | **released** | The speed kill is live once the recognizer runs on std5k.pl, inside the step; recognition-only vs RD's full parse favours the chart, so it can't fire falsely early. |
| A29 | A | **released** | §10 states the coupling is intended. |
| B10 | B | **released** via B-N7 / B-N14 | See B-N14. |
| B-N7 | B | **released** via B-N14 | See B-N14. |
| B-N11 | B | **released** via B-N13 | See B-N13. |
| B-N13 | B | **released** | std5k.pl is in the measures before the checkpoint, "valid only while both parsers accept it and perl compiles it"; the kill covers it; convergence has a multi-statement corpus; each step records the previous-set ratio. Not verified: that the template can be written out and accepted by G1; acceptance is a precondition. |
| B-N14 | B | **released** | Rule 4: sum-product over the statement's own packed nodes, nested statements counted and collapsed on their own; §9 adds "at least 5". At the waypoint, with no resolver, the cap counts raw readings, which is consistent. |
| F2 | C | **re-raised, narrowed (stand-aside)** | Half (b) met (std5k.pl from the start, §9 kills on std5k.pl's ratio). Half (a) not: "aggregate ratio" still undefined, and pre-checkpoint other files are snippets that "time setup rather than throughput". Downgraded because the std5k.pl arm is well defined and the override exists. Input validity is C-N23. |
| C-N13 | C | **released** via C-N18 | See C-N18. |
| C-N18 | C | **released** | "Counting its whole cost including the infrastructure share; the figure without that share is reported beside it." |
| C-N19 | C | **released** | Each step records the aggregate, std5k.pl's and the previous-set ratio. std5k.pl is a like-for-like baseline only if fixed (C-N23). |
| C-N20 | C | **re-raised as C-N23 (block)** | Edits now come from std5k.pl and multi-statement corpora, but pre-checkpoint std5k.pl is the only multi-statement input, and its repetition makes convergence close to guaranteed (inferred). |
| C-N21 | C | **held (stand-aside), narrowed** | Released for the leg to the checkpoint (a whole-ratchet estimate is recorded). Held for the parity leg (about 9 features plus require/use loading), which has per-step estimates only. |
| C-N22 | C | **released** | RD cost is "the sum over every PR whose title or `Implements:` trailer names the feature, with the list recorded". (G-3 reads `Implements:` trailers in pvm as carrying issue UUIDs, not feature names.) |
| D4 | D | **released** | R14 drops the arbiter, and the text carries it out (§7.4 and emitters gone; §7.3 "recorded and reviewed, not gated"; yield "oracle sites decided for the chart"; §10 compare and argument-marker tests; §9 undecided row gone). D's round-4 objections no longer apply to anything in the doc; rulings are not D's to reopen. Residue is D-N10. |
| D5 | D | **released** | R13: std5k.pl with a well-defined acceptance test; the generator is one template repeated (`latency_fixture_test.go:110-131`), and `TestLatencyFixtureMatchesGenerator` is the pattern. Aggregate ratio still has no formula, but std5k.pl's own ratio is a separate kill. Residue is D-N7. |
| D19 | D | **released** | The speed kill is live once the recognizer runs on std5k.pl; a recognition-only ratio is a lower bound. |
| D-N5 | D | **released** | Convergence edits come from std5k.pl and multi-statement corpora. |
| D-N6 | D | **released** | §10 says the coupling is intended; a CI job and the orchestrating session run `make chart-ratchet`. |
| E8 | E | **released** | With the arbiter gone, parity gates only on the oracle and yield; §7.1 keeps "necessary, not sufficient"; R15 widens the markers; the M4–M6 verdict rests on cost. |
| E11 | E | **re-raised (stand-aside)** (was released in round 4) | Rule 3 still says "(`intuit_more` for `{`)"; the block-vs-hash `{` rule is `yyl_leftcurly` (spec 03 §3.3). The one-word fix flagged in round 4 was not taken. Read toke.c above `S_intuit_more`; subject l.85. |
| E12 | E | **released** | Edits drawn from std5k.pl and multi-statement corpora, fixed before the first step. |
| E17 | E | **held (stand-aside, minor)** | Files touched defined, perigrin judges the hack rule, whole cost decides; but whether "at most half the lines" means the sum or each kind (Go and grammar) is unstated. See also G-3. |
| E18 | E | **released** | Moot under R14. |
| E19 | E | **released** | R13 implemented; a recognition-only early measure is a lower bound, so an early firing is sound and does not reintroduce E5. |
| E20 | E | **released** | Moot with §7.4. |
| E21 | E | **released** | Set reuse is an every-step measure. |
| E22 | E | **released** | Sum-product over the statement's own packed nodes, after `chart/resolve`. |
| E23 | E | **released** | The keyword exclusion applies only to unqualified `SubNameNonQLike`, matching Guacamole.pm l.401–402. |
| F-1 | F | **released** | R14 removes §7.4; a grep finds no `arbiter`, `Deparse` or `10%`. Residue, not a reopening: scalar vs list assignment is a question of context, which no marker grades; §7.1 could say so. See also A32, D-N10. |
| F-2 | F | **released** | R14. |
| F-3 | F | **released** | R14. |
| F-4 | F | **released** | R14 removes the emitters. |
| F-5 | F | **held, stand-aside** | Rule 4 still defines WRONG broadly; measured WRONG is at marker sites only, and the response concedes "covered but undecided" goes unchecked outside them. See also A32, D-N10, G-5. |
| F-6 | F | **released** | Whole cost decides, with the other figure beside it. |
| F-7 | F | **released, with a new caveat (F-16)** | std5k.pl from the first step, and §9 kills on its ratio. |
| F-8 | F | **released** | Edits from std5k.pl and multi-statement corpora. Residues, not blocking: M4–M6 corpora are "fixed before the waypoint exits", so their edits can't be "fixed before the first step"; "M2 keeps its snapshot design" misnames M2's mechanism (in spec 06 §6.2.3 "snapshot" is the version fence; M2's mechanism is the token cache plus damage/repair). |
| F-9 | F | **released** | Only states whose conflicts are in force are compared, repeating as they join. |
| F-10 | F | **released** | Reworded as asked. |
| F-11 | F | **released** | The undecided ratio is gone; the forest kill needs at least 5 statements as well as 1%. |
| F-12 | F | **released** | One pointer replaces the ruling attributions; §9 drops the restating rows; keeping §4 and §11 is defensible. |
| F-13 | F | **released** | `chart/lex` owns the context facts and the chart never feeds them back. Inside a branch the paths might derive different facts; that is the branch-adding issue's. |
| F-14 | F | **released** | The speed kill is live inside the step; further splitting is refinement's. |

### Round-5 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A30 | A | 5 | §7.1, the sixth marker (R15) | Graded on RD too, so it becomes a mainline WRONG = 0 gate before the M4–M6 steps unless RD can hedge it (`unresolved` → wider). The definition also needs: perl's flattened operand count (`foo((1,2), 3)` 3 operand ops; `foo(1+2, 3)` 2 after folding); "each call" to include method calls (`foo $x, $y` undeclared is `method_named`); calls matched within a statement, since ops are attributed by statement line. The filehandle half is observable (`S` flag). See also C-N26, D-N9, G-1, G-2, G-10, F-18. | A mainline prerequisite with no estimate, or a marker without a hedge path that turns mainline red. | Ran Concise: `foo((1,2), 3)`, `foo(1+2, 3)`, `foo $x, $y` (method_named), `print $fh $x` (`print vKS`) vs `print $x` (`print vK`). Read conformance plan §5 M1; `verdict.go` `BucketWider`. | open, stand-aside | 7a430aa4 |
| A31 | A | 5 | §8 std5k.pl; §10 Isolation | `latencyFixture` is an unexported function in a `_test.go` file of `internal/psc`; producing std5k.pl changes `internal/psc` or copies the template, and §10 lists three exceptions, none this. See also C-N24, D-N8, E24, G-4. | A fourth unlisted exception, or a copied generator drifting from edit5k.pl, which the speed kill compares against. | Read `internal/psc/latency_fixture_test.go` `func latencyFixture() []byte`, `TestLatencyFixtureMatchesGenerator`. | open, stand-aside | 7a430aa4 |
| A32 | A | 5 | §3 rule 4 vs §7 (R14, R15) | With the arbiter gone a grouping-only WRONG is invisible to every gate: rule 4 still defines WRONG as asserting "what perl does not", but §7 grades only markers and argument structure. pvm's own example `print (1+2)*3` vs `print((1+2)*3)` (conformance plan M3) gives `print` one argument both ways, no marker differs, yet it is a lint finding. About stating R14's consequence, not reversing it. See also D-N10, G-5, F-5, C-N25, F-1. | A positive verdict or parity claim read as covering grouping correctness (prior-art §A1.3's coverage/fidelity confusion). | Read §3 rule 4, §7.1, §7.3; conformance plan §5 M3. Inferred from Concise: one operand op in both readings. | open, stand-aside | 7a430aa4 |
| B-N15 | B | 5 | §8 std5k.pl; §9 Speed and Context-leak rows | std5k.pl is one block repeated about 70 times (`latencyFixture` loops `fmt.Fprintf(&b, latencyBlock, n)`), so after the first copy nearly every Earley set interns and nearly every transition lookup hits; RD gets no such benefit, so the only file-scale pre-checkpoint ratio is biased toward the chart and set reuse starts inflated (as with YAEP's `gen.c` ×10). The toy grammar couldn't show the size. Remedy: record cache hit rate and set reuse beside the ratio and time once with the cache off, or vary the blocks. (B's own id: N15.) See also C-N23, D-N7, E25, F-16, G-4. | The speed kill under-fires on its own input; cache-dependent cost shows first on edit5k.pl and parity corpora, after the recommendation. | Read `latency_fixture_test.go` l.110-131; edit5k.pl 70 `package Fixture::Block`; YAEP README; Internals.txt item 8. Ran `probe/rec2`, R la=1, 20 000 statements: identical 11 unique sets 0.007 s; varied 39 unique sets 0.015 s over 2.4× tokens. Inconclusive. | open, stand-aside. Ruling R16 (made after this row was raised) accepts the repetition; not a release. | 7a430aa4 |
| C-N23 | C | 5 | §8 "A realistic input from the start", speed and convergence; §9 Speed | std5k.pl is realistic in size but not variety: `latencyFixture` appends `latencyBlock` to 5,000 lines substituting only the index (`%[1]d`: package name, `$VERSION`, `block%d`, `limit => %d`); identical token-kind sequences give identical (core, distances) sets, so interning and the cache hit nearly everywhere after block one, and convergence edits re-match almost at once. Re-raises C-N20. See also B-N15, D-N7, E25, F-16, G-4, F2. | Both the std5k.pl speed arm and the incremental verdict are evaluated on the chart's best case, biased to pass where continuing to parity is decided. | Read `latency_fixture_test.go:28-34`, `108-118`; `grep -c '^package Fixture::Block' edit5k.pl` = 70; §5 toy-probe sentence. Inferred: the interned sets coincide. Not measured. | open, block (several templates or seeded variations, or report beside the first block alone with the unique-set count). Ruling R16 (made after this row was raised) accepts the repetition; not a release. | 7a430aa4 |
| C-N24 | C | 5 | §10 Isolation vs §8 and §7.1 | The isolation list lacks two exceptions the revision creates: R13's std5k.pl comes from `latencyFixture` in an `internal/psc` test file; R15 grades the marker on RD, whose subject is `internal/parse/cmd/subject` (`internalParsePackage`), so `internal/parse` changes. See also A31, D-N8, E24. | §10's isolation claim is false as written; an implementer violates it or routes around it, which CLAUDE.md forbids. | Read §10; `internal/parseoracle/subject.go` `internalParsePackage`; `latency_fixture_test.go:110`. | open, stand-aside | 7a430aa4 |
| C-N25 | C | 5 | §7.1 and §8 Waypoint exit, under R14 | The waypoint exit is graded almost entirely by acceptance: the argument marker arrives "Before the M4–M6 steps", after the waypoint, so the oracle grades five kinds there (A2: 68 of 730 cases carry any site); yield holds for any well-formed SPPF; the span diff is "recorded and reviewed, not gated". C records the ruling's consequence and does not reopen it. See also A32, G-9, F-15. | A grammar that misgroups non-marker structure passes the waypoint; §2's "every step produces evidence" reads stronger than it is at step 1. | Read §7.1, §7.3, §8 Waypoint; ledger row A2. | open, stand-aside | 7a430aa4 |
| C-N26 | C | 5 | §7.1, R15 | The argument marker, "graded on recursive descent too, in both directions", will turn the oracle ratchet red on mainline unless it lands with a baseline, since RD has known attachment defects (the ~49 open "perl disagrees" issues). See also A30, G-10, D-N9. | The first commit is red (breaks the 100% rule) or lands with WRONG > 0 ratcheted, contradicting "WRONG = 0" unless RD's figure is the bar. | Read `internal/parseoracle/gate_test.go`, `ratchet_test.go`; §1's count. Inferred: RD scores WRONG on the new kind. | open, stand-aside | 7a430aa4 |
| D-N7 | D | 5 | §8 std5k.pl; §9 Speed | `latencyFixture` writes one block template repeated about 70 times, changing only the numeric suffix; interning and the transition cache should hit almost every set after the first block, so std5k.pl's ratio and set reuse overstate the chart, making the speed kill less likely to fire. See also B-N15, C-N23, E25, F-16, G-4. | A false pass on "a speed worth pursuing" at the checkpoint. | Read `latency_fixture_test.go:114-116`. Inferred from §5's design, not measured. | open, stand-aside. Ruling R16 (made after this row was raised) accepts the repetition; not a release. | 7a430aa4 |
| D-N8 | D | 5 | §10 Isolation; §5 | §10 still lists "three exceptions", but R15 needs RD's subject (`internal/parse/sites.go`, `internal/parse/cmd/subject/main.go`) to emit the marker, and R13 generates std5k.pl from `internal/psc/latency_fixture_test.go`. See also A31, C-N24, E24. | The isolation claim is false as written; a reviewer of the first such PR has no text to point to. | Ran `grep -rln 'SubjectCallSite\|call_sites'` over non-test Go files → `internal/parse/sites.go`, `internal/parse/cmd/subject/main.go`, two `parseoracle` files. Read §10 l.435–436. | open, stand-aside | 7a430aa4 |
| D-N9 | D | 5 | §7.1 argument marker | "Argument count of each call ... from perl's op tree" doesn't define what is counted: `foo((1,2),3)` has two direct operands under the call's `ex-list` but three constants in execution order; `foo(qw(a b c))` one `ex-list`, three constants. The oracle pairs by statement start line (`groupByStatement`), so `foo(bar($x), 1)` needs a per-call identity; graded both ways on RD, it may show WRONG > 0 on mainline. See also A30, G-1, F-18, C-N26, E26. | Spurious WRONG on either parser, or a mainline ratchet going red with no stated handling. | Ran `perl -MO=Concise` on the cases; read `compare_facts.go` `groupByStatement`; §1 "WRONG = 0". | open, stand-aside | 7a430aa4 |
| D-N10 | D | 5 | §3 rules 4 and 5 vs §7 | Rule 4 ("A reading is WRONG if it asserts what perl does not") and rule 5 ("never misparse") remain, but under R14 WRONG is graded only at the six marker kinds; §9's silent-misparse signal is "A `parsent` case accepted; WRONG > 0"; grouping misparses elsewhere are recorded, never graded. See also A32, F-5, G-5. | Readers take rules 4 and 5 as checked everywhere. | Read §3 rules 4–5; §7 intro; §9 Silent misparse row. | open, stand-aside | 7a430aa4 |
| E24 | E | 5 | §10 Isolation | §10's "three exceptions" omit (a) std5k.pl from `latencyFixture`, in a `_test.go` file of `psc` that can't be imported, and (b) the RD-graded marker, whose sites are built in `internal/parse/sites.go`. See also A31, C-N24, D-N8. | The isolation rule is quietly false from the first step. | Read `latency_fixture_test.go:4` (`package psc`), `:110`; ran `git grep -l "SubjectCallSite{"`. | open, stand-aside | 7a430aa4 |
| E25 | E | 5 | §8 std5k.pl, set reuse; §9 Speed | `latencyFixture` instantiates `latencyBlock` about 70 times with only `%[1]d` varying, so interning and the transition cache hit almost everywhere after block one; the std5k.pl ratio and set reuse flatter the chart. The text implements R13 faithfully; this is the generator's property. See also B-N15, C-N23, D-N7, F-16, G-4. | The std5k.pl kill leans toward passing; §9's context-leak signal starts inflated. | Read `latency_fixture_test.go:34` `const latencyBlock`; edit5k.pl 70 heredocs, one per block. Effect inferred. | open, stand-aside. Ruling R16 (made after this row was raised) accepts the repetition; not a release. | 7a430aa4 |
| E26 | E | 5 | §7.1 | The marker is "graded … in both directions", yet the same paragraph says "The oracle never sees surplus (`compareMarkers`)": if both directions means missing and surplus, the surplus sentence is false for the sixth marker; if it means both parsers, it says so twice. An extra attached argument is exactly a surplus. See also D-N9, A30. | The marker may be built on `compareMarkers`' missing-only logic, leaving over-attachment ungraded. | Read §7.1 l.270–277; `compare_facts.go:183` `compareMarkers`. | open, stand-aside | 7a430aa4 |
| F-15 | F | 5 | §7.1 (R15); §8 M4–M6 | The gating oracle does not see at least one named M4–M6 case: `sort $coderef @list` and `sort($coderef, @list)` have the same argument structure and differ only in the sort op's stacked flag, and §7.3 no longer gates. Unclear whether `s///e`, `(?{ … })` and unusual delimiters fall under `match`. Fix: each M4–M6 case names its grading marker, or the step adds one; a case with no marker is not checkpoint evidence. See also C-N25, A32. | A step can be "accepted under §7" and "clearly cheaper" while misparsing the very case it was chosen for. | Ran Concise: `sort $c @l` → `sort lKS`, `sort($c, @l)` → `sort lK`, same three kids. Ran `s/a/1+1/e` → `const[IV 2] s/FOLD` vs `s/a/1+1/` → `const[PV "1+1"]`. Whether a current marker grades it is inferred, unchecked. | open, block | 7a430aa4 |
| F-16 | F | 5 | §8 std5k.pl | std5k.pl as `latencyFixture` writes it is one 70-line block repeated 70 times; interning and transition caching may favour the chart's speed and set reuse, and the 200 convergence edits come from about a dozen statement shapes. "Valid only while both parsers accept it" means std5k.pl drops out of the measure, rather than signalling, if the chart fails it. See also B-N15, C-N23, D-N7, E25, G-4. | The speed kill fails to fire because the file is uniform, not because the chart is fast. | Read `latency_fixture_test.go` loop; ran `grep -c '^package Fixture::Block' edit5k.pl` → 70. Effect inferred, unmeasured. | open, stand-aside | 7a430aa4 |
| F-17 | F | 5 | §8 parity list | The parity list names "`print {$fh}` forms" as a Standard-Perl exclusion, but Standard Perl requires them (`print {$fh} $foo; # ok`, `print $fh $foo; # not ok`); std5k.pl's `printf {$out}` is already inside Standard Perl. Text unchanged since e34f8233. See also G-6. | A step planned for what the waypoint covers, while the real exclusion (`print $fh $x`) has no step. | Read `guacamole/lib/standard.pm` "Printing to filehandles with no brace"; subject l.371. | open, stand-aside | 7a430aa4 |
| F-18 | F | 5 | §7.1 (R15) | "Each call" is undefined: `entersub` only, or builtin list and named-unary operators too; the M4 named-unary case and `ref $x, $y` need builtins; method calls need an invocant rule. See also A30, D-N9, G-1. | Different readings grade different populations, and the floor is "at or above RD" on whichever was chosen. | Read §7.1 and §10's marker tests (user subs and `print` only). | open, stand-aside | 7a430aa4 |
| G-1 | G | 5 | §7.1; §10 "Argument marker" | The count is readable from the op tree (entersub `ex-list` children minus `pushmark` and trailing `ex-rv2cv`; filehandle via `print`'s STACKED flag), but scoring is unstated: the oracle compares presence per statement by principle ("never for op counts"), while this marker is a value per call; call identity, what counts as a call, and "in both directions" are undefined. See also A30, D-N9, F-18, E26. | Two implementers build two gates; the M4–M6 parity gate is fixed in code review, not the design. | Ran: `sub foo($); foo $x, $y;` → 1 argument; `sub foo; foo $x, $y` → 2; `foo((1,2),3)` → 2 children; `foo(0 ? $a : $b)` → 1 (FOLD); `foo(())` → 1 `stub`; `&foo;` → 0 with AMPER; `print $fh $x`, `print {$fh} $x`, `print STDOUT $x` → `print vKS`; `print $x`, `print $fh` → `print vK`. Read README, `compare_facts.go` `compareMarkers`/`decideMarker`. | open, block | 7a430aa4 |
| G-2 | G | 5 | §10 "Argument marker"; §8 waypoint preamble | The "`foo $x, $y` under no prototype" case is not a function call in perl: undeclared, it is indirect-object `$x->foo`, and `use feature ':5.36'` doesn't disable indirect syntax (`use v5.36` does). The marker needs a call-kind dimension, or the test is ill-posed. See also A30. | The first argument-marker test can't pass as written, or passes by ignoring call kind. | Ran: `foo $x, $y;` with or without `use strict` and with `use feature q(:5.36)` → `entersub … method_named[PV "foo"]`, `$y` dead; `use v5.36; … foo $x, $y;` → "syntax error … Do you need to predeclare "foo"?". | open, block ("folds into G-1") | 7a430aa4 |
| G-3 | G | 5 | §8 "cost of change", "Checkpoint" | The checkpoint's main number is undefined: (a) pvm's `Implements:` trailers carry issue UUIDs, never feature names, and title greps are noisy; (b) whether grammar lines count in "at most half the lines"; (c) whether the `internal/lexer` refactor is charged to the chart; (d) which baseline RD is charged against when the chart step includes building an exclusion RD already has. See also C-N22, E17. | The M4–M6 verdict could be argued either way after the fact. | Ran `git log origin/pu --format='%(trailers:key=Implements,valueonly)'`: top values all UUIDs (e.g. `01a17d7b-…` ×49), 2150 empty; `git log --grep=heredoc` → 90 commits, mostly fixes. | open, block | 7a430aa4 |
| G-4 | G | 5 | §8 "A realistic input" | std5k.pl can't come "from `latencyFixture`" as described: it takes no parameters, repeats one ~95-line template (70 copies), and lives in a `_test.go` file of `psc` that `internal/chart` can't import; the template uses a heredoc, autoquoted keys and paren-less calls (`lc $word`, `length $word`), all excluded, so a new template is needed. The 70 identical blocks also make interning and the cache hit far more than real code (inferred). See also B-N15, C-N23, D-N7, E25, F-16, A31, C-N24. | The early speed answer is biased toward continuing; set reuse on std5k.pl looks like success whatever happens. | Read `latency_fixture_test.go` (`latencyBlock`, `latencyFixture`); `standard.pm` ("HEREDOCs … not supported", "a string … always quoted", "All subroutines must use parentheses"); conformance plan §M2 "70 `sub scale` bodies". | open, block. R16 accepts std5k.pl's repetition but does not name this row. | 7a430aa4 |
| G-5 | G | 5 | §7.1 "Grouping … not graded"; §3 rule 5 | Rule 5 is not enforced for precedence or associativity: no marker grades it and the span diff isn't gated, though perl's op tree shows grouping (nested `subtract` for `$a - $b - $c` unless folded), so this is a choice, presented as settled. See also A32, D-N10, F-5. | A misassociating grammar can meet parity; RD's grouping isn't graded either, so parity is with an unchecked floor. | Read §7. Op-tree nesting inferred; no grouping probe run. | open, stand-aside | 7a430aa4 |
| G-6 | G | 5 | §8 parity list; §4 | (a) The doc lists "`print {$fh}` forms" as excluded; standard.pm allows `print {$fh} $foo` and excludes `print $fh $foo`. (b) Standard Perl also takes readings perl does not ("Left of arrow is always an invocant, never a function"), built into the grammar and bypassing rule 3; the ratchet lists only exclusions. See also F-17. | (a) a wrong parity step; (b) `Foo->new` with a declared `sub Foo` is a silent non-perl reading carried toward parity. | Read `lib/standard.pm` "Printing to filehandles with no brace", "Left of arrow is always an invocant". | open, stand-aside | 7a430aa4 |
| G-7 | G | 5 | §1; §2.3 | "About 49 open v0.1 issues … (a title-keyword count, approximate)" can't be reproduced, and defect yield's "open issues settled" has no fixed list; fix the list before the first step. | Defect yield can be counted generously afterwards. | Ran `git zhi list --format json`: 140 open v0.1; titles with "perl" 42, "disagree" 2, "parse" 21. | open, stand-aside | 7a430aa4 |
| G-8 | G | 5 | §5 `chart/grammar` | The listed SLIF subset is incomplete: G1 also uses a `+` sequence (`ElemSeq1 ::= Element+`, l.356) and `:default ::= action => …` (l.15). See also E10. | A small surprise on day one. | Ran a grep over `lib/Guacamole.pm`. | open, stand-aside | 7a430aa4 |
| G-9 | G | 5 | §8 waypoint exit | "Every perl-accepted `parses` case accepted under §7" is per case, but §7.1's gate is an aggregate against RD's score; what "accepted under §7" means for one snippet (exact or wider, no WRONG, yield) is unstated. See also C-N25. | The waypoint exit is argued case by case. | Read §7.1, §8. | open, stand-aside | 7a430aa4 |
| G-10 | G | 5 | §7.1 "mainline oracle work" | A sixth kind changes the oracle's frozen baselines: a subject reporting no site of a kind scores `no-answer`, so RD goes no-answer on every file with calls until `internal/parse/sites.go` reports arguments; a sizeable mainline milestone gating M4–M6 with no budget line. See also A30, C-N26. | The checkpoint slips for reasons outside the spike, unseen by the budget rule. | Read README ("A subject that reports no site of a kind anywhere … scores `no-answer`"); `testdata/ratchet/sites.txt` ("A move in either direction fails"). | open, stand-aside | 7a430aa4 |
| G-11 | G | 5 | §9 Speed row; §5 Leo paragraph | Dead weight: the "about 13.6 ms" YAEP-class extrapolation and the "toy YAEP probe" timings have no source or reproduction and feed no gate; the edit5k.pl RD numbers check out. See also D15. | Readers anchor on numbers nobody can re-run. | Read conformance plan BenchmarkLatencyParse 23.89/33.42 ms (matches); no source found for 13.6 ms. | open, stand-aside | 7a430aa4 |

### Round-6 status of rows open at 7a430aa4

Each status is in the raiser's own word, with its reason, as declared at 723440f1.

| id | raiser | round-6 status | reason given |
|---|---|---|---|
| A30 | A | **released** | §7.1: the kind "lands with recursive descent's measured score as its ratchet baseline, and its own budget line", a ratchet not a gate. Ran Concise (5.42.0): `foo((1,2), 3)` keeps an `ex-list` child, giving 2; undeclared `foo $x, $y` is `method_named`; with `sub foo($){}` it has 1 argument. Note: spec 07 §7.5.5 needs a doc-first update in the same mainline PR. (D-N11 and E28 find the count rule wrong under `padrange`.) |
| A31 | A | **released** | std5k.pl's generator is in `internal/chart` with its own template; `psc` unchanged. A's round-5 inference that the heredoc was edit5k.pl's only exclusion was wrong. |
| A32 | A | **released** | Rule 4 now scopes WRONG to the oracle's markers and states that grouping changing only a value is not graded. R14 implemented faithfully. |
| B-N15 | B | **released** (R16) | §8 says std5k.pl's ratio is a lower bound because repetition flatters interning and the cache; the direction is right. B's cache-off remedy was not adopted; set reuse still shows the bias. |
| F2 | C | **held (stand-aside)** | "Aggregate ratio" still undefined; pre-checkpoint snippets "time setup rather than throughput", so the aggregate arm can fire on setup cost. |
| C-N20 | C | **released** via C-N23 | See C-N23. |
| C-N21 | C | **held (stand-aside)** | Budgets unchanged; the parity leg (about 10 features plus require/use loading) has per-step estimates only. |
| C-N23 | C | **released (by ruling R16)** | §8: "its ratio is a lower bound. That is accepted…". C released it because rulings are not C's to reopen, not because the bias is gone. Implemented faithfully for speed; the convergence half is C-N30. Implicit fact: edit5k.pl comes from the same one-template generator, so only perl.git `t/` catches the bias, at the last parity step. |
| C-N24 | C | **released** | Generator in `internal/chart`; §10 lists a fourth exception for the marker's sites in `internal/parse` and `internal/parseoracle`. |
| C-N25 | C | **released** | Rule 4 states R14's consequence, and the waypoint exit spells out its grading. |
| C-N26 | C | **released** | The kind lands with RD's measured score as its ratchet baseline and its own budget line. Parity rule view in C-N29. |
| D-N7 | D | **released** | R16 implemented faithfully: "its ratio is a lower bound", with perl.git `t/` at parity as the backstop. |
| D-N8 | D | **released** | Generator in `internal/chart`; four exceptions, matching `internal/parse/sites.go` and `internal/parse/cmd/subject/main.go`. |
| D-N9 | D | **released, residue in D-N11 and D-N12** | §7.1 defines identity, count, special slot, two-way WRONG and an RD baseline. Ran: `foo((1,2),3)` two direct operands; undeclared `foo $x, $y` → `($x->foo, $y)`. |
| D-N10 | D | **released** | Rule 4 now scopes WRONG to the markers; rule 5 is unqualified but rule 4's clause, one line above, is enough. |
| E11 | E | **released** | Rule 3 now says "(`yyl_leftcurly` for `{`)". |
| E17 | E | **released** | "Go plus grammar, each also reported alone", so "half the lines" is the sum. |
| E24 | E | **released** | Generator in `internal/chart`; four exceptions listed. |
| E25 | E | **released** (R16) | R16 implemented faithfully; "lower bound" is the right direction. |
| E26 | E | **released** | "Never sees surplus" scoped to the five older kinds; the sixth is WRONG "in either direction". |
| F-5 | F | **released** | Rule 4 no longer claims more than is measured. |
| F-15 | F | **released** | §7.1 grades the sort comparator slot, and §10 tests `sort $c @l` vs `sort($c, @l)` (Concise: `sort lKS` vs `sort lK`). Depends on builtins actually being sites (F-18). |
| F-16 | F | **released in part, held in part (stand-aside)** | R16 implemented faithfully. Held: std5k.pl is still "valid only while both parsers accept it", so if the chart fails it the measure silently disappears; make acceptance an exit criterion of the first step. |
| F-17 | F | **released** | Parity list now reads "`print $fh` without braces and bareword filehandles". |
| F-18 | F | **re-raised, stand-aside** | §7.1's site list ("a sub call, `&`-call or method call") excludes `print`, `sort`, `map`, `grep`, so their slots are never graded, against §10's tests. Ran: map's block and expression forms are distinguishable (`null → scope` vs expression under `mapstart`). See also C-N27, D-N12, E27, G-12. |
| G-1 | G | **re-raised, narrower** (see G-12 and G-13) | §7.1 now defines identity, kind, count and slot, which answers round 5; what it newly says is defective. |
| G-2 | G | **released** | §10 lists undeclared `foo $x, $y` as an indirect method call. |
| G-3 | G | **released** | Lines are "Go plus grammar, each also reported alone"; each side is charged its own delta; the chart's includes the `internal/lexer` refactor; RD's PRs are listed by hand and reviewed. |
| G-4 | G | **released** | Generator in `internal/chart` with its own template; R16's lower bound is the right direction, and set interning is named. |
| G-5 | G | **released** | Rule 4 states the limit. G's objection is "a matter of preference, not of honesty". |
| G-6 | G | **released** | "`print $fh` without braces"; §4 and §8 add perl's readings where Standard Perl imposes others. |
| G-7 | G | **released** | §1 drops the count; the defect-yield list is fixed before the first step. |
| G-8 | G | **released** | §5 lists `X ::= Y+` and `:default`. |
| G-9 | G | **released** | The exit is per case plus per corpus. |
| G-10 | G | **released** | RD's measured score as the ratchet baseline, with its own budget line. |
| G-11 | G | **released** | The 13.6 ms and toy-probe figures are removed. |

### Round-6 findings

| id | raiser | round | section of the subject | finding | consequence if it ships | evidence | status | revision |
|---|---|---|---|---|---|---|---|---|
| A33 | A | 6 | §7.1 last sentence (the author call) | A view, not a defect: measuring parity against RD is right for a kind RD doesn't meet, but as a count it hides swaps (the chart WRONG at different sites with the same total). pvm's site ratchet is per site for this reason. The stricter form, the chart's WRONG sites a subset of RD's, costs nothing extra; new chart WRONG becomes a regression and fixed RD WRONG defect yield. See also B-N16, C-N29, G-14. | A chart trading one class of argument misparse for another passes parity unseen. | Read §7.1; `internal/parseoracle/sites_test.go` `TestSiteRatchetDetectsSwapAtEqualTotals`; `siteratchet.go` `CheckSites` ("compared both ways"). | open, stand-aside | 723440f1 |
| B-N16 | B | 6 | §7.1 argument marker (the author call); §4 "Standard Perl … imposes a few readings perl does not take" | B's view: a count rule lets the chart trade errors. Through the M4–M6 steps the chart carries known-wrong sites of its own from Standard Perl's imposed readings (e.g. `foo->bar` with `sub foo` declared, where perl calls `foo()` first), which RD's attachment defects can offset under a count. A set rule keeps the author's reason without the trade. (B's own id: N16.) See also A33, C-N29, G-14. | M4–M6 verdicts can call a step at parity while it introduces argument errors RD doesn't make. | Read §4 para 1, §7.1 last sentence, round-5 response. Trade inferred. Ran Concise: `foo((1,2),3)` 2, `foo((),3)` 2 (`stub`), `foo(1,2,)` 2, `foo(1 ? (2,3) : 4)` 1, `foo(1x2, 3)` 2, each matching if nulled `ex-*` ops count. Ran: `yyl_leftcurly` at toke.c l.6643. | open, stand-aside | 723440f1 |
| C-N27 | C | 6 | §7.1 argument marker | The site definition ("One site per call: a sub call, `&`-call or method call") excludes the builtins whose slots it then records ("a filehandle for `print`/`say`, a comparator for `sort`, a block for `map`/`grep`"), while §10 tests `print $fh $x` and `sort $c @l`. See also D-N12, E27, F-18, G-12. | An implementer following the first sentence doesn't grade the filehandle or comparator slots, the M4–M6 cases the marker exists for. | Read §7.1 para 2; §10; round-5 response R15 text. | open, stand-aside | 723440f1 |
| C-N28 | C | 6 | §7.1 vs parseoracle convention | Grading surplus "in either direction" contradicts the oracle's own reason for never grading surplus ("Perl's optree is post-peephole and only ever loses marker ops"): calls fold away too, and with matching by statement, name and order, one folded or surplus call shifts and misattributes every later site. See also G-13, F-19. | Spurious WRONG for both parsers on every inlined constant sub and dead-code call, cascading. | Ran perl 5.42.0: `use constant PI=>3; my $z = PI * 2;` → 0 `entersub`; `perl -MO=Deparse -e 'use constant PI=>3; my $z = PI * 2; foo() if 0;'` → `my $z = 6; '???';`. Read `internal/parseoracle/README.md` "Five questions, five kinds". | open, stand-aside | 723440f1 |
| C-N29 | C | 6 | §7.1, author call | C's view: accept the relaxation but make it per site; as a count the chart can be WRONG where RD is right, offset where RD is wrong, and still pass, which is a regression for users; defect yield already credits the second set. See also A33, B-N16, G-14. | Parity on argument structure passes with new misparses if the chart fixes as many as it breaks. | Read §7.1 last sentence; §8 defect-yield bullet. | open, stand-aside | 723440f1 |
| C-N30 | C | 6 | §8 std5k.pl paragraph; Checkpoint | R16's lower-bound text covers speed ("its ratio") but not convergence, which is also drawn from std5k.pl, the only multi-statement Standard input before the checkpoint; repetition flatters re-convergence too. See also C-N23, F-16. | The incremental verdict reads as measured when it is optimistic; low cost, since it only gates post-parity work, but the reader isn't told. | Read §8 std5k.pl, convergence bullet, Checkpoint; §2 Output. Flattery of convergence inferred, not measured. | open, stand-aside | 723440f1 |
| C-N31 | C | 6 | §9 Silent-misparse row vs §7.1 | The signal reads "WRONG > 0" unqualified, while §7.1 accepts WRONG > 0 on argument structure up to RD's level. See also E29, F-20, G-14. | Read literally the signal fires as soon as the marker lands. | Read §9 row; §7.1. | open, stand-aside | 723440f1 |
| D-N11 | D | 6 | §7.1 count rule | "The call's list children, less `pushmark` and the callee" is wrong under perl's `padrange` optimisation (any argument list starting with lexicals), which replaces `pushmark` and keeps nulled `padsv`s, so `padrange` is counted as an argument. Also say the invocant counts (`$o->m(1)` → 2). See also E28, A30. | The commonest call shape, `foo($x, $y)` on lexicals, is graded as 3 arguments; both parsers WRONG and RD's baseline polluted. | Ran `perl -MO=Concise -e 'my ($x,$y); sub foo {} foo($x,$y);'` → `padrange[$x; $y] /range=2`, `ex-padsv`, `ex-padsv`, `ex-rv2cv`, no `pushmark`; nested `foo(($x,$y),$z)` → `padrange`, `ex-list`, `ex-padsv`. | open, stand-aside | 723440f1 |
| D-N12 | D | 6 | §7.1 site scope vs special slot | Sites are "a sub call, `&`-call or method call" (`entersub`), yet the same paragraph records builtin special slots, and §10 tests `print $fh $x` and `sort $c @l`; those builtins aren't sites under the stated scope. See also C-N27, E27, F-18, G-12. | An implementer skips the builtin slots (losing two §10 tests) or widens the scope silently. | Read §7.1 para 2; §10. Op-level claim inferred, not re-run. | open, stand-aside | 723440f1 |
| E27 | E | 6 | §7.1 argument marker; §10 | The site set is `entersub` only, yet its fields include the builtin special slots and §10 tests them; under the stated set those builtins have no site. See also C-N27, D-N12, F-18, G-12. | The filehandle and comparator cases R15 was ruled for go ungraded, or the set widens without review. | Read §7.1, §10. Ran `perl -MO=Concise,-exec`: `print $fh $x` is `<@> print`, `sort $c @l` is `<@> sort`, not `entersub`. | open, stand-aside | 723440f1 |
| E28 | E | 6 | §7.1 | The count rule is right for `foo((1,2),3)` only if nulled ops count, and wrong when `padrange` replaces `pushmark`: `foo(($a,$b),$c)` children are `padrange`, `ex-list`, `padsv $c`, `ex-rv2cv`, so the rule gives 3 where the document's reading gives 2. See also D-N11, A30. | Lexical-argument calls, most real calls, are miscounted; both parsers graded against a mis-derived value, or errors cancel and hide defects. | Ran `perl -MO=Concise` (tree): `foo((1,2),3)` → `pushmark`, `ex-list lKPM`, `const 3`, `ex-rv2cv`; `foo(($a,$b),$c)` → `padrange … /range=3`, `ex-list lKPM`, `padsv $c`, `ex-rv2cv`. | open, stand-aside | 723440f1 |
| E29 | E | 6 | §8 waypoint exit vs §7.1 | The waypoint requires every `parses` case "accepted with no WRONG site", while §7.1's author call makes argument parity "the chart's WRONG at or below recursive descent's"; if the kind lands before the waypoint exits, the waypoint demands zero argument WRONG while later parity does not. See also F-20, C-N31. | The waypoint is stricter than parity, or ambiguous on which kinds "no WRONG site" covers. | Read §7.1 last sentence; §8 waypoint Exit. | open, stand-aside | 723440f1 |
| F-19 | F | 6 | §7.1 argument marker | The text doesn't say what happens to a site present on one side and absent on the other; identity by "statement, callee name and order" means one added or dropped call shifts every later site, and perl's `nextstate` anchor carries file and line but no column, so two statements on one line share an anchor. The invocant convention needs stating. See also C-N28, G-13. | An unmatched site may be dropped rather than counted WRONG (restoring surplus blindness), or one extra call cascades into several WRONG sites. | Ran: `my ($x,$y); foo $x, $y;` → one `entersub` with `method_named "foo"`, `$y` outside; `foo((1,2),3)` → outer `ex-list` over inner `ex-list` and `const 3`; a method invocant (`padsv $o`) is a child. | open, stand-aside | 723440f1 |
| F-20 | F | 6 | §7.1; §8 exit; §9 | WRONG now means three things: argument parity "at or below recursive descent's" (§7.1), "no WRONG site" per case (§8 exit), and §9's "WRONG > 0"; with RD's nonzero baseline every step trips §9 while meeting parity, and whether "keep everything already accepted accepted" applies the no-WRONG rule retroactively is unclear. See also E29, C-N31, G-14. | The signal becomes permanent noise, and the gate is ambiguous on the cases that matter. | Read §7.1 final sentence; §8 waypoint Exit; §9 Silent misparse row. | open, stand-aside | 723440f1 |
| G-12 | G | 6 | §7.1 sixth kind | Builtins both have sites and do not: "One site per call: a sub call, `&`-call or method call" excludes `print`, `sort`, `map`, `grep`, yet the paragraph records their special slots and §10 tests them; nothing says how a builtin site is identified (no callee name). Fix: a fourth site kind, identified by the builtin's name. Re-raises G-1. See also C-N27, D-N12, E27, F-18. | The §10 tests can't be built from §7.1 without inventing the builtin site. | Read §7.1, §10. Ran: `sort $c @l` → `sort lKS` with `$c` under a `null`; `sort($c, @l)` → `sort lK`; `map { $_ } @l` keeps a `scope` null under `mapstart`, `map $_, @l` doesn't; `grep { }` the same; `print $fh $x` → `print vKS`. Perl 5.44.0. | open, block | 723440f1 |
| G-13 | G | 6 | §7.1 sixth kind | Perl's folding is graded WRONG: a site is WRONG "if any of those differs, in either direction", but perl removes calls at compile time (constant subs inlined, dead branches dropped), and with identity by "its order among that statement's calls" one call missing on perl's side shifts every later call. Source vs op-tree order is unstated (they differ for nested calls: `foo(bar(1))` runs `bar` first), and "callee's name" is undefined for `$o->$m(1)` and `$c->(1,2)`. The five older kinds forgive folding; this kind reverses that silently. Fix: a site perl has no op for is not graded, and an order rule that survives a missing site (order among calls to the same callee, in source order; dynamic callees by position). Re-raises G-1. See also C-N28, F-19. | Spurious WRONG on both parsers; RD's baseline absorbs them, but every WRONG becomes triage noise, and the author call compares one noise floor with another. | Ran: `sub PI() {3} my $x = PI;` → `const[IV 3] s*/FOLD`, no entersub; `sub foo{} if (0) { foo() } foo(1);` and `my $x = 0 && foo(); foo(1);` leave only `foo(1)`; `$o->$m(1)` → `method K/1` over `padsv[$m]`, no name; `$c->(1,2)` and `&$c` → `ex-rv2cv` over `padsv[$c]`. | open, block | 723440f1 |
| G-14 | G | 6 | §7.1 last sentence; §9 "Silent misparse" | G's view on the author call: relaxing WRONG = 0 to the floor is right, but a count lets the chart be WRONG where RD is right if its total is no higher, the silent misparse rule 5 forbids; the site-level form (the chart's WRONG sites a subset of RD's, per file and site) costs nothing extra. Separately, §9's "WRONG > 0" reads as unconditional, so §7.1 and §9 disagree on the sixth kind. See also A33, B-N16, C-N29, C-N31, F-20. | The sixth-kind parity gate admits new misparses for fixed ones, and §9 contradicts §7.1. | Read §7.1; §9 row; §3 rule 5. Trade-off inferred. | open, stand-aside | 723440f1 |

## Clusters

Counted by the first section each finding names in its "section" column. Many findings name more than one. Round-1 counts use 4ec0955a's section numbering.

### Round 1 (41 findings, at 4ec0955a)

| section | count | findings |
|---|---|---|
| §1 | 1 | A13 |
| §2 | 2 | A10, F13 |
| §4 | 2 | B6, B8 |
| §5 | 12 | A5, A6, A7, B1, B2, B3, B4, B5, B7, B11, B12, B13 |
| §7 | 4 | A1, A2, A3, A9 |
| §8 | 12 | A4, A8, A11, B9, F1, F2, F3, F4, F6, F9, F10, F11 |
| §9 | 3 | B10, F7, F8 |
| §10 | 3 | A12, F5, F12 |
| the session brief | 2 | A14, A15 |

### Round 2 new findings (41, at 47951e0a)

| section | count | findings |
|---|---|---|
| §1 | 1 | D1 |
| §2 | 5 | A16, C-N1, C-N3, D6, D11 |
| §3 | 3 | C-N6, D9, D10 |
| §4 | 2 | B-N2, D8 |
| §5 | 13 | A17, A19, A20, B-N1, B-N3, B-N4, B-N5, B-N6, C-N5, D7, D12, D13, D15 |
| §7 | 2 | D4, D14 |
| §8 | 10 | C-N2, C-N7, C-N8, C-N9, D2, D3, D5, D16, D19, D20 |
| §9 | 4 | A18, B-N7, C-N4, D17 |
| §10 | 1 | D18 |

### Rows open at 47951e0a (52: the 41 round-2 findings plus the 11 re-raised round-1 rows)

B1, B10 and B12 are re-raised as B-N1, B-N7 and B-N2, so they appear in this count twice, once under each id.

| section | count | re-raised round-1 rows |
|---|---|---|
| §1 | 1 | none |
| §2 | 5 | none |
| §3 | 3 | none |
| §4 | 2 | none |
| §5 | 17 | A6, A7, B1, B12 |
| §7 | 3 | A2 |
| §8 | 14 | A4, F2, F4, F9 |
| §9 | 6 | B10, F7 |
| §10 | 1 | none |

### Round 3 new findings (39, at b5b10428)

| section | count | findings |
|---|---|---|
| §2 | 2 | C-N12, D-N3 |
| §3 | 2 | E8, E11 |
| §4 | 1 | E10 |
| §5 | 9 | A21, A23, B-N8, B-N9, B-N10, C-N11, D-N4, E6, E13 |
| §7 | 7 | A24, A25, C-N10, E1, E2, E3, E9 |
| §8 | 12 | A22, B-N11, B-N12, C-N13, C-N14, D-N1, D-N2, E4, E5, E7, E12, E17 |
| §9 | 1 | C-N15 |
| §10 | 3 | C-N16, C-N17, E14 |
| §12 | 1 | E15 |
| Status line | 1 | E16 |

### Rows open at b5b10428 (46: the 39 round-3 findings plus 7 carried rows)

B10 is carried by B-N7, so it appears in this count twice, once under each id.

| section | count | carried rows |
|---|---|---|
| §1 | 1 | D1 |
| §2 | 2 | none |
| §3 | 2 | none |
| §4 | 1 | none |
| §5 | 9 | none |
| §7 | 9 | D4, D14 |
| §8 | 14 | F2, D5 |
| §9 | 3 | B10, B-N7 |
| §10 | 3 | none |
| §12 | 1 | none |
| Status line | 1 | none |

### Round 4 new findings (33, at e34f8233)

| section | count | findings |
|---|---|---|
| §3 | 2 | B-N14, E22 |
| §5 | 2 | E23, F-13 |
| §7 | 8 | A26, E18, E20, F-1, F-2, F-3, F-4, F-5 |
| §8 | 16 | A27, A28, B-N13, C-N18, C-N19, C-N20, C-N21, C-N22, D-N5, E19, F-6, F-7, F-8, F-9, F-10, F-14 |
| §9 | 2 | E21, F-11 |
| §10 | 2 | A29, D-N6 |
| whole document | 1 | F-12 |

### Rows open at e34f8233 (45: the 33 round-4 findings, 10 carried rows and 2 re-raised released rows)

Six rows are re-raised through a newer row and so appear in this count twice, once under each id: B10 and B-N7 (B-N14), B-N11 (B-N13), C-N13 (C-N18), E12 (E19) and A18 (A27).

| section | count | carried or re-raised rows |
|---|---|---|
| §3 | 3 | E8 |
| §5 | 2 | none |
| §7 | 9 | D4 |
| §8 | 23 | F2, D5, B-N11, C-N13, E12, E17, D19 |
| §9 | 5 | B10, B-N7, A18 |
| §10 | 2 | none |
| whole document | 1 | none |

### Round 5 new findings (30, at 7a430aa4)

| section | count | findings |
|---|---|---|
| §1 | 1 | G-7 |
| §3 | 2 | A32, D-N10 |
| §5 | 1 | G-8 |
| §7 | 10 | A30, C-N25, C-N26, D-N9, E26, F-15, F-18, G-1, G-5, G-10 |
| §8 | 11 | A31, B-N15, C-N23, D-N7, E25, F-16, F-17, G-3, G-4, G-6, G-9 |
| §9 | 1 | G-11 |
| §10 | 4 | C-N24, D-N8, E24, G-2 |

### Rows open at 7a430aa4 (36: the 30 round-5 findings, 5 carried rows and 1 re-raised released row)

C-N20 is re-raised through C-N23, so it appears in this count twice, once under each id.

| section | count | carried or re-raised rows |
|---|---|---|
| §1 | 1 | none |
| §3 | 3 | E11 |
| §5 | 1 | none |
| §7 | 11 | F-5 |
| §8 | 15 | F2, C-N20, C-N21, E17 |
| §9 | 1 | none |
| §10 | 4 | none |

### Round 6 new findings (17, at 723440f1)

| section | count | findings |
|---|---|---|
| §7 | 14 | A33, B-N16, C-N27, C-N28, C-N29, D-N11, D-N12, E27, E28, F-19, F-20, G-12, G-13, G-14 |
| §8 | 2 | C-N30, E29 |
| §9 | 1 | C-N31 |

### Rows open at 723440f1 (22: the 17 round-6 findings and 5 carried rows)

G-1 is re-raised through G-12 and G-13, so it appears in this count beside them.

| section | count | carried rows |
|---|---|---|
| §7 | 16 | F-18, G-1 |
| §8 | 5 | F2, C-N21, F-16 |
| §9 | 1 | none |

## Participants

**Round 1.** Dispatched: A (codebase and architecture), B (parsing theory and technical feasibility), C (spec and plan quality, via paad:pushback v2.0.2). All three reported in full, and none appears truncated. C also recorded six candidates it dropped: Leo-by-reference against Chalk's `top_origin`; the "about 49" open-issues count; the semiring-separability assumption; the cold-parse extrapolation; security (no trust boundary); and the §11 SoN oracle.

**Round 2.** Dispatched: A, B and C (resumed) and D (new, reading cold with no prior findings). D states that it did not read round 1's minute or the other participants' scratch files. All four reported in full, and none appears truncated: A ends with "Not re-examined", B and C with "What I did not re-examine", and D with "What I did not check". A, B and C each disclosed a bias towards finding their own round-1 bars met.

Round-2 evidence A re-ran (code unchanged since round 1):
- extraction gave 791/126, identical to round 1;
- 730 cases pass `perl -c` with the preamble, and 68 of those carry a marker site;
- recursive descent accepted 777 `parses` cases and refused 14, with the same 9 perl-valid refusals as round 1;
- round-trip held on all 917 snippets;
- Deparse matches the source on 409 of 730 snippets after rough normalisation.

**Round 3.** Dispatched: A, B, C and D (resumed) and E (new, reading cold). E states that it did not read prior findings, `docs/assessments/` or the design branch's commit messages. All five reported in full, and none appears truncated: A ends with "Not re-examined", B, C and D with "What I did not re-examine", and E with "What I did not check".
- **Scope.** B read sections 4 to 10 in full and skimmed the rest. D only skimmed the ledger at 69582c78, and says a conflict between its releases and other participants' holds is possible.
- **Bias.** A, B, C and D each disclosed a bias. B found its own round-2 B-N7 proposal wrong. C wrote that C-N10 and C-N11 are failures of its own remedies.
- **Perl versions.** E used perl 5.44.0. The other participants cite 5.42.0.
- **Document growth.** The author asked whether the growth from 429 to 469 lines is earned. A and C both answered "mostly", with about 10–15 lines of review residue. A also named budget checking, the build tag and the duplicated override clause as not earned. E16 is E's view of the same residue.
- **E's verified-accurate list.** E recorded these as accurate: every cited symbol (`TokenCache.resyncsAt`, `State.Equal`, `TestLexerRatchet`, `sameState`, `keepsLead`, `compareMarkers`, `FuzzIncremental`, `FuzzIncrementalSequence`); the `expect.go` Word hedge at l.288–309; RD lexing up front (`parse.go:679-681`); 23.89/33.42 ms; 791/126/769/124; spec 03's eleven states; and Chalk's seven ambiguity classes, `LR0DFA.pm` and the Aycock thesis.
- **Counts that differ.** For perly.y, A and D count 14 writes and E counts 18 lines. For edit5k.pl, A releases A18 with heredocs as the only part G1 lacks a rule for, while B, C, D and E hold that its auto-quoting is excluded too. The minute records both positions and does not decide between them.

**Round 4.** Dispatched: A, B, C, D and E (resumed) and F (new, reading cold with no prior findings). F states that it did not read the design branch history, `docs/assessments/` or other participants' files. All six reported in full, and none appears truncated: A ends with "Not re-examined", B, C, D and E with "What I did not re-examine", and F with "What I did not check".
- **Scope.** B read §3 rule 4, §5 chart/grammar and chart/lex, §8, §9 and §10 in full, and skimmed the rest. D read §1–§10 in full, and used only the author's D4 check from the ledger. E read its own rows and R11/R12 in the ledger at 1d929dd4.
- **Bias.** A to E each disclosed a bias. B found its per-statement idea flawed again as written. C wrote that C-N18 is partly its own wording's defect. E retracted its round-3 E15 count.
- **Perl versions.** A and D used perl 5.42.0. E and F used 5.44.0. A notes that its arbiter probes ran on 5.42.0, while the author's ran on 5.44.0.
- **Document growth.** A and C record that the document shrank from 469 to 447 lines. C counts about eight remaining ruling attributions, and F-12 counts seven.
- **C's answers to the coordinator.** C answered questions from the coordinator about the checkpoint thresholds, the denominator, budgets, stale cross-references and growth. C found no stale phase references.
- **F's verified list.** F verified these: the cited symbols (`TokenCache.resyncsAt` at tokencache.go:94, `State.Equal` at checkpoint.go:102, and the others); `perlgitt.ratchet` "619 files, 619 clean"; 347/347; the edit5k timings at conformance-plan:620; the perly.y counts of 14 writes and 4 tests; Guacamole 2838253 at v0.008; and CORE.tpm at `internal/parse/declarations/CORE.tpm`.
- **Counts that differ.** For the attribute errata, F counts 12 invalid `:attr` cases among the 23 under perl 5.44.0, where D counted 10 in round 2. F also counts 708 of 769 unique `parses` cases accepted with stubs absent. The minute records both counts and does not reconcile them.

**Round 5.** Dispatched: A to F (resumed) and G (new, reading cold). G states that it read no prior findings, assessments or branch commit messages. All seven reported in full, and none appears truncated: A ends with "Not checked", and B to G each end with "What I did not check".
- **Scope.** B read §3 rules 3–6, §7, §8 and §9 in full and the rest by grep. D relied on the diff against e34f8233 plus targeted greps, and did not re-read the full text. F read the subject in full and diffed it. E did not re-run perl or Guacamole.
- **The minute.** D did not read the minute's recording of R13–R15. F was told not to read the minute that §2 now points to, and did not.
- **Bias.** Every resumed participant disclosed a bias. C wrote that C-N23 is a defect in a remedy it had itself suggested for F2.
- **Perl versions.** D used perl 5.42.0. G used 5.44.0.
- **Document growth.** The document grew from 447 to 464 lines. C calls the growth earned, and notes a cosmetic run-on at §8 Waypoint l.330.
- **G's verified list.** G verified `TokenCache.resyncsAt`, `State.Equal`, `TestLexerRatchet`, `compareMarkers`, the `expect.go` Word case and `internal/parse/declarations/CORE.tpm`, and that the site ratchet carries no WRONG rows. G ran `git zhi`: 140 open v0.1 issues, and m2-incremental with 8 open.
- **Readings that differ.**
  - On `latencyFixture`'s template, A infers that the heredoc is the only exclusion. C, D and E record a heredoc. G reads autoquoted keys and paren-less calls as well.
  - On `Implements:` trailers, C released C-N22, and G reads pvm's trailers as carrying issue UUIDs rather than feature names (G-3).
  - The minute records each reading and does not decide between them.
- **R16** came in during round 5, after B, C, D and E had raised the repetition finding.

**Round 6.** Dispatched: A to G, all resumed, with no new participant. All seven reported in full, and none appears truncated: each ends with "What I did not check". The round-6 brief is quoted under "Outcome of round 6".
- **Scope.** Most participants read the 7a430aa4..723440f1 diff together with `round5-response.md`. B read the whole diff, and E also read the changed sections in full. D read the diff on top of its earlier full reads. F did not read the minute.
- **Perl versions.** A, C and D used perl 5.42.0. G used 5.44.0.
- **Document growth.** The subject grew from 464 to 495 lines.
- **Bias.** C and D each disclosed a bias. C released C-N23 "by ruling" rather than on its own bar. D checked its own D-N9 harder, which is how D-N11 surfaced.
- **Readings that differ.** On the count rule, A (A30) and B (B-N16) ran Concise on their own cases and found that it matches, provided nulled ops count. D (D-N11) and E (E28) found it wrong under `padrange`. The minute records both and does not reconcile them.

Scratch evidence is under `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/assess-chart/`: `A/`, `A/r2/`, `A/r3/` and `A/r4/`; `B/probe/` and `B/r2/`; `D/`; `E/`; `F/`; and `G/`. D's probe `internal/parse/zz_d_probe_test.go` exists only in D's scratch extraction, not in any repository.

## Not checked

### Round 1

- **A:** the "about 49 open v0.1 issues" count (`git zhi list` shows 173 lines across 5 milestones; A did not reproduce the title sort). Chalk and YAEP internals (Leo `top_origin`, set reuse at about 70%), which are outside A's lens.
- **B:** B did not reread Scott (2008), Leo (1991), Aycock–Horspool (2002), Wiren (1989) or the libmarpa internals, and marked those claims [recalled]. B did not run Chalk or YAEP on Perl, measure an Aycock–Horspool state count, or build any of the chart in Go. B did not check YAEP's error recovery against interning, the incremental phase, rebel's commonplacebook analysis, whether Guacamole's tests pass, or B::SoN. `seedrec.py` is B's own minimal recogniser. B counted 548/88 calls in 38 files against the design's 546/87 and did not examine the difference.
- **C:** whether edit5k.pl falls within Standard Perl (relevant to F2); the open-issue population for the probes.

### Round 2

- **A:** A reasoned the A7 rejoin problem from code and did not build a lattice. A did not re-examine the Chalk/YAEP claims (the 0.007 s / 87.6 s probe, Leo side tables, post-hoc SPPF, seed provenance 28c6029e), the R5 SLIF subset against Guacamole.pm, whether the 9 RD issues were filed, the ledger at 279c1476 (A read only the response's summary of it), the 49-issue count, or the phase-3 probe list against the open chain.
- **B:** B did not re-examine the A- and F-series rows, the Deparse arbiter, the oracle subject, budgets, probe costing or the corpus extraction counts. B did not check whether `TestLexerRatchet` can carry the lexer refactor, the B::SoN material, or real Marpa behaviour (latm, lexeme priorities): B's G1 probe ignores G0 and priorities. B built no lattice recogniser, so B-N1 and B-N4 rest on the heredoc probe and on reading YAEP.
- **C:** C did not re-examine §4 or §5's recognizer and forest internals. C did not verify the YAEP probe numbers, the 23.89/33.42 ms baseline or the 791/126 counts (taken from ledger A4), and did not check whether `perl -MO=Deparse,-p` preserves grouping on the probe constructs (`s///e`, `(?{ })`). C did not re-examine §11, or §12 beyond the Kegel citation.
- **D:** the 49-issue count (`git zhi` is not installed in D's container). Anything in YAEP (no local copy), including the D15 figures. Guacamole's pass rate on its own suite (Marpa::R2 is not installed). Chalk's claims beyond the existence of the files, the commits and the seven ambiguity classes. Whether RD accepts the 23 errata, and how the per-file span diff treats snippets. B::SoN beyond the existence of its design doc.

### Round 3

- **A:** interleaved-mode feasibility, Leo side tables and the YAEP/Chalk claims (B's lens). The R5 SLIF subset, and whether "about 200 per-keyword rules" survive the token cut. Deparse-equality rates on phase-3 corpora: 722/730 is phase 1 only. Whether §7.4's normalisation turns legitimate shape differences into chart "wrong" verdicts now that round 2's "both match → legitimate" clause is gone. A left that to the compare tests and raised no finding on it yet. Whether the round-1 and round-2 RD defects were filed. The ledger at 69582c78 beyond the response's summary.
- **B:** phase-0 tokens against G1 lexemes beyond the `Kind` enum. The `HeredocOpen`/`HeredocBody`, `Prototype` and `FormatBody` tokens, which have no G1 rules, so a Standard-Perl fixture must not contain them. B did not run the production lexer on edit5k.pl. `collapse.py` is B's reading of R10 over 31 inputs, with G0 and priorities otherwise ignored and Marpa not run. B also did not re-examine the arbiter, budgets, phase 2/3 costing, the corpus counts, the A, C and D rows, or whether 10% undecided is a sound threshold.
- **C:** §4, §5's recognizer and forest internals, the token cut and the rejoin rule. C did not verify the 769/124 counts, the 13.6 ms extrapolation or perly.y's 14 writes. C did not run Guacamole's grammar on edit5k.pl: F2 rests on documented exclusions and grep counts. C did not check whether Faithful holds on the Guacamole snippets, which bears on C-N10's phase-1 side.
- **D:** D read the ledger only to its header, and did not read the other participants' rows. D did not re-examine phase-2 cost accounting beyond the D3 note. YAEP (no local copy). Guacamole's own pass rate (no Marpa::R2). Whether pvm canon has a fully parenthesised mode, which D's D4 fix would want. How many `parsent` cases are lexical (D-N4). The 49-issue count (no git-zhi).
- **E:** E did not run Guacamole (no Marpa::R2), so E4 rests on standard.pm's exclusions and grep counts, not a parse. E did not reproduce the YAEP toy probe or the 13.6 ms extrapolation. E did not verify the "about 49" count: E's own keyword grep over 143 open v0.1 lines gives 15, with different keywords. E did not check Chalk's `perl.bnf`@00399607 claims, YAEP at 7aba93e, or the commonplacebook reference. E did not check whether the 68 marker-site figure holds with stubs, or whether Deparse,-p of RD canon equals the source on the Guacamole corpus.

### Round 4

- **A:** B-series internals (Leo side tables, interning across the lattice, error recovery through the `error` terminal). The token-cut rewrites and the `SubName`/`SubNameNonQLike` claim. The forest-cap kill. The ledger at 1d929dd4 beyond the response. Whether the arbiter probes hold on perl 5.44.0.
- **B:** §1, §2, §4, §6 and §7 (the arbiter, the oracle, the 10% threshold). Budgets, cost accounting, the isolation exceptions and `make chart-ratchet`. The A, C, D and E rows. B did not probe whether the production lexer's default for each lexical conflict agrees with G1's single reading of Standard Perl, which B calls "worth one probe at the waypoint". All of B's derivation counts come from its own extractor, not Marpa or pvm's lexer.
- **C:** §4, and §5's token cut, rejoin and recognizer internals. C did not run the literal-renaming Deparse check itself, whether the waypoint corpus needs any token branch, or a re-verification of the edit5k.pl counts beyond round 3.
- **D:** the ledger beyond the D4 note, and other participants' rows. §4 and §11–§12, apart from the perly.y line. The forest cap and the token-cut rewrites, taken on the doc's word. Whether pvm canon can emit paren-minimal or fully parenthesised output. M4–M6 cost fairness beyond D3. The 49-issue count, YAEP, and Guacamole's pass rate.
- **E:** rows A–F and their releases. The `{`/`<<`/`?` deferral: whether XBLOCK, XTERMBLOCK, XATTRBLOCK, XATTRTERM and XBLOCKTERM count as "unreproduced" by construction (see F-9). The error-terminal design beyond its presence. Budgets and §2 Output beyond a read. E23 rests on reading the SLIF: Marpa::R2 is not installed.
- **F:** the "about 49" count (no `git zhi`). The YAEP 0.007 s / 87.6 s probe and the Chalk references (YAEP not on disk). Whether Guacamole's exclusion list matches `standard.pm` and its POD in full. The stub-module approach beyond the 38 cases. F ran no `make test` or Go code. F did not read the design branch history, `docs/assessments/` or other participants' files.

### Round 5

- **A:** B-series internals. The R13 generator's output, since std5k.pl doesn't exist yet, and whether edit5k.pl's 70 blocks contain nothing excluded beyond the heredoc. The reworded expectation-class comparison. The `POSIX::exit(0)` claim. The forest-cap rule. Whether A's round-3 and round-1 RD defects were filed. The ledger beyond the response.
- **B:** whether `latencyFixture`'s block, with exclusions written out, is accepted by G1. Whether perl's op tree exposes argument counts unambiguously, and the marker's cost to the oracle. §7.1's oracle mechanics, cost accounting, budgets and §10's coupling. The A, C, D and E rows. The diff outside §3, §7, §8 and §9.
- **C:** C did not generate std5k.pl, so C-N23's interning claim is inferred. Whether the op tree exposes argument count uniformly (e.g. `(foo $x), $y` under `($)`). §4, §5's token cut, the lattice and the recognizer internals. The ledger at 1d929dd4 and later, beyond the cited IDs.
- **D:** D did not re-read the full text line by line. The minute's recording of R13–R15. Whether the template holds exclusions beyond the heredoc and auto-quoting (e.g. `\*STDOUT`). D-N7's effect, which is inferred. Concise for `print $fh $x` and `($)` calls. YAEP, Guacamole's pass rate, the 49-issue count, and other participants' rows.
- **E:** whether the op tree yields an argument count for every listed call shape. Whether line-keyed sites can tell two calls on one line apart. Whether the template can be written as Standard Perl without changing its token mix. Rows A–F and the F-9/F-10 rewording beyond reading them. E ran no perl or Guacamole this round.
- **F:** whether the existing markers (especially `match`) grade `s///e`, `(?{ … })` or unusual delimiters (F read neither `parse_facts.pl` nor the call_sites protocol). The size of the repetition effect. The minute, which F was told not to read. The parts the diff left unchanged. F ran no Go code and no `make` target.
- **G:** Guacamole's suite under perl. A recording over the 32 files that load without `Test::Fatal` gave 567 `parses` and 76 `parsent`, which fits 791/126 but does not confirm it. No grouping probe (G-5 is inferred). The size of the interning effect (G-4 is inferred). Spec 03's eleven states and the perly.y 14/4 claim. YAEP and Chalk sources. `make test`.

### Round 6

- **A:** B-series internals. The SLIF additions against Guacamole.pm. The "left of arrow is always an invocant" claim and how many cases it affects. The defect-yield list, which doesn't exist yet. The minute change beyond the response. Whether A's earlier RD defects were filed: the 9 refusals, the 6 number-literal mis-lexes such as `0xfp-1`, `$# {…}`, and the `while () {}` canon.
- **B:** method calls under the marker (indirect object, `->$name`, `->SUPER::`) and how `&foo;` is counted. Whether site identity stays stable when the parsers disagree about which calls exist. std5k.pl's new template, which doesn't exist yet. Cost accounting, defect-yield lists, budgets, §10's exceptions, and other participants' rows.
- **C:** the six §10 marker cases against perl's op tree, except `foo $x, $y`, which C ran (it deparses to `$x->foo, $y`). Whether the op tree reports direct-argument counts uniformly for method calls. §4, §5 additions, the lattice and the recognizer. Whether a Standard-Perl std5k.pl that both parsers accept can be built.
- **D:** the text outside the diff hunks. Concise for `print $fh $x`, `sort $c @l` and block `map`/`grep`, so D-N12's op-level claim is inferred. Whether start-line attribution carries call order stably across multi-line calls. The minute's recording of R16 and the cost-accounting changes. YAEP, Guacamole's pass rate, and other participants' rows.
- **E:** whether `method_named` and indirect-method sites keep a stable op-tree order. Whether `map`/`grep` blocks and `sort` comparators are recoverable from nulled ops in every form. Other participants' rows. Sections the diff did not touch. E did not run Guacamole (Marpa::R2 absent).
- **F:** how pvm's oracle identifies call sites (`parse_facts.pl`, the call_sites protocol), so F-19 is inferred from the op tree. Whether a Standard-Perl std5k.pl template is accepted by perl and RD. Unchanged parts of the document. The minute. F ran no Go code.
- **G:** whether `internal/parse` can report call order and kind cheaply. Whether `parse_facts.pl`'s B walk reaches non-exec `null`/`scope` ops for the map-block slot, which G saw only in Concise's tree view. The document outside the diff. `make test`.
