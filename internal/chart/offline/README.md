<!-- ABOUTME: The chart spike's offline scripts: run by hand, never by make test or the ratchet. -->
<!-- ABOUTME: Covers the pinned -DDEBUGGING perl build, the bison referees and the perl labels, and what each costs to run. -->

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
