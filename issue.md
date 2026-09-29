---
title: The lexer splits ~~ into two bitwise nots in operator position, so every smartmatch breaks its expression
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:25:06.896787187Z
updated: 2026-09-29T04:25:06.896787187Z
---

The lexer emits `~~` as two `~` tokens, so a smartmatch splits the expression
at Unknown=0 or refuses:

    my $r = $x ~~ @a;       canon  my $r = $x;~~@a;
    $d = $v ~~ 1;           Unknown=1

The parser already has `~~` in its precedence and comparison tables; the
operator never reached it. toke.c lexes it as one token only in operator
position:

    if (FEATURE_SMARTMATCH_IS_ENABLED &&
        s[1] == '~' && (PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR))
        ... NCEop(OP_SMARTMATCH);

In term position `~~$x` is two bitwise nots, an idiom for forcing scalar
context, and must stay that way. Measured on 5.42.0: `$x ~~ @a` deparses as
`($x ~~ (\@a))`, and `my $n = ~~$x;` unchanged.

Reach: op/smartmatch.t, op/tie_fetch_count.t, op/switch.t and more.

## Acceptance criteria

- [ ] `~~` is one operator in operator position and two bitwise nots in term position (`go test ./internal/lexer/ -run '^TestSmartmatchIsOneOperator$' -count=1 -v | grep -q '^--- PASS: TestSmartmatchIsOneOperator'`)
- [ ] REGRESSION GUARD: the lexer's golden token stream is unchanged (`go test ./internal/lexer/ -count=1`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
