---
title: "Whitespace between a sigil and its name ends the variable: `my $ bits` splits"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:05:59.943691725Z
updated: 2026-09-29T08:36:16.427391626Z
---

perl allows whitespace between a sigil and its name. The lexer ends the variable at the space, so `my $ bits = sub { ... }->()` became `my $ ;bits() = ...` at Unknown=0, and now refuses as trailing tokens.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $ bits = 1; print $ bits;'
    my $bits = 1;
    print $bits;

perl.git t/op/caller.t:279. toke.c's scan_ident skips spaces after the sigil (`s = skipspace(s)` when the next byte is not a name).

The same cause reaches signatures, where yyl_sigvar skips space after the sigil too. perl.git t/op/signatures.t:1009 (t085) spells a signature one token per line, `$\n    a\n    ,`, and t086/t087 put comments between sigil and name. Measured with perl 5.42, `sub t085 ( $\n a , $ b = 333 ) { 1 }` deparses as `sub t085 ($a, $b = 333) {`.

## Acceptance criteria

- [ ] `$ name` with whitespace after the sigil is one variable (`go test ./internal/lexer/ -run '^TestSpaceAfterSigil$' -count=1 -v | grep -q '^--- PASS: TestSpaceAfterSigil'`)
