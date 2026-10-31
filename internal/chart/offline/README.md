<!-- ABOUTME: The chart spike's offline scripts: run by hand, never by make test or the ratchet. -->
<!-- ABOUTME: Covers the bison referees of the pinned perly.y and what each script costs to run. -->

# internal/chart/offline

Scripts run by hand against perl blead 94e50866 in `/opt/corpora/perl5`.
Nothing here runs in `make test` or `make chart-ratchet`; their output, where
they have any, is committed and read by the Go tests.

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

Cost (K6): `check.sh` takes 1–3 s wall clock on this machine (4 cores,
bison 3.8.2, gcc 13.3), nearly all of it the two bison and cc builds.
