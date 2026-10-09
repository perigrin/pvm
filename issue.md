---
title: The identifier scanner eats the digits of x3, losing binary x against a bare number
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T09:56:37.076237627Z
updated: 2026-10-09T05:12:54.42810467Z
sessions:
- start_sha: 98d028499fdad30e12a1e2e54cc03ddb455833f4
  end_sha: 93c26e8152e5d8e6b6914956ecf5014d163c7342
  commits: 2
  started_at: 2026-10-09T04:48:36.970381097Z
  ended_at: 2026-10-09T05:12:54.42810467Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T04:48:36.970381097Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T05:12:54.42810467Z
observed_paths:
- internal/lexer/repetitionglued_test.go
- internal/lexer/scanrows_test.go
- internal/parse/t2_test.go
---

The identifier scanner eats the digits of `x3`, so binary `x` against a bare
number is lost. Found by the TIER_2 gate ground of `01a0ce57` and confirmed
independently by the file-test work, where it costs a node in `base/lex.t`.

## Measured

Perl, 5.42.0 at `/home/perigrin/.local/bin/perl`, all three spellings are one
reading:

    perl -MO=Deparse -e 'my @a = (1)x3;'      ->  my(@a) = (1) x 3;
    perl -MO=Deparse -e 'my @a = (1) x 3;'    ->  my(@a) = (1) x 3;
    perl -MO=Deparse -e 'my $s="a"; $s x3;'   ->  $s x 3;

Our lexer emits `Word(x3)` for the glued spelling: `scanIdentRunes` takes the
digits, because a digit is a legal non-initial identifier byte, and by the time
anything could ask "was that `x` followed by a number" the run is already
consumed.

## Why it is not the `x=` gap

`01a0ce57` taught `takeRepeatAssign` to extend a word-position `x` by one `=`.
This reproduces with and without that fix: the failure is in the identifier
run, not in the operator suffix, and the character swallowed is a digit rather
than `=`.

## A docstring that is wrong about its own example

`internal/lexer/expect.go:14` lists

    (1)x3       repetition    -- operator position

as a settled case of the expect state deciding repetition. The state IS right
there -- operator position is expected -- and the token is still wrong, because
the decision happens in a scanner that never consults it. A comment naming an
example that does not work is worse than no example.

## THE SCOPE QUESTION IS SETTLED: the NARROW rule is correct

Step 1 below asked whether the rule should be `x` exactly or any word-shaped
operator glued to a number, and said to measure `cmp3`, `lt3`, `and3` first.
Measured 2026-09-26, and the general rule would be WRONG:

    perl -e 'my $r = 1 cmp3;'
      Bareword found where operator expected (Missing operator before "cmp3"?)
    perl -e 'my $r = 1 lt3;'    same
    perl -e 'my $r = 1 and3;'   same

So `x` is the ONLY word-shaped operator perl glues to a following number. A
general rule over the word-shaped operator set would make three syntax errors
parse.

The rest of the adjacent set, for whoever writes the test:

    (1)x3            ->  (1) x 3
    5x3              ->  '555'            folded, repetition of a number
    5x-1             ->  5 x -1           negative count, not folded
    "a"x3            ->  'aaa'            folded
    $s x3            ->  $s x 3           runtime operand

`0-5x-l{0}` is perl #123711's case and reads as `0 - (5 x (-l {0}))`; it is a
second file where this gap costs a node, measured by the file-test work in
`base/lex.t`.

The B::SoN session offered to act as an optree oracle for this family and has
the first three measured on its side; route any further `x`-adjacent question
there rather than re-deriving.

## Scope

1. In operator position, an identifier run beginning `x` and continuing only
   with digits is the binary `x` plus a Number. The NARROW rule -- `x` exactly,
   then digits -- is the correct one; see above.
2. Correct `expect.go:14`'s example, either by making it true or by citing one
   that is.

## Context

- paths: `internal/lexer/scan.go` (`scanWord`, `scanIdentRunes`,
  `takeRepeatAssign`), `internal/lexer/expect.go:14`
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: `scanWord`, immediately after `scanIdentRunes` returns

## Acceptance criteria

- [ ] `(1)x3` lexes as a repetition, not one Word (`go test ./internal/lexer/ -run '^TestRepetitionGluedToNumber$' -count=1 -v | grep -q '^--- PASS: TestRepetitionGluedToNumber'`)
- [ ] `cmp3`, `lt3` and `and3` still lex as one Word, because perl rejects them as operators (`go test ./internal/lexer/ -run '^TestRepetitionGluedToNumber$' -count=1 | grep -q '^ok'`)
- [ ] `expect.go`'s cited example is one that holds (`go test ./internal/lexer/ -run '^TestExpectDocstringExamples$' -count=1 -v | grep -q '^--- PASS: TestExpectDocstringExamples'`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
