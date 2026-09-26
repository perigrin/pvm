# Parsing perl.git t/ cleanly

**Goal.** Every `.t` file in perl.git's `t/` directory parses with zero Unknown
nodes.

**Baseline, measured 2026-09-26 at `59029a2c`:** 185 of 620 clean (29.8%),
7,745 Unknown nodes, 435 dirty files.

## This is not the corpus, and it is not T1

Three populations, and they have been conflated in conversation before:

| name | what it is | size | state |
|---|---|---|---|
| the graded corpus | `conformance/mdtest/*.md`, cases we wrote | 216 cases | **216 clean** |
| T1 | PerlOnJava's `unit/*.t`, top level only | 986 files | 554 clean (56.2%) |
| T2 core | perl.git `t/{base,cmd,comp,opbasic,class}` | 56 files | 22 clean (39.3%) |
| **perl.git t/** | **this goal**, walked recursively | **620 files** | **185 clean (29.8%)** |

`t1.ratchet`'s own header says T1 is PerlOnJava. T2 core is a 56-file subset of
this goal's 620, so today's 22/56 covers 9% of the target.

## Where the failures are

    dir        clean/total          dir        clean/total
    op            31/227            mro           52/73
    re            31/ 80            porting       15/37
    io             7/ 44            uni            4/30
    comp           6/ 25            run            6/28
    bigmem         8/ 12            base           6/ 9
    class          5/ 12            lib            4/10
    cmd            3/  5            win32          2/ 9
    opbasic        2/  5            perf           0/ 5

`op/` is 227 files with 31 clean -- 37% of the goal's population and 32% of its
dirty files. `mro/` is already 71% and `perf/` is 0 of 5.

## By refusal code

    trailing_tokens          5982 nodes   in 412 files
    not_a_term               1584 nodes   in 192 files
    unimplemented_statement   169 nodes   in  15 files
    chain_class_mismatch        5 nodes   in   2 files
    ternary_no_colon            4 nodes   in   3 files
    missing_operand             1 node    in   1 file

`trailing_tokens` in 412 of 435 dirty files means nearly every failure is a
statement that ended sooner than the source did, which is a symptom rather than
a cause -- the same shape the corpus work kept finding.

## By what the Unknown span CONTAINS

    1118  "}"          the single largest bucket
     164  ")"
      52  ":"
      38  "is $s, join('', map chr,"
      35  "like $@, _create_mismatc"
      33  ","
      31  "->"

A bare `}` as an Unknown span means a block did not close and the brace was
left over. That bucket is eight times the next one.

## What actually causes it: measured in isolation, and most guesses were wrong

Presence in a dirty file is not causation, and testing each construct ALONE
changed the priority list completely:

    labelled bare block   SKIP: { ... }           Unknown=1
    the same with last    SKIP: { last SKIP; }    Unknown=2
    given/when/default                            Unknown=2
    eval BLOCK                                    Unknown=0   <- not a cause
    eval BLOCK then if                            Unknown=0   <- not a cause
    for my ($k,$v) (%h)                           Unknown=0   <- not a cause
    sub() { 42 }                                  Unknown=0   <- not a cause
    plain bare block                              Unknown=0
    if/else                                       Unknown=0

`eval BLOCK` appeared in **177 of 435 dirty files** and parses perfectly. Had
this document been written from the presence counts it would have aimed the
first fix at the largest population and moved nothing. `for my ($k,$v)` and
`sub()` are the same trap at smaller scale.

All three refusing constructs are valid perl 5.42, verified:

    perl -e 'SKIP: { print "in\n"; last SKIP; }'                  -> in
    perl -e 'use feature "switch"; given (1) { when (1) {...} }'   -> one
    perl -e 'use v5.36; for my ($k,$v) (%h) {...}'                 -> a=1

## Reach of the two real causes

A LABELLED BARE BLOCK appears in 148 of the 620 files and in 83 of `op/`'s 227.
It is the only construct measured so far whose reach matches the size of the
`}` bucket, and `last LABEL` inside one costs a second node.

GIVEN/WHEN appears in only 3 dirty files -- `op/switch.t` and two others -- but
accounts for ~45 of the brace-context hits because that one file is dense with
it. Small reach, cheap fix, and it is a deprecated feature that still compiles.

## What is NOT yet measured, and must be before any more work is planned

The two causes above account for at most a few hundred of 7,745 nodes. The
remaining mass is unattributed, and the honest statement is that this document
does not yet know what it is. Specifically:

- The `}` bucket is 1,118 nodes; a labelled block explains some unknown share
  of it. The brace-context probe walked back to the nearest `{`, which finds an
  INNER brace when blocks nest, so its top three entries (`}`, `)`, empty) are
  probe artifacts rather than findings.
- `not_a_term` at 1,584 nodes in 192 files has had no bucketing pass at all.
- The `like $@, qr/.../` and `is $s, join(...)` spans that recur are
  parenless-call shapes. Today's declared/undeclared work closed those when the
  callee is declared in-file; these come from `test.pl`, which is `require`d at
  runtime and deliberately unresolved.

## The chain

Issues, in dependency order. Each is measurement-first because the measurements
above overturned three of four guesses.

1. **Bucket `not_a_term`'s 1,584 nodes by what the span contains.** No fix. The
   `trailing_tokens` bucketing produced this document; the second code has had
   none.
2. **Attribute the `}` bucket properly.** The probe that produced it has a known
   artifact. Walk to the MATCHING open brace rather than the nearest one, and
   report what share a labelled block actually owns.
3. **A labelled bare block is a statement.** `SKIP: { ... }` and `last LABEL`
   inside it. Blocked by (2), because its reach is the thing (2) measures.
4. **given/when/default.** Three files, deprecated, cheap. Not blocked.
5. **Decide what `require`d-at-runtime `test.pl` means for this goal.** Perl's
   own suite gets `ok`, `is`, `like` and `plan` from a file it `require`s, which
   is deliberately unresolved (`use.go:118`). Every `like $@, qr/.../` span is
   downstream of that decision. It may be that this goal cannot reach 620/620
   without resolving it, and that is a decision rather than a defect.

Items 1, 2 and 5 produce no code. That is deliberate: 435 files is too many to
fix one at a time, and the only two causes measured so far were found by
isolating constructs rather than by counting their appearances.
