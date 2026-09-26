---
title: A parenthesised and/or/xor refuses, and the bare form is a different program
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T04:10:38.190903837Z
updated: 2026-09-26T04:10:38.190903837Z
---

`($a and $b)` refuses with `not_a_term`, and so do `or` and `xor`. The
bare forms parse. The gap is the three LOW-PRECEDENCE word operators in a
parenthesised expression.

Filed by `crochet:chain-review` round 2, which found this case owned by
nobody: the glossary issue (`01a0dba9`) explicitly disclaims it, the
citation issue (`01a0d0c7`) never mentions it, and the issue that would
have covered it was cancelled on a false premise.

## Measured, at 226f38db

    bare   $a and $b       clean
    paren ($a and $b)      UNKNOWN   not_a_term
    bare   $a or $b        clean
    paren ($a or $b)       UNKNOWN   not_a_term
    bare   $a xor $b       clean
    paren ($a xor $b)      UNKNOWN   not_a_term
    paren ($a cmp $b)      clean
    paren ($a x 3)         clean
    paren, no assign: print(($a and 1), "\n")   UNKNOWN

`not $a` parenthesised is CLEAN -- it is prefix, not infix, so it is not
part of this. `cmp` and `x` are clean because they bind tighter than `=`
and the parser reaches them by a different route; only the three operators
BELOW assignment in precedence fail.

So this is not "word-shaped operators are unimplemented". It is narrower:
a word operator at perlop's lowest levels is not accepted as an infix
operator inside a parenthesised term.

## Why the parenthesised form is the one that matters

perl, 5.42.0 -- the paren changes the PROGRAM, not just the style:

    $ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = ($a and $b);'
    my $y = $a && $b;

    $ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = $a and $b;'
    $b if my $y = $a;

Bare `and` binds looser than `=`, so the assignment swallows `$a` and the
statement becomes a modifier on `$b`. Parenthesised, it is an ordinary
`&&`. The corpus case `comparison.md`'s "`&&` and `and` are one op at two
precedences" exists to document exactly that cliff, and we parse the half
that regroups while refusing the half that does not.

## What refuses today

Three corpus cases, all `refuses: unfiled` with `refusal: not_a_term`
where recorded:

    comparison.md   The five logical operators          ($a xor $b)
    comparison.md   `&&` and `and` are one op at two
                    precedences                         see note
    adjacency-04_operators.md   The whole tier in one body

The middle case is SHARED with `01a0dba9`: its source has a bare
`$a and 9`, which parses, and its failing claim is the token fact `one
operator whose text is "and"`. It is listed here because the citation
sweep will touch the same `refuses:` line -- coordinate, do not both write
it.

## Context

- paths: `internal/parse/expr.go`, `internal/parse/call.go`, `conformance/mdtest/comparison.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/parse/precedence.go:59-61` (`or`/`xor` at level 4, `and` at 5) and `internal/parse/call.go:343` (`startsTerm`)

## What is already right, so the fix is narrower than it looks

Measured before filing, so the work does not re-litigate settled code:

  - `and`, `or` and `xor` ARE in the precedence table already
    (`precedence.go:59-61`), at perlop's levels 4 and 5, with `xor`
    correctly sharing `or`'s level and a `toke.c:9224` citation.
  - `startsTerm` ALREADY knows a Word that is an infix operator does not
    start a term (`call.go:347`, changed for this exact shape in
    `01a0cf3e`).

So both halves of the knowledge exist. What fails is reaching them from
inside a parenthesised term: the paren's sub-expression is not parsed down
to level 4/5. That is the thing to measure first.

## Do not

Do not fix this by adding `and`/`or`/`xor` to the term-starter set. The
`01a0cf3e` change went the other way deliberately -- a Word that is an
operator is NOT a term -- and reversing it would reintroduce the bug that
change fixed (`print FOO eq "x" ? ...`).

## Acceptance criteria

- [ ] The parenthesised low-precedence word operators parse (`go test ./internal/parse/ -run '^TestParenthesisedWordInfixParses$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedWordInfixParses'`)
- [ ] The bare forms still parse, so the fix did not trade one shape for the other (`go test ./internal/parse/ -run '^TestBareWordInfixStillRegroups$' -count=1 -v | grep -q '^--- PASS: TestBareWordInfixStillRegroups'`)
- [ ] The five-logical-operators case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/comparison.md/The_five_logical_operators'`)
- [ ] `not` is still prefix and still clean (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] The T2 ratchet moved in the right direction or held (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
