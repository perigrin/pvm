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
- `not_a_term` HAS now been bucketed; see the two tables below. The finding is
  that 1,543 of its 1,584 nodes cannot be causes at all.
- The `like $@, qr/.../` and `is $s, join(...)` spans that recur are
  parenless-call shapes. Today's declared/undeclared work closed those when the
  callee is declared in-file; these come from `test.pl`, which is `require`d at
  runtime and deliberately unresolved.

### === not_a_term bucket table, by what the span contains ===

Measured 2026-09-26 at `2dc301fd`, over the same 620 files: 7,745 Unknown, 435
dirty, 1,584 `not_a_term` in 81 distinct span buckets. The top 25; every bucket
below the 25th is a single node in a single file.

    nodes  files  span (leading 24 bytes, newlines escaped)
     1118    157  "}"
      164     59  ")"
       52     23  ":"
       33      6  ","
       31     15  "->"
       16     11  ";"
       15      5  "::"
       11      6  "="
       11      7  "};"
       10      7  ");"
        8      3  "..."
        7      1  ".."
        7      2  ":;"
        7      3  "]"
        6      4  "*"
        5      2  "()"
        5      2  "^"
        4      1  "=>"
        4      1  "before Apack.pm at - lin"
        4      1  "no warnings 'experimenta"
        2      1  "        use threads;\n   "
        2      1  "Assembling pattern from "
        2      1  "Bareword found where ope"
        2      1  "before: Apack.pm\nbefore:"
        2      1  "package DB;\nsub DB::sub "

**The table is almost worthless on its own, and that is the finding.** 97% of
`not_a_term`'s mass is one-character punctuation: `}`, `)`, `:`, `,`, `->`. A
one-byte span names no construct. Grouping the 55 bytes PRECEDING each span
gave 1,049 distinct contexts for the 1,118 `}` nodes -- a long tail of uniques,
which is the signature of a cascade rather than of a construct.

**What separates cause from cascade: is the node the FIRST Unknown in its file?**

    not_a_term as its file's FIRST Unknown      41 nodes in  41 files
    not_a_term with an earlier Unknown before   1543 nodes

1,543 of 1,584 (97.4%) appear only after the parser has already refused
something earlier in the same file. They are fallout from a prior refusal
resynchronising, and no fix aimed at them can exist. **`not_a_term` is a
1,584-node bucket with a 41-node cause surface.** The whole of the remaining
work in this code is in those 41 files.

### === not_a_term: each construct measured ALONE ===

All 41 first-in-file cases, reduced to the smallest program holding one
construct, then narrowed until the refusal either survived or vanished. Perl
5.42.0 (`/home/perigrin/.local/bin/perl`) ran every candidate: **all eleven
refusing forms are valid perl**, so none of these refusals is correct.

REFUSES in isolation:

    my $r = $::{foo};                 Unknown=1  not_a_term ":"
    my @k = keys %::;                 Unknown=1  not_a_term ":;"
    ::ok(1);                          Unknown=1  trailing_tokens
    my $r = ::f();                    Unknown=1  not_a_term "::"
    my sub x () { 8 }                 Unknown=1  not_a_term "()"
    my sub x ($a) { 8 }               Unknown=1  not_a_term "($a)"
    my $s = sub :lvalue { 1 };        Unknown=1  trailing_tokens ":lvalue { 1 }"
    sub f :lvalue { 1 }               Unknown=1  trailing_tokens ":lvalue { 1 }"
    my @x = grep -e, @l;              Unknown=1  not_a_term ","
    my $x = -e;                       Unknown=1  not_a_term ";"
    for my ($a,) (1) { }              Unknown=1  not_a_term ")"

NOT A CAUSE -- parses clean in isolation. This is the valuable half, and it is
what stops the next person aiming a fix at a construct that already works:

    $ {$CX} = 17;                     Unknown=0   space between sigil and brace
    my $y = ${^XY};                   Unknown=0   caret var in braces
    formline $f, "a";                 Unknown=0
    our sub x { 1 }                   Unknown=0   NO prototype
    state sub x { 1 }                 Unknown=0   NO prototype
    my sub x { 8 }                    Unknown=0   NO prototype
    goto &g;                          Unknown=0
    goto LABEL;                       Unknown=0
    my $x = -t STDIN;                 Unknown=0   filetest WITH an operand
    my $x = -e $f;                    Unknown=0   filetest WITH an operand
    my @a = (1, <<'EOF', 2);          Unknown=0   heredoc as a list element
    my %h = (time => 1);              Unknown=0
    my $x = (time\n=>);               Unknown=0   the `=>` quoting a keyword
    my $v = bless \(my $d = 1), __PACKAGE__;      Unknown=0
    $main'object = 1;                 Unknown=0   apostrophe package separator
    threads->create( sub { 1 } );     Unknown=0
    local *Foo::m = sub { 1 };        Unknown=0
    my $r = \&UNIVERSAL::isa;         Unknown=0
    my $re = qr/ (?i: a{2} ) /x;      Unknown=0
    { my @a; @a=(1..4)\n}             Unknown=0   no semicolon before `}`
    my $r = $Foo::{bar};              Unknown=0   a NAMED package's stash
    my @k = keys %Foo::;              Unknown=0   a NAMED package's stash
    *Foo:: = \%Bar::;                 Unknown=0
    main::ok(1);                      Unknown=0   the explicit `main::` form
    my $x = \%Foo::;                  Unknown=0
    my ($a,) = (1);                   Unknown=0   trailing comma in plain `my`
    for my ($a,$b) (1,2) { }          Unknown=0   the well-formed list
    sub x () { 8 }                    Unknown=0   a NAMED sub's prototype
    my $s = sub { 1 };                Unknown=0   anon sub with NO attribute

Three of those negatives narrow a refusal to its real edge, and the narrowing
is the point:

- A lexical sub refuses only WITH a parameter list. `my sub x { 8 }` is clean;
  `my sub x () { 8 }` is not. The defect is the signature/prototype on
  `my`/`our`/`state sub`, not lexical subs.
- A stash refuses only for the ANONYMOUS `main` stash. `$Foo::{bar}` and
  `keys %Foo::` are clean; `$::{bar}` and `keys %::` are not. The defect is the
  empty package name before `::`, not stash access.
- A filetest refuses only with NO operand. `-e $f` is clean; `-e` and `grep -e,`
  are not. The defect is the implicit-`$_` form.
- `(time\n=>)` was measured a cause and is NOT one; the `=>`-quotes-a-keyword
  hit in `lex.t` was cascade from an earlier refusal in that file.

### Reach of the refusing constructs, over the same 620 files

    stash elem   $PKG::{name} / %::    in  58 files,  56 of them dirty
    leading ::   ::name(...)           in  84 files,  76 of them dirty
    anon/named sub w/ attribute        in  26 files,  25 of them dirty
    filetest w/ no operand  -e,        in  19 files,  19 of them dirty
    lexical sub w/ sig or proto        in   6 files,   6 of them dirty
    for my (list) w/ empty slot        in   1 file,    1 of them dirty

`for my (list)` with an empty slot is `op/for-many.t` alone -- a deliberate
torture file. It is 63 of the `,` bucket's nodes and one file, so it is not
worth an issue of its own.

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

### Filed out of the not_a_term bucketing

Item 1 is done, and it produced four fixes -- each one a construct that refuses
in isolation and appears in more than a handful of files. Ordered by reach:

    01a0de97-3c5d-714d-9ded-d68fd50ad296   leading ::, 84 files
    01a0de97-77fe-733d-8482-cc0c6ad25a51   $::{n} and %::, 58 files
    01a0de97-96ac-791e-bbff-d0b15fa70f60   sub :lvalue etc, 26 files
    01a0de97-b3bd-7c49-a55b-f99acee8f932   filetest with no operand, 19 files

Two refusing constructs were deliberately NOT filed, because reach does not
justify an issue:

- A lexical sub with a signature or prototype -- `my sub x () { 8 }`,
  `my sub x ($a) { 8 }`, and the same for `our` and `state` -- refuses, and is
  valid perl. Six files. Worth folding into whichever issue next touches the
  sub-declaration parse site rather than carrying alone.
- `for my ($a,) (1)` with an empty list slot refuses, and is valid perl. One
  file, `op/for-many.t`, which exists to torture that syntax.
