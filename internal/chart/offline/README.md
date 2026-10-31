<!-- ABOUTME: The chart spike's offline scripts: run by hand, never by make test or the ratchet. -->
<!-- ABOUTME: Covers the perl build, the bison referees, the perl labels and the -DT capture, and what each costs to run. -->

# internal/chart/offline

Scripts run by hand against perl blead 94e50866 in `/opt/corpora/perl5`.

- They never run in `make test` or `make chart-ratchet`.
- Nothing they build (perls, referee binaries) is committed.
- Output that tests read is committed.

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
depends on load from other jobs). Installed size:
107 MB.

## perly-alternatives.sh

`sh internal/chart/offline/perly-alternatives.sh /opt/corpora/perl5` runs
bison 3.8.2 on the tree's `perly.y` in a temporary directory and prints one
line per alternative, `perly.y:<line> <lhs>: <rhs>`, where `<line>` is
bison's `yyrline` for the alternative. It refuses any other bison. Its output
is committed as `internal/chart/grammar/testdata/perly/alternatives.txt`, the
numbering the port ledger and the family issues cite.

## referee/

Bison builds of the pinned `perly.y` (spec §8), the evidence for the
precedence category: a viable prefix without precedence, refused with it.

- `build.sh <perl5> <outdir>` refuses any `perl5` HEAD but 94e50866, then
  writes `referee-prec` and `referee-noprec` into `outdir`. `strip.pl` makes
  their grammars: every action stripped except those calling `yyerror`, kept
  as refusals; `catch_paren: empty` omitted, as the port omits it;
  `referee-noprec` also drops every precedence declaration and `%prec` and is
  built with `%glr-parser`, keeping every derivation.
- Each referee reads `perly.y` terminal names, one per line, on stdin, and
  prints `accept`, or with `-p` (prefix mode) `viable`, else `refuse@N`, N the
  token being read at the first error. Any error refuses, even one `yyparse`
  recovers from through `sideff: error`.
- `check.sh <perl5>` builds into a temporary directory and checks every case
  in `testdata/cases.txt` against both referees, failing on any differing
  verdict, naming the case, and on zero cases.

Cost (K6): `check.sh` takes about 0.65 s wall clock on this machine (0.62–0.67 s
over three runs; 4 cores, bison 3.8.2, gcc 13.3), nearly all of it the two
bison and cc builds.

## label.sh

`T1_CORPUS=<unit dir> sh internal/chart/offline/label.sh <perl>` labels every
waypoint-1 corpus file with the pinned perl's `perl -c` verdict (spec §9.1)
and writes `testdata/labels/in-tree.txt` (the Guacamole `parses` and
`parsent` cases, `std5k.pl`, `edit5k.pl`) and `testdata/labels/t1.txt` (T1's
986 top-level `*.t`). It refuses a perl that is not a `-DDEBUGGING` build at
94e50866, and a T1 checkout not at PerlOnJava 73be162b.

- Each file runs under `env -i` with only `PATH` and `PERL5LIB` set and a 30 s
  timeout. In-tree files run from the repository root with the Guacamole
  `stubs/` on `PERL5LIB`; T1 files run from the PerlOnJava root, where their
  `use lib 'src/test/resources/unit/lib'` resolves, with `PERL5LIB` empty.
- Each label file opens with a provenance header (`# key: value`): the perl,
  its version, `git_commit_id` and `ccflags`, the corpus, working directory,
  root, `PERL5LIB`, timeout and the first-error limit.
- Then one tab-separated line per file: `<path> accept`, or
  `<path> refuse <category> <status> <first error> [<referee>]`. The category
  is one of grammar, precedence, scope, semantic, environment; the status is
  `proposed` or `reviewed`. A timeout is labelled environment with the timeout
  as its evidence, never omitted.
- The category is proposed from perl's first error, its first stderr line,
  by the first of the script's patterns it matches. A first line no pattern
  matches stops the script; a later line never decides the category. The script never proposes precedence. That label needs the
  referee field, `referee-noprec viable; referee-prec refuse@N`, which
  LABELS-review adds from the file's `-DT` stream.

`TestPerlLabels` reads the labels; it never runs perl.

Cost (K6): one run labels 1,881 files in about 42 s wall clock on this
machine (41.6 s; 4 cores, files one at a time).

## tdt-capture.sh

`sh internal/chart/offline/tdt-capture.sh <perl> <file>` runs `<perl> -c -DpT
<file>` and prints the file's own toke.c token stream on stdout. `PERL5LIB`
passes through from the environment (the Guacamole snippets need
`testdata/guacamole/stubs` on it); the snippets carry their preamble in-file,
so the script takes none. Perl's stdin is `/dev/null`.

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
- Tokens reported under `### forced token:` are skipped there; they print
  where toke.c delivers them. `EOF` is not printed; the `;` toke.c inserts
  before a `}` and at end of file are.
- A file perl refuses ends with the line `# refused: perl exited N`, and the
  script exits 1. A syntax error stops toke.c, so the stream ends at the token
  perl refused; a deferred error (strict, "Can't modify") leaves the whole
  stream before the marker.
- A file whose perl runs past the timeout ends with `# timeout: perl ran past
  N s`, and the script exits 1 like a refused capture (perigrin: "Timeout, mark
  refused"). The bound is `TDT_TIMEOUT` seconds, 60 by default: six times the
  slowest corpus file measured below.
- A file perl stops with status 0 before its `EOF` (an `exit` in `BEGIN`, as
  `plan skip_all` does) ends with `# truncated: perl exited 0 before the file
  reached EOF`, and the script exits 2. A stream whose last line starts with
  `#` is never a whole one.
- Exit 2 otherwise is a capture failure with nothing useful on stdout: a
  missing file, a perl without `-DDEBUGGING` (which warns about `-D` and runs
  on, so it would otherwise yield an empty stream), or lost bracketing.

Only the first compilation unit's tokens print. Each unit (a loaded file, or a
string `eval` run at compile time, as modules do) starts where `-Dp` prints
`lookahead token is token GRAMPROG` and ends at its `-DT` `EOF`, and units nest.
`-DT` alone marks no start: a file starts at line 0 but a string eval at line
1 like any line, which is why `p` is on. Units run after the main file's `EOF`
(`CHECK` blocks) are dropped, and a `require` in a sub that `-c` never runs
contributes nothing. A file perl accepts whose units do not all close is
exit 2, never a silently short stream.

`testdata/tdt-sample.pl` and its hand-checked `testdata/tdt-sample.expected`
cover a `use` of a stub module (`testdata/tdt-stubs/TdtStub.pm`, itself
loading `strict` and running a string eval in `BEGIN`; its `TDT_SENTINEL_word`
must not appear), `use lib` with `FindBin` (real modules), a variable, `qq{}`
interpolation, a `=>` word, an interpolating heredoc and a `require` inside a
sub.

Why the timeout: perl `-Dp` spins without end once `BEGIN` puts an
`:encoding` layer on STDERR (`perl -Dp -e 'BEGIN { binmode STDERR,
":encoding(UTF-8)" }'` never returns, Encode preloaded or not; `-DT` alone
returns). T1's `io_pipe.t` and `io_layers.t` do this and end in `# timeout`.
`testdata/tdt-hang.pl` is the smallest such file; this check, with a short
bound, shows the capture ends it with the marker and a non-zero exit:

```sh
sh -c 'out=$(TDT_TIMEOUT=2 sh internal/chart/offline/tdt-capture.sh /opt/perl-94e50866-debug/bin/perl internal/chart/offline/testdata/tdt-hang.pl 2>/dev/null); s=$?; test "$s" -ne 0 && printf "%s\n" "$out" | tail -1 | grep -q "^# timeout: "'
```

Cost (K6), measured 2026-10-31 on the 4-core machine while other jobs ran:
the sample takes 0.5 s; `std5k.pl` and `edit5k.pl` 0.9 s each; the 893
Guacamole snippets 60 s in all (0.07 s each); T1's other 984 top-level `.t`
files 1,966 s in all (2.0 s each on average, at most 10 s), almost all of it
perl loading `Test::More` and friends under `-DpT`. Over those, 819 captures are
whole, 133 refused and 32 truncated by an `exit` in `BEGIN`.
