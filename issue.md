---
title: A filetest with no operand defaults to $_, and -e alone refuses
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:41:19.036325043Z
updated: 2026-09-26T16:41:19.036325043Z
---

A filetest operator with no operand -- `-e`, `-t`, `grep -e, @list` -- refuses.
Perl defaults such a filetest to `$_`. The same operator WITH an operand parses
clean, so the defect is the implicit-`$_` form alone.

## Measured 2026-09-26 at `2dc301fd`

Each program is the whole file:

    my $x = -e;              Unknown=1  not_a_term ";"
    my $x = -t;              Unknown=1  not_a_term ";"
    my @x = grep -e, @l;     Unknown=1  not_a_term ","
    my @x = grep { -e } @l;  Unknown=1  trailing_tokens "-e } @l;"

    my $x = -e $f;           Unknown=0  <- WITH an operand is fine
    my $x = -t STDIN;        Unknown=0  <- WITH an operand is fine

Valid perl 5.42.0:

    /home/perigrin/.local/bin/perl -e '$_ = "/etc"; my $x = -e; print "ok $x\n"'         -> ok 1
    /home/perigrin/.local/bin/perl -e 'my @x = grep -e, ("/etc"); print "ok @x\n"'       -> ok /etc
    /home/perigrin/.local/bin/perl -e 'my @x = grep { -e } ("/etc"); print "ok @x\n"'    -> ok /etc

The `grep { -e }` case is worth naming separately: it refuses with
`trailing_tokens` rather than `not_a_term`, because the operand-less filetest
inside the block leaves the block's close as trailing. One fix at the filetest
parse site should close both.

## Reach

Appears in 19 of perl.git t/'s 620 files, all 19 dirty. `t/op/stat_errors.t` is
the dense case -- it iterates every filetest letter with no operand -- and
`t/op/filetest.t` and `t/porting/*` supply the rest.

## Not to be confused with the cascade

`not_a_term` is 1,584 nodes but only 41 are the first Unknown in their file; the
other 1,543 are resync fallout. This construct is one of the 41-node cause
surface. See the `not_a_term` section of
`docs/plans/2026-09-26-perl-git-t-parses.md`.

## Acceptance criteria

- [ ] `my $x = -e;` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `my @x = grep -e, @l;` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `my @x = grep { -e } @l;` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `-e $f` and `-t STDIN` still parse, no regression (`go test ./internal/parse/ -count=1`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -count=1`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
