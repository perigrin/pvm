---
title: "A sub attribute ends the sub early: sub :lvalue leaves its body trailing"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:41:11.595550955Z
updated: 2026-09-26T16:41:11.595550955Z
---

A sub with an attribute -- `sub :lvalue { 1 }`, `sub :prototype($) { 1 }` --
refuses, anonymous and named alike. The same sub without the attribute parses
clean, so the defect is the attribute list between the name and the body.

## Measured 2026-09-26 at `2dc301fd`

Each program is the whole file:

    my $s = sub :lvalue { 1 };           Unknown=1  trailing_tokens ":lvalue { 1 }"
    my $s = sub :prototype($) { 1 };     Unknown=1  trailing_tokens ":prototype($) { 1 };"
    sub f :lvalue { 1 }                  Unknown=1  trailing_tokens ":lvalue { 1 }"

    my $s = sub { 1 };                   Unknown=0  <- no attribute is fine
    sub x () { 8 }                       Unknown=0  <- a prototype in parens is fine

The span shows the parser finishing the sub at `sub` and leaving the attribute
and the whole body as trailing tokens, which is why one attribute costs the rest
of the statement.

Valid perl 5.42.0:

    /home/perigrin/.local/bin/perl -e 'my $s = sub :lvalue { 1 }; print $s->(),"\n"'     -> 1
    /home/perigrin/.local/bin/perl -e 'sub f :lvalue { 1 } print f(),"\n"'               -> 1
    /home/perigrin/.local/bin/perl -e 'my $s = sub :prototype($) { 1 }; print $s->(1)'   -> 1

## Reach

Appears in 26 of perl.git t/'s 620 files, 25 of them dirty. `t/op/sub_lval.t`
and `t/op/lex.t` are the dense cases; `:lvalue` dominates, with `:prototype`
and `:method` in the tail.

## Not to be confused with the cascade

`not_a_term` is 1,584 nodes but only 41 are the first Unknown in their file; the
other 1,543 are resync fallout. This construct is one of the 41-node cause
surface. See the `not_a_term` section of
`docs/plans/2026-09-26-perl-git-t-parses.md`.

## Acceptance criteria

- [ ] `my $s = sub :lvalue { 1 };` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `sub f :lvalue { 1 }` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] An attribute taking an argument, `sub :prototype($) { 1 }`, parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `sub x () { 8 }` and `sub { 1 }` still parse, no regression (`go test ./internal/parse/ -count=1`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -count=1`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
