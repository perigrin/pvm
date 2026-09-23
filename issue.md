---
title: Forty-nine negative token facts name text their source never contains
state: pending
urgency: normal
milestone: corpus-complete
created: 2026-09-23T19:16:11.345924198Z
updated: 2026-09-23T19:16:11.345924198Z
---

Forty-nine of the corpus's 115 negative token facts assert nothing.
Measured by `TestNegativeFactsNameTextTheSourceContains`, which landed
with them ratcheted in `knownVacuous` so the check could arrive without
going red.

## The measurement

`checkTokenFact` compares `src[tk.Start:tk.End]` against the claimed
text. So a token whose text is `X` requires the bytes `X` somewhere in
the source. A source not containing those bytes can produce no such
token under ANY lexing, correct or broken -- the count is always zero,
a `no` fact always passes, and it holds nothing up.

Verified by hand on `04_operators/07_precedence.t`, whose fact is
`no operator whose text is "**"` and whose source is:

    my $a = $ARGV[0] // 2;
    my $b = $ARGV[1] // 3;
    my $c = $ARGV[2] // 4;
    my $default = $a + $b * $c;
    my $grouped = ($a + $b) * $c;
    print "$default $grouped\n";

No `**` anywhere. The fact cannot fail.

## Spread

    10  08_references
    10  04_operators
     8  11_oo
     6  06_control
     5  03_context
     3  07_subroutines
     3  02_variables
     2  12_packages
     1  10_io
     1  01_literals

Ten tiers, which is why this is its own issue rather than part of the
one that found it.

## Two vacuity mechanisms, and this is the second

`TestStringLiteralFactsSpellTheirDelimiters` catches a fact naming text
the LEXER can never emit -- a `string literal` fact without its
delimiters, or an operator spelling absent from `scan.go`'s table.

This is the other one: text the lexer CAN emit but THIS SOURCE cannot
produce. A reader checking only the table passes it, which is exactly
how `no operator whose text is "<="` shipped in
`32_chained_comparison.t` in the same commit that fixed two facts of
the first kind.

## What the repair is, per fact

Each is a fact to DELETE OR REPLACE, not to defend. Three shapes:

- The fact names a construct the file is ABOUT but never writes --
  `no operator whose text is "**"` in a precedence file that uses `+`
  and `*`. Replace with a spelling the source contains, or drop it.
- The fact names a near-miss the file's construct could be mis-lexed
  INTO. Those are the valuable ones and they need the spelling to be
  reachable: `27_incdec.t`'s `no operator whose text is "+"` works
  because the source writes `++`, whose bytes contain a `+`.
- No honest fact exists for the file. Drop the section and say why in
  the header, as `30_unary_plus.t` does.

## How it lands

Remove each entry from `knownVacuous` as its fact is repaired. The
check fails on a listed entry that is no longer vacuous, so the list
cannot rot -- it shrinks as the work lands and reaches zero when the
map can be deleted along with the ratchet branch.

## Acceptance criteria

- [ ] Every negative token fact names text its own source contains (`grep -q 'knownVacuous = map\[string\]bool{}' internal/conformance/vacuous_test.go`)
- [ ] The vacuity check passes with no ratcheted entries (`go test ./internal/conformance/ -run '^TestNegativeFactsNameTextTheSourceContains$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The whole conformance suite is green with the repairs in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
