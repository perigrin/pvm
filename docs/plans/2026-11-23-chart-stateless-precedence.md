# Chart parser: precedence without parser state

Status: proposed, for crochet:assess (revision 5). Date: 2026-11-23.
Assessment: `docs/assessments/chart-stateless-precedence.md`. Its evidence
directory, `docs/assessments/chart-stateless-precedence/`, holds:

- the prototypes, as gzipped diffs that apply to `pu` 8c3317079;
- every round's reports;
- the measurements.

Revisions 1 to 3 were `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.

The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, branch
`docs/chart-recovery-plan`; perigrin's D2, 2026-11-13).

## 1. Rulings

All perigrin's, of 2026-11-23. They are recorded here, and at refinement in
the milestone's Records.

- "Do just B first and let's re-measure".
- "Rule them in": the three stage keys of §4.
- The goal: `speed-ratio-aggregate-longlived` at or below 5; budget 7 days.
- "Relative and ≤ 5": the after-reading's median must be below the
  before-reading's minimum, taken interleaved in one session, and at or
  below 5.
- "Measure first, then decide".
- "We need to eliminate the map".
- "V2: both (Recommended)", after the combined prototype.
- "V1 now, V2 measured after (Recommended)", after round 4. Ship the exact
  state-free filter first. The recognizer filter becomes a follow-up decision
  with its own measurement, once its effect on `Expected` and on refusal
  reasons is designed for.
- "Let's get ourselves to 3.92 first then look": allocation and the other
  stages come after this work.

## 2. Evidence

All of it was taken on `pu` 8c3317079 and is in the evidence directory. Ratios
are over recursive descent's time on the same files. They were measured on one
machine, where today's whole pipeline reads 5.62 against the committed
baseline's 4.79, so compare them with each other, not with `baseline.txt`.

- **Where the time is.** On the ratchet's aggregate (2,211 files, three quiet
  runs), the precedence constraint `forest.Constrain` is 2.2–2.4, about 43%
  of the pipeline. It judges each derivation in bison's LR(0) parser state,
  memoised in hash maps keyed by node and state. Removing in advance
  everything it drops still leaves 0.88 of its time.
- **The parser state is not needed.** A filter that tracks no parser state
  matched the constraint's verdict, kept set and readings count everywhere it
  was run:
  - the referee's 1,296 sentences;
  - 25,490 toy forests, and 4,395 toy sentences against bison;
  - the 2,211 aggregate files;
  - 11,574 corpus files;
  - 20,083 operator sentences, 185 of them refused by precedence;
  - 15,000 Leo-chain and nullable-wrapped sentences, 989 of them refused by
    precedence.
- **Its speed.** With the memo held in per-node lists and dense tables, and
  no hash map, the pipeline takes 0.73 (0.72–0.74) of today's time and the
  precedence stage 0.41. No node holds more than 24 memo entries.
- **What was deferred, and why.** Adding bison's reduce side to recognition
  took the pipeline from 0.73 to 0.69. It also changes the refusal reason and
  byte on inputs shaped like `print $fh [];`. And on broken input such as
  `print $fh [ ) ;` it drops `]` from `Expected`. That is perigrin's "V2
  measured after".

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

   The tables are generated from the grammar, by the reading of levels,
   `%prec` and associativity that `forest.NewPrecedence` makes, and by
   bison's resolution of each (rule, terminal) pair over its automaton.
2. **What generation guards.** If any key's decision differs between bison's
   states, table generation returns an error through `NewPrecedence`'s
   existing error path, so such a grammar is refused before it parses
   anything. This guard proves the tables are well-defined. It does not
   prove the evaluator supplies the right key; the oracle of §3.4 covers
   that. A toy grammar that makes generation fail is added to the tests.
3. **No hashing on the evaluation path.** No hash table, built-in or
   hand-written, is read or written per derivation. Per-node lists may grow,
   and a test forces several entries onto one node. The maps `NewPrecedence`
   builds once per grammar stay. Per-call state lives on the call, never on
   the shared, cached tables.
4. **Identical output.** For every input, these equal today's:
   - the parse verdict;
   - the refusal reason and byte;
   - `Expected`;
   - the kept set;
   - the readings count.

   Today's `Constrain` is kept as a test-only oracle. A differential test
   compares all five over:
   - the aggregate;
   - the referee's sentences, refused ones included;
   - the toy referees;
   - the precedence-refused fixtures (`waypoint2Refusals`, the port's refused
     cases).

   Both forest digests and every ratchet count key are unchanged.
5. **The referee judges production.** `chart.CountPrecedence` and the toy
   referee tests run the production judgement, so `precedence-disagree` and
   `precedence-covered` measure what ships. The oracle's own agreement with
   bison stays tested.

**What this does not amend.** Recognition stays Boolean and items stay
(state, distance), as v2 §3 rule 12 has it. Precedence is still applied after
the chart, as a constraint on the forest, as v2 §2, §8 and rule 6 have it.
v2 §8 defines correctness by bison's resolution, pinned by generated tests;
this decision meets that standard by test rather than by replaying bison's
states. The live documents that describe the mechanism change first, in the
same pull request: the chart README's precedence-constraint section, and the
ratchet key table.

## 4. Measurement

The three stage keys come first, in a commit of their own on `pu`, before the
precedence change:

- `speed-stage-precedence-aggregate-longlived`;
- `speed-stage-resolve-aggregate-longlived`;
- `speed-stage-readings-aggregate-longlived`.

Each timed pass works on a fresh copy of the forest, made outside the span,
because the constraint and resolve change their input. They are reported
timing keys, never gated and never compared with the baseline, and the commit
that adds them reports their cost to a ratchet run.

The last issue takes the verdict reading: three quiet measurements of the
before-state and the after-state, interleaved in one session. "Quiet" means a
1-minute load at or below the core count. It records them in Records with the
commit, load, core count and kernel, giving every stage key's median, minimum
and maximum before and after.

## 5. Exit, stops and what the result decides

**Exit:**

- §3 is built;
- identical output holds;
- perigrin's relative exit is met by `speed-ratio-aggregate-longlived`.

**Stops,** reported to perigrin and never tuned toward:

- anything the differential test compares differs from the oracle;
- table generation finds a state-dependent key that §3.1's context does not
  settle;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- that median is below the before-reading's minimum, but
  `speed-ratio-aggregate-longlived` stays above 5;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget.

**What the result decides.** The before-and-after stage table goes to
perigrin. The next step is his ruling: the recognizer's reduce-side filter
(V2), allocation, or another stage.

## 6. Out of scope

- The recognizer's reduce-side filter (V2), deferred by ruling.
- A value on items.
- Grammar rewriting.
- Changing what precedence decides, including the shortcut where an
  alternative kept in one context is kept in all.
- Speed work on any other stage.

Only `internal/chart` changes, plus this document and its assessment.
