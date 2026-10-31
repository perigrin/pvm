<!-- ABOUTME: Orientation for internal/chart, the chart-parser research spike beside internal/parse. -->
<!-- ABOUTME: Covers the Guacamole corpus, the convergence edits, the chart-ratchet output contract, the lexer lists, the recognizer, and the generated tables. -->

# internal/chart

An Earley chart parser for Perl, built as a research spike beside the
recursive-descent parser in `internal/parse` and measured against it. The
design is `docs/plans/2026-10-29-chart-parser-design.md`. Nothing outside the
spike consumes this tree.

## Status: stopped by the speed kill

The `chart-waypoint` milestone was stopped on 2026-10-30 by its §9 speed
kill: lexing and recognition ran 3.9–5.0× recursive descent over the corpus,
against a 3× line. `docs/postmortems/chart-waypoint.md` records why. perigrin
decided to re-attempt the spike with a grammar ported from `perly.y` over
`internal/lexer`'s own tokens. The LR(0) generator, the recognizer, Leo and
the ratchet carry over. Guacamole's grammar, the SLIF reader, the token cut,
`chart/lex`'s predicate mapping, the observed-conflict probes and the CORE.tpm
cross-check are removed, and no grammar is committed until the port. The
history of what they were is in `docs/postmortems/chart-waypoint.md`. The
forest, error recovery, `chart/compare` and the oracle adapter were never
built.

## The Guacamole corpus

`testdata/guacamole/` holds the waypoint's corpus: every `parses` and
`parsent` argument of Guacamole's own suite, recorded by running it with the
stand-in `testdata/guacamole/Guacamole/Test.pm`, one case file per unique
text. Each case is `preamble.pl` (one line) followed by Guacamole's bytes
unchanged, so the whitespace pre-commit hooks skip the case directories.
`stubs/` supplies the modules cases `use` that perl cannot find.

Which cases perl accepts is said by the committed perl labels,
`testdata/labels/in-tree.txt`, not by running perl: the pinned `-DDEBUGGING`
perl's `perl -c` verdict on every case, with the stubs as `PERL5LIB`, and a
category for each refusal (spec §9.1). `offline/label.sh` writes them and
`offline/README.md` gives their format. A `parses` case perl refuses is not
scored. `TestPerlLabels` checks every case is labelled under the pinned perl.

`extract.pl` regenerates the whole directory from a Guacamole checkout.
`PROVENANCE` gives the revision and the command. Nothing here reads the
checkout afterwards. `TestGuacamoleCasesUnique` and
`TestGuacamoleCasesStartWithPreamble` guard the files.

## The convergence edits

`testdata/convergence-edits.jsonl` holds the 200 single-statement edits to
`testdata/std5k.pl` that §8 fixes before the first step.
`testdata/convergence-edits-guacamole.jsonl` holds the same kind of edits to
the corpus's three multi-statement `parses` cases, those with at least two
editable statements after the preamble. Each line is one edit:
`offset`, `removed` and `inserted` bytes, and the statement span
`stmt_start`..`stmt_end` it lies in. The corpus list also names its `case`.
`drawEdits` draws both lists with a fixed seed, over recursive descent's
`;`-terminated statements that hold no other such statement. The guards hold
each edit inside its statement and require the edited file to parse with no
refusal. **The lists are fixed once committed.** A later step may add a list,
but must never redraw one: `-chart.update-convergence-edits` exists only to
write a list for the first time.

## The `make chart-ratchet` output contract

`make chart-ratchet` runs `cmd/ratchet`. Every issue of the spike closes
against it, so this section is a contract. The key set below was fixed by the
reset of `docs/plans/2026-10-30-chart-parser-v2-design.md` §9.5: **no key is
added, renamed or retired after it.** Every key of the first attempt is
retired.

- **Line form.** `make -s chart-ratchet` prints one line per key on stdout,
  `chart-ratchet: <key> <value>`, every key once, in the fixed order of the
  table below, and nothing else there. Diagnostics go to stderr.
- **`na`.** A key nothing measures yet prints `na`. The issue that makes it
  print a number takes its baseline in the same commit. A key whose baseline
  is `na` and that prints a number fails with "baseline still na; take a
  baseline", and a number may not become `na`.
- **Exit rule.** `cmd/ratchet` exits 1 on a regression against
  `testdata/ratchet/baseline.txt`, naming the key, the baseline and the value
  on stderr, and 2 on an error: a measurement that cannot be taken, a missing
  corpus or label, a generator refusal, or a baseline that is malformed,
  names a key the ratchet does not print, or omits one. It never exits
  non-zero because the waypoint's exit is still unmet. `go run` turns every
  non-zero exit into 1 and `make` then returns 2, so through `make` only the
  stderr text (`exit status 2`) tells an error from a regression.
- **Inputs.** The ratchet reads only the working tree: no commit, no
  `refs/zhi`, no network. It reads nothing under Guacamole's paths and
  does not import `chart/lex`.
- **Budget.** One run takes at most 300 s on the development machine.
  Corpus-wide sweeps and benchmarks belong here, never in `make test`.
- **Mainline.** `make test` never runs the ratchet. Its CI job,
  `.github/workflows/chart-ratchet.yml`, runs on changes under
  `internal/chart/**` and `internal/lexer/**` and is advisory
  (`continue-on-error: true`).

**Isolation (§11).** The chart reads `internal/lexer` and `internal/parse`
and never changes them. A lexer defect found here is filed against the v0.1
milestone and listed — on `lex-unmapped`'s known list, or on the lexer-defect
list — never worked around.

To take a baseline after an intended change, run
`go run ./internal/chart/cmd/ratchet > /tmp/baseline.txt`, move it over
`internal/chart/testdata/ratchet/baseline.txt` and review the diff.

### The rules

- **May not rise:** `tables-drift`, `leo-disagree`, `ledger-gaps`,
  `rd-accepted-chart-refused`, `chart-accepted-both-refused`, `lex-unmapped`.
  **May not fall:** `tdt-aligned`.
- **Timing** keys (`lex-overhead*`, `speed-ratio-*`, `alloc-per-token-*`) are
  never compared with the baseline. `lex-overhead` fails **absolutely** when
  it prints a number above 1.3, whatever its baseline.
- **From the port's completion**, read from the same run's `ledger-gaps`
  being `0`: `tdt-mismatch`, `tdt-unseen-unlisted` and
  `tdt-unaligned-unlisted` must be 0; `terminals-per-token` and `tdt-unseen`
  may not rise; and the files `lex-unmapped` counts must equal
  `testdata/lex-unmapped-known.txt` as `(path, first unmapped byte)` pairs.
  Each entry is `<path> <byte> <issue ref>` and the comparison reads its
  first two fields. A listed file refused at another byte is a mismatch, so
  an entry cannot name an arbitrary byte of a refused file. Before
  completion `lex-unmapped` is only a may-not-rise count: the chart's own
  table is still growing, and those files are not lexer defects.
- **Reported** only: every other key.
- A rule exists only for a key the ratchet prints.

| Key | Meaning | Rule |
|---|---|---|
| `tables-drift` | committed tables differing from what `make chart-tables` writes | may not rise; 0 at the exit |
| `leo-disagree` | files where Leo on and off differ in acceptance or `Expected` at any position | may not rise; 0 at the exit |
| `ledger-gaps` | reference alternatives with no ledger mark, plus grammar citations not in the ledger | may not rise; 0 at the exit |
| `rd-accepted-chart-refused` | superset-population files the chart refuses | may not rise; 0 at the exit |
| `chart-accepted-both-refused` | perl grammar refusals recursive descent also refuses that the chart accepts, less both lexer lists | may not rise; 0 at the exit |
| `lex-unmapped` | files with a token the chart's table leaves unmapped | may not rise; known list from completion |
| `tdt-aligned` | tokens aligned with perl's `-DT` | may not fall |
| `tdt-mismatch` | aligned tokens whose terminal set misses perl's | 0 from completion |
| `tdt-unseen` | table pairs perl's `-DT` never showed | may not rise from completion |
| `terminals-per-token` | the table's width | may not rise from completion |
| `tdt-unseen-unlisted` | `tdt-unseen` pairs not on the reviewed list | 0 from completion |
| `tdt-unaligned-unlisted` | unaligned tokens outside a reviewed fitting family | 0 from completion |
| `lex-overhead` | chart/lex time over `lexer.Tokenize` time, median of three | timing; ≤ 1.3 absolutely |
| `lex-overhead-std5k` | the same on `std5k.pl` | timing |
| `speed-ratio-std5k-longlived` | chart over recursive descent on `std5k.pl`, long-lived (§9.3) | timing |
| `speed-ratio-std5k-cold` | the same, cold | timing |
| `speed-ratio-aggregate-longlived` | the same over the aggregate, long-lived | timing |
| `speed-ratio-aggregate-cold` | the same, cold | timing |
| `alloc-per-token-chart` | the chart's allocations per token | timing |
| `alloc-per-token-rd` | recursive descent's allocations per token | timing |
| `items-per-token` | Earley items per token | reported |
| `set-reuse` | share of positions whose Earley set was already interned | reported |
| `ledger-excluded` | ledger alternatives marked excluded | reported |
| `superset-population` | files both perl and recursive descent accept | reported |
| `superset-recognized` | share of that population the chart accepts | reported |
| `rd-accepted-perl-refused` | files recursive descent accepts and perl refuses | reported |
| `chart-accepted-perl-refused` | files the chart accepts and perl refuses | reported |

### The speed protocol

Every timing ratio — `lex-overhead*` and `speed-ratio-*` — comes from one
protocol in `cmd/ratchet/speed.go` (spec §9.3), code over injected sides, a
clock and a collector, so it is tested on fake timings
(`TestSpeedProtocol`, `TestSpeedProtocolExclusionBoundary`,
`TestSpeedProtocolRefusesUnknownTimedPath`):

- **Alternation.** The chart side and the reference side (recursive descent,
  or `lexer.Tokenize` for `lex-overhead`) are timed in alternating passes in
  one process: chart, reference, chart, reference. Each pass covers every
  file of the set once. A GC is forced before each pass, outside the timed
  span; within a pass collection is left to the runtime.
- **A measurement** is 2 warm-up pairs then 9 timed pairs; its ratio is the
  median of the 9 per-pair ratios (chart pass over reference pass), with min
  and max printed on stderr. A figure is the median of three measurements.
- **Lifetimes.** Long-lived: one parser per pass, reused across its files.
  Cold: a fresh parser per file.
- **The file set is an input.** `lex-overhead` passes every corpus file
  until the superset population exists and the aggregate after it; the speed
  ratios pass `std5k.pl` alone and the aggregate.
- **The aggregate** is the timed set less `std5k.pl`, `edit5k.pl` and any
  file whose recursive-descent time alone is more than 10% of recursive
  descent's time over the **whole** population, whatever subset is timed.
  Exactly 10% stays in. Recursive descent is passed to `Aggregate`
  explicitly, even when the reference side is `lexer.Tokenize`, so every
  figure reads one aggregate. A timed file outside the population is
  refused, never kept unchecked. Each excluded file and each corpus's share
  of the aggregate's recursive-descent time are printed on stderr.

Wiring real chart and recursive-descent runs into the protocol is the job of
the issues that make those keys print a number.

### The lexer lists

Two lists name the files a lexer defect takes out of the gates (§11):
`testdata/lex-unmapped-known.txt`, for a token the chart's table leaves
unmapped, and `testdata/lexer-defects.txt`, for a token that is mapped but
wrong. Refusal counts exclude the files on either list and no others. Both
start empty; an entry is added only by the issue that meets the defect, filed
against the v0.1 milestone, and the gate review confirms the construct at the
entry's byte is that issue's.

Each non-blank line is one entry, `<path> <byte> <issue ref>`:

- **path** names a corpus file. An in-tree file is repository-relative:
  `internal/chart/testdata/std5k.pl`, `internal/psc/testdata/latency/edit5k.pl`,
  or a `*.pl` directly under `internal/chart/testdata/guacamole/parses/` or
  `.../parsent/`. A T1 file is `T1/<name>.t`, a top-level `*.t` of T1,
  resolved as `$T1_CORPUS/<name>.t`.
- **byte** is a decimal offset inside that file (below its size), with no
  sign or leading zero, since the ratchet matches `<path> <byte>` as text: for
  `lex-unmapped`, the first byte the chart leaves unmapped; for a defect, the
  first byte of the wrong token.
- **issue ref** is a prefix of the open lexer issue's git-zhi uuid made of
  whole segments, at least two (`01a1882d-d8a4`), extended segment by segment
  until it names one issue: the first two segments are the creation time to
  the millisecond, so issues from one refinement run share them. Never the
  second segment alone or the 8-character short id.

`TestLexerLists` checks every entry's form, that its path is a corpus file
that exists and that its byte lies inside it. It never calls `git zhi`, so
`make test` never reads the chain; a separate criterion resolves the refs and
checks the issues are open. When a list holds a T1 entry and `T1_CORPUS` is
unset or names no directory, the test fails naming `T1_CORPUS`; it never
skips, so run it with `T1_CORPUS` set.

## The recognizer: `recognize`

`recognize.Recognize` runs a YAEP-style Earley recognizer over the generated
tables and a `chart/lex` lattice. Error recovery is not built yet.

- **Input** is a lattice of terminal-set ids; the recognizer never sees a
  terminal name. Each edge's token carries one id, `lex.Token.TerminalSet`,
  an index into `grammar.Tables.TerminalSets`, whose entries are sorted
  lists of terminal symbol ids; a token is read as every terminal in its
  set, so a token whose set holds several is scanned as each of them.
  `BuildTables` leaves `TerminalSets` empty and whoever builds the lattice
  fills it (the tests put one singleton set per terminal). Nodes are
  numbered so that every edge runs forward. `Recognize` validates this and
  returns an error, never a panic, for an edge with a node outside the
  lattice, an edge that does not run forward (a zero-length one included),
  or a set id outside the table. A lattice with no edges is valid: the
  grammar accepts or refuses it.
- **Class keys are set ids.** A token's class is its set id. The lookahead
  at a node is the set id of the edges leaving it, or one end-of-input class
  past the table's last set where none leave; each class's bitset is built
  once in `New`, so the per-token lookup allocates nothing
  (`TestClassKeysAllocationFree`). Where edges with different sets leave one
  node, its lookahead is their union, built there, and the transition cache
  is not used at that node.
- **Not safe for concurrent use.** A `Recognizer`'s intern table and
  transition cache change with every recognition, across inputs.
- **Items** are an LR(0) state and the distance in lattice nodes back to the
  item's origin. A set is its items, sorted, interned as a (core, distances)
  pair, so two positions with equal content share one `*Set`. A node that an
  edge passes over is inside a branch: its set is built but not interned.
- **Lookahead.** In up-front mode a complete item's rule completes only if
  the next token, or the end of input, can follow its LHS. That stops
  right-recursive `StatementSeq` from completing back through every earlier
  statement, which is what keeps sets equal from one statement to the next.
  `Expected(pos)` is the union of the expected terminals of every state in
  the set at `pos` as recognition without lookahead builds it: it closes
  the set again under every completion, outside the interned content, so a
  refusal point still names what was expected there.
- **The transition cache** is keyed by (set, token, lookahead) and the
  edge's span. A key alone does not fix the result: completion reads the
  sets where complete items began, which are not part of the source set. So
  each entry records those sets, and a hit requires them to be the same.
  `TestCacheChecksOriginSets` fails without that check. The cache is not
  used inside a branch, nor where two edges meet, and a transition that read
  any set inside a branch is never cached (`TestCacheNotKeyedOnBranchSet`).
- **Leo memos** (Leo 1991) complete a right recursion in one step. A memo is
  made at a node for a symbol B only where the node's whole set holds one
  item awaiting B, and that item is a Leo item of the tables (`leo.txt`). It
  resolves the chain of completions to its top item, by reference to the
  memo at the item's origin. Memos live in a per-node side table, never in
  a set, so two nodes whose sets differ only in their tops share one set
  (`TestLeoTopOutsideSetIdentity`). With lookahead, a chain whose top cannot
  be followed by the next token is completed link by link, as without Leo.
  Every set a chain read is a transition-cache dependency. Leo never changes
  acceptance. At lookahead 0, where interleaved mode will run, it makes
  `S ::= A S | A` and a toy right-recursive, nullable statement sequence
  (the shape of Guacamole's `StatementSeq`) linear, where without it they are
  quadratic (`TestLeoLinearAtLookaheadZero`,
  `TestLeoStatementSeqAtLookaheadZero`). Linear here means the item count,
  not time: with the cache on, each use of a chain walks it to record its
  sets, so time stays quadratic until memos are interned. A chain stops at
  a start-rule item from node 0, which acceptance must see. `Options.NoLeo`
  turns it off; it is the only Leo switch.

`make fuzz-chart-recognize` runs `FuzzRecognize`. It checks that recognition
never panics, that acceptance is the same with and without the cache, the
lookahead and Leo and with a dead-end branch added, and that `Expected` is the
same with and without lookahead. Each input byte is the singleton set id of
one terminal of the toy statement grammar, so every input is a valid
terminal string.

## The generated tables: `make chart-tables`

`make chart-tables` runs `grammar/cmd/gen`, which builds the tables
(`grammar.BuildTables`) of the committed grammar and writes them to
`grammar/testdata/tables/`. With no grammar committed there is nothing to
generate: it writes no table file, says so on stderr and exits 0. The output
depends only on the grammar's rule order, never on map iteration or the clock,
so regenerating it is byte-identical (§10).

The tables are the split LR(0) automaton with nullable closure, Aycock and
Horspool's ε-DFA:

- A sequence rule is expanded first: `X ::= Y*` becomes `X ::=` and
  `X ::= X Y`, and `X ::= Y+` becomes `X ::= Y` and `X ::= X Y`.
- A **kernel** state holds the items a transition reached. A **predicted**
  state holds the items a kernel state predicts. Both are closed under dot
  advancement past nullable symbols.
- An Earley item's origin is its kernel's origin in a kernel state, and the
  current position in a predicted state.
- **The start state is state 0**, the start symbol's predicted state. The
  grammar has no augmented start rule, so the recognizer **accepts** when the
  last Earley set holds a complete item of one of the start symbol's rules
  whose origin is 0, meaning the item's rule appears in the reduce list of a
  state in that set.

Every line is space-separated numbers and names:

- **`symbols.txt`**: `<id> <kind> <name>`, where kind is `terminal`,
  `nonterminal` or `nullable` (a nonterminal that derives the empty string).
  The nonterminals come first, in order of their first rule, then the
  terminals in name order.
- **`rules.txt`**: `<id> <LHS> ::= <RHS>`, with sequences expanded. An empty
  RHS is an empty alternative.
- **`items.txt`**: `<id> <rule> <dot>`, the integer core items, rule by rule
  and dot by dot.
- **`states.txt`**: `start <symbol>`, then one line per state:
  `<id> kernel|predicted <ε-state> items <item ids> reduce <rule ids>`. The
  ε-state is a kernel state's predicted state, or `-1`. The reduce list holds
  the rules of the state's complete items.
- **`transitions.txt`**: one line per state, `<id> <symbol>:<target> ...`,
  in symbol order. The target is always a kernel state.
- **`expected.txt`**: one line per state, `<id> <terminal ids>`, the
  terminals after a dot in the state's items, in symbol order.
- **`leo.txt`**: one line per Leo item, `<id> <item>`. A Leo item is a
  state's one item awaiting some nonterminal B, with B its rule's last
  symbol and deriving the rule's own LHS at its right edge (`S ::= A . S`).
  Other items in the state do not matter, so `S ::= A . S` is a Leo item
  beside `S ::= A .`. The right edge reaches past trailing nullables, but an
  item with anything after B is not a Leo item: what follows B could still
  read a token, and the memo skips the item that would read it. This is
  necessary, not sufficient: the recognizer also requires the whole Earley
  set to hold one item awaiting B.

These are separate files so that `states.txt` stays under the
check-added-large-files limit.
