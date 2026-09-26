---
title: The identifier scanner eats the digits of x3, losing binary x against a bare number
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T09:56:37.076237627Z
updated: 2026-09-26T09:56:37.076237627Z
---

The identifier scanner eats the digits of `x3`, so binary `x` against a bare
number is lost. Found by the TIER_2 gate ground of `01a0ce57` and confirmed
independently.

## Measured

Perl, 5.42.0 at `/home/perigrin/.local/bin/perl`, all three spellings are
one reading:

    perl -MO=Deparse -e 'my @a = (1)x3;'      ->  my(@a) = (1) x 3;
    perl -MO=Deparse -e 'my @a = (1) x 3;'    ->  my(@a) = (1) x 3;
    perl -MO=Deparse -e 'my $s="a"; $s x3;'   ->  $s x 3;

Our lexer emits `Word(x3)` for the glued spelling: `scanIdentRunes` takes
the digits, because a digit is a legal non-initial identifier byte, and by
the time anything could ask "was that `x` followed by a number" the run is
already consumed.

## Why it is not the `x=` gap

`01a0ce57` taught `takeRepeatAssign` to extend a word-position `x` by one
`=`. This reproduces with and without that fix: the failure is in the
identifier run, not in the operator suffix, and the character that gets
swallowed is a digit rather than `=`.

## A docstring that is wrong about its own example

`internal/lexer/expect.go:14` lists

    (1)x3       repetition    -- operator position

as a settled case of the expect state deciding repetition. The state is
indeed right there -- operator position is expected -- and the token is
still wrong, because the decision happens in a scanner that never consults
it. A comment naming an example that does not work is worse than no
example.

## Scope

1. In operator position, an identifier run beginning `x` and continuing
   only with digits is the binary `x` plus a Number. Decide whether the
   narrow rule (`x` exactly, then digits) or the general one (any
   word-shaped operator glued to a number) is correct -- measure `cmp3`,
   `lt3`, `and3` against perl before choosing, because the general rule may
   be wrong.
2. Correct `expect.go:14`'s example, either by making it true or by citing
   one that is.

## Context

- paths: `internal/lexer/scan.go` (`scanWord`, `scanIdentRunes`,
  `takeRepeatAssign`), `internal/lexer/expect.go:14`
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: `scanWord`, immediately after `scanIdentRunes` returns

## Acceptance criteria

- [ ] `(1)x3` lexes as a repetition, not one Word (`go test ./internal/lexer/ -run '^TestRepetitionGluedToNumber$' -count=1 -v | grep -q '^--- PASS: TestRepetitionGluedToNumber'`)
- [ ] `expect.go`'s cited example is one that holds (`go test ./internal/lexer/ -run '^TestExpectDocstringExamples$' -count=1 -v | grep -q '^--- PASS: TestExpectDocstringExamples'`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
