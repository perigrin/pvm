---
title: The readline BUILTIN is not a readline site, only the <> syntax is
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:15:49.825526237Z
updated: 2026-10-09T17:21:19.182936984Z
sessions:
- start_sha: 09e38fa97c039422168735f3da9b607b14677d80
  end_sha: 34b103f60e9548502ff3afe1f4cbb943b3397863
  commits: 1
  started_at: 2026-09-19T20:38:30.029913515Z
  ended_at: 2026-09-19T20:59:06.235065848Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T20:38:30.029913515Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T20:59:06.235065848Z
observed_paths:
- internal/parse/readline_test.go
- internal/parse/sites.go
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

## Acceptance criteria

- [x] `readline $fh` reports a readline site (`go test ./internal/parse/ -run '^TestReadlineBuiltinIsASite$' -count=1 -v | grep -q '^--- PASS: TestReadlineBuiltinIsASite'`)
- [x] The angle form still reports one, and `<*.c>` glob still does not (`go test ./internal/parse/ -run '^TestReadlineBuiltinIsASite$' -count=1 -v | grep -q '^--- PASS: TestReadlineBuiltinIsASite'`)
- [x] WRONG readline sites fall, measured (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2.


## Done, 2026-09-19 (34b103f6)

    corpus readline sites  25 -> 4
    corpus WRONG sites    163 -> 142, files 51 -> 45
    corpus exact          401 -> 403
    full suite green, 0 FAIL across the repo

## The test found a second, older defect

isReadlineTerm accepted ANY angle-delimited leaf, so a GLOB was reported as
a readline -- wrong before this change and in the other direction. Measured:

    <>  <$fh>  <FH>  <STDIN>  <My::Handle>  <a:b>    readline
    <*.c>  <$h{x}>  <a b>                            glob

readline iff the content is empty, a plain `$scalar`, or a bareword. perl's
rule is looser than a package name -- `<a:b>` is readline -- so this follows
the measurement rather than the tidier guess.

## Three corrections, each caught by measuring

**1.** My identifier loop was ASCII-only, so `<hòฟ>` became a glob and
t/uni/readline.t gained 6 WRONG sites. perl compiles it to readline. Rune
iteration now -- the same defect 0db016fd fixed in the lexer's quote guard,
and twice in one session is a pattern worth naming.

**2.** `$obj->readline` is a METHOD: Binary "->" whose right operand is a
resolved Call, where perl emits entersub. My first comment claimed that
shape "never reaches here"; it does.

**3.** Skipping the method NAME must not skip its ARGUMENTS. Dropping the
whole right subtree lost the reference in `threads->create(\&f, $i)` --
op/threads.t went clean -> 3 WRONG srefgen and re/pat.t gained two, taking
srefgen 38 -> 43. With arguments walked it is 38 again.

Each is a regression row in the test.

## The remaining 4 are the -n/-p switches

    run/switcha.t:1   #!./perl -na
    run/switchn.t:1   #!./perl -n
    run/switchp.t:1   #!./perl -p
    run/switchF.t:1   #!./perl -anFx+

These wrap the program in an implicit `while (<>)`, so perl emits a readline
the SOURCE does not contain. A static parser cannot see that without reading
the shebang and synthesising a loop.

Worth its own decision later: whether the subject should DECLINE such a file
rather than answer it, which would move these from WRONG to no-answer and is
the honest bucket for "the source I was given is not the program perl ran".
Not filed yet -- four sites is below the bar for an issue until someone
wants the row.
