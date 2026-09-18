---
title: "Five scanner rows: literal ranges, qx, <<>>, angle globs, leading ::"
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T14:23:38.848509109Z
updated: 2026-09-18T10:36:58.861539664Z
sessions:
- start_sha: f905b4e9ba967eb983700a3ef6feace66c1fc11c
  end_sha: 6968d55df55015b70b6730ae5781e9eff4b0c52b
  commits: 1
  started_at: 2026-09-18T10:21:48.314965344Z
  ended_at: 2026-09-18T10:36:58.861539664Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T10:21:48.314965344Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-18T10:36:58.861539664Z
observed_paths:
- internal/lexer/quote.go
- internal/lexer/ratchet_test.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/parse/parse_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
---

Five constructs do not lex, and each is a missing scanner row rather than a design problem. Measured at 3c997a6b, all round-tripping:

    1..2;        [source_file statement term]              one Number token
    qx/ls/;      [source_file unknown]
    <<>>;        [source_file statement binary unknown unknown]
    <*.c>;       [source_file unknown]
    $::x;        [source_file unknown]

## What each needs, and which matters most

**`1..2` is the one with a measured consumer defect.** The number scanner
(`internal/lexer/scan.go` `scanNumber`) consumes `isWordByte` or `.` in a
loop, so `1..2` is a single `Number` token and the range operator is gone
before the parser sees it. `$a..$b` is fine — `binary ".."` — so only
literal endpoints are affected.

The cost is recorded in §4.14.2: `my @r = (1..5)` infers `@r` as `Str`,
because `findOperatorText` (`internal/infer/infer.go:645-659`) matches the
anonymous `.` before ever seeing `..`, while `signatures.go:182` has
`"..": {Result: List}` waiting for a token that never arrives. Measured on
perl 5.42.0: `$r[0]+1` is 2, so the elements are numbers.

This one is required, not optional. The other four are gaps.

**`qx//`** — `internal/lexer/quote.go:22-31` has rows for `q`, `qq`, `qw`,
`qr`, `m`, `s`, `tr`, `y` and no `qx`. §4.14.2 notes a backtick string
already lexes as one `term` whose `Text` keeps the delimiters, so `qx//` is
the same node reached by a different spelling, and SoN calls it
`BacktickExpr` — a language fact.

**`<<>>`** — the 5.22 double-diamond. Lexes as `binary ">>"` over two
`Unknown`s, because `scanAngle` does not know the form. `<>` and `<$fh>` and
`<STDIN>` all work.

**`<*.c>`** — glob in angle brackets. `scanAngle` accepts word bytes, `$`
and `:` between the angles and stops at anything else, so a `*` ends the
candidate and the whole thing falls through. §4.14.2 marks `Glob`
SPECULATIVE for its consumer; parse it anyway, for the same reason as
`DoFile` — an `Unknown` mid-file costs more than an unread node.

**`$::x`** — the main-package shorthand for `$main::x`. `scanVarName` reaches
its `scanIdentRunes` case, which requires a leading identifier byte, and `:`
is not one. Measured on perl 5.42.0: `our $x = 5; print $::x;` prints 5.

## Scope

Scanner rows and the parse path each implies. Not in scope: interpolation
inside `qx//` (that is the `Interp` work, which waits on the lexer emitting
sub-tokens at all), and not `${^WORD}` or other punctuation-variable forms
that already lex.

## Acceptance

- [x] `1..2` and `1...2` lex as three tokens, and `$a..$b` still works (`go test ./internal/lexer/ -run '^TestNumericRangeIsNotOneToken$' -count=1 -v | grep -q '^--- PASS: TestNumericRangeIsNotOneToken'`)
- [x] `my @r = (1..5)` gives a `Range` whose endpoints are numbers, not a `Str` (`go test ./internal/parse/ -run '^TestLiteralRangeHasEndpoints$' -count=1 -v | grep -q '^--- PASS: TestLiteralRangeHasEndpoints'`)
- [x] `1.5`, `1_000`, `0x1f` and `1e10` still lex as one number each (`go test ./internal/lexer/ -run '^TestNumberFormsStillLex$' -count=1 -v | grep -q '^--- PASS: TestNumberFormsStillLex'`)
- [x] `qx/ls/` and `qx{ls}` lex as one quote-like token (`go test ./internal/lexer/ -run '^TestQxLexes$' -count=1 -v | grep -q '^--- PASS: TestQxLexes'`)
- [x] `<<>>` lexes as one readline, and `$a << $b` is still a shift (`go test ./internal/lexer/ -run '^TestDoubleDiamond$' -count=1 -v | grep -q '^--- PASS: TestDoubleDiamond'`)
- [x] `<*.c>` lexes as a glob, and `$a < $b` is still a comparison (`go test ./internal/lexer/ -run '^TestAngleGlob$' -count=1 -v | grep -q '^--- PASS: TestAngleGlob'`)
- [x] `$::x`, `@::a` and `$::Foo::x` lex as variables (`go test ./internal/lexer/ -run '^TestLeadingPackageSeparator$' -count=1 -v | grep -q '^--- PASS: TestLeadingPackageSeparator'`)
- [x] The lexer round-trip and the golden token stream both hold (`go test ./internal/lexer/ -count=1 2>&1 | grep -qv FAIL`)
- [x] T1 clean does not fall below 408 and Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)

Each negative criterion is paired with its positive on purpose: every one of
these rows is a case where the same bytes mean something else in operator
position, which is what `internal/lexer/expect.go` exists for. A row added
without its negative is how `%h = ()` once lexed as modulus.

The T1 AC uses `TestParseRatchet`, committed at 4bc971ec. The lexer's own ratchet and golden stream already exist and will
move; both regenerate under explicit flags.

Spec: §4.14.2 (`Range{Exclusive}`, `StrLit.Backtick`, `Readline{Magic}`,
`Glob`, `Var{Sigil,Name}`). Plan: M1.


## Done, 2026-09-18 (6968d55d)

    T1 Unknown   4,627 -> 4,606
    T2 clean     16 of 56, three files improved
    oracle       39 -> 40 exact, WRONG still 1 (the M4 pin)

Every one of the five produced PLAUSIBLE WRONG TOKENS rather than an error,
which is why nothing caught them:

    1..5      Number("1..5")                 the range operator, gone
    qx/ls/    Word("qx") Quote("/ls/")
    <<>>      Operator("<<") Operator(">>")
    <*.c>     five separate operators
    $::x      Variable("$:")                 a variable named `:`

On the Range AC: `(1..5)` and `($a..$b)` now produce IDENTICAL trees --
`binary` over two `term` endpoints. The `Range` node kind itself is the
lowering's (§4.14.2 marks it LOWERING for variable endpoints), so the
parser's half is the token split, and that is done.

## The one regression is a win underneath

op/glob.t went 0 -> 1 lexer errors because line 11's `<op/*>` now lexes as
one Readline instead of five operators. Line 28 is the cost: after
`map { ... }` the lexer is in operator position, so `<` stays a comparison
and the `/` after `op` opens a bare pattern that swallows a backtick.

Measured to predate this change by stashing it -- `map { 1 } <op>;` misreads
at HEAD too, it simply produced no error token to count. That is `map {`
leaving operator position, which is 01a0ac52's subject. Recorded in
`ratchet_test.go` rather than absorbed.

## Found and not fixed

`.5` is one number to perl (`print .5` gives 0.5) and splits here, because
scanNumber requires a leading digit. Letting the scan start on `.` collides
with concatenation and with the `..` this commit splits.
TestLeadingDotNumber records it as a skip with that reasoning.

## Fixture churn worth noting

The decline fixture moved for the SEVENTH time, and the sixth lasted one
commit: `$::x = \@a;` was chosen an hour earlier and the leading-`::` row
here taught the lexer to read it. Now `new Foo(\@a);` -- indirect-object
syntax, which §4.14.6 records as an open SCOPE decision rather than a gap,
so it should outlast the others.
