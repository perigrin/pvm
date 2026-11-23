# Chart parser: precedence without parser state

Status: proposed, for crochet:assess (revision 4). Date: 2026-11-23.
Assessment: `docs/assessments/chart-stateless-precedence.md`.

Revisions 1 to 3 were `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.
The file was renamed at revision 4, because what is decided changed twice
along the way; the assessment minute records each revision.

The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, branch
`docs/chart-recovery-plan`; perigrin's D2, 2026-11-13).

## 1. Rulings

All perigrin's, of 2026-11-23. They are recorded here, and at refinement in
the milestone's Records.

- "Do just B first and let's re-measure". B was the reduce-side filter of
  revision 2.
- "Rule them in": the three stage keys of §5.
- The goal: `speed-ratio-aggregate-longlived` at or below 5, with a budget
  of 7 days.
- "Relative and ≤ 5": the exit is met when the after-reading's median is
  below the before-reading's minimum, taken interleaved in one session, and
  is at or below 5.
- "Measure first, then decide".
- "We need to eliminate the map".
- "V2: both (Recommended)", chosen from the combined prototype's three
  variants (§2): a state-free, map-free precedence filter in the forest, plus
  the reduce-side filter in recognition and the forest build. Today's
  constraint stays as a test-only oracle, and the files are renamed.

## 2. Evidence

All of it was taken in scratch worktrees on `pu` 8c3317079 and is recorded in
the assessment minute. Ratios are over recursive descent's time on the same
files. They were measured on one machine, where today's whole pipeline reads
5.62 against the committed baseline's 4.79, so compare them with each other,
not with `baseline.txt`.

- **Where the time is.** Over the ratchet's own aggregate (2,211 files, three
  quiet runs), the precedence constraint (`forest.Constrain`) is 2.2–2.4,
  about 43% of the pipeline. It judges each derivation in bison's LR(0)
  parser state, and memoises those judgements in hash maps keyed by node and
  state.
- **Removing alternatives first does not help.** With everything the
  constraint drops removed beforehand, it still takes 0.88 of its time. Its
  cost is in judging the forest it keeps, not in what it removes.
- **The parser state is not needed.** A filter that tracks no parser state
  gives the same verdict and kept set as the constraint everywhere tested:
  - the referee's 1,296 sentences;
  - 25,490 toy forests, and 4,395 toy sentences against bison;
  - all 2,211 aggregate files.

  It keys its judgements on the rule, the terminal after the span, the parent
  item (needed only for `indirob` and `myterm` before `[` or `{`), and the
  spine-top item, the item the node's left spine hangs from. Each of these
  matters: dropping any one of them changes output on the aggregate or the
  toys.
- **The combination is exact and faster.** The prototype compared three
  variants with today's code:
  - V1: the state-free filter, its memo held in per-node lists with no hash
    map, and dense tables;
  - V2: V1, plus bison's reduce-side precedence applied at completion in
    recognition and in the forest build.

  Over the aggregate, against today's code:

  | | V1 | V2 |
  |---|---|---|
  | Whole pipeline | 0.73 (0.72–0.74) | 0.69 (0.69–0.72) |
  | Precedence stage | 0.41 | 0.40 |
  | Differences from today: referee, toys or aggregate | 0 | 0 |
  | Recognition acceptance | | unchanged on every aggregate file |
  | Items per token | | 4.70 → 4.47 |
  | Packed alternatives | | 5.65% fewer |
  | Recognize stage | | 0.89, but its range (0.88–1.10) straddles noise |

  Memo contexts per node: at most 10 when evaluating and 24 in all.

## 3. The decision

**Precedence is judged without parser state, and is partly applied before
the forest is built.**

1. **The forest's precedence judgement uses no parser state.** Its tables
   are generated from the grammar, by the reading of levels, `%prec` and
   associativity that `forest.NewPrecedence` makes, and bison's resolution of
   each (rule, terminal) pair. A judgement depends only on:
   - the rule;
   - the terminal after the span;
   - the parent item, where a reduction's decision differs between states;
   - the spine-top item, for the shift side.

   If table generation finds any key whose decision differs between bison's
   states, it fails, so a grammar outside this property is refused when the
   tables are built, never mis-parsed.
2. **No hashing on the evaluation path.** No hash table, built-in or
   hand-written, is read or written per derivation. Per-node lists may grow.
   A test forces several entries onto one node, so the multi-entry path runs.
   The maps `NewPrecedence` builds once per grammar stay. Per-call state
   lives on the call, never on the shared, cached tables.
3. **Bison's reduce side is applied before the forest.** Recognition refuses
   to complete a rule when every terminal that can follow it at that point is
   one before which bison's precedence never reduces it. The forest build
   refuses the same completions, except over empty spans. A Leo memo never
   completes a link the filter refuses. With lookahead off, the filter does
   not run.

What must become true:

- **Identical output.** For every input, the parse verdict, the refusal
  reason, the kept set and the readings count equal today's.
  - Today's `Constrain` is kept as a test-only oracle. A differential test
    compares verdict and exact kept set per file over the aggregate, the
    referee and the toy referees.
  - `resolved-forest-digest.golden` and every ratchet count key are
    unchanged.
  - `forest-digest.golden`, which is taken before precedence, moves once,
    because §3.3 removes alternatives from the built forest. It moves under a
    check that every alternative it loses is one the oracle also drops.
- **Recognition's contract.** With lookahead on, the recognizer accepts a
  subset of what the plain-Earley reference accepts, and at least what the
  reference accepts and precedence leaves a derivation for. With lookahead
  off, it accepts exactly what the reference accepts. `Expected` is computed
  with lookahead off, so it is unchanged. The reference stays without
  precedence.
- **The live documents change first,** in the same pull request as the code:
  - the chart README's description of the constraint;
  - the recognizer's `Options` contract;
  - the ratchet key table.

## 4. Measurement

The three stage keys come first, in a commit of their own on `pu`, before any
change to precedence:

- `speed-stage-precedence-aggregate-longlived`;
- `speed-stage-resolve-aggregate-longlived`;
- `speed-stage-readings-aggregate-longlived`.

Each timed pass works on a fresh copy of the forest made outside the span,
because the constraint and resolve change their input. They are reported
timing keys, like every `speed-` key: never gated and never compared with the
baseline. Their cost to a ratchet run is reported with the commit that adds
them.

Speed is read as chart-recovery's verdict was. The last issue takes three
quiet measurements of the before-state and the after-state, interleaved in
one session, and records them in Records with the commit, load, core count
and kernel. "Quiet" means a 1-minute load at or below the core count. The
reading gives every stage key's median, minimum and maximum, before and after.

## 5. Exit, stops and what the result decides

**Exit:**

- §3's three parts are built;
- identical output holds;
- the recognition contract holds;
- perigrin's relative exit is met by `speed-ratio-aggregate-longlived`.

**Stops,** reported to perigrin and never tuned toward:

- any output under identical output differs from the oracle;
- table generation finds a state-dependent key it cannot settle with the
  context named in §3.1;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- the precedence stage is faster but `speed-ratio-aggregate-longlived` stays
  above 5;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget.

**What the result decides.** The before-and-after stage table goes to
perigrin. Whether there is a next stage of speed work, and which stage it
targets, is his ruling.

## 6. Out of scope

- A value on items (revision 1's Option C). The forest build applies the
  shift side, where each child's rule is known.
- Grammar rewriting.
- Changing what precedence decides, including the shortcut where an
  alternative kept in one context is kept in all.
- Speed work on lex, resolve and readings.

Only `internal/chart` changes, plus this document and its assessment.
`TestOracleCatchesSkewUnderOneOption`, which counts the recognizer's
`Options`, is extended in the same change.
