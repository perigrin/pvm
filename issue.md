---
title: The block form of the filehandle slot refuses
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-23T23:54:24.963405482Z
updated: 2026-09-23T23:54:24.963405482Z
---

`print {$fh} "x"` is the block form of the filehandle slot and it
refuses. `parseFilehandleSlot` has a branch for it that does not yield
a parsing statement.

Found by the gate on `01a0cf3e`. Pre-existing: it refuses at
`831eaa1d^` too, so no change in this milestone caused it.

## Measured

Valid perl:

    $ perl -MO=Deparse -e 'open my $fh, ">", "/dev/null"; print {$fh} "x";'
    open my $fh, '>', '/dev/null';
    print {$fh;} 'x';

Our parser:

    print {$fh} "x";      REFUSED  [source_file unknown]

The other three forms are fine:

    print STDERR "x";           parses
    print $fh "x";              parses
    printf STDERR "%s", "x";    parses

## Where it is

`internal/parse/call.go`, `parseFilehandleSlot`:

    case tok.OpensBlock && p.text(tok) == "{":
        // `print {$fh} "a"` -- the block form, which exists precisely to
        // disambiguate an expression in the slot.
        return p.parseBlock(tok)

The branch exists and returns a block node; something downstream does
not accept it where a handle belongs. Worth checking whether
`OpensBlock` is even set here -- the lexer decides that with perl's
`intuit_curly`, and `{$fh}` is the shape a hash subscript also wears.

## Why it matters

This is the form that EXISTS to disambiguate. `print $x "a"` is
ambiguous between a handle and a first argument, and `print {$x} "a"`
is how a program says it meant the handle. A parser that refuses the
disambiguator cannot read code written to be unambiguous.

## Acceptance criteria

- [ ] The block form of the filehandle slot parses (`go test ./internal/parse/ -run '^TestBlockFilehandleSlot$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The other three handle forms still parse (`go test ./internal/parse/ -run '^TestHandleSlotSurvivesTheWordFix$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The whole suite is green (`make test`)
