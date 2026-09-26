---
title: scanNumber stops early on a signed exponent and never forms a vstring
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:08:00.127866819Z
updated: 2026-09-26T04:15:31.959483583Z
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

**Vstrings.** `vstrings.md`, two cases, and they fail DIFFERENTLY -- the
first draft of this issue quoted only one fact and showed the other as bare
source, which hid that:

    my $v = 65.66.67;    lexes as ONE Number("65.66.67")
                         right boundary, WRONG CATEGORY
                         fails: no numeric literal whose text is
                         "65.66.67"   (got 1, want 0)

    my $v = v65.66.67;   lexes as Word("v65") Operator(".") Number("66.67")
                         WRONG BOUNDARY, two facts fail
                         fails: no operator whose text is "."
                                no word whose text is "v65"

So the bare form is a category error (a number where perl has a string)
and the `v` form is a segmentation error (`versionPrefix` in
`internal/lexer/ident.go` claims `v65` for the word scanner before
`scanNumber` sees it). One root -- `scanNumber` has no dot-counting -- two
distinct symptoms, and each case's facts say which.

perl:

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

## Criteria corrected by chain-review, 2026-09-26

The first draft's criteria were vacuous. `grep -q 'PASS: TestCorpus/numeric-point.md'`
passes TODAY with the bug unfixed, because `TestCorpus` emits one subtest
per CASE and numeric-point has four cases that already pass -- so a
file-level grep matches a sibling. Each criterion below names its case, and
each was verified to FAIL now (rc=1) and therefore to be falsifiable.

The falsifiable pair is a SCOPE requirement rather than its own criterion:
`no operator whose text is "-"` fails today for the same reason as
`one numeric literal whose text is "5e-1"`, and both are token facts on the
same case. The named-case criterion covers both, because `verdict()`
collects every failing fact on a case before deciding -- a fix that merged
the literal while leaving a stray `-` token would keep the case SKIPping
and the criterion failing. A separate grep would have been a second,
weaker copy of that.

## Acceptance criteria

- [ ] The signed-exponent case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/numeric-point.md/The_sign_of_an_exponent_is_part_of_the_literal'`)
- [ ] The bare vstring case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/vstrings.md/Bare:_three_dot-separated_parts'`)
- [ ] The v-prefix case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/vstrings.md/The_.v._prefix'`)
- [ ] Neither vstring case leaves a refusal recorded (`! grep -q 'refuses:' conformance/mdtest/vstrings.md`)
- [ ] The exponent topic leaves no refusal recorded (`! grep -q 'refuses:' conformance/mdtest/numeric-point.md`)
- [ ] The lexer's own suite still passes (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
