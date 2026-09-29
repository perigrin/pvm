---
title: "`$ $name` is not a deref: a spaced second sigil is not seen"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:43:12.746102284Z
updated: 2026-09-29T12:43:12.746102284Z
---

A deref sigil separated by spaces from the sigil it applies to -- `$ $name1`, `@ $r` -- does not lex as a deref: startsDerefExpression skips spaces before a brace but checks the byte right after the sigil for a second `$`.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $x = $ $name1; my @y = @ $r; my $z = $$ ;'
    my $x = $$name1;
    my(@y) = @$r;
    my $z = $$;

perl.git t/op/postfixderef.t:224, `is ($ $name1, undef, ...)`.

## Acceptance criteria

- [ ] A spaced second sigil makes a deref, and `$$ ;` stays the pid (`go test ./internal/lexer/ -run '^TestSpacedDeref$' -count=1 -v | grep -q '^--- PASS: TestSpacedDeref'`)
