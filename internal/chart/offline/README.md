<!-- ABOUTME: The chart spike's offline scripts: run by hand, never by make test or the ratchet. -->
<!-- ABOUTME: Covers the perl build, the bison referees, the perl labels, the -DT capture and its alignment, and what each costs to run. -->

# internal/chart/offline

Scripts run by hand against perl blead 94e50866 in `/opt/corpora/perl5`.

- They never run in `make test` or `make chart-ratchet`.
- Nothing they build (perls, referee binaries) is committed.
- Output that tests read is committed.
- The entries below are in dependency order: `build-perl.sh` first, then the
  scripts that need only the tree or bison, then those that use the perl,
  ending with `label.sh`, which needs the referees and the captures.
- Every script that reads a pinned tree (the perl5 tree, and PerlOnJava for
  `label.sh` and `tdt-align.sh`) requires it to be at its pinned commit **and
  clean**: `git status --porcelain` empty, untracked files included. An edited
  `perly.y` at the pinned commit is refused. A refusal names the script and
  the tree on stderr and exits 1. Build and check in a `git clone --shared`
  copy, never by editing the pinned tree.
- Inputs: the perl5 tree at 94e50866 (`/opt/corpora/perl5`), the pinned perl
  (`/opt/perl-94e50866-debug/bin/perl`), `T1_CORPUS`
  (`/opt/corpora/PerlOnJava/src/test/resources/unit`, PerlOnJava 73be162b),
  bison 3.8.2 and the Go toolchain. Each command below is run from the
  repository root.

## build-perl.sh

`sh internal/chart/offline/build-perl.sh <perl5-tree> <prefix>` builds the
`-DDEBUGGING` perl of blead 94e50866 and prints the installed binary's path.
It refuses a tree whose HEAD is not 94e50866 before building or creating the
prefix. It builds in a `git clone --shared` of the tree under `$TMPDIR`, so the
tree is never written, and removes the clone on exit. `-Uversiononly` makes
the install `<prefix>/bin/perl` rather than `perl5.45.0`.

This milestone uses `sh internal/chart/offline/build-perl.sh /opt/corpora/perl5 /opt/perl-94e50866-debug`.

Build time: 382 s wall clock on the 4-core development machine, including
`make install` (a first run with the same options took 146 s; the figure
depends on load from other jobs). Installed size: 107 MB. A re-run on
2026-11-04 into a scratch prefix took 250 s, exited 0, printed
`<prefix>/bin/perl` and reported `v5.45.0-68-g94e5086608`; the binary differs
byte for byte from `/opt/perl-94e50866-debug/bin/perl` (the prefix is
embedded), so compare `perl -v`, not a checksum.

## perly-alternatives.sh

`sh internal/chart/offline/perly-alternatives.sh /opt/corpora/perl5` runs
bison 3.8.2 on the tree's `perly.y` in a temporary directory and prints one
line per alternative, `perly.y:<line> <lhs>: <rhs>`, where `<line>` is
bison's `yyrline` for the alternative. It refuses any other bison. Its output
is committed as `internal/chart/grammar/testdata/perly/alternatives.txt`, the
numbering the port ledger and the family issues cite. It needs no perl and
takes about 1 s (1.1 s on 2026-11-04); the 348 lines it printed were
byte-identical to the committed file (`cmp`).

## perly-tokens.sh

`sh internal/chart/offline/perly-tokens.sh /opt/corpora/perl5` prints the
terminals perly.y declares, one per line: every name on the `%token`, `%left`,
`%right` and `%nonassoc` lines before the first `%%`, `<type>` tags stripped,
each once, in first-declaration order, unfiltered (the test keeps its own
exclusions). It refuses any perl5 HEAD but 94e50866 and needs no bison and no
perl. Its output is committed as
`internal/chart/grammar/testdata/perly/declared-tokens.txt`, which
`TestClassSlotsAreThePerlyYTokens` reads, and matches it byte for byte (`cmp`).

## referee/

Bison builds of the pinned `perly.y` (spec §8), the evidence for the
precedence category: a viable prefix without precedence, refused with it.

- `sh internal/chart/offline/referee/build.sh <perl5> <outdir>` refuses any `perl5` HEAD but 94e50866, then
  writes `referee-prec` and `referee-noprec` into `outdir`. `referee/strip.pl` makes
  their grammars: every action stripped except those calling `yyerror`, kept
  as refusals; `catch_paren: empty` omitted, as the port omits it;
  `referee-noprec` also drops every precedence declaration and `%prec` and is
  built with `%glr-parser`, keeping every derivation.
- Each referee reads `perly.y` terminal names, one per line, on stdin, and
  prints `accept`, or with `-p` (prefix mode) `viable`, else `refuse@N`, N the
  token being read at the first error. Any error refuses, even one `yyparse`
  recovers from through `sideff: error`.
- With `-g` a referee prints a second line, the **grouping**: for each
  reduction of a rule carrying a precedence (a `%prec`, or its last
  terminal's), `line:from-to`, the rule's `perly.y` line (as
  `perly-alternatives.sh` numbers it) and the half-open span of 0-based
  terminal indexes it covers (`GRAMPROG` is 0), comma-separated in reduction
  order up to the first error, or `-` if there are none. A rule whose action
  refuses (a refusal kept from `perly.y`) is traced before it refuses, so the
  grouping includes the reduction that makes the error. `referee-noprec`'s
  grouping of an ambiguous span holds every derivation's reductions.
- `sh internal/chart/offline/referee/check.sh /opt/corpora/perl5` builds into a temporary directory and checks every case
  in `referee/testdata/cases.txt` against both referees, and `referee-prec`'s
  grouping, failing on any difference, naming the case, and on zero cases.

`referee/build.sh` writes the two referees and their generated `.y` and `.c` files
into `outdir`; none is committed. `referee/check.sh` prints
`check.sh: 7 cases agree`.

Cost (K6): `check.sh` takes about 0.65 s wall clock on this machine (0.62–0.67 s
over three runs; 4 cores, bison 3.8.2, gcc 13.3), nearly all of it the two
bison and cc builds. On 2026-11-04, with other jobs running, it took 1.1 to
1.5 s over three runs (the first, cold, 5.8 s) and `build.sh` alone 0.9 s.

### The generated sentences (spec §8)

`sh internal/chart/offline/referee/sentences.sh /opt/corpora/perl5 <outdir>`
generates the precedence referee's sentences. It refuses, exiting 1, a tree
whose HEAD is not 94e50866 or that has changes (`git status --porcelain`),
and any bison but 3.8.2. It writes into `<outdir>`:

- `bison-conflicts.txt`: every `Conflict between rule N and token T resolved
  as …` line of bison's report of the pinned `perly.y`, verbatim, in report
  order. Bison 3.8.2's `-v` alone no longer prints them; the script runs
  `bison -v --report=solved`. Each line is one (rule, token) resolution, the
  unit of coverage; 1,347 at 94e50866.
- `resolutions.txt`: per resolution, in the same order, tab-separated: rule,
  token, action (`shift`, `reduce`, `error`) and the comma-separated ids of
  the sentences covering it, or `-`.
- `sentences.txt`: per sentence, tab-separated: id, its `perly.y` terminals,
  `referee-prec`'s verdict and its grouping (`-g`, above).
- `uncovered-reviewed.txt`: a proposal, the resolutions no sentence covers
  under `reason: generation timed out`.

**A sentence** is a string of terminals on which `referee-prec` and
`referee-noprec` differ: one accepts and the other refuses, or both accept
with different groupings. Two refusals are never compared by grouping, since
each stops at its own first error. **A resolution is covered** by a sentence
on which `referee-prec` (built again with `YYDEBUG` to trace its states)
enters the conflict's state with that token next, at the index k the search
put it, and the difference sits at that token: on a shift, `referee-noprec`'s
grouping reduces the rule to end at k and `referee-prec`'s does not; on a
reduce, `referee-prec`'s does and `referee-noprec`'s has a span across k that
`referee-prec`'s lacks; on an error, `referee-prec` refuses reading k. So
coverage is checked against both referees' output, not inferred from the
search. After writing, `sentences.sh` runs both referees again on every
sentence and exits 1 unless they differ. An error resolution is covered only together with an **accepted
neighbour**: a sentence `referee-prec` accepts that reaches the same state
with another token next, on which the referees also differ (`++$x++` beside
`++$x .. $y`).

**The search** (`referee/sentences.pl`) runs over the LR(0) automaton of the
referee grammar, the same with and without precedence: shortest paths over
(state, item) from `grammar`'s alternatives, `GRAMPROG`'s preferred, where a
step reads a symbol (its shortest yield) or descends into a rule, paying for
the rest of the rule it leaves. Its candidates for a resolution are the
state's shift items reading the token, and the rule reduced with the token
next, each completed by shortest yields, shortest first. Rules that refuse in
an action or read `error` stand in no yield. **The bound** is 12 candidates
per resolution, a count rather than a clock so that a run is reproducible. A
resolution whose search reaches the bound is proposed as `generation timed
out`. One whose candidates all fail before the bound fits no reason spec §8
allows: the script names it on stderr, lists it nowhere, and exits 1, a stop
to report to perigrin. At 94e50866 neither happens: every resolution is
covered, by 1,341 sentences.

Committed under `internal/chart/testdata/precedence/` are `bison-conflicts.txt`
and `resolutions.txt`, byte for byte as the script writes them; `sentences.txt`
with the rendering column below added; and `uncovered-reviewed.txt` once
reviewed. `TestPrecedenceFixturesWellFormed` reads them and never runs bison.

**The rendering and its round-trip filter** (spec §8). `testdata/precedence/rendering.txt`
gives one lexeme to each `perly.y` terminal the sentences use that has one
(`GRAMPROG`, a start selector and no token, renders to nothing). A sentence is
rendered by joining its terminals' lexemes with spaces, and kept only if the
rendering round-trips: `internal/lexer` then chart/lex read it as exactly one
token per terminal, each token's terminal set holding its terminal. A
rendering the lexer reads as another terminal (`ref (` as `FUNC1`), as more or
fewer tokens (two tokens for one terminal), or as tokens whose text is not the
lexemes (a boundary that shifted), is not kept. The lexer inserts no `;`, the
recognizer does, so the inserted-`;` case is modelled as a written `;`.
A terminal's row in `rendering.txt` may carry a third column, a **preamble**:
source the lexer needs before the word to read it as that terminal. `NCRELOP`
is `isa`, a word until the feature is on, so its preamble is `use feature
"isa";`. A sentence's preambles (each distinct one once) are written before its
words and are part of the stored rendering, but their own tokens are left out of
the comparison, which still requires one token per terminal, with exactly the
lexeme's text, after them. Without the preamble the same sentence is not kept.
`sentences.txt`'s fifth column is the kept rendering, or `-`. Every sentence
not kept is on `nonroundtrip-reviewed.txt` under a `reason:` line, the
cause `TestPrecedenceRenderingRoundTrips` recomputes: a terminal with no
single-lexeme rendering (the `PLUGIN_*` terminals until a library operator is
in scope; `POSTJOIN`,
which only an interpolated string has; `PERLY_DOLLAR`, which merges with the
word after it), or the terminal the lexer reads differently. (`UNIOPSUB`
renders as `foo`, any word, and `BLKLSTOP` as `any`.) At 94e50866 the filter
keeps 708 of the 1,341 sentences and drops 633, every one for a terminal with no rendering; none
fails the lexer's read of a rendering. A resolution is covered only by a kept
sentence, so the 631 whose sentences are all dropped are on
`uncovered-reviewed.txt` under `reason: the rendering cannot reach it`
(`TestPrecedenceUnreachableIsNonRoundTrip`), and 716 of 1,347 are covered.
`go test ./internal/chart/ -run TestPrecedenceRenderingRoundTrips -chart.update-rendering`
rewrites the rendering column and both lists from `rendering.txt`.

Cost (K6): 75 s to 4.5 min wall clock end to end on the 4-core machine,
depending on load from other workers (263 s on 2026-11-10, after the
coverage checks; 29 s of user time, 74 s of system time), most of it
starting the referees on the candidates. Two runs
wrote byte-identical files (`cmp`).

## tdt-capture.sh

`sh internal/chart/offline/tdt-capture.sh <perl> <file>` runs `<perl> -c -DpT
<file>` and prints the file's own toke.c token stream on stdout. Perl runs
under `env -i` with only `PATH` and `PERL5LIB` set, as `label.sh` runs it, so
`PERL5OPT` cannot add tokens; the Guacamole snippets need
`testdata/guacamole/stubs` on `PERL5LIB`. The snippets carry their preamble
in-file, so the script takes none. Perl's stdin is `/dev/null`.

The form, one token per line, in the order toke.c delivered them to the parser:

- `TYPE` or `TYPE<TAB>VALUE`. `TYPE` is the `perly.y` terminal name; `-DT`
  prints `?? N` for the few terminals toke.c's `debug_tokens` lacks
  (`PERLY_DOLLAR`, `PERLY_PAREN_CLOSE`, ...), and the script names them from the
  perl's own `CORE/perly.h`.
- `VALUE` is what `-DT` shows, first match wins: a constant's `SvPEEK`
  (`PV("greeting"\0)`, `IV(1)`), else the pending identifier toke.c resolved
  for the token (`$name`, `%h`), else the op (`op_concat`, `op_null` for a
  plain `=`). Integer `ival`s are dropped: most are addresses or line numbers
  that change between runs. `-DT` shows no source text.
- A `use` or `no` is followed by a `BAREWORD` with no value: toke.c's
  `S_force_version` always forces a `BAREWORD` for the version slot, its
  `opval` null when no version is written (`use FindBin;` is `KW_USE_or_NO`,
  `BAREWORD`, `BAREWORD PV("FindBin"\0)`, `PERLY_SEMICOLON`); the
  `use` rule at `perly.y:714` reads the two as `version` and `module`.
- Tokens reported under `### forced token:` are skipped there; they print
  where toke.c delivers them. `EOF` is not printed; the `;` toke.c inserts
  before a `}` and at end of file are.
- A file perl refuses ends with the line `# refused: perl exited N`, and the
  script exits 1. A syntax error stops toke.c, so the stream ends at the token
  perl refused; a deferred error (strict, "Can't modify") leaves the whole
  stream before the marker.
- A file whose perl runs past the timeout ends with `# timeout: perl ran past
  N s`, and the script exits 1 like a refused capture (perigrin: "Timeout, mark
  refused"). The bound is `TDT_TIMEOUT` seconds, 60 by default: over six times the
  slowest corpus file measured below. Status 124 is read as the timeout only
  while a unit is still open; a file that itself exits 124 (from `BEGIN`) has
  closed its units and is refused.
- A file perl stops with status 0 before its `EOF` (an `exit` in `BEGIN`, as
  `plan skip_all` does) ends with `# truncated: the -DT stream stopped before
  the file's EOF (perl exited 0)`, and the script exits 2. A stream whose last line starts with
  `#` is never a whole one.
- Exit 2 otherwise is a capture failure with nothing useful on stdout: a
  missing file, a perl without `-DDEBUGGING` (which warns about `-D` and runs
  on, so it would otherwise yield an empty stream), lost bracketing, or perl
  never running: status 125 to 127 with no unit begun is `timeout` or `env`
  failing, and stderr carries their own message.

Only the first compilation unit's tokens print. Each unit (a loaded file, or a
string `eval` run at compile time, as modules do) starts where `-Dp` prints
`lookahead token is token GRAMPROG` and ends where `-Dp` prints `clearing the
parse stack`, which perl prints whenever a parse ends, accepted or aborted, so
a compile-time eval that fails to compile still closes; units nest. The file's
own unit closing before its `-DT` `EOF` with perl's status 0 is a truncation.
`-DT` alone marks no start: a file starts at line 0 but a string eval at line
1 like any line, which is why `p` is on. Units run after the main file's `EOF`
(`CHECK` blocks) are dropped, and a `require` in a sub that `-c` never runs
contributes nothing. A file perl accepts whose units do not all end is
exit 2, never a silently short stream. A sub-parse an XS keyword plugin runs
also prints `clearing the parse stack`, so such a file ends with `a parse ended
outside any unit` and exit 2.

`testdata/tdt-sample.pl` and its hand-checked `testdata/tdt-sample.expected`
cover a `use` of a stub module (`testdata/tdt-stubs/TdtStub.pm`, itself
loading `strict` and running a string eval in `BEGIN`; its `TDT_SENTINEL_word`
must not appear), `use lib` with `FindBin` (real modules), a variable, `qq{}`
interpolation, a `=>` word, an interpolating heredoc, a `require` inside a
sub, and a `BEGIN` string eval that fails to compile followed by a statement
that must still appear.

Why the timeout: perl `-Dp` spins without end once `BEGIN` puts an
`:encoding` layer on STDERR (`perl -Dp -e 'BEGIN { binmode STDERR,
":encoding(UTF-8)" }'` never returns, Encode preloaded or not; `-DT` alone
returns). T1's `io_pipe.t` and `io_layers.t` do this and end in `# timeout`.
`testdata/tdt-hang.pl` is the smallest such file; this check, with a short
bound, shows the capture ends it with the marker and a non-zero exit:

```sh
sh -c 'out=$(TDT_TIMEOUT=2 sh internal/chart/offline/tdt-capture.sh /opt/perl-94e50866-debug/bin/perl internal/chart/offline/testdata/tdt-hang.pl 2>/dev/null); s=$?; test "$s" -ne 0 && printf "%s\n" "$out" | tail -1 | grep -q "^# timeout: "'
```

`testdata/tdt-exit124.pl` exits 124 from `BEGIN`, timeout's own status; this
check shows it is refused, not called a timeout:

```sh
sh -c 'out=$(sh internal/chart/offline/tdt-capture.sh /opt/perl-94e50866-debug/bin/perl internal/chart/offline/testdata/tdt-exit124.pl 2>/dev/null); s=$?; test "$s" -eq 1 && printf "%s\n" "$out" | tail -1 | grep -qx "# refused: perl exited 124"'
```

With no `timeout` on `PATH`, `env` cannot run perl; this check shows the
capture exits 2 with nothing on stdout and names the cause:

```sh
sh -c 'd=$(mktemp -d); for t in awk grep env; do ln -s "$(command -v $t)" "$d/$t"; done; out=$(PATH=$d /bin/sh internal/chart/offline/tdt-capture.sh /opt/perl-94e50866-debug/bin/perl internal/chart/offline/testdata/tdt-sample.pl 2>"$d/err"); s=$?; e=$(cat "$d/err"); rm -rf "$d"; test "$s" -eq 2 && test -z "$out" && printf "%s" "$e" | grep -q timeout'
```

Cost (K6), measured 2026-10-31 on the 4-core machine, files one at a time,
run as `label.sh` runs perl: in-tree files from the repository root with
`internal/chart/testdata/guacamole/stubs` on `PERL5LIB`, T1 files from the
PerlOnJava root as `src/test/resources/unit/<file>` with `PERL5LIB` empty.
The sample takes 0.5 s; `std5k.pl` 2.2 s and `edit5k.pl` 1.6 s; the 893
Guacamole snippets 106 s in all (0.12 s each). T1's 986 top-level `.t` files
take 2,686 s in all (2.7 s each on average), 120 s of it the two timeouts;
the slowest other file takes 8.9 s, almost all of it perl loading
`Test::More` and friends under `-DpT`. Over T1, 831 captures are whole, 121
refused, 2 timed out and 32 truncated by an `exit` in `BEGIN`; the in-tree
files give 827 whole and 68 refused.

perl.git `t/`, measured 2026-11-12 the same way from the shim's `t/` with
`PERL5LIB` empty, other workers running: 619 of the 620 `.t` files take 369 s
in all (0.6 s each); 575 captures are whole, 31 refused, 13 truncated by an
`exit` in `BEGIN`, none timed out.

**`re/pat_psycho.t` is not captured** (perigrin, 2026-11-12: "Exclude it from
capture"). It calls `watchdog()` in `BEGIN`, which forks, so the child's
`-DpT` trace interleaves with the parent's on the shared stderr. A capture
then either fails (`unexpected token report`, exit 2) or ends whole but holds
the child's tokens, and either way waits on the pipe for the watchdog's 300 s.
`testdata/tdt/perlgit-uncaptured.txt` lists it; `tdt-align.sh` captures no
file on that list, `label.sh` keeps its `perl -c` label (`accept`) without
reading a capture, and `TestPerlGitUncaptured` holds the list to that one
file.

## shim/

`go run ./internal/chart/offline/shim <perl> <perl5-tree> <dest>` builds the
tree perl.git `t/` compiles in, for `tdt-align.sh` and `label.sh`, which run
it into a temporary directory with `$PERL5_CORPUS` (default
`/opt/corpora/perl5`) as the tree. `t/test.pl` clears `@INC` and unshifts
`../lib`, so a perl.git test compiles only from a `t/` whose `../lib` holds a
built library. It calls `parseoracle.BuildShim`, the builder the oracle facts
use (`offline/oracle`), so the labels, the captures and the oracle facts are
measured in one shim, and refuses a tree not at 94e50866 or not clean,
exiting 1 with the tree and the reason on stderr. What the shim lacks
(`TestInit.pm`, `Maintainers.pm`, XS::APItest, Win32, files outside `t/` and
`lib/`) shows as environment labels. Cost (K6): 0.9 s wall clock with `go
run`, and 155 MB of temporary disk (2026-11-12).

## tdt-align.sh

`TDT_FIXTURES=<dir> T1_CORPUS=<unit dir> [PERL5_CORPUS=<dir>] sh internal/chart/offline/tdt-align.sh <perl>`
captures every waypoint-1 corpus file, every perl.git `t/` file but those
on `testdata/tdt/perlgit-uncaptured.txt` (above), and the committed sample, with `tdt-capture.sh` into a temporary directory, run as
`label.sh` runs perl (in-tree files from the repository root with the
Guacamole `stubs/` on `PERL5LIB`, T1 files from the PerlOnJava root with
`PERL5LIB` empty, perl.git `t/` files from a shim `t/` with `PERL5LIB` empty;
see `shim/` below), then runs the aligner, `go run
./internal/chart/offline/tdt/align <capture-dir> <t1-dir> <perlgit-t-dir>
<out-dir>`, which reads perl.git sources from `$PERL5_CORPUS/t`. A capture
failure (exit 2 without a `# truncated` line) stops the run. The aligner
writes one fixture per corpus file to `$TDT_FIXTURES/<key>.tdt`, the key
being the repository path, `T1/<name>.t`, or `perlgit/<path under t/>`
(`perlgit/op/sub.t`), removing the old `*.tdt` files there and nothing else, and prints each file's unmatched
toke.c runs and a total to stdout. The sample's fixture becomes the committed
`testdata/tdt/sample.tdt`. The script makes `TDT_FIXTURES` absolute before it
changes directory and refuses any directory inside the repository or holding
it, resolving it through its nearest existing parent so a refused directory is
never created; the aligner refuses a directory holding `go.mod` or `.git`. Give `TDT_FIXTURES`
as an absolute path: `TestTDTFixtures` reads a relative one from
`internal/chart/`.

**The corpus fixtures are generated offline, not committed** (perigrin:
"Don't commit fixtures"). They come to about 16 MB for the 1,881 files, and
three are over the pre-commit hook's 500 KB limit on an added file (T1's
`code_too_large.t` 2,831 KB, `std5k.pl` 811 KB, `edit5k.pl` 719 KB). Unlike
T1, which is in the tree, `TDT_FIXTURES` names an input that lives outside it. This
milestone generates them with:

```sh
TDT_FIXTURES=/var/tmp/pvm-chart-tdt T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit sh internal/chart/offline/tdt-align.sh /opt/perl-94e50866-debug/bin/perl
```

`TestTDTFixtures` checks the fixtures in `TDT_FIXTURES`, perl.git `t/`'s
among them where `$PERL5_CORPUS/t` exists; it never runs perl or the aligner. With `TDT_FIXTURES` unset it skips, so `make test` and CI
pass without the debug perl; a `TDT_FIXTURES` that names no directory fails.
Committed are the aligner (`tdt/align/main.go`, checked by `tdt/align/main_test.go`), this script, `testdata/tdt/fitting-families.txt`
and `testdata/tdt/sample.tdt`, the sample's golden alignment, which
`TestAlignSample` compares a fresh alignment of the sample against.

The alignment, spec §12's:

- pvm's tokens are `lexer.Tokenize`'s less whitespace, comments, POD and
  `__END__`/`__DATA__`, which toke.c delivers nothing for.
- The toke.c stream is first grouped by the fitting families in
  `testdata/tdt/fitting-families.txt`: the tokens toke.c splits, invents or
  sublexes are folded onto the one token carrying the pvm token's text (the
  head), so `$name`'s `PERLY_DOLLAR PRIVATEREF` is one group matching pvm's
  Variable.
- Groups are matched to pvm tokens **in order by text**: a head's text is its
  punctuation, its op's operator or keyword, or its `PV`; a literal
  (`THING`, `QWLIST`) pairs with any pvm literal, and a sub toke.c resolved by
  scope with any word, since their values do not give the source text. At a
  mismatch the side that reaches a match sooner is skipped; when both reach
  one at the same distance neither is paired, so nothing pairs by reordering.
- A skipped group that is an invented token (`NOAMP`, the `&` before a
  called sub, the `sub` before a phaser) folds onto the next pair; any other
  is an **unmatched toke.c run**, reported per file. Module tokens leaking past
  the capture's bracketing show as such a run, as does a source filter's
  rewriting.
- A constant sub called with an empty list, `NAME()`, is one toke.c token
  (yyl_just_a_word folds the `()` in, toke.c:8202), so the pvm `(` and `)`
  after a word paired with a constant-sub BAREWORD are passed over rather than
  matched against the tokens after it.
- **TDTALIGN amendment 2** (perigrin, 2026-11-06, "Make amendment 2"): pvm
  tokens left unaligned for five reasons each carry a family, so
  `tdt-unaligned-unlisted` counts only the unexplained: `constant-parens`
  (the folded `()` above), `constant-string` (a word whose constant's string
  value toke.c delivers in place of its name), `lexical-sub` (the `my`, `our`
  or `state` before `sub`, and an `our sub`'s package-qualified name), and,
  per file in the aligner's `relabeled` list, `source-filter` and
  `attribute-resync`. Each family's reason is its line in
  `fitting-families.txt`.
- **TDTALIGN amendment 3** (perigrin, 2026-11-06, "Allow amendment 3"):
  five matching rules, each pairing a token the aligner left unaligned the
  way toke.c spells it, no family added: PERLY_TILDE pairs with `~.` as well
  as `~` (yyl_tilde, toke.c:7136-7138, under the bitwise feature); a
  BAREWORD's PV pairs with a word spelling its `::` as the old `'`
  (`Legacy'Package`); a variable's name drops only a `main::` or `::`
  qualifier, so `$main_regex` keeps its name; whitespace may stand between
  a variable's sigil and its name (`$ '`); and a heredoc body with no
  waiting pvm opener pairs with the interpolated string on its marker line
  that holds `<<`, whose fold carries toke.c's tokens for it (toke.c:11616).
  `v97`, the sixth unaligned token, is a lexer root instead: the lexer reads
  a v-string with no dot as a word (`corpus-covers-t2` `01a0dc72-2609`), so
  it is on `testdata/tdt/lexer-rooted.txt`. No further amendment of the
  definition is allowed.
- A refused, timed-out or truncated capture is not aligned (perigrin:
  "Timeout, mark refused"); its fixture holds only the capture's marker.

A fixture opens with `# tdt: <key>` and `# capture: whole`, or the marker,
then one line per pvm token in file order, spans as half-open byte ranges:

- `<start> <end> <TYPE>`: paired by text with one toke.c token;
- `<start> <end> <TYPE> <family> <TYPE>...`: paired through a family, the
  types after the family folded onto it; a family with no types after it
  paired the token through that family with nothing folded;
- `<start> <end> -` or `<start> <end> - <family>`: unaligned, or unaligned
  but explained by a family (a label's `:`, `return`, which `-DT` never
  reports);
- `unmatched <byte> <TYPE>...`: an unmatched toke.c run, before the pvm token
  at that byte.

`testdata/tdt/sample.tdt` is the sample's alignment in this form, the one
the gate review reads.

Cost (K6), measured 2026-11-02 on the 4-core machine, files one at a time,
the committed script run end to end: capturing all 1,881 files and the sample
takes 2,300 s and aligning them 1.0 s (2,302 s wall clock in all). Of the
corpus captures, 1,658 are whole and aligned, 189 refused, 2 timed out and 32
truncated, as TDTCAP's sweep found. 376,434 of 378,604 pvm tokens in the whole
captures are aligned; of the 2,170 unaligned, 1,965 are `return`, 84 a
label's `:` and 13 attribute words, and 108 are in no family. There are 37
unmatched toke.c runs. A re-run on 2026-11-04 with a freshly built perl and a
scratch `TDT_FIXTURES` took 2,198 s end to end and printed `aligned 1882
files: 376482 of 378652 pvm tokens aligned, 37 unmatched toke.c runs, 0.8 s`
(the 1,881 files and the sample); its 1,881 fixtures (the script moves the
sample's out to `testdata/tdt/sample.tdt`) matched `/var/tmp/pvm-chart-tdt`
byte for byte (`diff -r`) and `testdata/tdt/sample.tdt` was unchanged.

With perl.git `t/` (2026-11-12, other workers running, files
one at a time): capturing the 1,881 waypoint-1 files, 619 perl.git files and
the sample took 2,031 s and aligning them 3.8 s (2,062 s wall clock in all),
printing `aligned 2501 files: 880180 of 884314 pvm tokens aligned, 329
unmatched toke.c runs, 3.8 s`. The waypoint-1 fixtures matched
`/var/tmp/pvm-chart-tdt` byte for byte (`diff -r`), and
`testdata/tdt/sample.tdt` was unchanged. The perl.git fixtures come to 16 MB
(32 MB for all 2,500), none over 500 KB, and are not committed either. Their
tdt findings, measured by a scratch run of the ratchet's measures, are in
the commit that lists the lexer-rooted ones on `testdata/tdt/lexer-rooted.txt`
as `perlgit/<path>`.

## label.sh

Build the referees somewhere that persists first, for example
`sh internal/chart/offline/referee/build.sh /opt/corpora/perl5 /var/tmp/pvm-chart-referee`,
and pass that directory as `<referee-dir>`.

`T1_CORPUS=<unit dir> [PERL5_CORPUS=<dir>] sh internal/chart/offline/label.sh <perl> <referee-dir>`
labels every corpus file with the pinned perl's `perl -c` verdict
(spec §9.1) and writes `testdata/labels/in-tree.txt` (the Guacamole `parses`
and `parsent` cases, `std5k.pl`, `edit5k.pl`), `testdata/labels/t1.txt`
(T1's 986 top-level `*.t`) and `testdata/labels/perlgit.txt` (perl.git
`t/`'s 620 `*.t` at 94e50866, by path under `t/`, `op/sub.t`; the ratchet
names them `perlgit/<path>`). `<referee-dir>` holds `referee/build.sh`'s two
referees; `TDT_FIXTURES` (default `/var/tmp/pvm-chart-tdt`) holds
`tdt-align.sh`'s captures. It refuses a perl that is not a `-DDEBUGGING` build
at 94e50866, a T1 checkout not at PerlOnJava 73be162b, and (through
`shim/`) a perl5 tree not at 94e50866 or not clean.

- Each file runs under `env -i` with only `PATH` and `PERL5LIB` set and a 30 s
  timeout. In-tree files run from the repository root with the Guacamole
  `stubs/` on `PERL5LIB`; T1 files run from the PerlOnJava root, where their
  `use lib 'src/test/resources/unit/lib'` resolves, with `PERL5LIB` empty;
  perl.git `t/` files run from the shim's `t/` (`shim/`), where
  `t/test.pl` finds `../lib`, with `PERL5LIB` empty and `# cwd: .` in the
  header.
- Each label file opens with a provenance header (`# key: value`): the perl,
  its version, `git_commit_id` and `ccflags`, the corpus, working directory,
  root, `PERL5LIB`, timeout, the first-error limit, the status and how the
  review evidence is taken.
- Then one tab-separated line per file: `<path> accept`, or
  `<path> refuse <category> <status> <first error> [<review>]`. The category
  is one of grammar, precedence, scope, semantic, environment; the status is
  `proposed` or `reviewed`. A timeout is labelled environment with the timeout
  as its evidence, never omitted.
- The category is proposed from perl's first error, its first stderr line,
  by the first of the script's patterns it matches. A first line no pattern
  matches stops the script; a later line never decides the category.
- A grammar refusal's `-DT` stream (`tdt-capture.sh`) goes to both referees:
  `referee-prec` (perl's grammar) on the whole stream, then `referee-noprec`
  on the stream up to the token `referee-prec` refuses. The review field
  records both, and the label is **precedence** exactly when that prefix is
  `viable` without precedence: `referee-noprec viable; referee-prec
  refuse@N`. A toke.c croak leaves a stream both find `viable`.
- A file `perl -c` accepts whose capture in `TDT_FIXTURES` is `# truncated`
  stopped at a skip (`plan skip_all`, or a TAP skip and `exit`) in `BEGIN`
  without compiling the rest; it is refused as environment (perigrin's ruling
  "Label environment"), the capture its review field.
- Status 124 is read as a timeout by status alone (a `shortcut:` in the
  script): no label carries timeout evidence, so no corpus file exits 124.

**The review.** label.sh writes every category `proposed`; a person or agent
checks each against its evidence and marks it `reviewed`, and
`TestPerlLabels` refuses any refusal still `proposed`, so a rerun is reviewed
again before it lands. LABELS-review (2026-11-03) reviewed all of them:

| Category | In-tree | T1 | Evidence |
|---|---|---|---|
| grammar | 52 | 0 | 33 streams both referees refuse at the same token (30 syntax errors, 3 toke.c croaks: version format, anonymous sub); 13 toke.c croaks on a `viable` stream (string terminator, bare `<<`, digits); 6 `$#`/`$*` |
| precedence | 0 | 0 | no prefix is viable without precedence and refused with it |
| scope | 0 | 0 | no action refused under scope or features |
| semantic | 11 | 1 | op.c `Can't modify` (1); `Invalid CODE attribute` after the sub is parsed (10); a `strict subs` bareword (T1, 1) |
| environment | 5 | 152 | `no 5.010` too modern (5); no threads (59); missing modules (56); `-T` (5); a skip in `BEGIN` (32) |

One proposal was overturned: `$# / $* is no longer supported` was proposed
semantic and is grammar. `-DpT` shows gv.c croaking from toke.c's
`S_pending_ident`, after `PERLY_DOLLAR` and before the name token reaches the
parser; the referees find the stream `viable`. That is a fatal lexing error,
§9.1's grammar, not a check after parsing.

PERLGITLABELS (01a1c324-1b3f, 2026-11-11) reviewed perl.git `t/`'s 44
refusals the same way; 576 of the 620 files are accepted:

| Category | perl.git `t/` | Evidence |
|---|---|---|
| grammar | 1 | `comp/final_line_num.t`: both referees refuse its stream at the same token |
| precedence, scope, semantic | 0 | none |
| environment | 43 | modules the shim lacks: `TestInit.pm`, `Maintainers.pm`, XS::APItest, Win32 (24); `-T` or `-t` on the `#!` line (5); a `./` or `../` require outside `t/` (2); a skip in `BEGIN`: no ithreads, no dtrace, Windows or cygwin only, `encoding.pm` gone, XS::APItest or `PERL_RC_STACK` absent (12) |

One proposal was overturned: `comp/final_line_num.t` was proposed
environment as a skip in `BEGIN` (its capture is truncated), but its
`$SIG{__DIE__}` exits 0 on the syntax error at EOF the file provokes on
purpose. Printing `$_[0]` in a copy recovers the error, `syntax error at
comp/final_line_num.t line 13, at EOF`, and both referees refuse its `-DT`
stream at the same token: grammar. A rerun proposes environment again.

`TestPerlLabels` reads the labels; it never runs perl. Its perl.git subtest
reads the `.t` files under `$PERL5_CORPUS/t` (default `/opt/corpora/perl5`)
and skips where that tree is absent.

Cost (K6): one run labels 1,881 files in about 2.5 min wall clock on this
machine (147 s, 4 cores, files one at a time, other workers running; 42 s
before the review evidence was added; 122 s on 2026-11-04).

With perl.git `t/` (2026-11-12, against fixtures from the run above and
freshly built referees) one run labels 2,501 files in 367 s. Against the
committed files, the only differences were the `# status:` header,
`reviewed` becoming `proposed`, and `comp/final_line_num.t`, which the run
proposes environment and the review overturned (above); `re/pat_psycho.t`
stays `accept` with no capture read.

The run overwrites all three committed label files with every category
`proposed`: on 2026-11-04, against the fixtures in `/var/tmp/pvm-chart-tdt` and freshly built referees,
the only differences from the committed files were the `# status:` header and
`reviewed` becoming `proposed` on 221 lines (68 in-tree, 153 T1). Restore the reviewed files with
`git checkout -- internal/chart/testdata/labels`, or review the rerun again.
`TestPerlLabels` checks the result.

## oracle-facts.sh

`T1_CORPUS=<unit dir> sh internal/chart/offline/oracle-facts.sh <perl> <perl5-tree> <outdir>`
asks the pinned perl for its per-file parse facts (`parseoracle`'s
`parse_facts.pl`) over the oracle-graded corpora and writes `in-tree.txt`,
`t1.txt` and `perlgit.txt` into `<outdir>`; commit them as
`internal/chart/testdata/oracle/`. The ratchet then grades without spawning
perl (a live sweep is about six minutes, over its 300 s budget). It refuses a
perl that is missing or not the `-DDEBUGGING` build at 94e50866, a perl5 tree
not at 94e50866 or not clean, and a T1 checkout not at PerlOnJava 73be162b or
not clean, exiting 1 with its name, the tree and the reason on stderr.

- The oracle-graded corpora are the files perl accepts by the committed
  labels (the in-tree Guacamole snippets, `std5k.pl` and `edit5k.pl`: 827;
  T1: 833), and every `.t` file of perl.git `t/` (620), refused or not; for
  perl.git, `perlgit.txt`'s labels say which perl accepts.
- Each file starts with the label files' provenance header: the perl, its
  version, `git_commit_id` and `ccflags`, the corpus, working directory, root,
  `PERL5LIB`, the environment and the timeout (180 s per file).
- Then one tab-separated line per file, keyed `<path>` (in-tree, from the
  repository root), `T1/<name>.t` or `perlgit/<path under t/>`:
  `<key> facts <json>` with `ok`, `walked`, `op_count`, `sites` and
  `prototypes`, the fields `parseoracle.CompareFacts` reads (not the op
  sequence); or `<key> none <reason>` for a file perl produced no facts for.
  `t/re/pat_psycho.t` is the one `none` (its watchdog outlives the timeout);
  `ReadOracleFacts` never reads it as accepted, and fails on a truncated or
  malformed line.
- perl.git `t/` compiles only from a shim `t/` whose `lib/` holds the perl's
  own library, which `parseoracle.BuildShim` builds in a temporary directory.
- `T1/cpan_tooling_regressions.t` is accepted by its label (`perl -c`) but
  its facts say `ok` false (`-MO=Concise` fails on it); the record keeps perl's
  facts and the script reports the disagreement on stderr. The same holds
  for `perlgit/re/pat_psycho.t`, accepted by its label and recorded `none`;
  perl.git's 26 other records with `ok` false are files its labels refuse.
  `TestOracleFactsLabelAcceptedWithoutOKFacts` pins both.

Cost: 7 m 43 s wall clock on the 4-core development machine under load
average about 13 (3 m 34 s user), most of it perl.git `t/`. Commit the output
only from a run that exits 0.
