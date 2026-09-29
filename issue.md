---
title: "`5x3` is one Number: a decimal literal takes every letter after it"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:32:18.260426258Z
updated: 2026-09-29T21:32:18.260426258Z
---

A decimal literal takes every word byte after its digits, so `5x3` lexes as Number("5x3"), `1if $b` as Number("1if") and `0-5x-l{0}` as Number("5x") followed by `-l`. toke.c's scan_num takes a decimal's `e` only when a sign, digit or `_` follows it (toke.c:13142-13144), and it takes no other letter. Radix literals (0x, 0b, 0o, and a leading-zero octal such as the octal float `01.1p0`) keep their letters.

Measured with perl 5.42:

    $ perl -MO=Deparse -e '0-5x-l{0}; $a=5x3; print 1if $b; $c = 1eq 1;'
    0 - 5 x (-l {0});
    $a = '555';
    print 1 if $b;
    $c = !0;

scanVString shares the run, and it had been taking `v10.v257` whole. It now stops before a dot with no digit after it (perl measures `"\n\x{101}"`, a concatenation), so perl.git t/op/lc.t:198 and t/op/pack.t:926 stay clean. The dotless `v10` Word is the gap scanVString already documents.

perl.git t/base/lex.t:516.

## Acceptance criteria

- [ ] A decimal literal ends at a letter that is not its exponent (`go test ./internal/lexer/ -run '^TestDecimalEndsAtLetter$' -count=1 -v | grep -q '^--- PASS: TestDecimalEndsAtLetter'`)
