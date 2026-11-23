# Chart parser: precedence without parser state

Status: proposed, for crochet:assess (revision 7). Date: 2026-11-23.
Assessment: `docs/assessments/chart-stateless-precedence.md`. The evidence is
in `docs/assessments/chart-stateless-precedence/`, and its `INDEX.md` maps
every file to where it was produced. Revisions 1 to 3 were
`docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.

The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, branch
`docs/chart-recovery-plan`; perigrin's D2, 2026-11-13).

## 1. Rulings in force

All perigrin's, of 2026-11-23. The minute records every ruling, including
the superseded ones, and refinement records these in the milestone's Records.

- "V1 now, V2 measured after (Recommended)". Ship the exact state-free filter
  first. The recognizer's reduce-side filter (V2) becomes a follow-up
  decision with its own measurement, once its effect on `Expected` and on
  refusal reasons is designed for.
- "Drop the number". There is no numeric target for V1 and V2 together; the
  V2 decision sets its own after this decision's verdict reading. This
  supersedes the 3.92 target.
- "Split it (Recommended)": where the aggregate check runs (§3.4).
- "Rule them in": the three stage keys of §4.
- "Relative and ≤ 5", applied to the goal's `speed-ratio-aggregate-longlived`
  at or below 5. The after-reading's median must be below the
  before-reading's minimum, taken interleaved in one session, and at or
  below 5. The budget is 7 days.
- "We need to eliminate the map".
- "Let's get ourselves to 3.92 first then look". Under "Drop the number",
  this means allocation and the other stages wait until the V2 decision is
  made.

## 2. Evidence

All of it was taken on `pu` 8c3317079, on one machine. Ratios are over
recursive descent's time on the same files. Compare them with each other,
not with `baseline.txt`, whose `speed-ratio-aggregate-longlived` reads 4.79
for today's code.

- **Where the time is.** On the ratchet's aggregate, `Constrain` takes most
  of the pipeline's time: it judges each derivation in bison's LR(0) parser
  state, memoised in hash maps keyed by node and state.

  | Source | `Constrain` | Share of pipeline |
  |---|---|---|
  | `constrain-runs/` | 2.33–2.45 | 43% |
  | `combined-runs/` | 2.18–2.23 | 39% |

- **The parser state is not needed.** A filter that tracks no parser state
  matched `Constrain`'s verdict, kept set and readings count everywhere it
  was run:

  | Input | Size |
  |---|---|
  | The referee's sentences | 1,296 |
  | Toy forests | 25,490 |
  | Toy sentences, checked against bison | 4,395 |
  | Aggregate files | 2,211 |
  | Corpus files | 11,574 |
  | Operator sentences (seed 4; 185 refused by precedence) | 20,083 |
  | Leo-chain and nullable-wrapped sentences (seed 44; 989 refused) | 15,000 |

  Round 6 re-ran both generators and the 11,574 files.
- **V1's speed.** The memo is held in per-node lists and dense tables, with
  no hash map.

  | Reading | Whole pipeline (× today) | Ratio over recursive descent |
  |---|---|---|
  | Combined prototype, three processes | 0.73 (0.72–0.74) | 4.03–4.17 |
  | Round 5, independent | 0.727 | 4.12 |

  - The precedence stage is 0.39–0.42 of today's, per run.
  - No node holds more than 24 memo entries.
  - Building the tables costs about 90 ms, once per process.
- **V2, for its later decision.** In the combined prototype V2 measured
  3.85–4.02 over recursive descent. Round 5's independent run measured 4.10,
  against V1's 4.12 in the same run. Its gain over V1 did not reproduce.

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
   states, `NewPrecedence` returns an error, and `ParseWith` turns that error
   into a refusal of every file. The guard proves the tables are
   well-defined. It does not prove the evaluator supplies the right key: the
   reference of §3.4 covers that. The guard is stricter than today's
   `Constrain`, so it may refuse a grammar `Constrain` handles; perly.y
   passes. A toy grammar that makes generation fail is new work.
3. **No hashing on the evaluation path.** No hash table, built-in or
   hand-written, is read or written per derivation. Per-node lists may grow,
   and a test forces several entries onto one node. The maps `NewPrecedence`
   builds once per grammar stay. Per-call state lives on the call, never on
   the shared `Precedence`.
4. **Identical output, checked against today's judgement.**
   - **The reference.** Today's method stays as
     `(*Forest).ConstrainReference`, documented as used only by tests.
     `(*Forest).Constrain` becomes the state-free judgement, so every
     existing caller switches to it: `ParseWith`, the referee, and the forest
     and chart precedence tests. The reference cannot leave `forest`,
     because it reads `Precedence`'s unexported automaton. A new test holds
     the reference's own agreement with bison.
   - **What is compared.** At the constraint's own boundary, on the same
     built forest (after `libraryOperators`), the test compares the verdict
     and the exact kept set. Equal kept sets give equal readings. Nothing
     before the constraint changes, so recognition, the refusal byte and
     `Expected` cannot differ.
   - **The inputs, and where each runs.** Under "Split it":

     | Input | Runs in | Corpus needed |
     |---|---|---|
     | The referee's sentences, refused ones included | `make test` | none |
     | The toy referees | `make test` | none |
     | `waypoint2Refusals` | `make test` | none |
     | The seeded generators of §2 (committed) | `make test` | none |
     | The superset population's in-tree corpora (T1, Guacamole, std5k) | `make test` | none |
     | The superset population's perl.git files | `make chart-precedence-reference`, behind its own build tag | perl.git at 94e50866 |

     - The seeded generators' test asserts agreement on every sentence, and
       that at least one is refused. It does not assert the 185 and 989 of
       today's grammar, which would trip on a legitimate grammar change. If
       it proves slow, the bulk moves behind the build tag and `make test`
       keeps the refused sentences and a sample.
     - Population is the superset population, which is deterministic, not
       the timing-chosen aggregate.
     - The perl.git target is not run by `make test`. CI's corpus-backed
       Ratchets job runs it with `PERL5_CORPUS` set, and fails unless it
       passes, as that job's other corpus steps do.
     - A developer without perl.git runs `make test`. The perl.git part is
       checked in CI.
   - **What stays unchanged:** both forest digests and every ratchet count
     key.
5. **The referee judges production.** `chart.CountPrecedence` and the toy
   referee tests run the production judgement, so `precedence-disagree` and
   `precedence-covered` measure what ships.

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

That commit is the before-state.

New work they need:

- an exported way to copy a forest, so each timed pass works on a fresh copy
  made outside the span;
- exported access to `libraryOperators`, resolve and the readings count,
  which today are unexported.

How they behave:

- The precedence stage times the forest after `libraryOperators`.
- Like every `speed-` key, they are printed, listed in `baseline.txt`, and
  never gated.
- The commit adding them changes the ratchet's key list, its tests and the
  README key table, and reports the keys' cost to a ratchet run.
- That cost must keep the run inside its CI job's timeout.
- Library operators get no key, so the stage keys do not sum to the whole
  pipeline.

The last issue takes the verdict reading and records it in Records:

- three quiet measurements of the before-state and the after-state,
  interleaved in one session;
- "quiet" means a 1-minute load at or below the core count;
- the commit, load, core count and kernel for each measurement;
- every stage key's median, minimum and maximum, before and after.

## 5. Exit, stops and what the result decides

**Exit:**

- §3 is built;
- the checks of §3.4 pass;
- perigrin's relative exit is met by `speed-ratio-aggregate-longlived`.

**Stops,** reported to perigrin and never tuned toward:

- a §3.4 check differs from the reference;
- table generation finds a state-dependent key that §3.1's context does not
  settle;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- §3 is built and its checks pass, but the exit is otherwise not met;
- a quiet session cannot be had for the verdict reading. Then the reading
  records the load, and perigrin rules on it;
- the stage keys push the ratchet run past its CI job's timeout;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget. At the budget itself, the milestone reports its status
  to perigrin.

**What the result decides.** The before-and-after stage table goes to
perigrin. Under his rulings, the V2 decision comes next and sets its own
target. Allocation and the other stages wait.

## 6. Out of scope

- The recognizer's reduce-side filter (V2), deferred by ruling.
- A value on items.
- Grammar rewriting.
- Changing what precedence decides, including the shortcut where an
  alternative kept in one context is kept in all.
- Speed work on any other stage.

Changes are limited to `internal/chart`, the `Makefile` target and the CI job
step of §3.4, this document, and its assessment. The `Makefile` and CI paths
are named here for perigrin to rule in at refinement, keyed to the issue that
adds the target.
