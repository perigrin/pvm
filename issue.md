---
title: "Five scanner rows: literal ranges, qx, <<>>, angle globs, leading ::"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-17T14:23:38.848509109Z
updated: 2026-09-17T14:23:38.848509109Z
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

- [ ] `1..2` and `1...2` lex as three tokens, and `$a..$b` still works (`go test ./internal/lexer/ -run '^TestNumericRangeIsNotOneToken$' -count=1 -v | grep -q '^--- PASS: TestNumericRangeIsNotOneToken'`)
- [ ] `my @r = (1..5)` gives a `Range` whose endpoints are numbers, not a `Str` (`go test ./internal/parse/ -run '^TestLiteralRangeHasEndpoints$' -count=1 -v | grep -q '^--- PASS: TestLiteralRangeHasEndpoints'`)
- [ ] `1.5`, `1_000`, `0x1f` and `1e10` still lex as one number each (`go test ./internal/lexer/ -run '^TestNumberFormsStillLex$' -count=1 -v | grep -q '^--- PASS: TestNumberFormsStillLex'`)
- [ ] `qx/ls/` and `qx{ls}` lex as one quote-like token (`go test ./internal/lexer/ -run '^TestQxLexes$' -count=1 -v | grep -q '^--- PASS: TestQxLexes'`)
- [ ] `<<>>` lexes as one readline, and `$a << $b` is still a shift (`go test ./internal/lexer/ -run '^TestDoubleDiamond$' -count=1 -v | grep -q '^--- PASS: TestDoubleDiamond'`)
- [ ] `<*.c>` lexes as a glob, and `$a < $b` is still a comparison (`go test ./internal/lexer/ -run '^TestAngleGlob$' -count=1 -v | grep -q '^--- PASS: TestAngleGlob'`)
- [ ] `$::x`, `@::a` and `$::Foo::x` lex as variables (`go test ./internal/lexer/ -run '^TestLeadingPackageSeparator$' -count=1 -v | grep -q '^--- PASS: TestLeadingPackageSeparator'`)
- [ ] The lexer round-trip and the golden token stream both hold (`go test ./internal/lexer/ -count=1 2>&1 | grep -qv FAIL`)
- [ ] T1 clean does not fall below 408 and Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestT1Ratchet$' -count=1 -v | grep -q '^--- PASS: TestT1Ratchet'`)

Each negative criterion is paired with its positive on purpose: every one of
these rows is a case where the same bytes mean something else in operator
position, which is what `internal/lexer/expect.go` exists for. A row added
without its negative is how `%h = ()` once lexed as modulus.

The T1 AC depends on `TestT1Ratchet`, owned by the M1 gate and not yet
written. The lexer's own ratchet and golden stream already exist and will
move; both regenerate under explicit flags.

Spec: §4.14.2 (`Range{Exclusive}`, `StrLit.Backtick`, `Readline{Magic}`,
`Glob`, `Var{Sigil,Name}`). Plan: M1.
