# Milestone chart-perly-w2 Postmortem

This file has two parts: part 1, the waypoint-2 speed-kill postmortem, kept
as written, and part 2, at the end, on the run resumed after chart-recovery.

**Verdict: killed at waypoint 2 by spec §10's long-lived speed kill.** With the
forest added to the chart side's timed span, using the ratchet's own protocol
at 3487d258:

| Reading | Forest timed: aggregate long-lived | Forest not timed: aggregate long-lived | Forest timed: std5k long-lived |
|---|---|---|---|
| 1 | 3.51 | 2.10 | 2.23 |
| 2 | 3.64 | 2.20 | 2.22 |
| 3 | 3.47 | 2.22 | 2.18 |
| Median | **3.51** | 2.20 | 2.22 |

The limit is 3. All nine measurements were between 3.42 and 3.70. Lex-overhead
stayed within its limit throughout (medians 1.14 to 1.19).

Before the stop, waypoint 2 merged and closed 13 issues. 85 commits went
into the issues' branches, and all of them carry an `Implements:` trailer. 15
issues are pending, 3 are cancelled, and PERLGITPARITY stopped after measuring,
by design.

## What Worked Well

- **The kill was measured early and properly, not projected.** FOREST's gate
  projected 3.9 to 4.2. perigrin ruled "Measure properly first". The answer
  came from three quiet readings under the unchanged protocol, 6 to 12 minutes
  after FOREST merged (01:44 UTC; readings at 01:50, 01:53 and 01:56). It came
  before precedence, resolve or readings were built on a forest that cannot
  fit.
- **Gates caught real defects before merge.** Examples:
  - CTORONCE: a `weak.Make` crash on non-heap tables.
  - FOREST: superlinear forest construction (if/elsif ×2000 took 5.2 s; after
    the fix, 10.6 ms) and exponential enumeration.
  - REFRENDER: two terminals falsely listed as unrenderable, which put 20
    resolutions back under the referee.
  - PERLGITLABELS: two lexer sites that were never listed.

  No issue went past three reopens.
- **Gates re-ran mutations and references.** CACHEGAPS's and CTORONCE's gates
  re-ran each named mutation; XWORD's gate ran two against its known-limit
  test. FOREST's corpus forests were compared node for node against a
  reference built with no pruning, 2,678 files in all.

## What Didn't Work

- **The kill was predictable from waypoint 1.** w1 left lex plus recognize at a
  long-lived aggregate of about 2.2. That left 0.8 for the forest, precedence,
  resolve and readings together. Refinement put the speed checkpoint at the end
  of the chain (01a1c324-2009) and timed nothing per stage. The forest alone
  took 1.3.
- **The ratchet times less than the waypoint builds.** Its speed keys still time
  lex plus recognize. A forest-timed reading needed a scratch harness. The
  issue that would have put the whole pipeline in the timed path
  (01a1c324-1f72) sat behind readings.
- **The w1 close took eight attempts.**
  - Two were killed at the session's 2-hour limit.
  - One was killed by a container restart.
  - One failed on the 60 s suite, which ABSENTFAST then fixed.
  - Three were refused by criterion defects. Each needed a ruling: "Fix the
    check", "Repair the 3 checks", and "Convert 1–2, move 3–4".

  Four merged w2 issues carry the same unrunnable-criterion defect (4202, 1b0f,
  41c7 and LEXLOAD), so a w2 close would have been refused for that too.
- **Criteria that pass without the work.** Milestone criterion 7 is green, but
  `print STDOUT $foo bar();` is still accepted. A FOREST commit silently broke
  w1's repaired `speed.go` check.
- **Shared-machine incidents.**
  - A worker's `pkill -f "make test"` killed another runner's test.
  - A worker used bare `git stash` across shared worktrees (orchestrator
    session, 2026-11-12).
  - The Go build cache filled the disk and stopped a test runner with "no
    space left on device" (run-forest-2, 2026-11-12).

  The worker brief now forbids pattern kills and bare stash.

## What Puzzles Us

- **The cold key changed meaning.** After CTORONCE the aggregate cold ratio fell
  from 4.99 to 2.72, but long-lived barely moved. "Cold" now reuses tables an
  earlier `New` built.
- **Garbage collection costs more than the forest.** In the forest-timed
  profile, forest build took 2.71 s and GC mark workers 4.66 s of 13.32 s of
  samples. The profile cannot split GC by source; by inference a large part of
  it traces the forest's pointer graph.
- **The perl.git corpus is heavy.** It is 38% of the speed aggregate in the one
  non-quiet run that included it. It raised the aggregate before any forest
  existed.

## Where A Human Had To Act

**Collaboration, as designed.** The milestone's Records hold 24 rulings by
perigrin during the waypoint. Among them:

- the library-operator design;
- three w1 criterion repairs;
- the `isa` preamble;
- the toy Leo guard;
- the `pat_psycho.t` exclusion;
- "merge 566 when all checks clear";
- "Measure properly first".

Each of these was his call.

**Friction to engineer away** (from Records and the orchestrator's session):

- Long closes outlived the 2-hour background limit and container restarts. They
  ran detached and were re-run by hand.
- The disk-full stop.
- The environment rewrites the PR body footer on every PR (ORACLEFACTS gate F3,
  PRs #561 to #565).
- The Windows `internal/updater` test timed out on PRs #564 and #565 and was
  rerun by hand.
- Stop-hook prompts to commit `test/e2e/.perl-version` while runs held it
  modified.

## What Will We Change

- **Spend the speed budget at refinement.** Give each stage a share of the
  remaining ratio, and give each stage an issue a criterion that the timed path
  including that stage stays within its share. Here that would have been forest
  ≤ 0.3 against 2.2. This would have stopped the waypoint at FOREST's gate
  rather than after its merge.
- **Put each stage into the timed path the moment it exists.** The ratchet's
  speed keys must time everything built so far, not wait for a late issue.
- **Lint criteria at refinement:**
  - Refuse any criterion with no paren-wrapped command.
  - Refuse any criterion that passes on the unchanged tree.
  - Refuse any criterion pinned to a file a later issue may edit. FOREST broke
    one such pin.
- **Make a GC budget part of a stage's design review.** Before refining another
  attempt, the forest's issue should state its memory shape and be measured
  for collector cost as well as build time; the review found index-based,
  pointer-free nodes (ponytail S1) the only lever on the GC cost here.
- **Check whether a successor can fit at all.** With lex plus recognize at 2.2,
  a successor has to answer whether precedence, resolve and readings can fit in
  0.8, before anyone refines a third attempt.

---
Generated by crochet:postmortem on 2026-11-13.

Data sources:
- `git zhi milestone show chart-perly-w2` (Records)
- crochet:review's three lens reports
- the forest-timed reading report
- `git log 03e4097b7..3487d258`
- gocyclo

Not gathered: `git zhi verify` (pending issues make it fail by construction),
`docs health`, and sanbao.

# Part 2: resumed after chart-recovery, to completion (2026-11-15 to 2026-11-20)

**Verdict: every issue closed, with no speed ceiling.** perigrin ruled on
2026-11-15 that "dropping the speed gate was universal not just for the review
… speed is only one component", so part 1's kill no longer stops the
milestone. The resumed run, 22ccdfcc4..842eb6913, merged 33 PRs (#592 to
#624) and closed 33 issues; all 49 are done (46) or cancelled (3). All 225
commits carry an `Implements:` trailer. Oracle grading at 842eb6913:

| Corpus | Chart exact | RD exact | Chart WRONG | RD WRONG |
|---|---|---|---|---|
| in-tree | 815 | 809 | 0 | 0 |
| T1 | 2153 | 2139 | 0 | 3 |
| perl.git | 5556 | 5397 | 0 | 1 |
| Total | **8524** | 8345 | **0** | 4 |

The whole lead is on rv2hv (+204); on the other four markers the chart is 25
behind. Speed is reported, not gated, and is far worse than at the stop.
Checkpoint 3 (EXIT, #618, two readings at c4eb1e46): aggregate long-lived
5.18 and 5.11, std5k long-lived 8.56 and 8.39, against part 1's 3.51.
lex-overhead 1.213 and 1.236, within 1.3. One ratchet run takes 543 s on pu
against a 300 s budget (input 01a1ec3c).

## What Worked Well

- **Stops went to perigrin, not into tuning.** Eleven: FEATWORDS's recognizer
  bounds, PLUGIN's missing export, WRONGZERO twice, PERLGITPARITY twice,
  LEXMAPBRACE's fidelity row, KEYS's 300 s run, REVIEWFIX3 and REVIEWFIX4
  needing parse files, and HASHFACTS's §10 stop. Each came back as a short
  ruling.
- **Gates probed perl, not the diff.** They found gated-shape
  over-acceptances no corpus holds, so no count could see them: `A *^R;`
  (lgc-1), `use My::Ops (1 plus 2);` (plg-1, against an XS stub) and `require
  v5_1. 1 if 0;` (PERLGITPARITY pass 9). PERLGITPARITY took 11 gate passes,
  LEXARRAYTOP 6.
- **A new reader joined the review in every round from 2 to 7.** fresh5, new
  at round 5, found F5-1, the flaw in the exit claim; round 1's Q3 had seen
  the pooling but not the surplus. fresh7, new at round 7, tested the exit
  claim statement by statement. Six lenses accepted at round 7.
- **The per-file corpus scan caught what aggregate counts hid.** The
  aggregate showed a 193-site lead (8538 against 8345); per file, the chart
  decided rv2hv at 4383 perl.git sites where perl builds 570 (fresh5-r6). The
  scan became a required gate step, and found 0 of 2,231 graded files with a
  surplus at every REVIEWFIX5 gate.

## What Didn't Work

- **The exit was claimed on a grading artefact.** WRONGZERO (#604) and EXIT
  (#618) reported chart exact 8538 against RD's 8345. F5-1 showed the chart's
  oracle facts reported rv2hv at every hash-shaped token, `my %h; $h{a}`
  included, and the grader never penalises a surplus. perigrin ruled "Fix
  facts, re-measure". Graded honestly, HASHFACTS hit a §10 stop (T1 2098
  against 2139, WRONG 2); "Model deref contexts" and "Accept as a stated
  limit" (two op/range.t sites) followed, and it merged at 8524. fresh7 found
  the lead comes from deciding where RD hedges, not from over-reporting.
  perigrin ruled that "if it agrees with Perl and disagrees with RD that
  should always be acceptable".
- **The review took 7 rounds and five fix issues, and the fixes regressed
  what gates then caught.** REVIEWFIX's always-main `qualify` broke an
  always-main glob import (round 3 F1), let the chart accept `our sub ENV`
  for the whole file (R4-1) and refuse a required file's always-main subs
  (N6). REVIEWFIX3's `local` branch refused a glob nested in a `local` value
  (R4-2). REVIEWFIX4's gate found parse's bare-name pass overwriting a real
  import (16 forms). HASHFACTS needed 8 gates; REVIEWFIX5's first cut
  regressed op/ver.t's v-string keys. The cost: six of the run's 33 PRs
  (#619 to #624), the 34 hours from EXIT's merge to REVIEWFIX5's, six rulings
  admitting parse files, and 28 inputs filed after EXIT's merge.
- **Rules that decided by spelling leaked.** HASHFACTS's sort rule was
  narrowed by spelling five times before a token rule replaced it.
  REVIEWFIX5's quote keys needed G7 (a comment after whitespace) and G8 (form
  feed and vertical tab) before taking perl's own isSPACE_A set. The
  orchestrator's own fixes a59a2b915 and f15daffe0 each encoded a wrong rule
  about perl; the next gate caught each.
- **Speed rose with nothing timing it per stage.** VERDICT (#613) put the
  whole pipeline on the timed side: aggregate long-lived 1.69 to 4.95, std5k
  0.96 to 9.18. The worker's probe put about 115 ms of std5k in the
  precedence constraint, against about 25 ms for lex plus recognize.
- **Closes marked done with criteria unverified:** PRECPERLY 1/11, SCOPESUBS
  1/12 and PRINTFH 2/7 (all under load), KEYS 1/12 and REVIEWFIX4 1/5.
- **Criteria and records.**
  - **C8.** LEXGLOBCARET's closed negative criterion says "`* ^R` with a
    space … is not read as a glob name". That is false about perl:
    scan_ident skips spaces in term position, and `-DT` shows the glob. The
    command passes, because the tests follow toke.c. perigrin ruled "Leave
    it, note it".
  - **C9.** Coverage round 1: "Records hold no merge or close entries for
    LEXGLOBCARET, WRONGZERO, KEYS or EXIT. The final figures (WRONG 0, exact
    8538) and the 300 s overrun appear only in commits and the README."
    Round 2 added REVIEWFIX. They are:

    | Issue | PR, merge | Close | Figures from the merge |
    |---|---|---|---|
    | LEXGLOBCARET | #614, 6a2c3eb02 | 3/3 | perl.git no-answer 363 to 16 |
    | KEYS | #617, 4dce33fc8 | 1/12 unverified | 44 keys; runs of 683 s and 664 s |
    | WRONGZERO | #604, 34e063a0a | 7/7 | WRONG 13 to 0; exact 6424 to 8538 |
    | EXIT | #618, 27abc578e | 12/12 | checkpoint 3, above |
    | REVIEWFIX | #619, b36eeab41 | 5/5 | lex-overhead 1.234; std5k 7.72 |

  - **N2.** REVIEWFIX had no lex-overhead criterion. Its run read 1.234, and
    REVIEWFIX2 changed no measured code.
  - **N3.** Checkpoint 3 was timed before REVIEWFIX changed `lex.Map` and
    resolve.
  - **N4.** The amended criterion 12 admits the listed lexer and parse files
    for any w2 issue.
  - **Backticks inside a criterion's command span** made git zhi mark it
    unverifiable; REVIEWFIX2's say criterion had to be amended ("Amend the
    text").

## What Puzzles Us

- **Where the speed goes.** At checkpoint 3 the three stage keys sum to 2.25
  of the 5.18 aggregate; precedence, resolve and readings have no stage key.
- **Which speed is the waypoint's.** REVIEWFIX's non-quiet run read std5k
  7.72 and HASHFACTS's 7.68, below checkpoint 3's quiet 8.39 to 8.56 (N3).
- **Whether the rv2hv lead is merit or RD's caution.** RD hedges 194 perl.git
  rv2hv sites (a package hash read after a `use` or `BEGIN`); the chart
  decides them and agrees with perl. Sites pool to top-level statements, so
  fresh7 called its own evidence weak.

## Where A Human Had To Act

**Collaboration, as designed.** From 2026-11-15 to 2026-11-19 the Records
hold 39 rulings by perigrin. Paths outside internal/chart came in step by
step:

- internal/parse: "File an issue for use builtin" (BUILTINIMPORTS alone);
  "InfixInfo (PLUGIN), interpolatesCode, isReadlineTerm, hidesStructure";
  "Accept" (test-only fallout of a lexer fix); "Allow" (term.go, for $#);
  "Retake, file canon issue" (t1fidelity.ratchet).
- internal/lexer: "Rule lexer issues into w2" (five lexer issues); "Fix both
  in the chart" (the lexical-sub fact); "Rule both into w2" (LEXMAPBRACE,
  LEXGLOBCARET).
- After the fact and in review: "Rule the path in" (this file, for EXIT);
  "Amend, incl. decl/call" (criterion 12); then "Rule parse in", "Rule
  shape.go in", "Rule decl.go in", "REVIEWFIX4, parse in" and "Rule parse.go
  in".

In all, 21 internal/lexer and 18 internal/parse files changed.

**Friction to engineer away** (Records and run reports):

- **Runner defects:** `pkill -f "^make test"` (pex-2); two forbidden
  session-management calls (pgp-9, vrd-1), both refused; an 84-minute stall
  after finishing (leb-1); repeated ratchet runs that clobbered output
  (leb-2); no retry after a load refusal (lgc-5); no criterion-12 step
  (run-hf-2); the wrong T1_CORPUS (run-fwd-1, run-keys-1). Briefs were
  tightened after several; the next defect was a new one.
- **The shell's default T1_CORPUS is wrong.** /opt/corpora/PerlOnJava fails
  the chart's corpus tests; src/test/resources/unit under it passes.
- **Container restarts lost agents repeatedly:** PERLGITPARITY's worker,
  LEXGLOBCARET's and PLUGIN's original workers, and the rf5-2 runner.
- **The Windows updater test** timed out on at least six PRs, each passing on
  a rerun; ratchet guards refused to measure at load (9.60 against 4).

**Inputs.** 37 were filed to chart-perly-w2-inputs during this run, among
them 01a1ec3c (the 543 s ratchet), 01a1ed19 (grading at the innermost
statement), 01a1f60f-4476 (facts.go code health), 01a1f650 (signature hash
parameters after an attribute list), 01a1f6ac (malformed UTF-8 in a `qw`),
01a1f6c0 (a final `$` in an interpolating string) and 01a1f6c5 (README and
comment text).

## What Will We Change

- **Hedge by default.** A fact or token rule decides only where it follows a
  named perl routine (scan_const, isSPACE_A, S_simplify_sort) over its whole
  input class; everything else hedges.
- **Run the per-file corpus scan at every oracle-facts gate,** and have the
  grader report a decided surplus, so no aggregate can hide one.
- **Make a shared rule once.** The review's keying regressions came from
  the chart and parse keying one name differently; export one inMain (input
  01a1f2c6-fca9).
- **Lint criteria at filing:** no backticks in a command span, a claim about
  perl checked against perl, criterion 12 keyed by issue (N4).
- **Retake checkpoint 3 after any change on the timed path (N3),** and add
  stage keys for precedence, resolve and readings before any speed work.
- **Give runners a script, not a brief,** with T1_CORPUS set in it, and block
  pattern kills and session tools at the platform rather than in prose.

---
Generated by crochet:postmortem on 2026-11-20.

Data sources: milestone Records from 2026-11-15; `git zhi issue show` for
this milestone and its inputs; `git log 22ccdfcc4..842eb6913` and the merge
messages of #592 to #624; crochet:review rounds 1 to 7 and their briefs; the
execute-w2c run, gate, merge and close reports.

Not gathered: `git zhi verify` on the milestone, which runs at the close and
is recorded in the milestone's Records; `docs health`; and sanbao.
