<!-- ABOUTME: Orientation for internal/chart, the chart-parser research spike beside internal/parse. -->
<!-- ABOUTME: Covers its status, the perly.y grammar and its fits, the corpora, the ratchet contract, chart/lex, the recognizer, the forest and the tables. -->

# internal/chart

An Earley chart recognizer for Perl, built as a research spike beside the
recursive-descent parser in `internal/parse` and measured against it. The
design is `docs/plans/2026-10-30-chart-parser-v2-design.md` (on the branch
`docs/chart-parser-v2-design`). Nothing outside the spike consumes this tree.

## Status: waypoint 2 stopped by its speed kill

Waypoint 2 (spec §9.4, milestone `chart-perly-w2`) was stopped on 2026-11-13
by spec §10's long-lived speed kill. With the forest timed on the chart side
(lex, recognize, forest build) under the ratchet's own protocol, the
aggregate long-lived chart/recursive-descent ratio read 3.51, 3.64 and 3.47
in three quiet readings at 3487d258, a median of 3.51 against the limit of 3;
the same runs read 2.20 without the forest. The committed ratchet does not
yet time the forest, so its own `speed-ratio-aggregate-longlived` still reads
about 2.2. `docs/postmortems/chart-perly-w2.md` records why the waypoint
stopped and what a successor must know.

What waypoint 2 built before the stop:

- **the forest** over a recognized chart, adding no facts to its items
  (`forest`, below);
- **perl's oracle facts** over the oracle-graded corpora
  (`testdata/oracle/`), and **perl.git `t/` labelled** with the pinned perl's
  verdicts (`testdata/labels/perlgit.txt`), its lexer-rooted findings on the
  lexer lists (below) and its `-DT` captures made offline but not committed;
- **the precedence referee's sentences**, generated from bison's resolutions
  of perly.y and rendered to source where they round-trip through the lexer
  (`testdata/precedence/`);
- the recognizer's grammar-only tables built once per tables value, `x` at a
  statement start read as a word, and the cache and Leo test gaps closed.

Not done, among the issues still open on `chart-perly-w2`: the precedence
constraint, resolve by scope, the readings count, oracle grading, `PLUGIN_*`,
the whole-pipeline ratchet keys, perl.git `t/` parity and speed checkpoint 3.
Where a section below calls a limit "waypoint 2's" work, that work was not
done before the stop and is left to a successor.

### Waypoint 1

This is the spike's second attempt. The first (`chart-waypoint`, design
`docs/plans/2026-10-29-chart-parser-design.md`) was stopped on 2026-10-30 by
its speed kill; `docs/postmortems/chart-waypoint.md` records why. Its LR(0)
generator, recognizer, Leo memos and ratchet comparison carry over; Guacamole's
grammar, the SLIF reader, the token cut, chart/lex's predicate mapping and the
conflict probes are gone. Guacamole's CORE.tpm cross-check is replaced by the
`parse.CoreShapes()` cross-check (spec §5.3; "Builtin classes against
`parse.CoreShapes()`", below). Guacamole's snippets stay, as corpus.

Waypoint 1 (spec §9.2) is built, as the milestone `chart-perly-w1`:

- **the grammar**, `perly.y` at perl blead 94e50866 ported over
  `internal/lexer`'s own tokens, the port complete (`ledger-gaps 0`; below);
- **chart/lex**, one table lookup per significant token (`lex`, below);
- **the recognizer** over terminal-set ids, with the `;` toke.c inserts where
  a scope ends (`recognize`, below);
- **perl's verdicts**, the pinned `-DDEBUGGING` perl's labels and its `-DT`
  tokens aligned with pvm's (`offline/`, `testdata/labels/`, `testdata/tdt/`);
- **the cross-checks** of the builtin classes against `parse.CoreShapes()`
  and of the precedence levels against chapter 4 and
  `parse.LevelsAccountedFor()`, each holding a reviewed difference list
  (below);
- **the ratchet**, with spec §9.5's key set and §9.3's speed protocol.

Whether the waypoint's exit holds is what the ratchet and the milestone's
criteria say, read from a run; speed checkpoint readings are recorded in the
milestone, not here, since a timing is true only of the run that took it.
Waypoint 1's speed figure is a lower bound: it can stop the spike, not pass it
(spec §2).

**Not built.** Recognition reads no precedence and, but for the `isa`
feature, no scope, so a form perl refuses only by precedence
(`$a <=> $b <=> $c;`) or for want of a declaration is accepted here; the
sections below name the ones known. Error
recovery and interleaved lexing (spec §13) are not built either.

Some forms perl refuses as grammar fit neither category: nothing in the
tokens the chart reads marks them, so it accepts them. On the corpus there
are 14 such files, which recursive descent accepts too (so no gated key
counts them; they sit in `chart-accepted-perl-refused`): 10
malformed number literals (`08`, `0xap_+2`, `0x_.8`, `0b2`, `0b_` and kin) in
`parsent/Statements-Expressions-Value-Literal-0{05,08..16}` (v0.1
01a18830-730b); `$foo->$ *` and `$foo->$#` in
`parsent/Statements-Expressions-Value-ArrowDerefVariable-00{1,6}`; and `$#`
and `$*` alone in
`parses/Statements-Expressions-Value-NonLiteral-Variable-0{17,24}`. Perl
refuses the last four since 5.30, which removed `$#` and `$*`; the lexer emits
them as ordinary variables (`$foo->$ *` as `Variable "$ *"`, `$foo->$#` as
`Variable "$#"`), a root filed as v0.1 `01a1b717-ad05`.

Three limits hold off the corpus, where the gates cannot see them:

- **Refused where perl accepts.** `$y = x { 1 };` and `sub x {1} $y = x 1, 2;`
  are refused (`perl -c`: syntax OK). After a statement-start token (a file's
  start, `;`, a block's `{` or `}`, a label's colon) the table
  (`grammar/perly.go`, `addStatementX`) gives the word `x` every reading any
  plain word gets there, so `x {1};`, `sub x {} x {1};` and
  `sub x {1} x 1, 2;` are accepted as perl accepts them
  (`TestScopeDecidedParity`); elsewhere `x` where no operator is
  expected is still `BAREWORD` only, plus `PRIVATEREF` before `(`. The ruling
  (01a18fa6-ec70) makes `x` a word wherever an operator is not expected; the
  rest of that is a shortcut, pinned in `TestPortScopeDecidedKnownLimits`.
  `catch =>` followed by a `$` variable that is no plain name is refused in any
  list (`catch => $_;`, `(catch => $_)`, `foo(catch => $_);`,
  `x(catch => $Foo::x);`; `perl -c`: syntax OK, `catch` is a hash key there) by
  the PERLY_DOLLAR class that guards a `catch (` variable, pinned in the same
  test.
- **Accepted where perl refuses.** `print STDOUT $foo bar();` is accepted
  (perl: syntax error). A non-keyword word before a `Variable` may read as
  `METHCALL0` even after `print` or `say`, and `tdt-unseen` cannot see an
  over-width inside a key it has already seen (`TestPortScopeDecidedKnownLimits`
  pins the `print STDOUT $foo bar();` row). This is waypoint 2's refusal
  work.

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
reviewed category for each refusal (spec §9.1). `offline/label.sh` writes them and
`offline/README.md` gives their format. A `parses` case perl refuses is not
scored. `TestPerlLabels` checks every case is labelled under the pinned perl and
every refusal's category is reviewed.

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
added, renamed or retired after it**, except `lex-overhead-aggregate`
(perigrin, 2026-11-07, spec §9.3). Every key of the first attempt is
retired.

- **Line form.** `make -s chart-ratchet` prints one line per key on stdout,
  `chart-ratchet: <key> <value>`, every key once, in the fixed order of the
  table below, and nothing else there. Diagnostics go to stderr.
- **`na`.** A key nothing measures prints `na`. Every key prints a number
  since waypoint 1 was built; the rule stays because spec §9.5 states it. A
  key whose baseline is `na` and that prints a number fails with "baseline
  still na; take a baseline", and a number may not become `na`.
- **Exit rule.** `cmd/ratchet` exits 1 on a regression against
  `testdata/ratchet/baseline.txt`, naming the key, the baseline and the value
  on stderr, and 2 on an error: a measurement that cannot be taken, a missing
  corpus or label, a generator refusal, or a baseline that is malformed,
  names a key the ratchet does not print, or omits one. It never exits
  non-zero because the waypoint's exit is still unmet. `go run` turns every
  non-zero exit into 1 and `make` then returns 2, so through `make` only the
  stderr text (`exit status 2`) tells an error from a regression.
- **Inputs.** The ratchet reads the working tree and one directory outside
  it: the aligned `-DT` pairs (`TDT_FIXTURES`), with a default (below).
  T1 is in the tree: `testdata/perlonjava/unit` (see
  `testdata/perlonjava/README.md`), and `T1_CORPUS` overrides it. It reads
  no commit, no
  `refs/zhi` and no network. Of Guacamole it reads only the corpus snippets
  under `testdata/guacamole/`, never its grammar. It imports `chart/grammar`
  for the tables, `chart/lex` to time it and to count `lex-unmapped`,
  `recognize` for `leo-disagree`, the verdict keys and the speed ratios,
  `chart` for the perl labels, `internal/lexer` for `lex-overhead`'s
  reference side and the tdt keys, and `internal/parse` for recursive
  descent's verdicts and times.
- **Budget.** One run takes at most 300 s on the development machine.
  Corpus-wide sweeps and benchmarks belong here, never in `make test`.
  `BenchmarkLex`, `BenchmarkRecognize` and `BenchmarkExpected` (on
  `std5k.pl`) keep their bodies in `bench.go`, outside a test file, and
  `BenchmarkForest` in `forest/bench.go`; the ratchet runs the same bodies
  with `testing.Benchmark`, after every key is measured, printing one stderr
  line each: `chart-ratchet: BenchLex <n> ns/op <n> allocs/op` (likewise
  `BenchRecognize`, `BenchExpected`, `BenchForest`). They are not keys and
  nothing compares them. On a file the chart refuses, `BenchForest` (and so
  `measureBench`) is an error, not a figure: there is no forest to time
  (`TestMeasureBenchNeedsAcceptance`). Run by hand there,
  `BenchmarkRecognize` and `BenchmarkExpected` time only the prefix the
  grammar reaches.
- **Mainline.** `make test` never runs the ratchet. Its CI job,
  `.github/workflows/chart-ratchet.yml`, runs on changes under
  `internal/chart/**` and `internal/lexer/**` and is advisory
  (`continue-on-error: true`). It errors (exit 2) on every run, since CI has
  no tdt fixtures, so it observes nothing; T1 is in the tree.

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
- **The tdt keys read `TDT_FIXTURES`**, the directory `offline/tdt-align.sh`
  writes the aligned pairs into (they are not committed). Unset, it is
  `/var/tmp/pvm-chart-tdt`, where this milestone generates them; naming no directory, or a corpus file with no fixture there, is an
  error (exit 2). A pair is a token class and the terminal it declares, written
  `<terminal> <key>` as in `testdata/tdt/unseen-reviewed.txt`, where each entry
  sits under a `reason:` line saying why toke.c never produces it over the
  corpora; an entry naming no pair of the table is an error. Only whole
  captures are counted. On stderr the ratchet names each mismatched token,
  each unlisted unseen pair and each unaligned token outside a family.
- **The tdt comparison's fits and skips** (perigrin's TDTREV rulings). A
  fitted terminal stands for the toke.c pair its family folds, keyed by
  family: PRIVATEREF for a call's `&` and name (noamp) and for a constant
  sub's `NAME()` BAREWORD (constant-sub), and KW_SUB_named for the
  KW_SUB_named or KW_SUB_named_sig a statement word's phaser family folds
  (the `_sig` half because the chart does not track signatures; perigrin:
  "Both stand"). Two lists skip single tokens, from the mismatch check and
  from the unaligned count alike (perigrin: "Both stand"):
  `testdata/tdt/lexer-rooted.txt`
  (`<path> <byte> <issue ref>`, a token whose lexer root is filed, plus the
  `}` closing a listed `{`) and `testdata/tdt/overrides.txt`
  (`<path> <byte> <reason>`, a builtin the file overrides through
  CORE::GLOBAL, `use subs` or an import, which toke.c resolves as the sub).
  **There are two lexer lists because they answer different questions:**
  `lexer-defects.txt` drops whole files from the superset population,
  where `lexer-rooted.txt` skips one token in the width check, so a file the
  chart recognizes keeps its place in the population. A listed byte where no
  token starts is an error, and so is a reviewed tdt-unseen entry for a pair
  that is seen, whose reason no longer holds. A reason saying "no corpus
  token matches the key" is held to a scan of every corpus file with the key
  compiled alone, each file tokenized once and mapped only for the claims it
  could hold a token of (`TestTDTAbsentReasonsHold`, about 5 s;
  `TestTDTAbsentReasonsMatchPerClaimScan` holds it to the unfiltered scan).
- **The verdict keys read perl's committed labels** (`testdata/labels/`,
  reviewed categories); a missing label file, or a corpus file with no label,
  is an error (exit 2).
- A rule exists only for a key the ratchet prints.

| Key | Meaning | Rule |
|---|---|---|
| `tables-drift` | committed tables differing from what `make chart-tables` writes | may not rise; 0 at the exit |
| `leo-disagree` | files where Leo on and off differ in acceptance or `Expected` at any position, over every corpus file | may not rise; 0 at the exit |
| `ledger-gaps` | reference alternatives with no ledger mark, plus grammar citations not in the ledger | may not rise; 0 at the exit |
| `rd-accepted-chart-refused` | superset-population files the chart refuses | may not rise; 0 at the exit |
| `chart-accepted-both-refused` | perl grammar refusals recursive descent also refuses that the chart accepts, less both lexer lists | may not rise; 0 at the exit |
| `lex-unmapped` | files with a token the chart's table leaves unmapped | may not rise; known list from completion |
| `tdt-aligned` | tokens aligned with perl's `-DT` | may not fall |
| `tdt-mismatch` | aligned tokens whose terminal set misses perl's | 0 from completion |
| `tdt-unseen` | table pairs perl's `-DT` never showed | may not rise from completion |
| `terminals-per-token` | the table's width: terminals per mapped token (a token the table cannot map holds none and is not counted) | may not rise from completion |
| `tdt-unseen-unlisted` | `tdt-unseen` pairs not on the reviewed list | 0 from completion |
| `tdt-unaligned-unlisted` | unaligned tokens outside a reviewed fitting family | 0 from completion |
| `lex-overhead` | chart/lex time over `lexer.Tokenize` time, long-lived, on the two psc latency buffers: the larger of the `std5k.pl` and `edit5k.pl` figures, each the median of three, each pass lexing its buffer 25 times | timing; ≤ 1.3 absolutely |
| `lex-overhead-std5k` | the same on `std5k.pl` alone | timing |
| `lex-overhead-aggregate` | the same over the aggregate, never gated | timing |
| `speed-ratio-std5k-longlived` | chart over recursive descent on `std5k.pl`, long-lived (§9.3) | timing |
| `speed-ratio-std5k-cold` | the same, cold | timing |
| `speed-ratio-aggregate-longlived` | the same over the aggregate, long-lived | timing |
| `speed-ratio-aggregate-cold` | the same, cold | timing |
| `alloc-per-token-chart` | the chart's heap allocations (`runtime.MemStats.Mallocs`) per lexer token, cold, over the aggregate | timing |
| `alloc-per-token-rd` | recursive descent's, the same way and over the same tokens | timing |
| `items-per-token` | Earley items per token, over files the chart accepts (§7's cost on recognized input) | reported |
| `set-reuse` | share of positions whose Earley set was already interned, over files the chart accepts, each recognized by a fresh recognizer (cold), so only reuse within a file counts | reported |
| `ledger-excluded` | ledger alternatives marked excluded (reasons, a closed list in `grammar/ledger.go`: `keyword-plugin`, `refused-in-action`, `conditional-refusal`, `error-recovery`, `interpolation-only`, `folded-into-token`) | reported |
| `superset-population` | files both perl and recursive descent accept, less both lexer lists | reported |
| `superset-recognized` | share of that population the chart accepts | reported |
| `rd-accepted-perl-refused` | files recursive descent accepts and perl refuses | reported |
| `chart-accepted-perl-refused` | files the chart accepts and perl refuses | reported |

### The speed protocol

Every timing ratio — `lex-overhead*` and `speed-ratio-*` — comes from one
protocol in `cmd/ratchet/speed.go` (spec §9.3), code over injected sides, a
clock and a collector, so it is tested on fake timings
(`TestSpeedProtocol`, `TestSpeedProtocolExclusionBoundary`,
`TestSpeedProtocolRefusesUnknownTimedPath`, `TestSpeedLifetimes`):

- **Alternation.** The chart side and the reference side (recursive descent,
  or `lexer.Tokenize` for `lex-overhead`) are timed in alternating passes in
  one process: chart, reference, chart, reference. Each pass covers every
  file of the set once. A GC is forced before each pass, outside the timed
  span; within a pass collection is left to the runtime.
- **A measurement** is 2 warm-up pairs then 9 timed pairs; its ratio is the
  median of the 9 per-pair ratios (chart pass over reference pass), with min
  and max printed on stderr. A figure is the median of three measurements.
- **Lifetimes.** Long-lived: one parser per pass, reused across its files.
  Cold: a fresh recognizer per file. chart/lex's pooled lookup cache
  (`grammar.LookupCache`, from a `sync.Pool`) persists from file to file
  within a pass, cold or not; the GCs forced before each pass empty the pool,
  so no pass replays a cache warmed by an earlier pass. A `lex-overhead`
  pass lexes its one buffer 25 times (below), so within a pass a copy can
  meet a cache warmed by the copy before it, until a GC inside the span
  drops it: an LSP re-lexing one buffer in a warm process meets the same.
- **The file set is an input.** `lex-overhead` passes `std5k.pl` and
  `edit5k.pl` each alone, long-lived, and takes the larger figure (the edit5k
  figure is a `lex-overhead-edit5k` measurement line on stderr, not a key);
  `lex-overhead-aggregate` passes the aggregate; the speed ratios pass
  `std5k.pl` alone and the aggregate. A `lex-overhead` pass lexes its buffer
  25 times (`lexRepeat`), about 200 ms of work rather than 8 to 10 ms: perigrin
  ruled it on 2026-11-10, the definition's fourth amendment, after a pass of
  one buffer read 1.336 under load on code that reads about 1.15 quiet. On a
  loaded machine the measuring thread waits for a CPU for most of a short
  pass, so a one-buffer pair's wall ratio was decided by the scheduler, not
  the work (per-pair ratios 0.09 to 8.5 at load 15 to 21; GC was not the
  cause: every span held one collection on each side, and `GOGC=off` left the
  spread unchanged). The 9 pairs, the medians, the wall clock and the GC
  placement are unchanged (`TestLexOverheadDefinitionUnchanged`,
  `TestLexOverheadProductionClock`). `TestLexOverheadStableUnderLoad` shows
  why on a fake clock: delays landing independently per unit of work average
  out over 25 copies. Real load did not behave so: under heavier load (1-minute
  load 10 to 14 rising to 26 and 27) the 25-copy pass still read 1.306 and
  1.553, which is why lex-overhead is now measured only on a quiet machine
  (next point). The 1.3 protects chart/lex's share of an
  LSP call, which re-lexes one buffer in a warm process; over many small
  files each call starts with cold CPU caches, so the aggregate measures
  per-call start-up and is reported, never gated.
- **A quiet machine only** (perigrin, 2026-11-11). Just before lex-overhead
  is measured, chart-ratchet reads the 1-minute load average once from
  `/proc/loadavg`; above the CPU count (`runtime.NumCPU()`, 4 on the machine
  the readings came from) it measures nothing and exits 2 with
  `chart-ratchet: lex-overhead: machine too busy to measure: 1-minute load
  <load> above <limit>`, printing no keys, so it never passes; a close waits
  for a quiet machine. The limit is the CPU count because above it runnable
  threads queue for a CPU, the cause LEXLOAD established. The evidence for
  it is thin and is stated by pass length (loads are 1-minute averages when
  the run started). On the 25-copy pass: one ratchet started at 3.76 (3.8 to
  4.9 while lex-overhead ran) read 1.179, per-pair ratios 0.94 to 1.56; the
  failures, beside a busy loop per CPU and a second ratchet, started
  at 14.46 (1.306) and 10.44 (1.553); one ratchet alone started at 22.48
  read 1.160, so a busy machine can also pass by luck. On the one-buffer
  pass, before the amendment: a ratchet started at 3.88 (rising to 8.24)
  read 1.169, and the reproduction started at 6.37 (rising to 22.34) read
  1.336. The limit sits at the highest load the 25-copy pass was seen
  reliable at and below the lowest load any failure started from. The
  ratchet's own thread counts toward the load it reads. An unreadable load
  is an error, never a pass; off Linux there is no `/proc/loadavg`, so there the ratchet fails
  (a shortcut, pinned in `TestLoadAvg`). Only chart-ratchet reads the load:
  `make test` tests the gate on an injected load
  (`TestLexOverheadRefusesBusyMachine`), and `TestLexOverheadProductionLoad`
  pins the production path and limit.
- **What the speed ratios run.** The chart side is `chart/lex` then
  `recognize.Recognize` on one recognizer: long-lived reuses it across a
  pass's files (its intern table and cache persist), cold makes a fresh
  one per file. Recursive descent is `parse.Parse`, which holds no state
  between files, so its two lifetimes coincide. On a file the chart refuses,
  every position after the refusal holds no items and costs little, so
  until the chart recognizes the population (`superset-recognized`) a ratio
  is a lower bound on the finished recognizer's, reported for warning (§9.3
  checkpoints). On `std5k.pl` alone the long-lived figure is effectively the
  cold one (quiet readings at 52b8900a: long-lived 1.2207 and 1.1476 against
  cold 1.1479 and 1.1565; the baseline's single run is noisier than this),
  because one file is one pass and the recognizer has nothing to carry from
  file to file. chart/lex's pooled
  `grammar.LookupCache` does persist within a pass. A `sync.Pool` keeps its
  objects through one GC and drops them at the second, and two GCs are forced
  between consecutive chart passes (one before the reference pass between
  them), so each chart pass starts with the pool empty. Allocations are one
  cold pass after a GC, counted in `Mallocs`, over the lexer's tokens on the
  aggregate (`lexer.Tokenize`'s count, the same denominator for both sides).
- **The aggregate** is the timed set less `std5k.pl`, `edit5k.pl` and any
  file whose recursive-descent time alone is more than 10% of recursive
  descent's time over the **whole** population, whatever subset is timed.
  Exactly 10% stays in. Recursive descent is passed to `Aggregate`
  explicitly, even when the reference side is `lexer.Tokenize`, so every
  figure reads one aggregate. A timed file outside the population is
  refused, never kept unchecked. Each excluded file and each corpus's share
  of the aggregate's recursive-descent time are printed on stderr.

  The 10% rule excludes one file, `T1/code_too_large.t`: 10,001 lines of flat
  top-level statements, 34.3% to 36.1% of recursive descent's population time,
  read from two `make -s chart-ratchet` runs at 03e4097b: 36.1% on the first
  (load not recorded) and 34.3% on the second, at load average 8.15 (one
  minute). Other runs have read 37.6% to 40.1%. The share moves with load, and
  so does which files the rule excludes: a gate run at load 16 to 20 read 17.7%
  and also excluded `T1/require_module_return.t` (26.3%). Re-read it rather than
  trust it. No key prints the chart's ratio on it, so it is
  recorded here, observed at c6de5c76 (before) and on the branch of
  01a1b879-ae82 (after), on a 4-core machine at load average 1.2 to 1.3.
  Each figure is chart/lex plus a fresh recognizer with default `Options`
  against `parse.Parse`, 11 pairs with a GC before each side, the
  median of the last 9; recognition alone is the median of 9 of 11 runs.

  | input | lines | chart/RD before | chart/RD after | recognition before | after | with `NoCache: true` |
  |---|---|---|---|---|---|---|
  | `code_too_large.t`, first lines | 1,251 | 3.07 | 2.86 | 27 ms | 24 ms | 26 ms |
  | `code_too_large.t`, first lines | 2,501 | 3.71 | 2.94 | 71 ms | 49 ms | 56 ms |
  | `code_too_large.t`, first lines | 5,001 | 4.92 | 2.35 | 268 ms | 98 ms | 98 ms |
  | `code_too_large.t`, whole | 10,001 | 10.10 | 2.95 | 1,231 ms | 212 ms | 212 ms |
  | `std5k.pl` | 5,019 | 1.03 | 1.13 | 9.5 ms | 10.1 ms | 46 ms |
  | `edit5k.pl` | 5,054 | 1.23 | 1.25 | 10.9 ms | 9.6 ms | 41 ms |

  The before column was quadratic: the transition cache scanned a per-key
  list that grew by one entry per statement (see the recognizer's transition
  cache). After, the cache costs no more than no cache on this file and
  leaves `std5k.pl` and `edit5k.pl` where they were (their ratios move
  within the run-to-run spread; recognition alone on `std5k.pl` was 10.7 ms
  before and 10.4 ms after, the medians of five alternating benchmark runs).
  The whole-file ratio is still near 3, but no longer grows with the file:
  recognition time doubles from 5,001 lines to 10,001.

### The lexer lists

Two lists name the files a lexer defect takes out of the gates (§11):
`testdata/lex-unmapped-known.txt`, for a token the chart's table leaves
unmapped, and `testdata/lexer-defects.txt`, for a token that is mapped but
wrong. Refusal counts exclude the files on either list and no others. Both
started empty; an entry is added only by the issue that meets the defect, filed
against the v0.1 milestone, and the gate review confirms the construct at the
entry's byte is that issue's.

**The chart refuses some `return {…}` and `bless {…}, $class`.** The lexer
defect 01a1ad1d-bc7b guesses the open brace's type there, and where it guesses
wrong the chart's table leaves the brace unmapped; `return { "a", 1 }` and
`bless { a => 1 }, $c` happen to be guessed right and are accepted. Observed
at aa5aed2f, counting `lex-unmapped-known.txt` against the committed perl labels
(`testdata/labels/`): the defect accounts for 105 of its 114 entries, and the
labels mark 91 of those 105 accepted (and 92 of all 114). The 91 are 90 T1
files and one in-tree file (`parses/…Value-Literal-063.pl`); at their
unmapped brace, 89 read `bless {` (one inside `Storable::dclone(…)`) and 2
`return {`. The other 14 are files perl refuses: 12 at `bless {` and the two
`sprintf {…}` files under `parsent/`.
Recursive descent already refuses 33 of the 91, so the list takes 58 files
out of the superset population (§9.1 allows this), not 91.
`TestLexerListHidesNoChartGap` rewrites every brace the defect leaves
unmapped as `+{` in each of the 91 and holds the chart to accepting all of
them, and the count to 91, so the 01a1ad1d-bc7b entries hide no gap of the
chart's own. That is all it tests. The 92nd perl-accepted entry,
`T1/nested_heredoc.t 1821 01a1a846-5aea`, is a heredoc lexer defect, and
nothing checks that the chart would accept that file lexed right.

Each non-blank line is one entry, `<path> <byte> <issue ref>`:

- **path** names a corpus file. An in-tree file is repository-relative:
  `internal/chart/testdata/std5k.pl`, `internal/psc/testdata/latency/edit5k.pl`,
  or a `*.pl` directly under `internal/chart/testdata/guacamole/parses/` or
  `.../parsent/`. A T1 file is `T1/<name>.t`, a top-level `*.t` of T1,
  resolved as `<name>.t` under `testdata/perlonjava/unit`, or under
  `$T1_CORPUS` when that is set. A perl.git file is `perlgit/<path under t/>`,
  a `.t` file of the pinned perl5 tree's `t/` (`$PERL5_CORPUS`, by default
  `/opt/corpora/perl5`). `lexer-defects.txt` holds 13 such entries and
  `tdt/lexer-rooted.txt` 193; `lex-unmapped-known.txt` holds none yet, since
  the ratchet does not measure perl.git.
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
checks the issues are open. T1 entries resolve
against the in-tree copy, so the test does not skip. `T1_CORPUS` set but
naming no directory fails the test. perl.git entries are checked only where
the perl5 tree is present: without it the test leaves them unchecked rather
than failing.

## The grammar: `grammar/perly.go`

`grammar/perly.go` holds `perly.y`'s rules as data, one rule per line in
`perly.y` order, each citing the alternative it implements:
`lhs: rhs # perly.y:N`, or `# perly.y:N, fitted: <why>` where the rule takes
what `internal/lexer` emits in place of what toke.c does. The committed
tables hold 349 rules over 112 nonterminals and 124 terminals.

- **The reference list**, `grammar/testdata/perly/alternatives.txt`, is
  bison's report of `perly.y` at 94e50866: 348 alternatives over 114
  nonterminals. `offline/perly-alternatives.sh` regenerates it byte for byte.
- **The ledger**, `grammar/testdata/perly/ledger.txt`, marks every
  alternative: 289 ported, 37 fitted, 18 excluded and 4
  refused-by-precedence. An exclusion's reason is one of the closed list in
  `grammar/ledger.go`; today 5 `refused-in-action`, 8 `folded-into-token`,
  2 `keyword-plugin`, 1 `conditional-refusal`, 1 `error-recovery` and 1
  `interpolation-only`. `TestPortLedger` holds the citations and the ledger
  equal, `TestLedgerExclusionReasons` the reasons to the list, and
  `TestUnfittedRulesMatchPerlyY` every rule not marked `fitted:` to the text
  of the alternative it cites, symbol for symbol, less bison's mid-rule
  actions and `%empty`.
- **Refusals.** The five alternatives whose actions refuse unconditionally
  (`perly.y:1105`, `1109`, `1541`, `1546`, `1553`, on
  `grammar/testdata/perly/refused-in-action.txt`) have no rule, and the four
  comparison chains (`1466`, `1468`, `1484`, `1486`) are
  `refused-by-precedence`. `TestLedgerRefusals` holds, for each,
  `testdata/refusals/<line>/refused.pl`, which the chart refuses (a chain's
  dangling form, `$a <=> $b <=>`), beside `accepted.pl`, a nearby form the
  chart accepts; the pinned perl gives both verdicts.
  `testdata/port/refused-where.txt` holds where each port snippet under a
  `*-refused` directory is refused, one `<snippet> <table|grammar> <byte>`
  line each: `table` where chart/lex finds a token no class admits
  (`lex-unmapped`), `grammar` where the recognizer refuses.
  `TestRefusedSnippetsRefusedWhere` fails on a snippet
  that moves to the other cause or byte until the list is reviewed.
- **Start forms.** All seven of `perly.y`'s start forms (`perly.y:187-268`)
  are start rules. The program form is the one a file is recognized by, and
  the only one graded. Each of the other six begins with its `GRAM*`
  selector, a token perl's parser pushes and the lexer never emits, so
  `TestStartForms` puts the selector before the lexed tokens itself, and tests
  each form only on input that lexes the same from statement start:
  `internal/lexer` exports only statement-start states.
- **Token classes** are declared beside the rules by terminal name, one slot
  per lexed token `perly.y` declares (`%token`, `%left`, `%right`,
  `%nonassoc`), in the order it first declares it; the `GRAM*` selectors,
  `PLUGEXPR`, `PLUGSTMT`, the `PLUGIN_*` terminals, `PREC_LOW` and `UMINUS`
  have none. They are compiled into the token table
  (`grammar/tokens.go`, spec §7). toke.c is the authority for which terminal
  a token is, cited by line where a class is not plain. Where pvm's key cannot see what toke.c reads,
  perigrin has ruled extensions of §7, each named in the section that uses
  it: the neighbour flags `Glob`, `Name`, `Label`, `SortParen` and `Array`;
  the token's own `Global`, `LineColon` and `Blanks`; the prefixes for a
  version, a package name and `CORE::`; and the fitted terminals `CAST_*`
  and `amper: PRIVATEREF`.
- **Precedence** is declared and ignored by recognition (below).

Each family's section below gives its fits and its known limits. A limit
carries a `shortcut:` comment in the code and, where a test can reach it, a
pinned test row that fails when the limit goes.

### Attribute lists

`internal/lexer` emits `my $x :Foo(arg) Bar` as a colon, words and a
parenthesised argument scanned as Perl; toke.c emits one COLONATTR and one
ATTRLIST holding every attribute and its argument as an opaque string
(perly.y:1070 is fitted to the lexer's tokens). Known limits:

- An attribute's argument is one token, never `(` or `)`. `:Foo(a, b)`,
  `:Foo(1+2)` and `:Foo(())` are refused, where perl accepts them, until the
  lexer scans an argument as one token
  (`01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a`). No corpus file meets this, so
  `lexer-defects.txt` has no entry for it.
- An attribute name is a plain identifier to toke.c. pvm emits a
  package-qualified word as one token, so `:Foo::Bar`, `:Foo::` and `:Foo'Bar`
  are accepted where perl refuses them ("Invalid separator character"), until
  the same lexer issue (`01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a`) lands.
- An attribute after one whose name some class names (e.g. `use`, `my`) is
  not matched: `:use shared`, `:my shared` and `:no Foo` are refused where
  perl accepts them. `grammar.OtherWord` stands for a word no class names, so
  the set of such names grows as classes name words.
- `state` maps to KW_MY with or without `use feature 'state'`, so
  `state $x :Foo(1) = 2;` is accepted where perl refuses it without the
  feature (`01a1a29c-cd50-79a8-b87a-e9615c52af65`).
- `when` maps to KW_WHEN with or without `use feature 'switch'`, so
  `$y = 1 when $x;` is accepted where perl refuses it without the feature
  (`01a1a29c-cd50-79a8-b87a-e9615c52af65`).
- `given` and `default` map to KW_GIVEN and KW_DEFAULT with or without
  `use feature 'switch'`, so `given ($x) { 1 }` and `given ($x) { default { 1 } }`
  are accepted where perl refuses them without the feature, as a bareword
  call (`01a1a29c-cd50-79a8-b87a-e9615c52af65`). `when` and `default` outside
  a topicalizer are accepted too, which is parity: under `use feature 'switch'`
  the pinned perl gives syntax OK for `when (1) { 1 }` and `default { 1 }` and
  fails only at run time, so there is no scope refusal to make.
- `for our (…)` and `for state (…)` at a statement start are refused, as toke.c's
  `yyl_foreach` croaks "Missing $ on loop variable" there (only `my` may open a
  parenthesised list); `1 for our ($x);` stays accepted. The statement start is
  read as the token two back from `our`: a file's start, `;`, a block's `{`, a
  block's `}`, a label's `:`. The label's colon is told from an attribute list's by
  a neighbour flag (`Label`, on a label's word and on its colon), so
  `L: for our ($x) (1) {}` is refused as perl croaks, while
  `my $y : for our ($x);` stays accepted.
  One leftover: a `do`, `eval` or anonymous-sub block's `}` reads as a statement
  start, so `do {1} for our ($x);` is refused where perl accepts it, because the
  lexer does not mark what a `}` closes (`01a1a422-9176-7153-8625-6535b483a4b4`).
  The croak is not a syntax error, so these forms are inline rows in
  `TestPortLoopsRefusals`, not snippets.
- `my $x :if;` is accepted, where toke.c refuses the keyword as an attribute.
- A `:` after `my`, `our` or `state` and a variable is the attribute colon when
  an identifier character follows (toke.c's `yyl_colon`), so
  `$c ? my $x : shared;` is refused as perl refuses it. The key sees two
  tokens back, so two cases stay accepted where perl refuses them, until the
  lexer marks the `)` that closes a `my` list and the colon that separates
  attributes (`01a1a314-e3f8-765f-abf2-159425ed764e`):
  `$c ? my ($x) : shared;`, which looks like `$c ? (my $x) : shared;`, and
  `$c ? my $x : shared : 1;`, whose second colon looks like the one in
  `$a ? $b ? 1 : c : d;`. `TestPortTermsAttributeColon` pins both. A word opens
  an identifier only with a letter, `_` or a UTF-8 byte (under `use utf8`), so `$c ? my $x : ::foo;` is accepted
  as perl accepts it.
- Because `state` is a declarator with or without `use feature 'state'`
  (`01a1a29c-cd50-79a8-b87a-e9615c52af65`), `$c ? state $x : shared;` is refused
  without the feature, where perl reads `state` as a word and accepts it.
- A quote-like neighbour's class also records whether it opens with a letter
  (`qw`, `q`, `qq`, `m`, `s`, `tr`, `y`), and an unnamed word's whether it opens
  with a letter, `_` or a UTF-8 byte. This extends spec §7's neighbour class, which names the
  text only of words, operators and closing brackets, so that a class can say
  "an identifier character follows". Lookup stays constant-time and
  allocation-free: the extra ids are fixed at compile time and `Intern` adds one
  branch.
- `tdt-unseen` counts seven ATTRLIST pairs (a word after a word or `)`, and a
  word, number, quote or variable after `(`) that no `-DT` fixture shows, an
  artifact of the fitting; they are reviewed at completion.

The string-bitwise operators are feature-blind. `internal/lexer` emits `|.`,
`&.`, `^.` and `|.=` as operators with or without `use feature 'bitwise'`
(toke.c lexes them only under it), so `$y = $a |. $b;` is accepted where perl
refuses it without the feature. Known limit, until the lexer tracks the
feature (`01a1a24e-f0aa-7a97-90d8-e4f86efa1e3b`).

`try`, `catch`, `finally` and `defer` map to their keywords with or without
`use feature 'try'` or `'defer'` (toke.c reads them as words without it), so
`try { 1 } catch ($e) { 2 }` is accepted where perl refuses it without the
feature (`01a1a29c-cd50-79a8-b87a-e9615c52af65`, which covers these words too).
The other side is under-acceptance: where perl reads them as plain words,
`sub try {} try 1;` and `catch(1)` are refused (`try => 1;` and `$h{try}` are
accepted, as every word is there).

A LABEL is a word with no package separator and a one-byte colon after it on
its own line (blanks only between; the word's `LineColon`), where a statement may start: toke.c's label
scan reads a line at a time and scans the word without `::`. `L\n: 1;`,
`L #c\n: 1;` and `L::M: 1;` are refused.

`catch (` takes a plain lexical name: `catch ($::e)`, `catch ($Foo::e)`,
`catch ($_)`, `catch ($1)`, `catch ($^W)` and `catch ($')` are refused, as the
action declares the variable with `my`, and `catch ($e)`, `catch ($a)` and
`catch ($ARGV)` are accepted. The refusal is a PERLY_DOLLAR class excluded two
tokens after the word `catch` (a `Global` part: no plain identifier, or `$_`).
shortcut: that class also refuses `catch =>` followed by a variable that is no
plain name, in any list (`catch => $_;`, `foo(catch => $_);`), which perl
accepts (`catch` is a hash key). `$p->catch($Foo::x);` is accepted because a word
after `->` is a name (`Neighbour.Name`), which the exclusion's `catch` word does
not match. The exclusion sees only the word two tokens back, not the token
between, so the upgrade is to let it see `=>` as the variable's predecessor
(Prev1); it is pinned in `TestPortScopeDecidedKnownLimits`.

### Version tokens after `package`, `use` and `no`

toke.c forces a version into a BAREWORD after `package NAME` and after `use`
or `no` (and an empty one when none is written); `grammar/versions.go` gives
pvm's Number or v-string that class. The key drops a number's text and what
follows it, so chart/lex derives both into the Number's `Prefix`
(`lex.versionShape`): `grammar.VersionStrict` for what `force_strict_version`
reads (`1`, `1.23`, not `01`, `1_0`, `1.`, `0x10`, `1e3`), `grammar.VersionLax`
for what `force_version` reads (digits, `_` and `.`), and neither unless
whitespace, `;`, `{`, `}` or the end of the file follows, as `lib 1.2, 3` is a
term to toke.c. Known limits:

- `package Foo v1.2.3;` and `package Foo v1.4.0 {1}` are refused: after a
  package name the lexer does not scan a v-string (`scanVString` wants a term),
  so it emits `v1`, `.`, `2.3`. The lexer issue is
  `01a1a481-ebef-7f2d-99bd-fab0cc8aaf7c`; the two Guacamole files that hit it
  are on `testdata/lexer-defects.txt`.
- `use VERSION MODULE` (`use 5.036 strict;`) is refused. toke.c lexes that
  module as an ordinary term, so taking it would take `use 5.036 foo bar;` too.
- The version BAREWORD is also the term BAREWORD, and no terminal tells them
  apart. After a version that follows a module, `TokenClass.NotNext` (a list)
  refuses the comma and the operators that only join two terms
  (`=> . x ? == != || // && .. ... =~ !~ ** <=> eq ne lt gt le ge cmp = += -> | ^ >> > >=`
  and the other assignment operators, not `**=` and `//=`). `- + * % & < / <<`
  and `**=`, `//=` can open a term (a glob, a match), so they are accepted, as
  perl accepts them; `<=` is a readline's `<` to toke.c, so it stays accepted
  too. `use lib 1.2 x;`, where `x` is a word, is refused, and `isa` is a plain
  word there to perl, so `use lib 1.2 isa 3;` is accepted where perl refuses
  `isa 3`, as on the base.
- A bare word after `use MODULE` is a term (`use constant FOO => 1`,
  `use lib foo bar`, `use lib foo %h`), and refused when what follows is
  something toke.c refuses after an unknown bareword: a number, a quote, `[`,
  `\`, `!`, `~`, an array or `$#` variable (the neighbour's `Array` flag), a
  `::` word and any keyword that is not an operator word, `not` included.
  `use lib foo x;`, `use lib foo x, 1;` and `use lib x x;` are accepted where perl
  refuses them, since refusing `x` after a word would also refuse `use lib foo x 3;`,
  which perl accepts (pinned in `TestPortPackageAndUseRefusals`). `use lib foo { }` and
  `use lib foo $$x` are accepted, as perl does (pinned beside the other
  `use lib` rows), and `use lib foo __END__` is refused where perl accepts it,
  for want of the indirect-object family.
- `Neighbour.Array` is on a variable opening with `@` or `$#`, so a word's key
  can tell `foo @x` from `foo $x`. Like `Glob`, `Name` and `Label` it is a
  neighbour flag decided from the token's own text, an extension of spec section
  7's neighbour class by perigrin's ruling; INTERN's reference derives it
  independently (`refClassOf`) and `keyText` renders it with the neighbour.
- A delimited number two after `use` or `no` is a version and no term, unless
  `=>` stands between, after which toke.c reads a term (`(no => 1 , 2)`).
- The Number version shape extends spec section 7's key with a part it did not
  have: the number's own text and the byte after it, derived by chart/lex
  into `Prefix`. perigrin ruled it so, as he did for `Glob` and `Name`.

### The repetition operator `x`

toke.c's `KEY_x` makes `x` the repetition operator exactly when `PL_expect`
is XOPERATOR, which a token that completes a term leaves; otherwise it is a
word. A token class marked `Completes` is such a token, and `init` in
`grammar/perly.go` derives MULOP's `x` classes from the marks. As a word `x`
is `BAREWORD`, `PRIVATEREF` before `(`, and after each statement-start token
(`stateStarts`: a file's start, `;`, a block's `{` or `}`, a label's colon)
every scope-decided reading a plain word gets there (`addStatementX`:
`FUNC0SUB`, `UNIOPSUB`, `LSTOPSUB`, `NOAMP`, `METHCALL0`, `METHCALL`), with the
`{` after it a block as after any such word, so `x {1};`, `x { a => 1 } 1;` and
`sub x {1} x 1, 2;` are accepted as perl accepts them. A block's `}` includes
the `}` of `map`, `grep` and `sort`, so `@a = sort { 1 } x 1;` and
`@a = map { 1 } x 3;` are accepted too, as perl accepts them once `sub x;` is
declared. After a term that is not a block the set is unchanged
(`TestXOperatorAfterTermUnchanged`). After the `}` of a `do`, `eval` or
anonymous sub, where perl expects an operator, the readings follow as well,
by the shortcut that reads such a `}` as a statement start
(`01a1a422-9176-7153-8625-6535b483a4b4`): `$y = do { 1 } x 3;` gives `x`
`BAREWORD FUNC0SUB MULOP NOAMP UNIOPSUB`. No verdict changes, because the
grammar never sets two terms side by side; the set is pinned in
`TestXAfterBlockBraceKnownLimit`. The widening raised
`tdt-unseen` from 860 to 915: no corpus file starts a statement with a call to
a sub named `x`, so perl's `-DT` never shows the 55 new pairs, each reviewed in
`testdata/tdt/unseen-reviewed.txt`. `terminals-per-token` stayed 1.213, as no
corpus token is `x` after a statement-start token. Known limits:

- A block's `}` is not treated as completing a term for the `*` that follows
  it (`if (1) { } *print x 3;` is accepted, as perl does), because the
  neighbour cannot tell a statement block's `}` from a `do` block's. So
  `$y = do { 1 } * ref x 3;` is accepted, where perl refuses it, until the
  lexer marks what a `}` closes (`01a1a422-9176-7153-8625-6535b483a4b4`).
  The row is pinned in `TestPortListOperatorsXRule`.
- A class that ends a term must be marked `Completes`, or `x` after it is
  read as a word: `]`, the `}` closing a subscript or an anonymous hash, a
  quote-like (`qw`), a variable, a number and a string are marked; a family
  that adds one (`PMFUNC`, `PRIVATEREF`) marks it too.
- A bareword after `sort` (or `sort(`) is the sort sub's name and does not
  complete a term, so `sort foo x 3;` is refused as perl does. After the other
  list operators that take a filehandle the bareword is one only when no sub
  of that name is declared, so `print FH x 3;` is scope's to decide and is
  accepted (see Indirect objects).
- The `{` after a list operator whose `PL_expect` is XTERM is an anonymous
  hash to toke.c, whatever the lexer decided, so `sprintf {$fh} $x;` is
  refused as perl does; the `{` of a sub named like such an operator
  (`sub open {`, `sub join {`) is a block again through a class conditioned on
  `sub` two tokens back.

### Subs and phasers

**Citing a rule inside a fitted alternative.** The ledger marks an
alternative, not a rule. An alternative with one rule as written and another
fitted (638) is marked `fitted`, and its as-written rule is cited
`perly.y:N, fitted: as written; ...`, naming the rule that is the fit.

A named sub's attributes are the same ATTRLIST fit, with the same limits
(`01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a`). `internal/lexer` emits the argument of
`:prototype(...)` as one Prototype token, which the fit takes as an argument.
Beyond them:

- `sub foo :lvalue ($$) { 1 }` is accepted where perl refuses it: the tokens
  are the same with or without a space before the parenthesis, so the chart
  cannot tell an argument from a stray list. Until the lexer scans an argument
  only where the parenthesis touches the name (`01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a`).
- The `{` that ends a sub head after an anonymous sub's prototype
  (`sub ($$) { ... }`, `sub () { ... }`) or after an attribute colon with no
  attribute (`sub foo : { ... }`, `sub foo :lvalue : { ... }`) is not marked as
  a block by the lexer, so those forms are refused where perl accepts them
  (`01a1a445-61f0-73c8-81db-1cc65a3620f5`). Five T1 files hit the first and are
  on `lexer-defects.txt`.
- The lexer scans a prototype only after `sub NAME`, so the `(` after `BEGIN`,
  `END` and their kin is an ordinary parenthesis and `BEGIN ($$) { 1 }` is
  refused where perl accepts it (`01a1a44e-8f26-706b-96d4-b3e6d7f55c3f`). No corpus file meets this.
- `sub` is KW_SUB_named when a name follows it (an identifier, or a word
  opening with `::` or `'`, as toke.c's yyl_sub reads it) and KW_SUB_anon when
  none does. BEGIN, END, CHECK, UNITCHECK, INIT, AUTOLOAD and DESTROY are
  KW_SUB_named when no name follows, and rule 638 is fitted to take them with
  no name. `TokenClass.NextName` and `NotNextName` test this through a derived
  table part (one bit per next class), not a list of classes.

### Signatures

toke.c makes `sub` KW_SUB_named_sig (or KW_SUB_anon_sig) under the signatures feature and KW_SUB_named
otherwise (`yyl_sub`, `toke.c:5965-5975`, `is_sigsub`), and scans a prototype only without it. The chart does not
track the feature, so each `sub` is both terminals of its pair, and the split is the lexer's tokens: a Prototype
token after the name is the traditional rule's, parentheses holding parameters are the signature's.
`sub f { 1 }` and `sub f;` are therefore parsed both ways. The feature line itself is an ordinary `use` statement.

`internal/lexer` emits a parameter's sigil and name as one Variable token (`$x`, `@r`, `%h`) and a bare placeholder
as the sigil alone (`$`, `@`, `%`), which is the sigil's terminal either way (as `scalar: PERLY_DOLLAR` is fitted at
perly.y:1846). So `sigvar: PRIVATEREF` (1089) is `excluded folded-into-token`, and `sigvar: %empty` (1087) is the
only sigvar the chart ever uses. A parameter's default is a `term`, as written. Rules 1105 and 1109 (a slurpy
parameter with a default) and 1546 (an anonymous sub with no body) refuse in their actions and have no rule:
`TestPortSignaturesRefusals` pins them. A signature after the attributes is the grammar's order, and one before them
(`sub f ($x) : lvalue { 1 }`) is refused, as perl does.

Known limits, each pinned in `TestPortSignaturesKnownLimits`, accepted where perl refuses:

- A conditional refusal in an action is not ported: `sub f (@r, $y) { 1 }` (slurpy not last),
  `sub f (@r, @s) { 1 }` (two slurpies), `sub f (:$n, $x) { 1 }` (positional after named),
  `sub f ($x = 1, $y) { 1 }` (mandatory after optional) and `sub f ($x = ) { 1 }` (an empty default is legal for the
  placeholder `$=` only, and `$x` and `$` are one terminal).
- toke.c reads a signature in a state of its own (`yyl_sigvar`, `toke.c:5524`), which `internal/lexer` does not
  model (v0.1 `01a1a903-2d8d-713e-b264-bc8119a6a737`), so its tokens are the general ones: `sub f ($x += 1) { 1 }` lexes `Variable:$x Operator:+= Number:1`, where
  yyl_sigvar reads only `=`, `//=` and `||=` after a parameter; `sub f ($x::y) { 1 }` lexes one `Variable:$x::y`, where it
  reads a name with no package; and `sub f ($_) { 1 }` and `sub f (@_) { 1 }` are the globals, which perl refuses in a signature.
- Named parameters (`:$n`) take `optcolon` as written; `:@n` and `:%h` are refused, as perl does.
- toke.c makes BEGIN, END and their kin KW_SUB_named_sig under the feature (`toke.c:8315`), and the chart gives those
  words no signature alternative, so `END ($x) { 1 }` is refused where perl accepts it
  (`TestPortSignaturesPhaserLimit`; BEGIN dies at run time under `-c`, so END is the pinned row). The lexer scans
  neither a prototype nor a signature after those words (`01a1a44e-8f26-706b-96d4-b3e6d7f55c3f`). A fitted 608 without the name and seven more classes would take it, and would
  accept `BEGIN ($$) { 1 }` through the signature (that row in `TestPortSubsRefusals` would flip).

### Classes, methods and fields

`class NAME VERSION` is fitted as `package` is (perly.y:278, 294): pvm emits the name and then the version, which toke.c
forces first, so each rule has a form with and without the version BAREWORD, and a word two after `class` is never the
version (`afterWordSlot`). The name is a BAREWORD wherever it stands, keyword or not (`toke.c` KEY_class forces it with no
keyword check), as is a method's (`yyl_sub`). `method` is KW_METHOD_named or KW_METHOD_anon on the same split as `sub`
(a name follows or not), and `class`, `field` and `method` are keywords wherever they stand: the chart does not track the
feature. The mid-rule actions (`$@8`, `$@24`) are not ported. 1553 (`method` with no body) refuses in its action and has
no rule. Known limits, pinned in `TestPortClassesKnownLimits`:

- Refused where perl accepts: `class Foo v1.2.3;` (the package limit above, `01a1a481-ebef-7f2d-99bd-fab0cc8aaf7c`);
  `field $x :param //= 1;` (the lexer scans `//` after an attribute as a pattern, `01a1aa23-746a-7830-abd3-900a685fa6b0`);
  and `method m : { }` (the `sub foo : { }` limit above).
- Attribute arguments are the sub family's limit (`01a19fb5-79d7-76c5-a8b1-25fd04ddfb4a`). `class Bar { } class Foo :isa(Bar 1.2);`
  is refused where perl accepts it: pvm emits `( Word Number )` where toke.c delivers one ATTRLIST. And `method m :lvalue ($$) { 1 }`,
  `method m :lvalue (1) { 1 }` and `my $c = method :lvalue (1) { 1 };` are accepted where perl refuses them as bad signatures: pvm emits
  `( Variable:$$ )` and `( Number )` with or without a space before the parenthesis, so an argument and a stray list look alike.
- There is no refusal row for an attribute list after a class block: neither tokenizer marks that colon as an attribute colon.
- Accepted where perl refuses, each by an action or a scope rather than the grammar: `field` and `method` outside a class
  (waypoint 2's scope refusals), a field named `$::x`, `$_` or `$1`, and a class that inherits from itself.

### Formats

`internal/lexer` emits a format's picture lines, argument lines and closing
`.` as one FormatBody token (`01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b`); toke.c
emits FORMRBRACK, THING and FORMLBRACK … FORMRBRACK terminals. Rule 787 is
fitted to the token (FORMRBRACK stands for it), and 836, 837, 906, 924 and 925,
which build that sequence, are marked `excluded folded-into-token` (a reason
added to the closed list in `grammar/ledger.go`; `ledger-excluded` counts them).
Known limit:

- The argument lines are not parsed, so `format STDOUT =\n@<<<\n$x +\n.\n`
  (a dangling operator) is accepted where perl refuses it, until the lexer
  emits the body's tokens (`01a1a5b7-1bcb-7c4c-bbdc-1abedccd7e6b`). A pinned row
  in `TestPortFormatsRefusals` fails when the limit goes.

### Quote-likes and heredocs

`internal/lexer` emits a quote-like with its interior as one token (Quote,
HeredocOpen or Readline) where toke.c sublexes it, so rule 1328 is fitted as
`listop: FUNC` and rule 1699 as `term: PMFUNC`, each terminal standing for its
whole `SUBLEXSTART … SUBLEXEND` run; `optrepl` (1797, 1798), the substitution's
replacement inside the token, is `excluded folded-into-token`. The terminal
follows toke.c's `S_sublex_start`: THING for `'…'`, `q` and `<<'EOF'`; FUNC
for `qq`, `` `…` ``, `qx`, the other heredocs and a glob `<*.c>`; PMFUNC for
`m//`, `//`, `qr`, `s///`, `tr` and `y`; QWLIST for `qw`. A `"…"` is THING
when it holds no `$`, `@`, `\` or non-ASCII byte and FUNC otherwise, and a
readline `<FH>` THING where a glob is FUNC; the key cannot tell them apart, so
those tokens admit both. A quote-like's FUNC is always followed by toke.c's
SUBLEXSTART, never by `(`, so the FUNC classes exclude a `(` next (`NotNext`):
otherwise `"abc"(1)` would be read as a call (1326). Heredoc bodies are trivia
(see chart/lex below). Known limits:

- Interpolated code, `s///e` bodies and `qw` lists are inside their one token,
  so `$y = "@{[ 1 + ]}";` is accepted where perl refuses it. It counts in
  `chart-accepted-perl-refused`; a pinned row in `TestPortQuoteLikesRefusals`
  fails when the limit goes.
- A multi-line quote-like begun on a heredoc opener's line swallows the
  heredoc's body (`<<E1 . "a\nbody\nE1\nb"`), so the opener has no body and
  the file is refused where perl accepts it (`01a1a846-5aea`);
  `T1/nested_heredoc.t` is on `lexer-defects.txt`.

### Subscripts

The `;` perly.y spells before a subscript's `}` (`expr PERLY_SEMICOLON PERLY_BRACE_CLOSE`) is toke.c's artifact,
so the three brace rules (1366, 1369, 1373) are fitted without it, and a written `;` there (`$h{1;}`) is
refused, as perl does. A subscript's `{` is `PERLY_BRACE_OPEN` when the lexer did not open it as a block and it
follows a variable, `->`, or a closing `]` or `}` (the cases `ClassOf` calls `BraceSubscript`).

A word alone in a subscript's braces is a string (toke.c's `yyl_leftcurly`, `toke.c:6656-6679`):
`$h{time}`, `$h->{shift}`, `$h{-if}`. `+{ time }` is an anonymous hash and keeps its call reading. toke.c forces the
word only when it is a plain identifier (`scan_word(..., FALSE, ...)` reads no package part, so `Foo::bar`,
`Foo'bar`, `::foo` and `foo::` are not forced) and only spaces or tabs separate it from the `{`, from a minus (but see the filetest letters below)
before it, and from the `}`; a newline or a comment anywhere there leaves the word to the grammar, so
`$h{if\n}` is refused and `$h{time\n}` calls `time`, as perl does.

`TokenClass.Lone` states it: the Lone class names its surroundings (`Prev2`, `Prev1`, `Next` and `Blanks`) and
`CompileTokens` cuts every other class of a word wherever they all hold, as one copy of the class per part the Lone
class names with that part's value excluded, so `Lookup` does no more work. The blanks and the plain name are one
token bit, **`Blanks`**, a §7 extension by perigrin's ruling ("Add the blanks bit"): `chart/lex` sets it from the
gaps between raw tokens, in constant time and without allocating, only on a word a closing bracket follows (the
other words never pay for it), and the table tests it as key part 11. INTERN's reference reads the same bit off the
source bytes (`refBlanks`), and `TestBlanksAgreeWithReference` holds the two together.

One limit, pinned in `TestPortSubscriptsKnownLimits`. `$h{-e}` (and every filetest letter) reads as a filetest,
where toke.c forces the string, because `internal/lexer` emits `-e` as one token (v0.1 `01a1a805-a94e-7b39-8021-c61fe44b13ec`);
the minus rule above holds for every other word. `f()->{foo}` is accepted through the scope-decided family's
call (see "Scope-decided words").

### Indirect objects

`indirob` is ported as written: a bareword, a scalar, a block or a
PRIVATEREF, after a list operator (`print STDERR 1`, `print $fh 1`,
`print {$fh} 1`, `print(STDERR 1)`), a method word (`new Foo`,
`new Foo(1)`, `meth $obj 1`, `meth { $obj } 1`) or `&` (perly.y:1842's
as-written rule, so `&$code`, `&$code(1)` and `\&$code` are accepted).

- A word before a variable, a block or another word is METHCALL0 (toke.c's
  yyl_just_a_word and intuit_method); before another word it is METHCALL as
  well, since only the token after that word says which. A declared sub
  would be called instead, which scope decides, so the word stays a BAREWORD
  too.
- A `{` the lexer took for an anonymous hash after a word no class names is
  also a block: without a declared sub toke.c's METHCALL0 leaves PL_expect at
  XBLOCKTERM, so `foo { a => 1 };` is a method call with a block indirob, and
  with one the `{` is a hash, which scope decides.
- After `sort` a bareword is the sort sub's name whether or not a sub of that
  name is declared (toke.c:8158, "indir obj regardless for sort"), and
  PL_expect is XTERM after it, so `sort foo x 3;` is refused. The `(` of
  `sort(` carries the neighbour flag `SortParen` (spec §7's neighbour class,
  extended by perigrin's ruling as `Glob`, `Name` and `Label` are), since
  toke.c's yyl_leftparen lets the word after it be the name too, so
  `sort(foo x 3);` is refused. The `(` of an attribute named `sort`
  (`my $x :sort(1);`) carries the flag too, so its argument has ATTRLIST
  classes after the flagged `(`. `init` gives the `x` derived from a class its
  `NotPrev1` as `NotPrev2`.
- A sub's name after `&` is PRIVATEREF and BAREWORD (a lexical sub's or a
  package sub's, scope decides). A glob's name and a bareword filehandle are
  BAREWORD only.

Accepted where perl refuses them, because a declaration would make perl
accept them (scope, waypoint 2), each pinned in
`TestPortIndirectObjectsRefusals`:

- `print FH x 3;`, `print(FH x 3);`, `print x x 3;` and their kin after
  printf, say, open, binmode, exec and system: toke.c reads the word as a
  filehandle only when no sub of that name is declared, and
  `use constant FH => 1; print FH x 3;` is valid perl.
- `new Foo x 3;`: with no declared `new` intuit_method leaves PL_expect at
  XTERM after `Foo`.
- `foo { a => 1 } x 3;`: a block after a method, a hash after a declared sub.

`print $fh [];` is accepted: perl refuses it through
`indirob: scalar %prec PREC_LOW`, a precedence refusal (waypoint 2).

### Slices, glob elements, postfix dereference and sigil casts

The slices (1594-1624, `sliceme`, `kvslice`) and glob elements (1349, `gelem`) are ported as written, the brace
forms fitted without toke.c's `;` as the subscripts are. A `{` after a cast's sigil, a glob's `*` or a sub's `&`
opens a block (`*{"g"}`, `&{$c}(1)`, PREREF) unless its `}` follows at once, which toke.c reads as `HASHBRACK`, so
`${}` and `&{}` are refused (that `{` admits nothing, since no `-DT` capture of a file perl accepts holds one); these are classes keyed on the `{`'s neighbours, and `ClassOf` does not record them. A
`{` past the glob's name is a subscript (`*g{CODE}`, TERM): `ClassOf` calls the `{` after any name
`BraceSubscript`, so a lone word there is a string as in `$h{if}`.

**The fitted cast terminals**, an extension of spec §7 by perigrin's ruling ("Fitted cast terminal"): `CAST_DOLLAR`,
`CAST_SNAIL`, `CAST_PERCENT_SIGN` and `CAST_STAR` are the only terminals perly.y does not declare
(`grammar.CastTerminals`, held by `TestClassSlotsAreThePerlyYTokens`). pvm emits a whole variable (`$x`) as one token
the variable heads fit as `PERLY_DOLLAR`, and a cast's sigil (`$$x`, `${ … }`, `@{ … }`) as its own `DerefSigil`;
were the sigil `PERLY_DOLLAR` too, `scalar: PERLY_DOLLAR indirob` would read `$y = $x $z;`, `$x{1; 2};` and
`$x foo;` as casts. So a `DerefSigil` maps to the cast terminal, and `scalar`, `ary` and `hsh` gain
`CAST_* indirob` rules fitted to 1846-1856; `star: PERLY_STAR indirob` (1868) needs none, since pvm already emits a
glob's `*` apart from its name. The ratchet's tdt comparison reads a cast terminal as the perly.y terminal it names,
matched against toke.c's head or, for the star of a postfix dereference, a type the `postfix-deref` family folds
onto the token.

**The fitted call terminal**, by perigrin's ruling ("Extend fitted-terminal"): `amper: PRIVATEREF`, perly.y:1842's
fit for a word before `(`, stands for toke.c's zero-width `&` and the name. The tdt comparison reads the chart's
PRIVATEREF as matching a token the `noamp` family aligns with `PERLY_AMPERSAND` folded onto it (`callFit`), keyed on
the fixture's family and fold, never on the token's text.

**Postfix dereference** is one `Variable` token after `->` (`$*`, `@*`, `%*`, `&*`, `**`, `$#*`), and a bare
`@` or `%` there opens a slice. By perigrin's ruling ("Yes, if measured small") the variable heads leave `->` out
(`NotPrev1`), and classes keyed on the text after `->` take it: `@*` and `%*` map to their sigil's cast terminal,
`$*` and `**` to `CAST_STAR` (`$*` as its star, since `CAST_DOLLAR` there would read `$r->$*{a}` as a method named
by a cast), `&*` to `PERLY_AMPERSAND` and `$#*` to `DOLSHARP`, which nothing else after `->` is. So `$r->@*[0]`,
`$r->%*{a}`, `$r->**{CODE}`, `$r->$#*[0]` and `$r->$*(1)` are refused, and a `$` variable after `->` stays a
method's name. The lone `$` pvm emits for `->${` is a cast's sigil, `CAST_DOLLAR`, so a method named by a cast
(`$x->${\ 'm'}`, `$x->${$m}(1)`) is accepted. The generator's literal check lexes a class keyed on `->` after one, and a `DerefSigil`'s before a
variable. Interleaved Intern and Lex benchmarks against the base show no per-token cost.

A blank between `->` and `@` leaves toke.c's `->` a `TERM(ARROW)`, and the nameless `@` is then `PREREF(PERLY_SNAIL)`
(yyl_snail), so `$r->@ {"a"}` is the slice `$r->@{"a"}`: the `DerefSigil` after `->` maps to `PERLY_SNAIL`, never to
`CAST_SNAIL`. After `->% {`, `->* {` and `->& {` toke.c reads an operator and perl refuses; `->* {` is accepted, since
the key cannot tell `*{` from `* {`.

One limit, pinned in `TestPortSlicesAndDerefsKnownLimits`: a cast of a block that opens with `{}` is accepted
(`${ {} }`, `@{ {} }`, `%{ {} }`, `&{ {} }`, `*{ {} }`, `$x->${ {} }`, `${ {} }{a}`, `print ${ {} };`), where
toke.c's `scan_ident` reads `${ {` after any sigil as the punctuation variable `{` and perl refuses the `}` left
over ("Unmatched right curly bracket"); `internal/lexer` emits a `DerefSigil` and two `{` there, so the root is the
lexer's token (v0.1 `01a1a9d4`).

### Scope-decided words

Whether a bareword is a declared sub, and of which prototype, is scope, decided after the chart (waypoint 2).
A word the neighbours leave undecided (`TestPortScopeDecidedSuperpositionBounded`) is every terminal it could be:
BAREWORD, NOAMP, FUNC0SUB and UNIOPSUB, LSTOPSUB too before a `{`, and METHCALL0 where the indirect objects
put it (perly.y:1046, 1330, 1646, 1678, 1680, 1691). The neighbours decide the rest, as toke.c's yyl_just_a_word does:

- Before `=>` a word is BAREWORD (toke.c:8176), a keyword too: toke.c checks for `=>` before it looks a word up
  (yyl_fatcomma, toke.c:9332-9336 and 9440-9453), so init's `cutBeforeFatComma` takes every class keyed on a word's
  text away there, `x`'s included (`f(eq => 1, print => 2)`, `{ y => 2 }`; `$a x => 1` is refused, as by perl).
- `AUTOLOAD`, `DESTROY`, `BEGIN`, `UNITCHECK`, `CHECK`, `INIT` and `END` name a sub only where a statement starts
  and are any word elsewhere (toke.c:8313-8322), so init's `addStatementWords` gives them each class of a word no
  class names, after anything but a statement start (`$y = AUTOLOAD();`, `$y = END x 3`).
- `CORE::kw`, kw a keyword, is kw (yyl_key_core, toke.c:9230-9257): by perigrin's ruling ("Strip CORE:: in chart/lex")
  chart/lex gives such a word kw's text, so kw's classes and every neighbour check on kw apply, and the Prefix
  `CoreQualified` (`C`), a value of an existing key part as `PackageName` is. toke.c reaches yyl_key_core before its
  fat-comma check, so `CORE::eq =>` stays CHEQOP and is refused as by perl: the fat-comma BAREWORD classes leave
  `CoreQualified` out, and `cutBeforeFatComma` keeps a copy of each class it cuts for it. The prefix test runs only on a
  word opening with `C`; INTERN's reference reads it with a regular expression (`refCore`). `CORE::GLOBAL::foo` is a word.
- Before `(` it is a call: toke.c emits a zero-width `&` and the
  name (toke.c:8195-8211), which the word stands for as PRIVATEREF, by a fit of perly.y:1842 (`amper:
  PRIVATEREF`), so `foo(1)` and `f()->{foo}` are accepted.
- A word ending in `::` is a package name, never a call (safebw, toke.c:8098-8111), so `a::()`, `Foo:: 1` and
  `Foo:: $x` are refused (the last was a METHCALL0 before; those classes now leave a package name out too). chart/lex gives such a word the Prefix `PackageName` (`:`), read from its last two bytes: a value of
  an existing key part, as a Number's version shape is, and an extension of spec §7 that perigrin accepted
  (2026-11-05, "Accept").
  INTERN's reference reads it with a regular expression (`refPackageName`).
- After `->`, `&`, a glob's `*`, `sort`, `require`, the loop exits and the words that name something (`sub`,
  `package`, `use`, `no`, `format`) toke.c forces the word to a name (`forcedWord`), so it carries no
  scope-decided reading, and before `(` after `require` or a loop exit it stays BAREWORD, which perl refuses. After
  `do` toke.c forces the word only before `(` (yyl_do), so `do foo(1);` is refused and `do foo;` and `do foo 1;` stay
  scope's (`callForced`). Neither forces a keyword, so `next x(1);`, `require x(1);` and `do x(1);` are calls
  (`xForced`), and yyl_do leaves `CORE` too (`sub CORE {1} do CORE(1);`), where `next CORE(1);` stays refused.
- NOAMP stands for the zero-width NOAMP and the word as `subname` (perly.y:1646's fit). A sub's name is PRIVATEREF
  as well as BAREWORD (a lexical sub's, force_ident_maybe_lex), and `my sub`, `our sub` and `state sub` are a
  fit of perly.y:638 and 768 with the declarator's KW_MY before KW_SUB_named or KW_SUB_named_sig, which toke.c reads
  silently (yyl_my).
- FUNC0SUB completes a term (`TERM`), so `x` after it is the repetition operator; NOAMP, UNIOPSUB and LSTOPSUB
  leave PL_expect at XTERM and do not.

Limits, each with a `shortcut:` comment, pinned in `TestPortScopeDecidedKnownLimits` where a test can: a declared sub
named `x` carries the scope-decided readings only after a statement-start token (`addStatementX`), so
`sub x {1} $y = x 1, 2;` is refused where perl accepts it; a word two after `use` or `no` carries no scope-decided reading (afterUse, the version family's).
`my sub Foo::bar {1}`, its `our` and `state` forms, `CORE::foo(1);` and `$y = CORE::;` are accepted, where toke.c croaks while it
scans the word; by perigrin's ruling ("Pin as known limits") the root is the lexer's (v0.1 `01a1aa1e`).
`CORE::qw`, `CORE::q`, `CORE::qq`, `CORE::m`, `CORE::qr`, `CORE::qx`, `CORE::s`, `CORE::tr` and `CORE::y` with their
delimiters (`@y = CORE::qw(a b);`, `$y = CORE::q{x};`, `$y = CORE::tr/a/b/;`) are refused where perl accepts them:
internal/lexer emits a `CORE::` quote-like as a word and loose tokens (v0.1 `01a1ae56`), so stripping `CORE::` leaves
the operator's name with no delimiters.
`print FH x 3;`, `new Foo x 3;`, `foo bar x 3;` and `foo { a => 1 } x 3;` stay accepted (scope, above), and `x 1;` is
accepted as `foo 1;` is: perl refuses both, and accepts each once a sub of that name is declared (`sub x; x 1;`).
`terminals-per-token` and `items-per-token` rise by design.

### Builtin classes against `parse.CoreShapes()`

`TestCoreShapeDifferences` (`grammar/coreshapes.go`) compares the words the port classes as `FUNC0`,
`FUNC0OP`, `UNIOP`, `FUNC1`, `LSTOP` and `BLKLSTOP` with the shapes `parse.CoreShapes()` gives them (spec §5.3,
§3 rule 9), under a stated correspondence: niladic is `FUNC0` or `FUNC0OP`, unary is `UNIOP` or `FUNC1` (toke.c
decides by the next token, so either or both), list is `LSTOP`. Every word with a shape and no such class, a class
and no shape, or a class its shape does not correspond to is a difference, and
`grammar/testdata/core-shape-differences.txt` lists each with both sides and a reason decided by toke.c and
`perl -c`. The test holds the computed differences equal to that list, so one appears or goes only by a reviewed
edit. Neither side is changed to shorten the list: `internal/parse` is read-only and a class change is its
family's.

The correspondence is a hypothesis, amended once: `FUNC0OP` is niladic (`FUN0OP`, toke.c:263, differs from
`FUN0` only in the token and the op it returns). `BLKLSTOP` is not folded into `LSTOP`; `all` and `any` are listed.

### Precedence levels against chapter 4 and `parse.LevelsAccountedFor()`

`grammar/perly.go` states perly.y's 32 precedence declarations (`perlyPrecedence`, `%left|%right|%nonassoc
TOKEN... # perly.y:N`, loosest first, each citing one line of `perly.y:150-182`), and every rule whose alternative
carries a `%prec` carries it too, as `lhs: rhs %prec TOKEN # perly.y:N` (`TestPrecedenceCitations` holds them to
perly.y's 39 uses). **Recognition ignores both** at waypoint 1 (spec §8): `TestPrecedenceIgnoredByRecognition`
shows `$a <=> $b <=> $c;`, `1 .. 2 .. 3;` and `print $fh [];` still accepted. They are waypoint 2's data, and the
generator's unused-precedence refusal reads them. The declaration of perly.y:177 is fitted: `POSTJOIN` is left out,
since its one rule (1512) is excluded interpolation-only and the refusal would refuse it.

`TestLevelDifferences` (`grammar/levels.go`) compares each level, by number, with chapter 4's level table
(associativity and tokens) and with `parse.LevelsAccountedFor()` (presence only: its account is prose), and holds
every difference equal to `grammar/testdata/level-differences.txt`, each entry with both sides and a reason.
Associativity agrees with chapter 4 on all 32 levels; the differences are tokens chapter 4 leaves out or spells by character.
`internal/parse` and chapter 4 are read, never changed.

## The thin chart/lex: `lex`

`lex.Lex(src, table)` runs the production lexer up front (`lexer.Tokenize`)
and then does **one table lookup per significant token** through the
compiled token table (`grammar.TokenTable`), yielding its terminal-set id
(spec §5.2, §7). The lookup allocates nothing, and it uses no regexp. It
walks the table laid out flat (`TokenTable.Lookup`), behind a cache of answers
by whole key (`grammar.LookupCache`) that each `Map` call borrows from a pool.

**The stream is the lexer's tokens, not a copy of them.** `Stream.Raw` is
the slice `lexer.Tokenize` returned; what the chart adds sits beside it.
`Stream.Tokens` holds one `lex.Token` per significant token: its index into
`Raw` and its terminal-set id, no pointers, so the garbage collector never
scans it. The first unmapped byte is `Map`'s second result, and heredoc
bodies are a side table. `Map` allocates the same small number of times
whatever the file's length.

- **Trivia** (`lexer.IsTrivia`: whitespace, comments, pod, the data section)
  attaches to the next significant token by index: `Stream.Lead(i)` is the
  `Raw` run between token `i` and the one before it, and trivia after the
  last one is `Stream.Tail()`. Nothing is dropped: `Stream.Text` gives back
  the source.
- **Heredoc bodies** attach to their heredoc's opening token, never to the
  token after the opener's line: each body goes to the first waiting opener
  whose tag its terminator line spells, in the stream's `Bodies` side table
  under its opener's index (`Stream.Body(i)`). Trivia a body interrupts still
  goes to the next significant token; `Lead` leaves the body out. A body no
  waiting opener names has its opener inside a quote-like (`"@{[ <<EOF ]}"`,
  one Quote token), and goes to the token before it, a known limit pinned in `TestPortQuoteLikesHeredocBodyIsTrivia`
  (only `Stream.Body` and the tdt pairing read which token holds a body), and
  an opener no body follows gets set 0, as perl finds no terminator for it.
- **Every token is mapped.** A token whose key matches no class gets set 0,
  the empty set, and the stream goes on, so a file is always lexed whole;
  `Lex` also returns the byte of the first such token, or -1. The recognizer
  refuses there, and the ratchet's `lex-unmapped` counts the file.
- **The key** a token is looked up by is its lexer kind, text, flags and
  prefix (a variable's sigil, `#` for `$#`; a quote's opening quote or
  operator, `Q` for `qq`), and the classes of the two significant tokens
  before it and the one after. Each token is interned once and its class id
  reused where it is a neighbour. A neighbour's class carries what its own
  predecessor decided: a `{` its brace (subscript, hash or block), a `*`
  whether it is a glob sigil (`Glob`: the token before it completes no term,
  toke.c's `yyl_star`), and a word whether it names something (`Name`: it
  follows `->`, `&` or a glob's `*`). This extends spec §7's neighbour class,
  by perigrin's ruling on the `x` fitting; the key still reads three classes,
  so the lookup stays constant-time and allocation-free.
- **Interning** (`TokenTable.Intern`) writes the token's class, class id and
  text id into a `grammar.Token` the caller holds, reading its predecessor's
  class through a pointer: `Map` keeps the last token and the one being
  interned in two alternating slots, and takes the token's `TokenView` by
  pointer too. Copying a freshly written class or view by value stalled on
  store forwarding, and was most of Intern's cost. A text whose first byte
  and length no dictionary text has, and a class whose kind no neighbour
  class has, get id 0 without being hashed; the rest are looked up in two
  open-addressed tables keyed by an integer (`classKey`, and a text of at
  most seven bytes packed with its length), a multiply and a probe where a
  Go map hashed, and a longer text in the text map.
  `TestInternIdentityOverCorpus` holds every corpus file's ids to a
  reference that matches each class directly.

- **The isa feature** is the one bit of scope `Map` tracks, by perigrin's ruling ("Track the isa feature"): toke.c
  reads `isa` as a keyword only where the feature is on (keywords.c:353), and elsewhere `Map` gives the word the
  text `grammar.PlainIsa`, which no class names, so `ok(isa Foo('x'))` is an indirect method call and `$x isa Foo`
  is refused. A `use` or `no` statement sets the bit as feature.pm does (`isaAfter`): `feature` or `experimental`
  naming `isa`, `:all` or a bundle from `:5.35`, a bare `no feature`, and `use VERSION` (on from 5.35, off
  below it, `use 5.008` and `use 5` included). Every `{` saves it and the matching `}` restores it, which is the block scope; a hash's or a
  subscript's braces save and restore it too, harmlessly, since no `use` stands inside one. The work is a kind
  and length test per token, and a scan of the statement at a `use` or `no`. INTERN's reference reads the same
  statements with regular expressions (`refIsaAfter`). Not seen, each a `shortcut:` in `isaAfter`: a feature a
  module or a `BEGIN` block imports, or one a string eval turns on. `use v5;` leaves the bit, where perl clears it,
  because internal/lexer emits a bare `v5` as a Word rather than a version (`corpus-covers-t2` `01a0dc72`); pinned in
  `TestPortRelationalAndUnaryIsaLimits`.

`lex.Map` is the same over tokens the caller already has; `Lex` is `Map` over
`lexer.Tokenize`'s, and the ratchet times `Map` with its `Tokenize` call
inside the span, so `lex-overhead`'s numerator always includes the lexer.

## The recognizer: `recognize`

`recognize.Recognize` runs a YAEP-style Earley recognizer over the generated
tables and a `chart/lex` lattice. Error recovery is not built yet.

- **Input** is a lattice of terminal-set ids; the recognizer never sees a
  terminal name. Each edge's token carries one id, `lex.Token.TerminalSet`,
  an index into `grammar.Tables.TerminalSets`, whose entries are sorted
  lists of terminal symbol ids; a token is read as every terminal in its
  set, so a token whose set holds several is scanned as each of them. **Id 0
  is the empty set**, so a token never given an id refuses rather than
  scanning. `BuildTables` leaves `TerminalSets` empty; `grammar.Generate`
  fills it from the grammar's compiled token classes, and tests on toy
  grammars fill it themselves (the empty set, then one singleton set per
  terminal). Sets may be appended after `New`, and a recognizer built
  earlier reads them (`TestRecognizeSetAddedAfterNew`); a set already
  there is never changed. Nodes are
  numbered so that every edge runs forward. `Recognize` validates this and
  returns an error, never a panic, for an edge with a node outside the
  lattice, an edge that does not run forward (a zero-length one included),
  or a set id outside the table. A lattice with no edges is valid: the
  grammar accepts or refuses it.
- **Class keys are set ids.** A token's class is its set id. The lookahead
  at a node is the set id of the edges leaving it, or the end-of-input class,
  key -2, where none leave, so a set appended later never takes its key; each
  class's bitset is built once per tables value (below), so the per-token
  lookup allocates nothing
  (`TestClassKeysAllocationFree`). Where edges with different sets leave one
  node, its lookahead is their union, built there, and the transition cache
  is not used at that node.
- **Construction.** What depends only on the tables (each state's LHS
  sets, the FOLLOW bitsets, the Leo items, each set's lookahead bitset and
  whether it holds a `}`) is built by the first `New` over a
  `*grammar.Tables`, kept beside that value (`Tables.Derived`), and shared
  read-only by every recognizer over it, from any goroutine; a later `New`
  allocates 6 objects where it allocated 1,672 at c99366f8 (`BenchmarkNew`).
  A tables value anywhere works, a package variable's included
  (`TestNewOverAnyTablesValue`). A set appended to the tables is covered when
  a recognizer first meets its id, without writing what other recognizers
  read; two recognizers doing so at once agree, race-free
  (`TestRecognizersExtendSharedTablesConcurrently`). The tables' states,
  rules, items, symbols and FOLLOW sets must not change after the first
  `New`; a copy of a tables value is a value of its own and shares nothing
  (`TestRecognizersOverDistinctTablesShareNothing`). The first `New` over a
  tables value writes beside it, so do not copy `*tab` while that `New` runs;
  copying it beside a later `New` is safe (`TestCopyTablesBesideLaterNew`).
- **Not safe for concurrent use.** A `Recognizer`'s intern table and
  transition cache change with every recognition, across inputs, and stay
  its own (`TestRecognizerStateNotShared`).
- **Items** are an LR(0) state and the distance in lattice nodes back to the
  item's origin. A set is its items, sorted, interned as a (core, distances)
  pair, so two positions with equal content share one `*Set`. A node that an
  edge passes over is inside a branch: its set is built but not interned.
- **Lookahead.** In up-front mode a complete item's rule completes only if
  the next token, or the end of input, can follow its LHS. That keeps
  out of each set the completions the next token cannot continue. Observed
  at aa5aed2f, summing `len(Set.Items())` over every node of `Chart.Sets`
  (node 0 included), as every item count below does: 100 statements
  `$x = 1;` give 2,001 items with lookahead and 2,201 without; both are
  linear, since perly's `stmtseq` is left-recursive.
  `Expected(pos)` is the union of the expected terminals of every state in
  the set at `pos` as recognition without lookahead builds it: it closes
  the set again under every completion, outside the interned content, so a
  refusal point still names what was expected there.
- **The transition cache** is keyed by (set, token, lookahead) and the
  edge's span. A key alone does not fix the result: completion reads the
  sets where complete items began, which are not part of the source set. So
  each entry records those sets, and a hit requires them to be the same.
  `TestCacheChecksOriginSets` fails without that check. The cache is not
  used inside a branch (`TestCacheNotKeyedOnBranchSet`), nor where two edges
  meet (`TestCacheNotUsedAtJoin`), and a transition that read any set inside a
  branch is never cached. A key's entries are a trie over those origin sets,
  in the order the transition read them: which set it reads next follows
  from the key and the sets already read, so a lookup reads one set per
  dependency and follows its id. A lookup returns what a scan of the entries
  in the order stored would, the first whose sets hold
  (`TestCacheMatchesNoCache`, which also holds a warm cache to no cache on
  acceptance and `Expected` at every node of the corpus sample's whole
  files). At c6de5c76 each key held a list that every lookup scanned, and
  on flat top-level statements a key's list grew by one entry per statement,
  each differing only in the set where its statement began: recognition time
  was quadratic in the statement count though the item count was linear
  (`TestCacheLookupBounded` holds the longest scan of one lookup equal from
  100 such statements to 400). The cost on the files the aggregate excludes
  is under "The aggregate" in the speed protocol. The `Expected` memo had
  the same per-statement growth and is keyed the same way, in the same trie
  under span 0 (`TestExpectedMemoBounded`, `TestExpectedMemoMatchesNoCache`):
  observed at eab40199 (before) and on the branch of 01a1b8a6-e81e (after),
  `Expected` at every node after recognition took 1,595–1,778 ms before and
  363–392 ms after on the whole `code_too_large.t` (its longest memo list
  had 9,997 entries; a lookup now examines at most 4 distances), and
  15.7–17.5 ms before and 18.5–19.9 ms after on `std5k.pl`, load 1.6–2.4.
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
  linear, where without it they are quadratic
  (`TestLeoLinearAtLookaheadZero` and `TestLeoStatementSeqAtLookaheadZero`).
  perly has one production of that shape, `else: KW_ELSIF … mblock else |
  empty`. No test reproduces the counts that follow; they were observed at
  aa5aed2f. `if (1) { 1 }` followed by 50, 100 and 200 ` elsif (1) { 1 }` arms
  has 1,375, 2,725 and 5,425 items at lookahead 0 with Leo and 2,551, 7,576
  and 25,126 without. With lookahead on, the default, it is linear either way
  (1,028, 2,028, 4,028 with Leo; 1,076, 2,126, 4,226 without). At waypoint 1
  operator chains get no Leo benefit, because precedence is not applied yet:
  `$a` followed by 50, 100 and 200 ` = $a` and a `;` has 5,463, 20,913 and
  81,813 items at lookahead 0, with Leo or without, quadratic. Linear here
  means the item count, not time: with the cache on, each use of a chain walks it to record its
  sets, so time stays quadratic until memos are interned. Leo chains were
  not the only such term: at c6de5c76 the transition cache's lookup was
  quadratic in the number of top-level statements in default mode (see the
  transition cache above). A chain stops at
  a start-rule item from node 0, which acceptance must see. `Options.NoLeo`
  turns it off; it is the only Leo switch.
- **The `;` toke.c inserts, by a narrow ruby slipper** (spec §5.2). toke.c
  inserts a `;` where a scope ends: before the `}` that closes it
  (`yyl_rightcurly`, which forces the `}` and returns `PERLY_SEMICOLON`
  first, toke.c:6890-6897 at 94e50866) and at the end of the file
  (`yyl_fake_eof`, toke.c:7670). pvm's lexer emits neither, and its tokens do
  not change. So at a node whose next token's set holds `PERLY_BRACE_CLOSE`,
  or that ends the input, and whose set expects `PERLY_SEMICOLON`, the
  recognizer scans a zero-width `PERLY_SEMICOLON` there **beside** the next
  token, never instead of it. This models the language, not toke.c's
  mechanism (spec §5.2, amended at 63a92338): a statement's terminating `;`
  is optional where a scope ends, and a `;` written there is a null
  statement, so the insertion accepts nothing perl refuses. A `}` whose set
  expects no `;` gets none, and a missing `;` before any other token is
  still refused. The items the `;` advances are closed beside the
  node's set, never in it: `Sets[pos]` and `Expected(pos)` stay what the
  source may contain next, and the next node scans the `}` from both, the
  inserted items' origin sets being cache dependencies like any other. At
  such a node the lookahead that prunes completion holds `;` as well as the
  next token. `(*Chart).Inserted()` is the nodes where a `;` was scanned,
  ascending. The reference recognizer inserts the same `;` with a set of
  its own, so it stays the oracle (`TestSlipperReferenceAgrees`). Nothing
  else is inserted; interleaved lexing is future work (spec §13). No
  statement rule is fitted for the inserted `;`. The subscript,
  anonymous-hash and slice rules, where `perly.y` spells toke.c's `;` before
  every real `}` (`expr PERLY_SEMICOLON PERLY_BRACE_CLOSE`, perly.y:1349,
  1366, 1369, 1373, 1536, 1614, 1624; toke.c:6897), are fitted without it
  (`fitted: the ; before } is toke.c's, not the source's`), because that `;`
  never appears in source; so `$h{1;}` is refused, as perl refuses it
  (spec §5.2).

`make fuzz-chart-recognize` runs `FuzzRecognize`. It checks that recognition
never panics, that acceptance is the same with a dead-end branch added and
through a cache warmed by earlier inputs, that the warm recognizer's `Expected`
(memoized across inputs) equals the reference's, and that, under every `Options`
combination, acceptance and `Expected` at every position equal the reference
recognizer's (`recognize/reference_test.go`). The reference is a textbook
Earley recognizer, test-only: predict, scan and complete over the grammar's
rules, run to a fixed point, with no LR(0) states, interning, cache or Leo, and
sharing no code with `recognize` (`TestReferenceSharesNoCode`), so it fails
independently of the recognizer it checks. With no grammar argument each input
byte is the singleton set id of one terminal of the toy statement grammar, so
every input is a valid terminal string; with one, its bytes decode to a small
grammar (the first rule's left-hand side is the start, and may sit on a right-hand side) and the input's bytes
pick among that grammar's terminals.

`TestReferenceRecognizer` (`recognize/agree_test.go`) holds the recognizer to
the same reference, under every `Options` combination, on acceptance, on
`Inserted` and on `Expected` at every node: over every short input of toy
grammars (left and right recursion, nullable chains, ambiguity, a dead end,
`;`-terminated statements in braces) and over a deterministic corpus sample
(`std5k.pl`, the port snippets and every 24th Guacamole case, each lexed
through `chart/lex`, at most 300 tokens). Agreement does not need acceptance:
`TestReferenceAgreesOnTruncatedFiles` cuts each sample file mid-statement (after
its first `=`, else its first token but `;`) and requires both to refuse at the same node with the same `Expected`.
`TestReferenceRecognizerCatchesDroppedTransition` removes a start-state
transition the sample uses and requires the disagreement to be reported.

`TestSetDigestGolden` (`recognize/setdigest_test.go`) holds set identity to
its golden, `recognize/testdata/set-digest.golden`. Over the same corpus
sample, each file recognized whole, with one recognizer per option set
(`Options{}`, `NoCache`, `NoLeo`, `NoCache` with `NoLeo`, `NoLookahead`), it
hashes every set at every node (`Set.Items()`), `Inserted()` and `Expected` at
every 37th node, and writes one line per option set: its name, the accepted
count, the cached count and the digest. Only
`go test ./internal/chart/recognize -run '^TestSetDigestGolden$' -recognize.update-set-digest`
rewrites the golden; nothing regenerates it by default.
`TestSetDigestDetectsChange` keeps it from passing vacuously: every option set
accepts files, the cached count is 0 under `NoCache` and above 0 otherwise, and
one changed item, `Inserted` entry or sampled `Expected` each moves the digest.
A change that moves the golden has changed set identity: it is never
regenerated to make a change pass, and in chart-recovery a regeneration after
P2 is a stop reported to perigrin.

## The forest: `forest`

`forest.Parse(tab, src)` lexes and recognizes `src`, then builds its shared
packed forest after recognition (spec §9.4). It is the single entry the
precedence constraint, resolve and the readings count extend. A file the
chart refuses gets a `Refusal` and no forest, never a partial one
(`TestForestRefusedChartHasNoForest`). `Parse` recognizes first, as the
port's refusal list does: the refusal is at the token whose set died, or at
the end of the file, and is `lex-unmapped` only when that byte is the first
token no terminal admits, else `grammar` (`TestParseRefusalMatchesRecognition`
over `testdata/port/refused-where.txt`; an unmapped token past the dead set
does not move it, `TestParseRefusalUnmappedAfterDeadSet`).

- **No fact is added to an Earley item** (spec §3 rule 12, §10). Items stay
  (state, distance), recognition stays Boolean, and no existing file of
  `recognize` changes for the forest. It reads the chart through two
  read-only accessors in `recognize/finished.go`. A rule the chart completed
  is a reduce of a state in some set, or one a Leo memo skipped
  (`LeoLink`), or one the inserted `;` completed beside a set (`Slipped`).
  Each is a derivation the chart proved, so every node has one. Every node on
  a derivation of the file is completed in the chart, so none is missing.
- **A node** is a symbol over a span of lattice nodes, or a rule position
  there: a rule's first symbols, which binarizes a long rule so the forest
  stays cubic. A node's `Alts` are the ways its span divides, each the rule's
  earlier symbols (`Left`) and its last (`Right`). A node is tried only by
  the rules the chart completed for it. A split is tried only where the
  rule's earlier symbols stand as an item in the chart. The candidate starts
  are read from whichever list is shorter: the last symbol's completions at
  the node's end, or the nodes where an item from its start stands. That keeps
  a right recursion and a left one linear (`TestForestRightRecursionLinear`).
- **Leo.** Where a memo completed a right recursion in one step, the links it
  skipped are added at a node only where the chain's top is among that
  node's completions, which is where recognition took the chain. Each memo
  is read once, as `LeoLink` makes it from the sets. A node the transition
  cache served ran no completion and made no memo, though its set holds the
  chain's top, so `LeoLink` reads the sets rather than the side table, and
  stores nothing. Leo on and off give the same forest, on a file of cached
  sets included (`TestForestFromChart`).
- **The slipper's `;`** is a zero-width `PERLY_SEMICOLON` leaf at a node in
  `Inserted()`. `Slipped` closes the items it advances with Leo off, so
  replaying it after a cache hit writes no memo. A derivation's `;` leaves
  are a subset of `Inserted()`, not all of it: `{ 1; }` uses none, though
  the slipper inserted at the `}`. Aligning leaves to `-DT`'s tokens must
  allow for that.
- **Cycles.** No node has an `Alt` whose child is the node itself; such an
  `Alt` is dropped as it is built, since no derivation can use it. That
  removes perly.y's one cycle shape, `stmtseq` over a span beside a
  zero-width `;` statement. So a perly.y forest is acyclic, and a count of
  readings taken as a product over its packed nodes is exact. The test holds
  every forest of the corpus sample to that with a full cycle search, not
  only self-loops (`TestForestFromChart`; `TestForestCycleCheck` shows the
  search finds a two-node cycle). A cycle in a perly.y forest is a defect to
  fix, not something to count around. A toy grammar can still build a
  longer cycle, which `Derivations` survives by treating a node met again on
  its own path as having no derivation there.
- **Derivations.** `(*Forest).Derivations(n, limit)` enumerates at most
  `limit` trees of node `n`. It also reports `cut` when a bound stopped it:
  the limit, reached even if exactly `limit` trees exist, or a budget of
  steps proportional to `limit` and the forest's size. A derivation never holds a (symbol, span) node inside itself. So a
  cycle adds none, and a slipper `;` appears only where a statement needs
  it, never as a null statement beside one that does not (`{ 1; }` has one
  derivation, not one per `;`; `TestForestFromChart`). Enumeration stops at
  its limit on a cyclic, nullable, exponentially ambiguous toy. On a diamond
  of unit-rule cycles, whose dead ends grow exponentially with its depth, it
  stops at its budget and says so (`TestForestEnumerationBounded`).
- **Outside reference** (spec rule 11). On toy grammars (left and right
  recursion, nullable chains, ambiguity, a cycle, Leo beside ambiguity), under
  every recognizer option, every derivation is checked against the grammar's
  rules, and their number equals a brute-force enumerator's, written in the
  test from the rules alone. Toy grammars naming `PERLY_SEMICOLON` and
  `PERLY_BRACE_CLOSE` put the slipper in that comparison, with an enumerator
  that admits a zero-width `;` exactly before `}` and at end of input, Leo on
  and off (`TestForestSlipperBruteForce`). Over the corpus sample (`std5k.pl`, the port
  snippets, every 24th Guacamole case) a file has a forest exactly where the
  chart accepts it, and a derivation whose leaves are its tokens in order. An
  empty file, one of comments only, and files of `;` do too
  (`TestForestDegenerateFiles`).
- **Linear lattices only.** shortcut: the token ending at node j is edge
  j-1, and an inserted `;` is followed only by its `}`, as up-front lexing
  makes them; `build` refuses any other lattice
  (`TestForestLinearLatticesOnly`). Index the edges by node, and keep a
  `;`'s completions apart, when interleaved lexing branches.
- **Forest identity.** `TestForestDigestGolden`
  (`forest/forestdigest_test.go`) holds the forests of the corpus sample to
  `forest/testdata/forest-digest.golden`: the accepted count, then one line
  per accepted file, its path and the hash of a text dump of its forest. The
  dump visits every node reachable from the root and writes them in a fixed
  order, by span, then symbol name or rule position, never by pointer, map or
  allocation order: a symbol node gives its symbol's name and span, a rule
  position its rule, dot and span, and each alternative its rule, its split
  (where its last symbol starts, or the node's end for an empty rule) and both
  children's keys, a child's key being its label and span
  (`s<symbol>:<from>-<to>` or `r<rule>.<dot>:<from>-<to>`, `-` for none);
  the alternatives are sorted. It names no Go type, so a rewrite of
  the node types rewrites the dump and leaves the golden. Only
  `go test ./internal/chart/forest -run '^TestForestDigestGolden$' -forest.update-forest-digest`
  rewrites the golden; nothing regenerates it by default.
  `TestForestDigestDetectsChange` keeps it from passing vacuously: the sample
  has accepted files; one alternative's split, its Left rewired to a
  same-span node or nil, its Right to a same-start node, its Left and Right
  swapped around a zero-width Left, or one node's span each changes the
  dump; and the same nodes built in reverse order with reversed alternatives
  do not. A change that moves the golden has changed the forest: it is never
  regenerated to make a change pass, and in chart-recovery a regeneration
  after P5 is a stop reported to perigrin.
- **Cost.** `BenchmarkForest` times the forest of `std5k.pl` after
  recognition. When it was added, in five rounds interleaved with
  `BenchmarkRecognize` (`-benchtime 40x`, load 7.5–8.1), it took a median
  29.6 ms per op, 21.2 MB and 20,509 allocations, beside recognition's
  19.4 ms. An earlier build of the forest, which expanded every Leo chain
  and tried every split, took 36.0 ms in the same rounds. The span checkpoint 3 times includes it (spec §9.3).

## The generated tables: `make chart-tables`

`make chart-tables` runs `grammar/cmd/gen`, which builds the tables
(`grammar.BuildTables`) of the committed grammar and writes them to
`grammar/testdata/tables/`, with the token table compiled from the
grammar's token classes (`tokens.txt`). The ratchet's `tables-drift` holds
every one of them to what the generator writes. A grammar with no rule is
refused, as any other inconsistency is.

The generator refuses an inconsistent grammar (`grammar/refuse.go`, spec §5.2):
it prints every offence to stderr, writes no table, leaves those already there
as they were and exits 1. The offences are a rule without a citation, or one
outside the reference list (`testdata/perly/alternatives.txt`); a class literal
the lexer does not lex to the class's kind and text, between two terms or from
a statement's start through `lexer.Checkpoints` and `lexer.Resume`; a class
naming a terminal no rule uses, or a nonterminal; a terminal no class admits, but `PLUGIN_*`, the seven `GRAM*` start selectors, `PLUGEXPR`
and `PLUGSTMT`; an unreachable or unproductive nonterminal; and a precedence
token no rule and no `%prec` uses. `TestCommittedGrammarConsistent` holds the
committed grammar to them. Token classes are declared in `grammar/perly.go` by
terminal name, one slot per lexed token as "The grammar" above states, and resolved to ids
at generation (`Grammar.ResolveClasses`), so a rule added elsewhere does not
move a class off its terminal. The output depends only on the grammar's rule order, never on map iteration or the clock,
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

- **`tokens.txt`**: the compiled token table (`TokenTable.Text`). First
  `set <id> <terminal ids>`, one line per terminal set, set 0 the empty set;
  then the table's nodes, children before parents, each shared node once:
  `node <n> leaf set <id>`, or `node <n> part <part> otherwise <n>` followed
  by one `  <value> -> <n>` line per value that leads elsewhere. A part is
  0 kind, 1 text, 2 `OpensBlock`, 3 `TermFollows`, 4 prefix, 5–7 the
  classes of the second and first previous and the next token, 8 whether a
  sub's name follows, 9 `Global`, 10 `LineColon` and 11 `Blanks`; a text is
  printed quoted and a neighbour class as its fields. Last, `root <n>`.

These are separate files so that `states.txt` stays under the
check-added-large-files limit.
