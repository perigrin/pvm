# Chart parser: precedence without parser state

Status: accepted by perigrin (revision 9), 2026-11-23, after crochet:assess
reached its bound of eight rounds. Date: 2026-11-23.
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

All are perigrin's, of 2026-11-23. The minute records every ruling, including
superseded ones, and refinement records these in the milestone's Records.

- **"Apply fixes, accept (Recommended)".** The assessment reached its bound
  of eight rounds without a fixed point. Three participants accepted. Round
  8's fresh participant raised four one-sentence decision fixes (M1–M4) and a
  labelling fix (M5). perigrin ruled that revision 9 applies them and accepted
  it.
- **"V1 now, V2 measured after (Recommended)".** Ship the exact state-free
  filter first. The recognizer's reduce-side filter (V2) becomes a follow-up
  decision with its own measurement, once its effect on `Expected` and on
  refusal reasons is designed for.
- **"Drop the number".** There is no numeric target for V1 and V2 together.
  The V2 decision sets its own after this decision's verdict reading.
- **"V2's verdict reading (Recommended)".** Allocation and the other stages
  wait until V2 is built and measured.
- **"Split it (Recommended)".** Where the aggregate check runs (§3.4).
- **"Rule them in (Recommended)".** The `Makefile` target
  `chart-precedence-reference` and the `.github/workflows/ci.yml` Ratchets
  job step that runs it are ruled in, for the issue that adds that target.
- **"Rule them in".** The three stage keys of §4.
- **"Relative and ≤ 5".** This applies to the goal's
  `speed-ratio-aggregate-longlived` at or below 5. The exit is met when the
  after-reading's median is below the before-reading's minimum, read as §4
  defines, and is at or below 5. The budget is 7 days.
- **"We need to eliminate the map".**

## 2. Evidence

All of it was taken on `pu` 8c3317079, on one machine. Ratios are over
recursive descent's time on the same files. Compare them with each other, not
with `baseline.txt`, whose `speed-ratio-aggregate-longlived` reads 4.79 for
today's code.

### Where the time goes

`Constrain` takes the pipeline's largest share. It judges each derivation in
bison's LR(0) parser state, and memoises those judgements in hash maps keyed
by node and state.

| Source | `Constrain` | Share of the pipeline |
|---|---|---|
| `constrain-runs/` | 2.33–2.45 | 43% |
| `combined-runs/` | 2.18–2.23 | 39% |

### The parser state is not needed

A filter that tracks no parser state matched `Constrain`'s verdict, kept set
and readings count everywhere it was run:

| Input | Size |
|---|---|
| The referee's sentences | 1,296 |
| Toy forests | 25,490 |
| Toy sentences, checked against bison | 4,395 |
| Aggregate files | 2,211 |
| Corpus files, perl-refused files included | 11,574 |
| Operator sentences (seed 4), 185 refused by precedence | 20,083 |
| Leo-chain and nullable-wrapped sentences (seed 44), 989 refused | 15,000 |

The last three rows are reported, not kept. Their generators and harnesses
are in the evidence, but their outputs are not. Rounds 4 and 6 ran them; the
records are `reports/round4-fresh.md` and `reports/round6-fresh.md`.

### V1's speed

V1 holds its memo in per-node lists and dense tables, with no hash map.

| Reading | Pipeline, as a share of today's | Ratio over recursive descent |
|---|---|---|
| Combined prototype, three processes | 0.73 (0.72–0.74) | 4.03–4.17 |
| Round 5, independent | 0.727 | 4.12 |

- The precedence stage takes 0.39–0.42 of today's time, per run.
- No node holds more than 24 memo entries.
- Building V1's dense tables took 105–137 ms once per process, on top of
  `NewPrecedence`'s own build. This is reported in `reports/round7-fresh.md`,
  not kept.

### V2, for its later decision

V2 measured 3.85–4.02 in the combined prototype. Round 5's independent run
measured 4.10, against V1's 4.12 in the same run, so V2's gain over V1 did
not reproduce.

## 3. The decision

**The forest's precedence judgement uses no parser state and no hashing.
Recognition and the built forest are unchanged.**

### 3.1 Keys

A judgement depends only on these:

- **Reduce side:** the rule and the terminal after the span. For the six
  (rule, terminal) pairs whose decision differs between bison's states
  (`indirob` and `myterm` before `[` and `{`), the parent item is added.
- **Shift side:** the spine-top item (the item the node's left spine hangs
  from, or the start item at the root), the rule-position item before the
  shift, and the terminal shifted.

The tables are generated when `forest.NewPrecedence` runs, from the grammar's
levels, `%prec` and associativity and from bison's resolution of each (rule,
terminal) pair over its automaton. They are stored on the `Precedence` value.

### 3.2 What generation guards

If any key's decision differs between bison's states, `NewPrecedence` returns
an error, and `ParseWith` turns it into a refusal of every file.

- **What it proves:** the tables are well-defined.
- **What it doesn't prove:** that the evaluator supplies the right key. The
  reference of §3.4 covers that.
- **What it may refuse:** the guard is stricter than today's `Constrain`, so
  it may refuse a grammar `Constrain` handles. perly.y passes.
- **New work:** a toy grammar that makes generation fail.

### 3.3 No hashing on the evaluation path

- No hash table, built-in or hand-written, is read or written per derivation.
- Per-node lists may grow, and a test forces several entries onto one node.
- The maps `NewPrecedence` builds once per grammar stay.
- Per-call state lives on the call, never on the shared `Precedence`.

### 3.4 Identical output, checked against today's judgement

**The reference.** Today's method stays as `(*Forest).ConstrainReference`,
documented as used only by tests.

- `(*Forest).Constrain` becomes the state-free judgement, so every existing
  caller switches to it: `ParseWith`, the referee, and the forest and chart
  precedence tests.
- The reference cannot leave `forest`, because it reads `Precedence`'s
  unexported automaton.
- New work: a test holding the reference's agreement with bison.

**What is compared.** Each check takes the forest its caller passes to the
constraint today and compares the verdict and the exact kept set:

- For `ParseWith` and the corpus checks, that is the forest after
  `libraryOperators`.
- For the referee, it is `Build`'s forest narrowed to the sentence's
  terminals.

Equal kept sets give equal readings. Nothing before the constraint changes,
so recognition, the refusal byte and `Expected` cannot differ.

**The inputs, and where each runs.** Under "Split it":

| Input | Runs in | Corpus needed |
|---|---|---|
| The referee's sentences, refused ones included | `make test` | none |
| The toy referees | `make test` | none |
| `waypoint2Refusals` | `make test` | none |
| The seeded generators of §2 (new tests, rewritten from the evidence harnesses against production and the reference) | `make test` | none |
| Every recognized file of the ratchet's in-tree corpora (T1, Guacamole parses and parsent, std5k, edit5k), perl-refused files included (new test in `cmd/ratchet`, using §4's new API) | `make test` | none |
| Every recognized perl.git `t/` file (new) | `make chart-precedence-reference` (new target, own build tag) | perl.git at 94e50866 |

- **The seeded generators' test** asserts agreement on every sentence and
  that at least one sentence is refused. It does not assert today's 185 and
  989.
- **If a `make test` check proves slow,** its bulk moves behind a build tag
  of its own that `make test` still runs, as it runs `chart-absence-reference`.
  It never moves to the perl.git target, which `make test` does not run.
- **The perl.git target** is not run by `make test`. CI's corpus-backed
  Ratchets job runs it with `PERL5_CORPUS` set, and it fails unless it
  passes. A developer without perl.git runs `make test`.

**What stays unchanged:** both forest digests and every ratchet count key.

### 3.5 The referee judges production

`chart.CountPrecedence` and the toy referee tests run the production
judgement, so `precedence-disagree` and `precedence-covered` measure what
ships.

### What this does not amend

- Recognition stays Boolean and items stay (state, distance), as v2 §3 rule
  12 has it.
- Precedence is still applied after the chart, as a constraint on the forest,
  as v2 §2, §8 and rule 6 have it.
- v2 §8 defines correctness by bison's resolution, pinned by generated tests.
  This decision meets that standard by test, rather than by replaying bison's
  states.
- v2 §11's "only `internal/chart` changes" is departed from only by the
  `Makefile` and CI paths perigrin ruled in.
- v2 §9.5 says no key is added after the reset. It is departed from only by
  the three stage keys of §4, which perigrin ruled in. The Recovery Plan's D5
  did the same for its keys.

The live documents change first, in the same pull request: the chart README's
precedence-constraint section, and the ratchet key table.

## 4. Measurement

The three stage keys come first, in a commit of their own on `pu`, before the
precedence change. That commit is the before-state.

- `speed-stage-precedence-aggregate-longlived`
- `speed-stage-resolve-aggregate-longlived`
- `speed-stage-readings-aggregate-longlived`

**New work they need:**

- A per-pass preparation hook in the ratchet's timing protocol. Today
  `Protocol.pass` starts the clock before anything per pass runs. The hook
  makes a fresh copy of each forest outside the span, holding one pass's
  copies at a time.
- An exported way to copy a forest.
- Exported access to `libraryOperators`, resolve and the readings count. All
  three are unexported today.

**How they behave:**

- The precedence stage times the forest after `libraryOperators`.
- Like every `speed-` key, they are printed, listed in `baseline.txt`, and
  never gated.
- The commit adding them changes the ratchet's key list, its tests and the
  README key table.
- It reports the keys' cost to a ratchet run. That cost must keep the run
  inside `chart-ratchet.yml`'s timeout.
- Library operators get no key, so the stage keys do not sum to the whole
  pipeline.

**The verdict reading.** The last issue takes it, and records it in Records.

- **One measurement** is one ratchet run's printed value of a key. That value
  is the protocol's own median of its three internal measurements.
- **Runs:** three quiet runs of the before-state and three of the
  after-state, interleaved in one session. "Quiet" means a 1-minute load at
  or below the core count.
- **What it records:** for `speed-ratio-aggregate-longlived` and every stage
  key, before and after, the median, minimum and maximum over those three
  runs. It also records each run's commit, load, core count and kernel.

## 5. Exit, stops and what the result decides

**Exit:**

- §3 is built;
- the checks of §3.4 pass;
- perigrin's relative exit is met by `speed-ratio-aggregate-longlived` in the
  verdict reading.

**Stops,** reported to perigrin and never tuned toward. A stop overrides the
exit: if both hold, the milestone stops.

- any §3.4 check fails, whether it differs from the reference, the reference
  disagrees with bison, an assertion fails, or the perl.git target fails or
  times out in CI. A forest digest or ratchet count key moving is also a stop;
- table generation finds a state-dependent key that §3.1's context does not
  settle;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- §3 is built and its checks pass, but the exit is otherwise not met;
- a quiet session cannot be had for the verdict reading. Then the reading
  records the load, and perigrin rules on it;
- the stage keys push the ratchet run past `chart-ratchet.yml`'s timeout;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget. At the budget itself, the milestone reports its status to
  perigrin.

**What the result decides.** The before-and-after table goes to perigrin, and
the V2 decision comes next, setting its own target. Allocation and the other
stages wait for V2's verdict reading.

## 6. Out of scope

- The recognizer's reduce-side filter (V2), deferred by ruling.
- A value on items.
- Grammar rewriting.
- Changing what precedence decides. That includes the shortcut where an
  alternative kept in one context is kept in all.
- Speed work on any other stage.

Changes are limited to:

- `internal/chart`;
- the `Makefile` target and the CI job step ruled in (§1);
- this document and its assessment.
