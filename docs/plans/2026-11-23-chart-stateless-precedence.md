# Chart parser: precedence without parser state

Status: proposed, for crochet:assess (revision 6). Date: 2026-11-23.
Assessment: `docs/assessments/chart-stateless-precedence.md`.

The evidence is in `docs/assessments/chart-stateless-precedence/`. Its
`INDEX.md` maps every file to where it was produced. The directory holds:

- the prototypes, as gzipped diffs that apply to `pu` 8c3317079;
- the raw timing runs;
- the harnesses;
- every round's reports.

Revisions 1 to 3 were `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.

The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, branch
`docs/chart-recovery-plan`; perigrin's D2, 2026-11-13).

## 1. Rulings in force

All perigrin's, of 2026-11-23. They are recorded at refinement in the
milestone's Records. The minute records every ruling, including the ones
these superseded.

- "V1 now, V2 measured after (Recommended)": ship the exact state-free filter
  first. The recognizer's reduce-side filter (V2) becomes a follow-up
  decision, with its own measurement once its effect on `Expected` and on
  refusal reasons is designed for. This supersedes "Do just B first" and
  "V2: both".
- "Target for V1 then V2 (Recommended)": 3.92 is where this decision and the
  later V2 decision should land together. It is not this decision's exit. V1
  measured 4.12–4.16 on the machine where V2 measured 3.92.
- "Let's get ourselves to 3.92 first then look": allocation and the other
  stages wait until both decisions are in.
- "Rule them in": the three stage keys of §4.
- "Relative and ≤ 5", applied to the goal's `speed-ratio-aggregate-longlived`
  at or below 5. The after-reading's median must be below the
  before-reading's minimum, taken interleaved in one session, and at or
  below 5. Budget: 7 days.
- "We need to eliminate the map".

## 2. Evidence

All of it was taken on `pu` 8c3317079. Ratios are over recursive descent's
time on the same files, on one machine. On that machine today's whole
pipeline read 5.62 in the combined prototype's runs and 5.49 in the earlier
measurement, against the committed baseline's 4.79. So compare these figures
with each other, not with `baseline.txt`.

- **Where the time is.** On the ratchet's aggregate (2,211 files, three quiet
  runs), `forest.Constrain` is 2.2–2.4, about 43% of the pipeline. It judges
  each derivation in bison's LR(0) parser state, memoised in hash maps keyed
  by node and state.
- **The parser state is not needed.** A filter that tracks no parser state
  matched `Constrain`'s verdict, kept set and readings count everywhere it was
  run:
  - the referee's 1,296 sentences;
  - 25,490 toy forests, and 4,395 toy sentences against bison;
  - the 2,211 aggregate files;
  - 11,574 corpus files;
  - 20,083 operator sentences, 185 of them refused by precedence;
  - 15,000 Leo-chain and nullable-wrapped sentences, 989 of them refused by
    precedence.
- **Its speed.** With its memo held in per-node lists and dense tables, and
  no hash map, the whole pipeline runs at:
  - 0.73 of today's time (0.72–0.74 over three processes), recorded in
    `combined-runs/`;
  - 0.727 in round 5's independent reproduction.

  The precedence stage runs at 0.40–0.41 of today's. No node holds more than
  24 memo entries.

## 3. The decision

**The forest's precedence judgement uses no parser state and no hashing.
Recognition and the built forest are unchanged.**

1. **Keys.** A judgement depends only on these:
   - **Reduce side:** the rule and the terminal after the span. For the six
     (rule, terminal) pairs whose decision differs between bison's states
     (`indirob` and `myterm` before `[` and `{`), the parent item is added.
   - **Shift side:** the spine-top item (the item the node's left spine hangs
     from, or the start item at the root), the rule-position item before the
     shift, and the terminal shifted.

   The tables are generated when `forest.NewPrecedence` runs, from the
   grammar's levels, `%prec` and associativity and bison's resolution of each
   (rule, terminal) pair over its automaton. They are stored on the
   `Precedence` value.
2. **What generation guards.** If any key's decision differs between bison's
   states, `NewPrecedence` returns an error. `ParseWith` already turns that
   error into a refusal of every file, so the failure is loud.
   - The guard proves the tables are well-defined. It does not prove the
     evaluator supplies the right key; §3.4's oracle covers that.
   - The guard is stricter than today's `Constrain`, so it may refuse a
     grammar that `Constrain` handles. perly.y passes it.
   - A toy grammar that makes generation fail is new work, added with the
     tests.
3. **No hashing on the evaluation path.** No hash table, built-in or
   hand-written, is read or written per derivation. Per-node lists may grow,
   and a test forces several entries onto one node. The maps `NewPrecedence`
   builds once per grammar stay. Per-call state lives on the call, never on
   the shared `Precedence`.
4. **Identical output, checked against today's `Constrain`.**
   - **The reference.** Today's `Constrain` stays, exported as a reference
     implementation, `forest.ConstrainReference`. It is documented as used
     only by tests, and no production path calls it. It cannot move out of
     `forest`, because it reads `Precedence`'s unexported automaton.
   - **What is compared.** At the constraint's own boundary, on the same
     built forest, it compares the verdict, the exact kept set and the
     readings count. Recognition, the refusal byte and `Expected` cannot
     differ, because nothing before the constraint changes.
   - **The inputs:**
     - the referee's sentences, refused ones included;
     - the toy referees;
     - `waypoint2Refusals`;
     - precedence-refused sentences, from a generator committed with a fixed
       seed, which reproduces the 185 and 989 refused sentences of the
       evidence;
     - the ratchet aggregate.
   - **Where it runs.** The aggregate needs the T1 and perl.git corpora, so
     it runs as `make chart-precedence-reference` behind its own build tag.
     It follows the precedent of `make chart-absence-reference`: `make test`
     runs it after the suite, and CI's corpus-backed Ratchets job runs it.
     When its corpora are missing, the target fails; it never skips. The
     other inputs run in the ordinary suite.
   - **What stays unchanged:** both forest digests and every ratchet count
     key.
5. **The referee judges production.** `chart.CountPrecedence` and the toy
   referee tests run the production judgement, so `precedence-disagree` and
   `precedence-covered` measure what ships. The reference's own agreement with
   bison stays tested.

**What this does not amend.** Recognition stays Boolean and items stay
(state, distance), as v2 §3 rule 12 has it. Precedence is still applied after
the chart, as a constraint on the forest, as v2 §2, §8 and rule 6 have it.
v2 §8 defines correctness by bison's resolution, pinned by generated tests;
this decision meets that standard by test rather than by replaying bison's
states. The live documents change first, in the same pull request: the chart
README's precedence-constraint section, and the ratchet key table.

## 4. Measurement

The three stage keys come first, in a commit of their own on `pu`, before the
precedence change:

- `speed-stage-precedence-aggregate-longlived`;
- `speed-stage-resolve-aggregate-longlived`;
- `speed-stage-readings-aggregate-longlived`.

How they behave:

- Each timed pass works on a fresh copy of the forest made outside the span,
  because the constraint and resolve change their input.
- Like every `speed-` key, they are printed, listed in `baseline.txt` and
  never gated.
- That commit also changes the ratchet's key list, its tests and the README
  key table, and reports the keys' cost to a ratchet run.
- Library operators get no stage key, so the stage keys do not sum to the
  whole pipeline.

The last issue takes the verdict reading and records it in Records with the
commit, load, core count and kernel:

- three quiet measurements of the before-state and the after-state,
  interleaved in one session, where "quiet" means a 1-minute load at or below
  the core count;
- every stage key's median, minimum and maximum, before and after.

## 5. Exit, stops and what the result decides

**Exit:**

- §3 is built;
- the differential checks of §3.4 pass;
- perigrin's relative exit is met by `speed-ratio-aggregate-longlived`.

**Stops,** reported to perigrin and never tuned toward:

- a differential check differs from the reference;
- table generation finds a state-dependent key that §3.1's context does not
  settle;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- §3 is built and its checks pass, but the exit is otherwise not met;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget.

**What the result decides.** The before-and-after stage table goes to
perigrin. Under his ruling, the V2 decision comes next, toward 3.92 for V1 and
V2 together. Allocation and the other stages wait.

## 6. Out of scope

- The recognizer's reduce-side filter (V2), deferred by ruling.
- A value on items.
- Grammar rewriting.
- Changing what precedence decides, including the shortcut where an
  alternative kept in one context is kept in all.
- Speed work on any other stage.

Changes are limited to `internal/chart`, the `Makefile` target and the CI job
step of §3.4, this document, and its assessment. The `Makefile` and CI paths
are named here for perigrin to rule in at refinement.
