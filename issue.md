---
title: for $i (LIST) without my swaps its variable and its list at Unknown=0
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-27T01:18:44.077340739Z
updated: 2026-09-27T01:18:44.077340739Z
---

`for $i (LIST)` without `my` SWAPS its variable and its list, at Unknown=0. The
variable is emitted as the list and the list as a bare term.

Found by tracing the `}` bucket of perl.git t/ at `e87dee9c` (11 files, 222
nodes) to its MATCHING opener with a depth counter, then building the shape up
from a minimal case.

## Measured at `e87dee9c`

    source                                     U   canon
    for my $t (1) { for $i (2) { 3 if 1; } }   0   for my $t (1) {for ($i) 2 {3 if 1;} }
    for $i (2) { 3; }                          0   for ($i) 2 {3;}

The variable `$i` is emitted INSIDE the parens where the list belongs, and the
list `2` is emitted bare after them. The `my` spelling is correct, which is why
this survived every earlier pass: `for my $i (2)` canons as `for my $i (2)`.

That is the thirteenth zero-Unknown-and-wrong finding in this parser. The count
cannot see it; only the canon can.

## Perl keeps the order

    perl -MO=Deparse -e 'our $i; for $i (1,2) { print $i }'
      ->  foreach $i (1, 2) { print $i; }

Variable first, list in the parens. The construct is the C-styleless `foreach`
with an explicit package variable rather than a lexical, and perl's own suite
uses it wherever the loop variable must outlive the loop or be `local`ised.

## Scope

1. The `for`/`foreach` parse takes a `my`-declared variable and a bare variable
   down different paths; the bare path assigns them to the wrong slots. Find the
   site and make the bare spelling fill the same slots as the `my` spelling.
2. Measure `for our $i (...)`, `for local $i (...)` and `for $pkg::i (...)` --
   each may take a third path.
3. A corpus case asserting the CANON, since Unknown=0 passes today.
4. The `}` bucket this was found in is 11 files; do NOT assume this defect owns
   them. Its siblings (`01a0e04d`, filed alongside) came out of the same trace,
   and the bucket's openers are ordinary blocks -- the causes are inside them,
   one file at a time.

## Context

- paths: `internal/parse/control.go` (the loop parse), `internal/parse/canon.go`
  (`Loop` emission)
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- related: `01a0dfc7` (the six-keyword modifier canon defect, same file),
  `01a0e04d` (a bare block inside a loop lets a modifier escape it)

## Acceptance criteria

- [ ] `for $i (2) { 3; }` canons with the variable before the parens (`go test ./internal/parse/ -run '^TestForLoopBareVariable$' -count=1 -v | grep -q '^--- PASS: TestForLoopBareVariable'`)
- [ ] The `my` spelling is unchanged (`go test ./internal/parse/ -run '^TestForLoopBareVariable$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the C-style three-part `for` still parses (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
