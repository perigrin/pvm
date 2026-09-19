---
title: The readline BUILTIN is not a readline site, only the <> syntax is
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:15:49.825526237Z
updated: 2026-09-19T18:15:49.825526237Z
---

`isReadlineTerm` (sites.go:299) matches a leaf whose text is angle-delimited,
so `<FH>`, `<$fh>` and `<>` are reported and the `readline` FUNCTION is not
reported at all. Perl emits the same op for both:

    $ perl -MO=Concise,-exec -e 'open my $fh,"<","/dev/null"; my $l = readline $fh;' | grep -c readline
    1

20 of the 25 WRONG readline sites are the builtin, measured over the 620-file
corpus at afdd06c9:

    op/readline.t   10    $line = readline $fh;  readline($yunk);
                          readline $f{g};  readline *{$f{g}};  readline undef;
    io/pvbm.t        2    ok (!eval { readline $pvbm }, ...)
    io/socketpair.t  2    is(readline($b), "aa\n", ...)
    op/gv.t          2    my $x = readline $v{v};
    porting/ss_dup.t 2    while (readline $fh) {
    uni/gv.t         2    my $x = readline $Ｖ{Ｖ};
    op/flip.t        1    readline(UNKNOWN);

The doc comment says the angle form is trusted because "scanAngle emits a
Readline token only in term position, so a leaf whose text is
angle-delimited is the lexer's own decision". That reasoning is sound and
says nothing about the builtin, which arrives as an ordinary Call.

## The other five are the -n/-p switches

    run/switchF.t:1   #!./perl -anFx+
    run/switcha.t:1   #!./perl -na
    run/switchn.t:1   #!./perl -n
    run/switchp.t:1   #!./perl -p

`-n` and `-p` wrap the program in an implicit `while (<>) { ... }`, so perl
emits a readline the SOURCE does not contain. A static parser cannot see
that without reading the shebang and synthesising a loop. Out of scope here;
worth its own decision about whether the subject should decline such a file
rather than answer it.

## Acceptance

- [ ] `readline $fh` reports a readline site (`go test ./internal/parse/ -run '^TestReadlineBuiltinIsASite$' -count=1 -v | grep -q '^--- PASS: TestReadlineBuiltinIsASite'`)
- [ ] The angle form still reports one, and `<*.c>` glob still does not (`go test ./internal/parse/ -run '^TestReadlineBuiltinIsASite$' -count=1 -v | grep -q '^--- PASS: TestReadlineBuiltinIsASite'`)
- [ ] WRONG readline sites fall, measured (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2.
