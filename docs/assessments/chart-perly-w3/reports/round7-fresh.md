# Round 7, fresh participant

Subject: docs/plans/2026-11-23-chart-stateless-precedence.md, blob ce616721c9ed717dc119cf4c5f3329daf802f4cd. The blob is the same at 9afff6c63 and at the worktree HEAD, 1ff6e3f6a. I did not read the minute or reports/ before writing this.

## Findings

**L1 (§4 and §5: the reading does not record the key the exit is judged on, and "measurement" has two meanings).** §5's exit is judged on `speed-ratio-aggregate-longlived`. §4's verdict reading records "every stage key's median, minimum and maximum", and the exit key is not a stage key. "Three quiet measurements" also clashes with the ratchet's own term: `Protocol.Ratio` (speed.go:68) already returns the median of 3 measurements, each of 2+9 pairs, which is also v2 §9.3's sense. "Median" and "minimum" can therefore be read three ways: over run values, over measurement medians, or over round minima. In combined-runs, today's run medians are 5.59–5.68 but its round minimum is 4.99. *If it ships:* the exit and the precedence-stage stop each have two defensible bars, and the verdict can be disputed after the fact. *Fix:* name the exit key in the reading, and define one measurement as one ratchet run's printed value.

**L2 (§4: "fresh copy made outside the span" needs a protocol change that is not listed).** `Protocol.pass` (speed.go:54–65) starts the clock and then calls `side()`. No per-pass setup runs outside the span. Precedence and resolve both mutate the forest, so a fresh copy outside the span means one of two things. Either a new hook in `Protocol`, which §4's "New work" list leaves out. Or pre-built copies for 3×11 passes per mutating key, which inflates the heap, and constrain-measurement.md blames a larger heap for its 5.49 against 4.79. *If it ships:* the before-state commit either times the copying or changes the protocol mid-issue, which is a §5 stop. *Fix:* add the hook to New work and state how many copies are live at once.

**L3 (§1 against §6: one ruling is both in force and deferred).** §1 lists "Split it" as in force, and it is what gives the perl.git part its make target and CI step. §6 says those paths are "named here for perigrin to rule in at refinement". Under crochet, refinement records an outcome and does not decide one. v2 §11 says "Only internal/chart changes", and the "does not amend" paragraph does not mention §11. *If it ships:* if the ruling at refinement is no, §3.4's sixth row has nowhere to run. *Fix:* either cite the earlier ruling that let chart-absence-reference into the Makefile and CI and drop the deferral, or list §11 as amended.

**L4 (§2: the table-build cost comes from a different prototype).** "About 90 ms" comes from stateless.diff (measurement/result-run1..3: 105, 122 and 85 ms; compare 86). That prototype keeps its tables in maps (`byParent`, `shiftForbid`), plus a global `statelessFor` map. V1's dense tables were never timed in the evidence. I applied combined.diff to 8c3317079 and timed `NewStateless` five times at load 0.16: 105–137 ms. `NewPrecedence` itself took a further 297–343 ms. The consequence is small because the cost is paid once per process, but the figure should come from V1.

**L5 (§3.4: "after libraryOperators" is not true of every input).** The referee judges `builtForest` (Build only), narrowed to the sentence's terminals, with no libraryOperators (chart/precedence.go `judge`). The stated boundary holds only for the population and generator rows. *Risk:* an implementer adds libraryOperators to the referee comparison and changes what it compares. *Fix:* say "the forest each caller passes to Constrain today".

**L6 (§3.4 against §4: an unstated dependency, and an incomplete list).** The superset population is computed in cmd/ratchet, which is package main. The in-tree population differential therefore has to be a cmd/ratchet test, and it needs §4's forest copy and its exported libraryOperators. §4 justifies those exports by the stage keys alone. Separately, the in-tree population includes edit5k.pl (lex.go `corpus()`), which the parenthetical "(T1, Guacamole, std5k)" leaves out.

**L7 (§3.4: new work is not marked as new).** "The seeded generators of §2 (committed)" exist only as evidence: harness/fresh4/*.go.txt, which call the prototype's `fresh4Run` and `CStats`. Those will not exist once the prototype is gone. Turning the generators into tests is new work, and so are `make chart-precedence-reference` and its build tag. §3.2 and §4 mark new work explicitly, and this item should be marked the same way.

**L8 (§3.4: the standing checks drop the corpus refusals that §2 covered).** The superset population excludes every file perl refuses, which is where real-code precedence refusals live. §2's 11,574-file run included such files. Refusals stay covered by the referee, `waypoint2Refusals` and the generators, but not by real corpus files. *Suggest:* compare on every recognized corpus file. That set is just as deterministic as the population.

**L9 (§2 and INDEX: two claims the committed evidence does not back).**
- "Round 6 re-ran both generators and the 11,574 files": no output from that re-run is committed under measurement/ or harness/. INDEX only says round 6 confirmed the list mapping.
- INDEX.md says "the decision quotes 5.62", but revision 7 no longer quotes it, so that line is stale.

## Checked and consistent

- `NewPrecedence` already returns an error, and `parseStream` already turns it into a refusal. Strictly, that covers only files that reach the constraint, and the ratchet's `pipelineReason` makes any other refusal reason a loud error.
- `Precedence.states` is unexported, which matches §3.4's claim that the reference reads its unexported automaton.
- The callers of `Constrain` are exactly the four §3.4 lists.
- These exist as named: `CountPrecedence`, `waypoint2Refusals`, `precedence-disagree`, `precedence-covered`, and the two digest goldens (forestdigest_test.go).
- The CI Ratchets job sets PERL5_CORPUS, and the perl5 pin is 94e5086. The job has no timeout-minutes, so "its CI job's timeout" can only be chart-ratchet.yml's advisory 20 minutes.
- T1, the perl labels and T1's RD loader are all in-tree, so `make test` needs no corpus.
- The six mixed pairs, confirmed by a run: three `myterm` rules and `indirob: scalar`, each before `[` and `{`.
- Every §2 figure matches the evidence, apart from L4:
  - 2.33–2.45 and 43%; 2.18–2.23 and 39%;
  - 0.73 (0.72–0.74) and 4.03–4.17; 0.727 and 4.12;
  - 0.39–0.42; max 24; V2 3.85–4.02 and 4.10;
  - 1,296; 25,490; 4,395; 2,211.

## Not examined

- The minute, reports/, the recovery plan and c1.
- The 11,574, 20,083 and 15,000 counts, which appear only in reports.
- Whether chart-ratchet.yml's 20 minutes leaves room for the new keys.
- What the make-test additions cost on the macOS and Windows runners.
- The correctness argument for V1 beyond the evidence counts.

## Verdict

Modify. The direction is sound and the figures hold. L1–L3 are internal contradictions or under-specifications that would surface mid-milestone as stops. L2 is the one that would.
