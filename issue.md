---
title: Thirty-five parsing-distinctive constructs T1 uses that the corpus never names
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-22T03:37:32.481722445Z
updated: 2026-09-22T03:37:32.481722445Z
---

The corpus was built construct-first and has never been measured against
T1. Measured now, it reaches about a tenth of it and makes no claim at all
about 35 constructs whose PARSE is distinctive.

## The measurement

    T1       986 files  2,175,885 bytes  65,241 lines
    corpus   138 files    243,925 bytes   6,935 lines
                              11.2%          10.6%

File counts mislead -- a T1 file is a whole test program, a corpus file is
one construct -- so volume is the fairer comparison, and it is not the real
answer either.

**All twelve hard markers ARE exercised.** Every construct `easy_test.go`
used to mark a T1 file hard now has a tier claiming it and a corpus file
using it. That was the coverage question the milestone set itself and it is
met. But hard markers were only ever the KNOWN-hard set: 326 T1 files (33%)
carry one, 660 (67%) carry none.

Against T1's whole surface, **82 perl constructs T1 uses appear in no
corpus file**. Triaged, 47 are ordinary named operators that parse like any
list operator (`unlink`, `chdir`, `mkdir`, `fileno` ...) -- a corpus file
for those would assert nothing a hundred other calls do not, and their
absence is the corpus being graded by construct rather than by coverage.

**35 are parsing-distinctive**, and those are unmade claims.

## The 35, by T1 frequency

     316 (32.0%)  eval       block AND string form; the string re-enters the parser
     286 (29.0%)  defined    parenless unary with a special rule for &sub and %h
     218 (22.1%)  die        list operator whose argument extent is the classic ambiguity
     191 (19.4%)  undef      operator and lvalue; `undef $x` differs from `$x = undef`
     139 (14.1%)  push       first argument is an ARRAY, not an expression
     132 (13.4%)  can        the ->can(...)->() idiom
     126 (12.8%)  do         do BLOCK, do FILE, do SUB -- three unrelated parses
      74 ( 7.5%)  caller     context-sensitive return
      71 ( 7.2%)  values     aggregate-argument operator like keys
      69 ( 7.0%)  map        BLOCK LIST with no comma, or EXPR, LIST -- a parse fork
      44 ( 4.5%)  tie        TIE VARIABLE, CLASS, LIST
      44 ( 4.5%)  grep       same BLOCK/EXPR fork as map
      39 ( 4.0%)  ord
      38 ( 3.9%)  say        print-family with its own feature gate
      37 ( 3.8%)  tied
      34 ( 3.4%)  unpack     template mini-language
      34 ( 3.4%)  sprintf    format mini-language
      32 ( 3.2%)  substr     four-argument, lvalue-capable
      30 ( 3.0%)  split      first argument is a PATTERN, not an expression
      29 ( 2.9%)  pack       template mini-language
      29 ( 2.9%)  isa        an infix OPERATOR under 5.32+, not only a method
      29 ( 2.9%)  chr
      28 ( 2.8%)  warn       same extent question as die
      26 ( 2.6%)  time       niladic -- parses with no argument at all
      25 ( 2.5%)  each
      23 ( 2.3%)  select     one-arg and four-arg forms with unrelated meanings
      23 ( 2.3%)  qq         quote-like with arbitrary delimiters
      23 ( 2.3%)  prototype  reads a prototype -- the ($) hazard tier 07 measured
      17 ( 1.7%)  index
      16 ( 1.6%)  pos        lvalue-capable, tied to the regex engine
      15 ( 1.5%)  unshift    same array-first rule as push
      14 ( 1.4%)  lock       bareword-argument operator
      14 ( 1.4%)  elsif      tier 06 has if/else and NO elsif
       7 ( 0.7%)  tr         quote-like taking TWO delimited regions, like s///
      40 ( 4.1%)  exit

## Two measurements that show these are not bookkeeping

**`eval` is two constructs wearing one keyword**, measured under 5.42.0:

    eval { 1 }     enter nextstate entertry nextstate const leavetry ...
    eval "1"       enter nextstate const entereval ...

Different ops entirely. And `entereval` is currently claimed by
`14_recursive` for `s///ee`, so plain `eval STRING` would COLLIDE with an
existing claim -- `claimonce_test.go` would reject it. Where `eval` lands is
a real placement question, not a filing detail.

**`do { 1 }` emits no op at all** -- it folds to a bare `const`. Same
lesson tier 01 records for `my $x = 1+2`: the optimiser erases the
construct, so only a token fact can catch a parser that mishandles it.

## What this issue is not

Not an argument that the corpus is wrong. It is graded by construct, the
spec scopes tiers to parsing surface rather than to T1 coverage, and the
gate it set itself is met. This is the gap between "every construct we
named is covered" and "every construct T1 exercises is named", which
nobody had measured.

`elsif` is the sharpest single instance: tier 06 covers `if`, `unless`,
`if/else`, and the corpus contains no `elsif` anywhere, while 14 T1 files
use it.

## Shape of the work

These do not belong in one issue. The 35 split across existing tiers --
`eval`/`do` to 06 or a new tier, `push`/`values`/`each` to 02, `map`/`grep`/
`sort BLOCK` to 07, `split`/`pos`/`tr` to 09, `sprintf`/`pack` to a
mini-language tier that does not exist -- and several raise the placement
question `eval` raises. A follow-up should decide placement first, then
write files tier by tier, exactly as the fourteen tiers were done.

Measured at a3a0ed0c with scripts in the session scratchpad; the method is
a regex over source rather than a parse, deliberately, since using our own
parser to decide what T1 contains would make the answer depend on the thing
being measured.
