# Round 3, fresh participant

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` at 3e4d0c43b, blob a8960d05a84284fb0f9fcec803d101949fd9c405 (the live file matches the committed blob). Of the assessment minute I read only "Measurement after round 2".

Code read: `internal/chart/forest/precedence.go` (all of it), `forest.go` (parseStream), `forestdigest_test.go` (what the goldens cover), `cmd/ratchet/speed.go` (measureSpeed, measureStages), `testdata/ratchet/baseline.txt`, and the v2 design's §8 and speed lines.

## Experiment (scratch worktree at origin/pu 8c3317079, now removed)

I ran three variants of Constrain, interleaved within each round, against an unmodified control. 13 rounds, the first 2 discarded. Each run restored the forest from a snapshot before the call. Population: 859 accepted forests from `internal/chart/testdata/guacamole` plus std5k.pl, 165,699 nodes. A second run used the 40 largest of those files. Load was 0.4–0.9 on 4 cores. Every variant's output was compared with the control's: every `Alts[].Next` and every `Nodes[].Alt` head was identical. Under the no-map variant the forest package's tests pass, including both digest goldens (TestResolveKeepsCorpusAccepted fails on both variants alike, because /opt/corpora/PerlOnJava is absent here).

| Variant | Median Constrain time (three runs) |
|---|---|
| as on pu | 111 / 97 / 107 ms |
| maps given a `len(f.Nodes)` size hint | 111 / 91 / 104 ms (no change beyond noise) |
| dense `[state*nsym]int32` goto table in `walk` | 110 / 91 / 106 ms (about 0–7%) |
| no map: per-node `[]{s,t,out}` for eval and per-node `[][3]int32` for mark, by linear scan, plus the dense goto | **40 / 36 / 35 ms (2.6–3.0× faster)** |

The deepest eval recursion was 810 nested calls. Go 1.24's CPU profile keeps at most 64 frames (`maxCPUProfStack = 64` in runtime/cpuprof.go). The prototype diff and harness are kept beside this file: `fresh3/prototype.diff` and `fresh3/exp_test.go.txt`.

## Findings

**G1. §1 and §4: the exit's absolute half is already met, or depends on which host takes the reading.**
- Consequence if it ships: the exit cannot tell perigrin whether the work paid for itself.
- Evidence: the committed baseline (`baseline.txt`, retaken at 83fdd50f4) reads `speed-ratio-aggregate-longlived 4.7915`. ParseWith already called Constrain at that commit, so the ratio is at or below 5 before any change. The measurement reads 5.49, but its own minute says it is "not comparable with the ratchet's keys" (a different host, and every forest held in memory). README records the same key moving across a host change.
- Modify: state that "≤ 5" is read on the session's own interleaved before/after pair. Report the before-reading next to the committed 4.79, so a before that already meets the exit is visible.

**G2. §4: no stop for the middle outcome.**
- Consequence if it ships: if §3 is built, the output is identical and the stage is faster, but the ratio stays above 5, nothing stops the work. It meets neither the exit nor any stop, so it runs until twice the budget.
- Modify: add a stop for "built and identical, stage key improved, exit not met". It reports to perigrin with the stage table.

**G3. §2: the profile's breakdown is unreliable. Its conclusion holds.**
- Consequence if it ships: a reader of §2 expects a size hint to recover much of the cost ("growing maps made with no size hint", 7.4 s of the 12.4 s). In my experiment the hint does nothing measurable.
- The recursion runs 810 calls deep, so roughly 1,600 or more frames, against a 64-frame profile limit. Samples taken deep in the recursion lose their root frames, so "57% of Constrain's own CPU" is approximate in a direction nobody knows.
- The wall-clock prototype supports the decision more strongly than the profile: removing the eval and mark maps cuts about 64% of Constrain's time. The goto map matters least.
- Modify: cite wall-clock evidence for the decision, not profile shares.

**G4. §3.1: nothing checks "for every file" across the aggregate.**
- Consequence if it ships: a change in which alternatives survive on an aggregate file outside the sample passes every named instrument.
- Evidence: `forest-digest.golden` and `resolved-forest-digest.golden` cover the corpus sample (about 440 files). The ratchet count keys and the referee's two keys count verdicts and disagreements; they do not show which alternatives each file keeps.
- Belongs in an issue: a differential check that runs the current Constrain, kept as a test-only oracle, and the replacement over the 2,211-file aggregate. It compares node heads and `Alt.Next`, as my harness did.

**G5. §3.2: it prescribes a mechanism, not a property.**
- Consequence if it ships: "allocated once per call, not grown entry by entry" would forbid the design that measured fastest. That design grows small per-node slices by append for the 14% of nodes with more than one context.
- eval's results are per-entry allocations however they are stored. Read literally, the clause pushes toward one flat arena. In an arena, an `append` to a returned sub-slice overwrites its neighbour: eval returns memoized slices directly, and seq returns `right` unchanged.
- Modify: "no hash map; no table indexed by node, context or state grows during the call."

**G6. §3: the memo semantics that must survive are left unstated.**
- Consequence if it ships: the issue author has to rediscover them. "Computes exactly what it computes today" is the right contract, so this is not a defect in the decision.
- What a replacement must keep:
  - mark records the visit before it recurses (deduplication).
  - mark's key includes `first`, -1 meaning any. That is where the 32-context maximum comes from.
  - eval memoizes only after computing, with no in-progress sentinel. That is safe only because the forest is acyclic.
  - Memoized slices are shared and must never be mutated.
  - The kept-in-one-context shortcut is `kept []bool` by alternative. It is not a map and does not change.
- Belongs in an issue: list these in its description. My prototype kept all of them and its output is identical.

**G7. §3.4: the issues have to land in a fixed order.**
- `speed-stage-precedence/resolve/readings-aggregate-longlived` do not exist on pu. `measureStages` has only lex, recognize and forest, and the forest stage is `forest.Build` alone.
- The keys need a commit of their own, before any change to Constrain. The interleaved "before" must be taken on that commit, not on 8c3317079.
- Adding the keys also touches `main.go`'s key list, `main_test.go`, the README key table and the baseline.
- Belongs in an issue.

**G8. Direction: the exit's thresholds will not separate outcomes.**
- If the guacamole speedup holds on the aggregate, Constrain falls from 2.38 to about 0.8–0.9 in the harness's units, and the pipeline from about 5.5 to about 4.0. Both the exit and the stage stop pass by a wide margin.
- What perigrin needs after this is where the time goes next: recognize 0.94, build 0.73, lex 0.66.
- Modify: the exit report gives the full seven-stage table, before and after, in one session, not only the ratio.

## Not examined

- The aggregate itself. T1 under /opt/corpora is absent here and I did not assemble perlgit. My population is 5% of the aggregate's nodes, mostly small files.
- The ratchet run, the referee's `precedence-disagree` and `precedence-covered`, the GC share, and recursion depth on the aggregate.
- Rounds 1 and 2, the recovery plan beyond a grep, and the README beyond its key table and §4 notes.
- Whether deep recursion (stack growth) costs anything in its own right.

## Verdict

**Modify.** The direction is right, and a wall-clock prototype with identical output supports it more strongly than the profile does. Fix the exit (G1, G2, G8), restate §2's evidence (G3) and §3.2 as a property (G5). G4, G6 and G7 go into the issues.
