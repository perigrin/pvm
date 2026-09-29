---
title: "The yada statement ... refuses: sub f { ... } is perl's own placeholder body"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T05:11:56.035357072Z
updated: 2026-09-29T05:11:56.035357072Z
---

The yada-yada statement `...` refuses everywhere:

    sub f { ... }             Unknown=1
    eval { $i++; ...; 1 };    Unknown=1

`...` is the range operator's spelling in the infix table, and nothing reads it
at a statement start. perl allows it ONLY as a whole statement. Measured on
5.42.0:

    $ perl -MO=Deparse -e 'sub f { ... }'
    sub f { die 'Unimplemented'; }
    $ perl -c -e 'my $x = ...;'        syntax error near "= ..."
    $ perl -c -e '... if 0;'           syntax error near "... if"

So it is read at a statement start, followed by a terminator or closer, with no
modifier -- and still refuses as a term.

Reach: op/yadayada.t (6 Unknowns); base/lex.t line 507, `map{...} @_`;
op/const-optree.t.

## Acceptance criteria

- [ ] `...` as a whole statement parses and canons to a fixpoint (`go test ./internal/parse/ -run '^TestYadaYadaStatement$' -count=1 -v | grep -q '^--- PASS: TestYadaYadaStatement'`)
- [ ] NEGATIVE: `...` as a term still refuses (`go test ./internal/parse/ -run '^TestYadaYadaStatement$' -count=1 -v | grep -q '^--- PASS: TestYadaYadaStatement'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
