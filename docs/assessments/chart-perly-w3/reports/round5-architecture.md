# crochet:assess round 5: architecture and direction lens

**Subject:** revision 5 of `docs/plans/2026-11-23-chart-stateless-precedence.md`, at 2782fd181.
**Blob read:** d97415a147b923df589b27bed3cad0317f6bbfe2. I checked it with `git hash-object`; it matches the
expected `d97415a1`.
**Minute:** at 157b0f820, which is HEAD of the worktree.
**Recommendation: accept, with three small edits that do not change the decision.** A33 needs perigrin's
reading of one ruling. A34 and A35 are each a sentence. None of them blocks.

**Bias.** I raised A26–A31, and I am the only one who may release them, so I lean toward finding my own bar met.
Each release below quotes the text it rests on.

## What I read and ran

**Read, in this order:**

1. The subject, in full.
2. The evidence directory, which I listed.
3. `git diff 6dc1d9558 2782fd181 --stat`, last.

**Checked by running:**

- **Reports.** I compared all 15 files under `reports/` with `cmp` against the session's scratchpad originals.
  All 15 are identical.
- **Prototype diffs.**
  - All three gzipped diffs pass `gzip -t`.
  - All three apply with `git apply --check` to a `git archive 8c3317079 internal/chart` tree.
  - `no-map.diff.gz` is byte-identical to the scratchpad's `fresh3/prototype.diff`.
- **Measurements.** I read the first lines of `measurement/combined-compare.txt` and
  `measurement/stateless-compare.txt`.
- **Figures in §2.** I grepped the evidence directory for them.
  - "11,574", "20,083" and "15,000" appear only in `reports/round4-fresh.md`.
- **The minute.** I read its header (lines 1–63) and grepped it for scratchpad paths and the old file name.

## Not examined

- I did not re-run any prototype or harness, and I did not re-derive any figure in §2.
- I did not check the claims about bison's states: "six (rule, terminal) pairs" and "46 relevant spine tops".
  Those belong to the codebase lens.
- I did not read the minute's round-4 section beyond its header.
- I did not check:
  - whether the pre-commit hooks pass on the evidence files. The commit exists, and the hooks ran on it;
  - the size limit for `stateless-compare.txt` (315 KB);
  - whether the differential test over the aggregate fits the chart suite's time budget (A35 raises the question).

## A26–A31

| id | disposition | ground |
|---|---|---|
| A26 | **moot: V2 is deferred** | §3: "Recognition and the built forest are unchanged." §6: "The recognizer's reduce-side filter (V2), deferred by ruling." With recognition precedence-free, v2 §8's 2026-10-30 ruling, v2 §2, rule 6 and rule 12 are all untouched. The amendment carries to the V2 decision whenever it comes, and I agree it should. |
| A27 | **moot here; carries to the V2 decision** | Its subject, recognition's contract, is gone from this decision. Round 4's fresh participant confirmed the concern: "`print $fh [ ) ;` … drops `]` from `Expected`" (§2). §1's ruling makes it a design input: "once its effect on `Expected` and on refusal reasons is designed for." I agree it carries. |
| A28 | **released** | §5's fourth stop now reads: "that median is below the before-reading's minimum, but `speed-ratio-aggregate-longlived` stays above 5". This defines "faster" as stop 3's complement. §6 no longer says the forest build applies the shift side. |
| A29 | **released** | The minute's ABOUTME reads "now precedence without parser state … renamed at 4". The title is "Chart parser: precedence without parser state". The Subject is the new path and names the old one. A revision table gives commit, blob, path and title for each revision. The remaining scratchpad pointers are A34. |
| A30 | **released**, with a residue raised as A34 | The plan's header names the evidence directory. I verified the contents: the prototypes apply to 8c3317079, and the reports are verbatim. Gzip is a sound way past the whitespace hooks for patches whose whitespace is significant. |
| A31 | **released** | §3.5: "`chart.CountPrecedence` and the toy referee tests run the production judgement, so `precedence-disagree` and `precedence-covered` measure what ships. The oracle's own agreement with bison stays tested." |

## Assessment of revision 5

### "What this does not amend"

The paragraph is correct.

- Recognition and `Build` are unchanged, so these all hold as written:
  - rule 12's "Boolean … (state, distance)";
  - v2 §2's "resolved in the forest";
  - v2 §8's "precedence leaves recognition";
  - rule 6's "after the chart".
- Its standard sentence is accurate: "v2 §8 defines correctness by bison's resolution, pinned by generated
  tests; this decision meets that standard by test rather than by replaying bison's states". v2 §8's own words
  are "the definition is bison's resolution of each conflict, and the generated tests below are what pin it".
  v2 §8 never requires parser states; only the README's mechanism uses them.
- The paragraph names the README's precedence-constraint section and the key table as the live documents that
  change first.

§3.2 is now also honest about what the generation guard proves: "proves the tables are well-defined. It does not
prove the evaluator supplies the right key; the oracle of §3.4 covers that." That is the split I argued in round
4, stated in the document.

### §4–§5

The sections are consistent.

- **Stage keys first.** They land in their own commit, with a fresh forest copy per pass, and their cost is
  reported.
- **Verdict reading.** It follows chart-recovery's practice: one session, interleaved, quiet, with commit, load,
  cores and kernel recorded.
- **The two speed stops.** They partition the outcomes: the precedence stage not faster, or faster while the
  ratio stays above 5.
- **The exit.** It is perigrin's relative ≤ 5.
- **What the result decides.** "the recognizer's reduce-side filter (V2), allocation, or another stage" is the
  right set.

The one inconsistency is the "3.92" ruling (A33).

### The evidence directory as a record

It is adequate. I checked it rather than took it on trust: the reports are verbatim, and the diffs are intact and
apply to the named base. The gaps are small (A34).

### The minute across the rename

Its header now reads cleanly. The residue is that some pointers still go to the scratchpad (A34).

## New findings

### A33. §1 and §5: "3.92 first" is V2's figure, and this decision builds V1 (for perigrin; the text should reconcile it)

- **What the record shows.**
  - §1 records "Let's get ourselves to 3.92 first then look", listed after "V1 now, V2 measured after".
  - In the minute's combined-prototype table, 3.92 is V2's whole-pipeline ratio.
  - V1's ratio is 4.16, both on the machine where V0 read 5.62.
  - §2 itself gives V1 as 0.73 of today's time. 5.62 × 0.73 is about 4.1, not 3.92.
- **Why it matters.** Read literally, the ruling's target is reachable only with V2, which §6 defers. The
  decision's exit is perigrin's relative ≤ 5, which V1 is projected to meet.
- **Consequence if it ships.** Two readings of "done" coexist.
  - The exit is met at about 4.1, yet "3.92 first" is unmet.
  - So either V2 follows at once, ahead of allocation, or the ruling's figure is read loosely.
  - The next decision starts from a disagreement nobody wrote down.
- **Modify.** Add one sentence under the ruling: "3.92 is V2's figure on the prototype machine; this decision's V1
  projects about 4.1 there, and V2's decision is where 3.92 is reached". Or ask perigrin which reading he meant.

### A34. The minute still points at the scratchpad, and a few cited files are not in the directory (should fix)

- **Pointers.** The minute's header still says "The full reports are at the paths below, in the session
  scratchpad (`/tmp/claude-0/…`)" (line 45) and lists `assess-w3/…` paths. Line 63 and the "Experiments after
  round 3" paragraph also place the prototypes at `stateless/stateless.diff` and `combined/combined.diff` "in the
  same scratchpad". They should point to `chart-stateless-precedence/reports/` and `prototypes/`.
- **Renames with no recorded mapping.** `no-map.diff.gz` was `fresh3/prototype.diff`. I verified that by
  `cmp`, but nothing records it.
- **Files cited but not copied:**
  - fresh3's harness `fresh3/exp_test.go.txt`;
  - round 4's codebase probes under `codebase/ov4/`;
  - the C1-spike report, `chart-recovery/c1-report.md`, which the minute's author's statement cites.
- **Figures whose generator is not preserved.** §2's "11,574 corpus files", "20,083 operator sentences" and
  "15,000 Leo-chain … sentences" are recorded only in `reports/round4-fresh.md`. That report used
  `combined.diff`'s harness plus its own additions, which are not in the directory.
- **Consequence.** Once the scratchpad is cleared, the minute's links die, and those figures can be read but not
  reproduced.
- **Modify:**
  - repoint the minute;
  - add a short index (a README in the directory, or a table in the minute) giving each file's original path;
  - copy the missing files, or record that they are not kept and why.

### A35. §3.4: where the differential test runs is unstated (low)

- **The issue.** The differential runs today's `Constrain` against the new judgement over the 2,211-file
  aggregate. That aggregate needs `T1_CORPUS` and perl.git at `/opt/corpora/perl5`, and the prototype's
  preparation alone took 17 s.
- **Two precedents:**
  - `make chart-absence-reference`: a build tag, run by `make test` after the suite and in CI's corpus-backed
    job;
  - the ratchet, which is not in `make test`.
- **Consequence.** If the differential lands in the default suite, it may push the chart suite past its 50 s
  budget. If it lands nowhere that runs regularly, the oracle stops guarding anything after merge.
- **Modify.** One clause saying where the aggregate differential runs. The referee and toy parts can stay in the
  default suite.

## Summary

| id | status |
|---|---|
| A26 | moot (V2 deferred); carries to V2's decision |
| A27 | moot here; carries to V2's decision |
| A28 | released |
| A29 | released |
| A30 | released (residue A34) |
| A31 | released |
| A32 | not used |
| A33 | new: "3.92 first" is V2's figure; reconcile it, or perigrin reads the ruling |
| A34 | new, should fix: minute pointers, file index, missing files and generators |
| A35 | new, low: say where the aggregate differential runs |

**Recommendation: accept, with edits.**

- The decision is now narrow, exact by test, and correct about what it does not amend.
- Its measurement plan and stops are consistent with the README and with chart-recovery.
- Its evidence is durable and checked.
- A33 needs one sentence, or perigrin's reading. A34 is housekeeping in the minute and the evidence directory.
  A35 is one clause. None of them changes the decision.
