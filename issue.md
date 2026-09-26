---
title: scanNumber stops early on a signed exponent and never forms a vstring
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:08:00.127866819Z
updated: 2026-09-26T02:08:00.127866819Z
---

Two tier-01 cases refuse on a token fact with a clean parse, and both are
real lexer gaps rather than a vocabulary dispute.

They are separated from the four word-shaped-operator facts deliberately:
those four are a glossary question, these two are simply wrong output.

## Measured, at b0f1a9e9

**The exponent sign.** `numeric-point.md`, "The sign of an exponent is
part of the literal":

    my $x = 5e-1;
      Word("my") Variable("$x") Operator("=")
      Number("5e") Operator("-") Number("1") Semicolon(";")

The literal is split three ways. The case asserts both halves of the
claim and both fail:

    one numeric literal whose text is "5e-1"    got 0, want 1
    no operator whose text is "-"               got 1, want 0

perl reads `5e-1` as one number:

    $ perl -e 'print 5e-1'
    0.5

Note the second fact is the FALSIFIABLE pair of the first -- it fails
today for the same reason, and both must move together.

**Vstrings.** `vstrings.md`, two cases:

    my $v = 65.66.67;     no numeric literal whose text is "65.66.67"
    my $v = v65.66.67;

Neither forms a vstring. perl:

    $ perl -e 'my $v = 65.66.67; print $v'      ABC
    $ perl -e 'my $v = v65.66.67; print $v'     ABC

Both spellings produce the same three-character string, which is what
makes the bare form a literal rather than two `.` operators on numbers.

## Why they belong together

Both are `scanNumber` deciding where a numeric literal ENDS -- one
stopping too early at `e-`, one not recognising that a second `.` changes
the literal's kind entirely. The glossary already records that this
scanner's precedence is load-bearing:

> The `.` in `.5` is not an operator, and `scanNumber` running before
> operator scanning is what makes that true.

That is the same rule these two need extended, and `01a0c13f-97f5`
(leading decimal, done) is the precedent.

## Context

- paths: `internal/lexer/scan.go`, `conformance/mdtest/numeric-point.md`, `conformance/mdtest/vstrings.md`
- commands: `go test ./internal/lexer/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/lexer/scan.go` (`scanNumber`)

## Scope

Fix the lexer, then replace `refuses: unfiled` with nothing -- the cases
should PASS, not point at an issue. If either turns out larger than it
looks, file it and cite the id; do not leave `unfiled` behind.

## Acceptance criteria

- [ ] An exponent with a sign is one numeric literal (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] Both vstring spellings lex as one literal (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/vstrings.md'`)
- [ ] The exponent case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/numeric-point.md'`)
- [ ] Neither topic leaves a refusal unfiled (`! grep -q 'refuses: unfiled' conformance/mdtest/vstrings.md`)
- [ ] The exponent topic leaves no refusal unfiled (`! grep -q 'refuses: unfiled' conformance/mdtest/numeric-point.md`)
- [ ] The tier-01 lint and perl validation still hold (`go test ./internal/conformance/ -run '^TestTierLiterals' -count=1 | grep -q '^ok'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
