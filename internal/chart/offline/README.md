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

## referee/

Bison builds of the pinned `perly.y` (spec §8), the evidence for the
precedence category: a viable prefix without precedence, refused with it.

- `referee/build.sh <perl5> <outdir>` refuses any `perl5` HEAD but 94e50866, then
  writes `referee-prec` and `referee-noprec` into `outdir`. `referee/strip.pl` makes
  their grammars: every action stripped except those calling `yyerror`, kept
  as refusals; `catch_paren: empty` omitted, as the port omits it;
  `referee-noprec` also drops every precedence declaration and `%prec` and is
  built with `%glr-parser`, keeping every derivation.
- Each referee reads `perly.y` terminal names, one per line, on stdin, and
  prints `accept`, or with `-p` (prefix mode) `viable`, else `refuse@N`, N the
  token being read at the first error. Any error refuses, even one `yyparse`
  recovers from through `sideff: error`.
- `sh internal/chart/offline/referee/check.sh /opt/corpora/perl5` builds into a temporary directory and checks every case
  in `referee/testdata/cases.txt` against both referees, failing on any differing
  verdict, naming the case, and on zero cases.

`build.sh` writes the two referees and their generated `.y` and `.c` files
into `outdir`; none is committed. `referee/check.sh` prints
`check.sh: 7 cases agree`.

Cost (K6): `check.sh` takes about 0.65 s wall clock on this machine (0.62–0.67 s
over three runs; 4 cores, bison 3.8.2, gcc 13.3), nearly all of it the two
bison and cc builds. On 2026-11-04, with other jobs running, it took 1.1 to
1.5 s over three runs (the first, cold, 5.8 s) and `build.sh` alone 0.9 s.

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

## tdt-align.sh

`TDT_FIXTURES=<dir> T1_CORPUS=<unit dir> sh internal/chart/offline/tdt-align.sh <perl>`
captures every waypoint-1 corpus file, and the committed sample, with
`tdt-capture.sh` into a temporary directory, run as `label.sh` runs perl
(in-tree files from the repository root with the Guacamole `stubs/` on
`PERL5LIB`, T1 files from the PerlOnJava root with `PERL5LIB` empty), then
runs the aligner, `go run ./internal/chart/offline/tdt/align <capture-dir>
<t1-dir> <out-dir>`. A capture failure (exit 2 without a `# truncated` line)
stops the run. The aligner writes one fixture per corpus file to
`$TDT_FIXTURES/<key>.tdt`, the key being the repository path or
`T1/<name>.t`, removing the old `*.tdt` files there and nothing else, and prints each file's unmatched
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
`code_too_large.t` 2,831 KB, `std5k.pl` 811 KB, `edit5k.pl` 719 KB). Like
`T1_CORPUS`, `TDT_FIXTURES` names an input that lives outside the tree. This
milestone generates them with:

```sh
TDT_FIXTURES=/var/tmp/pvm-chart-tdt T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit sh internal/chart/offline/tdt-align.sh /opt/perl-94e50866-debug/bin/perl
```

`TestTDTFixtures` checks the fixtures in `TDT_FIXTURES`; it never runs perl
or the aligner. With `TDT_FIXTURES` unset it skips, so `make test` and CI
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
(the 1,881 files and the sample); its 1,882 fixtures matched the committed
`TDT_FIXTURES` byte for byte (`diff -r`) and `testdata/tdt/sample.tdt` was
unchanged.

## label.sh

`T1_CORPUS=<unit dir> sh internal/chart/offline/label.sh <perl> <referee-dir>`
labels every waypoint-1 corpus file with the pinned perl's `perl -c` verdict
(spec §9.1) and writes `testdata/labels/in-tree.txt` (the Guacamole `parses`
and `parsent` cases, `std5k.pl`, `edit5k.pl`) and `testdata/labels/t1.txt`
(T1's 986 top-level `*.t`). `<referee-dir>` holds `referee/build.sh`'s two
referees; `TDT_FIXTURES` (default `/var/tmp/pvm-chart-tdt`) holds
`tdt-align.sh`'s captures. It refuses a perl that is not a `-DDEBUGGING` build
at 94e50866, and a T1 checkout not at PerlOnJava 73be162b.

- Each file runs under `env -i` with only `PATH` and `PERL5LIB` set and a 30 s
  timeout. In-tree files run from the repository root with the Guacamole
  `stubs/` on `PERL5LIB`; T1 files run from the PerlOnJava root, where their
  `use lib 'src/test/resources/unit/lib'` resolves, with `PERL5LIB` empty.
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

`TestPerlLabels` reads the labels; it never runs perl.

Cost (K6): one run labels 1,881 files in about 2.5 min wall clock on this
machine (147 s, 4 cores, files one at a time, other workers running; 42 s
before the review evidence was added; 122 s on 2026-11-04).

The run overwrites both committed label files with every category
`proposed`: on 2026-11-04, against the committed `TDT_FIXTURES` and referees,
the only differences from the committed files were the `# status:` header and
`reviewed` becoming `proposed` on 223 lines. Restore the reviewed files with
`git checkout -- internal/chart/testdata/labels`, or review the rerun again.
`TestPerlLabels` checks the result.
