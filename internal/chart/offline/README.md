<!-- ABOUTME: The chart spike's offline scripts: run by hand, never by make test or the ratchet. -->
<!-- ABOUTME: Covers the pinned -DDEBUGGING perl build and the bison referees, and what each costs to run. -->

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
