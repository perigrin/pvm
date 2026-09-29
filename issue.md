---
title: "Whitespace between a sigil and its name ends the variable: `my $ bits` splits"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T08:05:59.943691725Z
updated: 2026-09-29T08:05:59.943691725Z
---

perl allows whitespace between a sigil and its name. The lexer ends the variable at the space, so `my $ bits = sub { ... }->()` became `my $ ;bits() = ...` at Unknown=0, and now refuses as trailing tokens.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $ bits = 1; print $ bits;'
    my $bits = 1;
    print $bits;

perl.git t/op/caller.t:279. toke.c's scan_ident skips spaces after the sigil (`s = skipspace(s)` when the next byte is not a name).

## Acceptance criteria

- [ ] `$ name` with whitespace after the sigil is one variable (`go test ./internal/lexer/ -run '^TestSpaceAfterSigil$' -count=1 -v | grep -q '^--- PASS: TestSpaceAfterSigil'`)
